# gt convoy - Batch Work Tracking

**Manage convoys - the primary unit for tracking batched work.**

## Synopsis

```bash
gt convoy create "name" <beads>      # Create convoy
gt convoy add <convoy-id> <beads>    # Add more issues
gt convoy list                       # Dashboard view
gt convoy status <id>                # Detailed progress
```

## What Is A Convoy?

A convoy is a **persistent tracking unit** that monitors related issues across
rigs. When you kick off work (even a single issue), a convoy tracks it so
you can see when it lands and what was included.

### Convoy vs Swarm

| Concept | What It Is | ID | Lifecycle |
|---------|------------|-------|-----------|
| **Convoy** | Persistent tracking unit | `hq-cv-*` | Created → Auto-closes when done |
| **Swarm** | Workers assigned to convoy issues | Uses convoy ID | Ephemeral - dissolves when done |

## Why Convoys Matter

1. **Dashboard visibility** - `gt convoy list` shows all active work
2. **Progress tracking** - See which issues are done vs pending
3. **Auto-close notification** - Know when batch completes
4. **Cross-rig capable** - Track issues across multiple rigs

## Basic Usage

### Create a Convoy

```bash
# Track multiple issues together
gt convoy create "Feature X" gt-abc gt-def gt-ghi
# -> Created: hq-cv-xyz tracking 3 issues

# With notification on completion
gt convoy create "Feature X" gt-abc gt-def --notify
```

### View Convoys

```bash
# Dashboard (primary view)
gt convoy list

# Detailed progress for one convoy
gt convoy status hq-cv-xyz
```

### Add Issues to Convoy

```bash
# Add more issues (reopens if closed)
gt convoy add hq-cv-xyz gt-jkl gt-mno
```

## Tracking Semantics

- **Non-blocking**: Tracked issues don't block the convoy
- **Cross-prefix**: Convoy in `hq-*` can track `gt-*`, `ap-*`, etc.
- **Landed**: All tracked issues closed → notification sent

## Auto-Convoy (via gt sling)

When slinging a single issue, `gt sling` auto-creates a convoy:

```bash
gt sling gt-abc gastown              # Creates "Work: <title>" convoy
gt sling gt-abc gastown --no-convoy  # Skip auto-convoy
```

This ensures even "swarm of one" assignments appear in the dashboard.

## Subcommands

| Subcommand | Description |
|------------|-------------|
| `gt convoy create` | Create a new convoy |
| `gt convoy add` | Add issues to existing convoy |
| `gt convoy status` | Show convoy progress |
| `gt convoy list` | Dashboard of all convoys |
| `gt convoy check` | Check and auto-close completed convoys |
| `gt convoy stranded` | Find convoys with ready work but no workers |

## Convoy Lifecycle

```
1. Create Convoy
   gt convoy create "Wave 1" gt-abc gt-def gt-ghi
   -> hq-cv-xyz (status: open)

2. Dispatch Work
   gt sling gt-abc gastown
   gt sling gt-def gastown
   gt sling gt-ghi gastown
   -> Workers assigned, convoy tracks progress

3. Work Progresses
   (polecat closes gt-abc)
   (polecat closes gt-def)
   -> Convoy shows partial completion

4. All Issues Complete
   (polecat closes gt-ghi)
   -> Convoy auto-closes
   -> Notification sent (if --notify)

5. Reopenable
   gt convoy add hq-cv-xyz gt-new
   -> Convoy reopens with new issue
```

## Dashboard Output

```
$ gt convoy list

CONVOY           | STATUS | PROGRESS | ISSUES
-----------------+--------+----------+------------------
Feature X        | active | 2/3      | gt-abc, gt-def, gt-ghi
Auth Refactor    | active | 0/5      | ap-001..ap-005
Done: Q4 Sprint  | closed | 12/12    | various
```

## Detailed Status

```
$ gt convoy status hq-cv-xyz

Convoy: Feature X (hq-cv-xyz)
Status: active
Progress: 2/3 (66%)

Issues:
  [x] gt-abc - Fix login bug (closed)
  [x] gt-def - Add OAuth support (closed)
  [ ] gt-ghi - Update docs (in_progress)

Workers (Swarm):
  gastown/Toast -> gt-ghi
```

## Use Cases

### Parallel Wave Dispatch

```bash
# Create convoy for wave
gt convoy create "Wave 1" gt-abc gt-def gt-ghi

# Dispatch in parallel
gt sling gt-abc gastown
gt sling gt-def gastown
gt sling gt-ghi gastown

# Monitor progress
gt convoy list
gt convoy status hq-cv-xyz
```

### Multi-Rig Coordination

```bash
# Track issues across rigs
gt convoy create "Full Stack Feature" ap-001 gt-002 ho-003

# Each to correct rig
gt sling ap-001 ai-platform
gt sling gt-002 gastown
gt sling ho-003 houston

# Single dashboard view
gt convoy status <id>
```

### Finding Stalled Work

```bash
# Convoys with ready work but no active workers
gt convoy stranded

# Results:
# hq-cv-abc: 2 issues ready, 0 workers
#   -> gt-def (ready), gt-ghi (ready)
```

## Flags Reference

### convoy create

| Flag | Description |
|------|-------------|
| `--notify` | Send notification when convoy completes |

### convoy list

| Flag | Description |
|------|-------------|
| `-i, --interactive` | Interactive tree view |

## Best Practices

1. **Create convoy before dispatch** - Enables tracking from start
2. **Use auto-convoy** - Single slings get visibility too
3. **Check stranded** - Find work that needs workers
4. **Monitor via list** - Primary dashboard view
5. **Group related work** - One convoy per logical batch
