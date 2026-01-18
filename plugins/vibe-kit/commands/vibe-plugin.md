---
description: Validate plugins with L13 semantic verification
version: 1.0.0
argument-hint: <plugin-path> [--all] [--deep]
model: opus
---

# /vibe-plugin

Invoke the **sk-vibe** skill for plugin validation mode.

## Arguments

| Argument | Purpose |
|----------|---------|
| `<plugin-path>` | Path to plugin (skill, command, agent) |
| `--all` | Validate all plugins |
| `--commands` | Only validate commands |
| `--skills` | Only validate skills |
| `--deep` | Force LLM semantic analysis |

## Checks Performed

1. **Description Truthfulness** - Claims match implementation
2. **Trigger Accuracy** - Phrases handled correctly (skills)
3. **Argument Consistency** - Declared args are used (commands)
4. **Progressive Disclosure** - L1/L2/L3 content structure
5. **Painted Doors** - Documented features that don't exist

## Related

- **Skill**: `~/.claude/skills/sk-vibe/SKILL.md`
- **Script**: `~/.claude/scripts/validate-plugin.sh`
