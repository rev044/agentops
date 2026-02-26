#!/usr/bin/env bash
# 50-agentops-bootstrap.sh — AgentOps setup hook for Gas Town Type 3 (rig setup hooks)
#
# Place this script in <rig>/.runtime/setup-hooks/ to auto-bootstrap AgentOps
# in new worktrees. Gas Town calls these scripts when creating worktrees.
#
# Contract:
#   - Receives GT_WORKTREE_PATH environment variable
#   - Numeric prefix (50-) controls execution order
#   - Must be idempotent (safe to run multiple times)
#   - Silent on success, stderr on failure
#   - Must complete within 10 seconds
#
# Usage:
#   cp hooks/examples/50-agentops-bootstrap.sh <rig>/.runtime/setup-hooks/
#   chmod +x <rig>/.runtime/setup-hooks/50-agentops-bootstrap.sh

set -euo pipefail

WORKTREE="${GT_WORKTREE_PATH:?GT_WORKTREE_PATH must be set}"

# Create .agents/ directory structure
for dir in research learnings patterns council knowledge/pending plans rpi ao handoff; do
    mkdir -p "$WORKTREE/.agents/$dir" 2>/dev/null || true
done

# Install minimal AgentOps hooks (SessionStart + SessionEnd + Stop)
if command -v ao &>/dev/null; then
    ao settings hooks install --target "$WORKTREE/.claude/settings.json" 2>/dev/null || true
fi

# Ensure .agents/ is gitignored
GITIGNORE="$WORKTREE/.gitignore"
if [ -f "$GITIGNORE" ]; then
    grep -q '\.agents/' "$GITIGNORE" 2>/dev/null || \
        printf '\n# AgentOps session artifacts\n.agents/\n' >> "$GITIGNORE"
elif [ -d "$WORKTREE/.git" ]; then
    printf '# AgentOps session artifacts\n.agents/\n' > "$GITIGNORE"
fi
