#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-local.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

assert_contains() {
  local pattern="$1"
  local label="$2"

  if rg -q "$pattern" "$SCRIPT"; then
    pass "$label"
  else
    fail "$label"
  fi
}

echo "== test-validate-local =="
assert_contains 'install-dev-hooks\.sh' "validate-local references repo hook bootstrap"
assert_contains 'core\.hooksPath' "validate-local warns when core.hooksPath is not .githooks"
assert_contains 'Manual Local Validation' "validate-local describes itself as manual validation"
assert_contains 'pre-push-gate\.sh' "validate-local wraps the shared push gate"

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
