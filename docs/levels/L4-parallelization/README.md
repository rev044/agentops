# L4 — Parallelization

Execute independent tasks in parallel with wave-based execution using the swarm pattern.

## What You'll Learn

- Identifying independent (unblocked) work via TaskList
- Using `/swarm` for parallel multi-agent execution
- Wave-based dependency resolution
- The Ralph Wiggum pattern for fresh context

## Prerequisites

- Completed L3-state-management
- Understanding of task dependencies (blockedBy)
- Comfortable with TaskCreate/TaskUpdate/TaskList

## Available Commands

| Command | Purpose |
|---------|---------|
| `/swarm` | Execute unblocked tasks in parallel via background agents |
| `/plan <goal>` | Same as L3 |
| `/research <topic>` | Same as L2 |
| `/implement [id]` | Execute single task |
| `/retro [topic]` | Same as L2 |

## Key Concepts

- **Wave**: Set of independent tasks executed together
- **Background agents**: Each task spawned via `Task(run_in_background=true)`
- **Fresh context**: Each agent spawn = clean slate (Ralph Wiggum pattern)
- **Dependency resolution**: Only unblocked tasks run in each wave

## The Ralph Wiggum Pattern

The swarm follows Ralph Wiggum's core insight: fresh context per iteration.

```
Ralph's loop:               Swarm equivalent:
while :; do                 Mayor identifies ready tasks
  cat PROMPT.md | claude    Mayor spawns background agents
done                        Agents complete, Mayor gets notified
                            Repeat for next wave
```

Why this matters:
- **Internal loops accumulate context** → degrades over iterations
- **Fresh spawns stay effective** → each agent is a clean slate

## Wave Workflow

```
1. TaskList → identifies unblocked tasks
2. /swarm → spawns background agents for wave
3. Agents complete work → <task-notification> arrives automatically
4. Mayor reconciliation:
   a. Verify work (check files/git)
   b. TaskUpdate(status="completed") for each
   c. TaskList to find newly unblocked tasks
5. Next wave begins
```

## Mayor Reconciliation Step

After notifications arrive, Mayor must verify before marking complete:

```
# For each completed agent:
1. Check the files created/modified
2. Run tests (npm test, pytest, etc.)
3. Run lint (npm run lint, etc.)
4. If valid: TaskUpdate(taskId="N", status="completed")
5. If invalid: Note issues, may need re-run

# After all verified:
TaskList() → shows newly unblocked tasks → ready for next wave
```

This prevents marking broken work as complete.

## Agent Prompts (Atomic)

Each spawned agent gets a simple, single-task prompt:

```
# Good (atomic):
"Create users endpoint in src/routes/users.ts. Include GET /users,
POST /users, GET /users/:id routes. Follow existing patterns."

# Bad (complex loop):
"Create users endpoint, then test it, then if tests fail fix them,
then validate, then update status, then check for more work..."
```

Agents do ONE thing. Mayor handles orchestration.

## Example Session

```
1. /plan "Build auth system"
   → Creates tasks with dependencies:
   #1 [pending] Create User model
   #2 [pending] Add password hashing (blockedBy: #1)
   #3 [pending] Create login endpoint (blockedBy: #1)
   #4 [pending] Write tests (blockedBy: #2, #3)

2. /swarm
   → Wave 1: Spawns agent for #1 (only unblocked)
   → Agent completes, Mayor marks #1 completed

3. /swarm
   → Wave 2: Spawns agents for #2 and #3 in parallel
   → Both complete

4. /swarm
   → Wave 3: Spawns agent for #4
   → All done

5. /vibe → Validate everything
```

## What's NOT at This Level

- No `/crank` (full autonomous execution without human wave triggers)
- Human triggers each wave

## Next Level

Once comfortable with waves, progress to [L5-orchestration](../L5-orchestration/) for full autonomy with `/crank`.
