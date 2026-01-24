---
name: research
description: >
  Deep codebase exploration. Triggers: "research", "explore", "investigate",
  "understand", "deep dive", "current state".
version: 3.0.0
tier: solo
author: "AI Platform Team"
license: "MIT"
context: inline
allowed-tools: "Read,Write,Bash,Grep,Glob,Task"
skills:
  - beads
---

# Research Skill

Deep codebase exploration → `~/gt/.agents/<rig>/research/`

## Role in the Brownian Ratchet

Research is the **chaos source** - broad exploration that gathers raw material:

| Component | Research's Role |
|-----------|-----------------|
| **Chaos** | Multiple exploration paths, parallel reads, divergent investigation |
| **Filter** | Human synthesis decision consolidates findings |
| **Ratchet** | Synthesis artifact locked in `.agents/synthesis/` |

### The Synthesis Ratchet (MANDATORY)

> **Research is chaos. Synthesis is the ratchet.**

Raw research is unusable at scale. The mandatory synthesis step filters
and consolidates, creating a single source of truth.

**Gate Criteria:**
- [ ] Single canonical reference created
- [ ] Conflicting findings resolved
- [ ] Artifact written to `~/gt/.agents/$RIG/synthesis/`

**Without synthesis, research is noise.**

```
/research (chaos) → SYNTHESIS (ratchet) → /plan or /product
```

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

**Check ALL knowledge sources before starting new research:**

```bash
# 1. Prior research (fastest, most relevant)
ls ~/gt/.agents/$RIG/research/ | grep -i "$TOPIC"

# 2. Learnings from post-mortems (lessons already learned)
ls ~/gt/.agents/$RIG/learnings/ 2>/dev/null | grep -i "$TOPIC"

# 3. Retros from similar work (what went wrong/right before)
ls ~/gt/.agents/$RIG/retros/ 2>/dev/null | grep -i "$TOPIC"

# 4. Patterns that might apply (reusable solutions)
ls ~/gt/.agents/$RIG/patterns/ 2>/dev/null

# 5. MCP memory recall (stored insights across sessions)
mcp__ai-platform__memory_recall(query="$TOPIC", limit=5)

# 6. Semantic search (conceptual matches)
mcp__smart-connections-work__lookup --query="$TOPIC"
```

**Why this matters:** These sources contain knowledge from previous implementations.
If you skip them, you risk repeating mistakes or reinventing solutions.

If prior exists: **reference it**, don't duplicate.

## Context Discovery (6-Tier)

| Tier | Source | Why |
|------|--------|-----|
| 1 | Code-map (`docs/code-map/`) | Fastest, most authoritative |
| 2 | Semantic search (MCP + memory_recall) | Finds conceptual matches + stored insights |
| 3 | Scoped grep/glob | Keyword precision |
| 4 | Source code | Direct evidence |
| 5 | Knowledge artifacts (`.agents/`) | Historical context from the knowledge loop |
| 6 | External (web) | Last resort |

**Tier 5 breakdown** (knowledge loop outputs):
- `.agents/research/` - Prior research documents
- `.agents/learnings/` - Extracted lessons from post-mortems
- `.agents/retros/` - Retrospective artifacts
- `.agents/patterns/` - Reusable solution patterns
- `.agents/specs/` - Specs with "Post-Implementation Learnings"

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
- `domain-kit/skills/standards/references/rag-formatting.md` - RAG-optimized knowledge formatting

## Phase Completion (RPI Workflow)

When research is complete:

```bash
~/.claude/scripts/checkpoint.sh research "Brief description of findings"
```

This will:
1. Save a checkpoint to `~/gt/.agents/$RIG/checkpoints/`
2. Remind you to start a fresh session
3. Provide recovery commands for the next phase

**Why fresh session?** Research context (files read, patterns explored) can bias
the planning phase. The RPI workflow enforces clean boundaries between phases.

## Synthesis Checkpoint (L1)

**From 2026-01 Post-Mortem:** Synthesis step is mandatory.

**CRITICAL:** Never plan directly from raw research. After research:

```
/research -> SYNTHESIS -> /plan or /product
```

### Why Synthesis?

Raw research is unusable at scale:
- 5 research docs with conflicting info → confusion during planning
- Synthesis creates single-source-of-truth → clear epic scope

### What Synthesis Produces

1. **Consolidated decisions** - One answer per question, not "option A vs B"
2. **Single canonical reference** - ~10-20K chars, not 50K+ scattered
3. **Clear constraints** - What we decided NOT to do and why

### RSI Pattern (P4)

Research-Synthesis-Implement is the proven flow:

```
Research (broad) -> Synthesis (consolidate) -> Plan (tasks) -> Implement (execute)
                          ↑
                          |
                    MANDATORY STEP
```

**Artifact:** After research, create:
`~/gt/.agents/$RIG/synthesis/YYYY-MM-DD-{topic}.md`

**See:** `~/.claude/CLAUDE.md` Post-Mortem Learnings (2026-01)

---

## Next

```
/research -> SYNTHESIS -> /plan or /product
```

After checkpoint, start fresh session. Create synthesis doc, then run `/plan` or `/product`.
