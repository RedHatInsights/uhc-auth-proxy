# Contributing to uhc-auth-proxy

## Start Here

Read these docs before writing any code:

- **[AGENTS.md](AGENTS.md)** — repo layout, conventions, architectural notes, and common pitfalls
- **[CLAUDE.md](CLAUDE.md)** — build/test/lint commands and CI expectations
- **[README.md](README.md)** — project purpose, data flow, and configuration

The `docs/` guideline files cover security, performance, error handling, API contracts, testing, and integration patterns. Read the relevant ones before touching the corresponding code.

---

## 1. Local Development Setup

```bash
# Install the binary locally
go install ./...

# Build (verify compilation)
go build ./...

# Start the server on :8080
uhc-auth-proxy start
```

No Makefile — use `go` commands directly. Do not add one.

---

## 2. Tests and Linting

Run these before every push. They match CI exactly.

```bash
# Tests (race detector required — matches Konflux CI)
go test -v -race --coverprofile=coverage.txt --covermode=atomic ./...

# Lint (matches GitHub Actions golangci-lint workflow)
golangci-lint run
```

Tests use Ginkgo v2 / Gomega. See [docs/testing-guidelines.md](docs/testing-guidelines.md) for conventions, including the requirement to call `cache.Clear()` in every `BeforeEach`.

---

## 3. Adding a New Operator Prefix

This is the most common non-trivial contribution. Two files must change together — if they drift, tests will fail.

**`server/server.go`**

1. Add a named constant for the new prefix (follow the existing `const` block):
   ```go
   myNewOperatorPrefix = `my-new-operator/`
   ```
2. Add the constant to the `operatorPrefixes` array and update its size:
   ```go
   operatorPrefixes = [10]string{..., myNewOperatorPrefix}
   ```

**`server/server_test.go`**

3. Add the operator name (without the trailing `/`) to `validOperatorAgents`:
   ```go
   validOperatorAgents := []string{..., "my-new-operator"}
   ```

Run `go test -v -race ./...` and confirm the new operator appears in passing test output.

---

## 4. PR Expectations

**CI checks that must pass:**

| Check | What it does |
|-------|-------------|
| `golangci-lint` (GitHub Actions) | Lint on every PR — fix all errors before pushing |
| Konflux/Tekton unit tests | `go test -v -race ./...` in a container build |
| JSON/YAML validation | Malformed config files fail automatically |

**What reviewers look for:**

- All outbound HTTP goes through `client.Wrapper`, never `http.DefaultClient`
- New tests follow Ginkgo v2 / Gomega patterns and clear the cache in `BeforeEach`
- Prometheus metric names stay prefixed `uhc_auth_proxy_`
- JSON field tags on `Identity` and related structs use `snake_case` — do not rename without coordinating with consumers
- No secrets or credentials in code or config files

---

## 5. Note for AI Agents

Read [AGENTS.md](AGENTS.md) first — it is the authoritative orientation document. [CLAUDE.md](CLAUDE.md) has the exact commands for build, test, and lint. Do not create a Makefile. Run tests with `-race`.
