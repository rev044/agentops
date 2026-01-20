#!/bin/bash
# Validate code-map-standard skill
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

echo "=== Code-Map-Standard Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify skill references exist
check_exists "References directory" "$SKILL_DIR/references"
if [ -d "$SKILL_DIR/references" ]; then
    # Check for template reference
    if ls "$SKILL_DIR/references/"*.md >/dev/null 2>&1; then
        echo "✓ Reference templates exist"
        CHECKS=$((CHECKS + 1))
    else
        echo "✗ No reference templates found"
        ERRORS=$((ERRORS + 1))
        CHECKS=$((CHECKS + 1))
    fi
fi

# Verify code-map workflow patterns in SKILL.md
check_pattern "SKILL.md has code-map documentation" "$SKILL_DIR/SKILL.md" "code-map|[Cc]ode [Mm]ap"
check_pattern "SKILL.md has feature documentation" "$SKILL_DIR/SKILL.md" "[Ff]eature"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Code-map-standard skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Code-map-standard skill validation passed"
    exit 0
fi
