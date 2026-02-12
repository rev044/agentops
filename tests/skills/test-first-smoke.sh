#!/usr/bin/env bash
# Smoke test for --test-first flow structural patterns
# Validates that all skill files contain required structural elements
# for the spec-first TDD pipeline (SPEC WAVE → TEST WAVE → GREEN mode).
#
# Usage: ./tests/skills/test-first-smoke.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

total=0
passed=0
failed=0

pass() { echo -e "${GREEN}  PASS${NC} $1"; ((total++)) || true; ((passed++)) || true; }
fail() { echo -e "${RED}  FAIL${NC} $1"; ((total++)) || true; ((failed++)) || true; }

cd "$REPO_ROOT"

# =============================================================================
# 1. Contract Template checks
# =============================================================================
echo -e "${BLUE}[TEST-FIRST]${NC} Contract template (skills/crank/references/contract-template.md)"

CONTRACT="skills/crank/references/contract-template.md"

if [[ -f "$CONTRACT" ]]; then
    pass "File exists"
else
    fail "File missing: $CONTRACT"
    # All subsequent contract checks will fail; skip to next section
    echo -e "${BLUE}[TEST-FIRST]${NC} Skipping remaining contract checks (file missing)"
fi

if [[ -f "$CONTRACT" ]]; then
    # 1a: All 7 required ## headings
    for heading in "Problem" "Inputs" "Outputs" "Invariants" "Failure Modes" "Out of Scope" "Test Cases"; do
        if grep -qE "^## ${heading}$" "$CONTRACT"; then
            pass "Contract has ## $heading heading"
        else
            fail "Contract missing ## $heading heading"
        fi
    done

    # 1b: Contract Granularity section
    if grep -qE "^## Contract Granularity$" "$CONTRACT"; then
        pass "Contract has ## Contract Granularity section"
    else
        fail "Contract missing ## Contract Granularity section"
    fi

    # 1c: YAML frontmatter with framework field
    if head -20 "$CONTRACT" | grep -q '```yaml' && grep -q 'framework:' "$CONTRACT"; then
        pass "Contract has YAML frontmatter with framework field"
    else
        fail "Contract missing YAML frontmatter with framework field"
    fi

    # 1d: Minimum 30 lines (structural check, not a stub)
    line_count=$(wc -l < "$CONTRACT" | tr -d ' ')
    if [[ "$line_count" -ge 30 ]]; then
        pass "Contract has $line_count lines (>= 30 minimum)"
    else
        fail "Contract has only $line_count lines (< 30 minimum)"
    fi
fi

# =============================================================================
# 2. Crank SKILL.md checks
# =============================================================================
echo -e "${BLUE}[TEST-FIRST]${NC} Crank SKILL.md (skills/crank/SKILL.md)"

CRANK="skills/crank/SKILL.md"

if [[ ! -f "$CRANK" ]]; then
    fail "File missing: $CRANK"
else
    # 2a: --test-first in a table row (pipe-delimited, not just prose mention)
    if grep -qE '^\|.*--test-first.*\|' "$CRANK"; then
        pass "--test-first appears in a table row (flag table)"
    else
        fail "--test-first not found in any table row"
    fi

    # 2b: Step 3b: SPEC WAVE as ### heading
    if grep -qE '^### Step 3b: SPEC WAVE' "$CRANK"; then
        pass "Step 3b: SPEC WAVE section heading exists"
    else
        fail "Step 3b: SPEC WAVE section heading missing"
    fi

    # 2c: Step 3c: TEST WAVE as ### heading
    if grep -qE '^### Step 3c: TEST WAVE' "$CRANK"; then
        pass "Step 3c: TEST WAVE section heading exists"
    else
        fail "Step 3c: TEST WAVE section heading missing"
    fi

    # 2d: Category-based skip logic (spec-eligible or docs/chore pattern)
    if grep -qE 'spec-eligible|spec.eligible' "$CRANK" && grep -qE 'docs.*chore|chore.*docs' "$CRANK"; then
        pass "Category-based skip logic present (spec-eligible + docs/chore)"
    else
        fail "Category-based skip logic missing (need spec-eligible AND docs/chore references)"
    fi

    # 2e: Backward compat — Step 4 still exists
    if grep -qE '^### Step 4:' "$CRANK"; then
        pass "Backward compat: Step 4 still exists"
    else
        fail "Backward compat: Step 4 missing (standard wave execution removed)"
    fi

    # 2f: Backward compat — Step 0 still exists
    if grep -qE '^### Step 0:' "$CRANK"; then
        pass "Backward compat: Step 0 still exists (Load Knowledge Context)"
    else
        fail "Backward compat: Step 0 missing (Load Knowledge Context removed)"
    fi
fi

# =============================================================================
# 3. Wave-patterns.md checks
# =============================================================================
echo -e "${BLUE}[TEST-FIRST]${NC} Wave patterns (skills/crank/references/wave-patterns.md)"

WAVES="skills/crank/references/wave-patterns.md"

if [[ ! -f "$WAVES" ]]; then
    fail "File missing: $WAVES"
else
    # 3a: Spec-First Wave Model section
    if grep -qE '^## Spec-First Wave Model' "$WAVES"; then
        pass "Spec-First Wave Model section exists"
    else
        fail "Spec-First Wave Model section missing"
    fi

    # 3b: RED gate documented
    if grep -qE 'RED (confirmation|gate)|RED Confirmation Gate' "$WAVES"; then
        pass "RED gate documented"
    else
        fail "RED gate not documented"
    fi

    # 3c: GREEN gate documented
    if grep -qE 'GREEN (confirmation|gate)|GREEN Confirmation Gate' "$WAVES"; then
        pass "GREEN gate documented"
    else
        fail "GREEN gate not documented"
    fi

    # 3d: Category-based skip documented
    if grep -qE 'Category.Based Skip|category.based skip' "$WAVES"; then
        pass "Category-based skip documented"
    else
        fail "Category-based skip not documented"
    fi
fi

# =============================================================================
# 4. Implement SKILL.md checks
# =============================================================================
echo -e "${BLUE}[TEST-FIRST]${NC} Implement SKILL.md (skills/implement/SKILL.md)"

IMPL="skills/implement/SKILL.md"

if [[ ! -f "$IMPL" ]]; then
    fail "File missing: $IMPL"
else
    # 4a: ### GREEN Mode section heading
    if grep -qE '^### GREEN Mode' "$IMPL"; then
        pass "### GREEN Mode section heading exists"
    else
        fail "### GREEN Mode section heading missing"
    fi

    # 4b: Test immutability rule documented
    if grep -qE 'Do NOT modify test|MUST NOT modify.*test|tests are immutable' "$IMPL"; then
        pass "Test immutability rule documented"
    else
        fail "Test immutability rule not documented"
    fi
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${BLUE}=============================================${NC}"

if [[ $failed -gt 0 ]]; then
    echo -e "${RED}FAILED${NC} - $passed/$total passed, $failed failed"
    exit 1
else
    echo -e "${GREEN}PASSED${NC} - All $total checks passed"
    exit 0
fi
