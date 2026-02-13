# Skill Structure Standard

**Version:** 1.0.0
**Last Updated:** 2026-02-13
**Source:** Anthropic Official Skills Guide (docs/anthropic-skills-guide.md)
**Purpose:** Defines the required structure, frontmatter, and quality standards for all AgentOps skills.

---

## Table of Contents

1. [File Structure](#file-structure)
2. [YAML Frontmatter](#yaml-frontmatter)
3. [Description Field](#description-field)
4. [Body Structure](#body-structure)
5. [Progressive Disclosure](#progressive-disclosure)
6. [Quality Checklist](#quality-checklist)
7. [AgentOps Extensions](#agentops-extensions)

---

## File Structure

```
skill-name/
├── SKILL.md              # Required — exact case, no variations
├── scripts/              # Optional — executable code
├── references/           # Optional — progressive disclosure docs
└── assets/               # Optional — templates, fonts, icons
```

### Rules

| Rule | ALWAYS | NEVER |
|------|--------|-------|
| Entry point | `SKILL.md` (exact case) | `skill.md`, `SKILL.MD`, `Skill.md` |
| Folder name | kebab-case (`bug-hunt`) | spaces, underscores, capitals |
| Name match | Folder name = `name:` field | Mismatch between folder and frontmatter |
| README | None inside skill folder | `README.md` in skill directories |
| Reserved | Any valid kebab-case name | `claude-*` or `anthropic-*` prefixes |

---

## YAML Frontmatter

### Required Fields

```yaml
---
name: skill-name
description: 'What it does. When to use it. Trigger phrases.'
---
```

### Anthropic-Defined Optional Fields

| Field | Type | Max Length | Purpose |
|-------|------|-----------|---------|
| `license` | string | — | Open source license (MIT, Apache-2.0) |
| `allowed-tools` | string | — | Restrict tool access |
| `compatibility` | string | 500 chars | Environment requirements |
| `metadata` | object | — | Custom key-value pairs |

### AgentOps Extension Fields (under `metadata:`)

AgentOps uses these custom fields under `metadata:` for tooling integration:

```yaml
metadata:
  tier: solo          # solo, team, orchestration, library, background, meta
  dependencies:       # List of skill names this skill depends on
    - standards
    - council
  internal: true      # true for non-user-facing skills
  replaces: old-name  # Deprecated skill this replaces
```

**Tier values and their constraints:**

| Tier | Max Lines | Purpose |
|------|-----------|---------|
| `solo` | 200 | Single-agent, no spawning |
| `team` | 500 | Spawns workers |
| `orchestration` | 500 | Coordinates multiple skills/teams |
| `library` | 200 | Referenced by other skills, not invoked directly |
| `background` | 200 | Hooks/automation, not user-invoked |
| `meta` | 200 | Explains the system itself |

### Security Restrictions

- No XML angle brackets (`<` `>`) in frontmatter
- No `claude` or `anthropic` in skill names
- YAML safe parsing only (no code execution)

---

## Description Field

The description is the **most critical field** — it determines when Claude loads the skill.

### Structure

```
[What it does] + [When to use it] + [Key capabilities]
```

### Requirements

- Under 1024 characters
- MUST include trigger phrases users would actually say
- MUST explain what the skill does (not just when)
- No XML tags

### Good Examples

```yaml
# Specific + actionable + triggers
description: 'Investigate suspected bugs with git archaeology and root cause analysis. Triggers: "bug", "broken", "doesn''t work", "failing", "investigate bug".'

# Clear value prop + multiple triggers
description: 'Comprehensive code validation. Runs complexity analysis then multi-model council. Answer: Is this code ready to ship? Triggers: "vibe", "validate code", "check code", "review code", "is this ready".'
```

### Bad Examples

```yaml
# Too vague
description: Helps with projects.

# Missing triggers
description: Creates sophisticated multi-page documentation systems.

# Too technical, no user triggers
description: Implements the Project entity model with hierarchical relationships.
```

### Internal Skills Exception

Library/background/meta skills that are auto-loaded (not user-invoked) may describe their loading mechanism instead of user triggers:

```yaml
description: 'Auto-loaded by /vibe, /implement based on file types.'
```

---

## Body Structure

### Recommended Template

```markdown
---
name: skill-name
description: '...'
metadata:
  tier: solo
---

# Skill Name

## Quick Start

Example invocations showing common usage patterns.

## Instructions

### Step 1: [First Major Step]
Specific, actionable instructions with exact commands.

### Step 2: [Next Step]
...

## Examples

### Example 1: [Common scenario]
User says: "..."
Actions: ...
Result: ...

## Troubleshooting

### Error: [Common error]
Cause: ...
Solution: ...
```

### Requirements

| Aspect | Requirement |
|--------|-------------|
| Size | Under 5,000 words |
| Instructions | Specific and actionable (exact commands, not "validate the data") |
| Examples | At least 2-3 usage examples for user-facing skills |
| Error handling | Troubleshooting section for common failures |
| References | Link to `references/` for detailed docs (don't inline everything) |

---

## Progressive Disclosure

Skills use three levels:

1. **Frontmatter** — Always in system prompt. Minimal: name + description.
2. **SKILL.md body** — Loaded when skill is relevant. Core instructions.
3. **references/** — Loaded on-demand. Detailed docs, schemas, examples.

### Rules

- Keep SKILL.md focused on core workflow
- Move detailed reference material to `references/`
- Explicitly link to references: "Read `references/api-patterns.md` for..."
- Move scripts >20 lines to `scripts/` directory
- Move inline bash >30 lines to `scripts/` or `references/`

---

## Quality Checklist

### Before Commit

- [ ] `SKILL.md` exists (exact case)
- [ ] Folder name matches `name:` field
- [ ] Folder name is kebab-case
- [ ] Description includes WHAT + WHEN (triggers)
- [ ] Description under 1024 characters
- [ ] No XML tags in frontmatter
- [ ] No `claude`/`anthropic` in name
- [ ] `metadata.tier` is set and valid
- [ ] SKILL.md under 5,000 words
- [ ] User-facing skills have examples section
- [ ] User-facing skills have troubleshooting section
- [ ] Detailed docs in references/, not inlined
- [ ] No README.md in skill folder

### Trigger Testing

- [ ] Triggers on 3+ obvious phrases
- [ ] Triggers on paraphrased requests
- [ ] Does NOT trigger on unrelated topics

---

## AgentOps Extensions

These are AgentOps-specific patterns not in the Anthropic spec:

### Tier System

Controls line limits and categorization. Enforced by `tests/skills/lint-skills.sh`.

### Dependencies

Declared under `metadata.dependencies`. Validated by `tests/skills/validate-skill.sh`.

### Skill Tiers Document

Full taxonomy at `skills/SKILL-TIERS.md`.

### Standards Loading

Language standards loaded JIT by `/vibe`, `/implement` — see `standards-index.md`.
