# gt polecat nuke - Complete Destruction

The nuclear option for polecat cleanup.

## Synopsis

```bash
gt polecat nuke <rig>/<polecat>...
gt polecat nuke <rig> --all
```

## What It Does

Complete destruction in order:

1. **Kill Claude session** (if running)
2. **Delete git worktree** (bypasses all safety checks with --force)
3. **Delete polecat branch**
4. **Close agent bead** (if exists)

## Safety Checks

By default, nuke REFUSES if:

| Condition | Why It Blocks |
|-----------|---------------|
| Unpushed commits | Work would be lost |
| Uncommitted changes | Work would be lost |
| Open merge request | MR bead still active |
| Work on hook | Task incomplete |

Use `--force` to bypass ALL safety checks.

## Usage

### Single Polecat

```bash
# With safety checks
gt polecat nuke daedalus/Toast

# Bypass safety (LOSES WORK)
gt polecat nuke daedalus/Toast --force
```

### Multiple Polecats

```bash
gt polecat nuke daedalus/Toast daedalus/Furiosa daedalus/Max
```

### All Polecats in Rig

```bash
# Preview first!
gt polecat nuke daedalus --all --dry-run

# Execute
gt polecat nuke daedalus --all

# Force all (dangerous!)
gt polecat nuke daedalus --all --force
```

## Flags

| Flag | Description |
|------|-------------|
| `--all` | Nuke all polecats in rig |
| `--dry-run` | Show what would happen |
| `-f, --force` | Bypass all safety checks |

## Pre-Nuke Verification

Before nuking, verify state:

```bash
# See what would be lost
gt polecat git-state daedalus/Toast

# Check if recovery needed
gt polecat check-recovery daedalus/Toast

# Full status
gt polecat status daedalus/Toast
```

## Common Scenarios

### Post-Merge Cleanup

```bash
# Work merged, safe to nuke
gt polecat nuke daedalus/Toast
gt polecat gc daedalus  # Clean branches
```

### Stuck Polecat

```bash
# Check first
gt polecat check-recovery daedalus/Toast

# If safe, force nuke
gt polecat nuke daedalus/Toast --force
```

### Batch Cleanup

```bash
# Preview
gt polecat nuke daedalus --all --dry-run

# Execute with safety
gt polecat nuke daedalus --all

# Force (last resort)
gt polecat nuke daedalus --all --force
```

## Error Handling

### Safety Check Failures

```
Error: polecat has unpushed commits
Hint: Use --force to bypass (LOSES WORK)
```

Options:
1. Push the work first, then nuke
2. Use `--force` if work is truly disposable

### Worktree Not Found

```
Error: worktree not found: ~/gt/daedalus/polecats/Toast
```

Polecat directory missing. Branch may still exist - use `gt polecat gc` to clean.

### Session Won't Die

```
Error: failed to kill session gt-daedalus-Toast
```

Manually kill:
```bash
tmux kill-session -t gt-daedalus-Toast
gt polecat nuke daedalus/Toast
```

## What Gets Deleted

| Artifact | Location | Deleted? |
|----------|----------|----------|
| Worktree | `~/gt/<rig>/polecats/<name>/` | Yes |
| Branch | `polecat/<name>-<timestamp>` | Yes (local) |
| Tmux session | `gt-<rig>-<name>` | Yes |
| Agent bead | `.beads/` | Closed |
| Remote branch | `origin/polecat/...` | No (manual) |

Remote branches persist. Clean manually or via `gt polecat gc` after merge.

## Post-Nuke

After nuking:

```bash
# Clean orphaned branches
gt polecat gc <rig>

# Verify removal
gt polecat list <rig>
```

## Best Practices

1. **Always --dry-run first** for batch operations
2. **Check git-state** before force nuke
3. **Push important work** before destruction
4. **Run gc after nuke** to clean branches
5. **Verify with list** after cleanup
