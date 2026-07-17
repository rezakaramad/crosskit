<p align="center" width="100%">
	<img width="24%" src="./logo.png">
</p>
<p align="center" >
	<img src="https://img.shields.io/badge/go-00ADD8?style=flat&logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/crossplane-326CE5?style=flat&logo=crossplane&logoColor=white" />

	<img src="https://img.shields.io/badge/github%20actions-CI-2088FF?style=flat&logo=githubactions&logoColor=white" />
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
| [hack/](./hack/) | Internal development scripts such as local fix and verify helpers. |
| [.github/](.github/) | Automation for CI, releases, and package publishing. |

## Typical workflow

| If you want to... | Start here |
| --- | --- |
| Add a new type | [types/README.md](./types/README.md) |
| Add a new function | [functions/README.md](./functions/README.md) |
| Use `runner` inside a function | [modules/runner/README.md](./modules/runner/README.md) |
| Work on generation tooling | [cmd/gen-xrd/README.md](./cmd/gen-xrd/README.md) |

## Local checks

- Run `task install:hooks` once after cloning — it installs a pre-push git hook that auto-fixes lint and formatting before every push.
- VS Code is set up to use [golangci-lint](https://github.com/golangci/golangci-lint) on save.
- We use a [Taskfile](https://taskfile.dev) to keep the common checks in one place.
- Run `task --list` to see what is available, then pick the check you need for one function or for all of them.
- The underlying scripts live in [hack/](./hack/), following the Kubernetes/CNCF convention for internal dev tooling.
- [Taskfile.yml](./Taskfile.yml) is the real source of truth for local commands.
- [.vscode/tasks.json](./.vscode/tasks.json) is only VS Code convenience. It makes the same commands clickable from `Tasks: Run Task`, but it duplicates what the Taskfile already defines.


A few useful ones:

```sh
$ task --list
task: Available tasks for this project:
* check:function:               Run tidy, lint, and tests for one function module.
* check:functions:              Run tidy, lint, and tests for all function modules.
* check:xtenant-render:         Run checks for xtenant-render.
* check:xtenant-validate:       Run checks for xtenant-validate.
* fix:function:                 Auto-fix tidy, lint, and formatting issues for one function module.
* fix:functions:                Auto-fix tidy, lint, and formatting issues for all function modules.
* fix:xtenant-render:           Auto-fix xtenant-render.
* fix:xtenant-validate:         Auto-fix xtenant-validate.

$ task fix:xtenant-validate
$ task check:xtenant-validate
$ task check:functions
```

## Functions: build and publish

The CI automatically discovers every directory under `functions/` and runs validation (lint + tests + build) on branches and pull requests. Packaging is verified on PRs; publishing to GHCR happens automatically as part of the release flow after the release PR is merged.

Artifacts produced by the release build:

- `ghcr.io/<owner>/<name>-runtime:<version>` — distroless runtime image containing the compiled Go binary
- `ghcr.io/<owner>/function-<name>:<version>` — the `.xpkg` Crossplane package (runtime image + `package/crossplane.yaml`)

Required files per function:

```
functions/<name>/
├── Dockerfile            # two-stage: golang → distroless
├── package/
│   └── crossplane.yaml   # name + capabilities; name must match functionRef.name in Compositions
└── go.mod
```

**How to add a new function (minimal):**

1. Scaffold the function with the Crossplane CLI:
   ```sh
   crossplane xpkg init <name> function-template-go --directory functions/<name>
   ```
   This creates the `Dockerfile`, `go.mod`, `package/crossplane.yaml`, and Go boilerplate under `functions/<name>/`.
2. Add the module to `go.work` (so local CI resolves modules consistently).
3. Open a PR and merge to `main`.
4. Publish by pushing a tag `functions/<name>/v<version>` — CI builds and pushes the runtime image and Crossplane package to GHCR automatically.

Notes:

- Tags drive releases: pushing `functions/<name>/v<version>` triggers the publish workflow.
- Keep one version series per package path (e.g. `functions/xtenantentra/v0.1.0`, `modules/composer/v0.1.0`).

## Release flow

**What's automatic vs manual:**

| Step | Who | Result |
| --- | --- | --- |
| Bump versions + changelogs | **You** | Commit updated `CHANGELOG.md` for each package |
| Create and push tags | **You** | Tags like `functions/<name>/v<version>` created |
| Create GitHub Releases | **You** (`gh release create`) | Release notes published |
| Build and publish | CI | Runtime image + Crossplane package pushed to GHCR (on function tag push) |

**Steps:**

1. Merge your PR to `main`; if you changed a library API, update function `go.mod` to the new library version first (CI compiles fine via `go.work`).
2. Tag each package you want to release and push the tags:
   ```sh
   git tag modules/composer/v0.1.0
   git tag functions/xtenantentra/v0.1.0
   git push origin modules/composer/v0.1.0 functions/xtenantentra/v0.1.0
   ```
3. Create GitHub Releases (optional but recommended for notes):
   ```sh
   gh release create modules/composer/v0.1.0 --title "modules/composer v0.1.0" --notes ""
   ```

Pushing a `functions/<name>/v<version>` tag triggers the publish workflow, which builds the runtime image and Crossplane package and pushes them to GHCR.

> Never delete and recreate a tag — bump the patch version instead. Moving a tag after the Go module proxy has cached it causes checksum mismatches.

**If a function build fails or you need to re-run it**, dispatch or re-push:
```sh
gh workflow run "Publish Functions" --ref main
# or re-trigger by pushing a new patch tag
```

**Useful commands:**
```sh
gh run list --workflow "Publish Functions" --limit 10
gh run view <run-id> --log-failed
gh release list --limit 20
gh api '/users/rezakaramad/packages?package_type=container&per_page=100' \
  --jq '.[] | select(.repository.full_name == "rezakaramad/crosskit") | .name'
```

**Version source of truth:** git tags (`<package-path>/v<semver>`) and each package's `CHANGELOG.md`.

Made with 🤓, 🐧 and 🍷.
