---
name: meta
description: >
  Use when: "context", "session", "multi-agent", "workflow", "memory", "persistence",
  "retrospective", "retro", "patterns", "learnings", "stigmergy", "autonomous",
  "coordination", "observer", "blackboard".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Meta Skill

Context management, memory, retrospectives, and multi-agent coordination patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Context Manager** | Session coordination, multi-agent | Long workflows |
| **Change Executor** | Mechanical implementation | Executing plans |
| **Autonomous Worker** | Stigmergy, independent work | Parallel tasks |
| **Meta Observer** | Session monitoring, synthesis | Coordination |
| **Memory Manager** | Persistence, retrieval | Cross-session |
| **Retro Analyzer** | Pattern extraction, insights | Learning |

---

## Context Management

### Session Context Pattern

```markdown
## Session Context

### Active Task
- **Goal**: [Current objective]
- **Phase**: [Research | Plan | Implement | Review]
- **Progress**: [What's done, what's next]

### Key Decisions
1. [Decision 1] - [Rationale]
2. [Decision 2] - [Rationale]

### Blockers
- [Blocker 1] - [Status]

### Files Modified
- [file1.py] - [Change summary]
- [file2.ts] - [Change summary]

### Next Steps
1. [Step 1]
2. [Step 2]
```

### Context Preservation

For long-running tasks or compaction recovery:

```markdown
## Context Recovery

### What Was Happening
[Brief description of the work in progress]

### Current State
- Branch: [branch name]
- Last commit: [hash] - [message]
- Uncommitted changes: [list files]

### Continuation Instructions
[What to do next to resume work]
```

---

## Change Execution

### Mechanical Implementation

When executing a well-defined plan:

1. **Read the plan** - Understand exact changes
2. **Execute sequentially** - One change at a time
3. **Validate each step** - Run tests after changes
4. **Document progress** - Update tracking

### Execution Log

```markdown
## Execution Log

### Plan: [Plan Name]

| Step | File | Change | Status |
|------|------|--------|--------|
| 1 | src/api.py:45 | Add endpoint | ✅ |
| 2 | tests/test_api.py | Add tests | ✅ |
| 3 | README.md | Update docs | ⏳ |

### Issues Encountered
- [Issue 1]: [Resolution]

### Deviation from Plan
- [What changed and why]
```

---

## Autonomous Work (Stigmergy)

### Stigmergic Coordination

Workers coordinate through shared artifacts (blackboard pattern):

```
.agents/
├── blackboard/           # Shared coordination
│   ├── active-tasks.md   # Who's doing what
│   ├── completed.md      # What's done
│   └── blockers.md       # What's blocked
└── research/             # Shared knowledge
```

### Blackboard Protocol

```markdown
## Active Tasks

| Worker | Task | Started | Status |
|--------|------|---------|--------|
| Worker-A | Implement auth | 14:00 | In progress |
| Worker-B | Write tests | 14:15 | In progress |

## Completed
- [14:30] Worker-A: Auth middleware done
- [14:45] Worker-B: Auth tests passing

## Blockers
- [14:20] Worker-C: Needs database schema from Worker-A
```

### Independent Work Pattern

```markdown
## Autonomous Worker Session

### Assignment
[Task description from orchestrator]

### Context Loaded
- [Relevant file 1]
- [Relevant file 2]

### Work Log
1. [Time]: [Action taken]
2. [Time]: [Action taken]

### Output
[Deliverable or update to blackboard]

### Handoff
[What next worker needs to know]
```

---

## Session Observation

### Observer Pattern

For monitoring multiple parallel sessions:

```markdown
## Observer Report

### Sessions Active: 3

| Session | Worker | Task | Progress |
|---------|--------|------|----------|
| 001 | Worker-A | Auth | 75% |
| 002 | Worker-B | Tests | 50% |
| 003 | Worker-C | Docs | 25% |

### Synthesis
- [Pattern observed across sessions]
- [Coordination needed]

### Interventions
- [Action taken to help Worker-C]

### Artifacts Produced
- [List of outputs from sessions]
```

### Session Synthesis

```markdown
## Synthesis: [Topic]

### Sources
- Session 001 findings
- Session 002 findings
- Session 003 findings

### Common Patterns
1. [Pattern found in multiple sessions]
2. [Pattern found in multiple sessions]

### Conflicts
- [Where sessions disagree]

### Recommended Action
[Based on synthesized findings]
```

---

## Memory Management

### Persistent Memory Structure

```
.agents/
├── research/         # Research bundles (dated)
├── plans/            # Implementation plans
├── patterns/         # Extracted patterns
├── learnings/        # Session learnings
├── retros/           # Retrospectives
└── archive/          # Old/completed items
```

### Memory Operations

**Store:**
```markdown
## Memory: [Topic]
**Date**: YYYY-MM-DD
**Type**: [research | pattern | learning | decision]

### Content
[The information to remember]

### Context
[When/why this is relevant]

### Related
- [Link to related memory]
```

**Retrieve:**
```bash
# Find memories by topic
grep -r "authentication" .agents/

# Find recent memories
find .agents/ -mtime -7 -name "*.md"

# Find patterns
ls .agents/patterns/
```

### Memory Lifecycle

| Stage | Location | Action |
|-------|----------|--------|
| Active | .agents/research/ | Current work |
| Learned | .agents/patterns/ | Extracted pattern |
| Archived | .agents/archive/ | Old but retrievable |

---

## Retrospective Analysis

### Retro Template

```markdown
# Retrospective: [Session/Project]

**Date**: YYYY-MM-DD
**Duration**: [Time spent]
**Outcome**: [Success | Partial | Failed]

## Summary
[What was accomplished]

## What Went Well
- [Positive 1]
- [Positive 2]

## What Could Improve
- [Improvement 1]
- [Improvement 2]

## Patterns Identified

### Pattern: [Name]
**Context**: [When this applies]
**Problem**: [What it solves]
**Solution**: [How to apply]
**Evidence**: [Where it worked]

## Action Items
- [ ] [Action for next time]
- [ ] [Process improvement]

## Learnings
1. [Key learning 1]
2. [Key learning 2]
```

### Pattern Extraction

```markdown
## Pattern: [Name]

### Context
[Situations where this pattern applies]

### Problem
[What problem this solves]

### Solution
[How to implement the pattern]

### Example
[Concrete example from codebase]

### Related Patterns
- [Related pattern 1]
- [Related pattern 2]

### Source
[Session/project where discovered]
```

---

## Workflow Coordination

### Multi-Phase Workflow

```markdown
## Workflow: [Name]

### Phase 1: Research
- **Input**: [Requirements]
- **Output**: Research bundle
- **Gate**: Assumptions validated

### Phase 2: Plan
- **Input**: Research bundle
- **Output**: Implementation plan
- **Gate**: Plan approved

### Phase 3: Implement
- **Input**: Approved plan
- **Output**: Working code
- **Gate**: Tests passing

### Phase 4: Review
- **Input**: Implementation
- **Output**: Merged code
- **Gate**: Review approved
```

### Handoff Protocol

```markdown
## Handoff: [From Phase] → [To Phase]

### Deliverable
[What's being handed off]

### Key Decisions Made
1. [Decision 1]
2. [Decision 2]

### Assumptions
- [Assumption 1] - [Validated: Y/N]

### Open Questions
- [Question 1]

### Recommended Next Steps
1. [Step 1]
2. [Step 2]
```

---

## Session Close Protocol

Before ending a session:

```markdown
## Session Close Checklist

- [ ] git status (check changes)
- [ ] git add (stage changes)
- [ ] bd sync (sync beads)
- [ ] git commit (commit code)
- [ ] git push (push to remote)
- [ ] Update .agents/ if needed
- [ ] Document handoff for next session
```
