# Dispatch Patterns

## Overview

Dispatch work to isolated polecat sessions via `gt sling`. Each polecat runs in its own tmux session with independent context.

## Rig Selection: Match Bead to Codebase

**Before slinging, determine which rig OWNS the work:**

```bash
# 1. READ the bead first
bd show <bead-id>

# 2. Match bead prefix to rig
#    - ap-* beads → athena rig
#    - gt-* beads → daedalus rig
#    - ho-* beads → argus rig

# 3. Sling to the CORRECT rig
gt sling <bead-id> <rig-name>
```

| Prefix | Rig | Prefix | Rig |
|--------|-----|--------|-----|
| `ap` | athena | `he` | hephaestus |
| `be` | chronicle | `kubic` | kubic-cm |
| `fr` | cyclopes | `mam` | mcp_agent_mail |
| `gt` | daedalus | `personal` | personal-site |
| `gitops` | gitops | `re` | release-engineering |
| `ho` | argus | `starport` | starport |
| `jc` | jren_cm | `vibe` | vibe-check |

**Never pick a random rig.** The polecat works in that rig's codebase. Wrong rig = wrong code context.

## Basic Dispatch

```bash
# Dispatch single issue to rig
gt sling gt-abc daedalus

# What happens:
# 1. Creates polecat worktree: ~/gt/daedalus/polecats/<name>/
# 2. Starts tmux session with Claude Code
# 3. Hooks issue to polecat
# 4. SessionStart hook runs → polecat finds work → executes
```

## Batch Dispatch

```bash
# Dispatch wave of issues
wave_issues="gt-abc gt-def gt-ghi"

# Create convoy first (for tracking)
gt convoy create "Wave 1" $wave_issues

# Dispatch each issue
for issue in $wave_issues; do
    gt sling $issue daedalus
done
```

## Multi-Rig Dispatch

When dispatching issues across different rigs:

```bash
# Each issue goes to the rig that owns its codebase
gt sling ap-123 athena         # API changes
gt sling gt-456 daedalus       # Gas Town changes
gt sling ho-789 argus          # Argus changes

# All three polecats work in parallel, each in the correct codebase
```

## Cross-Rig Routing

For the orchestrator to see beads across rigs:

```bash
# Register rigs for visibility (done once per rig)
bd repo add ~/gt/athena/mayor/rig
bd repo add ~/gt/daedalus/mayor/rig

# Sync to hydrate rig beads into town view
bd repo sync

# Now you can see/sling rig beads from town
bd show ap-123      # Works after repo sync
gt sling ap-123 athena
```

### Troubleshooting "bead not found"

```bash
# 1. Verify bead exists in rig
bash -c 'cd ~/gt/daedalus/mayor/rig && bd show <id>'

# 2. Sync the rig's beads
bash -c 'cd ~/gt/daedalus/mayor/rig && bd sync'

# 3. Re-sync multi-repo from town
bd repo sync

# 4. Retry
gt sling <id> daedalus
```

## Dispatch Options

```bash
# With molecule (workflow template)
gt sling gt-abc daedalus --molecule=shiny

# Limit concurrent polecats
gt sling gt-abc daedalus --max-polecats=4

# Specify polecat name
gt sling gt-abc daedalus --name=oauth-worker
```

## Convoy Creation

Always create convoy before dispatching:

```bash
# Create convoy tracking issues
gt convoy create "Feature X Implementation" gt-abc gt-def gt-ghi
# → Created: hq-cv-xyz

# With notification on completion
gt convoy create "Feature X" gt-abc gt-def --notify
```

## Dispatch Flow

```
Orchestrator                     Gas Town                    Polecat
    │                               │                           │
    │  gt convoy create             │                           │
    │ ─────────────────────────────▶│                           │
    │                               │ Create convoy bead        │
    │  ◀──────────────────────────  │                           │
    │  hq-cv-xyz                    │                           │
    │                               │                           │
    │  gt sling gt-abc daedalus     │                           │
    │ ─────────────────────────────▶│                           │
    │                               │ Create polecat worktree   │
    │                               │ Start tmux session        │
    │                               │ ─────────────────────────▶│
    │                               │                           │ SessionStart hook
    │                               │                           │ gt hook → finds work
    │                               │                           │ Execute issue
    │                               │                           │
    │  (no output - polecat works   │                           │ bd close gt-abc
    │   in isolation)               │                           │ ─────────────────▶ beads
```

## No Output to Orchestrator

**Critical:** `gt sling` returns immediately. Polecat output stays in its tmux session.

```bash
# This returns immediately
gt sling gt-abc daedalus
# → Dispatched: gt-abc to daedalus/polecats/nux

# Orchestrator context: ~20 tokens
# vs Task() which would return full implementation output
```

## Error Handling

```bash
# Check if dispatch succeeded
result=$(gt sling gt-abc daedalus 2>&1)

if echo "$result" | grep -q "Error"; then
    echo "Dispatch failed: $result"
    # Fall back to Task() agent
fi
```

## Creating Slingable Beads (Mayor Pattern)

⚠️ **HQ beads (`hq-*`) CANNOT be hooked by polecats!**

`gt sling` uses `bd update` which lacks cross-database routing. Beads must exist
in the target rig's database to be hookable.

| Work Type | Create From | Gets Prefix | Can Sling? |
|-----------|-------------|-------------|------------|
| Mayor coordination | `~/gt` | `hq-*` | ❌ No |
| Rig bug/feature | Rig's beads | `gt-*`, `ap-*`, etc. | ✅ Yes |

**From Mayor, to create a slingable bead:**

```bash
# Use BEADS_DIR to target the rig's beads database
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="Fix X" --type=bug
# Creates: gt-xxxxx (daedalus prefix)

BEADS_DIR=~/gt/athena/mayor/rig/.beads bd create --title="Add Y" --type=feature
# Creates: ap-xxxxx (athena prefix)
```

**Then sling normally:**
```bash
gt sling gt-xxxxx daedalus   # Works - bead is in daedalus's database
gt sling ap-xxxxx athena     # Works - bead is in athena's database
```

## Gotchas

**Wrong prefix for rig work.** Creating from `~/gt` gives `hq-*` which polecats can't hook.
- WRONG: `bd create --title="daedalus bug"` → `hq-xxx` (unhookable)
- RIGHT: `BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create ...` → `gt-xxx` (slingable)

**Temporal language inverts dependencies.** "Phase 1 blocks Phase 2" is backwards.
- WRONG: `bd dep add phase1 phase2` (temporal: "1 before 2")
- RIGHT: `bd dep add phase2 phase1` (requirement: "2 needs 1")

**Rule**: Think "X needs Y", not "X comes before Y". Verify with `bd blocked`.

## Best Practices

1. **Read the bead first** - Understand what codebase it affects
2. **Match prefix to rig** - Send work to the correct codebase
3. **Create convoy first** - Enables dashboard tracking
4. **Dispatch in parallel** - Multiple gt sling calls are independent
5. **Don't wait for completion** - Use convoy monitoring instead
6. **Handle dispatch errors** - Fall back to Task() if Gas Town unavailable
7. **Use BEADS_DIR from Mayor** - Create beads in target rig's database for slinging
