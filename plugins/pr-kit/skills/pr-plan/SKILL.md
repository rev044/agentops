---
name: pr-plan
description: >
  Strategic contribution planning for open source PRs. Takes pr-research output
  and produces plan artifact with scope, acceptance criteria, and risk assessment.
  Triggers: "pr plan", "contribution plan", "plan PR", "plan contribution".
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Bash,Grep,Glob,Task"
skills:
  - pr-research
---

# PR Plan Skill

Strategic planning for open source contributions. Produces structured plans with
scope definition, acceptance criteria, and risk assessment.

## Overview

Create a contribution plan that bridges research and implementation. This skill
takes pr-research output and produces an actionable plan with clear scope,
success criteria, and risk mitigation strategies.

**Output:** `~/gt/.agents/<rig>/plans/YYYY-MM-DD-pr-plan-{repo-slug}.md`

**When to Use**:
- After completing /pr-research
- Planning contribution strategy for external repos
- Evaluating contribution feasibility
- Before starting implementation

**When NOT to Use**:
- Haven't researched the repo yet (run /pr-research first)
- Trivial contributions (fix typos, obvious bugs)
- Internal project planning (use /plan or /formulate)

---

## Workflow

```
0.  Input Discovery     -> Find/load pr-research artifact
1.  Rig Detection       -> Determine output location
2.  Scope Definition    -> What exactly to contribute
3.  Target Selection    -> Which issues/areas to address
4.  Criteria Definition -> Acceptance criteria from research
5.  Risk Assessment     -> What could go wrong
6.  Strategy Formation  -> Implementation approach
7.  Output              -> Write plan artifact
8.  Confirm             -> Verify and next steps
```

---

## Phase 0: Input Discovery

**Identify the pr-research artifact to base the plan on.**

### 0.1 If Path Provided

```bash
# Verify research artifact exists
ls -la "$INPUT"
cat "$INPUT" | head -50
```

### 0.2 If Topic/Description Provided

```bash
# Search for existing pr-research artifacts
ls ~/gt/.agents/*/research/ 2>/dev/null | xargs grep -l "<topic>" 2>/dev/null
ls ~/gt/.agents/_oss/research/ 2>/dev/null | grep -i "<keywords>"

# If not found, suggest running /pr-research first
```

### 0.3 Research Artifact Validation

| Check | Action |
|-------|--------|
| File exists | Proceed |
| File is pr-research format | Extract key sections |
| File missing | Suggest: `/pr-research <repo>` first |
| File is other research type | Adapt or suggest pr-research |

---

## Phase 1: Rig Detection

**CRITICAL**: All `.agents/` artifacts go to `~/gt/.agents/<rig>/` based on context.

**Detection Logic**:
1. If research artifact path contains rig name, use that rig
2. If contributing FROM a rig's context, use that rig
3. For general OSS work, use `_oss` subfolder
4. If unclear, ask user

```bash
# Set RIG variable for output paths
RIG="_oss"  # Default for external contributions
mkdir -p ~/gt/.agents/$RIG/plans/
```

---

## Phase 2: Scope Definition

**Define the exact contribution.**

### 2.1 Extract from Research

Pull from the pr-research artifact:
- Repository overview
- Contribution opportunities identified
- Avoided areas noted

### 2.2 Scope Questions

| Question | Why It Matters |
|----------|----------------|
| What specific functionality? | Clear deliverable |
| Which files/packages? | Limits impact surface |
| What's explicitly out of scope? | Prevents scope creep |
| Single PR or series? | Sets expectations |

### 2.3 Scope Statement Template

```markdown
## Scope

**Contribution**: [1-2 sentences describing the change]

**Affected Areas**:
- `path/to/file.go` - [what changes]
- `path/to/other.go` - [what changes]

**Out of Scope**:
- [Related but excluded work]
- [Future enhancements]
- [Things that might seem related but aren't]
```

---

## Phase 3: Target Selection

**Choose specific issues or areas to address.**

### 3.1 Issue Selection Criteria

From pr-research, evaluate issues by:

| Factor | Weight | How to Assess |
|--------|--------|---------------|
| Alignment with scope | High | Does it match your contribution goal? |
| Difficulty level | High | Match your familiarity with codebase |
| Activity | Medium | Recent comments? Active discussion? |
| Assignee status | Medium | Unassigned = available |
| Label signals | Medium | "good first issue", "help wanted" |

### 3.2 Selection Process

```bash
# Review issues from research
grep -A 20 "## Contribution Opportunities" "$RESEARCH_FILE"

# Verify issue is still open
gh issue view <number> --json state,assignees,labels

# Check for recent activity
gh issue view <number> --json comments --jq '.comments | length'
```

### 3.3 Target Documentation

```markdown
## Target

**Primary Issue**: #N - [title]
- Status: Open, unassigned
- Labels: [relevant labels]
- Activity: Last comment X days ago

**Why This Issue**:
- [Alignment with scope]
- [Appropriate difficulty]
- [Good fit for contribution]

**Alternative Targets** (if primary blocked):
- #M - [brief description]
```

---

## Phase 4: Acceptance Criteria

**Define success from maintainer perspective.**

### 4.1 Extract Maintainer Expectations

From pr-research, pull:
- PR patterns (size, style, testing requirements)
- Review process requirements
- CI/CD requirements
- Documentation requirements

### 4.2 Acceptance Criteria Template

```markdown
## Acceptance Criteria

### Code Quality
- [ ] Follows project coding style (from CONTRIBUTING.md)
- [ ] Passes all existing tests
- [ ] Adds tests for new functionality
- [ ] No new linting warnings

### PR Requirements
- [ ] Title follows convention: `type(scope): description`
- [ ] Body uses project template
- [ ] Size within acceptable range (< X files, < Y lines)
- [ ] Single logical change (no scope creep)

### Review Process
- [ ] CI passes all checks
- [ ] Documentation updated (if required)
- [ ] Responds to review feedback promptly

### Project-Specific
- [ ] [Any project-specific requirements from research]
```

---

## Phase 5: Risk Assessment

**Identify what could go wrong and how to mitigate.**

### 5.1 Risk Categories

| Category | Examples |
|----------|----------|
| **Technical** | Breaking changes, API compatibility |
| **Process** | Long review times, maintainer bandwidth |
| **Scope** | Feature creep, hidden complexity |
| **External** | Competing PRs, architecture changes |

### 5.2 Risk Matrix Template

```markdown
## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| PR review takes > 2 weeks | Medium | Medium | Start small, be responsive |
| Scope expands during review | Medium | High | Define scope clearly upfront |
| Breaking change discovered | Low | High | Test against multiple versions |
| Competing PR submitted | Low | Medium | Check open PRs before starting |
| Architecture change in progress | Low | High | Monitor repo activity |

### Risk Mitigation Strategy

1. **Start small**: First PR should be minimal viable contribution
2. **Communicate early**: Comment on issue before implementing
3. **Stay responsive**: Address review feedback quickly
4. **Monitor activity**: Watch for conflicting changes
```

---

## Phase 6: Implementation Strategy

**Plan the approach to implementation.**

### 6.1 Strategy Components

| Component | Description |
|-----------|-------------|
| **Approach** | How to implement the change |
| **Order** | What to do first, second, etc. |
| **Validation** | How to verify correctness |
| **Timing** | When to submit PR |

### 6.2 Strategy Template

```markdown
## Implementation Strategy

### Approach

1. **Setup**: Fork repo, configure dev environment
2. **Understand**: Read existing code in affected area
3. **Implement**: Make changes following project patterns
4. **Test**: Run existing tests + add new tests
5. **Document**: Update any affected documentation
6. **Submit**: Create PR following project conventions

### Pre-Implementation Checklist

- [ ] Fork created and up-to-date with upstream
- [ ] Dev environment working (build, test pass)
- [ ] Issue claimed or comment posted
- [ ] Recent repo activity reviewed (no conflicts)

### Estimated Complexity

| Aspect | Estimate |
|--------|----------|
| Files changed | N |
| Lines of code | ~X |
| Test coverage needed | Y% |
| Difficulty | Easy/Medium/Hard |

### Dependencies

- [ ] [Any prerequisites]
- [ ] [Required knowledge]
- [ ] [External dependencies]
```

---

## Phase 7: Output

Write to `~/gt/.agents/$RIG/plans/YYYY-MM-DD-pr-plan-{repo-slug}.md`

### Output Template

```markdown
---
date: YYYY-MM-DD
type: PR-Plan
upstream: owner/repo
research: path/to/pr-research-artifact.md
tags: [plan, oss, contribution]
status: READY
---

# PR Plan: {repo-name}

## Executive Summary

{2-3 sentences: what you're contributing, why, expected outcome}

## Research Reference

**Source**: `{path to pr-research artifact}`
**Repo**: {owner/repo}
**URL**: {https://github.com/owner/repo}

## Scope

**Contribution**: {description of change}

**Affected Areas**:
- `path/to/file` - {description}

**Out of Scope**:
- {explicitly excluded items}

## Target

**Primary Issue**: #{N} - {title}
- Status: {Open/In Progress}
- Labels: {labels}
- Rationale: {why this issue}

**Alternative Targets**:
- #{M} - {backup option}

## Acceptance Criteria

### Code Quality
- [ ] Follows project coding style
- [ ] Passes existing tests
- [ ] Adds appropriate tests
- [ ] No linting warnings

### PR Requirements
- [ ] Title: `type(scope): description`
- [ ] Body uses template
- [ ] Size: < X files
- [ ] Single logical change

### Project-Specific
- [ ] {project requirements}

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| {risk} | {L/M/H} | {L/M/H} | {strategy} |

## Implementation Strategy

### Approach

1. {step 1}
2. {step 2}
3. {step 3}

### Pre-Implementation Checklist

- [ ] Fork ready
- [ ] Dev environment working
- [ ] Issue claimed
- [ ] No conflicting PRs

### Estimated Complexity

| Aspect | Estimate |
|--------|----------|
| Files | N |
| Lines | ~X |
| Difficulty | Easy/Medium/Hard |

## Next Steps

1. Claim/comment on target issue
2. Fork and set up development environment
3. Implement following strategy above
4. Run `/pr-prep` when ready to submit

---

**Ready to implement? Start with the pre-implementation checklist above.**
```

---

## Phase 8: Confirm

```bash
ls -la ~/gt/.agents/$RIG/plans/
```

Tell user:
```
PR Plan output: ~/gt/.agents/$RIG/plans/YYYY-MM-DD-pr-plan-{repo}.md

Next steps:
1. Review the plan
2. Claim/comment on target issue in upstream repo
3. Fork and implement
4. When ready: /pr-prep
```

---

## Reviewing Existing PR Stacks

When analyzing existing PRs (instead of planning new ones), provide **analysis, not approvals**.

### Our Role

| WE DO | WE DON'T |
|-------|----------|
| Analyze technical details | Approve or reject PRs |
| Identify merge order/dependencies | Say "LGTM" |
| Flag concerns and questions | Make merge decisions |
| Help maintainer review | Speak for maintainer |

### PR Stack Analysis Structure

```markdown
## Analysis: [Stack Name]

### PR Dependencies
[Which PRs depend on others]

### Merge Order Recommendation
1. #N - [why first]
2. #M - [why second]

### Conflict Detection
- #X and #Y both modify `file.go` - coordinate merge

### Questions for Each PR
[Technical questions, not verdicts]
```

### Comment Guidelines

**DON'T write approval language:**
```markdown
✅ LGTM
The fix is correct. ✅ LGTM
Ready to merge.
```

**DO write analysis:**
```markdown
## Analysis Notes

This addresses [problem] with [approach].

**Open questions:**
1. [Question]
2. [Question]

**Merge order**: After #N because [reason].
```

See: `pr-plan/lessons/`

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Plan without research | Run /pr-research first |
| Vague scope | Define specific files/changes |
| Skip risk assessment | Identify risks upfront |
| Ignore maintainer expectations | Base criteria on research |
| Plan large PRs | Start small, iterate |
| Skip issue claim | Comment before implementing |
| Use approval language when reviewing | Provide analysis, let maintainer decide |

---

## Workflow Integration

```
/pr-research <repo> -> /pr-plan <research> -> implement -> /pr-prep
         ↓                    ↓                   ↓           ↓
    Understand            Plan with          Code it      Submit
    the project           clear scope                     safely
```

---

## Quick Example

**User**: `/pr-plan ~/gt/.agents/_oss/research/2026-01-12-pr-gastown.md`

**Agent workflow**:

```bash
# Phase 0: Input Discovery
cat ~/gt/.agents/_oss/research/2026-01-12-pr-gastown.md | head -50
# Found: gastown research with contribution opportunities

# Phase 1: Rig Detection
RIG="_oss"
mkdir -p ~/gt/.agents/$RIG/plans/

# Phase 2: Scope Definition
# Contribution: Add formula validation to suggest.go
# Affected: internal/suggest/suggest.go
# Out of scope: Other packages, refinery, convoy

# Phase 3: Target Selection
# Primary: Issue #150 - formula validation missing
# Alternative: Issue #152 - better error messages

# Phase 4: Acceptance Criteria
# From research: conventional commits, < 5 files, tests required

# Phase 5: Risk Assessment
# Main risk: Review latency (mitigate: small PR, responsive)

# Phase 6: Implementation Strategy
# 1. Fork, 2. Read suggest.go, 3. Add validation, 4. Add tests, 5. PR

# Phase 7: Output
# Write ~/gt/.agents/_oss/plans/2026-01-12-pr-plan-gastown.md

# Phase 8: Confirm
```

**Result**: Plan ready for implementation.

---

## References

- **PR Research**: `pr-research/SKILL.md`
- **PR Preparation**: `pr-prep/SKILL.md`
- **Internal Planning**: `plan/SKILL.md`
