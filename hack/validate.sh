#!/usr/bin/env bash

set -euo pipefail

# Validates Go packages: gofmt, mod-tidy, golangci-lint, govulncheck, unit tests.
# Run hack/lint-fix.sh first to auto-fix formatting and tidy go.mod/go.sum.
#
# Usage:
#   hack/validate.sh                        # validate all modules/ and functions/
#   hack/validate.sh modules/gcpdns         # validate a specific package
#   hack/validate.sh functions/xtenant-render

repo_root=$(cd "$(dirname "$0")/.." && pwd)
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

validate_package() {
  local pkg_dir="$1"
  local pkg_name
  pkg_name=$(realpath --relative-to="$repo_root" "$pkg_dir")

  echo "==> $pkg_name: gofmt"
  (
    cd "$pkg_dir"
    unformatted=$(gofmt -l .)
    if [[ -n "$unformatted" ]]; then
      echo "error: the following files are not gofmt-formatted:" >&2
      echo "$unformatted" >&2
      exit 1
    fi
  )

  echo "==> $pkg_name: go mod tidy"
  (
    cd "$pkg_dir"
    go mod tidy
    if ! git diff --exit-code -- go.mod go.sum; then
      echo "error: go.mod/go.sum are not tidy, run 'go mod tidy'" >&2
      exit 1
    fi
  )

  echo "==> $pkg_name: golangci-lint"
  (
    cd "$pkg_dir"
    golangci-lint run --config "$golangci_config"
  )

  echo "==> $pkg_name: govulncheck"
  (
    cd "$pkg_dir"
    govulncheck ./...
  )

  echo "==> $pkg_name: go test"
  (
    cd "$pkg_dir"
    go test ./...
  )
}

if [[ $# -gt 0 ]]; then
  target="$repo_root/$1"
  if [[ ! -d "$target" ]]; then
    echo "error: '$1' not found under $repo_root" >&2
    exit 1
  fi
  validate_package "$target"
  exit 0
fi

# No argument: validate all modules/, functions/, and types/
for base_dir in "$repo_root/modules" "$repo_root/functions" "$repo_root/types"; do
  [[ -d "$base_dir" ]] || continue
  while IFS= read -r -d '' dir; do
    validate_package "$dir"
  done < <(find "$base_dir" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)
done
