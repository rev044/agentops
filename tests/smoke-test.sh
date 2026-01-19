#!/bin/bash
# Smoke test for AgentOps marketplace
# Usage: ./tests/smoke-test.sh [--verbose]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERBOSE="${1:-}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

errors=0
warnings=0

log() { echo -e "${BLUE}[TEST]${NC} $1"; }
pass() { echo -e "${GREEN}  ✓${NC} $1"; }
fail() { echo -e "${RED}  ✗${NC} $1"; ((errors++)) || true; }
warn() { echo -e "${YELLOW}  ⚠${NC} $1"; ((warnings++)) || true; }

cd "$REPO_ROOT"

# =============================================================================
# Test 1: Validate JSON files
# =============================================================================
log "Validating JSON files..."

for jf in .claude-plugin/marketplace.json .claude-plugin/plugin.json plugins/*/.claude-plugin/plugin.json; do
    [[ ! -f "$jf" ]] && continue
    if python3 -m json.tool "$jf" > /dev/null 2>&1; then
        pass "$jf"
    else
        fail "$jf - invalid JSON"
    fi
done

# =============================================================================
# Test 2: Validate marketplace references
# =============================================================================
log "Validating marketplace references..."

# Extract plugin names and sources from marketplace.json
while IFS= read -r line; do
    name=$(echo "$line" | cut -d'|' -f1)
    source=$(echo "$line" | cut -d'|' -f2)

    [[ -z "$name" ]] && continue

    # Check source exists
    if [[ "$source" == "." ]]; then
        pj=".claude-plugin/plugin.json"
    else
        pj="${source}/.claude-plugin/plugin.json"
    fi

    if [[ -f "$pj" ]]; then
        pass "$name -> $source"
    else
        fail "$name: $pj not found"
    fi
done < <(python3 -c "
import json
with open('.claude-plugin/marketplace.json') as f:
    mp = json.load(f)
for p in mp.get('plugins', []):
    src = p['source'].lstrip('./')
    if not src: src = '.'
    print(f\"{p['name']}|{src}\")
")

# =============================================================================
# Test 3: Validate skill structure
# =============================================================================
log "Validating skill structure..."

skill_count=0
skill_errors=0

for plugin_dir in plugins/*/; do
    skills_dir="${plugin_dir}skills"
    [[ ! -d "$skills_dir" ]] && continue

    for skill_dir in "$skills_dir"/*/; do
        [[ ! -d "$skill_dir" ]] && continue
        if [[ -f "${skill_dir}SKILL.md" ]]; then
            skill_count=$((skill_count + 1))
        else
            fail "$(basename "$plugin_dir")/skills/$(basename "$skill_dir")/SKILL.md missing"
            skill_errors=$((skill_errors + 1))
        fi
    done
done

if [[ $skill_errors -eq 0 ]]; then
    pass "All $skill_count skills have SKILL.md"
fi

# =============================================================================
# Test 4: Validate commands structure
# =============================================================================
log "Validating commands structure..."

cmd_count=0
[[ -d "commands" ]] && cmd_count=$(find commands -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
for plugin_dir in plugins/*/; do
    [[ -d "${plugin_dir}commands" ]] && cmd_count=$((cmd_count + $(find "${plugin_dir}commands" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')))
done

pass "Found $cmd_count commands"

# =============================================================================
# Test 5: Check for placeholder patterns
# =============================================================================
log "Checking for issues..."

# Check for [your-email] placeholders
placeholder_count=$(grep -rc "\[your-email\]" --include="*.md" . 2>/dev/null | grep -v ":0$" | wc -l || true)
if [[ "$placeholder_count" -gt 0 ]]; then
    fail "$placeholder_count [your-email] placeholders found"
    [[ "$VERBOSE" == "--verbose" ]] && grep -rn "\[your-email\]" --include="*.md" . 2>/dev/null | head -3
else
    pass "No placeholder emails"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${BLUE}═══════════════════════════════════════════${NC}"

if [[ $errors -gt 0 ]]; then
    echo -e "${RED}FAILED${NC} - $errors errors, $warnings warnings"
    exit 1
elif [[ $warnings -gt 0 ]]; then
    echo -e "${YELLOW}PASSED WITH WARNINGS${NC} - $warnings warnings"
    exit 0
else
    echo -e "${GREEN}PASSED${NC} - All smoke tests passed"
    exit 0
fi
