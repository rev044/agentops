#!/bin/bash
# Validate beads skill matches actual bd CLI behavior
set -euo pipefail

ERRORS=0
CHECKS=0

check() {
    local desc="$1"
    local cmd="$2"
    local expected="$3"

    CHECKS=$((CHECKS + 1))
    if eval "$cmd" 2>/dev/null | grep -qi "$expected"; then
        echo "✓ $desc"
    else
        echo "✗ $desc"
        echo "  Command: $cmd"
        echo "  Expected to find: $expected"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== Beads Skill Validation ==="
echo ""

# Verify bd binary exists
check "bd binary exists" "which bd" "bd"

# Verify bd version
check "bd version works" "bd version" "bd version"

# Verify core commands exist
check "bd new command" "bd new --help" "Create a new"
check "bd show command" "bd show --help" "Show"
check "bd list command" "bd list --help" "List"
check "bd update command" "bd update --help" "Update"
check "bd close command" "bd close --help" "Close"
check "bd ready command" "bd ready --help" "ready"
check "bd sync command" "bd sync --help" "Sync"
check "bd blocked command" "bd blocked --help" "blocked"

# Verify documented flags
check "bd new has --type flag" "bd new --help" "type"
check "bd new has --label flag" "bd new --help" "label"
check "bd list has --status flag" "bd list --help" "status"
check "bd update has --status flag" "bd update --help" "status"

# Verify dependency commands
check "bd dep command exists" "bd dep --help" "dep"
check "bd dep add subcommand" "bd dep --help" "add"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Skill documentation may be out of sync with CLI"
    exit 1
else
    echo ""
    echo "PASS: Skill documentation matches CLI"
    exit 0
fi
