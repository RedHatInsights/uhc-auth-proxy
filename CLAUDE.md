@AGENTS.md

## Commands

```bash
# Build
go build ./...

# Test (matches CI)
go test -v -race --coverprofile=coverage.txt --covermode=atomic ./...

# Lint (matches GitHub Actions golangci-lint workflow)
golangci-lint run
```

## CI Checks on PRs

- **golangci-lint** runs on every PR via GitHub Actions. Fix all lint errors before pushing.
- **Konflux/Tekton** builds a container image and runs `go test -v -race ./...` on every PR.
- JSON/YAML files are validated automatically; malformed config files will fail CI.

## Notes for Claude Code

- There is no Makefile. Use `go build` and `go test` directly.
- The `-race` flag is required in CI — run tests with it locally to catch data races before pushing.
- Do not add a Makefile unless explicitly requested.
