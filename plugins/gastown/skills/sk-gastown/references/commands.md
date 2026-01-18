# Subcommands Reference

## Overview

The `/gastown` command supports multiple modes via subcommands and flags.

## Command Syntax

```bash
/gastown [subcommand|epic-id|"goal"] [options]
```

## Subcommands

### status

Show active convoys and polecats:

```bash
/gastown status
```

Runs:
```bash
# Active convoys
gt convoy list

# Rig status (polecats, daemon)
gt rig status gastown
```

Output example:
```
Active Convoys:
  hq-cv-xyz: Wave 1: OAuth  [2/3] ●
  hq-cv-abc: Bug fixes      [5/5] ✓

Rig: gastown
  Polecats: 3 running
  Daemon: active
```

### resume

Continue from last checkpoint:

```bash
/gastown resume
```

Runs:
```bash
# Find hooked work
gt hook
# → ga-tq9

# Read saved state
bd comments ga-tq9 | grep "GASTOWN_STATE:"
# → {wave: 2, convoy: hq-cv-xyz, ...}

# Resume from state
```

### peek

Investigate specific polecat:

```bash
/gastown peek gastown/polecats/nux
```

Runs:
```bash
gt peek gastown/polecats/nux
```

### stop

Stop all polecats gracefully:

```bash
/gastown stop
```

Runs:
```bash
# Down each active polecat
for polecat in $(gt rig status gastown --json | jq -r '.polecats[]'); do
    gt down $polecat
done
```

## Flags

### --full

Enable R→P→I workflow:

```bash
/gastown "Add OAuth" --full
```

See `rpi-flow.md` for details.

### --rig <name>

Specify target rig:

```bash
/gastown ga-tq9 --rig myproject
```

Default: auto-detect from `gt rig list`.

### --max <n>

Limit concurrent polecats:

```bash
/gastown ga-tq9 --max 4
```

Default: 8.

### --dry-run

Preview without executing:

```bash
/gastown ga-tq9 --dry-run
```

Shows:
- Waves and issue assignments
- Convoy that would be created
- Polecats that would be spawned

### --resume

Explicit resume (alternative to `resume` subcommand):

```bash
/gastown ga-tq9 --resume
```

## Argument Detection

The skill auto-detects argument type:

| Input | Detection | Action |
|-------|-----------|--------|
| `status` | Subcommand | Show status |
| `resume` | Subcommand | Resume from checkpoint |
| `ga-xxx` | Epic ID pattern | Execute epic |
| `"..."` | Quoted string | Full RPI workflow |

## Examples

```bash
# Full workflow with goal
/gastown "Add user authentication"

# Execute existing epic
/gastown ga-tq9

# Check status
/gastown status

# Resume paused work
/gastown resume

# Dry run preview
/gastown ga-tq9 --dry-run

# Limited parallelism
/gastown ga-tq9 --max 4

# Specific rig
/gastown ga-tq9 --rig myproject

# Investigate polecat
/gastown peek gastown/polecats/nux

# Stop all polecats
/gastown stop
```

## Error Messages

| Condition | Message |
|-----------|---------|
| No Gas Town | "Gas Town not installed. Run setup first." |
| Daemon down | "Daemon not running. Start with: gt daemon start" |
| No rig | "No rig configured. Add with: gt rig add <name> <url>" |
| Epic not found | "Epic <id> not found in beads." |
| No work to resume | "No hooked work found. Nothing to resume." |
