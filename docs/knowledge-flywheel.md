# The Knowledge Flywheel

> Agents are stateless. The repo learns.

## The Problem

Coding agents forget everything between sessions. Notes alone do not fix that. If a solved problem is not extracted, curated, retrieved, and reused, the repo keeps paying for the same lesson.

## The Solution

AgentOps turns session output into durable environment state. Every session automatically extracts knowledge, scores it, and feeds the best back into future sessions — making each one smarter than the last.

## The Flywheel

```
┌───────────────────────────────────────────────────────────────────────┐
│                     THE KNOWLEDGE FLYWHEEL                            │
│                                                                       │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐       │
│  │  1. WORK   │─>│  2. FORGE  │─>│  3. POOL   │─>│ 4. PROMOTE │       │
│  │  Session   │  │  Extract   │  │  Score &   │  │  Graduate  │       │
│  │            │  │            │  │  Queue     │  │            │       │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘       │
│       ^                                                  │            │
│       │         ┌────────────┐  ┌────────────┐           │            │
│       └─────────│  6. INJECT │<─│5. LEARNINGS│<──────────┘            │
│                 │  Surface   │  │  Permanent │                        │
│                 │  & Cite    │  │  Knowledge │                        │
│                 └────────────┘  └────────────┘                        │
│                                                                       │
│  Each citation feeds back: utility scores update, high-utility        │
│  knowledge surfaces more often, low-utility decays. This is the       │
│  compounding effect — sessions get smarter because the best           │
│  knowledge rises and the noise sinks.                                 │
└───────────────────────────────────────────────────────────────────────┘
```

## The Six Stages

### Stage 1: Work

You use Claude to build, debug, research, or plan. A transcript (JSONL) is created automatically.

### Stage 2: Forge

At session end, `ao forge transcript` parses the transcript and extracts structured knowledge — decisions, solutions, learnings, failures, and references. Each becomes a markdown file in `.agents/knowledge/pending/`.

### Stage 3: Pool

`ao flywheel close-loop` ingests pending files and scores each on five dimensions:

| Dimension | What it measures |
|-----------|-----------------|
| Specificity | Names concrete files, functions, error messages |
| Actionability | A future session can act on this without more context |
| Novelty | New knowledge, not repetition |
| Context | Explains WHY, not just WHAT |
| Confidence | How certain the extraction is |

Candidates are tiered: **Gold** (>0.85), **Silver** (0.70–0.85), **Bronze** (0.50–0.70), or **Discard** (<0.50).

### Stage 4: Promote

Candidates that pass the promotion gate graduate to permanent knowledge:

- **Age gate:** Must be >24h old (prevents promoting noise from the current session)
- **Citation gate:** Must have been cited at least once (proves another session found it useful)
- **Tier gate:** Gold and Silver auto-promote. Bronze requires 3+ citations.

### Stage 5: Learnings

Promoted knowledge lives in `.agents/learnings/` and `.agents/patterns/`. The maturity lifecycle:

```
provisional → established → archived
```

Maintenance runs automatically: deduplication, contradiction detection, staleness archival.

### Stage 6: Inject

At session start and during work, `ao inject` or `ao lookup` retrieves the most relevant learnings for the current task. Each retrieval creates a **citation** — the signal that drives the feedback loop.

Citations with positive feedback increase the learning's utility score → higher utility → ranked higher in next injection → cited more → utility increases more → **compounding**.

## The Compounding Math

The flywheel equation:

```
dK/dt = I(t) - δ·K + σ·ρ·K
```

- **δ** — Knowledge decay rate (0.17/week baseline)
- **σ** — Retrieval effectiveness (what fraction of relevant knowledge is surfaced)
- **ρ** — Citation rate (how often retrieved knowledge is actually used)
- **Escape velocity:** When σρ > δ, knowledge compounds faster than it decays

### Golden Signals

Four signals measure whether the flywheel is actually compounding:

```bash
ao flywheel status
```

| Signal | Question | Healthy |
|--------|----------|---------|
| Velocity Trend | Is σρ-δ improving over time? | Positive slope |
| Citation Pipeline | Are citations delivering value? | >60% high-utility |
| Research Closure | Is research being mined into learnings? | <10% orphaned |
| Reuse Concentration | Is the whole pool active or just a few items? | Gini < 0.4 |

## Knowledge Stores

| Store | Content | Updated By |
|------|---------|------------|
| `.agents/knowledge/pending/` | Forge output awaiting pool ingestion | `ao forge` (automatic at session end) |
| `.agents/pool/` | Scored candidates awaiting promotion | `ao flywheel close-loop` |
| `.agents/learnings/` | Promoted, permanent knowledge | Pool promotion pipeline |
| `.agents/patterns/` | Promoted decision patterns | Pool promotion pipeline |
| `.agents/research/` | Scoped investigations | `/research` |
| `.agents/findings/registry.jsonl` | Reusable findings | `/pre-mortem`, `/vibe`, `/post-mortem` |
| `.agents/ao/citations.jsonl` | Citation trail | `ao inject`, `ao lookup` |
| `.agents/ao/feedback.jsonl` | Utility feedback | `ao flywheel close-loop` |
| `.agents/ao/metrics/` | Baseline snapshots for trend tracking | `ao metrics baseline` |

## The Compounding Effect

| Without the flywheel | With the flywheel |
|----------------------|-------------------|
| The same bug is rediscovered each session | The prior failure is retrieved before planning |
| Handoffs rely on chat memory | Handoffs and phased state live on disk |
| Notes accumulate without pressure | Useful findings get promoted into rules and constraints |
| Stale knowledge pollutes retrieval | Curation and maturity controls keep the corpus usable |
| Session 50 starts from scratch | Session 50 starts with 50 sessions of accumulated wisdom |

## See Also

- [Context Lifecycle Contract](context-lifecycle.md)
- [How It Works](how-it-works.md)
- [Primitive Chains](architecture/primitive-chains.md)
- [Brownian Ratchet](brownian-ratchet.md)
- [The Science](the-science.md)
