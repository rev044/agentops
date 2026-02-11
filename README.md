<div align="center">

# AgentOps

### Goal in, production code out. Every cycle makes the next one better.

[![Version](https://img.shields.io/badge/version-2.4.0-brightgreen)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-32-7c3aed)](skills/)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

[Install](#install) · [See It Work](#see-it-work) · [The Workflow](#the-workflow) · [Skills](#skills) · [FAQ](#faq) · [Docs](docs/)

</div>

---

## What Is It

- **One command, six phases.** `/rpi "goal"` runs research, planning, failure simulation, parallel implementation, code validation, and learning extraction — hands-free.
- **Self-correcting.** Failed validation retries with failure context. No human escalation.
- **Self-improving.** Every cycle extracts learnings and suggests the next `/rpi` command. The system improves its own tools.
- **All local.** Everything lives in `.agents/` and git. No cloud, no telemetry, no external services.

Works with **Claude Code**, **Codex CLI**, **Cursor**, **Open Code** — any agent that supports [Skills](https://skills.sh).

---

## Install

```bash
npx skills@latest add boshu2/agentops --all -g
```

Then inside your coding agent:

```bash
/quickstart
```

That's it.

<details>
<summary>Other install methods</summary>

**Claude Code plugin path:**
```bash
claude plugin add boshu2/agentops
```

**Install script** (plugin + optional CLI + hooks):
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install.sh)
```

**Manual:**
```bash
npx skills@latest add boshu2/agentops --all -g   # Plugin
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops   # CLI (optional)
ao hooks install   # Hooks (optional)
```

</details>

<details>
<summary>Troubleshooting</summary>

If slash commands don't appear: `npx skills@latest update`

More: [docs/troubleshooting.md](docs/troubleshooting.md)

</details>

---

## See It Work

```
> /rpi "add rate limiting to the API"

[research]    Exploring codebase... → .agents/research/rate-limiting.md
[plan]        3 issues, 2 waves → epic ag-0057
[pre-mortem]  3 judges → Verdict: PASS
[crank]       Wave 1: ███ 2/2 · Wave 2: █ 1/1
[vibe]        3 judges → Verdict: PASS
[post-mortem] 3 learnings extracted → .agents/
[flywheel]    Next: /rpi "add consistency-check finding category to /vibe"
```

You type one command and walk away. Come back, copy-paste the next command, walk away again.

<details>
<summary>More examples</summary>

**Multi-model validation:**
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
> /crank ag-0042

[crank] Epic: ag-0042 — 6 issues, 3 waves
[wave-1] ██████ 3/3 complete
[wave-2] ████── 2/2 complete
[wave-3] ██──── 1/1 complete
[vibe] PASS — all gates locked
[post-mortem] 4 learnings extracted
```

**Council standalone** (no setup, no workflow):
```
> /council validate this PR
> /council brainstorm caching strategies for the API
> /council research Redis vs Memcached for our use case
```

</details>

---

## The Workflow

1. **`/research`** — Explores your codebase. Produces a research artifact with findings and recommendations.

2. **`/plan`** — Decomposes the goal into issues with dependency waves. Creates a beads epic.

3. **`/pre-mortem`** — 3 judges simulate failures before you write code. FAIL? Re-plan with feedback and try again (max 3).

4. **`/crank`** — Spawns parallel agents in waves. Each worker gets fresh context. Lead validates and commits. Runs until every issue is closed.

5. **`/vibe`** — 3 judges validate the code. FAIL? Re-crank with failure context and re-vibe (max 3).

6. **`/post-mortem`** — Council validates the implementation. Retro extracts learnings. Synthesizes skill enhancements. **Suggests the next `/rpi` command.**

`/rpi "goal"` runs all six, end to end. Use `--interactive` if you want human gates at research and plan.

---

## The Flywheel

This is the core idea. Every cycle feeds the next one.

```
  /rpi "goal A"
    │
    ├── research → plan → pre-mortem → crank → vibe
    │
    ▼
  /post-mortem
    ├── council validates what shipped
    ├── retro extracts what you learned
    ├── synthesize skill enhancements       ← learnings improve the tools
    └── "Next: /rpi <enhancement>" ────┐
                                       │
  /rpi "goal B" ◄──────────────────────┘
    │
    └── ...repeat
```

Post-mortem doesn't just wrap up — it proposes how to make the skills better and hands you the command to do it. The system compounds. Each run makes the next one smarter.

---

## Skills

32 skills across three tiers, plus 10 internal skills that fire automatically.

### Orchestration

| Skill | What it does |
|-------|-------------|
| `/rpi` | Goal to production — 6-phase lifecycle with self-correcting retry loops |
| `/council` | Multi-model consensus — parallel judges, consolidated verdict |
| `/crank` | Autonomous epic execution — runs waves until all issues closed |
| `/swarm` | Parallel agents with fresh context — Codex sub-agents or Claude teams |
| `/codex-team` | Parallel Codex execution agents |

### Workflow

| Skill | What it does |
|-------|-------------|
| `/research` | Deep codebase exploration |
| `/plan` | Decompose goal into issues with dependency waves |
| `/implement` | Single issue, full lifecycle |
| `/vibe` | Complexity analysis + multi-model validation gate |
| `/pre-mortem` | Simulate failures before coding |
| `/post-mortem` | Validate implementation + extract learnings + suggest next cycle |
| `/release` | Pre-flight checks, changelog, version bumps, tag |

### Utilities

| Skill | What it does |
|-------|-------------|
| `/status` | Dashboard — current work, next action |
| `/quickstart` | Interactive onboarding |
| `/retro` | Extract learnings from completed work |
| `/knowledge` | Query knowledge base |
| `/bug-hunt` | Root cause analysis with git archaeology |
| `/complexity` | Code complexity metrics |
| `/doc` | Documentation generation and validation |
| `/trace` | Trace design decisions through history |
| `/handoff` | Structured session handoff |
| `/inbox` | Agent Mail monitoring |

<details>
<summary>Internal skills (auto-loaded, 10 total)</summary>

| Skill | Trigger | What it does |
|-------|---------|-------------|
| `inject` | Session start | Load prior knowledge into context |
| `extract` | On demand | Pull learnings from artifacts |
| `forge` | Session end | Mine transcript for decisions and patterns |
| `flywheel` | On demand | Knowledge health metrics |
| `ratchet` | On demand | Progress gates — once locked, stays locked |
| `standards` | By `/vibe`, `/implement` | Language-specific coding rules |
| `beads` | By `/plan`, `/implement` | Git-native issue tracking |
| `provenance` | On demand | Trace knowledge artifact lineage |
| `shared` | By distributed skills | Shared reference documents |
| `using-agentops` | Auto-injected | Workflow guide |

</details>

---

## How It Works

<details>
<summary><strong>The Brownian Ratchet</strong> — chaos in, locked progress out</summary>

```
  ╭─ agent-1 ─→ ✓ ─╮
  ├─ agent-2 ─→ ✗ ─┤   3 attempts, 1 fails
  ├─ agent-3 ─→ ✓ ─┤   council catches it
  ╰─ council ──→ PASS   ratchet locks the result
                  ↓
          can't go backward
```

Spawn parallel agents (chaos), validate with multi-model council (filter), merge to main (ratchet). Failed agents are cheap — fresh context means no contamination.

</details>

<details>
<summary><strong>Ralph Loops</strong> — fresh context every wave</summary>

```
  Wave 1:  spawn 3 workers → write files → lead validates → lead commits
  Wave 2:  spawn 2 workers → ...same pattern, zero accumulated context
```

Every wave gets a fresh worker set. Every worker gets clean context. No bleed-through between waves. The lead is the only one who commits.

Supports both Codex sub-agents (`spawn_agent`) and Claude agent teams (`TeamCreate`).

</details>

<details>
<summary><strong>Agent Backends</strong> — runtime-native orchestration</summary>

Skills auto-select the best available backend:

1. Codex sub-agents (`spawn_agent`)
2. Claude native teams (`TeamCreate` + `SendMessage`)
3. Background task fallback (`Task(run_in_background=true)`)

```
  Council:                               Swarm:
  ╭─ judge-1 ──╮                  ╭─ worker-1 ──╮
  ├─ judge-2 ──┼→ consolidate     ├─ worker-2 ──┼→ validate + commit
  ╰─ judge-3 ──╯                  ╰─ worker-3 ──╯
```

**Claude teams setup** (optional):
```json
// ~/.claude/settings.json
{
  "teammateMode": "tmux",
  "env": { "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1" }
}
```

</details>

<details>
<summary><strong>Hooks</strong> — the workflow enforces itself</summary>

11 hooks. All have a kill switch: `AGENTOPS_HOOKS_DISABLED=1`.

| Hook | Trigger | What it does |
|------|---------|-------------|
| Push gate | `git push` | Blocks push if `/vibe` hasn't passed |
| Worker guard | `git commit` | Blocks workers from committing (lead-only) |
| Dangerous git guard | `force-push`, `reset --hard` | Blocks destructive git commands |
| Ratchet nudge | Any prompt | "Run /vibe before pushing" |
| Task validation | Task completed | Validates metadata before accepting |

</details>

<details>
<summary><strong>The <code>ao</code> CLI</strong> — optional engine for the flywheel</summary>

All 32 skills work without it. The CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle.

```bash
ao inject              # Load prior knowledge
ao forge transcript    # Extract learnings from session
ao ratchet status      # Check progress gates
ao flywheel status     # Knowledge health metrics
ao session close       # Full lifecycle close
ao hooks install       # Install session hooks
```

Install: `brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops`

</details>

---

## FAQ

<details>
<summary><strong>Why not just use your coding agent directly?</strong></summary>

Your coding agent can spawn agents and write code. AgentOps makes it autonomous:

- **Self-correcting** — failed validation retries with failure context, not human escalation
- **Can't regress** — ratchet locks progress after each phase
- **Fresh context** — each worker gets clean context, no accumulated hallucinations
- **Self-improving** — every cycle proposes how to make the tools better
- **Cross-vendor** — `--mixed` mode adds Codex judges alongside Claude
- **Self-enforcing** — hooks block bad pushes, enforce lead-only commits

</details>

<details>
<summary><strong>What data leaves my machine?</strong></summary>

Nothing. All state lives in `.agents/` (git-tracked, local). No telemetry, no cloud, no external services.

</details>

<details>
<summary><strong>Can I use this with other AI coding tools?</strong></summary>

Yes — Claude Code, Codex CLI, Cursor, Open Code, anything supporting [Skills](https://skills.sh). The `--mixed` council mode adds Codex judges alongside Claude. Knowledge artifacts are plain markdown.

</details>

<details>
<summary><strong>How do I uninstall?</strong></summary>

```bash
npx skills@latest remove boshu2/agentops -g
brew uninstall agentops  # if installed
```

</details>

---

## Built On

| Project | Role |
|---------|------|
| [Ralph Wiggum pattern](https://ghuntley.com/ralph/) | Fresh context per agent — no bleed-through |
| [Multiclaude](https://github.com/dlorenc/multiclaude) | Validation gates that lock — no regression |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [Glossary](docs/GLOSSARY.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
