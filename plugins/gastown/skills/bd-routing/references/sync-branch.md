# beads-sync Branch

**For:** Understanding multi-worktree beads synchronization

## Overview

Rig-level beads use a dedicated `beads-sync` branch to coordinate changes across multiple git worktrees (crew/*, polecats/*).

## Why beads-sync?

### The Problem

Multiple agents work in different worktrees:
- `gastown/crew/boden/.beads/`
- `gastown/polecats/toast/.beads/`
- `gastown/polecats/waffle/.beads/`

Without coordination, beads changes would conflict.

### The Solution

All worktrees sync beads through the `beads-sync` branch:

```
[polecat/toast] --push--> [beads-sync] <--pull-- [crew/boden]
                              ^
                              |
[polecat/waffle] --push-------+
```

## How It Works

### Creating/Updating Beads

1. Agent creates/updates a bead locally
2. `bd sync` pushes changes to `beads-sync` branch
3. Other worktrees pull from `beads-sync`

### The Sync Command

```bash
bd sync                  # Full sync: commit, push, pull
bd sync --from-main      # Pull from main (ephemeral branches)
bd sync --status         # Check sync status
```

## Town vs Rig Sync

| Level | Branch | Sync Needed? |
|-------|--------|--------------|
| Town | main | No (single clone) |
| Rig | beads-sync | Yes (multiple worktrees) |

## Ephemeral Branches

Polecats work on ephemeral branches like `polecat/cheedo-mk5rs7g3`. These branches:
- Have no upstream tracking
- Merge to main locally (not pushed)
- Need `--from-main` to pull beads updates

```bash
# On ephemeral branch, pull beads from main
bd sync --from-main
```

## Workflow Examples

### Polecat Session End

```bash
# 1. Commit code changes
git add .
git commit -m "feat: implement feature X"

# 2. Sync beads (pushes to beads-sync)
bd sync

# 3. Merge to main
git checkout main
git merge polecat/cheedo-mk5rs7g3
git push
```

### Pulling Beads Updates

```bash
# In any worktree, pull latest beads
bd sync --from-main

# Or full sync (push your changes, pull others)
bd sync
```

### Checking Sync Status

```bash
bd sync --status
# Shows: local changes, remote changes, sync needed?
```

## Merge Conflicts

Beads files (`.beads/issues.jsonl`) are append-only, so conflicts are rare. If they occur:

```bash
# Accept theirs (append-only means all entries are valid)
git checkout --theirs .beads/issues.jsonl
git add .beads/issues.jsonl
git commit -m "merge: resolve beads conflict"
```

## Troubleshooting

| Issue | Fix |
|-------|-----|
| "beads-sync branch not found" | `git branch beads-sync` on main |
| Sync fails on ephemeral branch | Use `bd sync --from-main` |
| Local changes not syncing | Check `bd sync --status` |
| Merge conflict | Accept theirs (append-only) |
