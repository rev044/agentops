#!/bin/bash
# Lightweight skill lint gate — checks line count limits on SKILL.md edits.
# Runs as PreToolUse hook on Write|Edit. Fast path: exits immediately if
# the file being edited is not a SKILL.md.
set -euo pipefail

# Fast path: only check SKILL.md files
FILE_PATH=$(echo "${TOOL_INPUT:-}" | jq -r '.file_path // empty' 2>/dev/null)
case "$FILE_PATH" in
  */skills/*/SKILL.md) ;;
  *) exit 0 ;;
esac

# File must exist (new files won't have content yet)
[ -f "$FILE_PATH" ] || exit 0

SKILL_DIR=$(dirname "$FILE_PATH")
SKILL_NAME=$(basename "$SKILL_DIR")

# Extract tier from frontmatter
TIER=$(awk 'BEGIN{n=0} /^---$/{n++; if(n==2) exit; next} n==1{print}' "$FILE_PATH" \
  | grep '^[[:space:]]*tier:' | head -1 | sed 's/^[[:space:]]*tier:[[:space:]]*//' | tr -d '\r')

# Set limit based on tier
case "$TIER" in
  library|background) LIMIT=200 ;;
  orchestration)      LIMIT=550 ;;
  *)                  LIMIT=500 ;;
esac

LINE_COUNT=$(wc -l < "$FILE_PATH" | tr -d ' ')

if [ "$LINE_COUNT" -gt "$LIMIT" ]; then
  echo "⚠️ SKILL LINT: ${SKILL_NAME} is ${LINE_COUNT} lines (limit: ${LIMIT} for tier=${TIER}). Extract content to references/."
fi

# Always exit 0 — this is a warning, not a blocker
exit 0
