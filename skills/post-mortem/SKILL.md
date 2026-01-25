---
name: post-mortem
description: >
  Comprehensive post-implementation validation. Combines retro (learnings),
  vibe (code validation), security scanning (Talos), and knowledge extraction
  into a single unified workflow. Triggers: "post-mortem", "validate completion",
  "final check", "wrap up epic", "close out", "what did we learn".
version: 1.0.0
tier: solo
author: "AI Platform Team"
license: "MIT"
context: inline
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task,TodoWrite"
skills:
  - retro
  - vibe
  - beads
---

# Post-Mortem Skill

**The RPI capstone: validate + learn + extract + feed back into the flywheel.**

## Role in the Brownian Ratchet

Post-mortem is the **knowledge ratchet** - the final stage that locks learnings:

| Component | Post-Mortem's Role |
|-----------|-------------------|
| **Chaos** | Implementation produced outcomes (good and bad) |
| **Filter** | Retro extracts what matters, discards noise |
| **Ratchet** | Learnings locked in `.agents/`, MCP, specs |

> **Learnings never go backward. Once ratcheted, knowledge compounds.**

Post-mortem closes the knowledge loop:
```
Implementation → POST-MORTEM → Ratcheted Knowledge → Next Research
                    │
                    ├── .agents/retros/     (locked)
                    ├── .agents/learnings/  (locked)
                    ├── .agents/patterns/   (locked)
                    └── MCP memories        (locked)
```

**The Flywheel Effect:** Each ratcheted learning makes the next cycle faster.
This is why `/post-mortem` is mandatory - skipping it breaks the compounding.

## Philosophy

> "Implementation isn't done until we've validated it, learned from it, and fed that knowledge back into the system."

Post-mortem is the comprehensive POST phase of RPI that closes the knowledge loop. It combines everything that feeds back into the flywheel:

| Component | Purpose | Flywheel Feed |
|-----------|---------|---------------|
| **Retro** | What went wrong/right? | `.agents/retros/`, `.agents/learnings/` |
| **Vibe** | Code quality validation | Issues for findings, quality metrics |
| **Security** | Vulnerability scanning (Talos) | Security posture, CVE tracking |
| **Extract** | Knowledge persistence | MCP memories, `.agents/patterns/` |
| **Spec Update** | Lessons back to source | Enhanced specs for next iteration |

**All roads lead back to Research.** Every output feeds the next cycle.

## Quick Start

```bash
/post-mortem <epic-id>           # Full post-mortem on completed epic
/post-mortem                      # Auto-discover recently completed epic
/post-mortem --skip-security      # Skip security scan (faster)
/post-mortem --update-spec        # Update original spec with lessons
```

---

## Workflow

```
┌───────────────────────────────────────────────────────────────────┐
│                      POST-MORTEM WORKFLOW                          │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Phase 0: Discover     Session-aware epic discovery               │
│       ▼                (parses transcripts, handles compaction)   │
│  Phase 0.5: Chain      Cross-compaction chain traversal           │
│       ▼                (merges work across context resets)        │
│  Phase 0.6: Trace      Root cause tracing                         │
│       ▼                (research → spec → plan → commits)         │
│  Phase 1: Retro        Gather context, identify friction          │
│       ▼                                                           │
│  Phase 2: Vibe         Validate code quality (all 8 aspects)      │
│       ▼                                                           │
│  Phase 3: Security     Talos/guardian security scan               │
│       ▼                                                           │
│  Phase 4: Extract      Learnings, patterns, memories              │
│       ▼                                                           │
│  Phase 5: Feed Back    Update spec, store knowledge               │
│       ▼                                                           │
│  Phase 6: Report       Unified summary + provenance chain         │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

---

## Phase 0: Session-Aware Discovery (Enhanced)

**Problem solved:** The old heuristic (`bd list --since "24 hours ago"`) grabbed the wrong
epic when multiple epics closed recently. The new approach parses session transcripts to
find what was actually worked on.

### Primary: Session Transcript Analysis

```python
# Using the lib/transcript.py module
from lib.transcript import (
    find_session_transcripts,
    parse_transcript,
    extract_epic_from_session,
)
from lib.compaction import find_compaction_chain, build_composite_session

# 1. Find current project's transcripts
transcripts = find_session_transcripts(project_path)

# 2. Parse most recent transcript (handles compaction)
if transcripts:
    chain = find_compaction_chain(transcripts[0])
    session = build_composite_session(chain)
    EPIC = extract_epic_from_session(session)
```

**How it works:**
1. Finds Claude Code transcripts in `~/.claude/projects/{project-path}/`
2. Parses JSONL to extract beads operations (`bd update`, `bd close`)
3. Identifies epic from session work (not just recent closures)
4. Handles compaction chains (work spanning context resets)

### Fallback: Beads Heuristic

Only if session analysis fails (no transcripts, parsing error):

```bash
# Fallback to beads-based discovery
EPIC=$(bd list --type epic --status closed --since "24 hours ago" | head -1 | awk '{print $1}')

if [[ -z "$EPIC" ]]; then
    EPIC=$(bd list --type epic --status in_progress | head -1 | awk '{print $1}')
    OPEN_CHILDREN=$(bd list --parent=$EPIC --status open | wc -l)
    if [[ "$OPEN_CHILDREN" -gt 0 ]]; then
        echo "Epic $EPIC still has $OPEN_CHILDREN open children"
        exit 1
    fi
fi
```

### Gather Changed Files

```bash
# Get all commits from the epic's work
COMMITS=$(git log --oneline --since="7 days ago" --grep="$EPIC" | awk '{print $1}')

# Get changed files
CHANGED_FILES=$(git diff --name-only $COMMITS | sort -u)
```

---

## Phase 0.5: Cross-Compaction Chain (NEW)

**Problem solved:** Work spanning context compactions was lost. Now we merge
all transcript files in the compaction chain.

```python
from lib.compaction import find_compaction_chain, build_composite_session

# Find all related transcripts (same session slug)
chain = find_compaction_chain(transcript_path)

# Merge into unified view
composite = build_composite_session(chain)

# composite now has:
# - All beads operations from before AND after compaction
# - All file changes
# - All commits made
# - Compaction marker locations
```

**Compaction detection markers:**
- "session is being continued"
- "summary below covers"
- "context has been compacted"

---

## Phase 0.6: Root Cause Trace (NEW)

**Problem solved:** Post-mortem only validated code, couldn't trace back to
research/spec/pre-mortem that led to the implementation.

```python
from lib.trace import trace_epic_provenance, format_provenance_report

# Trace full chain from research to result
chain = trace_epic_provenance(
    epic_id=EPIC,
    session=session,
    project_path=project_root,
)

# chain contains:
# - research_artifact: .agents/research/*.md
# - product_brief: .agents/products/*.md
# - pre_mortem_results: .claude/plans/*-agent*.md
# - spec_artifact: .formula.toml or .agents/specs/*.md
# - plan_artifact: .claude/plans/*.md
# - implementation_commits: git commits
# - decisions_made: extracted from session
```

**Include in report:**
```python
provenance_section = format_provenance_report(chain)
```

---

## Phase 1: Retro

**Reuse `/retro` skill logic.**

### 1.1 Context Gathering

```bash
# Epic details
bd show $EPIC

# Recent commits
git log --oneline --since="7 days ago" | head -20

# Blackboard state
ls .agents/blackboard/
```

### 1.2 Friction Identification

Look for:
- Issues that took multiple attempts
- Blocked issues
- Issues with extensive comments
- Unexpected problems

```bash
# Issues with retry history
bd list --parent=$EPIC | while read issue; do
    bd show $issue | grep -i "retry\|blocked\|failed"
done
```

### 1.3 Learnings Extraction

| Learning Type | Source | Output |
|---------------|--------|--------|
| What went wrong | Blocked issues, retries | Learnings artifact |
| What went right | Fast completions | Patterns artifact |
| Process improvements | Session analysis | CLAUDE.md proposals |

---

## Phase 2: Vibe Validation

**Run full vibe on changed code.**

```bash
/vibe recent --create-issues
```

### Aspects to Check

| Aspect | Focus | Critical Threshold |
|--------|-------|-------------------|
| Quality | Code smells, patterns | 0 CRITICAL |
| Security | OWASP, injection, auth | 0 CRITICAL |
| Architecture | Boundaries, coupling | 2 HIGH |
| Complexity | CC, nesting, size | 5 HIGH |
| Semantic | Docstrings, names | 2 HIGH |
| Performance | N+1, leaks | 1 CRITICAL |
| Slop | AI artifacts | 3 HIGH |

### Gate Criteria

| Level | Action |
|-------|--------|
| 0 CRITICAL | Proceed |
| 1+ CRITICAL | Block - must fix before considering complete |
| >10 HIGH | Warning - create follow-up issues |

---

## Phase 3: Security Scan

**Talos-class security validation.**

### 3.1 Static Analysis

```bash
# Secrets detection
gitleaks detect --source . --verbose 2>/dev/null || echo "gitleaks not installed"

# Dependency vulnerabilities (if applicable)
if [[ -f "requirements.txt" ]]; then
    pip-audit -r requirements.txt 2>/dev/null || echo "pip-audit not installed"
fi

if [[ -f "go.mod" ]]; then
    govulncheck ./... 2>/dev/null || echo "govulncheck not installed"
fi

if [[ -f "package.json" ]]; then
    npm audit 2>/dev/null || echo "npm audit failed"
fi
```

### 3.2 Pattern Detection

Run security-focused grep patterns:

```bash
# SQL injection patterns
grep -rn "execute.*%s\|execute.*format\|f\".*SELECT" --include="*.py" . 2>/dev/null

# Command injection patterns
grep -rn "subprocess.*shell=True\|os.system\|eval(" --include="*.py" . 2>/dev/null

# Hardcoded secrets
grep -rn "password.*=.*['\"].*['\"]|api_key.*=.*['\"]" --include="*.py" . 2>/dev/null
```

### 3.3 Expert Review (Optional)

For CRITICAL or HIGH security findings, spawn security expert:

```bash
Task(subagent_type="security-expert", prompt="Deep dive on security findings: ...")
```

---

## Phase 4: Knowledge Extraction

### 4.1 Learnings

Extract concrete learnings to artifacts:

```bash
# Output location
LEARNING_PATH=".agents/learnings/$(date +%Y-%m-%d)-$EPIC.md"
```

**Learning Template:**
```markdown
# Learning: [Title]

**Date:** YYYY-MM-DD
**Epic:** <epic-id>
**Tags:** [learning, topic1, topic2]

## Context
What were we trying to do?

## What We Learned
Concrete insight or pattern discovered.

## Evidence
- Commit abc123: [description]
- Issue xyz: [what happened]

## Application
How to use this knowledge in the future.

## Discovery Provenance
| Insight | Source Type | Source Detail |
|---------|-------------|---------------|
| ... | grep/code-map/etc | file:line |
```

### 4.2 Patterns

If a reusable pattern was discovered:

```bash
PATTERN_PATH=".agents/patterns/$PATTERN_NAME.md"
```

### 4.3 Memories

Store in MCP for future recall:

```bash
mcp__ai-platform__memory_store(
    content="...",
    memory_type="fact",
    source="post-mortem:$EPIC",
    tags=["epic:$EPIC", "pattern:$PATTERN"]
)
```

---

## Phase 5: Feed Back

### 5.1 Spec Update (Optional)

If `--update-spec` is set and original spec exists:

```bash
# Find original spec
SPEC=$(grep -l "$EPIC" .agents/specs/*.md | head -1)

if [[ -n "$SPEC" ]]; then
    # Append lessons learned section
    echo -e "\n## Post-Implementation Learnings\n" >> $SPEC
    echo "Added after $EPIC completion on $(date +%Y-%m-%d):" >> $SPEC
    echo "- [Learning 1]" >> $SPEC
    echo "- [Learning 2]" >> $SPEC
fi
```

### 5.2 Formula Update (Optional)

If a formula was used and can be improved:

```bash
FORMULA=$(ls .formula.toml 2>/dev/null)
if [[ -n "$FORMULA" ]]; then
    # Note: Manual review recommended for formula updates
    echo "Formula found: $FORMULA"
    echo "Consider updating with lessons learned"
fi
```

---

## Phase 6: Report

### Summary Template

```markdown
# Post-Mortem Complete: [Epic Title]

**Epic:** <epic-id>
**Date:** YYYY-MM-DD
**Duration:** X days (from start to close)
**Discovery Method:** Session transcript / Beads heuristic

## Provenance Chain

### Research Phase
- **Artifact:** `.agents/research/YYYY-MM-DD-topic.md`
- **Key Findings:** [extracted from research]

### Product Phase
- **Brief:** `.agents/products/product-name.md` (if applicable)

### Pre-Mortem Phase
- **Results:** `.claude/plans/xxx-agent-xxx.md`
- **Issues Found:** N critical, M important

### Implementation Phase
- **Plan:** `.claude/plans/xxx.md`
- **Commits:** N
- **Key Decisions:**
  1. [Decision from AskUserQuestion or comment]
  2. [Decision from commit message]

## Retro Summary
- **Friction Points:** N identified
- **Key Learning:** [Most important insight]
- **Process Improvements:** M proposed

## Vibe Results
| Severity | Count | Action |
|----------|-------|--------|
| CRITICAL | 0 | N/A |
| HIGH | 3 | Issues created |
| MEDIUM | 7 | Follow-up |

## Security Scan
- **Secrets:** 0 detected
- **Vulnerabilities:** 0 found
- **Patterns:** 2 flagged for review

## Knowledge Extracted
- **Learnings:** .agents/learnings/YYYY-MM-DD-epic.md
- **Patterns:** 1 new pattern discovered
- **Memories:** 3 stored in MCP

## Follow-Up Issues Created
- <issue-1>: Fix HIGH vibe finding
- <issue-2>: Address security pattern

## Spec Feedback
- Original spec: [path]
- Updated: [yes/no]

---

**Next Steps:**
1. Review and merge follow-up issues
2. Consider spec updates for next similar project
3. Knowledge automatically available for next /research
```

---

## Arguments

| Arg | Purpose | Default |
|-----|---------|---------|
| `<epic-id>` | Epic to analyze | Auto-discover |
| `--skip-vibe` | Skip vibe validation | false |
| `--skip-security` | Skip security scan | false |
| `--update-spec` | Update original spec with lessons | false |
| `--deep` | Full security audit + expert routing | false |
| `--create-issues` | Create beads issues for findings | true |

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Skip vibe because "it's done" | Always validate - done isn't done |
| Ignore HIGH findings | Create follow-up issues |
| Generic learnings | Specific, actionable insights |
| Skip memory storage | Store for future recall |
| Forget the spec | Feed lessons back to spec |

---

## Execution Checklist

- [ ] Discovered epic from session transcripts (Phase 0)
- [ ] Fell back to beads heuristic if needed (Phase 0)
- [ ] Handled compaction chain if present (Phase 0.5)
- [ ] Traced provenance chain (Phase 0.6)
- [ ] Ran retro context gathering (Phase 1)
- [ ] Identified friction points (Phase 1)
- [ ] Ran vibe on changed code (Phase 2)
- [ ] Addressed CRITICAL findings (Phase 2)
- [ ] Created issues for HIGH findings (Phase 2)
- [ ] Ran security scan (Phase 3)
- [ ] No secrets or vulnerabilities (Phase 3)
- [ ] Extracted learnings to artifact (Phase 4)
- [ ] Stored memories in MCP (Phase 4)
- [ ] Updated spec if applicable (Phase 5)
- [ ] Generated summary report with provenance (Phase 6)
- [ ] Synced beads (`bd sync`)
- [ ] Committed and pushed (`git push`)

---

## Quick Example

**User**: "/post-mortem"

**Agent workflow**:

```bash
# Phase 0: Session-Aware Discovery
# Parse session transcript to find worked-on epic
python lib/transcript.py --transcript ~/.claude/projects/.../session.jsonl --extract-epic
# Output: ol-qg8

# Verify with beads
bd show ol-qg8
# Epic: R1 Knowledge Loop Closure
# Status: closed
# Children: 5/5 closed

# Phase 0.5: Compaction Chain
# Check for compaction markers in transcript
# Merge all chain files if compacted

# Phase 0.6: Root Cause Trace
python lib/trace.py --epic ol-qg8 --project . --report
# Research: .agents/research/2026-01-22-r1-loop.md
# Pre-mortem: .claude/plans/drifting-riding-thacker-agent.md
# Plan: .claude/plans/drifting-riding-thacker.md
# Commits: 3

# Phase 1: Retro
# Friction: Wrong epic discovered in previous post-mortem
# Learning: Session transcripts more reliable than beads heuristic

# Phase 2: Vibe
/vibe recent
# 0 CRITICAL, 2 HIGH (spec clarity), 5 MEDIUM

# Phase 3: Security
gitleaks detect --source .
# 0 secrets found

# Phase 4: Extract
# Learning: Pre-mortem simulation caught 10 failure modes
# Pattern: Simulate N iterations before implementation
# Memory stored: "Pre-mortem prevents implementation failures"

# Phase 5: Feed Back
# Spec updated with lessons learned section

# Phase 6: Report
# Summary generated at .agents/retros/2026-01-22-jc-9tx6.md
```

---

## ao CLI Integration

Post-mortem closes the Knowledge Flywheel via ao ratchet:

```bash
# 1. Index all knowledge artifacts
ao forge index .agents/retros/<epic>.md
ao forge index .agents/learnings/<epic>.md
ao forge index .agents/patterns/*.md

# 2. Record completion with provenance chain
ao ratchet record post-mortem \
  --input "epic:<epic-id>" \
  --output "retro:$(ls .agents/retros/*<epic>* | head -1)" \
  --output "learnings:$(ls .agents/learnings/*<epic>* | head -1)"

# 3. Verify loop closure
ao ratchet verify --epic "<epic-id>"
```

The ratchet locks progress: once indexed, future `/research` will discover these artifacts.

---

## References

### JIT-Loadable Documentation

| Topic | Reference |
|-------|-----------|
| Security patterns | `references/security-patterns.md` |
| Learning templates | `references/learning-templates.md` |
| Vibe integration | `~/.claude/skills/vibe/SKILL.md` |
| Retro integration | `~/.claude/skills/retro/SKILL.md` |
| RAG formatting | `domain-kit/skills/standards/references/rag-formatting.md` |

### Related Skills

- **retro**: Detailed retrospective (subset of post-mortem)
- **vibe**: Comprehensive validation (called by post-mortem)
- **pre-mortem**: Pre-implementation simulation (before)

---

## Workflow Integration

```
/research → /product → /pre-mortem → /vibe --spec → /plan → /crank → /post-mortem
    ↑                                                                           │
    │                                                                           ▼
    │                                                               ┌───────────────────┐
    │                                                               │ KNOWLEDGE OUTPUTS │
    │                                                               ├───────────────────┤
    │                                                               │ • .agents/retros/ │
    │                                                               │ • .agents/learnings/│
    │◄──────────────────────────────────────────────────────────────│ • .agents/patterns/│
    │              KNOWLEDGE LOOP                                   │ • MCP memories    │
    │         (consumed by next /research)                          └───────────────────┘
```

**Post-mortem is the capstone skill** that closes the loop, ensuring every implementation produces validated code and extracted knowledge that improves future work.

---

## Knowledge Loop Closure (CRITICAL)

**The loop is only closed when post-mortem outputs become research inputs.**

### Output → Input Mapping

| Post-mortem Output | Location | How /research Consumes It |
|--------------------|----------|---------------------------|
| **Retro artifact** | `.agents/retros/YYYY-MM-DD-epic.md` | Prior art search (Tier 5) |
| **Learnings** | `.agents/learnings/YYYY-MM-DD-epic.md` | Prior art search (Tier 5) |
| **Patterns** | `.agents/patterns/*.md` | Prior art search (Tier 5) |
| **MCP memories** | PostgreSQL via ETL | `mcp__ai-platform__memory_recall()` (Tier 2) |
| **Updated specs** | `.agents/specs/*.md` | Prior art search (Tier 5) |

### The 5 Connection Points

```
POST-MORTEM OUTPUTS              →    RESEARCH INPUTS
═══════════════════════════════════════════════════════════════

1. RETROS (.agents/retros/)
   "What went wrong with auth?"
                                 →    ls .agents/retros/ | grep auth
                                      # Found: 2026-01-15-auth-migration-retro.md
                                      # Contains: "Token refresh race condition"

2. LEARNINGS (.agents/learnings/)
   "Concrete insight about testing"
                                 →    ls .agents/learnings/ | grep test
                                      # Found: 2026-01-10-testing-learnings.md
                                      # Contains: "Mock external services, not internal"

3. PATTERNS (.agents/patterns/)
   "Reusable solution to X"
                                 →    ls .agents/patterns/
                                      # Found: retry-with-backoff.md
                                      # Applies to: network failures

4. MCP MEMORIES
   mcp__ai-platform__memory_store(
       content="Always validate JWT exp claim",
       tags=["auth", "jwt", "security"]
   )
                                 →    mcp__ai-platform__memory_recall(
                                          query="JWT authentication patterns"
                                      )
                                      # Returns: "Always validate JWT exp claim"

5. UPDATED SPECS
   Original spec + learnings
                                 →    grep -l "$TOPIC" .agents/specs/
                                      # Found spec with "Post-Implementation Learnings" section
```

### Research Skill Integration (Required)

For the loop to work, `/research` MUST check these sources:

```bash
# Phase: Prior Art (Tier 5) - MUST include:

# 1. Retros from similar work
ls .agents/retros/ 2>/dev/null | grep -i "$TOPIC" | head -5

# 2. Learnings from similar work
ls .agents/learnings/ 2>/dev/null | grep -i "$TOPIC" | head -5

# 3. Patterns that might apply
ls .agents/patterns/ 2>/dev/null

# 4. MCP memory recall
mcp__ai-platform__memory_recall(query="$TOPIC", limit=5)

# 5. Prior research (already in /research)
ls .agents/research/ | grep -i "$TOPIC"
```

### Verification: Is the Loop Closed?

Run this check before starting new research:

```bash
# Check if knowledge from previous work exists
RIG="athena"
TOPIC="authentication"

echo "=== Checking Knowledge Loop for: $TOPIC ==="

echo -e "\n1. Retros:"
ls .agents/retros/*$TOPIC* 2>/dev/null || echo "   None"

echo -e "\n2. Learnings:"
ls .agents/learnings/*$TOPIC* 2>/dev/null || echo "   None"

echo -e "\n3. Patterns:"
grep -l "$TOPIC" .agents/patterns/*.md 2>/dev/null || echo "   None"

echo -e "\n4. MCP Memories:"
# mcp__ai-platform__memory_recall(query="$TOPIC", limit=3)

echo -e "\n5. Prior Research:"
ls .agents/research/*$TOPIC* 2>/dev/null || echo "   None"
```

### Anti-Pattern: Broken Loop

| Symptom | Cause | Fix |
|---------|-------|-----|
| Same mistakes repeated | Retros not consulted | Add retro check to research |
| Reinventing solutions | Patterns not discovered | Add pattern search to research |
| Lost context | MCP memories not recalled | Add memory_recall to research |
| Specs don't improve | Learnings not appended | Enable `--update-spec` in post-mortem |

### Mandatory Loop Closure Checklist

Before considering post-mortem complete:

- [ ] Retro artifact written to `.agents/retros/`
- [ ] Key learnings extracted to `.agents/learnings/`
- [ ] Reusable patterns documented in `.agents/patterns/`
- [ ] Critical insights stored via `mcp__ai-platform__memory_store()`
- [ ] Original spec updated with learnings (if applicable)
- [ ] **VERIFIED**: Future `/research` will find these artifacts
