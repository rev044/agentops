# gt polecat gc - Branch Garbage Collection

Clean up orphaned polecat branches.

## Synopsis

```bash
gt polecat gc <rig>
gt polecat gc <rig> --dry-run
```

## The Problem

Polecats use timestamped branches:
- `polecat/Toast-mk5rtfes`
- `polecat/Toast-ab7xyz12`
- `polecat/Furiosa-cd9efgh3`

Over time, branches accumulate:
- Old timestamps when polecats are reused
- Branches for polecats that no longer exist
- Leftover after nuke operations

## What GC Does

Removes orphaned branches:

1. **Old timestamps** - Keeps only current branch per polecat
2. **Non-existent polecats** - Removes branches for deleted polecats

Does NOT remove:
- Current branches for active polecats
- Remote branches (local only)

## Usage

### Preview (Recommended First Step)

```bash
gt polecat gc gastown --dry-run
```

Output:
```
Would delete: polecat/Toast-ab7xyz12 (old timestamp)
Would delete: polecat/Nux-defunct (no polecat)
Would keep: polecat/Toast-mk5rtfes (current)
```

### Execute

```bash
gt polecat gc gastown
```

Output:
```
Deleted: polecat/Toast-ab7xyz12
Deleted: polecat/Nux-defunct
Kept: polecat/Toast-mk5rtfes
GC complete: 2 deleted, 1 kept
```

## When to Run GC

Run after:
- Nuking polecats
- Batch cleanup
- Noticing stale branches
- Before starting fresh wave of work

Frequency:
- After each cleanup session
- Weekly maintenance
- When branch count gets high

## Workflow Integration

### Post-Merge Cleanup

```bash
# 1. Nuke completed polecat
gt polecat nuke gastown/Toast

# 2. Clean branches
gt polecat gc gastown
```

### Batch Cleanup

```bash
# 1. Nuke all polecats
gt polecat nuke gastown --all

# 2. Verify removal
gt polecat list gastown

# 3. Clean branches
gt polecat gc gastown
```

### Maintenance Window

```bash
# Across all rigs
for rig in gastown ai-platform houston; do
  echo "=== $rig ==="
  gt polecat stale $rig
  gt polecat gc $rig --dry-run
done
```

## Branch Anatomy

Understanding branch naming helps GC:

```
polecat/Toast-mk5rtfes
│       │     └─ timestamp (8 chars)
│       └─ polecat name
└─ namespace
```

GC logic:
- Groups branches by polecat name
- Identifies current timestamp from worktree
- Deletes all other timestamps for that name
- Deletes all branches for non-existent polecats

## Remote Branches

GC only cleans LOCAL branches. Remote cleanup:

```bash
# Delete remote branch manually
git push origin --delete polecat/Toast-ab7xyz12

# Or after merge, prune
git fetch --prune
```

## Dry Run Output

```bash
gt polecat gc gastown --dry-run
```

Categories:
- **Would delete (old timestamp)** - Stale version of existing polecat
- **Would delete (no polecat)** - Branch for deleted polecat
- **Would keep (current)** - Active polecat branch
- **Would keep (unknown)** - Can't determine status, safer to keep

## Safety

GC is non-destructive:
- Only deletes clearly orphaned branches
- Keeps anything uncertain
- `--dry-run` shows exactly what happens
- No data loss (work should be pushed/merged first)

## Troubleshooting

### Branches Not Deleted

```
Keeping: polecat/Mystery-xyz (unknown status)
```

Possible causes:
- Polecat exists but worktree is corrupted
- Branch name doesn't match expected pattern
- Permission issues

Manual cleanup:
```bash
git branch -D polecat/Mystery-xyz
```

### Too Many Branches

If you have many stale branches:

```bash
# List all polecat branches
git branch | grep polecat/

# Aggressive cleanup (careful!)
git branch | grep polecat/ | xargs -n 1 git branch -D
```

### Remote Branch Accumulation

GC doesn't touch remote. Clean periodically:

```bash
# See remote polecat branches
git branch -r | grep polecat/

# Prune merged remotes
git fetch --prune
```

## Best Practices

1. **Always dry-run first** - See what will be deleted
2. **Run after nuke** - Part of cleanup workflow
3. **Weekly maintenance** - Prevent accumulation
4. **Push before cleanup** - Ensure work is preserved
5. **Handle remote separately** - GC is local only
