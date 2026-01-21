# Markdown Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical Markdown standards for vibe skill validation

---

## Table of Contents

1. [AI-Agent Optimization](#ai-agent-optimization)
2. [Document Structure](#document-structure)
3. [Heading Conventions](#heading-conventions)
4. [Code Blocks](#code-blocks)
5. [Tables](#tables)
6. [Links](#links)
7. [Lists](#lists)
8. [Validation](#validation)
9. [Compliance Assessment](#compliance-assessment)

---

## AI-Agent Optimization

### Principles

| Principle | Implementation | Why |
|-----------|----------------|-----|
| **Tables over prose** | Use tables for comparisons | Parallel parsing, scannable |
| **Explicit rules** | ALWAYS/NEVER, not "try to" | Removes ambiguity |
| **Decision trees** | If/then logic in lists | Executable reasoning |
| **Named patterns** | Anti-patterns with names | Recognizable error states |
| **Progressive disclosure** | Quick ref → details JIT | Context window efficiency |
| **Copy-paste ready** | Complete examples | Reduces inference errors |

---

## Document Structure

### SKILL.md Template

```markdown
# Skill Name

> **Triggers:** "phrase 1", "phrase 2", "phrase 3"

## Quick Reference

| Action | Command | Notes |
|--------|---------|-------|
| ... | ... | ... |

## When to Use

| Scenario | Action |
|----------|--------|
| Condition A | Do X |
| Condition B | Do Y |

## Workflow

1. Step one
2. Step two
3. Step three
```

---

## Heading Conventions

### Hierarchy Rules

| Level | Use For | Example |
|-------|---------|---------|
| `#` | Document title (one per file) | `# Style Guide` |
| `##` | Major sections | `## Installation` |
| `###` | Subsections | `### macOS Setup` |
| `####` | Minor divisions (sparingly) | `#### Homebrew` |

**NEVER:**
- Skip heading levels (`#` → `###`)
- Use bold text as fake headings
- Start with `##` (missing `#` title)

---

## Code Blocks

### Language Hints (Required)

ALWAYS specify language for syntax highlighting:

````markdown
```python
def hello():
    print("world")
```
````

### Common Language Hints

| Language | Fence | Use For |
|----------|-------|---------|
| `bash` | ` ```bash ` | Shell commands |
| `python` | ` ```python ` | Python code |
| `go` | ` ```go ` | Go code |
| `typescript` | ` ```typescript ` | TypeScript |
| `yaml` | ` ```yaml ` | YAML config |
| `json` | ` ```json ` | JSON data |
| `text` | ` ```text ` | Plain text, diagrams |

---

## Tables

### When to Use

| Situation | Use Table? | Alternative |
|-----------|------------|-------------|
| Comparing 3+ items | Yes | - |
| Key-value mappings | Yes | - |
| Command reference | Yes | - |
| Step-by-step | No | Numbered list |
| Narrative | No | Paragraphs |

### Table Formatting

```markdown
# Good - Aligned, readable
| Column A | Column B | Column C |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |
```

---

## Links

### Internal Links

```markdown
# Good - Relative paths
[Guide](./other-doc.md)

# Good - Anchor links
[Code Blocks](#code-blocks)

# Bad - Absolute paths
[Guide](/Users/me/project/docs/guide.md)
```

---

## Lists

### Unordered Lists

Use `-` consistently:

```markdown
- Item one
- Item two
  - Nested item
```

### Ordered Lists

Use `1.` for all items:

```markdown
1. First step
1. Second step
1. Third step
```

---

## Validation

### markdownlint Configuration

```yaml
# .markdownlint.yml
default: true

MD013:
  line_length: 100
  code_blocks: false
  tables: false

MD033:
  allowed_elements:
    - kbd
    - br
    - details
    - summary

MD004:
  style: dash

MD003:
  style: atx
```

### Validation Commands

```bash
# Lint Markdown files
npx markdownlint '**/*.md' --ignore node_modules

# Check links
npx markdown-link-check README.md

# Format with Prettier
npx prettier --write '**/*.md'
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 errors, single H1, 100% code hints, 0 broken links |
| A | <5 warnings, good structure |
| A- | <15 warnings, mostly correct |
| B | <30 warnings |
| C | Significant issues |

### Validation Commands

```bash
# Lint Markdown
npx markdownlint '**/*.md' --ignore node_modules

# Check heading hierarchy
grep -r "^# " docs/*.md | wc -l
ls docs/*.md | wc -l

# Code blocks without language
grep -rP '```\s*$' docs/ | wc -l
```

---

## Additional Resources

- [CommonMark Spec](https://spec.commonmark.org/)
- [markdownlint Rules](https://github.com/DavidAnson/markdownlint)
- [GitHub Flavored Markdown](https://github.github.com/gfm/)
