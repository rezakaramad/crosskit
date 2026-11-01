package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/functions/xtenant/argocd"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenant/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// ArgoCDApplicationSet composes an ApplicationSet that deploys one tenant workload resources to the workload clusters (Development/Test/Production).
type ArgoCDApplicationSet struct {
	XComposer
	role             string // workload
	config           inputv1beta1.ApplicationSetConfig
	ObservedResource *argocd.ApplicationSet
}

// NewWorkloadApplicationSet creates the composer for the ApplicationSet that deploys the tenant workload resources.
func NewWorkloadApplicationSet(f XContext) (*ArgoCDApplicationSet, error) {
	return newApplicationSet(f, "workload", "ArgoCDWorkloadApplicationSetReady", f.Defaults.Workload)
}

// newApplicationSet creates an ApplicationSet composer and deserializes any observed ApplicationSet.
func newApplicationSet(
	f XContext,
	role string,
	conditionType string,
	config inputv1beta1.ApplicationSetConfig,
) (*ArgoCDApplicationSet, error) {
	resourceName := resource.Name(fmt.Sprintf("xtenant-applicationset-%s-%s", role, f.XR.Name))
	observedStructured, err := composer.ConvertObserved[argocd.ApplicationSet](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &ArgoCDApplicationSet{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   conditionType,
		},
		role:             role,
		config:           config,
		ObservedResource: observedStructured,
	}, nil
}

// ComposeDesiredResource builds the desired ApplicationSet.
func (r *ArgoCDApplicationSet) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return r.ComposeDesiredResourceFrom(r.createResource())
}

// IsReady returns true when the ApplicationSet controller reports ResourcesUpToDate.
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

// createResource constructs the ApplicationSet with a clusters generator
// that creates one workload Application per matching cluster.
func (r *ArgoCDApplicationSet) createResource() *argocd.ApplicationSet {
	xr := r.FunctionContext.XR
	defaults := r.FunctionContext.Defaults
	cfg := r.config

	tenant := xr.GetName()
	envRef := labelRef(cfg.TargetClusters.EnvironmentKey)
	labels := map[string]string{defaults.TenantLabelKey: tenant}

	clusters := argocd.ApplicationSetGenerator{
		Clusters: &argocd.ClusterGenerator{
			Selector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      cfg.TargetClusters.ClusterTypeKey,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{cfg.TargetClusters.ClusterType},
					},
					{
						Key:      cfg.TargetClusters.EnvironmentKey,
						Operator: metav1.LabelSelectorOpIn,
						Values:   cfg.TargetClusters.Environments,
					},
				},
			},
			Template: argocd.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argocd.ApplicationSetTemplateMeta{
					Name:   fmt.Sprintf("%s-%s-%s", tenant, r.role, envRef),
					Labels: labels,
				},
				Spec: argocd.ApplicationSpec{
					Project: defaults.Project,
					Source: applicationSource(
						cfg.RepoURL,
						cfg.Path,
						cfg.TargetRevision,
						cfg.Helm,
						tenant,
						xr.Spec.ShortName,
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
			Name:      fmt.Sprintf("%s-%s", tenant, r.role),
			Namespace: defaults.Namespace.Argo,
			Labels:    labels,
		},
		Spec: argocd.ApplicationSetSpec{
			GoTemplate:        true,
			GoTemplateOptions: []string{"missingkey=error"},
			Generators:        []argocd.ApplicationSetGenerator{clusters},
			Template: argocd.ApplicationSetTemplate{
				ApplicationSetTemplateMeta: argocd.ApplicationSetTemplateMeta{
					Labels: labels,
				},
				Spec: argocd.ApplicationSpec{
					Project:    defaults.Project,
					SyncPolicy: &argocd.SyncPolicy{SyncOptions: syncOptions},
				},
			},
			IgnoreApplicationDifferences: []argocd.ApplicationSetResourceIgnoreDifferences{
				{JSONPointers: []string{"/spec/syncPolicy"}},
			},
		},
	}
}
