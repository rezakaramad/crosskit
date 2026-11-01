package main

import (
	"context"
	"sort"
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
	providerConfig = `"providerConfigRef": {"name": "azuread-pc", "kind": "ProviderConfig"}`

	argocdApp = `{
		"name": "argocd",
		"displayName": "ArgoCD",
		"appRegObjectID": "argocd-appreg-uuid",
		"enterpriseAppObjectID": "argocd-entapp-uuid"
	}`
	vaultApp = `{
		"name": "vault",
		"displayName": "Vault",
		"appRegObjectID": "vault-appreg-uuid",
		"enterpriseAppObjectID": "vault-entapp-uuid"
	}`
)

func inputWith(apps ...string) *fnv1.RunFunctionRequest {
	list := ""
	for i, a := range apps {
		if i > 0 {
			list += ","
		}
		list += a
	}
	return &fnv1.RunFunctionRequest{
		Meta: &fnv1.RequestMeta{Tag: "compose"},
		Observed: &fnv1.State{
			Composite: &fnv1.Resource{
				Resource: resource.MustStructJSON(`{
					"apiVersion": "idp.rezakara.demo/v1beta1",
					"kind": "XTenantSSO",
					"metadata": {"name": "pillow-factory"},
					"spec": {}
				}`),
			},
		},
		Input: resource.MustStructJSON(`{
			"apiVersion": "defaults.fn.crossplane.io/v1beta1",
			"kind": "Input",
			"apps": [` + list + `],
			` + providerConfig + `
		}`),
	}
}

func withObservedGroup(req *fnv1.RunFunctionRequest, name, objectID string) *fnv1.RunFunctionRequest {
	if req.Observed.Resources == nil {
		req.Observed.Resources = map[string]*fnv1.Resource{}
	}
	req.Observed.Resources[name] = &fnv1.Resource{
		Resource: resource.MustStructJSON(`{
			"apiVersion": "groups.azuread.m.upbound.io/v1beta1",
			"kind": "Group",
			"status": {"atProvider": {"objectId": "` + objectID + `"}}
		}`),
	}
	return req
}

func TestRunFunction(t *testing.T) {
	cases := map[string]struct {
		reason string
		req    *fnv1.RunFunctionRequest
		want   []string
	}{
		"ComposesGroupAndRoleForEveryApp": {
			reason: "Group and AppRole are composed for each app; the assignments wait for the groups",
			req:    inputWith(argocdApp, vaultApp),
			want: []string{
				"group-argocd-pillow-factory", "app-role-argocd-pillow-factory",
				"group-vault-pillow-factory", "app-role-vault-pillow-factory",
			},
		},
		"ComposesAssignmentOnceGroupHasObjectID": {
			reason: "An app's assignment is composed once its group has an object ID, independently of the other apps",
			req:    withObservedGroup(inputWith(argocdApp, vaultApp), "group-argocd-pillow-factory", "11111111-1111-1111-1111-111111111111"),
			want: []string{
				"group-argocd-pillow-factory", "app-role-argocd-pillow-factory", "roleassignment-argocd-sso-role-pillow-factory",
				"group-vault-pillow-factory", "app-role-vault-pillow-factory",
			},
		},
		"SingleApp": {
			reason: "The number of resources follows the apps list",
			req:    inputWith(vaultApp),
			want:   []string{"group-vault-pillow-factory", "app-role-vault-pillow-factory"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(context.Background(), tc.req)
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

			var got []string
			for k := range rsp.GetDesired().GetResources() {
				got = append(got, k)
			}
			sort.Strings(got)
			want := append([]string(nil), tc.want...)
			sort.Strings(want)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("%s\ndesired resources -want +got:\n%s", tc.reason, diff)
			}
		})
	}
}
