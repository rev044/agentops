#!/bin/bash
# Dangerous Git Operations Guard
# Blocks destructive git commands and suggests safe alternatives.

[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

# Read all stdin
INPUT=$(cat)

# Extract tool_input.command from JSON
COMMAND=$(echo "$INPUT" | grep -o '"command"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/^"command"[[:space:]]*:[[:space:]]*"//;s/"$//')

# Hot path: no git, no problem
echo "$COMMAND" | grep -q "git" || exit 0

# Allow-list (checked before block-list)
echo "$COMMAND" | grep -qE 'push.*--force-with-lease' && exit 0

# Block-list with safe alternatives
if echo "$COMMAND" | grep -qE 'push\s+.*(-f|--force)'; then
  echo "Blocked: force push. Use --force-with-lease instead." >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'reset\s+--hard'; then
  echo "Blocked: hard reset. Use git stash or git reset --soft." >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'clean\s+-f'; then
  echo "Blocked: force clean. Review with git clean -n first." >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'checkout\s+\.'; then
  echo "Blocked: checkout dot. Use git stash to preserve changes." >&2
  exit 2
fi

if echo "$COMMAND" | grep -qE 'branch\s+-D'; then
  echo "Blocked: force branch delete. Use git branch -d (safe delete)." >&2
  exit 2
fi

# No match — allow
exit 0
