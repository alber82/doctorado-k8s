#!/bin/bash

#if [[ "${TIMESCALEDB_HOST}" == "" ]]; then
#  echo 'Missing TIMESCALEDB_HOST'
#  exit 1
#fi

trap shutdown INT

function shutdown() {
  pkill -SIGINT scheduler
}

echo /scheduler \
       --metric-name=${METRIC_NAME:-metricname} \
       --metric-start-date="${METRIC_START_DATE:-"now()- INTERVAL '130 DAYS'"}" \
       --metric-end-date="${METRIC_END_DATE:-'now()'}" \
       --metric-operation=${METRIC_OPERATION:-'max'} \
       --metric-priority-order=${METRIC_PRIORITY_ORDER:-'desc'} \
       --metric-is-second-level=${METRIC_IS_SECOND_LEVEL:-'false'} \
       --metric-filter-clause="${METRIC_FILTER_CLAUSE:-''}" \
       --metric-second-level-group="${METRIC_SECOND_LEVEL_GROUP:-''}" \
       --metric-second-level-operation="${METRIC_SECOND_LEVEL_OPERATION:-''}" \
       --influxdb-host=${INFLUXDB_HOST:-'influxdb-influxdb2.monitoring'} \
       --influxdb-port=${INFLUXDB_PORT:-8086} \
       --influxdb-token=${INFLUXDB_TOKEN:-'klsjdaioqwehrqoikdnmxcq'} \
       --influxdb-organization=${INFLUXDB_ORGANIZATION:-'uclm'} \
       --influxdb-bucket=${INFLUXDB_BUCKET:-'doctorado'} \
       --scheduler-name=${SCHEDULER_NAME:-'random'} \
       --log-level=${LOG_LEVEL:-'info'} \
       --filtered-nodes=${FILTERED_NODES:-''} \
       --timeout=${TIMEOUT:-20}

/scheduler \
  --metric-name=${METRIC_NAME:-metricname} \
  --metric-start-date="${METRIC_START_DATE:-"now()- INTERVAL '130 DAYS'"}" \
  --metric-end-date="${METRIC_END_DATE:-'now()'}" \
  --metric-operation=${METRIC_OPERATION:-'max'} \
  --metric-priority-order=${METRIC_PRIORITY_ORDER:-'desc'} \
  --metric-is-second-level=${METRIC_IS_SECOND_LEVEL:-'false'} \
  --metric-filter-clause=${METRIC_FILTER_CLAUSE:-''} \
  --metric-second-level-group=${METRIC_SECOND_LEVEL_GROUP:-''} \
  --metric-second-level-operation="${METRIC_SECOND_LEVEL_OPERATION:-''}" \
  --influxdb-host=${INFLUXDB_HOST:-'influxdb-influxdb2.monitoring'} \
  --influxdb-port=${INFLUXDB_PORT:-8086} \
  --influxdb-token=${INFLUXDB_TOKEN:-'klsjdaioqwehrqoikdnmxcq'} \
  --influxdb-organization=${INFLUXDB_ORGANIZATION:-'uclm'} \
  --influxdb-bucket=${INFLUXDB_BUCKET:-'doctorado'} \
  --scheduler-name=${SCHEDULER_NAME:-'metricscheduler'} \
  --log-level=${LOG_LEVEL:-'info'} \
  --filtered-nodes=${FILTERED_NODES:-''} \
  --timeout=${TIMEOUT:-20}
