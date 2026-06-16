# Security Guidelines for uhc-auth-proxy

## Purpose

This proxy authenticates OpenShift cluster operators (insights-operator, cost-mgmt-operator, etc.) by validating their `cluster_id` and `authorization_token` against the UHC accounts management API. These guidelines codify the security conventions in the codebase so that contributors and AI agents produce consistent, secure changes.

## Architecture Security Model

The proxy sits between cluster operators and the Red Hat accounts management API. It receives a Bearer token and cluster ID from operator User-Agent headers, forwards them as an `AccessToken` credential to the upstream API, and returns an identity JSON used by downstream services. The proxy itself holds an offline access token (`OAT`) used to obtain short-lived access tokens from SSO.

**Trust boundary:** The proxy trusts nothing from the inbound HTTP request. Every request must pass User-Agent validation and Bearer token extraction before any upstream call is made.

## 1. User-Agent Allowlist

Only requests from recognized operator prefixes are accepted. The allowlist is defined in `server/server.go` as `operatorPrefixes`.

Rules:
- Do not remove the `strings.HasPrefix` check or the `cluster/` prefix requirement in `getClusterID`.
- When adding a new operator, add its prefix to the `operatorPrefixes` array and add a corresponding test case in the `validOperatorAgents` slice in `server/server_test.go`.
- The User-Agent must match the format `<operator-prefix><version> cluster/<cluster_id>`. Requests that do not match get a 400 response.
- Do not fall back to a default cluster ID if parsing fails.

## 2. Bearer Token Handling

Rules:
- The `getToken` function in `server/server.go` requires the exact prefix `Bearer ` (with trailing space). Do not accept `Bearer:` or other variants.
- The `getToken` error message currently includes the raw authorization header value in the error string (`"not a bearer token: '%s'"`), and this error is logged via `zap.Error`. Avoid extending this pattern — do not log successfully extracted token values. If modifying `getToken`, consider redacting the header value in the error message.
- The `makeKey` function rejects registrations where either `ClusterID` or `AuthorizationToken` is empty, returning a 500. Do not weaken this check.

## 3. Cache Key Construction

The in-memory cache in `cache/cache.go` uses `clusterID:authorizationToken` as the key.

Rules:
- The cache key must always include both the cluster ID and the full authorization token. This ensures that a stolen or rotated token cannot retrieve a cached identity from a previous token.
- Cache entries expire after 2 hours (hardcoded in `cache.Set`). Do not increase this TTL without security review.
- Always call `cache.Clear()` in test `BeforeEach` blocks to prevent cross-test cache pollution.

## 4. Upstream Credential Forwarding

The `HTTPWrapper.AddHeaders` method in `requests/client/wrapper.go` constructs the upstream authorization header:

```go
req.Header.Add("Authorization", fmt.Sprintf("AccessToken %s:%s", cluster_id, authorization_token))
```

Rules:
- Do not change this format without coordinating with the upstream accounts management API team.
- Do not add the raw `Bearer` token from the inbound request to outbound requests — only use the `AccessToken` format.
- The HTTP client has a configurable timeout (`TIMEOUT_SECONDS`). Do not set this to zero or remove the timeout.

## 5. SSO Token Management

The `requests/client/access.go` file manages the proxy's own SSO access token using the `OAT` (offline access token) environment variable.

Rules:
- The offline access token is fetched from the `OAT` env var, sourced from a Kubernetes secret (`uhc-auth-proxy-secret`). Never hardcode this value.
- The `CLIENT_ID` is sourced from a secret and used in the SSO token request. `CLIENT_SECRET` is defined in the Kubernetes template but is not currently referenced in the Go code. Do not add defaults for these in config files.
- Token refresh uses a mutex-protected package-level variable with expiry. Do not remove the `mutex.Lock()`/`defer mutex.Unlock()` pattern — it prevents concurrent token refresh races.
- The `ACCESS_TOKEN_URL` default points to `sso.redhat.com`. Do not change this default to a non-HTTPS URL.

## 6. Error Handling and Information Disclosure

Rules:
- Error responses to clients should not leak internal URLs, stack traces, or upstream response bodies. The current pattern wraps errors with user-facing messages like `"Could not authenticate"`. Note that `getToken` currently includes the raw authorization header in its error message — exercise caution when modifying error messages to avoid expanding information disclosure.
- When the upstream API returns an error body, it is parsed into `AccountError` and logged server-side with `zap.Object` for structured diagnostics. The client receives only the reason, code, and ID fields via `AccountError.Error()` — do not add raw upstream response bodies to client-facing output.
- Use `errors.As` for error type switching (see `getErrorStatusCode`). Do not use type assertions directly.
- Default to HTTP 401 when the upstream error type is unknown. Do not default to 200 or 500 for authentication failures.

## 7. Secret and Configuration Management

All secrets are injected via environment variables sourced from Kubernetes secrets (see `openshift/uhc-auth-proxy-template.yaml`).

Rules:
- Sensitive env vars: `OAT`, `CLIENT_ID`, `CLIENT_SECRET`, `CW_AWS_ACCESS_KEY_ID`, `CW_AWS_SECRET_ACCESS_KEY`. Never log these values.
- Do not add `viper.SetDefault` for any secret value. Defaults are only acceptable for non-sensitive configuration like `SERVER_PORT`, `TIMEOUT_SECONDS`, `LOG_LEVEL`, CloudWatch log group/region/stream, and upstream API URLs.
- The `.gitignore` and `.dockerignore` should not be modified to include secret files.

## 8. HTTP Server Configuration

Rules:
- The server uses `chi` middleware: `request_id.ConfiguredRequestID`, `RealIP`, `Logger`, `Recoverer`, and `StripSlashes`. Do not remove `Recoverer` — it prevents panics from crashing the process and leaking stack traces.
- The `/metrics` endpoint exposes Prometheus metrics. It does not require authentication (it is accessed by internal scrapers), but it should not expose secret values in metric labels.
- The `/status` endpoint is used for liveness/readiness probes. It must remain unauthenticated and lightweight.
- The server binds to a configurable port (default 8080). It does not configure TLS directly — TLS termination is handled by the OpenShift router/ingress.

## 9. Logging Security

Rules:
- All logging uses structured JSON via `zap`. Do not use `fmt.Println` for operational logging (it is acceptable in CLI `cmd/` code).
- Do not add `zap.String("token", ...)` or log any field containing a Bearer token, OAT, or AWS secret key. Note that `getToken` currently includes the raw authorization header in its error string when the format is invalid — avoid extending this pattern to valid tokens.
- The `cluster_id` is safe to log and is already included in error logs. The `authorization_token` is not — maintain this distinction.
- CloudWatch integration uses static AWS credentials. If modifying `logger/cloudwatch.go`, ensure credentials are only read from `logConfig` (which sources from env vars), not from config files on disk.

## 10. Testing Security Patterns

Rules:
- Use `FakeWrapper`, `ErrorWrapper`, and `ErrorWithBodyWrapper` (defined in `requests/cluster/types.go`) for test mocking. Do not make real HTTP calls to upstream APIs in unit tests.
- Test both valid and invalid inputs for all parsing functions (`getClusterID`, `getToken`, `makeKey`). The existing test suite covers empty auth, malformed User-Agent, and missing cluster prefixes.
- When adding a new operator to the allowlist, add a test in `server/server_test.go` that exercises the full handler path (via the `call` helper function), not just the `getClusterID` parser.
- Use `httptest.NewRecorder` for handler tests. Do not start a real HTTP server in unit tests.

## 11. Dockerfile Security

Rules:
- The final image uses `ubi9/ubi-minimal`, not a full OS image. Do not switch to a larger base without justification.
- The binary is built with `CGO_ENABLED=0` for a static binary. Do not enable CGO unless absolutely necessary.
- CVE patches for base image packages are applied via `microdnf update` for specific packages. When adding CVE fixes, target only the affected packages rather than running a blanket `microdnf update`.
- The build stage runs as `root` for dependency fetching and compilation. The final stage does not set a USER directive — consider adding `USER 1001` if the deployment does not already enforce a non-root security context.

## 12. Dependency Management

Rules:
- Dependencies are managed via `go.mod`. Review any new dependency for security implications before adding it.
- Automated dependency update PRs (from Mintmaker/Konflux) should be reviewed for breaking changes in security-sensitive packages (`net/http`, TLS libraries, authentication libraries).
- The `go.sum` file must always be committed alongside `go.mod` changes to ensure integrity verification.
