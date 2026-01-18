# ODMCR Loop Specification

> **Observe-Dispatch-Monitor-Collect-Retry**: The reconciliation loop powering autonomous execution.

## Overview

ODMCR is a Kubernetes-inspired reconciliation pattern adapted for agent orchestration. Like a Kubernetes controller continuously reconciling desired vs actual state, ODMCR continuously drives an epic toward completion.

**Design philosophy**: Declare the goal (epic complete), let the loop figure out how to get there.

## Loop Phases

### OBSERVE Phase

**Purpose**: Build current state snapshot.

**Inputs**: Epic ID

**Outputs**: State object

```bash
# Commands executed
bd ready --parent=<epic>                    # Ready issues
bd list --parent=<epic> --status=in_progress  # In-flight
bd list --parent=<epic> --status=closed       # Completed
bd blocked --parent=<epic>                    # Waiting on deps
gt convoy list                               # Active convoys
gt polecat list <rig>                        # Worker status
```

**State object structure**:

```yaml
epic_state:
  epic_id: gt-0100
  total_children: 8
  ready: [gt-0101, gt-0102]
  in_progress: [gt-0103, gt-0104]
  closed: [gt-0105, gt-0106]
  blocked: [gt-0107, gt-0108]

  # Derived
  remaining: 6  # total - closed
  dispatchable: 2  # ready count
  capacity: 2  # MAX_POLECATS - in_progress
  complete: false  # remaining == 0

  # Convoy tracking
  active_convoys:
    - id: cv-001
      issues: [gt-0103, gt-0104]
      status: running

  # Polecat mapping
  polecats:
    gt-0103: cheedo
    gt-0104: singe
```

**Token cost**: ~200-300 tokens (mostly from bd output)

### DISPATCH Phase

**Purpose**: Send work to available polecats.

**Inputs**: State object, retry queue

**Outputs**: Updated convoy, dispatched issue list

**Decision logic**:

```python
def dispatch_phase(state, retry_queue):
    to_dispatch = []

    # Priority 1: Scheduled retries that are due
    for issue, scheduled_time in retry_queue:
        if now() >= scheduled_time:
            to_dispatch.append(issue)
            retry_queue.remove(issue)

    # Priority 2: Fresh ready issues
    for issue in state.ready:
        if issue not in to_dispatch:
            to_dispatch.append(issue)

    # Respect capacity
    available_slots = state.capacity
    to_dispatch = to_dispatch[:available_slots]

    # Execute dispatch
    for issue in to_dispatch:
        gt_sling(issue, rig)

    return to_dispatch
```

**Dispatch commands**:

```bash
# Individual dispatch (auto-creates convoy for tracking)
gt sling <issue> <rig>

# Multiple issue dispatch (sling loop)
for issue in $(bd ready --parent=<epic> | awk '{print $1}' | head -$MAX_POLECATS); do
    gt sling "$issue" <rig>
done

# Manual convoy creation (if needed for grouping)
gt convoy create "Wave N" <issue1> <issue2> ...
```

**Token cost**: ~50 tokens per dispatch (command + confirmation)

### MONITOR Phase

**Purpose**: Track in-flight work without consuming context.

**Inputs**: Active convoys, poll interval

**Outputs**: Status updates, completion/failure events

**Monitoring strategy**:

```bash
# Primary: Convoy dashboard (LOWEST token cost)
gt convoy status <convoy-id>
# Output: ~100 tokens with all polecat statuses

# Secondary: Individual polecat check (if convoy unclear)
gt polecat status <rig>/<name>
# Output: ~50 tokens

# Tertiary: Peek at polecat work (debugging only)
tmux capture-pane -t gt-<rig>-<polecat> -p | tail -20
# Output: ~200 tokens - use sparingly
```

**Poll interval**: 30 seconds

**Why 30 seconds?**
- Too fast (5s): Wastes tokens, polecats need time
- Too slow (5m): Delayed failure detection, idle capacity
- 30s: Balance between responsiveness and efficiency

**Status interpretation**:

| Convoy Status | Meaning | Action |
|---------------|---------|--------|
| `running` | Polecats working | Continue monitoring |
| `partial` | Some complete, some running | Collect completed, continue |
| `complete` | All convoy issues done | Move to COLLECT |
| `failed` | One or more failed | Collect successes, RETRY failures |
| `stalled` | No progress for N polls | Investigate, possibly nudge |

**Stall detection**:

```python
def detect_stall(convoy_id, history):
    """Convoy is stalled if no status change in 5 polls (2.5 min)."""
    recent = history[-5:]
    if len(set(recent)) == 1:  # All same status
        return True
    return False
```

### COLLECT Phase

**Purpose**: Verify completions, update tracking, harvest results.

**Inputs**: Completion events from MONITOR

**Outputs**: Verified completions, updated epic state

**Collection steps**:

```bash
# 1. Verify beads status actually changed
bd show <issue> | grep "status: closed"

# 2. Verify git work product exists
git -C ~/gt/<rig>/polecats/<polecat> log -1 --oneline

# 3. Check for validation artifacts
ls ~/gt/<rig>/polecats/<polecat>/.agents/validations/

# 4. Update local tracking
# (internal state management)
```

**Collection validation**:

```python
def verify_completion(issue, polecat):
    """Verify issue was actually completed, not just abandoned."""

    # Check beads
    status = bd_show(issue).status
    if status != 'closed':
        return False, "Status not closed"

    # Check git
    commits = git_log(polecat_path, count=1)
    if not commits:
        return False, "No commits found"

    # Check commit message references issue
    if issue not in commits[0].message:
        return False, "Commit doesn't reference issue"

    return True, "Verified"
```

**Post-collection cleanup**:

```bash
# After successful collection, polecat can be:
# Option A: Reset for reuse
gt polecat reset <rig>/<name>

# Option B: Leave for later GC
# (handled by gt polecat gc after merge)
```

### RETRY Phase

**Purpose**: Handle failures with backoff and escalation.

**Inputs**: Failed issues, retry counts

**Outputs**: Scheduled retries, escalation actions

**Retry policy**:

| Attempt | Backoff | Action |
|---------|---------|--------|
| 1 | 30s | Re-sling to fresh polecat |
| 2 | 60s | Re-sling with nudge context |
| 3 | 120s | Re-sling with explicit hints |
| 4+ | - | Escalate: BLOCKER + mail |

**Backoff calculation**:

```python
def calculate_backoff(attempt):
    """Exponential backoff: 30s * 2^(attempt-1)"""
    return 30 * (2 ** (attempt - 1))

# attempt 1: 30s
# attempt 2: 60s
# attempt 3: 120s
```

**Retry execution**:

```bash
# Standard retry
gt sling <issue> <rig>

# Retry with context (after 2nd failure)
bd comments add <issue> "Previous attempt failed: <reason>. Try: <hint>"
gt sling <issue> <rig>
```

**Escalation (after MAX_RETRIES)**:

```bash
# Mark as blocker
bd update <issue> --labels=BLOCKER

# Add failure context
bd comments add <issue> "AUTO-ESCALATED: Failed 3 attempts.
Reasons: 1) <reason1> 2) <reason2> 3) <reason3>
Human review required."

# Mail human (--human, not mayor/ since we ARE mayor)
gt mail send --human -s "BLOCKER: <issue> failed 3 attempts" -m "..."

# Continue with other issues (don't halt epic)
```

## State Machine

```
                    ┌─────────────────────────────────────┐
                    │                                     │
                    ▼                                     │
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐ │
│ OBSERVE │───►│DISPATCH │───►│ MONITOR │───►│ COLLECT │─┘
└─────────┘    └─────────┘    └─────────┘    └─────────┘
     │              │              │              │
     │              │              │              │
     │              │              ▼              │
     │              │         ┌─────────┐        │
     │              │         │  RETRY  │────────┘
     │              │         └─────────┘
     │              │              │
     │              │              │ (schedule)
     │              │              ▼
     │              │         ┌─────────┐
     │              └────────►│  QUEUE  │
     │                        └─────────┘
     │
     │ (complete=true)
     ▼
┌─────────┐
│  EXIT   │
└─────────┘
```

## Loop Invariants

Invariants that must hold throughout execution:

1. **Progress**: Each loop iteration must make progress OR escalate
2. **Bounded**: Retry counts are bounded, escalation is guaranteed
3. **Idempotent**: Re-running OBSERVE produces same state for same beads
4. **Recoverable**: State can be reconstructed from beads alone

## Concurrency Model

**Single Mayor, Multiple Polecats**:

```
Mayor (ODMCR Loop)
    │
    ├── Polecat 1 (working on gt-0101)
    ├── Polecat 2 (working on gt-0102)
    ├── Polecat 3 (working on gt-0103)
    └── Polecat 4 (working on gt-0104)
```

**Coordination via beads**:
- Mayor updates issue status via `bd update`
- Polecats work independently
- Status synced via `bd sync` on both sides

**Race conditions avoided by**:
- Single Mayor making all dispatch decisions
- Polecats only touch their assigned issues
- Beads provides atomic status updates

## Token Budget

Per ODMCR iteration (30s):

| Phase | Tokens | Notes |
|-------|--------|-------|
| OBSERVE | ~300 | bd queries |
| DISPATCH | ~100 | gt sling commands |
| MONITOR | ~100 | convoy status |
| COLLECT | ~150 | verification |
| RETRY | ~100 | if failures |
| **Total** | ~750 | per iteration |

**Per hour**: ~90,000 tokens (120 iterations)
**Per 8-hour run**: ~720,000 tokens

This is sustainable for long-running autonomous execution.

## Error Recovery

**Mayor session crash**:
```bash
# State is in beads, not memory
# Simply restart crank
/crank <epic> <rig>  # Resumes from beads state
```

**Polecat orphaned**:
```bash
# Detected by stall detection
# Resolution: nuke and re-dispatch
gt polecat nuke <rig>/<name> --force
gt sling <issue> <rig>
```

**Beads sync conflict**:
```bash
# Append-only design makes this rare
# Resolution: theirs wins for issues.jsonl
git checkout --theirs .beads/issues.jsonl
git add .beads/issues.jsonl
bd sync
```

## Tuning Parameters

| Parameter | Default | Tuning Guidance |
|-----------|---------|-----------------|
| `MAX_POLECATS` | 4 | Increase for large epics, decrease for complex issues |
| `POLL_INTERVAL` | 30s | Decrease for fast issues, increase to save tokens |
| `MAX_RETRIES` | 3 | Increase for flaky tests, decrease for clean codebases |
| `BACKOFF_BASE` | 30s | Increase for rate-limited APIs |
| `STALL_THRESHOLD` | 5 polls | Decrease for tight deadlines |
