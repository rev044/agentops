# AgentOps

AI-assisted development workflows for Claude Code.

## The Problem

You're deep in a codebase, trying to add a feature. You've got 20 tabs open, grep results scattered across terminals, and you're losing context faster than you can build it.

AgentOps changes this. It gives Claude a structured workflow that builds knowledge over time instead of losing it.

## Install

```bash
claude /plugin add boshu2/agentops
```

## The RPI Workflow

```
Research → Plan → Implement → Validate
    ↑                            │
    └──── Knowledge Flywheel ────┘
```

Every time you complete work, learnings feed back into research. Your AI assistant gets smarter about YOUR codebase.

## Skills

| Skill | What It Does |
|-------|--------------|
| `/research` | Deep codebase exploration |
| `/plan` | Decompose goals into trackable issues |
| `/implement` | Execute a single issue |
| `/crank` | Autonomous multi-issue execution |
| `/vibe` | Code validation (security, quality, architecture) |
| `/retro` | Extract learnings from completed work |
| `/post-mortem` | Full validation + knowledge extraction |
| `/beads` | Git-native issue tracking |
| `/bug-hunt` | Root cause analysis |
| `/knowledge` | Query knowledge artifacts |
| `/complexity` | Code complexity analysis |
| `/doc` | Documentation generation |
| `/pre-mortem` | Simulate failures before implementing |

## Natural Language

Just describe what you want:

> "I need to understand how auth works" → `/research`

> "Check my code for issues" → `/vibe`

> "What could go wrong with this design?" → `/pre-mortem`

> "Execute this epic" → `/crank`

## Knowledge Artifacts

AgentOps stores knowledge in `.agents/`:

```
.agents/
├── research/     # Exploration findings
├── learnings/    # Extracted lessons
├── patterns/     # Reusable patterns
├── retros/       # Retrospective reports
└── products/     # Product briefs
```

Future `/research` commands discover these automatically.

## ao CLI Integration

For full workflow orchestration, install the [ao CLI](https://github.com/boshu2/ao):

```bash
brew install agentops
```

The ao CLI provides:
- `ao forge search` - Semantic knowledge search
- `ao forge index` - Index knowledge artifacts
- `ao ratchet` - Track progress with the Brownian Ratchet pattern

## Requirements

- [Claude Code](https://github.com/anthropics/claude-code) v1.0+
- Optional: [beads](https://github.com/steveyegge/beads) for issue tracking
- Optional: [ao CLI](https://github.com/boshu2/ao) for full workflow orchestration

## License

MIT
