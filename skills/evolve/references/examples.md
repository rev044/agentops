# /evolve Examples

## Infinite Autonomous Improvement

**User says:** `/evolve`

**What happens:**
1. Agent checks kill switch files (none found, continues).
2. Agent first reads `.agents/rpi/next-work.jsonl`, claims the highest-value harvested item, and runs `/rpi` on it.
3. The cycle's `/post-mortem` harvests 2 new follow-up items; evolve immediately re-reads the queue instead of trusting the pre-cycle snapshot.
4. With harvested work drained, evolve checks `bd ready` and lands the top unblocked bead.
5. With beads drained, evolve measures GOALS.yaml, finds a directive gap, and runs `/rpi` on that goal.
6. Once goals/directives are healthy, evolve generates testing work from thin coverage and lands the best regression-test improvement.
7. Testing producers dry up, so evolve runs validation tightening / bug-hunt and fixes the highest-value finding.
8. When remediation layers are empty, evolve mines hotspot/TODO/stale-doc drift and turns any real findings into durable work.
9. If all remediation layers stay empty, evolve writes a concrete feature suggestion as durable work and starts the next `/rpi`.
10. Only after repeated empty queue + generator passes does dormancy trigger and teardown begin.
11. To stop earlier: create `~/.config/evolve/KILL` or `.agents/evolve/STOP`.

**Result:** Runs as an always-on compounding loop. Empty queues trigger more work discovery; they do not end the run.

## Dry-Run Mode

**User says:** `/evolve --dry-run`

**What happens:**
1. Agent measures fitness.
2. Agent reports the next harvested/beads/goals item it would work on.
3. If those are empty, agent reports the next generator layer it would run (testing, validation, drift, or feature suggestion).
4. Agent stops without executing.

**Result:** Next-action preview without code changes.

## Regression with Revert

**User says:** `/evolve --max-cycles=3`

**What happens:**
1. Agent claims a harvested queue item in cycle 1 and starts `/rpi`.
2. Post-cycle fitness shows a regression.
3. Agent reverts the cycle's changes.
4. Agent clears the queue claim and leaves `consumed: false`, so the work is available again.
5. Agent logs the regression and continues.

**Result:** Fitness regressions are auto-reverted, and claimed work is re-queued instead of being lost.

## Worked Overnight Ladder

**User says:** `/evolve --athena`

**What happens:**
1. Athena warmup surfaces a stale research note about runtime smoke coverage.
2. `bd ready` has one open docs/runtime parity bead, so evolve runs that first.
3. That bead's `/post-mortem` harvests an implementation follow-up into `next-work.jsonl`; evolve re-reads the queue and runs it immediately.
4. The queue empties, so evolve measures goals and fixes one directive gap via `/rpi`.
5. All goals now pass. Evolve generates testing work from thin coverage and lands a new regression test.
6. Testing producers dry up, so evolve runs a bug-hunt / validation sweep and tightens a missing validation gate.
7. No bug-hunt findings remain, so evolve mines complexity/TODO/stale-doc drift and queues one cleanup item.
8. After that cleanup, the remediation ladder is empty, so evolve writes a concrete feature suggestion bead and starts the next `/rpi`.
9. Only after harvested work, beads, goals, testing, bug hunt, drift mining, and feature suggestions all come up empty across repeated passes does dormancy trigger.

**Result:** One long-running session compounds across beads -> harvested work -> goals -> testing -> bug hunt -> feature suggestion instead of stopping at the first empty queue.
