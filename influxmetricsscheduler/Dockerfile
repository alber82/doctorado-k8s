# build stage
FROM golang:1.23.2 AS builder
ARG TARGETOS
ARG TARGETARCH

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

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o scheduler main.go

FROM alpine:3.20.3

WORKDIR /
COPY --from=builder /workspace/scheduler .

#COPY influxmetricsscheduler/start.sh /start.sh

ENTRYPOINT ["/scheduler"]
