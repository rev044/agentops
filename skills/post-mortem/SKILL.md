---
name: post-mortem
description: 'Wrap up completed work. Council validates the implementation, then extract learnings. Triggers: "post-mortem", "wrap up", "close epic", "what did we learn".'
dependencies:
  - council  # multi-model judgment
  - retro    # extracts learnings
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
/post-mortem --deep recent      # thorough council review
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

### Step 2: Council Validates the Work

Run `/council validate` on the completed work:

```
/council validate <epic-or-recent>
```

**Council reviews:**
- Did implementation match the plan?
- Are there gaps or shortcuts taken?
- Security concerns?
- Technical debt introduced?

### Step 3: Extract Learnings

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

### Step 4: Write Post-Mortem Report

**Write to:** `.agents/post-mortems/YYYY-MM-DD-<topic>.md`

```markdown
# Post-Mortem: <Epic/Topic>

**Date:** YYYY-MM-DD
**Epic:** <epic-id or "recent">
**Duration:** <how long it took>

## Council Verdict: PASS / WARN / FAIL

| Judge | Verdict | Key Finding |
|-------|---------|-------------|
| Pragmatist | ... | ... |
| Skeptic | ... | ... |

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

### Step 5: Report to User

Tell the user:
1. Council verdict on implementation
2. Key learnings
3. Any follow-up items
4. Location of post-mortem report

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
