#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-codex-override-coverage.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

[[ -x "$SCRIPT" ]] || {
  echo "FAIL: missing script: $SCRIPT" >&2
  exit 1
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

write_prompt() {
  local path="$1"
  local name="$2"
  cat > "$path" <<EOF
# $name

Codex-native prompt for $name.

## Codex Execution Profile

1. Treat the canonical skill as source of truth.

## Guardrails

1. Keep behavior machine-checkable.
EOF
}

setup_fixture() {
  local fixture="$1"
  mkdir -p \
    "$fixture/skills/alpha" \
    "$fixture/skills/beta" \
    "$fixture/skills/gamma" \
    "$fixture/skills-codex/alpha" \
    "$fixture/skills-codex/beta" \
    "$fixture/skills-codex/gamma" \
    "$fixture/skills-codex-overrides/alpha"

  for skill in alpha beta gamma; do
    cat > "$fixture/skills/$skill/SKILL.md" <<EOF
---
name: $skill
description: fixture
---
EOF
    cat > "$fixture/skills-codex/$skill/SKILL.md" <<EOF
---
name: $skill
description: fixture
---
EOF
  done

  write_prompt "$fixture/skills-codex-overrides/alpha/prompt.md" "alpha"
  cp "$fixture/skills-codex-overrides/alpha/prompt.md" "$fixture/skills-codex/alpha/prompt.md"

  cat > "$fixture/skills-codex/beta/prompt.md" <<'EOF'
# beta

## Instructions

Load and follow the skill instructions from the sibling `SKILL.md` file for this skill.
EOF

  cat > "$fixture/skills-codex/gamma/prompt.md" <<'EOF'
# gamma

## Instructions

Load and follow the skill instructions from the sibling `SKILL.md` file for this skill.
EOF

  cat > "$fixture/skills-codex-overrides/catalog.json" <<'EOF'
{
  "version": 1,
  "waves": [
    {"id": "wave-a", "description": "fixture"},
    {"id": "wave-b", "description": "fixture"}
  ],
  "skills": [
    {"name": "alpha", "treatment": "bespoke", "wave": "wave-a", "reason": "Needs Codex-native wording."},
    {"name": "beta", "treatment": "parity_only", "wave": "wave-b", "reason": "Default generated prompt is enough."},
    {"name": "gamma", "treatment": "bespoke", "wave": "wave-b", "reason": "Needs Codex-native review structure."}
  ]
}
EOF
}

test_fixture_passes_with_complete_wave_filter() {
  local fixture="$TMP_DIR/pass"
  setup_fixture "$fixture"
  mkdir -p "$fixture/skills-codex-overrides/gamma"
  write_prompt "$fixture/skills-codex-overrides/gamma/prompt.md" "gamma"
  cp "$fixture/skills-codex-overrides/gamma/prompt.md" "$fixture/skills-codex/gamma/prompt.md"

  if bash "$SCRIPT" --repo-root "$fixture" --wave wave-a >/dev/null; then
    pass "supports repo-root fixtures and wave filtering"
  else
    fail "should validate a filtered wave in a fixture repo"
  fi
}

test_fails_when_bespoke_override_missing() {
  local fixture="$TMP_DIR/missing"
  setup_fixture "$fixture"

  if bash "$SCRIPT" --repo-root "$fixture" >/dev/null 2>&1; then
    fail "should fail when a bespoke catalog skill lacks an override prompt"
  else
    pass "fails when bespoke override prompt is missing"
  fi
}

test_fails_when_parity_skill_has_override() {
  local fixture="$TMP_DIR/parity"
  setup_fixture "$fixture"
  mkdir -p "$fixture/skills-codex-overrides/beta"
  write_prompt "$fixture/skills-codex-overrides/beta/prompt.md" "beta"
  mkdir -p "$fixture/skills-codex-overrides/gamma"
  write_prompt "$fixture/skills-codex-overrides/gamma/prompt.md" "gamma"
  cp "$fixture/skills-codex-overrides/gamma/prompt.md" "$fixture/skills-codex/gamma/prompt.md"

  if bash "$SCRIPT" --repo-root "$fixture" >/dev/null 2>&1; then
    fail "should fail when a parity-only skill has a prompt override"
  else
    pass "fails when parity-only skill has an unexpected override"
  fi
}

test_repo_catalog_is_complete() {
  if bash "$SCRIPT" --repo-root "$ROOT" >/dev/null 2>&1; then
    pass "repository catalog validates end to end"
  else
    fail "repository catalog should validate end to end"
  fi
}

echo "== test-codex-override-coverage =="
test_fixture_passes_with_complete_wave_filter
test_fails_when_bespoke_override_missing
test_fails_when_parity_skill_has_override
test_repo_catalog_is_complete

echo
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
