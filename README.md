<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**A knowledge flywheel for AI coding agents — your agent remembers across sessions.**

Maximize flow. Shorten feedback loops. Compound what you learn.

[Install](#install) · [Tiers](#choose-your-tier) · [What This Is](#what-this-is) · [Quick Start](#quick-start) · [How It Works](#how-it-works) · [Docs](docs/)

</div>

---

## Install

```bash
npx skills@latest add boshu2/agentops --all -g
```

Or with Claude Code:
```bash
claude plugin add boshu2/agentops
```

## Choose Your Tier

All skills install together. Tiers are about which tools to **use first** and when to add dependencies:

| Tier | What to Use | Extra Dependencies | When to Graduate |
|------|-------------|-------------------|-----------------|
| **Tier 0** | `/research`, `/pre-mortem`, `/vibe` | None | When you find yourself re-explaining context every session |
| **Tier 1** | + knowledge flywheel, session hooks | `ao` CLI | When you have multi-issue epics to track |
| **Tier 2** | + issue tracking, epic orchestration | `ao` + `beads` | When you want cross-vendor validation |
| **Tier 3** | + cross-vendor consensus (`--mixed`) | `ao` + `beads` + `codex` | You're at full power |

**Start at Tier 0.** Add tools as you need them.

---

## What This Is

Every AI coding session starts from zero. AgentOps changes that. Learnings from each session persist to `.agents/` (git-tracked), and the next session starts with that knowledge automatically injected. Each session feeds the next — knowledge compounds.

**What's automatic vs what you run:**

| | What | How |
|---|---|---|
| **Auto** | Inject prior knowledge at session start | Hook — runs before you type anything |
| **Auto** | Extract learnings at session end | Hook — runs when session closes |
| **You** | `/pre-mortem` → `/crank` → `/vibe` → `/post-mortem` | Slash commands you invoke |

The hooks close the loop. Without them, you have a pipeline. With them, you have a flywheel — each session feeds the next.

<details>
<summary>View the knowledge flywheel</summary>

```
                    THE KNOWLEDGE FLYWHEEL

  SESSION START
  +------------------------------------+
  | auto-inject prior knowledge        |  <-- hook (automatic)
  +------------------+-----------------+
                     |
                     v
  +------------------------------------+
  | /pre-mortem   catch risks early    |  <-- you run this
  +------------------+-----------------+
                     |
                     v
  +------------------------------------+
  | /crank        parallel execution   |  <-- you run this
  |   +- /swarm   fresh-context agents |
  +------------------+-----------------+
                     |
                     v
  +------------------------------------+
  | /vibe         validation gate      |  <-- you run this
  +------------------+-----------------+
                     |
                     v
  +------------------------------------+
  | /post-mortem  extract learnings    |  <-- you run this
  +------------------+-----------------+
                     |
                     v
  SESSION END
  +------------------------------------+
  | auto-extract new learnings         |  <-- hook (automatic)
  +------------------+-----------------+
                     |
                     v
               .agents/  (git-tracked, compounds across sessions)
               |-- learnings/
               |-- patterns/
               |-- plans/
               +-- council/
                     |
                     +--------> next session starts here --------+
                                                                 |
                     +-------------------------------------------+
                     |
                     v
               (back to top)
```

</details>

**Bring your own spec tool.** Use [superpowers](https://github.com/anthropics/superpowers), SDD, or write plans by hand. AgentOps executes them reliably.

---

## Quick Start

**Catch risks before you code:**
```
/pre-mortem "add OAuth integration"
```

**Validate before you commit:**
```
/vibe
```

**Parallelize independent tasks (fresh context per agent):**
```
/swarm
```

**Execute an entire epic autonomously:**
```
/crank epic-123
```

Want a hands-on walkthrough? Run `/quickstart` for a guided tour on your actual codebase.

**Ready for Tier 1?** Install the `ao` CLI to enable the knowledge flywheel — see [Installation Options](#installation-options).

---

## Skills (What They Are)

AgentOps is delivered as **skills**: Markdown playbooks your agent runs via slash commands.

- Skills live in `skills/<name>/SKILL.md` — install with `npx skills` or `claude plugin add`
- Some skills are **orchestrators** that compose other skills:
  - `/crank` — runs the epic loop, dispatches `/swarm` for each wave
  - `/swarm` — spawns fresh-context agents to execute tasks in parallel

---

## Why Agents Need This

| Problem | Without AgentOps | With AgentOps |
|---------|------------------|---------------|
| Same mistakes repeated | Agent rediscovers bugs every session | Learnings persist and compound |
| Context bloat | Performance degrades over long tasks | Fresh context per spawned agent |
| Progress regresses | "Fixed" things break again | Validation gates lock progress |
| Parallel chaos | Agents step on each other | Orchestrated execution, atomic work |
| Slow feedback | Find problems after shipping | Shift-left: `/pre-mortem` before code |

---

## The Pipeline

| Stage | Skill | What It Does |
|-------|-------|--------------|
| **Shift-left** | `/pre-mortem` | Simulate failures BEFORE you write code |
| **Execute** | `/crank` | Orchestrate epic loop, dispatch `/swarm` for each wave |
| **Execute** | `/swarm` | Spawn fresh-context agents for parallel work |
| **Validate** | `/council` | Multi-model consensus (2-6 judges, cross-vendor, debate mode) |
| **Gate** | `/vibe` | Complexity analysis + council validation — must pass to merge |
| **Learn** | `/post-mortem` | Extract learnings to feed future sessions |

### What's New

**`/council` — Multi-model consensus validation.** Spawn parallel judges with different perspectives (pragmatist, skeptic, visionary) or custom presets (security-audit, architecture, ops). Supports cross-vendor review (`--mixed` adds Codex judges), adversarial debate (`--debate` for two-round review), and explorer sub-agents (`--explorers=N` for deep research). Council is the validation primitive — `/vibe`, `/pre-mortem`, and `/post-mortem` all use it.

**`/swarm` — Native team coordination.** Workers spawn as teammates on native teams (`TeamCreate` + `SendMessage`), enabling retry-via-message (no re-spawn), task self-service (workers claim via `TaskUpdate`), and fresh context per wave. Falls back to fire-and-forget if teams unavailable.

**`/ratchet` — Progress gates that lock.** Tracks Research → Plan → Implement → Validate stages. Once a gate passes, it's locked — prevents regression. `/crank` records gates automatically as it completes waves.

## Execution Modes

`/swarm`, `/crank`, and `/implement` support two execution modes:

| | Local (default) | Distributed (`--mode=distributed`) |
|---|---|---|
| **How** | Task tool background agents | tmux sessions + Agent Mail |
| **Dependencies** | None (Claude-native) | `tmux`, `claude` CLI, Agent Mail MCP |
| **Context** | Fresh per agent | Fresh per agent |
| **Persistence** | Dies if mayor disconnects | Survives disconnection |
| **Debugging** | Read output file | Attach to tmux session |
| **Coordination** | TaskList only | Agent Mail + file reservations |

**When to use which:**

| Scenario | Mode |
|----------|------|
| Quick parallel tasks (<5 min each) | Local |
| Long-running work (>10 min each) | Distributed |
| Need to debug stuck workers | Distributed |
| Multi-file changes across workers | Distributed |
| Mayor might disconnect | Distributed |
| No extra tooling installed | Local |

**Distributed mode dependencies:**
```bash
brew install tmux                    # Session management
# claude CLI - already installed if you're using Claude Code
```

Agent Mail is an MCP server for inter-agent messaging. Distributed mode requires it for coordination. See `docs/agent-mail.md` for setup options.

> **Note:** Local mode works out of the box with zero extra dependencies. Only set up distributed mode if you need persistence or complex coordination.

## Which Skill Should I Use?

| You Want | Use | Why |
|----------|-----|-----|
| Parallel tasks (fresh context each) | `/swarm` | Spawns agents, mayor owns the loop |
| Execute an entire epic | `/crank` | Orchestrates waves via `/swarm` until done |
| Single issue, full lifecycle | `/implement` | Claim → execute → validate → close |
| Gate progress without executing | `/ratchet` | Records/checks gates only |

---

## How It Works

Four patterns that make the pipeline reliable:

| Pattern | Problem | Solution |
|---------|---------|----------|
| **Fresh context per agent** | Context bloat degrades performance | Each spawned agent gets clean context |
| **Validation gates** | Work regresses or breaks | Must pass `/vibe` to commit |
| **Orchestrated execution** | Chaos with multiple agents | Mayor owns the loop, agents work atomically |
| **Compounding memory** | Same bugs rediscovered | `/post-mortem` → `.agents/` → `/inject` |

<details>
<summary>View architecture diagram</summary>

```
MAYOR (orchestrator)                AGENTS (executors)
--------------------                ------------------

/crank epic-123
  |
  +-> Get ready issues -----------> /swarm creates team per wave
  |                                   |
  +-> Create tasks ----------------> +-> Workers join as teammates
  |                                   |
  +-> Workers report completion <---- +-> Fresh context, execute atomically
  |     (via SendMessage)             |
  +-> /vibe (validation gate)         +-> Return result via SendMessage
  |     |
  |     +-> PASS = progress locked (/ratchet)
  |     +-> FAIL = fix first
  |
  +-> Loop until DONE
  |
  +-> /post-mortem ----------------> .agents/learnings/
                                       |
NEXT SESSION                           |
------------                           |
auto-inject (hook) <-------------------+
  |
  +-> Starts with prior knowledge
```

</details>

<details>
<summary>View full workflow stages</summary>

```
INPUT: SPEC (from superpowers, SDD, or your workflow)
  └── Plan, issues, acceptance criteria

STAGE 1: PRE-MORTEM [validation gate]
  /pre-mortem → Simulate failures BEFORE implementing

STAGE 2: EXECUTE [orchestrated + fresh context]
  /crank → Autonomous loop
    └── /swarm → Parallel agents (fresh context each)

STAGE 3: VALIDATE [validation gate]
  /vibe → 8-aspect check, must pass to commit

STAGE 4: LEARN [compounding memory]
  /post-mortem → Extract learnings for next session

OUTPUT: LEARNINGS (feed your next spec)
  └── .agents/learnings/, .agents/patterns/
```

</details>

---

## Installation Options

<details>
<summary>Install for specific agents</summary>

```bash
# Codex
npx skills@latest add boshu2/agentops -g -a codex -s '*' -y

# OpenCode
npx skills@latest add boshu2/agentops -g -a opencode -s '*' -y

# Cursor
npx skills@latest add boshu2/agentops -g -a cursor -s '*' -y

# Update all
npx skills@latest update
```

</details>

<details>
<summary>Install CLI (optional, enables hooks)</summary>

```bash
# macOS
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops

# Any OS with Go
go install github.com/boshu2/agentops/cli/cmd/ao@latest

# Enable auto-hooks (Claude Code)
ao hooks install && ao hooks test
```

</details>

> **Note:** There's a [known bug](https://github.com/anthropics/claude-code/issues/15178) where plugin skills don't appear when pressing `/`. Skills still work — just type them directly.

---

## Tool Dependencies

The `/vibe` and `/post-mortem` skills run `scripts/toolchain-validate.sh`, which uses available linters and scanners. **All tools are optional** — missing ones are skipped gracefully.

| Tool | Purpose | Install |
|------|---------|---------|
| **gitleaks** | Secret scanning | `brew install gitleaks` |
| **semgrep** | SAST security patterns | `brew install semgrep` |
| **trivy** | Dependency vulnerabilities | `brew install trivy` |
| **gosec** | Go security | `go install github.com/securego/gosec/v2/cmd/gosec@latest` |
| **hadolint** | Dockerfile linting | `brew install hadolint` |
| **ruff** | Python linting | `pip install ruff` |
| **radon** | Python complexity | `pip install radon` |
| **golangci-lint** | Go linting | `brew install golangci-lint` |
| **shellcheck** | Shell linting | `brew install shellcheck` |

**Quick install (recommended):**
```bash
brew install gitleaks semgrep trivy hadolint shellcheck golangci-lint
pip install ruff radon
```

More tools = more coverage. But even with zero tools installed, the workflow still runs.

---

## Troubleshooting

- Plugin skills don’t show up when you press `/` in Claude Code: type the skill directly (e.g. `/pre-mortem`). (See the Claude Code issue linked above.)
- `ao` not found: ensure it’s on your `PATH` (`which ao`). For hook setup help, see `cli/docs/HOOKS.md`.

---

## The `/vibe` Validator

Not just "does it compile?" — **does it match the spec?**

| Aspect | What It Checks |
|--------|----------------|
| Semantic | Does code do what spec says? |
| Security | SQL injection, auth bypass, hardcoded secrets |
| Quality | Dead code, copy-paste, magic numbers |
| Architecture | Layer violations, circular deps, god classes |
| Complexity | Cyclomatic > 10, deep nesting |
| Performance | N+1 queries, unbounded loops, resource leaks |
| Slop | AI hallucinations, cargo cult, over-engineering |
| Accessibility | Missing ARIA, broken keyboard nav, contrast |

**Gate rule:** 0 critical = pass. 1+ critical = blocked until fixed.

---

## All Skills

| Skill | Purpose |
|-------|---------|
| `/pre-mortem` | Simulate failures before coding |
| `/crank` | Autonomous epic execution (orchestrator; runs waves via `/swarm`) |
| `/swarm` | Parallel agents with fresh context (native teams) |
| `/council` | Multi-model consensus (validate, brainstorm, critique, research, analyze) |
| `/vibe` | Complexity + council validation gate |
| `/implement` | Single issue execution |
| `/post-mortem` | Extract learnings |
| `/research` | Deep codebase exploration |
| `/plan` | Break goal into tracked issues |
| `/ratchet` | Progress gates that lock (Research → Plan → Implement → Validate) |
| `/beads` | Git-native issue tracking |

<details>
<summary>More skills</summary>

| Skill | Purpose |
|-------|---------|
| `/retro` | Quick retrospective |
| `/inject` | Manually load prior knowledge |
| `/knowledge` | Query knowledge base |
| `/bug-hunt` | Root cause analysis |
| `/complexity` | Code complexity metrics |
| `/doc` | Documentation generation |
| `/standards` | Language-specific rules |
</details>

---

## CLI Reference

The `ao` CLI is what makes the flywheel turn. It handles knowledge persistence with MemRL two-phase retrieval, confidence decay (stale knowledge ages out), and citation-tracked provenance so you can trace learnings back to the session that produced them.

```bash
ao quick-start --minimal  # Create .agents/ structure
ao hooks install          # Enable auto-hooks (Claude Code)
ao hooks test             # Verify hooks are working
ao inject [topic]         # Load prior knowledge (auto at session start)
ao search "query"         # Semantic search across learnings
ao flywheel status        # Knowledge growth rate, escape velocity
ao metrics report         # Flywheel health dashboard
ao forge transcript       # Extract learnings from session transcripts
ao ratchet status         # RPI progress gates (Research → Plan → Implement → Validate)
ao pool list              # Show knowledge by quality tier
```

---

## Built On

| Project | What We Use |
|---------|-------------|
| [Ralph Wiggum pattern](https://ghuntley.com/ralph/) | Fresh context per agent spawn |
| [Brownian Ratchet](https://github.com/dlorenc/multiclaude) | Validation gates that lock progress |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for memory |

---

## License

Apache-2.0 · [Documentation](docs/) · [Changelog](CHANGELOG.md)
