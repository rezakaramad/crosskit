# crossplane-toolkit

Small Go toolkit for building Crossplane functions and related tooling.

It currently includes:
- `modules/runner`: typed helpers for writing Crossplane composition functions
- `modules/generator`: generate Crossplane XRDs from annotated Go types using [controller-tools](https://github.com/kubernetes-sigs/controller-tools)
