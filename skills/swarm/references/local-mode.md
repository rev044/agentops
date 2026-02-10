# Swarm Local Mode: Detailed Execution

## Step 3: Create Team + Spawn Workers

**Create a team for this wave:**

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
1. Execute your pre-assigned task autonomously — create/edit files as needed, verify your work
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
completion via SendMessage. The team lead will commit."

## Step 4: Wait for Completion Messages

Workers send completion messages to the team lead via `SendMessage`:
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
     # Send retry instructions to existing worker (wakes from idle with full context)
     SendMessage(
       type="message",
       recipient="worker-<task-id>",
       content="Validation failed: <specific failure>. Fix the issue and try again.",
       summary="Retry: validation failed"
     )
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

## Step 5a: Cleanup Team

After wave completes:
```
# Shutdown each worker
SendMessage(type="shutdown_request", recipient="worker-<task-id>", content="Wave complete")

# Delete team for this wave
TeamDelete()
```

> **Note:** `TeamDelete()` deletes the team associated with this session's `TeamCreate()` call. If running concurrent teams (e.g., council inside crank), each team is cleaned up in the session that created it.

### Reaper Cleanup Pattern

Team cleanup MUST succeed even on partial failures. Follow this sequence:

1. **Attempt graceful shutdown:** Send shutdown_request to each worker
2. **Wait up to 30s** for shutdown_approved responses
3. **If any worker doesn't respond:** Log warning, proceed anyway
4. **Always call TeamDelete()** -- even if some workers are unresponsive
5. **TeamDelete cleans up** the team regardless of member state

**Failure modes and recovery:**

| Failure | Behavior |
|---------|----------|
| Worker hangs (no response) | 30s timeout -> proceed to TeamDelete |
| shutdown_request fails | Log warning -> proceed to TeamDelete |
| TeamDelete fails | Log error -> team orphaned (manual cleanup: delete ~/.claude/teams/<name>/) |
| Lead crashes mid-swarm | Team orphaned until session ends or manual cleanup |

**Never skip TeamDelete.** A lingering team config pollutes future sessions.

### Team Timeout Configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| Worker timeout | 180s | Max time for worker to complete its task |
| Shutdown grace period | 30s | Time to wait for shutdown_approved |
| Wave timeout | 600s | Max time for entire wave before forced cleanup |

## Step 6: Repeat if Needed

If more tasks remain:
1. Check TaskList for next wave
2. Create NEW team (`TeamCreate` with new epoch) -- fresh context per wave
3. Spawn new workers as teammates
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
