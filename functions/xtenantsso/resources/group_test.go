package resources

import (
	"testing"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
)

func TestGroup_createResource(t *testing.T) {
	cases := map[string]struct {
		app             inputv1beta1.AppConfig
		wantName        string
		wantDisplayName string
		wantResource    string
		wantCondition   string
	}{
		"argocd": {
			app:             argocdApp,
			wantName:        "acl-plt-argocd-tenant-pillow-factory",
			wantDisplayName: "ACL.PLT.ArgoCD.Tenant.PillowFactory",
			wantResource:    "group-argocd-pillow-factory",
			wantCondition:   "ArgoCDGroupReady",
		},
		"vault": {
			app:             vaultApp,
			wantName:        "acl-plt-vault-tenant-pillow-factory",
			wantDisplayName: "ACL.PLT.Vault.Tenant.PillowFactory",
			wantResource:    "group-vault-pillow-factory",
			wantCondition:   "VaultGroupReady",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := NewGroup(testXContext(), tc.app)
			if err != nil {
				t.Fatal(err)
			}
			got := r.createResource()

			if string(r.ResourceName) != tc.wantResource {
				t.Errorf("resource name = %q, want %q", r.ResourceName, tc.wantResource)
			}
			if r.GetConditionType() != tc.wantCondition {
				t.Errorf("condition type = %q, want %q", r.GetConditionType(), tc.wantCondition)
			}
			if got.Name != tc.wantName {
				t.Errorf("metadata.name = %q, want %q", got.Name, tc.wantName)
			}
			if got.Spec.ForProvider.DisplayName == nil || *got.Spec.ForProvider.DisplayName != tc.wantDisplayName {
				t.Errorf("DisplayName = %v, want %q", got.Spec.ForProvider.DisplayName, tc.wantDisplayName)
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
		})
	}
}

func TestGroup_IsReady(t *testing.T) {
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
			r := &Group{ObservedResource: observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
