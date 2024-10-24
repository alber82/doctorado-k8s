/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"scheduler-operator/internal/controller/metricscheduler/common"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	schedulerv1 "scheduler-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MetricSchedulerReconciler reconciles a MetricScheduler object
type MetricSchedulerReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

const (
	ReconciliationOnError time.Duration = 20 * time.Second
	ReconciliationOnOk    time.Duration = 120 * time.Second
)

//+kubebuilder:rbac:groups=scheduler.uclm.es,resources=metricschedulers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scheduler.uclm.es,resources=metricschedulers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scheduler.uclm.es,resources=metricschedulers/finalizers,verbs=update

// Annotation for generating RBAC role for scheduler Objects
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=create;get;list;patch;update;watch;delete;deletecollection
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=get	;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Annotation for generating RBAC role for writing Events
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MetricScheduler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *MetricSchedulerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("metricscheduler", req.NamespacedName)
	log.V(1).Info("Reconciling metricScheduler")

	var metricScheduler schedulerv1.MetricScheduler

	if err := r.Get(ctx, req.NamespacedName, &metricScheduler); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	labels := metricScheduler.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[common.MetricSchedulerNameLabel] = metricScheduler.Name

	metricSchedulerList := &schedulerv1.MetricSchedulerList{}
	_ = r.Client.List(ctx, metricSchedulerList, client.InNamespace(req.Namespace))

	switch {
	case metricScheduler.IsDelete():
		if metricScheduler.HasFinalizer() {
			if err := r.deleteMetricScheduler(ctx, &metricScheduler, log); err != nil {
				log.Error(err, "Cannot complete metric scheduler deletion")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: ReconciliationOnError,
				}, err
			}

			metricScheduler.RemoveFinalizer()
			if err := r.Update(ctx, &metricScheduler); err != nil {
				log.Error(err, "Cannot update metric scheduler after removing finalizer")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: ReconciliationOnError,
				}, err
			}
			log.Info("Removed finalizer successfully")
		}
	case !metricScheduler.HasFinalizer():
		metricScheduler.AddFinalizer()
		if err := r.Update(ctx, &metricScheduler); err != nil {
			log.Error(err, "Cannot update metric scheduler after adding finalizer")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: ReconciliationOnError,
			}, err
		}
		log.Info("Added finalizer successfully")
	}

	_, err := r.createOrUpdateServiceAccount(ctx, &metricScheduler, log, labels)

	if err != nil {
		log.Error(err, "There was an error on create/update service account")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: ReconciliationOnError,
		}, err
	}

	//_, err = r.createOrUpdateClusterRoleBinding(ctx, &metricScheduler, log, labels)
	//
	//if err != nil {
	//	log.Error(err, "There was an error on create/update cluster role binding")
	//	return ctrl.Result{
	//		Requeue:      true,
	//		RequeueAfter: ReconciliationOnError,
	//	}, err
	//}

	_, err = r.createOrUpdateDeployment(ctx, &metricScheduler, log, labels)

	if err != nil {
		log.Error(err, "There was an error on create/update metricScheduler deployment")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: ReconciliationOnError,
		}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MetricSchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulerv1.MetricScheduler{}).
		Complete(r)
}
