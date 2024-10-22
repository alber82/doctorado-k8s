package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	MetricSchedulerNameLabel string = "metricscheduler.uclm.es/metricscheduler-name"
)

// Update ...
func Update(owner, object metav1.Object, scheme *runtime.Scheme, labels map[string]string, mutateFns ...ctrlUtil.MutateFn) ctrlUtil.MutateFn {
	return func() error {

		if err := ctrl.SetControllerReference(owner, object, scheme); err != nil {
			return err
		}

		for _, f := range mutateFns {
			if err := f(); err != nil {
				return err
			}
		}
		return nil
	}
}

func CreateIntPtr(x int) *int {
	return &x
}

func CreateInt32Ptr(x int32) *int32 {
	return &x
}

func CreateInt64Ptr(x int64) *int64 {
	return &x
}

func CreateBoolPtr(x bool) *bool {
	return &x
}
