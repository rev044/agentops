#!/bin/bash
# commit-review-gate.sh - PreToolUse hook: inject staged diff before git commit
# Forces Claude to see its own changes before committing.
# Non-blocking (always exit 0). Injects diff as additionalContext.

# Kill switches
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0
[ "${AGENTOPS_COMMIT_REVIEW_DISABLED:-}" = "1" ] && exit 0

# Read stdin
INPUT=$(cat)

# Extract tool name and command
TOOL_NAME="${CLAUDE_TOOL_NAME:-}"
COMMAND="${CLAUDE_TOOL_INPUT_COMMAND:-}"
if [ -z "$TOOL_NAME" ]; then
    TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""' 2>/dev/null) || exit 0
fi
if [ -z "$COMMAND" ]; then
    COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // ""' 2>/dev/null) || exit 0
fi

# Only fire on Bash + git commit
[ "$TOOL_NAME" = "Bash" ] || exit 0
echo "$COMMAND" | grep -q 'git commit' || exit 0

# Don't fire on --amend with no new changes (just message edit)
echo "$COMMAND" | grep -qE '\-\-amend.*\-\-no-edit' && exit 0

# Get staged diff summary
DIFF_STAT=$(git diff --cached --stat 2>/dev/null)
[ -z "$DIFF_STAT" ] && exit 0

# Count changed files
FILE_COUNT=$(git diff --cached --name-only 2>/dev/null | wc -l | tr -d ' ')
[ "$FILE_COUNT" = "0" ] && exit 0

# Get abbreviated diff (truncate at 200 lines)
DIFF_CONTENT=$(git diff --cached 2>/dev/null | head -200)
DIFF_LINES=$(git diff --cached 2>/dev/null | wc -l | tr -d ' ')
TRUNCATED=""
if [ "$DIFF_LINES" -gt 200 ]; then
    TRUNCATED=" (showing first 200 of $DIFF_LINES lines — run 'git diff --cached' for full diff)"
fi

# Build review context
REVIEW_MSG="SELF-REVIEW before committing ($FILE_COUNT files changed):
Check for: wrong variable references, changed defaults, removed error handling, silent data loss, YAML syntax errors.

Staged changes:
$DIFF_STAT
${TRUNCATED}

${DIFF_CONTENT}"

# Inject as additionalContext
if command -v jq >/dev/null 2>&1; then
    jq -n --arg ctx "$REVIEW_MSG" '{"hookSpecificOutput":{"additionalContext":$ctx}}'
else
    # Fallback: escape for JSON
    safe_msg=${REVIEW_MSG//\\/\\\\}
    safe_msg=${safe_msg//\"/\\\"}
    safe_msg=$(echo "$safe_msg" | tr '\n' ' ')
    echo "{\"hookSpecificOutput\":{\"additionalContext\":\"$safe_msg\"}}"
fi

exit 0
