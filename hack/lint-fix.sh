#!/usr/bin/env bash

set -euo pipefail

# Auto-fixes Go packages: tidies go.mod/go.sum and applies golangci-lint
# auto-fixes (formatting, import ordering, and linters marked auto-fix).
# Run this locally before validate.sh to clear mechanical issues.
#
# Usage:
#   hack/lint-fix.sh                        # fix all modules/ and functions/
#   hack/lint-fix.sh modules/gcpdns         # fix a specific package
#   hack/lint-fix.sh functions/xtenant-render

repo_root=$(cd "$(dirname "$0")/.." && pwd)
golangci_config="$repo_root/.golangci.yml"
export GOWORK=off

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "error: golangci-lint is not installed or not on PATH" >&2
  echo "install it from https://golangci-lint.run/welcome/install/" >&2
  exit 1
fi

fix_package() {
  local pkg_dir="$1"
  local pkg_name
  pkg_name=$(realpath --relative-to="$repo_root" "$pkg_dir")

  echo "==> $pkg_name: go mod tidy"
  (
    cd "$pkg_dir"
    go mod tidy
  )

  echo "==> $pkg_name: golangci-lint --fix"
  (
    cd "$pkg_dir"
    golangci-lint run --fix --config "$golangci_config"
  )
}

if [[ $# -gt 0 ]]; then
  target="$repo_root/$1"
  if [[ ! -d "$target" ]]; then
    echo "error: '$1' not found under $repo_root" >&2
    exit 1
  fi
  fix_package "$target"
  exit 0
fi

# No argument: fix all modules/ and functions/
for base_dir in "$repo_root/modules" "$repo_root/functions"; do
  [[ -d "$base_dir" ]] || continue
  while IFS= read -r -d '' dir; do
    fix_package "$dir"
  done < <(find "$base_dir" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)
done
