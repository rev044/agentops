#!/usr/bin/env bash
# run-headless-tests.sh — Test AgentOps skills in OpenCode headless mode
#
# Usage:
#   ./run-headless-tests.sh [--tier N] [--skill NAME] [--timeout SECS] [--help]
#
# Runs OpenCode headless (opencode run) against AgentOps skills using the
# configured model (default: devstral/devstral-2). Captures output, exit code,
# duration, and generates per-skill logs.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/.agents/opencode-tests"
DATE=$(date +%Y-%m-%d)
MODEL="${OPENCODE_TEST_MODEL:-devstral/devstral-2}"
TIMEOUT=180
TIER=""
SKILL_FILTER=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  --tier N        Run only tier N (1, 2, or 3)
  --skill NAME    Run only a specific skill
  --timeout SECS  Timeout per test (default: 180)
  --model MODEL   Model to use (default: devstral/devstral-2)
  --help          Show this help

Environment:
  NODE_TLS_REJECT_UNAUTHORIZED  Set to 0 for self-signed certs (auto-set)
  OPENCODE_TEST_MODEL           Override default model
  OH_MY_OPENCODE_DISABLED       Set to 1 to disable oh-my-opencode hooks

Output:
  .agents/opencode-tests/<date>-<skill>.log   Per-skill output
  .agents/opencode-tests/<date>-summary.txt   Test summary
EOF
    exit 0
}

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --tier) TIER="$2"; shift 2 ;;
        --skill) SKILL_FILTER="$2"; shift 2 ;;
        --timeout) TIMEOUT="$2"; shift 2 ;;
        --model) MODEL="$2"; shift 2 ;;
        --help) usage ;;
        *) echo "Unknown option: $1"; usage ;;
    esac
done

mkdir -p "$OUTPUT_DIR"

# TLS bypass for self-signed RunAI certs
export NODE_TLS_REJECT_UNAUTHORIZED=0

# Strip ANSI escape codes from output
strip_ansi() {
    sed 's/\x1b\[[0-9;]*m//g' | sed 's/\x1b\[0m//g' | sed 's/\r//g'
}

# Run a single skill test
run_test() {
    local skill="$1"
    local prompt="$2"
    local tier="$3"
    local logfile="$OUTPUT_DIR/${DATE}-${skill}.log"

    printf "${BLUE}[T${tier}]${NC} Testing %-20s ... " "$skill"

    local start_time
    start_time=$(date +%s)

    # Run opencode headless
    local exit_code=0
    timeout "$TIMEOUT" opencode run \
        -m "$MODEL" \
        --title "test-${skill}" \
        "$prompt" \
        > "$logfile.raw" 2>&1 || exit_code=$?

    local end_time
    end_time=$(date +%s)
    local duration=$(( end_time - start_time ))

    # Strip ANSI codes
    strip_ansi < "$logfile.raw" > "$logfile"
    rm -f "$logfile.raw"

    # Check results
    local output_size
    output_size=$(wc -c < "$logfile" | tr -d ' ')
    local line_count
    line_count=$(wc -l < "$logfile" | tr -d ' ')

    # Determine pass/fail
    local status="FAIL"
    local color="$RED"
    if [[ $exit_code -eq 0 && $output_size -gt 50 ]]; then
        status="PASS"
        color="$GREEN"
    elif [[ $exit_code -eq 124 ]]; then
        status="TIMEOUT"
        color="$YELLOW"
    elif [[ $exit_code -eq 0 && $output_size -le 50 ]]; then
        status="EMPTY"
        color="$YELLOW"
    fi

    printf "${color}%-8s${NC} (exit=%d, %d bytes, %d lines, %ds)\n" \
        "$status" "$exit_code" "$output_size" "$line_count" "$duration"

    # Append to summary
    printf "%-20s %-8s tier=%d exit=%d bytes=%d lines=%d duration=%ds\n" \
        "$skill" "$status" "$tier" "$exit_code" "$output_size" "$line_count" "$duration" \
        >> "$OUTPUT_DIR/${DATE}-summary.txt"

    return 0
}

# Define test cases
# Format: "skill|prompt|tier"
declare -a TESTS=()

# Tier 1: Should work (read-only / tool-independent)
TIER1_TESTS=(
    "status|Load the status skill and run it. Show current project status.|1"
    "knowledge|Load the knowledge skill and search for 'council patterns'. Show results.|1"
    "complexity|Load the complexity skill and analyze the file skills/council/SKILL.md for complexity.|1"
    "doc|Load the doc skill and check documentation coverage for skills/research/ directory.|1"
    "handoff|Load the handoff skill and create a handoff summary for this test session.|1"
    "retro|Load the retro skill and extract learnings from the most recent work in .agents/learnings/.|1"
)

# Tier 2: Degraded mode (fallback possible)
TIER2_TESTS=(
    "research|Load the research skill and research the testing infrastructure in this repo. Use --auto mode. Do inline exploration only, do not spawn agents.|2"
    "plan|Load the plan skill and plan adding a README badge for test coverage. Use --auto mode.|2"
    "pre-mortem|Load the pre-mortem skill with --quick and validate the most recent plan in .agents/plans/.|2"
    "implement|Load the implement skill. Check what beads issues are ready to work on using bd ready.|2"
    "vibe|Load the vibe skill with --quick and review the file skills/status/SKILL.md for quality.|2"
    "bug-hunt|Load the bug-hunt skill and investigate any test failures in the tests/ directory.|2"
    "learn|Load the learn skill and save this insight: OpenCode headless mode requires NODE_TLS_REJECT_UNAUTHORIZED=0 for self-signed certs on RunAI clusters.|2"
    "trace|Load the trace skill and trace the decision history for the council architecture in this project.|2"
)

# Tier 3: Expected failure (hard blockers)
TIER3_TESTS=(
    "council|Load the council skill and validate skills/status/SKILL.md using multi-model consensus.|3"
    "crank|Load the crank skill. Show what epic would need to be cranked.|3"
    "swarm|Load the swarm skill and describe what it would do to spawn 2 workers.|3"
    "rpi|Load the rpi skill. Show what the full RPI lifecycle would look like for this project.|3"
    "codex-team|Load the codex-team skill and describe what it would do to spawn 2 codex agents.|3"
)

# Filter tests by tier
for test in "${TIER1_TESTS[@]}"; do
    if [[ -z "$TIER" || "$TIER" == "1" ]]; then
        TESTS+=("$test")
    fi
done
for test in "${TIER2_TESTS[@]}"; do
    if [[ -z "$TIER" || "$TIER" == "2" ]]; then
        TESTS+=("$test")
    fi
done
for test in "${TIER3_TESTS[@]}"; do
    if [[ -z "$TIER" || "$TIER" == "3" ]]; then
        TESTS+=("$test")
    fi
done

# Header
echo ""
echo "================================================================"
echo " OpenCode Headless Skills Test"
echo " Model: $MODEL"
echo " Date:  $DATE"
echo " Repo:  $REPO_ROOT"
echo " Tests: ${#TESTS[@]}"
echo " Timeout: ${TIMEOUT}s per test"
echo "================================================================"
echo ""

# Clear summary
> "$OUTPUT_DIR/${DATE}-summary.txt"

# Run tests
pass=0
fail=0
warn=0
total=0

for test_spec in "${TESTS[@]}"; do
    IFS='|' read -r skill prompt tier <<< "$test_spec"

    # Filter by skill name if specified
    if [[ -n "$SKILL_FILTER" && "$skill" != "$SKILL_FILTER" ]]; then
        continue
    fi

    total=$((total + 1))
    run_test "$skill" "$prompt" "$tier"

    # Count results from summary
    last_status=$(tail -1 "$OUTPUT_DIR/${DATE}-summary.txt" | awk '{print $2}')
    case "$last_status" in
        PASS) pass=$((pass + 1)) ;;
        FAIL) fail=$((fail + 1)) ;;
        *) warn=$((warn + 1)) ;;
    esac
done

# Footer
echo ""
echo "================================================================"
echo " Results: ${pass} PASS / ${warn} WARN / ${fail} FAIL  (${total} total)"
echo " Summary: $OUTPUT_DIR/${DATE}-summary.txt"
echo " Logs:    $OUTPUT_DIR/${DATE}-*.log"
echo "================================================================"
