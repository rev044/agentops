#!/bin/bash
# PreCompact hook: snapshot team state before context compaction
# Captures active teams, git status, branch info for recovery after compaction
# Fail-open: all errors are non-fatal, always exit 0

ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo ".")
TEAMS_DIR="$HOME/.claude/teams"
AGENTS_DIR="$ROOT/.agents"
SNAP_DIR="$ROOT/.agents/compaction-snapshots"

# Check if there's anything worth snapshotting
has_teams=false
has_agents=false
[[ -d "$TEAMS_DIR" ]] && ls "$TEAMS_DIR"/*/config.json >/dev/null 2>&1 && has_teams=true
[[ -d "$AGENTS_DIR" ]] && has_agents=true

if ! $has_teams && ! $has_agents; then
  exit 0
fi

# Create snapshot directory
mkdir -p "$SNAP_DIR" 2>/dev/null || exit 0

TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
SNAP_FILE="$SNAP_DIR/${TIMESTAMP}.md"

# Gather data
BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
GIT_STATUS=$(git status --short 2>/dev/null | head -20)
GIT_DIFF_STAT=$(git diff --stat 2>/dev/null | tail -5)

TEAM_NAMES=""
if $has_teams; then
  for cfg in "$TEAMS_DIR"/*/config.json; do
    tname=$(basename "$(dirname "$cfg")")
    TEAM_NAMES="${TEAM_NAMES:+$TEAM_NAMES, }$tname"
  done
fi

# Write snapshot file
{
  echo "# Compaction Snapshot"
  echo ""
  echo "**Timestamp:** $TIMESTAMP"
  echo "**Branch:** $BRANCH"
  echo ""
  if [[ -n "$TEAM_NAMES" ]]; then
    echo "## Active Teams"
    echo "$TEAM_NAMES"
    echo ""
  fi
  if [[ -n "$GIT_STATUS" ]]; then
    echo "## Git Status"
    echo '```'
    echo "$GIT_STATUS"
    echo '```'
    echo ""
  fi
  if [[ -n "$GIT_DIFF_STAT" ]]; then
    echo "## Diff Stat"
    echo '```'
    echo "$GIT_DIFF_STAT"
    echo '```'
  fi
} > "$SNAP_FILE" 2>/dev/null

# Build compact summary for additionalContext (<500 bytes)
STATUS_COUNT=$(echo "$GIT_STATUS" | grep -c . 2>/dev/null || echo "0")
SUMMARY="branch=$BRANCH teams=[$TEAM_NAMES] files_changed=$STATUS_COUNT snapshot=$TIMESTAMP"
# Truncate to stay under 500 bytes
SUMMARY="${SUMMARY:0:480}"

# Output JSON for hook system
echo "{\"hookSpecificOutput\":{\"additionalContext\":\"$SUMMARY\"}}"

# Cleanup: keep last 5 snapshots, remove older
if [[ -d "$SNAP_DIR" ]]; then
  ls -t "$SNAP_DIR"/*.md 2>/dev/null | tail -n +6 | while read -r old; do
    rm -f "$old" 2>/dev/null
  done
fi

exit 0
