# Spawning Polecats

How polecats come into existence and start working.

## Primary Method: gt sling

**This is how 99% of polecats spawn.** Don't use `gt polecat add` unless you have
a specific reason.

```bash
gt sling <bead> <rig>
```

### What Happens

1. **Worktree creation**: `~/gt/<rig>/polecats/<name>/`
2. **Branch creation**: `polecat/<name>-<timestamp>`
3. **Tmux session**: `gt-<rig>-<name>`
4. **Work hooked**: Bead attached to polecat
5. **Claude started**: SessionStart hook finds work, begins execution

### Spawn Flow

```
Orchestrator                     Gas Town                    Polecat
    |                               |                           |
    |  gt sling gt-abc gastown      |                           |
    | ----------------------------->|                           |
    |                               | git worktree add          |
    |                               | tmux new-session          |
    |                               | bd update --assignee      |
    |                               | ------------------------->|
    |                               |                           | claude
    |                               |                           | SessionStart
    |                               |                           | gt hook -> work!
    |  (returns immediately)        |                           | Execute...
```

### Options

```bash
# Basic spawn
gt sling gt-abc gastown

# To specific (existing) polecat
gt sling gt-abc gastown/Toast

# Multiple beads (parallel polecats)
gt sling gt-abc gt-def gt-ghi gastown

# Force spawn even with unread mail
gt sling gt-abc gastown --force

# Naked mode (no tmux auto-start)
gt sling gt-abc gastown --naked

# Specific Claude account
gt sling gt-abc gastown --account work
```

### Naming

Polecat names are auto-generated:
- Format: `<adjective>` or similar short identifier
- Unique within rig
- Used in branch name: `polecat/<name>-<timestamp>`

## Alternative: gt polecat add

Manual polecat creation. Rarely needed.

```bash
gt polecat add <rig> <name>
```

Use cases:
- Pre-staging polecats before dispatch
- Testing infrastructure
- Custom naming requirements

**Note:** Creates worktree but doesn't dispatch work. You'd need to separately
hook or sling work to it.

## Timestamped Branches

Polecats use timestamped branches:

```
polecat/Toast-mk5rtfes
        └────┬────────┘
          name-timestamp
```

Why timestamps?
- Prevents drift when polecat is reused
- Each "reset" gets fresh history from main
- Old branches cleaned via `gt polecat gc`

## After Spawn

Once spawned, polecat:

1. Finds work on hook (SessionStart)
2. Executes autonomously (Propulsion Principle)
3. Commits work
4. Pushes branch
5. Closes bead
6. Awaits cleanup

## Spawn Failures

Common issues:

| Error | Cause | Fix |
|-------|-------|-----|
| "rig not found" | Invalid rig name | Check `gt rig list` |
| "bead not found" | Wrong prefix | Verify bead exists in rig's database |
| "worktree exists" | Name collision | Let auto-naming pick new name |

## Best Practices

1. **Use sling, not add** - Sling handles everything
2. **Let names auto-generate** - Don't micromanage
3. **One bead per polecat** - Polecats are single-issue focused
4. **Match prefix to rig** - Bead prefix indicates correct rig
