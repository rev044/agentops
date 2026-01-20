# AgentOps

```
    ___                    __  ____
   /   | ____ ____  ____  / /_/ __ \____  _____
  / /| |/ __ `/ _ \/ __ \/ __/ / / / __ \/ ___/
 / ___ / /_/ /  __/ / / / /_/ /_/ / /_/ (__  )
/_/  |_\__, /\___/_/ /_/\__/\____/ .___/____/
      /____/                    /_/
```

[![Version](https://img.shields.io/badge/version-0.1.1-orange)](https://github.com/boshu2/agentops/releases/tag/v0.1.1)
[![CI](https://github.com/boshu2/agentops/actions/workflows/validate.yml/badge.svg)](https://github.com/boshu2/agentops/actions/workflows/validate.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude_Code-2.1.12-blueviolet)](https://docs.anthropic.com/en/docs/claude-code)
[![Plugins](https://img.shields.io/badge/plugins-9-blue)](plugins/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> **v0.1.1** - Added general-kit (zero dependencies), standards library, context inference. Feedback welcome!

Claude Code plugins for AI-assisted development workflows. Just describe what you want - skills trigger automatically.

---

## Just Talk Naturally

Skills trigger from natural language. No slash commands required:

| You Say | Triggers |
|---------|----------|
| "I need to understand how auth works" | `/research` |
| "Let's plan out this feature" | `/formulate` |
| "Validate my changes" | `/vibe` |
| "Check for security issues" | `/vibe --security` |
| "What did we learn?" | `/retro` |

The skills detect intent and activate. Slash commands work too if you prefer them.

---

## The Killer Workflow: Plan → Crank

The recommended meta for complex work:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  1. PLAN SESSION                    2. CRANK SESSION                        │
│  ┌─────────────────────┐            ┌─────────────────────┐                │
│  │                     │            │                     │                │
│  │  Shift+Tab (plan)   │  handoff   │  /crank <epic>      │                │
│  │        ↓            │ ────────►  │        ↓            │                │
│  │  /formulate         │  (fresh    │  wave 1: issues     │                │
│  │        ↓            │  context)  │  wave 2: issues     │                │
│  │  creates beads      │            │  wave 3: issues     │                │
│  │  organizes waves    │            │        ↓            │                │
│  │                     │            │  ALL CLOSED         │                │
│  └─────────────────────┘            └─────────────────────┘                │
│                                                                             │
│  Planning uses context.             Execution gets fresh context +          │
│  When done, clear it.               plan artifact + pointer to old convo.   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### The Steps

1. **Shift+Tab** to enter Claude's native plan mode
2. **`/formulate`** to invoke the skill (slash commands guarantee activation)
3. Claude explores, creates beads issues, organizes into dependency waves
4. **Handoff** clears context and starts fresh crank session
5. **`/crank <epic>`** executes all waves until done

### Why This Works

- **Plan mode + /formulate** gives you review gates before execution
- **Explicit `/slash` commands** guarantee the skill triggers (natural language works but isn't 100%)
- **Handoff clears context** - planning burns tokens exploring, crank needs room to execute
- **Crank session gets**: the plan artifact, the beads, and a pointer to the old conversation JSONL if needed

This is how we run everything. Plan in one session, crank in a fresh one.

### Example

```
You: [Shift+Tab to enter plan mode]
You: /formulate "add OAuth support to the API"

Claude: [explores codebase, understands patterns]
        [creates epic with 12 child issues]
        [organizes into 4 waves by dependency]
        [presents plan for review]

You: [approve plan]

Claude: Ready to crank. Starting fresh session...
        [clears context, hands off plan]

--- new session ---

Claude: [receives plan artifact]
        [cranks wave 1: 3 issues in parallel]
        [cranks wave 2: 4 issues in parallel]
        [cranks wave 3: 3 issues in parallel]
        [cranks wave 4: 2 issues in parallel]

        Epic complete. 12/12 issues closed.
```

---

## The Full Workflow

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
| **Research** | `/research` | Deep codebase exploration |
| **Plan** | `/formulate` | Create issues with dependencies, organize waves |
| **Execute** | `/crank` | Run all waves until epic is closed |
| **Validate** | `/vibe` | Check security, quality, architecture |
| **Learn** | `/retro` | Extract learnings for future sessions |

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

**Want the full plan→crank workflow?**
```bash
brew install beads
claude plugin install core-kit@agentops-marketplace
claude plugin install beads-kit@agentops-marketplace
```

**Multi-agent parallel execution?**
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

## Learn More

| Resource | Description |
|----------|-------------|
| [levels/](levels/) | Progressive tutorials from basics to full automation |
| [reference/](reference/) | Framework docs (PDC, FAAFO, failure patterns) |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |

---

## License

MIT
