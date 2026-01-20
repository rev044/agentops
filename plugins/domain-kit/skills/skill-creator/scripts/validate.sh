#!/bin/bash
# Validate skill-creator skill
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

echo "=== Skill-Creator Skill Validation ==="
echo ""


# Check SKILL.md exists
check_exists "SKILL.md exists" "$SKILL_DIR/SKILL.md"

# Check comprehensive-skill-pattern reference (the only real reference)
check_exists "Comprehensive skill pattern reference" "$SKILL_DIR/references/comprehensive-skill-pattern.md"

# Check for key patterns in SKILL.md
check_pattern "SKILL.md has skill creation workflow" "$SKILL_DIR/SKILL.md" "[Ss]kill.*[Cc]reat|[Cc]reat.*[Ss]kill"
check_pattern "SKILL.md has domain skill guidance" "$SKILL_DIR/SKILL.md" "[Dd]omain|[Rr]eference"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Skill-creator skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Skill-creator skill validation passed"
    exit 0
fi
