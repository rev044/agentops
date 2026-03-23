# Evolve Autonomous Execution Rules

## Fully Autonomous

Evolve runs without human intervention from start to teardown. Every `/rpi` invocation uses `--auto`. Do NOT ask the user for confirmation, clarification, or approval at any point. Do NOT pause between cycles. Do NOT summarize and wait. The user's only touchpoint is the teardown report at the very end.

## Each Cycle = Complete /rpi Run

All 3 phases (discovery → implementation → validation). Never invoke a partial RPI. If a task is too large for one cycle, break it into smaller sub-tasks during discovery and let `/crank` handle the waves. Evolve's job is to keep the loop turning, not to micro-manage individual tasks.

## Break Large Work into Sub-RPI Cycles

When work selection identifies a massive task (7+ issues, multi-subsystem scope), decompose it during `/rpi`'s discovery phase into an epic with waves. One evolve cycle = one `/rpi` run = one complete lifecycle. If the epic is too large for a single session, `/rpi`'s built-in retry and `--from=` resume handle continuation.

## Respect Repo-Local Program Contracts

When `PROGRAM.md` or `AUTODEV.md` exists, treat it as a hard operational constraint for cycle selection and keep/revert decisions:
- prefer work that can plausibly land inside mutable scope
- do not intentionally select work that requires immutable-scope edits
- treat out-of-scope work as escalation or backlog material, not permission to widen scope in place
- let `/rpi` enforce the active program contract during execution, then use its validation and decision policy during the evolve regression gate

## Anti-Patterns (DO NOT)

| Anti-Pattern | Why It's Wrong | Correct Behavior |
|--------------|----------------|------------------|
| Ask the user anything during execution | Evolve is fully autonomous — questions break the loop | Make best judgment, report in teardown |
| Stop after one `/rpi` cycle and summarize | Evolve loops until kill switch, max-cycles, or dormancy | Increment cycle and re-enter Step 1 |
| Run `/rpi` without `--auto` | Non-auto `/rpi` has human gates that halt the loop | Always pass `--auto` to `/rpi` |
| Run partial `/rpi` (skip validation) | Each cycle must be a complete 3-phase lifecycle | Let `/rpi` run all 3 phases autonomously |
| Pause between cycles to explain progress | The user wants results, not narration | Log cycle results, immediately start next cycle |
| Treat "no queued work" as "stop" | Generator layers (testing, validation, drift, features) produce work | Run all generator layers before considering dormancy |
| Select work that obviously violates `PROGRAM.md` | Scope escape is a tracked outcome, not a license to widen the loop | Escalate or re-queue out-of-scope work and keep selection inside mutable scope |
