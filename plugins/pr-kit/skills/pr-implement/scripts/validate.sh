#!/bin/bash
# Validate pr-implement skill
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

echo "=== PR Implement Skill Validation ==="
echo ""


# Verify dependent skills exist
check_exists "PR-plan skill exists" "$HOME/.claude/skills/pr-plan/SKILL.md"
check_exists "PR-prep skill exists" "$HOME/.claude/skills/pr-prep/SKILL.md"
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify pr-implement workflow patterns in SKILL.md
check_pattern "SKILL.md has fork-based implementation" "$SKILL_DIR/SKILL.md" "[Ff]ork"
check_pattern "SKILL.md has isolation check" "$SKILL_DIR/SKILL.md" "[Ii]solation.*[Cc]heck"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: PR-implement skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: PR-implement skill validation passed"
    exit 0
fi
