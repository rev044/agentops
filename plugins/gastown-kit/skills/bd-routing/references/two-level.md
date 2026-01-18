# Two-Level Beads Architecture

**For:** AI agents in Gas Town multi-workspace environments

## Overview

Gas Town operates with two distinct beads databases:

```
~/gt/                        ← TOWN LEVEL
├── .beads/                  ← Town beads (hq-* prefix)
│   ├── issues.jsonl
│   └── routes.jsonl         ← Prefix routing table
│
├── daedalus/                ← RIG
│   ├── .beads/              ← IGNORED (runtime state only)
│   ├── mayor/rig/.beads/    ← Canonical rig beads (gt-* prefix)
│   ├── crew/boden/.beads/   ← Worktree beads (synced via beads-sync)
│   └── polecats/*/.beads/   ← Worktree beads (synced via beads-sync)
│
└── athena/                  ← RIG
    ├── mayor/rig/.beads/    ← Canonical rig beads (ap-* prefix)
    └── ...
```

## Town Level

**Location:** `~/gt/.beads/`

**Prefix:** `hq-*`

**Purpose:**
- Mail and messaging between agents
- HQ coordination tasks
- Cross-rig meta-work

**Sync behavior:**
- Single clone (no worktrees)
- Commits directly to main branch
- No `beads-sync` branch needed

**Key characteristic:** Town beads CANNOT be hooked by polecats.

## Rig Level

**Location:** `<rig>/mayor/rig/.beads/` (canonical), synced to worktrees

**Prefix:** Rig-specific (`gt-*`, `ap-*`, `ho-*`, etc.)

**Purpose:**
- Actual project work
- Features, bugs, tasks
- Work that polecats execute

**Sync behavior:**
- Multiple worktrees share beads via `beads-sync` branch
- `bd sync` coordinates across worktrees
- `bd sync --from-main` pulls updates (for ephemeral branches)

**Key characteristic:** Rig beads CAN be slung to polecats.

## Why Two Levels?

### Separation of Concerns

| Concern | Level | Example |
|---------|-------|---------|
| Coordination | Town | "Dispatch wave 2 after wave 1" |
| Execution | Rig | "Fix authentication bug" |
| Communication | Town | Mail between agents |
| Implementation | Rig | Code changes |

### Isolation

- Town decisions don't pollute rig history
- Rigs can operate independently
- Prefix routing enables cross-rig references

### Slinging

The `gt sling` command only works with rig beads because:
1. `bd update` (used by sling) targets the local database
2. Polecats work in rig worktrees
3. They only see rig-level beads

## The Gitignored .beads

**Location:** `<rig>/.beads/` (directly in rig root)

**Status:** Gitignored

**Purpose:** Runtime state only (not persistent)

This is NOT the canonical rig beads location. The canonical location is
`<rig>/mayor/rig/.beads/`, which is inside the git repository.

## Determining Your Level

```bash
# Check current location
pwd

# If in ~/gt → Town level
# If in ~/gt/<rig>/* → Rig level

# Explicit check
BD_DEBUG_ROUTING=1 bd list
```

## Example: Creating Cross-Level Work

```bash
# Mayor creates coordination task (town level)
cd ~/gt
bd create --title="Coordinate security audit" --type=task
# Creates: hq-xxxxx

# Mayor creates actual work (rig level via BEADS_DIR)
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create \
  --title="Audit auth module" --type=task
# Creates: gt-xxxxx (slingable!)

# Sling the rig-level work
gt sling gt-xxxxx daedalus
```
