# The Knowledge Flywheel

> Agents are stateless. The repo learns.

## The Problem

Coding agents forget everything between sessions. Notes alone do not fix that. If a solved problem is not extracted, curated, retrieved, and reused, the repo keeps paying for the same lesson.

## The Solution

AgentOps turns session output into durable environment state:

- research and design artifacts in `.agents/`
- learnings and patterns extracted from completed work
- reusable findings captured in a normalized registry
- next-work queues and ratchet checkpoints for continuity
- curation signals that keep retrieval focused on what is still useful

## The Flywheel

```text
Do work -> extract signal -> curate -> retrieve -> apply -> reinforce
    ^                                                      |
    |______________________________________________________|
```

The loop is not just memory. It is memory plus validation plus loop closure.

## Lifecycle

| Stage | Surfaces | What happens |
|------|----------|--------------|
| Capture | `/research`, `/post-mortem`, `/retro`, `/forge` | Current work is written down as research, learnings, or findings |
| Curate | `ao maturity`, `ao dedup`, `ao contradict`, constraint review | Stale or conflicting artifacts lose weight; useful ones stay retrievable |
| Retrieve | `ao lookup`, `ao search`, startup hooks, phased handoffs, ranked packets | The next task starts with repo-native context instead of a blank window |
| Apply | `/plan`, `/pre-mortem`, `/implement`, `/vibe` | Prior lessons shape current choices and validation, using the best-matching packet rather than generic recall |
| Reinforce | citations, repeated use, promotion into constraints | Frequently useful knowledge hardens into planning rules or gates |

## Knowledge Stores

| Store | Content | Updated By |
|------|---------|------------|
| `.agents/research/` | Scoped understanding and repo investigations | `/research` |
| `.agents/brainstorm/` | Problem framing and option exploration | `/brainstorm` |
| `.agents/learnings/` | Reusable lessons and retrospective signal | `/retro`, `/post-mortem`, `/forge` |
| `.agents/findings/registry.jsonl` | Reusable findings before they become rules or constraints | `/pre-mortem`, `/vibe`, `/post-mortem` |
| `.agents/rpi/next-work.jsonl` | Harvested next steps | `/post-mortem`, `/evolve` |
| `.agents/ao/` | Ratchet trail, provenance, session metadata | `ao ratchet`, `ao forge`, `ao flywheel` |

The strongest compounding pattern is a **ranked stigmergic packet**: for a given goal or review target, select the best matching compiled prevention, active findings, and high-severity queued follow-up work, then carry that packet forward consistently across discovery, planning, review, and status surfaces.

## The Compounding Effect

| Without the flywheel | With the flywheel |
|----------------------|-------------------|
| The same integration bug is rediscovered in a new session | The prior failure is retrieved before planning or validation |
| Handoffs rely on chat memory | Handoffs and phased state live on disk |
| Notes accumulate without pressure | Useful findings get promoted into rules, checks, or constraints |
| Stale knowledge pollutes retrieval | Curation, contradiction checks, and maturity controls keep the corpus usable |

The practical result is simple: each completed cycle can leave behind a more capable environment than the one it started in.

## See Also

- [Context Lifecycle Contract](context-lifecycle.md)
- [How It Works](how-it-works.md)
- [Primitive Chains](architecture/primitive-chains.md)
- [Brownian Ratchet](brownian-ratchet.md)
