package controllers

import (
	schedulerv1 "scheduler-operator/api/v1"
	"scheduler-operator/internal/common"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *MetricSchedulerReconciler) recordEvent(metricScheduler *schedulerv1.MetricScheduler, event common.Event, reason common.Reason, message string) {
	r.Recorder.Event(metricScheduler, string(event), string(reason), message)
}

func (r *MetricSchedulerReconciler) recordEventFromOperationResult(metricScheduler *schedulerv1.MetricScheduler, opResult controllerutil.OperationResult, message string) {
	switch s := opResult; s {
	case controllerutil.OperationResultCreated:
		r.recordEvent(metricScheduler, common.Normal, common.Created, message)
	case controllerutil.OperationResultUpdated:
		r.recordEvent(metricScheduler, common.Normal, common.Updated, message)
	default:
		// Nothing
	}
}
