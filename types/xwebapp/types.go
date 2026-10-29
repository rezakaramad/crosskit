// Package xwebapp defines the XWebApp composite resource type.
package xwebapp

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() >= 3 && self.metadata.name.size() <= 30 && self.metadata.name.matches('^[a-z][a-z0-9-]*[a-z0-9]$')",message="metadata.name must be 3-30 chars, a valid RFC 1035 DNS label (lowercase letters, digits, hyphens; start with a letter and end alphanumeric)"
type XWebApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              XWebAppSpec   `json:"spec"`
	Status            XWebAppStatus `json:"status,omitempty"`
}

// XWebAppSpec mirrors the Deployment fields developers own. Everything else
// (selector, service account, scheduling) is set by the platform.
type XWebAppSpec struct {
	// Replicas defaults to 2 when unset.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Strategy appsv1.DeploymentStrategy `json:"strategy,omitempty"`

	// Template is copied into the Deployment as is; containers, initContainers
	// and volumes are managed by the developer.
	Template corev1.PodTemplateSpec `json:"template"`

	// +optional
	PlatformCapabilities *PlatformCapabilities `json:"platformCapabilities,omitempty"`
}

// PlatformCapabilities requests optional platform features.
type PlatformCapabilities struct {
	// Hostname exposes the app through the shared gateway.
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$`
	// +optional
	Hostname string `json:"hostname,omitempty"`
}

type XWebAppStatus struct{}
