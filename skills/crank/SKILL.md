---
name: crank
description: >
  Fully autonomous epic execution. Runs until ALL children are CLOSED.
  Loops through beads issues, runs /implement on each, validates with /vibe.
  NO human prompts, NO stopping.
version: 2.2.0
tier: orchestration
context: inline
triggers:
  - "crank"
  - "crank this epic"
  - "run autonomously"
  - "execute to completion"
  - "crank it"
  - "full auto"
  - "run unattended"
allowed-tools: Bash, Read, Glob, Grep, TodoWrite, Task
skills:
  - beads
  - implement
  - vibe
---

# crank: Autonomous Epic Execution

> **Runs until epic is CLOSED. Auto-adapts to Mayor (parallel) or Crew (sequential) mode.**

## Philosophy: The Brownian Ratchet (FIRE Loop)

Crank is the full implementation of the Brownian Ratchet pattern via the **FIRE loop**:

| FIRE Phase | Ratchet Role | Description |
|------------|--------------|-------------|
| **FIND** | Read state | Identify ready work, burning work, reaped work |
| **IGNITE** | **Chaos** | Spark parallel polecats (Mayor) or start work (Crew) |
| **REAP** | **Filter + Ratchet** | Validate, merge (permanent), close issues |
| **ESCALATE** | Recovery | Retry failures or escalate blockers to human |

The FIRE loop IS the ratchet:
```
FIND → IGNITE (chaos) → REAP (filter + ratchet) → ESCALATE → loop
```

**Key insight:** Polecats can fail independently. Each successful merge ratchets forward.
The system extracts progress from parallel attempts, filtering failures automatically.

See [fire.md](fire.md) for full loop specification.

## Role Detection

Crank auto-detects execution mode based on current context:

```bash
# Check role by directory structure
if [[ "$PWD" == */mayor/* ]] || [[ "$PWD" == ~/gt ]]; then
    ROLE="mayor"    # Can spawn polecats
else
    ROLE="crew"     # Execute directly
fi
```

| Role | Execution Style | Parallelism | Command |
|------|-----------------|-------------|---------|
| **Mayor** | Dispatch to polecats via `gt sling` | Up to 8 concurrent | `gt sling <issue> <rig>` |
| **Crew** | Execute directly via `/implement` | Sequential | `/implement <issue>` |

## Quick Start

```bash
# Auto-discover epic (finds open epic in current context)
/crank

# Explicit epic - USE IT DIRECTLY, NO DISCOVERY
/crank <epic-id>

# Force a specific mode
/crank <epic-id> --mode=crew     # Sequential, no polecats
/crank <epic-id> --mode=mayor    # Parallel via polecats
```

## CRITICAL: Argument Handling

**RULE: If an epic ID is provided, USE IT IMMEDIATELY. Do NOT run discovery.**

```python
def parse_args(args):
    """Parse crank arguments."""
    if args and args[0].startswith(('ol-', 'ap-', 'gt-', 'be-', 'he-', 'ho-')):
        # Explicit epic ID provided - USE IT, NO QUESTIONS
        return {'epic': args[0], 'mode': parse_mode(args)}

    # Only run discovery if NO epic ID provided
    return {'epic': discover_epic(), 'mode': parse_mode(args)}
```

**Anti-pattern (DO NOT DO):**
```
User: /crank ol-rg3p
Claude: "I found multiple epics, which one?" <- WRONG! User said ol-rg3p!
```

**Correct behavior:**
```
User: /crank ol-rg3p
Claude: [Immediately starts cranking ol-rg3p, no questions]
```

## Discovery (ONLY when no epic ID provided)

When invoked with just `/crank` (no arguments), infer the target:

### Priority 1: Conversational Context

If the user mentions a topic (e.g., "/crank flywheel"), search:

```bash
bd search "flywheel" --type epic --status open
```

### Priority 2: Beads Discovery

```bash
EPICS=$(bd list --type epic --status open 2>/dev/null | head -5)
EPIC_COUNT=$(echo "$EPICS" | grep -c '^' 2>/dev/null || echo 0)

if [[ "$EPIC_COUNT" -eq 1 ]]; then
    EPIC_ID=$(echo "$EPICS" | awk '{print $1}')
    # USE IT - one epic, no ambiguity
elif [[ "$EPIC_COUNT" -gt 1 ]]; then
    # ASK - multiple epics, need clarification
    echo "[CRANK] Multiple open epics. Please specify: /crank <epic-id>"
fi
```

### Priority 3: Recent Context

Check conversation history for recently-mentioned epic IDs.

## The ODMCR Loop

Both modes use the same reconciliation loop, just different dispatch mechanisms:

```
┌─────────────────────────────────────────────────────────┐
│                    ODMCR LOOP                           │
│                                                         │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐          │
│  │ OBSERVE  │───►│ DISPATCH │───►│ MONITOR  │          │
│  └──────────┘    └──────────┘    └──────────┘          │
│       ▲                               │                 │
│       │          ┌──────────┐         │                 │
│       │          │  RETRY   │◄────────┤                 │
│       │          └──────────┘         │                 │
│       │               │               ▼                 │
│       │          ┌──────────┐    ┌──────────┐          │
│       └──────────│ (loop)   │◄───│ COLLECT  │          │
│                  └──────────┘    └──────────┘          │
│                                                         │
│  EXIT: All children status=closed                       │
└─────────────────────────────────────────────────────────┘
```

---

## Crew Mode Execution

When running as crew (in a crew workspace), crank executes issues sequentially:

### Crew OBSERVE
```bash
# Same as mayor - query epic state
bd ready --parent=<epic>
bd list --parent=<epic> --status=in_progress
bd list --parent=<epic> --status=closed
```

### Crew DISPATCH
```bash
# Execute directly, one at a time
issue=$(bd ready --parent=<epic> --limit=1 | awk '{print $1}')
/implement $issue
```

**Key difference**: No polecats, no convoy. Just sequential `/implement` calls.

### Crew MONITOR
```bash
# Check if current issue completed
bd show <issue> | grep "CLOSED"
```

### Crew COLLECT
```bash
# Verify completion
git status                    # Check for uncommitted work
bd sync                       # Sync beads
git add . && git commit       # If needed
```

### Crew RETRY
```bash
# On failure, retry the same issue
# After MAX_RETRIES, skip and mail escalation
```

### Crew Execution Flow

```python
def crank_crew(epic_id):
    """Sequential execution for crew context."""

    retry_counts = {}

    while not epic_complete(epic_id):
        # OBSERVE
        ready = bd_ready(parent=epic_id, limit=1)
        if not ready:
            # All remaining are blocked or in_progress elsewhere
            sleep(30)
            continue

        issue = ready[0]

        # DISPATCH (sequential - execute directly)
        bd_update(issue, status='in_progress')
        result = implement(issue)  # Calls /implement skill

        # COLLECT
        if result.success:
            bd_close(issue, reason=result.summary)
            git_commit_and_push()
        else:
            # RETRY
            retry_counts[issue] = retry_counts.get(issue, 0) + 1
            if retry_counts[issue] >= MAX_RETRIES:
                bd_update(issue, status='blocked',
                         notes=f"Failed after {MAX_RETRIES} attempts")
                mail_escalation(issue)
            else:
                # Will retry on next loop iteration
                pass

    # Complete
    bd_close(epic_id, reason="All children completed via crank (crew mode)")
```

---

## Mayor Mode Execution

When running as mayor (at town root or in mayor/), crank dispatches to polecats:

### Mayor DISPATCH
```bash
# Parallel dispatch to polecats (batch slinging)
READY=$(bd ready --parent=<epic> | awk '{print $1}')
IN_PROGRESS=$(bd list --parent=<epic> --status=in_progress | wc -l)
SLOTS=$((MAX_POLECATS - IN_PROGRESS))

# Batch sling - all issues in one command, each gets own polecat
BATCH=$(echo "$READY" | head -$SLOTS | tr '\n' ' ')
gt sling $BATCH <rig>

# Or check for stranded convoys first
gt convoy stranded  # Find convoys with ready work but no workers
```

### Mayor MONITOR
```bash
# Poll convoy status (low-token)
gt convoy list | grep <epic-prefix>
gt convoy status <convoy-id>
```

### Mayor Execution Flow

```python
def crank_mayor(epic_id, rig):
    """Parallel execution via polecats."""

    retry_counts = {}

    while not epic_complete(epic_id):
        # OBSERVE
        ready = bd_ready(parent=epic_id)
        in_progress = bd_list(parent=epic_id, status='in_progress')

        # DISPATCH (parallel to polecats)
        slots = MAX_POLECATS - len(in_progress)
        for issue in ready[:slots]:
            if retry_counts.get(issue, 0) < MAX_RETRIES:
                gt_sling(issue, rig)

        # MONITOR
        sleep(POLL_INTERVAL)
        convoy_status = gt_convoy_status()

        # COLLECT
        for completed in convoy_status.completed:
            verify_completion(completed)

        # RETRY
        for failed in convoy_status.failed:
            retry_counts[failed] = retry_counts.get(failed, 0) + 1
            if retry_counts[failed] >= MAX_RETRIES:
                handle_exhausted(failed)
            else:
                schedule_retry(failed, backoff=30 * 2**retry_counts[failed])

    # Complete
    bd_close(epic_id, reason="All children completed via crank (mayor mode)")
    mail_human(f"Epic {epic_id} completed")
```

---

## Mode Comparison

| Aspect | Crew Mode | Mayor Mode |
|--------|-----------|------------|
| **Parallelism** | Sequential (1 at a time) | Up to 8 concurrent |
| **Dispatch** | `/implement` directly | `gt sling` to polecats |
| **Context** | Uses current session | Spawns new sessions |
| **Speed** | Slower but simpler | Faster with parallelism |
| **Monitoring** | Inline (same session) | Convoy dashboard |
| **Best for** | Small epics, testing | Large epics, overnight |

---

## Arguments

| Arg | Purpose | Default |
|-----|---------|---------|
| `<epic-id>` | Epic to execute | Auto-discover (see above) |
| `--mode` | Force `crew` or `mayor` | Auto-detect |
| `--max` | Max concurrent polecats (mayor only) | 8 |
| `--dry-run` | Preview waves without executing | false |
| `status` | Show current crank progress | - |
| `stop` | Graceful stop at next checkpoint | - |

---

## Constants

| Constant | Default | Description |
|----------|---------|-------------|
| `MAX_POLECATS` | 8 | Max concurrent workers (mayor) |
| `POLL_INTERVAL` | 30s | Seconds between status checks |
| `MAX_RETRIES` | 3 | Retries before escalation |
| `BACKOFF_BASE` | 30s | Base for exponential backoff |

---

## Prerequisites

Before cranking:

```bash
# 1. Epic exists with children
bd show <epic-id>

# 2. At least one issue is ready
bd ready --parent=<epic>

# 3. If mayor mode: rig exists
gt rig list
```

---

## Example Sessions

### Crew Mode
```bash
> /crank ap-68ohb

[CRANK] Detected role: crew (sequential mode)
[CRANK] Starting autonomous execution of ap-68ohb
[OBSERVE] Epic ap-68ohb: 27 children, 0 closed, 27 ready
[DISPATCH] Implementing ap-68ohb.2 directly...
[IMPLEMENT] E2E: Gateway LLM proxy streaming and non-streaming
... implementation happens ...
[COLLECT] ap-68ohb.2 completed
[DISPATCH] Implementing ap-68ohb.4 directly...
... continues sequentially ...
[COMPLETE] Epic ap-68ohb finished. 27/27 children closed.
```

### Mayor Mode
```bash
> /crank ap-68ohb --mode=mayor

[CRANK] Running in mayor mode (parallel via polecats)
[CRANK] Target rig: ai_platform
[OBSERVE] Epic ap-68ohb: 27 children, 0 closed, 27 ready
[DISPATCH] Slinging ap-68ohb.2, .4, .5, .6, .7, .8, .9, .10 to ai_platform
[MONITOR] Convoy cv-xyz active, 8 polecats working
[MONITOR] ... (30s) ...
[COLLECT] ap-68ohb.2 completed, verified
... continues until all closed ...
[COMPLETE] Epic ap-68ohb finished. 27/27 children closed.
```

---

## Safety Rails

Even in full-auto mode, crank respects:

1. **Retry limits**: Issues don't retry forever
2. **Blocker escalation**: Humans notified of stuck work
3. **No force operations**: Never `--force` on git
4. **Beads integrity**: Always `bd sync` before state changes
5. **Commit after each issue (crew)**: Progress is never lost

---

## Session Recovery

If session ends mid-crank:

```bash
# State is recoverable from beads
bd list --parent=<epic> --status=in_progress  # What was running
bd list --parent=<epic> --status=closed       # What completed

# Resume with
/crank <epic>  # Picks up where it left off
```

---

## ao CLI Integration

Crank uses ao ratchet for the Brownian Ratchet pattern:

```bash
# Gate check before starting
ao ratchet check crank --epic <epic-id>

# Record each completed issue (called by /implement)
ao ratchet record implement --input "issue:<id>" --output "commits:..."

# Epic completion closes the flywheel
ao ratchet record crank --input "epic:<epic-id>" --output "closed + commits"

# Verify all children are ratcheted
ao ratchet verify --epic <epic-id>
```

Each REAP phase locks progress: completed issues cannot regress.

---

## Related

- [fire.md](fire.md) - FIRE loop specification (FIND/IGNITE/REAP/ESCALATE)
- [failure-taxonomy.md](failure-taxonomy.md) - Failure types and handling
- `/vibe` - Validation before merging
- `/implement` - Single issue execution (used by crew mode)
