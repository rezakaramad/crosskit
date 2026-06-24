package resources

import (
	"fmt"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	commonv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/rezakaramad/crosskit/modules/composer"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoCDGroup composes a Kubernetes Entra ID Group for the tenant.
type ArgoCDGroup struct {
	XComposer
	ObservedResource *groupsv1beta1.Group
}

// NewArgoCDGroup creates a new ArgoCDGroup composer. It looks up the observed ArgoCDGroup
// resource and deserializes it for readiness check. Returns an error if the
// observed resource exists but cannot be deserialized.
func NewArgoCDGroup(f XContext) (composer.ComposableResource, error) {
	resourceName := resource.Name(fmt.Sprintf("group-argocd-%s", f.XR.Name))
	observedStructured, err := composer.ConvertObserved[groupsv1beta1.Group](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &ArgoCDGroup{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "ArgoCDGroupReady",
		},
		ObservedResource: observedStructured,
	}, nil
}

// ComposeDesiredResource builds the desired ArgoCDGroup and wraps it as a
// DesiredResource for inclusion in the function response.
func (s *ArgoCDGroup) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return s.ComposeDesiredResourceFrom(s.CreateResource())
}

// IsReady returns true if the observed resource has a Ready condition with status True.
func (s *ArgoCDGroup) IsReady() bool {
	if s.ObservedResource == nil {
		return false
	}
	for _, c := range s.ObservedResource.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			return true
		}
	}
	return false
}

// CreateResource constructs the Kubernetes Group spec for the tenant.
func (s *ArgoCDGroup) CreateResource() *groupsv1beta1.Group {
	xr := s.FunctionContext.XR
	defaults := s.FunctionContext.Defaults

	displayName := "ACL.PLT.ArgoCD.Tenant." + pascal(xr.GetName())

	group := &groupsv1beta1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acl-plt-argocd-tenant-" + xr.GetName(),
		},
		Spec: groupsv1beta1.GroupSpec{
			ForProvider: groupsv1beta1.GroupParameters{
				DisplayName:           &displayName,
				Description:           new("This application is managed by the Platform team."),
				PreventDuplicateNames: new(true),
				SecurityEnabled:       new(true),
			},
			ManagedResourceSpec: commonv2.ManagedResourceSpec{
				ManagementPolicies: commonv1.ManagementPolicies{"Create", "Observe", "Update"},
				ProviderConfigReference: &commonv1.ProviderConfigReference{
					Name: defaults.ProviderConfigRef.Name,
					Kind: defaults.ProviderConfigRef.Kind,
				},
			},
		},
	}

	return group
}
