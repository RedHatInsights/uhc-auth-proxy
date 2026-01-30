# Manual Build: 2022-03-10
############################
# STEP 1 build executable binary
############################
FROM registry.access.redhat.com/ubi9/go-toolset:9.7-1768393489 AS builder

LABEL name="uhc-auth-proxy" \
      summary="UHC Auth Proxy - OpenShift Cluster Authentication Service" \
      description="Authentication proxy service for OpenShift 4 clusters. Validates cluster_id and authorization_token against UHC services to enable insights-operator and other operators to send data without storing SSO credentials in clusters." \
      io.k8s.description="Authentication proxy service for OpenShift 4 clusters. Validates cluster_id and authorization_token against UHC services to enable insights-operator and other operators to send data without storing SSO credentials in clusters." \
      io.k8s.display-name="UHC Auth Proxy" \
      io.openshift.tags="insights,uhc,auth,proxy,authentication,openshift,cluster" \
      com.redhat.component="uhc-auth-proxy" \
      version="1.0" \
      release="1" \
      vendor="Red Hat, Inc." \
      url="https://github.com/redhatinsights/uhc-auth-proxy" \
      distribution-scope="private" \
      maintainer="platform-accessmanagement@redhat.com"

WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
# Fetch dependencies.
# Using go get requires root.
USER root

# Install Go 1.25.5 to address CVE-2025-61729 and CVE-2025-61727 (Until new ubi9 minimal image supports this go version)
RUN curl -LO https://go.dev/dl/go1.25.5.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz && \
    rm go1.25.5.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build -o /go/bin/uhc-auth-proxy
############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.7-1769056855

# Update vulnerable packages to address security vulnerabilities:
# - curl-minimal/libcurl-minimal: CVE-2025-9086 (Medium)
# - openssl-libs: CVE-2025-15467 (High), CVE-2025-69419, CVE-2025-11187 (Medium)
RUN microdnf update -y curl-minimal libcurl-minimal openssl-libs && \
    microdnf clean all

# Copy our static executable.
COPY --from=builder /go/bin/uhc-auth-proxy /go/bin/uhc-auth-proxy
# Default port
# EXPOSE 8080/tcp
# Run the hello binary.
ENTRYPOINT ["/go/bin/uhc-auth-proxy", "start"]
