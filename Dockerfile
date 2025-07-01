# Manual Build: 2022-03-10
############################
# STEP 1 build executable binary
############################
FROM registry.access.redhat.com/ubi9/go-toolset:9.6-1750969886 AS builder
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
# Fetch dependencies.
# Using go get requires root.
USER root
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 go build -o /go/bin/uhc-auth-proxy
############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.6-1751286687
# Copy our static executable.
COPY --from=builder /go/bin/uhc-auth-proxy /go/bin/uhc-auth-proxy
# Default port
# EXPOSE 8080/tcp
# Run the hello binary.
ENTRYPOINT ["/go/bin/uhc-auth-proxy", "start"]
