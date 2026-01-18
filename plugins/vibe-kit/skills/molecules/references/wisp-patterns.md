# Wisp Patterns

Wisps are ephemeral molecules that provide workflow structure without
cluttering the permanent database. They're stored in `.beads-wisp/`
and are NOT synced to git.

## When to Use Wisps

### Good Use Cases

| Pattern | Example |
|---------|---------|
| Operational loops | Patrol cycles, health checks, monitoring |
| One-shot orchestration | Temporary coordination tasks |
| Diagnostic runs | Debugging workflows, troubleshooting |
| High-frequency work | Routine tasks that would create noise |
| Experimental work | Prototyping workflows before committing |

### When NOT to Use Wisps

| Scenario | Use Instead |
|----------|-------------|
| Audit trail needed | `bd mol pour` (persistent mol) |
| Cross-session work | Persistent mol |
| Team collaboration | Persistent mol |
| Historical reference | Persistent mol + digest |

**Key insight:** Wisps prevent database bloat from routine operations while
still providing structure during execution.

## Creating Wisps

```bash
# From proto
bd mol wisp proto-name

# With variable substitution
bd mol wisp proto-name --var target=database --var env=staging

# spawn defaults to wisp
bd mol spawn proto-name  # Same as wisp
```

## Managing Wisps

### List Active Wisps

```bash
bd mol wisp list
bd mol wisp list --json  # Machine-readable
```

### Check Wisp Progress

```bash
bd mol progress wisp-id
bd mol current wisp-id
```

## Ending Wisps

### Squash (Create Digest)

Compresses wisp execution into a permanent summary:

```bash
# Auto-generate summary from child issues
bd mol squash wisp-id

# Provide explicit summary
bd mol squash wisp-id --summary "Patrol complete: 3 issues found, 2 resolved"

# Keep children, just create digest
bd mol squash wisp-id --keep-children

# Preview before squashing
bd mol squash wisp-id --dry-run
```

**When to squash:**
- Work had notable findings worth recording
- Pattern identified for future reference
- Summary useful for handoff

### Burn (Delete Without Trace)

```bash
bd mol burn wisp-id
```

**When to burn:**
- Routine work with no archival value
- Failed/aborted runs
- Clean runs with nothing to report

### Garbage Collection

Clean up orphaned wisps:

```bash
bd mol wisp gc
```

## Common Wisp Patterns

### Pattern: Patrol Cycle

Regular operational check that shouldn't clutter history:

```bash
# Create patrol wisp
bd mol wisp mol-patrol

# Execute patrol work...
# - Check system health
# - Review alerts
# - Triage issues

# End patrol
bd mol squash wisp-id --summary "Patrol: 2 alerts cleared, 1 new issue filed"
```

### Pattern: Diagnostic Investigation

Temporary debugging workflow:

```bash
# Start investigation
bd mol wisp mol-debug --var issue=connection-timeout

# Run diagnostics...
# - Collect logs
# - Check metrics
# - Test connections

# Found the problem
bd mol squash wisp-id --summary "Root cause: DNS TTL too high"

# Or if nothing found
bd mol burn wisp-id
```

### Pattern: Ephemeral Review

Quick review that doesn't need permanent record:

```bash
bd mol wisp mol-code-review --var pr=123

# Review...

# No major findings
bd mol burn wisp-id
```

### Pattern: Compound Wisp

Attach additional work to a wisp:

```bash
# Start main workflow
bd mol wisp mol-deploy

# Need additional checks
bd mol bond wisp-id mol-extra-validation --type parallel

# Complete
bd mol squash wisp-id --summary "Deploy verified"
```

### Pattern: Promote to Persistent

If ephemeral work becomes important:

```bash
# Started as wisp
bd mol wisp mol-feature

# Work is more significant than expected...
# Squash to create persistent digest
bd mol squash wisp-id --keep-children --summary "Important discovery"
# Children become persistent issues
```

## Wisp vs Mol Decision Tree

```
Is audit trail required?
├─ Yes → Use mol (pour)
└─ No
   ├─ Is this routine/operational?
   │  └─ Yes → Use wisp
   ├─ Will this span multiple sessions?
   │  └─ Yes → Use mol (pour)
   ├─ Does team need to see this?
   │  └─ Yes → Use mol (pour)
   └─ Default → Use wisp (can promote later)
```

## Troubleshooting

**"Wisp not found"**
- Wisps stored in `.beads-wisp/` (separate from `.beads/`)
- Check `bd mol wisp list` for active wisps

**"Wisp commands fail"**
- Ensure `.beads-wisp/` directory exists
- Check for file permissions

**"Too many wisps"**
- Run `bd mol wisp gc` to clean up
- Consider squashing completed wisps

## Related

- **Molecule lifecycle**: See `mol-lifecycle.md`
- **Formula syntax**: See `formula-toml.md`
- **Cooking**: See `cooking.md`
