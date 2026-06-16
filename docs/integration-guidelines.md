# Integration Guidelines

Rules and patterns for making external HTTP calls in uhc-auth-proxy.

## Architecture Overview

All outbound HTTP calls go through a single path:

1. `requests/client/wrapper.go` — shared HTTP client, auth headers, metrics, error handling
2. `requests/cluster/cluster.go` — account management service facade
3. `server/server.go` — maps upstream errors to HTTP responses, caches identity results

## External Endpoints

All external URLs are configured via environment variables with defaults set in `init()` functions:

| Variable | Default | Package |
|---|---|---|
| `CURRENT_ACCOUNT_URL` | `https://api.openshift.com/api/accounts_mgmt/v1/current_account` | `cluster` |
| `GET_ACCOUNTID_URL` | `https://api.openshift.com/api/accounts_mgmt/v1/cluster_registrations` | `cluster` |
| `ACCOUNT_DETAILS_URL` | `https://api.openshift.com/api/accounts_mgmt/v1/accounts/%s` | `cluster` |
| `ORG_DETAILS_URL` | `https://api.openshift.com/api/accounts_mgmt/v1/organizations/%s` | `cluster` |
| `ACCESS_TOKEN_URL` | `https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token` | `client` |
| `TIMEOUT_SECONDS` | (none — must be set via env var) | `client` |

Use `viper.GetString()` to read endpoint URLs at call time rather than storing them at init time. This allows tests and deployments to override values after init. Note that `TIMEOUT_SECONDS` is read once at package load time (in the `var` block of `wrapper.go`, before viper defaults apply) and cannot be changed at runtime.

## The Wrapper Pattern

All outbound HTTP calls to the account management API go through the `client.Wrapper` interface:

```go
type Wrapper interface {
    Do(req *http.Request, label string, cluster_id string, authorization_token string) ([]byte, error)
}
```

Rules:
1. Never call `http.Client.Do()` directly from business logic. Always go through a `Wrapper`.
2. The `label` parameter is used for Prometheus metrics, not routing. Pass the URL string (e.g., the value returned by `viper.GetString("CURRENT_ACCOUNT_URL")`), not the variable name.
3. `Do()` adds Authorization, Accept, and Content-Type headers automatically. Do not set these on the request before calling `Do()`.
4. The Authorization header uses the format `AccessToken {cluster_id}:{authorization_token}`. This is not a Bearer token — it is specific to the accounts management API.

## HTTP Client Configuration

A single package-level `http.Client` is shared across all requests in `requests/client/wrapper.go`.

Rules:
1. Do not create additional `http.Client` instances. Use the shared client.
2. The timeout is set at package load time from `TIMEOUT_SECONDS`. There is no retry logic anywhere in the codebase. If a request fails, it fails.
3. There is no custom Transport, connection pooling config, or TLS settings — the Go defaults apply.

## Error Handling

### Two-tier error model

`wrapper.Do()` returns `([]byte, error)` where both can be non-nil simultaneously:
- **Status >= 400**: Returns the response body AND a `*client.HttpError` with `StatusCode` and `Message`.
- **Transport error**: Returns `nil` body and the underlying error.

Rules:
1. When `wrapper.Do()` returns an error, always check whether the body (`[]byte`) is also non-nil. A non-nil body on error means the upstream returned a structured error response.
2. Use `errors.As(err, &httpError)` to extract the status code from `*client.HttpError`.
3. Default to HTTP 401 when the error is not an `HttpError`.

### AccountError (structured upstream errors)

When `GetCurrentAccount` receives both a body and an error, it attempts to unmarshal the body into `*cluster.AccountError`:

```go
if b != nil {
    res := &AccountError{}
    if json.Unmarshal(b, res) == nil {
        res.Inner = err
        return nil, res
    }
}
```

Rules:
1. `AccountError` wraps the original `HttpError` via its `Inner` field and implements `Unwrap()`. Use `errors.As` to reach either layer.
2. When logging an `AccountError`, use `zap.Object("account_error_verbose", accErr)` — it implements `zapcore.ObjectMarshaler`.
3. Any new external call that returns structured error JSON should follow this same pattern.

## Known Bug: Nil dereference in wrapper.Do

There is a latent nil-pointer risk in `wrapper.go`: `resp.StatusCode` is accessed before the `err != nil` check. If the HTTP call fails with a transport error (DNS failure, timeout), `resp` will be nil and this line will panic. Any fix must move the metric increment to after the nil check.

## Caching

- Successful identity lookups are cached for 2 hours in `cache/cache.go`. Cache key is `{cluster_id}:{authorization_token}`.
- Cache is checked before calling `GetIdentity`. If the cache returns non-nil, no external call is made.
- Only cache successful results. Errors are never cached.
- Call `cache.Clear()` at the start of every test that exercises the handler.

## SSO Token Management

`requests/client/access.go` manages OAuth tokens for the SSO endpoint:
1. Tokens are cached in a package-level variable with a mutex.
2. A token is refreshed when expired (`now >= expires`) or when the cached token string is empty.
3. This is a `refresh_token` grant type flow against Red Hat SSO.

## Testing Patterns

### Use the in-repo fakes, not mocks

| Fake | Purpose |
|---|---|
| `FakeWrapper` | Returns a preconfigured `*Account` response |
| `ErrorWrapper` | Always returns an error with nil body |
| `ErrorWithBodyWrapper` | Returns both a JSON body and a `*client.HttpError` |

When testing error paths that involve structured error bodies, use `ErrorWithBodyWrapper` and set both `AccountError` and `StatusCode`.

### Always clear cache between tests

Call `cache.Clear()` in `BeforeEach` blocks. The global in-memory cache persists across test cases.

### Test error unwrapping explicitly

Verify error chain with `errors.As` or `errors.Unwrap`, not string matching:

```go
unwrap := errors.Unwrap(err)
Expect(unwrap).To(BeAssignableToTypeOf(&AccountError{}))
```

### Valid user-agent format in handler tests

Requests must have a user-agent matching `{operator-prefix}/{version} cluster/{cluster_id}`. The recognized operator prefixes are defined in `server/server.go`.

## Prometheus Metrics

Any new external call must record request duration and status code using the existing `promauto` registration pattern. Use the `uhc_auth_proxy_` prefix for all metric names.
