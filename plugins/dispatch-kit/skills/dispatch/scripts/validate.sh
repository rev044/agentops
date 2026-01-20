#!/bin/bash
# Validate dispatch skill matches actual gt CLI behavior
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2034  # Reserved for future use
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0

check() {
    local desc="$1"
    local cmd="$2"
    local expected="$3"

    CHECKS=$((CHECKS + 1))
    if eval "$cmd" 2>/dev/null | grep -q "$expected"; then
        echo "✓ $desc"
    else
        echo "✗ $desc"
        echo "  Command: $cmd"
        echo "  Expected to find: $expected"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== Dispatch Skill Validation ==="
echo ""

# Verify gt binary exists
check "gt binary exists" "which gt" "gt"

# Verify gt sling command exists
check "gt sling command documented" "gt sling --help" "Sling work"

# Verify gt hook command exists
check "gt hook command documented" "gt hook --help" "hook"

# Verify gt convoy command exists
check "gt convoy --help" "gt convoy --help" "convoy"

# Verify documented flags exist
check "gt sling target resolution" "gt sling --help" "Target Resolution"
check "gt convoy has create subcommand" "gt convoy --help" "create"
check "gt convoy has list subcommand" "gt convoy --help" "list"

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
