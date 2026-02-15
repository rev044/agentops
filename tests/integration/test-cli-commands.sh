#!/usr/bin/env bash
# CLI command smoke tests for ao
# Tests basic subcommands with minimal setup

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

errors=0

log() { echo -e "${BLUE}[TEST]${NC} $1"; }
pass() { echo -e "${GREEN}  ✓${NC} $1"; }
fail() { echo -e "${RED}  ✗${NC} $1"; ((errors++)) || true; }
warn() { echo -e "${YELLOW}  ⚠${NC} $1"; }

# Pre-flight: check for Go
if ! command -v go &>/dev/null; then
    fail "Go not available — cannot build ao CLI"
    echo -e "${RED}FAILED${NC} - Prerequisites not met"
    exit 1
fi

# Build ao from source
log "Building ao CLI from source..."

TMPDIR="${TMPDIR:-/tmp}"
TMPBIN="$TMPDIR/ao-test-$$"
TMPDIR_TEST="$TMPDIR/ao-test-dir-$$"

# Trap cleanup
cleanup() {
    [[ -f "$TMPBIN" ]] && rm -f "$TMPBIN"
    [[ -d "$TMPDIR_TEST" ]] && rm -rf "$TMPDIR_TEST"
}
trap cleanup EXIT

if (cd "$REPO_ROOT/cli" && go build -o "$TMPBIN" ./cmd/ao 2>/dev/null); then
    pass "Built ao CLI successfully"
else
    fail "go build failed"
    echo -e "${RED}FAILED${NC} - Build failed"
    exit 1
fi

# Set up minimal .agents/ directory for commands that need it
log "Setting up test environment..."
mkdir -p "$TMPDIR_TEST/.agents/learnings"
mkdir -p "$TMPDIR_TEST/.agents/research"
mkdir -p "$TMPDIR_TEST/.agents/pool"
mkdir -p "$TMPDIR_TEST/.agents/rpi"
pass "Created test directory: $TMPDIR_TEST"

# Change to test dir for commands
cd "$TMPDIR_TEST"

# =============================================================================
# Test subcommands
# =============================================================================

test_command() {
    local cmd="$1"
    local name="$2"
    local output
    local exit_code=0

    log "Testing: $name"

    if output=$($cmd 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi

    # Check exit code 0
    if [[ $exit_code -eq 0 ]]; then
        pass "Exit code 0"
    else
        fail "Exit code $exit_code (expected 0)"
        [[ -n "$output" ]] && echo "$output" | head -5 | sed 's/^/    /'
        return
    fi

    # Check non-empty output
    if [[ -n "$output" ]]; then
        pass "Non-empty output (${#output} chars)"
    else
        fail "Empty output"
    fi
}

# Test 1: ao status
test_command "$TMPBIN status" "ao status"

# Test 2: ao version
test_command "$TMPBIN version" "ao version"

# Test 3: ao search
test_command "$TMPBIN search test" "ao search \"test\""

# Test 4: ao ratchet status
test_command "$TMPBIN ratchet status" "ao ratchet status"

# Test 5: ao flywheel status
test_command "$TMPBIN flywheel status" "ao flywheel status"

# Test 6: ao pool list
test_command "$TMPBIN pool list" "ao pool list"

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${BLUE}═══════════════════════════════════════════${NC}"

if [[ $errors -gt 0 ]]; then
    echo -e "${RED}FAILED${NC} - $errors errors"
    exit 1
else
    echo -e "${GREEN}PASSED${NC} - All CLI command tests passed"
    exit 0
fi
