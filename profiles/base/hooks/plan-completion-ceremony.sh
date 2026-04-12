#!/usr/bin/env bash
# Plan Completion Ceremony - Guided workflow to mark plan as complete
#
# Usage:
#   plan-completion-ceremony.sh <plan_filename>
#
# Example:
#   plan-completion-ceremony.sh PLAN_FOO.md
#
# What it does:
#   1. Validates all completion criteria are checked
#   2. Updates status to "Complete"
#   3. Moves plan to completed/ directory
#   4. Commits with standardized message
#
# Exit codes:
#   0 - Success
#   1 - Error (validation failed, user cancelled, etc.)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARSER="$SCRIPT_DIR/plan-metadata-parser.sh"
PLANS_DIR="docs/reference/plans"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

error() {
  echo -e "${RED}❌ $1${NC}" >&2
}

success() {
  echo -e "${GREEN}✅ $1${NC}"
}

info() {
  echo -e "${BLUE}ℹ️  $1${NC}"
}

warning() {
  echo -e "${YELLOW}⚠️  $1${NC}"
}

# Check arguments
PLAN_NAME="${1:-}"
if [ -z "$PLAN_NAME" ]; then
  error "Usage: $0 <plan_filename>"
  echo "Example: $0 PLAN_FOO.md" >&2
  exit 1
fi

# Find plan file
PLAN_FILE=""
if [ -f "$PLANS_DIR/active/$PLAN_NAME" ]; then
  PLAN_FILE="$PLANS_DIR/active/$PLAN_NAME"
elif [ -f "$PLANS_DIR/$PLAN_NAME" ]; then
  PLAN_FILE="$PLANS_DIR/$PLAN_NAME"
else
  error "Plan not found: $PLAN_NAME"
  echo "Searched in:" >&2
  echo "  - $PLANS_DIR/active/$PLAN_NAME" >&2
  echo "  - $PLANS_DIR/$PLAN_NAME" >&2
  exit 1
fi

echo ""
info "Plan Completion Ceremony"
echo "========================="
echo ""
echo "Plan: $(basename "$PLAN_FILE")"
echo "Location: $PLAN_FILE"
echo ""

# Step 1: Check current status
CURRENT_STATUS=$("$PARSER" "$PLAN_FILE" "status" 2>/dev/null || echo "UNKNOWN")
info "Current status: $CURRENT_STATUS"

if [ "$CURRENT_STATUS" = "Complete" ]; then
  warning "Plan is already marked as Complete"
  echo ""
  read -p "Do you want to move it to completed/ directory anyway? [y/N] " -n 1 -r
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Ceremony cancelled."
    exit 0
  fi
fi

# Step 2: Validate completion criteria
echo ""
info "Checking completion criteria..."
CRITERIA=$("$PARSER" "$PLAN_FILE" "completion_criteria" 2>/dev/null || echo "")

if [ -z "$CRITERIA" ]; then
  warning "No completion criteria found in plan"
  echo ""
  read -p "Mark as complete anyway? [y/N] " -n 1 -r
  echo ""
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Ceremony cancelled."
    exit 1
  fi
else
  TOTAL=$(echo "$CRITERIA" | wc -l | tr -d ' ' | tr -d '\n')
  COMPLETED=$(echo "$CRITERIA" | grep -c "\[x\]" || echo "0")
  INCOMPLETE=$(echo "$CRITERIA" | grep -c "\[ \]" || echo "0")

  echo ""
  echo "Completion criteria status:"
  echo "  Total: $TOTAL"
  echo "  Completed: $COMPLETED"
  echo "  Incomplete: $INCOMPLETE"
  echo ""

  if [ "$INCOMPLETE" -gt 0 ]; then
    error "Plan has $INCOMPLETE incomplete criteria"
    echo ""
    echo "Incomplete criteria:"
    echo "$CRITERIA" | grep "\[ \]"
    echo ""
    read -p "Mark as complete anyway? [y/N] " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Ceremony cancelled. Please complete all criteria first."
      exit 1
    fi
  else
    success "All completion criteria are checked!"
  fi
fi

# Step 3: Update status to Complete
echo ""
info "Updating status to 'Complete'..."

# Use sed to update the status field in YAML frontmatter
sed -i.bak 's/^status: .*/status: Complete/' "$PLAN_FILE"
rm -f "${PLAN_FILE}.bak"

success "Status updated"

# Step 4: Move to completed/ directory
DEST_DIR="$PLANS_DIR/completed"
mkdir -p "$DEST_DIR"

DEST_FILE="$DEST_DIR/$(basename "$PLAN_FILE")"

echo ""
info "Moving plan to completed/ directory..."
mv "$PLAN_FILE" "$DEST_FILE"

success "Plan moved to: $DEST_FILE"

# Step 5: Stage changes
echo ""
info "Staging changes for commit..."
git add "$DEST_FILE"

# Remove old file from git if it exists (it was moved)
if git ls-files --error-unmatch "$PLAN_FILE" >/dev/null 2>&1; then
  git rm "$PLAN_FILE" >/dev/null 2>&1 || true
fi

success "Changes staged"

# Step 6: Check if on plan branch - if so, validate and merge
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [[ "$CURRENT_BRANCH" == plan/* ]]; then
  echo ""
  echo "═══════════════════════════════════════════════════"
  info "Plan branch detected: $CURRENT_BRANCH"
  echo "═══════════════════════════════════════════════════"
  echo ""
  info "Initiating branch merge workflow..."
  echo ""

  # Commit plan completion on the branch
  info "Committing plan completion..."
  git commit -m "docs: Mark $(basename "$PLAN_FILE" .md) as complete

Context: Plan execution completed. All completion criteria met.

Solution: Updated plan status to Complete and moved to completed/ directory.

Learning: Plan lifecycle enforcement ensures plans stay current and don't
become stale anti-patterns. The ceremony workflow makes completion easy.

Impact: Plan properly archived. Repository hygiene maintained.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

  success "Plan completion committed"
  echo ""

  # Run validation
  info "Running validation suite before merge..."
  echo ""
  if ! bash "$SCRIPT_DIR/plan-branch-validate.sh"; then
    echo ""
    error "Validation failed. Plan marked complete but not merged."
    echo ""
    echo "Options:"
    echo "1. Fix issues and run: make plan-validate"
    echo "2. Then merge: make plan-merge"
    echo "3. Or abandon: make plan-abandon"
    exit 1
  fi

  echo ""

  # Merge to main
  info "Merging to main..."
  echo ""
  if ! bash "$SCRIPT_DIR/plan-branch-merge.sh" "$CURRENT_BRANCH"; then
    error "Merge failed"
    echo ""
    echo "Branch: $CURRENT_BRANCH (still exists)"
    echo "Plan: Marked complete but not merged"
    echo ""
    echo "Resolve issues and retry: make plan-merge"
    exit 1
  fi

  echo ""

  # Cleanup branch
  info "Cleaning up branch..."
  echo ""
  bash "$SCRIPT_DIR/plan-branch-cleanup.sh" "$CURRENT_BRANCH" --merged

  echo ""
  echo "═══════════════════════════════════════════════════"
  success "Plan complete, validated, and merged to main!"
  echo "═══════════════════════════════════════════════════"
  echo ""
  echo "Branch: $CURRENT_BRANCH (deleted)"
  echo "Current branch: main"
  echo "Tag created: plan-completed/${CURRENT_BRANCH#plan/}"
  echo ""
  echo "Next step: Push to remote"
  echo ""
  echo "Run: git push origin main --tags"
  echo ""

else
  # Not on plan branch - use standard ceremony
  echo ""
  echo "═══════════════════════════════════════════════════"
  success "Ceremony complete!"
  echo "═══════════════════════════════════════════════════"
  echo ""
  echo "Next step: Commit the changes"
  echo ""
  echo "Suggested commit message:"
  echo ""
  echo "---"
  echo "docs: Mark $(basename "$PLAN_FILE" .md) as complete"
  echo ""
  echo "Context: Plan execution completed. All completion criteria met."
  echo ""
  echo "Solution: Updated plan status to Complete and moved to completed/ directory."
  echo ""
  echo "Learning: Plan lifecycle enforcement ensures plans stay current and don't"
  echo "become stale anti-patterns. The ceremony workflow makes completion easy."
  echo ""
  echo "Impact: Plan properly archived. Repository hygiene maintained."
  echo ""
  echo "🤖 Generated with [Claude Code](https://claude.com/claude-code)"
  echo ""
  echo "Co-Authored-By: Claude <noreply@anthropic.com>"
  echo "---"
  echo ""
  echo "Run: git commit"
  echo ""
fi

exit 0
