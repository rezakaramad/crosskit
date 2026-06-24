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

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ProvisionesEntraResourcesWhenGroupReady": {
			reason: "The Function should compose group + appRole + roleAssignment when the group has an objectId",
			args: args{
				ctx: context.Background(),
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "provision"},
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "idp.rezakara.demo/v1beta1",
								"kind": "XTenantEntra",
								"metadata": {"name": "pillow-factory"},
								"spec": {}
							}`),
						},
						Resources: map[string]*fnv1.Resource{
							"group-argocd-pillow-factory": {
								Resource: resource.MustStructJSON(`{
									"apiVersion": "groups.azuread.m.upbound.io/v1beta1",
									"kind": "Group",
									"status": {
										"atProvider": {
											"objectId": "11111111-1111-1111-1111-111111111111"
										}
									}
								}`),
							},
						},
					},
					Input: resource.MustStructJSON(`{
						"apiVersion": "defaults.fn.crossplane.io/v1beta1",
						"kind": "Input",
						"argocdApp": {
							"appRegObjectID": "argocd-appreg-uuid",
							"enterpriseAppObjectID": "argocd-entapp-uuid"
						},
						"providerConfigRef": {
							"name": "azuread-pc",
							"kind": "ProviderConfig"
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "provision", Ttl: durationpb.New(response.DefaultTTL)},
				},
			},
		},
		"WaitsForGroupBeforeRoleAssignment": {
			reason: "The Function should compose group + appRole but skip roleAssignment when group has no objectId yet",
			args: args{
				ctx: context.Background(),
				req: &fnv1.RunFunctionRequest{
					Meta: &fnv1.RequestMeta{Tag: "waiting"},
					Observed: &fnv1.State{
						Composite: &fnv1.Resource{
							Resource: resource.MustStructJSON(`{
								"apiVersion": "idp.rezakara.demo/v1beta1",
								"kind": "XTenantEntra",
								"metadata": {"name": "pillow-factory"},
								"spec": {}
							}`),
						},
					},
					Input: resource.MustStructJSON(`{
						"apiVersion": "defaults.fn.crossplane.io/v1beta1",
						"kind": "Input",
						"argocdApp": {
							"appRegObjectID": "argocd-appreg-uuid",
							"enterpriseAppObjectID": "argocd-entapp-uuid"
						},
						"providerConfigRef": {
							"name": "azuread-pc",
							"kind": "ProviderConfig"
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1.RunFunctionResponse{
					Meta: &fnv1.ResponseMeta{Tag: "waiting", Ttl: durationpb.New(response.DefaultTTL)},
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
		})
	}
}
