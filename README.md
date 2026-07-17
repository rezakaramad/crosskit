<p align="center" width="100%">
	<img width="24%" src="./logo.png">
</p>
<p align="center" >
	<img src="https://img.shields.io/badge/go-00ADD8?style=flat&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/crossplane-326CE5?style=flat&logo=crossplane&logoColor=white" />
</p>

A monorepo for Crossplane-related libraries, functions, code generation tools, and shared API types.

Each top-level directory has a pretty specific job.

## Repository structure

| Path | What it is for |
| --- | --- |
| [functions/](./functions/) | Crossplane composition functions. |
| [modules/](./modules/) | Shared Go modules for the rest of the repository. |
| [cmd/](./cmd/) | Entry points for repo-owned CLI tools. |
| [types/](./types/) | Shared API types and schemas used by the rest of the repo. |
| [.github/](.github/) | Automation for CI. |

## Releasing

### Go modules (`modules/`, `types/`)

Tag the module path and push.

```sh
git tag modules/composer/v0.1.0
git tag types/xtenantentra/v0.1.0
git push origin modules/composer/v0.1.0 types/xtenantentra/v0.1.0
```

> Never move a tag after it has been pushed — the proxy caches it. Cut a new patch version instead.

### Function packages (`functions/`)

Tag the function path and push. CI (`publish-functions.yml`) fires automatically and:

1. Builds a multi-arch Docker runtime image → `ghcr.io/<owner>/function-<name>-runtime:<version>`
2. Builds a Crossplane `.xpkg` package embedding that image → `ghcr.io/<owner>/function-<name>:<version>`

```sh
git tag functions/xtenantentra/v0.1.0
git push origin functions/xtenantentra/v0.1.0
```

Then optionally create a GitHub Release for changelog/notes:

```sh
gh release create functions/xtenantentra/v0.1.0 --title "functions/xtenantentra v0.1.0" --notes ""
```

**Useful commands:**

```sh
gh run list --workflow "Publish Functions" --limit 10
gh run view <run-id> --log-failed
gh release list --limit 20
```

Made with 🤓, 🐧 and 🍷.
