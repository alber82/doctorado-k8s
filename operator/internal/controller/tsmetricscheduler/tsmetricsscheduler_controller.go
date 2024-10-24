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
	"scheduler-operator/internal/controller/tsmetricscheduler/common"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	schedulerv1 "scheduler-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TsMetricsSchedulerReconciler reconciles a TsMetricsScheduler object
type TsMetricsSchedulerReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

const (
	ReconciliationOnError time.Duration = 20 * time.Second
	ReconciliationOnOk    time.Duration = 120 * time.Second
)

// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=tsmetricsschedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=tsmetricsschedulers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=tsmetricsschedulers/finalizers,verbs=update

// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=configmaps,verbs=create;get;list;patch;update;watch;delete;deletecollection
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=get	;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Annotation for generating RBAC role for writing Events
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TsMetricsScheduler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *TsMetricsSchedulerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("metricscheduler", req.NamespacedName)
	log.V(1).Info("Reconciling metricScheduler")

	var tsMetricsScheduler schedulerv1.TsMetricsScheduler

	if err := r.Get(ctx, req.NamespacedName, &tsMetricsScheduler); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	labels := tsMetricsScheduler.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[common.MetricSchedulerNameLabel] = tsMetricsScheduler.Name

	tsMmetricsSchedulerList := &schedulerv1.TsMetricsSchedulerList{}
	_ = r.Client.List(ctx, tsMmetricsSchedulerList, client.InNamespace(req.Namespace))

	switch {
	case tsMetricsScheduler.IsDelete():
		if tsMetricsScheduler.HasFinalizer() {
			if err := r.deleteTsMetricsScheduler(ctx, &tsMetricsScheduler, log); err != nil {
				log.Error(err, "Cannot complete metric scheduler deletion")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: ReconciliationOnError,
				}, err
			}

			tsMetricsScheduler.RemoveFinalizer()
			if err := r.Update(ctx, &tsMetricsScheduler); err != nil {
				log.Error(err, "Cannot update metric scheduler after removing finalizer")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: ReconciliationOnError,
				}, err
			}
			log.Info("Removed finalizer successfully")
		}
	case !tsMetricsScheduler.HasFinalizer():
		tsMetricsScheduler.AddFinalizer()
		if err := r.Update(ctx, &tsMetricsScheduler); err != nil {
			log.Error(err, "Cannot update metric scheduler after adding finalizer")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: ReconciliationOnError,
			}, err
		}
		log.Info("Added finalizer successfully")
	}

	_, err := r.createOrUpdateServiceAccount(ctx, &tsMetricsScheduler, log, labels)

	if err != nil {
		log.Error(err, "There was an error on create/update service account")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: ReconciliationOnError,
		}, err
	}

	_, err = r.createOrUpdateClusterRoleBinding(ctx, &tsMetricsScheduler, log, labels)

	if err != nil {
		log.Error(err, "There was an error on create/update cluster role binding")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: ReconciliationOnError,
		}, err
	}

	_, err = r.createOrUpdateDeployment(ctx, &tsMetricsScheduler, log, labels)

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
func (r *TsMetricsSchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulerv1.TsMetricsScheduler{}).
		Complete(r)
}
