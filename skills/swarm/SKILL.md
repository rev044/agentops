---
name: swarm
description: 'Spawn background crank loops for parallel task execution. Pure Claude-native using Task tool. Triggers: "swarm", "spawn agents", "parallel work", "crank in parallel".'
---

# Swarm Skill

Spawn isolated crank loops as background agents to execute tasks in parallel.

## Architecture (Mayor-First)

```
Mayor (this session)
    |
    +-> Plan: TaskCreate with dependencies (blockedBy)
    |
    +-> Identify wave: tasks with no blockers
    |
    +-> Spawn: Task tool (run_in_background=true) for each
    |       Each agent runs crank loop on its task
    |
    +-> Monitor: TaskOutput to check progress
    |
    +-> Validate: Review changes when complete
    |
    +-> Repeat: New plan if more work needed
```

## Execution

Given `/swarm`:

### Step 1: Ensure Tasks Exist

Use TaskList to see current tasks. If none, create them:

```
TaskCreate(
  subject="Implement feature X",
  description="Full details...",
  blockedBy=[]  # or list of task IDs
)
```

### Step 2: Identify Wave

Find tasks that are:
- Status: `pending`
- No blockedBy (or all blockers completed)

These can run in parallel.

### Step 3: Spawn Crank Loops

For each ready task, spawn a background agent:

```
Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are a crank loop agent.

Your task ID: #<id>
Subject: <subject>
Description: <description>

CRANK LOOP:
1. Read the task details
2. Do the work (edit files, write code, run tests)
3. Commit your changes with a clear message
4. Signal completion

Work in the directory: <cwd>

Start now. Complete the task fully."
)
```

The Task tool returns a `task_id` for monitoring.

### Step 4: Monitor Progress

Use TaskOutput to check on background agents:

```
TaskOutput(task_id="<agent-task-id>", block=false)
```

Or wait for completion:
```
TaskOutput(task_id="<agent-task-id>", block=true, timeout=300000)
```

### Step 5: Validate & Review

When agents complete:
1. Check git status for changes
2. Review diffs
3. Run tests/validation
4. Commit combined work if needed

### Step 6: Repeat if Needed

If more tasks remain:
1. Check TaskList for next wave
2. Spawn new agents
3. Continue until all done

## Example Flow

```
Mayor: "Let's build a user auth system"

1. /plan → Creates tasks:
   #1 [pending] Create User model
   #2 [pending] Add password hashing (blockedBy: #1)
   #3 [pending] Create login endpoint (blockedBy: #1)
   #4 [pending] Add JWT tokens (blockedBy: #3)
   #5 [pending] Write tests (blockedBy: #2, #3, #4)

2. /swarm → Spawns agent for #1 (only unblocked task)

3. Agent #1 completes → #1 now completed
   → #2 and #3 become unblocked

4. /swarm → Spawns agents for #2 and #3 in parallel

5. Continue until #5 completes

6. /vibe → Validate everything
```

## Key Points

- **Pure Claude-native** - No tmux, no external scripts
- **Background agents** - `run_in_background=true` for isolation
- **Wave execution** - Only unblocked tasks spawn
- **Mayor orchestrates** - You control the flow
- **Crank loops** - Each agent works until task done

## Integration with AgentOps

This ties into the full workflow:

```
/research → Understand the problem
/plan → Decompose into tasks
/swarm → Execute in parallel
/vibe → Validate results
/post-mortem → Extract learnings
```

The knowledge flywheel captures learnings from each crank loop.

## Monitoring Commands

```
# Check background agent
TaskOutput(task_id="abc123", block=false)

# Wait for agent to finish
TaskOutput(task_id="abc123", block=true)

# List all tasks
TaskList()
```

## When to Use Swarm vs Crank

| Scenario | Use |
|----------|-----|
| Multiple independent tasks | `/swarm` (parallel) |
| Sequential dependencies | `/crank` (serial) |
| Mix of both | `/swarm` spawns waves, each wave parallel |

## Why This Works: Ralph Wiggum Pattern

This architecture follows the [Ralph Wiggum Pattern](https://ghuntley.com/ralph/) for autonomous agents.

**Core Insight:** Each `Task(run_in_background=true)` spawn = fresh context.

```
Ralph's bash loop:          Our swarm:
while :; do                 Mayor spawns Task → fresh context
  cat PROMPT.md | claude    Mayor spawns Task → fresh context
done                        Mayor spawns Task → fresh context
```

Both achieve the same thing: **fresh context per execution unit**.

### Why Fresh Context Matters

| Approach | Context | Problem |
|----------|---------|---------|
| Internal loop in agent | Accumulates | Degrades over iterations |
| Mayor spawns agents | Fresh each time | Stays effective at scale |

Making demigods loop internally would violate Ralph - context accumulates within the session. The loop belongs in Mayor (lightweight orchestration), fresh context belongs in demigods (heavyweight work).

### Key Properties

- **Mayor IS the loop** - Orchestration layer, manages state
- **Demigods are atomic** - One task, one spawn, one result
- **TaskList as memory** - State persists in task status, not context
- **Filesystem for artifacts** - Files written, commits made

This is **Ralph + parallelism**: the while loop is distributed across wave spawns, with multiple agents per wave.
