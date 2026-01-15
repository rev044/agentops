# Crew Refresh Pattern

Cycling crew workspace sessions with context handoff.

## Overview

`gt crew refresh` restarts a crew workspace session while preserving work context.
It sends a handoff mail to the workspace's own inbox, then restarts the session.
The new session reads the handoff and resumes work.

## Synopsis

```bash
gt crew refresh <name> [flags]
```

## Arguments

| Argument | Description |
|----------|-------------|
| `name` | Name of the crew workspace to refresh |

## Flags

| Flag | Description |
|------|-------------|
| `--rig` | Rig to use (if multiple rigs have same crew name) |
| `-m, --message` | Custom handoff message |
| `-h, --help` | Help for refresh |

## Examples

### Basic Refresh

```bash
# Refresh with auto-generated handoff
gt crew refresh dave
```

### With Custom Message

```bash
# Include specific context
gt crew refresh dave -m "Working on gt-123, left off at tests"
```

### Specify Rig

```bash
# When crew name exists in multiple rigs
gt crew refresh dave --rig gastown
```

## When to Use

| Scenario | Action |
|----------|--------|
| Context approaching limit | `gt crew refresh` |
| Session seems sluggish | `gt crew refresh` |
| Need fresh perspective | `gt crew refresh` |
| Long-running implementation | Proactive refresh |

## What Happens

1. **Mail created**: Handoff mail sent to crew's own inbox
2. **Session terminated**: Current Claude instance ends
3. **New session spawns**: Fresh context starts
4. **SessionStart hook runs**: `gt prime` loads context
5. **Mail read**: New session sees handoff context
6. **Work continues**: Resume from where left off

## Flow Diagram

```
Current Session (context: 40%)
    │
    ├─ gt crew refresh dave
    │       │
    │       ├─ Creates mail: "HANDOFF: Auto-refresh"
    │       ├─ Sends to: gastown/crew/dave (self)
    │       └─ Terminates session
    │
    ▼
New Session (context: 0%)
    │
    ├─ SessionStart hook: gt prime
    │       │
    │       ├─ Checks hook → (hooked work if any)
    │       └─ Checks mail → "HANDOFF: Auto-refresh"
    │
    └─ Resumes work with full context capacity
```

## Crew vs Polecat Refresh

| Aspect | Crew Refresh | Polecat Handoff |
|--------|--------------|-----------------|
| Command | `gt crew refresh` | `gt handoff` / `gt done --status DEFERRED` |
| Lifecycle | Self-managed | Witness-managed |
| Context | Persists via mail | May be reassigned |
| Typical use | Long implementation | Parallel task batches |

## Context Thresholds for Crew

| Usage | Action |
|-------|--------|
| < 30% | Continue normally |
| 30-35% | Consider refresh if complex work ahead |
| > 35% | Refresh immediately |

## Best Practices

### Proactive Refresh

Don't wait until context is critical:

```bash
# At ~30% with significant work remaining
gt crew refresh dave -m "Proactive refresh before algorithm impl"
```

### Include Key Context

```bash
gt crew refresh dave -m "$(cat <<'EOF'
## Current Work
Implementing user authentication for gt-auth-123.

## Completed
- Login endpoint
- Session management

## In Progress
- Logout (started, see src/auth/logout.py)

## Next
- Password reset flow
- Tests
EOF
)"
```

### Commit Before Refresh

```bash
# Save all work first
git add .
git commit -m "wip: auth logout (pre-refresh)"
bd sync

# Then refresh
gt crew refresh dave
```

## Automatic State Collection

The refresh command can auto-collect state:

- Current hooked work
- Recent beads activity
- Git status

This is included in the handoff mail even without `-m`.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Crew not found | Wrong name or rig | Check `gt crew list` |
| Mail not received | Crew inbox issue | Verify with `gt mail inbox crew/<name>` |
| Work lost | Not committed | Always commit before refresh |
| Context still high | Large codebase | Split work into smaller beads |

## Integration with Beads

Crew refresh works with the beads workflow:

```bash
# Before refresh
bd update gt-123 --status in_progress
bd comments add gt-123 "Pausing for context refresh"

# Refresh
gt crew refresh dave

# After (new session)
bd show gt-123  # Get context
# Continue work
```

## See Also

- `gt crew list` - List crew workspaces
- `gt crew status` - Check crew status
- `gt handoff` - General handoff command
- `gt prime` - Context recovery
