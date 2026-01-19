# AgentOps 🤖

[![CI](https://github.com/boshu2/agentops/actions/workflows/validate.yml/badge.svg)](https://github.com/boshu2/agentops/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude_Code-2.1.12-blueviolet)](https://docs.anthropic.com/en/docs/claude-code)
[![Plugins](https://img.shields.io/badge/plugins-8-blue)](plugins/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Claude Code plugins for AI-assisted development workflows.

Built on [beads](https://github.com/steveyegge/beads) (git-native issue tracking) and [gastown](https://github.com/steveyegge/gastown) (multi-agent orchestration).

## 📦 Install

```bash
# Add marketplace
claude plugin marketplace add boshu2/agentops

# Install what you need
claude plugin install core-kit@agentops-marketplace
claude plugin install vibe-kit@agentops-marketplace
```

## 🧩 Plugins

| Plugin | What it does |
|--------|--------------|
| 🔧 **core-kit** | `/research`, `/plan`, `/implement`, `/crank` - the main workflow |
| ✅ **vibe-kit** | `/vibe`, `/bug-hunt`, `/complexity` - validation and quality |
| 🔀 **pr-kit** | `/pr-research` → `/pr-retro` - open source contribution flow |
| 📋 **beads-kit** | `/beads`, `/status` - git-based issue tracking |
| 📝 **docs-kit** | `/doc`, `/oss-docs` - documentation generation |
| 📬 **dispatch-kit** | `/handoff`, `/mail` - multi-agent orchestration |
| 🏭 **gastown-kit** | `/gastown`, `/crew` - Gas Town worker management |
| 🌐 **domain-kit** | Reference knowledge across 17 domains |

### 💡 Recommended

**Getting started:** `core-kit` + `vibe-kit`

**Full setup:** Add `beads-kit` for issue tracking, `docs-kit` for documentation

## 🔄 Basic Workflow

```
/research → /plan → /implement → /retro
```

| Command | Purpose |
|---------|---------|
| 🔍 `/research` | Explore codebase, understand the problem |
| 📐 `/plan` | Break down into trackable beads issues |
| ⚡ `/implement` | Execute a **single** beads issue |
| 🎓 `/retro` | Extract learnings |

### 🚀 Autonomous Execution

`/crank` is the autonomous implementation loop. It runs until an entire epic (and all child issues) are closed:

```
/crank <epic-id>   # Runs until ALL children are CLOSED
```

- **Crew mode**: Executes issues sequentially via `/implement`
- **Mayor mode**: Dispatches to parallel workers via gastown

## 📚 Learn More

| Resource | Description |
|----------|-------------|
| 📖 [levels/](levels/) | Progressive tutorials from basics to full automation |
| 📋 [reference/](reference/) | Framework docs (PDC, FAAFO, failure patterns) |
| 🌐 [12factoragentops.com](https://12factoragentops.com) | Interactive examples |

## 📄 License

MIT
