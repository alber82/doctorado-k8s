package influxdbmetricsscheduler

import (
	"context"
	"github.com/go-logr/logr"
	schedulerv1 "scheduler-operator/api/v1"
)

func (r *InfluxdbMetricsSchedulerReconciler) deleteTsMetricsScheduler(ctx context.Context, influxdbMetricsScheduler *schedulerv1.InfluxdbMetricsScheduler, log logr.Logger) (err error) {

	return nil
}
