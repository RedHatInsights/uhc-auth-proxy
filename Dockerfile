# Manual Build: 2022-03-10
############################
# STEP 1 build executable binary
############################
FROM registry.access.redhat.com/ubi9/go-toolset:9.7-1771417345 AS builder

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

ENV PATH="/usr/local/go/bin:${PATH}"

RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build -o /go/bin/uhc-auth-proxy
############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.7-1771346502

# Copy our static executable.
COPY --from=builder /go/bin/uhc-auth-proxy /go/bin/uhc-auth-proxy
# Default port
# EXPOSE 8080/tcp
# Run the hello binary.
ENTRYPOINT ["/go/bin/uhc-auth-proxy", "start"]
