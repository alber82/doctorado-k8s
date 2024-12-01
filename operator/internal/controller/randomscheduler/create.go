package randomscheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"scheduler-operator/internal/controller/randomscheduler/common"
	"strings"

	schedulerv1 "scheduler-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *RandomSchedulerReconciler) createOrUpdateClusterRoleBinding(ctx context.Context, randomScheduler *schedulerv1.RandomScheduler, log logr.Logger, labels map[string]string) (clusterRoleBinding *rbacv1.ClusterRoleBinding, err error) {
	roleBindingName := randomScheduler.Name

	clusterRoleBinding = &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ClusterRoleBinding"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        randomScheduler.Name,
			Labels:      labels,
			Annotations: randomScheduler.Annotations,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     "system:kube-scheduler",
		},

		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      randomScheduler.Name,
				Namespace: randomScheduler.Namespace,
			},
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, common.Update(randomScheduler, clusterRoleBinding, r.Scheme, labels, func() error {

		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      randomScheduler.Name,
				Namespace: randomScheduler.Namespace,
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
		r.recordEventFromOperationResult(randomScheduler, opResult, fmt.Sprintf("Created or updated Role Binding %s", clusterRoleBinding.Name))
	}

	return clusterRoleBinding, nil
}

func (r *RandomSchedulerReconciler) createOrUpdateServiceAccount(ctx context.Context, randomScheduler *schedulerv1.RandomScheduler, log logr.Logger, labels map[string]string) (svcAccount *corev1.ServiceAccount, err error) {

	saLabels := make(map[string]string)

	for k, v := range labels {
		saLabels[k] = v
	}

	saLabels["app"] = randomScheduler.Name
	saLabels["component"] = randomScheduler.Name

	svcAccountName := randomScheduler.Name
	svcAccount = &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: rbacv1.ServiceAccountKind},
		ObjectMeta: metav1.ObjectMeta{
			Name:        svcAccountName,
			Namespace:   randomScheduler.Namespace,
			Labels:      saLabels,
			Annotations: randomScheduler.Annotations,
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, svcAccount, common.Update(randomScheduler, svcAccount, r.Scheme, labels))

	if err != nil {
		error := fmt.Errorf("could __NOT__ create or update Service Account, name=%s, error=%w", svcAccountName, err)
		return nil, errors.Unwrap(error)
	}

	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Service Account", "name", svcAccount.Name, "operation", opResult)
		r.recordEventFromOperationResult(randomScheduler, opResult, fmt.Sprintf("Created or updated Service Account %s", svcAccount.Name))
	}
	return svcAccount, nil
}

func (r *RandomSchedulerReconciler) createOrUpdateDeployment(ctx context.Context, randomScheduler *schedulerv1.RandomScheduler, log logr.Logger, labels map[string]string) (deployment *appsv1.Deployment, err error) {

	deploymentLabels := make(map[string]string)

	for k, v := range labels {
		deploymentLabels[k] = v
	}

	deploymentLabels["app"] = randomScheduler.Name

	if randomScheduler.Spec.Image == "" {
		err := errors.New("cannot find influxdbMetricsScheduler configuration, please check your influxdbMetricsScheduler")
		return nil, err
	}

	deploymentName := randomScheduler.Name

	deployment = &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName,
			Namespace:   randomScheduler.Namespace,
			Labels:      deploymentLabels,
			Annotations: randomScheduler.Annotations,
		},
	}

	opResult, err := ctrl.CreateOrUpdate(ctx, r.Client, deployment, common.Update(randomScheduler, deployment, r.Scheme, labels, func() error {

		if deployment.ObjectMeta.CreationTimestamp.IsZero() {
			deployment.Spec = appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: deploymentLabels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      deploymentLabels,
						Annotations: randomScheduler.Annotations,
					},
					Spec: corev1.PodSpec{},
				},
			}
		}

		deployment.Spec.Replicas = randomScheduler.Spec.Instances
		deployment.Spec.Strategy = randomScheduler.Spec.UpdateStrategy

		deployment.Spec.Template.Spec = corev1.PodSpec{
			ServiceAccountName: randomScheduler.Name,
			Hostname:           "randomscheduler",
			Subdomain:          randomScheduler.Name,
			Containers: []corev1.Container{
				{
					Name:            randomScheduler.Name,
					Image:           randomScheduler.Spec.Image,
					ImagePullPolicy: randomScheduler.Spec.ImagePullPolicy,

					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    *randomScheduler.Spec.Resources.Requests.Cpu(),
							corev1.ResourceMemory: *randomScheduler.Spec.Resources.Requests.Memory(),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    *randomScheduler.Spec.Resources.Limits.Cpu(),
							corev1.ResourceMemory: *randomScheduler.Spec.Resources.Limits.Memory(),
						},
					},
					Env: []corev1.EnvVar{
						{Name: "SCHEDULER_NAME", Value: randomScheduler.Name},
						//OTHERS
						{Name: "LOG_LEVEL", Value: randomScheduler.Spec.LogLevel},
						{Name: "TIMEOUT", Value: randomScheduler.Spec.Timeout},
						{Name: "FILTERED_NODES", Value: strings.Join(randomScheduler.Spec.FilterNodes, ",")},
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

		if randomScheduler.Spec.PriorityClassName != nil && *randomScheduler.Spec.PriorityClassName != "" {
			deployment.Spec.Template.Spec.PriorityClassName = *randomScheduler.Spec.PriorityClassName
		}

		return nil
	}))

	if err != nil {
		error := fmt.Errorf("Could __NOT__ create or update Deployment, name=%s, error=%w", deploymentName, err)
		return nil, errors.Unwrap(error)
	}
	if opResult != ctrlUtil.OperationResultNone {
		log.V(0).Info("Created or updated Deployment", "name", deployment.Name, "operation", opResult)
		r.recordEventFromOperationResult(randomScheduler, opResult, fmt.Sprintf("Created or updated Deployment %s", deployment.Name))
	}

	return deployment, nil
}
