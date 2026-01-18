#!/bin/bash
# Run all skill validation scripts
set -euo pipefail

SKILLS_DIR="${HOME}/.claude/skills"
PASSED=0
FAILED=0
SKIPPED=0

echo "╔════════════════════════════════════════════╗"
echo "║     Skill Validation Test Suite            ║"
echo "╚════════════════════════════════════════════╝"
echo ""

# Find all validate.sh scripts
for skill_dir in "$SKILLS_DIR"/*/; do
    skill_name=$(basename "$skill_dir")
    validate_script="$skill_dir/scripts/validate.sh"

    if [ -f "$validate_script" ]; then
        echo "━━━ Testing: $skill_name ━━━"
        chmod +x "$validate_script"

        if "$validate_script"; then
            PASSED=$((PASSED + 1))
        else
            FAILED=$((FAILED + 1))
        fi
        echo ""
    else
        SKIPPED=$((SKIPPED + 1))
    fi
done

echo "╔════════════════════════════════════════════╗"
echo "║               RESULTS                      ║"
echo "╠════════════════════════════════════════════╣"
printf "║  ✓ Passed:  %-28s ║\n" "$PASSED"
printf "║  ✗ Failed:  %-28s ║\n" "$FAILED"
printf "║  ○ Skipped: %-28s ║\n" "$SKIPPED (no validate.sh)"
echo "╚════════════════════════════════════════════╝"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "OVERALL: FAIL"
    exit 1
else
    echo ""
    echo "OVERALL: PASS"
    exit 0
fi
