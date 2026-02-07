<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

**Your AI agent forgets everything between sessions. AgentOps fixes that.**

</div>

## Install

```bash
npx skills@latest add boshu2/agentops --all -g
```

Or: `claude plugin add boshu2/agentops`

## See It Work

**Catch risks before you code:**
```
> /pre-mortem "add OAuth integration"

[council] Spawning 3 judges: missing-requirements, feasibility, scope
[missing-requirements] RISK: No token revocation strategy defined
[feasibility] RISK: Refresh token rotation — silent failure on expiry
[scope] WARN: OAuth adds 3 new dependencies, consider scope
Verdict: WARN — 2 significant risks, address before implementing
```

**Validate before you commit:**
```
> /vibe

[complexity] radon: all functions grade A-B ✓
[council] Spawning 3 judges: error-paths, api-surface, spec-compliance
[spec-compliance] All acceptance criteria met
Verdict: PASS — ready to ship
```

**Parallelize with fresh context per agent:**
```
> /swarm

[swarm] Creating team swarm-1738900000...
[worker-1] ✓ implement-auth
[worker-2] ✓ add-rate-limiter
[worker-3] ✓ update-tests
[lead] Validated + committed. Run /vibe to gate.
```

**Run an entire epic hands-free:**
```
> /crank

[crank] Epic: ag-0042 — 6 issues, 3 waves
[wave-1] ██████ 3/3 complete
[wave-2] ████── 2/2 complete
[wave-3] ██──── 1/1 complete
[vibe] PASS — all gates locked
[post-mortem] 4 learnings extracted → .agents/
```

## How It Works

```
  ╭── Research → Plan → Implement → Validate ──╮
  │                                            │
  ╰─────────── Knowledge Flywheel ◀────────────╯
```

```
  ╭─ worker-1 ─→ ✓ task      Fresh context per agent.
  ├─ worker-2 ─→ ✓ task      No bleed-through.
  ├─ worker-3 ─→ ✓ task      Lead validates + commits.
  ╰─ lead ─────→ validate
```

```
  ╭─ Claude ─→ PASS ─╮
  ├─ Claude ─→ WARN ─┼→ Consensus: WARN
  ╰─ Codex ──→ PASS ─╯
```

```
  Session 1:  learn → .agents/
  Session 2:  .agents/ → learn → .agents/
  Session 3:  .agents/ → learn → .agents/
              memory compounds ↑
```

```
  Hooks auto-enforce the loop:
  ┌─────────────────────────────────────────┐
  │ push without /vibe?        → blocked    │
  │ worker tries git commit?   → blocked    │
  │ forgot /pre-mortem?        → nudged     │
  │ force push to main?        → blocked    │
  └─────────────────────────────────────────┘
```

## Skills

### Orchestration

| Skill | What it does |
|-------|-------------|
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

## Going Deeper

**Optional: `ao` CLI** — auto-inject knowledge at session start, auto-extract at session end.

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops
ao hooks install
```

**Optional: `beads`** — git-native issue tracking. Lets `/crank` orchestrate multi-issue work from a tracked backlog.

See [docs/reference.md](docs/reference.md) for CLI reference, execution modes (local vs distributed), and tool dependencies.

## FAQ

**Why not just use Claude Code directly?**
Claude Code has agent spawning built in. AgentOps adds what it lacks:
- Cross-session memory (agents forget everything when the session ends)
- Codified patterns (isolation, validation contracts, debate protocol) that agents won't discover on their own
- Cross-vendor validation (`--mixed` mode adds Codex judges alongside Claude)
- Hooks that enforce the workflow without requiring discipline

**What data leaves my machine?**
Nothing. All state lives in `.agents/` (git-tracked, local). No telemetry, no cloud sync.

**How do I uninstall?**
```bash
npx skills@latest remove boshu2/agentops -g
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
