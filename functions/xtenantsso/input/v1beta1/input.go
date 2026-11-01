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

	// Apps lists the applications the tenant gets single sign-on to. A Group,
	// an AppRole and a RoleAssignment are created for each one.
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=name
	Apps []AppConfig `json:"apps"`

	// ProviderConfigRef references the AzureAD ProviderConfig to use.
	ProviderConfigRef ProviderConfigRef `json:"providerConfigRef"`
}

// AppConfig identifies one application in Entra ID.
type AppConfig struct {
	// Name is a short identifier used in resource names, such as argocd.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=20
	// +kubebuilder:validation:Pattern=`^[a-z][a-z0-9-]*[a-z0-9]$`
	Name string `json:"name"`

	// DisplayName is the application's name in group names, role descriptions and
	// condition types, such as ArgoCD.
	// +kubebuilder:validation:Pattern=`^[A-Za-z][A-Za-z0-9]*$`
	DisplayName string `json:"displayName"`

	// AppRegObjectID is the object ID of the application's App Registration.
	// +kubebuilder:validation:MinLength=1
	AppRegObjectID string `json:"appRegObjectID"`

	// EnterpriseAppObjectID is the object ID of the application's Enterprise Application.
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
