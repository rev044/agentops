#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SYNC_SCRIPT="$ROOT/scripts/sync-codex-native-skills.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

[[ -x "$SYNC_SCRIPT" ]] || {
  echo "FAIL: missing script: $SYNC_SCRIPT" >&2
  exit 1
}

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

OVERRIDES_DIR="$TMP_DIR/overrides"
OUT_DIR="$TMP_DIR/skills-codex"

mkdir -p "$OVERRIDES_DIR/research"
cp "$ROOT/skills-codex-overrides/research/prompt.md" "$OVERRIDES_DIR/research/prompt.md"
cat > "$OVERRIDES_DIR/catalog.json" <<'EOF'
{
  "version": 1,
  "waves": [
    {"id": "backbone", "description": "fixture"}
  ],
  "skills": [
    {"name": "research", "treatment": "bespoke", "wave": "backbone", "reason": "fixture"}
  ]
}
EOF

if bash "$SYNC_SCRIPT" --src "$ROOT/skills" --out "$OUT_DIR" --overrides "$OVERRIDES_DIR" --only research >/dev/null; then
  pass "sync completes with a fixture catalog"
else
  fail "sync should succeed with a fixture catalog"
fi

MANIFEST="$OUT_DIR/.agentops-manifest.json"
if jq -e '.codex_override_catalog.skills[0].name == "research"' "$MANIFEST" >/dev/null; then
  pass "generated manifest embeds codex override catalog"
else
  fail "generated manifest should embed codex override catalog"
fi

if jq -e '.codex_override_catalog_hash | strings | length > 0' "$MANIFEST" >/dev/null; then
  pass "generated manifest includes catalog hash"
else
  fail "generated manifest should include catalog hash"
fi

echo
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
