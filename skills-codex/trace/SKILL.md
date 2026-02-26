---
name: trace
description: 'Trace design decisions and concepts through session history, handoffs, and git. Triggers: "trace decision", "how did we decide", "where did this come from", "design provenance", "decision history".'
---


# Trace Skill

> **Quick Ref:** Trace design decisions through CASS sessions, handoffs, git, and artifacts. Output: `.agents/research/YYYY-MM-DD-trace-*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## When to Use

- Trace HOW architectural decisions evolved
- Find WHEN a concept was introduced
- Understand WHY something was designed a certain way
- Build provenance chain for design decisions

For knowledge artifact lineage (learnings, patterns, tiers), use `$provenance` instead.

**CLI dependencies:** cass (session search). If cass is unavailable, skip transcript search and rely on git log, handoff docs, and `.agents/` artifacts for decision tracing.

## Execution Steps

Given `$trace <concept>`:

### Step 1: Classify Target Type

Determine what kind of provenance to trace:

```
IF target is a file path (contains "/" or "."):
  → Use $provenance (artifact lineage)

IF target is a git ref (sha, branch, tag):
  → Use git-based tracing (Step 2b)

ELSE (keyword/concept):
  → Use design decision tracing (Step 2a)
```

### Step 2a: Design Decision Tracing (Concepts)

Launch 4 parallel search agents (CASS, Handoff, Git, Research) and wait for all to complete.

**Backend:** Agents use `Task(subagent_type="Explore")` which maps to `task(subagent_type="explore")` in OpenCode. See `skills/shared/SKILL.md` ("Runtime-Native Spawn Backend Selection") for the shared contract.

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

## Relationship to $provenance

| Skill | Purpose | Input | Output |
|-------|---------|-------|--------|
| `$provenance` | Artifact lineage | File path | Tier/promotion history |
| `$trace` | Design decisions | Concept/keyword | Timeline of evolution |

Use `$provenance` for: "Where did this learning come from?"
Use `$trace` for: "How did we decide on this architecture?"

## Examples

```bash
# Trace a design decision
$trace "three-level architecture"

# Trace a role/concept
$trace "Chiron"

# Trace a pattern
$trace "brownian ratchet"

# Trace a feature
$trace "parallel wave execution"
```

### Tracing an Architectural Decision

**User says:** `$trace "agent team protocol"`

**What happens:**
1. Agent classifies target as concept (not file path or git ref)
2. Agent launches 4 parallel agents: CASS search, handoff search, git log search, research artifact search
3. CASS finds 8 sessions mentioning "agent team", handoff finds 2 docs, git finds 3 commits, research finds 1 analysis
4. Agent builds chronological timeline from 2026-01-15 (first mention) to 2026-02-08 (latest update)
5. Agent extracts 5 key decisions: initial send-message design, team-create addition, deliberation protocol, in-process mode, delegate mode
6. Agent writes trace report to `.agents/research/2026-02-13-trace-agent-team-protocol.md` with full timeline and citations

**Result:** Complete evolution timeline showing how agent team protocol developed across 7 sessions with source citations.

### Tracing from Git Commit

**User says:** `$trace abc1234`

**What happens:**
1. Agent detects git ref format (short sha)
2. Agent runs git-based tracing commands to get commit details, changed files, related commits
3. Agent uses `git log --grep` to find related work
4. Agent searches `.agents/` for contemporary research/plans
5. Agent builds timeline focused on that specific change
6. Agent writes report showing commit context, what changed, why (from commit message and related docs)

**Result:** Trace report links commit to broader design context from surrounding artifacts.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| CASS returns no results | Session search not installed or query too specific | Check `which cass`. If missing, skip CASS and rely on handoffs/git/research. Try broader query terms. |
| Timeline has gaps | Not all decisions documented in searchable artifacts | Note gaps in report. Suggest interviewing team members or checking Slack/email archives for missing context. |
| Too many results (>50 matches) | Very broad concept or high-frequency term | Read `references/edge-cases.md` for ambiguous concept handling. Narrow query or filter by date range. Ask user for more specific aspect to trace. |
| Empty trace report (all sources failed) | Concept genuinely undocumented or typo | Verify spelling. Try synonyms. Report to user: "No documented history found. This may be a new concept or may need different search terms." |

---

## References

### discovery-patterns.md

# Discovery Patterns

Parallel search agent definitions for design decision tracing.

## Design Decision Tracing (Concepts) — Parallel Agents

Launch all 4 agents in parallel using the Task tool, then wait for all to complete.

### Agent 1: CASS Session Search

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  model: "haiku"
  description: "CASS search: <concept>"
  prompt: |
    Search session transcripts for: <concept>

    Run this command:
    cass search "<concept>" --json --limit 10

    Parse the JSON output and extract:
    - Session dates (created_at field, convert from Unix ms)
    - Session paths (source_path field)
    - Agents used (agent field)
    - Relevance scores (score field)
    - Key snippets (snippet/content fields)

    Return a structured list sorted by date (oldest first).
```

### Agent 2: Handoff Search

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  model: "haiku"
  description: "Handoff search: <concept>"
  prompt: |
    Search handoff documents for: <concept>

    1. List handoff files:
       ls -la .agents/handoff/*.md 2>/dev/null

    2. Search for concept mentions:
       grep -l "<concept>" .agents/handoff/*.md 2>/dev/null

    3. For each matching file, extract:
       - File date (from filename YYYY-MM-DD)
       - Context around the mention (grep -B5 -A5)
       - Related decisions or questions

    Return a structured list sorted by date.
```

### Agent 3: Git History Search

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  model: "haiku"
  description: "Git search: <concept>"
  prompt: |
    Search git history for: <concept>

    1. Search commit messages:
       git log --oneline --grep="<concept>" | head -20

    2. For interesting commits, get details:
       git show --stat <commit-sha>

    3. Extract:
       - Commit dates
       - Commit messages
       - Files changed
       - Authors

    Return a structured list sorted by date.
```

### Agent 4: Research/Learnings Search

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  model: "haiku"
  description: "Research search: <concept>"
  prompt: |
    Search research and learning artifacts for: <concept>

    1. Search research docs:
       grep -l "<concept>" .agents/research/*.md 2>/dev/null

    2. Search learnings:
       grep -l "<concept>" .agents/learnings/*.md 2>/dev/null

    3. Search patterns:
       grep -l "<concept>" .agents/patterns/*.md 2>/dev/null

    4. For each match, extract:
       - File date (from filename or modification time)
       - Context around the mention
       - Related concepts

    Return a structured list sorted by date.
```

## Git-Based Tracing (Commits/Refs)

For git refs, trace the commit history:

```bash
# Get commit details
git show --stat <ref>

# Get commit ancestry
git log --oneline --ancestry-path <ref>..HEAD | head -20

# Find related commits
git log --oneline --all --grep="$(git log -1 --format=%s <ref> | head -c 50)" | head -10
```

### edge-cases.md

# Edge Cases

How to handle common failure modes during trace execution.

## No CASS Results

```
IF cass search returns 0 results:
  - Log: "No session transcripts mention '<concept>'"
  - Continue with other sources
  - Note in report: "Concept not found in session history"
```

## No Handoff Documents

```
IF .agents/handoff/ doesn't exist OR no matches:
  - Log: "No handoff documents mention '<concept>'"
  - Continue with other sources
  - Note in report: "Concept not documented in handoffs"
```

## Ambiguous Concept (Too Many Results)

```
IF CASS returns >20 results:
  - Show top 10 by score
  - Ask user: "Many sessions mention this. Want to narrow by date range or workspace?"
  - Suggest related but more specific concepts
```

## All Sources Empty

```
IF all 4 searches return nothing:
  - Report: "No provenance found for '<concept>'"
  - Suggest: "Try related terms: <suggestions>"
  - Ask: "Is this concept documented somewhere else?"
```

### report-template.md

# Trace Report Template

Write trace reports to: `.agents/research/YYYY-MM-DD-trace-<concept-slug>.md`

## Full Template

```markdown
# Trace: <Concept>

**Date:** YYYY-MM-DD
**Query:** <original concept>
**Sources searched:** CASS, Handoffs, Git, Research

## Summary

<2-3 sentence overview of how the concept evolved>

## Timeline

| Date | Source | Event | Evidence |
|------|--------|-------|----------|
| ... | ... | ... | ... |

## Key Decisions

### Decision 1: <title>
- **Date:** YYYY-MM-DD
- **Source:** <CASS session / Handoff / Git commit>
- **What:** <what was decided>
- **Why:** <reasoning if known>
- **Evidence:** <link/path>

### Decision 2: <title>
...

## Evolution Summary

<How the concept changed over time, key inflection points>

## Current State

<Where the concept stands now based on most recent evidence>

## Related Concepts

- <related concept 1> - see `$trace <concept1>`
- <related concept 2> - see `$trace <concept2>`

## Sources

### CASS Sessions
| Date | Session Path | Score |
|------|--------------|-------|
| ... | ... | ... |

### Handoff Documents
| Date | File | Context |
|------|------|---------|
| ... | ... | ... |

### Git Commits
| Date | SHA | Message |
|------|-----|---------|
| ... | ... | ... |

### Research/Learnings
| Date | File |
|------|------|
| ... | ... |
```

## Timeline Construction

Merge results from all sources into the Timeline table.

**Deduplication rules:**
- Same content within 24 hours = single event (note multiple sources)
- Same session ID = single event
- Preserve ALL sources as evidence

**Sorting:**
- Chronological order (oldest first)
- Show evolution of the concept over time


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0
check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "name is trace" "grep -q '^name: trace' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "mentions provenance" "grep -qi 'provenance' '$SKILL_DIR/SKILL.md'"
check "mentions decision history" "grep -qi 'decision' '$SKILL_DIR/SKILL.md'"
check "mentions timeline" "grep -qi 'timeline' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


