package resources

import (
	"testing"

	"github.com/rezakaramad/crosskit/functions/xtenant/argocd"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenant/input/v1beta1"
	xtenantv1beta1 "github.com/rezakaramad/crosskit/types/xtenant/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testTenantLabelKey = "platform.rezakara.demo/tenant"

func testInput() *inputv1beta1.Input {
	return &inputv1beta1.Input{
		Project:        "tenant-system",
		TenantLabelKey: testTenantLabelKey,
		Management: inputv1beta1.ApplicationConfig{
			RepoURL:         "https://github.com/rezakaramad/kubepave",
			Path:            "kubernetes/charts/tenant-management",
			TargetRevision:  "HEAD",
			TargetNamespace: "tenant-system",
			Helm: inputv1beta1.HelmConfig{
				Parameters: []inputv1beta1.HelmParameter{{Name: "vault.enabled", Value: "true"}},
			},
		},
		Workload: inputv1beta1.ApplicationSetConfig{
			RepoURL:        "https://github.com/rezakaramad/kubepave",
			Path:           "kubernetes/charts/tenant-workload",
			TargetRevision: "HEAD",
			Helm: inputv1beta1.HelmConfig{
				ValueFiles: []string{`values-{{ index .metadata.labels "argocd.rezakara.demo/environment" }}.yaml`},
			},
			TargetClusters: inputv1beta1.TargetClustersConfig{
				ClusterTypeKey: "argocd.rezakara.demo/cluster-type",
				ClusterType:    "tenant",
				EnvironmentKey: "argocd.rezakara.demo/environment",
				Environments:   []string{"development"},
			},
		},
		Namespace: inputv1beta1.NamespaceConfig{Argo: "argocd"},
	}
}

func testXContext() XContext {
	xr := &xtenantv1beta1.XTenant{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantv1beta1.XTenantSpec{ShortName: "pil"},
	}
	return XContext{XR: xr, Defaults: testInput()}
}

func TestArgoCDApplication_CreateResource(t *testing.T) {
	cases := map[string]struct {
		build    func(XContext) (*ArgoCDApplication, error)
		wantName string
		wantPath string
		wantHelm []argocd.HelmParameter
	}{
		"management": {
			build:    NewManagementApplication,
			wantName: "pillow-factory-management",
			wantPath: "kubernetes/charts/tenant-management",
			wantHelm: []argocd.HelmParameter{
				{Name: "tenant.name", Value: "pillow-factory"},
				{Name: "tenant.tenantShortName", Value: "pil"},
				{Name: "vault.enabled", Value: "true"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, err := tc.build(testXContext())
			if err != nil {
				t.Fatal(err)
			}
			got := r.createResource()

			if got.Name != tc.wantName {
				t.Errorf("metadata.name = %q, want %q", got.Name, tc.wantName)
			}
			if got.Namespace != "argocd" {
				t.Errorf("metadata.namespace = %q, want argocd", got.Namespace)
			}
			if got.Labels[testTenantLabelKey] != "pillow-factory" {
				t.Errorf("label %q = %q, want pillow-factory", testTenantLabelKey, got.Labels[testTenantLabelKey])
			}
			if len(got.Finalizers) != 1 || got.Finalizers[0] != "resources-finalizer.argocd.argoproj.io" {
				t.Errorf("finalizers = %v, want [resources-finalizer.argocd.argoproj.io]", got.Finalizers)
			}
			if got.Spec.Project != "tenant-system" {
				t.Errorf("spec.project = %q, want tenant-system", got.Spec.Project)
			}
			if got.Spec.Source == nil || got.Spec.Source.Path != tc.wantPath {
				t.Fatalf("spec.source = %+v, want path %q", got.Spec.Source, tc.wantPath)
			}
			if len(got.Spec.Source.Helm.Parameters) != len(tc.wantHelm) {
				t.Fatalf("helm.parameters = %+v, want %+v", got.Spec.Source.Helm.Parameters, tc.wantHelm)
			}
			for i, p := range tc.wantHelm {
				if got.Spec.Source.Helm.Parameters[i] != p {
					t.Errorf("helm.parameters[%d] = %+v, want %+v", i, got.Spec.Source.Helm.Parameters[i], p)
				}
			}
			if got.Spec.Destination.Name != "in-cluster" || got.Spec.Destination.Namespace != "tenant-system" {
				t.Errorf("spec.destination = %+v, want in-cluster/tenant-system", got.Spec.Destination)
			}
			if got.Spec.SyncPolicy == nil || got.Spec.SyncPolicy.Automated == nil || !got.Spec.SyncPolicy.Automated.SelfHeal {
				t.Errorf("spec.syncPolicy = %+v, want automated selfHeal", got.Spec.SyncPolicy)
			}
		})
	}
}

func TestArgoCDApplication_ManagementCluster(t *testing.T) {
	cases := map[string]struct {
		configured string
		want       string
	}{
		"FallsBackToInCluster": {configured: "", want: "in-cluster"},
		"UsesConfiguredName":   {configured: "mgmt", want: "mgmt"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := testXContext()
			f.Defaults.ManagementCluster = tc.configured
			r, err := NewManagementApplication(f)
			if err != nil {
				t.Fatal(err)
			}
			if got := r.createResource().Spec.Destination.Name; got != tc.want {
				t.Errorf("destination.name = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestArgoCDApplication_IsReady(t *testing.T) {
	app := func(health, sync string) *argocd.Application {
		return &argocd.Application{Status: argocd.ApplicationStatus{
			Health: argocd.HealthStatus{Status: health},
			Sync:   argocd.SyncStatus{Status: sync},
		}}
	}

	cases := []struct {
		name     string
		observed *argocd.Application
		want     bool
	}{
		{name: "nil observed", observed: nil, want: false},
		{name: "empty status", observed: &argocd.Application{}, want: false},
		{name: "progressing and synced", observed: app("Progressing", "Synced"), want: false},
		{name: "degraded and synced", observed: app("Degraded", "Synced"), want: false},
		{name: "healthy and out of sync", observed: app("Healthy", "OutOfSync"), want: false},
		{name: "healthy and synced", observed: app("Healthy", "Synced"), want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &ArgoCDApplication{ObservedResource: tc.observed}
			if got := r.IsReady(); got != tc.want {
				t.Errorf("IsReady() = %v, want %v", got, tc.want)
			}
		})
	}
}
