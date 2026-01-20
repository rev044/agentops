#!/bin/bash
# Validate status skill
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
    if eval "$cmd" 2>/dev/null | grep -qi "$expected"; then
        echo "✓ $desc"
    else
        echo "✗ $desc"
        echo "  Command: $cmd"
        echo "  Expected to find: $expected"
        ERRORS=$((ERRORS + 1))
    fi
}

check_pattern() {
    local desc="$1"
    local file="$2"
    local pattern="$3"

    CHECKS=$((CHECKS + 1))
    if grep -qiE "$pattern" "$file" 2>/dev/null; then
        echo "✓ $desc"
    else
        echo "✗ $desc (pattern '$pattern' not found in $file)"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== Status Skill Validation ==="
echo ""


# Verify required CLIs exist
check "bd binary exists" "which bd" "bd"
check "gt binary exists" "which gt" "gt"
check "git binary exists" "which git" "git"

# Verify status workflow patterns in SKILL.md
check_pattern "SKILL.md has work state check" "$SKILL_DIR/SKILL.md" "[Ww]ork.*[Ss]tate|[Ss]napshot"
check_pattern "SKILL.md has git status" "$SKILL_DIR/SKILL.md" "git status|git"
check_pattern "SKILL.md has beads status" "$SKILL_DIR/SKILL.md" "bd|[Bb]eads"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Status skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Status skill validation passed"
    exit 0
fi
