#!/bin/bash
# Validate crank skill
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

echo "=== Crank Skill Validation ==="
echo ""

# Verify beads CLI exists and has required commands
check "bd binary exists" "which bd" "bd"
check "bd list command" "bd list --help" "list"
check "bd show command" "bd show --help" "show"

# Verify gastown CLI exists and has required commands
check "gt binary exists" "which gt" "gt"
check "gt sling command" "gt sling --help" "sling"
check "gt convoy command" "gt convoy --help" "convoy"

# Verify dependent skills exist
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"
check_exists "Gastown skill exists" "$HOME/.claude/skills/gastown/SKILL.md"
check_exists "Implement skill exists" "$HOME/.claude/skills/implement/SKILL.md"

# Verify crank workflow patterns in SKILL.md
check_pattern "SKILL.md has ODMCR loop" "$SKILL_DIR/SKILL.md" "ODMCR"
check_pattern "SKILL.md has role detection" "$SKILL_DIR/SKILL.md" "Role Detection|ROLE="
check_pattern "SKILL.md documents Mayor mode" "$SKILL_DIR/SKILL.md" "Mayor"
check_pattern "SKILL.md documents Crew mode" "$SKILL_DIR/SKILL.md" "Crew"

# Verify supporting documentation exists
check_exists "ODMCR reference" "$SKILL_DIR/odmcr.md"
check_exists "Failure taxonomy" "$SKILL_DIR/failure-taxonomy.md"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Crank skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Crank skill validation passed"
    exit 0
fi
