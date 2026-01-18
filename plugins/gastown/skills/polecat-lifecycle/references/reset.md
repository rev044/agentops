# Resetting Polecats

How to return a polecat to a clean state for reuse.

## Current Approach: Nuke and Re-Spawn

There is no dedicated `gt polecat reset` command. The standard pattern is:

```bash
# 1. Destroy the polecat
gt polecat nuke <rig>/<name> --force

# 2. Spawn fresh with new work
gt sling <new-bead> <rig>
```

This creates a completely clean polecat with:
- Fresh worktree from main
- New timestamped branch
- Clean state (no prior work artifacts)

## Why Nuke+Spawn?

### Isolation Guarantees

Each polecat execution is isolated:
- Fresh checkout from main
- No leftover files from prior work
- No git history entanglement

### Branch Freshness

Timestamped branches ensure:
- Work starts from current main
- No merge conflicts from drift
- Clean git history

### Simplicity

The model is simple:
- Polecat spawns → does work → gets destroyed
- No complex state management
- No "reuse this same environment" complexity

## When to Reset

Reset (nuke+spawn) when:
- Polecat completed its work and you have more
- Polecat got stuck and needs fresh start
- You want to reassign to different work
- State is somehow corrupted

## Pre-Reset Checklist

Before nuking:

```bash
# Check what would be lost
gt polecat git-state <rig>/<name>

# See if recovery needed vs safe to nuke
gt polecat check-recovery <rig>/<name>

# Check for unpushed work
gt polecat status <rig>/<name>
```

## Preserving Work

If polecat has unmerged work:

```bash
# Option 1: Push first, then nuke
cd ~/gt/<rig>/polecats/<name>
git push -u origin HEAD
gt polecat nuke <rig>/<name>

# Option 2: Force nuke (LOSES WORK)
gt polecat nuke <rig>/<name> --force
```

## Soft Reset Alternative

For minor resets without full destruction:

```bash
# Inside polecat worktree
cd ~/gt/<rig>/polecats/<name>
git reset --hard origin/main
git clean -fdx
```

This keeps the worktree but resets to main. Use sparingly - full nuke+spawn
is cleaner.

## Reset Patterns

### After Successful Completion

```bash
# Work done, merged to main
gt polecat nuke <rig>/<name>        # Clean up
gt sling <next-bead> <rig>          # New work spawns new polecat
```

### After Stuck/Failed Work

```bash
# Something went wrong
gt polecat check-recovery <rig>/<name>  # Assess damage
gt polecat nuke <rig>/<name> --force    # Nuclear option
gt sling <same-or-new-bead> <rig>       # Fresh start
```

### Reassigning Work

```bash
# Move work to different polecat
gt unsling                               # Unhook from current
gt polecat nuke <rig>/<name> --force    # Destroy old
gt sling <bead> <different-rig>         # Reassign elsewhere
```

## Future: Dedicated Reset Command

A `gt polecat reset` command may be added for:
- Faster recycling without full nuke
- Preserving polecat identity/name
- Simpler single-command workflow

Until then, nuke+spawn is the canonical pattern.
