# Error Recovery

## Overview

Gas Town provides automatic crash recovery via daemon. Orchestrator's job is to detect issues and escalate when needed.

## Automatic Recovery

When a polecat crashes:

```
1. Tmux detects crash (pane-died hook)
2. Daemon notices on patrol (every 3 min)
3. Daemon respawns polecat
4. SessionStart hook runs
5. Polecat discovers hook via gt hook
6. Work resumes from last checkpoint
```

**Orchestrator action:** None. Just wait and poll convoy.

## Investigating Stuck Polecats

### Peek at Polecat

```bash
# Capture recent output (last 100 lines)
gt peek gastown/polecats/nux

# Get more context
gt peek gastown/polecats/nux -n 200
```

### Nudge Polecat

```bash
# Send a message to resume or redirect
gt nudge gastown/polecats/nux "Resume work on gt-abc"

# More specific nudge
gt nudge gastown/polecats/nux "Focus on the failing test in auth_test.go"
```

### Kill and Respawn

```bash
# Bring down the polecat
gt down gastown/polecats/nux

# Re-sling the work
gt sling gt-abc gastown
```

## Detecting Blockers

Polecats report blockers via beads comments:

```bash
# Polecat writes:
bd comments add gt-abc "BLOCKER: Need database credentials to test"

# Orchestrator detects:
for issue in $wave_issues; do
    blocker=$(bd comments $issue 2>/dev/null | grep "BLOCKER:")
    if [ -n "$blocker" ]; then
        echo "Issue $issue blocked: $blocker"
        # Escalate to human
    fi
done
```

## Error Types and Responses

| Error | Detection | Response |
|-------|-----------|----------|
| **Polecat crash** | Daemon auto-detects | Wait, daemon restarts |
| **Polecat stuck** | Convoy timeout | `gt peek`, `gt nudge` |
| **Explicit blocker** | BLOCKER comment | Escalate to human |
| **Daemon down** | `gt daemon status` fails | Prompt human to restart |
| **Rig missing** | `gt rig list` empty | Guide rig setup |
| **Beads error** | `bd` commands fail | Run `bd doctor` |

## Escalation Protocol

When orchestrator detects unrecoverable error:

```bash
# Log checkpoint for human
bd comments add $epic_id "CHECKPOINT: Issue $issue_id blocked - $blocker_message"

# Output clear message
echo "CHECKPOINT: Human intervention required"
echo "Issue $issue_id reports: $blocker_message"
echo ""
echo "Options:"
echo "1. Resolve blocker and run: /gastown $epic_id --resume"
echo "2. Skip issue: bd close $issue_id --reason 'Skipped'"
echo "3. Cancel: bd close $epic_id --reason 'Cancelled'"
```

## Defensive Patterns

```python
def execute_wave(issues: list, rig: str):
    # Create convoy first
    convoy_id = create_convoy(issues)

    # Dispatch with error handling
    for issue in issues:
        try:
            result = bash(f"gt sling {issue} {rig}")
            if "Error" in result:
                log(f"Failed to sling {issue}: {result}")
                # Don't fail whole wave - continue with others
        except Exception as e:
            log(f"Exception slinging {issue}: {e}")

    # Monitor with timeout
    result = monitor_convoy(convoy_id, timeout_minutes=60)

    if result == "timeout":
        investigate_stuck_polecats(issues)
        escalate_to_human("Wave timeout")

    if result == "blocked":
        blockers = get_blocker_messages(issues)
        escalate_to_human(f"Wave blocked: {blockers}")

    return result
```

## Recovery Commands Quick Reference

```bash
# Check polecat output
gt peek <rig>/polecats/<name>

# Nudge polecat
gt nudge <rig>/polecats/<name> "Message"

# Kill polecat
gt down <rig>/polecats/<name>

# Respawn with new work
gt sling <issue> <rig>

# Check daemon
gt daemon status

# Restart daemon
gt daemon stop && gt daemon start

# Diagnose beads
bd doctor
```

## Fallback to Task Agents

If Gas Town completely unavailable:

```python
def dispatch_with_fallback(issue, rig):
    # Try polecat first
    result = bash(f"gt sling {issue} {rig} 2>&1")

    if "Error" in result or "not found" in result:
        # Fall back to Task agent
        log(f"Gas Town unavailable, using Task agent for {issue}")
        return Task(
            subagent_type="general-purpose",
            prompt=f"Implement {issue}",
            run_in_background=True
        )

    return None  # Polecat handling it
```
