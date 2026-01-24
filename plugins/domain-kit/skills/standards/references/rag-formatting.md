# RAG-Optimized Knowledge Formatting

> **Purpose:** Formatting standards for knowledge artifacts to maximize retrieval quality with embedding-based search systems.

## Scope

This document covers: markdown formatting rules, frontmatter conventions, and section structure for knowledge artifacts that will be indexed by embedding models for retrieval-augmented generation (RAG).

**Related:**
- [Tag Vocabulary](./tags.md) - Tag selection and categories
- [Markdown Style Guide](./markdown.md) - General markdown conventions

---

## Quick Reference

| Rule | Target | Rationale |
|------|--------|-----------|
| **One concept per H2** | Each H2 = separate embedding | Focused semantics, precise retrieval |
| **200-400 chars/section** | Embedding model sweet spot | Too short merges, too long dilutes |
| **Front-load key terms** | First sentence most important | Embeddings weight early text heavily |
| **No filler words** | Every word adds semantic value | "It is important to note" = noise |
| **Action-oriented headings** | "Validate JWT tokens" not "About tokens" | Matches search intent |

---

## Why These Rules Matter

Modern embedding models convert text to fixed-size vectors. When sections are:
- **Too short** (<100 chars): May be merged with adjacent content, losing distinct meaning
- **Too long** (>500 chars): Semantic meaning gets averaged/diluted
- **Unfocused**: Multiple topics in one section confuse retrieval

The 200-400 character range per section is optimal for most embedding models (384-1024 dimensions) and provides the best retrieval precision.

---

## Frontmatter Conventions

```yaml
---
type: fact|pattern|episode|preference  # Required: knowledge type
tier: 0-4                 # Required: maturity level
tags: [tag1, tag2]        # Required: 3-5 tags, type first
source: path/to/origin    # Optional: where learned
created: YYYY-MM-DD       # Optional: creation date
---
```

### Tier Definitions

| Tier | Name | Description |
|------|------|-------------|
| 0 | Observation | Raw, unvalidated insight |
| 1 | Documented | Written down, not verified |
| 2 | Validated | Tested in at least one context |
| 3 | Pattern | Proven reusable across contexts |
| 4 | Skill | Internalized, automatic application |

### What NOT to Store

| Field | Why Wrong |
|-------|-----------|
| `confidence` | Query-dependent - similarity computed at search time |
| `relevance` | Query-dependent - computed per search |
| `quality` | Subjective, conflates with tier |

**Key insight:** The same document has different similarity scores for different queries. Storing "confidence" on the document is meaningless - relevance is always query-dependent.

---

## Section Structure

### Good Section (285 chars)

```markdown
## Problem

API endpoints accepting JWT tokens without signature verification allow
attackers to forge authentication. This vulnerability affects all protected
routes using the shared middleware.
```

### Bad Section (too short, 45 chars)

```markdown
## Problem

JWT tokens not verified.
```

**Why it's bad:** Below minimum threshold, may merge with adjacent content. Loses semantic precision.

### Bad Section (too long, 800+ chars)

```markdown
## Problem

When users submit API requests to our gateway, the system needs to verify
that the JWT token included in the Authorization header is valid. However,
the current implementation does not properly check the signature of the
token. This means that an attacker could potentially create their own JWT
token with any claims they want, and the system would accept it as valid.
This is a serious security vulnerability because... [continues]
```

**Why it's bad:** Dilutes semantic focus, embedding averages over too much content, reduces retrieval precision.

---

## Common Errors

| Error | Example | Fix |
|-------|---------|-----|
| Filler phrases | "It should be noted that X" | Just say "X" |
| Passive voice | "Tokens are validated by the middleware" | "Middleware validates tokens" |
| Vague headings | "## Overview" | "## JWT Validation Process" |
| Missing key terms | Section about auth without "authentication" | Front-load domain terms |
| Too many topics | H2 covers 3 concepts | Split into 3 H2s |
| Empty sections | "## Implementation\nTBD" | Remove or fill |

---

## Anti-Patterns

| Name | Pattern | Why Bad | Instead |
|------|---------|---------|---------|
| **Mega-Section** | 1000+ char H2 section | Diluted semantics, poor retrieval | Split by concept |
| **Stub Section** | <100 char section | Merged with parent, loses identity | Expand or merge explicitly |
| **Keyword Stuffing** | Repeating terms for "SEO" | Embeddings detect this, reduces quality | Natural language |
| **Nested Headers** | H2 > H3 > H4 > H5 | Deep nesting fragments meaning | Flatten to H2/H3 max |
| **Confidence Field** | `confidence: 0.85` in frontmatter | Meaningless - similarity is query-time | Use `tier` for maturity |
| **Copy-Paste Docs** | Importing external docs verbatim | Context mismatch, bloat | Summarize key points |

---

## AI Agent Guidelines

When creating knowledge artifacts:

| Guideline | Rationale |
|-----------|-----------|
| ALWAYS use 200-400 chars per H2 section | Optimal for embedding models |
| ALWAYS front-load key terms in first sentence | Early text weighted heavily |
| ALWAYS include `type` and `tier` in frontmatter | Enables filtering and ranking |
| ALWAYS use action-oriented H2 headings | Matches user search intent |
| NEVER include confidence/relevance fields | These are query-time, not storage-time |
| NEVER create sections under 100 chars | Will merge unexpectedly |
| NEVER exceed 500 chars per section | Semantic dilution |
| PREFER H2 over H3 for main concepts | H2s get separate embeddings |
| PREFER concrete examples over abstractions | Embeddings anchor on specifics |

---

## Document Types

| Type | Purpose | Example |
|------|---------|---------|
| `research` | Exploration findings | "2026-01-15-auth-analysis.md" |
| `learning` | Extracted insights | "2026-01-15-jwt-gotcha.md" |
| `pattern` | Reusable solutions | "retry-with-backoff.md" |
| `retro` | Session retrospectives | "2026-01-15-epic-retro.md" |
| `plan` | Implementation plans | "2026-01-15-auth-plan.md" |

---

## Template

```markdown
---
type: pattern
tier: 2
tags: [pattern, auth, jwt, security]
created: 2026-01-15
---

# JWT Token Validation Pattern

## Problem

API endpoints accepting JWT tokens without proper signature verification
allow attackers to forge authentication. This vulnerability affects all
protected routes using the shared auth middleware.

## Solution

Always verify RS256 signatures using the public key from the JWKS endpoint.
Validate exp, iat, and iss claims before accepting the token.

## Evidence

Caught 3 vulnerabilities in auth service during 2026-01-15 security review.
Now standardized across all microservices.

## Code Reference

See `services/gateway/middleware/auth.py:45-89` for implementation.
```

---

## Validation Checklist

Quick check for RAG-readiness:

- [ ] Each H2 section is 200-400 characters
- [ ] No filler phrases ("It is important to note...")
- [ ] Headings are action-oriented, not generic
- [ ] Key terms appear in first sentence of each section
- [ ] No confidence/relevance fields in frontmatter
- [ ] Tier and type fields present in frontmatter
- [ ] No sections under 100 characters
- [ ] Maximum nesting depth is H3

---

## Summary

**The Golden Rule:** Write each H2 section as if it's the only thing a user will see when they search. Make it self-contained, semantically focused, and actionable.

**The Science:** Embedding models compress text to vectors. Focused sections = precise vectors = better retrieval. Unfocused sections = blurred vectors = poor retrieval.
