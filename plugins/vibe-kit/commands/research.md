---
description: Deep codebase exploration to .agents/research/
version: 1.0.0
argument-hint: <topic-or-goal>
model: opus
---

# /research

Invoke the **sk-research** skill for deep codebase exploration.

## Arguments

| Argument | Purpose |
|----------|---------|
| `<topic>` | The topic or goal to research (required) |

## Execution

This command invokes the `sk-research` skill with the provided topic.

The skill handles:
- Prior art discovery (semantic + local search)
- Parallel sub-agent exploration
- Structured output document generation
- Workflow integration with `/plan`

**Output:** `.agents/research/YYYY-MM-DD-{topic-slug}.md`

**Next Step:** `/plan .agents/research/YYYY-MM-DD-{topic-slug}.md`

## Related

- **Skill**: `~/.claude/skills/sk-research/SKILL.md`
- **Standard**: `~/.claude/standards/command-skill-architecture.md`
