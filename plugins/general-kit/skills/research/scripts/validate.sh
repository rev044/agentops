#!/bin/bash
# Validate research skill
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

echo "=== Research Skill Validation ==="
echo ""

# Verify beads CLI exists (primary dependency)
check "bd binary exists" "which bd" "bd"

# Verify research output directories pattern
check_exists "Agent artifacts base dir" "$HOME/gt/.agents"

# Verify skill references exist
check_exists "Context discovery reference" "$SKILL_DIR/references/context-discovery.md"
check_exists "Document template reference" "$SKILL_DIR/references/document-template.md"
check_exists "Failure patterns reference" "$SKILL_DIR/references/failure-patterns.md"
check_exists "Vibe methodology reference" "$SKILL_DIR/references/vibe-methodology.md"

# Verify research workflow patterns in SKILL.md
check_pattern "SKILL.md has 6-tier context discovery" "$SKILL_DIR/SKILL.md" "6-[Tt]ier"
check_pattern "SKILL.md has Prior Art section" "$SKILL_DIR/SKILL.md" "Prior Art"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Research skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Research skill validation passed"
    exit 0
fi
