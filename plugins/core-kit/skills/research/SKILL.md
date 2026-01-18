---
name: research
description: >
  Deep codebase exploration. Triggers: "research", "explore", "investigate",
  "understand", "deep dive", "current state".
version: 3.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Bash,Grep,Glob,Task"
skills:
  - beads
---

# Research Skill

Deep codebase exploration → `~/gt/.agents/<rig>/research/`

## Quick Start

```bash
/research authentication flows in services/auth
```

## Workflow

```
1. Rig Detection   -> Where does output go?
2. Prior Art       -> What already exists? (CRITICAL)
3. Context Discovery -> 6-tier systematic exploration
4. Synthesis       -> Analyze, identify patterns
5. Output          -> Write research doc
```

## Rig Detection

Output goes to `~/gt/.agents/<rig>/research/` based on code being explored:

| Code Location | Rig | Output |
|---------------|-----|--------|
| `~/gt/athena/**` | athena | `~/gt/.agents/athena/research/` |
| `~/gt/daedalus/**` | daedalus | `~/gt/.agents/daedalus/research/` |
| Multiple rigs | _cross-rig | `~/gt/.agents/_cross-rig/research/` |

## Prior Art (Never Skip)

```bash
mcp__smart-connections-work__lookup --query="$TOPIC"
ls ~/gt/.agents/$RIG/research/ | grep -i "$TOPIC"
```

If prior exists: reference it, don't duplicate.

## Context Discovery (6-Tier)

| Tier | Source | Why |
|------|--------|-----|
| 1 | Code-map (`docs/code-map/`) | Fastest, most authoritative |
| 2 | Semantic search (MCP) | Finds conceptual matches |
| 3 | Scoped grep/glob | Keyword precision |
| 4 | Source code | Direct evidence |
| 5 | Prior research (`.agents/`) | Historical context |
| 6 | External (web) | Last resort |

**Details:** `references/context-discovery.md`

## Output

Write to `~/gt/.agents/$RIG/research/YYYY-MM-DD-{topic}.md`

**Required sections:**
- Executive Summary
- Current State (key files)
- Findings (with `file:line` evidence)
- Constraints & Risks
- Recommendation
- Next Steps

**Template:** `references/document-template.md`

## Key Rules

| Rule | Why |
|------|-----|
| Stay under 40% context | Prevents hallucination |
| Always cite `file:line` | Verifiable claims |
| Check prior art first | Prevents re-solving |
| Scope all searches | Context efficiency |
| Verify before trusting | Reality over model |

## References

- `references/context-discovery.md` - 6-tier hierarchy details
- `references/document-template.md` - Output format
- `references/failure-patterns.md` - 12 patterns to watch for
- `references/vibe-methodology.md` - Core principles

## Next

```
/research -> /plan or /product
```
