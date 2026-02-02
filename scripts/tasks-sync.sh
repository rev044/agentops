#!/usr/bin/env bash
# tasks-sync.sh - Beads Light: sync tasks to/from .tasks.json
# Usage: tasks-sync.sh export|import|claim|complete|list

set -euo pipefail

TASKS_FILE=".tasks.json"
TASKS_LOCK=".tasks.lock"

# Acquire lock (simple file-based)
acquire_lock() {
    local timeout=10
    local count=0
    while [[ -f "$TASKS_LOCK" ]] && [[ $count -lt $timeout ]]; do
        sleep 1
        ((count++))
    done
    echo $$ > "$TASKS_LOCK"
}

release_lock() {
    rm -f "$TASKS_LOCK"
}

# Initialize empty tasks file if needed
init_tasks() {
    if [[ ! -f "$TASKS_FILE" ]]; then
        echo "[]" > "$TASKS_FILE"
    fi
}

# List tasks (for demigods to read)
cmd_list() {
    init_tasks
    cat "$TASKS_FILE" | jq -r '.[] | "\(.id)\t\(.status)\t\(.owner // "none")\t\(.subject)"'
}

# Get ready tasks (pending, no blockers, no owner)
cmd_ready() {
    init_tasks
    cat "$TASKS_FILE" | jq -r '.[] | select(.status == "pending" and (.owner == null or .owner == "")) | .id'
}

# Show a specific task
cmd_show() {
    local task_id="$1"
    init_tasks
    cat "$TASKS_FILE" | jq ".[] | select(.id == \"$task_id\")"
}

# Claim a task (set status=in_progress, owner=demigod-N)
cmd_claim() {
    local task_id="$1"
    local owner="${2:-$$}"

    acquire_lock
    init_tasks

    # Update the task
    local tmp
    tmp="$(mktemp)"
    cat "$TASKS_FILE" | jq "map(if .id == \"$task_id\" then .status = \"in_progress\" | .owner = \"$owner\" else . end)" > "$tmp"
    mv "$tmp" "$TASKS_FILE"

    release_lock
    echo "Claimed: $task_id by $owner"
}

# Complete a task
cmd_complete() {
    local task_id="$1"

    acquire_lock
    init_tasks

    local tmp
    tmp="$(mktemp)"
    cat "$TASKS_FILE" | jq "map(if .id == \"$task_id\" then .status = \"completed\" else . end)" > "$tmp"
    mv "$tmp" "$TASKS_FILE"

    release_lock
    echo "Completed: $task_id"
}

# Add a task (for manual creation)
cmd_add() {
    local subject="$1"
    local description="${2:-}"

    acquire_lock
    init_tasks

    # Generate ID
    local max_id
    max_id="$(cat "$TASKS_FILE" | jq -r '.[].id' | sort -n | tail -1)"
    local new_id
    new_id="$((${max_id:-0} + 1))"

    local tmp
    tmp="$(mktemp)"
    cat "$TASKS_FILE" | jq ". + [{\"id\": \"$new_id\", \"subject\": \"$subject\", \"description\": \"$description\", \"status\": \"pending\", \"owner\": null, \"blockedBy\": []}]" > "$tmp"
    mv "$tmp" "$TASKS_FILE"

    release_lock
    echo "Created task #$new_id: $subject"
}

# Export from description (Claude provides task data as JSON on stdin)
cmd_import_json() {
    acquire_lock
    cat > "$TASKS_FILE"
    release_lock
    echo "Imported tasks to $TASKS_FILE"
}

# Main
case "${1:-list}" in
    list)     cmd_list ;;
    ready)    cmd_ready ;;
    show)     cmd_show "$2" ;;
    claim)    cmd_claim "$2" "${3:-}" ;;
    complete) cmd_complete "$2" ;;
    add)      cmd_add "$2" "${3:-}" ;;
    import)   cmd_import_json ;;
    *)        echo "Usage: $0 list|ready|show|claim|complete|add|import"; exit 1 ;;
esac
