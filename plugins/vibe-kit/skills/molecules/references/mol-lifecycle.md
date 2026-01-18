# Molecule Lifecycle

The molecule system uses a chemistry metaphor to describe workflow phases
from template creation through execution and cleanup.

## Phase Overview

```
Formula (.toml)
    |
    | bd cook
    v
  Proto (solid)
    |
    |---> bd mol pour ---> Mol (liquid) ---> Work complete
    |
    |---> bd mol wisp ---> Wisp (vapor)
                              |
                              |---> bd mol squash ---> Digest
                              |
                              |---> bd mol burn ---> Nothing
```

## Phase Details

| Phase | Name | Storage | Synced | Persistence |
|-------|------|---------|--------|-------------|
| File | Formula | `.formula.toml` | Git | Permanent |
| Solid | Proto | `.beads/` | Yes | Permanent template |
| Liquid | Mol | `.beads/` | Yes | Persistent issues |
| Vapor | Wisp | `.beads-wisp/` | No | Ephemeral |
| Digest | Summary | `.beads/` | Yes | Permanent record |

## Phase Transitions

### Formula -> Proto (cook)

Transform a TOML template into a proto bead:

```bash
# Preview first
bd cook workflow.formula.toml --dry-run

# Compile-time (keep {{variables}})
bd cook workflow.formula.toml

# Runtime (substitute variables)
bd cook workflow.formula.toml --mode=runtime --var key=value

# Persist to database
bd cook workflow.formula.toml --persist
```

### Proto -> Mol (pour)

Create persistent issues from a proto:

```bash
# Basic pour
bd mol pour proto-name

# With variable substitution
bd mol pour proto-name --var component=auth

# Run: pour + assign + pin
bd mol run proto-name --var version=2.0
```

**Use `pour` when:**
- Work needs audit trail
- Cross-session persistence required
- Team collaboration needed

### Proto -> Wisp (wisp)

Create ephemeral workflow:

```bash
# Create wisp
bd mol wisp proto-name

# With variables
bd mol wisp proto-name --var target=db

# List wisps
bd mol wisp list
```

**Use `wisp` when:**
- Operational loops (patrols, health checks)
- Diagnostic runs
- High-frequency ephemeral work
- No archival value needed

### Wisp -> Digest (squash)

Compress wisp execution into a permanent summary:

```bash
# Auto-generate summary
bd mol squash wisp-id

# Provide summary
bd mol squash wisp-id --summary "Patrol complete: 3 issues found"

# Keep children (just create digest)
bd mol squash wisp-id --keep-children

# Preview
bd mol squash wisp-id --dry-run
```

### Wisp -> Nothing (burn)

Delete wisp without any trace:

```bash
bd mol burn wisp-id
```

**Use `burn` when:**
- Routine work with no archival value
- Failed/aborted runs
- Clean slate needed

### Ad-hoc Epic -> Proto (distill)

Extract a reusable template from completed work:

```bash
# Basic distill
bd mol distill epic-id --as "Release Workflow"

# With variable extraction
bd mol distill epic-id --var feature_name=auth --var version=1.0.0
```

**Use `distill` when:**
- Team develops good workflow organically
- Capturing tribal knowledge
- Creating template from real execution

## Bonding (Combining Workflows)

Combine protos or molecules:

```bash
# Sequential: B runs after A completes
bd mol bond A B

# Parallel: B runs alongside A
bd mol bond A B --type parallel

# Conditional: B runs only if A fails
bd mol bond A B --type conditional
```

**Operand combinations:**

| A | B | Result |
|---|---|--------|
| proto | proto | Compound proto |
| proto | mol | Spawn proto, attach to mol |
| mol | proto | Spawn proto, attach to mol |
| mol | mol | Join into compound |

## Lifecycle Best Practices

### Starting Work

```bash
# Option 1: Ephemeral (default spawn)
bd mol spawn proto-name

# Option 2: Persistent
bd mol pour proto-name

# Option 3: Durable with recovery (pour + assign + pin)
bd mol run proto-name
```

### During Execution

```bash
# Check progress
bd mol progress mol-id

# Show current position
bd mol current mol-id

# Bond additional work
bd mol bond mol-id additional-proto
```

### Ending Work

```bash
# For wisps - squash or burn
bd mol squash wisp-id --summary "Completed successfully"
# or
bd mol burn wisp-id

# For mols - just close the epic
bd close mol-epic-id

# Garbage collect orphaned wisps
bd mol wisp gc
```

### Detecting Stale Mols

```bash
# Find complete-but-unclosed molecules
bd mol stale
```

## Common Patterns

### Weekly Review Workflow

```bash
# Create proto once
bd create "Weekly Review" --type epic --label template
# Add children...

# Use each week
bd mol pour mol-weekly-review
```

### Ephemeral Patrol

```bash
bd mol wisp mol-patrol
# Execute patrol...
bd mol squash wisp-id --summary "Patrol complete"
```

### Feature with Rollback

```bash
bd mol spawn mol-deploy --attach mol-rollback --attach-type conditional
# If deploy fails, rollback becomes unblocked
```

## Related

- **Formula syntax**: See `formula-toml.md`
- **Cooking details**: See `cooking.md`
- **Wisp patterns**: See `wisp-patterns.md`
