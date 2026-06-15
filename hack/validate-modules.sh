#!/usr/bin/env bash

set -euo pipefail

# Validates all modules: gofmt, mod-tidy, golangci-lint, govulncheck, unit tests.
# Run hack/lint-fix-modules.sh first to auto-fix formatting and tidy go.mod/go.sum.
# You can pass a single module name as an argument to validate only that module.

repo_root=$(cd "$(dirname "$0")/.." && pwd)
modules_dir="$repo_root/modules"
golangci_config="$repo_root/.golangci.yml"
export GOWORK=off

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "error: golangci-lint is not installed or not on PATH" >&2
  echo "install it from https://golangci-lint.run/welcome/install/" >&2
  exit 1
fi

if ! command -v govulncheck >/dev/null 2>&1; then
  echo "error: govulncheck is not installed or not on PATH" >&2
  echo "install it with: go install golang.org/x/vuln/cmd/govulncheck@latest" >&2
  exit 1
fi

validate_module() {
  local module_dir="$1"
  local module_name
  module_name=$(basename "$module_dir")

  echo "==> $module_name: gofmt"
  (
    cd "$module_dir"
    unformatted=$(gofmt -l .)
    if [[ -n "$unformatted" ]]; then
      echo "error: the following files are not gofmt-formatted:" >&2
      echo "$unformatted" >&2
      exit 1
    fi
  )

  echo "==> $module_name: go mod tidy"
  (
    cd "$module_dir"
    go mod tidy
    if ! git diff --exit-code -- go.mod go.sum; then
      echo "error: go.mod/go.sum are not tidy, run 'go mod tidy'" >&2
      exit 1
    fi
  )

  echo "==> $module_name: golangci-lint"
  (
    cd "$module_dir"
    golangci-lint run --config "$golangci_config"
  )

  echo "==> $module_name: govulncheck"
  (
    cd "$module_dir"
    govulncheck ./...
  )

  echo "==> $module_name: go test"
  (
    cd "$module_dir"
    go test ./...
  )
}

if [[ $# -gt 0 ]]; then
  target="$modules_dir/$1"
  if [[ ! -d "$target" ]]; then
    echo "error: module '$1' not found under $modules_dir" >&2
    exit 1
  fi
  validate_module "$target"
  exit 0
fi

while IFS= read -r -d '' dir; do
  validate_module "$dir"
done < <(find "$modules_dir" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)
