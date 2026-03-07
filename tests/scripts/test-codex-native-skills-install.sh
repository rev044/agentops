#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
NATIVE_SCRIPT="$ROOT/scripts/install-codex-native-skills.sh"
PLUGIN_SCRIPT="$ROOT/scripts/install-codex-plugin.sh"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

if [[ ! -f "$NATIVE_SCRIPT" ]]; then
  echo "FAIL: missing script: $NATIVE_SCRIPT" >&2
  exit 1
fi
if [[ ! -f "$PLUGIN_SCRIPT" ]]; then
  echo "FAIL: missing script: $PLUGIN_SCRIPT" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

setup_fixture() {
  local fixture="$1"

  mkdir -p \
    "$fixture/scripts" \
    "$fixture/.codex-plugin" \
    "$fixture/.agents/plugins" \
    "$fixture/skills" \
    "$fixture/skills-codex"

  cp "$NATIVE_SCRIPT" "$fixture/scripts/install-codex-native-skills.sh"
  cp "$PLUGIN_SCRIPT" "$fixture/scripts/install-codex-plugin.sh"
  chmod +x \
    "$fixture/scripts/install-codex-native-skills.sh" \
    "$fixture/scripts/install-codex-plugin.sh"

  cat > "$fixture/scripts/sync-codex-native-skills.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
OUT=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)
      OUT="${2:-}"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done
[[ -n "$OUT" ]] || exit 1
mkdir -p "$OUT/source-skill"
cat > "$OUT/source-skill/SKILL.md" <<'INNER'
---
name: source-skill
description: fixture
---
INNER
cat > "$OUT/source-skill/prompt.md" <<'INNER'
# source-skill
INNER
EOF

  cat > "$fixture/scripts/export-claude-skills-to-codex.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
SRC=""
DST=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --src)
      SRC="${2:-}"
      shift 2
      ;;
    --dst)
      DST="${2:-}"
      shift 2
      ;;
    --backup|--only)
      shift 2
      ;;
    --dry-run)
      shift
      ;;
    *)
      shift
      ;;
  esac
done
[[ -n "$SRC" && -n "$DST" ]] || exit 1
mkdir -p "$DST"
cp -R "$SRC/." "$DST/"
EOF
  chmod +x \
    "$fixture/scripts/sync-codex-native-skills.sh" \
    "$fixture/scripts/export-claude-skills-to-codex.sh"

  cat > "$fixture/.codex-plugin/plugin.json" <<'EOF'
{
  "name": "agentops",
  "skills": "./skills-codex"
}
EOF

  cat > "$fixture/.agents/plugins/marketplace.json" <<'EOF'
{
  "name": "agentops-marketplace",
  "plugins": [
    {
      "name": "agentops",
      "source": {
        "source": "local",
        "path": "./"
      }
    }
  ]
}
EOF
}

test_native_plugin_metadata_refreshes_plugin_cache() {
  local fixture="$TMP_DIR/plugin-mode"
  local home_dir="$TMP_DIR/home-plugin"
  local plugin_root="$home_dir/.codex/plugins/cache/agentops-marketplace/agentops/local"

  setup_fixture "$fixture"
  mkdir -p "$home_dir/.codex"
  cat > "$home_dir/.codex/.agentops-codex-install.json" <<EOF
{
  "install_mode": "native-plugin",
  "plugin_root": "$plugin_root"
}
EOF

  if ! HOME="$home_dir" bash "$fixture/scripts/install-codex-native-skills.sh" >/dev/null; then
    fail "native-plugin refresh should succeed"
    return
  fi

  if [[ ! -f "$plugin_root/skills-codex/source-skill/SKILL.md" ]]; then
    fail "native-plugin refresh should write skills into plugin cache"
    return
  fi
  if [[ -e "$home_dir/.codex/skills/source-skill/SKILL.md" ]]; then
    fail "native-plugin refresh should not write raw skills into ~/.codex/skills"
    return
  fi
  if rg -q '"source": "install-codex-plugin.sh"' "$home_dir/.codex/.agentops-codex-install.json" && \
    rg -q '"install_mode": "native-plugin"' "$home_dir/.codex/.agentops-codex-install.json"; then
    pass "native-plugin metadata refreshes active plugin cache"
  else
    fail "native-plugin refresh should preserve plugin install metadata"
  fi
}

test_explicit_dest_keeps_raw_skill_install_behavior() {
  local fixture="$TMP_DIR/raw-mode"
  local home_dir="$TMP_DIR/home-raw"
  local raw_dst="$home_dir/custom-skills"
  local plugin_root="$home_dir/.codex/plugins/cache/agentops-marketplace/agentops/local"

  setup_fixture "$fixture"
  mkdir -p "$home_dir/.codex"
  cat > "$home_dir/.codex/.agentops-codex-install.json" <<EOF
{
  "install_mode": "native-plugin",
  "plugin_root": "$plugin_root"
}
EOF

  if ! HOME="$home_dir" bash "$fixture/scripts/install-codex-native-skills.sh" --dest "$raw_dst" >/dev/null; then
    fail "explicit --dest raw install should succeed"
    return
  fi

  if [[ ! -f "$raw_dst/source-skill/SKILL.md" ]]; then
    fail "explicit --dest should install raw skills into requested destination"
    return
  fi
  if [[ -e "$plugin_root/skills-codex/source-skill/SKILL.md" ]]; then
    fail "explicit --dest should not refresh plugin cache"
    return
  fi
  if rg -q '"source": "install-codex-native-skills.sh"' "$home_dir/.codex/.agentops-codex-install.json"; then
    pass "explicit --dest preserves raw skill install behavior"
  else
    fail "explicit --dest should record raw skill install metadata"
  fi
}

echo "== test-codex-native-skills-install =="
test_native_plugin_metadata_refreshes_plugin_cache
test_explicit_dest_keeps_raw_skill_install_behavior

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
