#!/usr/bin/env bash
# run-rpi-headless.sh — Chain RPI phases as separate headless Claude invocations
#
# Each phase invokes the actual AgentOps skills (/research, /plan, /crank, /vibe, etc.)
# so the full skill chain executes — council validation, knowledge flywheel, beads
# tracking, and all other skill behavior works exactly as in interactive mode.
#
# Skills use unpredictable tools internally (WebFetch, WebSearch, Agent sub-agents,
# etc.), so --dangerously-skip-permissions is required. Timeouts and --max-budget-usd
# provide the safety rails instead of tool scoping.
#
# Usage:
#   bin/run-rpi-headless.sh "add user authentication"
#   bin/run-rpi-headless.sh --from=implementation ag-5k2
#   bin/run-rpi-headless.sh --phase=1 "add user auth"     # discovery only
#   bin/run-rpi-headless.sh --phase=2 ag-5k2               # implementation only
#   bin/run-rpi-headless.sh --phase=3                       # validation only
#
# Environment:
#   CLAUDE_MODEL        — Model to use (default: unset, uses Claude default)
#   PHASE1_TIMEOUT      — Phase 1 timeout in seconds (default: 600)
#   PHASE2_TIMEOUT      — Phase 2 timeout in seconds (default: 900)
#   PHASE3_TIMEOUT      — Phase 3 timeout in seconds (default: 600)
#   PHASE1_MAX_TURNS    — Phase 1 max turns (default: 15)
#   PHASE2_MAX_TURNS    — Phase 2 max turns (default: 30)
#   PHASE3_MAX_TURNS    — Phase 3 max turns (default: 15)
#   MAX_BUDGET_USD      — Per-phase cost guardrail (default: 5.00)
#   RPI_PLUGIN_DIR      — Plugin directory (default: auto-detect repo root)
#   RPI_DRY_RUN         — If set, print commands without executing
#   RPI_VERBOSE         — If set, use --verbose and --output-format stream-json

set -euo pipefail

# --- Configuration ---

PHASE1_TIMEOUT="${PHASE1_TIMEOUT:-600}"
PHASE2_TIMEOUT="${PHASE2_TIMEOUT:-900}"
PHASE3_TIMEOUT="${PHASE3_TIMEOUT:-600}"
PHASE1_MAX_TURNS="${PHASE1_MAX_TURNS:-15}"
PHASE2_MAX_TURNS="${PHASE2_MAX_TURNS:-30}"
PHASE3_MAX_TURNS="${PHASE3_MAX_TURNS:-15}"
MAX_BUDGET_USD="${MAX_BUDGET_USD:-5.00}"
MAX_RETRIES=2

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="${RPI_PLUGIN_DIR:-$(cd "$SCRIPT_DIR/.." && pwd)}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# --- Parse arguments ---

START_PHASE=1
END_PHASE=3
GOAL=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --from=discovery|--phase=1)
            START_PHASE=1; END_PHASE=1; shift ;;
        --from=implementation|--from=crank|--phase=2)
            START_PHASE=2; END_PHASE=2; shift ;;
        --from=validation|--from=vibe|--phase=3)
            START_PHASE=3; END_PHASE=3; shift ;;
        --from=*)
            echo "ERROR: Unknown --from value: $1" >&2; exit 1 ;;
        --phase=*)
            echo "ERROR: Unknown --phase value: $1" >&2; exit 1 ;;
        --all)
            START_PHASE=1; END_PHASE=3; shift ;;
        --help|-h)
            echo "Usage: $0 [--from=<phase>|--phase=<n>|--all] <goal-or-epic-id>"
            echo ""
            echo "Phases: discovery (1), implementation (2), validation (3)"
            echo "Default: runs all 3 phases sequentially"
            echo ""
            echo "Examples:"
            echo "  $0 \"add user authentication\"        # full RPI"
            echo "  $0 --from=implementation ag-5k2      # crank existing epic"
            echo "  $0 --phase=3                          # validate recent changes"
            echo ""
            echo "Environment vars: PHASE{1,2,3}_TIMEOUT, PHASE{1,2,3}_MAX_TURNS,"
            echo "  MAX_BUDGET_USD, CLAUDE_MODEL, RPI_DRY_RUN, RPI_VERBOSE"
            exit 0
            ;;
        -*)
            echo "ERROR: Unknown flag: $1" >&2; exit 1 ;;
        *)
            GOAL="$1"; shift ;;
    esac
done

# Phase 3 (validation) can run without a goal — it reviews recent changes
if [[ -z "$GOAL" && $START_PHASE -lt 3 ]]; then
    echo "ERROR: Goal is required for phases 1-2" >&2
    echo "Usage: $0 [--from=<phase>] <goal-or-epic-id>" >&2
    exit 1
fi

# --- Helpers ---

log() { echo -e "${BLUE}[RPI]${NC} $1"; }
ok()  { echo -e "${GREEN}[RPI]${NC} $1"; }
err() { echo -e "${RED}[RPI]${NC} $1" >&2; }
warn(){ echo -e "${YELLOW}[RPI]${NC} $1"; }

# Build claude command as an array (preserves spaces in arguments)
# Skills need full tool access — they chain into sub-skills that use
# unpredictable tools (WebFetch, WebSearch, Agent, etc.).
build_claude_cmd() {
    local max_turns="$1"
    CLAUDE_CMD=(claude -p)

    CLAUDE_CMD+=(--plugin-dir "$REPO_ROOT")
    CLAUDE_CMD+=(--dangerously-skip-permissions)
    CLAUDE_CMD+=(--max-turns "$max_turns")
    CLAUDE_CMD+=(--no-session-persistence)
    CLAUDE_CMD+=(--max-budget-usd "$MAX_BUDGET_USD")

    if [[ -n "${CLAUDE_MODEL:-}" ]]; then
        CLAUDE_CMD+=(--model "$CLAUDE_MODEL")
    fi

    if [[ -n "${RPI_VERBOSE:-}" ]]; then
        CLAUDE_CMD+=(--output-format stream-json --verbose)
    else
        CLAUDE_CMD+=(--output-format json)
    fi
}

# Run a phase with retry logic
run_phase() {
    local phase_num="$1"
    local prompt="$2"
    local max_turns="$3"
    local timeout_s="$4"

    local attempt=1
    local exit_code=0

    while [[ $attempt -le $((MAX_RETRIES + 1)) ]]; do
        if [[ $attempt -gt 1 ]]; then
            warn "Phase $phase_num: retry $((attempt - 1))/$MAX_RETRIES"
        fi

        build_claude_cmd "$max_turns"

        if [[ -n "${RPI_DRY_RUN:-}" ]]; then
            echo "[DRY RUN] timeout $timeout_s ${CLAUDE_CMD[*]} \"${prompt:0:100}...\""
            return 0
        fi

        local output_file
        output_file="$(mktemp)"

        if timeout "$timeout_s" "${CLAUDE_CMD[@]}" "$prompt" > "$output_file" 2>&1; then
            # Extract result from JSON output
            if [[ -z "${RPI_VERBOSE:-}" ]]; then
                local result
                result=$(jq -r '.result // empty' "$output_file" 2>/dev/null) || true
                if [[ -n "$result" ]]; then
                    echo "$result"
                else
                    cat "$output_file"
                fi
            else
                cat "$output_file"
            fi
            rm -f "$output_file"
            return 0
        fi

        exit_code=$?
        if [[ $exit_code -eq 124 ]]; then
            warn "Phase $phase_num: timed out after ${timeout_s}s (attempt $attempt)"
        else
            warn "Phase $phase_num: failed with exit code $exit_code (attempt $attempt)"
        fi

        # Show last output on failure
        if [[ -f "$output_file" ]]; then
            tail -20 "$output_file" >&2 || true
            rm -f "$output_file"
        fi

        attempt=$((attempt + 1))
    done

    err "Phase $phase_num: failed after $((MAX_RETRIES + 1)) attempts"
    return "$exit_code"
}

# --- Prompts (invoke actual skills) ---

SLUG=$(echo "${GOAL:-validation}" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -cd 'a-z0-9-' | head -c 40)
DATE=$(date +%Y-%m-%d)
STATE_DIR=".agents/rpi"
mkdir -p "$STATE_DIR"

# Phase 1: Discovery — invoke /research, /plan, /pre-mortem
PHASE1_PROMPT="Run RPI Phase 1 (Discovery) for: ${GOAL}

Execute these skills in order:
1. /research ${GOAL}
2. /plan ${GOAL}
3. /pre-mortem

After all three complete, write a summary to ${STATE_DIR}/phase-1-summary-${DATE}-${SLUG}.md
with the epic ID and pre-mortem verdict.

Do NOT ask questions — execute autonomously. Do NOT invoke /rpi itself."

# Phase 2: Implementation — invoke /crank with the epic
PHASE2_PROMPT="Run RPI Phase 2 (Implementation) for: ${GOAL}

Read ${STATE_DIR}/phase-1-summary-${DATE}-${SLUG}.md to find the epic ID.
Then run: /crank <epic-id>

If no phase-1 summary exists, find the latest open epic with: bd list --type epic --status open

After crank completes, write a summary to ${STATE_DIR}/phase-2-summary-${DATE}-${SLUG}.md

Do NOT ask questions — execute autonomously. Do NOT invoke /rpi itself."

# Phase 3: Validation — invoke /vibe and /post-mortem
PHASE3_PROMPT="Run RPI Phase 3 (Validation).

Execute these skills in order:
1. /vibe recent
2. /post-mortem

After both complete, write a summary to ${STATE_DIR}/phase-3-summary-${DATE}-${SLUG}.md
with the vibe verdict and any learnings extracted.

Do NOT ask questions — execute autonomously. Do NOT invoke /rpi itself."

# --- Main ---

echo ""
echo -e "${BLUE}═══════════════════════════════════════════${NC}"
echo -e "${BLUE}  Headless RPI: ${GOAL:-validation}${NC}"
echo -e "${BLUE}  Phases: ${START_PHASE}–${END_PHASE}${NC}"
echo -e "${BLUE}═══════════════════════════════════════════${NC}"
echo ""

overall_exit=0

if [[ $START_PHASE -le 1 && $END_PHASE -ge 1 ]]; then
    log "Phase 1: Discovery (timeout: ${PHASE1_TIMEOUT}s, turns: ${PHASE1_MAX_TURNS})"
    if run_phase 1 "$PHASE1_PROMPT" "$PHASE1_MAX_TURNS" "$PHASE1_TIMEOUT"; then
        ok "Phase 1: DONE"
    else
        err "Phase 1: FAILED"
        overall_exit=1
    fi
    echo ""
fi

if [[ $START_PHASE -le 2 && $END_PHASE -ge 2 && $overall_exit -eq 0 ]]; then
    log "Phase 2: Implementation (timeout: ${PHASE2_TIMEOUT}s, turns: ${PHASE2_MAX_TURNS})"
    if run_phase 2 "$PHASE2_PROMPT" "$PHASE2_MAX_TURNS" "$PHASE2_TIMEOUT"; then
        ok "Phase 2: DONE"
    else
        err "Phase 2: FAILED"
        overall_exit=1
    fi
    echo ""
fi

if [[ $START_PHASE -le 3 && $END_PHASE -ge 3 && $overall_exit -eq 0 ]]; then
    log "Phase 3: Validation (timeout: ${PHASE3_TIMEOUT}s, turns: ${PHASE3_MAX_TURNS})"
    if run_phase 3 "$PHASE3_PROMPT" "$PHASE3_MAX_TURNS" "$PHASE3_TIMEOUT"; then
        ok "Phase 3: DONE"
    else
        err "Phase 3: FAILED"
        overall_exit=1
    fi
    echo ""
fi

echo -e "${BLUE}═══════════════════════════════════════════${NC}"
if [[ $overall_exit -eq 0 ]]; then
    echo -e "${GREEN}  Headless RPI: ALL PHASES COMPLETE${NC}"
else
    echo -e "${RED}  Headless RPI: FAILED (see output above)${NC}"
fi
echo -e "${BLUE}═══════════════════════════════════════════${NC}"

exit "$overall_exit"
