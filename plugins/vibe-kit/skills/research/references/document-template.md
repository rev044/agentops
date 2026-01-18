# Research Document Template

## Filename Format

`.agents/research/YYYY-MM-DD-{topic-slug}.md`

Convert topic to kebab-case slug:
- "authentication flow" -> `2026-01-03-authentication-flow.md`
- "MCP server architecture" -> `2026-01-03-mcp-server-architecture.md`

---

## Required Sections

### 1. Frontmatter

```yaml
---
date: YYYY-MM-DD
type: Research
topic: "Topic Name"
tags: [research, domain, tech]
status: COMPLETE
supersedes: []
---
```

### 2. Executive Summary

2-3 sentences: what found, what recommend.

### 3. Current State

- What exists today
- Key files table: | File | Purpose |
- Existing patterns

### 4. Findings

Each finding with:
- Evidence: `file:line`
- Implications

### 5. Constraints

| Constraint | Impact | Mitigation |
|------------|--------|------------|

### 6. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|

### 7. Recommendation

- Recommended approach
- Rationale
- Alternatives considered and rejected

### 8. Discovery Provenance

Track which sources provided key insights (enables flywheel optimization).

**Purpose**: Create an audit trail showing which discovery method found each insight. This enables post-hoc analysis: "Which sources led to successful implementation?"

**When to complete**: As you research, add one row per significant finding showing its source.

**Example**:
```markdown
| Finding | Source Type | Source Detail | Confidence |
|---------|-------------|---------------|------------|
| Gateway request flow | code-map | docs/code-map/gateway.md | 1.0 |
| Middleware pattern | smart-connections | "request middleware chain" | 0.95 |
| Error handling at L45 | grep | services/gateway/middleware.py | 1.0 |
| Rate limiting precedent | prior-research | 2026-01-10-ratelimit.md | 0.85 |
| OAuth2 RFC | web-search | "RFC 6749 OAuth 2.0" | 0.80 |
```

**Source Types by Tier** (higher tier = better quality):

**Tier 1 (Authoritative)**
- `code-map` - Structured architecture documentation (highest confidence)

**Tier 2 (Semantic)**
- `smart-connections` - Obsidian semantic search
- `athena-knowledge` - MCP ai-platform search

**Tier 3 (Scoped Search)**
- `grep` - Pattern matching in code
- `glob` - File pattern matching

**Tier 4 (Source Code)**
- `read` - Direct file reading
- `lsp` - Language Server Protocol queries

**Tier 5 (Prior Art)**
- `prior-research` - Previous research documents
- `prior-retro` - Retrospective learnings
- `prior-pattern` - Reusable patterns
- `memory-recall` - Semantic memory search

**Tier 6 (External)**
- `web-search` - Web search results
- `web-fetch` - Direct URL fetch

**Other**
- `conversation` - User-provided context

**Confidence scoring**:
- `1.0` - Source is authoritative/written down
- `0.95` - Semantic match, high relevance
- `0.85` - Good match, may need verification
- `0.70` - Reasonable match, verify
- < 0.70 - Use sparingly, needs verification

### 9. Next Steps

Point to `/plan` for implementation.

---

## Tag Vocabulary

**Rules:** 3-5 tags total. First tag MUST be `research`.

| Category | Valid Tags |
|----------|------------|
| **Core Domains** | `agents`, `data`, `api`, `infra`, `security`, `auth` |
| **Quality** | `testing`, `reliability`, `performance`, `monitoring` |
| **Process** | `ci-cd`, `workflow`, `ops`, `docs` |
| **Governance** | `architecture`, `compliance`, `standards`, `ui` |
| **Languages** | `python`, `shell`, `typescript`, `go`, `yaml` |
| **Platforms** | `helm`, `kubernetes`, `openshift`, `docker`, `argocd` |
| **AI Stack** | `mcp`, `litellm`, `neo4j`, `postgres`, `redis`, `fastapi` |

**Examples:**
- `[research, agents, mcp]` - MCP server research
- `[research, data, neo4j]` - Data storage research
- `[research, security, auth]` - Authentication research

---

## Status Values

| Status | Meaning |
|--------|---------|
| `COMPLETE` | Ready for planning |
| `IN_PROGRESS` | Ongoing research |
| `SUPERSEDED` | Newer research exists |
