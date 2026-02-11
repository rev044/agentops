# AgentOps Architecture

> AI-assisted development workflows for skills-protocol coding agents.

## Overview

AgentOps is a skills plugin providing the RPI workflow with Knowledge Flywheel.

```
.
├── .claude-plugin/
│   └── plugin.json      # Plugin manifest
├── skills/              # 32 skills (21 user-facing, 11 internal)
│   ├── council/         # orchestration — Multi-model validation (core primitive)
│   ├── crank/           # orchestration — Autonomous epic execution
│   ├── swarm/           # orchestration — Parallel agent spawning
│   ├── codex-team/      # orchestration — Parallel Codex execution
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
│   └── ...              # 10 hook scripts total
├── lib/                 # Shared code
│   ├── skills-core.js
│   └── scripts/prescan.sh
└── docs/                # Documentation
```

---

## The RPI Workflow

```
Research → Plan → Implement → Validate
    ↑                            │
    └──── Knowledge Flywheel ────┘
```

### Phases

| Phase | Skills | Output |
|-------|--------|--------|
| **Research** | `/research`, `/knowledge` | `.agents/research/` |
| **Plan** | `/pre-mortem`, `/plan` | Beads issues |
| **Implement** | `/implement`, `/crank` | Code, tests |
| **Validate** | `/vibe`, `/retro`, `/post-mortem` | `.agents/learnings/`, `.agents/patterns/` |

### Knowledge Flywheel

Every `/post-mortem` feeds back to `/research`:

1. Learnings extracted → `.agents/learnings/`
2. Patterns discovered → `.agents/patterns/`
3. Research enriched → Future sessions benefit

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

Future `/research` commands discover these automatically via:
1. File pattern matching
2. Semantic search (ao forge)
3. Smart Connections MCP (if available)

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
