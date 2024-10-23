package controllers

import (
	"context"
	"github.com/go-logr/logr"
	schedulerv1 "scheduler-operator/api/v1"
)

func (r *MetricSchedulerReconciler) deleteMetricScheduler(ctx context.Context, metricScheduler *schedulerv1.MetricScheduler, log logr.Logger) (err error) {

	return nil
}
