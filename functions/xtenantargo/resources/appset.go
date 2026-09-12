// Package resources implements the composers that the Crossplane function uses
// to build and reconcile the child resources.
package resources

// This file describes one child resource (an Argo CD ApplicationSet) as a "composer":
// A composer is an object that knows how to produce and monitor exactly one composed (child) resource of the XR.

import (
	"fmt"

	"github.com/rezakaramad/crosskit/functions/xtenantargo/argocd"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantargo/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

const tenantLabelKey = "platform.talktorubberduck/tenant"

// ArgoCDApplicationSet composes the Argo CD ApplicationSet that deploys the
// tenant's platform charts to the management and workload clusters.
// For every child resource, we define a struct that encapsulates the observed resource
// and the desired resource specification.
type ArgoCDApplicationSet struct {
	XComposer
	ObservedResource *argocd.ApplicationSet
}

// NewArgoCDApplicationSet creates a new ArgoCDApplicationSet composer.
// It looks up the observed ApplicationSet resource and deserializes it for the readiness check.
// Returns an error if the observed resource exists but cannot be deserialized.
func NewArgoCDApplicationSet(f XContext) (*ArgoCDApplicationSet, error) {
	resourceName := resource.Name(fmt.Sprintf("applicationset-argocd-%s", f.XR.Name))
	observedStructured, err := composer.ConvertObserved[argocd.ApplicationSet](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &ArgoCDApplicationSet{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "ArgoCDApplicationSetReady",
		},
		ObservedResource: observedStructured,
	}, nil
}

// ComposeDesiredResource tells Crossplane what resource we want to create/update in the cluster and how to check if it's ready.
func (r *ArgoCDApplicationSet) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return r.ComposeDesiredResourceFrom(r.createResource())
}

// IsReady returns true when the ApplicationSet controller has finished
// reconciling and all generated Applications are up to date
// (ResourcesUpToDate condition = True).
func (r *ArgoCDApplicationSet) IsReady() bool {
	if r.ObservedResource == nil {
		return false
	}
	for _, condition := range r.ObservedResource.Status.Conditions {
		if condition.Type == "ResourcesUpToDate" && condition.Status == "True" {
			return true
		}
	}
	return false
}

// createResource constructs the two-generator ApplicationSet for the tenant:
// a list generator for the management cluster app and a clusters generator for
// the per-environment workload cluster apps.
func (r *ArgoCDApplicationSet) createResource() *argocd.ApplicationSet {
	xr := r.FunctionContext.XR
	defaults := r.FunctionContext.Defaults

	tenant := xr.GetName()
	shortName := xr.Spec.ShortName
	appSetNamespace := defaults.Namespace.ApplicationSet
	envRef := labelRef(defaults.Workload.TargetClusters.EnvironmentKey)
	labels := map[string]string{tenantLabelKey: tenant}

	// Management cluster ApplicationSet generator.
	m := argocd.ApplicationSetGenerator{
		List: &argocd.ListGenerator{
			Elements: []map[string]string{{"cluster": "management"}},
			Template: argocd.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argocd.ApplicationSetTemplateMeta{
					Name: tenant + "-management",
				},
				Spec: argocd.ApplicationSpec{
					Project: defaults.Project,
					Source: applicationSource(
						defaults.Management.RepoURL,
						defaults.Management.Path,
						defaults.Management.TargetRevision,
						defaults.Management.Helm,
						tenant,
						shortName,
					),
					Destination: argocd.ApplicationDestination{
						Name:      "in-cluster",
						Namespace: defaults.Management.TargetNamespace,
					},
				},
			},
		},
	}

	// Workload cluster ApplicationSet generator.
	w := argocd.ApplicationSetGenerator{
		Clusters: &argocd.ClusterGenerator{
			Selector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      defaults.Workload.TargetClusters.ClusterTypeKey,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{defaults.Workload.TargetClusters.ClusterType},
					},
					{
						Key:      defaults.Workload.TargetClusters.EnvironmentKey,
						Operator: metav1.LabelSelectorOpIn,
						Values:   defaults.Workload.TargetClusters.Environments,
					},
				},
			},
			Template: argocd.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argocd.ApplicationSetTemplateMeta{
					Name:   fmt.Sprintf("%s-workload-%s", tenant, envRef),
					Labels: labels,
				},
				Spec: argocd.ApplicationSpec{
					Project: defaults.Project,
					Source: applicationSource(
						defaults.Workload.RepoURL,
						defaults.Workload.Path,
						defaults.Workload.TargetRevision,
						defaults.Workload.Helm,
						tenant,
						shortName,
					),
					Destination: argocd.ApplicationDestination{
						Name:      "{{ .name }}",
						Namespace: tenant,
					},
				},
			},
		},
	}

	return &argocd.ApplicationSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant,
			Namespace: appSetNamespace,
			Labels:    labels,
		},
		Spec: argocd.ApplicationSetSpec{
			GoTemplate:        true,
			GoTemplateOptions: []string{"missingkey=error"},
			Generators:        []argocd.ApplicationSetGenerator{m, w},
			Template: argocd.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argocd.ApplicationSetTemplateMeta{
					Labels: labels,
				},
				Spec: argocd.ApplicationSpec{
					Project: defaults.Project,
					SyncPolicy: &argocd.SyncPolicy{
						SyncOptions: []string{
							"ServerSideApply=true",
							"ApplyOutOfSyncOnly=true",
							"RespectIgnoreDifferences=true",
						},
					},
				},
			},
			IgnoreApplicationDifferences: []argocd.ApplicationSetResourceIgnoreDifferences{
				{JSONPointers: []string{"/spec/syncPolicy"}},
			},
		},
	}
}

// applicationSource builds an ApplicationSource with tenant identity and Helm configuration.
func applicationSource(
	repoURL string,
	path string,
	targetRevision string,
	helmConfig inputv1beta1.WorkloadHelmConfig,
	tenant string,
	shortName string,
) *argocd.ApplicationSource {
	helm := &argocd.HelmSource{
		ValueFiles: helmConfig.ValueFiles,
		Parameters: []argocd.HelmParameter{
			{Name: "tenant.name", Value: tenant},
			{Name: "tenant.tenantShortName", Value: shortName},
		},
	}
	for _, p := range helmConfig.Parameters {
		helm.Parameters = append(helm.Parameters, argocd.HelmParameter{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return &argocd.ApplicationSource{
		RepoURL:        repoURL,
		Path:           path,
		TargetRevision: targetRevision,
		Helm:           helm,
	}
}
