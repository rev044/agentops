# AgentOps Levels — Progressive Learning Path

Inspired by Kelsey Hightower's approach: learn by doing, one level at a time.

## Philosophy

Each level builds on the previous. Master L1 before attempting L2. The levels are designed to be:

- **Progressive** — Each level adds ONE new concept
- **Practical** — Real demos, not just theory
- **Incremental** — Break changes into small, verifiable steps

## The Five Levels

| Level | Name | Core Concept | New Capability |
|-------|------|--------------|----------------|
| L1 | Basics | Single-session work | `/research`, `/implement` (single issue) |
| L2 | Persistence | `.agents/` output | State survives sessions |
| L3 | State Management | Issue tracking | `/plan`, beads integration |
| L4 | Parallelization | Wave execution | `/implement-wave` |
| L5 | Orchestration | Full autonomy | `/crank`, gastown multi-agent |

## Progression Path

```
L1 (Gateway)
    ↓
L2 (Add persistence)
    ↓
L3 (Add tracking)
    ↓
L4 (Add parallelism)
    ↓
L5 (Full autonomy)
```

## Level Details

### L1 — Basics
Single-session work. No state persistence. Use `/research` to explore, `/implement` to build. Changes exist only in git.

### L2 — Persistence
Add `.agents/` directory. Research documents, patterns, and learnings persist across sessions. The AI has memory.

### L3 — State Management
Add issue tracking with beads. `/plan` creates structured work items. Track progress across sessions.

### L4 — Parallelization
Execute independent work in parallel. `/implement-wave` runs multiple issues concurrently. Speed through unblocked work.

### L5 — Orchestration
Full autonomous operation. `/crank` handles epic-to-completion via the ODMCR reconciliation loop. Integrates with gastown for multi-agent parallelization.

## Getting Started

Start with [L1-basics/](./L1-basics/). Read its README, run the demos, then progress to L2.

## Directory Contents

Each level directory contains:
- `README.md` — What you'll learn, prerequisites, available commands
- `demo/` — Real session transcripts showing the level in action
