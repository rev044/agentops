---
name: molecules
description: >
  This skill should be used when the user asks to "workflow template",
  "formula", "molecule", "wisp", "proto", or needs guidance on
  workflow templates and formula TOML for reusable agent patterns.
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - beads
---

# Molecules Skill

Workflow templates and formula TOML for reusable agent patterns.

## Overview

Molecules are bd's chemistry-inspired system for reusable work templates and
ephemeral workflows. This skill provides reference documentation for:

- **Formulas**: TOML templates that define workflow structure
- **Protos**: Compiled formulas (solid phase) - reusable templates
- **Mols**: Instantiated workflows (liquid phase) - persistent issues
- **Wisps**: Ephemeral workflows (vapor phase) - temporary, no audit trail

**When to Use**: Creating reusable workflow patterns, understanding bd's
molecule system, cooking formulas, managing wisps.

**When NOT to Use**: Simple one-off tasks (use beads directly), research
(use research).

---

## Quick Reference

### Phase Transitions

| Phase | Name | Storage | Synced? | Use Case |
|-------|------|---------|---------|----------|
| Solid | Proto | `.beads/` | Yes | Reusable template |
| Liquid | Mol | `.beads/` | Yes | Persistent instance |
| Vapor | Wisp | `.beads-wisp/` | No | Ephemeral instance |

**Transitions:**
- `cook`: Formula file (.toml) -> Proto
- `pour`: Proto -> Mol (persistent)
- `wisp`: Proto -> Wisp (ephemeral)
- `squash`: Wisp -> Digest (permanent summary)
- `burn`: Wisp -> Nothing (deleted)
- `distill`: Ad-hoc epic -> Proto

### Essential Commands

| Command | Purpose |
|---------|---------|
| `bd formula list` | List available formulas/protos |
| `bd cook <file.toml>` | Transform formula to proto |
| `bd mol pour <proto>` | Create persistent mol |
| `bd mol wisp <proto>` | Create ephemeral wisp |
| `bd mol squash <wisp>` | Compress to digest |
| `bd mol burn <wisp>` | Delete without trace |
| `bd mol distill <epic>` | Extract proto from work |

---

## Instructions

### When to Use Molecules

**Use Protos/Mols when:**
- Repeatable patterns (releases, reviews, onboarding)
- Encoding tribal knowledge as templates
- Audit trail matters
- Cross-session persistence needed

**Use Wisps when:**
- Operational loops (patrol cycles, health checks)
- One-shot orchestration
- Diagnostic runs
- High-frequency ephemeral work

### Basic Workflow

```bash
# 1. Create formula file
# See references/formula-toml.md for syntax

# 2. Cook formula (preview first)
bd cook path/to/workflow.formula.toml --dry-run

# 3a. Create persistent mol
bd mol pour workflow-name --var key=value

# 3b. Or create ephemeral wisp
bd mol wisp workflow-name --var key=value

# 4. Execute work...

# 5. End wisp (if applicable)
bd mol squash <wisp-id> --summary "Completed"
# or
bd mol burn <wisp-id>  # No trace
```

---

## References

### JIT-Loadable Documentation

| Topic | Reference |
|-------|-----------|
| Formula TOML syntax | `references/formula-toml.md` |
| Molecule lifecycle | `references/mol-lifecycle.md` |
| Wisp patterns | `references/wisp-patterns.md` |
| Cooking formulas | `references/cooking.md` |

### Related Skills

- **beads**: Core issue tracking
- **formulate**: Creating formulas from goals
- **crank**: Executing molecule work

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Use mols for one-off work | Create beads directly |
| Use wisps for auditable work | Use mols (persistent) |
| Forget to squash wisps | `bd mol squash` or `bd mol burn` |
| Hardcode values in formulas | Use `{{variables}}` |
| Skip `--dry-run` | Preview first, cook second |
