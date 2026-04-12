#!/usr/bin/env bash
#
# plan-branch-create.sh - Create isolated branch for plan work
#
# Usage: bash tools/hooks/plan-branch-create.sh <plan-name>
# Example: bash tools/hooks/plan-branch-create.sh hooks-lifecycle
#
# This script:
# 1. Creates branch: plan/<plan-name>
# 2. Creates plan file from template (if doesn't exist)
# 3. Checks out the new branch
#

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get plan name from argument
PLAN_NAME="${1:-}"

if [ -z "$PLAN_NAME" ]; then
  echo -e "${RED}❌ Error: Plan name required${NC}"
  echo ""
  echo "Usage: bash tools/hooks/plan-branch-create.sh <plan-name>"
  echo "Example: bash tools/hooks/plan-branch-create.sh hooks-lifecycle"
  exit 1
fi

# Construct branch and file names
BRANCH="plan/$PLAN_NAME"
PLAN_FILE="docs/reference/plans/active/PLAN_${PLAN_NAME}.md"
PLAN_TEMPLATE="docs/reference/plans/PLAN_TEMPLATE.md"

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Creating Plan Branch: $BRANCH${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

# Check if branch already exists
if git show-ref --verify --quiet "refs/heads/$BRANCH"; then
  echo -e "${YELLOW}⚠️  Branch already exists: $BRANCH${NC}"
  echo ""
  read -p "Checkout existing branch? [y/N] " -n 1 -r
  echo ""
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    git checkout "$BRANCH"
    echo -e "${GREEN}✅ Checked out existing branch: $BRANCH${NC}"
    exit 0
  else
    echo "Aborted."
    exit 1
  fi
fi

# Create branch from current HEAD
git checkout -b "$BRANCH"
echo -e "${GREEN}✅ Created branch: $BRANCH${NC}"

# Create plan file from template if it doesn't exist
if [ -f "$PLAN_FILE" ]; then
  echo -e "${YELLOW}ℹ️  Plan file already exists: $PLAN_FILE${NC}"
else
  if [ ! -f "$PLAN_TEMPLATE" ]; then
    echo -e "${RED}❌ Error: Template not found: $PLAN_TEMPLATE${NC}"
    exit 1
  fi

  # Create plan from template
  mkdir -p "$(dirname "$PLAN_FILE")"
  cp "$PLAN_TEMPLATE" "$PLAN_FILE"

  # Update date in plan file
  TODAY=$(date +%Y-%m-%d)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/date: YYYY-MM-DD/date: $TODAY/" "$PLAN_FILE"
  else
    # Linux
    sed -i "s/date: YYYY-MM-DD/date: $TODAY/" "$PLAN_FILE"
  fi

  echo -e "${GREEN}✅ Created plan file: $PLAN_FILE${NC}"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Plan branch ready!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Edit plan file: $PLAN_FILE"
echo "2. Make changes and commit to this branch"
echo "3. When done: make plan-complete PLAN=$(basename "$PLAN_FILE")"
echo ""
echo -e "${YELLOW}💡 Tip: All work on this branch is isolated from main${NC}"
echo ""
