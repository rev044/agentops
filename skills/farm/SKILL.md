---
name: farm
description: 'Spawn Agent Farm for parallel issue execution. Mayor orchestrates crew via tmux + MCP Agent Mail. Replaces /crank for multi-agent work. Triggers: "farm", "spawn agents", "parallel work", "multi-agent".'
---

# Farm Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Spawn an Agent Farm to execute multiple tasks in parallel with witness monitoring.

## Work Source Priority

Farm uses the first available work source:
1. **Native Tasks** (TaskList) - Claude Code's built-in task management
2. **Beads** (.beads/issues.jsonl) - Git-native issue tracking

This allows farm to work without beads installed.

## Architecture

```
Mayor (this session)
    |
    +-> Detect work source (tasks vs beads)
    |
    +-> Spawn agents in tmux (serial, 30s stagger)
    |       |
    |       +-> Each agent: claim task → execute → mark complete
    |
    +-> Spawn witness in separate tmux session
    |
    +-> Monitor via inbox/TaskList
    |
    +-> "FARM COMPLETE" when all done
    |
    +-> /post-mortem (extract learnings)
```

## Execution Steps

Given `/farm [--agents N]`:

### Step 1: Detect Work Source & Pre-Flight

**First, detect work source by checking for pending tasks:**

Use the TaskList tool to check for pending tasks. If tasks exist with status "pending" and no blockedBy, use native tasks. Otherwise fall back to beads.

**Work source detection:**
- If TaskList returns pending tasks → use NATIVE TASKS
- Else if `.beads/issues.jsonl` exists → use BEADS
- Else STOP: "No work found. Create tasks or run /plan first."

**For native tasks, count ready work:**
- Ready = status "pending" AND blockedBy is empty or all blockedBy tasks completed

**For beads, run pre-flight:**
```bash
ao farm validate 2>/dev/null
```

Or manual checks:
```bash
bd ready 2>/dev/null | wc -l
```

**Always check disk space:**
```bash
df -h . | awk 'NR==2 {print $4}'
```
If < 5GB, warn user.

### Step 2: Determine Agent Count

**Default:** `N = min(5, ready_count)`

**If --agents specified:** Use that value, but cap at ready count.

**For native tasks:**
- Count pending tasks with no blockers from TaskList output
- READY_COUNT = number of such tasks

**For beads:**
```bash
READY_COUNT=$(bd ready 2>/dev/null | wc -l | tr -d ' ')
```

```
AGENTS = min(N or 5, READY_COUNT)
```

If AGENTS = 0, STOP: "No work available for agents."

### Step 3: Spawn Agent Farm

**For native tasks, spawn agents via Task tool:**

Use the Task tool to spawn parallel agents. Each agent gets instructions to:
1. Check TaskList for pending tasks with no blockers
2. Claim a task (TaskUpdate: status=in_progress, owner=agent-N)
3. Execute the task (read description, do the work)
4. Mark complete (TaskUpdate: status=completed)
5. Repeat until no pending tasks remain

**Spawn agents in parallel using Task tool:**
```
For each agent 1..N:
  Task(subagent_type="general-purpose", prompt="
    You are Agent-{N} in an agent farm.

    Your job:
    1. Use TaskList to find pending tasks
    2. Pick the lowest ID task that has no blockedBy (or all blockedBy completed)
    3. Use TaskUpdate to set status=in_progress and owner=agent-{N}
    4. Use TaskGet to read full description
    5. Execute the task (edit files, run commands as needed)
    6. Use TaskUpdate to set status=completed
    7. Repeat until TaskList shows no pending tasks

    When done, output: AGENT-{N} COMPLETE
  ")
```

**For beads, use ao farm:**
```bash
ao farm start --agents $AGENTS --epic <epic-id> 2>&1
```

**Expected output:**
```
Farm started: N agents
Work source: native tasks (or beads)
Agents spawned in parallel via Task tool
```

### Step 4: Monitor Progress

**Tell the user farm is running, then monitor:**

**For native tasks:**
- Use TaskList periodically to check progress
- Count: pending vs in_progress vs completed
- Farm complete when all tasks are completed

**For beads:**
```
Commands:
  ao inbox              - Check messages
  ao farm status        - Show agent states
  ao farm stop          - Graceful shutdown
```

**Progress check (native tasks):**
Use TaskList and report:
- Pending: X tasks
- In Progress: X tasks (owned by agent-N)
- Completed: X tasks

### Step 5: Handle Completion

**For native tasks:**
1. All agents return "AGENT-N COMPLETE"
2. Verify via TaskList: all tasks status=completed
3. Report: "Farm complete. X tasks completed."

**For beads:**
1. Verify all issues closed:
```bash
bd list --status open 2>/dev/null | wc -l
```

2. Clean up farm resources:
```bash
ao farm stop 2>/dev/null
```

**Next step:**
```
Farm complete. X tasks completed.
Run /post-mortem to extract learnings.
```

## Error Handling

### Circuit Breaker

If >50% agents fail within 60 seconds:
```bash
ao farm stop --reason "circuit-breaker"
```

Tell user: "Circuit breaker triggered. >50% agents failed. Check logs."

### Witness Death

If witness process dies (detected via PID check):
```bash
if ! kill -0 $(cat .witness.pid 2>/dev/null) 2>/dev/null; then
    echo "ERROR: Witness died. Stopping farm."
    ao farm stop --reason "witness-died"
fi
```

### Orphaned Issues

If issues stuck in_progress after farm stop:
```bash
ao farm resume
```

## Key Rules

- **Pre-flight first** - Never spawn without validation
- **Serial spawn** - 30s stagger prevents rate limits
- **Cap agents** - Never more agents than ready issues
- **Monitor witness** - Check PID health every 30s
- **Graceful stop** - Clean up all child processes
- **Resume capability** - Recover from disconnects

## Without ao CLI

If `ao farm` commands not available:

1. **Manual tmux spawn:**
```bash
# Create session
tmux new-session -d -s ao-farm

# For each agent
tmux send-keys -t ao-farm "claude --prompt 'Run /implement on next ready issue'" Enter
sleep 30  # Wait before next agent
```

2. **Manual witness:**
```bash
tmux new-session -d -s ao-farm-witness
tmux send-keys -t ao-farm-witness "claude --prompt 'Monitor tmux session ao-farm, summarize every 5m'" Enter
```

3. **Manual inbox (beads messages):**
```bash
bd list --type message --to mayor 2>/dev/null
```

## Agent Farm Behavior

### Native Tasks Mode

Each spawned agent (via Task tool):
1. Checks TaskList for pending tasks with no blockers
2. Claims task via TaskUpdate (status=in_progress, owner=agent-N)
3. Reads full task via TaskGet
4. Executes the work described in task description
5. Marks complete via TaskUpdate (status=completed)
6. Repeats until no pending tasks remain
7. Returns "AGENT-N COMPLETE"

Mayor monitors via TaskList - no witness needed since Task tool returns results.

### Beads Mode

Each spawned agent (via tmux):
1. Runs `/implement` loop until no ready issues
2. Claims issues atomically via `bd claim`
3. Sends completion message via Agent Mail
4. Exits when no more work

Witness (beads mode only):
1. Polls agent tmux panes every 60s
2. Summarizes to mayor every 5m
3. Escalates blockers immediately
4. Sends "FARM COMPLETE" when all agents idle and no ready issues

## Exit Conditions

Farm exits when:
- All tasks/issues completed (success)
- Circuit breaker triggers (>50% fail in beads mode)
- All agents return COMPLETE (native tasks mode)
- User cancels (manual)
