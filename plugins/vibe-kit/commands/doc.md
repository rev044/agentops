---
description: Generate docs from code, configs, and templates
version: 1.0.0
argument-hint: [discover|gen|all|sync|coverage] [target]
model: haiku
---

# /doc

Invoke the **sk-doc** skill for documentation generation and validation.

## Arguments

| Argument | Purpose |
|----------|---------|
| `discover` | Scan for documentable features |
| `gen <feature>` | Generate/update specific doc |
| `all` | Update all documentation |
| `sync` | Pull from canonical source |
| `coverage` | Validate docs match code |

## Execution

Invokes the `sk-doc` skill which auto-detects project type (CODING, INFORMATIONAL, OPS) and routes to appropriate generator.

The skill handles:
- Project type detection via signal scoring
- Feature discovery and coverage analysis
- Type-specific doc generation
- Validation with exact issue counts

## Related

- **Skill**: `~/.claude/skills/sk-doc/SKILL.md`
- **Coverage**: `/doc-coverage` for detailed validation
