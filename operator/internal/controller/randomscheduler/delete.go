package randomscheduler

import (
	"context"
	"github.com/go-logr/logr"
	schedulerv1 "scheduler-operator/api/v1"
)

func (r *RandomSchedulerReconciler) deleteRandomScheduler(ctx context.Context, randomScheduler *schedulerv1.RandomScheduler, log logr.Logger) (err error) {

	return nil
}
