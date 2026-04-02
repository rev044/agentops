# ao - AgentOps Software Factory CLI

`ao` is the explicit operator surface for AgentOps when you want the repo to
behave like a software factory instead of a loose collection of agent
primitives.

The short version:

```bash
ao quick-start
ao factory start --goal "fix auth startup"
ao rpi phased "fix auth startup"
ao codex stop
```

That lane keeps four concerns explicit:

1. `ao factory start` surfaces a bounded work order by compiling a goal-time
   briefing when the corpus supports it, then running Codex startup.
2. `ao rpi phased` runs the delivery line with fresh context per phase.
3. `ao rpi status` lets the operator inspect long-running factory work.
4. `ao codex stop` closes the flywheel so the session leaves behind learnings,
   citations, and handoff state.

## Install

```bash
go install github.com/boshu2/agentops/cli/cmd/ao@latest
```

## Quick Start

```bash
# From your repo root: create .agents/, starter knowledge surfaces, and hooks
ao quick-start

# Start a goal with briefing-first runtime context
ao factory start --goal "fix auth startup"

# Run the delivery lane
ao rpi phased "fix auth startup"
```

If you prefer the skill-first path, use `/rpi "fix auth startup"` after
`ao factory start`.

That's it. In Claude Code, `CLAUDE.md` remains the startup surface. The
installed hooks stay silent and only prepare runtime state for the factory lane.

## Operator Surfaces

**SessionStart**: Performs startup maintenance, recovers handoff state, and can
stage `factory-goal.txt` / `factory-briefing.txt` without injecting context.

**UserPromptSubmit**: When startup lacked a goal, the first substantive prompt
can be captured as silent factory intake and staged for later explicit use.

**SessionEnd**: Extracts learnings and updates the feedback loop.

| Command | Purpose |
|---------|---------|
| `ao factory start --goal "<goal>"` | Compile briefing-first startup context, then run explicit Codex start |
| `ao knowledge brief --goal "<goal>"` | Build the task-time briefing directly |
| `ao codex start` | Lower-level hookless Codex startup |
| `ao rpi phased "<goal>"` | CLI-first Discovery -> Implementation -> Validation lane |
| `ao rpi status` | Monitor long-running phased work |
| `ao codex stop` | Close the loop explicitly at session end |

## Lower-Level Substrate

These commands remain important, but they sit below the factory lane:

- `ao knowledge` for belief/playbook/briefing refresh and gap reporting
- `ao context assemble` for a five-section task briefing
- `ao lookup` and `ao search` for direct retrieval
- `ao forge transcript` and `ao flywheel close-loop` for manual lifecycle work

## Reference

- [Software Factory Surface](../docs/software-factory.md)
- [Session Lifecycle](../docs/workflows/session-lifecycle.md)
- [CLI Reference](docs/COMMANDS.md)
