# Swarm Local Mode: Runtime-Aware Detailed Execution

## Step 3: Select Backend + Spawn Workers

Choose the local backend in this order:

1. `spawn_agent` available -> **Codex experimental sub-agents**
2. `TeamCreate` available -> **Claude native teams**
3. Otherwise -> **Task(run_in_background=true)** fallback

### Codex backend (preferred when available)

For each ready task, pre-assign and spawn a worker:

```
# 1. Pre-assign the task (deterministic, race-free)
TaskUpdate(taskId="<id>", owner="worker-<task-id>", status="in_progress")

# 2. Spawn Codex sub-agent for that task
spawn_agent(
  message="You are worker-<task-id>.
Task #<id>: <subject>
<description>

Rules:
- Work only this task
- Do not commit
- On completion return JSON envelope with status, detail, artifacts
- If blocked, return JSON blocked envelope"
)
```

Track `worker-<task-id> -> <agent-id>` mapping for waits/retries/cleanup.

### Claude teams backend

Create a team for this wave:

```
TeamCreate(team_name="swarm-<epoch>")
```

Team naming: `swarm-<epoch>` (e.g., `swarm-1738857600`). New team per wave = fresh context per spawn (Ralph Wiggum preserved).

**For each ready task, assign it to a worker BEFORE spawning:**

```
# 1. Pre-assign the task to the worker (deterministic, race-free)
TaskUpdate(taskId="<id>", owner="worker-<task-id>", status="in_progress")

# 2. Spawn the worker with the assignment baked into its prompt
Task(
  subagent_type="general-purpose",
  model="opus",
  team_name="swarm-<epoch>",
  name="worker-<task-id>",
  timeout=180000,  # 3 minutes per worker
  prompt="You are Worker on swarm team \"swarm-<epoch>\".

Your Assignment: Task #<id>: <subject>
<description>

Instructions:
1. Execute your pre-assigned task independently — create/edit files as needed, verify your work
2. When done, mark complete: TaskUpdate(taskId=\"<id>\", status=\"completed\")
3. Send completion message to team lead with summary of what you did
4. If blocked, message team lead explaining the issue

When reporting completion or failure, use the structured envelope format.
Your message MUST start with a JSON code block:

On success:
\`\`\`json
{
  \"type\": \"completion\",
  \"issue_id\": \"<task-id>\",
  \"status\": \"done\",
  \"detail\": \"Summary of what was done\",
  \"artifacts\": [\"path/to/changed/file1\", \"path/to/changed/file2\"]
}
\`\`\`

If blocked:
\`\`\`json
{
  \"type\": \"blocked\",
  \"issue_id\": \"<task-id>\",
  \"status\": \"blocked\",
  \"detail\": \"Reason you cannot proceed\"
}
\`\`\`

Rules:
- Work only on YOUR pre-assigned task
- Do NOT call TaskList or claim other tasks
- Do NOT message other workers
- Only update YOUR task status: in_progress → completed
- Do NOT run git add, git commit, or git push. Write your files and report completion via SendMessage. The team lead will commit."
)
```

## Race Condition Prevention

Workers do NOT race-claim tasks from TaskList. The team lead assigns each task
to a specific worker BEFORE spawning. This prevents:
- Two workers claiming the same task
- Workers seeing stale TaskList state
- Non-deterministic assignment order

Workers only transition their assigned task: in_progress -> completed.

## Git Commit Policy

**Workers MUST NOT commit.** The team lead is the sole committer.

| Actor | Git Permissions |
|-------|----------------|
| Team lead (mayor) | git add, commit, push |
| Workers | Read-only git. Write files only. |

**Rationale:**
- Workers share a worktree (per native teams semantics research)
- Concurrent git add/commit from multiple workers corrupts the index
- Lead-only commits ensure atomic, reviewable changesets per wave

**Worker instructions:** Include in every worker prompt:
"Do NOT run git add, git commit, or git push. Write your files and report
completion through your runtime channel (Codex reply or SendMessage). The team lead will commit."

## Step 4: Wait for Completion

Completion channel depends on backend:
- Codex backend: `wait(ids=[...], timeout_ms=<worker-timeout>)`
- Claude teams: workers send completion via `SendMessage`
- Fallback: `TaskOutput(..., block=true)`

For Claude teams specifically:
- Messages arrive automatically as conversation turns (no polling needed)
- Each message includes a summary of what the worker did
- Workers go idle after sending -- this is normal, not an error
- **CRITICAL**: Do NOT mark complete yet - validation required first

## Step 4a: Validate Before Accepting (MANDATORY)

> **TRUST ISSUE**: Agent completion claims are NOT trusted. Verify then trust.

**The Validation Contract**: Before marking any task complete, Mayor MUST run validation checks. See `skills/shared/validation-contract.md` for full specification.

**Validation flow:**

```
<task-notification> arrives
        |
        v
    RUN VALIDATION
        |
    +---+---+
    |       |
  PASS    FAIL
    |       |
    v       v
 complete  retry/escalate
```

**For each completed task notification:**

1. **Check task metadata for validation requirements:**
   ```
   TaskList() -> find task -> check metadata.validation
   ```

2. **Execute validation checks (in order):**

   | Check Type | Command | Pass Condition |
   |------------|---------|----------------|
   | `files_exist` | `ls -la <paths>` | All files exist |
   | `command` | Run specified command | Exit code 0 |
   | `content_check` | `grep <pattern> <file>` | Pattern found |
   | `tests` | `<test_command>` | Tests pass |
   | `lint` | `<lint_command>` | No errors |

3. **On validation PASS:**
   ```
   TaskUpdate(taskId="<id>", status="completed")
   ```

4. **On validation FAIL:**
   - Increment retry count for task
   - If retries < MAX_RETRIES (default: 3):
     ```
     # Codex backend: send follow-up to existing sub-agent
     send_input(
       id="<agent-id-for-task>",
       message="Validation failed: <specific failure>. Fix and retry."
     )

     # Claude teams backend:
     SendMessage(type="message", recipient="worker-<task-id>", content="Validation failed: <specific failure>. Fix and retry.", summary="Retry: validation failed")
     ```
   - If retries >= MAX_RETRIES:
     ```
     TaskUpdate(taskId="<id>", status="blocked")
     # Escalate to user
     ```

**Minimal validation (when no metadata):**

If task has no explicit validation requirements, apply default checks:

```bash
# Check that worker wrote the expected files
git status --porcelain  # Should show unstaged changes from worker

# Check for obvious failures in recent output
# (agent should not have ended with errors)
```

**Example task with validation metadata:**

```
TaskCreate(
  subject="Add authentication middleware",
  description="...",
  metadata={
    "validation": {
      "files_exist": ["src/middleware/auth.py", "tests/test_auth.py"],
      "command": "pytest tests/test_auth.py -v",
      "content_check": {"file": "src/middleware/auth.py", "pattern": "def authenticate"}
    }
  }
)
```

## Step 5: Review & Finalize

When workers complete AND pass validation:
1. Check git status for changes (workers wrote files but did not commit)
2. Review diffs
3. Run any additional tests/validation
4. Team lead commits all changes for the wave (sole committer)

## Step 5a: Cleanup

After wave completes:
```
# Codex backend: close all spawned workers
close_agent(id="<agent-id-1>")
close_agent(id="<agent-id-2>")

# Claude teams backend:
SendMessage(type="shutdown_request", recipient="worker-<task-id>", content="Wave complete")
TeamDelete()
```

> **Note:** `TeamDelete()` applies only to Claude team backend. Codex backend requires explicit `close_agent()` for each spawned worker.

### Reaper Cleanup Pattern

Team cleanup MUST succeed even on partial failures. Follow this sequence:

1. **Attempt graceful shutdown/close:** `close_agent` (Codex) or `shutdown_request` (Claude)
2. **Wait up to 30s** for shutdown_approved responses
3. **If any worker doesn't respond:** Log warning, proceed anyway
4. **Always cleanup backend resources** (`close_agent` or `TeamDelete`)
5. **Cleanup must run even on partial failures**

**Failure modes and recovery:**

| Failure | Behavior |
|---------|----------|
| Worker hangs (no response) | 30s timeout -> proceed with cleanup |
| close/shutdown fails | Log warning -> continue cleanup |
| TeamDelete fails | Log error -> team orphaned (manual cleanup: delete ~/.claude/teams/<name>/) |
| close_agent fails | Log error -> leaked sub-agent handle (best effort close later) |
| Lead crashes mid-swarm | Cleanup may be deferred to session end |

**Never skip cleanup.** Lingering workers/teams pollute future sessions.

### Team Timeout Configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| Worker timeout | 180s | Max time for worker to complete its task |
| Shutdown grace period | 30s | Time to wait for shutdown_approved |
| Wave timeout | 600s | Max time for entire wave before forced cleanup |

## Step 6: Repeat if Needed

If more tasks remain:
1. Check TaskList for next wave
2. Spawn a NEW wave worker set (new sub-agents or new team) for fresh context
3. Execute the next wave
4. Continue until all done

## Partial Completion

**Worker timeout:** 180s (3 minutes) default per worker.

**Timeout behavior:**
1. Log warning: "Worker {name} timed out on task {id}"
2. Mark task as failed with reason "timeout"
3. Add to retry queue for next wave
4. Continue with remaining workers

**Quorum:** Swarm does not require quorum -- each worker is independent.
Each completed task is accepted individually.
