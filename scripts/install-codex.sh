#!/usr/bin/env bash
# install-codex.sh — Install AgentOps Codex-native skills into ~/.codex/skills
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash
#
# What it does:
#   1. Downloads a temporary AgentOps archive (no local git clone)
#   2. Copies pre-built Codex-native skills into ~/.codex/skills
#   3. Writes local install metadata
#
# Update policy:
#   Re-run this installer when new AgentOps releases land.

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
SKILLS_DST="${CODEX_DIR}/skills"
INSTALL_META="${CODEX_DIR}/.agentops-codex-install.json"
ARCHIVE_URL="https://codeload.github.com/boshu2/agentops/tar.gz/refs/heads/main"
UPDATE_CMD="curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh | bash"

echo "Installing AgentOps Codex skills..."
echo ""

for cmd in curl tar; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    fail "Missing required command: $cmd"
  fi
done

if ! command -v codex >/dev/null 2>&1; then
  warn "Codex CLI not found in PATH. Install from https://github.com/openai/codex"
  warn "Continuing anyway — skills will be ready when Codex is installed."
fi

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
SKILLS_SRC="${SRC_ROOT}/skills-codex"
[ -d "$SKILLS_SRC" ] || fail "Pre-built Codex skills not found in bundle"

mkdir -p "$SKILLS_DST"

installed=0
for skill_dir in "$SKILLS_SRC"/*/; do
  [ -d "$skill_dir" ] || continue
  skill_name="$(basename "$skill_dir")"
  dst="${SKILLS_DST}/${skill_name}"
  rm -rf "$dst"
  cp -R "${skill_dir%/}" "$dst"
  installed=$((installed + 1))
done

INSTALLED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
mkdir -p "$(dirname "$INSTALL_META")"
cat > "$INSTALL_META" <<EOF
{
  "installed_at": "$INSTALLED_AT",
  "source": "install-codex.sh",
  "version": "main",
  "update_command": "$UPDATE_CMD"
}
EOF

echo ""
SKILL_COUNT=$(find "$SKILLS_DST" -name "SKILL.md" -maxdepth 2 2>/dev/null | wc -l | tr -d ' ')
info "Installation complete!"
echo "  Skills installed: $installed"
echo "  Skill index count: $SKILL_COUNT"
echo "  Location: $SKILLS_DST"
info "Install metadata written: $INSTALL_META"
echo ""
echo "Update note:"
echo "  AgentOps ships frequent updates."
echo "  Re-run this installer regularly (especially after new releases):"
echo "  $UPDATE_CMD"
