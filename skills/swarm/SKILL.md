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
    +-> Spawn: Task(team_name=..., name="worker-<id>") for each
    |       Workers join team, claim tasks, execute atomically
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

**For each ready task, spawn a worker as a teammate:**

```
Task(
  subagent_type="general-purpose",
  team_name="swarm-<epoch>",
  name="worker-<task-id>",
  prompt="You are Worker on swarm team \"swarm-<epoch>\".

Your Assignment: Task #<id>: <subject>
<description>

Instructions:
1. Claim your task: TaskUpdate(taskId=\"<id>\", status=\"in_progress\", owner=\"worker-<task-id>\")
2. Execute autonomously — create/edit files as needed, verify your work
3. Send completion message to team lead with summary of what you did
4. If blocked, message team lead explaining the issue

Rules:
- Work only on YOUR assigned task
- Do NOT claim other tasks or message other workers
- Commit with a message referencing the task ID"
)
```

Workers can access `TaskList`/`TaskUpdate` to claim and update their assigned tasks. Mayor still validates before accepting.

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
# Check for uncommitted changes (agent should have committed)
git status --porcelain

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
1. Check git status for changes
2. Review diffs
3. Run any additional tests/validation
4. Commit combined work if needed

### Step 5a: Cleanup Team

After wave completes:
```
# Shutdown each worker
SendMessage(type="shutdown_request", recipient="worker-<task-id>", content="Wave complete")

# Delete team for this wave
TeamDelete()
```

> **Note:** `TeamDelete()` deletes the team associated with this session's `TeamCreate()` call. If running concurrent teams (e.g., council inside crank), each team is cleaned up in the session that created it.

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

## Key Points

- **Pure Claude-native** - No tmux, no external scripts
- **Native teams** - `TeamCreate` + `Task(team_name=...)` + `SendMessage` for coordination
- **Workers access TaskList** - Workers claim tasks and update status via `TaskUpdate`
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
- **Filesystem for artifacts** - Files written, commits made
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

## Distributed Mode: tmux + Agent Mail

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
  model: "opus-4.5"
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

## References

- **Agent Mail Protocol:** See `skills/shared/agent-mail-protocol.md` for message format specifications
- **Parser (Go):** `cli/internal/agentmail/` - shared parser for all message types
