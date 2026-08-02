package resources

import (
	"testing"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantentra/input/v1beta1"
	xtenantentra "github.com/rezakaramad/crosskit/types/xtenantentra"
	appv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/app/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVaultSsoRoleAssignment_CreateResource(t *testing.T) {
	xr := &xtenantentra.XTenantEntra{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantentra.XTenantEntraSpec{},
	}
	d := &inputv1beta1.Input{
		VaultApp: inputv1beta1.VaultAppConfig{
			AppRegObjectID:        "vault-appreg-uuid",
			EnterpriseAppObjectID: "vault-entapp-uuid",
		},
		ProviderConfigRef: inputv1beta1.ProviderConfigRef{Name: "azuread-pc", Kind: "ProviderConfig"},
	}
	groupID := "vault-group-uuid"
	r := &VaultSsoRoleAssignment{
		XComposer: XComposer{FunctionContext: XContext{XR: xr, Defaults: d}},
		Group: &groupsv1beta1.Group{
			Status: groupsv1beta1.GroupStatus{AtProvider: groupsv1beta1.GroupObservation{ObjectID: &groupID}},
		},
	}

	got := r.CreateResource()

	if got.Name != "acl-plt-vault-sso-role-pillow-factory" {
		t.Errorf("metadata.name = %q, want %q", got.Name, "acl-plt-vault-sso-role-pillow-factory")
	}
	if got.Spec.ForProvider.ResourceObjectID == nil || *got.Spec.ForProvider.ResourceObjectID != "vault-entapp-uuid" {
		t.Errorf("ResourceObjectID = %v, want %q", got.Spec.ForProvider.ResourceObjectID, "vault-entapp-uuid")
	}
	if got.Spec.ForProvider.PrincipalObjectID == nil || *got.Spec.ForProvider.PrincipalObjectID != groupID {
		t.Errorf("PrincipalObjectID = %v, want %q (observed group ObjectID)", got.Spec.ForProvider.PrincipalObjectID, groupID)
	}
	// Contract: the assignment's AppRoleID must equal the same UID that
	// app-reg-role-vault uses for its RoleID (vault- prefixed seed).
	wantUID := generateRoleUID("vault-pillow-factory")
	if got.Spec.ForProvider.AppRoleID == nil || *got.Spec.ForProvider.AppRoleID != wantUID {
		t.Errorf("AppRoleID = %v, want %q", got.Spec.ForProvider.AppRoleID, wantUID)
	}
}

// TestVaultSsoRoleAssignment_ComposeDesired_NoObservedGroup verifies that
// when the Vault group hasn't yet surfaced its ObjectID, the composer skips
// composition (returns nil, nil) instead of building an assignment with a
// dangling principal.
func TestVaultSsoRoleAssignment_ComposeDesired_NoObservedGroup(t *testing.T) {
	xr := &xtenantentra.XTenantEntra{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantentra.XTenantEntraSpec{},
	}
	d := &inputv1beta1.Input{}
	r := &VaultSsoRoleAssignment{
		XComposer: XComposer{FunctionContext: XContext{XR: xr, Defaults: d}},
		Group:     nil,
	}

	got, err := r.ComposeDesiredResource()
	if err != nil {
		t.Fatalf("ComposeDesiredResource error: %v", err)
	}
	if got != nil {
		t.Errorf("ComposeDesiredResource = %v, want nil when observed group is missing", got)
	}
}

func TestVaultSsoRoleAssignment_IsReady(t *testing.T) {
	for _, tt := range readyCases() {
		t.Run(tt.name, func(t *testing.T) {
			var observed *appv1beta1.RoleAssignment
			if tt.conditions != nil {
				observed = &appv1beta1.RoleAssignment{
					Status: appv1beta1.RoleAssignmentStatus{
						ResourceStatus: commonv1.ResourceStatus{ConditionedStatus: commonv1.ConditionedStatus{Conditions: tt.conditions}},
					},
				}
			}
			r := &VaultSsoRoleAssignment{ObservedResource: observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
