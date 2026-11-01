// Package resources implements the composers that the Crossplane function uses
// to build and reconcile the child resources.
package resources

import (
	"fmt"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	commonv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// Group composes the tenant's Entra ID ACL group for one app.
type Group struct {
	XComposer
	app              inputv1beta1.AppConfig
	ObservedResource *groupsv1beta1.Group
}

// groupResourceName is the internal name of an app's group. The RoleAssignment looks the group up by it.
func groupResourceName(app inputv1beta1.AppConfig, tenant string) resource.Name {
	return resource.Name(fmt.Sprintf("group-%s-%s", app.Name, tenant))
}

// NewGroup creates a Group composer and deserializes any observed Group.
func NewGroup(f XContext, app inputv1beta1.AppConfig) (*Group, error) {
	resourceName := groupResourceName(app, f.XR.Name)
	observed, err := composer.ConvertObserved[groupsv1beta1.Group](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &Group{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   app.DisplayName + "GroupReady",
		},
		app:              app,
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired Group.
func (g *Group) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return g.ComposeDesiredResourceFrom(g.createResource())
}

// IsReady returns true if the observed resource has a Ready condition with status True.
func (g *Group) IsReady() bool {
	if g.ObservedResource == nil {
		return false
	}
	for _, c := range g.ObservedResource.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			return true
		}
	}
	return false
}

// createResource constructs the Group for the tenant.
func (g *Group) createResource() *groupsv1beta1.Group {
	tenant := g.FunctionContext.XR.GetName()
	defaults := g.FunctionContext.Defaults

	displayName := fmt.Sprintf("ACL.PLT.%s.Tenant.%s", g.app.DisplayName, pascal(tenant))

	return &groupsv1beta1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("acl-plt-%s-tenant-%s", g.app.Name, tenant),
		},
		Spec: groupsv1beta1.GroupSpec{
			ForProvider: groupsv1beta1.GroupParameters{
				DisplayName:           &displayName,
				Description:           new("This application is managed by the Platform team."),
				PreventDuplicateNames: new(true),
				SecurityEnabled:       new(true),
			},
			ManagedResourceSpec: commonv2.ManagedResourceSpec{
				ManagementPolicies: commonv1.ManagementPolicies{"Create", "Observe", "Update", "Delete"},
				ProviderConfigReference: &commonv1.ProviderConfigReference{
					Name: defaults.ProviderConfigRef.Name,
					Kind: defaults.ProviderConfigRef.Kind,
				},
			},
		},
	}
}
