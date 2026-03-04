#!/usr/bin/env bash
# PreToolUse hook: warn when skill runs without matching context isolation
# Warn-only (exit 0) — does not block execution

# Kill switches
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

# Read tool input from stdin (Claude Code passes JSON on stdin, NOT env vars)
INPUT=$(cat)

# Extract skill name from JSON input
if command -v jq >/dev/null 2>&1; then
    SKILL_NAME=$(echo "$INPUT" | jq -r '.tool_input.skill // ""' 2>/dev/null)
else
    SKILL_NAME=$(echo "$INPUT" | grep -o '"skill"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"skill"[[:space:]]*:[[:space:]]*"//;s/"$//')
fi
[[ -z "$SKILL_NAME" ]] && exit 0

# Strip namespace prefix (e.g. "agentops:research" → "research")
SKILL_NAME="${SKILL_NAME##*:}"

# Reject skill names with path traversal characters
[[ "$SKILL_NAME" =~ [./\\] ]] && exit 0

# Resolve SKILL.md path
SKILL_DIR=""
for candidate in \
    "${CLAUDE_PLUGIN_ROOT}/skills/${SKILL_NAME}" \
    "${HOME}/.claude/skills/${SKILL_NAME}"; do
    [[ -f "${candidate}/SKILL.md" ]] && SKILL_DIR="$candidate" && break
done
[[ -z "$SKILL_DIR" ]] && exit 0

# Extract context.window from frontmatter
WINDOW=$(awk '/^---$/{n++; next} n==1 && /^context:/{found=1; next} found && /window:/{print $2; exit} n==2{exit}' "${SKILL_DIR}/SKILL.md" 2>/dev/null)
[[ -z "$WINDOW" ]] && exit 0

# Warn if isolated/fork skill runs in inherited context
if [[ "$WINDOW" == "isolated" || "$WINDOW" == "fork" ]]; then
    echo "WARN: Skill '${SKILL_NAME}' declares context.window=${WINDOW} but may be running in caller's context. Use 'ao inject --for=${SKILL_NAME}' for filtered context." >&2
fi

exit 0
