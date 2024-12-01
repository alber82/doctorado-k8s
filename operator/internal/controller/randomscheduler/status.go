package randomscheduler

import (
	schedulerv1 "scheduler-operator/api/v1"
	"scheduler-operator/internal/controller/randomscheduler/common"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *RandomSchedulerReconciler) recordEvent(randomScheduler *schedulerv1.RandomScheduler, event common.Event, reason common.Reason, message string) {
	r.Recorder.Event(randomScheduler, string(event), string(reason), message)
}

func (r *RandomSchedulerReconciler) recordEventFromOperationResult(randomScheduler *schedulerv1.RandomScheduler, opResult controllerutil.OperationResult, message string) {
	switch s := opResult; s {
	case controllerutil.OperationResultCreated:
		r.recordEvent(randomScheduler, common.Normal, common.Created, message)
	case controllerutil.OperationResultUpdated:
		r.recordEvent(randomScheduler, common.Normal, common.Updated, message)
	default:
		// Nothing
	}
}
