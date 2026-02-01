---
name: swarm
description: 'Spawn parallel Claude sessions for task execution. Uses .tasks.json for shared state, tmux for isolation. Triggers: "swarm", "spawn agents", "parallel work".'
---

# Swarm Skill

Spawn parallel Claude Code sessions (demigods) to work on tasks.

## Beads Light

Swarm uses `.tasks.json` - a simple file-based task system that demigods can share:

```json
[
  {"id": "1", "subject": "...", "description": "...", "status": "pending", "owner": null},
  {"id": "2", "subject": "...", "description": "...", "status": "in_progress", "owner": "demigod-1"}
]
```

**Commands** (via `scripts/tasks-sync.sh`):
- `./scripts/tasks-sync.sh list` - Show all tasks
- `./scripts/tasks-sync.sh ready` - Show ready task IDs
- `./scripts/tasks-sync.sh claim <id> <owner>` - Claim a task
- `./scripts/tasks-sync.sh complete <id>` - Mark complete
- `./scripts/tasks-sync.sh add "subject" "description"` - Add task

## Execution

Given `/swarm [--agents N]`:

### Step 1: Export Tasks to File

First, export your TaskList to `.tasks.json`:

```bash
# Create tasks array from current session tasks
# (You'll need to manually build this from TaskList output)
```

Or use the script to add tasks directly:
```bash
./scripts/tasks-sync.sh add "Implement feature X" "Full description here..."
./scripts/tasks-sync.sh add "Fix bug Y" "Details about the bug..."
```

### Step 2: Check Ready Tasks

```bash
./scripts/tasks-sync.sh ready
# Output: 1 2 3 (IDs of ready tasks)

READY_COUNT=$(./scripts/tasks-sync.sh ready | wc -w)
echo "$READY_COUNT tasks ready"
```

### Step 3: Spawn Demigods

For each ready task, spawn a demigod:

```bash
PROJECT=$(basename $(pwd))
SCRIPT_PATH="$(pwd)/scripts/tasks-sync.sh"

for TASK_ID in $(./scripts/tasks-sync.sh ready | head -${N:-5}); do
    # Get task details
    TASK_JSON=$(./scripts/tasks-sync.sh show $TASK_ID)
    SUBJECT=$(echo "$TASK_JSON" | jq -r '.subject')
    DESCRIPTION=$(echo "$TASK_JSON" | jq -r '.description')

    # Claim the task
    ./scripts/tasks-sync.sh claim $TASK_ID "demigod-$TASK_ID"

    # Spawn demigod
    tmux new-session -d -s "demigod-${PROJECT}-${TASK_ID}" \
        "cd $(pwd) && claude -p 'You are demigod-${TASK_ID}.

Your task: ${SUBJECT}

Details: ${DESCRIPTION}

When complete, run: ./scripts/tasks-sync.sh complete ${TASK_ID}

Do the work now.' 2>&1 | tee .demigod-${TASK_ID}.log"

    echo "Spawned demigod for task #$TASK_ID: $SUBJECT"
    sleep 30  # Stagger
done
```

### Step 4: Monitor

```bash
# Check tmux sessions
tmux list-sessions | grep demigod

# Check task status
./scripts/tasks-sync.sh list

# View demigod output
tail -f .demigod-<id>.log
```

### Step 5: Review & Cleanup

When all tasks show `completed`:

```bash
git status
git diff --stat

# Kill tmux sessions
tmux kill-session -t demigod-*

# Clean up
rm -f .demigod-*.log
```

## Demigod Instructions

Each demigod receives:
1. Task subject and description in the prompt
2. Command to run when complete: `./scripts/tasks-sync.sh complete <id>`

Demigods work independently and signal completion via the shared `.tasks.json` file.

## Quick Example

```bash
# Add some tasks
./scripts/tasks-sync.sh add "Create user model" "Add User struct with name, email fields"
./scripts/tasks-sync.sh add "Add validation" "Validate email format in User model"
./scripts/tasks-sync.sh add "Write tests" "Unit tests for User model"

# Check ready
./scripts/tasks-sync.sh ready  # 1 2 3

# Spawn swarm
/swarm --agents 3

# Monitor
./scripts/tasks-sync.sh list
tmux list-sessions | grep demigod

# When done
git add -A && git commit -m "User model with validation and tests"
```

## Key Points

- **`.tasks.json` is the source of truth** - Shared across all demigods
- **No external dependencies** - Just bash + jq
- **File locking** - Simple lock prevents race conditions
- **Claim before work** - Prevents duplicate work
- **Complete when done** - Updates shared state
