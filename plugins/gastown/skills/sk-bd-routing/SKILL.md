---
name: sk-bd-routing
description: >
  Two-level beads architecture and routing. Understanding where beads live,
  how prefix routing works, and when to use BEADS_DIR. Critical for Mayor
  understanding which database to use for different work types.
version: 1.0.0
triggers:
  - "beads routing"
  - "two-level beads"
  - "prefix routing"
  - "which database"
  - "town beads"
  - "rig beads"
  - "hq beads"
  - "BEADS_DIR"
  - "beads sync"
  - "slingable bead"
  - "where do beads live"
  - "cross-rig beads"
allowed-tools: Bash, Read, Glob, Grep
---

# sk-bd-routing - Two-Level Beads Routing

Understanding where beads live and how to route commands to the right database.

> **Core Question**: Which database does my bead belong to?

## Overview

**What this skill covers:** The two-level beads architecture in Gas Town and how to route commands correctly.

| User Says | Claude Does |
|-----------|-------------|
| "which database?" | Explain Town vs Rig beads |
| "beads routing" | Show prefix routing table |
| "can't sling hq bead" | Explain BEADS_DIR workaround |
| "two-level beads" | Describe Town/Rig architecture |
| "where do beads live" | Show location hierarchy |

---

## Quick Reference

```bash
# Check what database you're in
pwd                          # Location determines default database
BD_DEBUG_ROUTING=1 bd show <id>  # Debug routing decision

# Target a specific database
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="..."
```

---

## The Two Levels

| Level | Location | Prefix | Purpose | Can Sling? |
|-------|----------|--------|---------|------------|
| **Town** | `~/gt/.beads/` | `hq-*` | Mail, coordination | No |
| **Rig** | `<rig>/crew/*/.beads/` | `gt-*`, `ap-*`, `da-*`, etc. | Project work | Yes |

**The key distinction:**
- Town beads are for Mayor coordination and mail
- Rig beads are for actual project work that polecats execute

---

## Prefix Routing

`bd` routes commands based on bead ID prefix:

```bash
bd show hq-abc    # Routes to ~/gt/.beads/
bd show gt-xyz    # Routes to daedalus beads
bd show ap-123    # Routes to athena beads
```

**Route registration:** `~/gt/.beads/routes.jsonl`

---

## Creating Slingable Beads

**Problem:** HQ beads (`hq-*`) can't be hooked by polecats.

**Solution:** Use `BEADS_DIR` to create beads in the target rig:

```bash
# From Mayor context, create a daedalus bead
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="Fix bug" --type=bug
# Creates: gt-xxxxx (slingable!)

# Then sling it
gt sling gt-xxxxx daedalus
```

---

## Common Patterns

### Creating Work for a Rig (from Mayor)

```bash
# 1. Target the rig's database
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create \
  --title="Feature X" --type=feature --priority=2

# 2. Sling to the rig
gt sling gt-<id> daedalus
```

### Checking Which Database You're Using

```bash
# Option 1: Use debug flag
BD_DEBUG_ROUTING=1 bd list

# Option 2: Check current directory
pwd  # Database determined by location
```

### Syncing Across Worktrees

```bash
# Rig beads use beads-sync branch
bd sync                 # Full sync
bd sync --from-main     # Pull from main (ephemeral branches)
```

---

## Gotchas

### Wrong Prefix for Rig Work

- **WRONG:** `bd create --title="gastown bug"` from `~/gt`
  - Creates `hq-xxx` (polecats can't hook)
- **RIGHT:** Use `BEADS_DIR` to target rig's database
  - Creates `gt-xxx` (slingable)

### Bead Not Found

If `bd show <id>` says "not found", check routing:
```bash
BD_DEBUG_ROUTING=1 bd show <id>
```

The prefix may not match any registered route.

---

## References

Load JIT when needed:

| Reference | When to Load |
|-----------|--------------|
| `references/two-level.md` | Town vs Rig architecture details |
| `references/prefix-routing.md` | Full prefix routing mechanics |
| `references/sync-branch.md` | beads-sync branch workflow |
| `references/BEADS_DIR.md` | Environment variable usage |

---

## See Also

- `/beads` - General beads workflow
- `/sk-dispatch` - Work dispatch (uses routing)
- `/sk-gastown` - Gas Town status and management
