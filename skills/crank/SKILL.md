---
name: crank
description: 'Fully autonomous epic execution. Runs until ALL children are CLOSED. Local mode uses /swarm (Task tool). Distributed mode uses /swarm --mode=distributed (tmux + Agent Mail) for persistence and coordination. NO human prompts, NO stopping.'
dependencies:
  - swarm       # required - executes each wave
  - vibe        # required - final validation
  - implement   # required - individual issue execution
  - beads       # optional - issue tracking via bd CLI (fallback: TaskList)
  - post-mortem # optional - suggested for learnings extraction
---

# Crank Skill

> **Quick Ref:** Autonomous epic execution. Local mode: `/swarm` for each wave. Distributed mode: `/swarm --mode=distributed` (tmux + Agent Mail). Output: closed issues + final vibe.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Autonomous execution: implement all issues until the epic is DONE.

**CLI dependencies:** bd (issue tracking), ao (knowledge flywheel). Both optional — see `skills/shared/SKILL.md` for fallback table. If bd is unavailable, use TaskList for issue tracking and skip beads sync. If ao is unavailable, skip knowledge injection/extraction.

**Requires:** bd CLI (beads) for issue tracking.

## Architecture: Crank + Swarm

```
Crank (orchestrator)           Swarm (executor)
    |                              |
    +-> bd ready (wave issues)     |
    |                              |
    +-> TaskCreate from beads  --->+-> Spawn agents (fresh context)
    |                              |
    +-> /swarm                 --->+-> Execute in parallel
    |                              |
    +-> Verify + bd update     <---+-> Results
    |                              |
    +-> Loop until epic DONE       |
```

**Separation of concerns:**
- **Crank** = Beads-aware orchestration, epic lifecycle, knowledge flywheel
- **Swarm** = Fresh-context parallel execution (Ralph Wiggum pattern)

**Requires:** bd CLI (beads) for issue tracking.

## Global Limits

**MAX_EPIC_WAVES = 50** (hard limit across entire epic)

This prevents infinite loops on circular dependencies or cascading failures.

**Why 50?**
- Typical epic: 5-10 issues
- With retries: ~5 waves max
- 50 = safe upper bound

## Completion Enforcement (The Sisyphus Rule)

**THE SISYPHUS RULE:** Not done until explicitly DONE.

After each wave, output completion marker:
- `<promise>DONE</promise>` - Epic truly complete, all issues closed
- `<promise>BLOCKED</promise>` - Cannot proceed (with reason)
- `<promise>PARTIAL</promise>` - Incomplete (with remaining items)

**Never claim completion without the marker.**

## Execution Steps

Given `/crank [epic-id]`:

### Step 0: Load Knowledge Context (ao Integration)

**Search for relevant learnings before starting the epic:**

```bash
# If ao CLI available, inject prior knowledge about epic execution
if command -v ao &>/dev/null; then
    # Search for relevant learnings
    ao search "epic execution implementation patterns" 2>/dev/null | head -20

    # Check flywheel status
    ao flywheel status 2>/dev/null

    # Get current ratchet state
    ao ratchet status 2>/dev/null
fi
```

If ao not available, skip this step and proceed. The knowledge flywheel enhances but is not required.

### Step 1: Identify the Epic

**If epic ID provided:** Use it directly. Do NOT ask for confirmation.

**If no epic ID:** Discover it:
```bash
bd list --type epic --status open 2>/dev/null | head -5
```

If multiple epics found, ask user which one.

### Step 1a: Initialize Wave Counter

```bash
# Initialize crank tracking in epic notes
bd update <epic-id> --append-notes "CRANK_START: wave=0 at $(date -Iseconds)" 2>/dev/null
```

Track in memory: `wave=0`

### Step 2: Get Epic Details

```bash
bd show <epic-id> 2>/dev/null
```

### Step 3: List Ready Issues (Current Wave)

Find issues that can be worked on (no blockers):
```bash
bd ready 2>/dev/null
```

**`bd ready` returns the current wave** - all unblocked issues. These can be executed in parallel because they have no dependencies on each other.

### Step 3a: Pre-flight Check - Issues Exist

**Verify there are issues to work on:**

**If 0 ready issues found:**
```
STOP and return error:
  "No ready issues found for this epic. Either:
   - All issues are blocked (check dependencies)
   - Epic has no child issues (run /plan first)
   - All issues already completed"
```

Do NOT proceed with empty issue list - this produces false "epic complete" status.

### Step 4: Execute Wave via Swarm

**BEFORE each wave:**
```bash
# Increment wave counter
wave=$((wave + 1))
bd update <epic-id> --append-notes "CRANK_WAVE: $wave at $(date -Iseconds)" 2>/dev/null

# CHECK GLOBAL LIMIT
if [[ $wave -ge 50 ]]; then
    echo "<promise>BLOCKED</promise>"
    echo "Global wave limit (50) reached. Remaining issues:"
    bd children <epic-id> --status open 2>/dev/null
    # STOP - do not continue
fi
```

**Wave Execution via Swarm:**

1. **Get ready issues from Step 3**
2. **Create TaskList tasks from beads issues:**

For each ready beads issue, create a corresponding TaskList task:
```
TaskCreate(
  subject="<issue-id>: <issue-title>",
  description="Implement beads issue <issue-id>.

Details from beads:
<paste issue details from bd show>

Execute using /implement <issue-id>. Mark complete when done.",
  activeForm="Implementing <issue-id>"
)
```

3. **Add dependencies if issues have beads blockedBy:**
```
TaskUpdate(taskId="2", addBlockedBy=["1"])
```

4. **Invoke swarm to execute the wave:**
```
Tool: Skill
Parameters:
  skill: "agentops:swarm"
```

Swarm will:
- Find all unblocked TaskList tasks
- Spawn background agents with fresh context (Ralph pattern)
- Execute them in parallel
- Wait for notifications

5. **After swarm completes, verify beads status:**
```bash
# For each completed TaskList task, close the beads issue
bd update <issue-id> --status closed 2>/dev/null
```

### Step 5: Validate Before Accepting (MANDATORY)

> **CRITICAL**: Swarm completion claims are NOT trusted. Run validation BEFORE closing issues.

Swarm executes its own per-task validation (see `skills/shared/validation-contract.md`), but crank adds an additional verification layer for beads integration.

**For each issue reported complete by swarm:**

1. **Verify swarm validation passed:**
   ```
   TaskList() → check task status == "completed" (not just notified)
   ```
   If task is still pending/blocked after notification, swarm validation failed.

2. **Run issue-level validation checks:**

   | Check | Command | Fail Action |
   |-------|---------|-------------|
   | Commit exists | `git log --oneline -1 \| grep <issue-id>` | Retry |
   | Files changed | `git diff --name-only HEAD~1` | Verify expected |
   | Tests pass | `<project test command>` | Retry with hint |
   | No regressions | `git diff HEAD~1 \| grep -v "^+"` | Review |

3. **On validation PASS:**
   ```bash
   bd update <issue-id> --status closed 2>/dev/null
   ```

4. **On validation FAIL:**
   - Do NOT close the beads issue
   - Add failure context to issue:
     ```bash
     bd comments add <issue-id> "Validation failed: <reason>. Retrying..."
     ```
   - Re-add to next wave (swarm will re-execute)
   - After 3 failures, escalate:
     ```bash
     bd update <issue-id> --labels BLOCKER
     bd comments add <issue-id> "ESCALATED: 3 validation failures. Human review required."
     ```

**After all validations complete:**

5. Track changed files:
```bash
git diff --name-only HEAD~5 2>/dev/null | sort -u
```

6. **Record ratchet progress (ao integration):**
```bash
# If ao CLI available, record wave completion
if command -v ao &>/dev/null; then
    ao ratchet record implement 2>/dev/null
    echo "Ratchet: recorded wave $wave completion"
fi
```

**Note:** Per-issue validation is lightweight. Full vibe is still batched at the end for comprehensive review.

### Step 6: Check for More Work

After completing a wave:
1. Clear completed tasks from TaskList
2. Check if new beads issues are now unblocked: `bd ready`
3. If yes, return to Step 4 (create new TaskList tasks, invoke swarm)
4. If no more issues after 3 retry attempts, proceed to Step 7
5. **Max retries:** If issues remain blocked after 3 checks, escalate: "Epic blocked - cannot unblock remaining issues"

### Step 7: Final Batched Validation

When all issues complete, run ONE comprehensive vibe on recent changes:

```bash
# Get list of changed files from recent commits
git diff --name-only HEAD~10 2>/dev/null | sort -u
```

**Run vibe on recent changes:**
```
Tool: Skill
Parameters:
  skill: "agentops:vibe"
  args: "recent"
```

**If CRITICAL issues found:**
1. Fix them
2. Re-run vibe on affected files
3. Only proceed to completion when clean

### Step 8: Extract Learnings (ao Integration)

**Before reporting completion, extract learnings from the session:**

```bash
# If ao CLI available, forge learnings from this epic execution
if command -v ao &>/dev/null; then
    # Extract learnings from recent session transcripts
    ao forge transcript ~/.claude/projects/*/conversations/*.jsonl 2>/dev/null

    # Show flywheel status post-execution
    echo "=== Flywheel Status ==="
    ao flywheel status 2>/dev/null

    # Show pending learnings for review
    ao pool list --tier=pending 2>/dev/null | head -10
fi
```

If ao not available, skip learning extraction. Recommend user runs `/post-mortem` manually.

### Step 9: Report Completion

Tell the user:
1. Epic ID and title
2. Number of issues completed
3. Total iterations used (of 50 max)
4. Final vibe results
5. Flywheel status (if ao available)
6. Suggest running `/post-mortem` to review and promote learnings

**Output completion marker:**
```
<promise>DONE</promise>
Epic: <epic-id>
Issues completed: N
Iterations: M/50
Flywheel: <status from ao flywheel status>
```

If stopped early:
```
<promise>BLOCKED</promise>
Reason: <global limit reached | unresolvable blockers>
Issues remaining: N
Iterations: M/50
```

## The FIRE Loop

Crank follows FIRE for each wave:

| Phase | Action |
|-------|--------|
| **FIND** | `bd ready` - get unblocked beads issues |
| **IGNITE** | Create TaskList tasks, invoke `/swarm` |
| **REAP** | Swarm collects results, crank syncs to beads |
| **ESCALATE** | Fix blockers, retry failures |

**Parallel Wave Model (via Swarm):**
```
Wave 1: bd ready → [issue-1, issue-2, issue-3]
        ↓
        TaskCreate for each issue
        ↓
        /swarm → spawns 3 fresh-context agents
                  ↓         ↓         ↓
               DONE      DONE      BLOCKED
                                     ↓
                               (retry in next wave)
        ↓
        bd update --status closed for completed

Wave 2: bd ready → [issue-4, issue-3-retry]
        ↓
        TaskCreate for each
        ↓
        /swarm → spawns 2 fresh-context agents
        ↓
        bd update for completed

Final vibe on all changes → Epic DONE
```

Loop until all beads issues are CLOSED.

## Key Rules

- **If epic ID given, USE IT** - don't ask for confirmation
- **Swarm for each wave** - delegates parallel execution to swarm
- **Fresh context per issue** - swarm provides Ralph pattern isolation
- **Batch validation at end** - ONE vibe at the end saves context
- **Fix CRITICAL before completion** - address findings before reporting done
- **Loop until done** - don't stop until all issues closed
- **Autonomous execution** - minimize human prompts
- **Respect wave limit** - STOP at 50 waves (hard limit)
- **Output completion markers** - DONE, BLOCKED, or PARTIAL (required)
- **Knowledge flywheel** - load learnings at start, forge at end (ao optional)
- **Beads ↔ TaskList sync** - crank bridges beads issues to TaskList for swarm

---

## Distributed Mode: Agent Mail Orchestration

> **When:** Agent Mail MCP tools are available AND `--mode=distributed` is set

Distributed mode transforms /crank from a TaskList-based orchestrator to a persistent orchestrator that runs waves via `/swarm --mode=distributed` and coordinates through Agent Mail.

### Why Distributed Mode?

| Local (Task Tool) | Distributed (Agent Mail) |
|-------------------|--------------------------|
| Subagents inside session | Independent Claude sessions |
| TaskOutput for results | Agent Mail messages |
| No help routing | Chiron pattern for HELP_REQUESTs |
| Race conditions on files | File reservations |
| In-process monitoring | Inbox-based monitoring |

**Use distributed mode when:**
- Running long epics that may exceed session limits
- Need coordination across multiple Claude sessions
- Want Chiron (expert helper) to answer stuck demigods
- Need advisory file locking between parallel workers

### Mode Detection

```bash
# Auto-detect Agent Mail availability
AGENT_MAIL_AVAILABLE=false

# Method 1: Check HTTP endpoint
if curl -s http://localhost:8765/health 2>/dev/null | grep -q "ok"; then
    AGENT_MAIL_AVAILABLE=true
fi

# Method 2: Check MCP tools in current session
# (if mcp__mcp-agent-mail__* tools are available)

# Explicit flag takes precedence
# /crank epic-123 --mode=distributed
```

**Mode selection:**
| Condition | Mode |
|-----------|------|
| `--mode=distributed` AND Agent Mail available | Distributed |
| `--mode=distributed` AND Agent Mail unavailable | Error: "Agent Mail required for distributed mode" |
| No flag AND Agent Mail available | Local (default) |
| No flag AND Agent Mail unavailable | Local |

### Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--mode=distributed` | Force distributed orchestration mode | `local` |
| `--agent-mail` | Enable Agent Mail (alias for `--mode=distributed`) | `false` |
| `--orchestrator-id` | Crank's identity in Agent Mail | `crank-<epic-id>` |
| `--chiron` | Enable Chiron pattern for help requests | `true` in distributed |
| `--max-parallel` | Max concurrent demigods per wave | `5` |

### Distributed Mode Architecture

```
Crank (orchestrator)              Agent Mail              Demigods
    |                                 |                       |
    +-> bd ready (wave issues)        |                       |
    |                                 |                       |
    +-> Reserve files for wave ------>|                       |
    |                                 |                       |
    +-> /swarm --mode=distributed ----|---- spawns workers -->|
    |                                 |                       |
    +-> Poll inbox <------------------|<-- BEAD_ACCEPTED -----|
    |                                 |<-- PROGRESS ----------|
    |                                 |<-- HELP_REQUEST ------|
    |   (route to Chiron) ----------->|                       |
    |                                 |<-- OFFERING_READY ----|
    |                                 |                       |
    +-> Verify + bd update            |                       |
    |                                 |                       |
    +-> Release file reservations --->|                       |
    |                                 |                       |
    +-> Loop until epic DONE          |                       |
```

**Separation of concerns:**
- **Crank** = Beads-aware orchestration, Agent Mail coordination, file reservations
- **Demigods** = Fresh-context parallel execution with Agent Mail reporting
- **Chiron** = Expert helper that responds to HELP_REQUESTs

### Distributed Mode Execution Steps

When `--mode=distributed` is enabled:

#### Step 0: Initialize Orchestrator Identity

```bash
# Register crank as an orchestrator in Agent Mail
ORCHESTRATOR_ID="${ORCHESTRATOR_ID:-crank-$(echo $EPIC_ID | tr -d '-')}"
PROJECT_KEY=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
```

**MCP Tool Call:**
```
Tool: mcp__mcp-agent-mail__register_agent
Parameters:
  project_key: "<project-key>"
  program: "crank-skill"
  model: "claude-opus-4-5-20250101"
  task_description: "Orchestrating epic <epic-id>"
```

#### Step 1: Reserve Files for Wave (Before Spawn)

**Before spawning demigods for a wave, reserve files to prevent conflicts:**

1. **Analyze wave issues to identify files:**
```bash
# For each issue in the wave, predict affected files
# Use issue description + codebase patterns
```

2. **Reserve files:**
```
Tool: mcp__mcp-agent-mail__file_reservation_paths
Parameters:
  project_key: "<project-key>"
  agent_name: "<orchestrator-id>"
  paths:
    - "src/auth.py"           # Issue 1 will modify
    - "src/models/user.py"    # Issue 2 will modify
    - "tests/test_auth.py"    # Issue 1 will modify
  exclusive: false  # Advisory reservations
```

**File reservation strategy:**
| Scenario | Action |
|----------|--------|
| Files already reserved by another | Log warning, proceed (advisory) |
| Cannot predict files | Skip reservation, rely on demigod to reserve |
| Conflicting issues in same wave | Serialize those issues (don't spawn parallel) |

#### Step 2: Execute Wave via /swarm (Distributed Mode)

In distributed mode, crank delegates parallel execution to swarm’s distributed mode:

```
/swarm --mode=distributed --bead-ids <issue-1>,<issue-2>,<issue-3> --wait
```

Swarm handles worker spawning (tmux) and coordination (Agent Mail). See `skills/swarm/SKILL.md` for the distributed worker lifecycle and monitoring.

#### Step 3: Monitor via Inbox (Not TaskOutput)

**Polling loop for wave monitoring:**

```bash
POLL_INTERVAL=30  # seconds
MAX_WAIT=$((30 * 60))  # 30 minutes per wave max
ELAPSED=0

while [ $ELAPSED -lt $MAX_WAIT ]; do
    # Fetch inbox for all messages
    # MCP: mcp__mcp-agent-mail__fetch_inbox

    # Process messages by type:
    # - BEAD_ACCEPTED: Log demigod started
    # - PROGRESS: Update tracking
    # - HELP_REQUEST: Route to Chiron
    # - OFFERING_READY: Mark demigod complete
    # - FAILED: Mark demigod failed
    # - CHECKPOINT: Handle context exhaustion

    # Check if all demigods in wave complete
    if all_complete; then
        break
    fi

    sleep $POLL_INTERVAL
    ELAPSED=$((ELAPSED + POLL_INTERVAL))
done
```

**MCP Tool Call (fetch inbox):**
```
Tool: mcp__mcp-agent-mail__fetch_inbox
Parameters:
  project_key: "<project-key>"
  agent_name: "<orchestrator-id>"
```

**Message handling:**
| Message Subject | Action |
|-----------------|--------|
| `BEAD_ACCEPTED` | Log: "Demigod <id> accepted <issue>" |
| `PROGRESS` | Log: Update progress tracker |
| `HELP_REQUEST` | Route to Chiron (see Step 4) |
| `OFFERING_READY` | Verify + close beads issue |
| `FAILED` | Log failure, add to retry queue |
| `CHECKPOINT` | Handle partial progress, spawn replacement |

#### Step 4: Chiron Pattern for Help Requests

**When HELP_REQUEST received, route to Chiron:**

```
Tool: mcp__mcp-agent-mail__send_message
Parameters:
  project_key: "<project-key>"
  sender_name: "<orchestrator-id>"
  to: "chiron@olympus"
  subject: "HELP_ROUTE"
  body_md: |
    ## Help Request Routed
    From: <demigod-id>
    Issue: <issue-id>

    ## Original Request
    <paste help request body>

    ## Context
    Epic: <epic-id>
    Wave: <wave-number>

    Please respond to <demigod-id> with HELP_RESPONSE.
  thread_id: "<issue-id>"
  ack_required: false
```

**Chiron is expected to:**
1. Receive HELP_ROUTE
2. Analyze the problem
3. Send HELP_RESPONSE directly to demigod
4. Demigod continues with guidance

**Fallback if Chiron unavailable (timeout > 2 min):**
- Log: "No Chiron response - demigod must proceed with best judgment"
- Demigod either succeeds or fails on its own

#### Step 5: Verify Completion via Agent Mail

**When OFFERING_READY received:**

1. **Acknowledge message:**
```
Tool: mcp__mcp-agent-mail__acknowledge_message
Parameters:
  project_key: "<project-key>"
  message_id: "<message-id>"
```

2. **Verify work:**
```bash
# Check commit exists
git log --oneline -1

# Check files modified
git diff --name-only HEAD~1

# Run validation
# (same as local mode batched vibe)
```

3. **Close beads issue:**
```bash
bd update <issue-id> --status closed 2>/dev/null
```

4. **Send acknowledgment to demigod:**
```
Tool: mcp__mcp-agent-mail__send_message
Parameters:
  project_key: "<project-key>"
  sender_name: "<orchestrator-id>"
  to: "<demigod-id>"
  subject: "OFFERING_ACCEPTED"
  body_md: |
    Issue <issue-id> closed.
    Commit: <commit-sha>
    Thank you for your service.
  thread_id: "<issue-id>"
  ack_required: false
```

#### Step 6: Handle Failures

**When FAILED message received:**

1. **Log failure:**
```bash
bd update <issue-id> --append-notes "FAILED: <reason> at $(date -Iseconds)" 2>/dev/null
```

2. **Decide retry strategy:**
| Failure Type | Action |
|--------------|--------|
| `TESTS_FAIL` | Add to retry queue with hint |
| `BUILD_FAIL` | Add to retry queue |
| `SPEC_IMPOSSIBLE` | Mark blocked, escalate |
| `CONTEXT_HIGH` | Spawn fresh demigod with checkpoint |
| `ERROR` | Add to retry queue (max 3 attempts) |

3. **For retry:**
```bash
# Track retry count in beads notes
bd update <issue-id> --append-notes "RETRY: attempt $(( retry_count + 1 )) at $(date -Iseconds)" 2>/dev/null
```

#### Step 7: Release File Reservations

**After wave completes (success or failure):**

```
Tool: mcp__mcp-agent-mail__release_file_reservations
Parameters:
  project_key: "<project-key>"
  agent_name: "<orchestrator-id>"
```

#### Step 8: Handle Checkpoints (Context Exhaustion)

**When demigod sends CHECKPOINT due to context exhaustion:**

1. **Parse checkpoint info:**
```markdown
## From CHECKPOINT message:
- Partial commit: abc123
- Progress: Steps 1-3 complete, Step 4 in progress
- Next steps: Complete Step 4, then Steps 5-7
```

2. **Spawn replacement demigod:**
```
Tool: Skill
Parameters:
  skill: "agentops:spawn"
  args: "--issue <issue-id> --resume --checkpoint <commit-sha> --orchestrator <orchestrator-id>"
```

3. **Replacement demigod gets:**
- Issue context (from beads)
- Checkpoint commit (partial work)
- Guidance for what remains

### Distributed Mode FIRE Loop

Distributed mode uses the same FIRE pattern with Agent Mail coordination:

| Phase | Local | Distributed |
|-------|---------|---------|
| **FIND** | `bd ready` | `bd ready` |
| **IGNITE** | TaskCreate + /swarm | File reserve + `/swarm --mode=distributed` |
| **REAP** | TaskOutput notifications | fetch_inbox polling |
| **ESCALATE** | Retry via swarm | Chiron for help + retry via spawn |

### Distributed Mode Parallel Wave Model

```
Wave 1: bd ready → [issue-1, issue-2, issue-3]
        ↓
        Reserve files for all 3 issues
        ↓
        /swarm --mode=distributed --bead-ids issue-1,issue-2,issue-3 --wait
        ↓
        Poll inbox:
          - BEAD_ACCEPTED (×3)
          - PROGRESS updates
          - HELP_REQUEST → route to Chiron
          - OFFERING_READY (×2)
          - FAILED (×1)
        ↓
        bd update --status closed (×2)
        Add issue-3 to retry queue
        ↓
        Release file reservations

Wave 2: bd ready → [issue-4, issue-3-retry]
        ↓
        (repeat pattern)

Final vibe on all changes → Epic DONE
```

### Distributed Mode Key Rules

- **Reserve files BEFORE spawn** - prevents conflicts between demigods
- **Monitor via inbox** - not TaskOutput (demigods are independent sessions)
- **Route HELP_REQUESTs** - Chiron answers stuck demigods
- **Acknowledge completions** - closes coordination loop
- **Handle checkpoints** - spawn replacements for context-exhausted demigods
- **Release reservations** - after each wave completes
- **Same wave limit** - MAX_EPIC_WAVES = 50 still applies
- **Same completion markers** - DONE, BLOCKED, PARTIAL

### Distributed vs Local Mode Summary

| Aspect | Local | Distributed |
|--------|---------|---------|
| Spawn mechanism | Task tool | `/swarm --mode=distributed` (tmux + Agent Mail) |
| Monitoring | TaskOutput | fetch_inbox polling |
| Help requests | User prompt | Chiron pattern |
| File conflicts | Race conditions | Advisory reservations |
| Context exhaustion | Agent fails | Checkpoint + replacement |
| Session scope | Single session | Cross-session |
| External coordination | No | Yes (multiple Claude instances) |

### Without Agent Mail

If Agent Mail is not available and `--mode=distributed` is requested:

```markdown
Error: Distributed mode requires Agent Mail.

To enable Agent Mail:
1. Start MCP Agent Mail server:
   Start your Agent Mail MCP server (implementation-specific). See `docs/agent-mail.md`.

2. Add to ~/.claude/mcp_servers.json:
   {
     "mcp-agent-mail": {
       "type": "http",
       "url": "http://127.0.0.1:8765/mcp/"
     }
   }

3. Restart Claude Code session

Falling back to local mode.
```

### Integration with Other Skills

| Skill | Distributed Mode Integration |
|-------|------------------------------|
| `/swarm` | Executes waves (local or distributed mode) |
| `/implement` | Run by demigods with `--agent-mail` flag |
| `/inbox` | Used by crank for monitoring (or direct fetch_inbox) |
| `/chiron` | Receives HELP_ROUTE, responds with HELP_RESPONSE |
| `/vibe` | Final validation (same as local mode) |

### Example Distributed Mode Session

```bash
# Start with distributed mode
/crank ol-527 --mode=distributed

# Output:
# Distributed mode: Agent Mail orchestration enabled
# Orchestrator ID: crank-ol527
# Project: /Users/fullerbt/gt/olympus
#
# Wave 1: Spawning 3 demigods...
#   - /swarm --mode=distributed --bead-ids ol-527.1,ol-527.2,ol-527.3 --wait
#
# Monitoring inbox...
#   [00:15] BEAD_ACCEPTED from demigod-ol-527-1
#   [00:16] BEAD_ACCEPTED from demigod-ol-527-2
#   [00:17] BEAD_ACCEPTED from demigod-ol-527-3
#   [02:30] PROGRESS from demigod-ol-527-1: Step 4 complete
#   [03:45] HELP_REQUEST from demigod-ol-527-2 → routing to Chiron
#   [04:10] OFFERING_READY from demigod-ol-527-1
#   [05:00] OFFERING_READY from demigod-ol-527-3
#   [06:15] OFFERING_READY from demigod-ol-527-2
#
# Wave 1 complete: 3/3 issues closed
# Releasing file reservations...
#
# Wave 2: Spawning 2 demigods...
# ...
#
# <promise>DONE</promise>
# Epic: ol-527
# Issues completed: 8
# Iterations: 3/50
# Mode: distributed (Agent Mail)
```

---

## References

- **Agent Mail Protocol:** See `skills/shared/agent-mail-protocol.md` for message format specifications
- **Parser (Go):** `cli/internal/agentmail/` - shared parser for all message types
