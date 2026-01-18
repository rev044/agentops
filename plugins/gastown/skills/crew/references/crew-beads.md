# Crew Beads - Shared Issue Database

How crew workspaces interact with the beads issue tracking system.

## Beads Location

Each crew workspace has access to the rig's beads database:

```
<rig>/crew/<name>/.beads/
├── issues.jsonl       # Issue data
├── config.toml        # Beads configuration
└── routes.jsonl       # Prefix routing (if town-level)
```

### Shared vs Isolated

Crew workspaces typically share the rig's beads:

| Setup | Behavior |
|-------|----------|
| **Shared** (default) | All crew see same issues |
| **Isolated** | Each crew has own beads (rare) |

Most rigs use shared beads so crew members can coordinate.

---

## Beads Sync

Crew workspaces coordinate via `bd sync`:

```bash
# Commit beads changes to beads-sync branch
bd sync

# Check sync status
bd sync --status
```

### The beads-sync Branch

- Dedicated branch for beads coordination
- Separate from code branches
- Auto-merges on sync (append-only)

### Sync Workflow

```
Crew A: creates issue → bd sync → pushes beads-sync
Crew B: bd sync → pulls beads-sync → sees new issue
```

---

## Prefix Routing

Issue IDs route to the correct beads database:

| Prefix | Routes To |
|--------|-----------|
| `hq-*` | Town beads (`~/gt/.beads/`) |
| `gt-*` | Gastown rig beads |
| `ap-*` | AI-platform rig beads |
| etc. | Configured per rig |

### From Crew Perspective

```bash
# Working in gastown/crew/dave
bd create --title="Fix bug"
# Creates: gt-xxxx (gastown prefix)

# Can still view other rig issues
bd show ap-yyyy  # Routes to ai-platform beads
```

---

## Common Crew Beads Operations

### Creating Issues

```bash
# Create in current rig
bd create --title="Implement feature" --type=feature

# With priority
bd create --title="Fix bug" --type=bug --priority=1

# With description
bd create --title="Add tests" --type=task --body="Unit tests for auth module"
```

### Claiming Work

```bash
# See what's ready
bd ready

# Claim an issue
bd update gt-xxxx --status=in_progress
```

### Completing Work

```bash
# Close with reason
bd close gt-xxxx --reason="Implemented in commit abc123"

# Sync to share with team
bd sync
```

### Adding Comments

```bash
# Progress update
bd comments add gt-xxxx "Tests passing, starting integration"

# Block reason
bd comments add gt-xxxx "BLOCKED: Need API credentials"
```

---

## Multi-Crew Coordination

When multiple crew members work on the same rig:

### Avoiding Conflicts

1. **Sync frequently**: `bd sync` before starting work
2. **Claim before working**: `bd update --status=in_progress`
3. **Check ready issues**: `bd ready` shows unclaimed work

### Workflow Example

```bash
# Dave starts working
bd sync                          # Get latest
bd ready                         # See available work
bd update gt-123 --status=in_progress  # Claim it
bd sync                          # Share claim

# Emma checks for work
bd sync                          # Gets Dave's claim
bd ready                         # gt-123 not shown (Dave claimed it)
bd update gt-456 --status=in_progress  # Claim different issue
```

### Conflict Resolution

Beads conflicts are rare (append-only design), but if they occur:

```bash
# Accept theirs (merge both sets of changes)
git checkout --theirs .beads/issues.jsonl
git add .beads/issues.jsonl
git commit -m "merge: resolve beads conflict"
```

---

## Dependencies Between Issues

Crew members can set dependencies:

```bash
# Issue B depends on issue A (A blocks B)
bd dep add gt-456 gt-123

# View blocked issues
bd blocked

# View what blocks a specific issue
bd show gt-456  # Shows "Depends on: gt-123"
```

### Cross-Crew Dependencies

Crew members can depend on each other's work:

```
Dave working on gt-123 (auth)
Emma working on gt-456 (profile) which depends on gt-123

Emma: bd dep add gt-456 gt-123
Emma: bd ready  # gt-456 not shown until gt-123 closes
Dave: bd close gt-123 --reason="Done"
Dave: bd sync
Emma: bd sync
Emma: bd ready  # gt-456 now shows
```

---

## Best Practices

### Always Sync Before Starting

```bash
gt crew at dave
bd sync
bd ready
# Pick work
bd update <id> --status=in_progress
bd sync  # Share your claim
```

### Always Sync Before Stopping

```bash
bd close <id> --reason="Done"
bd sync
git push
gt crew stop dave
```

### Communicate Via Beads

```bash
# Instead of Slack/chat, use beads comments
bd comments add gt-123 "FYI: Changed approach, using JWT instead of sessions"

# Others see it when they sync
bd sync
bd show gt-123
```

### Use Dependencies for Coordination

```bash
# When your work blocks others
bd dep add <their-issue> <your-issue>

# When you're blocked
bd comments add <your-issue> "BLOCKED on <other-issue>, waiting for API changes"
```
