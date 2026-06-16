# Testing Guidelines for uhc-auth-proxy

## Framework and Setup

All tests use **Ginkgo v2** with **Gomega** matchers. Do not use standard `testing.T` assertions or `testify`. Run tests with:

```bash
go test -v -race --coverprofile=coverage.txt --covermode=atomic ./...
```

## Suite File Convention

Every package with tests has a `*_suite_test.go` bootstrap file. When adding tests to a new package, create this file first:

```go
package mypkg_test

import (
    "testing"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestMyPkg(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "MyPkg Suite")
}
```

## Package Access Conventions

Two patterns coexist — follow the one already used in the package you are modifying:

- **External test package** (`package foo_test`): Used in `cache` and `requests/cluster` packages. Import the package under test with dot-import.
- **Internal test package** (`package foo`): Used in `server` and `logger` packages. The `server` test is internal to access unexported functions like `getClusterID`, `getToken`, and `makeKey`. The `logger` test is internal to stub the unexported `getCloudwatchCore` function variable.

Do not switch a package from internal to external or vice versa.

## Dot-Imports

Ginkgo and Gomega are always dot-imported. The package under test is dot-imported only in external test packages. No other packages should be dot-imported.

## Describe/It Structure

Tests use a two-level nesting pattern. Prefer not to nest `Describe` blocks more than two levels deep. A common naming convention is `"When <condition>"` for Describe blocks and `"should <expected behavior>"` for It blocks, though existing tests do not apply this uniformly — follow the convention already used in the file you are modifying:

```go
Describe("When passed a valid user-agent header", func() {
    It("should return a cluster id", func() {
        // ...
    })
})
```

## BeforeEach for Test Isolation

Use `BeforeEach` to reset shared state before each `It` block:

1. **Cache state**: Call `cache.Clear()` in `BeforeEach` when testing handlers or any code that touches the cache.
2. **Wrapper and account setup**: Initialize `FakeWrapper`, `ErrorWrapper`, and `ErrorWithBodyWrapper` structs in `BeforeEach`, not at the `Describe` level. Fields on these structs (such as `AccountError` or `StatusCode` on `ErrorWithBodyWrapper`) may be set per `It` block when different tests require different values.

```go
BeforeEach(func() {
    account = &cluster.Account{
        Organization: cluster.Org{EbsAccountID: "123", ExternalID: "123"},
    }
    wrapper = &cluster.FakeWrapper{GetAccountResponse: account}
    cache.Clear()
})
```

## Mocking External HTTP Calls via Wrapper Interface

The codebase does **not** use `httptest.Server` to mock upstream API calls. Instead, it uses the `client.Wrapper` interface with fake implementations in `requests/cluster/types.go`:

| Wrapper | Purpose | Behavior |
|---|---|---|
| `FakeWrapper` | Happy path | Returns a marshaled `Account` based on `GetAccountResponse` |
| `ErrorWrapper` | Error without body | Returns `nil` bytes and a generic error |
| `ErrorWithBodyWrapper` | Error with OCM-style body | Returns marshaled `AccountError` bytes and a `client.HttpError` with configurable `StatusCode` |

When adding new error scenarios, extend or create a new wrapper in `requests/cluster/types.go` — do not create wrapper mocks inside test files.

## Testing HTTP Handlers

Handler tests use `httptest.NewRecorder` (not `httptest.NewServer`). The `server` package has a shared `call` helper function:

```go
func call(wrapper client.Wrapper, userAgent string, auth string) (*httptest.ResponseRecorder, *cluster.Identity)
```

Rules for handler tests:
- Prefer using `call()` rather than creating your own request/recorder setup.
- The returned `Identity` will be zero-value (`&cluster.Identity{}`) on error paths.
- Check `rr.Result().StatusCode` for HTTP status assertions.
- Check `rr.Body` with `io.ReadAll` and `ContainSubstring` for error body content.

## Testing Error Propagation

When testing error paths with `ErrorWithBodyWrapper`, set both `StatusCode` and `AccountError` fields before calling. Verify error chain behavior with `errors.Unwrap`:

```go
unwrap := errors.Unwrap(err)
Expect(unwrap).To(BeAssignableToTypeOf(&AccountError{}))
```

## Table-Driven Tests

For testing multiple valid inputs against the same assertion, use a loop over a slice inside a single `Describe` block:

```go
validOperatorAgents := []string{"insights-operator", "cost-mgmt-operator", ...}
for _, a := range validOperatorAgents {
    It(fmt.Sprintf("should return a valid Identity json for %s", a), func() {
        _, ident := call(wrapper, fmt.Sprintf("%s/abc cluster/123", a), "Bearer mytoken")
        Expect(ident.AccountNumber).To(Equal("123"))
    })
}
```

## Logger Tests and Environment Stubbing

Logger tests use `github.com/prashantv/gostub` to stub environment variables and function variables. Reset stubs with `defer stubs.Reset()` and reset global logger state in `AfterEach`:

```go
AfterEach(func() {
    resetLogger()
})
```

## Assertion Conventions

- Use `Expect(err).To(BeNil())` for nil error checks.
- Use `Expect(err).To(Not(BeNil()))` for non-nil error checks.
- Use `Equal()` for exact value matching.
- Use `ContainSubstring()` for partial string matching on response bodies.
- Use `BeAssignableToTypeOf()` for type assertions on error types.

## What Must Be Tested

When modifying code, ensure tests cover:

1. **Valid input happy path** — correct return values and types.
2. **Invalid input** — proper error returns and HTTP status codes.
3. **Empty/missing input** — edge cases like empty tokens or missing headers.
4. **Error propagation** — errors from wrappers surface correctly, including status codes from upstream services.
5. **Cache interaction** — behavior is correct with both empty and populated caches.

## Adding a New Operator Prefix

When adding a new operator to `operatorPrefixes` in `server/server.go`:
1. Add the prefix constant.
2. Add it to the `operatorPrefixes` array (update the array size literal).
3. Add the operator name (without trailing slash) to the `validOperatorAgents` slice in `server/server_test.go`.
4. The existing table-driven loop will automatically generate a test case.
