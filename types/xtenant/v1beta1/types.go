// Package v1beta1 defines the v1beta1 XTenant composite resource type.
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type XTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              XTenantSpec   `json:"spec"`
	Status            XTenantStatus `json:"status,omitempty"`
}

type XTenantSpec struct {
	// ShortName is a short DNS-label alias for the tenant, used to build gateway
	// hostnames such as <shortName>.<environment>.rezakara.demo. It is required
	// and must be a non-empty DNS label.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=10
	// +kubebuilder:validation:Pattern=`^[a-z][a-z0-9-]*[a-z0-9]$`
	ShortName string `json:"shortName"`
}

type XTenantStatus struct{}
