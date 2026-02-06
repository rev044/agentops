---
name: status
description: 'Single-screen dashboard showing current work, recent validations, flywheel health, and suggested next action. Triggers: "status", "dashboard", "what am I working on", "where was I".'
dependencies: []
---

# /status — AgentOps Dashboard

> **Purpose:** Single-screen overview of your current state. What am I working on? What happened recently? What should I do next?

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** bd, ao, gt — all optional. Shows what's available, skips what isn't.

## Execution Steps

### Step 1: Gather State

Run all checks in parallel where possible:

```bash
# Current beads work
if command -v bd &>/dev/null; then
  bd ready 2>/dev/null | head -5
  bd list --status in_progress 2>/dev/null | head -3
fi

# Recent council verdicts
ls -lt .agents/council/ 2>/dev/null | head -4

# Flywheel health
if command -v ao &>/dev/null; then
  ao flywheel status 2>/dev/null
fi

# Pending knowledge
ls .agents/knowledge/pending/ 2>/dev/null | wc -l

# Recent learnings
ls -lt .agents/learnings/ 2>/dev/null | head -3

# Git state
git log --oneline -3 2>/dev/null
git status --short 2>/dev/null | head -5

# Agent inbox
if command -v gt &>/dev/null; then
  gt mail inbox 2>/dev/null | head -5
fi
```

### Step 2: Render Dashboard

Present a compact single-screen summary:

```
=== AgentOps Status ===

CURRENT WORK
  <in-progress beads issues, or "No active work">
  <git branch + uncommitted changes count>

READY TO WORK
  <top 3 unblocked beads issues, or "No ready issues">

RECENT VALIDATIONS
  <last 3 council reports with verdict (PASS/WARN/FAIL)>
  <or "No recent validations">

KNOWLEDGE FLYWHEEL
  <ao flywheel status, or pending learnings count>
  <or "ao not installed — learnings in .agents/knowledge/pending/">

INBOX
  <pending messages count, or "No messages" or "gt not installed">

SUGGESTED NEXT ACTION
  <one concrete suggestion based on state>
```

### Step 3: Suggest Next Action

| State | Suggestion |
|-------|------------|
| In-progress issue exists | "Continue working on `<issue-id>`: `<title>`" |
| Ready issues available | "Pick up next issue: `/implement <issue-id>`" |
| Uncommitted changes | "Review changes: `/vibe recent`" |
| Recent WARN/FAIL verdict | "Address findings in `<report-path>`" |
| Pending knowledge items | "Promote learnings: `ao pool promote`" |
| Clean state | "Start with `/research` or `/plan` to find work" |
| Inbox has messages | "Check messages: `/inbox`" |

---

## See Also

- `skills/quickstart/SKILL.md` — First-time onboarding
- `skills/inbox/SKILL.md` — Agent mail monitoring
- `skills/knowledge/SKILL.md` — Query knowledge artifacts
