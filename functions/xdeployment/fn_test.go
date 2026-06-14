package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

func TestRunFunction(t *testing.T) {
	cases := map[string]struct {
		reason string
		req    *fnv1.RunFunctionRequest
		check  func(t *testing.T, rsp *fnv1.RunFunctionResponse)
	}{
		"FatalOnMissingObservedXR": {
			reason: "The Function should return a fatal result when no observed XR is present",
			req: &fnv1.RunFunctionRequest{
				Meta: &fnv1.RequestMeta{Tag: "no-xr"},
			},
			check: func(t *testing.T, rsp *fnv1.RunFunctionResponse) {
				t.Helper()
				if diff := cmpopts.AnyError.Error(); diff == "" {
					// just check structure
				}
				if rsp.Meta.Ttl.AsDuration() != durationpb.New(response.DefaultTTL).AsDuration() {
					t.Errorf("unexpected TTL")
				}
				if len(rsp.Results) == 0 || rsp.Results[0].Severity != fnv1.Severity_SEVERITY_FATAL {
					t.Errorf("expected fatal result, got: %v", rsp.Results)
				}
				if len(rsp.Conditions) == 0 {
					t.Errorf("expected FunctionSuccess=False condition, got none")
					return
				}
				c := rsp.Conditions[0]
				if c.Type != "FunctionSuccess" || c.Status != fnv1.Status_STATUS_CONDITION_FALSE || c.Reason != "InternalError" {
					t.Errorf("unexpected condition: type=%q status=%v reason=%q", c.Type, c.Status, c.Reason)
				}
			},
		},
		"RendersDeploymentForValidXDeployment": {
			reason: "The Function should render a Deployment for a valid XDeployment XR",
			req: &fnv1.RunFunctionRequest{
				Meta: &fnv1.RequestMeta{Tag: "render"},
				Observed: &fnv1.State{
					Composite: &fnv1.Resource{
						Resource: resource.MustStructJSON(`{
							"apiVersion": "idp.rezakara.demo/v1alpha1",
							"kind": "XDeployment",
							"metadata": {"name": "my-app"},
							"spec": {
								"image": "nginx:1.25",
								"replicas": 2,
								"namespace": "my-namespace"
							}
						}`),
					},
				},
			},
			check: func(t *testing.T, rsp *fnv1.RunFunctionResponse) {
				t.Helper()
				if len(rsp.Results) > 0 && rsp.Results[0].Severity == fnv1.Severity_SEVERITY_FATAL {
					t.Errorf("unexpected fatal result: %s", rsp.Results[0].Message)
				}
				if rsp.Desired == nil || len(rsp.Desired.Resources) != 1 {
					t.Errorf("expected 1 desired resource, got %d", len(rsp.GetDesired().GetResources()))
					return
				}
				if _, ok := rsp.Desired.Resources["my-app-deployment"]; !ok {
					t.Errorf("expected 'my-app-deployment' in desired resources, got: %v", rsp.Desired.Resources)
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(context.Background(), tc.req)
			if err != nil {
				t.Fatalf("%s: RunFunction returned unexpected error: %v", tc.reason, err)
			}
			tc.check(t, rsp)
		})
	}
}
