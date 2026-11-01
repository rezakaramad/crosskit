package resources

import (
	"fmt"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	commonv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	appv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/app/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// RoleAssignment assigns the tenant's role on an enterprise application to the tenant's group.
type RoleAssignment struct {
	XComposer
	app              inputv1beta1.AppConfig
	ObservedResource *appv1beta1.RoleAssignment
	Group            *groupsv1beta1.Group
}

// NewRoleAssignment creates a RoleAssignment composer. It also looks up the observed
// Group, whose object ID is the principal of the assignment.
func NewRoleAssignment(f XContext, app inputv1beta1.AppConfig) (*RoleAssignment, error) {
	resourceName := resource.Name(fmt.Sprintf("roleassignment-%s-sso-role-%s", app.Name, f.XR.Name))
	observed, err := composer.ConvertObserved[appv1beta1.RoleAssignment](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	group, err := composer.ConvertObserved[groupsv1beta1.Group](f.Observed, groupResourceName(app, f.XR.Name))
	if err != nil {
		return nil, err
	}

	return &RoleAssignment{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   app.DisplayName + "SsoRoleAssignmentReady",
		},
		app:              app,
		ObservedResource: observed,
		Group:            group,
	}, nil
}

// ComposeDesiredResource builds the desired RoleAssignment. It returns nil until the
// Group has surfaced its object ID; a later reconcile composes the assignment.
func (s *RoleAssignment) ComposeDesiredResource() (*composer.DesiredResource, error) {
	if s.Group == nil || s.Group.Status.AtProvider.ObjectID == nil {
		return nil, nil
	}
	return s.ComposeDesiredResourceFrom(s.createResource())
}

// IsReady returns true if the observed resource has a Ready condition with status True.
func (s *RoleAssignment) IsReady() bool {
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

// createResource constructs the RoleAssignment for the tenant.
func (s *RoleAssignment) createResource() *appv1beta1.RoleAssignment {
	tenant := s.FunctionContext.XR.GetName()
	defaults := s.FunctionContext.Defaults

	appRoleID := roleUID(s.app, tenant)

	return &appv1beta1.RoleAssignment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("acl-plt-%s-sso-role-%s", s.app.Name, tenant),
		},
		Spec: appv1beta1.RoleAssignmentSpec{
			ForProvider: appv1beta1.RoleAssignmentParameters{
				AppRoleID:         &appRoleID,
				ResourceObjectID:  &s.app.EnterpriseAppObjectID,
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
