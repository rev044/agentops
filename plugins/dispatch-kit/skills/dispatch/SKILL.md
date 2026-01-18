---
name: dispatch
description: >
  Unified dispatch for Gas Town work assignment. Covers gt sling (assign work),
  gt hook (attach work), gt convoy (batch tracking), and the Propulsion Principle.
  This is THE skill for understanding how to dispatch work to agents.
version: 1.0.0
context-budget:
  skill-md: 7KB
  references-total: 20KB
  typical-session: 10KB
triggers:
  - "dispatch work"
  - "sling"
  - "gt sling"
  - "assign to polecat"
  - "hook work"
  - "gt hook"
  - "convoy"
  - "gt convoy"
  - "create convoy"
  - "track work"
  - "propulsion principle"
  - "how do I assign work"
  - "batch dispatch"
  - "slingable bead"
allowed-tools: Bash, Read, Glob, Grep
---

# dispatch - Unified Work Dispatch

The definitive skill for dispatching work in Gas Town.

> **Core Principle**: If it's on your hook, YOU RUN IT.

## Overview

**What this skill covers:** How to assign, hook, and track work across Gas Town agents.

| User Says | Claude Does |
|-----------|-------------|
| "dispatch this to a polecat" | `gt sling <bead> <rig>` |
| "hook this work" | `gt hook <bead>` |
| "create a convoy" | `gt convoy create "name" <beads>` |
| "how do I sling?" | Explain dispatch patterns |
| "batch dispatch these issues" | Multiple `gt sling` or convoy create |

---

## Quick Reference

```bash
# THE main dispatch command
gt sling <bead> <target>           # Assign work + start immediately

# Hook (attach without starting)
gt hook <bead>                     # Attach to your hook

# Convoy (batch tracking)
gt convoy create "name" <beads>    # Create tracked batch
gt convoy list                     # Dashboard view
gt convoy status <id>              # Progress details
```

---

## The Three Dispatch Commands

### 1. gt sling - THE Dispatch Command

**Use this when:** Assigning work to any agent (including yourself).

```bash
gt sling gt-abc daedalus           # Auto-spawn polecat in rig (Olympian)
gt sling gt-abc daedalus/Toast     # Specific polecat
gt sling gt-abc crew               # Crew worker
gt sling gt-abc mayor              # Mayor
gt sling gt-abc                    # Self (current agent)
```

**Key behaviors:**
- Auto-spawns polecats when target is a rig
- Auto-creates convoy for dashboard visibility
- Hooks work and starts immediately (Propulsion Principle)

### 2. gt hook - Attach Without Starting

**Use this when:** You want to attach work but not trigger immediate execution.

```bash
gt hook <bead>                     # Attach to your hook
gt hook status                     # Show what's hooked
gt unsling                         # Remove from hook
```

**Key behaviors:**
- Work survives session restarts
- SessionStart hook finds attached work
- Does NOT start new session

### 3. gt convoy - Batch Tracking

**Use this when:** Tracking multiple related issues together.

```bash
gt convoy create "Wave 1" gt-abc gt-def gt-ghi
gt convoy add <convoy-id> gt-xyz   # Add more issues
gt convoy list                     # Dashboard
gt convoy status <id>              # Details
```

**Key behaviors:**
- Auto-closes when all tracked issues complete
- Cross-prefix capable (convoy in hq-* tracks gt-*, ap-*, etc.)
- Non-blocking (tracked issues don't block convoy)

---

## Dispatch Patterns

### Single Issue Dispatch

```bash
# Most common: dispatch to a rig
gt sling gt-abc daedalus

# What happens:
# 1. Creates polecat worktree: ~/gt/daedalus/polecats/<name>/
# 2. Starts tmux session with Claude Code
# 3. Hooks issue to polecat
# 4. SessionStart → polecat finds work → executes (Propulsion)
# 5. Auto-creates convoy for tracking
```

### Batch Dispatch (Parallel)

```bash
# Multiple beads to same rig
gt sling gt-abc gt-def gt-ghi daedalus

# Or create convoy first, then sling individually
gt convoy create "Feature X" gt-abc gt-def gt-ghi
gt sling gt-abc daedalus
gt sling gt-def daedalus
gt sling gt-ghi daedalus
```

### Multi-Rig Dispatch

```bash
# Each issue goes to correct rig based on bead prefix
gt sling at-123 athena         # Athena (Knowledge) changes
gt sling gt-456 daedalus       # Daedalus (Coordination) changes
gt sling ar-789 argus          # Argus (Observation) changes
```

### Self-Dispatch

```bash
# Hook work for yourself (current session)
gt sling gt-abc                # Hooks and continues
gt hook gt-abc                 # Just hooks (no immediate action)
```

---

## Creating Slingable Beads

**HQ beads (`hq-*`) CANNOT be hooked by polecats!**

`gt sling` uses `bd update` which requires beads in the target rig's database.

| Work Type | Create From | Gets Prefix | Slingable? |
|-----------|-------------|-------------|------------|
| Mayor coordination | `~/gt` | `hq-*` | No |
| Rig bug/feature | Rig's beads | `gt-*`, `ap-*`, etc. | Yes |

**From Mayor, create slingable beads:**

```bash
# Target the rig's beads database
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="Fix X" --type=bug
# Creates: gt-xxxxx (slingable!)

# Then sling normally
gt sling gt-xxxxx daedalus
```

---

## The Propulsion Principle

> **If you find work on your hook, YOU RUN IT.**

Gas Town is a steam engine. Agents are pistons. The system runs when agents
execute immediately upon finding hooked work.

**Startup behavior:**
1. Check hook (`gt hook`)
2. Work hooked? → **EXECUTE IMMEDIATELY** (no confirmation)
3. Hook empty? → Check mail, then wait

**Why it matters:**
- No supervisor polling for status
- The hook IS your assignment
- Waiting stalls the entire system

**Reference:** `references/propulsion.md`

---

## Gotchas

### Wrong prefix for rig work
- WRONG: `bd create --title="daedalus bug"` → `hq-xxx` (can't sling)
- RIGHT: `BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create ...` → `gt-xxx`

### Temporal language inverts dependencies
- WRONG: `bd dep add phase1 phase2` (temporal: "1 before 2")
- RIGHT: `bd dep add phase2 phase1` (requirement: "2 needs 1")

### No output to orchestrator
`gt sling` returns immediately. Polecat output stays in its tmux session.
Use `gt convoy status` to monitor, not wait for command output.

---

## Compare: sling vs hook vs handoff

| Command | Hooks Work | Starts Session | Keeps Context |
|---------|------------|----------------|---------------|
| `gt sling` | Yes | Yes | Yes |
| `gt hook` | Yes | No | Yes |
| `gt handoff` | Yes | Yes (new) | No (fresh) |

---

## References

Load JIT when needed:

| Reference | When to Load |
|-----------|--------------|
| `references/gt-sling.md` | Full sling documentation |
| `references/gt-hook.md` | Hook semantics and lifecycle |
| `references/gt-convoy.md` | Convoy creation and tracking |
| `references/propulsion.md` | The Propulsion Principle philosophy |

---

## See Also

- `/gastown` - Status checks, monitoring, polecat management
- `/crank` - Autonomous epic execution
- `bd ready` - Find unblocked work to dispatch
