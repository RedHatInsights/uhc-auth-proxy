# Error Handling Guidelines for uhc-auth-proxy

## Custom Error Types

### `client.HttpError`

Defined in `requests/client/types.go`. Carries an HTTP `StatusCode` and `Message`. Created by `HTTPWrapper.Do` when the upstream response has status >= 400. The `StatusCode` field is the **only** mechanism for propagating upstream HTTP status codes back to the caller.

### `cluster.AccountError`

Defined in `requests/cluster/types.go`. Represents an error body returned by the OCM accounts API. Implements `zapcore.ObjectMarshaler` for structured logging via `zap.Object()`. Has an `Inner` field and implements `Unwrap() error` so it participates in the `errors.As`/`errors.Is` chain.

## Error Wrapping Rules

### Rule 1: Prefer `%w` when crossing package boundaries

`cluster.GetIdentity` wraps errors from `GetCurrentAccount` using `fmt.Errorf("...: %w", err)`. This preserves the original error for type assertions upstream. Prefer `%w` (not `%v` or `%s`) when the caller needs to inspect the underlying error type.

### Rule 2: AccountError wraps HttpError, not the other way around

In `GetCurrentAccount`, when the wrapper returns both a body and an error, the body is unmarshalled into `AccountError` and the original error (typically `*client.HttpError`) is stored in `AccountError.Inner`. The chain is:

```
fmt.Errorf wrapping -> *AccountError -> (Inner) *client.HttpError
```

This means `errors.As(err, &accErr)` and `errors.As(err, &httpError)` can **both** succeed on the same error chain. The server relies on this.

### Rule 3: Return body bytes alongside errors from Wrapper.Do

`client.HTTPWrapper.Do` returns `(body, error)` where **both can be non-nil** when status >= 400. The body contains the upstream error response. Callers must check `b != nil` before attempting to unmarshal even when `err != nil`.

## HTTP Status Code Propagation

### Rule 4: Use `getErrorStatusCode` for mapping errors to response codes

`server.getErrorStatusCode` uses `errors.As` to extract `*client.HttpError` from the error chain and returns its `StatusCode`. The **default** status code when no `HttpError` is found is **401** (not 500).

### Rule 5: Input validation errors return 400, not 401

Errors from `getClusterID` and `getToken` (malformed request headers) are handled **before** calling `GetIdentity` and return HTTP 400 directly.

### Rule 6: Internal failures return 500

`makeKey` failures (incomplete Registration) and `json.Marshal` failures return HTTP 500 directly, bypassing `getErrorStatusCode`.

| Condition | Status Code |
|---|---|
| Invalid User-Agent | 400 |
| Invalid Authorization header | 400 |
| Incomplete Registration (empty token) | 500 |
| Upstream auth failure with `HttpError` | Forward upstream code |
| Upstream auth failure without `HttpError` | 401 |
| JSON marshal failure | 500 |

## Structured Logging with Zap

### Rule 7: Prefer `zap.Error(err)` over `fmt.Sprintf` for error logging

In most error paths in the server, `zap.Error(err)` is used as a structured field. Prefer this over `fmt.Sprintf` to inline errors into the message string. Note that `StatusHandler` currently uses `log.Error(fmt.Sprintf(...))` as an exception; new code in the request handler should use `zap.Error(err)` as the structured field instead.

### Rule 8: Use `getErrorSpecificFields` to attach domain context

When handling authentication failures, `getErrorSpecificFields` checks for `*cluster.AccountError` via `errors.As` and appends it as a `zap.Object` field. Extend this function when adding new error types that carry diagnostic data.

### Rule 9: Include `cluster_id` in authentication error logs

The server adds `zap.String("cluster_id", reg.ClusterID)` to the log fields for auth failures. Include it when logging errors related to a specific cluster.

### Rule 10: Attach `request_id` via logger context, not per-call

The server creates a per-request logger with `log.With(zap.String("request_id_header", reqID), ...)` and uses it for all subsequent log calls. Do not pass request ID as a separate field on each call.

## Error Construction Patterns

### Rule 11: Use `errors.New` for static sentinel errors

Use `errors.New` for fixed error messages with no dynamic content (see `makeKey`). Use `fmt.Errorf` only when interpolating values into the message.

### Rule 12: HttpError message should include URL, status code, and status text

Follow the pattern in `HTTPWrapper.Do`:

```go
fmt.Sprintf("request to %s failed: %d %s", req.URL.String(), resp.StatusCode, resp.Status)
```

## Testing Error Handling

### Rule 13: Use test wrapper types for error simulation

| Type | Behavior |
|---|---|
| `FakeWrapper` | Returns a configured successful response |
| `ErrorWrapper` | Always returns an error with nil body |
| `ErrorWithBodyWrapper` | Returns both a JSON body and a `*client.HttpError` |

### Rule 14: Verify error type with `errors.Unwrap` in tests

When testing error wrapping, unwrap to the expected depth and assert with `BeAssignableToTypeOf`. The cluster test package uses a dot-import of `requests/cluster`, so `AccountError` can be referenced unqualified:

```go
unwrap := errors.Unwrap(err)
Expect(unwrap).To(BeAssignableToTypeOf(&AccountError{}))
```

### Rule 15: Clear the cache in BeforeEach for error tests

`cache.Clear()` must be called in `BeforeEach` blocks for server handler tests. Cached successful responses mask error paths and cause false passes.

### Rule 16: Assert HTTP status codes on the ResponseRecorder

Server tests use `rr.Result().StatusCode` to verify the HTTP status code returned for each error scenario. Every new error path in the handler must have a corresponding status code assertion.

## Response Body on Errors

Write a human-readable message to the response body via `fmt.Fprintf`. The `AccountError.Error()` method formats as `[Reason: ..., Code: ..., ID: ...]` which propagates upstream diagnostic info to the caller. Do not return empty bodies on error.

## Wrapper Interface Contract

The `client.Wrapper` interface returns `([]byte, error)`:
- `(body, nil)` — success
- `(body, err)` — upstream error with response body (may contain structured error JSON)
- `(nil, err)` — transport or other failure with no body

When implementing a new `Wrapper`, return `*HttpError` for HTTP failures so the status code propagates. Return the body alongside the error when available.
