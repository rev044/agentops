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
| [GSD](comparisons/agentops-vs-gsd) | Spec-driven execution, 7 runtimes | Cross-session memory |
| [Compound Engineer](comparisons/agentops-vs-compound-engineer) | Knowledge compounding loop | Automated flywheel + validation gates |
| [Superpowers](comparisons/agentops-vs-superpowers) | Strict TDD enforcement | Memory that compounds + pre-mortem |
| [Claude-Flow / Ruflo](comparisons/agentops-vs-claude-flow) | 54+ agent swarms, WASM perf | Knowledge layer across sessions |
| [SDD Tools](comparisons/agentops-vs-sdd) | Spec as source of truth | Learning extraction + failure prevention |

[Full comparison matrix →](comparisons/)

## Links

- [GitHub Repository](https://github.com/boshu2/agentops)
- [Install Guide](https://github.com/boshu2/agentops#install)
- [CLI Reference](https://github.com/boshu2/agentops/blob/main/cli/docs/COMMANDS.md)
- [Changelog](https://github.com/boshu2/agentops/blob/main/docs/CHANGELOG.md)
