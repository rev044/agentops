# AgentOps

[![Version](https://img.shields.io/badge/version-0.1.0-orange)](https://github.com/boshu2/agentops/releases/tag/v0.1.0)
[![CI](https://github.com/boshu2/agentops/actions/workflows/validate.yml/badge.svg)](https://github.com/boshu2/agentops/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude_Code-2.1.12-blueviolet)](https://docs.anthropic.com/en/docs/claude-code)
[![Plugins](https://img.shields.io/badge/plugins-8-blue)](plugins/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> **v0.1.0** - Pre-release for testing. Feedback welcome!

Claude Code plugins for AI-assisted development workflows.

Built on [beads](https://github.com/steveyegge/beads) (git-native issue tracking) and [gastown](https://github.com/steveyegge/gastown) (multi-agent orchestration).

## Install

```bash
# Add marketplace
claude plugin marketplace add boshu2/agentops

# Install what you need
claude plugin install core-kit@agentops-marketplace
claude plugin install vibe-kit@agentops-marketplace
```

## Plugins

| Plugin | Skills | Purpose |
|--------|--------|---------|
| **core-kit** | `/research`, `/product`, `/formulate`, `/implement`, `/implement-wave`, `/crank`, `/retro` | The main workflow |
| **vibe-kit** | `/vibe`, `/vibe-docs`, `/validation-chain`, `/bug-hunt`, `/complexity` | Validation and quality |
| **pr-kit** | `/pr-research`, `/pr-plan`, `/pr-implement`, `/pr-validate`, `/pr-prep`, `/pr-retro` | Open source contribution |
| **beads-kit** | `/beads`, `/status`, `/molecules` | Git-based issue tracking |
| **docs-kit** | `/doc`, `/oss-docs`, `/golden-init` | Documentation generation |
| **dispatch-kit** | `/dispatch`, `/handoff`, `/mail`, `/roles` | Multi-agent orchestration |
| **gastown-kit** | `/gastown`, `/crew`, `/polecat-lifecycle`, `/bd-routing` | Gas Town worker management |
| **domain-kit** | 18 domain skills | Reference knowledge (auto-loaded) |

**Expert Agents** (vibe-kit): `security-expert`, `architecture-expert`, `code-quality-expert`, `ux-expert`

### Recommended

**Getting started:** `core-kit` + `vibe-kit`

**Full setup:** Add `beads-kit` for issue tracking, `docs-kit` for documentation

## Basic Workflow

```
/research -> /formulate -> /implement -> /vibe -> /retro
```

| Command | Purpose |
|---------|---------|
| `/research` | Explore codebase, understand the problem |
| `/formulate` | Break down into trackable beads issues |
| `/implement` | Execute a single beads issue |
| `/vibe` | Validate changes (security, quality, architecture) |
| `/retro` | Extract learnings |

### Two Types of Planning

| Type | When to Use |
|------|-------------|
| **Native plan mode** | Single-task implementation. Claude auto-enters, you review and approve, then Claude implements. |
| **/formulate** | Epic decomposition into beads issues with dependencies. For multi-issue work that `/crank` executes. |

### Autonomous Execution

`/crank` is the autonomous implementation loop. It runs until an entire epic (and all child issues) are closed:

```
/crank <epic-id>   # Runs until ALL children are CLOSED
```

- **Crew mode**: Executes issues sequentially via `/implement`
- **Mayor mode**: Dispatches to parallel workers via gastown

## Learn More

| Resource | Description |
|----------|-------------|
| [levels/](levels/) | Progressive tutorials from basics to full automation |
| [reference/](reference/) | Framework docs (PDC, FAAFO, failure patterns) |
| [12factoragentops.com](https://12factoragentops.com) | Interactive examples |

## License

MIT
