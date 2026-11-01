package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

const (
	managementName = "xtenant-application-management-pillow-factory"
	workloadName   = "xtenant-applicationset-workload-pillow-factory"
)

const input = `{
	"apiVersion": "defaults.fn.crossplane.io/v1beta1",
	"kind": "Input",
	"project": "tenant-system",
	"tenantLabelKey": "platform.rezakara.demo/tenant",
	"management": {
		"repoURL": "https://github.com/rezakaramad/kubepave",
		"path": "kubernetes/charts/tenant-management",
		"targetRevision": "HEAD",
		"targetNamespace": "tenant-system"
	},
	"workload": {
		"repoURL": "https://github.com/rezakaramad/kubepave",
		"path": "kubernetes/charts/tenant-workload",
		"targetRevision": "HEAD",
		"targetClusters": {
			"clusterTypeKey": "argocd.rezakara.demo/cluster-type",
			"clusterType": "tenant",
			"environmentKey": "argocd.rezakara.demo/environment",
			"environments": ["development"]
		}
	},
	"namespace": {"argo": "argocd"}
}`

func observedApplication(name, health, sync string) *fnv1.Resource {
	return &fnv1.Resource{Resource: resource.MustStructJSON(`{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind": "Application",
		"metadata": {"name": "` + name + `", "namespace": "argocd"},
		"status": {"health": {"status": "` + health + `"}, "sync": {"status": "` + sync + `"}}
	}`)}
}

func TestRunFunction(t *testing.T) {
	xr := &fnv1.Resource{Resource: resource.MustStructJSON(`{
		"apiVersion": "idp.rezakara.demo/v1beta1",
		"kind": "XTenant",
		"metadata": {"name": "pillow-factory"},
		"spec": {"shortName": "pil"}
	}`)}

	cases := map[string]struct {
		reason   string
		observed map[string]*fnv1.Resource
		want     map[string]fnv1.Ready
	}{
		"ComposesAllChildren": {
			reason:   "Both children are composed on the first reconcile, none ready yet",
			observed: nil,
			want: map[string]fnv1.Ready{
				managementName: fnv1.Ready_READY_FALSE,
				workloadName:   fnv1.Ready_READY_FALSE,
			},
		},
		"ReadinessIsPerChild": {
			reason: "Each child reports its own readiness; a ready management Application does not make the workload ApplicationSet ready",
			observed: map[string]*fnv1.Resource{
				managementName: observedApplication("pillow-factory-management", "Healthy", "Synced"),
			},
			want: map[string]fnv1.Ready{
				managementName: fnv1.Ready_READY_TRUE,
				workloadName:   fnv1.Ready_READY_FALSE,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := &fnv1.RunFunctionRequest{
				Meta: &fnv1.RequestMeta{Tag: "compose"},
				Observed: &fnv1.State{
					Composite: xr,
					Resources: tc.observed,
				},
				Input: resource.MustStructJSON(input),
			}

			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(context.Background(), req)
			if err != nil {
				t.Fatalf("%s\nRunFunction() error: %v", tc.reason, err)
			}

			wantMeta := &fnv1.ResponseMeta{Tag: "compose", Ttl: durationpb.New(response.DefaultTTL)}
			if diff := cmp.Diff(wantMeta, rsp.GetMeta(), cmpopts.IgnoreUnexported(fnv1.ResponseMeta{}, durationpb.Duration{})); diff != "" {
				t.Errorf("%s\nRunFunction() meta -want +got:\n%s", tc.reason, diff)
			}
			for _, r := range rsp.GetResults() {
				if r.GetSeverity() == fnv1.Severity_SEVERITY_FATAL {
					t.Errorf("%s\nunexpected fatal result: %s", tc.reason, r.GetMessage())
				}
			}

			got := map[string]fnv1.Ready{}
			for k, v := range rsp.GetDesired().GetResources() {
				got[k] = v.GetReady()
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s\ndesired resources -want +got:\n%s", tc.reason, diff)
			}
		})
	}
}
