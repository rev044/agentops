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
ROOT="$(cd "$ROOT" 2>/dev/null && pwd -P 2>/dev/null || printf '%s' "$ROOT")"
# Source hook-helpers from plugin install dir, not repo root (security: prevents malicious repo sourcing)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/hook-helpers.sh
. "$SCRIPT_DIR/../lib/hook-helpers.sh"

# Cold start: no chain = no enforcement
[ ! -f "$ROOT/.agents/ao/chain.jsonl" ] && exit 0

# Parse chain directly for speed (avoid spawning ao process)
# Look for the latest vibe step entry
# Schema naming: "gate" + "status" are CANONICAL (current schema)
#                "step" + "locked" are LEGACY (preserved for backward compat with old chain.jsonl entries)
VIBE_LINE=$(grep -E '"(step|gate)"[[:space:]]*:[[:space:]]*"vibe"' "$ROOT/.agents/ao/chain.jsonl" 2>/dev/null | tail -1)

LOG_DIR="$ROOT/.agents/ao"
mkdir -p "$LOG_DIR" 2>/dev/null

VIBE_DONE=false
if [ -z "$VIBE_LINE" ]; then
    # No vibe entry at all — vibe is pending, block
    :
else
    # Check if vibe is locked or skipped
    # CANONICAL: "status": "locked|skipped"  |  LEGACY: "locked": true
    if echo "$VIBE_LINE" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(locked|skipped)"' || echo "$VIBE_LINE" | grep -qE '"locked"[[:space:]]*:[[:space:]]*true'; then
        VIBE_DONE=true
    fi
fi

if [ "$VIBE_DONE" = "false" ]; then
    # Vibe not completed — block push
    if [ -n "$CLAUDE_AGENT_NAME" ] && echo "$CLAUDE_AGENT_NAME" | grep -q '^worker-'; then
        MSG="Push blocked: vibe check needed. Report to team lead."
    else
        MSG="BLOCKED: vibe not completed. Run /vibe before pushing.
Options:
  1. /vibe              -- full council validation
  2. /vibe --quick      -- fast inline check
  3. ao ratchet skip vibe --reason \"<why>\""
    fi

    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) GATE_BLOCK: push-gate blocked (vibe): $CMD" >> "$LOG_DIR/hook-errors.log" 2>/dev/null
    write_failure "push_gate_vibe" "git push" 2 "vibe not completed before push: $CMD"
    echo "$MSG" >&2
    exit 2
fi

# --- Post-mortem gate ---
# If vibe exists, check that post-mortem is also done before allowing push
# Schema naming: "gate" is CANONICAL, "step" is LEGACY (backward compat)
PM_LINE=$(grep -E '"(step|gate)"[[:space:]]*:[[:space:]]*"post-mortem"' "$ROOT/.agents/ao/chain.jsonl" 2>/dev/null | tail -1)

if [ -z "$PM_LINE" ]; then
    # Vibe exists but no post-mortem entry — block
    :
else
    # Check if post-mortem is locked or skipped
    # CANONICAL: "status": "locked|skipped"  |  LEGACY: "locked": true
    if echo "$PM_LINE" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(locked|skipped)"' || echo "$PM_LINE" | grep -qE '"locked"[[:space:]]*:[[:space:]]*true'; then
        # Post-mortem done — allow push
        exit 0
    fi
fi

# Post-mortem not completed — block push
if [ -n "$CLAUDE_AGENT_NAME" ] && echo "$CLAUDE_AGENT_NAME" | grep -q '^worker-'; then
    PM_MSG="Push blocked: post-mortem needed. Report to team lead."
else
    PM_MSG="BLOCKED: post-mortem not completed. Run /post-mortem to capture learnings before pushing.
Options:
  1. /post-mortem          -- full council wrap-up
  2. ao ratchet skip post-mortem --reason '<why>'"
fi

echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) GATE_BLOCK: push-gate blocked (post-mortem): $CMD" >> "$LOG_DIR/hook-errors.log" 2>/dev/null
write_failure "push_gate_postmortem" "git push" 2 "post-mortem not completed before push: $CMD"
echo "$PM_MSG" >&2
exit 2
