# gt sling - THE Work Dispatch Command

**The unified command for assigning work in Gas Town.**

## Synopsis

```bash
gt sling <bead-or-formula> [target] [flags]
```

## What It Does

`gt sling` is THE command for work assignment. It handles:

1. **Existing agents** - mayor, crew, witness, refinery
2. **Auto-spawning polecats** - when target is a rig
3. **Dispatching to dogs** - Deacon's helper workers
4. **Formula instantiation** - molecule/wisp creation
5. **Auto-convoy creation** - dashboard visibility

## Target Resolution

| Target | Description | Example |
|--------|-------------|---------|
| (none) | Self (current agent) | `gt sling gt-abc` |
| `crew` | Crew worker in current rig | `gt sling gt-abc crew` |
| `<rig>` | Auto-spawn polecat in rig | `gt sling gt-abc gastown` |
| `<rig>/<polecat>` | Specific polecat | `gt sling gt-abc gastown/Toast` |
| `mayor` | Mayor | `gt sling gt-abc mayor` |
| `deacon/dogs` | Auto-dispatch to idle dog | `gt sling gt-abc deacon/dogs` |
| `deacon/dogs/<name>` | Specific dog | `gt sling gt-abc deacon/dogs/alpha` |

## Auto-Convoy

Single-issue slings auto-create a convoy for tracking:

```bash
gt sling gt-abc gastown              # Creates "Work: <title>" convoy
gt sling gt-abc gastown --no-convoy  # Skip auto-convoy
```

This ensures all work appears in `gt convoy list`.

## Batch Slinging

Multiple beads to same rig:

```bash
gt sling gt-abc gt-def gt-ghi gastown
```

Each bead gets its own polecat - parallelizes dispatch.

## Spawning Options

When target is a rig:

```bash
gt sling gt-abc gastown --create     # Create polecat if missing
gt sling gt-abc gastown --naked      # No-tmux (manual start)
gt sling gt-abc gastown --force      # Ignore unread mail
gt sling gt-abc gastown --account work  # Specific Claude account
```

## Natural Language Args

Pass instructions to the executor:

```bash
gt sling gt-abc --args "patch release"
gt sling code-review --args "focus on security"
```

Stored in bead, shown via `gt prime`. The executor (LLM) interprets naturally.

## Formula Slinging

```bash
gt sling mol-release mayor/              # Cook + wisp + attach + nudge
gt sling towers-of-hanoi --var disks=3   # With variables
```

### Formula-on-Bead (--on flag)

Apply formula to existing work:

```bash
gt sling mol-review --on gt-abc        # Apply formula to existing bead
gt sling shiny --on gt-abc crew        # Apply formula, sling to crew
```

## Compare: sling vs hook vs handoff

| Command | Hooks | Starts Session | Keeps Context |
|---------|-------|----------------|---------------|
| `gt sling <bead>` | Yes | Yes | Yes |
| `gt hook <bead>` | Yes | No | Yes |
| `gt handoff <bead>` | Yes | Yes (new) | No (fresh) |

**Use sling when:** You want to assign and start now.
**Use hook when:** Just attaching, no immediate action.
**Use handoff when:** Fresh context needed (e.g., after compaction).

## Dispatch Flow

```
Orchestrator                     Gas Town                    Polecat
    |                               |                           |
    |  gt sling gt-abc gastown      |                           |
    | ----------------------------->|                           |
    |                               | Create polecat worktree   |
    |                               | Start tmux session        |
    |                               | ------------------------->|
    |                               |                           | SessionStart hook
    |                               |                           | gt hook -> finds work
    |                               |                           | Execute (Propulsion)
    |                               |                           |
    |  (returns immediately)        |                           | bd close gt-abc
    |                               |                           | ----------> beads
```

## No Output to Orchestrator

**Critical:** `gt sling` returns immediately. Polecat output stays in tmux.

```bash
gt sling gt-abc gastown
# -> Dispatched: gt-abc to gastown/polecats/nux
# Orchestrator context: ~20 tokens (vs full output from Task())
```

Use `gt convoy status` to monitor progress.

## Rig Selection

**Match bead prefix to rig** - the polecat works in that rig's codebase.

```bash
bd show <bead-id>              # 1. READ the bead
# ap-* -> ai-platform, gt-* -> gastown, etc.
gt sling <bead-id> <rig-name>  # 2. Sling to CORRECT rig
```

| Prefix | Rig | Prefix | Rig |
|--------|-----|--------|-----|
| `ap` | ai-platform | `kagent` | kagent |
| `bd` | beads | `kubic` | kubic-cm |
| `fractal` | fractal | `mam` | mcp_agent_mail |
| `gt` | gastown | `personal` | personal-site |
| `gitops` | gitops | `re` | release-engineering |
| `ho` | houston | `starport` | starport |

**Never pick a random rig.** Wrong rig = wrong code context.

## Error Handling

```bash
result=$(gt sling gt-abc gastown 2>&1)

if echo "$result" | grep -q "Error"; then
    echo "Dispatch failed: $result"
    # Fall back to Task() agent
fi
```

## Flags Reference

| Flag | Description |
|------|-------------|
| `--account <name>` | Claude Code account handle |
| `-a, --args <text>` | Natural language instructions |
| `--create` | Create polecat if missing |
| `-n, --dry-run` | Show what would be done |
| `--force` | Force spawn even with unread mail |
| `-m, --message <text>` | Context message |
| `--naked` | No-tmux mode (manual start) |
| `--no-convoy` | Skip auto-convoy creation |
| `--on <bead>` | Apply formula to existing bead |
| `-s, --subject <text>` | Context subject |
| `--var <key=value>` | Formula variable (repeatable) |

## Best Practices

1. **Read the bead first** - Know what codebase it affects
2. **Match prefix to rig** - Send to correct codebase
3. **Use auto-convoy** - Dashboard visibility by default
4. **Dispatch in parallel** - Multiple slings are independent
5. **Don't wait for completion** - Use convoy monitoring
6. **Create slingable beads** - Use BEADS_DIR for cross-rig creation
