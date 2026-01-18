---
description: Validate beads state matches reality
version: 1.0.0
argument-hint: [--fix|--dry-run]
model: haiku
---

# /beads-validate

Invoke the **beads** skill for validation mode.

## Arguments

| Argument | Purpose |
|----------|---------|
| `--fix` | Auto-fix discrepancies (with confirmation) |
| `--dry-run` | Show what would be fixed without changing anything |

## Execution

This command invokes the `beads` skill with validation arguments.

The skill handles:
- **State verification**: Check beads database against reality
- **Inconsistency detection**: Find epics with all children closed, stale in-progress issues, blocked issues ready to unblock
- **Fix application**: Apply fixes with user confirmation

## Validation Checks

1. **Complete epics not closed** - Epics where all children are closed
2. **Stale in-progress** - Issues in_progress for >24 hours without updates
3. **Blocked ready to unblock** - Blocked issues whose blockers are closed
4. **Doc issues with existing docs** - Doc issues where documentation exists

## Related

- **Skill**: `~/.claude/skills/beads/SKILL.md` (see Validation Workflow section)
- **Command**: `bd stats` for project health metrics
