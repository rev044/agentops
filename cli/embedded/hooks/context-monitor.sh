#!/usr/bin/env bash
set -euo pipefail

# context-monitor.sh - PostToolUse hook: context window usage monitoring via bridge pattern
#
# Reads context metrics from statusline bridge file and injects agent-facing
# warnings at 35% and 25% remaining context.
#
# Bridge file: /tmp/claude-ctx-{CLAUDE_SESSION_ID}.json
# Written by statusline command (user-configured).
#
# Env vars:
#   AGENTOPS_HOOKS_DISABLED=1          - disable all hooks
#   AGENTOPS_CONTEXT_MONITOR_DISABLED=1 - disable this hook only
#   AGENTOPS_CONTEXT_WARN_PERCENT=35   - warning threshold (default 35)
#   AGENTOPS_CONTEXT_CRIT_PERCENT=25   - critical threshold (default 25)

# Kill switches
[[ "${AGENTOPS_HOOKS_DISABLED:-}" == "1" ]] && exit 0
[[ "${AGENTOPS_CONTEXT_MONITOR_DISABLED:-}" == "1" ]] && exit 0

# Consume stdin (required even if unused)
cat > /dev/null

SESSION_ID="${CLAUDE_SESSION_ID:-}"
[[ -z "$SESSION_ID" ]] && exit 0

BRIDGE_FILE="/tmp/claude-ctx-${SESSION_ID}.json"
[[ ! -f "$BRIDGE_FILE" ]] && exit 0

# Bridge file must be recent (< 5 min old) to avoid stale data
if command -v stat >/dev/null 2>&1; then
    if [[ "$(uname)" == "Darwin" ]]; then
        FILE_AGE=$(( $(date +%s) - $(stat -f %m "$BRIDGE_FILE") ))
    else
        FILE_AGE=$(( $(date +%s) - $(stat -c %Y "$BRIDGE_FILE") ))
    fi
    [[ "$FILE_AGE" -gt 300 ]] && exit 0
fi

# Read remaining percent from bridge file
if command -v jq >/dev/null 2>&1; then
    REMAINING=$(jq -r '.remaining_percent // ""' "$BRIDGE_FILE" 2>/dev/null) || exit 0
else
    REMAINING=$(grep -o '"remaining_percent"[[:space:]]*:[[:space:]]*[0-9]*' "$BRIDGE_FILE" 2>/dev/null | grep -o '[0-9]*$') || exit 0
fi

[[ -z "$REMAINING" ]] && exit 0

WARN_THRESHOLD="${AGENTOPS_CONTEXT_WARN_PERCENT:-35}"
CRIT_THRESHOLD="${AGENTOPS_CONTEXT_CRIT_PERCENT:-25}"

MSG=""

if [[ "$REMAINING" -le "$CRIT_THRESHOLD" ]]; then
    MSG="Context window at ${REMAINING}% remaining. Plan final steps carefully — room for ~1-2 more significant operations. Consider /handoff."
elif [[ "$REMAINING" -le "$WARN_THRESHOLD" ]]; then
    MSG="Context window at ${REMAINING}% remaining. Consider wrapping up current work within ~2-3 tasks."
fi

[[ -z "$MSG" ]] && exit 0

# Emit agent-facing warning
if command -v jq >/dev/null 2>&1; then
    jq -n --arg ctx "$MSG" '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":$ctx}}'
else
    safe_msg=${MSG//\\/\\\\}
    safe_msg=${safe_msg//\"/\\\"}
    echo "{\"hookSpecificOutput\":{\"hookEventName\":\"PostToolUse\",\"additionalContext\":\"$safe_msg\"}}"
fi

exit 0
