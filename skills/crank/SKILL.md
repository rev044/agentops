---
name: crank
description: 'Fully autonomous epic execution. Runs until ALL children are CLOSED. Uses /swarm for parallel wave execution. NO human prompts, NO stopping.'
---

# Crank Skill

> **Quick Ref:** Autonomous epic execution. Uses `/swarm` for each wave until DONE. Output: closed issues + final vibe.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Autonomous execution: implement all issues until the epic is DONE.

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

### Step 5: Track Progress

After swarm completes the wave:

1. Update beads issues based on TaskList results:
```bash
bd update <issue-id> --status closed 2>/dev/null
```

2. Track changed files:
```bash
git diff --name-only HEAD~5 2>/dev/null | sort -u
```

3. **Record ratchet progress (ao integration):**
```bash
# If ao CLI available, record wave completion
if command -v ao &>/dev/null; then
    ao ratchet record implement 2>/dev/null
    echo "Ratchet: recorded wave $wave completion"
fi
```

**Note:** Skip per-wave vibe - validation is batched at the end to save context.

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
