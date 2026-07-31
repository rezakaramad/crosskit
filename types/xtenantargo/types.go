// Package xtenantargo defines the XTenantArgo composite resource type.
package xtenantargo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() >= 3 && self.metadata.name.size() <= 21 && self.metadata.name.matches('^[a-z][a-z0-9-]*[a-z0-9]$')",message="metadata.name must be 3-21 chars, a valid RFC 1035 DNS label (lowercase letters, digits, hyphens; start with a letter and end alphanumeric)"
type XTenantArgo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              XTenantArgoSpec   `json:"spec"`
	Status            XTenantArgoStatus `json:"status,omitempty"`
}

// XTenantArgoSpec defines the tenant-specific inputs the function forwards to
// the management and workload Applications composed for the tenant.
type XTenantArgoSpec struct {
	// ShortName is a short DNS-label alias for the tenant, used to build gateway
	// hostnames such as <shortName>.<environment>.rezakara.demo. It is required
	// and must be a non-empty DNS label.
	ShortName string `json:"shortName"`
}

type XTenantArgoStatus struct{}
