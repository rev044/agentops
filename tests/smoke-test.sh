#!/bin/bash
# Smoke test for AgentOps plugin
# Usage: ./tests/smoke-test.sh [--verbose]
# Updated for unified structure (skills/ at repo root)

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

for jf in .claude-plugin/plugin.json hooks/hooks.json; do
    if [[ ! -f "$jf" ]]; then
        fail "$jf - not found"
        continue
    fi
    if python3 -m json.tool "$jf" > /dev/null 2>&1; then
        pass "$jf"
    else
        fail "$jf - invalid JSON"
    fi
done

# =============================================================================
# Test 2: Validate plugin manifest
# =============================================================================
log "Validating plugin manifest..."

manifest=".claude-plugin/plugin.json"
valid_keys='["$schema","name","version","description","author","homepage","repository","license","keywords","commands","skills","agents","hooks"]'

if jq empty "$manifest" 2>/dev/null; then
    invalid=$(jq -r --argjson valid "$valid_keys" 'keys - $valid | .[]' "$manifest" 2>/dev/null || true)
    if [[ -n "$invalid" ]]; then
        fail "Invalid manifest keys: $invalid"
    else
        pass "Manifest keys valid"
    fi
else
    fail "Manifest not valid JSON"
fi

# =============================================================================
# Test 3: Validate skill structure (unified - skills/ at root)
# =============================================================================
log "Validating skill structure..."

skill_count=0
skill_errors=0

for skill_dir in skills/*/; do
    [[ ! -d "$skill_dir" ]] && continue
    skill_name=$(basename "$skill_dir")
    skill_md="${skill_dir}SKILL.md"

    if [[ -f "$skill_md" ]]; then
        # Check for frontmatter
        if head -1 "$skill_md" | grep -q "^---$"; then
            if grep -q "^name:" "$skill_md"; then
                skill_count=$((skill_count + 1))
            else
                fail "$skill_name - missing 'name' in frontmatter"
                skill_errors=$((skill_errors + 1))
            fi
        else
            fail "$skill_name - no YAML frontmatter"
            skill_errors=$((skill_errors + 1))
        fi
    else
        fail "$skill_name/SKILL.md missing"
        skill_errors=$((skill_errors + 1))
    fi
done

if [[ $skill_errors -eq 0 ]] && [[ $skill_count -gt 0 ]]; then
    pass "All $skill_count skills have valid SKILL.md"
fi

# =============================================================================
# Test 4: Validate agents structure
# =============================================================================
log "Validating agents structure..."

agent_count=0
[[ -d "agents" ]] && agent_count=$(find agents -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')

if [[ $agent_count -gt 0 ]]; then
    pass "Found $agent_count agents"
else
    warn "No agents found (optional)"
fi

# =============================================================================
# Test 5: Validate hooks
# =============================================================================
log "Validating hooks..."

if [[ -f "hooks/hooks.json" ]]; then
    if jq empty "hooks/hooks.json" 2>/dev/null; then
        pass "hooks.json valid"
    else
        fail "hooks.json invalid JSON"
    fi
else
    warn "No hooks.json found"
fi

# =============================================================================
# Test 6: Validate CLI builds (if Go available)
# =============================================================================
log "Validating ao CLI..."

if [[ -d "cli" ]]; then
    if command -v go &>/dev/null; then
        tmpdir="$(mktemp -d -t ao-test.XXXXXX)"
        trap 'rm -rf "$tmpdir"' EXIT
        tmpbin="$tmpdir/ao"
        if (cd "$REPO_ROOT/cli" && go build -o "$tmpbin" ./cmd/ao 2>/dev/null); then
            pass "ao CLI builds successfully"
        else
            fail "ao CLI build failed"
        fi
    else
        warn "Go not available - skipping CLI build test"
    fi
else
    warn "No cli/ directory"
fi

# =============================================================================
# Test 7: Check for placeholder patterns
# =============================================================================
log "Checking for issues..."

# Check for [your-email] placeholders
placeholder_files=$(grep -rl "\[your-email\]" --include="*.md" . 2>/dev/null | wc -l | tr -d ' ') || true
if [[ "${placeholder_files:-0}" -gt 0 ]]; then
    fail "$placeholder_files files with [your-email] placeholders"
    [[ "$VERBOSE" == "--verbose" ]] && grep -rn "\[your-email\]" --include="*.md" . 2>/dev/null | head -3 || true
else
    pass "No placeholder emails"
fi

# Check for TODO/FIXME in skills
todo_files=$(grep -l "TODO\|FIXME" skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ') || true
if [[ "${todo_files:-0}" -gt 0 ]]; then
    warn "$todo_files skills with TODO/FIXME"
else
    pass "No TODO/FIXME in skills"
fi

# =============================================================================
# Test 8: Claude CLI load test
# =============================================================================
log "Testing Claude CLI plugin load..."

if command -v claude &>/dev/null; then
    load_output=$(timeout 10 claude --plugin-dir . --help 2>&1) || true
    if echo "$load_output" | grep -qiE "invalid manifest|validation error|failed to load"; then
        fail "Claude CLI load failed"
        echo "$load_output" | grep -iE "invalid|failed|error" | head -3 | sed 's/^/    /'
    else
        pass "Claude CLI loads plugin"
    fi
else
    warn "Claude CLI not available for load test"
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
