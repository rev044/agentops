---
name: sk-handoff
description: >
  Cross-session continuity patterns. Hand off to fresh agent sessions,
  preserve context via hooks and mail, and resume work reliably.
version: 1.0.0
triggers:
  - "handoff"
  - "session continuity"
  - "context cycling"
  - "fresh session"
  - "hand off"
  - "end session"
  - "context full"
  - "restart session"
  - "cross-session"
  - "preserve context"
  - "continue in new session"
allowed-tools: Bash, Read, Glob, Grep
---

# sk-handoff - Cross-Session Continuity Skill

Reliable handoff patterns for Gas Town agents.

## Overview

**What this skill does:** Enables seamless work continuation across agent session boundaries.

| User Says | Claude Does |
|-----------|-------------|
| "hand off to fresh session" | `gt handoff` |
| "context is full" | Save state, hook work, request handoff |
| "continue in new session" | Check hook, load state, resume |
| "send handoff notes" | `gt mail send --self -s "HANDOFF" -m "..."` |
| "refresh this crew" | `gt crew refresh <name>` |

---

## Quick Reference

```bash
# Hand off current session
gt handoff                              # Basic handoff
gt handoff -c                           # Auto-collect state
gt handoff -s "Subject" -m "Notes"      # With context

# Hand off with bead attached
gt handoff gt-abc                       # Hook bead, then restart
gt handoff gt-abc -s "Fix auth bug"     # Hook with context

# Crew refresh (restart with handoff)
gt crew refresh dave                    # Auto-generated handoff
gt crew refresh dave -m "Working on X"  # Custom message

# Self-mail handoff (manual)
gt mail send --self -s "HANDOFF: Feature X" -m "Context here"
```

---

## Core Concepts

### What Survives a Restart

| Survives | Does NOT Survive |
|----------|------------------|
| Hooked work (`gt hook`) | In-memory context |
| Beads state (issues, comments) | Todo list |
| Git commits | Uncommitted changes |
| Mail messages | Session variables |
| CLAUDE.md files | Conversation history |

**Key insight:** Anything in beads or git survives. Save state there.

### The Hook as Anchor

The hook persists across restarts. On session start:

```bash
gt hook
# → Shows: gt-abc (Your Feature)
```

If work is hooked, the Propulsion Principle applies: **RUN IT.**

### Context Thresholds

| Usage | Action |
|-------|--------|
| < 30% | Continue normally |
| 30-35% | Prepare for handoff (save state) |
| > 35% | Execute handoff immediately |

Handoff at 35%, not when context is exhausted.

---

## Operations

### 1. Basic Handoff

End the current session and restart fresh:

```bash
# Simple handoff (no context)
gt handoff

# With auto-collected state
gt handoff -c

# With custom message
gt handoff -s "Working on auth" -m "Completed: login flow. Next: logout"
```

**What happens:**
1. Mail sent to your own inbox with handoff context
2. Session terminates
3. New session starts
4. SessionStart hook reads mail, finds hooked work
5. Work continues

### 2. Handoff with Work Attachment

Hook specific work before handing off:

```bash
# Hook bead and restart
gt handoff gt-abc

# Hook with context
gt handoff gt-abc -s "Bug fix" -m "Root cause identified in auth.py:42"
```

**When to use:** To ensure the next session picks up specific work.

### 3. State Preservation

For complex orchestration (epics, waves), save state to beads:

```bash
# Build state JSON
state=$(cat <<'EOF'
{
  "wave": 2,
  "completed": ["gt-abc", "gt-def"],
  "in_progress": ["gt-ghi"],
  "convoy": "hq-cv-xyz",
  "rig": "gastown"
}
EOF
)

# Save to bead comments
bd comments add gt-epic "HANDOFF_STATE: $state"

# Hook the epic
gt hook gt-epic

# Then handoff
gt handoff
```

### 4. Crew Refresh

Cycle a crew workspace with handoff:

```bash
# Basic refresh
gt crew refresh dave

# With custom message
gt crew refresh dave -m "Working on gt-123, paused at tests"
```

**Use case:** When crew context is full but work is mid-flight.

### 5. Mail-Based Handoff

For ad-hoc instructions that don't fit a bead:

```bash
# Send to self
gt mail send --self -s "HANDOFF: Priority shift" -m "$(cat <<'EOF'
New priority from human: Focus on security fixes.
Defer feature work until security audit complete.
Check: bd list --label=security
EOF
)"

# Hook the mail
gt hook attach <mail-id>
```

---

## Resume Patterns

### On Session Start

```bash
# 1. Check hook (automatic via SessionStart)
gt hook
# → gt-abc or mail-xyz

# 2. If bead hooked
bd show gt-abc
# Read any HANDOFF_STATE comments
bd comments gt-abc | grep "HANDOFF_STATE:" | tail -1

# 3. If mail hooked
gt mail read <mail-id>
# Execute the prose instructions

# 4. Continue work
```

### Resuming Epic Orchestration

```python
def resume_epic(epic_id):
    # Read saved state
    comments = bash(f"bd comments {epic_id}")
    state_line = find_last_match(comments, "HANDOFF_STATE:")

    if not state_line:
        return start_fresh(epic_id)

    state = json.loads(state_line.split("HANDOFF_STATE:")[1])

    # Check convoy for in-progress wave
    if state.get("convoy"):
        convoy = bash(f"gt convoy status {state['convoy']}")
        if "complete" in convoy:
            state["completed"].extend(state["in_progress"])
            state["in_progress"] = []
            state["wave"] += 1

    return continue_from_wave(epic_id, state)
```

---

## Common Patterns

### Context Cycling Pattern

Proactive handoff before context fills:

```bash
# Monitor context (Claude Code shows this)
# At 35%: initiate handoff

# 1. Save state
bd comments add $epic "HANDOFF_STATE: {wave: 2, ...}"

# 2. Hook work
gt hook $epic

# 3. Handoff
gt handoff -s "Context cycling" -m "Wave 2 in progress, 3 polecats active"
```

### Emergency Handoff

When something goes wrong:

```bash
# Save available state
bd comments add $current_work "EMERGENCY_HANDOFF: $(date)"
bd comments add $current_work "Last known state: $state"

# Hook work
gt hook $current_work

# Notify human
gt mail send --human -s "EMERGENCY: Session crash" -m "Work hooked: $current_work"

# Handoff
gt handoff
```

### Planned Handoff

For long-running work that spans sessions:

```bash
# End of session checklist
[ ] Commit code: git add . && git commit -m "..."
[ ] Sync beads: bd sync
[ ] Save state: bd comments add $work "SESSION_END: ..."
[ ] Hook work: gt hook $work
[ ] Handoff: gt handoff -c

# Next session pickup
gt hook       # Find work
bd show $id   # Get context
# Continue
```

---

## Handoff Message Format

When sending handoff context, include:

```markdown
## Context
Current work and rationale.

## Completed
- Item 1
- Item 2

## In Progress
- Current item (state: ...)

## Next
What the next session should do.

## Files
Key files touched: path/to/file.py:42
```

Example:
```bash
gt handoff -s "Feature: Auth logout" -m "$(cat <<'EOF'
## Context
Implementing logout flow for ticket gt-abc.

## Completed
- Backend endpoint: src/api/auth.py:logout()
- Session cleanup: src/services/session.py

## In Progress
- Frontend button (50% done)
- File: src/components/Header.tsx:42

## Next
1. Finish logout button click handler
2. Add confirmation modal
3. Write tests

## Files
src/api/auth.py:89
src/services/session.py:156
src/components/Header.tsx:42
EOF
)"
```

---

## Troubleshooting

| Problem | Diagnosis | Solution |
|---------|-----------|----------|
| Work lost after handoff | Hook not set | Always `gt hook` before `gt handoff` |
| State missing | Comments not saved | Check `bd comments <id>` |
| New session doesn't continue | SessionStart hook failed | Run `gt prime` manually |
| Handoff mail not received | Wrong address | Use `--self` flag |
| Context still full | Summarization failed | Split work into smaller beads |

### Recovery from Failed Handoff

```bash
# Check what's actually in beads (source of truth)
bd list --status=in_progress        # What's active
bd list --status=closed             # What's done
bd show <id>                        # Full context

# Rebuild state manually
# Beads always has the real state
```

---

## References

- `references/gt-handoff.md` - Full handoff command reference
- `references/mail-to-self.md` - Self-mail handoff pattern
- `references/session-continuity.md` - What survives restart
- `references/crew-refresh.md` - Crew workspace cycling
