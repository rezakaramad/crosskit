package resources

import (
	"fmt"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	commonv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/rezakaramad/crosskit/modules/composer"
	appv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/app/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoCDSsoRoleAssignment assigns the tenant-specific ArgoCD AppRole on the
// ArgoCD Enterprise Application to the tenant's ArgoCD ACL group.
type ArgoCDSsoRoleAssignment struct {
	XComposer
	ObservedResource *appv1beta1.RoleAssignment
	Group            *groupsv1beta1.Group
}

// NewArgoCDSsoRoleAssignment creates a new ArgoCDSsoRoleAssignment composer.
// It looks up both the observed RoleAssignment and the observed ArgoCD group
// (whose ObjectID is required as the principal) from the function's observed
// resources. Returns an error if either observed resource exists but cannot be
// deserialized.
func NewArgoCDSsoRoleAssignment(f XContext) (composer.ComposableResource, error) {
	resourceName := resource.Name(fmt.Sprintf("roleassignment-argocd-sso-role-%s", f.XR.Name))
	observedStructured, err := composer.ConvertObserved[appv1beta1.RoleAssignment](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	groupName := resource.Name(fmt.Sprintf("group-argocd-%s", f.XR.Name))
	observedGroup, err := composer.ConvertObserved[groupsv1beta1.Group](f.Observed, groupName)
	if err != nil {
		return nil, err
	}

	return &ArgoCDSsoRoleAssignment{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "ArgoCDSsoRoleAssignmentReady",
		},
		ObservedResource: observedStructured,
		Group:            observedGroup,
	}, nil
}

// ComposeDesiredResource builds the desired ArgoCDSsoRoleAssignment and wraps
// it as a DesiredResource for inclusion in the function response. Returns nil if
// the ArgoCD group has not yet been reconciled and surfaced its ObjectID — the
// assignment will be composed on a subsequent reconcile once the group exists
// upstream.
func (s *ArgoCDSsoRoleAssignment) ComposeDesiredResource() (*composer.DesiredResource, error) {
	if s.Group == nil || s.Group.Status.AtProvider.ObjectID == nil {
		return nil, nil
	}
	return s.ComposeDesiredResourceFrom(s.CreateResource())
}

// IsReady returns true if the observed resource has a Ready condition with status True.
func (s *ArgoCDSsoRoleAssignment) IsReady() bool {
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

// CreateResource constructs the Kubernetes RoleAssignment spec for the tenant.
func (s *ArgoCDSsoRoleAssignment) CreateResource() *appv1beta1.RoleAssignment {
	xr := s.FunctionContext.XR
	defaults := s.FunctionContext.Defaults

	appRoleID := generateRoleUID(xr.GetName())

	return &appv1beta1.RoleAssignment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acl-plt-argocd-sso-role-" + xr.GetName(),
		},
		Spec: appv1beta1.RoleAssignmentSpec{
			ForProvider: appv1beta1.RoleAssignmentParameters{
				AppRoleID:         &appRoleID,
				ResourceObjectID:  &defaults.ArgoCDApp.EnterpriseAppObjectID,
				PrincipalObjectID: s.Group.Status.AtProvider.ObjectID,
			},
			ManagedResourceSpec: commonv2.ManagedResourceSpec{
				ProviderConfigReference: &commonv1.ProviderConfigReference{
					Name: defaults.ProviderConfigRef.Name,
					Kind: defaults.ProviderConfigRef.Kind,
				},
			},
		},
	}
}
