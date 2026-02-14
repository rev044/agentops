# AgentOps Architecture

> Context quality is the primary lever for agent output quality. Orchestrate what enters each window, compound what comes out.

## Overview

AgentOps is a skills plugin that orchestrates context across agent windows and compounds results through a knowledge flywheel — each session is smarter than the last.

```
.
├── .claude-plugin/
│   └── plugin.json      # Plugin manifest
├── skills/              # 34 skills (24 user-facing, 10 internal)
│   ├── rpi/             # orchestration — Full RPI lifecycle orchestrator
│   ├── council/         # orchestration — Multi-model validation (core primitive)
│   ├── crank/           # orchestration — Autonomous epic execution
│   ├── swarm/           # orchestration — Parallel agent spawning
│   ├── codex-team/      # orchestration — Parallel Codex execution
│   ├── evolve/          # orchestration — Goal-driven fitness loop
│   ├── implement/       # team — Execute single issue
│   ├── quickstart/      # solo — Interactive onboarding
│   ├── status/          # solo — Single-screen dashboard
│   ├── research/        # solo — Deep codebase exploration
│   ├── plan/            # solo — Decompose epics into issues
│   ├── vibe/            # solo — Code validation (complexity + council)
│   ├── pre-mortem/      # solo — Council on plans
│   ├── post-mortem/     # solo — Council + retro (wrap up work)
│   ├── retro/           # solo — Extract learnings
│   ├── complexity/      # solo — Cyclomatic analysis
│   ├── knowledge/       # solo — Query knowledge artifacts
│   ├── bug-hunt/        # solo — Investigate bugs
│   ├── doc/             # solo — Generate documentation
│   ├── product/         # solo — Generate PRODUCT.md
│   ├── handoff/         # solo — Session handoff
│   ├── inbox/           # solo — Agent mail monitoring
│   ├── release/         # solo — Pre-flight, changelog, tag
│   ├── trace/           # solo — Trace design decisions
│   ├── beads/           # library — Issue tracking reference
│   ├── standards/       # library — Coding standards
│   ├── shared/          # library — Shared reference docs
│   ├── inject/          # background — Load knowledge at session start
│   ├── extract/         # background — Extract from transcripts
│   ├── forge/           # background — Mine transcripts
│   ├── provenance/      # background — Trace knowledge lineage
│   ├── ratchet/         # background — Progress gates
│   ├── flywheel/        # background — Knowledge health monitoring
│   └── using-agentops/  # meta — Workflow guide (auto-injected)
├── hooks/               # Session hooks
│   ├── hooks.json
│   ├── session-start.sh
│   └── ...              # 12 hook scripts total
├── lib/                 # Shared code
│   ├── skills-core.js
│   └── scripts/prescan.sh
└── docs/                # Documentation
```

---

## Design Philosophy

Three principles drive every architectural decision in AgentOps:

**The intelligence lives in the window.** Agent output quality is determined by context input quality. Bad answers mean wrong context was loaded. Contradictions mean context wasn't shared between agents. Hallucinations mean context was too sparse. Drifting means signal-to-noise collapsed. Every failure is a context failure — so every solution is a context solution.

**Least-privilege context loading.** Each agent receives only the context necessary for its task. Research gets prior knowledge. Plan gets a 500-token research summary. Crank workers get fresh context per wave with zero bleed-through. Vibe gets recent changes only. Phase summaries compress output between phases to prevent signal-to-noise collapse. The context window is treated as a security boundary — nothing enters without scoping.

**The cycle is the product.** No single skill is the value. The compounding loop — research, plan, validate, build, validate, learn, repeat — makes each successive context window smarter than the last. Post-mortem doesn't just extract learnings; it proposes the next cycle's work. The system feeds itself.

---

## The RPI Workflow

```
Research → Plan → Implement → Validate
    ↑                            │
    └──── Knowledge Flywheel ────┘
```

Each phase is a context boundary. The output of one phase is compressed and scoped before entering the next — preventing context contamination across phases.

### Phases

| Phase | Skills | Output |
|-------|--------|--------|
| **Research** | `/research`, `/knowledge` | `.agents/research/` |
| **Plan** | `/pre-mortem`, `/plan` | Beads issues |
| **Implement** | `/implement`, `/crank` | Code, tests |
| **Validate** | `/vibe`, `/retro`, `/post-mortem` | `.agents/learnings/`, `.agents/patterns/` |

### Knowledge Flywheel

Every `/post-mortem` feeds back into the next `/rpi` cycle:

1. Council validates the implementation
2. `/retro` extracts learnings → `.agents/learnings/`
3. Process improvement proposals synthesized from retro findings
4. Next-work items harvested → `.agents/rpi/next-work.jsonl`
5. **Suggested `/rpi` command presented** — ready to copy-paste

```
  /rpi "goal A"
    └─ post-mortem → retro → process improvements → "Next: /rpi goal B"
                                                          │
  /rpi "goal B" ←──────────────────────────────────────────┘
    └─ post-mortem → retro → process improvements → "Next: /rpi goal C"
```

The flywheel is self-perpetuating: every cycle learns, proposes improvements, and queues the next cycle. The system doesn't just get smarter — it tells you exactly what to improve next.

Learnings re-enter future context windows through quality gates: 5-dimension scoring (specificity, actionability, novelty, context, confidence) into gold/silver/bronze tiers, with freshness decay (MemRL two-phase retrieval, delta=0.17/week) ensuring stale knowledge loses priority automatically. The flywheel is curation, not just storage.

---

## Skills

### Core Workflow

| Skill | Purpose |
|-------|---------|
| `/research` | Deep codebase exploration |
| `/plan` | Decompose goals into trackable issues |
| `/implement` | Execute a single issue |
| `/crank` | Autonomous multi-issue execution |
| `/vibe` | Code validation (8 aspects) |
| `/retro` | Extract learnings |
| `/post-mortem` | Full validation + knowledge extraction |

### Orchestration

| Skill | Purpose |
|-------|---------|
| `/council` | Multi-model validation (core primitive) |
| `/swarm` | Parallel agent spawning |
| `/codex-team` | Parallel Codex execution agents |

### Utilities

| Skill | Purpose |
|-------|---------|
| `/beads` | Issue tracking operations |
| `/bug-hunt` | Root cause analysis |
| `/knowledge` | Query knowledge artifacts |
| `/complexity` | Code complexity analysis |
| `/doc` | Documentation generation |
| `/pre-mortem` | Failure simulation |
| `/handoff` | Session handoff |
| `/inbox` | Agent mail monitoring |
| `/release` | Pre-flight, changelog, version bumps, tag |
| `/status` | Single-screen dashboard |
| `/quickstart` | Interactive onboarding |
| `/trace` | Trace design decisions |

### Internal (not invoked directly)

| Skill | Purpose |
|-------|---------|
| `beads` | Issue tracking reference (loaded by /implement, /plan) |
| `standards` | Coding standards (loaded by /vibe, /implement, /doc) |
| `shared` | Shared reference documents |
| `inject` | Load knowledge at session start (hook-triggered) |
| `extract` | Extract from transcripts (hook-triggered) |
| `forge` | Mine transcripts for knowledge |
| `provenance` | Trace knowledge lineage |
| `ratchet` | Progress gates |
| `flywheel` | Knowledge health monitoring |
| `using-agentops` | Workflow overview (auto-injected on session start) |

### Meta

| Skill | Purpose |
|-------|---------|
| `using-agentops` | Workflow overview (auto-injected on session start) |

---

## ao CLI Integration

For full workflow orchestration, skills integrate with the ao CLI:

| Skill | ao Command |
|-------|------------|
| `/research` | `ao forge search` |
| `/retro` | `ao forge index` |
| `/post-mortem` | `ao ratchet record` |
| `/implement` | `ao ratchet claim/record` |
| `/crank` | `ao ratchet verify` |

The ao CLI provides:
- **ao forge** - Semantic knowledge search and indexing
- **ao ratchet** - Progress tracking with Brownian Ratchet pattern

---

## Subagents

Subagents are disposable. Each gets fresh context scoped to its role — no accumulated state, no bleed-through. Clean context in, validated output out, then terminate.

Subagent behaviors are defined inline within SKILL.md files — there is no separate `agents/` directory. Skills that use subagents (e.g., `/council`, `/vibe`, `/pre-mortem`, `/post-mortem`, `/research`) spawn them via runtime-native backends (Codex sub-agents, Claude teams, or fallback background tasks).

### Validation Subagents (spawned by /vibe, /council)

| Agent Role | Focus |
|------------|-------|
| Code reviewer | Quality, patterns, maintainability |
| Security reviewer | Vulnerabilities, OWASP |
| Security expert | Deep security analysis |
| Architecture expert | System design, cross-cutting |
| Code quality expert | Complexity, refactoring |
| UX expert | Accessibility, user experience |

### Post-Mortem Subagents (spawned by /post-mortem)

| Agent Role | Focus |
|------------|-------|
| Plan compliance expert | Compare implementation to plan |
| Goal achievement expert | Did we solve the problem? |
| Ratchet validator | Verify gates are locked |
| Flywheel feeder | Extract learnings with provenance |
| Technical learnings expert | Technical patterns |
| Process learnings expert | Process improvements |

### Pre-Mortem Subagents (spawned by /pre-mortem)

| Agent Role | Focus |
|------------|-------|
| Integration failure expert | Integration risks |
| Ops failure expert | Operational risks |
| Data failure expert | Data integrity risks |
| Edge case hunter | Edge cases and exceptions |

### Research Subagents (spawned by /research)

| Agent Role | Focus |
|------------|-------|
| Coverage expert | Research completeness |
| Depth expert | Depth of analysis |
| Gap identifier | Missing areas |
| Assumption challenger | Challenge assumptions |

---

## Knowledge Artifacts

`.agents/` stores knowledge generated during sessions:

```
.agents/
├── bundles/       # Grouped artifacts
├── council/       # Council/validation reports
├── handoff/       # Session handoff context
├── learnings/     # Extracted lessons
├── patterns/      # Reusable patterns
├── plans/         # Implementation plans
├── pre-mortems/   # Failure simulations
├── reports/       # General reports
├── research/      # Exploration findings
├── retros/        # Retrospective reports
├── specs/         # Validated specifications
└── tooling/       # Tooling documentation
```

Knowledge artifacts are the system's long-term memory. Future `/research` commands discover them via file pattern matching, semantic search (`ao forge`), or Smart Connections MCP (if available). Freshness decay ensures stale artifacts lose priority over time — the system forgets what's no longer relevant. Quality gates prevent low-confidence or context-specific learnings from polluting the shared knowledge base.

---

## Context Boundaries

The system enforces context isolation at three levels:

**Phase boundaries.** Each RPI phase produces a compressed summary (500 tokens max) that feeds the next phase. Raw output never crosses phase boundaries — only distilled signal.

**Worker boundaries.** Each crank worker gets fresh context scoped to its assigned issue. Workers cannot see each other's work-in-progress. Only the lead sees all workers' output and commits.

**Session boundaries.** Each session starts with injected knowledge (freshness-weighted, quality-gated) and ends with extracted learnings. The flywheel bridges sessions without carrying raw context forward.

---

## Session Hook

On session start, `hooks/session-start.sh`:
1. Creates `.agents/` directories if missing
2. Injects `using-agentops` skill content as context
3. Outputs JSON with `additionalContext` for compatible agent runtimes

---

## Installation

```bash
npx skills@latest add boshu2/agentops --all -g
```

Optional:
- [beads](https://github.com/steveyegge/beads) for issue tracking
- [ao CLI](https://github.com/boshu2/ao) for full orchestration
