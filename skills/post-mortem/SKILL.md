---
name: post-mortem
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

## See Also

- `skills/council/SKILL.md` — Multi-model validation council
- `skills/retro/SKILL.md` — Extract learnings
- `skills/vibe/SKILL.md` — Council validates code
- `skills/pre-mortem/SKILL.md` — Council validates plans
