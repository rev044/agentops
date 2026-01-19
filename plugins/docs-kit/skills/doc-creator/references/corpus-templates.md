# Corpus Documentation Templates

Detailed templates and frontmatter schemas for corpus documentation.

---

## Corpus Root Template

The corpus root (`README.md`) anchors the entire corpus hierarchy.

### Required Frontmatter

```yaml
---
title: "Corpus Name"           # Human-readable title
type: corpus-root              # MUST be corpus-root
corpus: name                   # Short identifier (helm, shell, python)
tokens: ~25000                 # Total estimated tokens for entire corpus
sections: [00-nav, 01-basic]   # List of section directories
tags: [corpus, name, domain]   # Searchable tags
---
```

### Full Template

```markdown
# [Corpus Name]: [Subtitle]

**[One-line purpose statement]**

---

## Philosophy

[2-3 sentences on who this corpus is for and mental model]

---

## Corpus Structure

| Section | Purpose | Time Investment |
|---------|---------|-----------------|
| [00-foundation](00-foundation/) | Basics | 1.5 hours |
| [01-patterns](01-patterns/) | Core patterns | 2 hours |

**Total accelerated path:** X hours for fluency

---

## Learning Paths

### Path 1: "Quick read" (X hours)
1. [Link to doc] - X min
2. [Link to doc] - X min

### Path 2: "Deep dive" (X hours)
...

---

## Token Budget

| Document | Tokens | Use When |
|----------|--------|----------|
| This README | ~2k | Overview |
| 00-foundation | ~4k | New to topic |
...

---

## Related Resources

- [Related standard](../../standards/...)
- [Related corpus](../other-corpus/)
```

---

## Corpus Section Template

Sections group related content documents. Each section has a README.

### Required Frontmatter

```yaml
---
title: "Section: Topic Name"   # Section title with topic
type: corpus-section           # MUST be corpus-section
corpus: name                   # Parent corpus identifier
section: 01-topic              # Section directory name
tokens: ~6000                  # Estimated tokens for section
time: "2 hours"                # Learning time estimate
prerequisites: [00-foundation] # Required prior sections (optional)
concepts: [concept-a, concept-b, concept-c]  # 5-10 key concepts
tags: [corpus, name, topic]    # Searchable tags
---
```

### Full Template

```markdown
# Section: [Topic Name]

[2-3 sentence overview of what this section covers]

---

## Prerequisites

- Complete [00-foundation](../00-foundation/) first
- Familiarity with [concept]

---

## Documents

| Document | Purpose | Time |
|----------|---------|------|
| [basics.md](basics.md) | Core concepts | 30 min |
| [patterns.md](patterns.md) | Common patterns | 45 min |

---

## Key Concepts

| Concept | Definition |
|---------|------------|
| `concept-a` | Brief definition |
| `concept-b` | Brief definition |

---

## Next Steps

After completing this section:
- Move to [02-advanced](../02-advanced/)
- Practice with [exercises](../exercises/)
```

---

## Content Document Template

Individual content documents are the learning units within sections.

### Required Frontmatter

```yaml
---
title: "Document Title"        # Document title
type: corpus-content           # MUST be corpus-content
corpus: name                   # Parent corpus identifier
section: 01-topic              # Parent section directory
tokens: ~2000                  # Estimated document tokens
time: "30 min"                 # Reading time estimate
concepts: [concept-a, concept-b]  # 3-8 specific concepts
relates_to: [other-doc.md]     # Related docs in same section (optional)
tags: [corpus, name, topic, subtopic]  # Searchable tags
---
```

### Full Template

```markdown
# [Document Title]

[1-2 paragraph introduction explaining what reader will learn]

---

## [Core Concept 1]

[Explanation with examples]

### Pattern

```language
# Example code or pattern
```

### When to Use

- Use case 1
- Use case 2

---

## [Core Concept 2]

...

---

## Common Mistakes

| Mistake | Why It's Wrong | Correct Approach |
|---------|----------------|------------------|
| [mistake] | [reason] | [fix] |

---

## Quick Reference

| Operation | Syntax |
|-----------|--------|
| [op] | `syntax` |

---

## See Also

- [Related doc](./related.md)
- [External reference](https://...)
```

---

## Concept Index Template

The concept index enables GraphRAG entity extraction and cross-referencing.

### Required Frontmatter

```yaml
---
title: "NAME Concept Index"
type: corpus-index
corpus: name
tags: [corpus, name, index, graphrag]
---
```

### Full Template

```markdown
# [NAME] Concept Index

Alphabetical index of concepts for GraphRAG entity extraction.

---

## A

### concept-alpha
**Definition:** Brief definition of concept.
**See also:** related-concept, another-concept
**Found in:** [Section](../01-section/doc.md)

---

## B

### concept-beta
...
```

---

## Concept Naming Guidelines

| Good Concept | Bad Concept |
|--------------|-------------|
| `sync-wave` | `section 2` |
| `parameter-expansion` | `example` |
| `errexit` | `important` |

**Naming rules:**
- Lowercase with hyphens: `sync-wave` not `SyncWave`
- Be specific: `helm-hook` not `hook`
- Match existing terminology

---

## Quick Start: Create New Corpus

```bash
# 1. Create directory structure
mkdir -p docs/corpus/NEW-corpus/{00-navigation,01-foundation}

# 2. Create root README with frontmatter
cat > docs/corpus/NEW-corpus/README.md << 'EOF'
---
title: "NEW Corpus"
type: corpus-root
corpus: NEW
tokens: ~5000
sections: [00-navigation, 01-foundation]
tags: [corpus, NEW]
---

# NEW Corpus: [Subtitle]

**Pattern-recognition-based training...**
EOF

# 3. Create concept index
cat > docs/corpus/NEW-corpus/00-navigation/index.md << 'EOF'
---
title: "NEW Concept Index"
type: corpus-index
corpus: NEW
tags: [corpus, NEW, index, graphrag]
---

# NEW Concept Index

Alphabetical index of concepts for GraphRAG entity extraction.
EOF

# 4. Validate
make validate-corpus-verbose
```

---

## Add Corpus to Enforcement

After creating a new corpus, add it to validation enforcement.

### Step 1: Edit Validation Script

```bash
# In tools/scripts/validate-corpus-frontmatter.sh
# Find ENFORCED_CORPORA array and add new corpus:

ENFORCED_CORPORA=("helm-corpus" "shell-corpus" "NEW-corpus")
```

### Step 2: Validate

```bash
make validate-corpus-verbose
```

---

## Validation Checklist

Before committing corpus content:

- [ ] All files have `---` frontmatter delimiters
- [ ] `type` field matches document role (root/section/content/index)
- [ ] `corpus` field uses consistent short identifier
- [ ] `tokens` estimated with ~4 words/token formula
- [ ] `concepts` field populated (not empty array)
- [ ] `tags` include at least `[corpus, NAME]`
- [ ] `make validate-corpus` passes 100%

---

## Error Recovery

### "Missing frontmatter" Error

```bash
# Check file starts with ---
head -1 docs/corpus/NAME-corpus/file.md
# Should output: ---

# If missing, add frontmatter block at file start
```

### "Invalid type" Error

Valid types: `corpus-root`, `corpus-section`, `corpus-content`, `corpus-index`

```bash
# Check current type
grep "^type:" docs/corpus/NAME-corpus/file.md
```
