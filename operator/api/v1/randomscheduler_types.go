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

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	randomSchedulerFinalizerName = "randomscheduler.finalizers.uclm.es"
)

// RandomSchedulerSpec defines the desired state of RandomScheduler
type RandomSchedulerSpec struct {
	// Image URI to retrieve the PgBouncer Task Docker. Depending on your case, you may point to a centralized repository with all your available images, to your Kubernetes Master machine, or to DockerHub (example value provided). Kubernetes will be in charge of downloading the image you specify and run it in the the most suitable agent for your case.
	// +optional
	Image string `json:"image"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/containers/images#updating-images
	// +optional
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent;
	// +kubebuilder:default=Always
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=2
	// The number of PgBouncer instances
	Instances *int32 `json:"instances"`

	// +optional
	// +kubebuilder:default={requests:{cpu:1,memory:"1024M"},limits:{cpu:1,memory:"1024M"}}
	// PgBouncer instances resources
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	// PgBouncer GoSec Healthchecks
	Healthchecks HealthchecksSpec `json:"healthchecks,omitempty"`

	// +optional
	// PgBouncer Constraints
	Constraints *ConstraintsSpec `json:"constraints,omitempty"`

	// +optional
	// +kubebuilder:default={type:"RollingUpdate",rollingUpdate:{maxSurge:"25%",maxUnavailable:"25%"}}
	// PgBouncer Update Strategy
	UpdateStrategy appsv1.DeploymentStrategy `json:"updateStrategy,omitempty"`

	// +optional
	// PriorityClassName is the name of the PriorityClassName cluster resource. This replaces the globalDefault priority class name. For. For more information, refer to the Kubernetes Priority Class documentation.
	PriorityClassName *string `json:"priorityClassName,omitempty"`

	// Scheduler name
	//+kubebuilder:default=influxdb-metrics-scheduler
	//+optional
	Name string `json:"name,omitempty"`

	// Log level
	//+kubebuilder:default=info
	//+optional
	LogLevel string `json:"logLevel,omitempty"`

	// +optional
	// +kubebuilder:default={"master01"}
	// User Filtered nodes
	FilterNodes []string `json:"filterNodes,omitempty"`

	// Timeout
	//+optional
	Timeout string `json:"timeout,omitempty"`
}

// RandomSchedulerStatus defines the observed state of RandomScheduler
type RandomSchedulerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RandomScheduler is the Schema for the randomschedulers API
type RandomScheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RandomSchedulerSpec   `json:"spec,omitempty"`
	Status RandomSchedulerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RandomSchedulerList contains a list of RandomScheduler
type RandomSchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RandomScheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RandomScheduler{}, &RandomSchedulerList{})
}

// IsDelete return true if the resource is being deleted
func (randomScheduler *RandomScheduler) IsDelete() bool {
	return !randomScheduler.ObjectMeta.DeletionTimestamp.IsZero()
}

// HasFinalizer returns true if finalizer is set
func (randomScheduler *RandomScheduler) HasFinalizer() bool {
	return containsString(randomScheduler.ObjectMeta.Finalizers, randomSchedulerFinalizerName)
}

// AddFinalizer adds the finalizer
func (randomScheduler *RandomScheduler) AddFinalizer() {
	randomScheduler.ObjectMeta.Finalizers = append(randomScheduler.ObjectMeta.Finalizers, randomSchedulerFinalizerName)
}

// RemoveFinalizer removes the finalizer
func (randomScheduler *RandomScheduler) RemoveFinalizer() {
	randomScheduler.ObjectMeta.Finalizers = removeString(randomScheduler.ObjectMeta.Finalizers, randomSchedulerFinalizerName)
}
