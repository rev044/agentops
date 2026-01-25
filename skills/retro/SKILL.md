---
name: retro
description: >
  Extract learnings from completed work. Trigger phrases: "run a retrospective",
  "extract learnings", "what did we learn", "capture lessons", "create a retro".
version: 2.1.0
tier: solo
author: "AI Platform Team"
license: "MIT"
context: inline
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob"
skills:
  - beads
---

# Retro Skill

Extract learnings, patterns, and insights from completed work.

## Overview

Transform session knowledge into durable, searchable artifacts:
- `.agents/retros/` - Session retrospective summaries
- `.agents/learnings/` - Reusable knowledge extracted
- `.agents/patterns/` - Reusable patterns discovered

**When to Use**:
- After completing significant work (epic closed)
- After encountering unexpected challenges
- After discovering useful patterns

---

## Setup

```bash
mkdir -p .agents/{retros,learnings,patterns}/
```

---

## Workflow

```
1. Gather Context      -> Analyze commits, beads, blackboard
1.5. Query Analytics   -> GET /memories/analytics/sources (NEW)
2. Identify Friction   -> Detect problems and patterns
3. Propose Changes     -> Generate concrete improvements
4. Supersession        -> Check for outdated artifacts
5. User Review         -> Present proposals for approval
6. Apply Changes       -> Execute approved edits
7-9. Write Outputs     -> Retro, learnings, patterns (include Source Performance)
10. Session Naming     -> /rename for future /resume (optional)
11. Finalize           -> Update blackboard, confirm
```

---

## Phase 1-2: Gather Context & Friction

See `references/context-gathering.md` for detailed commands.

**Quick version:**
```bash
bd show <epic-id>                          # Get issue details
git log --oneline --since="7 days ago"     # Recent commits
ls .agents/blackboard/                     # Shared state
```

---

## Phase 1.5: Query Analytics

**NEW**: Before generating proposals, query the analytics endpoint to get source performance data. This enables data-driven tier weight suggestions.

### Call Analytics Endpoint

```bash
# Query source analytics for the default tenant
curl -s "${AI_PLATFORM_URL:-http://localhost:8000}/memories/analytics/sources?collection=default" \
    -H "X-API-Key: ${AI_PLATFORM_API_KEY:-test}" | jq
```

**Or via MCP tool:**
```
mcp__ai-platform__search_knowledge with query: "source performance analytics"
```

### Response Schema

```json
{
  "sources": [
    {
      "source_type": "smart-connections",
      "tier": 2,
      "memory_count": 45,
      "total_access_count": 312,
      "avg_confidence": 0.87,
      "value_score": 0.85,
      "expected_weight": 0.8,
      "deviation": 0.05
    }
  ],
  "total_memories": 150,
  "recommendations": [
    "PROMOTE: 'grep' overperforming by 25%. Consider increasing tier weight from 0.6 to 0.75."
  ]
}
```

### Include in Retro Output

Add a **Source Performance** section to the retro summary when analytics data is available:

```markdown
## Source Performance

| Source | Tier | Value Score | Expected | Deviation |
|--------|------|-------------|----------|-----------|
| smart-connections | 2 | 0.85 | 0.80 | +0.05 |
| grep | 3 | 0.75 | 0.60 | +0.15 |
| web-search | 6 | 0.15 | 0.20 | -0.05 |

### Recommendations

- **PROMOTE**: 'grep' overperforming (+25%). Consider tier weight increase.
```

### Tier Weight Deviation Suggestions

Generate suggestions when deviation > 20%:

| Deviation | Action | Example |
|-----------|--------|---------|
| > +0.2 | PROMOTE | Source outperforming expectations, increase priority |
| < -0.2 | DEMOTE | Source underperforming, decrease priority |
| -0.2 to +0.2 | OK | Source performing as expected |

**Purpose**: Creates the B2 feedback loop - analytics inform tier weights, which affect future discovery prioritization.

### Session ID Detection

Detect session ID for conversation analysis:

```bash
# Source 1: Environment variable (preferred)
SESSION_ID="${CLAUDE_SESSION_ID:-}"

# Source 2: Most recent session file (fallback)
if [ -z "$SESSION_ID" ]; then
    SESSION_ID=$(find ~/.claude/projects -name "*.jsonl" -type f -mmin -60 \
        | grep -v "agent-" | head -1 | xargs basename 2>/dev/null | sed 's/.jsonl$//')
fi

# Source 3: Crank state (for polecat sessions)
if [ -z "$SESSION_ID" ]; then
    SESSION_ID=$(cat .agents/blackboard/crank-state.json 2>/dev/null | jq -r '.session_id // empty')
fi
```

If session ID found, run conversation analysis:
```bash
python3 ~/.claude/scripts/analyze-sessions.py --session=$SESSION_ID --limit=50
```

---

## Phase 3: Propose Changes

### Proposal Format

```markdown
### Proposal: [Title]
**Severity:** CRITICAL | RECOMMENDED | OPTIONAL
**Target File:** [path]
**Evidence:** [commit:abc123] | [beads:id]

#### Problem
[What went wrong]

#### Proposed Change
```diff
- old line
+ new line
```
```

### Severity Classification

| Severity | Criteria |
|----------|----------|
| **CRITICAL** | >2 occurrences, blocking, data loss risk |
| **RECOMMENDED** | 1-2 occurrences, friction reducer |
| **OPTIONAL** | Nice-to-have, polish |

---

## Phase 5.5: Auto-Update Tiers

| Tier | Behavior | Targets |
|------|----------|---------|
| 1 | Auto-Apply | Docs, examples in `.agents/patterns/` |
| 2 | Review | Skills, triggers |
| 3 | Human Required | CLAUDE.md, critical proposals |

---

## Phase 7-9: Write Outputs

See `references/output-templates.md` for full templates.

| Output | Location | First Tag |
|--------|----------|-----------|
| Retro | `.agents/retros/YYYY-MM-DD-{topic}.md` | `retro` |
| Learning | `.agents/learnings/YYYY-MM-DD-{topic}.md` | `learning` |
| Pattern | `.agents/patterns/{name}.md` | `pattern` |

### Discovery Provenance in Learnings

Include a `## Discovery Provenance` section in learning outputs to track which sources led to insights. This creates the feedback loop for the knowledge flywheel.

**Purpose**: Enable post-hoc analysis of which discovery sources (web-search, smart-connections, grep, etc.) produce the most valuable learnings.

**Format**:
```markdown
## Discovery Provenance

| Learning | Source Type | Source Detail |
|----------|-------------|---------------|
| MCP pattern in gateways | smart-connections | "gateway middleware patterns" |
| Error handling precedent | grep | services/gateway/middleware.py:L45-60 |
| Rate limit implementation | prior-research | 2025-12-15-ratelimit-pattern.md |
```

> **Note:** Do NOT include a "Confidence" column. Confidence/relevance are query-time metrics, not storage-time. See `domain-kit/skills/standards/references/rag-formatting.md`.

**Source types by tier**:
- **Tier 1**: `code-map` - Structured documentation (highest quality)
- **Tier 2**: `smart-connections`, `athena-knowledge` - Semantic search
- **Tier 3**: `grep`, `glob` - Pattern search
- **Tier 4**: `read`, `lsp` - Direct source reading
- **Tier 5**: `prior-research`, `prior-retro`, `prior-pattern`, `memory-recall` - Past work
- **Tier 6**: `web-search`, `web-fetch` - External (lowest priority, highest risk)

**How it works**:
1. During retro, learnings and their sources are identified
2. A provenance table shows source_type for each learning
3. Session analyzer extracts these and stores as memories with source_type
4. Analytics endpoint (`GET /memories/analytics/sources`) measures value_score per source
5. High-value sources get prioritized in future discovery hierarchies

**Example analysis**: If learnings extracted from `smart-connections` have 10x higher citation counts than `web-search` learnings, smart-connections should be ranked higher in the discovery tier ordering.

---

## Phase 10: Session Naming (Optional)

Name the session for future reference:

```bash
/rename "retro-{topic}-$(date +%Y-%m-%d)"
```

**Naming Convention:**
- `retro-oauth-2026-01-19` - For feature retros
- `retro-debug-mem-leak-2026-01-19` - For debugging retros
- `retro-epic-ap-123-2026-01-19` - For epic completion retros

**Future Access:**
- `/resume` - Browse recent sessions from same repo
- `--continue` - Continue most recent session

**When to Name:**
- Significant debugging sessions
- Epic completions
- Discovery of reusable patterns

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Create without searching | Search existing first |
| Skip friction analysis | Identify what went wrong |
| Auto-apply CLAUDE.md changes | Require human approval |
| Write vague learnings | Include specific paths |

---

## ao CLI Integration

When ao CLI is available, use it to close the knowledge loop:

```bash
# Index learnings for semantic search
ao forge index .agents/learnings/<learning>.md

# Record retro completion (feeds Knowledge Flywheel)
ao ratchet record retro --input "session" --output ".agents/retros/<retro>.md"

# Query existing patterns before creating new
ao forge search "pattern <topic>" --limit 5
```

The Knowledge Flywheel: `/retro` outputs feed back to `/research` via ao forge indexing.

---

## References

- **Context Gathering**: `references/context-gathering.md`
- **Output Templates**: `references/output-templates.md`

## Related Skills

```
/research -> /plan -> /implement -> /retro
```
