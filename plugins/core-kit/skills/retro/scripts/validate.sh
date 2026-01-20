#!/bin/bash
# Validate retro skill
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

echo "=== Retro Skill Validation ==="
echo ""

# Verify beads CLI exists
check "bd binary exists" "which bd" "bd"

# Verify dependent skill exists
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"

# Verify skill references exist
check_exists "References directory exists" "$SKILL_DIR/references"

# Verify retro workflow patterns in SKILL.md
check_pattern "SKILL.md has learnings extraction" "$SKILL_DIR/SKILL.md" "[Ll]earning"
check_pattern "SKILL.md mentions artifact output" "$SKILL_DIR/SKILL.md" "retros|artifact"

# Verify retros output location pattern
check_exists "Agent artifacts base dir" "$HOME/gt/.agents"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Retro skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Retro skill validation passed"
    exit 0
fi
