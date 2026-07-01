# AGENTS.md

This file provides orientation for AI agents and contributors working on **uhc-auth-proxy**. For project purpose and usage, see [README.md](README.md). For domain-specific rules, see the guideline files below.

## Guideline Documents

| File | Scope |
|------|-------|
| [docs/security-guidelines.md](docs/security-guidelines.md) | Token handling, user-agent validation, credential flow, and secrets management |
| [docs/performance-guidelines.md](docs/performance-guidelines.md) | In-memory cache behavior, shared HTTP client, mutex patterns, and Prometheus metrics |
| [docs/error-handling-guidelines.md](docs/error-handling-guidelines.md) | Custom error types (`HttpError`, `AccountError`), wrapping conventions, and status code propagation |
| [docs/api-contracts-guidelines.md](docs/api-contracts-guidelines.md) | Endpoint contracts, accepted user-agents, and identity payload shape |
| [docs/testing-guidelines.md](docs/testing-guidelines.md) | Ginkgo v2 / Gomega conventions, test wrappers, and cache clearing between tests |
| [docs/integration-guidelines.md](docs/integration-guidelines.md) | External HTTP call patterns, UHC account management API interaction, and the `client.Wrapper` interface |

## Repository Layout

```
main.go                     # Entrypoint -- delegates to cmd.Execute()
cmd/
  root.go                   # Cobra root command, Viper config init
  start.go                  # `start` subcommand -- launches the HTTP server
  run.go                    # `run` subcommand -- CLI one-shot identity fetch
server/
  server.go                 # chi router, middleware stack, RootHandler, Prometheus counters
cache/
  cache.go                  # In-memory TTL cache (2-hour expiry, mutex-guarded)
requests/
  client/
    wrapper.go              # HTTPWrapper / Wrapper interface -- all outbound HTTP
    access.go               # SSO token refresh with mutex-guarded caching
    config.go               # Viper defaults for ACCESS_TOKEN_URL, TIMEOUT_SECONDS
    types.go                # HttpError type
  cluster/
    cluster.go              # GetIdentity / GetCurrentAccount facade
    types.go                # Registration, Account, Identity, test fakes
    config.go               # Viper defaults for UHC API URLs
logger/
  logger.go                 # zap JSON logger with optional CloudWatch tee
```

## Cross-Cutting Conventions

### Configuration

All runtime configuration uses **Viper** with `AutomaticEnv()`. Defaults are set in package-level `init()` functions. There is no `.env` file; environment variables are the sole config source in deployed environments.

### Naming

- Package names are short, lowercase, singular (`cache`, `server`, `logger`).
- JSON field names use `snake_case` to match the UHC API.
- Prometheus metric names are prefixed `uhc_auth_proxy_`.

### Interface-Based HTTP Abstraction

The `client.Wrapper` interface is the seam for all outbound HTTP. Production code uses `HTTPWrapper`; tests use `FakeWrapper`, `ErrorWrapper`, or `ErrorWithBodyWrapper` (all defined in `requests/cluster/types.go`). Never call `http.DefaultClient` directly.

### Logging

Structured JSON logging via `zap`. Each package creates a named child logger (`log.Named("server")`). Always use `zap.Field` helpers (`zap.Error`, `zap.String`) rather than `fmt.Sprintf` in log calls.

### Build and CI

- **Dockerfile**: Multi-stage build using Hummingbird FIPS images (`hi/go:1.26.4-fips-builder` and `hi/core-runtime:2.42-openssl-fips`). FIPS mode is enabled via `GODEBUG=fips140=on`.
- **Konflux/Tekton**: Pipelines in `.tekton/`. Unit tests run via `konflux_unit_test.sh` with `-race` and coverage.
- **GitHub Actions**: `golangci-lint` on PRs. No Makefile — build with `go build`, test with `go test -v ./...`.

## Architectural Notes

- The service is stateless except for the in-memory cache. Restarting a pod clears all cached identities — this is intentional.
- Data flow: `RootHandler` validates user-agent and bearer token → checks cache → calls UHC API on miss → caches and returns identity JSON.
- Operator prefixes in `server.go` form an allowlist. Adding a new operator requires updating `operatorPrefixes` and the test's `validOperatorAgents` slice.

## Common Pitfalls

1. **Forgetting to clear cache in tests.** Every `BeforeEach` in server tests must call `cache.Clear()`.
2. **Adding operators without updating tests.** The `operatorPrefixes` array and `validOperatorAgents` slice must stay in sync.
3. **Nil-pointer in `wrapper.Do`.** `resp.StatusCode` is accessed before the `err != nil` check at line 61 — a nil `resp` will panic.
4. **Viper `init()` ordering.** Multiple packages set defaults in `init()`. Always use environment variables for non-default values in tests.
5. **JSON field mismatches.** The `Identity` struct uses `snake_case` JSON tags that downstream consumers depend on. Do not rename tags without coordinating with consumers.
