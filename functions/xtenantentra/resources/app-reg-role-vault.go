package resources

import (
	"fmt"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	commonv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/rezakaramad/crosskit/modules/composer"
	applicationsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/applications/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultAppRole composes an Azure AD AppRole on the Vault app registration
// granting access to the tenant.
type VaultAppRole struct {
	XComposer
	ObservedResource *applicationsv1beta1.AppRole
}

// NewVaultAppRole creates a new VaultAppRole composer. It looks up the
// observed AppRole resource and deserializes it for readiness check. Returns an
// error if the observed resource exists but cannot be deserialized.
func NewVaultAppRole(f XContext) (composer.ComposableResource, error) {
	resourceName := resource.Name(fmt.Sprintf("app-role-vault-%s", f.XR.Name))
	observedStructured, err := composer.ConvertObserved[applicationsv1beta1.AppRole](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &VaultAppRole{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "VaultAppRoleReady",
		},
		ObservedResource: observedStructured,
	}, nil
}

// ComposeDesiredResource builds the desired VaultAppRole and wraps it as a
// DesiredResource for inclusion in the function response.
func (a *VaultAppRole) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return a.ComposeDesiredResourceFrom(a.CreateResource())
}

// IsReady returns true if the observed resource has a Ready condition with status True.
func (a *VaultAppRole) IsReady() bool {
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

// CreateResource constructs the Kubernetes AppRole spec for the tenant.
func (a *VaultAppRole) CreateResource() *applicationsv1beta1.AppRole {
	xr := a.FunctionContext.XR
	defaults := a.FunctionContext.Defaults

	applicationID := "/applications/" + defaults.VaultApp.AppRegObjectID
	roleName := xr.GetName()
	description := "Vault Access for " + xr.GetName()
	value := xr.GetName()
	uid := generateRoleUID("vault-" + xr.GetName())

	return &applicationsv1beta1.AppRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-role-vault-" + xr.GetName(),
		},
		Spec: applicationsv1beta1.AppRoleSpec{
			ForProvider: applicationsv1beta1.AppRoleParameters_2{
				AllowedMemberTypes: []*string{
					new("User"),
				},
				ApplicationID: &applicationID,
				DisplayName:   &roleName,
				Description:   &description,
				Value:         &value,
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
