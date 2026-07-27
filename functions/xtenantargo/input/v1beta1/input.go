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

	// Project is the ArgoCD AppProject all generated Applications belong to.
	// +kubebuilder:validation:MinLength=1
	Project string `json:"project"`

	// Management configures the Application deployed to the management cluster.
	Management ManagementConfig `json:"management"`

	// Workload configures the Applications deployed to the workload clusters.
	Workload WorkloadConfig `json:"workload"`

	// Namespace configures the namespace that holds the ApplicationSet.
	Namespace NamespaceConfig `json:"namespace"`
}

// ManagementConfig configures the Application the ApplicationSet deploys to the
// management cluster.
type ManagementConfig struct {
	// RepoURL is the URL of the Git repository hosting the management chart.
	// +kubebuilder:validation:MinLength=1
	RepoURL string `json:"repoURL"`

	// Path is the path within the repository to the management chart.
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`

	// TargetRevision is the Git revision the Application tracks.
	// +kubebuilder:default="HEAD"
	TargetRevision string `json:"targetRevision,omitempty"`

	// TargetNamespace is the namespace in the management cluster the
	// Application deploys into.
	// +kubebuilder:validation:MinLength=1
	TargetNamespace string `json:"targetNamespace"`
}

// WorkloadConfig configures the Applications the ApplicationSet deploys to
// the workload clusters.
type WorkloadConfig struct {
	// RepoURL is the URL of the Git repository hosting the workload chart.
	// +kubebuilder:validation:MinLength=1
	RepoURL string `json:"repoURL"`

	// Path is the path within the repository to the workload chart.
	// +kubebuilder:validation:MinLength=1
	Path string `json:"path"`

	// TargetRevision is the Git revision the Applications track.
	// +kubebuilder:default="HEAD"
	TargetRevision string `json:"targetRevision,omitempty"`

	// Helm configures the Helm source parameters for the workload Applications.
	Helm WorkloadHelmConfig `json:"helm,omitempty"`

	// TargetClusters selects the workload clusters to deploy to.
	TargetClusters TargetClustersConfig `json:"targetClusters"`
}

// WorkloadHelmConfig configures Helm value files and parameters for the
// workload Applications.
type WorkloadHelmConfig struct {
	// ValueFiles lists Helm values files to apply. Go template expressions are
	// evaluated at ApplicationSet render time.
	ValueFiles []string `json:"valueFiles,omitempty"`

	// Parameters lists Helm parameters to pass. Values may contain Go template
	// expressions evaluated at ApplicationSet render time.
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

	// EnvironmentKey is the cluster label key identifying the environment.
	// +kubebuilder:default="argocd.rezakara.demo/environment"
	EnvironmentKey string `json:"environmentKey,omitempty"`

	// Environments are the environment label values the ApplicationSet targets.
	// +kubebuilder:default={"dev","test","prod","wl"}
	Environments []string `json:"environments,omitempty"`
}

// NamespaceConfig configures the namespaces the ApplicationSet uses.
type NamespaceConfig struct {
	// ApplicationSet is the namespace the ApplicationSet resource is created
	// in. This must be the namespace Argo CD's ApplicationSet controller
	// watches (typically the Argo CD installation namespace).
	// +kubebuilder:default="argocd"
	ApplicationSet string `json:"applicationSet,omitempty"`

	// Prefix is prepended to the tenant name to form the workload destination
	// namespace (e.g. "tn-" yields "tn-<tenant>").
	// +kubebuilder:default="tn-"
	Prefix string `json:"prefix,omitempty"`
}
