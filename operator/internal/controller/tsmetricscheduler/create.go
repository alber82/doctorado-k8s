package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduler-operator/internal/controller/tsmetricscheduler/common"
	"strconv"
	"strings"

	schedulerv1 "scheduler-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *TsMetricsSchedulerReconciler) createOrUpdateClusterRoleBinding(ctx context.Context, metricScheduler *schedulerv1.TsMetricsScheduler, log logr.Logger, labels map[string]string) (clusterRoleBinding *rbacv1.ClusterRoleBinding, err error) {
	roleBindingName := metricScheduler.Name

	clusterRoleBinding = &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ClusterRoleBinding"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        metricScheduler.Name,
			Labels:      labels,
			Annotations: metricScheduler.Annotations,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "system:kube-scheduler",
		},

		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      metricScheduler.Name,
				Namespace: metricScheduler.Namespace,
			},
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, common.Update(metricScheduler, clusterRoleBinding, r.Scheme, labels, func() error {

		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      metricScheduler.Name,
				Namespace: metricScheduler.Namespace,
			},
		}

		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "system:kube-scheduler",
		}

		return nil
	}))

	if err != nil {
		error := fmt.Errorf("could __NOT__ create or update Role Binding, name=%s, error=%w", roleBindingName, err)
		return nil, errors.Unwrap(error)
	}
	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Role Binding", "name", clusterRoleBinding.Name, "operation", opResult)
		r.recordEventFromOperationResult(metricScheduler, opResult, fmt.Sprintf("Created or updated Role Binding %s", clusterRoleBinding.Name))
	}

	return clusterRoleBinding, nil
}

func (r *TsMetricsSchedulerReconciler) createOrUpdateServiceAccount(ctx context.Context, tsMetricScheduler *schedulerv1.TsMetricsScheduler, log logr.Logger, labels map[string]string) (svcAccount *corev1.ServiceAccount, err error) {

	saLabels := make(map[string]string)

	for k, v := range labels {
		saLabels[k] = v
	}

	saLabels["app"] = tsMetricScheduler.Name
	saLabels["component"] = tsMetricScheduler.Name

	svcAccountName := tsMetricScheduler.Name
	svcAccount = &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: rbacv1.ServiceAccountKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:        svcAccountName,
			Namespace:   tsMetricScheduler.Namespace,
			Labels:      saLabels,
			Annotations: tsMetricScheduler.Annotations,
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, svcAccount, common.Update(tsMetricScheduler, svcAccount, r.Scheme, labels))

	if err != nil {
		error := fmt.Errorf("could __NOT__ create or update Service Account, name=%s, error=%w", svcAccountName, err)
		return nil, errors.Unwrap(error)
	}

	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Service Account", "name", svcAccount.Name, "operation", opResult)
		r.recordEventFromOperationResult(tsMetricScheduler, opResult, fmt.Sprintf("Created or updated Service Account %s", svcAccount.Name))
	}
	return svcAccount, nil
}

func (r *TsMetricsSchedulerReconciler) createOrUpdateDeployment(ctx context.Context, tsMetricScheduler *schedulerv1.TsMetricsScheduler, log logr.Logger, labels map[string]string) (deployment *appsv1.Deployment, err error) {

	deploymentLabels := make(map[string]string)

	for k, v := range labels {
		deploymentLabels[k] = v
	}

	deploymentLabels["app"] = tsMetricScheduler.Name

	if tsMetricScheduler.Spec.Image == "" {
		err := errors.New("cannot find pgBouncer configuration, please check your pgBouncer")
		return nil, err
	}

	deploymentName := tsMetricScheduler.Name

	deployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName,
			Namespace:   tsMetricScheduler.Namespace,
			Labels:      deploymentLabels,
			Annotations: tsMetricScheduler.Annotations,
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, deployment, common.Update(tsMetricScheduler, deployment, r.Scheme, labels, func() error {

		if deployment.ObjectMeta.CreationTimestamp.IsZero() {
			deployment.Spec = appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: deploymentLabels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      deploymentLabels,
						Annotations: tsMetricScheduler.Annotations,
					},
					Spec: corev1.PodSpec{},
				},
			}
		}

		deployment.Spec.Replicas = tsMetricScheduler.Spec.Instances
		deployment.Spec.Strategy = tsMetricScheduler.Spec.UpdateStrategy

		deployment.Spec.Template.Spec = corev1.PodSpec{
			ServiceAccountName: tsMetricScheduler.Name,
			Hostname:           "metricscheduler",
			Subdomain:          tsMetricScheduler.Name,
			Containers: []corev1.Container{
				{
					Name:            tsMetricScheduler.Name,
					Image:           tsMetricScheduler.Spec.Image,
					ImagePullPolicy: tsMetricScheduler.Spec.ImagePullPolicy,

					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    *tsMetricScheduler.Spec.Resources.Requests.Cpu(),
							corev1.ResourceMemory: *tsMetricScheduler.Spec.Resources.Requests.Memory(),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    *tsMetricScheduler.Spec.Resources.Limits.Cpu(),
							corev1.ResourceMemory: *tsMetricScheduler.Spec.Resources.Limits.Memory(),
						},
					},
					Env: []corev1.EnvVar{
						{Name: "SCHEDULER_NAME", Value: tsMetricScheduler.Name},
						//METRIC SPEC
						{Name: "METRIC_NAME", Value: tsMetricScheduler.Spec.Metric.Name},
						{Name: "METRIC_START_DATE", Value: tsMetricScheduler.Spec.Metric.StartDate},
						{Name: "METRIC_END_DATE", Value: tsMetricScheduler.Spec.Metric.EndDate},
						{Name: "METRIC_OPERATION", Value: strings.Replace(tsMetricScheduler.Spec.Metric.Operation, ";", ",", -1)},
						{Name: "METRIC_PRIORITY_ORDER", Value: tsMetricScheduler.Spec.Metric.PriorityOrder},
						{Name: "METRIC_FILTER_CLAUSE", Value: strings.Replace(strings.Join(tsMetricScheduler.Spec.Metric.FilterClause, ","), ";", ",", -1)},
						{Name: "METRIC_IS_SECOND_LEVEL", Value: strconv.FormatBool(tsMetricScheduler.Spec.Metric.IsSecondLevel)},
						{Name: "METRIC_SECOND_LEVEL_GROUP", Value: strings.Join(tsMetricScheduler.Spec.Metric.SecondLevelGroup, ",")},
						{Name: "METRIC_SECOND_LEVEL_SELECT", Value: strings.Replace(strings.Join(tsMetricScheduler.Spec.Metric.SecondLevelSelect, ","), ";", ",", -1)},
						//TIMESCALEDB SPEC
						{Name: "TIMESCALEDB_HOST", Value: tsMetricScheduler.Spec.Timescaledb.Host},
						{Name: "TIMESCALEDB_PORT", Value: tsMetricScheduler.Spec.Timescaledb.Port},
						{Name: "TIMESCALEDB_USER", Value: tsMetricScheduler.Spec.Timescaledb.User},
						{Name: "TIMESCALEDB_PASSWORD", Value: tsMetricScheduler.Spec.Timescaledb.Password},
						{Name: "TIMESCALEDB_DATABASE", Value: tsMetricScheduler.Spec.Timescaledb.Database},
						{Name: "TIMESCALEDB_AUTH_TYPE", Value: tsMetricScheduler.Spec.Timescaledb.AuthenticationType},
						//OTHERS
						{Name: "LOG_LEVEL", Value: tsMetricScheduler.Spec.LogLevel},
						{Name: "TIMEOUT", Value: tsMetricScheduler.Spec.Timeout},
						{Name: "FILTERED_NODES", Value: strings.Join(tsMetricScheduler.Spec.FilterNodes, ",")},
					},
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
				},
			},
			DeprecatedServiceAccount:      tsMetricScheduler.Name,
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: common.CreateInt64Ptr(30),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 "default-scheduler",
		}

		if tsMetricScheduler.Spec.PriorityClassName != nil && *tsMetricScheduler.Spec.PriorityClassName != "" {
			deployment.Spec.Template.Spec.PriorityClassName = *tsMetricScheduler.Spec.PriorityClassName
		}

		return nil
	}))

	if err != nil {
		error := fmt.Errorf("Could __NOT__ create or update Deployment, name=%s, error=%w", deploymentName, err)
		return nil, errors.Unwrap(error)
	}
	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Deployment", "name", deployment.Name, "operation", opResult)
		r.recordEventFromOperationResult(tsMetricScheduler, opResult, fmt.Sprintf("Created or updated Deployment %s", deployment.Name))
	}

	return deployment, nil
}
