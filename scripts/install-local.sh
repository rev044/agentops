#!/bin/bash
# Install a plugin locally for testing before publishing
# Usage: ./scripts/install-local.sh <plugin-name>
#
# This copies the plugin to ~/.claude/plugins/ so you can test it
# exactly as a user would experience it after marketplace install.
#
# Example:
#   ./scripts/install-local.sh solo-kit
#   claude   # Now has solo-kit skills available

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CLAUDE_PLUGINS_DIR="${HOME}/.claude/plugins"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

usage() {
    echo "Usage: $0 <plugin-name> [--uninstall]"
    echo ""
    echo "Examples:"
    echo "  $0 solo-kit          # Install solo-kit locally"
    echo "  $0 solo-kit --uninstall  # Remove local installation"
    echo "  $0 --list            # List available plugins"
    echo ""
    echo "Available plugins:"
    for p in "$REPO_ROOT/plugins"/*/; do
        echo "  - $(basename "$p")"
    done
    exit 1
}

if [[ $# -lt 1 ]]; then
    usage
fi

if [[ "$1" == "--list" ]]; then
    echo "Available plugins:"
    for p in "$REPO_ROOT/plugins"/*/; do
        name=$(basename "$p")
        skills=$(find "$p/skills" -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')
        echo "  $name ($skills skills)"
    done
    exit 0
fi

PLUGIN_NAME="$1"
PLUGIN_SRC="$REPO_ROOT/plugins/$PLUGIN_NAME"
PLUGIN_DST="$CLAUDE_PLUGINS_DIR/$PLUGIN_NAME"

if [[ ! -d "$PLUGIN_SRC" ]]; then
    echo -e "${RED}Error: Plugin not found: $PLUGIN_NAME${NC}"
    echo ""
    usage
fi

# Uninstall
if [[ "${2:-}" == "--uninstall" ]]; then
    if [[ -d "$PLUGIN_DST" ]]; then
        rm -rf "$PLUGIN_DST"
        echo -e "${GREEN}✓${NC} Uninstalled $PLUGIN_NAME from $PLUGIN_DST"
    else
        echo -e "${YELLOW}!${NC} $PLUGIN_NAME not installed at $PLUGIN_DST"
    fi
    exit 0
fi

# Validate first
echo -e "${BLUE}Validating $PLUGIN_NAME...${NC}"
if ! "$REPO_ROOT/scripts/validate-local.sh" "$PLUGIN_NAME"; then
    echo ""
    echo -e "${RED}Validation failed - not installing${NC}"
    exit 1
fi

# Create plugins directory if needed
mkdir -p "$CLAUDE_PLUGINS_DIR"

# Remove old installation if exists
if [[ -d "$PLUGIN_DST" ]]; then
    echo -e "${YELLOW}Removing existing installation...${NC}"
    rm -rf "$PLUGIN_DST"
fi

# Copy plugin
echo -e "${BLUE}Installing to $PLUGIN_DST...${NC}"
cp -r "$PLUGIN_SRC" "$PLUGIN_DST"

# Verify installation
skill_count=$(find "$PLUGIN_DST/skills" -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  $PLUGIN_NAME installed successfully!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
echo ""
echo "  Location: $PLUGIN_DST"
echo "  Skills:   $skill_count"
echo ""
echo "  Test it:"
echo "    claude --plugin-dir $PLUGIN_DST"
echo ""
echo "  Or enable in settings.json:"
echo "    \"enabledPlugins\": { \"$PLUGIN_NAME\": true }"
echo ""
echo "  Uninstall:"
echo "    $0 $PLUGIN_NAME --uninstall"
echo ""
