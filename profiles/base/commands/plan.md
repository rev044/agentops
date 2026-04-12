---
description: Planning phase - specify exact changes with file:line precision
---

# /plan - Phase 2: Specify Exact Changes

**Purpose:** Create detailed implementation specification from research findings.

**Philosophy:** Planning IS the work. Execution becomes trivial when planning is thorough.

**Token budget:** 40-60k tokens (20-30% of context window)

---

## When to Use

Use `/plan` when you:
- Have completed research phase
- Understand the problem and constraints
- Need to specify exact implementation steps
- Want human review before executing changes

**Don't use if:**
- Haven't done research (do `/research` first)
- Plan already exists (use `/implement`)
- Changes are trivial (just implement directly)

---

## Planning Process

### Step 1: Load Research (if available)

```bash
/bundle-load research-[topic]
```

### Step 2: Specify ALL Changes

**Every change must have:**
- Exact file path
- Line number (if modifying existing)
- Specific change description
- Validation command

**Example:**
```
File: apps/redis/kustomization.yaml:18
Change: Add redis service reference
Validation: make quick
```

### Step 3: Define Test Strategy

**What tests prove it works?**
- Unit tests
- Integration tests
- Validation commands
- Expected outcomes

### Step 4: Document Rollback

**How to undo if needed:**
- Rollback commands
- Restoration procedure
- Verification steps

### Step 5: Get Human Approval

**Before implementing:**
- Review plan with user
- Confirm approach
- Adjust if needed

### Step 6: Save Plan Bundle

```bash
/bundle-save plan-[topic]
```

---

## Plan Document Structure

```markdown
# Implementation Plan: [Topic Name]

## Summary
[What will be implemented and why]

## Changes Specified

### 1. Create [file-path]
**Purpose:** [Why this file?]
**Content:** [What goes in it?]
**Validation:** `make quick`

### 2. Modify [file-path:line]
**Current:** [What exists now]
**Change:** [What to change]
**Validation:** `make test`

### 3. Delete [file-path]
**Reason:** [Why removing?]
**Dependencies:** [Check nothing uses it]
**Validation:** `make ci-all`

## Test Strategy

### Validation Tests
- [ ] YAML syntax: `make quick`
- [ ] Full CI: `make ci-all`
- [ ] Deployment test: `kubectl apply --dry-run`

### Functional Tests
- [ ] Feature works as expected
- [ ] No regressions
- [ ] Performance acceptable

## Rollback Procedure

If implementation fails:
1. `git revert [commit-sha]`
2. Verify rollback: `make validate`
3. Return to planning phase
4. Adjust plan and retry

## Approval

- [ ] Approach validated
- [ ] Changes reviewed
- [ ] Risks acceptable
- [ ] Ready to implement

**Approved by:** [User name]
**Date:** [Date]

## Next Steps

1. Start fresh session
2. Load this plan: `/bundle-load plan-[topic]`
3. Execute: `/implement`
```

---

## Token Budget Management

```
Planning Phase: 40-60k tokens (20-30%)

Breakdown:
- Load research: 5-10k
- Specify changes: 20-30k
- Define tests: 5-10k
- Documentation: 5-10k
- Reserve: 5-10k

Monitor: Stay under 40% total
```

---

## Critical: Precision Required

**Every change must be specific:**

❌ **BAD:** "Update the configuration"
✅ **GOOD:** "Edit config.yaml:23 - Change timeout from 30s to 60s"

❌ **BAD:** "Add redis support"
✅ **GOOD:** "Create apps/redis/kustomization.yaml with redis:7.0 image"

❌ **BAD:** "Fix the bug"
✅ **GOOD:** "Modify api/handler.go:45 - Add nil check before accessing user.Email"

---

## Transition to Implementation

After plan approved:

```bash
# Save plan
/bundle-save plan-[topic]

# Start new session for implementation
# Load plan bundle
/bundle-load plan-[topic]

# Execute plan
/implement
```

---

## Success Criteria

Plan is complete when:

- [ ] Every change specified (file:line)
- [ ] Test strategy defined
- [ ] Rollback procedure documented
- [ ] Human approval received
- [ ] Plan bundle saved
- [ ] Ready for implementation phase

---

## Common Mistakes

**Vague specifications:**
- "Update the code" - WHERE in the code?
- "Add feature" - WHICH files changed?
- "Fix issue" - WHAT specific change?

**Missing validation:**
- No test commands
- No success criteria
- No verification steps

**No rollback:**
- Can't undo safely
- No recovery procedure
- High risk

---

**Next command:** `/implement` to execute the approved plan
