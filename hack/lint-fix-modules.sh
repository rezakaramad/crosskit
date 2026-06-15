#!/usr/bin/env bash

set -euo pipefail

# Auto-fixes modules: tidies go.mod/go.sum and applies golangci-lint
# auto-fixes (formatting, import ordering, and linters marked auto-fix).
# Run this locally before validate-modules.sh to clear mechanical issues.
# You can pass a single module name as an argument to fix only that module.

repo_root=$(cd "$(dirname "$0")/.." && pwd)
modules_dir="$repo_root/modules"
golangci_config="$repo_root/.golangci.yml"
export GOWORK=off

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "error: golangci-lint is not installed or not on PATH" >&2
  echo "install it from https://golangci-lint.run/welcome/install/" >&2
  exit 1
fi

fix_module() {
  local module_dir="$1"
  local module_name
  module_name=$(basename "$module_dir")

  echo "==> $module_name: go mod tidy"
  (
    cd "$module_dir"
    go mod tidy
  )

  echo "==> $module_name: golangci-lint --fix"
  (
    cd "$module_dir"
    golangci-lint run --fix --config "$golangci_config"
  )
}

if [[ $# -gt 0 ]]; then
  target="$modules_dir/$1"
  if [[ ! -d "$target" ]]; then
    echo "error: module '$1' not found under $modules_dir" >&2
    exit 1
  fi
  fix_module "$target"
  exit 0
fi

while IFS= read -r -d '' dir; do
  fix_module "$dir"
done < <(find "$modules_dir" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)
