---
name: polecat-lifecycle
description: >
  Polecat worker lifecycle management. Covers spawning polecats (gt sling),
  resetting them for reuse, nuking completed workers, and garbage collecting
  stale branches. For Witness and Mayor roles primarily.
version: 1.0.0
triggers:
  - "spawn polecat"
  - "reset polecat"
  - "nuke polecat"
  - "polecat gc"
  - "gt polecat"
  - "destroy polecat"
  - "cleanup polecat"
  - "stale polecat"
  - "polecat stuck"
  - "polecat lifecycle"
  - "remove polecat"
  - "add polecat"
  - "create polecat"
allowed-tools: Bash, Read
---

# polecat-lifecycle - Polecat Worker Lifecycle

Manage the full lifecycle of polecat worker agents in Gas Town.

> **For Witness and Mayor roles.** Polecats don't manage their own lifecycle.

## Overview

Polecats are worker agents that operate in isolated git worktrees. They spawn,
work, and get cleaned up when done.

| User Says | Claude Does |
|-----------|-------------|
| "spawn a polecat for this" | `gt sling <bead> <rig>` (auto-spawns) |
| "nuke that polecat" | `gt polecat nuke <rig>/<name> --force` |
| "clean up stale branches" | `gt polecat gc <rig>` |
| "is this polecat stuck?" | `gt polecat status <rig>/<name>` |
| "list polecats" | `gt polecat list <rig>` |

---

## Quick Reference

```bash
# Spawn (via sling - auto-creates polecat)
gt sling <bead> <rig>                # Auto-spawn + dispatch

# Manual add (rarely needed)
gt polecat add <rig> <name>          # Create polecat worktree

# Status
gt polecat list <rig>                # List all polecats
gt polecat status <rig>/<name>       # Detailed status

# Cleanup
gt polecat nuke <rig>/<name>         # Destroy (with safety checks)
gt polecat nuke <rig>/<name> --force # Destroy (bypass safety)
gt polecat gc <rig>                  # Clean orphaned branches

# Diagnostics
gt polecat stale <rig>               # Find stale polecats
gt polecat git-state <rig>/<name>    # Pre-nuke verification
gt polecat check-recovery <rig>/<name>  # Recovery needed?
```

---

## Lifecycle Stages

```
┌─────────┐    gt sling     ┌─────────┐    work done    ┌─────────┐
│ (none)  │ ─────────────→  │ working │ ─────────────→  │  done   │
└─────────┘                 └─────────┘                 └─────────┘
     ↑                           │                          │
     │                           │ stuck/error              │
     │                           ↓                          │
     │                      ┌─────────┐                     │
     │                      │  stuck  │                     │
     │                      └─────────┘                     │
     │                           │                          │
     │      gt polecat nuke      │                          │
     └───────────────────────────┴──────────────────────────┘
```

### Stage: Spawning

Polecats spawn via `gt sling`:

```bash
gt sling gt-abc daedalus
# 1. Creates worktree: ~/gt/daedalus/polecats/<name>/
# 2. Creates branch: polecat/<name>-<timestamp>
# 3. Starts tmux session
# 4. Hooks work to polecat
# 5. SessionStart triggers execution
```

**Reference:** `references/spawn.md`

### Stage: Working

Polecat executes autonomously (Propulsion Principle):

- Check status: `gt polecat status <rig>/<name>`
- Peek at work: `tmux capture-pane -t gt-<rig>-<name> -p | tail -20`
- Monitor progress: `gt convoy status <id>`

### Stage: Done

After polecat completes:

1. Branch pushed to origin
2. Bead closed
3. Ready for cleanup

### Stage: Stuck

Polecats can get stuck. Signs:

- No activity in tmux session
- Hit usage limit
- Error in execution
- Waiting for something that won't happen

**Reference:** `references/troubleshooting.md`

---

## Spawning Polecats

### Auto-Spawn via Sling (Preferred)

```bash
gt sling <bead> <rig>
```

`gt sling` automatically:
- Creates a new polecat if needed
- Picks a unique name
- Creates timestamped branch
- Starts tmux session
- Hooks the work

### Manual Add (Rare)

```bash
gt polecat add <rig> <name>
```

Only use when:
- Pre-staging polecats before work
- Creating without immediate work
- Testing polecat infrastructure

**Reference:** `references/spawn.md`

---

## Cleanup Operations

### Nuke - Complete Destruction

```bash
gt polecat nuke <rig>/<name>           # With safety checks
gt polecat nuke <rig>/<name> --force   # Bypass safety
gt polecat nuke <rig> --all            # All polecats in rig
gt polecat nuke <rig> --all --dry-run  # Preview
```

**Safety checks prevent nuking if:**
- Unpushed/uncommitted changes
- Open merge request
- Work still on hook

Use `--force` to bypass (LOSES WORK).

**Reference:** `references/nuke.md`

### GC - Branch Cleanup

```bash
gt polecat gc <rig>
gt polecat gc <rig> --dry-run
```

Removes orphaned branches:
- Branches for non-existent polecats
- Old timestamped branches

**Reference:** `references/gc.md`

### Stale Detection

```bash
gt polecat stale <rig>
gt polecat stale <rig> --cleanup
gt polecat stale <rig> --threshold 50
```

A polecat is stale if:
- No active tmux session
- Way behind main (>20 commits) OR no agent bead
- No uncommitted work

---

## Resetting Polecats

**Current approach:** Nuke and re-spawn.

```bash
# Clean slate for reuse
gt polecat nuke <rig>/<name> --force
gt sling <new-bead> <rig>
```

The nuke+sling pattern is the standard way to recycle a polecat slot.

**Reference:** `references/reset.md`

---

## Diagnostics

### Status Check

```bash
gt polecat status <rig>/<name>
gt polecat status <rig>/<name> --json
```

Shows:
- Lifecycle state (working/done/stuck/idle)
- Assigned issue
- Session status
- Last activity

### Git State (Pre-Nuke)

```bash
gt polecat git-state <rig>/<name>
```

Shows exactly what would be lost if nuked.

### Recovery Check

```bash
gt polecat check-recovery <rig>/<name>
```

Determines if polecat needs recovery vs safe to nuke.

---

## Common Workflows

### Post-Merge Cleanup

```bash
# After work merged to main
gt polecat nuke <rig>/<name>           # Destroy worker
gt polecat gc <rig>                    # Clean branches
```

### Batch Cleanup

```bash
# Nuke all in a rig
gt polecat nuke <rig> --all --dry-run  # Preview first!
gt polecat nuke <rig> --all --force    # Execute
```

### Stuck Polecat Recovery

```bash
# 1. Check status
gt polecat status <rig>/<name>

# 2. Check git state
gt polecat git-state <rig>/<name>

# 3. Decide: recover or nuke?
gt polecat check-recovery <rig>/<name>

# 4. If safe, nuke
gt polecat nuke <rig>/<name> --force
```

---

## References

Load JIT when needed:

| Reference | When to Load |
|-----------|--------------|
| `references/spawn.md` | Spawning patterns and options |
| `references/reset.md` | Resetting polecats for reuse |
| `references/nuke.md` | Destruction and safety checks |
| `references/gc.md` | Branch cleanup |
| `references/troubleshooting.md` | Stuck polecats, usage limits |

---

## See Also

- `/dispatch` - Work assignment (sling, hook, convoy)
- `/roles` - Understanding polecat vs other roles
- `/gastown` - Overall Gas Town operations
- `gt polecat --help` - Command reference
