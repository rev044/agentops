# Context Discovery Tiers

**Purpose**: Systematic approach to finding code/context before implementing.

**Rule**: Work top-to-bottom. Skip tiers if source unavailable.

---

## Tier Order

| Tier | Source | Tool/Command | When to Skip |
|------|--------|--------------|--------------|
| **1** | Code-Map | `Read docs/code-map/README.md` | No code-map in repo |
| **2** | Semantic Search | `mcp__smart-connections-work__lookup` | MCP not connected |
| **3** | Scoped Search | `Grep/Glob` with path limits | - |
| **4** | Source Code | `Read` files from Tier 1-3 signposts | - |
| **5** | Prior Knowledge | `ls .agents/research/` | Verify against source |
| **6** | External Docs | Context7, WebSearch | Last resort |

---

## Tier Details

### Tier 1: Code-Map (Fastest)

```bash
Read docs/code-map/README.md   # Find category
Read docs/code-map/{feature}.md  # Get signposts
```

**Why first**: Local, instant, gives exact paths and function names.

### Tier 2: Semantic Search

```bash
mcp__smart-connections-work__lookup --query="$TOPIC" --limit=10
```

**Why second**: Finds conceptual matches code-map might miss. Requires MCP.

### Tier 3: Scoped Search

```bash
Grep("pattern", path="services/auth/")   # SCOPED
Glob("services/etl/**/*.py")             # SCOPED
```

**Never**: `Grep("pattern")` or `Glob("**/*.py")` on large repos.

### Tier 4: Source Code

Read files identified by Tiers 1-3. Use function/class names, not line numbers.

### Tier 5: Prior Knowledge

```bash
ls .agents/research/ | grep -i "$TOPIC"
```

**Caution**: May be stale. Always verify findings against current source.

### Tier 6: External

- **Context7**: Library documentation
- **WebSearch**: External APIs, standards

---

## Quick Reference

```
Code-Map → Semantic → Grep/Glob → Source → .agents/ → External
   ↓           ↓          ↓          ↓         ↓          ↓
 paths      meaning    keywords    code     history    docs
```

---

## Tier Weights (Flywheel-Optimized)

Default weights based on typical value. Adjust based on `GET /memories/analytics/sources`:

| Tier | Source Type | Default Weight | Notes |
|------|-------------|----------------|-------|
| 1 | `code-map` | 1.0 | Local, authoritative |
| 2 | `smart-connections` | 0.95 | High semantic match |
| 3 | `grep`, `glob` | 0.85 | Keyword precision |
| 4 | `read` | 0.80 | Direct source |
| 5 | `prior-research`, `memory-recall` | 0.70 | May be stale |
| 6 | `web-search`, `web-fetch` | 0.60 | External, verify |

**Optimization loop**:
```bash
# Query source analytics
curl -H "X-API-Key: $KEY" "$ETL_URL/memories/analytics/sources?collection=default"

# Response includes per-source value_score metrics:
# {
#   "sources": [
#     {"source_type": "smart-connections", "value_score": 0.72},
#     {"source_type": "grep", "value_score": 0.61},
#     ...
#   ],
#   "recommendations": [...]
# }

# Adjust weights based on value_score:
# value_score = (total_citations / memory_count) × avg_confidence × recency_factor
#
# - value_score > 0.5: Move source up in priority (increase weight)
# - value_score 0.3-0.5: Maintain current position
# - value_score < 0.3: Consider deprioritizing
# - value_score < 0.1 with high count: Review quality - many memories but rarely cited
```

**Tool to source_type mapping** (for session analyzer):
```python
WebSearch → "web-search"
WebFetch → "web-fetch"
mcp__smart-connections-work__lookup → "smart-connections"
mcp__smart-connections-personal__lookup → "smart-connections"
mcp__ai-platform__search_knowledge → "athena-knowledge"
mcp__ai-platform__memory_recall → "memory-recall"
Grep → "grep"
Glob → "glob"
Read → "read"
LSP → "lsp"
```

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Start with Grep on full repo | Start with code-map |
| Read source before knowing where | Find signposts first |
| Trust .agents/ without verifying | Cross-check against source |
| Web search for internal code | Use Tiers 1-4 |
| Unscoped Glob/Grep | Always specify path |
