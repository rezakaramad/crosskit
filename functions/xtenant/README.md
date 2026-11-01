# function-xtenant

A [Crossplane composition function][functions] that composes the Argo CD
resources which deploy a tenant. Each `XTenant` produces two children:

| Resource | Name | Deploys | To |
| --- | --- | --- | --- |
| `Application` | `<tenant>-management` | `tenant-management` chart | management cluster |
| `ApplicationSet` | `<tenant>-workload` | `tenant-workload` chart | workload clusters |

The `ApplicationSet` generates one `Application` per matching workload cluster,
named `<tenant>-workload-<env>`.

```shell
# Run code generation - see input/generate.go
$ go generate ./...

# Run tests - see fn_test.go
$ go test ./...

# Build the function's runtime image - see Dockerfile
$ docker build . --tag=runtime

# Build a function package - see package/crossplane.yaml
$ crossplane xpkg build -f package --embed-runtime-image=runtime
```

## Argo CD types

The upstream `github.com/argoproj/argo-cd` Go module cannot be imported cleanly
as a library, so this function models the small subset of the `Application` and
`ApplicationSet` schema it emits in the local [`argocd`](argocd) package.

## Local development

`go.mod` points the `types/xtenant` module at the local directory with a
`replace` directive. Remove it and require the tagged version once
`types/xtenant` is released.

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
