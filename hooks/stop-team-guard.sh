#!/bin/bash
# Stop hook: prevent orphaned teams by checking for active team configs

# Kill switch
[ "${AGENTOPS_HOOKS_DISABLED:-}" = "1" ] && exit 0

TEAMS_DIR="$HOME/.claude/teams"

# If teams directory doesn't exist, nothing to guard
[ ! -d "$TEAMS_DIR" ] && exit 0

# Find team config files
configs=$(find "$TEAMS_DIR" -maxdepth 2 -name "config.json" 2>/dev/null)

# If no configs found, safe to stop
[ -z "$configs" ] && exit 0

# Extract team names from directory paths
team_names=""
while IFS= read -r cfg; do
    dir=$(dirname "$cfg")
    name=$(basename "$dir")
    if [ -z "$team_names" ]; then
        team_names="$name"
    else
        team_names="$team_names, $name"
    fi
done <<< "$configs"

echo "Active teams found: ${team_names}. Send shutdown_request to teammates or run TeamDelete before stopping." >&2
exit 2
