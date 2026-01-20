#!/bin/bash
# Validate vibe skill
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

echo "=== Vibe Skill Validation ==="
echo ""


# Verify dependent skills exist
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify skill references exist (if any)
if [ -d "$SKILL_DIR/references" ]; then
    check_exists "References directory exists" "$SKILL_DIR/references"
fi

# Verify vibe workflow patterns in SKILL.md
check_pattern "SKILL.md has Talos designation" "$SKILL_DIR/SKILL.md" "Talos"
check_pattern "SKILL.md has aspects documentation" "$SKILL_DIR/SKILL.md" "Aspects|aspects"
check_pattern "SKILL.md has security validation" "$SKILL_DIR/SKILL.md" "[Ss]ecurity"
check_pattern "SKILL.md has quality validation" "$SKILL_DIR/SKILL.md" "[Qq]uality"

# Verify expert agents exist (vibe uses these for deep validation)
check_exists "Security expert agent" "$HOME/.claude/agents/security-expert.md"
check_exists "Code quality expert agent" "$HOME/.claude/agents/code-quality-expert.md"
check_exists "Architecture expert agent" "$HOME/.claude/agents/architecture-expert.md"
check_exists "UX expert agent" "$HOME/.claude/agents/ux-expert.md"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Vibe skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Vibe skill validation passed"
    exit 0
fi
