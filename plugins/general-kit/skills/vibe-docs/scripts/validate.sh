#!/bin/bash
# Validate vibe-docs skill
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

echo "=== Vibe-Docs Skill Validation ==="
echo ""


# Verify vibe-docs workflow patterns in SKILL.md
check_pattern "SKILL.md has documentation validation" "$SKILL_DIR/SKILL.md" "[Dd]oc.*[Vv]alid|[Vv]alid.*[Dd]oc"
check_pattern "SKILL.md has deployment reality check" "$SKILL_DIR/SKILL.md" "[Dd]eploy|[Rr]eality"
check_pattern "SKILL.md has claim verification" "$SKILL_DIR/SKILL.md" "[Cc]laim|[Vv]erify|[Aa]udit"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Vibe-docs skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Vibe-docs skill validation passed"
    exit 0
fi
