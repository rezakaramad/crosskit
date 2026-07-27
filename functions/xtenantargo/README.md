# function-xtenantargo

A [Crossplane composition function][functions] that renders the Argo CD
`ApplicationSet` for a tenant. The `ApplicationSet` watches the tenant's
`platform-deploy-<tenant>` repository and deploys the applications it contains to
the workload clusters selected by cluster labels.

It is the function counterpart of the Helm-rendered `ApplicationSet` in the
`tenant-management` chart, and is consistent with the `xtenantentra` function in
coding style, structure, and design pattern.

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
as a library, so this function models the small subset of the `ApplicationSet`
schema it emits in the local [`argocd`](argocd) package, following the same
approach used by other functions in this repository.

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
