---
name: swarm
description: 'Spawn parallel Claude sessions for task execution. Uses native TaskList for work, tmux for isolation. Triggers: "swarm", "spawn agents", "parallel work".'
---

# Swarm Skill

Spawn parallel Claude Code sessions (demigods) to execute tasks.

## How It Works

```
Mayor (you)
    |
    +-> TaskList → find ready tasks (pending, no blockers)
    |
    +-> Group into wave (tasks that can run in parallel)
    |
    +-> For each task in wave:
    |       tmux new-session → claude --prompt "Complete task #N"
    |
    +-> Monitor → check task status
    |
    +-> Review when complete
```

## Execution

Given `/swarm [--agents N]`:

### Step 1: Get Ready Tasks

Use TaskList to find tasks that are:
- Status: pending
- No blockedBy (or all blockedBy tasks completed)

```
Ready tasks = pending tasks with no active blockers
Wave size = min(N or 5, ready count)
```

### Step 2: Spawn Demigods

For each task in the wave, spawn a tmux session:

```bash
PROJECT=$(basename $(pwd))

# For each ready task ID:
tmux new-session -d -s "demigod-${PROJECT}-${TASK_ID}" \
    "claude --print --prompt 'Complete task #${TASK_ID}. Use TaskGet to read it, do the work, then TaskUpdate status=completed when done.'"

echo "Spawned demigod for task #${TASK_ID}"
sleep 30  # Stagger to avoid rate limits
```

### Step 3: Monitor

Check progress:
```bash
# List active sessions
tmux list-sessions | grep demigod

# Check task status
# Use TaskList - look for completed vs in_progress
```

### Step 4: Review

When all demigods complete:
```bash
git status
git diff --stat
```

Then commit the combined work.

## Example

```bash
# You have 6 tasks, 4 ready (no blockers)
/swarm --agents 4

# Spawns:
# demigod-myproject-1
# demigod-myproject-2
# demigod-myproject-3
# demigod-myproject-4

# Each runs: claude --prompt "Complete task #N..."

# Monitor:
tmux list-sessions | grep demigod

# When done, review and commit
```

## Key Points

- **TaskList is the work source** - No external dependencies
- **Waves from blockedBy** - Only ready tasks spawn
- **tmux for isolation** - Each demigod is independent
- **30s stagger** - Prevents API rate limits
- **Mayor reviews** - Always review before committing

## Killing Demigods

```bash
# Kill one
tmux kill-session -t demigod-myproject-1

# Kill all
tmux kill-server
```

## Without tmux

If you can't use tmux, use the Task tool to spawn subagents instead. Less isolation but works:

```
Task(subagent_type="general-purpose", prompt="Complete task #N...")
```

This is faster but subagents share context and die with your session.
