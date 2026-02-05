---
name: pre-mortem
description: 'Validate a plan or spec before implementation using multi-model council. Answer: Is this good enough to implement? Triggers: "pre-mortem", "validate plan", "validate spec", "is this ready".'
dependencies:
  - council  # multi-model judgment
---

# Pre-Mortem Skill

> **Purpose:** Is this plan/spec good enough to implement?

Run `/council validate` on a plan or spec to get multi-model judgment before committing to implementation.

---

## Quick Start

```bash
/pre-mortem                                         # validates most recent plan
/pre-mortem path/to/PLAN.md                         # validates specific plan
/pre-mortem --deep path/to/SPEC.md                  # 3 judges instead of 2
/pre-mortem --mixed path/to/PLAN.md                 # cross-vendor (Claude + Codex)
/pre-mortem --preset=architecture path/to/PLAN.md   # architecture-focused review
/pre-mortem --explorers=3 path/to/SPEC.md           # deep investigation of plan
```

---

## Execution Steps

### Step 1: Find the Plan/Spec

**If path provided:** Use it directly.

**If no path:** Find most recent plan:
```bash
ls -lt .agents/plans/ 2>/dev/null | head -3
ls -lt .agents/specs/ 2>/dev/null | head -3
```

Use the most recent file. If nothing found, ask user.

### Step 2: Run Council Validation

Run `/council validate` on the plan/spec:

```
/council validate <plan-path>
```

**Default (2 judges):**
- Pragmatist: Is this implementable? What's missing?
- Skeptic: What could go wrong? What's over-engineered?

**With --deep (3 judges):**
```
/council --deep validate <plan-path>
```
Adds Visionary: Where does this lead? What's the 10x version?

**With --mixed (cross-vendor):**
```
/council --mixed validate <plan-path>
```
3 Claude + 3 Codex agents for cross-vendor plan validation.

**With preset override:**
```
/council --preset=architecture validate <plan-path>
```
Uses architecture-focused personas (scalability, maintainability, simplicity) for system design plans.

**With explorers:**
```
/council --explorers=3 validate <plan-path>
```
Each judge spawns 3 explorers to investigate aspects of the plan's feasibility against the codebase. Useful for complex migration or refactoring plans.

### Step 3: Interpret Council Verdict

| Council Verdict | Pre-Mortem Result | Action |
|-----------------|-------------------|--------|
| PASS | Ready to implement | Proceed |
| WARN | Review concerns | Address warnings or accept risk |
| FAIL | Not ready | Fix issues before implementing |

### Step 4: Write Pre-Mortem Report

**Write to:** `.agents/council/YYYY-MM-DD-pre-mortem-<topic>.md`

```markdown
# Pre-Mortem: <Topic>

**Date:** YYYY-MM-DD
**Plan/Spec:** <path>

## Council Verdict: PASS / WARN / FAIL

| Judge | Verdict | Key Finding |
|-------|---------|-------------|
| Pragmatist | ... | ... |
| Skeptic | ... | ... |
| Visionary | ... | (if --deep) |

## Shared Findings
- ...

## Concerns Raised
- ...

## Recommendation
<council recommendation>

## Decision Gate

[ ] PROCEED - Council passed, ready to implement
[ ] ADDRESS - Fix concerns before implementing
[ ] RETHINK - Fundamental issues, needs redesign
```

### Step 5: Report to User

Tell the user:
1. Council verdict (PASS/WARN/FAIL)
2. Key concerns (if any)
3. Recommendation
4. Location of pre-mortem report

---

## Integration with Workflow

```
/plan epic-123
    │
    ▼
/pre-mortem                    ← You are here
    │
    ├── PASS → /implement
    ├── WARN → Review, then /implement or fix
    └── FAIL → Fix plan, re-run /pre-mortem
```

---

## Examples

### Validate a Plan

```bash
/pre-mortem .agents/plans/2026-02-05-auth-system.md
```

Council reviews the auth system plan, reports whether it's ready to implement.

### Deep Validation for Complex Specs

```bash
/pre-mortem --deep .agents/specs/api-v2-spec.md
```

3 judges (pragmatist, skeptic, visionary) review the spec thoroughly.

### Cross-Vendor Plan Validation

```bash
/pre-mortem --mixed .agents/plans/2026-02-05-auth-system.md
```

3 Claude + 3 Codex agents validate the plan from different vendor perspectives.

### Auto-Find Recent Plan

```bash
/pre-mortem
```

Finds the most recent plan in `.agents/plans/` and validates it.

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/plan/SKILL.md` — Create implementation plans
- `skills/vibe/SKILL.md` — Validate code after implementation
