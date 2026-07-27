package resources

import (
	"testing"

	"github.com/rezakaramad/crosskit/functions/xtenantargo/argocd"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantargo/input/v1beta1"
	xtenantargo "github.com/rezakaramad/crosskit/types/xtenantargo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testInput() *inputv1beta1.Input {
	return &inputv1beta1.Input{
		Project: "tenant-system",
		Management: inputv1beta1.ManagementConfig{
			RepoURL:         "https://github.com/talktorubberduckdev/platform-hub",
			Path:            "charts/tenant-management",
			TargetRevision:  "HEAD",
			TargetNamespace: "tenant-platform-resources",
		},
		Workload: inputv1beta1.WorkloadConfig{
			RepoURL:        "https://github.com/talktorubberduckdev/platform-hub",
			Path:           "charts/tenant-workload",
			TargetRevision: "HEAD",
			Helm: inputv1beta1.WorkloadHelmConfig{
				ValueFiles: []string{`values-plt-{{ index .metadata.labels "argocd.rezakara.demo/environment" }}.yaml`},
				Parameters: []inputv1beta1.HelmParameter{
					{
						Name:  "environmentPrefix",
						Value: `{{ index .metadata.labels "argocd.rezakara.demo/environment" }}`,
					},
				},
			},
			TargetClusters: inputv1beta1.TargetClustersConfig{
				ClusterTypeKey: "argocd.rezakara.demo/cluster-type",
				ClusterType:    "tenant",
				EnvironmentKey: "argocd.rezakara.demo/environment",
				Environments:   []string{"dev", "test", "prod", "wl"},
			},
		},
		Namespace: inputv1beta1.NamespaceConfig{ApplicationSet: "argocd", Prefix: "tn-"},
	}
}

func TestArgoCDApplicationSet_CreateResource(t *testing.T) {
	xr := &xtenantargo.XTenantArgo{
		ObjectMeta: metav1.ObjectMeta{Name: "pillow-factory"},
		Spec:       xtenantargo.XTenantArgoSpec{},
	}
	r := &ArgoCDApplicationSet{XComposer: XComposer{FunctionContext: XContext{XR: xr, Defaults: testInput()}}}

	got, err := r.createResource()
	if err != nil {
		t.Fatalf("createResource() error = %v", err)
	}

	// --- metadata ---
	if got.Name != "pillow-factory" {
		t.Errorf("metadata.name = %q, want %q", got.Name, "pillow-factory")
	}
	if got.Namespace != "argocd" {
		t.Errorf("metadata.namespace = %q, want %q", got.Namespace, "argocd")
	}
	if got.Labels[tenantLabelKey] != "pillow-factory" {
		t.Errorf("label %q = %q, want %q", tenantLabelKey, got.Labels[tenantLabelKey], "pillow-factory")
	}
	if !got.Spec.GoTemplate {
		t.Errorf("goTemplate = false, want true")
	}

	// --- generators ---
	if len(got.Spec.Generators) != 2 {
		t.Fatalf("generators len = %d, want 2", len(got.Spec.Generators))
	}
	if got.Spec.Generators[0].List == nil {
		t.Fatal("generators[0].list is nil, want list generator")
	}
	if got.Spec.Generators[1].Clusters == nil {
		t.Fatal("generators[1].clusters is nil, want clusters generator")
	}

	// --- management generator ---
	listGen := got.Spec.Generators[0].List
	if len(listGen.Elements) != 1 || listGen.Elements[0]["cluster"] != "management" {
		t.Errorf("list.elements = %+v, want [{cluster:management}]", listGen.Elements)
	}
	mgmtTmpl := listGen.Template
	if mgmtTmpl.Name != "pillow-factory-management" {
		t.Errorf("management template.name = %q, want %q", mgmtTmpl.Name, "pillow-factory-management")
	}
	if mgmtTmpl.Spec.Project != "tenant-system" {
		t.Errorf("management template.spec.project = %q, want tenant-system", mgmtTmpl.Spec.Project)
	}
	if mgmtTmpl.Spec.Source == nil || mgmtTmpl.Spec.Source.RepoURL != "https://github.com/talktorubberduckdev/platform-hub" {
		t.Errorf("management source.repoURL = %v", mgmtTmpl.Spec.Source)
	}
	if mgmtTmpl.Spec.Source.Path != "charts/tenant-management" {
		t.Errorf("management source.path = %q, want charts/tenant-management", mgmtTmpl.Spec.Source.Path)
	}
	if mgmtTmpl.Spec.Destination.Name != "in-cluster" {
		t.Errorf("management destination.name = %q, want in-cluster", mgmtTmpl.Spec.Destination.Name)
	}
	if mgmtTmpl.Spec.Destination.Namespace != "tenant-platform-resources" {
		t.Errorf("management destination.namespace = %q, want tenant-platform-resources", mgmtTmpl.Spec.Destination.Namespace)
	}

	// --- workload generator ---
	clustersGen := got.Spec.Generators[1].Clusters
	exprs := clustersGen.Selector.MatchExpressions
	if len(exprs) != 2 {
		t.Fatalf("cluster selector matchExpressions = %d, want 2", len(exprs))
	}
	if exprs[0].Key != "argocd.rezakara.demo/cluster-type" || exprs[0].Values[0] != "tenant" {
		t.Errorf("cluster-type expr = %+v", exprs[0])
	}
	if exprs[1].Key != "argocd.rezakara.demo/environment" || len(exprs[1].Values) != 4 {
		t.Errorf("environment expr = %+v", exprs[1])
	}
	wlTmpl := clustersGen.Template
	wantWLName := `pillow-factory-workload-{{ index .metadata.labels "argocd.rezakara.demo/environment" }}`
	if wlTmpl.Name != wantWLName {
		t.Errorf("workload template.name = %q, want %q", wlTmpl.Name, wantWLName)
	}
	if wlTmpl.Spec.Project != "tenant-system" {
		t.Errorf("workload template.spec.project = %q, want tenant-system", wlTmpl.Spec.Project)
	}
	if wlTmpl.Spec.Source == nil {
		t.Fatal("workload source is nil")
	}
	if wlTmpl.Spec.Source.RepoURL != "https://github.com/talktorubberduckdev/platform-hub" {
		t.Errorf("workload source.repoURL = %q", wlTmpl.Spec.Source.RepoURL)
	}
	if wlTmpl.Spec.Source.Path != "charts/tenant-workload" {
		t.Errorf("workload source.path = %q, want charts/tenant-workload", wlTmpl.Spec.Source.Path)
	}
	if wlTmpl.Spec.Source.Helm == nil {
		t.Fatal("workload source.helm is nil")
	}
	if len(wlTmpl.Spec.Source.Helm.ValueFiles) != 1 {
		t.Errorf("workload helm.valueFiles len = %d, want 1", len(wlTmpl.Spec.Source.Helm.ValueFiles))
	}
	if len(wlTmpl.Spec.Source.Helm.Parameters) != 1 || wlTmpl.Spec.Source.Helm.Parameters[0].Name != "environmentPrefix" {
		t.Errorf("workload helm.parameters = %+v", wlTmpl.Spec.Source.Helm.Parameters)
	}
	if wlTmpl.Spec.Destination.Name != "{{ .name }}" {
		t.Errorf("workload destination.name = %q, want {{ .name }}", wlTmpl.Spec.Destination.Name)
	}
	if wlTmpl.Spec.Destination.Namespace != "tn-pillow-factory" {
		t.Errorf("workload destination.namespace = %q, want tn-pillow-factory", wlTmpl.Spec.Destination.Namespace)
	}

	// --- base template ---
	baseTmpl := got.Spec.Template
	if baseTmpl.Spec.Project != "tenant-system" {
		t.Errorf("base template.spec.project = %q, want tenant-system", baseTmpl.Spec.Project)
	}
	if baseTmpl.Spec.SyncPolicy == nil || len(baseTmpl.Spec.SyncPolicy.SyncOptions) == 0 {
		t.Errorf("base template.spec.syncPolicy = %+v, want syncOptions", baseTmpl.Spec.SyncPolicy)
	}

	// --- ignoreApplicationDifferences ---
	if len(got.Spec.IgnoreApplicationDifferences) != 1 ||
		got.Spec.IgnoreApplicationDifferences[0].JSONPointers[0] != "/spec/syncPolicy" {
		t.Errorf("ignoreApplicationDifferences = %+v", got.Spec.IgnoreApplicationDifferences)
	}
}

func TestArgoCDApplicationSet_IsReady(t *testing.T) {
	cases := []struct {
		name     string
		observed *argocd.ApplicationSet
		want     bool
	}{
		{name: "nil observed", observed: nil, want: false},
		{name: "no conditions", observed: &argocd.ApplicationSet{}, want: false},
		{
			name: "ResourcesUpToDate=True",
			observed: &argocd.ApplicationSet{Status: argocd.ApplicationSetStatus{
				Conditions: []argocd.ApplicationSetCondition{
					{Type: "ResourcesUpToDate", Status: "True"},
				},
			}},
			want: true,
		},
		{
			name: "ResourcesUpToDate=False",
			observed: &argocd.ApplicationSet{Status: argocd.ApplicationSetStatus{
				Conditions: []argocd.ApplicationSetCondition{
					{Type: "ResourcesUpToDate", Status: "False"},
				},
			}},
			want: false,
		},
		{
			name: "ErrorOccurred=True only",
			observed: &argocd.ApplicationSet{Status: argocd.ApplicationSetStatus{
				Conditions: []argocd.ApplicationSetCondition{
					{Type: "ErrorOccurred", Status: "True"},
				},
			}},
			want: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := &ArgoCDApplicationSet{ObservedResource: tt.observed}
			if got := r.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
