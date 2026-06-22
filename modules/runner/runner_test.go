package runner

import (
	"context"
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/resource/composite"
)

// mustJSON marshals v to a JSON string, panicking on error.
// Only used in tests to build raw unstructured XR content.
func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// ─── Shared test helpers ──────────────────────────────────────────────────────

// makeRequest builds a minimal RunFunctionRequest with the given XR name and
// an optional set of already-observed composed resources.
func makeRequest(xrName string, observed map[string]*fnv1.Resource) *fnv1.RunFunctionRequest {
	if observed == nil {
		observed = map[string]*fnv1.Resource{}
	}
	return &fnv1.RunFunctionRequest{
		Observed: &fnv1.State{
			Composite: &fnv1.Resource{
				Resource: resource.MustStructObject(&corev1.ConfigMap{
					TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
					ObjectMeta: metav1.ObjectMeta{Name: xrName},
				}),
			},
			Resources: observed,
		},
		Desired: &fnv1.State{Resources: map[string]*fnv1.Resource{}},
		Input: resource.MustStructObject(&corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "defaults"},
		}),
	}
}

// ─── Baseline unit tests ──────────────────────────────────────────────────────

type unusedStruct struct{ Name string }

func TestNewDecodeTarget_Pointer(t *testing.T) {
	got := newDecodeTarget[*unusedStruct]()
	if got == nil {
		t.Fatal("expected non-nil pointer for pointer type")
	}
	if got.Name != "" {
		t.Fatalf("expected zero value struct, got Name=%q", got.Name)
	}
}

func TestNewDecodeTarget_Value(t *testing.T) {
	got := newDecodeTarget[int]()
	if got != 0 {
		t.Fatalf("expected zero int, got %d", got)
	}
}

func TestIsTypedNil(t *testing.T) {
	var secret *corev1.Secret
	if !isTypedNil(secret) {
		t.Fatal("typed nil pointer should be treated as nil")
	}
	if isTypedNil(&corev1.Secret{}) {
		t.Fatal("non-nil pointer should not be treated as nil")
	}
	if !isTypedNil(nil) {
		t.Fatal("untyped nil should be treated as nil")
	}
}

func TestBuildConnectionSecret(t *testing.T) {
	c := composite.New()
	c.SetName("my-xr")
	xr := &resource.Composite{Resource: c}

	result, err := buildConnectionSecret(xr, map[string]string{
		"host": "db.internal",
		"port": "5432",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.name != resource.Name("my-xr-connection") {
		t.Fatalf("unexpected name: %q", result.name)
	}
	if result.desired.Ready != resource.ReadyTrue {
		t.Fatalf("connection secret should be ReadyTrue, got %v", result.desired.Ready)
	}

	var secret corev1.Secret
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		result.desired.Resource.UnstructuredContent(), &secret,
	); err != nil {
		t.Fatalf("failed to decode secret: %v", err)
	}
	if string(secret.Data["host"]) != "db.internal" {
		t.Fatalf("unexpected host: %q", string(secret.Data["host"]))
	}
}

// ─── Change 1: ObservedAs helper ─────────────────────────────────────────────

// TestObservedAs_ReturnsNilWhenMissing confirms that ObservedAs returns nil
// (not an error) for a resource that has not been created yet.
// Builders use this as the "not ready yet, retry next reconcile" signal.
func TestObservedAs_ReturnsNilWhenMissing(t *testing.T) {
	got, err := ObservedAs[*corev1.ConfigMap](
		map[resource.Name]resource.ObservedComposed{},
		resource.Name("not-there"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

// TestObservedAs_DecodesWhenPresent confirms that ObservedAs decodes a present
// resource into the typed struct.
func TestObservedAs_DecodesWhenPresent(t *testing.T) {
	cm := &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: "sibling"},
		Data:       map[string]string{"key": "value-from-sibling"},
	}
	composedCM, err := composed.From(cm)
	if err != nil {
		t.Fatalf("failed to build composed configmap: %v", err)
	}

	observed := map[resource.Name]resource.ObservedComposed{
		"sibling": {Resource: composedCM},
	}

	got, err := ObservedAs[*corev1.ConfigMap](observed, "sibling")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected decoded configmap, got nil")
	}
	if got.Data["key"] != "value-from-sibling" {
		t.Fatalf("unexpected Data[key]: %q", got.Data["key"])
	}
}

// ─── Change 2: sibling access inside a resource ──────────────────────────────

// siblingDependentResource reads a sibling ConfigMap from ctx.Observed and
// injects its data into the desired resource it produces.
// This models the objectId-piping pattern: resource A exists in ctx.Observed,
// resource B reads it and injects a field into its own desired spec.
type siblingDependentResource struct{}

func (siblingDependentResource) ConditionType() string { return "DependentReady" }

func (siblingDependentResource) ResourceName(_ Context[*corev1.ConfigMap, *corev1.ConfigMap]) resource.Name {
	return resource.Name("dependent")
}

// Desired reads the "upstream" sibling and returns nil when it is not present yet.
// When it is present, it copies the sibling's data into the desired resource.
func (siblingDependentResource) Compose(ctx Context[*corev1.ConfigMap, *corev1.ConfigMap]) (*corev1.ConfigMap, error) {
	upstream, err := ObservedAs[*corev1.ConfigMap](ctx.Observed, resource.Name("upstream"))
	if err != nil {
		return nil, err
	}
	if upstream == nil {
		// sibling not created yet — skip this reconcile
		return nil, nil
	}
	return &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: "dependent"},
		Data: map[string]string{
			// value piped from the sibling's observed status
			"fromSibling": upstream.Data["generatedID"],
		},
	}, nil
}

func (siblingDependentResource) IsReady(_ Context[*corev1.ConfigMap, *corev1.ConfigMap], obs *corev1.ConfigMap) bool {
	return obs != nil
}

func (siblingDependentResource) ConnectionDetails(_ Context[*corev1.ConfigMap, *corev1.ConfigMap], _ *corev1.ConfigMap) map[string]string {
	return nil
}

// TestResourceCanReadSibling_SkipsWhenSiblingMissing confirms that when the
// upstream sibling is not in ctx.Observed, the dependent resource returns nil
// and the runner skips it (no desired resource registered for it).
func TestResourceCanReadSibling_SkipsWhenSiblingMissing(t *testing.T) {
	req := makeRequest("xr-test", nil) // no observed composed resources

	r := New[*corev1.ConfigMap, *corev1.ConfigMap](req, logging.NewNopLogger())
	Register(r, siblingDependentResource{})

	rsp, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	resources := rsp.GetDesired().GetResources()
	if _, exists := resources["dependent"]; exists {
		t.Fatal("dependent should be skipped when upstream sibling is missing")
	}
}

// TestResourceCanReadSibling_ProducesDesiredWhenSiblingPresent confirms that
// when the sibling exists in ctx.Observed, the resource reads its data and
// injects it into the desired resource it produces.
func TestResourceCanReadSibling_ProducesDesiredWhenSiblingPresent(t *testing.T) {
	// Build an observed "upstream" ConfigMap that has already been created.
	upstream := &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: "upstream"},
		Data:       map[string]string{"generatedID": "abc-123"},
	}

	req := makeRequest("xr-test", map[string]*fnv1.Resource{
		"upstream": {Resource: resource.MustStructObject(upstream)},
	})

	r := New[*corev1.ConfigMap, *corev1.ConfigMap](req, logging.NewNopLogger())
	Register(r, siblingDependentResource{})

	rsp, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	resources := rsp.GetDesired().GetResources()
	dep, exists := resources["dependent"]
	if !exists {
		t.Fatal("expected dependent desired resource to be present")
	}

	var got corev1.ConfigMap
	if err := resource.AsObject(dep.GetResource(), &got); err != nil {
		t.Fatalf("failed to decode dependent: %v", err)
	}
	if got.Data["fromSibling"] != "abc-123" {
		t.Fatalf("expected fromSibling=abc-123, got %q", got.Data["fromSibling"])
	}
}

// ─── Change 3: WriteConnectionSecretToRef ────────────────────────────────────

// connectionResource always returns one connection detail.
type connectionResource struct{}

func (connectionResource) ConditionType() string { return "DBReady" }
func (connectionResource) ResourceName(_ Context[*corev1.ConfigMap, *corev1.ConfigMap]) resource.Name {
	return resource.Name("db")
}
func (connectionResource) Compose(_ Context[*corev1.ConfigMap, *corev1.ConfigMap]) (*corev1.ConfigMap, error) {
	return &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: "db"},
	}, nil
}
func (connectionResource) IsReady(_ Context[*corev1.ConfigMap, *corev1.ConfigMap], _ *corev1.ConfigMap) bool {
	return true
}
func (connectionResource) ConnectionDetails(_ Context[*corev1.ConfigMap, *corev1.ConfigMap], _ *corev1.ConfigMap) map[string]string {
	return map[string]string{"host": "db.internal"}
}

// TestConnectionSecret_DefaultName confirms that when the XR has no
// writeConnectionSecretToRef the secret is named "<xr-name>-connection".
func TestConnectionSecret_DefaultName(t *testing.T) {
	req := makeRequest("my-db", nil)

	r := New[*corev1.ConfigMap, *corev1.ConfigMap](req, logging.NewNopLogger())
	Register(r, connectionResource{})

	rsp, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	resources := rsp.GetDesired().GetResources()
	if _, ok := resources["my-db-connection"]; !ok {
		t.Fatalf("expected secret named my-db-connection, got keys: %v", resourceNames(resources))
	}
}

// TestConnectionSecret_UsesWriteConnectionSecretToRef confirms that when the XR
// has spec.writeConnectionSecretToRef.name set, the runner uses that name for
// the connection secret instead of the default "<xr-name>-connection".
func TestConnectionSecret_UsesWriteConnectionSecretToRef(t *testing.T) {
	// Build the XR's unstructured content manually so we can set
	// spec.writeConnectionSecretToRef.name — the field the runner reads.
	xrUnstructured := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   map[string]interface{}{"name": "my-db"},
		"spec": map[string]interface{}{
			"writeConnectionSecretToRef": map[string]interface{}{
				"name": "custom-db-secret",
			},
		},
	}

	req := &fnv1.RunFunctionRequest{
		Observed: &fnv1.State{
			Composite: &fnv1.Resource{
				Resource: resource.MustStructJSON(mustJSON(xrUnstructured)),
			},
			Resources: map[string]*fnv1.Resource{},
		},
		Desired: &fnv1.State{Resources: map[string]*fnv1.Resource{}},
		Input: resource.MustStructObject(&corev1.ConfigMap{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{Name: "defaults"},
		}),
	}

	r := New[*corev1.ConfigMap, *corev1.ConfigMap](req, logging.NewNopLogger())
	Register(r, connectionResource{})

	rsp, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	resources := rsp.GetDesired().GetResources()
	if _, ok := resources["custom-db-secret"]; !ok {
		t.Fatalf("expected secret named custom-db-secret, got keys: %v", resourceNames(resources))
	}
	if _, ok := resources["my-db-connection"]; ok {
		t.Fatal("default secret name should not be used when writeConnectionSecretToRef is set")
	}
}

// resourceNames returns sorted resource keys for human-readable test output.
func resourceNames(m map[string]*fnv1.Resource) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	return names
}
