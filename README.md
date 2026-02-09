<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

**Autonomous coding agent orchestration. One command, full lifecycle.**

</div>

## Install

```bash
# Skills (Claude Code plugin — skills, hooks, agent teams)
npx skills@latest add boshu2/agentops --all -g

# CLI (knowledge engine — inject, extract, ratchet gates)
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops
ao hooks install
```

Or: `claude plugin add boshu2/agentops`

## See It Work

**One command, goal to production:**
```
> /rpi "add rate limiting to the API"

[research]    Exploring codebase... → .agents/research/rate-limiting.md
[plan]        3 issues, 2 waves → epic ag-0057
[pre-mortem]  3 judges → Verdict: PASS
[crank]       Wave 1: ███ 2/2 · Wave 2: █ 1/1
[vibe]        3 judges → Verdict: PASS
[post-mortem] 3 learnings extracted → .agents/
```
No prompts. No babysitting. Failed validation → automatic retry with failure context. Ratchet locks each phase — progress can't go backward.

**Run an entire epic hands-free:**
```
> /crank ag-0042

[crank] Epic: ag-0042 — 6 issues, 3 waves
[wave-1] ██████ 3/3 complete
[wave-2] ████── 2/2 complete
[wave-3] ██──── 1/1 complete
[vibe] PASS — all gates locked
[post-mortem] 4 learnings extracted → .agents/
```

**Multi-model validation before you ship:**
```
> /council --deep validate the auth system

[council] 3 judges spawned
[judge-1] PASS — JWT implementation correct
[judge-2] WARN — rate limiting missing on /login
[judge-3] PASS — refresh rotation implemented
Consensus: WARN — add rate limiting before shipping
```

## How It Works

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
  Wave 1:  TeamCreate → spawn 3 workers (fresh context each)
           workers write files → lead validates → lead commits
           TeamDelete

  Wave 2:  TeamCreate → spawn 2 workers (fresh context each)
           ...same pattern, zero accumulated context
```

Every wave gets a new team. Every worker gets clean context. No bleed-through between waves. The lead is the only one who commits — eliminates merge conflicts across parallel workers.

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

### Knowledge Flywheel

```
  Session 1:  work → ao forge → extract learnings → .agents/
  Session 2:  ao inject → .agents/ loaded → work → extract → .agents/
  Session 3:  ao inject → richer context → work → extract → .agents/
                          compounds over time ↑
```

The `ao` CLI auto-injects relevant knowledge at session start and auto-extracts learnings at session end. Knowledge compounds across sessions — the agent gets smarter with every cycle.

## Skills

### Orchestration

| Skill | What it does |
|-------|-------------|
| `/rpi` | Goal to production — autonomous 6-phase lifecycle with self-correcting retry loops |
| `/council` | Multi-model consensus — spawns parallel judges, consolidates verdict |
| `/crank` | Autonomous epic execution — runs `/swarm` waves until all issues closed |
| `/swarm` | Parallel agents with fresh context — team per wave, lead commits |
| `/codex-team` | Parallel Codex agents orchestrated by Claude — cross-vendor execution |

### Workflow

| Skill | What it does |
|-------|-------------|
| `/research` | Deep codebase exploration → `.agents/research/` |
| `/plan` | Decompose goal into issues with dependency waves |
| `/implement` | Single issue, full lifecycle — explore, code, test, commit |
| `/vibe` | Complexity analysis + multi-model validation gate |
| `/pre-mortem` | Simulate failures before coding (3 judges: requirements, feasibility, scope) |
| `/post-mortem` | Validate implementation + extract learnings |
| `/release` | Pre-flight, changelog, version bumps, tag, GitHub Release |

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
<summary>Internal skills (auto-loaded)</summary>

| Skill | Trigger | What it does |
|-------|---------|-------------|
| `inject` | Session start | Load relevant prior knowledge with decay weighting |
| `extract` | On demand | Pull learnings from artifacts |
| `forge` | Session end | Mine transcript for decisions, learnings, failures, patterns |
| `flywheel` | On demand | Knowledge health — velocity, pool depths, staleness |
| `ratchet` | On demand | RPI progress gates — once locked, stays locked |
| `standards` | By `/vibe`, `/implement` | Language-specific coding rules (Python, Go, TS, Shell) |
| `beads` | By `/plan`, `/implement` | Git-native issue tracking reference |
| `provenance` | On demand | Trace knowledge artifact lineage and sources |

</details>

## The `ao` CLI

The engine behind the skills. Manages knowledge injection, extraction, ratchet gates, and session lifecycle.

```bash
ao inject              # Load prior knowledge into session
ao forge transcript <path>  # Extract learnings from session transcripts
ao ratchet status      # Check RPI progress gates
ao ratchet next        # What's the next phase?
ao search "topic"      # Semantic search across knowledge artifacts
ao session close       # Full lifecycle close: forge → extract → promote
ao flywheel status     # Knowledge health — velocity, pool depths
ao hooks install       # Install session hooks for auto-inject/extract
```

## Agent Teams

AgentOps uses Claude Code's native [agent teams](https://code.claude.com/docs/en/agent-teams) — `/council`, `/swarm`, and `/crank` create teams automatically.

```
  Council:                         Swarm:
  ╭─ judge-1 ──╮                  ╭─ worker-1 ──╮
  ├─ judge-2 ──┼→ team lead       ├─ worker-2 ──┼→ team lead
  ╰─ judge-3 ──╯   consolidates   ╰─ worker-3 ──╯   validates + commits
```

**Setup** (one-time):
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

## Hooks

Hooks auto-enforce the workflow — no discipline required.

| Hook | Trigger | What it does |
|------|---------|-------------|
| Push gate | `git push` | Blocks push if `/vibe` hasn't passed |
| Worker guard | `git commit` | Blocks workers from committing (lead-only) |
| Dangerous git guard | `force-push`, `reset --hard` | Blocks destructive git commands |
| Ratchet nudge | Any prompt | "Run /vibe before pushing", "Run /pre-mortem first" |
| Task validation | Task completed | Validates metadata before accepting completion |

All hooks have a kill switch: `AGENTOPS_HOOKS_DISABLED=1`.

## FAQ

**Why not just use Claude Code directly?**
Claude Code can spawn agents and write code. AgentOps turns it into an autonomous software engineering system:
- **Goal in, production code out** — `/rpi "goal"` runs 6 phases hands-free. You don't prompt, approve, or babysit.
- **Self-correcting** — Failed validation triggers retry with failure context, not human escalation. The system fixes its own mistakes.
- **Can't regress** — Brownian ratchet locks progress after each phase. Parallel agents generate chaos; multi-model council filters it; ratchet keeps only what passes.
- **Fresh context every time** — Ralph loops: each wave gets a new team, each worker gets clean context. No accumulated hallucinations, no bleed-through between tasks.
- **Gets smarter** — Knowledge flywheel: every session forges learnings into `.agents/`. Next session injects them. Your agents compound intelligence across sessions.
- **Cross-vendor** — `--mixed` mode adds Codex judges alongside Claude. Different models catch different bugs.
- **Self-enforcing** — Hooks block pushes without validation, prevent workers from committing, nudge agents through the lifecycle. No discipline required.

**What data leaves my machine?**
Nothing. All state lives in `.agents/` (git-tracked, local). No telemetry, no cloud sync.

**Do I need the `ao` CLI?**
Skills work without it. The CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle management. Recommended for full-lifecycle workflows.

**How do I uninstall?**
```bash
npx skills@latest remove boshu2/agentops -g
brew uninstall agentops  # if installed
```

## Built On

| Project | What we use it for |
|---------|-------------------|
| [Fresh context per agent](https://ghuntley.com/ralph/) | Each spawned agent gets clean context — no bleed-through |
| [Validation gates that lock](https://github.com/dlorenc/multiclaude) | Once a stage passes, it stays passed — no regression |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

## License

Apache-2.0 · [Documentation](docs/) · [Reference](docs/reference.md) · [Changelog](CHANGELOG.md)
