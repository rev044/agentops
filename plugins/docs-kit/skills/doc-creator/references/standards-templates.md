# Standards Documentation Templates

Detailed templates for standards, schemas, and vocabulary documentation.

---

## Coding Standard Template

Coding standards define rules for a specific language or technology.

### Full Template

```markdown
# [Language/Tool] Standards

> **Purpose:** [One sentence purpose]

## Scope

This document covers: [comma-separated list].

**Related:**
- [Other Standard](./...) - Description

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Version** | X.Y+ | `tool --version` |
| **Linter** | linter-name | `.linterrc` |
| **Threshold** | Value | Command |

---

## Required Patterns

### [Pattern Name]

Every [file/function/class] MUST:

```language
# Correct pattern
```

**Why:** [Technical justification]

### [Pattern 2]

...

---

## [Tool] Integration

### Repository Configuration

Create `.toolrc` at repo root:

```ini
# Configuration content
```

### Common Fixes

| Issue | Fix |
|-------|-----|
| [Error code] | [How to fix] |

---

## Validation

```bash
# Command to validate
tool check path/
```

---

## Exceptions

Document MUST comment why when deviating:

```language
# Disable reason: [justification]
# disable-rule
```

---

**Created:** YYYY-MM-DD
**Related:** [Corpus](../corpus/...) | [Methodology](../methodology/)
```

---

## Schema Documentation Template

Schemas define data structures with required fields and validation rules.

### Full Template

```markdown
# [Thing] Schema

**Purpose:** [What this schema defines and why]

**Related:**
- [Related Schema](./...) - Description
- [Validation Tool](./...) - How it's validated

---

## Why [Schema Name]?

| Feature | How It Works |
|---------|--------------|
| **Feature 1** | Description |
| **Feature 2** | Description |

---

## [Entity] Types

| Type | Use For | Example |
|------|---------|---------|
| `type-a` | Use case | `path/example` |
| `type-b` | Use case | `path/example` |

---

## Schema by Type

### type-a

[Description of when to use]

```yaml
---
field1: "value"
field2: value
---
```

| Field | Required | Description |
|-------|----------|-------------|
| `field1` | Yes | What this field does |
| `field2` | No | Optional field |

---

## Validation

```bash
# Command to validate
make validate-thing
```
```

---

## Tag Vocabulary Template

Tag vocabularies define controlled vocabularies for categorization.

### Full Template

```markdown
# Tag Vocabulary

**Purpose:** [Why tags matter]

---

## Tag Categories

### Category 1: [Name]

| Tag | Use For |
|-----|---------|
| `tag-a` | Description |
| `tag-b` | Description |

### Category 2: [Name]

...

---

## Usage Rules

1. **Rule 1:** Description
2. **Rule 2:** Description

---

## Adding New Tags

Before adding a new tag:

1. Check if existing tag covers the use case
2. Ensure tag is specific enough to be useful
3. Add to appropriate category
4. Update validation if applicable
```

---

## Normative Language (RFC 2119)

Standards MUST use RFC 2119 language for requirements:

| Term | Meaning |
|------|---------|
| **MUST** | Absolute requirement |
| **MUST NOT** | Absolute prohibition |
| **SHOULD** | Strong recommendation, exceptions need justification |
| **SHOULD NOT** | Strong discouragement |
| **MAY** | Optional, truly at developer discretion |

---

## Quick Start: Create New Standard

```bash
# 1. Create standard document
cat > docs/standards/NEW-standards.md << 'EOF'
# NEW Standards

> **Purpose:** [One sentence describing what this standard covers]

## Scope

This document covers: [list of topics].

**Related:**
- [Other Standard](./other-standards.md) - Brief description
- [Tag Vocabulary](./tag-vocabulary.md) - Documentation standards

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Tool** | toolname | `tool --check` |
| **Config** | .toolrc | At repo root |

---

## Required Patterns

### Pattern 1

Every file MUST:

```language
# Example code
```

**Why:** [Explanation]
EOF

# 2. Update README index
# Add row to docs/standards/README.md Standards Index table

# 3. Validate links
make check-links
```

---

## Update Standards Index

After creating any new standard, update the index in `docs/standards/README.md`:

```markdown
## Standards Index

| Language | Document | Gate |
|----------|----------|------|
| Python | [python-style-guide.md](./python-style-guide.md) | CC <= 10 |
| Shell/Bash | [shell-script-standards.md](./shell-script-standards.md) | shellcheck |
| **NEW** | [new-standards.md](./new-standards.md) | tool-name |
```

---

## Add Validation Integration

Every standard should have automated validation.

### Option 1: Makefile Target

```makefile
# In Makefile
validate-NEW: ## Validate NEW files
	@tool check path/
```

### Option 2: Validation Script

```bash
# In tools/scripts/validate-NEW.sh
#!/usr/bin/env bash
set -euo pipefail

# Validation logic here
tool check "$@"
```

### Option 3: Pre-commit Hook

```yaml
# In .pre-commit-config.yaml
repos:
  - repo: https://github.com/tool/tool
    rev: vX.Y.Z
    hooks:
      - id: tool-lint
        files: \.ext$
```

---

## Validation Checklist

Before committing new standards:

- [ ] Purpose statement is clear and concise
- [ ] Scope section lists what's covered
- [ ] Related links are valid
- [ ] Quick Reference table has validation column
- [ ] Normative language (MUST/SHOULD/MAY) used correctly
- [ ] Code examples are correct and copy-pasteable
- [ ] Validation command/tool documented
- [ ] README index updated
- [ ] `make check-links` passes

---

## Error Recovery

### "Standard not enforced" Issue

1. Add to pre-commit hooks
2. Add to CI pipeline
3. Add to Makefile targets
4. Document common violations with fixes

### "Standard conflicts with existing code" Issue

1. Document existing exceptions
2. Create migration path
3. Consider grandfather clause with deadline

### "Links broken" Issue

```bash
# Check all links
make check-links

# Fix relative paths
# From docs/standards/NEW.md to docs/corpus/
# Use: ../../corpus/ not ../corpus/
```
