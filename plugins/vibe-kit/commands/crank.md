---
description: Autonomous epic execution - runs until done (crew or mayor mode)
version: 2.0.0
argument-hint: <epic-id> [--mode=crew|mayor] [--dry-run] [--max N]
model: opus
---

# /crank

Fully autonomous epic execution. Runs until ALL children are CLOSED.

**Auto-detects role**: Mayor uses polecats (parallel), Crew executes directly (sequential).

## Arguments

| Arg | Purpose |
|-----|---------|
| `<epic-id>` | Epic to execute (e.g., `gt-abc`, `ap-1234`) |
| `--mode` | Force `crew` (sequential) or `mayor` (parallel) |
| `--dry-run` | Preview waves without executing |
| `--max <n>` | Limit concurrent polecats (default: 8, mayor only) |
| `status` | Show current crank progress |
| `stop` | Graceful stop at next checkpoint |

## Execution Modes

| Mode | Context | Execution | Best For |
|------|---------|-----------|----------|
| **Crew** | `~/gt/<rig>/crew/boden` | Sequential via `/implement` | Small epics, testing |
| **Mayor** | `~/gt` or `~/gt/<rig>/mayor` | Parallel via `gt sling` | Large epics, overnight |

## Key Properties

- **Never stops for human input** - escalates via mail instead
- **ODMCR loop** - Observe, Dispatch, Monitor, Collect, Retry
- **Auto-adapts** - detects role from current directory

## Examples

```bash
# From crew (sequential)
/crank ap-68ohb

# From mayor (parallel)
/crank ap-68ohb

# Force crew mode (even from mayor)
/crank ap-68ohb --mode=crew

# Force mayor mode with rig
/crank ap-68ohb --mode=mayor

# Dry run to preview
/crank ap-68ohb --dry-run
```

## Related

- **Skill**: `~/.claude/skills/sk-crank/SKILL.md`
- **ODMCR Loop**: `~/.claude/skills/sk-crank/odmcr.md`
- **Failure Taxonomy**: `~/.claude/skills/sk-crank/failure-taxonomy.md`
