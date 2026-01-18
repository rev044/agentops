---
description: Validate code does what it claims
version: 1.0.0
argument-hint: <target> [recent|all|path]
model: opus
---

# /vibe

Invoke the **sk-vibe** skill for full semantic code validation.

## Arguments

| Argument | Purpose |
|----------|---------|
| `recent` | Files from last commit |
| `all` | Full codebase (sampled) |
| `<path>` | Specific directory or file |

## Execution

Orchestrates prescan + semantic analysis:

1. **Prescan** - Fast static checks (P1, P4, P5, P8, P9, P12)
2. **Semantic** - LLM-powered analysis (docstrings, names, security, pragmatic, slop)
3. **Report** - JSON, JUnit XML, assessment artifact
4. **Issues** - Create beads for CRITICAL/HIGH findings

## Related

- **Skill**: `~/.claude/skills/sk-vibe/SKILL.md`
- **Commands**: `/vibe-prescan`, `/vibe-semantic`, `/vibe-plugin`
