# Performance Guidelines

## Architecture Overview

uhc-auth-proxy is a single-binary HTTP proxy that authenticates OpenShift cluster requests against UHC account management APIs. It uses an in-memory cache, a shared HTTP client, and mutex-protected critical sections. There are no goroutines launched by application code; all concurrency comes from `net/http` serving requests on separate goroutines.

## Critical Sections and Lock Ordering

### 1. Cache lock (`cache/cache.go`)

The cache uses a package-level `sync.Mutex` protecting a `map[string]item`. The lock is held only for the map read or write operation, never across I/O.

Rules:
- Never hold the cache mutex while performing network calls or any blocking I/O.
- Always lock/unlock around a single map operation (get or set), never across compound operations.
- Do not add a `sync.RWMutex` without profiling first; the current lock hold time is sub-microsecond and contention is unlikely to be a bottleneck relative to the upstream HTTP call.
- Expired entries are detected on read but never evicted. If adding eviction, use a separate goroutine with its own timer rather than scanning under the write lock.

### 2. Token lock (`requests/client/access.go`)

`GetToken` uses a `sync.Mutex` with `defer mutex.Unlock()` and holds the lock across a network call (`fetch`) when the token is expired. This function exists in the codebase but is not currently called from the production request path.

Rules:
- If `GetToken` is wired into the request flow, be aware it is a serialization point: when the token expires, all concurrent requests block behind a single `fetch` call. Do not make this worse by adding work inside the critical section.
- If you need to reduce contention, use a `singleflight.Group` keyed on the offline access token instead of a bare mutex.
- Never cache the token outside this mutex. The single-writer pattern here prevents token races.

### 3. Lock ordering

The cache package has one active mutex. `requests/client/access.go` defines a second mutex not currently invoked in the production request path. **Do not introduce a call path that acquires both locks** (e.g., calling `cache.Set` while holding the token mutex). If you must, always acquire the token lock first, then the cache lock.

## HTTP Client

### Shared client (`requests/client/wrapper.go`)

A single `http.Client` is initialized as a package-level variable in `wrapper.go` using `viper.GetDuration("TIMEOUT_SECONDS") * time.Second` for the timeout. Because this variable is initialized before `config.go`'s `init()` runs, the `viper.SetDefault("TIMEOUT_SECONDS", 30)` default has not yet been registered at that point. In practice, the client timeout is controlled entirely by the `TIMEOUT_SECONDS` environment variable, which must be set before the process starts. If the environment variable is absent, the client has no timeout. Both account management calls in `wrapper.go` and SSO token-fetch calls in `access.go` use this shared client.

Rules:
- Do not create per-request `http.Client` instances. The shared client reuses connections via the default `http.Transport` connection pool.
- The `TIMEOUT_SECONDS` value is read once at package variable initialization time. Changing the env var at runtime has no effect. If dynamic timeout configuration is needed, set the timeout on the `http.Request` context instead.
- The default `http.Transport` has `MaxIdleConns: 100` and `MaxIdleConnsPerHost: 2`. Since this service talks to a small number of upstream hosts, consider setting `MaxIdleConnsPerHost` to a higher value (e.g., 10-20) on a custom transport if connection reuse becomes a bottleneck.
- `io.ReadAll` is used to consume response bodies in both `wrapper.go` and `access.go`. For this service the payloads are small JSON (< 10 KB). Do not switch to streaming unless payload sizes grow significantly.

### Bug: nil-pointer dereference in `Do`

In `wrapper.go`, `resp.StatusCode` is accessed before the `err != nil` check. If `client.Do` returns `(nil, err)`, this panics. Any fix must preserve the Prometheus counter increment for all responses but move it after the nil check.

## Cache Behavior

Rules:
- Cache TTL is hardcoded to 2 hours (`cache.go`). Do not change this without understanding the downstream impact: a shorter TTL increases load on UHC account management APIs; a longer TTL delays propagation of account changes.
- Cache keys are `clusterID:authorizationToken` (`server.go` `makeKey`). This means a token rotation immediately bypasses the cache, which is correct behavior.
- The cache grows unboundedly. In practice this is acceptable because the number of distinct cluster+token pairs is finite and entries are small (`[]byte` of marshaled `Identity` JSON). If the service starts handling significantly more clusters, add an LRU eviction policy or bounded map.
- Always call `cache.Clear()` in test `BeforeEach` blocks to prevent cross-test cache pollution.

## Prometheus Metrics

| Metric | Type | Location | Labels |
|--------|------|----------|--------|
| `uhc_auth_proxy_request_time` | Histogram | `wrapper.go` | `url` |
| `uhc_auth_proxy_to_acct_mgmt_request_status` | Counter | `wrapper.go` | `code` |
| `uhc_auth_proxy_cache_hit` | Counter | `server.go` | none |
| `uhc_auth_proxy_cache_miss` | Counter | `server.go` | none |
| `uhc_auth_proxy_responses` | Counter | `server.go` | `code` |

Rules:
- Use `promauto` for registration (existing convention). Never call `prometheus.MustRegister` directly.
- The `url` label on `request_time` uses the full URL string. This is safe only because the service calls a fixed set of upstream URLs. Do not add metrics with unbounded label cardinality (e.g., cluster IDs, request IDs).
- Histogram buckets are tuned for sub-second to 5-second responses. If upstream latency characteristics change, adjust the buckets to avoid most observations falling into a single bucket.

## Server Configuration

| Env Var | Default | Effect |
|---------|---------|--------|
| `SERVER_PORT` | `8080` | Listen port |
| `TIMEOUT_SECONDS` | (none — must be set via env var) | HTTP client timeout (read at package init, before viper defaults apply) |
| `ACCESS_TOKEN_URL` | `https://sso.redhat.com/...` | SSO token endpoint |
| `CURRENT_ACCOUNT_URL` | `https://api.openshift.com/...` | Account lookup endpoint |
| `CLIENT_ID` | (none) | OAuth client ID |

Rules:
- The server does not configure `ReadTimeout`, `WriteTimeout`, or `IdleTimeout` on `http.Server`. If adding these, set `WriteTimeout` greater than `TIMEOUT_SECONDS` to avoid killing requests waiting for a slow upstream response.
- No graceful shutdown is implemented. `ListenAndServe` runs until process termination. If adding graceful shutdown, drain the cache and flush CloudWatch logs before exiting.

## Middleware Stack

The chi router uses these middlewares in order: `request_id`, `RealIP`, `Logger`, `Recoverer`, `StripSlashes`.

Rules:
- `Recoverer` catches panics per-request. Do not remove it.
- `Logger` runs on every request including `/metrics` and `/status`. If log volume becomes a problem, add a path-based filter to skip health check endpoints.

## What NOT to Optimize

- **String splitting in `getClusterID`**: Runs once per request on a short string. The linear scan over 9 operator prefixes is negligible.
- **JSON marshal/unmarshal of Identity**: Payloads are tiny. Do not introduce code generation or pooling for this.
- **Package-level init functions**: The viper defaults and logger initialization run once at startup. They are not hot paths.
