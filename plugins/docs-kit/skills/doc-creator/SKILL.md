---
name: doc-creator
description: >
  This skill should be used when the user asks to "create documentation",
  "write a corpus document", "create a standard", "add corpus content",
  "write GraphRAG documentation", or mentions corpus/standards documentation.
version: 2.0.0
tier: solo
context: inline
allowed-tools: "Read,Write,Edit,Glob,Grep"
skills:
  - standards
---

# Document Creator Skill

Create well-structured documentation with automatic context detection. Supports both corpus (educational) and standards (normative) documentation types.

---

## Context Detection

This skill automatically detects the type of documentation needed:

| If Creating... | Mode | Characteristics |
|----------------|------|-----------------|
| Corpus, training data, educational content | **Corpus Mode** | GraphRAG frontmatter, concept indexes, learning paths |
| Standards, guidelines, schemas, vocabularies | **Standards Mode** | RFC 2119 language, validation gates, normative rules |

**Detection triggers:**
- **Corpus Mode:** mentions "corpus", "training", "learning", "concepts", "GraphRAG", "educational"
- **Standards Mode:** mentions "standard", "guideline", "schema", "vocabulary", "normative", "MUST/SHOULD/MAY"

---

## Critical Rules

### Corpus Mode

1. **ALWAYS validate frontmatter** with `make validate-corpus` before committing
2. **NEVER skip the `concepts` field** - it powers GraphRAG entity extraction
3. **ALWAYS estimate tokens** using the ~4 words/token formula
4. **CHECK existing concept indexes** before adding new concepts

### Standards Mode

1. **Standards are NORMATIVE** - Use MUST/SHOULD/MAY language (RFC 2119)
2. **ALWAYS include validation** - Every standard needs a checkable gate
3. **UPDATE the README index** - Add new standards to `docs/standards/README.md`
4. **CONSIDER impact** - Standards affect all developers; get consensus first

---

## When to Use This Skill

**Corpus Mode - Use when:**
- Creating a new corpus (e.g., `go-corpus`, `terraform-corpus`)
- Adding new sections to existing corpora
- Creating individual content documents within sections
- Building or updating concept indexes for GraphRAG

**Standards Mode - Use when:**
- Creating a new coding standard (language, tool, or pattern)
- Creating schema documentation (like frontmatter schemas)
- Adding vocabulary definitions (like tag vocabularies)
- Documenting validation requirements

**Do not use this skill for:**
- API reference docs (different structure)
- General README files
- Process/methodology documentation

---

## Corpus Mode Quick Reference

### Document Types

| Type | Frontmatter `type` | Purpose |
|------|-------------------|---------|
| Corpus Root | `corpus-root` | Top-level README anchoring the corpus |
| Section | `corpus-section` | Groups related content documents |
| Content | `corpus-content` | Individual learning units |
| Concept Index | `corpus-index` | GraphRAG entity extraction |

### Required Frontmatter Fields

**All corpus documents require:**
- `title` - Human-readable title
- `type` - One of: `corpus-root`, `corpus-section`, `corpus-content`, `corpus-index`
- `corpus` - Short identifier (helm, shell, python)
- `tags` - Searchable tags, minimum `[corpus, NAME]`

**Additional by type:**
- `corpus-root`: `tokens`, `sections`
- `corpus-section`: `section`, `tokens`, `time`, `concepts`
- `corpus-content`: `section`, `tokens`, `time`, `concepts`
- `corpus-index`: (no additional required fields)

### Validation

```bash
make validate-corpus          # Quick validation
make validate-corpus-verbose  # Detailed output
```

---

## Standards Mode Quick Reference

### Document Types

| Type | Purpose |
|------|---------|
| Coding Standard | Rules for a specific language or technology |
| Schema | Data structures with required fields |
| Tag Vocabulary | Controlled vocabularies for categorization |

### Standard Document Structure

Every standard MUST include:
1. **Purpose statement** - One sentence describing coverage
2. **Scope section** - What topics are covered
3. **Quick Reference table** - With validation column
4. **Required Patterns** - Using MUST/SHOULD/MAY language
5. **Validation section** - Command to verify compliance

### Normative Language (RFC 2119)

| Term | Meaning |
|------|---------|
| **MUST** | Absolute requirement |
| **MUST NOT** | Absolute prohibition |
| **SHOULD** | Strong recommendation |
| **SHOULD NOT** | Strong discouragement |
| **MAY** | Optional |

### Validation

```bash
make check-links  # Validate internal links
```

After creating a standard, update `docs/standards/README.md` index.

---

## Corpus vs Standards

| Aspect | Corpus | Standards |
|--------|--------|-----------|
| **Language** | Patterns, examples (educational) | MUST/SHOULD (normative) |
| **Purpose** | Teach best practices | Define minimum bar |
| **Validation** | Optional | Required, automated |
| **Audience** | Learners, developers | Enforcers, tools |
| **Location** | `docs/corpus/` | `docs/standards/` |
| **Frontmatter** | GraphRAG-enabled | Simple metadata |

---

## Additional Resources

### Reference Files

For detailed templates and procedures, consult:

- **`references/corpus-templates.md`** - Full corpus document templates including:
  - Corpus root template with frontmatter schema
  - Section template with prerequisites
  - Content document template with examples
  - Concept index template for GraphRAG
  - Quick start commands and validation checklist

- **`references/standards-templates.md`** - Full standards document templates including:
  - Coding standard template
  - Schema documentation template
  - Tag vocabulary template
  - Quick start commands and validation checklist

### External References (JIT Load)

**Location:** `gitops/docs/` repository

| Reference | Path | When to Load |
|-----------|------|--------------|
| Frontmatter Schema | `docs/standards/corpus-frontmatter-schema.md` | Field requirements |
| Tag Vocabulary | `docs/standards/tag-vocabulary.md` | Valid tags |
| Standards README | `docs/standards/README.md` | Index structure |
| Helm Corpus | `docs/corpus/helm-corpus/README.md` | Example corpus root |
| Shell Standards | `docs/standards/shell-script-standards.md` | Example coding standard |

---

## Integration Points

**Note:** These paths are relative to the gitops repository.

- **Corpus Validation:** `make validate-corpus` / `make validate-corpus-verbose`
- **Standards Index:** `docs/standards/README.md` - Update for every new standard
- **Link Validation:** `make check-links` validates internal links
- **Schema:** `docs/standards/corpus-frontmatter-schema.md`
- **Tags:** `docs/standards/tag-vocabulary.md`

---

## Standards Library

When creating documentation, reference the standards library for consistent formatting:

| Content Type | Reference |
|--------------|-----------|
| Markdown files | `~/.claude/skills/standards/references/markdown.md` |
| Document tags | `~/.claude/skills/standards/references/tags.md` |
| JSON/JSONL data | `~/.claude/skills/standards/references/json.md` |
| YAML frontmatter | `~/.claude/skills/standards/references/yaml.md` |

---

**Created:** 2025-12-31
**Version:** 2.0.0 (refactored into SKILL.md + references/)
**Maintainer:** Platform team
