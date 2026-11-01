package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/functions/xtenant/argocd"
	"github.com/rezakaramad/crosskit/modules/composer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// ArgoCDApplication composes the Application that deploys the tenant management resources to the management cluster.
type ArgoCDApplication struct {
	XComposer
	ObservedResource *argocd.Application
}

// NewManagementApplication creates an Application composer and deserializes any observed Application.
func NewManagementApplication(f XContext) (*ArgoCDApplication, error) {
	resourceName := resource.Name(fmt.Sprintf("xtenant-application-management-%s", f.XR.Name))
	observed, err := composer.ConvertObserved[argocd.Application](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &ArgoCDApplication{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "ArgoCDManagementApplicationReady",
		},
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired Application.
func (r *ArgoCDApplication) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return r.ComposeDesiredResourceFrom(r.createResource())
}

// IsReady returns true when Argo CD reports the Application as Healthy and Synced.
func (r *ArgoCDApplication) IsReady() bool {
	if r.ObservedResource == nil {
		return false
	}
	status := r.ObservedResource.Status
	return status.Health.Status == "Healthy" && status.Sync.Status == "Synced"
}

// createResource constructs the Application for the tenant.
func (r *ArgoCDApplication) createResource() *argocd.Application {
	xr := r.FunctionContext.XR
	defaults := r.FunctionContext.Defaults
	cfg := defaults.Management

	tenant := xr.GetName()
	labels := map[string]string{defaults.TenantLabelKey: tenant}

	cluster := defaults.ManagementCluster
	if cluster == "" {
		cluster = "in-cluster"
	}

	return &argocd.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("%s-management", tenant),
			Namespace:  defaults.Namespace.Argo,
			Labels:     labels,
			Finalizers: []string{"resources-finalizer.argocd.argoproj.io"},
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
				Name:      cluster,
				Namespace: cfg.TargetNamespace,
			},
			SyncPolicy: &argocd.SyncPolicy{
				// Readiness requires Synced, so the Application must converge on its own.
				Automated:   &argocd.SyncPolicyAutomated{SelfHeal: true},
				SyncOptions: syncOptions,
			},
		},
	}
}
