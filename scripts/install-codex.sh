#!/usr/bin/env bash
# install-codex.sh — Install AgentOps skills for Codex CLI
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-codex.sh -o /tmp/install-codex.sh
#   bash /tmp/install-codex.sh
#   # or
#   ./scripts/install-codex.sh
#
# What it does:
#   1. Shallow-clones agentops repo (or pulls if exists)
#   2. Symlinks pre-built Codex-native skills into ~/.codex/skills/
#   3. Verifies installation
#
# To update later:
#   cd ~/.codex/agentops && git pull

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
SKILLS_DST="${CODEX_DIR}/skills"
REPO_URL="https://github.com/boshu2/agentops.git"

echo "Installing AgentOps for Codex CLI..."
echo ""

# Step 1: Check Codex is installed
if ! command -v codex &>/dev/null; then
  warn "Codex CLI not found in PATH. Install from https://github.com/openai/codex"
  warn "Continuing anyway — skills will be ready when Codex is installed."
fi

# Step 2: Clone or update repo
if [ -d "$AGENTOPS_DIR/.git" ]; then
  info "AgentOps repo exists, pulling latest..."
  git -C "$AGENTOPS_DIR" pull --ff-only 2>/dev/null || warn "git pull failed — using existing version"
else
  info "Cloning AgentOps..."
  mkdir -p "$(dirname "$AGENTOPS_DIR")"
  git clone --depth 1 "$REPO_URL" "$AGENTOPS_DIR"
fi

# Step 3: Verify pre-built Codex skills exist
SKILLS_SRC="$AGENTOPS_DIR/skills-codex"
if [ ! -d "$SKILLS_SRC" ]; then
  fail "Pre-built Codex skills not found at $SKILLS_SRC"
fi

# Step 4: Symlink skills into ~/.codex/skills/
mkdir -p "$SKILLS_DST"

linked=0
for skill_dir in "$SKILLS_SRC"/*/; do
  [ -d "$skill_dir" ] || continue
  skill_name="$(basename "$skill_dir")"
  dst="$SKILLS_DST/$skill_name"
  rm -rf "$dst"
  ln -s "${skill_dir%/}" "$dst"
  linked=$((linked + 1))
done

# Step 5: Verify
echo ""
SKILL_COUNT=$(find -L "$SKILLS_DST" -name "SKILL.md" -maxdepth 2 2>/dev/null | wc -l | tr -d ' ')
info "Installation complete!"
echo "  Skills: $SKILLS_DST ($SKILL_COUNT skills)"
echo ""
echo "Restart Codex to activate."
echo ""
echo "To update later:"
echo "  cd $AGENTOPS_DIR && git pull"
