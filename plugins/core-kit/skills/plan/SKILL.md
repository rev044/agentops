---
name: plan
description: >
  This skill should be used when the user asks to "create a plan",
  "plan implementation", "break down into tasks", "decompose into features",
  "create beads issues from research", "what issues should we create",
  "plan out the work", or needs to convert a goal into executable beads issues.
version: 1.1.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - beads
  - research
---

# Plan Skill

Create structured implementation plans from goals, research, or feature requests.
Produces beads issues with proper dependencies and wave computation for parallel execution.

## Overview

**Core Purpose**: Transform a goal into an executable plan with discrete beads issues,
dependency ordering, and wave-based parallelization for `/autopilot` and `/implement-wave`.

**Key Capabilities**:
- 6-tier context discovery hierarchy
- Prior plan discovery to prevent duplicates
- Feature decomposition with dependency modeling
- Beads issue creation with epic-child relationships
- Wave computation for parallel execution

**When to Use**: Work needs 2+ discrete issues with dependencies.
**When NOT to Use**: Single task (use `/implement`), exploratory (use `/research`).

---

## Instructions

### Phase 0: Rig Detection

**CRITICAL**: All `.agents/` artifacts go to `~/gt/.agents/<rig>/` based on the primary codebase being planned.

**Detection Logic**:
1. Identify which rig's code you're planning (e.g., files in `~/gt/ai-platform/` → `ai-platform`)
2. If planning across multiple rigs, use `_cross-rig`
3. If unknown/unclear, ask user

| Files Being Read | Target Rig | Output Base |
|------------------|------------|-------------|
| `~/gt/athena/**` | `athena` | `~/gt/.agents/athena/` |
| `~/gt/hephaestus/**` | `hephaestus` | `~/gt/.agents/hephaestus/` |
| `~/gt/daedalus/**` | `daedalus` | `~/gt/.agents/daedalus/` |
| Multiple rigs | `_cross-rig` | `~/gt/.agents/_cross-rig/` |

```bash
# Set RIG variable for use in output paths
RIG="athena"  # or hephaestus, daedalus, _cross-rig
mkdir -p ~/gt/.agents/$RIG/plans/
```

---

### Phase 1: Context Discovery

See `~/.claude/skills/research/references/context-discovery.md` for full 6-tier hierarchy.

**Quick version**: Code-Map → Semantic Search → Scoped Grep → Source → .agents/ → External

**Checklist**:
- [ ] Checked code-map index
- [ ] Ran semantic search (if MCP available)
- [ ] Followed signposts to source
- [ ] Reviewed .agents/ patterns WITH verification

---

### Phase 1.5: Prior Plan Discovery

Before creating new plans, check for existing work:

```bash
# Town-level plans (Mayor/orchestration work)
grep -l "<goal keywords>" ~/gt/.agents/$RIG/plans/*.md 2>/dev/null | head -5
grep -l "<goal keywords>" ~/gt/.agents/_cross-rig/plans/*.md 2>/dev/null | head -5

# Crew workspace plans (implementation work - may have older artifacts)
grep -l "<goal keywords>" ~/gt/$RIG/crew/boden/.agents/plans/*.md 2>/dev/null | head -5

# Existing beads epics
bd list --type=epic | grep -i "<goal keywords>"
```

**Note**: Prior plans may exist in either location:
- **Town-level** (`~/gt/.agents/<rig>/plans/`) - Mayor/orchestration plans
- **Crew workspace** (`~/gt/<rig>/crew/boden/.agents/plans/`) - Implementation plans

| Prior Plan Status | Action |
|-------------------|--------|
| COMPLETE | Reference it, don't duplicate |
| ACTIVE | Extend existing plan |
| BLOCKED | Understand blockers first |

---

### Phase 2: Research/Analysis

Deepen understanding with targeted exploration:

```
Task(
    subagent_type="Explore",
    model="haiku",
    prompt="Find all code related to: $GOAL"
)
```

**Model**: Use `haiku` for exploration (fast, cheap).

**Identify**: Affected files, existing patterns, related tests, blockers.

---

### Phase 3: Decompose into Features

Each feature should be:
- Completable in a single focused session
- Testable independently
- Following existing patterns

#### Dependency Direction (CRITICAL)

**Rule:** `bd dep add A B` means "A waits for B"

| Command | Meaning |
|---------|---------|
| `bd dep add A B` | A depends on B (B must complete first) |
| `bd dep add child epic` | **WRONG** - Creates deadlock! |

#### Epic-Child Relationship

**Epics and children have NO dependency relationship.** Track children via comment:

```bash
# 1. Create epic
bd create "Epic: OAuth2 Support" --type epic --priority P1
# -> ai-platform-101

# 2. Create children (NO dependency on epic!)
bd create "Add callback endpoint" --type feature --priority P1
# -> ai-platform-102

# 3. Dependencies ONLY between children
bd dep add ai-platform-103 ai-platform-102

# 4. Track children via comment (NOT dependency!)
bd comments add ai-platform-101 "Children: ai-platform-102, ai-platform-103"
```

#### Wave Computation

| Wave | Rule |
|------|------|
| Wave 1 | Issues with NO dependencies |
| Wave N | Issues where ALL deps are in Wave N-1 or earlier |

---

### Phase 3.5: Verify Agent Dependencies

For agent-related work (KAgents, MCP servers):

```bash
grep -A5 "agents:" charts/ai-platform/values.yaml
ls services/mcp-servers/
```

---

### Phase 4: Create Beads Issues

```bash
# Create epic
bd create "Epic: $GOAL" --type epic --priority P1

# Create features
bd create "Feature description" --type feature --priority P2

# Set dependencies
bd dep add ai-platform-103 ai-platform-102

# REQUIRED: File annotations for wave parallelization
bd comment <id> "Files affected: src/auth/middleware.py, tests/test_auth.py"
```

---

### Phase 5: Write Plan to Memory

Write to `~/gt/.agents/$RIG/plans/YYYY-MM-DD-{goal-slug}.md`

See `references/templates.md` for full template. Key elements:
- Frontmatter with date, goal, epic ID, tags
- Features table with dependencies
- Wave execution order table
- Dependency graph (ASCII)
- Autopilot handoff section

---

### Phase 6: Output Summary

Output structured summary with autopilot handoff:

```markdown
# Plan Complete: [Goal]

**Epic:** `ai-platform-xxx`
**Plan:** `~/gt/.agents/$RIG/plans/YYYY-MM-DD-goal-slug.md`
**Issues:** N features across M waves

## Wave Execution Order
| Wave | Issues | Can Parallel |
|------|--------|--------------|
| 1 | ai-platform-102, ai-platform-106 | Yes |
| 2 | ai-platform-103 | No |

## Ready for Autopilot
```bash
/autopilot ai-platform-xxx --dry-run
/autopilot ai-platform-xxx
```
```

---

## Epic Lifecycle Checklist

- [ ] Create epic with `bd create --type epic`
- [ ] Create children (NO dependency on epic!)
- [ ] Add children list to epic via `bd comment`
- [ ] **Start epic**: `bd update <epic> --status in_progress`
- [ ] Implement children via `/implement` or `/implement-wave`
- [ ] **Close epic**: `bd close <epic>` after all children done
- [ ] Run `bd sync && git push` before session end

**Common Gotcha**: Children can't become `ready` until epic is `in_progress`.

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Create children depending on epic | Track via comment on epic |
| Skip file annotations | Add "Files affected: ..." |
| Create 10+ features at once | Group into 3-5 per wave |
| Skip prior plan check | Search ~/gt/.agents/$RIG/plans/ first |
| Grep source blindly | Use code-map signposts |
| Forget to start epic | `bd update <epic> --status in_progress` |

---

## Execution Checklist

- [ ] Researched codebase context (6-tier hierarchy)
- [ ] Checked for prior plans
- [ ] Decomposed into discrete features
- [ ] Created beads issues with `bd create`
- [ ] Set dependencies with `bd dep add`
- [ ] Added file annotations with `bd comment`
- [ ] Added Children comment to epic
- [ ] Started epic with `bd update <epic> --status in_progress`
- [ ] Verified with `bd show <epic-id>`
- [ ] Synced with `bd sync`
- [ ] Output summary with autopilot handoff
- [ ] Wrote plan to `~/gt/.agents/$RIG/plans/`

---

## Quick Example

**User**: "/plan Add rate limiting to the API gateway"

**Agent workflow**:

```bash
# Phase 0: Context Discovery
mcp__smart-connections-work__lookup --query="rate limiting API gateway" --limit=10
cat docs/code-map/README.md | grep -i "gateway\|rate"

# Phase 0: Rig Detection (files in athena → RIG=athena)
RIG="athena"
mkdir -p ~/gt/.agents/$RIG/plans/

# Phase 1.5: Prior Plan Check
grep -l "rate" ~/gt/.agents/$RIG/plans/*.md
bd list --type=epic | grep -i rate
# No existing work found

# Phase 1: Research
# Found: services/gateway/middleware.py has existing auth middleware
# Pattern to follow: middleware chain in gateway

# Phase 2: Decomposition
# Feature 1: Rate limit middleware (P1, no deps)
# Feature 2: Rate limit config in Helm (P1, no deps, different files)
# Feature 3: Rate limit tests (P2, depends on F1)

# Phase 3: Create Issues
bd create "Epic: API Gateway Rate Limiting" --type epic --priority P1
# -> ai-platform-200

bd create "Add rate limit middleware" --type feature --priority P1
# -> ai-platform-201
bd comment ai-platform-201 "Files affected: services/gateway/middleware.py"

bd create "Add rate limit Helm config" --type feature --priority P1
# -> ai-platform-202
bd comment ai-platform-202 "Files affected: charts/ai-platform/values.yaml"

bd create "Rate limit integration tests" --type feature --priority P2
# -> ai-platform-203
bd dep add ai-platform-203 ai-platform-201

bd comments add ai-platform-200 "Children: ai-platform-201, ai-platform-202, ai-platform-203"
bd update ai-platform-200 --status in_progress

# Phase 5: Write plan to ~/gt/.agents/athena/plans/2026-01-03-rate-limiting.md

# Phase 5: Output summary with autopilot handoff
```

**Result**: 3 features, Wave 1 (201, 202 parallel), Wave 2 (203).

For more examples, see `references/examples.md`.

---

## References

### JIT-Loadable Documentation

| Topic | Reference |
|-------|-----------|
| Full templates | `references/templates.md` |
| Detailed examples | `references/examples.md` |
| Phase naming | `.claude/includes/phase-naming.md` |
| Beads workflows | `~/.claude/skills/beads/SKILL.md` |
| Decomposition patterns | `~/.claude/patterns/commands/plan/decomposition.md` |

### Essential Commands

| Command | Purpose |
|---------|---------|
| `bd create "Title" --type epic` | Create epic container |
| `bd create "Feature" --type feature` | Create feature issue |
| `bd dep add A B` | A waits for B |
| `bd comments add <id> "Children: ..."` | Track children on epic |
| `bd update <id> --status in_progress` | Start work |
| `bd show <id>` | View issue details |
| `bd ready` | Show unblocked issues |
| `bd sync` | Sync to git |

### Related Skills

- **beads**: Detailed issue management
- **research**: When goal needs exploration before planning
- **implement**: When executing a single issue from the plan

---

**Progressive Disclosure**: This skill provides core planning workflows. For detailed templates see `references/templates.md`, for examples see `references/examples.md`.
