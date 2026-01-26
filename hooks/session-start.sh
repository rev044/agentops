#!/usr/bin/env bash
# AgentOps Session Start Hook
# Creates .agents/ directories and injects using-agentops context

set -euo pipefail

# Get plugin directory (where this script lives)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Create .agents directories if they don't exist
AGENTS_DIRS=(".agents/research" ".agents/products" ".agents/retros" ".agents/learnings" ".agents/patterns")

for dir in "${AGENTS_DIRS[@]}"; do
    if [[ ! -d "$dir" ]]; then
        mkdir -p "$dir"
    fi
done

# Read the using-agentops skill content
using_agentops_content=$(cat "${PLUGIN_ROOT}/skills/using-agentops/SKILL.md" 2>&1 || echo "Error reading using-agentops skill")

# Escape outputs for JSON using pure bash
escape_for_json() {
    local input="$1"
    local output=""
    local i char
    for (( i=0; i<${#input}; i++ )); do
        char="${input:$i:1}"
        case "$char" in
            $'\\') output+='\\' ;;
            '"') output+='\"' ;;
            $'\n') output+='\n' ;;
            $'\r') output+='\r' ;;
            $'\t') output+='\t' ;;
            *) output+="$char" ;;
        esac
    done
    printf '%s' "$output"
}

using_agentops_escaped=$(escape_for_json "$using_agentops_content")

# Output context injection as JSON
cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "<EXTREMELY_IMPORTANT>\nYou have AgentOps superpowers.\n\n**Below is the full content of your 'agentops:using-agentops' skill - your introduction to using AgentOps skills. For all other skills, use the 'Skill' tool:**\n\n${using_agentops_escaped}\n</EXTREMELY_IMPORTANT>"
  }
}
EOF

exit 0
