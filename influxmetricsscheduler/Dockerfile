# build stage
FROM golang:1.23.2 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY influxmetricsscheduler/go.mod go.mod
COPY influxmetricsscheduler/go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY influxmetricsscheduler/main.go main.go
COPY influxmetricsscheduler/pkg/commons/commons.go pkg/commons/commons.go
COPY influxmetricsscheduler/pkg/influxdb/influxclient.go pkg/influxdb/influxclient.go

# Vet
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on go vet ./...
# Test
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on go test ./... -coverprofile cover.out -coverpkg=./...
# Build
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o scheduler -ldflags "-X main.Version=$VERSION" main.go

FROM alpine:3.20.3

COPY --from=builder /workspace/scheduler scheduler

COPY scheduler/start.sh /

ENTRYPOINT ["/start.sh"]

