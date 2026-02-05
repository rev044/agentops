---
name: judge
deprecated: true
replaced_by: council
description: 'DEPRECATED: Use /council instead. Multi-model validation council.'
---

# Judge Skill (DEPRECATED)

> **This skill is deprecated.** Use `/council` instead.

## Migration

| Old | New |
|-----|-----|
| `/judge recent` | `/council validate recent` |
| `/judge 2 opus` | `/council recent` |
| `/judge 3 opus` | `/council --deep recent` |
| `/judge --models=opus,codex` | `/council --mixed recent` |

See `skills/council/SKILL.md` for the full specification.
