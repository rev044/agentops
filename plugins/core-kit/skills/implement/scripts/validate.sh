#!/bin/bash
# Validate implement skill
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

echo "=== Implement Skill Validation ==="
echo ""

# Verify beads CLI exists and has required commands
check "bd binary exists" "which bd" "bd"
check "bd ready command" "bd ready --help" "ready"
check "bd update command" "bd update --help" "update"
check "bd close command" "bd close --help" "close"
check "bd comment command" "bd comment --help" "comment"
check "bd sync command" "bd sync --help" "sync"

# Verify just build tool (used for testing)
check "just binary exists" "which just" "just"

# Verify dependent skills exist
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify implementation workflow patterns in SKILL.md
check_pattern "SKILL.md has workflow phases" "$SKILL_DIR/SKILL.md" "Phase [0-9]"
check_pattern "SKILL.md mentions testing" "$SKILL_DIR/SKILL.md" "just test|MANDATORY"
check_pattern "SKILL.md has context/patterns reference" "$SKILL_DIR/SKILL.md" "[Cc]ontext|[Pp]attern|[Ll]int"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Implement skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Implement skill validation passed"
    exit 0
fi
