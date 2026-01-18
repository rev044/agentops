# Broadcast Reference

## Overview

`gt broadcast` sends a message to all active workers (polecats and crew) in Gas Town. Use for town-wide or rig-wide announcements.

## Basic Usage

```bash
gt broadcast "<message>"
```

## Command Reference

```bash
gt broadcast <message> [flags]

Flags:
  --all          Include infrastructure agents (mayor, witness, refinery)
  --rig <name>   Only broadcast to workers in this rig
  --dry-run      Preview what would be sent without sending
  -h, --help     Help for broadcast
```

## Examples

### Town-Wide Announcements

```bash
# All workers in town
gt broadcast "New priority work available"

# With infrastructure agents
gt broadcast --all "System maintenance in 5 minutes"

# Preview without sending
gt broadcast --dry-run "Test message"
```

### Rig-Specific Announcements

```bash
# Only gastown workers
gt broadcast --rig gastown "Rig-specific update"

# Only ai-platform workers
gt broadcast --rig ai-platform "API changes deployed"
```

## Delivery Mechanism

Broadcast uses `gt nudge` internally to send to each active worker's tmux session:

```
gt broadcast "message"
     │
     ├── gt nudge gastown/Toast "message"
     ├── gt nudge gastown/Nux "message"
     ├── gt nudge ai-platform/Alpha "message"
     └── ...
```

## Recipients

### Default Recipients (workers only)

- All active polecats across all rigs
- All active crew workers

### With --all Flag

Adds:
- Mayor
- Witness (per rig)
- Refinery (per rig)
- Deacon

## Use Cases

### Priority Shift

```bash
gt broadcast "Priority shift: Focus on security fixes"
```

### New Work Available

```bash
gt broadcast "Wave 2 issues ready - check bd ready"
```

### System Alert

```bash
gt broadcast --all "Build system down for 10 minutes"
```

### Rig-Specific Context

```bash
gt broadcast --rig gastown "New dependency added to gastown - run go mod download"
```

### Pre-Maintenance

```bash
gt broadcast --all "Shutting down in 5 minutes - commit and push work"
```

## DND Considerations

By default, broadcast respects DND (Do Not Disturb) settings. Agents with DND enabled will not receive the broadcast.

To send despite DND, you must nudge individually with `--force`:

```bash
# Broadcast doesn't have --force, use individual nudges
for polecat in $(gt polecat list --json | jq -r '.[].name'); do
    gt nudge "$polecat" "URGENT: Message" --force
done
```

## Best Practices

1. **Use sparingly** - Broadcasts interrupt all workers
2. **Be concise** - Workers see it in their session
3. **Use --dry-run first** - Verify recipients
4. **Prefer targeted mail** - For specific agents, use `gt mail send`
5. **Include action** - Tell workers what to do next

## Comparison with Mail

| Feature | Broadcast | Mail |
|---------|-----------|------|
| Delivery | Immediate (tmux) | Async (inbox) |
| Persistence | No record | Stored in beads |
| Targeting | All workers | Specific agents |
| Response | None expected | Can reply |
| Use case | Announcements | Work coordination |

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Missing recipients | Agent not active | Check `gt polecat list` |
| No one received | All on DND | Use individual nudges with --force |
| Wrong rig | Typo in --rig | Verify with `gt rig list` |
