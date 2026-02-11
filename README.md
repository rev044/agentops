<div align="center">

# AgentOps

### Goal in, production code out.

[![Version](https://img.shields.io/badge/version-2.3.0-brightgreen)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-32-7c3aed)](skills/)
[![Hooks](https://img.shields.io/badge/hooks-11-orange)](hooks/)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

**Turn your coding agent into an autonomous software engineering system.**<br/>
One command runs research, planning, multi-model validation, parallel implementation,<br/>
and knowledge extraction — with self-correcting retry loops and zero human prompts.

[What Is AgentOps](#what-is-agentops) · [Install](#install) · [See It Work](#see-it-work) · [Skills](#skills-catalog) · [FAQ](#faq) · [Docs](docs/)

</div>

---

## What Is AgentOps

AgentOps is a skills plugin that turns your coding agent into an autonomous software engineering system. It works with **Claude Code**, **Codex CLI**, **Cursor**, **Open Code** — any agent that supports the [Skills protocol](https://skills.sh). One command runs research, planning, multi-model validation, parallel implementation, and knowledge extraction. All state is local — stored in a `.agents/` directory and tracked in git.

## Prerequisites

| Dependency | Required? | What it's for |
|-----------|-----------|---------------|
| **Node.js 18+ + npm** | Yes | Installs the skills plugin (`npx skills`) |
| **A coding agent** | Yes | Claude Code, Cursor, Codex CLI, or any Skills-compatible agent |
| **Git** | Recommended | Knowledge artifacts tracked in `.agents/`, hooks gate on git ops |
| **Homebrew** | Optional | Easiest way to install the `ao` CLI |
| **`ao` CLI** | Optional | Knowledge flywheel, ratchet gates, session hooks |
| **`tmux`** | Optional | Agent team pane mode (`brew install tmux`) |

> **What works without `ao`?** All 32 skills work without the CLI. The `ao` CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle hooks.

## Install

**Option 1: npx (recommended)**

```bash
npx skills@latest add boshu2/agentops --all -g
```

One line. Done. Run `/quickstart` to get started.

**Option 2: Install script** (installs plugin + optional CLI + hooks)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install.sh)
```

<details>
<summary><strong>Option 3: Manual</strong></summary>

```bash
# 1. Plugin
npx skills@latest add boshu2/agentops --all -g

# 2. CLI (optional — adds knowledge flywheel)
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops

# 3. Hooks (optional — auto-enforces workflow)
ao hooks install
```

Or: `claude plugin add boshu2/agentops`

</details>

**Try it in 60 seconds** (run these inside your coding agent):

```bash
/quickstart                        # Guided tour on your actual codebase
/council validate this PR           # Standalone — no setup needed
```

## See It Work

```
> /rpi "add rate limiting to the API"

[research]    Exploring codebase... → .agents/research/rate-limiting.md
[plan]        3 issues, 2 waves → epic ag-0057
[pre-mortem]  3 judges → Verdict: PASS
[crank]       Wave 1: ███ 2/2 · Wave 2: █ 1/1
[vibe]        3 judges → Verdict: PASS
[post-mortem] 3 learnings extracted → .agents/
```

That's six phases — research, planning, failure simulation, parallel implementation, code validation, and learning extraction — running autonomously. Failed validation triggers a retry with the failure context. Ratchet gates lock each phase so progress can't go backward. You type one command and walk away.

<details>
<summary><strong>More examples</strong></summary>

**Autonomous epic execution:**
```
> /crank ag-0042

[crank] Epic: ag-0042 — 6 issues, 3 waves
[wave-1] ██████ 3/3 complete
[wave-2] ████── 2/2 complete
[wave-3] ██──── 1/1 complete
[vibe] PASS — all gates locked
[post-mortem] 4 learnings extracted → .agents/
```

**Multi-model validation council:**
```
> /council --deep validate the auth system

[council] 3 judges spawned
[judge-1] PASS — JWT implementation correct
[judge-2] WARN — rate limiting missing on /login
[judge-3] PASS — refresh rotation implemented
Consensus: WARN — add rate limiting before shipping
```

**Parallel agents with fresh context:**
```
> /swarm ag-0042 --wave=1

[swarm] Spawning 3 workers (clean context each)
[worker-1] ✓ ag-001 — add middleware
[worker-2] ✓ ag-002 — add rate limiter
[worker-3] ✓ ag-003 — add config schema
[lead] Validated → committed 3 files
```

</details>

## How It Works

AgentOps is built on four ideas: a physics metaphor for reliable agent output, a memory pattern for compounding intelligence, an execution model for safe parallelism, and an orchestration loop that wires them together.

### The Brownian Ratchet

```
  Chaos × Filter → Locked Progress

  ╭─ agent-1 ─→ ✓ ─╮
  ├─ agent-2 ─→ ✗ ─┤   3 attempts, 1 fails
  ├─ agent-3 ─→ ✓ ─┤   filter catches it
  ╰─ council ──→ PASS   ratchet locks the progress
                  ↓
          can't go backward
```

Spawn parallel agents (chaos), validate with multi-model council (filter), merge to main (ratchet). Failed agents are cheap — fresh context means no contamination. Progress only moves forward.

### Ralph Loops

```
  Wave 1:  Select backend (spawn_agent or TeamCreate) → spawn 3 workers
           workers write files → lead validates → lead commits
           cleanup backend resources

  Wave 2:  Select backend (spawn_agent or TeamCreate) → spawn 2 workers
           ...same pattern, zero accumulated context
```

Every wave gets a fresh worker set (new sub-agents or teammates). Every worker gets clean context. No bleed-through between waves. The lead is the only one who commits — eliminates merge conflicts across parallel workers.

### Knowledge Flywheel

```
  Session 1:  work → ao forge → extract learnings → .agents/
  Session 2:  ao inject → .agents/ loaded → work → extract → .agents/
  Session 3:  ao inject → richer context → work → extract → .agents/
                          compounds over time ↑
```

The `ao` CLI auto-injects relevant knowledge at session start and auto-extracts learnings at session end. Decisions, patterns, failures, and fixes accumulate in `.agents/` — your agents get smarter with every cycle.

### Context Orchestration

```
  /rpi "goal"
    │
    ├─ Phase 1: /research ─── explore codebase
    ├─ Phase 2: /plan ─────── decompose into issues + waves
    ├─ Phase 3: /pre-mortem ── 3-judge council validates plan
    │                          FAIL? → re-plan → re-validate (max 3)
    ├─ Phase 4: /crank ─────── autonomous wave execution
    │                          BLOCKED? → retry with context (max 3)
    ├─ Phase 5: /vibe ──────── 3-judge council validates code
    │                          FAIL? → re-crank → re-vibe (max 3)
    └─ Phase 6: /post-mortem ─ extract learnings → .agents/
```

Six phases, zero human gates. Council FAIL triggers retry loops — re-plan or re-crank with the failure context, then re-validate. Ratchet checkpoints after each phase lock progress. Use `--interactive` if you want human gates at research and plan.

## Skills Catalog

32 skills across three tiers — orchestration, workflow, and utilities — plus 10 internal skills that fire automatically.

### Orchestration

| Skill | What it does |
|-------|-------------|
| `/rpi` | Goal to production — autonomous 6-phase lifecycle with self-correcting retry loops |
| `/council` | Multi-model consensus — spawns parallel judges, consolidates verdict (default, `--deep`, `--mixed`) |
| `/crank` | Autonomous epic execution — runs `/swarm` waves until all issues closed |
| `/swarm` | Parallel agents with fresh context — runtime-native backend (Codex sub-agents or Claude teams), lead commits |
| `/codex-team` | Parallel Codex execution — prefers Codex sub-agents, falls back to Codex CLI |

### Workflow

| Skill | What it does |
|-------|-------------|
| `/research` | Deep codebase exploration → `.agents/research/` |
| `/plan` | Decompose goal into issues with dependency waves |
| `/implement` | Single issue, full lifecycle — explore, code, test, commit |
| `/vibe` | Complexity analysis + multi-model validation gate |
| `/pre-mortem` | Simulate failures before coding (3 judges: requirements, feasibility, scope) |
| `/post-mortem` | Validate implementation + extract learnings |
| `/release` | Pre-flight checks, changelog, version bumps, tag, GitHub Release |

### Utilities

| Skill | What it does |
|-------|-------------|
| `/status` | Dashboard — current work, recent validations, next action |
| `/quickstart` | Interactive onboarding — guided RPI cycle on your codebase |
| `/handoff` | Structured session handoff for continuation |
| `/retro` | Extract learnings from completed work |
| `/knowledge` | Query knowledge base across `.agents/` |
| `/bug-hunt` | Root cause analysis with git archaeology |
| `/complexity` | Code complexity metrics (radon, gocyclo) |
| `/doc` | Documentation generation and validation |
| `/trace` | Trace design decisions through history and git |
| `/inbox` | Monitor Agent Mail messages |

<details>
<summary><strong>Internal skills (auto-loaded, 10 total)</strong></summary>

| Skill | Trigger | What it does |
|-------|---------|-------------|
| `inject` | Session start | Load relevant prior knowledge into session context |
| `extract` | On demand | Pull learnings from artifacts |
| `forge` | Session end | Mine transcript for decisions, learnings, failures, patterns |
| `flywheel` | On demand | Knowledge health — velocity, pool depths, staleness |
| `ratchet` | On demand | RPI progress gates — once locked, stays locked |
| `standards` | By `/vibe`, `/implement` | Language-specific coding rules (Python, Go, TS, Shell) |
| `beads` | By `/plan`, `/implement` | Git-native issue tracking reference |
| `provenance` | On demand | Trace knowledge artifact lineage and sources |
| `shared` | By distributed skills | Shared reference documents for distributed mode |
| `using-agentops` | Auto-injected | AgentOps workflow guide |

</details>

## The `ao` CLI

The engine behind the skills. Written in Go. Manages knowledge injection, extraction, ratchet gates, and session lifecycle.

```bash
ao inject              # Load prior knowledge into session
ao forge transcript <path>  # Extract learnings from session transcript
ao ratchet status      # Check RPI progress gates
ao ratchet next        # What's the next phase?
ao search "topic"      # Semantic search across knowledge artifacts
ao session close       # Full lifecycle close: forge → extract → promote
ao flywheel status     # Knowledge health — velocity, pool depths
ao hooks install       # Install session hooks for auto-inject/extract
```

## Agent Backends

AgentOps orchestration skills are runtime-native. In local mode they select backend in this order:

1. Codex experimental sub-agents (`spawn_agent`)
2. Claude native teams (`TeamCreate` + `SendMessage`)
3. Background task fallback (`Task(run_in_background=true)`)

`/council`, `/swarm`, and `/crank` all follow this backend contract.

```
  Council:                               Swarm:
  ╭─ judge-1 ──╮                  ╭─ worker-1 ──╮
  ├─ judge-2 ──┼→ lead            ├─ worker-2 ──┼→ lead
  ╰─ judge-3 ──╯   consolidates   ╰─ worker-3 ──╯   validates + commits
```

**Claude teams setup** (optional, only if using Claude team backend):
```json
// ~/.claude/settings.json
{
  "teammateMode": "tmux",
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

| Mode | What you see |
|------|-------------|
| `"tmux"` | Each teammate gets own pane — click to interact |
| `"in-process"` | Single terminal — Shift+Up/Down cycles teammates |
| `"auto"` | Best available (default) |

Requires `tmux` for pane mode (`brew install tmux`).

Codex sub-agent backend requires no Claude team configuration.

## Hooks

11 hooks auto-enforce the workflow — no discipline required.

| Hook | Trigger | What it does |
|------|---------|-------------|
| Push gate | `git push` | Blocks push if `/vibe` hasn't passed |
| Worker guard | `git commit` | Blocks workers from committing (lead-only) |
| Dangerous git guard | `force-push`, `reset --hard` | Blocks destructive git commands |
| Ratchet nudge | Any prompt | "Run /vibe before pushing", "Run /pre-mortem first" |
| Task validation | Task completed | Validates metadata before accepting completion |

All hooks have a kill switch: `AGENTOPS_HOOKS_DISABLED=1`.

## Use Council Standalone

Use `/council` independently — no RPI workflow required, no `ao` CLI, no setup beyond plugin install. Validate any PR, brainstorm approaches, or research decisions with multi-model consensus.

```bash
/council validate this PR                                    # review current changes
/council --deep --preset=security-audit validate the auth system  # thorough security audit
/council brainstorm caching strategies for the API           # explore options with parallel judges
/council research Redis vs Memcached for our use case        # deep-dive technology comparison
```

Works on any codebase, any file, any question. See [skills/council/SKILL.md](skills/council/SKILL.md) for all modes and presets.

## FAQ

<details>
<summary><strong>Why not just use your coding agent directly?</strong></summary>

Your coding agent can spawn agents and write code. AgentOps turns it into an autonomous software engineering system:

- **Goal in, production code out** — `/rpi "goal"` runs 6 phases hands-free. You don't prompt, approve, or babysit.
- **Self-correcting** — Failed validation triggers retry with failure context, not human escalation. The system fixes its own mistakes.
- **Can't regress** — Brownian ratchet locks progress after each phase. Parallel agents generate chaos; multi-model council filters it; ratchet keeps only what passes.
- **Runtime-aware orchestration** — Same skills choose Codex sub-agents or Claude teams automatically based on the active runtime.
- **Fresh context every time** — Ralph loops: each wave gets fresh workers, each worker gets clean context. No accumulated hallucinations, no bleed-through between tasks.
- **Gets smarter** — Knowledge flywheel: every session forges learnings into `.agents/`. Next session injects them. Your agents compound intelligence across sessions.
- **Cross-vendor** — `--mixed` mode adds Codex judges alongside Claude. Different models catch different bugs.
- **Self-enforcing** — Hooks block pushes without validation, prevent workers from committing, nudge agents through the lifecycle. No discipline required.

</details>

<details>
<summary><strong>What data leaves my machine?</strong></summary>

Nothing. All state lives in `.agents/` (git-tracked, local). No telemetry, no cloud sync, no external services. Your knowledge stays yours.

</details>

<details>
<summary><strong>Do I need the ao CLI?</strong></summary>

Skills work without it. The CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle management. Recommended for full-lifecycle workflows. Install it with `brew install agentops`.

</details>

<details>
<summary><strong>How does the council actually work?</strong></summary>

`/council` spawns 2–6 parallel judge agents (Claude and/or Codex), each with a different perspective (security, performance, correctness, etc.). Judges deliberate asynchronously, then the lead consolidates into a single verdict: PASS, WARN, or FAIL with specific findings. Built-in presets: `default`, `security-audit`, `architecture`, `research`, `ops`.

</details>

<details>
<summary><strong>Can I use this with other AI coding tools?</strong></summary>

AgentOps works with any agent that supports the Skills protocol — Claude Code, Codex CLI, Cursor, Open Code, and others. The `--mixed` council mode adds Codex (OpenAI) judges alongside Claude for cross-vendor validation. The knowledge flywheel (`.agents/` directory) is plain markdown files that any tool can read.

</details>

<details>
<summary><strong>How do I uninstall?</strong></summary>

```bash
npx skills@latest remove boshu2/agentops -g
brew uninstall agentops  # if installed
```

</details>

## Built On

| Project | What we use it for |
|---------|-------------------|
| [Fresh context per agent](https://ghuntley.com/ralph/) | Each spawned agent gets clean context — no bleed-through |
| [Validation gates that lock](https://github.com/dlorenc/multiclaude) | Once a stage passes, it stays passed — no regression |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guide.

```bash
# Run tests
./tests/run-all.sh

# Test the plugin locally
claude --plugin ./
```

## License

Apache-2.0 · [Documentation](docs/INDEX.md) · [Glossary](docs/GLOSSARY.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
