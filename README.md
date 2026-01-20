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
4. **Review the plan** - you can:
   - Tweak it ("move issue X to wave 2")
   - Reject it ("start over, different approach")
   - Accept it ("looks good")
   - **Accept and clear context** ← this is the handoff
5. **Fresh session** receives the plan artifact and runs `/crank <epic>`

### The Handoff

When you accept the plan, Claude offers to clear context and hand off. This is key:

```
Claude: Plan ready. 12 issues across 4 waves.

        Options:
        - [Accept] Continue in this session
        - [Accept + Clear] Start fresh crank session (recommended)
        - [Revise] Make changes

You: [Accept + Clear]

Claude: Handing off to fresh session...
        - Plan artifact saved
        - Beads issues created
        - Conversation JSONL path noted

--- new session with fresh context ---
```

### Why This Works

- **Plan mode + /formulate** gives you review gates before execution
- **Explicit `/slash` commands** guarantee the skill triggers (natural language works but isn't 100%)
- **"Accept + Clear" handoff** - planning burns tokens exploring, crank needs room to execute
- **Fresh session gets**: the plan artifact, the beads, and a pointer to the old conversation JSONL if needed

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

# Start with solo-kit (any language, any project)
/plugin install solo-kit@agentops

# Add language-specific support
/plugin install python-kit@agentops    # if Python
/plugin install go-kit@agentops        # if Go
/plugin install typescript-kit@agentops # if TypeScript
/plugin install shell-kit@agentops     # if Shell/Bash
```

---

## Tiered Architecture

AgentOps scales from solo developer to multi-agent orchestration.

### Tier 1: Solo Developer (Any Project)

| Plugin | Skills | Purpose |
|--------|--------|---------|
| **solo-kit** | `/research`, `/vibe`, `/bug-hunt`, `/complexity`, `/doc`, `/oss-docs`, `/golden-init` | **Start here** - essential validation and exploration |

**Agents:** `code-reviewer`, `security-reviewer`
**Hooks:** Auto-format, debug statement warnings, git push review

```bash
/plugin install solo-kit@agentops
```

### Tier 2: Language Kits (Plug-in Based on Project)

| Plugin | Standards | Hooks | Purpose |
|--------|-----------|-------|---------|
| **python-kit** | `python.md` | ruff, mypy | Python development |
| **go-kit** | `go.md` | gofmt, golangci-lint | Go development |
| **typescript-kit** | `typescript.md` | prettier, tsc | TypeScript/JavaScript |
| **shell-kit** | `shell.md` | shellcheck | Shell scripting |

Each adds language-specific standards, linting hooks, and patterns.

### Tier 3: Team Workflows

| Plugin | Skills | Purpose | Requires |
|--------|--------|---------|----------|
| **beads-kit** | `/beads`, `/status`, `/molecules` | Git-based issue tracking | [beads](https://github.com/steveyegge/beads) |
| **pr-kit** | `/pr-research`, `/pr-plan`, `/pr-implement`, `/pr-validate` | PR workflows | beads |
| **dispatch-kit** | `/dispatch`, `/handoff`, `/mail`, `/roles` | Multi-agent coordination | beads |

### Tier 4: Multi-Agent Orchestration

| Plugin | Skills | Purpose | Requires |
|--------|--------|---------|----------|
| **crank-kit** | `/plan`, `/formulate`, `/implement`, `/implement-wave`, `/crank`, `/retro` | Autonomous execution | beads |
| **gastown-kit** | `/gastown`, `/crew`, `/polecat-lifecycle`, `/bd-routing` | Parallel workers | beads + [gastown](https://github.com/steveyegge/gastown) |

---

## Upgrade Path

```
solo-kit              ← Any developer starts here
    ↓
+ language-kit        ← Add your language(s)
    ↓
+ beads-kit           ← Track work across sessions
    ↓
+ crank-kit           ← Autonomous execution
    ↓
+ gastown-kit         ← Multi-agent orchestration
```

---

## Legacy Plugins

These plugins are being refactored into the tiered structure:

| Legacy | Migrates To |
|--------|-------------|
| `general-kit` | `solo-kit` |
| `core-kit` | `solo-kit` + `crank-kit` |
| `vibe-kit` | `solo-kit` |
| `domain-kit` | Language kits |
| `docs-kit` | `solo-kit`

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
