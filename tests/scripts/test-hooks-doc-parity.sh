#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-hooks-doc-parity.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

if [[ ! -x "$SCRIPT" ]]; then
  echo "FAIL: missing executable script: $SCRIPT" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

write_manifest() {
  local path="$1"
  cat > "$path" <<'JSON'
{
  "hooks": {
    "SessionStart": [],
    "SessionEnd": [],
    "Stop": []
  }
}
JSON
}

test_pass_when_no_stale_phrases() {
  local fixture="$TMP_DIR/pass"
  mkdir -p "$fixture/docs"
  write_manifest "$fixture/hooks.json"

  cat > "$fixture/AGENTS.md" <<'EOF'
All pushes run validate.yml.
EOF
  cat > "$fixture/docs/a.md" <<'EOF'
Hooks are governed by hooks/hooks.json runtime contract.
EOF

  if HOOKS_DOC_PARITY_MANIFEST="$fixture/hooks.json" \
     HOOKS_DOC_PARITY_FILES="$fixture/AGENTS.md $fixture/docs/a.md" \
     bash "$SCRIPT" > "$fixture/out.txt" 2>&1; then
    pass "passes when scoped files have no stale hook-count phrases"
  else
    fail "should pass when no stale hook-count phrases exist"
    sed 's/^/  /' "$fixture/out.txt"
  fi
}

test_fail_with_actionable_diff() {
  local fixture="$TMP_DIR/fail"
  mkdir -p "$fixture/docs"
  write_manifest "$fixture/hooks.json"

  cat > "$fixture/docs/b.md" <<'EOF'
AgentOps has 12 hooks that enforce governance.
EOF

  if HOOKS_DOC_PARITY_MANIFEST="$fixture/hooks.json" \
     HOOKS_DOC_PARITY_FILES="$fixture/docs/b.md" \
     bash "$SCRIPT" > "$fixture/out.txt" 2>&1; then
    fail "should fail when stale hook-count phrase exists"
    return
  fi

  if grep -q "HOOKS_DOC_PARITY: drift detected" "$fixture/out.txt"; then
    pass "reports drift header"
  else
    fail "missing drift header in failure output"
  fi

  if grep -q -- "- AgentOps has 12 hooks that enforce governance." "$fixture/out.txt" \
     && grep -q -- "+ AgentOps has 3 active hooks that enforce governance." "$fixture/out.txt"; then
    pass "prints diff-style replacement hint"
  else
    fail "missing actionable diff hint in failure output"
  fi
}

echo "== test-hooks-doc-parity =="
test_pass_when_no_stale_phrases
test_fail_with_actionable_diff

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
