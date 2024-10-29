package influxdbmetricsscheduler

import (
	schedulerv1 "scheduler-operator/api/v1"
	"scheduler-operator/internal/controller/influxdbmetricsscheduler/common"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *InfluxdbMetricsSchedulerReconciler) recordEvent(influxdbMetricsScheduler *schedulerv1.InfluxdbMetricsScheduler, event common.Event, reason common.Reason, message string) {
	r.Recorder.Event(influxdbMetricsScheduler, string(event), string(reason), message)
}

func (r *InfluxdbMetricsSchedulerReconciler) recordEventFromOperationResult(influxdbMetricsScheduler *schedulerv1.InfluxdbMetricsScheduler, opResult controllerutil.OperationResult, message string) {
	switch s := opResult; s {
	case controllerutil.OperationResultCreated:
		r.recordEvent(influxdbMetricsScheduler, common.Normal, common.Created, message)
	case controllerutil.OperationResultUpdated:
		r.recordEvent(influxdbMetricsScheduler, common.Normal, common.Updated, message)
	default:
		// Nothing
	}
}
