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
	metricSchedulerFinalizerName = "metricscheduler.finalizers.uclm.es"
)

// MetricSchedulerSpec defines the desired state of MetricScheduler
type MetricSchedulerSpec struct {
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
	// Timescaledb
	Timescaledb *TimescaledbSpec `json:"timescaledb,omitempty"`

	// +optional
	// Metric
	Metric *MetricSpec `json:"metric,omitempty"`

	// +optional
	// +kubebuilder:default={type:"RollingUpdate",rollingUpdate:{maxSurge:"25%",maxUnavailable:"25%"}}
	// PgBouncer Update Strategy
	UpdateStrategy appsv1.DeploymentStrategy `json:"updateStrategy,omitempty"`

	// +optional
	// PriorityClassName is the name of the PriorityClassName cluster resource. This replaces the globalDefault priority class name. For. For more information, refer to the Kubernetes Priority Class documentation.
	PriorityClassName *string `json:"priorityClassName,omitempty"`

	// Scheduler name
	//+optional
	Name string `json:"name,omitempty"`

	// Log level
	//+kubebuilder:default=info
	//+optional
	LogLevel string `json:"logLevel,omitempty"`

	// +optional
	// +kubebuilder:default={""}
	// User Filtered nodes
	FilterNodes []string `json:"filterNodes,omitempty"`

	// Timeout
	//+optional
	Timeout string `json:"timeout,omitempty"`
}

// TimescaledbSpec (TimescaledbSpec Specification)
type MetricSpec struct {
	// Metric name.
	Name string `json:"name,omitempty"`

	// +optional
	// Metric start date.
	StartDate string `json:"startDate,omitempty"`

	// +optional
	// Metric end date.
	EndDate string `json:"endDate,omitempty"`

	// +optional
	// +kubebuilder:default=max
	// Metric operation.
	Operation string `json:"operation,omitempty"`

	// +optional
	// +kubebuilder:default=desc
	// Metric priority order.
	PriorityOrder string `json:"priorityOrder,omitempty"`

	// +optional
	// +kubebuilder:default={""}
	// Others filters to apply
	FilterClause []string `json:"filters,omitempty"`

	// +optional
	// +kubebuilder:default=false
	// Others filters to apply
	IsSecondLevel bool `json:"isSecondLevel,omitempty"`

	// +optional
	// Others filters to apply
	SecondLevelGroup []string `json:"secondLevelGroup,omitempty"`

	// +optional
	// Others filters to apply
	SecondLevelSelect []string `json:"secondLevelSelect,omitempty"`
}

// TimescaledbSpec (TimescaledbSpec Specification)
type TimescaledbSpec struct {
	// +optional
	// +kubebuilder:default=timescaledb.monitoring
	// Host to connect to timescaledb.
	Host string `json:"host,omitempty"`

	// +optional
	// +kubebuilder:default="5432"
	// Port to connect to timescaledb.
	Port string `json:"port,omitempty"`

	// +optional
	// +kubebuilder:default=postgres
	// User to connect to timescaledb.
	User string `json:"user,omitempty"`

	// +optional
	// +kubebuilder:default=postgres
	// Password to connect to timescaledb.
	Password string `json:"password,omitempty"`

	// +optional
	// +kubebuilder:default=tsdb
	// Database to connect to timescaledb.
	Database string `json:"database,omitempty"`

	// +optional
	// +kubebuilder:default=md5
	// AuthenticationType to connect to timescaledb.
	AuthenticationType string `json:"authenticationType,omitempty"`
}

// ConstraintsSpec (Constraints Specification)
type ConstraintsSpec struct {
	// +optional
	// Describes affinity scheduling rules.
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +optional
	// Describes TopologySpreadConstraint scheduling rules.
	TopologySpreadConstraint []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	// +optional
	// Describes tolerations scheduling rules.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// +optional
	// Describes nodeSelector scheduling rules.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// HealthchecksSpec (Healthchecks Specification)
type HealthchecksSpec struct {
	// +optional
	// Startup Probe
	StartupProbe *corev1.Probe `json:"startupProbe,omitempty"`

	// +optional
	// +kubebuilder:default={initialDelaySeconds: 15, periodSeconds: 15, timeoutSeconds: 1, successThreshold: 1, failureThreshold: 3}
	// Readiness Probe
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// +optional
	// +kubebuilder:default={initialDelaySeconds: 2, periodSeconds: 5, timeoutSeconds: 1, successThreshold: 1, failureThreshold: 3}
	// Liveness Probe
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
}

// MetricScheduler is the Schema for the metricschedulers API
type MetricSchedulerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MetricScheduler is the Schema for the metricschedulers API
type MetricScheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetricSchedulerSpec   `json:"spec,omitempty"`
	Status MetricSchedulerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MetricSchedulerList contains a list of MetricScheduler
type MetricSchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetricScheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetricScheduler{}, &MetricSchedulerList{})
}

// IsDelete return true if the resource is being deleted
func (metricScheduler *MetricScheduler) IsDelete() bool {
	return !metricScheduler.ObjectMeta.DeletionTimestamp.IsZero()
}

// HasFinalizer returns true if finalizer is set
func (metricScheduler *MetricScheduler) HasFinalizer() bool {
	return containsString(metricScheduler.ObjectMeta.Finalizers, metricSchedulerFinalizerName)
}

// AddFinalizer adds the finalizer
func (metricScheduler *MetricScheduler) AddFinalizer() {
	metricScheduler.ObjectMeta.Finalizers = append(metricScheduler.ObjectMeta.Finalizers, metricSchedulerFinalizerName)
}

// RemoveFinalizer removes the finalizer
func (metricScheduler *MetricScheduler) RemoveFinalizer() {
	metricScheduler.ObjectMeta.Finalizers = removeString(metricScheduler.ObjectMeta.Finalizers, metricSchedulerFinalizerName)
}
