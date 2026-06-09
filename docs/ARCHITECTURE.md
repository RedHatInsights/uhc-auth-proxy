# Architecture

This document captures the reasoning behind uhc-auth-proxy's design decisions. For implementation details, see [AGENTS.md](../AGENTS.md). For domain-specific rules, see the guideline files in `docs/`.

## Why This Service Exists

OpenShift 4 clusters ship with an `insights-operator` that sends telemetry and diagnostic data to Red Hat. Authenticating these uploads requires associating each cluster with a Red Hat customer account. The naive solution would be to store SSO credentials on every cluster, but that creates a massive secret-management burden for customers and a large credential-sprawl attack surface.

uhc-auth-proxy solves this by accepting the `cluster_id` and `authorization_token` that are already provisioned with every OpenShift 4 cluster, and translating them into an identity document that downstream platform services understand. This lets clusters authenticate with zero additional configuration by the customer.

## The Two-Auth-Scheme Design

Inbound requests carry a `Bearer <token>` header. Outbound requests to the UHC accounts management API use a different format: `AccessToken <cluster_id>:<token>`. The proxy does not pass the Bearer token through unchanged.

This is intentional. The upstream UHC API requires the `AccessToken` scheme with the cluster ID and token combined. The Bearer scheme on the inbound side is a convention that lets the API gateway route the request to this proxy based on the User-Agent header without needing to understand the cluster-specific auth format. The proxy is the translation layer between these two auth worlds.

## Why an Operator Allowlist

The `operatorPrefixes` array in `server.go` restricts which User-Agent strings are accepted. Only known Red Hat operators (insights-operator, cost-mgmt-operator, etc.) can authenticate through this proxy.

An open model was rejected because this proxy returns identity documents that grant access to customer account data in downstream services. Allowing arbitrary callers would turn any cluster credential into a general-purpose customer-account token. The allowlist ensures that only operators with a legitimate need to report data can use this path. Adding a new operator is intentionally a code change that requires review.

## Why a 2-Hour Cache TTL

The cache key is `clusterID:authorizationToken`, and entries expire after 2 hours. This TTL balances three concerns:

1. **UHC API load**: Without caching, every request from every cluster hits the accounts management API. Operators report frequently (insights-operator reports every 2 hours by default), so caching for that interval eliminates nearly all redundant calls.
2. **Token rotation correctness**: Because the authorization token is part of the cache key, a rotated token immediately bypasses the cache. The 2-hour window only applies to unchanged credentials.
3. **Account change propagation**: If a customer's org ID or account number changes (rare), the stale cached identity persists for at most 2 hours. This was deemed acceptable since account metadata changes are infrequent and not time-critical for telemetry.

## Why In-Memory Cache Instead of an External Store

The service is intentionally stateless from a deployment perspective. A pod restart clears all cached identities, which is safe because the next request simply re-authenticates against the upstream API. An external cache would add operational complexity for a workload where cache misses are cheap and infrequent. Each pod in a scaled deployment maintains its own cache, but since operator requests are typically routed consistently, duplication is minimal.

## The SSO Token (GetToken) Path

`requests/client/access.go` contains a mutex-guarded SSO token cache that exchanges an offline access token (OAT) for a short-lived access token. This code exists but is not currently called in the production request path — the `AccessToken` scheme used by `wrapper.go` sends the cluster credentials directly without needing an SSO token. The SSO token path was part of an earlier auth flow and remains available as an alternative mechanism. The mutex serialization in `GetToken` would be a bottleneck under high concurrency if reactivated.

## Why the Wrapper Interface

All outbound HTTP goes through `client.Wrapper`, an interface with a single `Do` method. This exists purely as a test seam. The production implementation (`HTTPWrapper`) adds auth headers and makes real HTTP calls; test implementations (`FakeWrapper`, `ErrorWrapper`, `ErrorWithBodyWrapper`) return canned responses. This pattern was chosen over HTTP-level mocking because it keeps tests focused on the proxy's logic rather than HTTP plumbing.

## Notable Tradeoffs and Known Issues

- **No graceful shutdown**: The server runs `ListenAndServe` until the process is killed. In-flight requests are dropped. This is acceptable because the proxy is stateless and operator clients retry.
- **No cache eviction**: Expired entries are only detected on read, never proactively evicted. Memory grows monotonically during a pod's lifetime, but is bounded by the cluster population.
- **Nil-pointer bug in `wrapper.Do`**: `resp.StatusCode` is accessed before checking `err != nil`. If the HTTP call returns `(nil, error)`, the service panics. The `Recoverer` middleware prevents a crash but the request fails with a 500.
- **Client timeout initialization order**: The `http.Client` timeout reads `TIMEOUT_SECONDS` via Viper at package init time, before Viper defaults are registered. If the env var is not set, the client has no timeout.
- **Error message leaks auth header**: `getToken` in `server.go` includes the raw Authorization header in its error message when the format is invalid. This value is logged and returned to the caller.
