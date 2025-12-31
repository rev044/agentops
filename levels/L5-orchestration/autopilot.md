---
description: Execute an epic to completion with automated reconciliation
---

# /autopilot

Runs an entire epic through waves automatically. Validates between waves. Pauses for human review when needed.

## Usage

```
/autopilot <epic-id>
/autopilot agentops-xyz
/autopilot agentops-xyz --dry-run
```

## What's Different from L4

At L5, full autonomy:
- Executes waves without human trigger
- Validates between each wave
- Pauses at checkpoints for critical decisions
- Reconciles state after each phase

## How It Works

1. Load epic and compute waves
2. Execute Wave 1 (parallel)
3. Validate (tests, lint, semantic checks)
4. If validation passes, continue to Wave 2
5. If validation fails, pause for human
6. Repeat until all waves complete
7. Run `/retro` to capture learnings

## Flags

| Flag | Purpose |
|------|---------|
| `--dry-run` | Show what would happen without executing |
| `--resume` | Continue from saved state after pause |
| `--approve-threshold=high` | Pause on any validation issue |

## Output

```
/autopilot agentops-epic-123

Epic: "Add user dashboard"
Total: 8 issues across 4 waves

Wave 1/4: 3 issues...
  ✓ agentops-a1, agentops-b2, agentops-c3
  Validation: PASS

Wave 2/4: 2 issues...
  ✓ agentops-d4, agentops-e5
  Validation: PASS

Wave 3/4: 2 issues...
  ✓ agentops-f6, agentops-g7
  Validation: 1 warning (unused import)
  Continuing...

Wave 4/4: 1 issue...
  ✓ agentops-h8
  Validation: PASS

Epic complete! 8/8 issues closed.
Running /retro...
```

## Human Checkpoints

Autopilot pauses when:
- Validation finds HIGH severity issues
- Wave execution fails
- State mismatch detected

```
PAUSED - Validation found HIGH issue:
  "Potential SQL injection in user query"

Options:
1. Continue (ignore) → "continue"
2. Fix first         → "fix"
3. Abort             → "abort"

Your choice:
```

## Next

- `/retro` - Already run at completion
- Review `.agents/retros/` for learnings
