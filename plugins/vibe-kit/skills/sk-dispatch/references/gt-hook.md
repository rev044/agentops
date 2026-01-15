# gt hook - Work Attachment Semantics

**Show what's on your hook, or attach new work.**

## Synopsis

```bash
gt hook                              # Show what's hooked
gt hook <bead>                       # Attach bead to hook
gt hook status                       # Same as no args
gt hook show <agent>                 # Show another agent's hook
```

## What Is The Hook?

The hook is Gas Town's **durability primitive**. Work on your hook survives:

- Session restarts
- Context compaction
- Handoffs

When you restart (via `gt handoff` or session reconnect), your SessionStart
hook finds the attached work and you continue from where you left off.

## Basic Usage

```bash
# Check your hook
gt hook
gt hook status                       # Explicit status

# Attach work
gt hook gt-abc                       # Attach issue
gt hook gt-abc -s "Fix the bug"      # With subject for handoff mail

# Remove work
gt unsling                           # Clear your hook

# Check another agent's hook
gt hook show gastown/Toast           # Polecat's hook
gt hook show mayor                   # Mayor's hook
```

## Hook vs Sling vs Handoff

| Command | Hooks Work | Starts Session | Keeps Context |
|---------|------------|----------------|---------------|
| `gt hook <bead>` | Yes | No | Yes |
| `gt sling <bead>` | Yes | Yes | Yes |
| `gt handoff <bead>` | Yes | Yes (new) | No (fresh) |

**When to use hook:**
- Attaching work without triggering immediate execution
- Queuing work for later
- Manual session management

**When to use sling:**
- Assigning and starting immediately
- Dispatching to polecats

**When to use handoff:**
- Need fresh context (e.g., after compaction)
- Handing off to next session

## The Propulsion Principle

> **If you find work on your hook, YOU RUN IT.**

When a Gas Town agent starts (or restarts), it:

1. Runs SessionStart hook
2. Checks `gt hook`
3. If work found → **EXECUTES IMMEDIATELY**

No confirmation. No waiting. The hook IS your assignment.

## Hookable Mail

Mail beads can be hooked for ad-hoc instruction handoff:

```bash
gt hook attach <mail-id>             # Hook existing mail
gt handoff -m "..."                  # Create and hook new instructions
```

If you find mail on your hook, interpret the prose instructions and execute.
This enables ad-hoc tasks without creating formal beads.

## Hook Lifecycle

```
1. Work Created
   bd create "Fix X"
   -> gt-abc

2. Work Attached to Hook
   gt hook gt-abc
   OR
   gt sling gt-abc gastown (auto-hooks)
   -> Work is now on agent's hook

3. Agent Starts/Restarts
   SessionStart hook runs
   gt hook -> shows gt-abc
   -> Agent sees work

4. Propulsion Kicks In
   Agent reads bead
   Agent executes
   -> Work progressing

5. Work Completes
   bd close gt-abc
   gt unsling (automatic on close)
   -> Hook cleared
```

## Flags Reference

| Flag | Description |
|------|-------------|
| `-n, --dry-run` | Show what would be done |
| `-f, --force` | Replace existing incomplete hooked bead |
| `--json` | Output as JSON (for status) |
| `-m, --message <text>` | Message for handoff mail |
| `-s, --subject <text>` | Subject for handoff mail |

## Subcommands

| Subcommand | Description |
|------------|-------------|
| `gt hook show <agent>` | Show another agent's hook |
| `gt hook status` | Show your hook (default) |

## Related Commands

```bash
gt sling <bead>          # Hook + start now
gt handoff <bead>        # Hook + restart (fresh context)
gt unsling               # Remove work from hook
```

## Hooked vs Pinned

- **Hooked**: Work assigned to you. Triggers autonomous execution.
- **Pinned**: Permanent reference beads (different concept entirely).

Don't confuse these - they serve different purposes.

## Examples

### Queuing Work for Later

```bash
# Attach but don't start yet
gt hook gt-abc
# Work is hooked but current session continues

# Later, trigger execution
gt sling gt-abc          # Now starts the work
```

### Checking What's Hooked

```bash
# Your hook
gt hook
# -> Hooked: gt-abc - Fix the authentication bug

# Another agent's hook
gt hook show gastown/Toast
# -> Hooked: gt-def - Implement OAuth
```

### Force Replace

```bash
# Current hook has incomplete work
gt hook
# -> Hooked: gt-old (in_progress)

# Force replace with new work
gt hook gt-new --force
# -> Replaced: gt-old with gt-new
```

## Error Handling

If hook fails:

```bash
# Check what's currently hooked
gt hook status

# If something's there and you need to replace
gt hook <new-bead> --force

# Or clear first
gt unsling
gt hook <new-bead>
```
