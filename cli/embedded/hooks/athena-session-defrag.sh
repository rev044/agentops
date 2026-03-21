#!/usr/bin/env bash
set -euo pipefail

# Kill switch
[[ "${AGENTOPS_HOOKS_DISABLED:-}" == "1" ]] && exit 0

# Only run defrag if ao CLI available
if ! command -v ao &>/dev/null; then
  exit 0
fi

# Lightweight defrag only — prune and dedup, no mining
if ao defrag --prune --dedup 2>/dev/null; then
  STATUS="completed"
else
  STATUS="failed (non-fatal)"
fi

echo "{\"hookSpecificOutput\":{\"hookEventName\":\"SessionEnd\",\"additionalContext\":\"Athena defrag ${STATUS}\"}}"
