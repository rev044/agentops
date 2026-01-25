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

## Skills (20)

| Skill | What It Does |
|-------|--------------|
| `/research` | Deep codebase exploration |
| `/plan` | Decompose goals into trackable issues |
| `/implement` | Execute a single issue |
| `/crank` | Autonomous multi-issue execution |
| `/vibe` | Code validation (security, quality, architecture) |
| `/retro` | Extract learnings from completed work |
| `/post-mortem` | Full validation + knowledge extraction |
| `/pre-mortem` | Simulate failures before implementing |
| `/beads` | Git-native issue tracking |
| `/bug-hunt` | Root cause analysis |
| `/knowledge` | Query knowledge artifacts |
| `/complexity` | Code complexity analysis |
| `/doc` | Documentation generation |
| `/forge` | Knowledge forge operations |
| `/extract` | Extract decisions/learnings from transcripts |
| `/inject` | Inject knowledge into context |
| `/ratchet` | Brownian ratchet progress gates |
| `/flywheel` | Knowledge flywheel orchestration |
| `/provenance` | Track artifact provenance |
| `/using-agentops` | Usage guide |

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

## ao CLI

The ao CLI provides workflow orchestration outside Claude sessions.

**Install:**
```bash
brew install agentops      # From tap
# or
./scripts/install-ao.sh    # From source
```

**Commands:**
```bash
# Knowledge Forge
ao forge search <query>      # Semantic search across knowledge
ao forge index <path>        # Index knowledge artifacts
ao forge queue               # View pending extractions

# Ratchet (Progress Gates)
ao ratchet check <skill>     # Verify prerequisites
ao ratchet record <type>     # Record progress event
ao ratchet verify <epic>     # Verify epic completion

# Context Management
ao extract                   # Extract from last session
ao inject                    # Inject knowledge into context
ao status                    # Show session/knowledge status
```

See [docs/brownian-ratchet.md](docs/brownian-ratchet.md) for the philosophy.

## Requirements

- [Claude Code](https://github.com/anthropics/claude-code) v1.0+
- Optional: [beads](https://github.com/steveyegge/beads) for issue tracking
- Optional: Go 1.22+ (to build ao CLI from source)

## License

MIT
