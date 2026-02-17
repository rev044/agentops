# Swarm Local Mode: Runtime-Aware Detailed Execution

## Context Budget Rule

> **Workers write results to disk. The orchestrator reads only thin status files.**
>
> When N workers finish, their full output (file reads, tool calls, reasoning) must NOT flood back into the orchestrator context. This is the #1 cause of context explosion in multi-wave epics.

**Result protocol:**
1. Workers write `.agents/swarm/results/<task-id>.json` on completion
2. Orchestrator checks for result files (Glob/Read), NOT full Task/SendMessage output
3. SendMessage used only for coordination signals (blocked, need help) — kept under 100 tokens
4. Task tool return values are acknowledged but NOT parsed for work details

```bash
# Orchestrator creates result directory before spawning
mkdir -p .agents/swarm/results
```

## Step 2b: Pre-Spawn Worktree Setup (Multi-Epic Waves)

> **Skip this step** for single-epic waves or when `--no-worktrees` is set.
> **Required** for multi-epic dispatch or when `--worktrees` is set.

Evidence: shared-worktree multi-epic dispatch produced build breaks and algorithm duplication (`.agents/evolve/dispatch-comparison.md`).

### Detection

```bash
# Multi-epic: check if tasks span more than one epic prefix
# If wave tasks have subjects like "[ol-527] ..." and "[ol-531] ...", use worktrees.
# Single-epic: tasks share one prefix (e.g., all ol-527.*) → shared worktree OK.
```

### Create Worktrees

```bash
# For each epic ID in the wave:
git worktree add /tmp/swarm-<epic-id> -b swarm/<epic-id>
```

Track the mapping:
```
epic_worktrees = {
  "<epic-id>": "/tmp/swarm-<epic-id>",
  ...
}
```

### Inject into Worker Prompts

Each worker prompt must include:
```
WORKING DIRECTORY: /tmp/swarm-<epic-id>

All file reads, writes, and edits MUST use absolute paths rooted at /tmp/swarm-<epic-id>.
Do NOT operate on the main repo directly.
Result file: write to <main-repo>/.agents/swarm/results/<task-id>.json (always main repo path).
```

### Merge-Back After Validation

After each worker's task passes validation:
```bash
# From main repo:
git merge --no-ff swarm/<epic-id> -m "chore: merge swarm/<epic-id>"
git worktree remove /tmp/swarm-<epic-id>
git branch -d swarm/<epic-id>
```

Merge order must respect task blockedBy dependencies.

---

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
- On completion, write your result to .agents/swarm/results/<task-id>.json:
  {\"type\":\"completion\",\"issue_id\":\"<task-id>\",\"status\":\"done\",\"detail\":\"<one-line summary>\",\"artifacts\":[\"file1\",\"file2\"]}
- If blocked, write:
  {\"type\":\"blocked\",\"issue_id\":\"<task-id>\",\"status\":\"blocked\",\"detail\":\"<reason>\"}
- Keep your final response under 50 tokens — the result file IS your report"
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
2. Write your result to .agents/swarm/results/<task-id>.json (see format below)
3. Mark complete: TaskUpdate(taskId=\"<id>\", status=\"completed\")
4. Send a SHORT completion signal to team lead (under 100 tokens)
5. If blocked, write blocked result to same path and message team lead

RESULT FILE FORMAT (MANDATORY — write this BEFORE sending any message):

On success, write to .agents/swarm/results/<task-id>.json:
{\"type\":\"completion\",\"issue_id\":\"<task-id>\",\"status\":\"done\",\"detail\":\"<one-line summary max 100 chars>\",\"artifacts\":[\"path/to/file1\",\"path/to/file2\"]}

If blocked, write to .agents/swarm/results/<task-id>.json:
{\"type\":\"blocked\",\"issue_id\":\"<task-id>\",\"status\":\"blocked\",\"detail\":\"<reason max 200 chars>\"}

CONTEXT BUDGET RULE:
Your SendMessage to the team lead must be under 100 tokens.
Do NOT include file contents, diffs, or detailed explanations in messages.
The result JSON file IS your full report. The team lead reads the file, not your message.

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
"Do NOT run git add, git commit, or git push. Write your result to .agents/swarm/results/<task-id>.json, then send a short signal (under 100 tokens) via your runtime channel. The team lead reads result files, not messages."

## Step 4: Wait for Completion

**Completion signals** (how you know workers are done):
- Codex backend: `wait(ids=[...], timeout_ms=<worker-timeout>)`
- Claude teams: workers send short completion signal via `SendMessage`
- Fallback: `TaskOutput(..., block=true)` — acknowledge return, do NOT parse content

**Result data** (how you get work details — ALL backends):
```bash
# Read result files written by workers (thin JSON, ~200 bytes each)
for task_id in <wave-task-ids>; do
    if [ -f ".agents/swarm/results/${task_id}.json" ]; then
        cat ".agents/swarm/results/${task_id}.json"
    else
        echo "WARNING: No result file for ${task_id}"
    fi
done
```

**Why disk-based results:** Task tool returns and SendMessage content include the worker's FULL conversation (file reads, tool calls, reasoning). For 6 workers, that's 6 × 5-20K tokens flooding the orchestrator context. Result files are ~200 bytes each — 6 × 200 bytes = 1.2KB total.

For Claude teams specifically:
- Short messages arrive as conversation turns (under 100 tokens each)
- Workers go idle after sending — this is normal, not an error
- **CRITICAL**: Do NOT mark complete yet - validation required first
- **CRITICAL**: Do NOT parse the SendMessage content for work details — read the result file instead

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
```bash
# Clean up result files from this wave (prevent stale reads in next wave)
rm -f .agents/swarm/results/*.json
```

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
