# BEADS_DIR Environment Variable

**For:** Targeting specific beads databases from any location

## Overview

`BEADS_DIR` overrides the default database selection, allowing you to create or modify beads in a different rig's database without changing directories.

## Why BEADS_DIR?

### The Problem

Mayor sits in `~/gt` (town level). Creating beads there gives `hq-*` prefix:

```bash
cd ~/gt
bd create --title="Fix daedalus bug"
# Creates: hq-xxxxx (can't be slung to polecats!)
```

### The Solution

Use `BEADS_DIR` to target the rig's database:

```bash
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="Fix bug"
# Creates: gt-xxxxx (slingable!)
```

## Syntax

```bash
BEADS_DIR=<path-to-.beads> bd <command>
```

The path must point to a `.beads/` directory containing `issues.jsonl`.

## Common Patterns

### Creating Slingable Beads (from Mayor)

```bash
# Target daedalus
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create \
  --title="Feature X" --type=feature --priority=2
# Creates: gt-xxxxx

# Target athena
BEADS_DIR=~/gt/athena/mayor/rig/.beads bd create \
  --title="API fix" --type=bug
# Creates: ap-xxxxx

# Target argus
BEADS_DIR=~/gt/argus/mayor/rig/.beads bd create \
  --title="Pipeline update" --type=task
# Creates: ho-xxxxx
```

### Batch Creation

```bash
# Create multiple beads in daedalus
export BEADS_DIR=~/gt/daedalus/mayor/rig/.beads

bd create --title="Task 1" --type=task
bd create --title="Task 2" --type=task
bd create --title="Task 3" --type=task

unset BEADS_DIR  # Clear when done
```

### Cross-Rig Operations

```bash
# From daedalus crew, view athena beads
BEADS_DIR=~/gt/athena/mayor/rig/.beads bd list

# From anywhere, update a specific rig's bead
BEADS_DIR=~/gt/argus/mayor/rig/.beads bd update ho-abc --status=closed
```

## Database Paths by Rig

| Rig | BEADS_DIR Path |
|-----|----------------|
| Town (HQ) | `~/gt/.beads` |
| daedalus | `~/gt/daedalus/mayor/rig/.beads` |
| athena | `~/gt/athena/mayor/rig/.beads` |
| argus | `~/gt/argus/mayor/rig/.beads` |
| chronicle | `~/gt/chronicle/mayor/rig/.beads` |
| cyclopes | `~/gt/cyclopes/mayor/rig/.beads` |

**Pattern:** `~/gt/<rig>/mayor/rig/.beads`

## Important Notes

### BEADS_DIR Does NOT Affect Prefix Routing

Prefix routing (for `bd show <id>`) uses `~/gt/.beads/routes.jsonl`, not `BEADS_DIR`.

```bash
# This still uses prefix routing
BEADS_DIR=~/gt/argus/mayor/rig/.beads bd show gt-abc
# Routes to daedalus (based on gt- prefix), NOT argus!
```

### When to Use BEADS_DIR

| Use Case | Use BEADS_DIR? |
|----------|----------------|
| Creating slingable beads from Mayor | Yes |
| Batch operations in a specific rig | Yes |
| Reading beads by ID | No (prefix routing handles it) |
| Working in your assigned rig | No (already in correct location) |

### Shell Aliases

Consider creating aliases for common operations:

```bash
# In ~/.bashrc or ~/.zshrc
alias bd-gt='BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd'
alias bd-ap='BEADS_DIR=~/gt/athena/mayor/rig/.beads bd'
alias bd-ho='BEADS_DIR=~/gt/argus/mayor/rig/.beads bd'
```

Usage:
```bash
bd-gt create --title="Daedalus task"  # Creates gt-xxxxx
bd-ap list                             # Lists athena beads
```

## Troubleshooting

| Issue | Fix |
|-------|-----|
| "No .beads database found" | Check path exists: `ls $BEADS_DIR` |
| Wrong prefix created | Verify `BEADS_DIR` points to correct rig |
| Export not working | Use `export BEADS_DIR=...` or inline syntax |
