---
name: bug-hunt
description: >
  Investigate suspected bugs with git archaeology and root cause analysis.
  Triggers: "bug", "broken", "doesn't work", "failing", "investigate bug".
version: 1.0.0
context: fork
author: "boshu2"
license: "MIT"
allowed-tools: "Read,Write,Bash,Grep,Glob,Task,WebFetch"
skills:
  - standards
---

# Bug Hunt Skill

Systematic bug investigation that produces root cause analysis and fix design.

## Overview

When a bug is suspected, this skill guides through:
1. Confirming the bug exists
2. Tracing when/how it was introduced
3. Understanding the root cause
4. Designing a complete fix

**Output**: `~/gt/.agents/<rig>/research/YYYY-MM-DD-bug-{slug}.md`

---

## Workflow

```
0. Rig Detection -> Determine target rig from code paths
1. Reproduce     -> Confirm bug exists, capture symptoms
2. Archaeology   -> Git blame/log/bisect to find introduction
3. Root Cause    -> Trace the failure cascade
4. Fix Design    -> Solution + test plan
5. Output        -> Structured document for PR/issue
```

---

## Phase 0: Rig Detection

**CRITICAL**: All `.agents/` artifacts go to `~/gt/.agents/<rig>/` based on the codebase with the bug.

**Detection Logic**:
1. Identify which rig's code has the bug (e.g., files in `~/gt/gastown/` → `gastown`)
2. If bug spans multiple rigs, use `_cross-rig`
3. If unknown/unclear, ask user

| Bug Location | Target Rig | Output Base |
|--------------|------------|-------------|
| `~/gt/athena/**` | `athena` | `~/gt/.agents/athena/` |
| `~/gt/daedalus/**` | `daedalus` | `~/gt/.agents/daedalus/` |
| `~/gt/chronicle/**` | `chronicle` | `~/gt/.agents/chronicle/` |
| Multiple rigs | `_cross-rig` | `~/gt/.agents/_cross-rig/` |

```bash
# Set RIG variable for use in output paths
RIG="daedalus"  # or athena, chronicle, _cross-rig
mkdir -p ~/gt/.agents/$RIG/research/
```

---

## Phase 1: Reproduce

**Goal**: Confirm the bug is real and capture exact symptoms.

```bash
# Run the failing command/operation
# Capture exact error message
# Note environment (version, config, etc.)
```

**Checklist**:
- [ ] Bug reproduces consistently?
- [ ] Exact error/symptom captured?
- [ ] Environment noted (versions, config)?
- [ ] Minimal reproduction steps?

**If can't reproduce**: Stop - may be config issue or already fixed.

---

## Phase 2: Git Archaeology

**Goal**: Find when the bug was introduced.

### Key Commands

```bash
# Find commits touching relevant file
git log --oneline -20 -- <file>

# Find when a specific pattern was introduced
git log -S "problematic_code" --oneline

# Blame specific lines
git blame -L <start>,<end> <file>

# Check if bug exists at specific commit
git show <commit>:<file> | grep -A5 "pattern"

# Find merge that introduced it
git log --merges --oneline -- <file>

# Binary search for introducing commit (powerful for complex bugs)
git bisect start
git bisect bad HEAD
git bisect good <known-good-commit>
# Then test each suggested commit until found
git bisect reset
```

### Trace Pattern

```
Current (broken) → Prior commit → Prior commit → ... → Last working
```

Document:
- **Introducing commit**: `<hash>` - `<message>`
- **Introducing PR**: #<number> (if applicable)
- **Author**: Who introduced it
- **Date**: When
- **Intent**: What they were trying to do

---

## Phase 3: Root Cause Analysis

**Goal**: Understand WHY the bug exists, not just WHERE.

### The Cascade Pattern

Most bugs are cascades of well-intentioned changes:

```
Original working code
  ↓
Change A (fixed problem X, but...)
  ↓
Change B (partial fix for A's side effect)
  ↓
Current bug (B didn't fully address A's problem)
```

### Document the Cascade

| Commit | Intent | Side Effect |
|--------|--------|-------------|
| `abc123` | Fix daemon timing | Broke routing |
| `def456` | Partial routing fix | Still wrong dir |
| Current | - | Full failure |

### Key Questions

1. What was the original behavior?
2. What change broke it?
3. Why did that change seem correct at the time?
4. What did the author miss?
5. Were there warning signs (comments, tests)?

---

## Phase 4: Fix Design

**Goal**: Design a complete fix that won't create new problems.

### Fix Checklist

- [ ] Addresses root cause (not just symptom)?
- [ ] Doesn't reintroduce original problem?
- [ ] Follows existing patterns in codebase?
- [ ] Has test coverage for the fix?
- [ ] Has test coverage for regression?

### Test Plan

| Test Type | Coverage |
|-----------|----------|
| Unit | New helper functions |
| Integration | Cross-component interaction |
| E2E | Full workflow validation |
| Regression | Original bug doesn't recur |

### Code Pattern

```
// Before (broken):
<broken code with comment explaining why>

// After (fixed):
<fixed code with comment explaining the fix>
// See: <issue URL>
```

---

## Phase 5: Output

Write to `~/gt/.agents/<rig>/research/YYYY-MM-DD-bug-{slug}.md`

### Required Sections

```markdown
# Bug: {Title}

## Symptom
What fails, exact error, reproduction steps.

## Timeline
| Date | Event | Commit/PR |
|------|-------|-----------|

## Root Cause Analysis
### The Cascade
Trace of commits that led to bug.

### Key Insight
One sentence explaining the fundamental mistake.

## Fix Design
### Approach
What the fix does.

### Code Changes
Files and patterns.

### Test Coverage
What tests validate the fix.

## References
- Issue: <URL>
- Introducing commit: <hash>
- Fix commit: <hash> (after implementation)
```

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Fix symptom only | Trace to root cause |
| Skip git history | Always find introducing commit |
| Blame the author | Understand their intent |
| No tests | Test the fix AND regression |
| Vague "it's broken" | Exact symptoms + repro steps |

---

## Example: PR #149 Workflow

This skill was derived from the cross-rig sling fix:

1. **Reproduce**: `gt sling ap-xxx ai-platform` failed from town root
2. **Archaeology**: Traced to PR #138's BEADS_DIR changes
3. **Root Cause**: Cascade of 3 commits, each partially fixing prior
4. **Fix**: New `ResolveHookDir()` with prefix-based routing
5. **Output**: Issue #148, PR #149 with 19 test cases

See: `.agents/research/2026-01-06-cross-rig-sling-fix-workflow.md`

---

## Workflow Integration

```
/bug-hunt -> /plan (if complex) -> implement -> /retro
```

For simple bugs, skip `/plan` and implement directly after `/bug-hunt`.

---

## Standards Loading

When designing fixes (Phase 4), load relevant standards for consistent code:

| File Pattern | Load Reference |
|--------------|----------------|
| `*.py` | `domain-kit/skills/standards/references/python.md` |
| `*.go` | `domain-kit/skills/standards/references/go.md` |
| `*.ts`, `*.tsx` | `domain-kit/skills/standards/references/typescript.md` |
| `*.sh` | `domain-kit/skills/standards/references/shell.md` |
