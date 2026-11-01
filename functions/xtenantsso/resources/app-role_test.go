package resources

import (
	"testing"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	applicationsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/applications/v1beta1"
)

func TestAppRole_createResource(t *testing.T) {
	cases := map[string]struct {
		app             inputv1beta1.AppConfig
		wantName        string
		wantApplication string
		wantDescription string
	}{
		"argocd": {
			app:             argocdApp,
			wantName:        "app-role-argocd-pillow-factory",
			wantApplication: "/applications/argocd-appreg-uuid",
			wantDescription: "ArgoCD Access for pillow-factory",
		},
		"vault": {
			app:             vaultApp,
			wantName:        "app-role-vault-pillow-factory",
			wantApplication: "/applications/vault-appreg-uuid",
			wantDescription: "Vault Access for pillow-factory",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := NewAppRole(testXContext(), tc.app)
			if err != nil {
				t.Fatal(err)
			}
			got := r.createResource()

			if got.Name != tc.wantName {
				t.Errorf("metadata.name = %q, want %q", got.Name, tc.wantName)
			}
			if got.Spec.ForProvider.ApplicationID == nil || *got.Spec.ForProvider.ApplicationID != tc.wantApplication {
				t.Errorf("ApplicationID = %v, want %q", got.Spec.ForProvider.ApplicationID, tc.wantApplication)
			}
			if got.Spec.ForProvider.Description == nil || *got.Spec.ForProvider.Description != tc.wantDescription {
				t.Errorf("Description = %v, want %q", got.Spec.ForProvider.Description, tc.wantDescription)
			}
			if got.Spec.ForProvider.Value == nil || *got.Spec.ForProvider.Value != "pillow-factory" {
				t.Errorf("Value = %v, want %q (XR name)", got.Spec.ForProvider.Value, "pillow-factory")
			}
			// The RoleAssignment looks the role up by this same UID.
			wantUID := roleUID(tc.app, "pillow-factory")
			if got.Spec.ForProvider.RoleID == nil || *got.Spec.ForProvider.RoleID != wantUID {
				t.Errorf("RoleID = %v, want %q", got.Spec.ForProvider.RoleID, wantUID)
			}
		})
	}
}

func TestAppRole_IsReady(t *testing.T) {
	for _, tt := range readyCases() {
		t.Run(tt.name, func(t *testing.T) {
			var observed *applicationsv1beta1.AppRole
			if tt.conditions != nil {
				observed = &applicationsv1beta1.AppRole{
					Status: applicationsv1beta1.AppRoleStatus{
						ResourceStatus: commonv1.ResourceStatus{ConditionedStatus: commonv1.ConditionedStatus{Conditions: tt.conditions}},
					},
				}
			}
			r := &AppRole{ObservedResource: observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
