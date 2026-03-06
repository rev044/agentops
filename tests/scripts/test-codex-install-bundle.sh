#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/validate-codex-install-bundle.sh"

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
  local archived_body="$2"
  local generated_body="$3"

  mkdir -p \
    "$fixture/scripts" \
    "$fixture/skills/source-skill" \
    "$fixture/skills-codex/source-skill"

  cp "$SCRIPT" "$fixture/scripts/validate-codex-install-bundle.sh"

  cat > "$fixture/scripts/sync-codex-native-skills.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
OUT=""
while [[ \$# -gt 0 ]]; do
  case "\$1" in
    --out)
      OUT="\${2:-}"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done
[[ -n "\$OUT" ]] || exit 1
mkdir -p "\$OUT/source-skill"
cat > "\$OUT/source-skill/SKILL.md" <<'INNER'
$generated_body
INNER
EOF
  chmod +x "$fixture/scripts/sync-codex-native-skills.sh"

  cat > "$fixture/skills/source-skill/SKILL.md" <<'EOF'
---
name: source-skill
description: fixture
---
EOF

  cat > "$fixture/skills-codex/source-skill/SKILL.md" <<EOF
$archived_body
EOF

  (
    cd "$fixture"
    git init >/dev/null 2>&1
    git config user.name "Codex Test"
    git config user.email "codex@example.com"
    git add .
    git commit -m "fixture" >/dev/null 2>&1
  )
}

run_fixture() {
  local fixture="$1"
  local out_file="$2"

  (
    cd "$fixture"
    bash scripts/validate-codex-install-bundle.sh
  ) > "$out_file" 2>&1
}

test_pass_with_matching_bundle() {
  local fixture="$TMP_DIR/pass"
  local out="$fixture/out.txt"
  local body='---
name: source-skill
description: generated
---

# Source Skill

Bundle matches generated output.'

  setup_fixture "$fixture" "$body" "$body"

  if run_fixture "$fixture" "$out"; then
    pass "passes when archived bundle matches regenerated output"
  else
    fail "should pass when archived bundle matches regenerated output"
    sed 's/^/  /' "$out"
  fi
}

test_fail_with_stale_bundle() {
  local fixture="$TMP_DIR/fail"
  local out="$fixture/out.txt"
  local archived_body='---
name: source-skill
description: stale
---

# Source Skill

This archive is stale.'
  local generated_body='---
name: source-skill
description: current
---

# Source Skill

This output is current.'

  setup_fixture "$fixture" "$archived_body" "$generated_body"

  if run_fixture "$fixture" "$out"; then
    fail "should fail when archived bundle is stale"
    return
  fi

  if grep -q "Codex install bundle drift detected" "$out"; then
    pass "fails when archived bundle is stale"
  else
    fail "missing stale bundle error"
    sed 's/^/  /' "$out"
  fi
}

echo "== test-codex-install-bundle =="
test_pass_with_matching_bundle
test_fail_with_stale_bundle

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
