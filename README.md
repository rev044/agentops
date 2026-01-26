# AgentOps

AgentOps is a complete knowledge management system for your coding agents, built on top of composable "skills" and a memory layer that makes sure your AI actually learns from your codebase.

## How it works

It starts the moment you fire up Claude Code. Before you even type anything, AgentOps has already searched your `.agents/` directory for relevant learnings from past sessions and injected them into context. That OAuth bug you debugged two weeks ago? Claude already knows about it.

As you work, the skills kick in automatically. When you're about to build something, `/research` activates and deep-scans your codebase first. When you need to break down a feature, `/plan` creates tracked issues. When you say "go", `/crank` takes over and autonomously works through your plan—implementing, validating, and committing each piece.

At the end of your session, AgentOps extracts what you learned. That tricky edge case you discovered? That pattern that finally worked? It's captured in `.agents/learnings/` and will be there next time you need it. Git-tracked. Permanent. Searchable.

The magic is the flywheel: **knowledge compounds**. Each session makes the next one smarter. After a few weeks, Claude knows your codebase like a senior engineer who's been on the team for months.

And because the skills trigger automatically, you don't need to do anything special. Your coding agent just has a memory.

## Installation

```bash
# Install the CLI (manages your knowledge base)
brew install boshu2/agentops/agentops

# Add the plugin to Claude Code
claude mcp add boshu2/agentops

# Initialize in your repo
ao init

# Set up auto-hooks (one time)
ao hooks install
```

### Verify Installation

```bash
ao badge
```

You should see a knowledge dashboard. If you're starting fresh, it'll show 🌱 STARTING. That's normal—the flywheel needs sessions to turn.

## The Basic Workflow

1. **Session starts** → AgentOps automatically injects relevant knowledge from past sessions. Claude already knows your patterns.

2. **You describe what you want** → `/research` activates. Deep-scans your codebase, loads prior learnings, understands context before writing code.

3. **You approve the approach** → `/plan` breaks work into bite-sized issues. Each one has clear acceptance criteria.

4. **You say "go"** → `/crank` takes over. Implements each issue, validates with `/vibe`, commits, moves to the next. Hours of autonomous work.

5. **Session ends** → AgentOps extracts learnings. Patterns discovered, decisions made, edge cases found—all saved to `.agents/`.

6. **Next session** → Starts at step 1, but smarter. The flywheel turns.

**The agent checks for relevant skills before any task.** Mandatory workflows, not suggestions.

## What's Inside

### Skills Library

**Core Workflow**
- **research** - Deep codebase exploration before writing code
- **plan** - Decompose goals into tracked issues
- **implement** - Execute a single issue with full lifecycle
- **crank** - Autonomous multi-issue execution (the "ship it" button)
- **vibe** - Validate code quality, security, architecture

**Knowledge Management**
- **forge** - Mine transcripts for knowledge
- **inject** - Load relevant knowledge into session
- **retro** - Extract learnings from completed work
- **post-mortem** - Full validation + knowledge extraction
- **knowledge** - Query knowledge artifacts

**Risk & Quality**
- **pre-mortem** - Simulate failures before building
- **bug-hunt** - Systematic root cause analysis
- **complexity** - Find refactor targets

**Documentation**
- **doc** - Generate documentation
- **oss-docs** - Scaffold OSS documentation
- **golden-init** - Initialize repos with best practices

**Issue Tracking**
- **beads** - Git-native issue tracking
- **status** - Quick status check

**Open Source Contribution**
- **pr-research** - Upstream codebase exploration
- **pr-prep** - Prepare PRs with proper context

### Domain Expert Agents

When you need specialized review, AgentOps has 6 domain experts:

- **security-expert** - OWASP Top 10, vulnerability assessment
- **architecture-expert** - System design, cross-cutting concerns
- **code-quality-expert** - Complexity, patterns, maintainability
- **ux-expert** - Accessibility, UX validation
- **code-reviewer** - Code review analysis
- **security-reviewer** - Security-focused review

### The ao CLI

```bash
ao badge              # Knowledge flywheel health
ao inject             # Manually inject knowledge
ao search "oauth"     # Search your learnings
ao forge transcript   # Extract from past sessions
ao hooks install      # Set up auto-hooks
```

## The Knowledge Flywheel

```
╔═══════════════════════════════════════════╗
║         🏛️  AGENTOPS KNOWLEDGE             ║
╠═══════════════════════════════════════════╣
║  Retrieval (σ)     │  0.72  ███████░░░   ║
║  Citation Rate (ρ) │  0.34  ███░░░░░░░   ║
║  Decay (δ)         │  0.17  █░░░░░░░░░   ║
╠═══════════════════════════════════════════╣
║  σ×ρ = 0.24 > δ    │  🚀 COMPOUNDING     ║
╚═══════════════════════════════════════════╝
```

- **σ (sigma)** — When you need knowledge, how often is it found?
- **ρ (rho)** — When found, how often is it actually used?
- **δ (delta)** — Knowledge fades at ~17%/week without use

**Escape velocity:** When `σ × ρ > δ`, knowledge compounds faster than it decays. That's the goal.

| Status | Meaning |
|--------|---------|
| 🌱 STARTING | Just installed. Keep using it. |
| 📈 BUILDING | Flywheel turning. Approaching escape velocity. |
| 🚀 COMPOUNDING | Knowledge grows faster than it decays. |

## Knowledge Storage

Everything lives in `.agents/` (git-tracked, permanent):

```
.agents/
├── learnings/    # What we learned
├── patterns/     # How we solved it
├── research/     # What we found
├── retros/       # What went wrong/right
└── ao/
    ├── sessions/ # Mined transcripts
    └── index/    # Search index
```

## Philosophy

- **Memory over amnesia** - Your AI should remember what worked
- **Compound over reset** - Each session builds on the last
- **Automatic over manual** - The flywheel turns itself
- **Git-tracked over ephemeral** - Knowledge survives sessions, machines, team changes

Read more: [The Science Behind the Flywheel](docs/the-science.md)

## Requirements

- [Claude Code](https://github.com/anthropics/claude-code) v1.0+
- Optional: [beads](https://github.com/beads-ai/beads) for issue tracking

## Contributing

Skills live directly in this repository. To contribute:

1. Fork the repository
2. Create a branch for your skill
3. Follow the skill template in `skills/`
4. Submit a PR

## Updating

```bash
brew upgrade agentops
```

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: https://github.com/boshu2/agentops/issues
- **Releases**: https://github.com/boshu2/agentops/releases

---

*Stop renting intelligence. Own it.*
