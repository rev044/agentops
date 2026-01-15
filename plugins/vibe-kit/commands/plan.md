---
description: Research codebase and create beads plan
version: 1.0.0
model: opus
argument-hint: <goal-description>
---

# /plan

Create structured implementation plans from goals.

Invoke the **sk-plan** skill to research the codebase, decompose the goal into
discrete features, create beads issues with proper dependencies, and output
an autopilot-ready summary.

## Arguments

| Argument | Purpose |
|----------|---------|
| `<goal>` | The goal to plan for (required) |

## Execution

Invokes the `sk-plan` skill with the provided goal.

The skill handles:
- 6-tier context discovery (semantic search, code-map, source, patterns, docs, web)
- Prior plan discovery to prevent duplicates
- Feature decomposition with dependency ordering
- Beads issue creation with epic-child relationships
- Wave computation for parallel execution
- Plan file generation and autopilot handoff

## Related

- **Skill**: `~/.claude/skills/sk-plan/SKILL.md`
- **Patterns**: `~/.claude/patterns/commands/plan/decomposition.md`
- **Templates**: `~/.claude/patterns/commands/plan/templates.md`
