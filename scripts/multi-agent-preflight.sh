#!/usr/bin/env bash
# multi-agent-preflight.sh - Validate distributed multi-agent prerequisites.
# Checks:
#   1) Registration prerequisites
#   2) Quorum inputs
#   3) Claim-lock health

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TASKS_SYNC_SCRIPT="${TASKS_SYNC_SCRIPT:-${SCRIPT_DIR}/tasks-sync.sh}"
AGENT_MAIL_HEALTH_URL="${AGENT_MAIL_HEALTH_URL:-http://localhost:8765/health}"

WORKFLOW="swarm"
MAX_WORKERS="${MAX_WORKERS:-5}"
MIN_QUORUM="${MIN_QUORUM:-1}"
READY_COUNT=""
BEAD_IDS=""
QUIET=0

FAILED=0

usage() {
    cat <<'EOF'
Usage: multi-agent-preflight.sh [options]

Options:
  --workflow <swarm|crank|implement>   Execution workflow (default: swarm)
  --max-workers <n>                    Maximum workers considered for quorum (default: 5)
  --min-quorum <n>                     Minimum required ready workers (default: 1)
  --ready-count <n>                    Explicit ready count (overrides autodetect)
  --bead-ids <id1,id2,...>             Ready bead IDs (used for quorum count)
  --agent-mail-health-url <url>        Agent Mail health endpoint
  --tasks-sync-script <path>           Path to tasks-sync.sh for lock health checks
  --quiet                              Reduce output
  -h, --help                           Show this message
EOF
}

note() {
    if (( QUIET == 0 )); then
        printf "%s\n" "$*"
    fi
}

pass() {
    note "PASS: $1"
}

warn() {
    note "WARN: $1"
}

fail() {
    note "FAIL: $1"
    FAILED=1
}

is_positive_int() {
    [[ "$1" =~ ^[1-9][0-9]*$ ]]
}

is_non_negative_int() {
    [[ "$1" =~ ^[0-9]+$ ]]
}

count_csv_values() {
    local raw="$1"
    if [[ -z "$raw" ]]; then
        echo "0"
        return
    fi

    printf "%s" "$raw" \
        | tr ',' '\n' \
        | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' \
        | awk 'NF {count++} END {print count+0}'
}

detect_ready_count() {
    if [[ -n "$READY_COUNT" ]]; then
        echo "$READY_COUNT"
        return
    fi

    if [[ -n "$BEAD_IDS" ]]; then
        count_csv_values "$BEAD_IDS"
        return
    fi

    if command -v bd >/dev/null 2>&1 && command -v jq >/dev/null 2>&1; then
        local count
        count="$(bd ready --json 2>/dev/null | jq -r 'length' 2>/dev/null || true)"
        if is_non_negative_int "${count:-}"; then
            echo "$count"
            return
        fi
    fi

    echo "0"
}

check_registration_prerequisites() {
    note "[registration] checking prerequisites"

    if command -v claude >/dev/null 2>&1; then
        pass "claude CLI found"
    else
        fail "claude CLI not found in PATH"
    fi

    if command -v curl >/dev/null 2>&1; then
        pass "curl found"
    else
        fail "curl not found in PATH"
    fi

    case "$WORKFLOW" in
        swarm|crank)
            if command -v tmux >/dev/null 2>&1; then
                pass "tmux found"
            else
                fail "tmux required for $WORKFLOW workflow"
            fi
            ;;
        implement)
            if command -v tmux >/dev/null 2>&1; then
                pass "tmux found"
            else
                warn "tmux not found (allowed for standalone implement)"
            fi
            ;;
        *)
            fail "unsupported workflow '$WORKFLOW' (expected swarm|crank|implement)"
            ;;
    esac

    local project_key
    project_key="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
    pass "project key resolved: $project_key"

    if curl -fsS --max-time 2 "$AGENT_MAIL_HEALTH_URL" >/dev/null 2>&1; then
        pass "agent-mail endpoint reachable: $AGENT_MAIL_HEALTH_URL"
    else
        fail "agent-mail endpoint not reachable: $AGENT_MAIL_HEALTH_URL"
    fi

    if [[ -n "${OLYMPUS_DEMIGOD_ID:-}" || -n "${ORCHESTRATOR_ID:-}" ]]; then
        pass "agent identity present in environment"
    else
        warn "no agent identity preset (registration step must set one)"
    fi
}

check_quorum_inputs() {
    note "[quorum] validating inputs"

    if is_positive_int "$MAX_WORKERS"; then
        pass "max-workers is valid: $MAX_WORKERS"
    else
        fail "max-workers must be a positive integer (got '$MAX_WORKERS')"
        return
    fi

    if is_positive_int "$MIN_QUORUM"; then
        pass "min-quorum is valid: $MIN_QUORUM"
    else
        fail "min-quorum must be a positive integer (got '$MIN_QUORUM')"
        return
    fi

    if (( MIN_QUORUM > MAX_WORKERS )); then
        fail "min-quorum ($MIN_QUORUM) cannot exceed max-workers ($MAX_WORKERS)"
        return
    fi

    local ready_count
    ready_count="$(detect_ready_count)"

    if ! is_non_negative_int "$ready_count"; then
        fail "ready count must be a non-negative integer (got '$ready_count')"
        return
    fi

    local considered_ready="$ready_count"
    if (( considered_ready > MAX_WORKERS )); then
        considered_ready="$MAX_WORKERS"
    fi

    if (( considered_ready >= MIN_QUORUM )); then
        pass "quorum satisfied: ready=$ready_count considered=$considered_ready min=$MIN_QUORUM"
    else
        fail "quorum not met: ready=$ready_count considered=$considered_ready min=$MIN_QUORUM"
    fi
}

check_claim_lock_health() {
    note "[claim-lock] checking tasks-sync lock health"

    if [[ ! -x "$TASKS_SYNC_SCRIPT" ]]; then
        fail "tasks-sync script is missing or not executable: $TASKS_SYNC_SCRIPT"
        return
    fi

    local output
    if output="$("$TASKS_SYNC_SCRIPT" lock-health 2>&1)"; then
        pass "$output"
    else
        fail "$output"
    fi
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --workflow)
            WORKFLOW="${2:-}"
            shift 2
            ;;
        --max-workers)
            MAX_WORKERS="${2:-}"
            shift 2
            ;;
        --min-quorum)
            MIN_QUORUM="${2:-}"
            shift 2
            ;;
        --ready-count)
            READY_COUNT="${2:-}"
            shift 2
            ;;
        --bead-ids)
            BEAD_IDS="${2:-}"
            shift 2
            ;;
        --agent-mail-health-url)
            AGENT_MAIL_HEALTH_URL="${2:-}"
            shift 2
            ;;
        --tasks-sync-script)
            TASKS_SYNC_SCRIPT="${2:-}"
            shift 2
            ;;
        --quiet)
            QUIET=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

check_registration_prerequisites
check_quorum_inputs
check_claim_lock_health

if (( FAILED != 0 )); then
    note "Preflight result: FAILED"
    exit 1
fi

note "Preflight result: PASS"
