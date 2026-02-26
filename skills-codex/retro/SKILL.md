---
name: retro
description: 'Extract learnings from completed work. Trigger phrases: "run a retrospective", "extract learnings", "what did we learn", "lessons learned", "capture lessons", "create a retro".'
---


# Retro Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Extract learnings from completed work, propose proactive improvements, and feed the knowledge flywheel.

## Execution Steps

Given `$retro [topic] [--vibe-results <path>]`:

### Step 1: Identify What to Retrospect

**If vibe results path provided:** Read and incorporate validation findings:
```
Tool: Read
Parameters:
  file_path: <vibe-results-path>
```

This allows post-mortem to pass validation context without re-running vibe.

**If topic provided:** Focus on that specific work.

**If no topic:** Look at recent activity:
```bash
# Recent commits
git log --oneline -10 --since="7 days ago"

# Recent issues closed
bd list --status closed --since "7 days ago" 2>/dev/null | head -5

# Recent research/plans
ls -lt .agents/research/ .agents/plans/ 2>/dev/null | head -5
```

### Step 2: Gather Context

Read relevant artifacts:
- Research documents
- Plan documents
- Commit messages
- Code changes

Use the Read tool and git commands to understand what was done.

### Step 3: Identify Learnings

**If vibe results were provided, incorporate them:**
- Extract learnings from CRITICAL and HIGH findings
- Note patterns that led to issues
- Identify anti-patterns to avoid

Ask these questions:

**What went well?**
- What approaches worked?
- What was faster than expected?
- What should we do again?

**What went wrong?**
- What failed?
- What took longer than expected?
- What would we do differently?
- (Include vibe findings if provided)

**What did we discover?**
- New patterns found
- Codebase quirks learned
- Tool tips discovered
- Debugging insights

### Step 4: Extract Actionable Learnings

For each learning, capture:
- **ID**: L1, L2, L3...
- **Category**: debugging, architecture, process, testing, security
- **What**: The specific insight
- **Why it matters**: Impact on future work
- **Confidence**: high, medium, low

### Step 5: Write Learnings

**Write to:** `.agents/learnings/YYYY-MM-DD-<topic>.md`

```markdown
# Learning: <Short Title>

**ID**: L1
**Category**: <category>
**Confidence**: <high|medium|low>

## What We Learned

<1-2 sentences describing the insight>

## Why It Matters

<1 sentence on impact/value>

## Source

<What work this came from>

---

# Learning: <Next Title>

**ID**: L2
...
```

### Step 5.5: Classify Learning Scope

For each learning extracted in Step 5, classify:

**Question:** "Does this learning reference specific files, packages, or architecture in THIS repo? Or is it a transferable pattern that helps any project?"

- **Repo-specific** → Write to `.agents/learnings/` (existing behavior from Step 5). Use `git rev-parse --show-toplevel` to resolve repo root — never write relative to cwd.
- **Cross-cutting/transferable** → Rewrite to remove repo-specific context (file paths, function names, package names), then:
  1. Write abstracted version to `~/.agents/learnings/YYYY-MM-DD-<slug>.md` (NOT local — one copy only)
  2. Run abstraction lint check:
     ```bash
     file="<path-to-written-global-file>"
     grep -iEn '(internal/|cmd/|\.go:|/pkg/|/src/|AGENTS\.md|CLAUDE\.md)' "$file" 2>/dev/null
     grep -En '[A-Z][a-z]+[A-Z][a-z]+\.(go|py|ts|rs)' "$file" 2>/dev/null
     grep -En '\./[a-z]+/' "$file" 2>/dev/null
     ```
     If matches: WARN user with matched lines, ask to proceed or revise. Never block the write.

**Note:** Each learning goes to ONE location (local or global). No `promoted_to` needed — there's no local copy to mark when `$retro` writes directly to global.

**Example abstraction:**
- Local: "Athena's validate package needs O_CREATE|O_EXCL for atomic claims because Zeus spawns concurrent workers"
- Global: "Use O_CREATE|O_EXCL for atomic file creation when multiple processes may race on the same path"
### Step 5.6: Compile Constraint Templates

For each extracted learning scoring >= 4/5 on actionability AND tagged "constraint" or "anti-pattern", run `bash hooks/constraint-compiler.sh <learning-path>` to generate a constraint template.

```bash
# Compile high-scoring constraint/anti-pattern learnings into enforcement templates
for f in .agents/learnings/YYYY-MM-DD-*.md; do
    [ -f "$f" ] || continue
    bash hooks/constraint-compiler.sh "$f" 2>/dev/null || true
done
```

This produces draft constraint templates in `.agents/constraints/` that can later be activated via `ao quality constraint activate <id>`.

### Step 6: Write Retro Summary

**Write to:** `.agents/retros/YYYY-MM-DD-<topic>.md`

```markdown
# Retrospective: <Topic>

**Date:** YYYY-MM-DD
**Scope:** <what work was reviewed>

## Summary
<1-2 sentence overview>

## What Went Well
- <thing 1>
- <thing 2>

## What Could Be Improved
- <improvement 1>
- <improvement 2>

## Learnings Extracted
- L1: <brief>
- L2: <brief>

See: `.agents/learnings/YYYY-MM-DD-<topic>.md`

## Proactive Improvement Agenda

| # | Area | Improvement | Priority | Horizon | Effort | Evidence |
|---|------|-------------|----------|---------|--------|----------|
| 1 | repo / execution / CI | <improvement> | P0/P1/P2 | now/next-cycle/later | S/M/L | <retro evidence> |

### Recommended Next $rpi
$rpi "<highest-value item>"

## Action Items
- [ ] <any follow-up needed>
```

### Step 6.5: Proactive Improvement Agenda (MANDATORY)

After writing the retro summary, use the full context you just gathered to propose concrete improvements.

Ask explicitly:
1. **Repo:** What should we improve in the codebase/contracts/docs to reduce future defects?
2. **Execution:** What should we improve in planning/implementation/review workflow to increase throughput?
3. **CI/Automation:** What should we improve in validation gates/tooling to reduce noise and catch regressions earlier?

Requirements:
- Propose at least **5** items total.
- Cover all three areas above (repo, execution, CI/automation).
- Include at least **1 quick win** (small, low-risk, same-session viable).
- For each item include: `priority` (P0/P1/P2), `horizon` (now/next-cycle/later), `effort` (S/M/L), and one-line rationale tied to retro evidence.
- Mark one item as **Recommended Next $rpi**.

Write this into the retro file under:
```markdown
## Proactive Improvement Agenda

| # | Area | Improvement | Priority | Horizon | Effort | Evidence |
|---|------|-------------|----------|---------|--------|----------|
| 1 | CI | <improvement> | P0 | now | S | <retro evidence> |

### Recommended Next $rpi
$rpi "<highest-value item>"
```

### Step 7: Feed the Knowledge Flywheel (auto-extract)

```bash
# If ao available, index via forge, close session, and trigger flywheel
if command -v ao &>/dev/null; then
  ao know forge markdown .agents/learnings/YYYY-MM-DD-*.md 2>/dev/null
  echo "Learnings indexed in knowledge flywheel"

  # Apply feedback from completed tasks to associated learnings
  ao task-feedback 2>/dev/null
  echo "Task feedback applied"

  # Close session and trigger full flywheel close-loop
  ao work session close 2>/dev/null || true
  ao quality flywheel close-loop --quiet 2>/dev/null || true
  echo "Session closed, flywheel loop triggered"

  # Sync insights to MEMORY.md immediately (don't wait for session end)
  ao settings notebook update --quiet 2>/dev/null || true

  # Flag stale constraints for retirement
  ao quality constraint review 2>/dev/null || true
else
  # Learnings are already written to .agents/learnings/ by Step 5.
  # Without ao CLI, grep-based search in $research, $knowledge, and $inject
  # will find them directly — no copy to pending needed.

  # Build lightweight keyword index for faster search
  mkdir -p .agents/ao
  for f in .agents/learnings/YYYY-MM-DD-*.md; do
    [ -f "$f" ] || continue
    TITLE=$(head -1 "$f" | sed 's/^# //')
    echo "{\"file\": \"$f\", \"title\": \"$TITLE\", \"keywords\": [], \"timestamp\": \"$(date -Iseconds)\"}" >> .agents/ao/search-index.jsonl
  done
  echo "Learnings indexed locally (ao CLI not available — grep-based search active)"
fi
```

This auto-extraction step ensures every retro feeds the flywheel without requiring the user to remember manual commands.

### Step 8: Report to User

Tell the user:
1. Number of learnings extracted
2. Key insights (top 2-3)
3. Location of retro and learnings files
4. Knowledge has been indexed for future sessions
5. Top proactive improvements (top 3) + recommended next `$rpi`

## Key Rules

- **Be specific** - "auth tokens expire" not "learned about auth"
- **Be actionable** - learnings should inform future decisions
- **Cite sources** - reference what work the learning came from
- **Write both files** - retro summary AND detailed learnings
- **Be proactive** - always produce repo + execution + CI improvements from gathered context
- **Index knowledge** - make it discoverable

## The Flywheel

Learnings feed future research:
```
Work → $retro → improvements + learnings → ao know forge markdown → $research finds it
```

Future sessions start smarter because of your retrospective.

## Examples

### Retrospective After Implementation

**User says:** `$retro`

**What happens:**
1. Agent looks at recent activity via `git log --oneline -10`
2. Agent finds 8 commits related to authentication refactor
3. Agent reads commit messages, code changes, and related issue in beads
4. Agent asks: What went well? What went wrong? What was discovered?
5. Agent identifies 4 learnings: L1 (token expiry pattern), L2 (middleware ordering matters), L3 (test coverage caught edge case), L4 (documentation prevents support load)
6. Agent writes learnings file to `.agents/learnings/2026-02-13-auth-refactor.md`
7. Agent writes retro summary to `.agents/retros/2026-02-13-auth-refactor.md`
8. Agent runs `ao know forge markdown` to add learnings to knowledge base

**Result:** 4 learnings extracted and indexed, retro summary documents what went well and improvements needed.

### Post-Mortem with Vibe Results

**User says:** `$retro --vibe-results .agents/council/2026-02-13-vibe-api.md`

**What happens:**
1. Agent reads vibe results file showing 2 CRITICAL and 3 HIGH findings
2. Agent extracts learnings from validation findings (race condition pattern, missing input validation)
3. Agent reviews recent commits for context
4. Agent creates 6 learnings: 2 from vibe findings (what to avoid), 4 from successful patterns (what to repeat)
5. Agent writes both learnings and retro files
6. Agent indexes knowledge automatically via ao know forge

**Result:** Vibe findings incorporated into learnings, preventing same issues in future work.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| No recent activity found | Clean git history or work not committed yet | Ask user what to retrospect. Accept manual topic: `$retro "planning process improvements"`. Review uncommitted changes if needed. |
| Learnings too generic | Insufficient analysis or surface-level review | Dig deeper into code changes. Ask "why" repeatedly. Ensure learnings are actionable (specific pattern, not vague principle). Check confidence level. |
| ao know forge markdown fails | ao CLI not installed or .agents/ structure wrong | Graceful fallback: index learnings locally to `.agents/ao/search-index.jsonl`. Notify user ao not available. Learnings still in `.agents/learnings/` and discoverable via grep-based search. |
| Duplicate learnings extracted | Same insight from multiple sources | Deduplicate before writing. Check existing learnings with grep. Merge duplicates into single learning with multiple source citations. |

## Reference Documents

- [references/context-gathering.md](references/context-gathering.md)
- [references/output-templates.md](references/output-templates.md)

---

## References

### context-gathering.md

# Context Gathering

How to collect rich context from multiple sources for retrospectives.

## Target Identification

### If Epic ID Provided

```bash
bd show $ARGUMENTS
# Extract child issue IDs, query each for comments
```

### If Topic/Plan Provided

```bash
ls .agents/plans/*$ARGUMENTS* 2>/dev/null
ls .agents/research/*$ARGUMENTS* 2>/dev/null
```

### If No Argument

```bash
bd list --status closed | head -10
```

---

## Conversation Analysis

If a session ID is available, analyze the Codex conversation to extract:
- Decisions made during implementation
- Friction encountered (errors, retries, workarounds)
- Patterns discovered or followed
- Lessons learned

```bash
# Analyze specific session
python3 ~/.claude/scripts/analyze-sessions.py --session=$SESSION_ID

# Analyze with extraction limits for large sessions
python3 ~/.claude/scripts/analyze-sessions.py --session=$SESSION_ID --limit=50
```

### Conversation Data → Retro Output Mapping

| Conversation Data | Retro Output |
|-------------------|--------------|
| `DecisionExtraction` | `.agents/retros/` - Decisions section |
| `QualityResult.issues` | Friction detection |
| `DocExtraction(type="warning")` | `.agents/learnings/` |
| `DecisionExtraction(type="pattern")` | `.agents/patterns/` |
| `DecisionExtraction(type="lesson")` | What Worked / What Didn't |

### Session ID Sources

1. **Environment variable**: `$CLAUDE_SESSION_ID` (set by Codex)
2. **Recent session detection**: Find most recent `.jsonl` in `~/.claude/projects/`
3. **Beads comment**: Sessions may be recorded in crank state

### When No Session Available

Fall back to git analysis and beads comments. Note in retro that conversation
analysis was unavailable.

---

## Git Commit Analysis

```bash
git log --oneline --since="7 days ago" | grep -E "(ai-platform-[a-z0-9]+|$TOPIC)"
git show <commit-hash> --stat
```

**Extract:** Files modified, commit messages, lines changed.

---

## Beads Comments

```bash
bd show <epic-id>
bd show <child-id>
```

**Extract:** Decisions, blockers, workarounds, root causes.

---

## Blackboard State

```bash
ls .agents/blackboard/
cat .agents/blackboard/crank-state.json 2>/dev/null
```

---

## Friction Detection

### Friction Keywords

```bash
# Look for friction keywords in comments
grep -i "error\|failed\|retry\|workaround\|fixed by\|root cause" <comments>

# Look for fix commits
git log --oneline | grep -i "fix\|revert\|hotfix\|patch"
```

### Search Prior Solutions

```bash
ls .agents/learnings/ | grep -i "$TOPIC"
ls .agents/patterns/ | grep -i "$TOPIC"
```

**If found:** Reference in proposal.
**If not found:** Mark as "NEW" for learning extraction.

### Friction Categories

| Category | Indicators |
|----------|------------|
| Retry/Failure | "Error:", "Failed:", test failures |
| Manual Fix | User corrections after agent action |
| Blocking | Dependency issues |
| Pattern Deviation | Didn't follow established pattern |
| Missing Information | Had to search for docs |

### Friction → Fix Mapping

| Friction | Fix Location |
|----------|--------------|
| Command unclear | `.claude/commands/*.md` |
| Skill trigger missed | `.claude/skills/**/*.md` |
| Pattern not followed | `.agents/patterns/*.md` |
| Convention violated | `CLAUDE.md` |

---

## Supersession Check

### Search Existing Artifacts

```bash
mcp__smart-connections-work__lookup --query="$TOPIC" --limit=10
grep -rl "$TOPIC" .agents/learnings/ .agents/patterns/ 2>/dev/null
```

### Supersession Criteria

| Criterion | Supersede? |
|-----------|------------|
| Same topic, newer insight | Yes |
| Same topic, complementary | No (cross-reference) |
| Obsolete/incorrect info | Yes |

### Metadata

Old artifact:
```yaml
superseded_by: .agents/learnings/YYYY-MM-DD-new.md
```

New artifact:
```yaml
supersedes: .agents/learnings/YYYY-MM-DD-old.md
```

### output-templates.md

# Retro Output Templates

Document templates for retro, learnings, and patterns.

---

## Tag Vocabulary Reference

See `.claude/includes/tag-vocabulary.md` for the complete tag vocabulary.

**Document type tags:** `retro`, `learning`, `pattern`

**Examples:**
- `[retro, agents, mcp]` - MCP server implementation retro
- `[learning, data, neo4j]` - GraphRAG implementation learning
- `[pattern, testing, python]` - Python testing pattern

---

## Retro Summary Template

Write to `.agents/retros/YYYY-MM-DD-{topic}.md`:

```markdown
---
date: YYYY-MM-DD
type: Learning
topic: "[Topic]"
tags: [retro, domain-tag, optional-tech-tag]
status: COMPLETE
---

# Retrospective: [Topic]

**Date:** YYYY-MM-DD
**Epic:** [beads epic ID if applicable]
**Duration:** [Single session | Multi-session | Sprint]

---

## What We Accomplished

[Summary of work completed with commits, issues closed, metrics]

| Commit | Issue | Description |
|--------|-------|-------------|
| `abc123` | ai-platform-xxx | Feature description |

---

## What Went Well

- [Positive outcome 1]
- [Positive outcome 2]

---

## What Could Improve

- [Area for improvement 1]
- [Area for improvement 2]

---

## Patterns Worth Repeating

[Code blocks or descriptions of reusable patterns discovered]

---

## Remaining Work

[List of open issues or next steps]

---

## Session Stats

| Metric | Value |
|--------|-------|
| Issues closed | X |
| Lines added | ~Y |

## Source Performance

[Include if analytics data available from Phase 1.5]

| Source | Tier | Value Score | Expected | Deviation |
|--------|------|-------------|----------|-----------|
| [source_type] | [tier] | [value_score] | [expected_weight] | [deviation] |

### Tier Weight Recommendations

[List any recommendations from analytics endpoint]

- **PROMOTE/DEMOTE**: '[source_type]' [over/under]performing by [%]. Consider [action].
```

**Tag Rules:** First tag MUST be `retro`. Include domain tag.

---

## Learning File Template

Write to `.agents/learnings/YYYY-MM-DD-{topic}.md`:

**Tag Rules:** 3-5 tags. First tag MUST be `learning`. At least one domain tag required.

```markdown
---
date: YYYY-MM-DD
type: Learning
topic: "[Topic]"
source: "[beads ID or plan file]"
tags: [learning, domain-tag, optional-tech-tag]
status: COMPLETE
---

# Learning: [Topic]

## Context
[What were we trying to do?]

## What We Learned

### [Learning 1]
**Type:** Technical | Process | Pattern | Gotcha

[Description]

**Evidence:** [File path, beads comment, or observation]

**Application:** [How to use this knowledge in the future]

### [Learning 2]
...

## Discovery Provenance

Track which sources led to these learnings (enables flywheel optimization).

**Purpose**: Create measurement data for the knowledge flywheel. Analytics can then measure: "Which discovery sources produce the most cited, most valuable knowledge?"

**Format**:
```markdown
| Learning | Source Type | Source Detail |
|----------|-------------|---------------|
| [Learning 1] | [type] | [detail] |
| [Learning 2] | [type] | [detail] |
```

> **Note:** Do NOT include a "Confidence" column. Confidence/relevance are query-time metrics, not storage-time. See `domain-kit/skills/standards/references/rag-formatting.md`.

**Example**:
```markdown
| Middleware pattern works well | smart-connections | "request lifecycle" query |
| Rate limit algorithm at L89 | grep | services/limits.py:89 |
| Precedent from prior work | prior-research | 2026-01-01-limits.md |
```

**Source types by tier**:
- **Tier 1**: `code-map`
- **Tier 2**: `smart-connections`, `athena-knowledge`
- **Tier 3**: `grep`, `glob`
- **Tier 4**: `read`, `lsp`
- **Tier 5**: `prior-research`, `prior-retro`, `prior-pattern`, `memory-recall`
- **Tier 6**: `web-search`, `web-fetch`

**How it feeds the flywheel**:
1. You document source_type for each learning during retro
2. Session analyzer extracts these and stores as memories with source_type field
3. `GET /memories/analytics/sources` computes value_score for each source
4. High-value sources (value_score > 0.7) get promoted in discovery tier ordering
5. Future research prioritizes high-value sources = better decisions

## Related
- Plan: [link to plan file if applicable]
- Research: [link to research file if applicable]
- Issues: [beads IDs]
```

---

## Pattern File Template

Write to `.agents/patterns/`:

**Tag Rules:** 3-5 tags. First tag MUST be `pattern`. At least one domain tag required.

```markdown
---
date: YYYY-MM-DD
type: Pattern
category: "[Category]"
tags: [pattern, domain-tag, optional-tech-tag]
status: ACTIVE
---

# Pattern: [Name]

## When to Use
[Triggering conditions]

## The Pattern
[Step-by-step or code example]

## Why It Works
[Rationale]

## Examples
[Real examples from codebase with file paths]
```

---

## Progress Output Templates

### Context Summary

```
================================================================
CONTEXT GATHERED
================================================================

Epic: ai-platform-xxxx
Title: [Epic title]
Duration: [Days from first to last commit]

Sources Analyzed:
  - Commits: 12 (abc123..def456)
  - Issues: 5 (3 closed, 2 open)
  - Files modified: 8
  - Blackboard entries: 2
  - Commands used: 14

Key Files:
  - .claude/commands/retro.md (major changes)
  - services/etl/app/main.py (new)
  - tests/test_etl.py (new)

Ready for Phase 2: Identify Improvements
================================================================
```

### Friction Analysis

```
================================================================
FRICTION ANALYSIS COMPLETE
================================================================

Friction Points Found: 5
  [HIGH] Epic-child dependency confusion (3 occurrences)
  [MEDIUM] Wave detection unclear (2 occurrences)
  [LOW] Commit message format inconsistent (1 occurrence)

Improvement Opportunities: 3
  1. Update $plan command with dependency warning
  2. Add wave auto-detection to /load-epic
  3. Document commit message format in CLAUDE.md

New Patterns Discovered: 1
  - Comment-based epic-child linking

Ready for Phase 3: Propose Changes
================================================================
```

### User Review Display

```
================================================================
IMPROVEMENT PROPOSALS
================================================================

Found 4 improvements (1 critical, 2 recommended, 1 optional)

Would you like to:
1. Review each proposal individually
2. Apply all CRITICAL and RECOMMENDED (skip OPTIONAL)
3. Apply all proposals
4. Skip improvements (proceed to retro summary only)

================================================================
```

### Changes Applied

```
================================================================
CHANGES APPLIED
================================================================

Successfully applied: 3/3 proposals

Files modified:
  * .claude/commands/plan.md
  * .claude/commands/load-epic.md
  * CLAUDE.md

Commit: abc1234

Skipped (user choice): 1
  - .agents/patterns/comment-based-linking.md

Failed: 0

================================================================
```

### Supersession Report

```
================================================================
SUPERSESSION CHECK COMPLETE
================================================================

Searched for: "[topic]"
Candidates found: 3
Superseded: 1

Supersession applied:
  * .agents/learnings/2025-11-15-old-pattern.md
    -> superseded by: .agents/learnings/2025-12-31-new-pattern.md
    -> Reason: Updated approach with better performance

Cross-references added (not superseded):
  - .agents/patterns/related-pattern.md
    -> Added to "Related" section

No action needed:
  - .agents/retros/2025-10-01-unrelated.md
    -> Different topic, no relationship

Ready for Phase 5: User Review
================================================================
```

### Final Output

```
Retro complete:
- Summary: .agents/retros/YYYY-MM-DD-topic.md
- Learnings: .agents/learnings/YYYY-MM-DD-topic.md (if applicable)
- Patterns: [updated/created files] (if applicable)

This knowledge is now persistent and available to future sessions.
```


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
check "SKILL.md has name: retro" "grep -q '^name: retro' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 1 file" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 1 ]"
check "SKILL.md mentions .agents/learnings/ output" "grep -q '\.agents/learnings/' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions .agents/retros/ output" "grep -q '\.agents/retros/' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions knowledge flywheel" "grep -qi 'flywheel' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions ao know forge" "grep -q 'ao know forge' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions learning categories" "grep -qi 'category\|debugging\|architecture\|process' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions confidence levels" "grep -qi 'confidence.*high\|high.*medium.*low' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions vibe results integration" "grep -qi 'vibe.results\|vibe-results' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


