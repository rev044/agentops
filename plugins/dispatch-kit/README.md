# Dispatch Kit

Work assignment and orchestration primitives. 4 skills for multi-agent coordination.

## Install

```bash
/plugin install dispatch-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/dispatch` | auto-triggered | Work assignment with gt sling |
| `/handoff` | `/handoff` | Session continuity |
| `/roles` | auto-triggered | Role responsibilities |
| `/mail` | auto-triggered | Agent communication |

## The Propulsion Principle

> **If you find work on your hook, YOU RUN IT.**

No waiting for confirmation. The hook having work IS the assignment.

## Key Commands

### Dispatch

```bash
gt sling gt-1234 daedalus    # Send work to polecat
gt hook                       # Check your assigned work
gt convoy create "Wave 1" gt-1234 gt-1235  # Batch tracking
```

### Handoff

```bash
gt handoff                    # Cycle to fresh session
gt handoff -m "context..."    # With context notes
```

### Mail

```bash
gt mail send mayor/ -s "Subject" -m "Body"
gt mail inbox                 # Check messages
gt nudge daedalus/crew/peer "message"  # Wake agent
```

## Roles

| Role | Purpose |
|------|---------|
| **Mayor** | Global coordinator (dispatch, don't implement) |
| **Crew** | Human-guided developer |
| **Polecat** | Autonomous worker |
| **Witness** | Lifecycle monitor |
| **Refinery** | Merge processor |

## Philosophy

- **The Propulsion Principle** - work hooked → run it
- **Clean handoffs preserve context** - nothing lost between sessions
- **Roles define capabilities** - clear boundaries

## Related Kits

- **beads-kit** - Issues being dispatched
- **gastown-kit** - Full Gas Town orchestration
- **core-kit** - What gets implemented
