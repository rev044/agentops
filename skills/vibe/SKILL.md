---
name: vibe
tier: solo
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
/vibe                                    # validates recent changes
/vibe recent                             # same as above
/vibe src/auth/                          # validates specific path
/vibe --quick recent                     # fast inline check, no agent spawning
/vibe --deep recent                      # 3 judges instead of 2
/vibe --mixed recent                     # cross-vendor (Claude + Codex)
/vibe --preset=security-audit src/auth/  # security-focused review
/vibe --explorers=2 recent               # judges with explorer sub-agents
/vibe --debate recent                    # two-round adversarial review
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

**Pre-flight: If no files found:**
Return immediately with: "PASS (no changes to review) — no modified files detected."
Do NOT spawn agents for empty file lists.

### Step 2: Run Complexity Analysis

**Detect language and run appropriate tool:**

**For Python:**
```bash
# Check if radon is available
mkdir -p .agents/council
echo "$(date -Iseconds) preflight: checking radon" >> .agents/council/preflight.log
if ! which radon >> .agents/council/preflight.log 2>&1; then
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
echo "$(date -Iseconds) preflight: checking gocyclo" >> .agents/council/preflight.log
if ! which gocyclo >> .agents/council/preflight.log 2>&1; then
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

### Step 2a: Run Constraint Tests

**If the project has constraint tests, run them before council:**

```bash
# Check if constraint tests exist (Olympus pattern)
if [ -d "internal/constraints" ] && ls internal/constraints/*_test.go &>/dev/null; then
  echo "Running constraint tests..."
  go test ./internal/constraints/ -run TestConstraint -v 2>&1
  # If FAIL → include failures in council context as CRITICAL findings
  # If PASS → note "N constraint tests passed" in report
fi
```

**Why:** Constraint tests catch mechanical violations (ghost references, TOCTOU races, dead code at entry points) that council judges miss. Proven by Argus ghost ref in ol-571 — council gave PASS while constraint test caught it.

Include constraint test results in the council packet context. Failed constraint tests are CRITICAL findings that override council PASS verdict.

### Step 2b: Metadata Verification Checklist (MANDATORY)

Run mechanical checks BEFORE council — catches errors LLMs estimate instead of measure:
1. **File existence** — every path in `git diff --name-only HEAD~3` must exist on disk
2. **Line counts** — if a file claims "N lines", verify with `wc -l`
3. **Cross-references** — internal markdown links resolve to existing files
4. **Diagram sanity** — files with >3 ASCII boxes should have matching labels

Include failures in council packet as `context.metadata_failures` (MECHANICAL findings). If all pass, note in report.

### Step 2c: Deterministic Validation (Olympus)

**Guard:** Only run when `.ol/config.yaml` exists AND `which ol` succeeds. Skip silently otherwise.

If OL project detected: run `ol validate stage1 --quest <quest-id> --bead <bead-id> --worktree .`
- **`passed: false`** → Auto-FAIL the vibe. Do NOT proceed to council.
- **`passed: true`** → Include Stage1Result in council context. Proceed normally.
- **Error/non-zero exit** → Note "SKIPPED (ol error)" in report. Proceed to council.

### Step 2.5: Codex Review (if available)

Run a fast, diff-focused code review via Codex CLI before council:

```bash
echo "$(date -Iseconds) preflight: checking codex" >> .agents/council/preflight.log
if which codex >> .agents/council/preflight.log 2>&1; then
  codex review --uncommitted > .agents/council/codex-review-pre.md 2>&1 && \
    echo "Codex review complete — output at .agents/council/codex-review-pre.md" || \
    echo "Codex review skipped (failed)"
else
  echo "Codex review skipped (CLI not found)"
fi
```

**If output exists**, include in council packet as additional context:
```json
"codex_review": {
  "source": "codex review --uncommitted",
  "content": "<contents of .agents/council/codex-review-pre.md>"
}
```

This gives council judges a Codex-generated review as pre-existing context — cheap, fast, diff-focused. It does NOT replace council judgment; it augments it.

**Skip conditions:**
- Codex CLI not on PATH → skip silently
- `codex review` fails → skip silently, proceed with council only
- No uncommitted changes → skip (nothing to review)

### Step 3: Load the Spec (New)

Before invoking council, try to find the relevant spec/bead:

1. **If target looks like a bead ID** (e.g., `na-0042`): `bd show <id>` to get the spec
2. **Search for plan doc:** `ls .agents/plans/ | grep <target-keyword>`
3. **Check git log:** `git log --oneline | head -10` to find the relevant bead reference

If a spec is found, include it in the council packet's `context.spec` field:
```json
{
  "spec": {
    "source": "bead na-0042",
    "content": "<the spec/bead description text>"
  }
}
```

### Step 4: Run Council Validation

**With spec found — use code-review preset (3 judges):**
```
/council --deep --preset=code-review validate <target>
```
- `error-paths`: Trace every error handling path. What's uncaught? What fails silently?
- `api-surface`: Review every public interface. Is the contract clear? Breaking changes?
- `spec-compliance`: Compare implementation against the spec. What's missing? What diverges?

The spec content is injected into the council packet context so the `spec-compliance` judge can compare implementation against it.

**Without spec — 3 independent judges (no perspectives):**
```
/council --deep validate <target>
```
3 independent judges (no perspective labels). Vibe uses `--deep` by default (3 judges) for consistency with /pre-mortem and /post-mortem, but users can override with `--quick` (inline single-agent check) or `--mixed` (cross-vendor with Codex).

**Council receives:**
- Files to review
- Complexity hotspots (from Step 2)
- Git diff context
- Spec content (when found, in `context.spec`)

All council flags pass through: `--quick` (inline), `--mixed` (cross-vendor), `--preset=<name>` (override perspectives), `--explorers=N`, `--debate` (adversarial 2-round). See Quick Start examples and `/council` docs.

### Step 5: Council Checks

Each judge reviews for:

| Aspect | What to Look For |
|--------|------------------|
| **Correctness** | Does code do what it claims? |
| **Security** | Injection, auth issues, secrets |
| **Edge Cases** | Null handling, boundaries, errors |
| **Quality** | Dead code, duplication, clarity |
| **Complexity** | High cyclomatic scores, deep nesting |
| **Architecture** | Coupling, abstractions, patterns |

### Step 6: Interpret Verdict

| Council Verdict | Vibe Result | Action |
|-----------------|-------------|--------|
| PASS | Ready to ship | Merge/deploy |
| WARN | Review concerns | Address or accept risk |
| FAIL | Not ready | Fix issues |

### Step 7: Write Vibe Report

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
| Error-Paths | ... | ... (with spec — code-review preset) |
| API-Surface | ... | ... (with spec — code-review preset) |
| Spec-Compliance | ... | ... (with spec — code-review preset) |
| Judge 1 | ... | ... (no spec — 3 independent judges) |
| Judge 2 | ... | ... (no spec — 3 independent judges) |
| Judge 3 | ... | ... (no spec — 3 independent judges) |

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

### Step 8: Report to User

Tell the user:
1. Complexity hotspots (if any)
2. Council verdict (PASS/WARN/FAIL)
3. Key concerns
4. Location of vibe report

### Step 9: Record Ratchet Progress

After council verdict:
1. If verdict is PASS or WARN:
   - Run: `ao ratchet record vibe --output "<report-path>" 2>/dev/null || true`
   - Suggest: "Run /post-mortem to capture learnings and complete the cycle."
2. If verdict is FAIL:
   - Do NOT record. Tell user to fix issues and re-run /vibe.

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

## Consolidation

For conflict resolution between agent findings, follow the algorithm in `.agents/specs/conflict-resolution-algorithm.md`.

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/complexity/SKILL.md` — Standalone complexity analysis
- `skills/pre-mortem/SKILL.md` — Council validates plans
- `skills/post-mortem/SKILL.md` — Council validates completed work
- `skills/standards/SKILL.md` — Language-specific coding standards
- `.agents/specs/conflict-resolution-algorithm.md` — Conflict resolution algorithm
