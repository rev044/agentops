#!/bin/bash
# Validate complexity skill
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2034  # Reserved for future use
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0
WARNINGS=0

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

check_optional() {
    local desc="$1"
    local cmd="$2"
    local expected="$3"

    CHECKS=$((CHECKS + 1))
    if eval "$cmd" 2>/dev/null | grep -qi "$expected"; then
        echo "✓ $desc"
    else
        echo "⚠ $desc (optional, not installed)"
        WARNINGS=$((WARNINGS + 1))
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

echo "=== Complexity Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Check complexity analysis tools (optional - warn if missing)
check_optional "radon installed (Python complexity)" "which radon" "radon"
check_optional "gocyclo installed (Go complexity)" "which gocyclo" "gocyclo"

# Verify complexity workflow patterns in SKILL.md
check_pattern "SKILL.md documents radon" "$SKILL_DIR/SKILL.md" "radon"
check_pattern "SKILL.md documents gocyclo" "$SKILL_DIR/SKILL.md" "gocyclo"
check_pattern "SKILL.md has complexity grades" "$SKILL_DIR/SKILL.md" "[Cc]omplexity.*[Gg]rade|A-F"
check_pattern "SKILL.md has refactoring guidance" "$SKILL_DIR/SKILL.md" "[Rr]efactor"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Complexity skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Complexity skill validation passed"
    exit 0
fi
