# Manual Build: 2022-03-10
############################
# STEP 1 build executable binary
############################
FROM registry.access.redhat.com/hi/go:1.27.0-fips-builder@sha256:f61df82b9277aa825678ba9b19960bee25c84d6e07f5db269790fbd9462751f1 AS builder

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
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build -o /go/bin/uhc-auth-proxy
############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/hi/core-runtime:2.43-openssl-fips@sha256:3d4c13013f00581ae341ec62427e9b969d288a2f66b71b6de9028b3ef2c547a0

# Copy our static executable.
COPY --from=builder /go/bin/uhc-auth-proxy /go/bin/uhc-auth-proxy
# Default port
# EXPOSE 8080/tcp
# Run the hello binary.
ENV GODEBUG=fips140=on
USER 1001
ENTRYPOINT ["/go/bin/uhc-auth-proxy", "start"]
