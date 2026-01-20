#!/bin/bash
# Validate golden-init skill
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

echo "=== Golden-Init Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify golden-init workflow patterns in SKILL.md
check_pattern "SKILL.md documents .agents/ directory" "$SKILL_DIR/SKILL.md" "\\.agents/"
check_pattern "SKILL.md documents .beads/ directory" "$SKILL_DIR/SKILL.md" "\\.beads/"
check_pattern "SKILL.md documents justfile" "$SKILL_DIR/SKILL.md" "justfile"
check_pattern "SKILL.md documents pre-commit" "$SKILL_DIR/SKILL.md" "pre-commit"
check_pattern "SKILL.md has auto-detection" "$SKILL_DIR/SKILL.md" "[Aa]uto.*[Dd]etect|[Dd]etect"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Golden-init skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Golden-init skill validation passed"
    exit 0
fi
