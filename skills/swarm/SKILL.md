---
name: swarm
description: 'Spawn isolated agents for parallel task execution. Local mode: Task tool (pure Claude-native). Distributed mode: tmux + Agent Mail (process isolation, persistence). Triggers: "swarm", "spawn agents", "parallel work".'
dependencies:
  - implement # required - executes `/implement <bead-id>` in distributed mode
  - vibe      # optional - integration with validation
---

# Swarm Skill

Spawn isolated agents to execute tasks in parallel. Fresh context per agent (Ralph Wiggum pattern).

**Execution Modes:**
- **Local** (default) - Pure Claude-native using Task tool background agents
- **Distributed** (`--mode=distributed`) - tmux sessions + Agent Mail for robust coordination

**Integration modes:**
- **Direct** - Create TaskList tasks, invoke `/swarm`
- **Via Crank** - `/crank` creates tasks from beads, invokes `/swarm` for each wave

## Architecture (Mayor-First)

```
Mayor (this session)
    |
    +-> Plan: TaskCreate with dependencies
    |
    +-> Identify wave: tasks with no blockers
    |
    +-> Create team: TeamCreate(team_name="swarm-<epoch>")
    |
    +-> Assign: TaskUpdate(taskId, owner="worker-<id>", status="in_progress")
    |
    +-> Spawn: Task(team_name=..., name="worker-<id>") for each
    |       Workers receive pre-assigned task, execute atomically
    |
    +-> Wait: Workers send completion via SendMessage
    |
    +-> Validate: Review changes when complete
    |
    +-> Cleanup: shutdown_request workers, TeamDelete()
    |
    +-> Repeat: New team + new plan if more work needed
```

## Execution

Given `/swarm`:

### Step 1: Ensure Tasks Exist

Use TaskList to see current tasks. If none, create them:

```
TaskCreate(subject="Implement feature X", description="Full details...")
TaskUpdate(taskId="2", addBlockedBy=["1"])  # Add dependencies after creation
```

### Step 2: Identify Wave

Find tasks that are:
- Status: `pending`
- No blockedBy (or all blockers completed)

These can run in parallel.

### Step 3: Create Team + Spawn Workers

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

### Race Condition Prevention

Workers do NOT race-claim tasks from TaskList. The team lead assigns each task
to a specific worker BEFORE spawning. This prevents:
- Two workers claiming the same task
- Workers seeing stale TaskList state
- Non-deterministic assignment order

Workers only transition their assigned task: in_progress → completed.

### Git Commit Policy

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

### Step 4: Wait for Completion Messages

Workers send completion messages to the team lead via `SendMessage`:
- Messages arrive automatically as conversation turns (no polling needed)
- Each message includes a summary of what the worker did
- Workers go idle after sending — this is normal, not an error
- **CRITICAL**: Do NOT mark complete yet - validation required first

### Step 4a: Validate Before Accepting (MANDATORY)

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
   TaskList() → find task → check metadata.validation
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

### Step 5: Review & Finalize

When workers complete AND pass validation:
1. Check git status for changes (workers wrote files but did not commit)
2. Review diffs
3. Run any additional tests/validation
4. Team lead commits all changes for the wave (sole committer)

### Step 5a: Cleanup Team

After wave completes:
```
# Shutdown each worker
SendMessage(type="shutdown_request", recipient="worker-<task-id>", content="Wave complete")

# Delete team for this wave
TeamDelete()
```

> **Note:** `TeamDelete()` deletes the team associated with this session's `TeamCreate()` call. If running concurrent teams (e.g., council inside crank), each team is cleaned up in the session that created it.

#### Reaper Cleanup Pattern

Team cleanup MUST succeed even on partial failures. Follow this sequence:

1. **Attempt graceful shutdown:** Send shutdown_request to each worker
2. **Wait up to 30s** for shutdown_approved responses
3. **If any worker doesn't respond:** Log warning, proceed anyway
4. **Always call TeamDelete()** — even if some workers are unresponsive
5. **TeamDelete cleans up** the team regardless of member state

**Failure modes and recovery:**

| Failure | Behavior |
|---------|----------|
| Worker hangs (no response) | 30s timeout → proceed to TeamDelete |
| shutdown_request fails | Log warning → proceed to TeamDelete |
| TeamDelete fails | Log error → team orphaned (manual cleanup: delete ~/.claude/teams/<name>/) |
| Lead crashes mid-swarm | Team orphaned until session ends or manual cleanup |

**Never skip TeamDelete.** A lingering team config pollutes future sessions.

#### Team Timeout Configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| Worker timeout | 180s | Max time for worker to complete its task |
| Shutdown grace period | 30s | Time to wait for shutdown_approved |
| Wave timeout | 600s | Max time for entire wave before forced cleanup |

### Step 6: Repeat if Needed

If more tasks remain:
1. Check TaskList for next wave
2. Create NEW team (`TeamCreate` with new epoch) — fresh context per wave
3. Spawn new workers as teammates
4. Continue until all done

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

### Partial Completion

**Worker timeout:** 180s (3 minutes) default per worker.

**Timeout behavior:**
1. Log warning: "Worker {name} timed out on task {id}"
2. Mark task as failed with reason "timeout"
3. Add to retry queue for next wave
4. Continue with remaining workers

**Quorum:** Swarm does not require quorum — each worker is independent.
Each completed task is accepted individually.

## Key Points

- **Pure Claude-native** - No tmux, no external scripts
- **Native teams** - `TeamCreate` + `Task(team_name=...)` + `SendMessage` for coordination
- **Pre-assigned tasks** - Mayor assigns tasks before spawning; workers never race-claim
- **Team per wave** - Fresh team = fresh context (Ralph Wiggum preserved)
- **Wave execution** - Only unblocked tasks spawn
- **Mayor orchestrates** - You control the flow, workers report via `SendMessage`
- **Retry via message** - Send retry instructions to idle workers (no re-spawn needed)
- **Atomic execution** - Each worker works until task done
- **Graceful fallback** - If `TeamCreate` unavailable, fall back to `Task(run_in_background=true)`

## Integration with AgentOps

This ties into the full workflow:

```
/research → Understand the problem
/plan → Decompose into beads issues
/crank → Autonomous epic loop
    └── /swarm → Execute each wave in parallel
/vibe → Validate results
/post-mortem → Extract learnings
```

**Direct use (no beads):**
```
TaskCreate → Define tasks
/swarm → Execute in parallel
```

The knowledge flywheel captures learnings from each agent.

## Task Management Commands

```
# List all tasks
TaskList()

# Mark task complete after notification
TaskUpdate(taskId="1", status="completed")

# Add dependency between tasks
TaskUpdate(taskId="2", addBlockedBy=["1"])
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--mode=local\|distributed` | Execution mode | `local` |
| `--max-workers=N` | Max concurrent workers | 5 |
| `--from-wave <json-file>` | Load wave from OL hero hunt output (see OL Wave Integration) | - |
| `--bead-ids` | Specific beads to work (comma-separated, distributed mode) | Auto from `bd ready` |
| `--wait` | Wait for all workers to complete (distributed mode) | false |
| `--timeout` | Max time to wait if `--wait` (distributed mode) | 30m |

## When to Use Swarm

| Scenario | Use |
|----------|-----|
| Multiple independent tasks | `/swarm` (parallel) |
| Sequential dependencies | `/swarm` with blockedBy |
| Mix of both | `/swarm` spawns waves, each wave parallel |

## Why This Works: Ralph Wiggum Pattern

This architecture follows the [Ralph Wiggum Pattern](https://ghuntley.com/ralph/) for autonomous agents.

**Core Insight:** Each wave creates a new team = fresh context per spawn.

```
Ralph's bash loop:          Our swarm:
while :; do                 Wave 1: TeamCreate → spawn workers → fresh context each
  cat PROMPT.md | claude    Wave 2: TeamCreate → spawn workers → fresh context each
done                        Wave 3: TeamCreate → spawn workers → fresh context each
```

Both achieve the same thing: **fresh context per execution unit**.

### Team-Per-Wave = Ralph Wiggum

Native teams don't break Ralph — they enhance it:
- **New team per wave** = fresh context per spawn (workers don't persist across waves)
- **Workers are atomic** = one task, one spawn, one result
- **Team lead IS the loop** = orchestration layer, manages state across waves
- **Within a wave**, workers can retry via `SendMessage` (wake from idle with same context) — this is a feature, not a violation, since the retry happens within the same atomic unit of work

### Why Fresh Context Matters

| Approach | Context | Problem |
|----------|---------|---------|
| Internal loop in agent | Accumulates | Degrades over iterations |
| Mayor spawns team per wave | Fresh each time | Stays effective at scale |

Making workers loop internally would violate Ralph — context accumulates within the session. The loop belongs in Mayor (lightweight orchestration), fresh context belongs in workers (heavyweight work).

### Key Properties

- **Mayor IS the loop** - Orchestration layer, manages state
- **Workers are atomic** - One task, one spawn, one result
- **Team per wave** - `TeamCreate` → work → `TeamDelete` → repeat
- **TaskList as memory** - State persists in task status, not context
- **Filesystem for artifacts** - Files written by workers, committed by team lead
- **SendMessage for coordination** - Workers report to team lead, never to each other

This is **Ralph + parallelism + coordination**: the while loop is distributed across wave spawns, with multiple agents per wave communicating through the team lead.

## Integration with Crank

When `/crank` invokes `/swarm`:

1. **Crank** bridges beads issues to TaskList tasks
2. **Swarm** executes the TaskList wave with fresh-context agents
3. **Crank** syncs results back to beads

```
/crank epic-123
  └── bd ready → [ao-1, ao-2, ao-3]
      └── TaskCreate for each issue
          └── /swarm
              └── Spawn agents (fresh context each)
                  └── Complete TaskList tasks
      └── bd update --status closed
  └── Loop until epic DONE
```

This gives you:
- **Beads orchestration** (crank) - Epic lifecycle, issue tracking
- **Fresh-context execution** (swarm) - Ralph pattern for each issue

## Distinctions (Common Confusion)

| You Want | Use | Why |
|----------|-----|-----|
| Fresh-context parallel execution | `/swarm` | Each spawned background agent is a clean slate |
| Autonomous epic loop | `/crank` | Loops waves via swarm until epic closes |
| Just swarm, no beads | `/swarm` directly | Use TaskList only, skip beads integration |
| RPI progress gates | `/ratchet` | Tracks/locks progress; does not execute work by itself |

---

## Distributed Mode: tmux + Agent Mail (Experimental)

> **Status: Experimental.** Local mode (native teams) is the recommended execution method. Distributed mode requires Agent Mail and tmux and has not been battle-tested. Use for long-running epics where process isolation and persistence are critical.

> **When:** MCP Agent Mail is available AND you want true process isolation, persistent workers, and robust coordination.

Distributed mode spawns real tmux sessions instead of Task tool background agents. Each demigod runs in its own Claude process with full lifecycle management.

### Why Distributed Mode?

| Local (Task tool) | Distributed (tmux + Agent Mail) |
|---------------------|----------------------------|
| Background agents in Mayor's process | Separate tmux sessions |
| Coupled to Mayor lifecycle | Persistent if Mayor crashes |
| No inter-agent coordination | Agent Mail messaging |
| No file conflict prevention | File reservations |
| Simple, fast to spawn | More setup, more robust |
| Good for small jobs | Better for large/long jobs |

### Mode Detection

At skill start, detect which mode to use:

```bash
# Method 1: Explicit flag
# /swarm --mode=distributed <tasks>

# Method 2: Auto-detect Agent Mail availability
MODE="local"

# Check for Agent Mail MCP tools (look for register_agent tool)
if mcp-tools 2>/dev/null | grep -q "mcp-agent-mail"; then
    AGENT_MAIL_AVAILABLE=true
fi

# Check for Agent Mail HTTP endpoint
if curl -s http://localhost:8765/health >/dev/null 2>&1; then
    AGENT_MAIL_HTTP=true
fi

# Distributed requires: Agent Mail available + explicit flag
if [ "$MODE_FLAG" = "distributed" ] && [ "$AGENT_MAIL_AVAILABLE" = "true" -o "$AGENT_MAIL_HTTP" = "true" ]; then
    MODE="distributed"
fi
```

**Decision matrix:**

| `--mode` | Agent Mail | Result |
|----------|------------|--------|
| Not set | Not available | Local |
| Not set | Available | Local (explicit opt-in required) |
| `--mode=local` | Any | Local |
| `--mode=distributed` | Not available | **Error: Agent Mail required** |
| `--mode=distributed` | Available | Distributed |

### Distributed Mode Invocation

```
/swarm --mode=distributed [--max-workers=N]
/swarm --mode=distributed --bead-ids ol-527.1,ol-527.2,ol-527.3
```

**Parameters:**

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--mode=distributed` | Enable tmux + Agent Mail mode | - |
| `--max-workers=N` | Max concurrent demigods | 5 |
| `--bead-ids` | Specific beads to work (comma-separated) | Auto from `bd ready` |
| `--wait` | Wait for all demigods to complete | false |
| `--timeout` | Max time to wait (if --wait) | 30m |

### Distributed Mode Architecture

```
Mayor Session (this session)
    |
    +-> Mode Detection: `--mode=distributed` + Agent Mail available → Distributed
    |
    +-> Identify wave: bd ready → [ol-527.1, ol-527.2, ol-527.3]
    |
    +-> Spawn: tmux new-session for each worker
    |       demigod-ol-527-1 → runs `/implement ol-527.1 --mode=distributed`
    |       demigod-ol-527-2 → runs `/implement ol-527.2 --mode=distributed`
    |       demigod-ol-527-3 → runs `/implement ol-527.3 --mode=distributed`
    |
    +-> Coordinate: Agent Mail messages
    |       Each demigod sends ACCEPTED, PROGRESS, DONE/FAILED
    |       Mayor monitors via fetch_inbox
    |       File reservations prevent conflicts
    |
    +-> Validate: On DONE, optionally run /vibe --remote
    |
    +-> Repeat: New wave when workers complete
```

### Distributed Mode Execution Steps

Given `/swarm --mode=distributed`:

#### Step 1: Pre-flight Checks

```bash
# Check tmux is available
which tmux >/dev/null 2>&1 || {
    echo "Error: tmux required for distributed mode. Install: brew install tmux"
    exit 1
}

# Check claude CLI is available
which claude >/dev/null 2>&1 || {
    echo "Error: claude CLI required for distributed mode"
    exit 1
}

# Check Agent Mail is available
AGENT_MAIL_OK=false
if curl -s http://localhost:8765/health >/dev/null 2>&1; then
    AGENT_MAIL_OK=true
fi
# OR check MCP tools

if [ "$AGENT_MAIL_OK" != "true" ]; then
    echo "Error: Agent Mail required for distributed mode"
    echo "Start your Agent Mail MCP server (implementation-specific). See docs/agent-mail.md."
    exit 1
fi
```

#### Step 2: Register Mayor with Agent Mail

Register the Mayor session to receive messages from demigods.

```
MCP Tool: register_agent
Parameters:
  project_key: <absolute path to project>
  program: "claude-code"
  model: "opus"
  task_description: "Mayor orchestrating swarm for wave"
```

**Store the returned agent name as `MAYOR_NAME`.**

#### Step 3: Identify Wave (Ready Beads)

Same as local mode, get the beads to work:

```bash
# Get ready beads
READY_BEADS=$(bd ready --json 2>/dev/null | jq -r '.[].id' | head -$MAX_WORKERS)

# Or use explicit bead list
if [ -n "$BEAD_IDS" ]; then
    READY_BEADS=$(echo "$BEAD_IDS" | tr ',' '\n')
fi

# Count
WAVE_SIZE=$(echo "$READY_BEADS" | wc -l | tr -d ' ')
```

If no ready beads, exit with message.

#### Step 4: Spawn Demigods via tmux

For each ready bead, spawn a demigod session:

```bash
for BEAD_ID in $READY_BEADS; do
    # Generate session name
    SESSION_NAME="demigod-$(echo $BEAD_ID | tr '.' '-')"

    # Check for existing session
    if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
        echo "Session $SESSION_NAME already exists, skipping"
        continue
    fi

    # Spawn demigod in new tmux session
    tmux new-session -d -s "$SESSION_NAME" "claude -p '/implement $BEAD_ID --mode=distributed --thread-id $BEAD_ID'"

    # Verify session started
    if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
        echo "Spawned: $SESSION_NAME for $BEAD_ID"
        SPAWNED_COUNT=$((SPAWNED_COUNT + 1))
    else
        echo "Failed to spawn: $SESSION_NAME"
    fi

    # Rate limit to avoid API cascade
    sleep 2
done
```

**Key points:**
- Use `-d` for detached sessions (run in background)
- Session names derived from bead IDs for easy correlation
- Rate limit spawns (2 seconds) to avoid API rate limits
- Verify each session started before continuing

#### Step 5: Monitor via Agent Mail

Poll Agent Mail inbox for demigod messages:

```
MCP Tool: fetch_inbox
Parameters:
  project_key: <project path>
  agent_name: <MAYOR_NAME>
  limit: 50
  include_bodies: true
```

**Process messages by type:**

| Subject Pattern | Action |
|-----------------|--------|
| `[<bead-id>] ACCEPTED` | Log demigod started work |
| `[<bead-id>] PROGRESS` | Log progress, check health |
| `[<bead-id>] HELP_REQUEST` | Alert for manual intervention or route to Chiron |
| `[<bead-id>] DONE` | Mark complete, check if wave finished |
| `[<bead-id>] FAILED` | Log failure, decide on retry or escalate |

#### Step 6: Track Completion

Maintain completion state:

```bash
# Track per-bead status
declare -A BEAD_STATUS
for BEAD_ID in $READY_BEADS; do
    BEAD_STATUS[$BEAD_ID]="spawned"
done

# Update on DONE/FAILED messages
# BEAD_STATUS[$BEAD_ID]="done" or "failed"

# Check if wave complete
DONE_COUNT=0
FAILED_COUNT=0
for BEAD_ID in "${!BEAD_STATUS[@]}"; do
    case "${BEAD_STATUS[$BEAD_ID]}" in
        done) DONE_COUNT=$((DONE_COUNT + 1)) ;;
        failed) FAILED_COUNT=$((FAILED_COUNT + 1)) ;;
    esac
done

if [ $((DONE_COUNT + FAILED_COUNT)) -eq $WAVE_SIZE ]; then
    WAVE_COMPLETE=true
fi
```

#### Step 7: Report Results

When wave completes (or on `--wait` timeout):

```markdown
## Swarm Distributed Mode Results

**Wave completed:** <timestamp>
**Beads in wave:** <WAVE_SIZE>
**Successful:** <DONE_COUNT>
**Failed:** <FAILED_COUNT>

### Completed Beads
| Bead ID | Demigod | Commit | Summary |
|---------|---------|--------|---------|
| ol-527.1 | GreenCastle | abc123 | Added auth middleware |
| ol-527.2 | BlueMountain | def456 | Fixed rate limiting |

### Failed Beads
| Bead ID | Demigod | Reason | Recommendation |
|---------|---------|--------|----------------|
| ol-527.3 | RedValley | Tests failed | Re-run with spec clarification |

### Active Sessions
| Session | Bead | Status | Runtime |
|---------|------|--------|---------|
| demigod-ol-527-1 | ol-527.1 | done | 15m |
| demigod-ol-527-2 | ol-527.2 | done | 12m |
| demigod-ol-527-3 | ol-527.3 | failed | 18m |
```

#### Step 8: Cleanup Completed Sessions

Optionally clean up tmux sessions for completed beads:

```bash
# Clean up done sessions (keep failed for debugging)
for BEAD_ID in "${!BEAD_STATUS[@]}"; do
    if [ "${BEAD_STATUS[$BEAD_ID]}" = "done" ]; then
        SESSION_NAME="demigod-$(echo $BEAD_ID | tr '.' '-')"
        tmux kill-session -t "$SESSION_NAME" 2>/dev/null
    fi
done
```

**Or keep all sessions for review:** Use `--keep-sessions` flag to preserve all tmux sessions for post-mortem analysis.

### Distributed Mode Helpers

Use these helpers with distributed mode swarm:

| Helper | Purpose |
|--------|---------|
| `tmux list-sessions` | List running worker sessions |
| `tmux attach -t <session>` | Attach to a worker session for debugging |
| `/inbox` | Check Agent Mail for pending messages |
| `/vibe --remote <session>` | Validate a worker’s work before accepting |

### Distributed Mode File Reservations

File reservations prevent conflicts when multiple demigods edit files.

**How it works:**
1. Each demigod claims files before editing (via Agent Mail `file_reservation_paths`)
2. If another demigod tries to claim the same file, it sees a conflict warning
3. Demigods release reservations when done

**Mayor can view reservations:**

```
MCP Tool: get_file_reservations (if available)
Parameters:
  project_key: <project path>
```

**On conflict:**
- Demigod sends PROGRESS message noting the conflict
- Mayor decides: wait, reassign, or allow parallel work
- Advisory reservations don't block, just warn

### Distributed Mode Error Handling

#### Demigod Session Crashes

If a tmux session dies unexpectedly:

```bash
# Check session health
tmux has-session -t "$SESSION_NAME" 2>/dev/null || {
    echo "Session $SESSION_NAME died"

    # Check bead status
    STATUS=$(bd show $BEAD_ID --json 2>/dev/null | jq -r '.status')

    if [ "$STATUS" = "in_progress" ]; then
        # Unclaim bead for retry
        bd update $BEAD_ID --status open --assignee "" 2>/dev/null
        echo "Bead $BEAD_ID released for retry"
    fi
}
```

#### Agent Mail Server Crashes

If Agent Mail becomes unavailable mid-swarm:

1. Demigods continue working (graceful degradation)
2. Messages queue locally (if supported)
3. Mayor loses visibility but work continues
4. On restart, poll beads status directly:

```bash
for BEAD_ID in $READY_BEADS; do
    STATUS=$(bd show $BEAD_ID --json 2>/dev/null | jq -r '.status')
    echo "$BEAD_ID: $STATUS"
done
```

#### Timeout Handling

If `--wait` times out:

```markdown
## Swarm Timeout

Wave did not complete within <timeout>.

### Still Running
| Session | Bead | Runtime | Action |
|---------|------|---------|--------|
| demigod-ol-527-3 | ol-527.3 | 32m | Consider `tmux attach -t demigod-ol-527-3` |

### Options
1. Continue waiting: `/swarm --mode=distributed --wait --timeout 60m`
2. Attach to slow workers: `tmux attach -t demigod-ol-527-3`
3. Kill and retry: `tmux kill-session -t demigod-ol-527-3`
```

### Distributed vs Local Mode Summary

| Behavior | Local | Distributed |
|----------|---------|---------|
| Spawn mechanism | `TeamCreate` + `Task(team_name=...)` | `tmux new-session -d` |
| Worker entry point | Inline prompt (teammate) | `/implement <bead-id> --mode=distributed` |
| Process isolation | Team per wave (fresh context) | Separate processes |
| Persistence | Tied to Mayor | Survives Mayor crash |
| Coordination | `SendMessage` (native teams) | Agent Mail messages |
| File conflicts | Workers report to lead | File reservations |
| Retry mechanism | `SendMessage` to idle worker | Re-spawn |
| Debugging | Limited | `tmux attach -t <session>` |
| Resource overhead | Low | Medium (N tmux sessions) |
| Setup requirements | None (native teams built-in) | tmux + Agent Mail |

### When to Use Distributed Mode

| Scenario | Recommendation |
|----------|---------------|
| Quick parallel tasks (<5 min each) | Local |
| Long-running work (>10 min each) | Distributed |
| Need to debug stuck workers | Distributed |
| Multi-file changes across workers | Distributed (file reservations) |
| Mayor might disconnect | Distributed (persistence) |
| Complex coordination needed | Distributed |
| Simple, isolated tasks | Local |

### Example: Full Distributed Mode Swarm

```bash
# 1. Start Agent Mail (if not running)
# Start your Agent Mail MCP server (implementation-specific)
# See docs/agent-mail.md

# 2. In Claude session, run distributed mode swarm
/swarm --mode=distributed --max-workers=3 --wait

# Output:
# Pre-flight: tmux OK, Agent Mail OK
# Registered as Mayor: GoldenPeak
# Wave 1: 3 ready beads
# Spawning: demigod-ol-527-1 for ol-527.1
# Spawning: demigod-ol-527-2 for ol-527.2
# Spawning: demigod-ol-527-3 for ol-527.3
# Monitoring...
# [15:32] ACCEPTED from GreenCastle (ol-527.1)
# [15:32] ACCEPTED from BlueMountain (ol-527.2)
# [15:33] ACCEPTED from RedValley (ol-527.3)
# [15:40] PROGRESS from GreenCastle: Step 4 - implementing auth
# [15:45] DONE from GreenCastle (ol-527.1) - commit abc123
# [15:48] DONE from BlueMountain (ol-527.2) - commit def456
# [15:55] FAILED from RedValley (ol-527.3) - tests failed
#
# Wave complete: 2 done, 1 failed
# Sessions cleaned up (except failed)
# Use `tmux attach -t demigod-ol-527-3` to debug the failed worker
```

### Fallback Behavior

If distributed mode requested but requirements not met:

```
Error: Distributed mode requires tmux and Agent Mail.

Missing:
- [ ] tmux: Install with `brew install tmux`
- [x] Agent Mail: Running at localhost:8765

Falling back to local mode? [y/N]
```

If user confirms, degrade to local mode execution. Otherwise, exit with error.

---

## OL Wave Integration

When `/swarm --from-wave <json-file>` is invoked, the swarm reads wave data from an OL hero hunt output file and executes it with completion backflow to OL.

### Pre-flight

```bash
# --from-wave requires ol CLI on PATH
which ol >/dev/null 2>&1 || {
    echo "Error: ol CLI required for --from-wave. Install ol or use swarm without wave integration."
    exit 1
}
```

If `ol` is not on PATH, exit immediately with the error above. Do not fall back to normal swarm mode.

### Input Format

The `--from-wave` JSON file contains `ol hero hunt` output:

```json
{
  "wave": [
    {"id": "ol-527.1", "title": "Add auth middleware", "spec_path": "quests/ol-527/specs/ol-527.1.md", "priority": 1},
    {"id": "ol-527.2", "title": "Fix rate limiting", "spec_path": "quests/ol-527/specs/ol-527.2.md", "priority": 2}
  ],
  "blocked": [
    {"id": "ol-527.3", "title": "Integration tests", "blocked_by": ["ol-527.1", "ol-527.2"]}
  ],
  "completed": [
    {"id": "ol-527.0", "title": "Project setup"}
  ]
}
```

### Execution

1. **Parse the JSON file** and extract the `wave` array.

2. **Create TaskList tasks** from wave entries (one `TaskCreate` per entry):

```
for each entry in wave:
    TaskCreate(
        subject="[{entry.id}] {entry.title}",
        description="OL bead {entry.id}\nSpec: {entry.spec_path}\nPriority: {entry.priority}\n\nRead the spec file at {entry.spec_path} for full requirements.",
        metadata={
            "ol_bead_id": entry.id,
            "ol_spec_path": entry.spec_path,
            "ol_priority": entry.priority
        }
    )
```

3. **Execute swarm normally** on those tasks (Step 2 onward from main execution flow). Tasks are ordered by priority (lower number = higher priority).

4. **Completion backflow**: After each worker completes a bead task AND passes validation, the team lead runs the OL ratchet command to report completion back to OL:

```bash
# Extract quest ID from bead ID (e.g., ol-527.1 → ol-527)
QUEST_ID=$(echo "$BEAD_ID" | sed 's/\.[^.]*$//')

ol hero ratchet "$BEAD_ID" --quest "$QUEST_ID"
```

**Ratchet result handling:**

| Exit Code | Meaning | Action |
|-----------|---------|--------|
| 0 | Bead complete in OL | Mark task completed, log success |
| 1 | Ratchet validation failed | Mark task as failed, log the validation error from stderr |

5. **After all wave tasks complete**, report a summary that includes both swarm results and OL ratchet status for each bead.

### Example

```
/swarm --from-wave /tmp/wave-ol-527.json

# Reads wave JSON → creates 2 tasks from wave entries
# Spawns workers for ol-527.1 and ol-527.2
# On completion of ol-527.1:
#   ol hero ratchet ol-527.1 --quest ol-527 → exit 0 → bead complete
# On completion of ol-527.2:
#   ol hero ratchet ol-527.2 --quest ol-527 → exit 0 → bead complete
# Wave done: 2/2 beads ratcheted in OL
```

---

## References

- **Agent Mail Protocol:** See `skills/shared/agent-mail-protocol.md` for message format specifications
- **Parser (Go):** `cli/internal/agentmail/` - shared parser for all message types
