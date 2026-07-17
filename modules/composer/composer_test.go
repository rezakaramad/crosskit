package composer

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
)

func init() {
	// Register corev1 types with the scheme for testing
	_ = corev1.AddToScheme(composed.Scheme)
}

// TestFunctionContext holds test types for generic FunctionContext
type TestXR struct {
	Name string
}

type TestDefaults struct {
	Value string
}

// TestBaseComposer_GetConditionType verifies that GetConditionType returns
// whatever string was set on the BaseComposer, including the empty string.
func TestBaseComposer_GetConditionType(t *testing.T) {
	tests := []struct {
		name          string
		conditionType string
		want          string
	}{
		{
			name:          "returns configured condition type",
			conditionType: "DeploymentReady",
			want:          "DeploymentReady",
		},
		{
			name:          "returns empty string when not set",
			conditionType: "",
			want:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseComposer[TestXR, TestDefaults]{
				ConditionType: tt.conditionType,
			}

			got := b.GetConditionType()
			if got != tt.want {
				t.Errorf("GetConditionType() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBaseComposer_ComposeDesiredResourceFrom verifies that ComposeDesiredResourceFrom
// handles nil inputs gracefully and correctly converts typed K8s resources into
// a DesiredResource with ReadyFalse (resources start as not-ready until proven otherwise).
func TestBaseComposer_ComposeDesiredResourceFrom(t *testing.T) {
	tests := []struct {
		name               string
		structuredResource runtime.Object
		resourceName       resource.Name
		wantNil            bool
		wantErr            bool
	}{
		{
			name:               "returns nil for nil resource",
			structuredResource: nil,
			resourceName:       "test-resource",
			wantNil:            true,
			wantErr:            false,
		},
		{
			// A typed nil (e.g. var s *corev1.Service = nil) is non-nil as a runtime.Object
			// interface value, so ComposeDesiredResourceFrom must use reflection to detect it.
			name:               "returns nil for typed nil pointer",
			structuredResource: (*corev1.Service)(nil),
			resourceName:       "test-resource",
			wantNil:            true,
			wantErr:            false,
		},
		{
			name: "composes valid service resource",
			structuredResource: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-service",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
				},
			},
			resourceName: "test-resource",
			wantNil:      false,
			wantErr:      false,
		},
		{
			name: "composes valid configmap resource",
			structuredResource: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-configmap",
				},
				Data: map[string]string{
					"key": "value",
				},
			},
			resourceName: "configmap-resource",
			wantNil:      false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := FunctionContext[TestXR, TestDefaults]{
				FunctionResponse: &fnv1.RunFunctionResponse{},
				Log:              logging.NewNopLogger(),
			}

			b := &BaseComposer[TestXR, TestDefaults]{
				FunctionContext: ctx,
				ResourceName:    tt.resourceName,
				ConditionType:   "TestReady",
			}

			got, err := b.ComposeDesiredResourceFrom(tt.structuredResource)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil result, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil result, got nil")
			}
			if got.Name != tt.resourceName {
				t.Errorf("Name = %q, want %q", got.Name, tt.resourceName)
			}
			if got.Resource == nil {
				t.Fatal("expected non-nil Resource")
			}
			if got.Resource.Ready != resource.ReadyFalse {
				t.Errorf("Ready = %v, want %v", got.Resource.Ready, resource.ReadyFalse)
			}
		})
	}
}

// TestConvertObserved verifies that ConvertObserved correctly deserializes observed
// composed resources from the unstructured map into typed Go structs.
// A missing resource name returns nil without error — callers treat nil as "not yet created".
func TestConvertObserved(t *testing.T) {
	tests := []struct {
		name         string
		observed     map[resource.Name]resource.ObservedComposed
		resourceName resource.Name
		wantNil      bool
		wantErr      bool
		wantName     string
	}{
		{
			name:         "returns nil when resource not in observed map",
			observed:     map[resource.Name]resource.ObservedComposed{},
			resourceName: "missing-resource",
			wantNil:      true,
			wantErr:      false,
		},
		{
			name:         "returns nil when observed map is empty",
			observed:     map[resource.Name]resource.ObservedComposed{},
			resourceName: "any-resource",
			wantNil:      true,
			wantErr:      false,
		},
		{
			name: "converts observed service successfully",
			observed: map[resource.Name]resource.ObservedComposed{
				"test-service": {
					Resource: buildUnstructuredService("my-service", "10.0.0.1"),
				},
			},
			resourceName: "test-service",
			wantNil:      false,
			wantErr:      false,
			wantName:     "my-service",
		},
		{
			name: "converts observed configmap successfully",
			observed: map[resource.Name]resource.ObservedComposed{
				"test-configmap": {
					Resource: buildUnstructuredConfigMap("my-configmap"),
				},
			},
			resourceName: "test-configmap",
			wantNil:      false,
			wantErr:      false,
			wantName:     "my-configmap",
		},
		{
			name: "returns nil when looking for wrong resource name",
			observed: map[resource.Name]resource.ObservedComposed{
				"existing-resource": {
					Resource: buildUnstructuredService("my-service", "10.0.0.1"),
				},
			},
			resourceName: "different-resource",
			wantNil:      true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertObserved[corev1.Service](tt.observed, tt.resourceName)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil result, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil result, got nil")
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

// TestConvertObserved_ConfigMap is a focused test that also validates the Data field
// is preserved through the unstructured round-trip (the table test above only checks Name).
func TestConvertObserved_ConfigMap(t *testing.T) {
	observed := map[resource.Name]resource.ObservedComposed{
		"test-configmap": {
			Resource: buildUnstructuredConfigMap("my-configmap"),
		},
	}

	got, err := ConvertObserved[corev1.ConfigMap](observed, "test-configmap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if got.Name != "my-configmap" {
		t.Errorf("Name = %q, want %q", got.Name, "my-configmap")
	}
	if got.Data["test-key"] != "test-value" {
		t.Errorf("Data[test-key] = %q, want %q", got.Data["test-key"], "test-value")
	}
}

func TestConvertObserved_TypeMismatch(t *testing.T) {
	// Create a Service in the observed map
	observed := map[resource.Name]resource.ObservedComposed{
		"test-resource": {
			Resource: buildUnstructuredService("my-service", "10.0.0.1"),
		},
	}

	// Try to convert it as a ConfigMap - should succeed but with empty/default fields
	// because FromUnstructured is lenient
	got, err := ConvertObserved[corev1.ConfigMap](observed, "test-resource")
	// FromUnstructured doesn't error on type mismatch, it just sets fields it can
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	// The name field is common to all resources so it gets set
	if got.Name != "my-service" {
		t.Errorf("Name = %q, want %q", got.Name, "my-service")
	}
}

// TestFunctionContext_Fields verifies that all fields on FunctionContext are
// accessible and hold the values they were initialized with.
func TestFunctionContext_Fields(t *testing.T) {
	xr := TestXR{Name: "test-xr"}
	defaults := TestDefaults{Value: "test-value"}
	observed := map[resource.Name]resource.ObservedComposed{}
	rsp := &fnv1.RunFunctionResponse{}
	log := logging.NewNopLogger()

	ctx := FunctionContext[TestXR, TestDefaults]{
		Observed:         observed,
		FunctionResponse: rsp,
		XR:               xr,
		Defaults:         defaults,
		Log:              log,
	}

	if ctx.XR != xr {
		t.Errorf("XR = %+v, want %+v", ctx.XR, xr)
	}
	if ctx.Defaults != defaults {
		t.Errorf("Defaults = %+v, want %+v", ctx.Defaults, defaults)
	}
	if !reflect.DeepEqual(ctx.Observed, observed) {
		t.Errorf("Observed = %+v, want %+v", ctx.Observed, observed)
	}
	if ctx.FunctionResponse != rsp {
		t.Errorf("FunctionResponse = %p, want %p", ctx.FunctionResponse, rsp)
	}
	if ctx.Log == nil {
		t.Error("expected non-nil Log")
	}
}

// TestDesiredResource_Fields verifies that DesiredResource correctly pairs a
// resource name with its desired composed state, including the Ready status.
func TestDesiredResource_Fields(t *testing.T) {
	name := resource.Name("test-resource")
	desiredComposed := &resource.DesiredComposed{
		Ready: resource.ReadyTrue,
	}

	dr := DesiredResource{
		Name:     name,
		Resource: desiredComposed,
	}

	if dr.Name != name {
		t.Errorf("Name = %q, want %q", dr.Name, name)
	}
	if dr.Resource != desiredComposed {
		t.Errorf("Resource = %p, want %p", dr.Resource, desiredComposed)
	}
	if dr.Resource.Ready != resource.ReadyTrue {
		t.Errorf("Ready = %v, want %v", dr.Resource.Ready, resource.ReadyTrue)
	}
}

// Helper functions to build unstructured resources for testing

func buildUnstructuredService(name, clusterIP string) *composed.Unstructured {
	u := &composed.Unstructured{
		Unstructured: unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]any{
					"name": name,
				},
				"spec": map[string]any{
					"type":      "ClusterIP",
					"clusterIP": clusterIP,
				},
			},
		},
	}
	return u
}

func buildUnstructuredConfigMap(name string) *composed.Unstructured {
	u := &composed.Unstructured{
		Unstructured: unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata": map[string]any{
					"name": name,
				},
				"data": map[string]any{
					"test-key": "test-value",
				},
			},
		},
	}
	return u
}
