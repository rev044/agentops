---
description: Create reusable formula template from goal
version: 1.0.0
argument-hint: <topic> [--immediate]
model: opus
---

# /formulate

Create reusable `.formula.toml` templates from research or goals.

## Arguments

| Arg | Purpose |
|-----|---------|
| `<topic>` | Goal or topic to formulate (required) |
| `--immediate` | Skip formula, create beads directly |

## When to Use

| Scenario | Command |
|----------|---------|
| Repeatable workflow | `/formulate` |
| One-time implementation | `/formulate --immediate` or `/plan` |

## Output

- **Formula**: `~/gt/.agents/<rig>/formulas/<topic>.formula.toml`
- **Issues**: Via `bd cook <formula>` (or immediate with `--immediate`)

## Execution

This command invokes the **sk-formulate** skill.

```bash
/formulate "release checklist"           # Create reusable formula
/formulate "fix auth bug" --immediate    # Create beads directly
```

## Workflow

```
/research → /formulate → bd cook <formula> → bd mol pour
```

## Related

- **Skill**: `~/.claude/skills/sk-formulate/SKILL.md`
- **Templates**: `~/.claude/skills/sk-formulate/references/templates.md`
- **Molecules**: `~/.claude/skills/beads/references/MOLECULES.md`
