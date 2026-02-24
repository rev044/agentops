#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

if ! command -v bats >/dev/null 2>&1; then
  echo "SKIP: bats not installed; skipping tests/hooks/constraint-compiler.bats"
  exit 0
fi

bats tests/hooks/constraint-compiler.bats
