---
name: post-mortem
description: 'Wrap up completed work. Council validates the implementation, then extract learnings. Triggers: "post-mortem", "wrap up", "close epic", "what did we learn".'
---


# Post-Mortem Skill

> **Purpose:** Wrap up completed work — validate it shipped correctly and extract learnings.

Two steps:
1. `$council validate` — Did we implement it correctly?
2. `$retro` — What did we learn?

---

## Quick Start

```bash
$post-mortem                    # wraps up recent work
$post-mortem epic-123           # wraps up specific epic
$post-mortem --quick recent     # fast inline wrap-up, no spawning
$post-mortem --deep recent      # thorough council review
$post-mortem --mixed epic-123   # cross-vendor (Claude + Codex)
$post-mortem --explorers=2 epic-123  # deep investigation before judging
$post-mortem --debate epic-123      # two-round adversarial review
$post-mortem --skip-checkpoint-policy epic-123  # skip ratchet chain validation
```

---

## Execution Steps

### Pre-Flight Checks

Before proceeding, verify:
1. **Git repo exists:** `git rev-parse --git-dir 2>/dev/null` — if not, error: "Not in a git repository"
2. **Work was done:** `git log --oneline -1 2>/dev/null` — if empty, error: "No commits found. Run $implement first."
3. **Epic context:** If epic ID provided, verify it has closed children. If 0 closed children, error: "No completed work to review."

### Step 0.4: Reference Existence Preflight (MANDATORY)

Before Step 0.5 and Step 2.5, verify required reference docs exist:

```bash
REQUIRED_REFS=(
  "skills/post-mortem/references/checkpoint-policy.md"
  "skills/post-mortem/references/metadata-verification.md"
)

missing=0
for ref in "${REQUIRED_REFS[@]}"; do
  if [ ! -f "$ref" ]; then
    echo "WARN: missing required reference: $ref"
    missing=$((missing + 1))
  fi
done

if [ "$missing" -gt 0 ]; then
  echo "WARN: post-mortem reference preflight incomplete (${missing} missing)."
  echo "Add these as checkpoint warnings in council context and proceed only if intentionally deferred."
fi
```

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

### Step 2.5: Pre-Council Metadata Verification (MANDATORY)

Read `references/metadata-verification.md` for the full verification procedure. Mechanically checks: plan vs actual files, file existence in commits, cross-references in docs, and ASCII diagram integrity. Failures are included in the council packet as `context.metadata_failures`.

### Step 3: Council Validates the Work

Run `$council` with the **retrospective** preset and always 3 judges:

```
$council --deep --preset=retrospective validate <epic-or-recent>
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
$council --quick validate <epic-or-recent>
```
Single-agent structured review. Fast wrap-up without spawning.

**With debate mode:**
```
$post-mortem --debate epic-123
```
Enables adversarial two-round review for post-implementation validation. Use for high-stakes shipped work where missed findings have production consequences. See `$council` docs for full --debate details.

**Advanced options (passed through to council):**
- `--mixed` — Cross-vendor (Claude + Codex) with retrospective perspectives
- `--preset=<name>` — Override with different personas (e.g., `--preset=ops` for production readiness)
- `--explorers=N` — Each judge spawns N explorers to investigate the implementation deeply before judging
- `--debate` — Two-round adversarial review (judges critique each other's findings before final verdict)

### Step 4: Extract Learnings

Run `$retro` to capture what we learned:

```
$retro <epic-or-recent>
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
# Post-Mortem: <Epic/Topic>

**Date:** YYYY-MM-DD
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

## Learnings (from $retro)

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

### Recommended Next $rpi
$rpi "<highest-value improvement>"

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
- Include at least **5** improvements total.
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

### Recommended Next $rpi
$rpi "<highest-value improvement>"
```

### Step 5.6: Prior-Findings Resolution Tracking (MANDATORY)

After Step 5.5, compute and include prior-findings resolution tracking from `.agents/rpi/next-work.jsonl`:

```bash
NEXT_WORK=".agents/rpi/next-work.jsonl"

if [ -f "$NEXT_WORK" ]; then
  totals=$(jq -Rs '
    split("\n")
    | map(select(length>0) | fromjson)
    | reduce .[] as $e (
        {entries:0,total:0,resolved:0};
        .entries += 1
        | .total += ($e.items | length)
        | .resolved += (if ($e.consumed // false) then ($e.items | length) else 0 end)
      )
    | .unresolved = (.total - .resolved)
    | .rate = (if .total > 0 then ((.resolved * 10000 / .total) | round / 100) else 0 end)
  ' "$NEXT_WORK")

  per_source=$(jq -Rs '
    split("\n")
    | map(select(length>0) | fromjson)
    | map({
        source_epic,
        total: (.items | length),
        resolved: (if (.consumed // false) then (.items | length) else 0 end)
      })
    | group_by(.source_epic)
    | map({
        source_epic: .[0].source_epic,
        total: (map(.total) | add),
        resolved: (map(.resolved) | add),
        unresolved: ((map(.total) | add) - (map(.resolved) | add)),
        rate: (if (map(.total) | add) > 0
          then (((map(.resolved) | add) * 10000 / (map(.total) | add)) | round / 100)
          else 0 end)
      })
  ' "$NEXT_WORK")

  echo "Prior findings totals: $totals"
  echo "Prior findings by source epic: $per_source"
else
  echo "No next-work.jsonl found; resolution tracking unavailable."
fi
```

Write the totals and per-source rows into `## Prior Findings Resolution Tracking` in the post-mortem report.

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
else
  # Learnings are already in .agents/learnings/ from $retro (Step 4).
  # Without ao CLI, grep-based search in $research, $knowledge, and $inject
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
6. **Suggested next `$rpi` command** (ALWAYS — this is how the flywheel spins itself)
7. Top proactive improvements (top 3), including one quick win

**The next `$rpi` suggestion is MANDATORY, not opt-in.** After every post-mortem, present the highest-severity harvested item as a ready-to-copy command:

```markdown
## Flywheel: Next Cycle

Based on this post-mortem, the highest-priority follow-up is:

> **<title>** (<type>, <severity>)
> <1-line description>

Ready to run:
```
$rpi "<title>"
```

Or see all N harvested items in `.agents/rpi/next-work.jsonl`.
```

If no items were harvested, write: "Flywheel stable — no follow-up items identified."

### Step 8: Harvest Next Work

Scan the council report and retro for actionable follow-up items:

1. **Council findings:** Extract tech debt, warnings, and improvement suggestions from the council report (items with severity "significant" or "critical" that weren't addressed in this epic)
2. **Retro patterns:** Extract recurring patterns from retro learnings that warrant dedicated RPIs (items from "Do Differently Next Time" and "Anti-Patterns to Avoid")
3. **Process improvements:** Include all items from Step 5.5 (type: `process-improvement`). These are the flywheel's growth vector — each cycle makes development more effective.
4. **Write `## Next Work` section** to the post-mortem report:

```markdown
## Next Work

| # | Title | Type | Severity | Source | Target Repo |
|---|-------|------|----------|--------|-------------|
| 1 | <title> | tech-debt / improvement / pattern-fix / process-improvement | high / medium / low | council-finding / retro-learning / retro-pattern | <repo-name or *> |
```

5. **SCHEMA VALIDATION (MANDATORY):** Before writing, validate each harvested item against the schema contract (`.agents/rpi/next-work.schema.md`):

```bash
validate_next_work_item() {
  local item="$1"
  local title=$(echo "$item" | jq -r '.title // empty')
  local type=$(echo "$item" | jq -r '.type // empty')
  local severity=$(echo "$item" | jq -r '.severity // empty')
  local source=$(echo "$item" | jq -r '.source // empty')
  local description=$(echo "$item" | jq -r '.description // empty')
  local target_repo=$(echo "$item" | jq -r '.target_repo // empty')

  # Required fields
  if [ -z "$title" ] || [ -z "$description" ]; then
    echo "SCHEMA VALIDATION FAILED: missing title or description for item"
    return 1
  fi

  # target_repo required (v1.2)
  if [ -z "$target_repo" ]; then
    echo "SCHEMA VALIDATION FAILED: missing target_repo for item '$title'"
    return 1
  fi

  # Type enum validation
  case "$type" in
    tech-debt|improvement|pattern-fix|process-improvement) ;;
    *) echo "SCHEMA VALIDATION FAILED: invalid type '$type' for item '$title'"; return 1 ;;
  esac

  # Severity enum validation
  case "$severity" in
    high|medium|low) ;;
    *) echo "SCHEMA VALIDATION FAILED: invalid severity '$severity' for item '$title'"; return 1 ;;
  esac

  # Source enum validation
  case "$source" in
    council-finding|retro-learning|retro-pattern) ;;
    *) echo "SCHEMA VALIDATION FAILED: invalid source '$source' for item '$title'"; return 1 ;;
  esac

  return 0
}

# Validate each item; drop invalid items (do NOT block the entire harvest)
VALID_ITEMS=()
INVALID_COUNT=0
for item in "${HARVESTED_ITEMS[@]}"; do
  if validate_next_work_item "$item"; then
    VALID_ITEMS+=("$item")
  else
    INVALID_COUNT=$((INVALID_COUNT + 1))
  fi
done
echo "Schema validation: ${#VALID_ITEMS[@]}/$((${#VALID_ITEMS[@]} + INVALID_COUNT)) items passed"
```

6. **Write to next-work.jsonl** (canonical path: `.agents/rpi/next-work.jsonl`):

```bash
mkdir -p .agents/rpi

# Resolve current repo name for target_repo default
CURRENT_REPO=$(bd config --get prefix 2>/dev/null \
  || basename "$(git remote get-url origin 2>/dev/null)" .git 2>/dev/null \
  || basename "$(pwd)")

# Assign target_repo to each validated item (v1.2):
#   process-improvement → "*" (applies across all repos)
#   all other types     → CURRENT_REPO (scoped to this repo)
for i in "${!VALID_ITEMS[@]}"; do
  item="${VALID_ITEMS[$i]}"
  item_type=$(echo "$item" | jq -r '.type')
  if [ "$item_type" = "process-improvement" ]; then
    VALID_ITEMS[$i]=$(echo "$item" | jq -c '.target_repo = "*"')
  else
    VALID_ITEMS[$i]=$(echo "$item" | jq -c --arg repo "$CURRENT_REPO" '.target_repo = $repo')
  fi
done

# Append one entry per epic (schema v1.2: .agents/rpi/next-work.schema.md)
# Only include VALID_ITEMS that passed schema validation
# Each item: {title, type, severity, source, description, evidence, target_repo}
# Entry fields: source_epic, timestamp, items[], consumed: false
```

Use the Write tool to append a single JSON line to `.agents/rpi/next-work.jsonl` with:
- `source_epic`: the epic ID being post-mortemed
- `timestamp`: current ISO-8601
- `items`: array of harvested items (min 0 — if nothing found, write entry with empty items array)
- `consumed`: false, `consumed_by`: null, `consumed_at`: null

7. **Do NOT auto-create bd issues.** Report the items and suggest: "Run `$rpi --spawn-next` to create an epic from these items."

If no actionable items found, write: "No follow-up items identified. Flywheel stable."

---

## Integration with Workflow

```
$plan epic-123
    │
    ▼
$pre-mortem (council on plan)
    │
    ▼
$implement
    │
    ▼
$vibe (council on code)
    │
    ▼
Ship it
    │
    ▼
$post-mortem              ← You are here
    │
    ├── Council validates implementation
    ├── Retro extracts learnings
    ├── Synthesize process improvements
    └── Suggest next $rpi ──────────┐
                                    │
    ┌───────────────────────────────┘
    │  (flywheel: learnings become next work)
    ▼
$rpi "<highest-priority enhancement>"
```

---

## Examples

### Wrap Up Recent Work

**User says:** `$post-mortem`

**What happens:**
1. Agent scans recent commits (last 7 days)
2. Runs `$council --deep --preset=retrospective validate recent`
3. 3 judges (plan-compliance, tech-debt, learnings) review
4. Runs `$retro` to extract learnings
5. Synthesizes process improvement proposals
6. Harvests next-work items to `.agents/rpi/next-work.jsonl`
7. Feeds learnings to knowledge flywheel via `ao forge`

**Result:** Post-mortem report with learnings, tech debt identified, and suggested next `$rpi` command.

### Wrap Up Specific Epic

**User says:** `$post-mortem ag-5k2`

**What happens:**
1. Agent loads original plan from `bd show ag-5k2`
2. Council reviews implementation vs plan
3. Retro captures what went well and what was hard
4. Process improvements identified (e.g., "Add pre-commit lint check")
5. Next-work items harvested and written to JSONL

**Result:** Epic-specific post-mortem with 3 harvested follow-up items (2 tech-debt, 1 process-improvement).

### Cross-Vendor Review

**User says:** `$post-mortem --mixed ag-3b7`

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
| Retro fails but council succeeds | `$retro` skill unavailable or errors | Post-mortem proceeds with "⚠️ SKIPPED: retro unavailable" — council findings still captured |
| No next-work items harvested | Council found no tech debt or improvements | Flywheel stable — write entry with empty items array to next-work.jsonl |
| Schema validation failed | Harvested item missing required field or has invalid enum value | Drop invalid item, log error, proceed with valid items only |
| Checkpoint-policy preflight blocks | Prior FAIL verdict in ratchet chain without fix | Resolve prior failure (fix + re-vibe) or skip checkpoint-policy via `--skip-checkpoint-policy` |
| Metadata verification fails | Plan vs actual files mismatch or missing cross-references | Include failures in council packet as `context.metadata_failures` — judges assess severity |

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/retro/SKILL.md` — Extract learnings
- `skills/vibe/SKILL.md` — Council validates code (`$vibe` after coding)
- `skills/pre-mortem/SKILL.md` — Council validates plans (before implementation)


## Reference Documents

- [references/learning-templates.md](references/learning-templates.md)
- [references/plan-compliance-checklist.md](references/plan-compliance-checklist.md)
- [references/security-patterns.md](references/security-patterns.md)

---

## References

### checkpoint-policy.md

# Checkpoint-Policy Preflight

Validates the ratchet chain before running the post-mortem council. Ensures prior phases completed successfully and all artifacts are available.

## 1. Guard Clause

```bash
# Skip if --skip-checkpoint-policy flag is set
# Skip if chain file doesn't exist (standalone post-mortem is valid)
CHAIN_FILE=".agents/ao/chain.jsonl"
if [ ! -f "$CHAIN_FILE" ]; then
  echo "Checkpoint policy: SKIP (no chain file — standalone post-mortem)"
  # Continue to Step 1 without blocking
fi
```

## 2. Ratchet Chain Policy Checks

Load `chain.jsonl` and verify prior phases are locked:

1. **Parse entries using dual-schema:** Check for BOTH `gate` (old schema) and `step` (new schema) field names. Each line is a JSON object — use `jq` to extract either field:
   ```bash
   jq -r '.gate // .step' "$CHAIN_FILE"
   ```
2. **Required phases:** For each of `research`, `plan`, `pre-mortem`, `implement`/`crank`, `vibe`:
   - Check that at least one entry exists with `locked: true` or `status: "locked"`
   - Missing phases: **WARN** (logged, not blocking)
3. **Council verdict validation:** For `pre-mortem` and `vibe` entries:
   - Find the corresponding council report in `.agents/council/` (match by date and type in filename)
   - Read the `## Council Verdict:` line
   - If verdict is `FAIL`: **BLOCK** — do not proceed
4. **Cycle guard:** If `cycle > 1` in any entry, verify `parent_epic` is non-empty. Empty parent on multi-cycle: **WARN**

## 3. Artifact Availability Checks

For each chain entry's `output` path:

1. If output starts with `.agents/` or contains `/` (is a file path): verify file exists on disk
2. If output matches `epic:<id>` or `issue:<id>`: skip (not a file reference)
3. If output is `inline-pass`: skip (no artifact expected)
4. Missing artifacts: **WARN**

```bash
while IFS= read -r line; do
  output=$(echo "$line" | jq -r '.output // .artifact // empty')
  case "$output" in
    epic:*|issue:*|inline-pass|"") continue ;;
    *)
      if [[ "$output" == *"/"* ]] && [ ! -f "$output" ]; then
        echo "WARN: artifact missing: $output"
      fi
      ;;
  esac
done < "$CHAIN_FILE"
```

## 4. Idempotency Check

If an epic ID is provided, check `.agents/rpi/next-work.jsonl` for an existing entry with the same `source_epic`:

1. If found and `consumed: false`: **WARN** "Post-mortem already harvested for this epic. Re-running will create duplicate entries."
2. If found and `consumed: true`: **INFO** "Prior post-mortem consumed by `<consumed_by>`. Fresh harvest will be appended."
3. If not found: no action needed

```bash
NEXT_WORK=".agents/rpi/next-work.jsonl"
if [ -n "$EPIC_ID" ] && [ -f "$NEXT_WORK" ]; then
  existing=$(grep "\"source_epic\":\"$EPIC_ID\"" "$NEXT_WORK" | tail -1)
  if [ -n "$existing" ]; then
    consumed=$(echo "$existing" | jq -r '.consumed')
    if [ "$consumed" = "false" ]; then
      echo "WARN: Post-mortem already harvested for $EPIC_ID. Re-running will create duplicate entries."
    else
      consumed_by=$(echo "$existing" | jq -r '.consumed_by')
      echo "INFO: Prior post-mortem consumed by $consumed_by. Fresh harvest will be appended."
    fi
  fi
fi
```

## 5. Summary Report Table

Print the preflight summary before proceeding:

```
| Check              | Status    | Detail                    |
|--------------------|-----------|---------------------------|
| Chain loaded       | PASS/SKIP | path or "not found"       |
| Prior phases locked| PASS/WARN | list any unlocked         |
| No FAIL verdicts   | PASS/BLOCK| list any FAILed           |
| Artifacts exist    | PASS/WARN | list any missing          |
| Idempotency        | PASS/WARN/INFO | dedup status         |
```

## 6. Blocking Behavior

- **BLOCK** only on FAIL verdicts in prior gates (pre-mortem or vibe). If any check is BLOCK: stop post-mortem and report:
  > "Checkpoint-policy BLOCKED: `<reason>`. Fix the failing gate and re-run."
- **WARN** on everything else (missing phases, missing artifacts, idempotency). Warnings are logged, included in the council packet as `context.checkpoint_warnings`, and execution proceeds.
- **INFO** is purely informational — no action needed.

### learning-templates.md

# Learning Templates for Post-Mortem

Templates for extracting and documenting learnings during Phase 4.

---

## Learning Artifact Template

Write to: `.agents/learnings/YYYY-MM-DD-{topic}.md`

```markdown
# Learning: [Concise Title]

**Date:** YYYY-MM-DD
**Epic:** <epic-id>
**Tags:** [learning, topic1, topic2]
**Verified:** yes/no

---

## Context

What were we trying to accomplish? What was the situation?

- **Goal:** [What we were building]
- **Approach:** [How we approached it]
- **Environment:** [Relevant context]

---

## What We Learned

Concrete insight or pattern discovered. Be specific.

[2-3 sentences describing the learning]

### Key Insight

> [One-sentence summary that could be quoted]

---

## Verification Status

**Verified:** yes / no

**Verification Method:** [How this was verified]
- Tool-verified: [tool name and output file]
- Multi-observation: [list of files/commits where pattern appeared]
- Production-confirmed: [incident or metric that confirmed]
- Single-observation: [needs more data to confirm]

**NOTE:** Do NOT use confidence scores (0.92, 0.91). Use "verified: yes/no" with method.

---

## Evidence

| Source | Detail | Relevance |
|--------|--------|-----------|
| Commit | abc123 | [What it shows] |
| Issue | <issue-id> | [What happened] |
| Discussion | [link/reference] | [Key point] |

---

## Application

How to apply this learning in the future:

1. **When to use:** [Trigger conditions]
2. **How to apply:** [Concrete steps]
3. **What to avoid:** [Anti-pattern]

---

## Discovery Provenance

| Insight | Source Type | Source Detail |
|---------|-------------|---------------|
| [Learning point] | [grep/code-map/etc] | [file:line or query] |

---

## Related

- Previous learnings: [links]
- Related patterns: [links]
- Documentation: [links]
```

---

## Pattern Artifact Template

Write to: `.agents/patterns/{pattern-name}.md`

```markdown
# Pattern: [Pattern Name]

**Date:** YYYY-MM-DD
**Discovered In:** <epic-id>
**Tags:** [pattern, category, language]
**Maturity:** experimental | validated | established

---

## Summary

[One paragraph describing what this pattern does and when to use it]

---

## Problem

What problem does this pattern solve?

- [Pain point 1]
- [Pain point 2]

---

## Solution

### Structure

[Describe the pattern structure]

### Code Example

```language
// Example implementation
```

### When to Use

- [Condition 1]
- [Condition 2]

### When NOT to Use

- [Counter-indication 1]
- [Counter-indication 2]

---

## Trade-offs

| Pro | Con |
|-----|-----|
| [Benefit] | [Drawback] |

---

## Real-World Usage

### From Epic <epic-id>

[How we used this pattern in the epic]

**Before:**
```language
// Old approach
```

**After:**
```language
// Pattern applied
```

**Result:** [What improved]

---

## Related Patterns

- [Related pattern 1]: [How they differ]
- [Related pattern 2]: [How they complement]

---

## References

- [External link 1]
- [Internal doc 1]
```

---

## Memory Storage Template

For MCP `memory_store`:

```python
mcp__ai-platform__memory_store(
    content="[Concise learning statement - max 200 words]",
    memory_type="fact",  # fact | preference | episode
    source=f"post-mortem:{epic_id}",
    tags=[
        f"epic:{epic_id}",
        "learning",
        "topic:specific-topic",
        "rig:rig-name",
        "verified:yes"  # or "verified:no" - never use confidence scores
    ]
    # Note: Use "verified:yes/no" tags, NOT confidence scores (0.92, 0.91).
    # Verification requires source citation or multiple observations.
)
```

### Memory Types

| Type | Use For | Example |
|------|---------|---------|
| `fact` | Learned information | "Pre-mortem simulation catches 80% of spec issues" |
| `preference` | User/project choices | "This project prefers snake_case over camelCase" |
| `episode` | Significant events | "Wave 6 timeout issue resolved by increasing limit to 900s" |

### Tag Conventions

| Tag Prefix | Purpose | Example |
|------------|---------|---------|
| `epic:` | Link to source epic | `epic:jc-9tx6` |
| `topic:` | Subject area | `topic:security` |
| `rig:` | Which rig | `rig:athena` |
| `pattern:` | If a pattern | `pattern:pre-mortem` |
| `tool:` | If about a tool | `tool:upgrade.py` |

---

## Retro Summary Template

Write to: `.agents/retros/YYYY-MM-DD-{epic}.md`

```markdown
# Retro: [Epic Title]

**Epic:** <epic-id>
**Date:** YYYY-MM-DD
**Duration:** N days
**Mode:** crew | mayor | mixed

---

## Summary

[2-3 sentence overview of what was accomplished]

---

## What Went Well

1. **[Thing 1]:** [Why it went well]
2. **[Thing 2]:** [Why it went well]
3. **[Thing 3]:** [Why it went well]

---

## What Could Improve

1. **[Thing 1]:** [What to do differently]
   - **Action:** [Specific improvement]
2. **[Thing 2]:** [What to do differently]
   - **Action:** [Specific improvement]

---

## Friction Points

| Issue | Impact | Resolution | Learning |
|-------|--------|------------|----------|
| [Problem] | [Severity] | [How fixed] | [What we learned] |

---

## Metrics

| Metric | Value | Notes |
|--------|-------|-------|
| Issues completed | N | |
| Issues blocked | M | |
| Retries | K | |
| Duration | X days | |

---

## Learnings Extracted

- [Link to learning 1]
- [Link to learning 2]

## Patterns Discovered

- [Link to pattern 1]

## Memories Stored

- [Summary of memory 1]
- [Summary of memory 2]

---

## Source Performance

| Source | Tier | Value Score | Deviation |
|--------|------|-------------|-----------|
| smart-connections | 2 | 0.85 | +0.05 |
| grep | 3 | 0.75 | +0.15 |

### Recommendations

- **PROMOTE:** [source] overperforming by X%
- **DEMOTE:** [source] underperforming by X%

---

## Process Proposals

### Proposal: [Title]
**Severity:** CRITICAL | RECOMMENDED | OPTIONAL
**Target:** [file or process]

**Problem:** [What went wrong]

**Proposed Change:** [What to do]

**Status:** pending | approved | rejected

---

## Next Time

- [ ] [Action item 1]
- [ ] [Action item 2]
```

---

## Quick Extraction Workflow

```bash
# 1. Identify learnings from commits
git log --oneline --since="7 days ago" --grep="$EPIC" | while read commit; do
    echo "Commit: $commit"
    git show $commit --stat
    echo "---"
done

# 2. Identify friction from beads
bd list --parent=$EPIC | while read issue; do
    status=$(bd show $issue | grep "Status:")
    comments=$(bd show $issue | grep -c "Comment:")
    retries=$(bd show $issue | grep -c "retry\|Retry")
    echo "$issue: status=$status, comments=$comments, retries=$retries"
done

# 3. Create artifacts
mkdir -p .agents/{learnings,patterns,retros}

# 4. Store memories
for learning in "${LEARNINGS[@]}"; do
    mcp__ai-platform__memory_store(content="$learning", ...)
done
```

---

## Quality Criteria for Learnings

A good learning:
- [ ] Is specific (not generic advice)
- [ ] Has evidence (commits, issues, discussions)
- [ ] Is actionable (can be applied)
- [ ] Has context (when it applies)
- [ ] Is findable (good tags)
- [ ] Has verification status (verified: yes/no with method, NOT confidence scores)

A good pattern:
- [ ] Solves a real problem (not hypothetical)
- [ ] Has been used (not theoretical)
- [ ] Has trade-offs documented
- [ ] Has code examples
- [ ] Explains when NOT to use it

### metadata-verification.md

# Pre-Council Metadata Verification

**Mechanically verify delivered artifacts against the plan BEFORE council. Catches metadata errors that LLMs estimate instead of measure (L19, L22, L24).**

```bash
METADATA_FAILURES=""

# 1. Plan vs actual file list — did we deliver what we said we would?
if [ -n "$PLAN_DOC" ] && [ -f "$PLAN_DOC" ]; then
  # Extract file paths mentioned in plan
  for planned_file in $(grep -oP '`([^`]+\.(go|py|ts|js|md|yaml|yml|sh))`' "$PLAN_DOC" 2>/dev/null | tr -d '`' | sort -u); do
    if [ ! -f "$planned_file" ]; then
      METADATA_FAILURES="${METADATA_FAILURES}\n- PLANNED BUT MISSING: $planned_file (in plan but not on disk)"
    fi
  done
fi

# 2. Claimed metrics vs measured — line counts, issue counts, file counts
# Check git log for claimed counts in commit messages
for commit_msg in $(git log --oneline --since="7 days ago" --format="%s" 2>/dev/null); do
  # Handled inline during council context building
  :
done

# 3. File existence — all paths in recent commits exist
for f in $(git diff --name-only HEAD~10 2>/dev/null | sort -u); do
  if [ ! -f "$f" ] && ! git log --diff-filter=D --name-only --format="" HEAD~10..HEAD | grep -q "^${f}$"; then
    METADATA_FAILURES="${METADATA_FAILURES}\n- MISSING FILE: $f (in commits but not on disk, not intentionally deleted)"
  fi
done

# 4. Cross-references in delivered docs
for f in $(git diff --name-only HEAD~10 2>/dev/null | grep -E '\.(md|txt)$'); do
  if [ -f "$f" ]; then
    for ref in $(grep -oP '\[.*?\]\(((?!http)[^)]+)\)' "$f" 2>/dev/null | grep -oP '\(([^)]+)\)' | tr -d '()'); do
      ref_dir=$(dirname "$f")
      if [ ! -f "$ref_dir/$ref" ] && [ ! -f "$ref" ]; then
        METADATA_FAILURES="${METADATA_FAILURES}\n- BROKEN LINK: $f references $ref (not found)"
      fi
    done
  fi
done

# 5. ASCII diagram verification (>3 boxes per L22)
for f in $(git diff --name-only HEAD~10 2>/dev/null | grep -E '\.(md|txt)$'); do
  if [ -f "$f" ]; then
    box_count=$(grep -cP '┌|╔|\+--' "$f" 2>/dev/null || echo 0)
    if [ "$box_count" -gt 3 ]; then
      label_count=$(grep -cP '│\s+\S' "$f" 2>/dev/null || echo 0)
      if [ "$box_count" -gt "$label_count" ]; then
        METADATA_FAILURES="${METADATA_FAILURES}\n- DIAGRAM CHECK: $f has ${box_count} boxes but only ${label_count} label lines — verify"
      fi
    fi
  fi
done

# Report
if [ -n "$METADATA_FAILURES" ]; then
  echo "METADATA VERIFICATION FAILURES:"
  echo -e "$METADATA_FAILURES"
else
  echo "Metadata verification: all checks passed"
fi
```

**If failures found:** Include them in the council packet as `context.metadata_failures`. Tag as MECHANICAL — council judges should focus on plan compliance, tech debt, and learnings instead of re-discovering broken links or wrong counts.

**If plan doc found:** Compare planned deliverables (file paths, issue counts) against actual. Mismatches become pre-loaded findings for the `plan-compliance` judge.

**Why:** Post-mortem councils were spending judge cycles on metadata errors (wrong line counts, missing files, broken cross-refs) that are mechanically verifiable. Pre-council verification frees judges to focus on structural assessment and learning extraction.

### plan-compliance-checklist.md

# Plan Compliance Checklist

Use this checklist to mechanically verify implementation against plan.

## How to Use

1. Read the plan file
2. Extract each TODO/deliverable
3. For each item, fill in the table

## Checklist Template

| # | Plan Item | Expected File | File Exists? | Implementation Matches? | Evidence |
|---|-----------|---------------|--------------|------------------------|----------|
| 1 | <copy from plan> | <path> | yes/no | yes/partial/no | file:line |
| 2 | ... | ... | ... | ... | ... |

## Verification Rules

### "File Exists?" Column
- `yes` = file exists at expected path
- `no` = file missing (GAP)

### "Implementation Matches?" Column
- `yes` = code does what plan says
- `partial` = some of it, not all
- `no` = code does something different

### "Evidence" Column
- Must include file:line reference
- If no match, explain what's there instead

## Common Gaps

| Gap Type | Example | Action |
|----------|---------|--------|
| Missing file | Expected `auth.py`, not found | Create follow-up issue |
| Partial impl | Tests exist but don't cover edge cases | Document gap |
| Scope change | Plan said X, we built Y instead | Document rationale |

## Command-Surface Parity Addendum (CLI repos)

When plan scope includes `cli/cmd/ao/*.go` changes, add this parity table:

| Command File | Tested Run-path Evidence | Intentionally Uncovered? | Follow-up Issue |
|---|---|---|---|
| cli/cmd/ao/<command>.go | TestName / file:line | yes/no | ag-xxxx (if yes) |

Rules:
- Every modified command file must have either tested run-path evidence or an intentional-uncovered entry.
- "Intentionally uncovered" requires a concrete follow-up issue ID.
- Do not close plan-compliance as PASS when command-surface rows are missing.

## Example

From a real plan:

| # | Plan Item | Expected File | File Exists? | Implementation Matches? | Evidence |
|---|-----------|---------------|--------------|------------------------|----------|
| 1 | Create toolchain-validate.sh | scripts/toolchain-validate.sh | yes | yes | scripts/toolchain-validate.sh:1-375 |
| 2 | Support --json flag | scripts/toolchain-validate.sh | yes | yes | scripts/toolchain-validate.sh:26,339 |
| 3 | Add unit tests | tests/scripts/test-toolchain-validate.sh | yes | partial | Missing exit code tests |

### security-patterns.md

# Security Patterns for Post-Mortem

Security checks to run during post-mortem Phase 3.

---

## Tool Detection

Before running security tools, check availability:

```bash
command -v gitleaks &>/dev/null && HAS_GITLEAKS=true || HAS_GITLEAKS=false
command -v semgrep &>/dev/null && HAS_SEMGREP=true || HAS_SEMGREP=false
command -v govulncheck &>/dev/null && HAS_GOVULN=true || HAS_GOVULN=false
command -v pip-audit &>/dev/null && HAS_PIPAUDIT=true || HAS_PIPAUDIT=false
```

---

## Static Analysis Tools

### Gitleaks (Secret Detection)

```bash
# Full scan
gitleaks detect --source . --verbose --report-format json --report-path reports/gitleaks.json

# Changed files only
git diff --name-only HEAD~10 | xargs gitleaks detect --source
```

**Severity:**
- API keys, passwords: CRITICAL
- Generic secrets: HIGH
- Potential false positives: MEDIUM

### Semgrep (SAST)

```bash
# OWASP patterns
semgrep --config "p/owasp-top-ten" --json -o reports/semgrep.json .

# Python-specific
semgrep --config "p/python" .

# Go-specific
semgrep --config "p/golang" .
```

### Language-Specific Vulnerability Scanners

**Python:**
```bash
pip-audit -r requirements.txt --format json -o reports/pip-audit.json
safety check -r requirements.txt --json > reports/safety.json
```

**Go:**
```bash
govulncheck -json ./... > reports/govulncheck.json
```

**JavaScript:**
```bash
npm audit --json > reports/npm-audit.json
```

---

## Pattern-Based Detection

When tools aren't available, use grep patterns:

### SEC-P01: SQL Injection

```bash
# Python
grep -rn "execute.*%s\|execute.*format\|execute.*f\"\|cursor.execute.*+" --include="*.py" .

# Go
grep -rn "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|Query.*+" --include="*.go" .
```

**Fix:** Use parameterized queries.

### SEC-P02: Command Injection

```bash
# Python
grep -rn "os.system\|subprocess.*shell=True\|eval(\|exec(" --include="*.py" .

# Go
grep -rn "exec.Command.*Shell\|os.Exec" --include="*.go" .
```

**Fix:** Use safe APIs, never shell=True with user input.

### SEC-P03: Hardcoded Secrets

```bash
# All languages
grep -rn "password.*=.*['\"].*['\"]" --include="*.py" --include="*.go" --include="*.js" .
grep -rn "api_key.*=.*['\"]" --include="*.py" --include="*.go" --include="*.js" .
grep -rn "secret.*=.*['\"].*[a-zA-Z0-9]" --include="*.py" --include="*.go" --include="*.js" .
```

**Fix:** Use environment variables or secrets management.

### SEC-P04: Insecure Deserialization

```bash
# Python
grep -rn "pickle.load\|yaml.load.*Loader" --include="*.py" .

# Go
grep -rn "json.Unmarshal.*interface{}" --include="*.go" .
```

**Fix:** Use safe loaders, validate before deserializing.

### SEC-P05: XSS Patterns

```bash
# JavaScript/TypeScript
grep -rn "innerHTML.*=\|dangerouslySetInnerHTML\|v-html" --include="*.js" --include="*.ts" --include="*.vue" .

# Python (templates)
grep -rn "mark_safe\|{{.*\|raw}}\|autoescape false" --include="*.html" --include="*.jinja" .
```

**Fix:** Use safe templating, escape output.

### SEC-P06: Path Traversal

```bash
# All languages
grep -rn "open.*%s\|open.*format\|os.path.join.*input" --include="*.py" .
grep -rn "filepath.Join.*user" --include="*.go" .
```

**Fix:** Validate and sanitize file paths.

### SEC-P07: Insecure TLS

```bash
# Python
grep -rn "verify=False\|CERT_NONE" --include="*.py" .

# Go
grep -rn "InsecureSkipVerify.*true" --include="*.go" .
```

**Fix:** Never disable certificate verification in production.

### SEC-P08: Weak Cryptography

```bash
# Python
grep -rn "MD5\|SHA1\|DES\|RC4" --include="*.py" .

# Go
grep -rn "md5.\|sha1.\|des.\|rc4." --include="*.go" .
```

**Fix:** Use strong algorithms (SHA256+, AES).

---

## OWASP Top 10 Checklist

| OWASP ID | Category | Check | Pattern |
|----------|----------|-------|---------|
| A01 | Broken Access Control | Auth checks on routes | Missing `@login_required` |
| A02 | Cryptographic Failures | Weak crypto, plaintext | SEC-P03, SEC-P08 |
| A03 | Injection | SQL, command, XSS | SEC-P01, SEC-P02, SEC-P05 |
| A04 | Insecure Design | Business logic flaws | Manual review |
| A05 | Security Misconfiguration | Debug mode, defaults | `DEBUG=True`, default passwords |
| A06 | Vulnerable Components | Old dependencies | pip-audit, npm audit |
| A07 | Auth Failures | Weak auth, session | Timing attacks, session fixation |
| A08 | Data Integrity | Deserialization, CI/CD | SEC-P04 |
| A09 | Logging Failures | Missing logs, PII in logs | grep for sensitive in logs |
| A10 | SSRF | Server requests | `requests.get(user_input)` |

---

## Expert Agent Delegation

For CRITICAL findings, spawn security expert:

```python
Task(
    subagent_type="security-expert",
    prompt=f"""Deep security review for post-mortem.

Findings to analyze:
{findings}

Changed files:
{changed_files}

Please:
1. Verify each finding is real (not false positive)
2. Assess exploitability
3. Recommend specific fixes
4. Identify any additional vulnerabilities
"""
)
```

---

## Report Format

```markdown
## Security Scan Results

**Date:** YYYY-MM-DD
**Epic:** <epic-id>
**Changed Files:** N files

### Summary

| Category | CRITICAL | HIGH | MEDIUM | LOW |
|----------|----------|------|--------|-----|
| Secrets | 0 | 0 | 0 | 0 |
| Vulnerabilities | 0 | 0 | 0 | 0 |
| Patterns | 0 | 1 | 2 | 0 |

### Findings

#### SEC-001 [HIGH] Potential SQL Injection
- **File:** services/db.py:42
- **Pattern:** `cursor.execute(f"SELECT * FROM {table}")`
- **Fix:** Use parameterized query: `cursor.execute("SELECT * FROM ?", (table,))`
- **Issue Created:** <issue-id>

### Tools Used
- gitleaks: v8.18.0
- semgrep: v1.0.0
- grep patterns: SEC-P01 through SEC-P08

### Recommendations
1. Fix HIGH findings before merge
2. Consider adding pre-commit hook for gitleaks
3. Schedule full security audit (quarterly)
```

---

## Integration with Post-Mortem

In Phase 3 of post-mortem:

```python
def run_security_scan(changed_files, epic_id):
    findings = []

    # 1. Run available tools
    if HAS_GITLEAKS:
        findings += run_gitleaks(changed_files)

    if HAS_SEMGREP:
        findings += run_semgrep(changed_files)

    # 2. Run grep patterns
    findings += run_grep_patterns(changed_files)

    # 3. Check OWASP Top 10
    findings += check_owasp(changed_files)

    # 4. Delegate CRITICAL to expert
    critical = [f for f in findings if f.severity == 'CRITICAL']
    if critical:
        expert_review = spawn_security_expert(critical, changed_files)
        findings += expert_review.additional_findings

    # 5. Create issues for HIGH+
    for finding in findings:
        if finding.severity in ['CRITICAL', 'HIGH']:
            bd_create(
                title=f"Security: {finding.title}",
                type="bug",
                priority="P1" if finding.severity == 'CRITICAL' else "P2",
                description=finding.details
            )

    return SecurityReport(findings)
```


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: post-mortem" "grep -q '^name: post-mortem' '$SKILL_DIR/SKILL.md'"
check "references/ has at least 2 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 2 ]"
check "SKILL.md mentions Step 8" "grep -q 'Step 8' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions harvest" "grep -qi 'harvest' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


