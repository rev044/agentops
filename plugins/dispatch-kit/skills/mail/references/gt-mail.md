# Mail Commands Reference

## Overview

`gt mail` provides asynchronous communication between Gas Town agents. Messages are stored as beads with type=message.

## Architecture

```
Town (.beads/)
├── Mayor Inbox (mayor/)
└── Rig Mailboxes
    ├── <rig>/witness
    ├── <rig>/refinery
    ├── <rig>/<polecat>
    └── <rig>/crew/<name>
```

## Commands

### inbox

Check messages in an inbox.

```bash
# Current context (auto-detected)
gt mail inbox

# Specific inbox
gt mail inbox mayor/
gt mail inbox gastown/Toast
gt mail inbox gastown/witness

# Unread only
gt mail inbox -u
gt mail inbox --unread

# JSON output
gt mail inbox --json
```

### send

Send a message to an agent.

```bash
gt mail send <address> -s "<subject>" -m "<message>" [flags]
```

**Addresses:**

| Pattern | Example | Description |
|---------|---------|-------------|
| `mayor/` | Mayor inbox | Town coordinator |
| `<rig>/witness` | Rig observer | Per-rig lifecycle manager |
| `<rig>/refinery` | Merge queue | Per-rig merge processor |
| `<rig>/<polecat>` | Worker | Specific polecat |
| `<rig>/crew/<name>` | Crew member | Named crew worker |
| `<rig>/` | Rig broadcast | All agents in rig |
| `list:<name>` | Mailing list | Defined in messaging.json |
| `--human` | Human overseer | Escalation target |
| `--self` | Self (auto-detect) | Handoff to next session |

**Flags:**

| Flag | Description |
|------|-------------|
| `-s, --subject` | Message subject (required) |
| `-m, --message` | Message body |
| `--type` | Message type: task, scavenge, notification (default), reply |
| `--priority` | 0=urgent, 1=high, 2=normal (default), 3=low, 4=backlog |
| `--urgent` | Shortcut for --priority 0 |
| `--notify` | Send tmux nudge to recipient |
| `--cc` | CC recipients (can repeat) |
| `--reply-to` | Message ID being replied to |
| `--permanent` | Non-ephemeral (syncs to remote) |
| `--pinned` | Pin message (persists for handoff) |
| `--wisp` | Ephemeral message (default) |

**Examples:**

```bash
# Simple notification
gt mail send mayor/ -s "Status update" -m "Work complete"

# Task with notification
gt mail send gastown/Toast -s "New task" -m "Fix gt-abc" --type task --notify

# Urgent escalation
gt mail send --human -s "BLOCKER" -m "Need credentials" --urgent

# Handoff to self
gt mail send --self -s "HANDOFF" -m "Continue feature X" --pinned

# Reply to message
gt mail send mayor/ -s "Re: Status" -m "Acknowledged" --reply-to msg-abc123

# CC recipients
gt mail send gastown/Toast -s "Update" -m "FYI" --cc overseer --cc gastown/witness
```

### read

Read a specific message.

```bash
gt mail read <message-id>
gt mail read msg-abc123

# JSON output
gt mail read msg-abc123 --json
```

### reply

Reply to a message.

```bash
gt mail reply <message-id> -m "<reply>"
gt mail reply msg-abc123 -m "Acknowledged, working on it"
```

### thread

View a message thread.

```bash
gt mail thread <message-id>
```

### search

Search messages by content.

```bash
gt mail search "<query>"
gt mail search "BLOCKER"
gt mail search "feature X"
```

### mark

Mark messages read/unread.

```bash
gt mail mark <message-id> --read
gt mail mark <message-id> --unread
```

### archive

Archive messages.

```bash
gt mail archive <message-id>
```

### delete

Delete a message.

```bash
gt mail delete <message-id>
```

### clear

Clear all messages from inbox.

```bash
gt mail clear
```

### check

Check for new mail (used by hooks).

```bash
gt mail check
```

### peek

Preview first unread message.

```bash
gt mail peek
```

### claim / release

Queue message management.

```bash
# Claim message from queue
gt mail claim <message-id>

# Release claimed message
gt mail release <message-id>
```

### announces

List or read announce channels.

```bash
gt mail announces
gt mail announces --channel <name>
```

## Message Types

| Type | Purpose | Use Case |
|------|---------|----------|
| `notification` | Informational | Status updates, FYI messages |
| `task` | Required processing | Work assignments |
| `scavenge` | Optional first-come | Available work |
| `reply` | Response to message | Threaded replies |

## Priority Levels

| Priority | Name | Use Case |
|----------|------|----------|
| 0 | Urgent/Critical | Blockers, security issues |
| 1 | High | Important but not critical |
| 2 | Normal (default) | Standard messages |
| 3 | Low | When convenient |
| 4 | Backlog | Eventually |

## Patterns

### Status Reporting

```bash
# Polecat reports completion to Mayor
gt mail send mayor/ -s "Work complete: gt-abc" -m "$(cat <<'EOF'
Issue: gt-abc
Status: CLOSED
Summary: Fixed authentication bug
Files: src/auth.py, tests/test_auth.py
EOF
)"
```

### Work Request

```bash
# Mayor assigns task to polecat
gt mail send gastown/Toast -s "Task: gt-def" -m "$(cat <<'EOF'
Please work on gt-def - Add validation to form
Priority: Normal
Context: See bead for details
EOF
)" --type task --notify
```

### Session Handoff

```bash
# Self handoff for next session
gt mail send --self -s "🤝 HANDOFF: Feature X" -m "$(cat <<'EOF'
## Context
Implementing feature X for gastown

## Completed
- Database schema migration
- API endpoints

## Next Steps
1. Frontend components
2. Integration tests

## Files Modified
- src/api/feature.py
- src/models/feature.py
EOF
)" --pinned
```

## Best Practices

1. **Use descriptive subjects** - Easy to scan in inbox
2. **Include context** - Don't assume recipient knows background
3. **Use appropriate type** - task for work, notification for FYI
4. **Notify for urgency** - Use --notify for time-sensitive
5. **Pin handoffs** - Ensure persistence across sessions
6. **Reply to thread** - Use --reply-to for conversations
