#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
GATE_SCRIPT="$REPO_ROOT/scripts/validate-learning-coherence.sh"

PASS_COUNT=0
FAIL_COUNT=0
TMP_DIR="$(mktemp -d)"

# shellcheck disable=SC2329
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

pass() {
  echo "PASS: $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

fail() {
  echo "FAIL: $1"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

run_gate() {
  local learnings_dir="$1"
  local output_file="$2"

  set +e
  bash "$GATE_SCRIPT" "$learnings_dir" >"$output_file" 2>&1
  local status=$?
  set -e
  return "$status"
}

assert_gate_passes() {
  local description="$1"
  local learnings_dir="$2"
  local output_file
  output_file="$(mktemp "$TMP_DIR/pass-XXXXXX")"

  if run_gate "$learnings_dir" "$output_file"; then
    pass "$description"
  else
    echo "--- output ($description) ---"
    cat "$output_file"
    fail "$description (expected pass)"
  fi
}

assert_gate_fails_with() {
  local description="$1"
  local learnings_dir="$2"
  local expected="$3"
  local output_file
  output_file="$(mktemp "$TMP_DIR/fail-XXXXXX")"

  if run_gate "$learnings_dir" "$output_file"; then
    echo "--- output ($description) ---"
    cat "$output_file"
    fail "$description (expected failure)"
    return
  fi

  if grep -Fq "$expected" "$output_file"; then
    pass "$description"
  else
    echo "--- output ($description) ---"
    cat "$output_file"
    fail "$description (missing expected text: $expected)"
  fi
}

make_fixture_dir() {
  local name="$1"
  local dir="$TMP_DIR/$name/.agents/learnings"
  mkdir -p "$dir"
  echo "$dir"
}

test_script_executable() {
  if [[ -x "$GATE_SCRIPT" ]]; then
    pass "validate-learning-coherence.sh is executable"
  else
    fail "validate-learning-coherence.sh is not executable"
  fi
}

test_frontmatter_only_memrl_fails() {
  local dir
  dir="$(make_fixture_dir frontmatter-only)"
  cat >"$dir/2026-03-08-meta-only.md" <<'EOF'
---
utility: 0.80
confidence: 0.90
maturity: established
---
EOF

  assert_gate_fails_with \
    "frontmatter-only MemRL learning fails" \
    "$dir" \
    "frontmatter-only learning"
}

test_manual_frontmatter_only_fails() {
  local dir
  dir="$(make_fixture_dir manual-frontmatter-only)"
  cat >"$dir/2026-03-08-meta-only.md" <<'EOF'
---
id: learning-123
date: 2026-03-07
---
EOF

  assert_gate_fails_with \
    "frontmatter-only manual learning fails" \
    "$dir" \
    "frontmatter-only learning"
}

test_memrl_with_body_passes() {
  local dir
  dir="$(make_fixture_dir memrl-body)"
  cat >"$dir/2026-03-08-valid.md" <<'EOF'
---
utility: 0.72
confidence: 0.80
maturity: candidate
---
# Learning: Event mirroring preserves observability

When worktree runs emit events to a single root, dashboards and status tools can drift.
Mirror event writes across artifact roots so control-plane reads stay consistent.
EOF

  assert_gate_passes "memrl learning with body passes" "$dir"
}

test_non_learning_markdown_is_ignored() {
  local dir
  dir="$(make_fixture_dir non-learning-doc)"
  cat >"$dir/AGENTS.md" <<'EOF'
# Learnings

This helper file should not be treated as a learning artifact by the gate.
EOF
  cat >"$dir/2026-03-08-valid-learning.md" <<'EOF'
---
utility: 0.72
confidence: 0.80
maturity: candidate
---
# Learning: Ignore helper docs during coherence scans

Only date-prefixed learning artifacts should participate in coherence validation.
EOF

  assert_gate_passes "non-learning markdown is ignored" "$dir"
}

test_frontmatter_without_recognized_fields_fails() {
  local dir
  dir="$(make_fixture_dir missing-fields)"
  cat >"$dir/2026-03-08-bad-frontmatter.md" <<'EOF'
---
foo: bar
bar: baz
---
# Learning

This has body text but no recognized frontmatter fields.
EOF

  assert_gate_fails_with \
    "frontmatter without recognized fields fails" \
    "$dir" \
    "frontmatter has no recognized fields"
}

test_non_learning_markdown_is_ignored() {
  local dir
  dir="$(make_fixture_dir non-learning-doc)"
  cat >"$dir/AGENTS.md" <<'EOF'
# Learnings

Helper documentation file that should not be treated as a learning artifact.
EOF

  assert_gate_passes "non-learning markdown is ignored" "$dir"
}

echo "================================"
echo "Testing learning coherence gate"
echo "================================"
echo ""

test_script_executable
test_frontmatter_only_memrl_fails
test_manual_frontmatter_only_fails
test_memrl_with_body_passes
test_non_learning_markdown_is_ignored
test_frontmatter_without_recognized_fields_fails
test_non_learning_markdown_is_ignored

echo ""
echo "================================"
echo "Results: $PASS_COUNT PASS, $FAIL_COUNT FAIL"
echo "================================"

if [[ $FAIL_COUNT -gt 0 ]]; then
  exit 1
fi
exit 0
