// Package v1beta1 contains the input type for the xdeployment Function.
// +kubebuilder:object:generate=true
// +groupName=platform.rezakara.demo
// +versionName=v1beta1
package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Input is the configuration passed to this Function from the Composition pipeline step.
//
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NextInsight configures optional Next-Insight application metadata enrichment.
	// +optional
	NextInsight NextInsightInput `json:"nextInsight,omitempty"`
}

// NextInsightInput configures Next-Insight metadata enrichment for Deployment labels.
type NextInsightInput struct {
	// LabelPrefix is the Kubernetes label key prefix applied to all labels
	// produced from Next-Insight metadata (e.g. "nextinsight.rezakara.demo/").
	// +optional
	LabelPrefix string `json:"labelPrefix,omitempty"`
}
