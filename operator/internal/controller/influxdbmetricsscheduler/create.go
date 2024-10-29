package influxdbmetricsscheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduler-operator/internal/controller/influxdbmetricsscheduler/common"
	"strconv"
	"strings"

	schedulerv1 "scheduler-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *InfluxdbMetricsSchedulerReconciler) createOrUpdateClusterRoleBinding(ctx context.Context, influxdbMetricsScheduler *schedulerv1.InfluxdbMetricsScheduler, log logr.Logger, labels map[string]string) (clusterRoleBinding *rbacv1.ClusterRoleBinding, err error) {
	roleBindingName := influxdbMetricsScheduler.Name

	clusterRoleBinding = &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ClusterRoleBinding"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        influxdbMetricsScheduler.Name,
			Labels:      labels,
			Annotations: influxdbMetricsScheduler.Annotations,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "system:kube-scheduler",
		},

		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      influxdbMetricsScheduler.Name,
				Namespace: influxdbMetricsScheduler.Namespace,
			},
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, common.Update(influxdbMetricsScheduler, clusterRoleBinding, r.Scheme, labels, func() error {

		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      influxdbMetricsScheduler.Name,
				Namespace: influxdbMetricsScheduler.Namespace,
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
		r.recordEventFromOperationResult(influxdbMetricsScheduler, opResult, fmt.Sprintf("Created or updated Role Binding %s", clusterRoleBinding.Name))
	}

	return clusterRoleBinding, nil
}

func (r *InfluxdbMetricsSchedulerReconciler) createOrUpdateServiceAccount(ctx context.Context, influxdbMetricsScheduler *schedulerv1.InfluxdbMetricsScheduler, log logr.Logger, labels map[string]string) (svcAccount *corev1.ServiceAccount, err error) {

	saLabels := make(map[string]string)

	for k, v := range labels {
		saLabels[k] = v
	}

	saLabels["app"] = influxdbMetricsScheduler.Name
	saLabels["component"] = influxdbMetricsScheduler.Name

	svcAccountName := influxdbMetricsScheduler.Name
	svcAccount = &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: rbacv1.ServiceAccountKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:        svcAccountName,
			Namespace:   influxdbMetricsScheduler.Namespace,
			Labels:      saLabels,
			Annotations: influxdbMetricsScheduler.Annotations,
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, svcAccount, common.Update(influxdbMetricsScheduler, svcAccount, r.Scheme, labels))

	if err != nil {
		error := fmt.Errorf("could __NOT__ create or update Service Account, name=%s, error=%w", svcAccountName, err)
		return nil, errors.Unwrap(error)
	}

	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Service Account", "name", svcAccount.Name, "operation", opResult)
		r.recordEventFromOperationResult(influxdbMetricsScheduler, opResult, fmt.Sprintf("Created or updated Service Account %s", svcAccount.Name))
	}
	return svcAccount, nil
}

func (r *InfluxdbMetricsSchedulerReconciler) createOrUpdateDeployment(ctx context.Context, influxdbMetricsScheduler *schedulerv1.InfluxdbMetricsScheduler, log logr.Logger, labels map[string]string) (deployment *appsv1.Deployment, err error) {

	deploymentLabels := make(map[string]string)

	for k, v := range labels {
		deploymentLabels[k] = v
	}

	deploymentLabels["app"] = influxdbMetricsScheduler.Name

	if influxdbMetricsScheduler.Spec.Image == "" {
		err := errors.New("cannot find influxdbMetricsScheduler configuration, please check your influxdbMetricsScheduler")
		return nil, err
	}

	deploymentName := influxdbMetricsScheduler.Name

	deployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName,
			Namespace:   influxdbMetricsScheduler.Namespace,
			Labels:      deploymentLabels,
			Annotations: influxdbMetricsScheduler.Annotations,
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, deployment, common.Update(influxdbMetricsScheduler, deployment, r.Scheme, labels, func() error {

		if deployment.ObjectMeta.CreationTimestamp.IsZero() {
			deployment.Spec = appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: deploymentLabels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      deploymentLabels,
						Annotations: influxdbMetricsScheduler.Annotations,
					},
					Spec: corev1.PodSpec{},
				},
			}
		}

		deployment.Spec.Replicas = influxdbMetricsScheduler.Spec.Instances
		deployment.Spec.Strategy = influxdbMetricsScheduler.Spec.UpdateStrategy

		deployment.Spec.Template.Spec = corev1.PodSpec{
			ServiceAccountName: influxdbMetricsScheduler.Name,
			Hostname:           "influxdbMetricsScheduler",
			Subdomain:          influxdbMetricsScheduler.Name,
			Containers: []corev1.Container{
				{
					Name:            influxdbMetricsScheduler.Name,
					Image:           influxdbMetricsScheduler.Spec.Image,
					ImagePullPolicy: influxdbMetricsScheduler.Spec.ImagePullPolicy,

					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    *influxdbMetricsScheduler.Spec.Resources.Requests.Cpu(),
							corev1.ResourceMemory: *influxdbMetricsScheduler.Spec.Resources.Requests.Memory(),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    *influxdbMetricsScheduler.Spec.Resources.Limits.Cpu(),
							corev1.ResourceMemory: *influxdbMetricsScheduler.Spec.Resources.Limits.Memory(),
						},
					},
					Env: []corev1.EnvVar{
						{Name: "SCHEDULER_NAME", Value: influxdbMetricsScheduler.Name},
						//METRIC SPEC
						{Name: "METRIC_NAME", Value: influxdbMetricsScheduler.Spec.Metric.Name},
						{Name: "METRIC_START_DATE", Value: influxdbMetricsScheduler.Spec.Metric.StartDate},
						{Name: "METRIC_END_DATE", Value: influxdbMetricsScheduler.Spec.Metric.EndDate},
						{Name: "METRIC_OPERATION", Value: strings.Replace(influxdbMetricsScheduler.Spec.Metric.Operation, ";", ",", -1)},
						{Name: "METRIC_PRIORITY_ORDER", Value: influxdbMetricsScheduler.Spec.Metric.PriorityOrder},
						{Name: "METRIC_FILTER_CLAUSE", Value: strings.Replace(strings.Join(influxdbMetricsScheduler.Spec.Metric.FilterClause, ","), ";", ",", -1)},
						{Name: "METRIC_IS_SECOND_LEVEL", Value: strconv.FormatBool(influxdbMetricsScheduler.Spec.Metric.IsSecondLevel)},
						{Name: "METRIC_SECOND_LEVEL_GROUP", Value: strings.Join(influxdbMetricsScheduler.Spec.Metric.SecondLevelGroup, ",")},
						{Name: "METRIC_SECOND_LEVEL_OPERATION", Value: strings.Replace(strings.Join(influxdbMetricsScheduler.Spec.Metric.SecondLevelOperation, ","), ";", ",", -1)},
						//TIMESCALEDB SPEC
						{Name: "INFLUXDB_HOST", Value: influxdbMetricsScheduler.Spec.Influxdb.Host},
						{Name: "INFLUXDB_PORT", Value: influxdbMetricsScheduler.Spec.Influxdb.Port},
						{Name: "INFLUXDB_TOKEN", Value: influxdbMetricsScheduler.Spec.Influxdb.Token},
						{Name: "INFLUXDB_ORGANIZATION", Value: influxdbMetricsScheduler.Spec.Influxdb.Organization},
						{Name: "INFLUXDB_BUCKET", Value: influxdbMetricsScheduler.Spec.Influxdb.Bucket},
						//OTHERS
						{Name: "LOG_LEVEL", Value: influxdbMetricsScheduler.Spec.LogLevel},
						{Name: "TIMEOUT", Value: influxdbMetricsScheduler.Spec.Timeout},
						{Name: "FILTERED_NODES", Value: strings.Join(influxdbMetricsScheduler.Spec.FilterNodes, ",")},
					},
					TerminationMessagePath:   "/dev/termination-log",
					TerminationMessagePolicy: corev1.TerminationMessageReadFile,
				},
			},
			RestartPolicy:                 corev1.RestartPolicyAlways,
			TerminationGracePeriodSeconds: common.CreateInt64Ptr(30),
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 "default-scheduler",
		}

		if influxdbMetricsScheduler.Spec.PriorityClassName != nil && *influxdbMetricsScheduler.Spec.PriorityClassName != "" {
			deployment.Spec.Template.Spec.PriorityClassName = *influxdbMetricsScheduler.Spec.PriorityClassName
		}

		return nil
	}))

	if err != nil {
		error := fmt.Errorf("Could __NOT__ create or update Deployment, name=%s, error=%w", deploymentName, err)
		return nil, errors.Unwrap(error)
	}
	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Deployment", "name", deployment.Name, "operation", opResult)
		r.recordEventFromOperationResult(influxdbMetricsScheduler, opResult, fmt.Sprintf("Created or updated Deployment %s", deployment.Name))
	}

	return deployment, nil
}
