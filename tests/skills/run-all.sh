#!/bin/bash
# Run all skill validation scripts for agentops
# Updated for unified structure (skills/ at repo root)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SKILLS_DIR="$REPO_ROOT/skills"
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
echo "║  Tests: SKILL.md + frontmatter + skill-specific validate   ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Infrastructure/library skills (auto-loaded, not user-invoked)
INFRA_SKILLS=""

is_infra_skill() {
    local skill="$1"
    for infra in $INFRA_SKILLS; do
        if [ "$skill" = "$infra" ]; then
            return 0
        fi
    done
    return 1
}

echo -e "${BLUE}━━━ Skills Directory: $SKILLS_DIR ━━━${NC}"
echo ""

# Validate each skill in skills/
for skill_dir in "$SKILLS_DIR"/*/; do
    [ -d "$skill_dir" ] || continue
    skill_name=$(basename "$skill_dir")
    validate_script="$skill_dir/scripts/validate.sh"

    # Check if it's an infrastructure skill (library)
    if is_infra_skill "$skill_name"; then
        echo -e "  ${BLUE}○ $skill_name (library skill)${NC}"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Check SKILL.md exists with frontmatter
    skill_md="$skill_dir/SKILL.md"
    if [[ ! -f "$skill_md" ]]; then
        echo -e "  ${RED}✗ $skill_name (missing SKILL.md)${NC}"
        FAILED=$((FAILED + 1))
        continue
    fi

    if ! head -1 "$skill_md" | grep -q "^---$"; then
        echo -e "  ${RED}✗ $skill_name (no YAML frontmatter)${NC}"
        FAILED=$((FAILED + 1))
        continue
    fi

    if ! grep -q "^name:" "$skill_md"; then
        echo -e "  ${RED}✗ $skill_name (missing 'name' in frontmatter)${NC}"
        FAILED=$((FAILED + 1))
        continue
    fi

    # Run skill-specific validation if present
    if [ -f "$validate_script" ]; then
        chmod +x "$validate_script"
        if "$validate_script" > /dev/null 2>&1; then
            echo -e "  ${GREEN}✓ $skill_name${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "  ${RED}✗ $skill_name (validate.sh failed)${NC}"
            FAILED=$((FAILED + 1))
        fi
    else
        # No skill-specific tests, basic validation passed
        echo -e "  ${GREEN}✓ $skill_name${NC} ${YELLOW}(no validate.sh)${NC}"
        PASSED=$((PASSED + 1))
    fi
done

echo ""
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
