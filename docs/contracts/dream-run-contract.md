# Dream Run Contract

`ao overnight` is the headless automation surface for Dream, AgentOps' private overnight operator mode.

This contract defines the minimum behavior required for the shipped Dream surfaces:

- `ao overnight setup`
- `ao overnight start|run`
- `ao overnight report`

## Scope

V1 covers:

- local setup guidance with honest scheduler assistance artifacts
- local-first overnight runs against the real repo-local `.agents` corpus
- machine-readable and markdown morning summaries
- optional multimodel Dream Council synthesis over bounded runner reports
- built-in DreamScape terrain rendering inside the report
- explicit process supervision and lock behavior
- optional keep-awake assistance on macOS

V1 does not promise:

- guaranteed scheduled execution on sleeping laptops
- tracked source-code edits overnight
- free-form model-to-model chat overnight
- visualization outside the shared report contract

## Command Surface

Primary commands:

- `ao overnight setup`
- `ao overnight start`
- `ao overnight run`
  `run` is an alias for `start`
- `ao overnight report --from <dir-or-summary.json>`

Required flags for `start`:

- `--goal <text>`
- `--output-dir <path>`
- `--run-timeout <duration>`
- `--keep-awake`
- `--no-keep-awake`
- `--runner <name>` (repeatable)
- `--creative-lane`

Required flags for `setup`:

- `--apply`
- `--scheduler <manual|launchd|cron|systemd|auto>`
- `--at <HH:MM>`
- `--runner <name>` (repeatable)

## Process Model

One Dream run is a single bounded process with a stable output directory.

Defaults:

- output directory: `.agents/overnight/latest`
- lock file: `.agents/overnight/run.lock`
- log file: `<output-dir>/overnight.log`
- timeout: `8h`
- keep-awake: enabled by default, opt-out via config or `--no-keep-awake`

## Locking

Dream must prevent overlapping local runs.

Required behavior:

- acquire an exclusive non-blocking lock on the lock file before running
- fail fast if another run already holds the lock
- release the lock when the process exits

## Step Contract

The first slice runs these steps in order:

1. `ao flywheel close-loop --threshold 0h --json`
2. `ao --dry-run defrag --prune --dedup --oscillation-sweep`
3. `ao metrics health --json`
4. `ao retrieval-bench --live --json`
5. optional: `ao knowledge brief --goal <goal> --json`
6. optional: Dream Council packet generation
7. optional: one bounded runner pass per configured Dream runner
8. optional: Dream Council synthesis

Hard-fail steps:

- close-loop
- metrics health

Soft-fail steps:

- defrag preview
- retrieval bench
- knowledge brief
- Dream Council runner execution
- Dream Council synthesis when no runner completes

Soft failures must degrade the report, not delete it.

## Crash Behavior

If the Dream process fails after creating the output directory, it must still try to write:

- `summary.json`
- `summary.md`

The report must show:

- `status: failed`
- the last completed step
- degraded or failed steps
- log path

## Keep-Awake Behavior

V1 only auto-manages keep-awake on macOS via `caffeinate`.

Rules:

- default-on for local bedtime runs
- opt-out via config or `--no-keep-awake`
- if `caffeinate` is unavailable, continue and mark the run degraded
- non-macOS platforms must not fake scheduler guarantees

## Output Artifacts

Required artifacts:

- `<output-dir>/close-loop.json`
- `<output-dir>/defrag/latest.json` when the preview succeeds
- `<output-dir>/metrics-health.json`
- `<output-dir>/retrieval-bench.json` when live retrieval succeeds
- `<output-dir>/briefing.json` when a goal briefing succeeds
- `<output-dir>/council/packet.json` when Dream Council is configured
- `<output-dir>/council/<runner>.json` for each successful Dream Council runner
- `<output-dir>/council/synthesis.json` when Dream Council synthesis succeeds
- `<output-dir>/summary.json`
- `<output-dir>/summary.md`

## Relationship To CI

GitHub nightly remains the public proof harness.

`ao overnight` is the private compounding engine.

They may share primitive steps and report shapes, but they are not the same operational surface.

`ao overnight setup` helps persist `dream.*` config and generate host-specific
assistance artifacts. The host scheduler still owns actual scheduling semantics.
