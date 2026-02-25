#!/usr/bin/env bash
# AgentOps Session Start Hook (minimal flywheel)
# Creates .agents/ directories, optionally runs extract+inject, injects skill context.
#
# Startup modes (AGENTOPS_STARTUP_CONTEXT_MODE):
#   manual  (default) — MEMORY.md auto-loaded by Claude Code; emit pointer only, no extract/inject
#   lean    — extract + lean inject (shrinks when MEMORY.md is fresh)
#   legacy  — extract + full inject (pre-notebook behavior)

# Kill switches
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0
[ "${AGENTOPS_SESSION_START_DISABLED:-}" = "1" ] && exit 0

STARTUP_MODE="${AGENTOPS_STARTUP_CONTEXT_MODE:-manual}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
PLUGIN_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
ROOT="$(cd "$ROOT" 2>/dev/null && pwd -P 2>/dev/null || printf '%s' "$ROOT")"
AO_DIR="$ROOT/.agents/ao"

HOOK_ERROR_LOG="$AO_DIR/hook-errors.log"

log_hook_fail() {
    mkdir -p "$AO_DIR" 2>/dev/null || return 0
    echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) HOOK_FAIL: $1" >> "$HOOK_ERROR_LOG" 2>/dev/null || true
}

AO_TIMEOUT_BIN="timeout"
command -v "$AO_TIMEOUT_BIN" >/dev/null 2>&1 || AO_TIMEOUT_BIN="gtimeout"

run_ao_quick() {
    local seconds="$1"; shift
    if command -v "$AO_TIMEOUT_BIN" >/dev/null 2>&1; then
        "$AO_TIMEOUT_BIN" "$seconds" ao "$@" 2>/dev/null
        return $?
    fi
    ao "$@" 2>/dev/null
}

cd "$ROOT" 2>/dev/null || true

# Ensure global .agents/ directories exist (cross-repo knowledge)
mkdir -p "$HOME/.agents/learnings" "$HOME/.agents/patterns" 2>/dev/null

# Ensure local .agents/ directories exist
for dir in .agents/research .agents/products .agents/retros .agents/learnings \
           .agents/patterns .agents/council .agents/knowledge/pending \
           .agents/plans .agents/rpi .agents/ao; do
    mkdir -p "$ROOT/$dir" 2>/dev/null
done

# Auto-gitignore .agents/
if [ "${AGENTOPS_GITIGNORE_AUTO:-1}" != "0" ] && [ -d "$ROOT/.git" ]; then
    GITIGNORE="$ROOT/.gitignore"
    if [ -f "$GITIGNORE" ]; then
        grep -q '\.agents/' "$GITIGNORE" 2>/dev/null || \
            printf '\n# AgentOps session artifacts\n.agents/\n' >> "$GITIGNORE" 2>/dev/null
    else
        printf '# AgentOps session artifacts\n.agents/\n' > "$GITIGNORE" 2>/dev/null
    fi
fi
if [ ! -f "$ROOT/.agents/.gitignore" ]; then
    cat > "$ROOT/.agents/.gitignore" 2>/dev/null <<'EOF'
*
!.gitignore
!README.md
EOF
fi

# Flywheel behavior depends on startup mode
INJECTED_KNOWLEDGE=""
NOTEBOOK_LEAN_MODE=0

# Predecessor handoff discovery (used in all modes)
PREDECESSOR_FILE="${GT_PREDECESSOR_HANDOFF:-}"
if [ -z "$PREDECESSOR_FILE" ] && [ -d "$ROOT/.agents/handoff" ]; then
    PREDECESSOR_FILE=$(ls -t "$ROOT/.agents/handoff/"*.md 2>/dev/null | head -1)
fi

if [ "$STARTUP_MODE" = "manual" ]; then
    # Manual mode (default): MEMORY.md is auto-loaded by Claude Code.
    # No extract/inject — emit pointer-only context for JIT retrieval.
    MANUAL_CTX="MEMORY.md is auto-loaded by Claude Code for this project.
For on-demand retrieval: \`ao search \"<query>\"\` or \`ao lookup --query \"<query>\"\`"
    if [ -n "$PREDECESSOR_FILE" ] && [ -f "$PREDECESSOR_FILE" ]; then
        MANUAL_CTX="${MANUAL_CTX}
Predecessor handoff: ${PREDECESSOR_FILE}"
    fi
    INJECTED_KNOWLEDGE="$MANUAL_CTX"

elif command -v ao &>/dev/null; then
    # Lean/legacy mode: extract pending queue + inject prior knowledge
    INJECT_EXTRA_FLAGS=()
    if [ -n "${HOOK_BEAD:-}" ]; then
        INJECT_EXTRA_FLAGS+=(--bead "$HOOK_BEAD")
        run_ao_quick 5 extract --bead "$HOOK_BEAD" || log_hook_fail "ao extract --bead"
    else
        run_ao_quick 5 extract || log_hook_fail "ao extract"
    fi

    if [ -n "$PREDECESSOR_FILE" ] && [ -f "$PREDECESSOR_FILE" ]; then
        INJECT_EXTRA_FLAGS+=(--predecessor "$PREDECESSOR_FILE")
    fi

    # Notebook-aware lean injection (skip in legacy mode)
    MEMORY_DIR="$HOME/.claude/projects"
    if [ "$STARTUP_MODE" != "legacy" ] && [ -d "$MEMORY_DIR" ]; then
        PROJECT_PATH=$(printf '%s' "$ROOT" | tr '/' '-')
        MEMORY_FILE="$MEMORY_DIR/$PROJECT_PATH/memory/MEMORY.md"
        if [ -f "$MEMORY_FILE" ]; then
            MTIME=$(stat -f %m "$MEMORY_FILE" 2>/dev/null || stat -c %Y "$MEMORY_FILE" 2>/dev/null || echo "")
            if [ -n "$MTIME" ]; then
                MEMORY_AGE_DAYS=$(( ($(date +%s) - MTIME) / 86400 ))
            else
                MEMORY_AGE_DAYS=0  # stat failed but file exists — assume fresh
            fi
            if [ "$MEMORY_AGE_DAYS" -le 7 ]; then
                NOTEBOOK_LEAN_MODE=1
            fi
        fi
    fi

    INJECT_MODE_FLAGS=(--apply-decay --format markdown)
    if [ "$NOTEBOOK_LEAN_MODE" = "1" ]; then
        INJECT_MODE_FLAGS+=(--max-tokens 400)
    elif [ "${AGENTOPS_INDEX_INJECT:-0}" = "1" ]; then
        INJECT_MODE_FLAGS+=(--index-only --max-tokens 400)
    else
        INJECT_MODE_FLAGS+=(--max-tokens 800)
    fi

    # Use bead title as query for relevance-scoped injection
    if [ -n "${HOOK_BEAD:-}" ] && command -v bd &>/dev/null; then
        BEAD_TITLE=$("$AO_TIMEOUT_BIN" 3 bd show "$HOOK_BEAD" --json 2>/dev/null | jq -r '.title // empty' 2>/dev/null || true)
        if [ -n "$BEAD_TITLE" ]; then
            INJECT_EXTRA_FLAGS+=("$BEAD_TITLE")
        fi
    fi

    if ! INJECTED_KNOWLEDGE="$(run_ao_quick 5 inject "${INJECT_MODE_FLAGS[@]}" "${INJECT_EXTRA_FLAGS[@]}")"; then
        log_hook_fail "ao inject"
        INJECTED_KNOWLEDGE=""
    fi
fi

# Keep startup context lean: inject only fresh flywheel knowledge and a short skill pointer.
SKILL_FILE="${PLUGIN_ROOT}/skills/using-agentops/SKILL.md"
if [ -f "$SKILL_FILE" ]; then
    using_agentops_hint="AgentOps workflow context is available. Use the Skill tool to read ${SKILL_FILE} when needed."
else
    using_agentops_hint="(AgentOps skill content unavailable at ${SKILL_FILE})"
fi

# Notebook lean mode: MEMORY.md auto-loaded → shrink injection budget
if [ "${NOTEBOOK_LEAN_MODE:-0}" = "1" ]; then
    MAX_INJECT_CHARS=1500
else
    MAX_INJECT_CHARS=4000
fi
if [ -n "$INJECTED_KNOWLEDGE" ] && [ "${#INJECTED_KNOWLEDGE}" -gt "$MAX_INJECT_CHARS" ]; then
    # Truncate at last newline within budget (never mid-line)
    trimmed="${INJECTED_KNOWLEDGE:0:$MAX_INJECT_CHARS}"
    INJECTED_KNOWLEDGE="${trimmed%
*}

*[truncated by session-start hook]*"
fi

# Nudge agent if MEMORY.md is stale or missing
if [ -n "${MEMORY_FILE:-}" ] && [ -f "$MEMORY_FILE" ] && [ "${MEMORY_AGE_DAYS:-0}" -gt 7 ]; then
    INJECTED_KNOWLEDGE="${INJECTED_KNOWLEDGE}

*Note: Your MEMORY.md hasn't been updated in ${MEMORY_AGE_DAYS} days. Consider running \`ao notebook update\` or updating it manually.*"
fi

if [ -n "$INJECTED_KNOWLEDGE" ]; then
    full_content=$(cat <<HOOKCTX
<AGENTOPS_CONTEXT>
${INJECTED_KNOWLEDGE}

${using_agentops_hint}
</AGENTOPS_CONTEXT>
HOOKCTX
)
else
    full_content=$(cat <<HOOKCTX
<AGENTOPS_CONTEXT>
No prior flywheel knowledge was injected for this session.
${using_agentops_hint}
</AGENTOPS_CONTEXT>
HOOKCTX
)
fi

if command -v jq &>/dev/null; then
    additional_context=$(printf '%s' "$full_content" | jq -Rs '.')
    cat <<HOOKEOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": ${additional_context}
  }
}
HOOKEOF
else
    # Minimal fallback — escape newlines and quotes
    escaped=$(printf '%s' "$full_content" | sed 's/\\/\\\\/g; s/"/\\"/g' | tr '\n' ' ')
    cat <<HOOKEOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "${escaped}"
  }
}
HOOKEOF
fi

exit 0
