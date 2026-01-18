---
name: sk-crew
description: >
  Crew workspace management. Understanding and managing persistent human-guided
  workspaces in Gas Town. Covers lifecycle, workspace operations, and when to use
  crew vs polecats.
version: 1.0.0
triggers:
  - "crew workspace"
  - "create crew"
  - "crew vs polecat"
  - "persistent workspace"
  - "add crew"
  - "start crew"
  - "stop crew"
  - "crew refresh"
  - "crew pristine"
  - "where should I work"
  - "human-managed workspace"
  - "gt crew"
allowed-tools: Bash, Read
---

# sk-crew - Crew Workspace Management

Manage persistent, human-guided workspaces in Gas Town.

## Overview

**Crew workspaces** are persistent developer environments within a rig. Unlike polecats
(ephemeral, autonomous workers), crew members are human-managed and long-lived.

| Aspect | Crew | Polecat |
|--------|------|---------|
| **Persistence** | Long-lived | Ephemeral |
| **Management** | Human-guided | Witness-managed |
| **Hook behavior** | Show, await confirm | Auto-execute |
| **Scope** | Flexible, multi-issue | Single issue |
| **Identity** | Named (dave, emma) | Auto-generated |

**Use crew when**: Interactive development with human oversight is needed.
**Use polecat when**: Autonomous parallel execution is needed.

---

## Quick Reference

```bash
# Lifecycle
gt crew add <name>                # Create workspace
gt crew start <name>              # Start session
gt crew stop <name>               # Stop session
gt crew remove <name>             # Delete workspace

# Workspace Management
gt crew pristine [<name>]         # Sync with remote
gt crew refresh <name>            # Context cycle with handoff
gt crew rename <old> <new>        # Rename workspace

# Status
gt crew list                      # List all crew
gt crew status [<name>]           # Detailed status
gt crew at <name>                 # Attach to session
```

---

## When to Use Crew vs Polecat

| Scenario | Use |
|----------|-----|
| Interactive development with human | Crew |
| Parallel autonomous execution | Polecat |
| Long-running project work | Crew |
| Single-issue batch processing | Polecat |
| Exploratory/research work | Crew |
| Epic wave execution | Polecat |

**Key difference**: Crew waits for human direction. Polecats auto-execute hooked work.

---

## Crew Behavior

Crew members operate under the Propulsion Principle but with human confirmation:

```
1. Check hook (gt hook)
2. If work hooked → Show human, await confirmation
3. If not hooked → Wait for human instructions
```

**Unlike polecats**, crew does NOT auto-execute. The human may want to discuss or
modify the approach first.

---

## Beads Integration

Crew workspaces share the rig's beads database:

- **Location**: `<rig>/crew/<name>/.beads/` (symlinked or independent)
- **Sync**: Use `bd sync` to commit beads changes
- **Routing**: Issue prefixes route automatically (e.g., `gt-*` → gastown beads)

When multiple crew members work in the same rig, beads sync keeps them coordinated.

---

## Common Operations

### Create and Start Working

```bash
# Create workspace
gt crew add dave --rig gastown

# Start session
gt crew start gastown dave

# Or start all crew in a rig
gt crew start gastown
```

### End of Day

```bash
# Commit and push your work
git add .
git commit -m "type(scope): description"
git push -u origin HEAD

# Sync beads
bd sync

# Stop session
gt crew stop dave
```

### Context Refresh

When context gets stale or a fresh start is needed:

```bash
# Refresh with handoff (preserves continuity)
gt crew refresh dave -m "Working on gt-123, tests passing"

# Or pristine (sync with remote, no handoff)
gt crew pristine dave
```

---

## See Also

- `sk-roles` - Role responsibilities (Crew vs Mayor vs Polecat)
- `sk-gastown` - Gas Town status and utility operations
- `sk-dispatch` - Work dispatch patterns

### References

- `references/lifecycle.md` - Create, start, stop, remove crew
- `references/crew-vs-polecat.md` - When to use which
- `references/workspace-management.md` - Pristine, refresh, rename
- `references/crew-beads.md` - Shared beads database
