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
	influxdbMetricSchedulerFinalizerName = "influxdbmetricscheduler.finalizers.uclm.es"
)

// InfluxdbMetricsSchedulerSpec defines the desired state of InfluxdbMetricsScheduler
type InfluxdbMetricsSchedulerSpec struct {
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
	Influxdb *InfluxdbSpec `json:"influxdb,omitempty"`

	// +optional
	// Metric
	Metric *InfluxMetricSpec `json:"metric,omitempty"`

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

// TimescaledbSpec (TimescaledbSpec Specification)
type InfluxMetricSpec struct {
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
	SecondLevelOperation []string `json:"secondLevelOperation,omitempty"`
}

// InfluxdbSpec (InfluxdbSpec Specification)
type InfluxdbSpec struct {
	// +optional
	// +kubebuilder:default=influxdb-influxdb2.monitoring
	// Host to connect to influxdb.
	Host string `json:"host,omitempty"`

	// +optional
	// +kubebuilder:default="80"
	// Port to connect to influxdb.
	Port string `json:"port,omitempty"`

	// +optional
	// token to connect to influxdb.
	Token string `json:"token,omitempty"`

	// +optional
	// +kubebuilder:default=uclm
	// Organization where connect to influxdb.
	Organization string `json:"organization,omitempty"`

	// +optional
	// +kubebuilder:default=doctorado
	// Bucket where connect to influxdb.
	Bucket string `json:"bucket,omitempty"`

	// +optional
	// +kubebuilder:default=md5
	// AuthenticationType to connect to timescaledb.
	AuthenticationType string `json:"authenticationType,omitempty"`
}

// InfluxdbMetricsSchedulerStatus defines the observed state of InfluxdbMetricsScheduler
type InfluxdbMetricsSchedulerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// InfluxdbMetricsScheduler is the Schema for the influxdbmetricsschedulers API
type InfluxdbMetricsScheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InfluxdbMetricsSchedulerSpec   `json:"spec,omitempty"`
	Status InfluxdbMetricsSchedulerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InfluxdbMetricsSchedulerList contains a list of InfluxdbMetricsScheduler
type InfluxdbMetricsSchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InfluxdbMetricsScheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InfluxdbMetricsScheduler{}, &InfluxdbMetricsSchedulerList{})
}

// IsDelete return true if the resource is being deleted
func (metricScheduler *InfluxdbMetricsScheduler) IsDelete() bool {
	return !metricScheduler.ObjectMeta.DeletionTimestamp.IsZero()
}

// HasFinalizer returns true if finalizer is set
func (metricScheduler *InfluxdbMetricsScheduler) HasFinalizer() bool {
	return containsString(metricScheduler.ObjectMeta.Finalizers, influxdbMetricSchedulerFinalizerName)
}

// AddFinalizer adds the finalizer
func (metricScheduler *InfluxdbMetricsScheduler) AddFinalizer() {
	metricScheduler.ObjectMeta.Finalizers = append(metricScheduler.ObjectMeta.Finalizers, influxdbMetricSchedulerFinalizerName)
}

// RemoveFinalizer removes the finalizer
func (metricScheduler *InfluxdbMetricsScheduler) RemoveFinalizer() {
	metricScheduler.ObjectMeta.Finalizers = removeString(metricScheduler.ObjectMeta.Finalizers, influxdbMetricSchedulerFinalizerName)
}
