package resources

import (
	"testing"

	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantentra/input/v1beta1"
	xtenantentra "github.com/rezakaramad/crosskit/types/xtenantentra"
	applicationsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/applications/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestArgoCDAppRole_CreateResource(t *testing.T) {
	xr := &xtenantentra.XTenantEntra{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantentra.XTenantEntraSpec{},
	}
	d := &inputv1beta1.Input{
		ArgoCDApp:         inputv1beta1.ArgoCDAppConfig{AppRegObjectID: "argocd-appreg-uuid"},
		ProviderConfigRef: inputv1beta1.ProviderConfigRef{Name: "azuread-pc", Kind: "ProviderConfig"},
	}
	r := &ArgoCDAppRole{XComposer: XComposer{FunctionContext: XContext{XR: xr, Defaults: d}}}

	got := r.CreateResource()

	if got.Name != "app-role-argocd-pillow-factory" {
		t.Errorf("metadata.name = %q, want %q", got.Name, "app-role-argocd-pillow-factory")
	}
	if got.Spec.ForProvider.ApplicationID == nil || *got.Spec.ForProvider.ApplicationID != "/applications/argocd-appreg-uuid" {
		t.Errorf("ApplicationID = %v, want %q", got.Spec.ForProvider.ApplicationID, "/applications/argocd-appreg-uuid")
	}
	if got.Spec.ForProvider.Value == nil || *got.Spec.ForProvider.Value != "pillow-factory" {
		t.Errorf("Value = %v, want %q (XR name)", got.Spec.ForProvider.Value, "pillow-factory")
	}
	// Contract: the AppRole's RoleID must equal generateRoleUID(xr.GetName())
	// because the ArgoCD SSO RoleAssignment looks up the assignment target by
	// this same UID.
	wantUID := generateRoleUID("pillow-factory")
	if got.Spec.ForProvider.RoleID == nil || *got.Spec.ForProvider.RoleID != wantUID {
		t.Errorf("RoleID = %v, want %q", got.Spec.ForProvider.RoleID, wantUID)
	}
}

func TestArgoCDAppRole_IsReady(t *testing.T) {
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
			r := &ArgoCDAppRole{ObservedResource: observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
