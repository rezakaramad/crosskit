// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=defaults.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Input can be used to provide input to this Function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ArgoCDApp holds the ArgoCD App Registration and Enterprise Application IDs.
	ArgoCDApp ArgoCDAppConfig `json:"argocdApp"`

	// ProviderConfigRef references the AzureAD ProviderConfig to use.
	ProviderConfigRef ProviderConfigRef `json:"providerConfigRef"`
}

type ArgoCDAppConfig struct {
	// AppRegObjectID is the object ID of the ArgoCD App Registration.
	// +kubebuilder:validation:MinLength=1
	AppRegObjectID string `json:"appRegObjectID"`

	// EnterpriseAppObjectID is the object ID of the ArgoCD Enterprise Application.
	// +kubebuilder:validation:MinLength=1
	EnterpriseAppObjectID string `json:"enterpriseAppObjectID"`
}

type ProviderConfigRef struct {
	// Name is the name of the ProviderConfig to use.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Kind is the kind of the ProviderConfig.
	// +kubebuilder:validation:MinLength=1
	Kind string `json:"kind"`
}
