---
name: research
description: >
  Deep codebase exploration. Triggers: "research", "explore", "investigate",
  "understand", "deep dive", "current state".
version: 2.1.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Bash,Grep,Glob,Task"
skills:
  - beads
---

# Research Skill

Deep codebase exploration that produces structured findings in `~/gt/.agents/<rig>/research/`.

## Overview

Systematic exploration before planning or implementation. Research produces
evidence-based findings with file paths and actionable recommendations.

**When to Use**:
- Starting a new feature or investigation
- Understanding unfamiliar codebase areas
- Evaluating technical approaches

**When NOT to Use**:
- Simple questions (just answer directly)
- Prior research already covers topic (reference it)

---

## Workflow

```
0.  Rig Detection      -> Determine target rig from code paths
0.5 Setup              -> mkdir -p ~/gt/.agents/<rig>/research/
1.  Prior Art          -> Search existing research
2.  Research           -> Parallel sub-agent exploration
3.  Output             -> Write research document
4.  Confirm            -> Verify file, inform user
```

---

## Phase 0: Rig Detection

**CRITICAL**: All `.agents/` artifacts go to `~/gt/.agents/<rig>/` based on the primary codebase being researched.

**Detection Logic**:
1. Identify which rig's code is being explored (e.g., files in `~/gt/athena/` → `athena`)
2. If researching multiple rigs, use `_cross-rig`
3. If unknown/unclear, ask user

| Files Being Read | Target Rig | Output Base |
|------------------|------------|-------------|
| `~/gt/athena/**` | `athena` | `~/gt/.agents/athena/` |
| `~/gt/daedalus/**` | `daedalus` | `~/gt/.agents/daedalus/` |
| `~/gt/cyclopes/**` | `cyclopes` | `~/gt/.agents/cyclopes/` |
| `~/gt/hephaestus/**` | `hephaestus` | `~/gt/.agents/hephaestus/` |
| Multiple rigs | `_cross-rig` | `~/gt/.agents/_cross-rig/` |

```bash
# Set RIG variable for use in output paths
RIG="athena"  # or daedalus, cyclopes, hephaestus, _cross-rig
mkdir -p ~/gt/.agents/$RIG/research/
```

---

## Phase 1: Prior Art Discovery

**CRITICAL**: Check before creating new research.

```bash
# Semantic search (best for finding related work)
mcp__smart-connections-work__lookup --query="$TOPIC" --limit=5

# Town-level artifacts (Mayor/orchestration work)
ls -la ~/gt/.agents/$RIG/research/ | grep -i "<keywords>"
ls -la ~/gt/.agents/_cross-rig/research/ | grep -i "<keywords>"

# Crew workspace artifacts (implementation work - may have older artifacts)
ls -la ~/gt/$RIG/crew/boden/.agents/research/ 2>/dev/null | grep -i "<keywords>"
```

**Note**: Prior art may exist in either location:
- **Town-level** (`~/gt/.agents/<rig>/`) - Mayor/orchestration artifacts
- **Crew workspace** (`~/gt/<rig>/crew/boden/.agents/`) - Implementation artifacts

| Decision | When | Action |
|----------|------|--------|
| **Extension** | Prior incomplete | Build on existing |
| **Supersession** | Prior outdated | Create new with `supersedes:` |
| **Redundant** | Prior complete | Reference existing |

---

## Phase 2: Research

Launch ONE batched sub-agent for efficient exploration:

```
Task(
    subagent_type="Explore",
    model="haiku",
    prompt="Research $TOPIC comprehensively:
1. Find relevant code (file paths, key functions)
2. Find existing patterns to follow
3. Find related documentation
Group findings by category."
)
```

**Note**: Use haiku for exploration (fast, cheap). One batched query saves 40K tokens vs 3 separate agents.

### Research Checklist

- [ ] Understand current state
- [ ] Identify affected components
- [ ] Find existing patterns
- [ ] Locate related tests
- [ ] Identify constraints/dependencies
- [ ] Note potential risks

---

## Phase 3: Output

Write to `~/gt/.agents/$RIG/research/YYYY-MM-DD-{topic-slug}.md`

See `references/document-template.md` for full template.

**Required Sections**:
1. Frontmatter (date, type, tags, status)
2. Executive Summary (2-3 sentences)
3. Current State (key files, patterns)
4. Findings (with evidence file:line)
5. Constraints & Risks (tables)
6. Recommendation (approach + rationale)
7. **Discovery Provenance** (source attribution table)
8. Next Steps (→ /plan)

### Discovery Provenance Section

Track which sources provided key insights for flywheel optimization. This enables the knowledge flywheel to measure which discovery sources produce the most valuable knowledge.

**When to include:**
- Every research document outputs a Discovery Provenance table
- One row per key finding showing which source discovered it
- Enables post-hoc analysis: which sources led to successful decisions?

**Example**:
```markdown
## Discovery Provenance

| Finding | Source Type | Source Detail | Confidence |
|---------|-------------|---------------|------------|
| MCP pattern in ai-platform | smart-connections | "MCP server architecture" lookup | 0.95 |
| Authentication flows | grep | services/auth/*.py:auth_flow | 1.0 |
| Rate limiting precedent | prior-research | 2026-01-10-ratelimit-research.md | 0.85 |
| External OAuth standard | web-search | "RFC 6749 OAuth 2.0" | 0.80 |
```

**Source types**:
- Tier 1: `code-map`
- Tier 2: `smart-connections`, `athena-knowledge`
- Tier 3: `grep`, `glob`
- Tier 4: `read`, `lsp`
- Tier 5: `prior-research`, `prior-retro`, `prior-pattern`, `memory-recall`
- Tier 6: `web-search`, `web-fetch`
- Other: `conversation`, `code-map`

**Use case**: After the research is done, these provenance entries become memory candidates. The session analyzer extracts them and stores memories with source_type, enabling `GET /memories/analytics/sources` to measure which discovery tiers produce the most valuable knowledge.

---

## Phase 4: Confirm

```bash
ls -la ~/gt/.agents/$RIG/research/
```

Tell user:
```
Research output: ~/gt/.agents/$RIG/research/YYYY-MM-DD-topic.md
Next: /plan ~/gt/.agents/$RIG/research/YYYY-MM-DD-topic.md
```

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Vague findings | Cite `file:line` for evidence |
| Skip prior art check | Always run Phase 0.5 |
| Single sequential search | Parallel sub-agents |
| Missing recommendation | Always conclude with direction |

---

## References

- **Context Discovery**: `references/context-discovery.md` (6-tier hierarchy)
- **Document Template**: `references/document-template.md`
- **Tag Vocabulary**: `.claude/includes/tag-vocabulary.md`

## Workflow Integration

```
/research -> /plan -> /implement -> /retro
```
