package resources

import (
	"testing"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	appv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/app/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
)

func TestRoleAssignment_createResource(t *testing.T) {
	cases := map[string]struct {
		app             inputv1beta1.AppConfig
		wantName        string
		wantEnterprise  string
		wantResource    string
		wantCondition   string
		groupObjectID   string
		groupObservedAs string
	}{
		"argocd": {
			app:             argocdApp,
			wantName:        "acl-plt-argocd-sso-role-pillow-factory",
			wantEnterprise:  "argocd-entapp-uuid",
			wantResource:    "roleassignment-argocd-sso-role-pillow-factory",
			wantCondition:   "ArgoCDSsoRoleAssignmentReady",
			groupObjectID:   "argocd-group-uuid",
			groupObservedAs: "group-argocd-pillow-factory",
		},
		"vault": {
			app:             vaultApp,
			wantName:        "acl-plt-vault-sso-role-pillow-factory",
			wantEnterprise:  "vault-entapp-uuid",
			wantResource:    "roleassignment-vault-sso-role-pillow-factory",
			wantCondition:   "VaultSsoRoleAssignmentReady",
			groupObjectID:   "vault-group-uuid",
			groupObservedAs: "group-vault-pillow-factory",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := NewRoleAssignment(testXContext(), tc.app)
			if err != nil {
				t.Fatal(err)
			}
			r.Group = &groupsv1beta1.Group{
				Status: groupsv1beta1.GroupStatus{AtProvider: groupsv1beta1.GroupObservation{ObjectID: &tc.groupObjectID}},
			}

			got := r.createResource()

			if string(r.ResourceName) != tc.wantResource {
				t.Errorf("resource name = %q, want %q", r.ResourceName, tc.wantResource)
			}
			if r.GetConditionType() != tc.wantCondition {
				t.Errorf("condition type = %q, want %q", r.GetConditionType(), tc.wantCondition)
			}
			if string(groupResourceName(tc.app, "pillow-factory")) != tc.groupObservedAs {
				t.Errorf("group lookup name = %q, want %q", groupResourceName(tc.app, "pillow-factory"), tc.groupObservedAs)
			}
			if got.Name != tc.wantName {
				t.Errorf("metadata.name = %q, want %q", got.Name, tc.wantName)
			}
			if got.Spec.ForProvider.ResourceObjectID == nil || *got.Spec.ForProvider.ResourceObjectID != tc.wantEnterprise {
				t.Errorf("ResourceObjectID = %v, want %q", got.Spec.ForProvider.ResourceObjectID, tc.wantEnterprise)
			}
			if got.Spec.ForProvider.PrincipalObjectID == nil || *got.Spec.ForProvider.PrincipalObjectID != tc.groupObjectID {
				t.Errorf("PrincipalObjectID = %v, want %q (observed group ObjectID)", got.Spec.ForProvider.PrincipalObjectID, tc.groupObjectID)
			}
			// The assignment must use the same UID the AppRole sets as its RoleID.
			wantUID := roleUID(tc.app, "pillow-factory")
			if got.Spec.ForProvider.AppRoleID == nil || *got.Spec.ForProvider.AppRoleID != wantUID {
				t.Errorf("AppRoleID = %v, want %q", got.Spec.ForProvider.AppRoleID, wantUID)
			}
		})
	}
}

// When the group hasn't surfaced its ObjectID yet, the composer must skip composition
// instead of building an assignment with a dangling principal.
func TestRoleAssignment_ComposeDesired_NoObservedGroup(t *testing.T) {
	r, err := NewRoleAssignment(testXContext(), argocdApp)
	if err != nil {
		t.Fatal(err)
	}

	got, err := r.ComposeDesiredResource()
	if err != nil {
		t.Fatalf("ComposeDesiredResource error: %v", err)
	}
	if got != nil {
		t.Errorf("ComposeDesiredResource = %v, want nil when observed group is missing", got)
	}
}

func TestRoleAssignment_IsReady(t *testing.T) {
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
			r := &RoleAssignment{ObservedResource: observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
