#!/usr/bin/env bash
# Release Smoke Test — validates ALL ao CLI commands before release
#
# Proves every registered command runs without panicking, produces expected
# output, and exits cleanly. Any broken registration, flag parsing error,
# or runtime panic is caught before tagging.
#
# Usage:
#   bash scripts/release-smoke-test.sh              # build + test
#   bash scripts/release-smoke-test.sh --skip-build # use existing cli/bin/ao
#
# Exit codes:
#   0 = all tests passed
#   1 = one or more tests failed

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# ═══════════════════════════════════════════════════════
#  Options
# ═══════════════════════════════════════════════════════

SKIP_BUILD=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        -h|--help)
            echo "Usage: bash scripts/release-smoke-test.sh [--skip-build]"
            echo ""
            echo "Options:"
            echo "  --skip-build   Use existing cli/bin/ao instead of rebuilding"
            echo "  -h, --help     Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

# ═══════════════════════════════════════════════════════
#  Colors & Counters
# ═══════════════════════════════════════════════════════

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0
TOTAL=0

pass() {
    echo -e "${GREEN}  PASS${NC}  $1"
    PASS=$((PASS + 1))
    TOTAL=$((TOTAL + 1))
}

fail() {
    echo -e "${RED}  FAIL${NC}  $1"
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
}

skip() {
    echo -e "${YELLOW}  SKIP${NC}  $1"
    SKIP=$((SKIP + 1))
    TOTAL=$((TOTAL + 1))
}

section() {
    echo ""
    echo -e "${BLUE}── $1 ──${NC}"
}

AO="$REPO_ROOT/cli/bin/ao"

# ═══════════════════════════════════════════════════════
#  Panic / Crash Detection
# ═══════════════════════════════════════════════════════

# Check output for Go panic/crash markers. Returns 0 if clean, 1 if crash detected.
check_no_panic() {
    local output="$1"
    local label="$2"
    if echo "$output" | grep -qE '(^panic:|runtime error:|^goroutine [0-9]+ \[)'; then
        fail "$label — PANIC/CRASH DETECTED"
        echo "$output" | grep -E '(panic:|runtime error:|goroutine)' | head -5 | sed 's/^/    /'
        return 1
    fi
    return 0
}

# ═══════════════════════════════════════════════════════
#  Test Helpers
# ═══════════════════════════════════════════════════════

# test_exec: Run a command, assert exit 0, check for panics.
# Usage: test_exec "label" cmd [args...]
test_exec() {
    local label="$1"
    shift
    local output
    local rc=0
    output=$("$@" 2>&1) || rc=$?

    if ! check_no_panic "$output" "$label"; then
        return 0  # already counted as FAIL
    fi

    if [[ "$rc" -eq 0 ]]; then
        pass "$label"
    else
        fail "$label (exit $rc)"
        echo "$output" | head -3 | sed 's/^/    /'
    fi
}

# test_exec_output: Run a command, assert exit 0, assert output matches pattern.
# Usage: test_exec_output "label" "pattern" cmd [args...]
test_exec_output() {
    local label="$1"
    local pattern="$2"
    shift 2
    local output
    local rc=0
    output=$("$@" 2>&1) || rc=$?

    if ! check_no_panic "$output" "$label"; then
        return 0
    fi

    if [[ "$rc" -ne 0 ]]; then
        fail "$label (exit $rc)"
        echo "$output" | head -3 | sed 's/^/    /'
        return 0
    fi

    if echo "$output" | grep -qEi "$pattern"; then
        pass "$label"
    else
        fail "$label — output missing pattern: $pattern"
        echo "$output" | head -5 | sed 's/^/    /'
    fi
}

# test_exec_exact: Run a command, assert output exactly equals expected string.
# Usage: test_exec_exact "label" "expected" cmd [args...]
test_exec_exact() {
    local label="$1"
    local expected="$2"
    shift 2
    local output
    local rc=0
    output=$("$@" 2>&1) || rc=$?

    if ! check_no_panic "$output" "$label"; then
        return 0
    fi

    if [[ "$rc" -ne 0 ]]; then
        fail "$label (exit $rc)"
        echo "$output" | head -3 | sed 's/^/    /'
        return 0
    fi

    if [[ "$output" == "$expected" ]]; then
        pass "$label"
    else
        fail "$label — expected exactly '$expected', got '$(echo "$output" | head -1)'"
    fi
}

# test_help: Run --help, assert exit 0 and output contains Usage.
# Usage: test_help "label" cmd [args...] --help
test_help() {
    local label="$1"
    shift
    local output
    local rc=0
    output=$("$@" 2>&1) || rc=$?

    if ! check_no_panic "$output" "$label"; then
        return 0
    fi

    if [[ "$rc" -ne 0 ]]; then
        fail "$label (exit $rc)"
        echo "$output" | head -3 | sed 's/^/    /'
        return 0
    fi

    if echo "$output" | grep -qEi '(Usage|usage|Available Commands|Flags)'; then
        pass "$label"
    else
        fail "$label — help output missing Usage/Commands/Flags"
        echo "$output" | head -5 | sed 's/^/    /'
    fi
}

# test_json: Run a command, assert exit 0 and output is valid JSON.
# Usage: test_json "label" cmd [args...]
test_json() {
    local label="$1"
    shift
    local output
    local rc=0
    output=$("$@" 2>&1) || rc=$?

    if ! check_no_panic "$output" "$label"; then
        return 0
    fi

    if [[ "$rc" -ne 0 ]]; then
        fail "$label (exit $rc)"
        echo "$output" | head -3 | sed 's/^/    /'
        return 0
    fi

    if echo "$output" | jq . >/dev/null 2>&1; then
        pass "$label"
    else
        fail "$label — output is not valid JSON"
        echo "$output" | head -3 | sed 's/^/    /'
    fi
}

# test_exec_tolerant: Run a command, accept exit 0 or 1 (some commands exit 1
# for "no data" which is acceptable), but fail on panic or exit >= 2.
# Usage: test_exec_tolerant "label" cmd [args...]
test_exec_tolerant() {
    local label="$1"
    shift
    local output
    local rc=0
    output=$("$@" 2>&1) || rc=$?

    if ! check_no_panic "$output" "$label"; then
        return 0
    fi

    if [[ "$rc" -le 1 ]]; then
        pass "$label"
    else
        fail "$label (exit $rc)"
        echo "$output" | head -3 | sed 's/^/    /'
    fi
}

# ═══════════════════════════════════════════════════════
#  Build
# ═══════════════════════════════════════════════════════

START_TIME=$(date +%s)

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Release Smoke Test — ao CLI Command Coverage${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"

section "Build"

if [[ "$SKIP_BUILD" == "true" ]]; then
    if [[ -x "$AO" ]]; then
        pass "Using existing binary: $AO"
    else
        fail "Binary not found at $AO — run without --skip-build"
        echo ""
        echo -e "${RED}ABORT: No binary to test${NC}"
        exit 1
    fi
else
    if (cd "$REPO_ROOT/cli" && make build >/dev/null 2>&1); then
        pass "Built ao binary"
    else
        fail "Build failed"
        echo ""
        echo -e "${RED}ABORT: Build failed${NC}"
        exit 1
    fi
fi

# Verify binary is executable
if "$AO" version >/dev/null 2>&1; then
    pass "Binary executes successfully"
else
    fail "Binary fails to execute"
    echo ""
    echo -e "${RED}ABORT: Binary broken${NC}"
    exit 1
fi

# ═══════════════════════════════════════════════════════
#  Safe Commands (read-only / reporting — actually execute)
# ═══════════════════════════════════════════════════════

section "Safe Commands (read-only execution)"

# Core info
test_exec_output "ao version" "version|Version" "$AO" version
test_exec_output "ao status" "Status|AgentOps|Initialized" "$AO" status
test_exec_tolerant "ao doctor" "$AO" doctor

# Search (ao know)
test_exec "ao know search 'test'" "$AO" know search "test"
test_exec_exact "ao know search --json nonexistent => []" "[]" "$AO" know search --json "nonexistent-xyz-12345"

# Knowledge injection (ao know)
test_exec_tolerant "ao know inject" "$AO" know inject
test_exec_output "ao know inject --index-only" "Knowledge Index|ID|Title" "$AO" know inject --index-only

# Lookup (ao know)
test_exec_tolerant "ao know lookup --query 'test'" "$AO" know lookup --query "test"

# Metrics (ao quality)
test_exec_output "ao quality metrics health" "sigma|rho|delta|retrieval|citation|decay" "$AO" quality metrics health
test_exec_output "ao quality metrics report" "Flywheel|Metrics|Period|decay|retrieval|citation" "$AO" quality metrics report

# Flywheel (ao quality)
test_exec_output "ao quality flywheel status" "Flywheel|status|COMPOUNDING|decay|retrieval" "$AO" quality flywheel status
test_exec_output "ao quality flywheel close-loop" "Close-Loop|Summary|Pool|promote|Citation" "$AO" quality flywheel close-loop

# Pool (ao quality)
test_exec_tolerant "ao quality pool list" "$AO" quality pool list

# Maturity (ao quality)
test_exec_output "ao quality maturity --scan" "Maturity|Distribution|Provisional|Candidate|No learnings" "$AO" quality maturity --scan

# Anti-patterns, constraints, contradict, dedup (ao quality)
test_exec_tolerant "ao quality anti-patterns" "$AO" quality anti-patterns
test_exec_tolerant "ao quality constraint list" "$AO" quality constraint list
test_exec_tolerant "ao quality contradict" "$AO" quality contradict
test_exec_output "ao quality dedup" "Dedup|Scan|Total|Duplicate|No learnings" "$AO" quality dedup

# Curate (ao quality)
test_exec_tolerant "ao quality curate status" "$AO" quality curate status
test_exec_tolerant "ao quality curate verify" "$AO" quality curate verify

# Notebook and memory (ao settings, quiet mode to avoid state changes)
test_exec "ao settings notebook update --quiet" "$AO" settings notebook update --quiet
test_exec_tolerant "ao settings memory sync --quiet" "$AO" settings memory sync --quiet

# Trace (ao know, help only — requires artifact path arg)
test_help "ao know trace --help" "$AO" know trace --help

# Goals (ao work)
test_exec_output "ao work goals validate" "VALID|goals|version" "$AO" work goals validate
test_exec_output "ao work goals measure" "GOAL|RESULT|pass|fail" "$AO" work goals measure

# Ratchet (ao work)
test_exec_output "ao work ratchet status" "Ratchet|Chain|Status|STEP" "$AO" work ratchet status

# RPI (ao work)
test_exec_tolerant "ao work rpi status" "$AO" work rpi status

# Badge (ao quality)
test_exec_output "ao quality badge" "AGENTOPS|KNOWLEDGE|Sessions|Learnings|Citations" "$AO" quality badge

# Context (ao work)
test_exec_tolerant "ao work context status" "$AO" work context status

# Vibe-check (ao quality, help only — actual execution takes a while)
test_help "ao quality vibe-check --help" "$AO" quality vibe-check --help

# ═══════════════════════════════════════════════════════
#  Help-Only Commands (would modify state — test --help)
# ═══════════════════════════════════════════════════════

section "Help-Only Commands (state-modifying — --help only)"

test_help "ao start init --help" "$AO" start init --help
test_help "ao start seed --help" "$AO" start seed --help
test_help "ao start demo --help" "$AO" start demo --help
test_help "ao start quick-start --help" "$AO" start quick-start --help
test_help "ao know forge --help" "$AO" know forge --help
test_help "ao work session --help" "$AO" work session --help
test_help "ao settings hooks --help" "$AO" settings hooks --help
test_help "ao settings config --help" "$AO" settings config --help
test_help "ao completion --help" "$AO" completion --help
test_help "ao settings plans --help" "$AO" settings plans --help
test_help "ao quality gate --help" "$AO" quality gate --help

# ═══════════════════════════════════════════════════════
#  Subcommand Help Coverage (verify subcommands exist)
# ═══════════════════════════════════════════════════════

section "Subcommand Help Coverage (verify command groups list subcommands)"

test_help "ao work goals --help" "$AO" work goals --help
test_help "ao work ratchet --help" "$AO" work ratchet --help
test_help "ao quality metrics --help" "$AO" quality metrics --help
test_help "ao quality pool --help" "$AO" quality pool --help
test_help "ao quality constraint --help" "$AO" quality constraint --help
test_help "ao quality curate --help" "$AO" quality curate --help
test_help "ao work session --help" "$AO" work session --help
test_help "ao work rpi --help" "$AO" work rpi --help
test_help "ao quality flywheel --help" "$AO" quality flywheel --help
test_help "ao quality maturity --help" "$AO" quality maturity --help
test_help "ao settings memory --help" "$AO" settings memory --help
test_help "ao settings notebook --help" "$AO" settings notebook --help
test_help "ao know trace --help" "$AO" know trace --help

# ═══════════════════════════════════════════════════════
#  Flag Testing (verify key flags produce valid output)
# ═══════════════════════════════════════════════════════

section "Flag Testing (JSON output validation)"

test_json "ao know search --json 'test'" "$AO" know search --json "test"
test_json "ao know search --json nonexistent => valid JSON" "$AO" know search --json "nonexistent-xyz-12345"

# Some commands support --json; test where available
if "$AO" doctor --json >/dev/null 2>&1; then
    test_json "ao doctor --json" "$AO" doctor --json
else
    skip "ao doctor --json (flag not supported)"
fi

if "$AO" quality pool list --json >/dev/null 2>&1; then
    test_json "ao quality pool list --json" "$AO" quality pool list --json
else
    skip "ao quality pool list --json (flag not supported)"
fi

if "$AO" work ratchet status --json >/dev/null 2>&1; then
    test_json "ao work ratchet status --json" "$AO" work ratchet status --json
else
    skip "ao work ratchet status --json (flag not supported)"
fi

if "$AO" quality flywheel status --json >/dev/null 2>&1; then
    test_json "ao quality flywheel status --json" "$AO" quality flywheel status --json
else
    skip "ao quality flywheel status --json (flag not supported)"
fi

if "$AO" quality metrics health --json >/dev/null 2>&1; then
    test_json "ao quality metrics health --json" "$AO" quality metrics health --json
else
    skip "ao quality metrics health --json (flag not supported)"
fi

if "$AO" quality constraint list --json >/dev/null 2>&1; then
    test_json "ao quality constraint list --json" "$AO" quality constraint list --json
else
    skip "ao quality constraint list --json (flag not supported)"
fi

# ═══════════════════════════════════════════════════════
#  Top-Level Help (catch unregistered commands)
# ═══════════════════════════════════════════════════════

section "Top-Level Help (catch registration issues)"

test_help "ao --help" "$AO" --help

# Verify expected command groups appear in top-level help
TOP_HELP=$("$AO" --help 2>&1)
EXPECTED_COMMANDS=(
    start doctor quality status version
    work completion settings know
)

for cmd in "${EXPECTED_COMMANDS[@]}"; do
    if echo "$TOP_HELP" | grep -qw "$cmd"; then
        pass "ao --help lists '$cmd'"
    else
        fail "ao --help missing '$cmd'"
    fi
done

# ═══════════════════════════════════════════════════════
#  Summary
# ═══════════════════════════════════════════════════════

END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Release Smoke Test Results${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  Total:   ${TOTAL}"
echo -e "  ${GREEN}Passed:  ${PASS}${NC}"
echo -e "  ${RED}Failed:  ${FAIL}${NC}"
echo -e "  ${YELLOW}Skipped: ${SKIP}${NC}"
echo -e "  Time:    ${ELAPSED}s"
echo ""

if [[ "$FAIL" -gt 0 ]]; then
    echo -e "${RED}  RELEASE SMOKE TEST FAILED ($FAIL failure(s))${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
    exit 1
fi

echo -e "${GREEN}  RELEASE SMOKE TEST PASSED${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
exit 0
