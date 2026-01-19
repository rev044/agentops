# AgentOps

[![CI](https://github.com/boshu2/agentops/actions/workflows/validate.yml/badge.svg)](https://github.com/boshu2/agentops/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Claude Code plugins for AI-assisted development workflows.

## Install

```bash
# Add marketplace
claude plugin marketplace add boshu2/agentops

# Install what you need
claude plugin install core-kit@agentops-marketplace
claude plugin install vibe-kit@agentops-marketplace
```

## Plugins

| Plugin | What it does |
|--------|--------------|
| **core-kit** | `/research`, `/plan`, `/implement`, `/crank` - the main workflow |
| **vibe-kit** | `/vibe`, `/bug-hunt`, `/complexity` - validation and quality |
| **pr-kit** | `/pr-research` → `/pr-retro` - open source contribution flow |
| **beads-kit** | `/beads`, `/status` - git-based issue tracking |
| **docs-kit** | `/doc`, `/oss-docs` - documentation generation |
| **dispatch-kit** | `/handoff`, `/mail` - multi-agent orchestration |
| **gastown-kit** | `/gastown`, `/crew` - Gas Town worker management |
| **domain-kit** | Reference knowledge across 17 domains |

### Recommended

**Getting started:** `core-kit` + `vibe-kit`

**Full setup:** Add `beads-kit` for issue tracking, `docs-kit` for documentation

## Basic Workflow

```
/research → /plan → /implement → /retro
```

- `/research` - Explore codebase, understand the problem
- `/plan` - Break down into trackable issues
- `/implement` - Execute with validation
- `/retro` - Extract learnings

For autonomous execution: `/crank` runs the full cycle unattended.

## Learn More

- **[levels/](levels/)** - Progressive tutorials from basics to full automation
- **[reference/](reference/)** - Framework docs (PDC, FAAFO, failure patterns)
- **[12factoragentops.com](https://12factoragentops.com)** - Interactive examples

## License

MIT
