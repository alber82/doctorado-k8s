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

package influxdbmetricsscheduler

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"scheduler-operator/internal/controller/influxdbmetricsscheduler/common"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	schedulerv1 "scheduler-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReconciliationOnError time.Duration = 20 * time.Second
	ReconciliationOnOk    time.Duration = 120 * time.Second
)

// InfluxdbMetricsSchedulerReconciler reconciles a InfluxdbMetricsScheduler object
type InfluxdbMetricsSchedulerReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=influxdbmetricsschedulers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=influxdbmetricsschedulers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=influxdbmetricsschedulers/finalizers,verbs=update

// +kubebuilder:rbac:groups=scheduler.uclm.es,resources=configmaps,verbs=create;get;list;patch;update;watch;delete;deletecollection
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;patch;update;watch
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=get	;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Annotation for generating RBAC role for writing Events
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *InfluxdbMetricsSchedulerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("influxdbmetricscheduler", req.NamespacedName)
	log.V(1).Info("Reconciling metricScheduler")

	var influxdbMetricsScheduler schedulerv1.InfluxdbMetricsScheduler

	if err := r.Get(ctx, req.NamespacedName, &influxdbMetricsScheduler); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	labels := influxdbMetricsScheduler.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[common.MetricSchedulerNameLabel] = influxdbMetricsScheduler.Name

	influxdbMetricsSchedulerList := &schedulerv1.InfluxdbMetricsSchedulerList{}
	_ = r.Client.List(ctx, influxdbMetricsSchedulerList, client.InNamespace(req.Namespace))

	switch {
	case influxdbMetricsScheduler.IsDelete():
		if influxdbMetricsScheduler.HasFinalizer() {
			if err := r.deleteTsMetricsScheduler(ctx, &influxdbMetricsScheduler, log); err != nil {
				log.Error(err, "Cannot complete metric scheduler deletion")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: ReconciliationOnError,
				}, err
			}

			influxdbMetricsScheduler.RemoveFinalizer()
			if err := r.Update(ctx, &influxdbMetricsScheduler); err != nil {
				log.Error(err, "Cannot update metric scheduler after removing finalizer")
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: ReconciliationOnError,
				}, err
			}
			log.Info("Removed finalizer successfully")
		}
	case !influxdbMetricsScheduler.HasFinalizer():
		influxdbMetricsScheduler.AddFinalizer()
		if err := r.Update(ctx, &influxdbMetricsScheduler); err != nil {
			log.Error(err, "Cannot update metric scheduler after adding finalizer")
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: ReconciliationOnError,
			}, err
		}
		log.Info("Added finalizer successfully")
	}

	_, err := r.createOrUpdateServiceAccount(ctx, &influxdbMetricsScheduler, log, labels)

	if err != nil {
		log.Error(err, "There was an error on create/update service account")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: ReconciliationOnError,
		}, err
	}

	_, err = r.createOrUpdateClusterRoleBinding(ctx, &influxdbMetricsScheduler, log, labels)

	if err != nil {
		log.Error(err, "There was an error on create/update cluster role binding")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: ReconciliationOnError,
		}, err
	}

	_, err = r.createOrUpdateDeployment(ctx, &influxdbMetricsScheduler, log, labels)

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
func (r *InfluxdbMetricsSchedulerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&schedulerv1.InfluxdbMetricsScheduler{}).
		Complete(r)
}
