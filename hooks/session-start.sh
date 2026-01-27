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

# Detect and read AGENTS.md if present (competitor adoption: AGENTS.md standard)
agents_md_content=""
if [[ -f "AGENTS.md" ]]; then
    agents_md_content=$(cat AGENTS.md 2>/dev/null || echo "")
fi

# Read the using-agentops skill content (with safe error handling)
SKILL_FILE="${PLUGIN_ROOT}/skills/using-agentops/SKILL.md"
if [[ -f "$SKILL_FILE" ]]; then
    using_agentops_content=$(cat "$SKILL_FILE")
else
    # Generic fallback - don't leak path information
    using_agentops_content="(AgentOps skill content unavailable)"
fi

# escape_for_json: Escape string for safe JSON embedding
# Handles: backslash, quotes, newlines, carriage returns, tabs
# Parameters: $1 = input string
# Output: Escaped string suitable for JSON string value (without surrounding quotes)
# Note: This is used when jq is not available; prefer jq when possible
escape_for_json() {
    local input="$1"
    local output=""
    local i char
    for (( i=0; i<${#input}; i++ )); do
        char="${input:$i:1}"
        # shellcheck disable=SC1003
        case "$char" in
            '\'*) output+='\\\\' ;;
            '"') output+='\\"' ;;
            $'\n') output+='\\n' ;;
            $'\r') output+='\\r' ;;
            $'\t') output+='\\t' ;;
            *) output+="$char" ;;
        esac
    done
    printf '%s' "$output"
}

# Build AGENTS.md section if content exists
agents_md_section=""
if [[ -n "$agents_md_content" ]]; then
    agents_md_section="\n\n## Project Agent Instructions (AGENTS.md)\n\n${agents_md_content}"
fi

# Combine all content for context injection
full_content="<EXTREMELY_IMPORTANT>
You have AgentOps superpowers.

**Below is the full content of your 'agentops:using-agentops' skill - your introduction to using AgentOps skills. For all other skills, use the 'Skill' tool:**

${using_agentops_content}${agents_md_section}
</EXTREMELY_IMPORTANT>"

# Output context injection as JSON using jq for safe encoding (preferred)
# Falls back to manual escaping if jq unavailable
if command -v jq &>/dev/null; then
    # Use jq to safely encode the entire content as a JSON string
    additional_context=$(printf '%s' "$full_content" | jq -Rs '.')
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": ${additional_context}
  }
}
EOF
else
    # Fallback: manual escaping (less safe but functional)
    escaped_content=$(escape_for_json "$full_content")
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "${escaped_content}"
  }
}
EOF
fi

exit 0
