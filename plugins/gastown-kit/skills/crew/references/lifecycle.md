# Crew Lifecycle - Add, Start, Stop, Remove

Complete lifecycle management for crew workspaces.

## Creating Crew Workspaces

### Basic Creation

```bash
# Create single workspace
gt crew add dave

# Create in specific rig
gt crew add emma --rig gastown

# Create multiple at once
gt crew add murgen croaker goblin

# Create with feature branch
gt crew add fred --branch
```

Each workspace is created at `<rig>/crew/<name>/` with:
- Full git clone of the project repository
- Mail directory for message delivery
- CLAUDE.md with crew worker prompting
- Optional feature branch (`crew/<name>`)

### Directory Structure

```
<rig>/crew/<name>/
├── .beads/            # Beads database (may be symlinked)
├── .git/              # Full git clone
├── CLAUDE.md          # Crew-specific context
├── mail/              # Incoming messages
└── ...                # Project files
```

---

## Starting Crew Sessions

### Start Commands

```bash
# Start all crew in a rig
gt crew start gastown

# Start specific crew
gt crew start gastown dave

# Start with specific agent
gt crew start gastown dave --agent crew

# Start with specific account
gt crew start gastown dave --account main
```

### What Happens on Start

1. tmux session created: `gt-<rig>-crew-<name>`
2. Claude Code launched in the workspace directory
3. Session ready for interactive use

### Attaching to Sessions

```bash
# Attach to running session
gt crew at dave

# This opens the tmux session
# Detach with: Ctrl+b d
```

---

## Stopping Crew Sessions

### Stop Commands

```bash
# Stop specific crew
gt crew stop dave

# Stop all crew in a rig
gt crew stop gastown

# Stop all crew everywhere
gt crew stop --all

# Force stop (skip output capture)
gt crew stop dave --force
```

### What Happens on Stop

1. Output captured for debugging (unless `--force`)
2. Claude session terminated
3. tmux session killed
4. Workspace remains intact

**Note**: Stopping a crew session does NOT delete the workspace or any work.
Use `gt crew remove` to delete a workspace.

---

## Removing Crew Workspaces

### Remove Commands

```bash
# Remove single workspace
gt crew remove dave

# Remove multiple
gt crew remove dave emma
```

### What Happens on Remove

1. Running session stopped (if any)
2. Workspace directory deleted
3. Crew entry removed from state

**Warning**: This deletes all local work that hasn't been pushed!

### Before Removing

```bash
# Check for uncommitted work
gt crew status dave

# Or manually check
git -C <rig>/crew/dave status
```

---

## Lifecycle Summary

| Stage | Command | Result |
|-------|---------|--------|
| Create | `gt crew add <name>` | Workspace created |
| Start | `gt crew start <name>` | Session running |
| Attach | `gt crew at <name>` | Connected to session |
| Stop | `gt crew stop <name>` | Session ended |
| Remove | `gt crew remove <name>` | Workspace deleted |

---

## Common Patterns

### First-Time Setup

```bash
# Create workspace for a rig you'll work on regularly
gt crew add boden --rig ai-platform

# Start and begin working
gt crew start ai-platform boden
gt crew at boden
```

### Daily Workflow

```bash
# Start of day
gt crew start gastown

# Work...

# End of day
git push
bd sync
gt crew stop gastown
```

### Cleanup Old Workspaces

```bash
# List all crew
gt crew list

# Remove unused ones
gt crew remove old-workspace
```
