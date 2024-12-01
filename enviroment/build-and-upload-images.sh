#!/bin/bash
set -e

export DOCKER_BUILDKIT=1


docker build --platform linux/arm64 --build-arg VERSION=0.0.0 -t 192.168.1.229:5000/doctorado/scheduler-operator:0.0.0 -f operator/Dockerfile .
docker push 192.168.1.229:5000/doctorado/scheduler-operator:0.0.0

docker build --platform linux/arm64 --build-arg VERSION=0.0.0 -t 192.168.1.229:5000/doctorado/ts-scheduler:0.0.0 -f scheduler/Dockerfile .
docker push 192.168.1.229:5000/doctorado/ts-scheduler:0.0.0

docker build --platform linux/arm64  --build-arg VERSION=0.0.0 -t 192.168.1.229:5000/albertogomez/influxdb-scheduler:0.0.0 -f influxmetricsscheduler/Dockerfile .
docker push 192.168.1.229:5000/doctorado/influxdb-scheduler:0.0.0
