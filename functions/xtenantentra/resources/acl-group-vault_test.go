package resources

import (
	"testing"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantentra/input/v1beta1"
	xtenantentra "github.com/rezakaramad/crosskit/types/xtenantentra"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVaultGroup_CreateResource(t *testing.T) {
	xr := &xtenantentra.XTenantEntra{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantentra.XTenantEntraSpec{},
	}
	d := &inputv1beta1.Input{
		VaultApp:          inputv1beta1.VaultAppConfig{AppRegObjectID: "vault-appreg-uuid"},
		ProviderConfigRef: inputv1beta1.ProviderConfigRef{Name: "azuread-pc", Kind: "ProviderConfig"},
	}
	r := &VaultGroup{XComposer: XComposer{FunctionContext: XContext{XR: xr, Defaults: d}}}

	got := r.CreateResource()

	if got.Name != "acl-plt-vault-tenant-pillow-factory" {
		t.Errorf("metadata.name = %q, want %q", got.Name, "acl-plt-vault-tenant-pillow-factory")
	}
	if got.Spec.ForProvider.DisplayName == nil || *got.Spec.ForProvider.DisplayName != "ACL.PLT.Vault.Tenant.PillowFactory" {
		t.Errorf("DisplayName = %v, want %q", got.Spec.ForProvider.DisplayName, "ACL.PLT.Vault.Tenant.PillowFactory")
	}
	if got.Spec.ForProvider.SecurityEnabled == nil || !*got.Spec.ForProvider.SecurityEnabled {
		t.Errorf("SecurityEnabled = %v, want true", got.Spec.ForProvider.SecurityEnabled)
	}
	if got.Spec.ProviderConfigReference == nil || got.Spec.ProviderConfigReference.Name != "azuread-pc" {
		t.Errorf("ProviderConfigReference = %v, want name=azuread-pc", got.Spec.ProviderConfigReference)
	}
	wantPolicies := commonv1.ManagementPolicies{"Create", "Observe", "Update", "Delete"}
	if len(got.Spec.ManagementPolicies) != len(wantPolicies) {
		t.Errorf("ManagementPolicies = %v, want %v", got.Spec.ManagementPolicies, wantPolicies)
	}
}

func TestVaultGroup_IsReady(t *testing.T) {
	for _, tt := range readyCases() {
		t.Run(tt.name, func(t *testing.T) {
			var observed *groupsv1beta1.Group
			if tt.conditions != nil {
				observed = &groupsv1beta1.Group{
					Status: groupsv1beta1.GroupStatus{
						ResourceStatus: commonv1.ResourceStatus{ConditionedStatus: commonv1.ConditionedStatus{Conditions: tt.conditions}},
					},
				}
			}
			r := &VaultGroup{ObservedResource: observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
