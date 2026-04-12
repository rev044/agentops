#!/usr/bin/env bash
#
# plan-branch-validate.sh - Run full validation suite before merge
#
# Usage: bash tools/hooks/plan-branch-validate.sh
#
# This script runs 3 levels of validation:
# Level 1: Quick checks (syntax, plan format)
# Level 2: CI pipeline simulation
# Level 3: Full test suite
#
# Exit code: 0 if all pass, 1 if any fail
#

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track validation results
FAILED=0

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Running Plan Validation Suite${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

# Level 1: Quick validation
echo -e "${BLUE}Level 1/3: Quick Validation${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

echo "→ Running: make quick"
if make quick > /tmp/plan-validate-quick.log 2>&1; then
  echo -e "  ${GREEN}✅ Quick validation passed${NC}"
else
  echo -e "  ${RED}❌ Quick validation failed${NC}"
  echo "  Log: /tmp/plan-validate-quick.log"
  FAILED=1
fi

echo ""
echo "→ Running: make check-plans"
if make check-plans > /tmp/plan-validate-plans.log 2>&1; then
  echo -e "  ${GREEN}✅ Plan format validation passed${NC}"
else
  echo -e "  ${RED}❌ Plan format validation failed${NC}"
  echo "  Log: /tmp/plan-validate-plans.log"
  FAILED=1
fi

echo ""

# Level 2: CI simulation
echo -e "${BLUE}Level 2/3: CI Pipeline Simulation${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

echo "→ Running: make ci-all"
if make ci-all > /tmp/plan-validate-ci.log 2>&1; then
  echo -e "  ${GREEN}✅ CI pipeline passed${NC}"
else
  echo -e "  ${RED}❌ CI pipeline failed${NC}"
  echo "  Log: /tmp/plan-validate-ci.log"
  FAILED=1
fi

echo ""

# Level 3: Test suite
echo -e "${BLUE}Level 3/3: Test Suite${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

echo "→ Running: uv run pytest tests/"
if uv run pytest tests/ > /tmp/plan-validate-tests.log 2>&1; then
  echo -e "  ${GREEN}✅ Test suite passed${NC}"
else
  echo -e "  ${RED}❌ Test suite failed${NC}"
  echo "  Log: /tmp/plan-validate-tests.log"
  FAILED=1
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"

# Final result
if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}✅ All validation checks passed!${NC}"
  echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
  echo ""
  echo -e "${GREEN}Ready to merge to main${NC}"
  echo ""
  exit 0
else
  echo -e "${RED}❌ Validation failed${NC}"
  echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
  echo ""
  echo -e "${YELLOW}Review logs in /tmp/plan-validate-*.log${NC}"
  echo "Fix issues and run validation again."
  echo ""
  exit 1
fi
