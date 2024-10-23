# build stage
FROM golang:1.19.4 as builder

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

# Vet
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on go vet ./...
# Test
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on go test ./... -coverprofile cover.out -coverpkg=./...
# Build
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o scheduler -ldflags "-X main.Version=$VERSION" main.go

FROM centos:7

COPY --from=builder /workspace/scheduler scheduler

COPY scheduler/start.sh /

ENTRYPOINT ["/start.sh"]

