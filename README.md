# AgentOps

```
    ___                    __  ____
   /   | ____ ____  ____  / /_/ __ \____  _____
  / /| |/ __ `/ _ \/ __ \/ __/ / / / __ \/ ___/
 / ___ / /_/ /  __/ / / / /_/ /_/ / /_/ (__  )
/_/  |_\__, /\___/_/ /_/\__/\____/ .___/____/
      /____/                    /_/
```

[![Version](https://img.shields.io/badge/version-0.1.0-orange)](https://github.com/boshu2/agentops/releases/tag/v0.1.0)
[![CI](https://github.com/boshu2/agentops/actions/workflows/validate.yml/badge.svg)](https://github.com/boshu2/agentops/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude_Code-2.1.12-blueviolet)](https://docs.anthropic.com/en/docs/claude-code)
[![Plugins](https://img.shields.io/badge/plugins-9-blue)](plugins/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> **v0.1.0** - Pre-release for testing. Feedback welcome!

Claude Code plugins for AI-assisted development workflows. Research, plan, implement, validate, learn.

---

## The Workflow

```
┌──────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│   /research  ──►  /formulate  ──►  /implement  ──►  /vibe  ──►  /retro  │
│                                                                          │
│   understand      break down       execute          validate    extract  │
│   the problem     into issues      the work         changes     lessons  │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

| Stage | Command | What It Does |
|-------|---------|--------------|
| **Research** | `/research` | Deep codebase exploration, understand the problem |
| **Plan** | `/formulate` | Break down into trackable issues with dependencies |
| **Execute** | `/implement` | Execute a single issue end-to-end |
| **Validate** | `/vibe` | Check security, quality, architecture, accessibility |
| **Learn** | `/retro` | Extract and preserve learnings |

**Autonomous mode**: `/crank <epic>` runs the full loop until all issues are closed.

---

## Install

```bash
# Add marketplace
claude plugin marketplace add boshu2/agentops

# Start with general-kit (no dependencies)
claude plugin install general-kit@agentops-marketplace

# Add more as needed
claude plugin install core-kit@agentops-marketplace
claude plugin install vibe-kit@agentops-marketplace
```

---

## Plugins

### Dependency Tiers

| Tier | Plugins | Requirements |
|------|---------|--------------|
| **Standalone** | `general-kit`, `domain-kit` | None - works everywhere |
| **Beads** | `core-kit`, `vibe-kit`, `beads-kit`, `pr-kit`, `docs-kit` | [beads](https://github.com/steveyegge/beads) CLI |
| **Gas Town** | `gastown-kit`, `dispatch-kit` | [gastown](https://github.com/steveyegge/gastown) CLI |

> **Note**: This repo provides plugins FOR beads and gastown - it's not built on them. Start with `general-kit` which has no dependencies.

### All Plugins

| Plugin | Skills | Purpose |
|--------|--------|---------|
| **general-kit** | `/research`, `/vibe`, `/vibe-docs`, `/bug-hunt`, `/complexity`, `/validation-chain`, `/doc`, `/oss-docs`, `/golden-init` | **Start here** - no dependencies |
| **core-kit** | `/plan`, `/product`, `/formulate`, `/implement`, `/implement-wave`, `/crank`, `/retro` | Structured workflow |
| **vibe-kit** | `/vibe`, `/vibe-docs`, `/validation-chain`, `/bug-hunt`, `/complexity` | Validation and quality |
| **pr-kit** | `/pr-research`, `/pr-plan`, `/pr-implement`, `/pr-validate`, `/pr-prep`, `/pr-retro` | Open source contribution |
| **beads-kit** | `/beads`, `/status`, `/molecules` | Git-based issue tracking |
| **docs-kit** | `/doc`, `/doc-creator`, `/code-map-standard`, `/oss-docs`, `/golden-init` | Documentation generation |
| **dispatch-kit** | `/dispatch`, `/handoff`, `/mail`, `/roles` | Multi-agent communication |
| **gastown-kit** | `/gastown`, `/crew`, `/polecat-lifecycle`, `/bd-routing` | Parallel worker orchestration |
| **domain-kit** | 21 domain skills + `standards` library | Reference knowledge (auto-loaded) |

**Expert Agents** (general-kit, vibe-kit): `security-expert`, `architecture-expert`, `code-quality-expert`, `ux-expert`

---

## Recommended Setup

**Just exploring?**
```bash
claude plugin install general-kit@agentops-marketplace
```
Research, validation, documentation, expert agents - no external tools needed.

**Want structured workflows?**
```bash
brew install beads
claude plugin install core-kit@agentops-marketplace
claude plugin install beads-kit@agentops-marketplace
```

**Full multi-agent setup?**
```bash
brew install beads gastown
claude plugin install gastown-kit@agentops-marketplace
claude plugin install dispatch-kit@agentops-marketplace
```

---

## OpenCode Compatibility

These plugins can be translated to [opencode](https://github.com/opencode-ai/opencode) skills and commands:

```bash
# Use opencode's translation tool
opencode translate ./plugins/general-kit --format opencode
```

The skill format is designed to be portable across AI coding assistants.

---

## Two Types of Planning

| Type | When to Use |
|------|-------------|
| **Native plan mode** | Single task. Claude auto-enters plan mode, you review, Claude implements. |
| **/formulate** | Epic decomposition. Creates beads issues with dependencies for `/crank` to execute. |

---

## Learn More

| Resource | Description |
|----------|-------------|
| [levels/](levels/) | Progressive tutorials from basics to full automation |
| [reference/](reference/) | Framework docs (PDC, FAAFO, failure patterns) |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |

---

## License

MIT
