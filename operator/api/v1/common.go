package v1

import (
	corev1 "k8s.io/api/core/v1"
)

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

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
