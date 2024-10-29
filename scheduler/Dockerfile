# build stage
FROM golang:1.23.2 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY scheduler/go.mod go.mod
COPY scheduler/go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY scheduler/main.go main.go
COPY scheduler/pkg/commons.go pkg/commons.go
COPY scheduler/pkg/postgresql/dbclient.go pkg/postgresql/dbclient.go
COPY scheduler/pkg/postgresql/dbutils.go pkg/postgresql/dbutils.go

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o scheduler main.go

FROM centos:7

COPY --from=builder /workspace/scheduler scheduler

COPY scheduler/start.sh /

ENTRYPOINT ["/start.sh"]

