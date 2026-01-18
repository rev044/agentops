---
name: product
description: >
  Create lightweight product brief (PR/FAQ) that articulates customer need, success criteria,
  and scope before engineering decomposition. Triggers: "product brief", "customer value",
  "who is this for", "PR/FAQ", "working backwards", "define success", "scope this".
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - research
---

# Product Skill

Create lightweight product briefs that bridge `/research` and `/formulate` with customer-first thinking.

## Overview

Product briefs force articulation of **why** before **how**. Based on Amazon's PR/FAQ
and Google's Design Doc patterns, this skill ensures the customer problem is understood
before decomposing into engineering tasks.

**Output:** `~/gt/.agents/<rig>/products/YYYY-MM-DD-{topic-slug}.md`

**When to Use**:
- Multi-day work (3+ days)
- User-facing impact isn't immediately obvious
- Multiple valid approaches exist
- Unclear "why should this be built?"

**When NOT to Use**:
- Pure technical debt (no user-facing impact)
- Bug fixes with clear scope
- PM already wrote a PRD
- Single-day tasks

---

## Workflow

```
0.  Rig Detection       -> Determine target rig
0.5 Setup               -> mkdir -p ~/gt/.agents/<rig>/products/
1.  Prior Art           -> Check for existing PRDs/product docs
2.  Customer Discovery  -> Who, pain points, workarounds
3.  Solution            -> Headline, problem mapping
4.  Success Criteria    -> Measurable outcomes
5.  Scope               -> In-scope, non-goals
6.  Output              -> Write product brief
7.  Confirm             -> Verify file, next steps
```

---

## Phase 0: Rig Detection

**CRITICAL**: All `.agents/` artifacts go to `~/gt/.agents/<rig>/` based on the primary codebase.

**Detection Logic**:
1. Identify which rig's code is involved (e.g., files in `~/gt/ai-platform/` → `ai-platform`)
2. If work spans multiple rigs, use `_cross-rig`
3. If unknown/unclear, ask user

| Files Being Read | Target Rig | Output Base |
|------------------|------------|-------------|
| `~/gt/athena/**` | `athena` | `~/gt/.agents/athena/` |
| `~/gt/cyclopes/**` | `cyclopes` | `~/gt/.agents/cyclopes/` |
| `~/gt/daedalus/**` | `daedalus` | `~/gt/.agents/daedalus/` |
| Multiple rigs | `_cross-rig` | `~/gt/.agents/_cross-rig/` |

```bash
# Set RIG variable for use in output paths
RIG="athena"  # or cyclopes, daedalus, _cross-rig
mkdir -p ~/gt/.agents/$RIG/products/
```

---

## Phase 1: Prior Art Discovery

**Check before creating new product briefs.**

```bash
# Existing product docs
ls -la ~/gt/.agents/$RIG/products/ 2>/dev/null | grep -i "<keywords>"
ls -la ~/gt/$RIG/docs/product/ 2>/dev/null

# Existing PRDs
find ~/gt/$RIG -name "*PRD*" -o -name "*prd*" 2>/dev/null | head -5

# Semantic search
mcp__smart-connections-work__lookup --query="$TOPIC product requirements customer" --limit=5
```

| Prior Work Status | Action |
|-------------------|--------|
| PRD exists | Reference it, don't duplicate |
| Product brief exists | Extend or supersede |
| None | Create new brief |

---

## Phase 2: Customer Discovery

**Identify the user and their pain.**

### Questions to Answer

1. **Who exactly is the customer?**
   - Not "users" — be specific: "API consumers on metered billing plans"
   - Name the persona if possible

2. **What are their pain points? (ranked)**
   - List 2-3 specific problems
   - Rank by severity/frequency

3. **How do they solve this today?**
   - Current workarounds
   - Why those workarounds are insufficient

### Discovery Methods

```bash
# Check existing user feedback
grep -ri "pain\|problem\|frustrat\|workaround" ~/gt/$RIG/docs/ 2>/dev/null | head -10

# Check issue tracker for user complaints
bd list --type=bug | head -10
```

**If these questions cannot be answered**, run `/research` first.

---

## Phase 3: Solution Articulation

### The Headline Test

**State the value proposition in 10 words or less.**

Bad: "We're adding rate limiting to the API gateway"
Good: "Predictable API costs with automatic overrun protection"

The headline forces thinking from the customer perspective.

### Problem→Solution Mapping

Every feature must trace back to a stated customer problem:

| Problem | Solution |
|---------|----------|
| Unpredictable bills | Usage limits with alerts |
| No visibility into consumption | Real-time dashboard |
| Sudden service cuts | Graceful degradation mode |

**If a feature doesn't map to a problem, question whether to build it.**

---

## Phase 4: Success Criteria

Define measurable outcomes:

| Metric | Target | How Measured |
|--------|--------|--------------|
| Adoption | 80% of API users enable limits | Analytics |
| Cost reduction | 30% fewer overrun charges | Billing data |
| Support tickets | 50% reduction in billing complaints | Support system |

**Good criteria are:**
- Measurable (not "users are happier")
- Time-bound (within 3 months of launch)
- Tied to stated problems

---

## Phase 5: Scope Definition

### In Scope

What are we definitely building?

- Be specific
- These become `/formulate` inputs

### Non-Goals (Critical)

**What are we explicitly NOT building?**

This prevents scope creep and clarifies boundaries:

- "Not building DDoS protection (separate initiative)"
- "Not supporting enterprise tier (future phase)"
- "Not changing pricing model (out of scope)"

**Every non-goal should feel like something someone might reasonably expect.**

---

## Phase 6: Output

Write to `~/gt/.agents/$RIG/products/YYYY-MM-DD-{topic-slug}.md`

Use template from `references/template.md`:

```markdown
# Product Brief: [Feature Name]

**Date:** YYYY-MM-DD
**Author:** [name]
**Status:** draft | review | approved

## 1. Headline
[10 words or less]

## 2. Customer & Problem
**Who:** [specific persona]
**Pain Points:**
1. [Most painful]
2. [Second]
3. [Third]
**Current Workarounds:** [how they cope today]

## 3. Solution
[2-3 sentences]

| Problem | Solution |
|---------|----------|
| ... | ... |

## 4. Customer Quote
> "[What would a happy user say?]"
> — [Persona], [Role]

## 5. Success Criteria
| Metric | Target | Measurement |
|--------|--------|-------------|
| ... | ... | ... |

## 6. Scope
**In Scope:**
- ...

**Non-Goals:**
- ...

## 7. Open Questions
| Question | Impact | Owner |
|----------|--------|-------|
| ... | ... | ... |
```

---

## Phase 7: Confirm

```bash
ls -la ~/gt/.agents/$RIG/products/
```

Tell user:
```
Product brief: ~/gt/.agents/$RIG/products/YYYY-MM-DD-topic.md

Next: /formulate ~/gt/.agents/$RIG/products/YYYY-MM-DD-topic.md
```

---

## Time Budget

**Target: 15 minutes or less.**

This is a lightweight brief, not a full PRD. If it's taking longer:
- Scope is too broad (narrow it)
- More research needed (run `/research` first)
- Over-engineering (simplify)

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| "Users" as persona | Name specific persona type |
| Features without problems | Map every feature to a pain point |
| Unmeasurable success | Quantify with targets |
| Missing non-goals | Explicitly state what is NOT being done |
| 2-hour briefs | Keep to 15 minutes |
| Skip for "obvious" features | Even obvious needs articulation |

---

## Execution Checklist

- [ ] Detected target rig (Phase 0)
- [ ] Checked for prior product docs (Phase 1)
- [ ] Identified specific customer persona (Phase 2)
- [ ] Listed ranked pain points (Phase 2)
- [ ] Wrote 10-word headline (Phase 3)
- [ ] Mapped problems to solutions (Phase 3)
- [ ] Defined measurable success criteria (Phase 4)
- [ ] Listed in-scope items (Phase 5)
- [ ] Listed non-goals explicitly (Phase 5)
- [ ] Wrote product brief to products/ (Phase 6)
- [ ] Confirmed output and next steps (Phase 7)

---

## Quick Example

**User**: "/product Add rate limiting to the API gateway"

**Agent workflow**:

```bash
# Phase 0: Rig Detection
RIG="athena"
mkdir -p ~/gt/.agents/$RIG/products/

# Phase 1: Prior Art
ls ~/gt/.agents/$RIG/products/ | grep -i rate
# None found

# Phase 2: Customer Discovery
# Who: API consumers on usage-based billing
# Pain: Unpredictable costs, no visibility, sudden cutoffs

# Phase 3: Solution
# Headline: "Predictable API costs with automatic protection"
# Maps: Unpredictable→limits, No visibility→dashboard, Cutoffs→graceful degradation

# Phase 4: Success Criteria
# 80% adoption, 30% cost reduction, 50% fewer complaints

# Phase 5: Scope
# In: Usage limits, alerts, dashboard
# Non-goals: DDoS protection, enterprise tier, pricing changes

# Phase 6: Output
# Write ~/gt/.agents/athena/products/2026-01-11-api-rate-limiting.md

# Phase 7: Confirm
```

**Result**: Product brief ready for `/formulate`.

---

## References

### JIT-Loadable Documentation

| Topic | Reference |
|-------|-----------|
| Full template | `references/template.md` |
| Good/bad examples | `references/examples.md` |

### Related Skills

- **research**: For technical exploration first
- **formulate**: When ready to decompose into engineering tasks
- **plan**: Alternative for one-time plans (no reusable formula)

### Industry Sources

- [Amazon PR/FAQ](https://productstrategy.co/working-backwards-the-amazon-prfaq-for-product-innovation/)
- [Google Design Docs](https://www.industrialempathy.com/posts/design-docs-at-google/)

---

## Workflow Integration

```
/research → /product → /formulate → /crank → /retro
    ↓          ↓           ↓           ↓        ↓
Technical   Product     Engineering  Execute  Learn
Findings    Brief       Tasks        Code
```

**Progressive Disclosure**: This skill provides core product-thinking workflow. For detailed template see `references/template.md`, for examples see `references/examples.md`.
