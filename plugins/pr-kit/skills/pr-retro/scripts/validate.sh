#!/bin/bash
# Validate pr-retro skill
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

echo "=== PR Retro Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"

# Verify pr-retro workflow patterns in SKILL.md
check_pattern "SKILL.md has PR outcome analysis" "$SKILL_DIR/SKILL.md" "[Oo]utcome|[Aa]nalyze"
check_pattern "SKILL.md has lessons learned" "$SKILL_DIR/SKILL.md" "[Ll]esson|[Ll]earn"
check_pattern "SKILL.md mentions accept/reject" "$SKILL_DIR/SKILL.md" "[Aa]ccept|[Rr]eject"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: PR-retro skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: PR-retro skill validation passed"
    exit 0
fi
