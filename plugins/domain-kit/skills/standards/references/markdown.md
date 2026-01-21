# Markdown Style Guide - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~4KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/markdown-standards.md (~15KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for Markdown standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Standard | Value | Validation |
|----------|-------|------------|
| **Line Length** | 100 chars soft limit | markdownlint |
| **Heading Style** | ATX (`#`) | markdownlint |
| **List Marker** | `-` for unordered | markdownlint |
| **Code Fence** | Triple backtick + language | markdownlint |
| **Link Style** | Reference links for repeated | Visual check |

---

## AI-Agent Optimization

| Principle | Implementation |
|-----------|----------------|
| **Tables over prose** | Use tables for comparisons |
| **Explicit rules** | ALWAYS/NEVER, not "try to" |
| **Decision trees** | If/then logic in lists |
| **Named patterns** | Anti-patterns with names |
| **Copy-paste ready** | Complete examples |

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Broken internal links | Wrong relative path | Use `./` prefix |
| Code not highlighted | Missing language | Add language after fence |
| Table renders as text | No blank line before | Add blank line |
| List breaks | No blank line between | Add blank lines |
| Heading not rendered | No space after `#` | `# Title` not `#Title` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| Wall of Text | Long paragraphs | Break into lists/tables |
| Implicit Logic | "You might want to..." | "ALWAYS do X when Y" |
| Deep Nesting | 4+ bullet levels | Flatten or use headings |
| Fake Headings | `**Bold Text**` as title | Use `##` headings |
| Screenshot-Only | Instructions in images | Text + optional screenshot |

---

## Headings

| Level | Use For | Example |
|-------|---------|---------|
| `#` | Document title (one per file) | `# Style Guide` |
| `##` | Major sections | `## Installation` |
| `###` | Subsections | `### macOS Setup` |
| `####` | Minor divisions | `#### Homebrew` |

**NEVER:** Skip levels, bold as heading, start with `##`

---

## Code Blocks

````markdown
```python
def hello():
    print("world")
```
````

| Language | Use For |
|----------|---------|
| `bash` | Shell commands |
| `python` | Python code |
| `go` | Go code |
| `yaml` | YAML config |
| `json` | JSON data |
| `text` | Plain text, diagrams |

---

## Tables

```markdown
| Column A | Column B | Column C |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |
```

**Use tables for:** Comparisons, key-value maps, command reference
**Don't use for:** Step-by-step instructions, narrative

---

## Links

```markdown
# Internal (relative)
[Guide](./other-doc.md)

# Anchor
[Code Blocks](#code-blocks)

# Reference (repeated URLs)
[Docs][k8s-docs]
[k8s-docs]: https://kubernetes.io/docs/
```

---

## Lists

```markdown
# Unordered (use -)
- Item one
- Item two
  - Nested

# Ordered (all 1.)
1. First step
1. Second step
1. Third step
```

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Headings** | ATX style, no skipped levels |
| **Code** | Language hint on all fences |
| **Lists** | `-` for unordered, `1.` for ordered |
| **Links** | Relative for internal docs |
| **Tables** | Blank line before, aligned |
| **Structure** | Scannable, tables over prose |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-020 | Missing code fence language | Syntax highlighting |
| PRE-021 | Skipped heading levels | Document structure |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Document templates (SKILL.md, Reference)
- markdownlint configuration
- Link conventions and validation
- Emphasis and blockquote patterns
- Compliance assessment details

Load: `vibe/references/markdown-standards.md`
