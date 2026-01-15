# Session Continuity: What Survives a Restart

Understanding what persists across agent session boundaries.

## Overview

Agent sessions are ephemeral. When a session ends (handoff, crash, context limit),
a new session starts fresh with no memory of the previous conversation.

**The key to continuity is external state.**

## What Survives

| Persists | Storage | Access |
|----------|---------|--------|
| Hooked work | Beads | `gt hook` |
| Issue state | Beads | `bd show`, `bd list` |
| Issue comments | Beads | `bd comments <id>` |
| Git commits | Git | `git log` |
| Mail messages | Beads (mail) | `gt mail inbox` |
| CLAUDE.md files | Filesystem | Auto-loaded |
| Config files | Filesystem | `gt config` |

## What Does NOT Survive

| Lost | Why | Mitigation |
|------|-----|------------|
| Conversation history | Not persisted | Write key decisions to beads |
| In-memory context | Session-local | Save state before handoff |
| Todo list | Session-local | Use beads for tracking |
| Uncommitted changes | Not saved | Commit before handoff |
| Local variables | Memory only | Write to comments |

## The Hook as Anchor

The hook is the primary continuity mechanism:

```bash
# Before handoff
gt hook gt-abc

# After restart (new session)
gt hook
# → gt-abc (still there!)
```

Hooks persist in beads, which is in git. They survive everything.

## State Persistence Patterns

### For Simple Work

Just hook the bead:

```bash
gt hook gt-abc
gt handoff
```

Next session:
```bash
gt hook       # → gt-abc
bd show gt-abc  # Full context
```

### For Complex Orchestration

Save state to bead comments:

```bash
# Build state
state='{"wave": 2, "completed": ["gt-1", "gt-2"], "in_progress": ["gt-3"]}'

# Save to bead
bd comments add gt-epic "ORCHESTRATOR_STATE: $state"

# Hook and handoff
gt hook gt-epic
gt handoff
```

Next session:
```bash
# Retrieve state
bd comments gt-epic | grep "ORCHESTRATOR_STATE:" | tail -1
```

### For Ad-Hoc Context

Use mail-to-self:

```bash
gt mail send --self -s "HANDOFF: Notes" -m "Context here"
gt hook attach <mail-id>
gt handoff
```

## Context Recovery Flow

What happens when a new session starts:

```
1. SessionStart hook fires
   └─ Runs: gt prime

2. gt prime executes
   ├─ Checks: gt hook
   ├─ Checks: gt mail inbox
   └─ Loads: CLAUDE.md context

3. If work hooked
   ├─ Propulsion Principle applies
   └─ Work begins immediately

4. If no hook
   ├─ Check mail for instructions
   └─ Wait for user direction
```

## What gt prime Does

```bash
gt prime
```

1. Outputs hooked work (if any)
2. Shows unread mail count
3. Lists ready beads
4. Displays any urgent notifications

This gives the new session immediate context without manual exploration.

## Beads as Source of Truth

When in doubt, check beads:

```bash
# What's actually completed?
bd list --status=closed

# What's in progress?
bd list --status=in_progress

# What's blocked?
bd blocked

# Full context on specific work
bd show gt-abc
```

Beads always has the real state. If your saved state disagrees with beads,
trust beads.

## CLAUDE.md Files

These load automatically every session:

| File | Scope | Contains |
|------|-------|----------|
| `~/.claude/CLAUDE.md` | Global | Universal rules, shell config |
| `<project>/CLAUDE.md` | Project | Project-specific context |
| `<rig>/crew/*/CLAUDE.md` | Workspace | Role-specific context |

Put persistent instructions here, not in conversation.

## Emergency Recovery

If handoff fails and you lose context:

```bash
# 1. Check what's actually in beads
bd list --status=in_progress
bd list --status=open

# 2. Check for recent commits
git log --oneline -10

# 3. Check mail for clues
gt mail inbox

# 4. Look for state comments
bd comments <suspected-work-id> | grep -E "STATE|HANDOFF"

# 5. Rebuild mentally and continue
```

## Best Practices

### Before Any Handoff

```bash
[ ] Commit code changes
[ ] Save state to bead comments (if complex)
[ ] Hook work: gt hook <id>
[ ] Sync beads: bd sync
```

### Keep Comments Clean

Use consistent prefixes for state:

```bash
# Orchestration state
bd comments add $id "ORCHESTRATOR_STATE: {...}"

# Manual notes
bd comments add $id "NOTE: Paused due to blocker"

# Handoff context
bd comments add $id "HANDOFF: Context for next session"
```

### Don't Over-Persist

Not everything needs to survive:
- Exploratory thinking → Let it go
- Dead-end investigations → Let them go
- Routine operations → Don't log every step

Persist decisions, state, and actionable context.

## See Also

- `gt hook` - Hook management
- `gt prime` - Context recovery
- `bd comments` - State storage
- `gt handoff` - Session handoff
