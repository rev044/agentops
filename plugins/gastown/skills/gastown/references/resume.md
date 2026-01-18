# Session Continuity

## Overview

When orchestrator context fills up, save state to beads and hook for the next session.

## Context Monitoring

Check context usage periodically:

| Usage | Action |
|-------|--------|
| < 30% | Continue normally |
| 30-35% | Prepare for handoff (save state) |
| > 35% | Execute handoff immediately |

## Saving State

Before handoff, save orchestration state:

```bash
# Build state
state=$(cat <<EOF
{
  "wave": 2,
  "completed": ["gt-abc", "gt-def"],
  "in_progress": ["gt-ghi", "gt-jkl"],
  "convoy": "hq-cv-xyz",
  "rig": "gastown",
  "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)

# Save to epic
bd comments add $epic_id "GASTOWN_STATE: $state"

# Attach to hook
gt hook $epic_id
```

## Requesting Handoff

Signal that fresh session needed:

```bash
# Attach work to hook
gt hook $epic_id

# Output handoff message
echo "HANDOFF_REQUIRED: Context at 35%"
echo "To continue: /gastown resume"
echo "Or: /gastown $epic_id --resume"
```

## Resume Protocol

New session finds hooked work:

```python
def resume_gastown():
    # Check hook
    hook = bash("gt hook")

    if not hook:
        print("No hooked work found. Nothing to resume.")
        return

    epic_id = hook.strip()

    # Read saved state
    comments = bash(f"bd comments {epic_id}")
    state_line = find_last_match(comments, "GASTOWN_STATE:")

    if not state_line:
        # No saved state - start fresh
        print(f"No saved state for {epic_id}. Starting fresh.")
        return execute_epic(epic_id)

    state = json.loads(state_line.split("GASTOWN_STATE:")[1])

    # Check convoy for in-progress wave
    if state["convoy"]:
        convoy_status = bash(f"gt convoy status {state['convoy']}")

        # Parse progress
        match = re.search(r"Progress: (\d+)/(\d+)", convoy_status)
        if match and match.group(1) == match.group(2):
            # Wave completed while we were away
            print("Wave completed during handoff")
            state["completed"].extend(state["in_progress"])
            state["in_progress"] = []
            state["wave"] += 1

    # Continue from saved state
    print(f"Resuming {epic_id} from wave {state['wave']}")
    return continue_from_wave(epic_id, state)
```

## State Fields

| Field | Purpose |
|-------|---------|
| `wave` | Current wave number (1-indexed) |
| `completed` | List of completed issue IDs |
| `in_progress` | List of currently executing issue IDs |
| `convoy` | Current convoy ID |
| `rig` | Target rig |
| `started_at` | Timestamp |

## Multiple Handoffs

Each handoff appends new state:

```bash
# Session 1
bd comments $epic_id | grep "GASTOWN_STATE:"
# → GASTOWN_STATE: {..., wave: 1}

# Session 2
bd comments $epic_id | grep "GASTOWN_STATE:"
# → GASTOWN_STATE: {..., wave: 1}
# → GASTOWN_STATE: {..., wave: 2}

# Always use LAST state
bd comments $epic_id | grep "GASTOWN_STATE:" | tail -1
```

## Emergency Recovery

If handoff fails or state corrupted:

```bash
# Check what's actually completed
bd list --status=closed | grep -E "gt-abc|gt-def|..."

# Check what's in progress
bd list --status=in_progress | grep -E "..."

# Rebuild state from beads (source of truth)
# Beads always has the real state
```

## Integration with Daemon

Daemon preserves hooks across restarts:

```
Orchestrator crash
    ↓
Hook persists in beads
    ↓
Daemon restarts any crashed polecats
    ↓
New orchestrator session
    ↓
gt hook → Find epic
    ↓
Resume from saved state
```

**No work lost.** The hook IS your assignment, survives everything.

## Handoff Checklist

```bash
# Before handoff:
[ ] Save state: bd comments add $epic "GASTOWN_STATE: ..."
[ ] Attach hook: gt hook $epic
[ ] Log: bd comments add $epic "HANDOFF: reason"

# After resume:
[ ] Check hook: gt hook
[ ] Read state: bd comments $epic | grep GASTOWN_STATE
[ ] Check convoy: gt convoy status
[ ] Continue: execute remaining waves
```

## Example Flow

```
Session 1:
├─ Start epic gt-tq9
├─ Execute Wave 1 ✓
├─ Execute Wave 2 ✓
├─ Context at 35%
├─ Save state: {wave: 3, completed: [...], ...}
├─ Hook epic: gt hook gt-tq9
└─ Output: "HANDOFF_REQUIRED"

Session 2:
├─ gt hook → gt-tq9
├─ Read state: wave 3
├─ Check convoy → Wave 2 complete
├─ Execute Wave 3
├─ Execute Wave 4
├─ Close epic
└─ Done!
```
