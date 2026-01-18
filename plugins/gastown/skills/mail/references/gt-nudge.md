# Nudge Reference

## Overview

`gt nudge` sends a direct message to an agent's Claude Code tmux session. Use for immediate attention or unsticking stuck polecats.

## Basic Usage

```bash
gt nudge <target> "<message>"
gt nudge <target> -m "<message>"
```

## Command Reference

```bash
gt nudge <target> [message] [flags]

Flags:
  -m, --message <string>   Message to send
  -f, --force              Send even if target has DND enabled
  -h, --help               Help for nudge
```

## Target Formats

| Target | Example | Description |
|--------|---------|-------------|
| `<rig>/<polecat>` | `gastown/Toast` | Specific polecat |
| `mayor` | `mayor` | Mayor session (gt-mayor) |
| `deacon` | `deacon` | Deacon session (gt-deacon) |
| `witness` | `witness` | Current rig's witness |
| `refinery` | `refinery` | Current rig's refinery |
| `channel:<name>` | `channel:workers` | All members of channel |

## Examples

### Nudge Polecat

```bash
# Continue signal (most common)
gt nudge gastown/Toast "continue with your task"

# Status check
gt nudge gastown/Toast -m "What's your status?"

# Specific instruction
gt nudge gastown/Toast "Focus on completing gt-abc before starting new work"
```

### Nudge Infrastructure

```bash
# Nudge Mayor
gt nudge mayor "Status update requested"

# Nudge Witness
gt nudge witness "Check polecat health"

# Nudge Refinery
gt nudge refinery "Process merge queue"
```

### Nudge Channel

```bash
# Nudge all channel members
gt nudge channel:workers "New priority work available"
```

Channels are defined in `~/gt/config/messaging.json`:

```json
{
  "nudge_channels": {
    "workers": ["gastown/polecats/*", "ai-platform/polecats/*"],
    "oncall": ["mayor", "gastown/witness"]
  }
}
```

### Force Through DND

```bash
# Override DND for urgent message
gt nudge gastown/Toast "URGENT: Stop current work" --force
```

## Delivery Mechanism

Nudge uses tmux's `send-keys` with a reliable delivery pattern:

```
1. Send text in literal mode (-l flag)
2. Wait 500ms for paste to complete
3. Send Enter as separate command
```

This is the **only** supported way to send messages to Claude sessions.

## Use Cases

### Unstick Stuck Polecat

```bash
# Check if stuck
tmux capture-pane -t gt-gastown-Toast -p | tail -10

# Nudge to continue
gt nudge gastown/Toast "continue with your task"
```

### Request Status

```bash
gt nudge gastown/Toast "What issue are you working on? What's blocking you?"
```

### Redirect Work

```bash
gt nudge gastown/Toast "Stop current task. Priority: work on gt-xyz instead"
```

### Wake Idle Agent

```bash
gt nudge witness "Check if any polecats need attention"
```

## DND (Do Not Disturb)

Agents can enable DND to block nudges:

```bash
# Agent enables DND
gt dnd on

# Nudges are skipped
gt nudge gastown/Toast "message"  # → Skipped, DND enabled

# Force overrides DND
gt nudge gastown/Toast "URGENT" --force  # → Delivered
```

## Nudge vs Mail

| Use Nudge | Use Mail |
|-----------|----------|
| Immediate attention | Async coordination |
| "Continue" signals | Work assignment |
| Status checks | Task details |
| Unsticking | Formal handoff |
| No record needed | Audit trail needed |

## Common Patterns

### Unstick Loop

```bash
# Try nudge, wait, check progress
gt nudge gastown/Toast "continue"
sleep 30
status=$(tmux capture-pane -t gt-gastown-Toast -p | tail -5)

if echo "$status" | grep -q "stuck\|error"; then
    # Escalate to human
    gt mail send --human -s "Polecat stuck" -m "$status"
fi
```

### Status Collection

```bash
# Nudge all polecats for status
for polecat in $(gt polecat list gastown --json | jq -r '.[].name'); do
    gt nudge "gastown/$polecat" "Reply with current status"
done
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Nudge not delivered | Wrong session name | Check `tmux list-sessions` |
| Agent ignoring nudge | DND enabled | Use `--force` |
| No response | Agent crashed | Check `gt polecat status` |
| Session not found | Polecat not running | Use `gt polecat list` |

## Best Practices

1. **Keep messages short** - They appear inline in session
2. **Be specific** - Tell agent exactly what to do
3. **Use sparingly** - Nudges interrupt work flow
4. **Respect DND** - Only use --force for true emergencies
5. **Prefer mail for details** - Nudge to check mail, not for complex instructions
