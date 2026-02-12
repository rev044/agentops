#!/bin/bash
# pre-mortem-gate.sh - PreToolUse hook: block /crank when epic has 3+ issues and no pre-mortem
# Evidence: 6/6 consecutive positive pre-mortem ROI across epics.

# Kill switch
[ "${AGENTOPS_SKIP_PRE_MORTEM_GATE:-}" = "1" ] && exit 0

# Workers are exempt
[ "${AGENTOPS_WORKER:-}" = "1" ] && exit 0

# Read all stdin
INPUT=$(cat)

# Extract tool name and args
if command -v jq >/dev/null 2>&1; then
    TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""' 2>/dev/null)
    SKILL_NAME=$(echo "$INPUT" | jq -r '.tool_input.skill // ""' 2>/dev/null)
    SKILL_ARGS=$(echo "$INPUT" | jq -r '.tool_input.args // ""' 2>/dev/null)
else
    # Fallback without jq
    TOOL_NAME=$(echo "$INPUT" | grep -o '"tool_name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"tool_name"[[:space:]]*:[[:space:]]*"//;s/"$//')
    SKILL_NAME=$(echo "$INPUT" | grep -o '"skill"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"skill"[[:space:]]*:[[:space:]]*"//;s/"$//')
    SKILL_ARGS=$(echo "$INPUT" | grep -o '"args"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"args"[[:space:]]*:[[:space:]]*"//;s/"$//')
fi

# Only gate Skill tool calls for crank
[ "$TOOL_NAME" = "Skill" ] || exit 0
echo "$SKILL_NAME" | grep -qiE '^(crank|agentops:crank)$' || exit 0

# Check for --skip-pre-mortem bypass in args
echo "$SKILL_ARGS" | grep -q "\-\-skip-pre-mortem" && exit 0

# Extract epic-id from args (first arg that looks like a bead ID)
EPIC_ID=$(echo "$SKILL_ARGS" | grep -oE '[a-z]{2}-[a-z0-9]+' | head -1)
[ -z "$EPIC_ID" ] && exit 0  # No epic ID found, can't check — fail open

# Count children
if ! command -v bd &>/dev/null; then
    exit 0  # No bd CLI, can't count — fail open
fi
CHILD_COUNT=$(bd children "$EPIC_ID" 2>/dev/null | wc -l | tr -d ' ')
[ "$CHILD_COUNT" -lt 3 ] && exit 0  # Less than 3 issues, no gate needed

# Check for pre-mortem evidence (scoped to this epic)
# Method 1: Council artifacts — match epic-specific pattern
if ls .agents/council/*-pre-mortem-"$EPIC_ID"* >/dev/null 2>&1 || \
   ls .agents/council/*-pre-mortem-"${EPIC_ID%%.*}"* >/dev/null 2>&1; then
    exit 0
fi
# Method 1b: Fallback — any pre-mortem from today (same session)
TODAY=$(date +%Y-%m-%d)
if ls .agents/council/"$TODAY"-*pre-mortem* >/dev/null 2>&1; then
    exit 0
fi

# Method 2: Ratchet record
if command -v ao &>/dev/null; then
    if ao ratchet status -o json 2>/dev/null | grep -q '"pre-mortem"'; then
        exit 0
    fi
fi

# No evidence found — block
ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")
ROOT="$(cd "$ROOT" 2>/dev/null && pwd -P 2>/dev/null || printf '%s' "$ROOT")"
# shellcheck source=../lib/hook-helpers.sh
. "$ROOT/lib/hook-helpers.sh"

LOG_DIR="$ROOT/.agents/ao"
mkdir -p "$LOG_DIR" 2>/dev/null
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) GATE_BLOCK: pre-mortem-gate blocked crank for $EPIC_ID ($CHILD_COUNT children)" >> "$LOG_DIR/hook-errors.log" 2>/dev/null
write_failure "pre_mortem_gate" "bd children $EPIC_ID" 2 "Epic $EPIC_ID has $CHILD_COUNT issues, no pre-mortem evidence found"

cat >&2 <<EOMSG
BLOCKED: Epic $EPIC_ID has $CHILD_COUNT issues. Pre-mortem is mandatory for 3+ issue epics.
(6/6 consecutive positive ROI — this gate prevents implementation waste.)

Options:
  1. /pre-mortem                         -- run pre-mortem validation
  2. /crank $EPIC_ID --skip-pre-mortem   -- bypass with justification
  3. export AGENTOPS_SKIP_PRE_MORTEM_GATE=1  -- disable gate entirely
EOMSG
exit 2
