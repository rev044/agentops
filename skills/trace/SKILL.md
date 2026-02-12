---
name: trace
tier: solo
description: 'Trace design decisions and concepts through session history, handoffs, and git. Triggers: "trace decision", "how did we decide", "where did this come from", "design provenance", "decision history".'
dependencies:
  - provenance # alternative - for artifact lineage
---

# Trace Skill

> **Quick Ref:** Trace design decisions through CASS sessions, handoffs, git, and artifacts. Output: `.agents/research/YYYY-MM-DD-trace-*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## When to Use

- Trace HOW architectural decisions evolved
- Find WHEN a concept was introduced
- Understand WHY something was designed a certain way
- Build provenance chain for design decisions

For knowledge artifact lineage (learnings, patterns, tiers), use `/provenance` instead.

**CLI dependencies:** cass (session search). If cass is unavailable, skip transcript search and rely on git log, handoff docs, and `.agents/` artifacts for decision tracing.

## Execution Steps

Given `/trace <concept>`:

### Step 1: Classify Target Type

Determine what kind of provenance to trace:

```
IF target is a file path (contains "/" or "."):
  → Use /provenance (artifact lineage)

IF target is a git ref (sha, branch, tag):
  → Use git-based tracing (Step 2b)

ELSE (keyword/concept):
  → Use design decision tracing (Step 2a)
```

### Step 2a: Design Decision Tracing (Concepts)

Launch 4 parallel search agents (CASS, Handoff, Git, Research) and wait for all to complete.

Read `references/discovery-patterns.md` for agent definitions and prompts.

### Step 2b: Git-Based Tracing (Commits/Refs)

Read `references/discovery-patterns.md` for git-based tracing commands.

### Step 3: Build Timeline

Merge results from all sources into a single chronological timeline (oldest first). Deduplicate same-day/same-session events. Every claim needs a source citation.

### Step 4: Extract Key Decisions

For each event in timeline, identify:
- **What changed:** The decision or evolution
- **Why:** Reasoning if available
- **Who:** Session/author/commit author
- **Evidence:** Link to source (session path, file, commit)

### Step 5: Write Trace Report

**Write to:** `.agents/research/YYYY-MM-DD-trace-<concept-slug>.md`

Read `references/report-template.md` for the full report format and deduplication rules.

### Step 6: Report to User

Tell the user:
1. Concept traced successfully
2. Timeline of evolution (key dates)
3. Most significant decisions
4. Location of trace report
5. Related concepts to explore

## Handling Edge Cases

Read `references/edge-cases.md` for handling: no CASS results, no handoffs, ambiguous concepts (>20 results), and all-sources-empty scenarios. General principle: continue with remaining sources and note gaps in the report.

## Key Rules

- **Search ALL sources** - CASS, handoffs, git, research
- **Build timeline** - chronological evolution is the goal
- **Cite evidence** - every claim needs a source
- **Handle gaps gracefully** - not all concepts are in all sources
- **Write report** - trace must produce `.agents/research/` artifact

## Relationship to /provenance

| Skill | Purpose | Input | Output |
|-------|---------|-------|--------|
| `/provenance` | Artifact lineage | File path | Tier/promotion history |
| `/trace` | Design decisions | Concept/keyword | Timeline of evolution |

Use `/provenance` for: "Where did this learning come from?"
Use `/trace` for: "How did we decide on this architecture?"

## Examples

```bash
# Trace a design decision
/trace "three-level architecture"

# Trace a role/concept
/trace "Chiron"

# Trace a pattern
/trace "brownian ratchet"

# Trace a feature
/trace "parallel wave execution"
```
