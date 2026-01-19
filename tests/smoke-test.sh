#!/bin/bash
# Smoke test for AgentOps marketplace
# Usage: ./tests/smoke-test.sh [--verbose]
#
# Validates:
# - All JSON files are valid
# - All plugins have required structure
# - All skills have SKILL.md files
# - marketplace.json references exist
# - Plugin metadata is consistent

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERBOSE="${1:-}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ERRORS=()
WARNINGS=()

log() { echo -e "${BLUE}[TEST]${NC} $1"; }
pass() { echo -e "${GREEN}  ✓${NC} $1"; }
fail() { echo -e "${RED}  ✗${NC} $1"; ERRORS+=("$1"); }
warn() { echo -e "${YELLOW}  ⚠${NC} $1"; WARNINGS+=("$1"); }
verbose() { [[ "$VERBOSE" == "--verbose" ]] && echo -e "    $1" || true; }

cd "$REPO_ROOT"

# =============================================================================
# Test 1: Validate JSON files
# =============================================================================
log "Validating JSON files..."

validate_json() {
    local file="$1"
    if python3 -m json.tool "$file" > /dev/null 2>&1; then
        pass "$file"
        return 0
    else
        fail "$file - invalid JSON"
        return 1
    fi
}

# Root marketplace and plugin
validate_json ".claude-plugin/marketplace.json"
validate_json ".claude-plugin/plugin.json"

# All plugin.json files
for pj in plugins/*/.claude-plugin/plugin.json; do
    [[ -f "$pj" ]] && validate_json "$pj"
done

# =============================================================================
# Test 2: Validate marketplace references
# =============================================================================
log "Validating marketplace references..."

python3 << 'EOF'
import json
import os
import sys

with open('.claude-plugin/marketplace.json') as f:
    mp = json.load(f)

errors = []
for p in mp.get('plugins', []):
    name = p['name']
    source = p['source'].lstrip('./')
    if source == '':
        source = '.'

    # Check source directory exists
    if not os.path.isdir(source):
        errors.append(f"{name}: source directory '{source}' not found")
        continue

    # Check plugin.json exists
    if source == '.':
        pj_path = '.claude-plugin/plugin.json'
    else:
        pj_path = f'{source}/.claude-plugin/plugin.json'

    if not os.path.isfile(pj_path):
        errors.append(f"{name}: missing {pj_path}")
        continue

    # Verify name matches
    with open(pj_path) as f:
        pj = json.load(f)
    if pj.get('name') != name:
        errors.append(f"{name}: plugin.json name '{pj.get('name')}' doesn't match marketplace")

    print(f"  \033[0;32m✓\033[0m {name} -> {source}")

if errors:
    print()
    for e in errors:
        print(f"  \033[0;31m✗\033[0m {e}")
    sys.exit(1)
EOF

# =============================================================================
# Test 3: Validate skill structure
# =============================================================================
log "Validating skill structure..."

skill_count=0
missing_skill_md=()

for plugin_dir in plugins/*/; do
    plugin_name=$(basename "$plugin_dir")
    skills_dir="${plugin_dir}skills"

    if [[ ! -d "$skills_dir" ]]; then
        verbose "$plugin_name: no skills directory"
        continue
    fi

    for skill_dir in "$skills_dir"/*/; do
        [[ ! -d "$skill_dir" ]] && continue
        skill_name=$(basename "$skill_dir")
        skill_md="${skill_dir}SKILL.md"

        if [[ -f "$skill_md" ]]; then
            ((skill_count++))
            verbose "$plugin_name/$skill_name: OK"
        else
            missing_skill_md+=("$plugin_name/skills/$skill_name")
        fi
    done
done

if [[ ${#missing_skill_md[@]} -gt 0 ]]; then
    for s in "${missing_skill_md[@]}"; do
        fail "$s/SKILL.md missing"
    done
else
    pass "All $skill_count skills have SKILL.md"
fi

# =============================================================================
# Test 4: Validate commands structure
# =============================================================================
log "Validating commands structure..."

cmd_count=0

# Root commands
if [[ -d "commands" ]]; then
    for cmd in commands/*.md; do
        [[ -f "$cmd" ]] || continue
        ((cmd_count++))
        # Check for required frontmatter or header
        if head -5 "$cmd" | grep -q "^#\|^---"; then
            verbose "$(basename "$cmd"): OK"
        else
            warn "$(basename "$cmd"): missing header or frontmatter"
        fi
    done
fi

# Plugin commands
for plugin_dir in plugins/*/; do
    cmd_dir="${plugin_dir}commands"
    [[ ! -d "$cmd_dir" ]] && continue

    for cmd in "$cmd_dir"/*.md; do
        [[ -f "$cmd" ]] || continue
        ((cmd_count++))
    done
done

pass "Found $cmd_count commands"

# =============================================================================
# Test 5: Check for common issues
# =============================================================================
log "Checking for common issues..."

# Check for TODO markers in production code (exclude test files and docs)
todo_count=$(grep -r "TODO\|FIXME\|XXX" --include="*.sh" --include="*.py" \
    --exclude-dir=tests \
    --exclude-dir=.git \
    . 2>/dev/null | wc -l | tr -d ' ')

if [[ "$todo_count" -gt 0 ]]; then
    warn "$todo_count TODO/FIXME markers found in scripts"
    [[ "$VERBOSE" == "--verbose" ]] && grep -r "TODO\|FIXME\|XXX" --include="*.sh" --include="*.py" --exclude-dir=tests --exclude-dir=.git . 2>/dev/null | head -5
else
    pass "No TODO markers in scripts"
fi

# Check for placeholder emails
placeholder_count=$(grep -r "\[your-email\]\|example\.com" --include="*.md" \
    --exclude=CONTRIBUTING.md \
    . 2>/dev/null | wc -l | tr -d ' ')

if [[ "$placeholder_count" -gt 0 ]]; then
    warn "$placeholder_count placeholder emails found"
else
    pass "No placeholder emails"
fi

# =============================================================================
# Test 6: Validate plugin metadata consistency
# =============================================================================
log "Validating plugin metadata..."

python3 << 'EOF'
import json
import os

plugins_dir = 'plugins'
issues = []

for plugin_name in os.listdir(plugins_dir):
    plugin_path = os.path.join(plugins_dir, plugin_name)
    if not os.path.isdir(plugin_path):
        continue

    pj_path = os.path.join(plugin_path, '.claude-plugin', 'plugin.json')
    if not os.path.isfile(pj_path):
        continue

    with open(pj_path) as f:
        pj = json.load(f)

    # Check required fields
    required = ['name', 'version', 'description']
    for field in required:
        if field not in pj:
            issues.append(f"{plugin_name}: missing '{field}' in plugin.json")

    # Check version format
    version = pj.get('version', '')
    if version and not all(p.isdigit() for p in version.split('.')):
        issues.append(f"{plugin_name}: invalid version format '{version}'")

    # Check skills reference if present
    skills_ref = pj.get('skills')
    if skills_ref:
        skills_path = os.path.join(plugin_path, skills_ref.lstrip('./'))
        if not os.path.isdir(skills_path):
            issues.append(f"{plugin_name}: skills path '{skills_ref}' not found")

if issues:
    for i in issues:
        print(f"  \033[0;31m✗\033[0m {i}")
    exit(1)
else:
    print(f"  \033[0;32m✓\033[0m All plugin metadata valid")
EOF

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${BLUE}═══════════════════════════════════════════${NC}"

if [[ ${#ERRORS[@]} -gt 0 ]]; then
    echo -e "${RED}FAILED${NC} - ${#ERRORS[@]} errors, ${#WARNINGS[@]} warnings"
    echo ""
    echo "Errors:"
    for e in "${ERRORS[@]}"; do
        echo "  - $e"
    done
    exit 1
elif [[ ${#WARNINGS[@]} -gt 0 ]]; then
    echo -e "${YELLOW}PASSED WITH WARNINGS${NC} - ${#WARNINGS[@]} warnings"
    exit 0
else
    echo -e "${GREEN}PASSED${NC} - All smoke tests passed"
    exit 0
fi
