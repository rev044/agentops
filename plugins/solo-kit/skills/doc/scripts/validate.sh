#!/bin/bash
# Validate doc skill
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

echo "=== Doc Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify skill scripts exist (doc skill uses detection scripts)
check_exists "Project detection script" "$SKILL_DIR/scripts/detect-project.sh"

# Verify skill references exist
check_exists "References directory" "$SKILL_DIR/references"
if [ -d "$SKILL_DIR/references" ]; then
    check_exists "Project types reference" "$SKILL_DIR/references/project-types.md"
fi

# Verify doc workflow patterns in SKILL.md
check_pattern "SKILL.md has project detection" "$SKILL_DIR/SKILL.md" "[Pp]roject.*[Dd]etect|[Dd]etect.*[Pp]roject"
check_pattern "SKILL.md has coverage command" "$SKILL_DIR/SKILL.md" "coverage"
check_pattern "SKILL.md has discover command" "$SKILL_DIR/SKILL.md" "discover"
check_pattern "SKILL.md documents project types" "$SKILL_DIR/SKILL.md" "CODING|INFORMATIONAL|OPS"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: Doc skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: Doc skill validation passed"
    exit 0
fi
