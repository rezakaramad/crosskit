# composer

Generic building blocks for composing Crossplane composed (child) resources
from a composite resource (XR) and its parsed input.

## The problem this solves

A Crossplane composition function receives a request (the XR, its input, and the
currently observed child resources) and must return the set of child resources
it wants to exist, plus status conditions describing their readiness. For
**every** child resource, that means the same low-level chores:

1. Find the observed resource (the child resource) by name in an untyped map and deserialize it from
   `unstructured` into a typed Go struct.
2. Build the desired resource as a typed struct.
3. Convert that typed struct into Crossplane's `composed` form, handling the
   conversion error.
4. Decide whether to skip creation (a typed-nil pointer must not become an empty
   object in the response).
5. Inspect the observed struct to compute readiness, then set the matching
   status condition on the XR.

Without a shared abstraction, that boilerplate is copy-pasted into every
function's `RunFunction` and repeated per resource. The consequences:

- **Duplication** — the same convert/deserialize/skip/condition code in every
  file, drifting subtly over time.
- **Coupling** — `RunFunction` has to know the concrete type and quirks of each
  resource, so adding a resource means editing the main loop.
- **Easy mistakes** — forgetting the typed-nil skip check emits phantom empty
  resources; inconsistent error handling loses failures; condition names drift.

### What it looks like without the module

One child resource, inlined into `RunFunction`; and this block repeats,
copy-pasted, for every child resource:

```go
// --- ApplicationSet: deserialize observed ---
name := resource.Name("applicationset-argocd-" + xr.GetName())
var observed argocd.ApplicationSet
if obs, ok := req.Observed[name]; ok {
    if err := runtime.DefaultUnstructuredConverter.
        FromUnstructured(obs.Resource.UnstructuredContent(), &observed); err != nil {
        response.Fatal(rsp, err)
        return rsp, nil
    }
}

// --- ApplicationSet: build + convert desired ---
desiredObj := buildApplicationSet(xr, input) // *argocd.ApplicationSet
if desiredObj != nil { // easy to forget → phantom empty resource
    cd, err := composed.From(desiredObj)
    if err != nil {
        response.Fatal(rsp, err)
        return rsp, nil
    }

    // --- ApplicationSet: readiness + condition ---
    ready := false
    for _, c := range observed.Status.Conditions {
        if c.Type == "ResourcesUpToDate" && c.Status == "True" {
            ready = true
        }
    }
    if ready {
        response.ConditionTrue(rsp, "ArgoCDApplicationSetReady", "Available").TargetComposite()
    } else {
        response.ConditionFalse(rsp, "ArgoCDApplicationSetReady", "Unavailable").TargetComposite()
    }
    desired[name] = &resource.DesiredComposed{Resource: cd}
}
```

With the module, the deserialize/convert/skip/condition machinery lives in
`BaseComposer` and the loop, so the same resource is just:

```go
r, err := resources.NewArgoCDApplicationSet(fnContext) // observed load
// createResource() builds it; IsReady() checks ResourcesUpToDate
```

...plus one line registering it in `buildComposers`. The loop handles convert,
skip, and conditions for every composer identically.

## What the module gives instead

The module extracts all of that into two things:

- A generic **`FunctionContext`/`BaseComposer`** pair that carries the request
  data and implements the repetitive convert/skip/condition plumbing once.
- A small **`ComposableResource`** interface so `RunFunction` can treat every
  child resource uniformly: build a `[]ComposableResource`, loop over it, and
  call the same four methods on each, without knowing what any of them are.

Adding a new child resource becomes "write one **composer** and register it",
instead of "edit the main loop and re-derive the boilerplate". The rest of this
document describes the pieces that make that possible.

A **composer** is an object that knows how to produce and monitor exactly one
child resource of the XR. Each function builds a list of composers; the function
loop asks each one what to create and whether it's ready. A composer is a
**two-way adapter** between typed Go structs and Crossplane's untyped world:

- **Outbound (build):** `createResource()` makes a typed object →
  `ComposeDesiredResourceFrom` wraps it into a `DesiredResource`.
- **Inbound (check):** `ConvertObserved` reads cluster state into a typed object
  → `IsReady()` inspects it.

`BaseComposer` is a shared struct that gives concrete composers common fields and
a few generic helper methods, especially `GetConditionType`, `ComposeDesiredResourceFrom`,
and `GetConnectionDetails`

## The pieces

### Data types

- **`FunctionContext[XR, D]`**: the input bundle for one reconcile call:
  `Observed` (current cluster state), `FunctionResponse` (where results/errors
  go), `XR` (the parent), `Defaults` (the parsed input), `Log`. The type
  parameters `XR` and `D` are pinned per-function (e.g. via an `XContext` alias).
- **`DesiredResource`**: the output unit: a `Name` plus the desired
  `*resource.DesiredComposed`. A composer produces one, or `nil` to mean
  "don't create it".
- **`BaseComposer[XR, D]`**: the shared struct concrete composers embed
  (often aliased `XComposer`). Provides the `FunctionContext`, `ResourceName`,
  and `ConditionType` fields plus the methods below.

### Contract

- **`ComposableResource`**: the interface the function loop talks to. Being a
  composer means implementing:
  - `ComposeDesiredResource() (*DesiredResource, error)`: build it
  - `IsReady() bool`: is it done reconciling
  - `GetConditionType() string`: status condition name
  - `GetConnectionDetails() map[string]string`: optional secret data

  `BaseComposer` already implements `GetConditionType` and
  `GetConnectionDetails`, so each concrete composer usually only writes
  `ComposeDesiredResource` and `IsReady`.

### Helpers on `BaseComposer`

- **`GetConditionType()`**: returns the stored `ConditionType`.
- **`ComposeDesiredResourceFrom(obj)`**: the workhorse. Converts a typed object
  into a `*DesiredResource`. Returns `(nil, nil)` for a typed-nil pointer,
  signalling "skip this resource"; reports a fatal on conversion failure.
- **`GetConnectionDetails()`**: base returns `nil`; override to contribute
  secret data.

### Package functions

- **`ConvertObserved[T](observed, name)`**: the mirror of
  `ComposeDesiredResourceFrom`: looks up the observed resource by name and
  deserializes it into typed `T`, or returns `nil` if not yet created. Feeds
  `IsReady()`.
- **`ComposeConnectionSecret(name, details)`**: builds a Kubernetes `Secret`
  `DesiredResource` from key/value pairs.

## Writing a composer

1. Define a struct embedding `BaseComposer` (via the function's `XComposer`
   alias) and add an `ObservedResource *T` field.
2. Write `NewXxx(f XContext) (*Xxx, error)`: compute the `ResourceName`, load the
   observed resource with `composer.ConvertObserved[T]`, set `ConditionType`.
3. Implement `createResource()`:
   - Return just `*T` when the build can't fail (the common case).
   - Return `(*T, error)` only when there's genuinely fallible logic
     (e.g. a computed lookup or cross-field validation). Field-shape validation
     is better expressed with kubebuilder markers on the input type, which
     Crossplane enforces before the function runs.
4. Implement `ComposeDesiredResource()`:
   - No-error variant: `return r.ComposeDesiredResourceFrom(r.createResource())`
   - Error variant: call `createResource()`, check `err`, then
     `ComposeDesiredResourceFrom`.
5. Implement `IsReady()` by inspecting `ObservedResource`.
6. Register the composer in the function's `buildComposers` so the loop runs it.
