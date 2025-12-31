# L5 — Orchestration

Full autonomous operation with autopilot.

## What You'll Learn

- Using `/autopilot` for epic-to-completion
- Reconciliation loops
- Validation gates
- Human checkpoint patterns

## Prerequisites

- Completed L4-parallelization
- Comfortable with wave execution
- Understanding of validation patterns

## Available Commands

| Command | Purpose |
|---------|---------|
| `/autopilot` | Epic-to-completion with validation gates |
| `/implement-wave` | Same as L4 |
| `/plan <goal>` | Same as L3 |
| `/research <topic>` | Same as L2 |
| `/implement [id]` | Same as L3 |
| `/retro [topic]` | Same as L2 |

## Key Concepts

- **Autopilot**: Autonomous epic execution
- **Reconciliation loop**: Continuous state checking
- **Validation gates**: Quality checks between phases
- **Human checkpoints**: Pause points for review

## Autopilot Flow

```
/autopilot <epic>
    ↓
Plan decomposition
    ↓
Wave execution (loop)
    ↓
Validation gate
    ↓
Human checkpoint (if configured)
    ↓
Continue or adjust
    ↓
Epic complete
```

## Safety Mechanisms

- Validation gates catch drift
- Human checkpoints for critical decisions
- Reconciliation detects state mismatches
- Graceful degradation on failures

## Mastery

At L5, you can hand off entire epics to the agent system and trust the outcome.
