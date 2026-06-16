# API Contracts Guidelines

## Overview

uhc-auth-proxy is an authentication proxy that validates OpenShift operator requests against Red Hat's account management service and returns an identity JSON payload. It has exactly two functional endpoints and a narrow set of accepted clients.

## Routes

| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| GET | `/` | `RootHandler` | Primary auth endpoint |
| GET | `/api/uhc-auth-proxy/v1` | `RootHandler` | Versioned alias of `/` |
| GET | `/status` | `StatusHandler` | Health/readiness check |
| ANY | `/metrics` | `promhttp.Handler` | Prometheus metrics |

The `/` and `/api/uhc-auth-proxy/v1` routes are registered with `r.Get` and share the same handler variable. The `/metrics` route is registered with `r.Handle`, which accepts all HTTP methods, though Prometheus scrapers always use GET. Any new auth behavior added to the shared handler automatically applies to both auth routes.

## Request Contract

### Required Headers

Every request to the auth endpoint **must** include both headers. Missing or malformed values return `400`.

#### User-Agent

Format: `<operator-prefix><version> cluster/<cluster-id>`

The operator prefix must be one of these exact strings (including the trailing slash):
- `insights-operator/`
- `cost-mgmt-operator/`
- `marketplace-operator/`
- `acm-operator/`
- `assisted-installer-operator/`
- `cryostat-operator/`
- `openshift-lightspeed-operator/`
- `jws-operator/`
- `runtimes-inventory-operator/`

The second space-delimited segment must begin with `cluster/`. The cluster ID is everything after that prefix.

When adding a new operator, you must:
1. Add a `const` with the prefix string in `server.go` and add it to the `operatorPrefixes` array (update the array size literal). Note: existing constants use a `*Prefix` naming convention (e.g., `insightsOperatorPrefix`), except `acmOperator` which does not follow this pattern.
2. Add the operator name **without** the trailing slash (e.g., `"my-operator"`) to the `validOperatorAgents` slice in `server_test.go`.

#### Authorization

Format: `Bearer <token>`

Must start with exactly `Bearer ` (capital B, single space). The token value after the prefix must be non-empty or the request will fail with 500 (empty token creates an incomplete Registration).

### Request ID Propagation

The proxy reads `x-rh-insights-request-id` from the incoming request and uses `request_id.ConfiguredRequestID` middleware to generate/propagate it. Both the original header value and the middleware-generated ID are logged on every request.

## Response Contract

### Success (200)

Content-Type: `application/json`

```json
{
  "account_number": "<ebs_account_id>",
  "org_id": "<external_id>",
  "type": "System",
  "internal": {
    "org_id": "<external_id>"
  },
  "system": {
    "cluster_id": "<cluster_id>"
  }
}
```

Key facts:
- `type` is always the literal string `"System"` — never vary this.
- `account_number` comes from `organization.ebs_account_id` in the upstream account.
- `org_id` (top-level) and `internal.org_id` both come from `organization.external_id`. Keep them in sync.
- `system.cluster_id` is echoed from the request's User-Agent, not from the upstream service.
- Successful responses are cached for 2 hours keyed on `<cluster_id>:<token>`. Cached responses are served as raw bytes, so they are byte-identical to the original.

### Error Responses

Errors are returned as **plain text** (not JSON). Only success responses set `Content-Type: application/json`.

| Code | Condition | Body prefix |
|------|-----------|-------------|
| 400 | Invalid User-Agent | `Invalid user-agent:` |
| 400 | Invalid Authorization header | `Invalid authorization header:` |
| 500 | Empty token / incomplete Registration | `Could not form valid cluster registration object:` |
| 401 | Upstream auth failure without an `HttpError` (default) | `Could not authenticate:` |
| upstream code | Upstream returns an `HttpError` (any 4xx/5xx code) | `Could not authenticate:` |
| 500 | JSON marshal failure | `Unable to read identity:` |

Error status code logic (`getErrorStatusCode`):
- If the error wraps a `client.HttpError`, use its `StatusCode` field directly, regardless of the value.
- Otherwise, default to **401**.

### Status Endpoint

`GET /status` returns `{"status": "available"}` with no authentication required.

## Middleware Stack

Applied in this order via `chi.Use`:
1. `request_id.ConfiguredRequestID("x-rh-insights-request-id")` — generates/propagates request IDs
2. `middleware.RealIP` — trusts X-Forwarded-For / X-Real-IP
3. `middleware.Logger` — request logging
4. `middleware.Recoverer` — panic recovery returns 500
5. `middleware.StripSlashes` — trailing slashes normalized

Do not reorder these. The request-id middleware must be first.

## Caching Behavior

- Cache key: `<cluster_id>:<bearer_token>` (colon-separated)
- TTL: 2 hours (hardcoded in `cache.Set`)
- Cache stores raw JSON bytes, not deserialized structs
- Cache is an in-memory `map[string]item` with a global mutex (no distributed cache)
- On cache hit, the upstream account management service is **not** called
- `cache.Clear()` is used in tests between scenarios; there is no runtime cache invalidation endpoint

## Upstream Client Contract

The `HTTPWrapper.Do` method adds headers for the upstream call to the account management service:

```
Authorization: AccessToken <cluster_id>:<authorization_token>
Accept: application/json
Content-Type: application/json
```

Note: the upstream `Authorization` format is `AccessToken`, not `Bearer`. This is a different scheme from what the proxy itself accepts.

## Prometheus Metrics

All Prometheus metrics use the `uhc_auth_proxy_` prefix:

| Metric | Type | Labels |
|--------|------|--------|
| `uhc_auth_proxy_cache_hit` | Counter | none |
| `uhc_auth_proxy_cache_miss` | Counter | none |
| `uhc_auth_proxy_responses` | Counter | `code` |
| `uhc_auth_proxy_request_time` | Histogram | `url` |
| `uhc_auth_proxy_to_acct_mgmt_request_status` | Counter | `code` |

When adding new metrics, follow this naming convention and always use `promauto.New*`.

## Rules for Adding or Changing API Behavior

1. **No request body parsing.** The proxy authenticates purely from headers.
2. **Error responses are plain text.** Do not return JSON error bodies from the auth handler.
3. **New operators require two code changes and one test change:** add a const and array entry (updating the array size literal) in `server.go`, and add the operator name (without trailing slash) to `validOperatorAgents` in `server_test.go`.
4. **Status codes from upstream are forwarded.** Do not mask or remap upstream 4xx/5xx codes wrapped in `client.HttpError`.
5. **Cache invalidation does not exist at runtime.** The 2-hour TTL is the only mechanism.
6. **The `/` and `/api/uhc-auth-proxy/v1` routes share the same handler variable** and are therefore always in sync. Do not register separate handlers for these routes.
7. **Content-Type is only set on success.** Error paths in the auth handler intentionally omit it.
