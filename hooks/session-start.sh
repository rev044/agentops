#!/usr/bin/env bash
# AgentOps Session Start Hook (manual mode)
# Creates .agents/ directories, consumes handoffs, injects skill context.
# MEMORY.md is auto-loaded by Claude Code — no extract/inject needed.

# Kill switches
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0
[ "${AGENTOPS_SESSION_START_DISABLED:-}" = "1" ] && exit 0

# Worker environment sanitization
if [[ "${AGENTOPS_WORKER_SESSION:-}" == "1" ]]; then
    # Reset aliases to prevent interference
    unalias -a 2>/dev/null || true
fi

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

cd "$ROOT" 2>/dev/null || true

# Ensure global .agents/ directories exist (cross-repo knowledge)
mkdir -p "$HOME/.agents/learnings" "$HOME/.agents/patterns" 2>/dev/null

# Ensure local .agents/ directories exist
for dir in .agents/research .agents/products .agents/retros .agents/learnings \
           .agents/patterns .agents/council .agents/knowledge/pending \
           .agents/plans .agents/rpi .agents/ao .agents/handoff; do
    mkdir -p "$ROOT/$dir" 2>/dev/null
done

# Clear stale dedup flags from prior sessions (prevents cross-session suppression)
rm -f "$ROOT/.agents/ao/.intent-echo-fired" 2>/dev/null

# Auto-cleanup stale RPI runs (lightweight, <1s, dry-run only)
if command -v ao &>/dev/null; then
    ao rpi cleanup --all --stale-after 24h --dry-run 2>/dev/null | head -5 || true
fi

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

# Flywheel behavior
INJECTED_KNOWLEDGE=""

# Derive MEMORY.md path once (used for stale nudge)
MEMORY_DIR="$HOME/.claude/projects"
PROJECT_PATH=$(printf '%s' "$ROOT" | tr '/' '-')
MEMORY_FILE="$MEMORY_DIR/$PROJECT_PATH/memory/MEMORY.md"
MEMORY_AGE_DAYS=-1
if [ -f "$MEMORY_FILE" ]; then
    MTIME=$(stat -f %m "$MEMORY_FILE" 2>/dev/null || stat -c %Y "$MEMORY_FILE" 2>/dev/null || echo "")
    if [ -n "$MTIME" ]; then
        MEMORY_AGE_DAYS=$(( ($(date +%s) - MTIME) / 86400 ))
    fi
fi

# Structured handoff consumption (ao handoff JSON artifacts)
HANDOFF_CONTEXT=""
if [ -d "$ROOT/.agents/handoff" ] && command -v jq &>/dev/null; then
    # Find newest unconsumed .json handoff (exclude .consumed.json and .consuming.json)
    HANDOFF_JSON=$(find "$ROOT/.agents/handoff" -maxdepth 1 -name 'handoff-*.json' \
        -not -name '*.consumed.json' -not -name '*.consuming.json' 2>/dev/null \
        | sort -r | head -1)
    if [ -n "$HANDOFF_JSON" ] && [ -f "$HANDOFF_JSON" ]; then
        # Atomic claim: mv to .consuming prevents concurrent session race
        CONSUMING="${HANDOFF_JSON%.json}.consuming.json"
        if mv "$HANDOFF_JSON" "$CONSUMING" 2>/dev/null; then
            H_GOAL=$(jq -r '.goal // empty' "$CONSUMING" 2>/dev/null)
            H_SUMMARY=$(jq -r '.summary // empty' "$CONSUMING" 2>/dev/null)
            H_CONTINUATION=$(jq -r '.continuation // empty' "$CONSUMING" 2>/dev/null)
            H_TYPE=$(jq -r '.type // "manual"' "$CONSUMING" 2>/dev/null)
            # Finalize: write consumed metadata and rename to .consumed.json
            CONSUMED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
            jq --arg t "$CONSUMED_AT" '.consumed=true | .consumed_at=$t' \
                "$CONSUMING" > "${CONSUMING}.tmp" 2>/dev/null \
                && mv "${CONSUMING}.tmp" "${HANDOFF_JSON%.json}.consumed.json" 2>/dev/null \
                && rm -f "$CONSUMING" 2>/dev/null
            # Build injection context
            HANDOFF_CONTEXT="### Handoff Context (${H_TYPE})"
            [ -n "$H_GOAL" ] && HANDOFF_CONTEXT="${HANDOFF_CONTEXT}
- **Goal:** ${H_GOAL}"
            [ -n "$H_SUMMARY" ] && HANDOFF_CONTEXT="${HANDOFF_CONTEXT}
- **Summary:** ${H_SUMMARY}"
            [ -n "$H_CONTINUATION" ] && HANDOFF_CONTEXT="${HANDOFF_CONTEXT}
- **Continue:** ${H_CONTINUATION}"
            HANDOFF_CONTEXT="${HANDOFF_CONTEXT}
- **Source:** ${HANDOFF_JSON}"
        fi
    fi
fi

# Predecessor handoff discovery
PREDECESSOR_FILE="${GT_PREDECESSOR_HANDOFF:-}"
if [ -z "$PREDECESSOR_FILE" ] && [ -d "$ROOT/.agents/handoff" ]; then
    PREDECESSOR_FILE=$(ls -t "$ROOT/.agents/handoff/"*.md 2>/dev/null | head -1)
fi

# Build injection context (manual mode — MEMORY.md auto-loaded by Claude Code)
INJECTED_KNOWLEDGE="MEMORY.md is auto-loaded by Claude Code for this project.
For on-demand retrieval: \`ao search \"<query>\"\` or \`ao lookup --query \"<query>\"\`"
if [ -n "$HANDOFF_CONTEXT" ]; then
    INJECTED_KNOWLEDGE="${INJECTED_KNOWLEDGE}

${HANDOFF_CONTEXT}"
fi
if [ -n "$PREDECESSOR_FILE" ] && [ -f "$PREDECESSOR_FILE" ]; then
    INJECTED_KNOWLEDGE="${INJECTED_KNOWLEDGE}
Predecessor handoff: ${PREDECESSOR_FILE}"
fi

# Inject flywheel knowledge with predecessor context pass-through
if command -v ao >/dev/null 2>&1; then
    INJECT_ARGS="--apply-decay --format markdown --max-tokens 1000"
    [ -n "${H_GOAL:-}" ] && INJECT_ARGS="${INJECT_ARGS} --context ${H_GOAL}"
    [ -n "${H_TYPE:-}" ] && INJECT_ARGS="${INJECT_ARGS} --session-type ${H_TYPE}"
    [ -f "${PROJECT_ROOT:-.}/.agents/profile.md" ] && INJECT_ARGS="${INJECT_ARGS} --profile"
    # shellcheck disable=SC2086
    AO_INJECT_OUTPUT=$(timeout 5 ao inject ${INJECT_ARGS} 2>/dev/null) || true
    if [ -n "${AO_INJECT_OUTPUT:-}" ]; then
        INJECTED_KNOWLEDGE="${INJECTED_KNOWLEDGE}

## Injected Knowledge (ao inject)

${AO_INJECT_OUTPUT}"
    fi
fi

# Keep startup context lean: inject only fresh flywheel knowledge and a short skill pointer.
SKILL_FILE="${PLUGIN_ROOT}/skills/using-agentops/SKILL.md"
if [ -f "$SKILL_FILE" ]; then
    using_agentops_hint="AgentOps workflow context is available. Use the Skill tool to read ${SKILL_FILE} when needed."
else
    using_agentops_hint="(AgentOps skill content unavailable at ${SKILL_FILE})"
fi

# Truncation (static cap — no notebook mode logic needed)
MAX_INJECT_CHARS=4000
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
