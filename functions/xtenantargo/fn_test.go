package main

import (
	"context"
	"testing"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRunFunction(t *testing.T) {
	type args struct {
		ctx context.Context
		req *fnv1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1.RunFunctionResponse
		err error
	}

	input := `{
		"apiVersion": "defaults.fn.crossplane.io/v1beta1",
		"kind": "XTenantArgoDefaults",
		"project": "tenant-system",
		"management": {
			"repoURL": "https://github.com/talktorubberduckdev/platform-hub",
			"path": "charts/tenant-management",
			"targetRevision": "HEAD",
			"targetNamespace": "tenant-system"
		},
		"workload": {
			"repoURL": "https://github.com/talktorubberduckdev/platform-hub",
			"path": "charts/tenant-workload",
			"targetRevision": "HEAD",
			"helm": {
				"valueFiles": ["values-plt-{{ index .metadata.labels \"argocd.rezakara.demo/environment\" }}.yaml"],
				"parameters": [{"name": "environmentPrefix", "value": "{{ index .metadata.labels \"argocd.rezakara.demo/environment\" }}"}]
			},
			"targetClusters": {
				"clusterTypeKey": "argocd.rezakara.demo/cluster-type",
				"clusterType": "tenant",
				"environmentKey": "argocd.rezakara.demo/environment",
				"environments": ["dev", "test", "prod", "wl"]
			}
		},
		"namespace": {
			"applicationSet": "argocd",
			"prefix": "tn-"
		}
	}`

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ComposesApplicationSet": {
			reason: "The Function should compose the tenant ApplicationSet from the XR and input",
			args: args{
				ctx: context.Background(),
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "compose"},
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "idp.rezakara.demo/v1beta1",
								"kind": "XTenantArgo",
								"metadata": {"name": "pillow-factory"},
								"spec": {"shortName": "pil"}
							}`),
						},
					},
					Input: resource.MustStructJSON(input),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "compose", Ttl: durationpb.New(response.DefaultTTL)},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)
			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nRunFunction() error -want +got:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.rsp, rsp,
				cmpopts.IgnoreFields(fnv1.RunFunctionResponse{}, "Desired", "Results", "Conditions"),
				cmpopts.IgnoreUnexported(fnv1.RunFunctionResponse{}, fnv1.ResponseMeta{}, durationpb.Duration{}),
			); diff != "" {
				t.Errorf("%s\nRunFunction() response -want +got:\n%s", tc.reason, diff)
			}
			for _, r := range rsp.GetResults() {
				if r.GetSeverity() == fnv1.Severity_SEVERITY_FATAL {
					t.Errorf("%s\nunexpected fatal result: %s", tc.reason, r.GetMessage())
				}
			}

			desired := rsp.GetDesired().GetResources()
			if _, ok := desired["applicationset-argocd-pillow-factory"]; !ok {
				t.Errorf("%s\nexpected desired resource %q, got keys %v",
					tc.reason, "applicationset-argocd-pillow-factory", keys(desired))
			}
		})
	}
}

func keys(m map[string]*fnv1.Resource) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
