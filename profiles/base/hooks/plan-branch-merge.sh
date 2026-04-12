#!/usr/bin/env bash
#
# plan-branch-merge.sh - Merge plan branch to main with validation
#
# Usage: bash tools/hooks/plan-branch-merge.sh <plan-branch>
# Example: bash tools/hooks/plan-branch-merge.sh plan/hooks-lifecycle
#
# This script:
# 1. Validates you're on the plan branch
# 2. Runs full validation suite
# 3. Updates main from remote (if available)
# 4. Merges plan branch to main (--no-ff)
# 5. Tags the merge point
#

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get plan branch from argument
PLAN_BRANCH="${1:-}"

if [ -z "$PLAN_BRANCH" ]; then
  echo -e "${RED}❌ Error: Plan branch required${NC}"
  echo ""
  echo "Usage: bash tools/hooks/plan-branch-merge.sh <plan-branch>"
  echo "Example: bash tools/hooks/plan-branch-merge.sh plan/hooks-lifecycle"
  exit 1
fi

# Validate branch name format
if [[ ! "$PLAN_BRANCH" =~ ^plan/ ]]; then
  echo -e "${RED}❌ Error: Not a plan branch: $PLAN_BRANCH${NC}"
  echo "Plan branches must start with 'plan/'"
  exit 1
fi

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Must be on the plan branch
if [ "$CURRENT_BRANCH" != "$PLAN_BRANCH" ]; then
  echo -e "${RED}❌ Error: Not on plan branch: $PLAN_BRANCH${NC}"
  echo "Current branch: $CURRENT_BRANCH"
  echo ""
  echo "Switch to the plan branch first:"
  echo "  git checkout $PLAN_BRANCH"
  exit 1
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Merging Plan Branch: $PLAN_BRANCH${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

# Step 1: Run validation
echo -e "${BLUE}Step 1/4: Validation${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

if ! bash tools/hooks/plan-branch-validate.sh; then
  echo ""
  echo -e "${RED}❌ Validation failed. Cannot merge.${NC}"
  echo ""
  echo "Fix issues and run validation again:"
  echo "  make plan-validate"
  exit 1
fi

echo ""

# Step 2: Update main from remote (if remote exists)
echo -e "${BLUE}Step 2/4: Update Main${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

git checkout main

# Check if remote exists
if git remote | grep -q origin; then
  echo "→ Pulling latest from origin/main..."
  if git pull origin main --ff-only 2>&1; then
    echo -e "  ${GREEN}✅ Main branch updated${NC}"
  else
    echo -e "  ${YELLOW}⚠️  Could not fast-forward main${NC}"
    echo ""
    echo "Your main branch has diverged from origin."
    echo "Please rebase your plan branch:"
    echo "  git checkout $PLAN_BRANCH"
    echo "  git rebase main"
    echo "  make plan-merge"
    git checkout "$PLAN_BRANCH"
    exit 1
  fi
else
  echo -e "  ${YELLOW}ℹ️  No remote configured, skipping update${NC}"
fi

echo ""

# Step 3: Merge plan branch
echo -e "${BLUE}Step 3/4: Merge Branch${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

PLAN_NAME=$(echo "$PLAN_BRANCH" | sed 's/plan\///')
MERGE_MSG="Merge $PLAN_BRANCH (plan complete)

This merge completes the plan: $PLAN_NAME

All validation gates passed:
- Quick validation (syntax, format)
- CI pipeline simulation
- Full test suite

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

echo "→ Merging $PLAN_BRANCH to main..."
if git merge --no-ff "$PLAN_BRANCH" -m "$MERGE_MSG"; then
  echo -e "  ${GREEN}✅ Merge successful${NC}"
else
  echo -e "  ${RED}❌ Merge failed${NC}"
  echo ""
  echo "Resolve conflicts and complete merge manually:"
  echo "  # Fix conflicts"
  echo "  git add ."
  echo "  git commit"
  exit 1
fi

echo ""

# Step 4: Tag the merge
echo -e "${BLUE}Step 4/4: Tag Merge${NC}"
echo "─────────────────────────────────────────────────────"
echo ""

TAG_NAME="plan-completed/$PLAN_NAME"
TAG_MSG="Plan completed and merged: $PLAN_NAME

Merge commit: $(git rev-parse HEAD)
Branch: $PLAN_BRANCH

All validation gates passed."

echo "→ Creating tag: $TAG_NAME"
git tag -a "$TAG_NAME" -m "$TAG_MSG"
echo -e "  ${GREEN}✅ Tag created${NC}"

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Plan branch merged successfully!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Delete plan branch: bash tools/hooks/plan-branch-cleanup.sh $PLAN_BRANCH --merged"
echo "2. Push to remote: git push origin main --tags"
echo ""
echo -e "${YELLOW}💡 Tip: You're now on main branch${NC}"
echo ""
