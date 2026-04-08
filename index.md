---
title: "AgentOps — The Bookkeeping Layer for Coding Agents"
description: "Memory, validation, and feedback loops that compound between sessions. Hooks into Claude Code, Codex, and OpenCode."
layout: home
---

# AgentOps

**The bookkeeping layer for coding agents.** Memory, validation, and feedback loops that compound between sessions.

Agents are fast, capable, and forgetful. Session 1, your agent debugs a bug for 2 hours. Session 15, a different agent hits the same bug and starts from scratch. AgentOps fixes this.

- **Validation** — independent councils challenge your plan *and* your code before it ships
- **Memory** — solved problems stay solved. Your repo accumulates knowledge across sessions, agents, and runtimes
- **Loop closure** — every completed session produces better context for the next one

[GitHub Repository](https://github.com/boshu2/agentops) · [Install](https://github.com/boshu2/agentops#install) · [Changelog](https://github.com/boshu2/agentops/blob/main/docs/CHANGELOG.md)

---

## Quick Start

```bash
claude plugin marketplace add boshu2/agentops
claude plugin install agentops@agentops-marketplace
```

Then type `/quickstart` in your agent chat.

New here? Start with the [Newcomer Guide](docs/newcomer-guide.md) or the [Philosophy](docs/philosophy.md) behind the tool.

---

## Competitive Comparisons

How AgentOps stacks up against other AI coding agent tools:

| Tool | Focus | What AgentOps Adds |
|------|-------|--------------------|
| [GSD](docs/comparisons/vs-gsd.md) | Spec-driven execution, 7 runtimes | Cross-session memory |
| [Compound Engineer](docs/comparisons/vs-compound-engineer.md) | Knowledge compounding loop | Automated flywheel + validation gates |
| [Superpowers](docs/comparisons/vs-superpowers.md) | Strict TDD enforcement | Memory that compounds + pre-mortem |
| [Claude-Flow / Ruflo](docs/comparisons/vs-claude-flow.md) | 54+ agent swarms, WASM perf | Knowledge layer across sessions |
| [SDD Tools](docs/comparisons/vs-sdd.md) | Spec as source of truth | Learning extraction + failure prevention |

[Full comparison matrix →](docs/comparisons/README.md)

---

## Documentation

### Getting Started
- [Newcomer Guide](docs/newcomer-guide.md) — practical orientation
- [Quickstart](docs/README.md) — repo-level docs index
- [Philosophy](docs/philosophy.md) — why AgentOps exists
- [How It Works](docs/how-it-works.md) — the mechanism behind the flywheel
- [The Science](docs/the-science.md) — formal foundations

### Core Concepts
- [Architecture](docs/ARCHITECTURE.md) — five pillars and operational invariants
- [Context Lifecycle](docs/context-lifecycle.md) — three-tier injection model
- [Knowledge Flywheel](docs/knowledge-flywheel.md) — how memory compounds
- [Brownian Ratchet](docs/brownian-ratchet.md) — the mental model
- [Origin Story](docs/origin-story.md) — how we got here

### Reference
- [Skills Catalog](docs/SKILLS.md) — all 50+ skills
- [Skill Router](docs/SKILL-ROUTER.md) — which skill to run when
- [CLI Skills Map](docs/cli-skills-map.md) — ao CLI → skill mapping
- [Glossary](docs/GLOSSARY.md) — terminology reference
- [Environment Variables](docs/ENV-VARS.md) — configuration reference
- [FAQ](docs/FAQ.md) — common questions

### Operations
- [Evolve Setup](docs/evolve-setup.md) — fitness-scored improvement loop
- [Software Factory](docs/software-factory.md) — operator lane
- [Incident Runbook](docs/INCIDENT-RUNBOOK.md) — when things break
- [Troubleshooting](docs/troubleshooting.md) — common problems
- [Testing](docs/TESTING.md) — test architecture
- [Security](docs/SECURITY.md) — security model
- [CI/CD](docs/CI-CD.md) — pipeline overview

### Deep Dives
- [Agent Footguns](docs/agent-footguns.md) — what to avoid
- [Leverage Points](docs/leverage-points.md) — high-impact interventions
- [Scale Without Swarms](docs/scale-without-swarms.md) — alternatives to parallelism
- [Curation Pipeline](docs/curation-pipeline.md) — how knowledge gets promoted
- [Context Packet](docs/context-packet.md) — what gets injected

### Contributing
- [Contributing Guide](docs/CONTRIBUTING.md)
- [Code of Conduct](docs/CODE_OF_CONDUCT.md)
- [Releasing](docs/RELEASING.md)
- [Changelog](docs/CHANGELOG.md)
