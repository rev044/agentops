#!/bin/bash
# prompt-nudge.sh - UserPromptSubmit hook: ratchet-aware one-liner nudges
# Checks prompt keywords against RPI ratchet state. Injects reminders.
# Cap: one nudge line, < 200 bytes. No directory scanning.

# Kill switch
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

# Read all stdin
INPUT=$(cat)

# Extract prompt from JSON
if command -v jq >/dev/null 2>&1; then
    PROMPT=$(echo "$INPUT" | jq -r '.prompt // ""' 2>/dev/null)
else
    PROMPT=$(echo "$INPUT" | grep -o '"prompt"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"prompt"[[:space:]]*:[[:space:]]*"//;s/"$//')
fi

# No prompt → exit silently
[ -z "$PROMPT" ] || [ "$PROMPT" = "null" ] && exit 0

# Find repo root
ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")

# Cold start: no ratchet chain = no nudging
[ ! -f "$ROOT/.agents/ao/chain.jsonl" ] && exit 0

# Check ao availability
command -v ao >/dev/null 2>&1 || exit 0

# Get ratchet status as JSON
RATCHET=$(ao ratchet status -o json 2>/dev/null) || exit 0
[ -z "$RATCHET" ] && exit 0

# Parse steps (requires jq for JSON parsing)
command -v jq >/dev/null 2>&1 || exit 0

# Helper: check if a step is pending
step_pending() {
    echo "$RATCHET" | jq -e ".steps[] | select(.step == \"$1\" and .status == \"pending\")" >/dev/null 2>&1
}

# Lowercase prompt for matching
PROMPT_LOWER=$(echo "$PROMPT" | tr '[:upper:]' '[:lower:]')

NUDGE=""

# Check prompt keywords against ratchet state
if echo "$PROMPT_LOWER" | grep -qE '(implement|build|code|fix|create|add)'; then
    if step_pending "pre-mortem"; then
        NUDGE="Reminder: pre-mortem hasn't been run on your plan."
    fi
elif echo "$PROMPT_LOWER" | grep -qE '(commit|push|ship|deploy|release)'; then
    if step_pending "vibe"; then
        NUDGE="Reminder: run /vibe before pushing."
    fi
elif echo "$PROMPT_LOWER" | grep -qE '(done|finished|wrap|complete|close)'; then
    if step_pending "post-mortem"; then
        NUDGE="Reminder: run /post-mortem to capture learnings."
    fi
fi

# No nudge needed → exit silently
[ -z "$NUDGE" ] && exit 0

# Output nudge as additionalContext
# JSON-escape the nudge (minimal — no special chars expected)
printf '{"hookSpecificOutput":{"additionalContext":"%s"}}\n' "$NUDGE"

exit 0
