#!/bin/bash
# Run all skill validation scripts for agentops plugins
# Includes both generic dependency validation and skill-specific tests
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PLUGINS_DIR="$REPO_DIR/plugins"
TESTS_DIR="$REPO_DIR/tests/skills"
PASSED=0
FAILED=0
SKIPPED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "╔════════════════════════════════════════════════════════════╗"
echo "║   AgentOps Skill Validation Test Suite                     ║"
echo "╠════════════════════════════════════════════════════════════╣"
echo "║  Tests: Dependency validation + skill-specific validation  ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Infrastructure skills (library-style, auto-loaded via hooks)
INFRA_SKILLS="bd-routing crew dispatch handoff mail molecules polecat-lifecycle roles"

is_infra_skill() {
    local skill="$1"
    for infra in $INFRA_SKILLS; do
        if [ "$skill" = "$infra" ]; then
            return 0
        fi
    done
    return 1
}

# Run validation for each skill in each plugin
for plugin_dir in "$PLUGINS_DIR"/*-kit; do
    plugin_name=$(basename "$plugin_dir")
    echo -e "${BLUE}━━━ Plugin: $plugin_name ━━━${NC}"
    echo ""

    for skill_dir in "$plugin_dir"/skills/*/; do
        [ -d "$skill_dir" ] || continue
        skill_name=$(basename "$skill_dir")
        validate_script="$skill_dir/scripts/validate.sh"

        # Check if it's an infrastructure skill (library)
        if is_infra_skill "$skill_name"; then
            echo -e "  ${BLUE}○ $skill_name (library skill)${NC}"
            SKIPPED=$((SKIPPED + 1))
            continue
        fi

        if [ -f "$validate_script" ]; then
            chmod +x "$validate_script"

            # Run skill-specific validation
            if "$validate_script" > /dev/null 2>&1; then
                echo -e "  ${GREEN}✓ $skill_name${NC}"
                PASSED=$((PASSED + 1))
            else
                echo -e "  ${RED}✗ $skill_name${NC}"
                FAILED=$((FAILED + 1))
            fi
        elif [ -f "$TESTS_DIR/validate-skill.sh" ]; then
            # No skill-specific tests, run generic validation only
            if "$TESTS_DIR/validate-skill.sh" "$skill_name" > /dev/null 2>&1; then
                echo -e "  ${YELLOW}○ $skill_name (generic only)${NC}"
                PASSED=$((PASSED + 1))
            else
                echo -e "  ${RED}✗ $skill_name (generic failed)${NC}"
                FAILED=$((FAILED + 1))
            fi
        else
            echo -e "  ${YELLOW}○ $skill_name (no validation)${NC}"
            SKIPPED=$((SKIPPED + 1))
        fi
    done
    echo ""
done

echo "╔════════════════════════════════════════════════════════════╗"
echo "║                       RESULTS                              ║"
echo "╠════════════════════════════════════════════════════════════╣"
printf "║  ${GREEN}✓${NC} Passed:     %-42s ║\n" "$PASSED skills"
printf "║  ${RED}✗${NC} Failed:     %-42s ║\n" "$FAILED skills"
printf "║  ${BLUE}○${NC} Skipped:    %-42s ║\n" "$SKIPPED (library/no test)"
echo "╠════════════════════════════════════════════════════════════╣"
printf "║  Total Skills: %-40s ║\n" "$((PASSED + FAILED + SKIPPED))"
echo "╚════════════════════════════════════════════════════════════╝"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo -e "${RED}OVERALL: FAIL${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}OVERALL: PASS${NC}"
    exit 0
fi
