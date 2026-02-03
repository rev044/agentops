<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**DevOps for AI agents.**

Maximize flow. Shorten feedback loops. Compound what you learn.

[Install](#install) · [What This Is](#what-this-is) · [Quick Start](#quick-start) · [How It Works](#how-it-works) · [Docs](docs/)

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

---

## What This Is

DevOps doesn't write your code. It makes sure code flows reliably from commit to production.

**AgentOps doesn't write your spec.** It makes sure specs flow reliably from plan to working code.

```
Traditional DevOps:
  code → build → test → deploy → monitor → feedback
                                              ↓
                                        next commit

AgentOps:
  spec → /pre-mortem → /crank → /vibe → /post-mortem
                                              ↓
                                     .agents/learnings/
                                              ↓
                                         next spec
```

| DevOps Principle | AgentOps |
|------------------|----------|
| CI/CD pipelines | `/crank` + `/swarm` — automated, parallel execution |
| Shift-left testing | `/pre-mortem` — catch failures before you code |
| Quality gates | `/vibe` — must pass to merge |
| Continuous improvement | Flywheel — learnings feed future sessions |

**Bring your own spec tool.** Use [superpowers](https://github.com/anthropics/superpowers), SDD, or write plans by hand. AgentOps is the pipeline that executes them reliably

---

## Quick Start

**Validate before you code:**
```
/pre-mortem "add OAuth integration"
```

**Execute an epic (issue loop):**
```
/crank epic-123
```

**Parallelize independent tasks (Ralph loop / fresh context):**
```
/swarm
```

**Check before you commit:**
```
/vibe
```

That's it. AgentOps catches problems before they ship and remembers solutions for next time.

---

## Skills (What They Are)

AgentOps is delivered as **skills**: Markdown playbooks your agent runs via slash commands.

- Skills live in `skills/<name>/SKILL.md` — install with `npx skills` or `claude plugin add`
- Some skills are **orchestrators** that compose other skills:
  - `/crank` — runs the epic loop, dispatches `/swarm` for each wave
  - `/swarm` — spawns fresh-context agents to execute tasks in parallel

---

## Why Agents Need DevOps

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
| **Gate** | `/vibe` | 8-aspect validation — must pass to merge |
| **Learn** | `/post-mortem` | Extract learnings to feed future sessions |

## Execution Modes

`/swarm`, `/crank`, and `/implement` support two execution modes:

| | Local (default) | Distributed (`--distributed`) |
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

Agent Mail is an MCP server for inter-agent messaging. Distributed mode requires it for coordination. See [mcp_agent_mail](https://github.com/boshu2/acfs-research) for setup.

> **Note:** Local mode works out of the box with zero extra dependencies. Only set up distributed mode if you need persistence or complex coordination.

## Which Skill Should I Use?

| You Want | Use | Why |
|----------|-----|-----|
| Parallel tasks (fresh context each) | `/swarm` | Spawns agents, mayor owns the loop |
| Execute an entire epic | `/crank` | Orchestrates waves via `/swarm` until done |
| Single issue, full lifecycle | `/implement` | Claim → execute → validate → close |
| Gate progress without executing | `/ratchet` | Records/checks gates only |

**The pipeline:**
```
[spec] → /pre-mortem → /crank → /vibe → /post-mortem → [learnings]
              ↓           ↓        ↓          ↓             ↓
          Shift-left   Execute   Gate      Extract    Feed next run
```

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
MAYOR (orchestrator)              AGENTS (executors)
--------------------              ------------------

/crank epic-123
     |
     +-> Get ready issues ---------> /swarm spawns N agents
     |                                    |
     +-> Create tasks -------------->     +-> Fresh context each
     |                                    |
     +-> Wait for completion <------      +-> Execute atomically
     |                                    |
     +-> /vibe (validation gate)          +-> Return result
     |      |
     |      +-> PASS = progress locked
     |      +-> FAIL = fix first
     |
     +-> Loop until DONE
     |
     +-> /post-mortem -----------------> .agents/learnings/
                                              |
NEXT SESSION                                  |
------------                                  |
SessionStart <---- /inject <------------------+
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

## What Gets Captured

Everything lives in `.agents/` — git-tracked, portable, yours.

```
.agents/
├── learnings/     # "Auth bugs stem from token refresh"
├── patterns/      # "How we handle retries"
├── pre-mortems/   # Failure simulations
├── plans/         # Implementation plans
├── vibe/          # Validation reports
└── ...
```

With hooks enabled, the flywheel turns automatically:
- **SessionStart** → Injects relevant prior knowledge
- **SessionEnd** → Extracts learnings for next time

---

## All Skills

| Skill | Purpose |
|-------|---------|
| `/pre-mortem` | Simulate failures before coding |
| `/crank` | Autonomous epic execution (orchestrator; runs waves via `/swarm`) |
| `/swarm` | Parallel agents with fresh context |
| `/vibe` | 8-aspect validation gate |
| `/implement` | Single issue execution |
| `/post-mortem` | Extract learnings |
| `/research` | Deep codebase exploration |
| `/plan` | Break goal into tracked issues |
| `/beads` | Git-native issue tracking |

<details>
<summary>More skills</summary>

| Skill | Purpose |
|-------|---------|
| `/ratchet` | Track RPI progress gates |
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

```bash
ao quick-start --minimal  # Create .agents/ structure
ao hooks install          # Enable auto-hooks (Claude Code)
ao inject [topic]         # Load prior knowledge
ao search "query"         # Search knowledge base
ao flywheel status        # Check flywheel health
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
