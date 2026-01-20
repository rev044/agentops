#!/bin/bash
# Validate validation-chain skill
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

echo "=== Validation Chain Skill Validation ==="
echo ""


# Verify dependent skills exist
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"
check_exists "Vibe skill exists" "$HOME/.claude/skills/vibe/SKILL.md"
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify validation-chain workflow patterns in SKILL.md
check_pattern "SKILL.md has file classification" "$SKILL_DIR/SKILL.md" "[Cc]lassification|[Cc]lassify"
check_pattern "SKILL.md has specialist routing" "$SKILL_DIR/SKILL.md" "[Ss]pecialist"
check_pattern "SKILL.md has parallel dispatch" "$SKILL_DIR/SKILL.md" "[Pp]arallel|Task"
check_pattern "SKILL.md has triage matrix" "$SKILL_DIR/SKILL.md" "[Tt]riage|[Bb]locker"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Validation-chain skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Validation-chain skill validation passed"
    exit 0
fi
