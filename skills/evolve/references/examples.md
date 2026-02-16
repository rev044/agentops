# /evolve Examples

## Infinite Autonomous Improvement

**User says:** `/evolve`

**What happens:**
1. Agent checks kill switch files (none found, continues)
2. Agent measures fitness against GOALS.yaml (3 of 5 goals passing)
3. Agent selects worst-failing goal by weight (test-pass-rate)
4. Agent invokes `/rpi "Improve test-pass-rate"` with full lifecycle
5. Agent re-measures fitness post-cycle (test-pass-rate now passing, all others unchanged)
6. Agent logs cycle to history, increments cycle counter
7. Agent loops to next cycle, selects next failing goal
8. After all goals pass, agent checks harvested work from post-mortem — finds 3 items
9. Agent works through harvested items, each generating more via post-mortem
10. After 3 consecutive idle cycles (no failing goals, no harvested work), agent runs `/post-mortem` and writes session summary
11. To stop earlier: create `~/.config/evolve/KILL` or `.agents/evolve/STOP`

**Result:** Runs forever — fixing goals, consuming harvested work, re-measuring. Only stops on kill switch or stagnation (3 idle cycles).

## Dry-Run Mode

**User says:** `/evolve --dry-run`

**What happens:**
1. Agent measures fitness (3 of 5 goals passing)
2. Agent identifies worst-failing goal (doc-coverage, weight 5)
3. Agent reports what would be worked on: "Dry run: would work on 'doc-coverage' (weight: 5)"
4. Agent shows harvested work queue (2 items from prior RPI cycles)
5. Agent stops without executing

**Result:** Fitness report and next-action preview without code changes.

## Regression with Revert

**User says:** `/evolve --max-cycles=3`

**What happens:**
1. Agent improves goal A in cycle 1 (commit abc123)
2. Agent measures fitness post-cycle: goal A passes, but goal B now fails (regression)
3. Agent reverts commit abc123 with annotated message
4. Agent logs regression to history, moves to next goal
5. Agent completes 3 cycles (cap reached), runs post-mortem

**Result:** Fitness regressions are auto-reverted, preventing compounding failures.
