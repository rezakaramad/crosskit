package resources

import (
	commonv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	xtenantssov1beta1 "github.com/rezakaramad/crosskit/types/xtenantsso/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	argocdApp = inputv1beta1.AppConfig{
		Name:                  "argocd",
		DisplayName:           "ArgoCD",
		AppRegObjectID:        "argocd-appreg-uuid",
		EnterpriseAppObjectID: "argocd-entapp-uuid",
	}
	vaultApp = inputv1beta1.AppConfig{
		Name:                  "vault",
		DisplayName:           "Vault",
		AppRegObjectID:        "vault-appreg-uuid",
		EnterpriseAppObjectID: "vault-entapp-uuid",
	}
)

func testXContext() XContext {
	xr := &xtenantssov1beta1.XTenantSSO{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantssov1beta1.XTenantSSOSpec{},
	}
	d := &inputv1beta1.Input{
		Apps:              []inputv1beta1.AppConfig{argocdApp, vaultApp},
		ProviderConfigRef: inputv1beta1.ProviderConfigRef{Name: "azuread-pc", Kind: "ProviderConfig"},
	}
	return XContext{XR: xr, Defaults: d}
}

// readyCases is shared by every composer's IsReady test, since their readiness logic is
// identical: nil-observed is not ready; observed is ready when it has a Ready=True condition.
type readyCase struct {
	name       string
	conditions []commonv1.Condition
	want       bool
}

func readyCases() []readyCase {
	return []readyCase{
		{name: "nil observed", conditions: nil, want: false},
		{name: "no conditions", conditions: []commonv1.Condition{}, want: false},
		{name: "Ready=False", conditions: []commonv1.Condition{{Type: "Ready", Status: corev1.ConditionFalse}}, want: false},
		{name: "Ready=True", conditions: []commonv1.Condition{{Type: "Ready", Status: corev1.ConditionTrue}}, want: true},
		{name: "unrelated true", conditions: []commonv1.Condition{{Type: "Synced", Status: corev1.ConditionTrue}}, want: false},
	}
}
