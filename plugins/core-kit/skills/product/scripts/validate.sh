#!/bin/bash
# Validate product skill
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

echo "=== Product Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Research skill exists" "$HOME/.claude/skills/research/SKILL.md"

# Verify product workflow patterns in SKILL.md
check_pattern "SKILL.md has customer-first approach" "$SKILL_DIR/SKILL.md" "[Cc]ustomer"
check_pattern "SKILL.md has PR/FAQ documentation" "$SKILL_DIR/SKILL.md" "PR/FAQ|[Pp]ress [Rr]elease"
check_pattern "SKILL.md has success criteria" "$SKILL_DIR/SKILL.md" "[Ss]uccess.*[Cc]riteria"
check_pattern "SKILL.md has scope documentation" "$SKILL_DIR/SKILL.md" "[Ss]cope"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Product skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Product skill validation passed"
    exit 0
fi
