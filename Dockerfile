############################
# STEP 1 build executable binary
############################
FROM registry.redhat.io/rhel8/go-toolset:latest AS builder
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
# Use go mod
ENV GO111MODULE=on
# Fetch dependencies.
# Using go get requires root.
USER root
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build -o /go/bin/uhc-auth-proxy
############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/ubi8/ubi:latest
# Copy our static executable.
COPY --from=builder /go/bin/uhc-auth-proxy /go/bin/uhc-auth-proxy
# Default port
# EXPOSE 8080/tcp
# Run the hello binary.
ENTRYPOINT ["/go/bin/uhc-auth-proxy", "start"]
