# AgentOps Architecture

> AI-assisted development workflows for Claude Code.

## Overview

AgentOps is a single Claude Code plugin providing the RPI workflow with Knowledge Flywheel.

```
.
├── .claude-plugin/
│   └── plugin.json      # Plugin manifest
├── skills/              # All 21 skills
│   ├── research/
│   ├── plan/
│   ├── implement/
│   ├── crank/
│   ├── vibe/
│   ├── retro/
│   ├── post-mortem/
│   ├── beads/
│   ├── bug-hunt/
│   ├── knowledge/
│   ├── complexity/
│   ├── doc/
│   ├── pre-mortem/
│   └── using-agentops/  # Meta skill (injected on session start)
├── agents/              # 20 Subagent definitions
│   ├── code-reviewer.md           # Code quality review
│   ├── security-reviewer.md       # Security vulnerabilities
│   ├── security-expert.md         # Deep security analysis
│   ├── architecture-expert.md     # System design
│   ├── code-quality-expert.md     # Complexity, refactoring
│   ├── ux-expert.md               # Accessibility, UX
│   ├── plan-compliance-expert.md  # Plan vs implementation
│   ├── goal-achievement-expert.md # Did we solve the problem?
│   ├── ratchet-validator.md       # Gate validation
│   ├── flywheel-feeder.md         # Knowledge extraction
│   ├── technical-learnings-expert.md
│   ├── process-learnings-expert.md
│   ├── integration-failure-expert.md  # Pre-mortem
│   ├── ops-failure-expert.md
│   ├── data-failure-expert.md
│   ├── edge-case-hunter.md
│   ├── coverage-expert.md         # Research quality
│   ├── depth-expert.md
│   ├── gap-identifier.md
│   └── assumption-challenger.md
├── hooks/               # Session hooks
│   ├── hooks.json
│   └── session-start.sh
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

### Utilities

| Skill | Purpose |
|-------|---------|
| `/beads` | Issue tracking operations |
| `/bug-hunt` | Root cause analysis |
| `/knowledge` | Query knowledge artifacts |
| `/complexity` | Code complexity analysis |
| `/doc` | Documentation generation |
| `/pre-mortem` | Failure simulation |

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

Located in `agents/`. 20 specialized validators used by `/vibe`, `/pre-mortem`, `/post-mortem`, and `/research`.

### Validation Agents (used by /vibe)

| Agent | Focus |
|-------|-------|
| `code-reviewer` | Quality, patterns, maintainability |
| `security-reviewer` | Vulnerabilities, OWASP |
| `security-expert` | Deep security analysis |
| `architecture-expert` | System design, cross-cutting |
| `code-quality-expert` | Complexity, refactoring |
| `ux-expert` | Accessibility, user experience |

### Post-Mortem Agents (used by /post-mortem)

| Agent | Focus |
|-------|-------|
| `plan-compliance-expert` | Compare implementation to plan |
| `goal-achievement-expert` | Did we solve the problem? |
| `ratchet-validator` | Verify gates are locked |
| `flywheel-feeder` | Extract learnings with provenance |
| `technical-learnings-expert` | Technical patterns |
| `process-learnings-expert` | Process improvements |

### Pre-Mortem Agents (used by /pre-mortem)

| Agent | Focus |
|-------|-------|
| `integration-failure-expert` | Integration risks |
| `ops-failure-expert` | Operational risks |
| `data-failure-expert` | Data integrity risks |
| `edge-case-hunter` | Edge cases and exceptions |

### Research Agents (used by /research)

| Agent | Focus |
|-------|-------|
| `coverage-expert` | Research completeness |
| `depth-expert` | Depth of analysis |
| `gap-identifier` | Missing areas |
| `assumption-challenger` | Challenge assumptions |

---

## Knowledge Artifacts

`.agents/` stores knowledge (18 directories):

```
.agents/
├── research/      # Exploration findings
├── plans/         # Implementation plans
├── pre-mortems/   # Failure simulations
├── specs/         # Validated specifications
├── learnings/     # Extracted lessons
├── patterns/      # Reusable patterns
├── retros/        # Retrospective reports
├── vibe/          # Validation reports
├── complexity/    # Complexity analysis
├── doc/           # Generated documentation
├── assessments/   # Quality assessments
├── reports/       # General reports
├── products/      # Product briefs
├── synthesis/     # Synthesized knowledge
├── patches/       # Patch tracking
├── bundles/       # Grouped artifacts
├── docs/          # Documentation outputs
└── ao/            # CLI state
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
3. Outputs JSON with `additionalContext` for Claude Code

---

## Installation

```bash
claude /plugin add boshu2/agentops
```

Optional:
- [beads](https://github.com/steveyegge/beads) for issue tracking
- [ao CLI](https://github.com/boshu2/ao) for full orchestration
