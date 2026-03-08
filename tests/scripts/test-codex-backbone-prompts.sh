#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-codex-backbone-prompts.sh"

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
  local body="$3"
  mkdir -p "$(dirname "$path")"
  cat > "$path" <<EOF
# $name

$body
EOF
}

setup_fixture() {
  local fixture="$1"
  mkdir -p \
    "$fixture/skills-codex/alpha" \
    "$fixture/skills-codex/beta" \
    "$fixture/skills-codex-overrides"

  write_prompt "$fixture/skills-codex/alpha/prompt.md" "alpha" "Alpha prompt.

## Codex Execution Profile

1. Route issues into downstream execution.
2. Keep handoff artifacts durable.

## Guardrails

1. Prefer concise, exact execution notes.
2. Keep behavior findings-first."

  write_prompt "$fixture/skills-codex/beta/prompt.md" "beta" "Beta prompt.

## Codex Execution Profile

1. Report exact validations run.
2. Require remote verification.

## Guardrails

1. Do not skip failing gates.
2. Recover until push succeeds."

  cat > "$fixture/skills-codex-overrides/catalog.json" <<'EOF'
{
  "version": 1,
  "skills": [
    {
      "name": "alpha",
      "treatment": "bespoke",
      "wave": "backbone",
      "reason": "fixture",
      "operator_contract": {
        "required_sections": ["## Codex Execution Profile", "## Guardrails"],
        "required_markers": ["Route issues into downstream execution.", "Keep behavior findings-first."]
      }
    },
    {
      "name": "beta",
      "treatment": "bespoke",
      "wave": "backbone",
      "reason": "fixture",
      "operator_contract": {
        "required_sections": ["## Codex Execution Profile", "## Guardrails"],
        "required_markers": ["Report exact validations run.", "Recover until push succeeds."]
      }
    }
  ]
}
EOF
}

test_fixture_passes() {
  local fixture="$TMP_DIR/pass"
  setup_fixture "$fixture"

  if bash "$SCRIPT" --repo-root "$fixture" >/dev/null; then
    pass "passes when generated backbone prompts satisfy the catalog contract"
  else
    fail "should validate a matching backbone prompt fixture"
  fi
}

test_fails_when_marker_missing() {
  local fixture="$TMP_DIR/missing-marker"
  setup_fixture "$fixture"
  python3 - <<'PY' "$fixture/skills-codex/alpha/prompt.md"
from pathlib import Path
path = Path(__import__("sys").argv[1])
path.write_text(path.read_text().replace("Route issues into downstream execution.\n", ""))
PY

  if bash "$SCRIPT" --repo-root "$fixture" >/dev/null 2>&1; then
    fail "should fail when a required generated-prompt marker is missing"
  else
    pass "fails when a required generated-prompt marker is missing"
  fi
}

test_fails_when_sections_out_of_order() {
  local fixture="$TMP_DIR/section-order"
  setup_fixture "$fixture"
  python3 - <<'PY' "$fixture/skills-codex/beta/prompt.md"
from pathlib import Path
path = Path(__import__("sys").argv[1])
path.write_text("""# beta

Beta prompt.

## Guardrails

1. Do not skip failing gates.
2. Recover until push succeeds.

## Codex Execution Profile

1. Report exact validations run.
2. Require remote verification.
""")
PY

  if bash "$SCRIPT" --repo-root "$fixture" >/dev/null 2>&1; then
    fail "should fail when required sections are out of order"
  else
    pass "fails when required sections are out of order"
  fi
}

test_repo_catalog_passes() {
  if bash "$SCRIPT" --repo-root "$ROOT" >/dev/null 2>&1; then
    pass "repository backbone prompts validate end to end"
  else
    fail "repository backbone prompts should validate end to end"
  fi
}

echo "== test-codex-backbone-prompts =="
test_fixture_passes
test_fails_when_marker_missing
test_fails_when_sections_out_of_order
test_repo_catalog_passes

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
