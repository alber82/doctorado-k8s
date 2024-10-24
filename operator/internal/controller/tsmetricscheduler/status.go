package controllers

import (
	schedulerv1 "scheduler-operator/api/v1"
	"scheduler-operator/internal/controller/tsmetricscheduler/common"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *TsMetricsSchedulerReconciler) recordEvent(metricScheduler *schedulerv1.TsMetricsScheduler, event common.Event, reason common.Reason, message string) {
	r.Recorder.Event(metricScheduler, string(event), string(reason), message)
}

func (r *TsMetricsSchedulerReconciler) recordEventFromOperationResult(metricScheduler *schedulerv1.TsMetricsScheduler, opResult controllerutil.OperationResult, message string) {
	switch s := opResult; s {
	case controllerutil.OperationResultCreated:
		r.recordEvent(metricScheduler, common.Normal, common.Created, message)
	case controllerutil.OperationResultUpdated:
		r.recordEvent(metricScheduler, common.Normal, common.Updated, message)
	default:
		// Nothing
	}
}
