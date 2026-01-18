#!/bin/bash
# Validate gastown skill matches actual gt CLI behavior
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

echo "=== Gas Town Skill Validation ==="
echo ""

# Verify gt binary exists
check "gt binary exists" "which gt" "gt"

# Verify gt version
check "gt version works" "gt version" "gt version"

# Verify core commands exist
check "gt rig command" "gt rig --help" "rig"
check "gt polecat command" "gt polecat --help" "polecat"
check "gt mail command" "gt mail --help" "mail"
check "gt convoy command" "gt convoy --help" "convoy"
check "gt hook command" "gt hook --help" "hook"
check "gt sling command" "gt sling --help" "sling"
check "gt prime command" "gt prime --help" "prime"
check "gt handoff command" "gt handoff --help" "handoff"

# Verify rig subcommands
check "gt rig list subcommand" "gt rig --help" "list"
check "gt rig add subcommand" "gt rig --help" "add"

# Verify polecat subcommands
check "gt polecat list subcommand" "gt polecat --help" "list"
check "gt polecat spawn subcommand" "gt polecat --help" "spawn"
check "gt polecat nuke subcommand" "gt polecat --help" "nuke"

# Verify mail subcommands
check "gt mail inbox subcommand" "gt mail --help" "inbox"
check "gt mail send subcommand" "gt mail --help" "send"

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
