#!/bin/bash
# Validate pr-plan skill
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2034  # Reserved for future use
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0

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

check_exists() {
    local desc="$1"
    local path="$2"

    CHECKS=$((CHECKS + 1))
    if [ -e "$path" ]; then
        echo "✓ $desc"
    else
        echo "✗ $desc ($path not found)"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== PR Plan Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "PR-research skill exists" "$HOME/.claude/skills/pr-research/SKILL.md"

# Verify pr-plan workflow patterns in SKILL.md
check_pattern "SKILL.md has scope definition" "$SKILL_DIR/SKILL.md" "[Ss]cope"
check_pattern "SKILL.md has acceptance criteria" "$SKILL_DIR/SKILL.md" "[Aa]cceptance.*[Cc]riteria"
check_pattern "SKILL.md has risk assessment" "$SKILL_DIR/SKILL.md" "[Rr]isk"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: PR-plan skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: PR-plan skill validation passed"
    exit 0
fi
