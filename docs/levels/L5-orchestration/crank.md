---
description: Execute an epic to completion with automated reconciliation
---

# /crank

Runs an entire epic through the ODMCR loop automatically until ALL children are CLOSED.

## Usage

```
/crank <epic-id>
/crank agentops-xyz
/crank agentops-xyz --mode=crew
```

## What's Different from L4

At L5, full autonomy:
- Executes issues without human prompts
- Auto-detects Mayor vs Crew mode
- Reconciles state after each cycle
- NO stopping until epic is CLOSED

## How It Works

The ODMCR Loop:

1. **Observe**: `bd show <epic>`, `bd ready` - understand current state
2. **Dispatch**: Execute ready issues (mode-dependent)
3. **Monitor**: Track progress, check for completion
4. **Collect**: Close completed issues, update status
5. **Retry**: Handle failures, escalate blockers
6. Loop until all children are CLOSED

## Execution Modes

| Mode | Detection | Dispatch Method |
|------|-----------|-----------------|
| **Crew** | Default | Sequential `/implement` |
| **Mayor** | In `~/gt` or `*/mayor/*` | Parallel `gt sling` to polecats |

Force a mode:
```
/crank <epic> --mode=crew     # Sequential
/crank <epic> --mode=mayor    # Parallel via gastown
```

## Output

```
/crank agentops-epic-123

Epic: "Add user dashboard"
Mode: crew (sequential)
Total: 8 issues

[OBSERVE] 3 issues ready, 5 blocked
[DISPATCH] /implement agentops-a1
[COLLECT] agentops-a1 CLOSED
[OBSERVE] 2 issues ready, 4 blocked
[DISPATCH] /implement agentops-b2
...
[COLLECT] agentops-h8 CLOSED
[OBSERVE] 0 issues remaining

Epic CLOSED! 8/8 issues complete.
```

## Failure Handling

Crank handles failures automatically:
- Retry transient failures
- Skip blocked issues (revisit next cycle)
- Escalate persistent failures to human

## Next

- `/retro` - Extract learnings after epic completes
- Review `.agents/retros/` for captured patterns
