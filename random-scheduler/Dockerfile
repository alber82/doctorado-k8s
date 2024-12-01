# build stage
FROM golang:1.23.2 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY random-scheduler/go.mod go.mod
COPY random-scheduler/go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY random-scheduler/scheduler/main.go main.go

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o random-scheduler main.go

FROM alpine:3.20.3

COPY --from=builder /workspace/random-scheduler random-scheduler
ENTRYPOINT ["/random-scheduler"]

