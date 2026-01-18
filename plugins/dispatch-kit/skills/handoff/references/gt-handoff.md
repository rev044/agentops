# gt handoff Command Reference

End session and hand off to a fresh agent.

## Synopsis

```bash
gt handoff [bead-or-role] [flags]
```

## Description

The `gt handoff` command is the canonical way to end any agent session and ensure
work continues in a fresh context. It handles all Gas Town roles:

- **Mayor, Crew, Witness, Refinery, Deacon**: Respawns with fresh Claude instance
- **Polecats**: Calls `gt done --status DEFERRED` (Witness handles lifecycle)

## Arguments

| Argument | Description |
|----------|-------------|
| `bead-or-role` | Optional. A bead ID to hook before restart, or a role name to hand off |

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--subject` | `-s` | Subject for handoff mail (optional) |
| `--message` | `-m` | Message body for handoff mail (optional) |
| `--collect` | `-c` | Auto-collect state into handoff message |
| `--watch` | `-w` | Switch to new session (default: true) |
| `--dry-run` | `-n` | Show what would be done without executing |
| `--help` | `-h` | Help for handoff |

## Examples

### Basic Handoff

```bash
# End current session, restart fresh
gt handoff
```

### With Context

```bash
# Include handoff notes
gt handoff -s "Working on auth" -m "Login done, starting logout"

# Auto-collect state (hook, inbox, ready beads)
gt handoff -c
```

### Hook Work First

```bash
# Hook bead, then restart
gt handoff gt-abc

# Hook with context
gt handoff gt-abc -s "Bug fix" -m "Root cause in auth.py:42"
```

### Role-Specific Handoff

```bash
# Hand off crew session
gt handoff crew

# Hand off mayor session
gt handoff mayor
```

### Dry Run

```bash
# See what would happen
gt handoff -n

# Preview with bead
gt handoff gt-abc -n
```

## What Happens

1. **Mail created**: Handoff mail sent to your own inbox
2. **State collected** (if `-c`): Current status, inbox, beads state captured
3. **Session terminated**: Current Claude instance ends
4. **New session spawns**: Fresh context, full capacity
5. **SessionStart hook runs**: `gt prime` restores context
6. **Work continues**: Hook is checked, work resumes

## The --collect Flag

When `-c` is used, automatically gathers:

- Current hooked work
- Unread inbox messages
- Ready beads (from `bd ready`)
- In-progress issues

This provides context for the next session without manual summarization.

## Handoff vs Done

| Command | Use When | Next State |
|---------|----------|------------|
| `gt handoff` | Context full, want to continue | Fresh session continues |
| `gt done` | Work complete | Session ends, no continuation |

For polecats, `gt handoff` maps to `gt done --status DEFERRED` since the Witness
manages polecat lifecycle.

## Integration with Hooks

Work on the hook persists across handoff:

```bash
# Before handoff
gt hook gt-abc
gt handoff

# After restart (new session)
gt hook
# → gt-abc (still there)
```

Any molecule attached to the hook is auto-continued by the new session.

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| "No session to handoff" | Not in managed session | Use from crew/polecat workspace |
| "Mail send failed" | Mail system issue | Check `gt mail inbox` |
| "Hook failed" | Invalid bead ID | Verify with `bd show <id>` |

## See Also

- `gt done` - Mark work complete (no continuation)
- `gt hook` - Check/set hooked work
- `gt crew refresh` - Crew-specific handoff
- `gt mail send --self` - Manual handoff mail
