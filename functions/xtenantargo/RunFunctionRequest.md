# Crossplane Function gRPC Contract: RunFunctionRequest & RunFunctionResponse

Ref: https://buf.build/crossplane/crossplane/docs/main:apiextensions.fn.proto.v1 

## RunFunctionRequest: what Crossplane sends to `req`

| proto field | type | what it is | where in your code |
|---|---|---|---|
| `meta` (1) | `RequestMeta` | Request metadata: `tag` (dedup id) + `capabilities` (what this Crossplane version supports) | [fn.go](fn.go#L62) `req.GetMeta().GetTag()` |
| `observed` (2) | `State` | Actual current state of the XR + composed resources, as of pipeline start | via `request.GetObservedComposedResources(req)` and `request.GetObservedCompositeResource(req)` |
| `desired` (3) | `State` | Desired state **accumulated so far** by earlier functions in the pipeline (partial object) | via `request.GetDesiredComposedResources(req)` |
| `input` (4) | `Struct` (JSON) | The `input:` block from your Composition pipeline step | via `request.GetInput(req, input)` into your `inputv1beta1.Input` |
| `context` (5) | `Struct` | Arbitrary data passed between functions in the pipeline | not used here |
| `extra_resources` (6, **deprecated**) | `map<string,Resources>` | Old mechanism for fetched extra resources | not used |
| `credentials` (7) | `map<string,Credentials>` | Secrets declared in the Composition | not used |
| `required_resources` (8) | `map<string,Resources>` | Resources you asked for (via response `requirements.resources`) that Crossplane fetched and fed back | not used |
| `required_schemas` (9) | `map<string,Schema>` | OpenAPI schemas you requested | not used |

### The `State` message (used by `observed` and `desired`)

A composition never deals with a single resource; it deals with the XR (the parent) **and** all the resources it composes (the children). `State` is the container that bundles both:

```proto
message State {
  Resource composite = 1;              // the ONE parent XR
  map<string, Resource> resources = 2; // MANY children, keyed by name
}
```

- `composite` (1) → the XR itself (a single `Resource`, e.g. `XTenantEntra`). There is always exactly one per call. You read it via `request.GetObservedCompositeResource(req)`.
- `resources` (2) → `map<string, Resource>` of composed resources. The **key** is the resource's composition-local name (a logical label you choose, like `applicationset`), **not** the Kubernetes `metadata.name`. The map (rather than a list) gives each child a stable identity across reconciles, so Crossplane updates the same resource instead of creating duplicates. You read/write it via `request.GetObservedComposedResources(req)` / `request.GetDesiredComposedResources(req)`, and the line `desired[desiredResource.Name] = desiredResource.Resource` sets an entry in this map.

`observed` and `desired` share this exact structure; only their meaning differs:

| | `composite` | `resources` |
|---|---|---|
| **observed** | XR as it currently exists in the cluster | children as they currently exist |
| **desired** | what the XR *should* look like (status only) | children you *want* to exist |

At this point, I was a little puzzled by `observed` and `desired`. so I tried to wrap my head around them and eventually came to this:

**Crossplane runs a loop**:
Crossplane is constantly reconciling. Every few moments it asks: "*What does the world look like right now*", and "*what should it look like?*" Then it makes the world match. Those two questions are exactly the two words:

`observed` = what currently exists in the cluster (reality, read-only to you). Crossplane fills this in for you.
`desired` = what you want to exist (your intent). You build this and hand it back.

Crossplane's job = look at the gap between observed and desired, then create/update/delete real Kubernetes objects until reality matches desire.

flowchart LR
    O["observed<br/>(what IS)"] --> F["your function"]
    F --> D["desired<br/>(what SHOULD BE)"]
    D --> K["Crossplane makes<br/>the cluster match"]
    K -->|next reconcile| O

You never write `observed`; it's a report. You never read your own past writes from `observed` directly for intent; you express intent purely through `desired`.

**Why your function receives `desired` in the *request* too**

This is the confusing part. `desired` shows up in both the *request* and the *response*. Why?

Because a Composition can run **several functions in a pipeline**, one after another:

```shell
function A  →  function B  →  function C
```

Each function adds its piece to the same growing `desired` map. So when your function runs:

- the `desired` in the request = "*what functions before me already asked for*" (could be empty if you're first).
- the `desired` in the response = "*that, plus my additions*."

That's why your code does this at the top:

```go
desired, err := request.GetDesiredComposedResources(req)  // start from what's already there
```

instead of desired := map[...]{}. If you started fresh, you'd wipe out what earlier functions wanted. The golden rule from the schema: **pass through what you don't own, add what you do**.

The question that came up here was: *What would it happen if we had only one function in the pipeline?*
On the very first reconcile, the `desired` composed-resources map in your request is empty; because no function ran before you, so nobody has contributed any desired children yet.

**Two clarifications so it's not misleading**
1. "Empty" is about the composed resources map, not the whole request.
Even as the only function, your request still contains real data in `observed` (the actual cluster state), `input`, `meta`, etc. Only `desired.resources` (the children map) starts empty.

2. It's empty this run, but not necessarily empty across the object's lifetime.
There's a subtlety people miss: `desired` is rebuilt from scratch **on every reconcile**. Crossplane does not carry your previous `desired` map into the next reconcile's request. Each loop iteration, your function runs again and re-declares the full desired state from nothing:

flowchart LR
    subgraph R1["reconcile #1"]
        A1["req.desired = empty"] --> B1["you add applicationset"] --> C1["rsp.desired = {applicationset}"]
    end
    subgraph R2["reconcile #2"]
        A2["req.desired = empty again"] --> B2["you add applicationset"] --> C2["rsp.desired = {applicationset}"]
    end
    C1 -.->|Crossplane reconciles cluster, then loops| A2

So even in a one-function pipeline you must **always re-add** your resource every run. If your function ever returned an empty `desired`, Crossplane would read that as "you no longer want the ApplicationSet" and **delete** it.

**So why start from request.GetDesiredComposedResources(req) at all?**
For a single-function pipeline you could start from an empty map and it'd work identically; the inherited map is empty anyway. But starting from the request's desired is the correct, future-proof habit because:

- if someone later adds another function before yours in the pipeline, your code already respects their contributions instead of wiping them, and
- it matches the SDK's intended pattern (pass through what you don't own, add what you do).

## RunFunctionResponse: what you return in `rsp`

| proto field | type | what it is | where in your code |
|---|---|---|---|
| `meta` (1) | `ResponseMeta` | `tag` (must match request) + `ttl` (when Crossplane calls again) | set by `response.To(req, response.DefaultTTL)` at [fn.go](fn.go#L64) |
| `desired` (2) | `State` | The desired state you want — **partial**; anything you omit that was previously desired gets deleted | set by `response.SetDesiredComposedResources(rsp, desired)` at [fn.go](fn.go#L145) |
| `results` (3) | `repeated Result` | Observability events (Fatal / Warning / Normal) | `response.Fatal(rsp, ...)` calls produce these |
| `context` (4) | `Struct` | Context handed to the next function | not used |
| `requirements` (5) | `Requirements` | Ask Crossplane to fetch extra `resources` / `schemas` for the next reconcile | not used |
| `conditions` (6) | `repeated Condition` | Status conditions applied to the XR (and optionally claim) | all your `response.ConditionTrue/False(...).Target...()` calls |
| `output` (7) | `Struct` | Only for Operations; XRs discard it | not used |

`RunFunctionResponse` is the object your function builds and returns to tell Crossplane what to do. Think of the request as "here's the situation" and the response as "here’s what should happen next."

