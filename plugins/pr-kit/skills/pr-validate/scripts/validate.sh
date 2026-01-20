#!/bin/bash
# Validate pr-validate skill
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

echo "=== PR Validate Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "PR-prep skill exists" "$HOME/.claude/skills/pr-prep/SKILL.md"

# Verify pr-validate workflow patterns in SKILL.md
check_pattern "SKILL.md has isolation validation" "$SKILL_DIR/SKILL.md" "[Ii]solation"
check_pattern "SKILL.md has scope creep detection" "$SKILL_DIR/SKILL.md" "[Ss]cope.*[Cc]reep"
check_pattern "SKILL.md has upstream alignment" "$SKILL_DIR/SKILL.md" "[Uu]pstream.*[Aa]lign"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: PR-validate skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: PR-validate skill validation passed"
    exit 0
fi
