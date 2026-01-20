#!/bin/bash
# Validate tekton skill
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC2034  # Reserved for future use
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0
WARNINGS=0

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

echo "=== Tekton Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Check tekton CLI (optional - warn if missing)
check_optional "tkn CLI installed" "which tkn" "tkn"

# Verify tekton workflow patterns in SKILL.md
check_pattern "SKILL.md has pipeline documentation" "$SKILL_DIR/SKILL.md" "[Pp]ipeline"
check_pattern "SKILL.md has pipelinerun documentation" "$SKILL_DIR/SKILL.md" "[Pp]ipelinerun|[Pp]ipeline[Rr]un"
check_pattern "SKILL.md has build triggers" "$SKILL_DIR/SKILL.md" "[Bb]uild|[Ii]mage"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Tekton skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Tekton skill validation passed"
    exit 0
fi
