# AgentOps Architecture

> AI-assisted development workflows for Claude Code.

## Overview

AgentOps is a single Claude Code plugin providing the RPI workflow with Knowledge Flywheel.

```
.
├── .claude-plugin/
│   └── plugin.json      # Plugin manifest
├── skills/              # All 14 skills
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
├── agents/              # Subagent definitions
│   ├── code-reviewer.md
│   ├── security-reviewer.md
│   ├── architecture-expert.md
│   ├── code-quality-expert.md
│   ├── security-expert.md
│   └── ux-expert.md
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

Located in `agents/`. Used by `/vibe` and other validation skills.

| Agent | Focus |
|-------|-------|
| `code-reviewer` | Quality, patterns, maintainability |
| `security-reviewer` | Vulnerabilities, OWASP |
| `architecture-expert` | System design, cross-cutting |
| `code-quality-expert` | Complexity, refactoring |
| `security-expert` | Deep security analysis |
| `ux-expert` | Accessibility, user experience |

---

## Knowledge Artifacts

`.agents/` stores knowledge:

```
.agents/
├── research/     # Exploration findings
├── learnings/    # Extracted lessons (JSONL or MD)
├── patterns/     # Reusable patterns
├── retros/       # Retrospective reports
└── products/     # Product briefs
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
