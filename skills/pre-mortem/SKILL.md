---
name: pre-mortem
tier: solo
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
/pre-mortem --quick path/to/PLAN.md                 # fast inline check, no spawning
/pre-mortem --deep path/to/SPEC.md                  # 4 judges with plan-review preset
/pre-mortem --mixed path/to/PLAN.md                 # cross-vendor (Claude + Codex)
/pre-mortem --preset=architecture path/to/PLAN.md   # architecture-focused review
/pre-mortem --explorers=3 path/to/SPEC.md           # deep investigation of plan
/pre-mortem --debate path/to/PLAN.md                # two-round adversarial review
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

### Step 1a: Search Knowledge Flywheel

```bash
if command -v ao &>/dev/null; then
    ao search "plan validation lessons <goal>" 2>/dev/null | head -10
fi
```
If ao returns prior plan review findings, include them as context for the council packet. Skip silently if ao is unavailable or returns no results.

### Step 2: Run Council Validation

Run `/council` with the **plan-review** preset and always 4 judges (--deep):

```
/council --deep --preset=plan-review validate <plan-path>
```

**Default (4 judges with plan-review perspectives):**
- `missing-requirements`: What's not in the spec that should be? What questions haven't been asked?
- `feasibility`: What's technically hard or impossible here? What will take 3x longer than estimated?
- `scope`: What's unnecessary? What's missing? Where will scope creep?
- `spec-completeness`: Are boundaries defined? Do conformance checks cover all acceptance criteria? Is the plan mechanically verifiable?

Pre-mortem always uses 4 judges (`--deep`) because plans deserve thorough review. The spec-completeness judge validates SDD patterns; for plans without boundaries/conformance sections, it issues WARN (not FAIL) for backward compatibility.

**With --quick (inline, no spawning):**
```
/council --quick validate <plan-path>
```
Single-agent structured review. Fast sanity check before committing to full council.

**With --mixed (cross-vendor):**
```
/council --mixed --preset=plan-review validate <plan-path>
```
3 Claude + 3 Codex agents for cross-vendor plan validation with plan-review perspectives.

**With explicit preset override:**
```
/pre-mortem --preset=architecture path/to/PLAN.md
```
Explicit `--preset` overrides the automatic plan-review preset. Uses architecture-focused personas instead.

**With explorers:**
```
/council --deep --preset=plan-review --explorers=3 validate <plan-path>
```
Each judge spawns 3 explorers to investigate aspects of the plan's feasibility against the codebase. Useful for complex migration or refactoring plans.

**With debate mode:**
```
/pre-mortem --debate
```
Enables adversarial two-round review for plan validation. Use for high-stakes plans where multiple valid approaches exist. See `/council` docs for full --debate details.

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
| Missing-Requirements | ... | ... |
| Feasibility | ... | ... |
| Scope | ... | ... |

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

### Step 5: Record Ratchet Progress

```bash
ao ratchet record pre-mortem 2>/dev/null || true
```

### Step 6: Report to User

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

3 judges (missing-requirements, feasibility, scope) review the auth system plan.

### Cross-Vendor Plan Validation

```bash
/pre-mortem --mixed .agents/plans/2026-02-05-auth-system.md
```

3 Claude + 3 Codex agents validate the plan with plan-review perspectives.

### Architecture-Focused Review

```bash
/pre-mortem --preset=architecture .agents/specs/api-v2-spec.md
```

3 judges with architecture perspectives (scalability, maintainability, simplicity) review the spec.

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
