package controllers

import (
	"context"
	"github.com/go-logr/logr"
	schedulerv1 "scheduler-operator/api/v1"
)

func (r *TsMetricsSchedulerReconciler) deleteTsMetricsScheduler(ctx context.Context, tsMetricScheduler *schedulerv1.TsMetricsScheduler, log logr.Logger) (err error) {

	return nil
}
