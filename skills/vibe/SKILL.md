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

**Run mechanical checks BEFORE council. These catch errors that LLMs estimate instead of measure (L19, L22, L24).**

```bash
METADATA_FAILURES=""

# 1. File existence — every path referenced in recent changes must exist
for f in $(git diff --name-only HEAD~3 2>/dev/null); do
  if [ ! -f "$f" ]; then
    METADATA_FAILURES="${METADATA_FAILURES}\n- MISSING FILE: $f (in git diff but not on disk)"
  fi
done

# 2. Line counts — if any file claims a count (e.g., "# 150 lines"), verify
# Search for self-reported line counts in changed files
for f in $(git diff --name-only HEAD~3 2>/dev/null); do
  if [ -f "$f" ]; then
    claimed=$(grep -oP '(\d+)\s*lines' "$f" 2>/dev/null | head -1 | grep -oP '\d+')
    if [ -n "$claimed" ]; then
      actual=$(wc -l < "$f" | tr -d ' ')
      if [ "$claimed" -ne "$actual" ] 2>/dev/null; then
        METADATA_FAILURES="${METADATA_FAILURES}\n- LINE COUNT MISMATCH: $f claims ${claimed} lines, actual ${actual}"
      fi
    fi
  fi
done

# 3. Cross-references — internal doc links resolve
for f in $(git diff --name-only HEAD~3 2>/dev/null | grep -E '\.(md|txt)$'); do
  if [ -f "$f" ]; then
    for ref in $(grep -oP '\[.*?\]\(((?!http)[^)]+)\)' "$f" 2>/dev/null | grep -oP '\(([^)]+)\)' | tr -d '()'); do
      ref_dir=$(dirname "$f")
      if [ ! -f "$ref_dir/$ref" ] && [ ! -f "$ref" ]; then
        METADATA_FAILURES="${METADATA_FAILURES}\n- BROKEN LINK: $f references $ref (not found)"
      fi
    done
  fi
done

# 4. ASCII diagram sanity — boxes vs labels (>3 boxes need verification per L22)
for f in $(git diff --name-only HEAD~3 2>/dev/null | grep -E '\.(md|txt)$'); do
  if [ -f "$f" ]; then
    box_count=$(grep -cP '┌|╔|\+--' "$f" 2>/dev/null || echo 0)
    if [ "$box_count" -gt 3 ]; then
      label_count=$(grep -cP '│\s+\S' "$f" 2>/dev/null || echo 0)
      if [ "$box_count" -gt "$label_count" ]; then
        METADATA_FAILURES="${METADATA_FAILURES}\n- DIAGRAM CHECK: $f has ${box_count} boxes but only ${label_count} label lines — verify diagram accuracy"
      fi
    fi
  fi
done

# Report results
if [ -n "$METADATA_FAILURES" ]; then
  echo "METADATA VERIFICATION FAILURES:"
  echo -e "$METADATA_FAILURES"
else
  echo "Metadata verification: all checks passed"
fi
```

**If failures found:** Include them in the council packet as `context.metadata_failures`. These are MECHANICAL findings — council should not need to re-discover them. Council judges focus on structural and logical issues instead.

**If all pass:** Note "Metadata verification: N checks passed" in the vibe report.

**Why:** LLMs estimate metadata from content complexity, not measurement (L24). Line counts, cross-references, and diagram accuracy are mechanical — verify them mechanically. This frees council judges to focus on correctness, architecture, and security.

### Step 2c: Deterministic Validation (Olympus)

**Guard:** Only run when BOTH conditions are true:
1. `.ol/config.yaml` exists in the project root
2. `which ol` succeeds (ol CLI is on PATH)

If either condition fails, skip this step entirely (no-op). Non-Olympus projects are unaffected.

**Detection:**
```bash
if [ -f ".ol/config.yaml" ] && which ol > /dev/null 2>&1; then
  OL_PROJECT=true
else
  OL_PROJECT=false
fi
```

**If OL_PROJECT is true:**

1. Determine `<quest-id>` and `<bead-id>` from context:
   - Check `.ol/config.yaml` for current quest
   - Or extract from git branch name (e.g., `ol-572/bead-3`)
   - Or from the target argument if it looks like an OL bead ID

2. Run deterministic validation:
```bash
ol validate stage1 --quest <quest-id> --bead <bead-id> --worktree .
```

3. Parse the JSON output (Stage1Result format):
```json
{
  "quest_id": "ol-572",
  "bead_id": "ol-572.3",
  "worktree": "/path/to/worktree",
  "passed": true,
  "steps": [
    {"name": "go build", "passed": true, "duration": "1.2s"},
    {"name": "go vet", "passed": true, "duration": "0.8s"},
    {"name": "go test", "passed": true, "duration": "3.4s"}
  ],
  "summary": "all steps passed"
}
```

4. **If `passed: false`:** Auto-FAIL the vibe immediately. Do NOT proceed to council. Write the vibe report with:
   - Verdict: **FAIL**
   - Include all step details showing which steps failed
   - Recommendation: "Fix deterministic validation failures before running council review."

5. **If `passed: true`:** Record results and proceed to council (Step 4). Include the Stage1Result in the vibe report as a "Deterministic Validation" section and pass it as context to council.

6. **If `ol validate stage1` exits non-zero (error):** Log the error, note it in the report as "Deterministic validation: SKIPPED (ol error)", and proceed to council normally.

**Include in vibe report (Step 7)** — add before the "Council Verdict" section:

```markdown
## Deterministic Validation (Olympus)

**Status:** PASS | FAIL | SKIPPED
**Quest:** <quest-id> | **Bead:** <bead-id>

| Step | Result | Duration |
|------|--------|----------|
| go build | PASS | 1.2s |
| go vet | PASS | 0.8s |
| go test | PASS | 3.4s |

**Summary:** <summary from Stage1Result>
```

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
3 independent judges (no perspective labels). Vibe always uses `--deep` for consistency with /pre-mortem and /post-mortem.

**Council receives:**
- Files to review
- Complexity hotspots (from Step 2)
- Git diff context
- Spec content (when found, in `context.spec`)

**With --quick (inline, no spawning):**
```
/council --quick validate <target>
```
Single-agent structured self-review. Fast, cheap, good for mid-implementation checks.

**With explicit --deep (redundant — vibe always uses --deep):**
```
/council --deep validate <target>
```
3 independent judges (no perspective labels). Same as default vibe behavior.

**With --mixed (cross-vendor):**
```
/council --mixed validate <target>
```
3 Claude + 3 Codex agents for cross-vendor consensus.

**With explicit preset override:**
```
/vibe --preset=security-audit src/auth/
```
Explicit `--preset` overrides the automatic code-review preset. Uses security-focused personas instead.

**With explorers:**
```
/vibe --explorers=2 src/auth/
```
Each judge spawns 2 explorer sub-agents to investigate code patterns before judging. Useful for large codebases.

**With debate mode:**
```
/vibe --debate recent
```
Enables adversarial two-round review where judges critique each other's findings before final verdict. Use for high-stakes reviews where judges are likely to disagree. See `/council` docs for full --debate details.

**Timeout:** Vibe inherits council timeout settings. If judges time out,
the council report will note partial results. Vibe treats a partial council
report the same as a full report — the verdict stands with available judges.

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
