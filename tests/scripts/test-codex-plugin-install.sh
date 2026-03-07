#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/scripts/install-codex-plugin.sh"

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

  mkdir -p \
    "$fixture/.codex-plugin" \
    "$fixture/.agents/plugins" \
    "$fixture/skills-codex/research" \
    "$fixture/skills-codex/heal-skill/scripts"

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

  cat > "$fixture/skills-codex/research/SKILL.md" <<'EOF'
---
name: research
description: fixture
---
EOF

  cat > "$fixture/skills-codex/research/prompt.md" <<'EOF'
# research
EOF

  cat > "$fixture/skills-codex/heal-skill/SKILL.md" <<'EOF'
---
name: heal-skill
description: fixture
---
EOF

  cat > "$fixture/skills-codex/heal-skill/prompt.md" <<'EOF'
# heal-skill
EOF

  cat > "$fixture/skills-codex/heal-skill/scripts/heal.sh" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
  chmod +x "$fixture/skills-codex/heal-skill/scripts/heal.sh"

  cat > "$fixture/skills-codex/research/.agentops-generated.json" <<'EOF'
{
  "generator": "scripts/sync-codex-native-skills.sh",
  "source_skill": "skills/research",
  "layout": "modular",
  "source_hash": "fixture-source",
  "generated_hash": "fixture-generated"
}
EOF

  cat > "$fixture/skills-codex/heal-skill/.agentops-generated.json" <<'EOF'
{
  "generator": "scripts/sync-codex-native-skills.sh",
  "source_skill": "skills/heal-skill",
  "layout": "modular",
  "source_hash": "fixture-source",
  "generated_hash": "fixture-generated"
}
EOF

  cat > "$fixture/skills-codex/.agentops-manifest.json" <<'EOF'
{
  "generator": "scripts/sync-codex-native-skills.sh",
  "source_root": "skills",
  "layout": "modular",
  "skills": [
    {
      "name": "heal-skill",
      "source_skill": "skills/heal-skill",
      "source_hash": "fixture-source",
      "generated_hash": "fixture-generated"
    },
    {
      "name": "research",
      "source_skill": "skills/research",
      "source_hash": "fixture-source",
      "generated_hash": "fixture-generated"
    }
  ]
}
EOF
}

run_install() {
  local fixture="$1"
  local codex_home="$2"

  bash "$SCRIPT" \
    --repo-root "$fixture" \
    --codex-home "$codex_home" \
    --version "fixture-version" \
    --update-command "fixture-update" >/dev/null
}

test_installs_plugin_cache_and_config() {
  local fixture="$TMP_DIR/install"
  local codex_home="$TMP_DIR/codex-home"
  local plugin_root="$codex_home/plugins/cache/agentops-marketplace/agentops/local"
  local codex_skills_root="$codex_home/skills"
  local user_skills_root="$TMP_DIR/.agents/skills"

  setup_fixture "$fixture"

  if ! run_install "$fixture" "$codex_home"; then
    fail "install should succeed"
    return
  fi

  [[ -f "$plugin_root/.codex-plugin/plugin.json" ]] || {
    fail "plugin manifest copied into cache"
    return
  }
  [[ -f "$plugin_root/skills-codex/research/SKILL.md" ]] || {
    fail "skills copied into plugin cache"
    return
  }
  [[ -f "$codex_skills_root/research/SKILL.md" ]] || {
    fail "skills mirrored into CODEX_HOME/skills"
    return
  }
  [[ -f "$user_skills_root/research/SKILL.md" ]] || {
    fail "skills mirrored into ~/.agents/skills sibling"
    return
  }
  [[ -f "$plugin_root/.agentops-codex-state.json" ]] || {
    fail "plugin state file written"
    return
  }
  [[ -f "$codex_home/config.toml" ]] || {
    fail "config.toml written"
    return
  }
  if rg -q '^\[features\]$' "$codex_home/config.toml" && \
    rg -q '^plugins = true$' "$codex_home/config.toml" && \
    rg -q '^\[plugins\."agentops@agentops-marketplace"\]$' "$codex_home/config.toml" && \
    rg -q '^enabled = true$' "$codex_home/config.toml" && \
    rg -q "\"codex_home_skills_root\": \"$codex_skills_root\"" "$codex_home/.agentops-codex-install.json" && \
    rg -q "\"user_skills_root\": \"$user_skills_root\"" "$codex_home/.agentops-codex-install.json"; then
    pass "installs plugin cache and mirrors managed raw skills"
  else
    fail "config.toml missing plugin enablement"
  fi
}

test_updates_agentops_raw_skills_without_touching_custom_skills() {
  local fixture="$TMP_DIR/archive"
  local codex_home="$TMP_DIR/archive-home"
  local legacy_skills="$codex_home/skills"

  setup_fixture "$fixture"
  mkdir -p "$legacy_skills/research" "$legacy_skills/custom-skill"
  cat > "$legacy_skills/research/SKILL.md" <<'EOF'
legacy
EOF
  cat > "$legacy_skills/custom-skill/SKILL.md" <<'EOF'
custom
EOF

  if ! run_install "$fixture" "$codex_home"; then
    fail "install with legacy raw skills should succeed"
    return
  fi

  if ! rg -q '^name: research$' "$legacy_skills/research/SKILL.md"; then
    fail "managed AgentOps raw skill should be refreshed in ~/.codex/skills"
    return
  fi
  if [[ ! -e "$legacy_skills/custom-skill/SKILL.md" ]]; then
    fail "non-AgentOps custom skill should remain in ~/.codex/skills"
    return
  fi
  pass "refreshes AgentOps raw skills without touching custom skill folders"
}

test_refreshes_home_agents_agentops_skills_without_touching_custom_skills() {
  local fixture="$TMP_DIR/home-agents"
  local home_dir="$TMP_DIR/home-agents-root"
  local codex_home="$home_dir/.codex"
  local user_skills_root="$home_dir/.agents/skills"

  setup_fixture "$fixture"
  mkdir -p "$user_skills_root/research" "$user_skills_root/custom-skill"
  cat > "$user_skills_root/research/SKILL.md" <<'EOF'
official
EOF
  cat > "$user_skills_root/custom-skill/SKILL.md" <<'EOF'
custom
EOF

  if ! HOME="$home_dir" run_install "$fixture" "$codex_home"; then
    fail "install with ~/.agents/skills should succeed"
    return
  fi

  if ! rg -q '^name: research$' "$user_skills_root/research/SKILL.md"; then
    fail "plugin install should refresh managed AgentOps skills in ~/.agents/skills"
    return
  fi
  if [[ ! -f "$user_skills_root/custom-skill/SKILL.md" ]]; then
    fail "plugin install should not remove custom ~/.agents/skills folders"
    return
  fi
  pass "plugin install refreshes AgentOps skills without touching custom ~/.agents/skills folders"
}

echo "== test-codex-plugin-install =="
test_installs_plugin_cache_and_config
test_updates_agentops_raw_skills_without_touching_custom_skills
test_refreshes_home_agents_agentops_skills_without_touching_custom_skills

echo ""
echo "Results: $PASS PASS, $FAIL FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  exit 1
fi
exit 0
