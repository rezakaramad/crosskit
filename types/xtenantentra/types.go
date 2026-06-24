package xtenantentra

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() >= 3 && self.metadata.name.size() <= 21 && self.metadata.name.matches('^[a-z][a-z0-9-]*[a-z0-9]$')",message="metadata.name must be 3-21 chars, a valid RFC 1035 DNS label (lowercase letters, digits, hyphens; start with a letter and end alphanumeric)"
type XTenantEntra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              XTenantEntraSpec   `json:"spec"`
	Status            XTenantEntraStatus `json:"status,omitempty"`
}

type XTenantEntraSpec struct{}

type XTenantEntraStatus struct{}
