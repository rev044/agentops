<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**Multi-agent orchestration with validation gates and memory that compounds.**

[Install](#install) · [Quick Start](#quick-start) · [How It Works](#how-it-works) · [Documentation](docs/)

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

## Quick Start

**Validate before you code:**
```
/pre-mortem "add OAuth integration"
```

**Execute with parallel agents:**
```
/crank epic-123
```

**Check before you commit:**
```
/vibe
```

That's it. AgentOps catches problems before they ship and remembers solutions for next time.

---

## What It Does

| Problem | AgentOps Solution |
|---------|-------------------|
| Agents make the same mistakes repeatedly | **Memory that compounds** — learnings persist across sessions |
| Context degrades over long tasks | **Fresh context per agent** — each spawned agent starts clean |
| Progress slips backward | **Validation gates** — can't commit until checks pass |
| Parallel agents step on each other | **Orchestrated execution** — mayor coordinates, agents work atomically |

---

## Core Skills

| Skill | What It Does |
|-------|--------------|
| `/pre-mortem` | Simulate failures BEFORE you write code |
| `/crank` | Execute an epic autonomously with parallel agents |
| `/swarm` | Spawn fresh-context agents for parallel work |
| `/vibe` | 8-aspect validation gate before commit |
| `/post-mortem` | Extract learnings to feed future sessions |

**The workflow:**
```
/pre-mortem → /crank → /vibe → /post-mortem
     ↓           ↓        ↓          ↓
  Prevent    Execute   Validate   Remember
```

---

## How It Works

AgentOps combines four patterns that solve autonomous agent failures:

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
STAGE 1: UNDERSTAND
  /research → Deep-dive codebase
  /plan     → Break into tracked issues

STAGE 2: PRE-MORTEM [validation gate]
  /pre-mortem → Simulate failures BEFORE implementing

STAGE 3: EXECUTE [orchestrated + fresh context]
  /crank → Autonomous loop
    └── /swarm → Parallel agents (fresh context each)

STAGE 4: VALIDATE [validation gate]
  /vibe → 8-aspect check, must pass to commit

STAGE 5: LEARN [compounding memory]
  /post-mortem → Extract learnings for next session
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
| `/crank` | Autonomous epic execution (uses swarm) |
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
