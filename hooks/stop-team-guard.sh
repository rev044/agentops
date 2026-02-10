#!/bin/bash
# Stop hook: warn about teams that may have active members
# Only blocks stop if team members are actually running in tmux panes

# Kill switch
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

TEAMS_DIR="$HOME/.claude/teams"

# If teams directory doesn't exist, nothing to guard
[ ! -d "$TEAMS_DIR" ] && exit 0

# Find team config files
configs=$(find "$TEAMS_DIR" -maxdepth 2 -name "config.json" 2>/dev/null)

# If no configs found, safe to stop
[ -z "$configs" ] && exit 0

# Check each team for actually-running tmux panes
active_teams=""
while IFS= read -r cfg; do
    dir=$(dirname "$cfg")
    name=$(basename "$dir")

    # Extract tmux pane IDs from members (skip "in-process" and empty)
    pane_ids=$(grep -o '"tmuxPaneId"[[:space:]]*:[[:space:]]*"[^"]*"' "$cfg" 2>/dev/null \
        | sed 's/.*"tmuxPaneId"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' \
        | grep -v '^$' \
        | grep -v '^in-process$')

    # If no tmux panes, this team is stale (in-process agents die with session)
    [ -z "$pane_ids" ] && continue

    # Check if any pane is actually alive in tmux
    has_live_pane=false
    while IFS= read -r pane_id; do
        if tmux has-session -t "${pane_id%%.*}" 2>/dev/null; then
            has_live_pane=true
            break
        fi
    done <<< "$pane_ids"

    if [ "$has_live_pane" = "true" ]; then
        if [ -z "$active_teams" ]; then
            active_teams="$name"
        else
            active_teams="$active_teams, $name"
        fi
    fi
done <<< "$configs"

# Only block if there are teams with actually-running members
if [ -n "$active_teams" ]; then
    echo "Active teams with running members: ${active_teams}. Send shutdown_request to teammates or run TeamDelete before stopping." >&2
    exit 2
fi

# Stale team configs exist but no running members — safe to stop
exit 0
