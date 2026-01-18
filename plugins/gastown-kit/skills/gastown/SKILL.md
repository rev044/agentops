---
name: gastown
description: >
  Gas Town status and utility operations. Check convoy progress, peek into polecats,
  nudge stuck workers, cleanup completed work. For autonomous execution, use /crank.
version: 2.0.0
context: fork
context-budget:
  skill-md: 5KB
  references-total: 34KB
  typical-session: 15KB
triggers:
  - "gastown status"
  - "what polecats are running"
  - "convoy status"
  - "check convoy"
  - "polecat stuck"
  - "peek polecat"
  - "nudge polecat"
  - "cleanup polecats"
  - "dispatch work"
  - "sling issue"
  - "how do I sling"
  - "create slingable bead"
  - "propulsion principle"
  - "hooked work"
allowed-tools: Bash, Read, Glob, Grep
---

# gastown - Gas Town Status & Utility Skill

Monitor and manage Gas Town polecats and convoys.

> **For autonomous epic execution, use `/crank` instead.**
> This skill is for status checks, debugging, and utility operations.

## Overview

**What this skill does:** Provides visibility into Gas Town operations and utility commands for managing workers.

| User Says | Claude Does |
|-----------|-------------|
| "what's the status?" | `gt convoy list`, `gt polecat list` |
| "check convoy X" | `gt convoy status <id>` |
| "something's stuck" | Peek into polecat, diagnose |
| "nudge that polecat" | Send continue signal |
| "cleanup finished work" | `gt polecat gc`, cleanup branches |

**For dispatching work:** Use `gt sling` (single issue) or `/crank` (autonomous epic execution with parallel waves).

---

## Operations

### 1. Status Check

Get overview of active work:

```bash
# All active convoys (dashboard view)
gt convoy list

# Specific convoy progress
gt convoy status <convoy-id>

# Polecats in a rig
gt polecat list <rig>

# All polecats across town
gt polecat list
```

### 2. Peek into Polecats

See what a polecat is doing (without returning output to your context):

```bash
# View last N lines of polecat session
tmux capture-pane -t gt-<rig>-<polecat> -p | tail -20

# Full session dump (use sparingly - large output)
tmux capture-pane -t gt-<rig>-<polecat> -p -S -

# Check if polecat is responding
gt polecat status <rig>/<polecat>
```

**Common patterns to look for:**
- "You've hit your limit" - needs `/login` reset
- Stuck on confirmation - may need nudge
- Error messages - may need intervention

### 3. Nudge Stuck Polecats

Send a message to a polecat that seems stuck:

```bash
# Send continue signal
tmux send-keys -t gt-<rig>-<polecat> "continue with your task" Enter

# More specific nudge
tmux send-keys -t gt-<rig>-<polecat> "focus on completing the current issue" Enter
```

### 4. Cleanup Operations (Simplified in v0.2.5+)

**Self-cleaning polecats (v0.2.5+):** Polecats now self-nuke on completion via `gt done`.
No manual cleanup needed in most cases.

```bash
# Polecats self-clean automatically - just close the epic
bd close <epic-id> --reason "All children completed"
bd sync

# For stuck polecats that failed to self-clean:
gt polecat nuke <rig>/<polecat> --force

# Garbage collect merged branches (rarely needed now)
gt polecat gc <rig>
```

**Graceful shutdown (v0.2.5+):**
```bash
# Shut down polecats only (preserves witness/refinery)
gt down --polecats
```

---

## Dispatching Work

**For single-issue dispatch:**
```bash
gt sling <issue-id> <rig>
```

**For multiple issues in parallel (manual wave):**
```bash
# Sling each ready issue
for issue in $(bd ready --parent=<epic> | awk '{print $1}'); do
    gt sling "$issue" <rig>
done
```

**For autonomous epic execution:** Use `/crank` instead - it handles:
- Wave computation from dependencies
- Parallel dispatch via sling loop
- Progress monitoring
- Session continuity via hooks

---

## Troubleshooting

| Problem | Diagnosis | Solution |
|---------|-----------|----------|
| Polecat stuck | `tmux capture-pane ...` shows no activity | Nudge with `tmux send-keys` |
| Hit usage limit | See "You've hit your limit" in output | Wait or `/login` reset, then re-dispatch |
| Convoy not updating | `gt convoy status` stale | Check issue status with `bd show` |
| Polecat didn't self-clean | Session still exists after completion | `gt polecat nuke --force` |
| Context getting high | Approaching 40% usage | `gt down --polecats` for graceful wind-down |
| Need to check progress | Want visibility | Peek with `tmux capture-pane` |

---

## Quick Reference

```bash
# Status
gt convoy list                    # Dashboard of active convoys
gt convoy status <id>             # Detailed progress
gt polecat list <rig>             # Polecats in rig

# Peek (see what polecat is doing)
tmux capture-pane -t gt-<rig>-<polecat> -p | tail -20

# Nudge (unstick a polecat)
tmux send-keys -t gt-<rig>-<polecat> "continue" Enter

# Cleanup (v0.2.5+ - polecats self-clean)
bd close <epic-id>                # Just close epic - polecats self-nuke
gt down --polecats                # Graceful shutdown of all polecats
gt polecat nuke <rig>/<polecat> --force  # For stuck polecats only

# Dispatch (for manual use)
gt sling <issue> <rig>            # Single issue (auto-creates convoy)
# For multiple: use sling loop or /crank
```

---

## See Also

- `/crank` - Autonomous epic execution (preferred for full workflows)
- `/status` - Quick status check
- `gt --help` - Full CLI reference

### References

- `references/dispatch.md` - Full dispatch patterns and sling documentation
- `references/propulsion.md` - The Propulsion Principle (autonomous execution philosophy)
- `references/monitoring.md` - Convoy and polecat monitoring
- `references/recovery.md` - Recovery from stuck states
