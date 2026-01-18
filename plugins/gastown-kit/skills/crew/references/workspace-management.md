# Crew Workspace Management - Pristine, Refresh, Rename

Operations for maintaining and managing crew workspaces.

## Pristine - Sync with Remote

**Purpose**: Ensure workspace is up-to-date with remote.

```bash
# Pristine single workspace
gt crew pristine dave

# Pristine all crew in a rig
gt crew pristine --rig gastown

# Pristine all crew everywhere
gt crew pristine

# JSON output for scripting
gt crew pristine --json
```

### What Pristine Does

1. **Git pull**: Fetches and merges from remote
2. **Beads sync**: Runs `bd sync` to update issues
3. **Report**: Shows any uncommitted changes needing attention

### When to Use Pristine

| Scenario | Use Pristine |
|----------|--------------|
| Start of day | Yes - get latest changes |
| After other team members pushed | Yes - sync their work |
| Before starting new work | Yes - ensure clean state |
| Context feels stale | Maybe - consider refresh instead |

### Pristine vs Refresh

| | Pristine | Refresh |
|-|----------|---------|
| **Git operation** | Pull from remote | No git changes |
| **Session state** | Preserves | Restarts |
| **Context** | Preserves | Cycles (fresh start) |
| **Handoff** | No | Yes (mail to self) |

**Use pristine** for syncing remote changes.
**Use refresh** for context cycling.

---

## Refresh - Context Cycling with Handoff

**Purpose**: Start fresh session while preserving continuity.

```bash
# Basic refresh (auto-generated handoff)
gt crew refresh dave

# Refresh with custom message
gt crew refresh dave -m "Working on gt-123, tests passing"

# Specify rig
gt crew refresh dave --rig gastown
```

### What Refresh Does

1. **Captures context**: Current work state, progress
2. **Sends handoff mail**: To workspace's own inbox
3. **Restarts session**: Fresh Claude instance
4. **Reads handoff**: New session picks up context

### Handoff Mail Format

The refresh operation sends mail like:

```
To: gastown/crew/dave
Subject: Handoff
Body: <your message or auto-summary>
```

The new session reads this mail and continues work.

### When to Use Refresh

| Scenario | Use Refresh |
|----------|-------------|
| Context getting long | Yes - cycle for fresh start |
| Model feels confused | Yes - reset helps |
| Switching focus | Yes - clean break |
| Need latest code changes | No - use pristine instead |

### Custom Handoff Messages

Good handoff messages include:

```bash
# Status + what's next
gt crew refresh dave -m "Completed auth module. Tests pass. Next: implement logout."

# Blockers
gt crew refresh dave -m "Stuck on oauth config. Need to read docs. Issue gt-456."

# Decision context
gt crew refresh dave -m "Chose JWT over sessions. See commit abc123 for rationale."
```

---

## Rename - Change Workspace Identity

**Purpose**: Change the name of a crew workspace.

```bash
# Rename workspace
gt crew rename dave david

# Specify rig
gt crew rename old new --rig gastown
```

### What Rename Does

1. **Stops session**: If running, killed first
2. **Renames directory**: `<rig>/crew/dave/` → `<rig>/crew/david/`
3. **Updates state**: Crew registry updated
4. **New session name**: `gt-<rig>-crew-david`

### When to Use Rename

| Scenario | Use Rename |
|----------|------------|
| Standardizing names | Yes |
| Typo in name | Yes |
| Reorganizing crew | Yes |
| Name conflict | Yes |

### Considerations

- Running session will be stopped
- Git state preserved (no changes to branches)
- Mail directory moved with workspace
- CLAUDE.md identity updated

---

## Status Checking

Before making changes, check workspace status:

```bash
# List all crew with quick status
gt crew list

# Detailed status for specific workspace
gt crew status dave

# Check multiple
gt crew status dave emma
```

### Status Output Includes

- Running state (session active?)
- Git status (uncommitted changes?)
- Beads sync status
- Last activity

---

## Best Practices

### Daily Workflow

```bash
# Morning: sync with remote
gt crew pristine dave

# Work during day...

# Context getting heavy? Refresh
gt crew refresh dave -m "Progress: A done, B in progress"

# End of day: commit, push, stop
git push
bd sync
gt crew stop dave
```

### Before Major Changes

```bash
# Check status
gt crew status dave

# Ensure clean state
gt crew pristine dave

# Verify no uncommitted work
git status
```

### Workspace Hygiene

```bash
# Regular pristine keeps things synced
gt crew pristine  # all crew

# Remove unused workspaces
gt crew list
gt crew remove old-workspace

# Rename for clarity
gt crew rename temp proper-name
```
