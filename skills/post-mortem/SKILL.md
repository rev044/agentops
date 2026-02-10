---
name: post-mortem
tier: solo
description: 'Wrap up completed work. Council validates the implementation, then extract learnings. Triggers: "post-mortem", "wrap up", "close epic", "what did we learn".'
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
```

---

## Execution Steps

### Pre-Flight Checks

Before proceeding, verify:
1. **Git repo exists:** `git rev-parse --git-dir 2>/dev/null` — if not, error: "Not in a git repository"
2. **Work was done:** `git log --oneline -1 2>/dev/null` — if empty, error: "No commits found. Run /implement first."
3. **Epic context:** If epic ID provided, verify it has closed children. If 0 closed children, error: "No completed work to review."

### Step 1: Identify Completed Work

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

### Step 5: Write Post-Mortem Report

**Write to:** `.agents/council/YYYY-MM-DD-post-mortem-<topic>.md`

```markdown
# Post-Mortem: <Epic/Topic>

**Date:** YYYY-MM-DD
**Epic:** <epic-id or "recent">
**Duration:** <how long it took>

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

## Status

[ ] CLOSED - Work complete, learnings captured
[ ] FOLLOW-UP - Issues need addressing (create new beads)
```

### Step 6: Feed the Knowledge Flywheel

Post-mortem automatically feeds learnings into the flywheel:

```bash
mkdir -p .agents/knowledge/pending

if command -v ao &>/dev/null; then
  ao forge index .agents/learnings/ 2>/dev/null
  echo "Learnings indexed in knowledge flywheel"
else
  # Retro already wrote to .agents/learnings/ — copy to pending for future import
  cp .agents/learnings/YYYY-MM-DD-*.md .agents/knowledge/pending/ 2>/dev/null
  echo "Note: Learnings saved to .agents/knowledge/pending/ (install ao for auto-indexing)"
fi
```

### Step 7: Report to User

Tell the user:
1. Council verdict on implementation
2. Key learnings
3. Any follow-up items
4. Location of post-mortem report
5. Knowledge flywheel status

### Step 8: Harvest Next Work

Scan the council report and retro for actionable follow-up items:

1. **Council findings:** Extract tech debt, warnings, and improvement suggestions from the council report (items with severity "significant" or "critical" that weren't addressed in this epic)
2. **Retro patterns:** Extract recurring patterns from retro learnings that warrant dedicated RPIs (items from "Do Differently Next Time" and "Anti-Patterns to Avoid")
3. **Write `## Next Work` section** to the post-mortem report:

```markdown
## Next Work

| # | Title | Type | Severity | Source |
|---|-------|------|----------|--------|
| 1 | <title> | tech-debt / improvement / pattern-fix | high / medium / low | council-finding / retro-learning / retro-pattern |
```

4. **Write to next-work.jsonl** (canonical path: `.agents/rpi/next-work.jsonl`):

```bash
mkdir -p .agents/rpi

# Append one entry per epic (schema: .agents/rpi/next-work.schema.md)
# Each item: {title, type, severity, source, description, evidence}
# Entry fields: source_epic, timestamp, items[], consumed: false
```

Use the Write tool to append a single JSON line to `.agents/rpi/next-work.jsonl` with:
- `source_epic`: the epic ID being post-mortemed
- `timestamp`: current ISO-8601
- `items`: array of harvested items (min 0 — if nothing found, write entry with empty items array)
- `consumed`: false, `consumed_by`: null, `consumed_at`: null

5. **Do NOT auto-create bd issues.** Report the items and suggest: "Run `/rpi --spawn-next` to create an epic from these items."

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
    └── Retro extracts learnings
```

---

## Examples

### Wrap Up Recent Work

```bash
/post-mortem
```

Validates recent commits, extracts learnings.

### Wrap Up Specific Epic

```bash
/post-mortem epic-123
```

Council reviews epic-123 implementation, retro captures learnings.

### Thorough Review

```bash
/post-mortem --deep epic-123
```

3 judges review the epic.

### Cross-Vendor Review

```bash
/post-mortem --mixed epic-123
```

3 Claude + 3 Codex agents review the epic.

---

## Relationship to Other Skills

| Skill | When | Purpose |
|-------|------|---------|
| `/pre-mortem` | Before implementation | Council validates plan |
| `/vibe` | After coding | Council validates code |
| `/post-mortem` | After shipping | Council validates + extract learnings |
| `/retro` | Anytime | Extract learnings only |

---

## Consolidation

For conflict resolution between agent findings, follow the algorithm in `.agents/specs/conflict-resolution-algorithm.md`.

---

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/retro/SKILL.md` — Extract learnings
- `skills/vibe/SKILL.md` — Council validates code
- `skills/pre-mortem/SKILL.md` — Council validates plans
- `.agents/specs/conflict-resolution-algorithm.md` — Conflict resolution algorithm
