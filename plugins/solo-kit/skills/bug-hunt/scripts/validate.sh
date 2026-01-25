#!/bin/bash
# Validate bug-hunt skill
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

echo "=== Bug Hunt Skill Validation ==="
echo ""


# Verify git is available (required for archaeology)
check "git binary exists" "which git" "git"

# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify bug-hunt workflow patterns in SKILL.md
check_pattern "SKILL.md has git archaeology" "$SKILL_DIR/SKILL.md" "git blame|git log|git bisect|[Aa]rchaeology"
check_pattern "SKILL.md has root cause section" "$SKILL_DIR/SKILL.md" "[Rr]oot [Cc]ause"
check_pattern "SKILL.md has reproduce phase" "$SKILL_DIR/SKILL.md" "[Rr]eproduce"
check_pattern "SKILL.md has fix design phase" "$SKILL_DIR/SKILL.md" "[Ff]ix [Dd]esign|[Ff]ix.*[Pp]lan"
check_pattern "SKILL.md has rig detection" "$SKILL_DIR/SKILL.md" "[Rr]ig [Dd]etection"

# Verify output location pattern
check_exists "Agent artifacts base dir" "$HOME/gt/.agents"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Bug-hunt skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Bug-hunt skill validation passed"
    exit 0
fi
