# Plan Templates Reference

Detailed templates for plan documents and summaries.

---

## Plan File Template

**Location:** `.agents/plans/YYYY-MM-DD-{goal-slug}.md`

### Tag Vocabulary (REQUIRED)

Document type tag: `plan` (required first)

**Examples:**
- `[plan, agents, kagent]` - KAgent implementation plan
- `[plan, data, neo4j]` - GraphRAG implementation plan
- `[plan, auth, security]` - OAuth2 implementation plan
- `[plan, ci-cd, tekton]` - Tekton pipeline plan

### Full Template

```markdown
---
date: YYYY-MM-DD
type: Plan
goal: "[Goal description]"
tags: [plan, domain-tag, optional-tech-tag]
epic: "[beads epic ID]"
status: ACTIVE
---

# Plan: [Goal]

## Overview
[2-3 sentence summary of approach]

## Research Reference
[Link to research file if /research was run first, or "Inline research" if done in this command]

## Features (Dependency Order)

| ID | Feature | Priority | Depends On | Status |
|----|---------|----------|------------|--------|
| bd-xxx | Feature 1 | P1 | - | open |
| bd-yyy | Feature 2 | P2 | bd-xxx | open |

## Dependency Graph

```
Wave 1 (No Dependencies):
  bd-xxx: Feature 1
  bd-yyy: Feature 2
       |
       v unblocks
Wave 2 (Depends on Wave 1):
  bd-zzz: Feature 3
```

## Wave Execution Order

| Wave | Features | Can Parallel | Notes |
|------|----------|--------------|-------|
| 1 | Feature 1, Feature 2 | Yes | No dependencies, different files |
| 2 | Feature 3 | No | Depends on Feature 1 |
| 3 | Feature 4 | No | Depends on Feature 2 |

**Wave Computation Rules:**
- **Wave 1:** All features with no dependencies (blockers)
- **Wave N:** Features where all dependencies are in Wave N-1 or earlier
- **Can Parallel:** "Yes" if features in same wave affect different files, "No" if they share files

## Implementation Notes
[Key decisions, patterns to follow, risks identified]

## External Requirements
[Any prerequisites for autopilot: Langfuse enabled, secrets configured, migrations run, etc.]

## Next Steps
Run `/autopilot <epic-id> --dry-run` to validate, then `/autopilot <epic-id>` to execute.
```

---

## Plan Summary Template (Autopilot Handoff)

Output this after creating issues. This is the **handoff to autopilot**.

```markdown
---

# Plan Complete: [Goal Description]

**Epic:** `bd-epic-id`
**Plan:** `.agents/plans/YYYY-MM-DD-goal-slug.md`
**Issues:** N features across M waves

---

## Wave Execution Order

| Wave | Issues | Can Parallel | Ready Now |
|------|--------|--------------|-----------|
| 1 | bd-xxx, bd-yyy | Yes | Ready |
| 2 | bd-zzz | No | Blocked by Wave 1 |
| 3 | bd-aaa, bd-bbb | Yes | Blocked by Wave 2 |

## Features Created

| ID | Feature | Priority | Depends On |
|----|---------|----------|------------|
| bd-xxx | Setup auth middleware | P1 | - |
| bd-yyy | Add JWT validation | P2 | bd-xxx |
| bd-zzz | Add refresh token flow | P2 | bd-yyy |

## Dependency Graph

```
Wave 1 (No Dependencies):
  bd-xxx: Feature 1
  bd-yyy: Feature 2
       |
       v unblocks
Wave 2 (Depends on Wave 1):
  bd-zzz: Feature 3
       |
       v unblocks
Wave 3 (Depends on Wave 2):
  bd-aaa: Feature 4
  bd-bbb: Feature 5
```

---

## Ready for Autopilot

### Pre-Flight Checklist

- [x] Epic created with Children comment
- [x] All dependencies set via `bd dep add`
- [x] Files affected noted on each issue
- [ ] External requirements: [list any, e.g., "Langfuse enabled"]

### Execute

```bash
# Validate structure first
/autopilot bd-epic-id --dry-run

# Full execution
/autopilot bd-epic-id
```

### Alternative: Manual Execution

```bash
# Implement one at a time
bd ready
/implement bd-xxx
```
```

---

## Feature Template

For decomposition, use this template per feature:

```
Feature: [Short descriptive name]
Priority: P0|P1|P2|P3
Type: feature|bug|task|refactor
Depends On: [List of prerequisite features, if any]
Acceptance Criteria:
  - [ ] Criterion 1
  - [ ] Criterion 2
Test Strategy: [How to verify this works]
Files Affected: [List key files]
```
