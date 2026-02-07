#!/bin/bash
# push-gate.sh - PreToolUse hook: block git push/tag when vibe not completed
# Gates on RPI ratchet state. git commit is NOT blocked (local, reversible).
# Cold start (no chain.jsonl) = no enforcement.

# Kill switch
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

# Workers are exempt (they should never push anyway)
[ "${AGENTOPS_WORKER:-}" = "1" ] && exit 0

# Read all stdin
INPUT=$(cat)

# Extract tool_input.command from JSON
if command -v jq >/dev/null 2>&1; then
    CMD=$(echo "$INPUT" | jq -r '.tool_input.command // ""' 2>/dev/null)
else
    CMD=$(echo "$INPUT" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"command"[[:space:]]*:[[:space:]]*"//;s/"$//')
fi

# No command → pass through
[ -z "$CMD" ] || [ "$CMD" = "null" ] && exit 0

# Hot path: only care about git push/tag (<50ms for non-git commands)
echo "$CMD" | grep -qE 'git\s+(push|tag)' || exit 0

# Find repo root
ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
if [ -z "$ROOT" ]; then
    # Not in a git repo — can't enforce, fail open
    exit 0
fi

# Cold start: no chain = no enforcement
[ ! -f "$ROOT/.agents/ao/chain.jsonl" ] && exit 0

# Parse chain directly for speed (avoid spawning ao process)
# Look for the latest vibe step entry
VIBE_LINE=$(grep '"step"[[:space:]]*:[[:space:]]*"vibe"' "$ROOT/.agents/ao/chain.jsonl" 2>/dev/null | tail -1)

if [ -z "$VIBE_LINE" ]; then
    # No vibe entry at all — vibe is pending, block
    :
else
    # Check if vibe is locked or skipped
    if echo "$VIBE_LINE" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(locked|skipped)"'; then
        # Vibe completed or skipped — allow push
        exit 0
    fi
fi

# Determine message based on agent context
if [ -n "$CLAUDE_AGENT_NAME" ] && echo "$CLAUDE_AGENT_NAME" | grep -q '^worker-'; then
    MSG="Push blocked: vibe check needed. Report to team lead."
else
    MSG="BLOCKED: vibe not completed. Run /vibe before pushing.
Options:
  1. /vibe              -- full council validation
  2. /vibe --quick      -- fast inline check
  3. ao ratchet skip vibe --reason \"<why>\"
To disable all gates: export AGENTOPS_HOOKS_DISABLED=1"
fi

# Log the block
LOG_DIR="$ROOT/.agents/ao"
mkdir -p "$LOG_DIR" 2>/dev/null
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) GATE_BLOCK: push-gate blocked: $CMD" >> "$LOG_DIR/hook-errors.log" 2>/dev/null

echo "$MSG" >&2
exit 2
