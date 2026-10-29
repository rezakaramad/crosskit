package main

import (
	"context"
	"testing"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
)

const defaultsInput = `{
	"apiVersion": "defaults.fn.crossplane.io/v1beta1",
	"kind": "Input",
	"tolerations": [{"key": "dedicated", "value": "web", "effect": "NoSchedule"}],
	"gateway": {"name": "shared-gw", "namespace": "gateway-system"}
}`

func xr(spec string) string {
	return `{
		"apiVersion": "idp.rezakara.demo/v1beta1",
		"kind": "XWebApp",
		"metadata": {"name": "object-lister", "namespace": "acme", "labels": {"team": "storage"}},
		"spec": ` + spec + `
	}`
}

func run(t *testing.T, xrJSON string) *fnv1.RunFunctionResponse {
	t.Helper()

	f := &Function{log: logging.NewNopLogger()}
	rsp, err := f.RunFunction(context.Background(), &fnv1.RunFunctionRequest{
		Input:    resource.MustStructJSON(defaultsInput),
		Observed: &fnv1.State{Composite: &fnv1.Resource{Resource: resource.MustStructJSON(xrJSON)}},
	})
	if err != nil {
		t.Fatalf("RunFunction: %v", err)
	}
	return rsp
}

func fatal(rsp *fnv1.RunFunctionResponse) *fnv1.Result {
	for _, r := range rsp.GetResults() {
		if r.GetSeverity() == fnv1.Severity_SEVERITY_FATAL {
			return r
		}
	}
	return nil
}

func desired(t *testing.T, rsp *fnv1.RunFunctionResponse, name string) map[string]any {
	t.Helper()

	r, ok := rsp.GetDesired().GetResources()[name]
	if !ok {
		t.Fatalf("desired resource %q not found, have %v", name, rsp.GetDesired().GetResources())
	}
	return r.GetResource().AsMap()
}

func TestRunFunction_ComposesAllResources(t *testing.T) {
	rsp := run(t, xr(`{
		"replicas": 3,
		"template": {
			"metadata": {"labels": {"tier": "backend"}},
			"spec": {
				"initContainers": [{"name": "migrate", "image": "migrate:v2"}],
				"containers": [{
					"name": "app",
					"image": "app:v1",
					"ports": [{"name": "http", "containerPort": 8080}, {"name": "metrics", "containerPort": 9090}]
				}]
			}
		},
		"platformCapabilities": {"hostname": "acme.dev.example.com"}
	}`))

	if f := fatal(rsp); f != nil {
		t.Fatalf("unexpected fatal result: %s", f.GetMessage())
	}

	for _, name := range []string{
		"xwebapp-serviceaccount-object-lister",
		"xwebapp-deployment-object-lister",
		"xwebapp-service-object-lister",
		"xwebapp-httproute-object-lister",
	} {
		desired(t, rsp, name)
	}

	dep := desired(t, rsp, "xwebapp-deployment-object-lister")
	spec := dep["spec"].(map[string]any)
	if got := spec["replicas"]; got != float64(3) {
		t.Errorf("replicas = %v, want 3", got)
	}

	pod := spec["template"].(map[string]any)["spec"].(map[string]any)
	if got := pod["serviceAccountName"]; got != "object-lister" {
		t.Errorf("serviceAccountName = %v, want object-lister", got)
	}
	if len(pod["tolerations"].([]any)) != 1 {
		t.Errorf("platform tolerations not applied: %v", pod["tolerations"])
	}
	if len(pod["topologySpreadConstraints"].([]any)) != 1 {
		t.Errorf("topologySpreadConstraints not applied: %v", pod["topologySpreadConstraints"])
	}
	if len(pod["initContainers"].([]any)) != 1 || len(pod["containers"].([]any)) != 1 {
		t.Errorf("developer containers not preserved: %v", pod)
	}

	labels := spec["template"].(map[string]any)["metadata"].(map[string]any)["labels"].(map[string]any)
	if labels["tier"] != "backend" || labels["app.kubernetes.io/name"] != "object-lister" {
		t.Errorf("unexpected pod labels: %v", labels)
	}

	svcPorts := desired(t, rsp, "xwebapp-service-object-lister")["spec"].(map[string]any)["ports"].([]any)
	if len(svcPorts) != 2 || svcPorts[0].(map[string]any)["targetPort"] != "http" {
		t.Errorf("unexpected service ports: %v", svcPorts)
	}

	route := desired(t, rsp, "xwebapp-httproute-object-lister")["spec"].(map[string]any)
	parent := route["parentRefs"].([]any)[0].(map[string]any)
	if parent["name"] != "shared-gw" || parent["namespace"] != "gateway-system" {
		t.Errorf("unexpected parentRef: %v", parent)
	}
}

func TestRunFunction_NoHostnameSkipsHTTPRoute(t *testing.T) {
	rsp := run(t, xr(`{
		"template": {"spec": {"containers": [{"name": "app", "image": "app:v1", "ports": [{"name": "http", "containerPort": 8080}]}]}}
	}`))

	if f := fatal(rsp); f != nil {
		t.Fatalf("unexpected fatal result: %s", f.GetMessage())
	}
	if _, ok := rsp.GetDesired().GetResources()["xwebapp-httproute-object-lister"]; ok {
		t.Error("httproute should not be composed without a hostname")
	}
	if got := desired(t, rsp, "xwebapp-deployment-object-lister")["spec"].(map[string]any)["replicas"]; got != float64(2) {
		t.Errorf("default replicas = %v, want 2", got)
	}
}

func TestRunFunction_NoPortsSkipsService(t *testing.T) {
	rsp := run(t, xr(`{"template": {"spec": {"containers": [{"name": "app", "image": "app:v1"}]}}}`))

	if f := fatal(rsp); f != nil {
		t.Fatalf("unexpected fatal result: %s", f.GetMessage())
	}
	if _, ok := rsp.GetDesired().GetResources()["xwebapp-service-object-lister"]; ok {
		t.Error("service should not be composed without container ports")
	}
}

func TestRunFunction_InvalidSpecIsFatal(t *testing.T) {
	cases := map[string]string{
		"unnamed port":   `{"template": {"spec": {"containers": [{"name": "app", "image": "a", "ports": [{"containerPort": 8080}]}]}}}`,
		"duplicate port": `{"template": {"spec": {"containers": [{"name": "a", "image": "a", "ports": [{"name": "http", "containerPort": 1}]}, {"name": "b", "image": "b", "ports": [{"name": "http", "containerPort": 2}]}]}}}`,
		"no containers":  `{"template": {"spec": {"containers": []}}}`,
		"hostname without http port": `{
			"template": {"spec": {"containers": [{"name": "app", "image": "a", "ports": [{"name": "grpc", "containerPort": 1}]}]}},
			"platformCapabilities": {"hostname": "a.example.com"}
		}`,
	}

	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			rsp := run(t, xr(spec))
			if fatal(rsp) == nil {
				t.Error("expected a fatal result")
			}
			if len(rsp.GetDesired().GetResources()) != 0 {
				t.Errorf("no resources should be composed, got %v", rsp.GetDesired().GetResources())
			}
		})
	}
}
