package resources

import (
	"fmt"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	commonv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	applicationsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/applications/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// AppRole composes the tenant's role on one app registration.
type AppRole struct {
	XComposer
	app              inputv1beta1.AppConfig
	ObservedResource *applicationsv1beta1.AppRole
}

// NewAppRole creates an AppRole composer and deserializes any observed AppRole.
func NewAppRole(f XContext, app inputv1beta1.AppConfig) (*AppRole, error) {
	resourceName := resource.Name(fmt.Sprintf("app-role-%s-%s", app.Name, f.XR.Name))
	observed, err := composer.ConvertObserved[applicationsv1beta1.AppRole](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &AppRole{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   app.DisplayName + "AppRoleReady",
		},
		app:              app,
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired AppRole.
func (a *AppRole) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return a.ComposeDesiredResourceFrom(a.createResource())
}

// IsReady returns true if the observed resource has a Ready condition with status True.
func (a *AppRole) IsReady() bool {
	if a.ObservedResource == nil {
		return false
	}
	for _, c := range a.ObservedResource.Status.Conditions {
		if c.Type == "Ready" && c.Status == "True" {
			return true
		}
	}
	return false
}

// createResource constructs the AppRole for the tenant.
func (a *AppRole) createResource() *applicationsv1beta1.AppRole {
	tenant := a.FunctionContext.XR.GetName()
	defaults := a.FunctionContext.Defaults

	applicationID := "/applications/" + a.app.AppRegObjectID
	description := fmt.Sprintf("%s Access for %s", a.app.DisplayName, tenant)
	uid := roleUID(a.app, tenant)

	return &applicationsv1beta1.AppRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("app-role-%s-%s", a.app.Name, tenant),
		},
		Spec: applicationsv1beta1.AppRoleSpec{
			ForProvider: applicationsv1beta1.AppRoleParameters_2{
				AllowedMemberTypes: []*string{
					new("User"),
				},
				ApplicationID: &applicationID,
				DisplayName:   &tenant,
				Description:   &description,
				Value:         &tenant,
				RoleID:        &uid,
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
