#!/usr/bin/env bash
# Plan Lifecycle Validation - Check plan status consistency
#
# Validates that plans have consistent status, directory location, and completion state.
#
# Exit codes:
#   0 - All plans valid
#   1 - Validation errors found
#
# Usage:
#   plan-lifecycle-check.sh [--verbose]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARSER="$SCRIPT_DIR/plan-metadata-parser.sh"
PLANS_DIR="docs/reference/plans"
ERRORS=0
VERBOSE=0

# Parse arguments
if [ "${1:-}" = "--verbose" ]; then
  VERBOSE=1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

error() {
  echo -e "${RED}❌ $1${NC}" >&2
  ERRORS=$((ERRORS + 1))
}

warning() {
  echo -e "${YELLOW}⚠️  $1${NC}" >&2
}

success() {
  echo -e "${GREEN}✅ $1${NC}"
}

verbose() {
  if [ $VERBOSE -eq 1 ]; then
    echo "$1"
  fi
}

# Check if parser exists
if [ ! -x "$PARSER" ]; then
  error "Parser not found or not executable: $PARSER"
  exit 1
fi

# Check if plans directory exists
if [ ! -d "$PLANS_DIR" ]; then
  error "Plans directory not found: $PLANS_DIR"
  exit 1
fi

verbose "Validating plans in $PLANS_DIR..."
verbose ""

# Validation 1: Check plans in active/ directory
if [ -d "$PLANS_DIR/active" ]; then
  for plan in "$PLANS_DIR/active"/*.md; do
    [ -e "$plan" ] || continue  # Skip if no files
    [ "$(basename "$plan")" = "PLAN_TEMPLATE.md" ] && continue  # Skip template

    PLAN_NAME=$(basename "$plan")
    verbose "Checking: $PLAN_NAME"

    # Extract metadata
    STATUS=$("$PARSER" "$plan" "status" 2>/dev/null || echo "UNKNOWN")

    # Check 1: Status should be Active, Draft, or In Progress (not Complete)
    if [ "$STATUS" = "Complete" ] || [ "$STATUS" = "Deprecated" ] || [ "$STATUS" = "Superseded" ]; then
      error "$PLAN_NAME: Status is '$STATUS' but plan is in active/ directory"
      echo "   Fix: make plan-complete PLAN=$PLAN_NAME" >&2
      continue
    fi

    # Check 2: If all completion criteria checked, status should be Complete
    CRITERIA=$("$PARSER" "$plan" "completion_criteria" 2>/dev/null || echo "")
    if [ -n "$CRITERIA" ]; then
      TOTAL=$(echo "$CRITERIA" | wc -l | tr -d ' ' | tr -d '\n')
      COMPLETED=$(echo "$CRITERIA" | grep -c "\[x\]" || echo "0")

      # Only check if we have valid numbers
      if [[ "$TOTAL" =~ ^[0-9]+$ ]] && [[ "$COMPLETED" =~ ^[0-9]+$ ]]; then
        if [ "$TOTAL" -gt 0 ] && [ "$COMPLETED" -eq "$TOTAL" ]; then
          if [ "$STATUS" != "Complete" ]; then
            error "$PLAN_NAME: All completion criteria checked ($COMPLETED/$TOTAL), but status is '$STATUS'"
            echo "   Fix: make plan-complete PLAN=$PLAN_NAME" >&2
          fi
        fi
      fi
    fi
  done
fi

# Validation 2: Check plans in completed/ directory
if [ -d "$PLANS_DIR/completed" ]; then
  for plan in "$PLANS_DIR/completed"/*.md; do
    [ -e "$plan" ] || continue  # Skip if no files

    PLAN_NAME=$(basename "$plan")
    verbose "Checking: $PLAN_NAME (completed)"

    # Extract metadata
    STATUS=$("$PARSER" "$plan" "status" 2>/dev/null || echo "UNKNOWN")

    # Check: Status should be Complete (not Active/In Progress)
    if [ "$STATUS" = "Active" ] || [ "$STATUS" = "In Progress" ] || [ "$STATUS" = "Draft" ]; then
      error "$PLAN_NAME: Status is '$STATUS' but plan is in completed/ directory"
      echo "   Fix: Either update status to Complete or move back to active/" >&2
    fi
  done
fi

# Validation 3: Check plans in deprecated/ directory
if [ -d "$PLANS_DIR/deprecated" ]; then
  for plan in "$PLANS_DIR/deprecated"/*.md; do
    [ -e "$plan" ] || continue  # Skip if no files

    PLAN_NAME=$(basename "$plan")
    verbose "Checking: $PLAN_NAME (deprecated)"

    # Extract metadata
    STATUS=$("$PARSER" "$plan" "status" 2>/dev/null || echo "UNKNOWN")

    # Check: Status should be Deprecated or Superseded
    if [ "$STATUS" = "Active" ] || [ "$STATUS" = "In Progress" ] || [ "$STATUS" = "Complete" ]; then
      error "$PLAN_NAME: Status is '$STATUS' but plan is in deprecated/ directory"
      echo "   Fix: Either update status to Deprecated/Superseded or move to appropriate directory" >&2
    fi
  done
fi

# Summary
verbose ""
if [ $ERRORS -gt 0 ]; then
  echo ""
  error "Found $ERRORS plan lifecycle violation(s)"
  echo ""
  echo "Plans must have consistent status and directory location." >&2
  echo "Run with --verbose for full details." >&2
  exit 1
fi

success "All plans have consistent lifecycle status"
exit 0
