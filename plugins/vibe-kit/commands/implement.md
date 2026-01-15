---
description: Implement highest-priority unblocked issue
version: 1.0.0
argument-hint: [issue-id]
model: opus
allowed-tools: Bash, Edit, Glob, Grep, Read, Task, Write
---

# /implement

Invoke the **sk-implement** skill for executing a single beads issue with full lifecycle.

## Arguments

| Argument | Purpose | Default |
|----------|---------|---------|
| `[issue-id]` | Specific issue to implement | Auto-select from `bd ready` |

## Execution

This command invokes the `sk-implement` skill with the provided arguments.

The skill handles:
- Context discovery (6-tier hierarchy)
- Issue selection and status updates
- Implementation with progress tracking
- Mandatory test verification
- Closure, commit, and sync

**Target Issue:** $ARGUMENTS (or auto-select highest priority from `bd ready`)

## Related

- **Skill**: `~/.claude/skills/sk-implement/SKILL.md`
- **Patterns**: `~/.claude/patterns/commands/implement/` (lifecycle, validation)
- **Beads**: `~/.claude/skills/beads/SKILL.md`
