#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-codex-runtime-sections.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

if [[ ! -f "$SCRIPT" ]]; then
  echo "FAIL: missing script: $SCRIPT" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

setup_fixture() {
  local fixture="$1"
  local allowlist="$2"

  mkdir -p "$fixture/scripts/lint" "$fixture/skills-codex/fixture"
  cp "$SCRIPT" "$fixture/scripts/validate-codex-runtime-sections.sh"

  cat > "$fixture/scripts/lint/codex-residual-allowlist.txt" <<EOF
$allowlist
EOF
}

write_skill() {
  local path="$1"
  local body="$2"

  cat > "$path" <<EOF
---
name: fixture
description: fixture
---

# Runtime Setup

$body
EOF
}

run_fixture() {
  local fixture="$1"
  local out_file="$2"

  (
    cd "$fixture"
    bash scripts/validate-codex-runtime-sections.sh
  ) > "$out_file" 2>&1
}

test_pass_with_clean_fixture() {
  local fixture="$TMP_DIR/pass-clean"
  local out="$fixture/out.txt"

  setup_fixture "$fixture" "# empty allowlist for clean fixture"
  write_skill "$fixture/skills-codex/fixture/SKILL.md" \
    "Use Codex-only runtime instructions in this section."

  if run_fixture "$fixture" "$out"; then
    pass "passes with clean fixture"
  else
    fail "should pass with clean fixture"
    sed 's/^/  /' "$out"
  fi
}

test_fail_on_duplicate_runtime_setup_headings() {
  local fixture="$TMP_DIR/fail-duplicate-runtime-setup"
  local out="$fixture/out.txt"

  setup_fixture "$fixture" "# no residual markers allowlisted"
  cat > "$fixture/skills-codex/fixture/SKILL.md" <<'EOF'
---
name: fixture
description: fixture
---

# Runtime Setup

Primary setup details.

## Runtime Setup

Duplicated setup details.
EOF

  if run_fixture "$fixture" "$out"; then
    fail "should fail on duplicate runtime setup headings"
    return
  fi

  if grep -q "duplicate runtime setup section" "$out"; then
    pass "fails on duplicate runtime setup headings"
  else
    fail "missing duplicate runtime setup error"
  fi
}

test_allowlisted_marker_accepted() {
  local fixture="$TMP_DIR/pass-allowlisted-marker"
  local out="$fixture/out.txt"

  setup_fixture "$fixture" '\bteam-create\b'
  write_skill "$fixture/skills-codex/fixture/SKILL.md" \
    "Use team-create for orchestrated task dispatch."

  if run_fixture "$fixture" "$out"; then
    pass "accepts allowlisted marker"
  else
    fail "should accept allowlisted marker"
    sed 's/^/  /' "$out"
  fi
}

test_non_allowlisted_marker_fails() {
  local fixture="$TMP_DIR/fail-non-allowlisted-marker"
  local out="$fixture/out.txt"

  setup_fixture "$fixture" "# intentionally empty allowlist"
  write_skill "$fixture/skills-codex/fixture/SKILL.md" \
    "Anthropic runtime references should be rejected."

  if run_fixture "$fixture" "$out"; then
    fail "should fail on non-allowlisted marker"
    return
  fi

  if grep -q "residual mixed-runtime marker found" "$out"; then
    pass "fails on non-allowlisted marker"
  else
    fail "missing non-allowlisted marker error"
  fi
}

echo "== test-codex-runtime-sections =="
test_pass_with_clean_fixture
test_fail_on_duplicate_runtime_setup_headings
test_allowlisted_marker_accepted
test_non_allowlisted_marker_fails

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
