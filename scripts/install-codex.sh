#!/usr/bin/env bash
# install-codex.sh — Install AgentOps Codex-native skills (no repo clone)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh -o /tmp/install-codex.sh
#   bash /tmp/install-codex.sh
#   # or
#   ./scripts/install-codex.sh
#
# What it does:
#   1. Downloads a temporary AgentOps archive (no local git clone)
#   2. Installs runtime helper files into ~/.codex/agentops/
#   3. Installs pre-built Codex-native skills into ~/.codex/agentops/skills/
#   4. Writes local install metadata for stale-skill reminders
#
# To update later:
#   re-run this installer

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}✓${NC} $*"; }
warn()  { echo -e "${YELLOW}!${NC} $*"; }
fail()  { echo -e "${RED}✗${NC} $*"; exit 1; }

CODEX_DIR="${HOME}/.codex"
AGENTOPS_DIR="${CODEX_DIR}/agentops"
RUNTIME_DOT_CODEX="${AGENTOPS_DIR}/.codex"
RUNTIME_LIB_DIR="${AGENTOPS_DIR}/lib"
SKILLS_DST="${AGENTOPS_DIR}/skills"
LEGACY_PERSONAL_SKILLS="${CODEX_DIR}/skills"
INSTALL_META="${CODEX_DIR}/.agentops-codex-install.json"
ARCHIVE_URL="https://codeload.github.com/boshu2/agentops/tar.gz/refs/heads/main"
UPDATE_CMD="curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash"

echo "Installing AgentOps for Codex CLI..."
echo ""

for cmd in curl tar; do
  if ! command -v "$cmd" &>/dev/null; then
    fail "Missing required command: $cmd"
  fi
done

# Step 1: Check Codex is installed
if ! command -v codex &>/dev/null; then
  warn "Codex CLI not found in PATH. Install from https://github.com/openai/codex"
  warn "Continuing anyway — skills will be ready when Codex is installed."
fi

# Step 2: Download temporary archive (no persistent clone)
TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

ARCHIVE_FILE="${TMP_DIR}/agentops.tar.gz"
info "Downloading AgentOps bundle..."
curl -fsSL "$ARCHIVE_URL" -o "$ARCHIVE_FILE"

ARCHIVE_ROOT="$(tar -tzf "$ARCHIVE_FILE" | head -1 | cut -d/ -f1)"
[ -n "$ARCHIVE_ROOT" ] || fail "Could not determine archive root directory"
tar -xzf "$ARCHIVE_FILE" -C "$TMP_DIR"
SRC_ROOT="${TMP_DIR}/${ARCHIVE_ROOT}"

# Step 3: Resolve bundle sources
SKILLS_SRC="${SRC_ROOT}/skills-codex"
HELPER_SRC="${SRC_ROOT}/.codex/agentops-codex"
BOOTSTRAP_SRC="${SRC_ROOT}/.codex/agentops-bootstrap.md"
CORE_SRC="${SRC_ROOT}/lib/skills-core.js"
[ -d "$SKILLS_SRC" ] || fail "Pre-built Codex skills not found in bundle"
[ -f "$HELPER_SRC" ] || fail "Missing agentops-codex helper in bundle"
[ -f "$BOOTSTRAP_SRC" ] || fail "Missing bootstrap instructions in bundle"
[ -f "$CORE_SRC" ] || fail "Missing shared skills core in bundle"

# Step 4: Install runtime helper files
mkdir -p "$RUNTIME_DOT_CODEX" "$RUNTIME_LIB_DIR"
cp "$HELPER_SRC" "${RUNTIME_DOT_CODEX}/agentops-codex"
chmod +x "${RUNTIME_DOT_CODEX}/agentops-codex"
cp "$BOOTSTRAP_SRC" "${RUNTIME_DOT_CODEX}/agentops-bootstrap.md"
cp "$CORE_SRC" "${RUNTIME_LIB_DIR}/skills-core.js"
info "Installed runtime helper files to $AGENTOPS_DIR"

# Step 5: Install codex-native skills into ~/.codex/agentops/skills
rm -rf "$SKILLS_DST"
mkdir -p "$SKILLS_DST"
linked=0
for skill_dir in "$SKILLS_SRC"/*/; do
  [ -d "$skill_dir" ] || continue
  skill_name="$(basename "$skill_dir")"
  cp -R "$skill_dir" "${SKILLS_DST}/${skill_name}"
  linked=$((linked + 1))
done
info "Installed $linked Codex-native skills"

# Step 6: Remove legacy symlinked installs from ~/.codex/skills when present
legacy_removed=0
if [ -d "$LEGACY_PERSONAL_SKILLS" ]; then
  for skill_dir in "$SKILLS_SRC"/*/; do
    [ -d "$skill_dir" ] || continue
    skill_name="$(basename "$skill_dir")"
    legacy_path="${LEGACY_PERSONAL_SKILLS}/${skill_name}"
    if [ -L "$legacy_path" ]; then
      link_target="$(readlink "$legacy_path" || true)"
      case "$link_target" in
        *"/.codex/agentops/skills-codex/"*|*"/.codex/agentops/skills-codex")
          rm -f "$legacy_path"
          legacy_removed=$((legacy_removed + 1))
          ;;
      esac
    fi
  done
fi
if [ "$legacy_removed" -gt 0 ]; then
  info "Removed $legacy_removed legacy symlink(s) from $LEGACY_PERSONAL_SKILLS"
fi

# Step 7: Verify
echo ""
SKILL_COUNT=$(find "$SKILLS_DST" -name "SKILL.md" -maxdepth 2 2>/dev/null | wc -l | tr -d ' ')
info "Installation complete!"
echo "  Skills: $SKILLS_DST ($SKILL_COUNT skills)"

# Step 8: Write local install metadata for stale-skill reminders (no telemetry)
INSTALLED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
VERSION="main"
cat > "$INSTALL_META" <<EOF
{
  "installed_at": "$INSTALLED_AT",
  "source": "install-codex.sh",
  "version": "$VERSION",
  "update_command": "$UPDATE_CMD"
}
EOF
info "Install metadata written: $INSTALL_META"
echo ""
echo "Restart Codex to activate."
echo ""
echo "To update later:"
echo "  $UPDATE_CMD"
