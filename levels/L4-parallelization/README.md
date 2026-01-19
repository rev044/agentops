# L4 — Parallelization

Execute independent issues in parallel with wave execution.

## What You'll Learn

- Identifying independent (unblocked) work
- Using `/implement-wave` for parallel execution
- Batch close and single commit patterns
- Sub-agent coordination

## Prerequisites

- Completed L3-state-management
- Comfortable with beads issue tracking
- Understanding of issue dependencies

## Available Commands

| Command | Purpose |
|---------|---------|
| `/implement-wave` | Execute all unblocked issues in parallel |
| `/plan <goal>` | Same as L3 |
| `/research <topic>` | Same as L2 |
| `/implement [id]` | Execute single issue |
| `/retro [topic]` | Same as L2 |

## Key Concepts

- **Wave**: Set of independent issues executed together
- **Sub-agents**: Parallel workers for each issue
- **Batch close**: All wave issues closed in single commit
- **Dependency resolution**: Only unblocked issues run

## Wave Workflow

```
bd ready → identifies unblocked issues
/implement-wave → spawns sub-agents
Sub-agents complete work
Batch commit and close
Next wave begins
```

## What's NOT at This Level

- No `/crank` (full autonomous execution)
- Human triggers each wave

## Next Level

Once comfortable with waves, progress to [L5-orchestration](../L5-orchestration/) for full autonomy.
