#!/bin/bash
# Validate tekton-go-operator skill
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

echo "=== Tekton-Go-Operator Skill Validation ==="
echo ""


# Verify dependent skills exist
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"
check_exists "Tekton skill exists" "$HOME/.claude/skills/tekton/SKILL.md"
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Check optional tools
check_optional "tkn CLI installed" "which tkn" "tkn"

# Verify tekton-go-operator workflow patterns in SKILL.md
check_pattern "SKILL.md has Go operator documentation" "$SKILL_DIR/SKILL.md" "Go|operator|kubebuilder"
check_pattern "SKILL.md has Tekton CI documentation" "$SKILL_DIR/SKILL.md" "Tekton|CI"
check_pattern "SKILL.md has friction points documented" "$SKILL_DIR/SKILL.md" "[Ff]riction|dockerignore|envtest"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Tekton-go-operator skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Tekton-go-operator skill validation passed"
    exit 0
fi
