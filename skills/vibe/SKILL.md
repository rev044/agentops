---
name: vibe
description: 'Comprehensive code validation. Runs complexity analysis then multi-model council. Answer: Is this code ready to ship? Triggers: "vibe", "validate code", "check code", "review code", "is this ready".'
dependencies:
  - council    # multi-model judgment
  - complexity # complexity analysis
  - standards  # loaded for language-specific context
---

# Vibe Skill

> **Purpose:** Is this code ready to ship?

Two steps:
1. **Complexity analysis** — Find hotspots (radon, gocyclo)
2. **Council validation** — Multi-model judgment

---

## Quick Start

```bash
/vibe                      # validates recent changes
/vibe recent               # same as above
/vibe src/auth/            # validates specific path
/vibe --deep recent        # 3 judges instead of 2
/vibe --mixed recent       # cross-vendor (Claude + Codex)
```

---

## Execution Steps

### Step 1: Determine Target

**If target provided:** Use it directly.

**If no target or "recent":** Auto-detect from git:
```bash
# Check recent commits
git diff --name-only HEAD~3 2>/dev/null | head -20
```

If nothing found, ask user.

### Step 2: Run Complexity Analysis

**Detect language and run appropriate tool:**

**For Python:**
```bash
# Check if radon is available
if ! which radon > /dev/null 2>&1; then
  echo "⚠️ COMPLEXITY SKIPPED: radon not installed (pip install radon)"
  # Record in report that complexity was skipped
else
  # Run cyclomatic complexity
  radon cc <path> -a -s 2>/dev/null | head -30
  # Run maintainability index
  radon mi <path> -s 2>/dev/null | head -30
fi
```

**For Go:**
```bash
# Check if gocyclo is available
if ! which gocyclo > /dev/null 2>&1; then
  echo "⚠️ COMPLEXITY SKIPPED: gocyclo not installed (go install github.com/fzipp/gocyclo/cmd/gocyclo@latest)"
  # Record in report that complexity was skipped
else
  # Run complexity analysis
  gocyclo -over 10 <path> 2>/dev/null | head -30
fi
```

**For other languages:** Skip complexity with explicit note: "⚠️ COMPLEXITY SKIPPED: No analyzer for <language>"

**Interpret results:**

| Score | Rating | Action |
|-------|--------|--------|
| A (1-5) | Simple | Good |
| B (6-10) | Moderate | OK |
| C (11-20) | Complex | Flag for council |
| D (21-30) | Very complex | Recommend refactor |
| F (31+) | Untestable | Must refactor |

**Include complexity findings in council context.**

### Step 3: Run Council Validation

Run `/council validate` with complexity context:

```
/council validate <target>
```

**Council receives:**
- Files to review
- Complexity hotspots (from Step 2)
- Git diff context

**Default (2 judges):**
- Pragmatist: Is this implementable? What's overcomplicated?
- Skeptic: What could break? Security issues? Edge cases?

**With --deep (3 judges):**
```
/council --deep validate <target>
```
Adds Visionary: Architecture implications? Technical debt?

**With --mixed (cross-vendor):**
```
/council --mixed validate <target>
```
3 Claude + 3 Codex agents for cross-vendor consensus.

### Step 4: Council Checks

Each judge reviews for:

| Aspect | What to Look For |
|--------|------------------|
| **Correctness** | Does code do what it claims? |
| **Security** | Injection, auth issues, secrets |
| **Edge Cases** | Null handling, boundaries, errors |
| **Quality** | Dead code, duplication, clarity |
| **Complexity** | High cyclomatic scores, deep nesting |
| **Architecture** | Coupling, abstractions, patterns |

### Step 5: Interpret Verdict

| Council Verdict | Vibe Result | Action |
|-----------------|-------------|--------|
| PASS | Ready to ship | Merge/deploy |
| WARN | Review concerns | Address or accept risk |
| FAIL | Not ready | Fix issues |

### Step 6: Write Vibe Report

**Write to:** `.agents/council/YYYY-MM-DD-vibe-<target>.md`

```markdown
# Vibe Report: <Target>

**Date:** YYYY-MM-DD
**Files Reviewed:** <count>

## Complexity Analysis

**Status:** ✅ Completed | ⚠️ Skipped (<reason>)

| File | Score | Rating | Notes |
|------|-------|--------|-------|
| src/auth.py | 15 | C | Consider breaking up |
| src/utils.py | 4 | A | Good |

**Hotspots:** <list files with C or worse>
**Skipped reason:** <if skipped, explain why - e.g., "radon not installed">

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

## Decision

[ ] SHIP - Complexity acceptable, council passed
[ ] FIX - Address concerns before shipping
[ ] REFACTOR - High complexity, needs rework
```

### Step 7: Report to User

Tell the user:
1. Complexity hotspots (if any)
2. Council verdict (PASS/WARN/FAIL)
3. Key concerns
4. Location of vibe report

---

## Integration with Workflow

```
/implement issue-123
    │
    ▼
(coding, quick lint/test as you go)
    │
    ▼
/vibe                      ← You are here
    │
    ├── Complexity analysis (find hotspots)
    └── Council validation (multi-model judgment)
    │
    ├── PASS → ship it
    ├── WARN → review, then ship or fix
    └── FAIL → fix, re-run /vibe
```

---

## Examples

### Validate Recent Changes

```bash
/vibe recent
```

Runs complexity on recent changes, then council reviews.

### Validate Specific Directory

```bash
/vibe src/auth/
```

Complexity + council on auth directory.

### Deep Review

```bash
/vibe --deep recent
```

Complexity + 3 judges for thorough review.

### Cross-Vendor Consensus

```bash
/vibe --mixed recent
```

Complexity + Claude + Codex judges.

---

## Relationship to CI/CD

**Vibe runs:**
- Complexity analysis (radon, gocyclo)
- Council validation (multi-model judgment)

**CI/CD runs:**
- Linters
- Tests
- Security scanners
- Build

```
Developer workflow:
  /vibe recent → complexity + judgment

CI/CD workflow:
  git push → lint, test, scan → mechanical checks
```

Both should pass before shipping.

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/complexity/SKILL.md` — Standalone complexity analysis
- `skills/pre-mortem/SKILL.md` — Council validates plans
- `skills/post-mortem/SKILL.md` — Council validates completed work
- `skills/standards/SKILL.md` — Language-specific coding standards
