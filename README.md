# UHC Auth Proxy

uhc-auth-proxy is a service for authenticating Openshift 4 clusters. It does
so by handling the various API calls required to authenticate a `cluster_id`
and `authorization_token` with UHC services.

The `cluster_id` and `authorization_token` are two pieces of data that are
stored with each deployment of an Openshift 4 cluster. They are issued when
the cluster is provisioned.

This service provides an alternative way to authenticate and authorize
requests without requiring a customer to store their SSO credentials
somewhere within their cluster. This, in turn, enables the insights-operator
to send data with no additional configuration required.

![Image](../master/uhc-auth-proxy.png?raw=true)

## How it works

1. A request is made containing the `cluster_id` and `authorization_token` from a
   cluster
2. The auth proxy forms an authorization header from the provided `cluster_id`
   and `authorization_token` for a request to the OpenShift API
   account_management endpoint
3. Using the data returned from the OpenShift API, an identification document
   is built and returned.

An identification document should look something like this:

    {
         "account_number": "123456",
         "org_id": "1234567",
         "type": "System",
         "internal": {
             "org_id": "1234567"
         },
         "system":{"cluster_id":"1234567"}
    }

## Integration with current API broker

In practice, a request is made of the API broker from the insights-operator
with the following headers:

    User-agent: insights-operator/<git hash> cluster/<cluster id>
    Authentication: Bearer <authorization_token>

Because the header indicates the call came from the insights-operator the
uhc-auth-proxy is delegated to for authentication.

## Configuration

The following API will need to be accessed:

### GET https://api.openshift.com/api/accounts_mgmt/v1/current_account

The API returns something like:

    {
        "id": "string",
        "kind": "Account",
        "href": "/api/accounts_mgmt/v1/accounts/:account_id",
        "first_name": "string",
        "last_name": "string",
        "username": "string",
        "email": "string",
        "banned": false,
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "organization": {
            "id": "string",
            "kind": "Organization",
            "href": "/api/accounts_mgmt/v1/organizations/:org_id",
            "name": "Organization Name"
        }
    }

## CLI

You can try all the calls via the cli. To get started run `go install` from
the root of the project.

This will build and install the `uhc-auth-proxy` command. You can request an identity document like this:

    $ uhc-auth-proxy run --cluster-id $CLUSTER_ID --authorization-token $AUTHORIZATION_TOKEN

## Local Development

Any changes to code will require running `go install` and `go build` to rebuild.

This will start the service on port `8080`:

    $ uhc-auth-proxy start


Example of a curl call to local deployment:

    $ curl -X GET http://localhost:8080/api/uhc-auth-proxy/v1 -H 'user-agent: insights-operator/abcdef cluster/ghijkl' -H'Authorization: Bearer <Access token>'

## Open Endpoints

There are three open endpoints that could be reached.

`/api/uhc-auth-proxy/v1`

`/metrics`

`/status`

## Tests

To run tests locally:
```
$ go test -v ./...
```

## Tech Stack

| Concern | Library |
|---------|---------|
| HTTP router | [go-chi/chi v5](https://github.com/go-chi/chi) |
| CLI framework | [spf13/cobra](https://github.com/spf13/cobra) |
| Configuration | [spf13/viper](https://github.com/spf13/viper) |
| Structured logging | [go.uber.org/zap](https://pkg.go.dev/go.uber.org/zap) |
| Metrics | [prometheus/client_golang](https://github.com/prometheus/client_golang) |
| Test framework | [onsi/ginkgo v2](https://onsi.github.io/ginkgo/) + [onsi/gomega](https://onsi.github.io/gomega/) |
| RH platform middleware | [redhatinsights/platform-go-middlewares v2](https://github.com/RedHatInsights/platform-go-middlewares) |
| CloudWatch logging | [RedHatInsights/cloudwatch-v2](https://github.com/RedHatInsights/cloudwatch-v2) |

## Project Structure

```
main.go                     # Entrypoint — delegates to cmd.Execute()
cmd/
  root.go                   # Cobra root command, Viper config init
  start.go                  # `start` subcommand — launches the HTTP server
  run.go                    # `run` subcommand — CLI one-shot identity fetch
server/
  server.go                 # chi router, middleware stack, RootHandler, Prometheus counters
cache/
  cache.go                  # In-memory TTL cache (2-hour expiry, mutex-guarded)
requests/
  client/
    wrapper.go              # HTTPWrapper / Wrapper interface — all outbound HTTP
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

## Documentation

| Document | Description |
|----------|-------------|
| [AGENTS.md](AGENTS.md) | Orientation for AI agents and contributors: repo layout, cross-cutting conventions, architectural notes, and common pitfalls |
| [docs/security-guidelines.md](docs/security-guidelines.md) | Token handling, user-agent validation, credential flow, and secrets management |
| [docs/performance-guidelines.md](docs/performance-guidelines.md) | In-memory cache behavior, shared HTTP client, mutex patterns, and Prometheus metrics |
| [docs/error-handling-guidelines.md](docs/error-handling-guidelines.md) | Custom error types, wrapping conventions, and HTTP status code propagation |
| [docs/api-contracts-guidelines.md](docs/api-contracts-guidelines.md) | Endpoint contracts, accepted user-agents, and identity payload shape |
| [docs/testing-guidelines.md](docs/testing-guidelines.md) | Ginkgo v2 / Gomega conventions, test wrappers, and cache clearing between tests |
| [docs/integration-guidelines.md](docs/integration-guidelines.md) | External HTTP call patterns, UHC account management API interaction, and the `client.Wrapper` interface |
