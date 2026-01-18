---
description: Fast static scan for 6 failure patterns
version: 1.0.0
argument-hint: <target> [recent|all|path]
model: haiku
---

# /vibe-prescan

Invoke the **sk-vibe** skill for fast static prescan only.

## Arguments

| Argument | Purpose |
|----------|---------|
| `recent` | Files from last commit (default) |
| `all` | Full codebase |
| `<path>` | Specific directory or file |

## Patterns Detected

| ID | Pattern | Severity |
|----|---------|----------|
| P1 | Phantom Modifications | CRITICAL |
| P4 | Invisible Undone | HIGH |
| P5 | Eldritch Horror | HIGH |
| P8 | Cargo Cult Error Handling | HIGH |
| P9 | Documentation Phantom | MEDIUM |
| P12 | Zombie Code | MEDIUM |

## Execution

Fast static analysis - no LLM required.

## Related

- **Skill**: `~/.claude/skills/sk-vibe/SKILL.md`
- **Command**: `/vibe` for full validation
