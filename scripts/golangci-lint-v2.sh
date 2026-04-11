#!/usr/bin/env bash
set -euo pipefail

VERSION="${GOLANGCI_LINT_VERSION:-v2.11.4}"
DISPLAY_VERSION="${VERSION#v}"
MODULE="github.com/golangci/golangci-lint/v2/cmd/golangci-lint"

if [[ -n "${GOLANGCI_LINT_BIN:-}" ]]; then
  exec "$GOLANGCI_LINT_BIN" "$@"
fi

if command -v golangci-lint >/dev/null 2>&1 && golangci-lint version 2>/dev/null | grep -Eq "version v?${DISPLAY_VERSION}([ ,]|$)"; then
  exec golangci-lint "$@"
fi

if ! command -v go >/dev/null 2>&1; then
  echo "golangci-lint ${VERSION} is required and go is not installed to bootstrap it" >&2
  exit 127
fi

cache_root="${GOLANGCI_LINT_CACHE_BIN:-${XDG_CACHE_HOME:-$HOME/.cache}/agentops/golangci-lint}"
bin_dir="${cache_root}/${VERSION}"
bin="${bin_dir}/golangci-lint"

if [[ ! -x "$bin" ]]; then
  mkdir -p "$bin_dir"
  GOBIN="$bin_dir" go install "${MODULE}@${VERSION}"
fi

exec "$bin" "$@"
