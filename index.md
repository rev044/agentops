---
title: "AgentOps — The Bookkeeping Layer for Coding Agents"
description: "Memory, validation, and feedback loops that compound between sessions. Hooks into Claude Code, Codex, and OpenCode."
layout: default
---

# AgentOps

**The bookkeeping layer for coding agents.** Memory, validation, and feedback loops that compound between sessions.

Agents are fast, capable, and forgetful. Session 1, your agent debugs a bug for 2 hours. Session 15, a different agent hits the same bug and starts from scratch. AgentOps fixes this.

## Quick Start

```bash
claude plugin marketplace add boshu2/agentops
claude plugin install agentops@agentops-marketplace
```

Then type `/quickstart` in your agent chat.

## How It Compares

| Tool | Focus | What AgentOps Adds |
|------|-------|--------------------|
| [GSD](docs/comparisons/vs-gsd.md) | Spec-driven execution, 7 runtimes | Cross-session memory |
| [Compound Engineer](docs/comparisons/vs-compound-engineer.md) | Knowledge compounding loop | Automated flywheel + validation gates |
| [Superpowers](docs/comparisons/vs-superpowers.md) | Strict TDD enforcement | Memory that compounds + pre-mortem |
| [Claude-Flow / Ruflo](docs/comparisons/vs-claude-flow.md) | 54+ agent swarms, WASM perf | Knowledge layer across sessions |
| [SDD Tools](docs/comparisons/vs-sdd.md) | Spec as source of truth | Learning extraction + failure prevention |

[Full comparison matrix →](docs/comparisons/README.md)

## Links

- [GitHub Repository](https://github.com/boshu2/agentops)
- [Install Guide](https://github.com/boshu2/agentops#install)
- [CLI Reference](https://github.com/boshu2/agentops/blob/main/cli/docs/COMMANDS.md)
- [Changelog](https://github.com/boshu2/agentops/blob/main/docs/CHANGELOG.md)
