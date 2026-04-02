# Software Factory Surface

AgentOps already contains the machinery for a software factory. This document
names the operator surface explicitly so users do not have to infer it from
skills, hooks, CLI commands, and internal artifacts.

## Thesis

AgentOps is best understood as a **software-factory control plane**.

The environment carries:

- bounded briefing and context assembly
- tracked planning and scoped execution
- validation gates and ratchet checkpoints
- durable learning and loop closure
- isolated work lanes for long-running or parallel work

The workers remain replaceable. The environment carries continuity.

This follows the repo's stateful-environment/stateless-agents theory and its
own lifecycle/flywheel contracts: briefings and runtime state are the operator
surface; packets, chunks, topics, and builders are substrate.

## Operator Lane

For Codex and other explicit-runtime workflows, treat this as the canonical
lane:

```bash
ao factory start --goal "fix auth startup"
/rpi "fix auth startup"
# or: ao rpi phased "fix auth startup"
ao codex stop
```

That lane does four things:

1. `ao factory start` tries to compile a small goal-time briefing, then runs
   explicit Codex startup so the session begins with bounded context.
2. `/rpi` or `ao rpi phased` runs the delivery line: discovery,
   implementation, validation.
3. `ao rpi status` lets the operator inspect long-running phased work.
4. `ao codex stop` closes the flywheel explicitly so the session leaves behind
   learnings, citations, and handoff state.

## Surface Map

| Layer | Purpose | Primary surfaces |
|------|---------|------------------|
| Operator | What the human or lead agent should touch first | `ao factory start`, `/rpi`, `ao rpi phased`, `ao rpi status`, `ao codex stop` |
| Briefing + runtime | Bounded startup context and thread-time state | `ao knowledge brief`, `ao codex start`, `ao codex ensure-start`, `ao context assemble` |
| Delivery line | Research, planning, execution, validation | `/discovery`, `/plan`, `/crank`, `/validation`, `/rpi` |
| Learning loop | Convert completed work into future advantage | `ao codex stop`, `ao knowledge activate`, `ao flywheel close-loop`, `/retro`, `/forge` |
| Substrate | Retrieval, provenance, packetization, and promotion machinery | `.agents/packets/`, `.agents/topics/`, `.agents/briefings/`, `.agents/findings/`, hooks, builder logic |

## Why This Surface Exists

The factory framing matters because the repo already has the hard parts:

- RPI provides the conveyor belt.
- Context packets and briefings provide bounded work orders.
- The flywheel provides durable learning and loop closure.
- Codex lifecycle commands provide explicit runtime boundaries where hooks do
  not exist.

Without an explicit operator lane, users see a powerful collection of
primitives. With it, they see one product surface.

## Design Rules

- Prefer briefings over giant startup dumps.
- Keep substrate and operator surfaces distinct.
- Let external validation outrank self-report.
- Treat thin topics as discovery-only until evidence improves.
- Keep `athena` scoped to hygiene, not full operator-surface activation.

## Related Docs

- [How It Works](how-it-works.md)
- [Context Packet](context-packet.md)
- [Knowledge Flywheel](knowledge-flywheel.md)
- [Session Lifecycle](workflows/session-lifecycle.md)
- [CLI Reference](../cli/docs/COMMANDS.md)
