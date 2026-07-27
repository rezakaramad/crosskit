// Package argocd is a minimal stand-in for Argo CD's
// argoproj.io/v1alpha1 ApplicationSet. It provides just enough of the schema to
// construct desired resources in a Crossplane composition function and
// round-trip through runtime.Unstructured.
package argocd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	GroupVersion = schema.GroupVersion{Group: "argoproj.io", Version: "v1alpha1"}

	SchemeBuilder = runtime.NewSchemeBuilder(func(s *runtime.Scheme) error {
		s.AddKnownTypes(GroupVersion, &ApplicationSet{}, &ApplicationSetList{})
		metav1.AddToGroupVersion(s, GroupVersion)
		return nil
	})

	AddToScheme = SchemeBuilder.AddToScheme
)

type ApplicationSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ApplicationSetSpec   `json:"spec"`
	Status ApplicationSetStatus `json:"status,omitempty"`
}

type ApplicationSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ApplicationSet `json:"items"`
}

// ApplicationSetSpec is the spec of an ApplicationSet.
type ApplicationSetSpec struct {
	GoTemplate                   bool                                      `json:"goTemplate,omitempty"`
	Generators                   []ApplicationSetGenerator                 `json:"generators"`
	Template                     ApplicationSetTemplate                    `json:"template"`
	GoTemplateOptions            []string                                  `json:"goTemplateOptions,omitempty"`
	IgnoreApplicationDifferences []ApplicationSetResourceIgnoreDifferences `json:"ignoreApplicationDifferences,omitempty"`
}

type ApplicationSetGenerator struct {
	Matrix   *MatrixGenerator  `json:"matrix,omitempty"`
	List     *ListGenerator    `json:"list,omitempty"`
	Git      *GitGenerator     `json:"git,omitempty"`
	Clusters *ClusterGenerator `json:"clusters,omitempty"`
}

type MatrixGenerator struct {
	Generators []ApplicationSetGenerator `json:"generators"`
}

type ListGenerator struct {
	Elements []map[string]string `json:"elements"`
	Template ApplicationSetTemplate `json:"template,omitempty"`
}

type GitGenerator struct {
	RepoURL     string                      `json:"repoURL"`
	Revision    string                      `json:"revision"`
	Directories []GitDirectoryGeneratorItem `json:"directories,omitempty"`
}

type GitDirectoryGeneratorItem struct {
	Path string `json:"path"`
}

type ClusterGenerator struct {
	Selector metav1.LabelSelector `json:"selector,omitempty"`
	Template ApplicationSetTemplate `json:"template,omitempty"`
}

type ApplicationSetTemplate struct {
	ApplicationSetTemplateMeta `json:"metadata"`
	Spec                       ApplicationSpec `json:"spec"`
}

type ApplicationSetTemplateMeta struct {
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type ApplicationSpec struct {
	Project     string                 `json:"project"`
	Source      *ApplicationSource     `json:"source,omitempty"`
	Destination ApplicationDestination `json:"destination"`
	SyncPolicy  *SyncPolicy            `json:"syncPolicy,omitempty"`
}

type ApplicationSource struct {
	RepoURL        string      `json:"repoURL"`
	Path           string      `json:"path,omitempty"`
	TargetRevision string      `json:"targetRevision,omitempty"`
	Helm           *HelmSource `json:"helm,omitempty"`
}

type HelmSource struct {
	ValueFiles []string       `json:"valueFiles,omitempty"`
	Parameters []HelmParameter `json:"parameters,omitempty"`
}

type HelmParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ApplicationDestination struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type SyncPolicy struct {
	Automated   *SyncPolicyAutomated `json:"automated,omitempty"`
	SyncOptions []string             `json:"syncOptions,omitempty"`
}

type SyncPolicyAutomated struct {
	Enabled  *bool `json:"enabled,omitempty"`
	Prune    bool  `json:"prune,omitempty"`
	SelfHeal bool  `json:"selfHeal,omitempty"`
}

type ApplicationSetResourceIgnoreDifferences struct {
	Name         string   `json:"name,omitempty"`
	JSONPointers []string `json:"jsonPointers,omitempty"`
}

// ApplicationSetStatus is the observed state of an ApplicationSet.
type ApplicationSetStatus struct {
	Conditions        []ApplicationSetCondition         `json:"conditions,omitempty"`
	ApplicationStatus []ApplicationSetApplicationStatus `json:"applicationStatus,omitempty"`
}

type ApplicationSetCondition struct {
	Type               string `json:"type"`
	Message            string `json:"message,omitempty"`
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	Status             string `json:"status"`
	Reason             string `json:"reason,omitempty"`
}

type ApplicationSetApplicationStatus struct {
	Application        string   `json:"application"`
	LastTransitionTime string   `json:"lastTransitionTime,omitempty"`
	Message            string   `json:"message,omitempty"`
	Status             string   `json:"status,omitempty"`
	Step               string   `json:"step,omitempty"`
	TargetRevisions    []string `json:"targetRevisions,omitempty"`
}
