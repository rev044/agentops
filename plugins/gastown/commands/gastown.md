---
description: Gas Town status and utility operations
version: 1.0.0
argument-hint: [status|peek|convoy|resume]
---

# /gastown

Orchestrate multi-agent work via Gas Town polecats with context isolation.

Invoke the **sk-gastown** skill for parallel execution without context bloat.

## Arguments

| Argument | Purpose |
|----------|---------|
| `"goal"` | Full R→P→I workflow with polecats |
| `<epic-id>` | Execute existing epic |
| `status` | Show active convoys and polecats |
| `resume` | Continue from checkpoint |
| `peek <polecat>` | Investigate specific polecat |
| `stop` | Stop all polecats |

## Options

| Option | Purpose |
|--------|---------|
| `--full` | Enable Research→Plan→Implement |
| `--rig <name>` | Target rig (default: auto-detect) |
| `--max <n>` | Max concurrent polecats (default: 8) |
| `--dry-run` | Preview without executing |
| `--resume` | Explicit resume flag |

## Examples

```bash
# Full R→P→I workflow
/gastown "Add OAuth support to the API"

# Execute existing epic
/gastown gt-tq9

# Check what's running
/gastown status

# Resume after checkpoint
/gastown resume

# Preview waves
/gastown gt-tq9 --dry-run

# Limited parallelism
/gastown gt-tq9 --max 4
```

## Execution

Invokes `sk-gastown` skill which:
1. Creates convoy for tracking
2. Dispatches work via `gt sling`
3. Monitors via `gt convoy status`
4. Reads results from beads
5. Reports progress

## Related

- **Skill:** `~/.claude/skills/sk-gastown/SKILL.md`
- **Beads:** `bd` commands for issue tracking
- **Gas Town:** `gt` commands for orchestration
- **Autopilot:** `/autopilot` for Task-based execution
