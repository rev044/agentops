<div align="center">

# AgentOps

### Coding agents forget everything between sessions. This fixes that.

[![Version](https://img.shields.io/github/v/tag/boshu2/agentops?display_name=tag&sort=semver&label=version&color=8b5cf6)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-52-7c3aed)](skills/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

[How It Works](#how-it-works) · [See It Work](#see-it-work) · [Install](#install) · [The Path](#the-path) · [Skills](#skills) · [Deep Dive](#deep-dive) · [FAQ](#faq)

</div>

<div align="center">
<br>
<img src="docs/assets/swarm-6-rpi.png" alt="6 agents running full RPI development cycles in parallel — DevOps primitives, validation gates, knowledge flywheel, and human oversight composing into trusted autonomous execution" width="800">
<br>
<i>DevOps primitives, validation gates, a knowledge flywheel, hooks, and a CLI — composed into something you can actually trust.</i>
<br>
<sub>6 agents, each with <a href="https://ghuntley.com/ralph/">fresh context</a>, running full research → plan → implement cycles on independent epics.
Multi-model <a href="#skills">councils</a> gate every phase. Failures auto-retry. Regressions auto-revert. Hooks make the rules unavoidable.
A team leader coordinates. The human nudges from one terminal — or walks away.
Every session extracts learnings. The next session starts smarter. 50 files changed.</sub>
<br><br>
</div>

---

## How It Works

Coding agents get a blank context window every session. AgentOps automates the feedback loop so each session starts smarter than the last — learnings extracted, quality-gated, and re-injected automatically.

**DevOps' Three Ways:** flow, feedback, continual learning. AgentOps applies them to the agent loop, then compounds memory between sessions.

**The building blocks:** primitives you can mix and match into a custom pipeline that fits your workflow.

- **Flow:** orchestration skills that move WIP through the system. Research → plan → validate → build → review → learn — single-piece flow, minimizing context switches.
- **Feedback:** shorten the feedback loop until defects can't survive it. Multi-model councils catch issues before code ships. Hooks make the rules unavoidable — validation gates, push blocking, regression auto-revert. Problems found Friday don't wait until Monday.
- **Learning:** stop rediscovering what you already know. Every session extracts learnings, scores them, and re-injects them at the next session start. Knowledge compounds instead of decaying. Session 50 knows what session 1 learned the hard way.

---

## See It Work

**Use one piece.** No pipeline required — every skill works standalone:
```text
> /council validate this PR

[council] 3 judges spawned
[judge-1] PASS — JWT implementation correct
[judge-2] WARN — rate limiting missing on /login
[judge-3] PASS — refresh rotation implemented
Consensus: WARN — add rate limiting before shipping
```

**Three weeks later, different session:**
```text
> /knowledge "rate limiting"

1. .agents/learnings/2026-01-28-rate-limiting.md
   [established] Token bucket with Redis — chose over sliding window for burst tolerance
2. .agents/patterns/api-middleware.md
   Pattern: rate limit at middleware layer, not per-handler
```
Your agent reads these automatically at session start — no CLI required, just skills + `.agents/`.

**Wire it all together** when you're ready:
```text
> /rpi "add retry backoff to rate limiter"

[research]    Found 2 prior learnings on rate limiting (injected)
[plan]        2 issues, 1 wave → epic ag-0058
[pre-mortem]  4 judges → PASS (knew about Redis choice from prior session)
[crank]       Wave 1: ██ 2/2
[vibe]        3 judges → PASS
[post-mortem] 2 new learnings → .agents/
[flywheel]    Next: /rpi "add circuit breaker to external API calls"
```

<details>
<summary><b>More examples</b> — /crank, /evolve</summary>

<br>

**Parallel agents with fresh context:**
```text
> /crank ag-0042

[crank] Epic: ag-0042 — 6 issues, 3 waves
[wave-1] ██████ 3/3 complete
[wave-2] ████── 2/2 complete
[wave-3] ██──── 1/1 complete
[vibe] PASS — all gates locked
[post-mortem] 4 learnings extracted
```

**Goal-driven improvement loop:**
```text
> /evolve --max-cycles=5

[evolve] GOALS.yaml: 4 goals loaded
[cycle-1] Measuring fitness... 2/4 passing
         Worst gap: test-pass-rate (weight: 10)
         /rpi "Improve test-pass-rate" → 3 issues, 2 waves
         Re-measure: 3/4 passing ✓
[cycle-2] Worst gap: doc-coverage (weight: 7)
         /rpi "Improve doc-coverage" → 2 issues, 1 wave
         Re-measure: 4/4 passing ✓
[cycle-3] All goals met. Checking harvested work...
         Picked: "add smoke test for /evolve" (from post-mortem)
[teardown] /post-mortem → 5 learnings extracted
```

</details>

---

## Install

```bash
# Claude Code, Codex CLI, Cursor (most users)
npx skills@latest add boshu2/agentops --all -g

# OpenCode
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-opencode.sh | bash
```

Then type `/quickstart` in your agent chat.

```bash
# Claude Code plugin (alternative)
claude plugin add boshu2/agentops
```

<details>
<summary><b>Full setup</b> — CLI + hooks (optional)</summary>

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops
cd /path/to/your/repo
ao init --hooks --full
```

12 hooks across 8 lifecycle event types. Adds knowledge injection/extraction, validation gates, session lifecycle. All skills work without it.

</details>

<details>
<summary><b>OpenCode</b> — plugin + skills</summary>

Installs 7 hooks (tool enrichment, audit logging, compaction resilience) and symlinks all skills. Restart OpenCode after install. Details: [.opencode/INSTALL.md](.opencode/INSTALL.md)

</details>

<details>
<summary><b>Your data</b> — what it touches, where it lives, how to remove it</summary>

**Local-only. No telemetry. No cloud. No accounts.**

| What | Where | Reversible? |
|------|-------|:-----------:|
| Skills | Global skills dir (outside your repo; for Claude Code: `~/.claude/skills/`) | `npx skills@latest remove boshu2/agentops -g` |
| Knowledge artifacts | `.agents/` in your repo (git-ignored by default) | `rm -rf .agents/` |
| Hook registration | `.claude/settings.json` | `ao hooks uninstall` or delete entries |
| Git push gate | Pre-push hook (optional, only with CLI) | `AGENTOPS_HOOKS_DISABLED=1` |

Nothing modifies your source code. Nothing phones home. Everything is [open source](cli/) — audit it yourself.

</details>

Troubleshooting: [docs/troubleshooting.md](docs/troubleshooting.md)

---

## The Path

```text
/quickstart                          ← Day 1: guided tour on your codebase (~10 min)
    │
/council, /research, /vibe           ← Week 1: use skills standalone
    │
/rpi "goal"                          ← Week 2: full lifecycle — research → ship → learn
    │
/product → /goals generate           ← Define what good looks like
    │
/evolve                              ← Ongoing: measure goals, fix gaps, compound
```

Start with `/quickstart`. Use individual skills when you need them. Graduate to `/rpi` for end-to-end. When you're ready for hands-free improvement: `/product` defines your mission and personas, `/goals generate` scans for fitness goals, and `/evolve` pursues them.

---

## Deep Dive

Standard iterative development — research, plan, validate, build, review, learn — automated for agents that can't carry context between sessions.

This is DevOps thinking applied to agent work: the **Three Ways** as composable primitives.

- **Flow**: wave-based execution (`/crank`) + workflow orchestration (`/rpi`) to keep work moving.
- **Feedback**: shift-left validation (`/pre-mortem`, `/vibe`, `/council`) plus optional gates/hooks to make feedback unavoidable.
- **Continual learning**: post-mortems turn outcomes into reusable knowledge in `.agents/`, so the next session starts smarter.

```
  /rpi "goal"
    │
    ├── /research → /plan → /pre-mortem → /crank → /vibe
    │
    ▼
  /post-mortem
    ├── validates what shipped
    ├── extracts learnings → .agents/
    └── suggests next /rpi command ────┐
                                       │
   /rpi "next goal" ◄──────────────────┘
```

The post-mortem analyzes each learning, asks "what process would this improve?", and writes improvement proposals. It hands you a ready-to-copy `/rpi` command. Paste it, walk away.

Learnings pass quality gates (specificity, actionability, novelty) and land in tiered pools. Freshness decay ensures recent insights outweigh stale patterns. The [formal model](docs/the-science.md) is straightforward: if retrieval rate × usage rate exceeds decay rate, knowledge compounds. If not, it decays to zero.

<details>
<summary><b>Phase details</b> — what each step does</summary>

1. **`/research`** — Explores your codebase. Produces a research artifact with findings and recommendations.

2. **`/plan`** — Decomposes the goal into issues with dependency waves. Derives scope boundaries and conformance checks. Creates a [beads](https://github.com/steveyegge/beads) epic (git-native issue tracking).

3. **`/pre-mortem`** — Judges simulate failures before you write code, including a spec-completeness judge. FAIL? Re-plan with feedback (max 3 retries).

4. **`/crank`** — Spawns parallel agents in dependency-ordered waves. Each worker gets fresh context. Lead validates and commits. Runs until every issue is closed. `--test-first` for spec-first TDD.

5. **`/vibe`** — Judges validate the code. FAIL? Re-crank with failure context and re-vibe (max 3).

6. **`/post-mortem`** — Council validates the implementation. Retro extracts learnings. **Suggests the next `/rpi` command.**

`/rpi "goal"` runs all six end to end. Use `--interactive` for human gates at research and plan.

</details>

<details>
<summary><b>Phased RPI</b> — fresh context per phase for larger goals</summary>

`ao rpi phased "goal"` runs each phase in its own session — no context bleed between phases.

```bash
ao rpi phased "add rate limiting"      # Hands-free, fresh context per phase
ao rpi phased "add auth" &             # Run multiple in parallel (auto-worktrees)
ao rpi phased --from=implementation "fix perf"  # Resume at execution phase
ao rpi status --watch                   # Monitor active phased runs
```

Use `/rpi` when context fits in one session. Use `ao rpi phased` when it doesn't.

</details>

<details>
<summary><b>Goal-driven mode</b> — /evolve with GOALS.yaml</summary>

Bootstrap with `/goals generate` — it scans your repo (PRODUCT.md, README, skills, tests) and proposes mechanically verifiable goals. Or write them by hand:

```yaml
# GOALS.yaml
version: 1
goals:
  - id: test-pass-rate
    description: "All tests pass"
    check: "make test"
    weight: 10
```

Then `/evolve` measures them, picks the worst gap, runs `/rpi` to fix it, re-measures ALL goals (regressed commits auto-revert), and loops. It commits locally — you control when to push. Kill switch: `echo "stop" > ~/.config/evolve/KILL`

Maintain over time: `/goals` shows pass/fail status, `/goals prune` finds stale or broken checks.

</details>

<details>
<summary><b>References</b> — science, systems theory, prior art</summary>

Built on [Darr 1995](docs/the-science.md) (decay rates), Sweller 1988 (cognitive load), [Liu et al. 2023](docs/the-science.md) (lost-in-the-middle), [MemRL 2025](https://arxiv.org/abs/2502.06173) (RL for memory).

AgentOps concentrates on the high-leverage end of [Meadows' hierarchy](https://en.wikipedia.org/wiki/Twelve_leverage_points): information flows (#6), rules (#5), self-organization (#4), goals (#3). The bet: changing the loop beats tuning the output.

Deep dive: [docs/how-it-works.md](docs/how-it-works.md) — Brownian Ratchet, Ralph Loops, agent backends, hooks, context windowing.

</details>

---

## Skills

52 skills: 42 user-facing, 10 internal (fire automatically). Each level composes the ones below it.

| Scope | Skill | What it does |
|-------|-------|-------------|
| **Single review** | `/council` | Multiple judges (Claude + Codex) debate, surface disagreement, converge on a verdict. Customize with `--preset=security-audit`, `--perspectives="a,b,c"`, or `--perspectives-file` |
| **Single issue** | `/implement` | Full lifecycle for one task — research, plan, build, validate, learn |
| **Multi-issue waves** | `/crank` | Parallel agents in dependency-ordered waves with fresh context per worker |
| **Full lifecycle** | `/rpi` | Research → Plan → Pre-mortem → Crank → Vibe → Post-mortem — one command |
| **Hands-free loop** | `/evolve` | Measures fitness goals, picks the worst gap, ships a fix, rolls back regressions, repeats |

**Supporting skills:** `/research`, `/plan`, `/vibe`, `/pre-mortem`, `/post-mortem`, `/product`, `/goals`, `/readme`, `/status`, `/quickstart`, `/bug-hunt`, `/doc`, `/release`, `/knowledge`, `/handoff`

Full reference: [docs/SKILLS.md](docs/SKILLS.md)

<details>
<summary><b>Cross-runtime orchestration</b> — mix Claude, Codex, OpenCode</summary>

AgentOps orchestrates across runtimes. Claude can lead a team of Codex workers. Codex judges can review Claude's output.

| Spawning Backend | How it works | Best for |
|-----------------|-------------|----------|
| **Native teams** | `TeamCreate` + `SendMessage` — built into Claude Code | Tight coordination, debate |
| **Background tasks** | `Task(run_in_background=true)` — last-resort fallback | When no team APIs available |
| **Codex sub-agents** | `/codex-team` — Claude orchestrates Codex workers | Cross-vendor validation |
| **tmux + Agent Mail** | `/swarm --mode=distributed` — full process isolation | Long-running work, crash recovery |

</details>

---

## How AgentOps Fits With Other Tools

These are fellow experiments in making coding agents work. Use pieces from any of them.

| Alternative | What it does well | Where AgentOps focuses differently |
|-------------|-------------------|-------------------------------------|
| **[GSD](https://github.com/glittercowboy/get-shit-done)** | Clean subagent spawning, fights context rot | Cross-session memory (GSD keeps context fresh *within* a session; AgentOps carries knowledge *between* sessions) |
| **[Compound Engineer](https://github.com/EveryInc/compound-engineering-plugin)** | Knowledge compounding, structured loop | Multi-model councils and validation gates — independent judges debating before and after code ships |

[Detailed comparisons →](docs/comparisons/)

---

## The `ao` CLI

Optional. If you already use skills directly in chat, you can skip it.

The `ao` CLI runs the full RPI lifecycle from your terminal — no active chat session required. Each of the three phases (discovery, implementation, validation) gets its own fresh context window, so large goals don't hit context limits.

```bash
ao rpi phased "add rate limiting"              # 3 sessions: discover → build → validate
ao rpi phased "fix auth bug" &                 # Run multiple in parallel (auto-worktrees)
ao rpi phased --from=implementation "ag-058"   # Resume at build phase
ao rpi status --watch                          # Monitor active runs
```

Walk away, come back to committed code + extracted learnings.

```bash
ao search "query"      # Search knowledge across files and chat history
ao demo                # Interactive demo
```

Full reference: [CLI Commands](cli/docs/COMMANDS.md)

---

## FAQ

[docs/FAQ.md](docs/FAQ.md) — comparisons, limitations, subagent nesting, PRODUCT.md, uninstall.

---

<details>
<summary><b>Built on</b> — Ralph Wiggum, Multiclaude, beads, CASS, MemRL</summary>

[Ralph Wiggum](https://ghuntley.com/ralph/) (fresh context per agent) · [Multiclaude](https://github.com/dlorenc/multiclaude) (validation gates) · [beads](https://github.com/steveyegge/beads) (git-native issues) · [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search) (session search) · [MemRL](https://arxiv.org/abs/2502.06173) (cross-session memory)

</details>

## Contributing

<details>
<summary><b>Issue tracking</b> — Beads / bd</summary>

Git-native issues in `.beads/`. `bd onboard` (setup) · `bd ready` (find work) · `bd show <id>` · `bd close <id>` · `bd sync`. More: [AGENTS.md](AGENTS.md)

</details>

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [How It Works](docs/how-it-works.md) · [FAQ](docs/FAQ.md) · [Glossary](docs/GLOSSARY.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
