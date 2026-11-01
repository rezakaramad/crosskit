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

	// Project is the Argo CD AppProject all generated Applications belong to.
	// +kubebuilder:validation:MinLength=1
	Project string `json:"project"`

	// TenantLabelKey is the label key set on every generated resource, with the
	// tenant name as its value.
	// +kubebuilder:validation:MinLength=1
	TenantLabelKey string `json:"tenantLabelKey"`

	// ManagementCluster is the Argo CD destination name of the management
	// cluster. Defaults to "in-cluster", the cluster Argo CD itself runs in.
	// +kubebuilder:default="in-cluster"
	ManagementCluster string `json:"managementCluster,omitempty"`

	// SSO configures the Application that deploys the tenant-sso chart to the
	// management cluster.
	SSO ApplicationConfig `json:"sso"`

	// Management configures the Application that deploys the tenant-management
	// chart to the management cluster.
	Management ApplicationConfig `json:"management"`

	// Workload configures the ApplicationSet that deploys the tenant-workload
	// chart to the workload clusters.
	Workload ApplicationSetConfig `json:"workload"`

	// Namespace configures where the Argo CD resources are created.
	Namespace NamespaceConfig `json:"namespace"`
}

// ApplicationConfig configures an Application deployed to the management cluster.
type ApplicationConfig struct {
	// RepoURL is the URL of the Git repository hosting the chart.
	// +kubebuilder:validation:MinLength=1
	RepoURL string `json:"repoURL"`

	// Path is the path within the repository to the chart.
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`

	// TargetRevision is the Git revision the Application tracks.
	// +kubebuilder:default="HEAD"
	TargetRevision string `json:"targetRevision,omitempty"`

	// TargetNamespace is the namespace in the management cluster the
	// Application deploys into.
	// +kubebuilder:validation:MinLength=1
	TargetNamespace string `json:"targetNamespace"`

	// Helm configures additional Helm value files and parameters.
	Helm HelmConfig `json:"helm,omitempty"`
}

// ApplicationSetConfig configures the ApplicationSet that generates one Application
// per matching workload cluster.
type ApplicationSetConfig struct {
	// RepoURL is the URL of the Git repository hosting the workload chart.
	// +kubebuilder:validation:MinLength=1
	RepoURL string `json:"repoURL"`

	// Path is the path within the repository to the workload chart.
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`

	// TargetRevision is the Git revision the Applications track.
	// +kubebuilder:default="HEAD"
	TargetRevision string `json:"targetRevision,omitempty"`

	// Helm configures the Helm source for the generated Applications.
	Helm HelmConfig `json:"helm,omitempty"`

	// TargetClusters selects the workload clusters to deploy to.
	TargetClusters TargetClustersConfig `json:"targetClusters"`
}

// HelmConfig configures Helm value files and parameters for an Application source.
type HelmConfig struct {
	// ValueFiles lists Helm values files to apply. For the workload
	// ApplicationSet, Go template expressions are evaluated at render time.
	ValueFiles []string `json:"valueFiles,omitempty"`

	// Parameters lists Helm parameters to pass. For the workload
	// ApplicationSet, values may contain Go template expressions evaluated at
	// render time.
	Parameters []HelmParameter `json:"parameters,omitempty"`
}

// HelmParameter is a Helm --set parameter.
type HelmParameter struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Value string `json:"value"`
}

// TargetClustersConfig selects the workload clusters by label.
type TargetClustersConfig struct {
	// ClusterTypeKey is the cluster label key identifying the cluster type.
	// +kubebuilder:default="argocd.rezakara.demo/cluster-type"
	ClusterTypeKey string `json:"clusterTypeKey,omitempty"`

	// ClusterType is the cluster type value the ApplicationSet targets.
	// +kubebuilder:default="tenant"
	ClusterType string `json:"clusterType,omitempty"`

	// EnvironmentKey is the cluster label key identifying the environment. Its
	// value is used as <ENV> in the generated Application name.
	// +kubebuilder:default="argocd.rezakara.demo/environment"
	EnvironmentKey string `json:"environmentKey,omitempty"`

	// Environments are the environment label values the ApplicationSet targets.
	// +kubebuilder:default={"development"}
	Environments []string `json:"environments,omitempty"`
}

// NamespaceConfig configures the namespaces the Argo CD resources use.
type NamespaceConfig struct {
	// Argo is the namespace the Applications and the ApplicationSet are created
	// in. It must be a namespace Argo CD's controllers watch (typically the
	// Argo CD installation namespace).
	// +kubebuilder:default="argocd"
	Argo string `json:"argo,omitempty"`
}
