# function-xwebapp

A [Crossplane composition function][functions] that renders the resources of an
`XWebApp`: a `ServiceAccount`, a `Deployment`, a `Service`, and, when
`platformCapabilities.hostname` is set, an `HTTPRoute`.

Developers write a native Kubernetes pod template under `spec.template`. The
function deep-copies it into the `Deployment`, and the API server validates it.
The function only overwrites what the platform owns: the selector and pod
labels, `serviceAccountName`, `tolerations` and `topologySpreadConstraints`.
Those last two are hidden from the XRD schema (see `gen-xrd --hide`).

## Rules enforced by the function

- Every container port must have a unique, valid name; the names are used by the
  `Service` and by NetworkPolicies.
- `platformCapabilities.hostname` requires a container port named `http`.

A violation stops the composition with a `FunctionSuccess=False` condition.

## Generating the XRD

```shell
go run github.com/rezakaramad/crosskit/cmd/gen-xrd \
  --package github.com/rezakaramad/crosskit/types/xwebapp \
  --type    XWebApp \
  --group   idp.rezakara.demo \
  --version v1beta1 \
  --hide    spec.template.spec.tolerations \
  --hide    spec.template.spec.topologySpreadConstraints \
  | kubectl apply --server-side -f -
```

The XRD is large (about 600 KB), so it must be applied server-side.

```shell
# Run code generation - see input/generate.go
$ go generate -tags generate ./...

# Run tests - see fn_test.go
$ go test ./...

# Build the function's runtime image - see Dockerfile
$ docker build . --tag=runtime

# Build a function package - see package/crossplane.yaml
$ crossplane xpkg build -f package --embed-runtime-image=runtime
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
