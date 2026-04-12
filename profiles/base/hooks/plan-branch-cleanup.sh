#!/usr/bin/env bash
#
# plan-branch-cleanup.sh - Delete merged or abandoned plan branches
#
# Usage: bash tools/hooks/plan-branch-cleanup.sh <plan-branch> [--merged|--abandon]
# Example: bash tools/hooks/plan-branch-cleanup.sh plan/hooks-lifecycle --merged
#
# Modes:
#   --merged  : Delete branch that was merged (safe delete, -d)
#   --abandon : Force delete unmerged branch (dangerous, -D)
#

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get arguments
PLAN_BRANCH="${1:-}"
MODE="${2:---merged}"

if [ -z "$PLAN_BRANCH" ]; then
  echo -e "${RED}❌ Error: Plan branch required${NC}"
  echo ""
  echo "Usage: bash tools/hooks/plan-branch-cleanup.sh <plan-branch> [--merged|--abandon]"
  echo ""
  echo "Modes:"
  echo "  --merged  : Delete merged branch (safe, default)"
  echo "  --abandon : Force delete unmerged branch (dangerous)"
  echo ""
  echo "Examples:"
  echo "  bash tools/hooks/plan-branch-cleanup.sh plan/hooks-lifecycle --merged"
  echo "  bash tools/hooks/plan-branch-cleanup.sh plan/failed-experiment --abandon"
  exit 1
fi

# Validate branch name format
if [[ ! "$PLAN_BRANCH" =~ ^plan/ ]]; then
  echo -e "${RED}❌ Error: Not a plan branch: $PLAN_BRANCH${NC}"
  echo "Plan branches must start with 'plan/'"
  exit 1
fi

# Check if branch exists
if ! git show-ref --verify --quiet "refs/heads/$PLAN_BRANCH"; then
  echo -e "${RED}❌ Error: Branch does not exist: $PLAN_BRANCH${NC}"
  exit 1
fi

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Can't delete current branch
if [ "$CURRENT_BRANCH" = "$PLAN_BRANCH" ]; then
  echo -e "${RED}❌ Error: Cannot delete current branch${NC}"
  echo "Switch to another branch first:"
  echo "  git checkout main"
  exit 1
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Deleting Plan Branch: $PLAN_BRANCH${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""

# Handle based on mode
if [ "$MODE" = "--merged" ]; then
  echo -e "${BLUE}Mode: Safe delete (merged branches only)${NC}"
  echo ""

  # Check if branch is merged
  if git branch --merged main | grep -q "$PLAN_BRANCH"; then
    echo "→ Deleting merged branch..."
    if git branch -d "$PLAN_BRANCH"; then
      echo -e "  ${GREEN}✅ Branch deleted: $PLAN_BRANCH${NC}"
    else
      echo -e "  ${RED}❌ Failed to delete branch${NC}"
      exit 1
    fi
  else
    echo -e "${YELLOW}⚠️  Branch is not merged into main${NC}"
    echo ""
    echo "Cannot safely delete. Options:"
    echo "1. Merge first: make plan-merge"
    echo "2. Force delete: bash tools/hooks/plan-branch-cleanup.sh $PLAN_BRANCH --abandon"
    exit 1
  fi

elif [ "$MODE" = "--abandon" ]; then
  echo -e "${YELLOW}Mode: Force delete (abandon unmerged work)${NC}"
  echo ""

  # Check if branch is merged (warn if deleting merged branch)
  if git branch --merged main | grep -q "$PLAN_BRANCH"; then
    echo -e "${YELLOW}⚠️  Note: This branch is already merged${NC}"
    echo "You can use --merged mode instead."
    echo ""
  fi

  # Confirm before force delete
  echo -e "${RED}⚠️  WARNING: This will permanently delete unmerged work!${NC}"
  echo ""
  read -p "Are you sure you want to abandon $PLAN_BRANCH? [y/N] " -n 1 -r
  echo ""

  if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "→ Force deleting branch..."
    if git branch -D "$PLAN_BRANCH"; then
      echo -e "  ${GREEN}✅ Branch force deleted: $PLAN_BRANCH${NC}"
    else
      echo -e "  ${RED}❌ Failed to delete branch${NC}"
      exit 1
    fi
  else
    echo "Aborted."
    exit 1
  fi

else
  echo -e "${RED}❌ Error: Invalid mode: $MODE${NC}"
  echo "Use --merged or --abandon"
  exit 1
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Cleanup complete${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
