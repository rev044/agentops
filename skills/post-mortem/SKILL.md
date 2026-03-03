---
name: post-mortem
description: 'Wrap up completed work. Council validates the implementation, then extract learnings. Triggers: "post-mortem", "wrap up", "close epic", "what did we learn".'
skill_api_version: 1
metadata:
  tier: judgment
  dependencies:
    - council  # multi-model judgment
    - retro    # optional - extracts learnings (graceful skip on failure)
    - beads    # optional - for issue status
---

# Post-Mortem Skill

> **Purpose:** Wrap up completed work — validate it shipped correctly and extract learnings.

Two steps:
1. `/council validate` — Did we implement it correctly?
2. `/retro` — What did we learn?

---

## Quick Start

```bash
/post-mortem                    # wraps up recent work
/post-mortem epic-123           # wraps up specific epic
/post-mortem --quick recent     # fast inline wrap-up, no spawning
/post-mortem --deep recent      # thorough council review
/post-mortem --mixed epic-123   # cross-vendor (Claude + Codex)
/post-mortem --explorers=2 epic-123  # deep investigation before judging
/post-mortem --debate epic-123      # two-round adversarial review
/post-mortem --skip-checkpoint-policy epic-123  # skip ratchet chain validation
```

---

## Execution Steps

### Pre-Flight Checks

Before proceeding, verify:
1. **Git repo exists:** `git rev-parse --git-dir 2>/dev/null` — if not, error: "Not in a git repository"
2. **Work was done:** `git log --oneline -1 2>/dev/null` — if empty, error: "No commits found. Run /implement first."
3. **Epic context:** If epic ID provided, verify it has closed children. If 0 closed children, error: "No completed work to review."

### Step 0.4: Load Reference Documents (MANDATORY)

Before Step 0.5 and Step 2.5, load required reference docs into context using the Read tool:

```
REQUIRED_REFS=(
  "skills/post-mortem/references/checkpoint-policy.md"
  "skills/post-mortem/references/metadata-verification.md"
  "skills/post-mortem/references/closure-integrity-audit.md"
)
```

For each reference file, use the **Read tool** to load its content and hold it in context for use in later steps. Do NOT just test file existence with `[ -f ]` -- actually read the content so it is available when Steps 0.5 and 2.5 need it.

If a reference file does not exist (Read returns an error), log a warning and add it as a checkpoint warning in the council context. Proceed only if the missing reference is intentionally deferred.

### Step 0.5: Checkpoint-Policy Preflight (MANDATORY)

Read `references/checkpoint-policy.md` for the full checkpoint-policy preflight procedure. It validates the ratchet chain, checks artifact availability, and runs idempotency checks. BLOCK on prior FAIL verdicts; WARN on everything else.

### Step 1: Identify Completed Work and Record Timing

**Record the post-mortem start time for cycle-time tracking:**
```bash
PM_START=$(date +%s)
```

**If epic/issue ID provided:** Use it directly.

**If no ID:** Find recently completed work:
```bash
# Check for closed beads
bd list --status closed --since "7 days ago" 2>/dev/null | head -5

# Or check recent git activity
git log --oneline --since="7 days ago" | head -10
```

### Step 2: Load the Original Plan/Spec

Before invoking council, load the original plan for comparison:

1. **If epic/issue ID provided:** `bd show <id>` to get the spec/description
2. **Search for plan doc:** `ls .agents/plans/ | grep <target-keyword>`
3. **Check git log:** `git log --oneline | head -10` to find the relevant bead reference

If a plan is found, include it in the council packet's `context.spec` field:
```json
{
  "spec": {
    "source": "bead na-0042",
    "content": "<the original plan/spec text>"
  }
}
```

### Step 2.2: Load Implementation Summary

Check for a crank-generated phase-2 summary:

```bash
PHASE2_SUMMARY=$(ls -t .agents/rpi/phase-2-summary-*-crank.md 2>/dev/null | head -1)
if [ -n "$PHASE2_SUMMARY" ]; then
    echo "Phase-2 summary found: $PHASE2_SUMMARY"
    # Read the summary with the Read tool for implementation context
fi
```

If available, use the phase-2 summary to understand what was implemented, how many waves ran, and which files were modified.

### Step 2.3: Reconcile Plan vs Delivered Scope

Compare the original plan scope against what was actually delivered:

1. Read the plan from `.agents/plans/` (most recent)
2. Compare planned issues against closed issues (`bd children <epic-id>`)
3. Note any scope additions, removals, or modifications
4. Include scope delta in the post-mortem findings

### Step 2.4: Closure Integrity Audit (MANDATORY)

Read `references/closure-integrity-audit.md` for the full procedure. Mechanically verifies:

1. **Git evidence per child** — every closed child has at least one commit referencing it or touching its scoped files
2. **Phantom bead detection** — flags children with generic titles ("task") or empty descriptions
3. **Orphaned children** — beads in `bd list` but not linked to parent in `bd show`
4. **Multi-wave regression detection** — for crank epics, checks if a later wave removed code added by an earlier wave
5. **Stretch goal audit** — verifies deferred stretch goals have documented rationale

Include results in the council packet as `context.closure_integrity`. WARN on 1-2 findings, FAIL on 3+.

### Step 2.5: Pre-Council Metadata Verification (MANDATORY)

Read `references/metadata-verification.md` for the full verification procedure. Mechanically checks: plan vs actual files, file existence in commits, cross-references in docs, and ASCII diagram integrity. Failures are included in the council packet as `context.metadata_failures`.

### Step 2.6: Pre-Council Deep Audit Sweep

**Skip if `--quick` or `--skip-sweep`.**

Before council runs, dispatch a deep audit sweep to systematically discover issues across all changed files. This uses the same protocol as `/vibe --deep` — see the deep audit protocol in the vibe skill (`skills/vibe/`) for the full specification.

In summary:

1. Identify all files in scope (from epic commits or recent changes)
2. Chunk files into batches of 3–5 by line count (<=100 lines → batch of 5, 101–300 → batch of 3, >300 → solo)
3. Dispatch up to 8 Explore agents in parallel, each with a mandatory 8-category checklist per file (resource leaks, string safety, dead code, hardcoded values, edge cases, concurrency, error handling, HTTP/web security)
4. Merge all explorer findings into a sweep manifest at `.agents/council/sweep-manifest.md`
5. Include sweep manifest in council packet — judges shift to adjudication mode (confirm/reject/reclassify sweep findings + add cross-cutting findings)

**Why:** Post-mortem council judges exhibit satisfaction bias when reviewing monolithic file sets — they stop at ~10 findings regardless of actual issue count. Per-file explorers with category checklists find 3x more issues, and the sweep manifest gives judges structured input to adjudicate rather than discover from scratch.

**Skip conditions:**
- `--quick` flag → skip (fast inline path)
- `--skip-sweep` flag → skip (old behavior: judges do pure discovery)
- No source files in scope → skip (nothing to audit)

### Step 3: Council Validates the Work

Run `/council` with the **retrospective** preset and always 3 judges:

```
/council --deep --preset=retrospective validate <epic-or-recent>
```

**Default (3 judges with retrospective perspectives):**
- `plan-compliance`: What was planned vs what was delivered? What's missing? What was added?
- `tech-debt`: What shortcuts were taken? What will bite us later? What needs cleanup?
- `learnings`: What patterns emerged? What should be extracted as reusable knowledge?

Post-mortem always uses 3 judges (`--deep`) because completed work deserves thorough review.

**Timeout:** Post-mortem inherits council timeout settings. If judges time out,
the council report will note partial results. Post-mortem treats a partial council
report the same as a full report — the verdict stands with available judges.

The plan/spec content is injected into the council packet context so the `plan-compliance` judge can compare planned vs delivered.

**With --quick (inline, no spawning):**
```
/council --quick validate <epic-or-recent>
```
Single-agent structured review. Fast wrap-up without spawning.

**With debate mode:**
```
/post-mortem --debate epic-123
```
Enables adversarial two-round review for post-implementation validation. Use for high-stakes shipped work where missed findings have production consequences. See `/council` docs for full --debate details.

**Advanced options (passed through to council):**
- `--mixed` — Cross-vendor (Claude + Codex) with retrospective perspectives
- `--preset=<name>` — Override with different personas (e.g., `--preset=ops` for production readiness)
- `--explorers=N` — Each judge spawns N explorers to investigate the implementation deeply before judging
- `--debate` — Two-round adversarial review (judges critique each other's findings before final verdict)

### Step 4: Extract Learnings

Run `/retro` to capture what we learned:

```
/retro <epic-or-recent>
```

**Retro captures:**
- What went well?
- What was harder than expected?
- What would we do differently?
- Patterns to reuse?
- Anti-patterns to avoid?

**Error Handling:**

| Failure | Behavior |
|---------|----------|
| Council fails | Stop, report council error, no retro |
| Retro fails | Proceed, report learnings as "⚠️ SKIPPED: retro unavailable" |
| Both succeed | Full post-mortem with council + learnings |

Post-mortem always completes if council succeeds. Retro is optional enrichment.

### Step 4.5: Compile Constraint Templates

For each extracted learning scoring >= 4/5 on actionability AND tagged "constraint" or "anti-pattern", run `bash hooks/constraint-compiler.sh <learning-path>` to generate a constraint template.

```bash
# Compile high-scoring constraint/anti-pattern learnings into enforcement templates
for f in .agents/learnings/YYYY-MM-DD-*.md; do
    [ -f "$f" ] || continue
    bash hooks/constraint-compiler.sh "$f" 2>/dev/null || true
done
```

This produces draft constraint templates in `.agents/constraints/` that can later be activated via `ao constraint activate <id>`.

### Step 5: Write Post-Mortem Report

**Write to:** `.agents/council/YYYY-MM-DD-post-mortem-<topic>.md`

```markdown
---
id: post-mortem-YYYY-MM-DD-<topic-slug>
type: post-mortem
date: YYYY-MM-DD
source: "[[.agents/plans/YYYY-MM-DD-<plan-slug>]]"
---

# Post-Mortem: <Epic/Topic>

**Epic:** <epic-id or "recent">
**Duration:** <elapsed time from PM_START to now>
**Cycle-Time Trend:** <compare against prior post-mortems — is this faster or slower? Check .agents/retros/ for prior Duration values>

## Council Verdict: PASS / WARN / FAIL

| Judge | Verdict | Key Finding |
|-------|---------|-------------|
| Plan-Compliance | ... | ... |
| Tech-Debt | ... | ... |
| Learnings | ... | ... |

### Implementation Assessment
<council summary>

### Concerns
<any issues found>

## Learnings (from /retro)

### What Went Well
- ...

### What Was Hard
- ...

### Do Differently Next Time
- ...

### Patterns to Reuse
- ...

### Anti-Patterns to Avoid
- ...

## Proactive Improvement Agenda

| # | Area | Improvement | Priority | Horizon | Effort | Evidence |
|---|------|-------------|----------|---------|--------|----------|
| 1 | repo / execution / ci-automation | ... | P0/P1/P2 | now/next-cycle/later | S/M/L | ... |

## Prior Findings Resolution Tracking

| Metric | Value |
|---|---|
| Backlog entries analyzed | ... |
| Prior findings total | ... |
| Resolved findings | ... |
| Unresolved findings | ... |
| Resolution rate | ...% |

| Source Epic | Findings | Resolved | Unresolved | Resolution Rate |
|---|---:|---:|---:|---:|
| ... | ... | ... | ... | ...% |

## Command-Surface Parity Checklist

| Command File | Run-path Covered by Test? | Evidence (file:line or test name) | Intentionally Uncovered? | Reason |
|---|---|---|---|---|
| cli/cmd/ao/<command>.go | yes/no | ... | yes/no | ... |

### Recommended Next /rpi
/rpi "<highest-value improvement>"

## Status

[ ] CLOSED - Work complete, learnings captured
[ ] FOLLOW-UP - Issues need addressing (create new beads)
```

### Step 5.5: Synthesize Proactive Improvement Agenda (MANDATORY)

**After writing the post-mortem report, analyze retro + council context and proactively propose improvements to repo quality and execution quality.**

Read the retro output (from Step 4) and the council report (from Step 3). For each learning, ask:
1. **What process does this improve?** (build, test, review, deploy, documentation, automation, etc.)
2. **What's the concrete change?** (new check, new automation, workflow change, tooling improvement)
3. **Is it actionable in one RPI cycle?** (if not, split into smaller pieces)

Coverage requirements:
- Include **ALL** improvements found (no cap).
- Cover all three surfaces:
  - `repo` (code/contracts/docs quality)
  - `execution` (planning/implementation/review workflow)
  - `ci-automation` (validation/tooling reliability)
- Include at least **1 quick win** (small, low-risk, same-session viable).

Write process improvement items with type `process-improvement` (distinct from `tech-debt` or `improvement`). Each item must have:
- `title`: imperative form, e.g. "Add pre-commit lint check"
- `area`: which part of the development process to improve
- `description`: 2-3 sentences describing the change and why retro evidence supports it
- `evidence`: which retro finding or council finding motivates this
- `priority`: P0 / P1 / P2
- `horizon`: now / next-cycle / later
- `effort`: S / M / L

**These items feed directly into Step 8 (Harvest Next Work) alongside council findings. They are the flywheel's growth vector — each cycle makes the system smarter.**

Write this into the post-mortem report under `## Proactive Improvement Agenda`.

Example output:
```markdown
## Proactive Improvement Agenda

| # | Area | Improvement | Priority | Horizon | Effort | Evidence |
|---|------|-------------|----------|---------|--------|----------|
| 1 | ci-automation | Add validation metadata requirement for Go tasks | P0 | now | S | Workers shipped untested code when metadata didn't require `go test` |
| 2 | execution | Add consistency-check finding category in review | P1 | next-cycle | M | Partial refactoring left stale references undetected |

### Recommended Next /rpi
/rpi "<highest-value improvement>"
```

### Step 5.6: Prior-Findings Resolution Tracking (MANDATORY)

After Step 5.5, compute and include prior-findings resolution tracking from `.agents/rpi/next-work.jsonl`. Read `references/harvest-next-work.md` for the jq queries that compute totals and per-source resolution rates. Write results into `## Prior Findings Resolution Tracking` in the post-mortem report.

### Step 5.7: Command-Surface Parity Gate (MANDATORY)

Before marking post-mortem complete, enforce command-surface parity for modified CLI commands:

1. Identify modified command files under `cli/cmd/ao/` from the reviewed scope.
2. For each file, record at least one tested run-path (unit/integration/e2e) in `## Command-Surface Parity Checklist`.
3. Any intentionally uncovered command family must be explicitly listed with a reason and follow-up item.

If any modified command file is missing both coverage evidence and an intentional-uncovered rationale, post-mortem cannot be marked complete.

### Step 6: Feed the Knowledge Flywheel

Post-mortem automatically feeds learnings into the flywheel:

```bash
if command -v ao &>/dev/null; then
  ao forge markdown .agents/learnings/*.md 2>/dev/null
  echo "Learnings indexed in knowledge flywheel"

  # Validate and lock artifacts that passed council review
  ao temper validate .agents/learnings/YYYY-MM-DD-*.md 2>/dev/null || true
  echo "Artifacts validated for tempering"

  # Close session and trigger full flywheel close-loop (includes adaptive feedback)
  ao session close 2>/dev/null || true
  ao flywheel close-loop --quiet 2>/dev/null || true
  echo "Session closed, flywheel loop triggered"
else
  # Learnings are already in .agents/learnings/ from /retro (Step 4).
  # Without ao CLI, grep-based search in /research and /inject
  # will find them directly — no copy to pending needed.

  # Feedback-loop fallback: update confidence for cited learnings
  mkdir -p .agents/ao
  if [ -f .agents/ao/citations.jsonl ]; then
    echo "Processing citation feedback (ao-free fallback)..."
    # Read cited learning files and boost confidence notation
    while IFS= read -r line; do
      CITED_FILE=$(echo "$line" | grep -o '"learning_file":"[^"]*"' | cut -d'"' -f4)
      if [ -f "$CITED_FILE" ]; then
        # Note: confidence boost tracked via citation count, not file modification
        echo "Cited: $CITED_FILE"
      fi
    done < .agents/ao/citations.jsonl
  fi

  # Session-outcome fallback: record this session's outcome
  EPIC_ID="<epic-id>"
  echo "{\"epic\": \"$EPIC_ID\", \"verdict\": \"<council-verdict>\", \"cycle_time_minutes\": 0, \"timestamp\": \"$(date -Iseconds)\"}" >> .agents/ao/outcomes.jsonl

  # Skip ao temper validate (no fallback needed — tempering is an optimization)
  echo "Flywheel fed locally (ao CLI not available — learnings searchable via grep)"
fi
```

### Step 7: Report to User

Tell the user:
1. Council verdict on implementation
2. Key learnings
3. Any follow-up items
4. Location of post-mortem report
5. Knowledge flywheel status
6. **Suggested next `/rpi` command** (ALWAYS — this is how the flywheel spins itself)
7. ALL proactive improvements, organized by priority (highlight one quick win)

**The next `/rpi` suggestion is MANDATORY, not opt-in.** After every post-mortem, present the highest-severity harvested item as a ready-to-copy command:

```markdown
## Flywheel: Next Cycle

Based on this post-mortem, the highest-priority follow-up is:

> **<title>** (<type>, <severity>)
> <1-line description>

Ready to run:
```
/rpi "<title>"
```

Or see all N harvested items in `.agents/rpi/next-work.jsonl`.
```

If no items were harvested, write: "Flywheel stable — no follow-up items identified."

### Step 8: Harvest Next Work

Scan the council report and retro for actionable follow-up items:

1. **Council findings:** Extract tech debt, warnings, and improvement suggestions from the council report (items with severity "significant" or "critical" that weren't addressed in this epic)
2. **Retro patterns:** Extract recurring patterns from retro learnings that warrant dedicated RPIs (items from "Do Differently Next Time" and "Anti-Patterns to Avoid")
3. **Process improvements:** Include all items from Step 5.5 (type: `process-improvement`). These are the flywheel's growth vector — each cycle makes development more effective.
4. **Footgun entries (REQUIRED):** Extract platform-specific gotchas, surprising API behaviors, or silent-failure modes discovered during implementation. Each must include: trigger condition, observable symptom, and fix. Write as type `pattern-fix` with source `retro-learning`. If a footgun was discovered this cycle, it must appear in this harvest — do not defer.
5. **Write `## Next Work` section** to the post-mortem report:

```markdown
## Next Work

| # | Title | Type | Severity | Source | Target Repo |
|---|-------|------|----------|--------|-------------|
| 1 | <title> | tech-debt / improvement / pattern-fix / process-improvement | high / medium / low | council-finding / retro-learning / retro-pattern | <repo-name or *> |
```

6. **SCHEMA VALIDATION (MANDATORY):** Before writing, validate each harvested item against the schema contract (`.agents/rpi/next-work.schema.md`). Read `references/harvest-next-work.md` for the validation function and write procedure. Drop invalid items; do NOT block the entire harvest.

7. **Write to next-work.jsonl** (canonical path: `.agents/rpi/next-work.jsonl`). Read `references/harvest-next-work.md` for the write procedure (target_repo assignment, JSONL format, required fields).

8. **Do NOT auto-create bd issues.** Report the items and suggest: "Run `/rpi --spawn-next` to create an epic from these items."

If no actionable items found, write: "No follow-up items identified. Flywheel stable."

---

## Integration with Workflow

```
/plan epic-123
    │
    ▼
/pre-mortem (council on plan)
    │
    ▼
/implement
    │
    ▼
/vibe (council on code)
    │
    ▼
Ship it
    │
    ▼
/post-mortem              ← You are here
    │
    ├── Council validates implementation
    ├── Retro extracts learnings
    ├── Synthesize process improvements
    └── Suggest next /rpi ──────────┐
                                    │
    ┌───────────────────────────────┘
    │  (flywheel: learnings become next work)
    ▼
/rpi "<highest-priority enhancement>"
```

---

## Examples

### Wrap Up Recent Work

**User says:** `/post-mortem`

**What happens:**
1. Agent scans recent commits (last 7 days)
2. Runs `/council --deep --preset=retrospective validate recent`
3. 3 judges (plan-compliance, tech-debt, learnings) review
4. Runs `/retro` to extract learnings
5. Synthesizes process improvement proposals
6. Harvests next-work items to `.agents/rpi/next-work.jsonl`
7. Feeds learnings to knowledge flywheel via `ao forge`

**Result:** Post-mortem report with learnings, tech debt identified, and suggested next `/rpi` command.

### Wrap Up Specific Epic

**User says:** `/post-mortem ag-5k2`

**What happens:**
1. Agent loads original plan from `bd show ag-5k2`
2. Council reviews implementation vs plan
3. Retro captures what went well and what was hard
4. Process improvements identified (e.g., "Add pre-commit lint check")
5. Next-work items harvested and written to JSONL

**Result:** Epic-specific post-mortem with 3 harvested follow-up items (2 tech-debt, 1 process-improvement).

### Cross-Vendor Review

**User says:** `/post-mortem --mixed ag-3b7`

**What happens:**
1. Agent runs 3 Claude + 3 Codex judges
2. Cross-vendor perspectives catch edge cases
3. Verdict: WARN (missing error handling in 2 files)
4. Harvests 1 tech-debt item

**Result:** Higher confidence validation with cross-vendor review before closing epic.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Council times out | Epic too large or too many files changed | Split post-mortem into smaller reviews or increase timeout |
| Retro fails but council succeeds | `/retro` skill unavailable or errors | Post-mortem proceeds with "⚠️ SKIPPED: retro unavailable" — council findings still captured |
| No next-work items harvested | Council found no tech debt or improvements | Flywheel stable — write entry with empty items array to next-work.jsonl |
| Schema validation failed | Harvested item missing required field or has invalid enum value | Drop invalid item, log error, proceed with valid items only |
| Checkpoint-policy preflight blocks | Prior FAIL verdict in ratchet chain without fix | Resolve prior failure (fix + re-vibe) or skip checkpoint-policy via `--skip-checkpoint-policy` |
| Metadata verification fails | Plan vs actual files mismatch or missing cross-references | Include failures in council packet as `context.metadata_failures` — judges assess severity |

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/retro/SKILL.md` — Extract learnings
- `skills/vibe/SKILL.md` — Council validates code (`/vibe` after coding)
- `skills/pre-mortem/SKILL.md` — Council validates plans (before implementation)


## Reference Documents

- [references/harvest-next-work.md](references/harvest-next-work.md)
- [references/learning-templates.md](references/learning-templates.md)
- [references/plan-compliance-checklist.md](references/plan-compliance-checklist.md)
- [references/closure-integrity-audit.md](references/closure-integrity-audit.md)
- [references/security-patterns.md](references/security-patterns.md)
