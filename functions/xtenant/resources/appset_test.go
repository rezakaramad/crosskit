package resources

import (
	"testing"

	"github.com/rezakaramad/crosskit/functions/xtenant/argocd"
)

func TestArgoCDApplicationSet_CreateResource(t *testing.T) {
	r, err := NewWorkloadApplicationSet(testXContext())
	if err != nil {
		t.Fatal(err)
	}

	got := r.createResource()

	if got.Name != "pillow-factory-workload" {
		t.Errorf("metadata.name = %q, want pillow-factory-workload", got.Name)
	}
	if got.Namespace != "argocd" {
		t.Errorf("metadata.namespace = %q, want argocd", got.Namespace)
	}
	if got.Labels[testTenantLabelKey] != "pillow-factory" {
		t.Errorf("label %q = %q, want pillow-factory", testTenantLabelKey, got.Labels[testTenantLabelKey])
	}
	if !got.Spec.GoTemplate {
		t.Errorf("goTemplate = false, want true")
	}

	// Only workload clusters are targeted; the management side is composed as Applications.
	if len(got.Spec.Generators) != 1 || got.Spec.Generators[0].Clusters == nil {
		t.Fatalf("generators = %+v, want one clusters generator", got.Spec.Generators)
	}
	gen := got.Spec.Generators[0].Clusters

	exprs := gen.Selector.MatchExpressions
	if len(exprs) != 2 {
		t.Fatalf("cluster selector matchExpressions = %d, want 2", len(exprs))
	}
	if exprs[0].Key != "argocd.rezakara.demo/cluster-type" || exprs[0].Values[0] != "tenant" {
		t.Errorf("cluster-type expr = %+v", exprs[0])
	}
	if exprs[1].Key != "argocd.rezakara.demo/environment" || len(exprs[1].Values) != 1 || exprs[1].Values[0] != "development" {
		t.Errorf("environment expr = %+v", exprs[1])
	}

	tmpl := gen.Template
	wantName := `pillow-factory-workload-{{ index .metadata.labels "argocd.rezakara.demo/environment" }}`
	if tmpl.Name != wantName {
		t.Errorf("template.name = %q, want %q", tmpl.Name, wantName)
	}
	if tmpl.Spec.Project != "tenant-system" {
		t.Errorf("template.spec.project = %q, want tenant-system", tmpl.Spec.Project)
	}
	if tmpl.Spec.Source == nil || tmpl.Spec.Source.Path != "kubernetes/charts/tenant-workload" {
		t.Fatalf("template.spec.source = %+v", tmpl.Spec.Source)
	}
	helm := tmpl.Spec.Source.Helm
	if len(helm.ValueFiles) != 1 {
		t.Errorf("helm.valueFiles len = %d, want 1", len(helm.ValueFiles))
	}
	if len(helm.Parameters) != 2 ||
		helm.Parameters[0] != (argocd.HelmParameter{Name: "tenant.name", Value: "pillow-factory"}) ||
		helm.Parameters[1] != (argocd.HelmParameter{Name: "tenant.tenantShortName", Value: "pil"}) {
		t.Errorf("helm.parameters = %+v", helm.Parameters)
	}
	if tmpl.Spec.Destination.Name != "{{ .name }}" || tmpl.Spec.Destination.Namespace != "pillow-factory" {
		t.Errorf("template.spec.destination = %+v", tmpl.Spec.Destination)
	}

	if got.Spec.Template.Spec.SyncPolicy == nil || len(got.Spec.Template.Spec.SyncPolicy.SyncOptions) == 0 {
		t.Errorf("base template syncPolicy = %+v, want syncOptions", got.Spec.Template.Spec.SyncPolicy)
	}
	if len(got.Spec.IgnoreApplicationDifferences) != 1 ||
		got.Spec.IgnoreApplicationDifferences[0].JSONPointers[0] != "/spec/syncPolicy" {
		t.Errorf("ignoreApplicationDifferences = %+v", got.Spec.IgnoreApplicationDifferences)
	}
}

func TestArgoCDApplicationSet_IsReady(t *testing.T) {
	withCondition := func(status string) *argocd.ApplicationSet {
		return &argocd.ApplicationSet{Status: argocd.ApplicationSetStatus{
			Conditions: []argocd.ApplicationSetCondition{{Type: "ResourcesUpToDate", Status: status}},
		}}
	}

	cases := []struct {
		name     string
		observed *argocd.ApplicationSet
		want     bool
	}{
		{name: "nil observed", observed: nil, want: false},
		{name: "no conditions", observed: &argocd.ApplicationSet{}, want: false},
		{name: "ResourcesUpToDate=False", observed: withCondition("False"), want: false},
		{name: "ResourcesUpToDate=True", observed: withCondition("True"), want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &ArgoCDApplicationSet{ObservedResource: tc.observed}
			if got := r.IsReady(); got != tc.want {
				t.Errorf("IsReady() = %v, want %v", got, tc.want)
			}
		})
	}
}
