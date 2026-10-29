// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=defaults.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Input holds the platform-owned defaults that developers cannot set on an XWebApp.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Tolerations are applied to every pod of every XWebApp.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// Gateway is the Gateway that HTTPRoutes attach to.
	// +optional
	Gateway GatewayConfig `json:"gateway,omitempty"`
}

// GatewayConfig identifies the Gateway that HTTPRoutes attach to.
type GatewayConfig struct {
	// Name of the Gateway. Defaults to the XR's namespace.
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace of the Gateway. Defaults to the XR's namespace.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}
