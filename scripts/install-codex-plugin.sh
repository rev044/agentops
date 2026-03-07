#!/usr/bin/env bash
# install-codex-plugin.sh — Install the AgentOps native Codex plugin into CODEX_HOME.
#
# Usage:
#   bash scripts/install-codex-plugin.sh
#   bash scripts/install-codex-plugin.sh --repo-root /path/to/agentops --codex-home /tmp/codex-home

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}!${NC} $*"; }
fail()  { echo -e "${RED}✗${NC} $*"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CODEX_HOME="${HOME}/.codex"
PLUGIN_NAME="agentops"
MARKETPLACE_NAME="agentops-marketplace"
PLUGIN_KEY="${PLUGIN_NAME}@${MARKETPLACE_NAME}"
VERSION="${AGENTOPS_INSTALL_VERSION:-unknown}"
UPDATE_CMD="${AGENTOPS_UPDATE_COMMAND:-curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash}"

PLUGIN_MANIFEST="${REPO_ROOT}/.codex-plugin/plugin.json"
MARKETPLACE_FILE="${REPO_ROOT}/.agents/plugins/marketplace.json"
PLUGIN_SKILLS_SRC="${REPO_ROOT}/skills-codex"
PLUGIN_CACHE_ROOT="${CODEX_HOME}/plugins/cache/${MARKETPLACE_NAME}/${PLUGIN_NAME}/local"
PLUGIN_SKILLS_DST="${PLUGIN_CACHE_ROOT}/skills-codex"
LEGACY_SKILLS_DIR="${CODEX_HOME}/skills"
CONFIG_FILE="${CODEX_HOME}/config.toml"
INSTALL_META="${CODEX_HOME}/.agentops-codex-install.json"
LEGACY_BACKUP_ROOT=""

usage() {
  cat <<'EOF'
install-codex-plugin.sh

Install the AgentOps native Codex plugin into CODEX_HOME.

Options:
  --repo-root <dir>     AgentOps repo or extracted release bundle root
  --codex-home <dir>    Target Codex home (default: ~/.codex)
  --skills-src <dir>    Codex-native skills source root (default: <repo-root>/skills-codex)
  --version <value>     Version string to record in install metadata
  --update-command <s>  Update command to record in install metadata
  --help                Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root)
      REPO_ROOT="${2:-}"
      shift 2
      ;;
    --codex-home)
      CODEX_HOME="${2:-}"
      shift 2
      ;;
    --skills-src)
      PLUGIN_SKILLS_SRC="${2:-}"
      shift 2
      ;;
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --update-command)
      UPDATE_CMD="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ "$REPO_ROOT" != /* ]]; then
  REPO_ROOT="$(cd "$REPO_ROOT" && pwd)"
fi
if [[ "$CODEX_HOME" != /* ]]; then
  CODEX_HOME="$(cd "$CODEX_HOME" && pwd)"
fi

PLUGIN_MANIFEST="${REPO_ROOT}/.codex-plugin/plugin.json"
MARKETPLACE_FILE="${REPO_ROOT}/.agents/plugins/marketplace.json"
if [[ "$PLUGIN_SKILLS_SRC" != /* ]]; then
  PLUGIN_SKILLS_SRC="${REPO_ROOT}/${PLUGIN_SKILLS_SRC}"
fi
PLUGIN_CACHE_ROOT="${CODEX_HOME}/plugins/cache/${MARKETPLACE_NAME}/${PLUGIN_NAME}/local"
PLUGIN_SKILLS_DST="${PLUGIN_CACHE_ROOT}/skills-codex"
LEGACY_SKILLS_DIR="${CODEX_HOME}/skills"
CONFIG_FILE="${CODEX_HOME}/config.toml"
INSTALL_META="${CODEX_HOME}/.agentops-codex-install.json"

require_path() {
  local path="$1"
  local label="$2"
  [[ -e "$path" ]] || fail "Missing ${label}: $path"
}

upsert_toml_key() {
  local file="$1"
  local section="$2"
  local key="$3"
  local value="$4"
  local tmp

  mkdir -p "$(dirname "$file")"
  if [[ ! -f "$file" ]]; then
    printf '%s\n%s = %s\n' "$section" "$key" "$value" > "$file"
    return
  fi

  tmp="$(mktemp)"
  awk \
    -v section="$section" \
    -v key="$key" \
    -v value="$value" \
    '
    function emit_key() {
      print key " = " value
    }
    BEGIN {
      in_section = 0
      saw_section = 0
      wrote_key = 0
    }
    {
      if ($0 == section) {
        saw_section = 1
        in_section = 1
        print
        next
      }

      if (in_section && $0 ~ /^\[/) {
        if (!wrote_key) {
          emit_key()
          wrote_key = 1
        }
        in_section = 0
      }

      if (in_section && $0 ~ ("^[[:space:]]*" key "[[:space:]]*=")) {
        if (!wrote_key) {
          emit_key()
          wrote_key = 1
        }
        next
      }

      print
    }
    END {
      if (in_section && !wrote_key) {
        emit_key()
        wrote_key = 1
      }
      if (!saw_section) {
        if (NR > 0) {
          print ""
        }
        print section
        emit_key()
      }
    }
    ' "$file" > "$tmp"
  mv "$tmp" "$file"
}

move_legacy_agentops_skills() {
  local moved=0
  local timestamp
  local backup_dir
  local skill_name

  [[ -d "$LEGACY_SKILLS_DIR" ]] || return 0

  timestamp="$(date +%Y%m%d-%H%M%S)"
  backup_dir="${CODEX_HOME}/agentops-legacy-skills.${timestamp}"

  while IFS= read -r skill_name; do
    [[ -n "$skill_name" ]] || continue
    if [[ -f "$LEGACY_SKILLS_DIR/$skill_name/SKILL.md" ]]; then
      mkdir -p "$backup_dir"
      mv "$LEGACY_SKILLS_DIR/$skill_name" "$backup_dir/$skill_name"
      moved=$((moved + 1))
    fi
  done < <(find "$PLUGIN_SKILLS_SRC" -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort)

  if [[ "$moved" -gt 0 ]]; then
    LEGACY_BACKUP_ROOT="$backup_dir"
    info "Archived $moved legacy AgentOps skill folder(s) from ~/.codex/skills to $backup_dir"
  elif [[ -d "$backup_dir" ]]; then
    rmdir "$backup_dir" 2>/dev/null || true
  fi
}

stage_plugin_source() {
  local staging_root="$1"

  mkdir -p "$staging_root"
  cp -R "$REPO_ROOT/.codex-plugin" "$staging_root/.codex-plugin"
  cp -R "$PLUGIN_SKILLS_SRC" "$staging_root/skills-codex"

  if [[ -f "$REPO_ROOT/.mcp.json" ]]; then
    cp "$REPO_ROOT/.mcp.json" "$staging_root/.mcp.json"
  fi
  if [[ -f "$REPO_ROOT/.app.json" ]]; then
    cp "$REPO_ROOT/.app.json" "$staging_root/.app.json"
  fi
}

require_path "$PLUGIN_MANIFEST" "Codex plugin manifest"
require_path "$MARKETPLACE_FILE" "Codex marketplace manifest"
require_path "$PLUGIN_SKILLS_SRC" "Codex-native skill bundle"

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

info "Installing AgentOps Codex native plugin..."

move_legacy_agentops_skills

mkdir -p "$(dirname "$PLUGIN_CACHE_ROOT")"
rm -rf "$PLUGIN_CACHE_ROOT"
stage_plugin_source "$TMP_DIR/plugin"
cp -R "$TMP_DIR/plugin" "$PLUGIN_CACHE_ROOT"

upsert_toml_key "$CONFIG_FILE" "[features]" "plugins" "true"
upsert_toml_key "$CONFIG_FILE" "[plugins.\"${PLUGIN_KEY}\"]" "enabled" "true"

INSTALLED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
mkdir -p "$(dirname "$INSTALL_META")"
cat > "$INSTALL_META" <<EOF
{
  "installed_at": "$INSTALLED_AT",
  "source": "install-codex-plugin.sh",
  "install_mode": "native-plugin",
  "plugin_key": "$PLUGIN_KEY",
  "version": "$VERSION",
  "plugin_root": "$PLUGIN_CACHE_ROOT",
  "update_command": "$UPDATE_CMD"
}
EOF

SKILL_COUNT="$(find "$PLUGIN_SKILLS_DST" -mindepth 2 -maxdepth 2 -name SKILL.md 2>/dev/null | wc -l | tr -d ' ')"

info "Native Codex plugin installed"
echo "  Plugin key: $PLUGIN_KEY"
echo "  Plugin root: $PLUGIN_CACHE_ROOT"
echo "  Skills available: $SKILL_COUNT"
echo "  Config updated: $CONFIG_FILE"
info "Install metadata written: $INSTALL_META"
if [[ -n "$LEGACY_BACKUP_ROOT" ]]; then
  echo "  Legacy AgentOps raw skills backup: $LEGACY_BACKUP_ROOT"
fi
echo ""
echo "Restart Codex to pick up the native plugin."
