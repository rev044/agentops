# Autonomous Execution Rules

## The DAG Rule

RPI is a 4-step DAG. You enter at `--from` and run every step after it to completion. There are no natural stopping points between steps — the DAG is one unit of work. The human's only touchpoint is after STEP 4 (report).

Unless complexity == `fast`, STEP 3 (validation) is mandatory. Skipping it breaks the knowledge flywheel — no quality check, no learnings captured, no compounding.

## Fully Autonomous by Default

Unless `--interactive` is explicitly set, RPI runs hands-free. Do NOT:
- Ask the user for confirmation between steps
- Ask "want me to commit?" or "should I continue?"
- Pause to summarize and wait for input
- Request clarification mid-execution
- Stop to ask about approach or strategy

If something is genuinely blocked (3 retries exhausted), then and only then do you stop and report.

## Phase Completion Tracking

After each step, log progress:
```
STEP 1 COMPLETE ✓ (discovery) → STEP 2
STEP 2 COMPLETE ✓ (implementation) → STEP 3
STEP 3 COMPLETE ✓ (validation) → STEP 4
STEP 4 COMPLETE ✓ (report) — RPI DONE
```
