#!/usr/bin/env bash
# AgentOps Session Start Hook
# Creates .agents/ directories and injects using-agentops context

set -euo pipefail

# Get plugin directory (where this script lives)
PLUGIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Create .agents directories if they don't exist
AGENTS_DIRS=(".agents/research" ".agents/products" ".agents/retros" ".agents/learnings" ".agents/patterns")

for dir in "${AGENTS_DIRS[@]}"; do
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
    fi
done

# Read the using-agentops skill content
SKILL_PATH="${PLUGIN_DIR}/skills/using-agentops/SKILL.md"

if [[ -f "$SKILL_PATH" ]]; then
    # Extract content after frontmatter (skip YAML header)
    SKILL_CONTENT=$(awk '/^---$/{p=!p;next}p==0{print}' "$SKILL_PATH" | tail -n +2)

    # Output JSON with additionalContext (for Claude Code hook system)
    cat << EOF
{
  "additionalContext": $(echo "$SKILL_CONTENT" | jq -Rs .)
}
EOF
else
    # Fallback if skill file missing
    cat << 'EOF'
{
  "additionalContext": "# AgentOps\n\nAvailable skills: /research, /plan, /implement, /crank, /vibe, /retro, /post-mortem, /beads, /bug-hunt, /knowledge, /complexity, /doc"
}
EOF
fi
