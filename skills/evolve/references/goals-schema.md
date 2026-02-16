# GOALS.yaml Schema

```yaml
version: 1
mission: "What this repo does"

goals:
  - id: unique-identifier
    description: "Human-readable description"
    check: "shell command — exit 0 = pass, non-zero = fail"
    weight: 1-10  # Higher = fix first
```

Goals are checked in weight order (highest first). The first failing goal with the highest weight is selected for improvement.

## Fitness Snapshot Format

Each cycle writes a fitness snapshot with **continuous values** (not just pass/fail):

```json
{
  "cycle": 1,
  "timestamp": "2026-02-12T15:45:00-05:00",
  "cycle_start_sha": "abc1234",
  "goals": [
    {
      "id": "go-coverage-floor",
      "result": "pass",
      "weight": 2,
      "value": 86.1,
      "threshold": 80
    },
    {
      "id": "doc-coverage",
      "result": "pass",
      "weight": 2,
      "value": 20,
      "threshold": 16
    },
    {
      "id": "go-cli-builds",
      "result": "pass",
      "weight": 5,
      "value": null,
      "threshold": null
    }
  ]
}
```

- **value**: The continuous metric extracted from the check command (null for binary-only goals)
- **threshold**: The pass/fail threshold (null for binary-only goals)
- **cycle_start_sha**: Git SHA at cycle start, used for multi-commit revert on regression

Pre-cycle snapshots: `fitness-{N}.json`
Post-cycle snapshots: `fitness-{N}-post.json` (for full-fitness regression comparison)

## Cycle-0 Baseline

Before the first improvement cycle runs, the system captures a baseline fitness snapshot (`fitness-0-baseline.json`). This serves as the comparison anchor for measuring session-wide progress.

The baseline includes:
- **All goals** from GOALS.yaml, measured in their initial state
- **Cycle-0 report** (`cycle-0-report.md`) — summary of which goals are failing and their weights
- **No regression comparisons** — this is the starting point

When the session ends (at Teardown), the system computes the **session fitness trajectory** by comparing the baseline against the final cycle snapshot. This produces `session-fitness-delta.md`, which shows which goals improved, regressed, or stayed unchanged over the entire /evolve session.
