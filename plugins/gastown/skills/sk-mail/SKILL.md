---
name: sk-mail
description: >
  Agent communication patterns. Send mail, check inbox, broadcast to workers,
  nudge stuck polecats, and escalate to humans when needed.
version: 1.0.0
triggers:
  - "send mail"
  - "check inbox"
  - "check my mail"
  - "read mail"
  - "broadcast"
  - "nudge polecat"
  - "nudge worker"
  - "escalate"
  - "escalate to human"
  - "send message"
  - "notify worker"
  - "inter-agent communication"
allowed-tools: Bash, Read, Glob, Grep
---

# sk-mail - Agent Communication Skill

Inter-agent communication patterns for Gas Town.

## Overview

**What this skill does:** Enables structured communication between agents (Mayor, Witnesses, Refineries, Polecats, Crew).

| User Says | Claude Does |
|-----------|-------------|
| "check my mail" | `gt mail inbox` |
| "send mail to X" | `gt mail send <address> -s "..." -m "..."` |
| "broadcast to workers" | `gt broadcast "message"` |
| "nudge that polecat" | `gt nudge <target> "message"` |
| "escalate to human" | `gt mail send --human -s "..." -m "..."` |

---

## Quick Reference

```bash
# Check inbox
gt mail inbox                        # Your inbox (auto-detected)
gt mail inbox mayor/                 # Mayor's inbox
gt mail inbox gastown/Toast          # Polecat's inbox

# Send mail
gt mail send mayor/ -s "Subject" -m "Message"
gt mail send gastown/refinery -s "Task" -m "Details"
gt mail send --human -s "Escalation" -m "Need human help"

# Read specific message
gt mail read <message-id>

# Broadcast to all workers
gt broadcast "Check your mail"
gt broadcast --rig gastown "Rig-specific message"

# Nudge specific polecat/agent
gt nudge gastown/Toast "Continue with your task"
gt nudge witness "Check polecat health"
```

---

## Operations

### 1. Checking Inbox

```bash
# Your inbox (context auto-detected)
gt mail inbox

# Filter to unread only
gt mail inbox -u

# Specific agent's inbox
gt mail inbox mayor/
gt mail inbox gastown/witness
gt mail inbox gastown/Toast
```

**Output format:**
```
ID           FROM              SUBJECT                    DATE
msg-abc123   gastown/Toast     Work complete              2026-01-08 10:30
msg-def456   gastown/refinery  Merge ready                2026-01-08 10:15
```

### 2. Sending Mail

**Address formats:**
| Pattern | Example | Destination |
|---------|---------|-------------|
| `mayor/` | `gt mail send mayor/ ...` | Mayor inbox |
| `<rig>/witness` | `gt mail send gastown/witness ...` | Rig's Witness |
| `<rig>/refinery` | `gt mail send gastown/refinery ...` | Rig's Refinery |
| `<rig>/<polecat>` | `gt mail send gastown/Toast ...` | Specific polecat |
| `<rig>/crew/<name>` | `gt mail send gastown/crew/max ...` | Crew worker |
| `--human` | `gt mail send --human ...` | Human overseer |
| `list:<name>` | `gt mail send list:oncall ...` | Mailing list |

**Message types:**
```bash
# Notification (default) - informational
gt mail send mayor/ -s "Status" -m "Work complete"

# Task - requires processing
gt mail send gastown/Toast -s "Task" -m "Fix bug" --type task

# Scavenge - first-come optional work
gt mail send gastown/ -s "Available" -m "Bug fix available" --type scavenge

# Urgent (priority 0)
gt mail send mayor/ -s "Critical" -m "Blocker found" --urgent
```

**With notification (tmux nudge):**
```bash
gt mail send gastown/Toast -s "Action needed" -m "Review PR" --notify
```

### 3. Reading Messages

```bash
# Read specific message
gt mail read msg-abc123

# Reply to message
gt mail reply msg-abc123 -m "Acknowledged, working on it"
```

### 4. Broadcasting

Send to all active workers:

```bash
# All workers in town
gt broadcast "New priority work available"

# Workers in specific rig only
gt broadcast --rig gastown "Rig-wide announcement"

# Include infrastructure agents (mayor, witness, refinery)
gt broadcast --all "System maintenance in 5 minutes"

# Preview without sending
gt broadcast --dry-run "Test message"
```

### 5. Nudging

Direct message to agent's tmux session:

```bash
# Nudge polecat to continue
gt nudge gastown/Toast "Continue with your task"

# Nudge witness
gt nudge witness "Check polecat health"

# Nudge with message flag
gt nudge gastown/alpha -m "What's your status?"

# Nudge channel (all members)
gt nudge channel:workers "Priority work available"

# Force (override DND)
gt nudge gastown/Toast "Urgent" --force
```

**When to nudge vs mail:**
| Use Nudge | Use Mail |
|-----------|----------|
| Polecat seems stuck | Formal task assignment |
| Quick "continue" signal | Work handoff |
| Immediate attention needed | Async communication |
| Status check | Record of request |

### 6. Escalation

Escalate to human when:
- Blockers require human decision
- Repeated failures after retries
- Security or safety concerns
- Out-of-scope requests

```bash
# Escalate to human
gt mail send --human -s "ESCALATION: Blocker on gt-abc" -m "$(cat <<'EOF'
Issue: gt-abc
Problem: Cannot proceed - missing API credentials
Attempted: Checked env vars, config files, secrets
Need: Human to provide or configure credentials
EOF
)"
```

---

## Common Patterns

### Handoff Between Sessions

```bash
# Send handoff to self for next session
gt mail send --self -s "HANDOFF: Feature X" -m "$(cat <<'EOF'
Context: Implementing feature X
Completed: Database schema, API endpoints
Next: Frontend components
Files: src/api/feature.py, src/models/feature.py
EOF
)"
```

### Work Dispatch Pattern

```bash
# 1. Create work item
bd create --title="Fix bug Y" --type=bug

# 2. Send as task to polecat
gt mail send gastown/Toast -s "New task" -m "Fix gt-abc" --type task --notify
```

### Blocker Escalation Pattern

```bash
# Detect blocker → escalate
if bd show gt-abc | grep -q "BLOCKER:"; then
    reason=$(bd comments gt-abc | grep "BLOCKER:" | head -1)
    gt mail send --human -s "BLOCKER: gt-abc" -m "$reason" --urgent
fi
```

---

## Troubleshooting

| Problem | Diagnosis | Solution |
|---------|-----------|----------|
| Mail not received | Check address format | Verify with `gt mail inbox <address>` |
| Nudge not delivered | DND enabled | Use `--force` flag |
| Broadcast missed agents | Agents not active | Check `gt polecat list` |
| Can't find message | Check all inboxes | `gt mail search "keyword"` |

---

## References

- `references/gt-mail.md` - Full mail command reference
- `references/gt-broadcast.md` - Broadcast patterns
- `references/gt-nudge.md` - Nudge command details
- `references/escalation.md` - When and how to escalate
