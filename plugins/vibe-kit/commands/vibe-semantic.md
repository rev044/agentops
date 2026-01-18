---
description: Orchestrate semantic faithfulness analyses
version: 1.0.0
argument-hint: <target> [--only <types>]
model: opus
---

# /vibe-semantic

Invoke the **sk-vibe** skill for deep semantic analysis.

## Arguments

| Argument | Purpose |
|----------|---------|
| `<target>` | Directory or file to analyze |
| `--only <types>` | Limit to: docstrings,names,security,pragmatic,slop |

## Analysis Types

| Type | What It Checks |
|------|----------------|
| `docstrings` | Parameter/return/behavior claims |
| `names` | Function names match behavior |
| `security` | Validation theater, auth bypass |
| `pragmatic` | DRY, orthogonality, reversibility |
| `slop` | AI-generated boilerplate |

## Execution

LLM-powered semantic analysis using pattern files.

## Related

- **Skill**: `~/.claude/skills/sk-vibe/SKILL.md`
- **Command**: `/vibe-prescan` for fast static only
