<div align="center">

<img src="docs/assets/hero.jpg" alt="Context engineering — crafting what enters the window" width="700">

# AgentOps

### The missing DevOps layer for coding agents. Give it a goal, it ships validated code — and remembers what worked.

*Context orchestration for every phase — research, planning, validation, execution. Each session learns from the last, so your agent compounds knowledge over time.*

[![GitHub stars](https://img.shields.io/github/stars/boshu2/agentops?style=social)](https://github.com/boshu2/agentops)
[![Version](https://img.shields.io/github/v/tag/boshu2/agentops?display_name=tag&sort=semver&label=version&color=8b5cf6)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-38-7c3aed)](skills/)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

[See It Work](#see-it-work) · [Install](#install) · [The Workflow](#the-workflow) · [The Flywheel](#the-flywheel) · [Skills](#skills) · [CLI](#the-ao-cli) · [FAQ](#faq)

</div>

---

> [!IMPORTANT]
> **No data leaves your machine.** All state lives in `.agents/` inside your repo (git-ignored by default). No telemetry, no cloud, no accounts. Every gate has a kill switch (`AGENTOPS_HOOKS_DISABLED=1`). The `ao` CLI is [open source](cli/) — audit it yourself. Apache-2.0.

**Pick your runtime, one command:**

```bash
# Claude Code, Codex CLI, Cursor (most users)
npx skills@latest add boshu2/agentops --all -g

# OpenCode
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-opencode.sh | bash
```

Requires Node.js 18+. Then type `/quickstart` in your agent chat. Full lifecycle: `/rpi "your goal"`.

Slash commands not appearing? Restart your agent, then `npx skills@latest update`. More install options: [Install](#install).

---

If you use Claude Code, Codex CLI, Cursor, or OpenCode and wish your agent remembered what it learned last session — this is for you.

**Every coding agent session starts at zero.** One model of knowledge decay ([Darr 1995](docs/the-science.md)) suggests ~17% loss per week without reinforcement — your mileage will vary, but the direction is real: without a feedback loop, agents don't accumulate expertise.

I come from DevOps, so I applied the [Three Ways](https://itrevolution.com/articles/the-three-ways-principles-underpinning-devops/) to agent workflows:

- **Flow** — Knowledge streams from session to forge to store to the next session. No batching, just continuous context.
- **Feedback** — Validation gates at every phase. Pre-mortems on plans, councils on code, ratchets that lock progress.
- **Continuous Learning** — Retros extract patterns, post-mortems close the loop. Failures become permanent learnings, not one-off incidents.

That's the architecture: a pipeline with a knowledge flywheel on top, so each session builds on the last.

**What's worked for me:**

1. **Cross-session memory**
   After each session, learnings are extracted, quality-gated, and injected into the next one automatically. There's a [formal threshold](docs/the-science.md) for when this tips from decay to compounding.

2. **Multi-model validation**
   Pre-mortem simulates failures on the plan *before* coding. Council reviews the code *after* — Claude and Codex judges debating each other. Failures retry automatically with context.

3. **Composable pieces**
   Use one skill or all of them. Wire them together when you're ready. `/rpi "goal"` runs the full lifecycle, but you don't have to start there.

[Detailed comparisons →](docs/comparisons/) · [Glossary →](docs/GLOSSARY.md) · [How it works →](docs/how-it-works.md) · [The Science →](docs/the-science.md)

---

## See It Work

**Use one piece.** No pipeline required — all skills work standalone:
```text
> /council validate this PR

[council] 3 judges spawned
[judge-1] PASS — JWT implementation correct
[judge-2] WARN — rate limiting missing on /login
[judge-3] PASS — refresh rotation implemented
Consensus: WARN — add rate limiting before shipping
```

**It remembers.** Three weeks later, different session:
```text
> /knowledge "rate limiting"

1. .agents/learnings/2026-01-28-rate-limiting.md
   [established] Token bucket with Redis — chose over sliding window for burst tolerance
2. .agents/patterns/api-middleware.md
   Pattern: rate limit at middleware layer, not per-handler
```
Your agent reads these automatically at session start. No copy-paste, no "remember last time we..."

**Wire it all together** when you're ready — one command, all six phases:
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

Session 2 was faster and better because session 1's learnings were already in context. That's the flywheel.

<details>
<summary><b>More examples</b> — /crank, /evolve</summary>

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

| Runtime | Status | Install method |
|---------|:------:|----------------|
| **Claude Code** | ✅ Best | Skills or plugin |
| **Codex CLI** | ✅ Strong | Skills |
| **Cursor** | ✅ Good | Skills |
| **OpenCode** | ✅ Good | Install script |

### Quick install

```bash
# Most runtimes (Claude Code, Codex CLI, Cursor)
npx skills@latest add boshu2/agentops --all -g
```

```bash
# OpenCode
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-opencode.sh | bash
```

```bash
# Claude Code plugin (alternative to skills)
claude plugin add boshu2/agentops
```

This installs all 38 skills. Then type `/quickstart` in your agent chat. That's it.

> [!NOTE]
> **What changes in your repo:** Skills install to `~/.claude/skills/` (global, outside your repo). If you later add the optional CLI + hooks, `ao init` creates a `.agents/` directory (git-ignored) for knowledge artifacts and registers hooks in `.claude/settings.json`. The session-start hook also auto-appends `.agents/` to your project `.gitignore` on first run (`AGENTOPS_GITIGNORE_AUTO=0` to disable). Nothing modifies your source code. Disable all hooks instantly: `AGENTOPS_HOOKS_DISABLED=1`.

<details>
<summary><b>Full setup</b> — CLI + hooks (optional)</summary>
<code>brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops</code><br>
<code>ao init --hooks --full</code> — all 12 hooks across 8 lifecycle event types. Adds knowledge injection/extraction, ratchet gates, session lifecycle. All 38 skills work without it.
</details>

<details>
<summary><b>OpenCode</b> — plugin + skills</summary>
Installs 7 hooks (tool enrichment, audit logging, compaction resilience) and symlinks all 38 skills. Restart OpenCode after install. Details: <a href=".opencode/INSTALL.md">.opencode/INSTALL.md</a>
</details>

Troubleshooting: [docs/troubleshooting.md](docs/troubleshooting.md)

---

## The Path

```text
/quickstart                          ← Day 1: guided tour on your codebase (~10 min)
    │
/council, /research, /vibe           ← Week 1: use skills standalone, learn the pieces
    │
/rpi "goal"                          ← Week 2: full lifecycle — research → ship → learn
    │
GOALS.yaml + /evolve                 ← Ongoing: define fitness goals, the system compounds
```

Start with `/quickstart`. Use individual skills when you need them. Graduate to `/rpi` for end-to-end. Set `GOALS.yaml` and let `/evolve` compound from there.

**Where does the knowledge live?** Everything the system learns is stored in `.agents/` — a local knowledge vault in your repo. Learnings, research artifacts, plans, fitness snapshots, and session handoffs all live there. Knowledge survives across sessions automatically. You never configure it; skills read from and write to it automatically.

---

## The Workflow

1. **`/research`** — Explores your codebase. Produces a research artifact with findings and recommendations.

2. **`/plan`** — Decomposes the goal into issues with dependency waves. Derives three-tier boundaries (Always / Ask First / Never) to prevent scope creep, and conformance checks — verifiable assertions generated from the spec itself. Creates a [beads](https://github.com/steveyegge/beads) epic (git-native issue tracking).

3. **`/pre-mortem`** — 4 judges simulate failures before you write code, including a spec-completeness judge that validates plan boundaries and conformance checks. FAIL? Re-plan with feedback and try again (max 3).

4. **`/crank`** — Spawns parallel agents in waves. Each worker gets fresh context. Cross-cutting constraints from the plan are injected into every wave's validation pass. `--test-first` uses a spec-first TDD model — specs and tests before implementation in every wave. Lead validates and commits. Runs until every issue is closed.

5. **`/vibe`** — 3 judges validate the code. FAIL? Re-crank with failure context and re-vibe (max 3).

6. **`/post-mortem`** — Council validates the implementation. Retro extracts learnings. Synthesizes process improvements. **Suggests the next `/rpi` command.**

`/rpi "goal"` runs all six, end to end (micro-epics of 2 or fewer issues auto-detect fast-path: inline validation instead of full council, ~15 min faster with no quality loss). Use `--interactive` if you want human gates at research and plan.

### Phased RPI: Own Your Context Window

For larger goals, `ao rpi phased "goal"` runs each phase in its own fresh session — no context bleed between phases. Supports `--interactive` (human gates at research/plan), `--from=<phase>` (resume), and parallel worktrees.

```bash
ao rpi phased "add rate limiting"      # Hands-free, fresh context per phase
ao rpi phased "add auth" &             # Run multiple in parallel (auto-worktrees)
ao rpi phased --from=crank "fix perf"  # Resume from any phase
```

Use `/rpi` when context fits in one session. Use `ao rpi phased` when it doesn't.

---

## The Flywheel

This is what makes AgentOps different. The system doesn't just run — it compounds.

```
  /rpi "goal A"
    │
    ├── research → plan → pre-mortem → crank → vibe
    │
    ▼
  /post-mortem
    ├── council validates what shipped
    ├── retro extracts what you learned
    ├── proposes how to improve the skills   ← the tools get better
    └── "Next: /rpi <enhancement>" ────┐
                                       │
  /rpi "goal B" ◄──────────────────────┘
    │
    └── ...repeat forever
```

Post-mortem doesn't just wrap up. It analyzes every learning from the retro, asks "what process would this improve?", and writes concrete improvement proposals. Then it hands you a ready-to-copy `/rpi` command targeting the highest-priority improvement. You come back, paste it, walk away. The system grows its knowledge stock with each cycle.

Learnings pass quality gates (specificity, actionability, novelty) and land in gold/silver/bronze tiers. [MemRL](https://arxiv.org/abs/2502.06173)-inspired freshness decay ensures recent insights outweigh stale patterns.

<details>
<summary><b>The Science</b> — escape-velocity condition, citations</summary>
Knowledge either compounds or decays to zero. The math: <code>σ × ρ > δ</code> (retrieval × usage > decay). When true, each session makes the next one better. Built on: Darr 1995 (decay rates), Sweller 1988 (cognitive load), Liu et al. 2023 (lost-in-the-middle), MemRL 2025 (RL for memory), Brownian Ratchet (thermodynamics). Deep dive: <a href="docs/the-science.md">docs/the-science.md</a>
</details>

<details>
<summary><b>Systems Leverage</b> — Meadows leverage points mapped to AgentOps</summary>
AgentOps concentrates on the high-leverage end of <a href="https://en.wikipedia.org/wiki/Twelve_leverage_points">Meadows' hierarchy</a>: information flows (#6 — hooks + <code>.agents/</code>), rules (#5 — ratchet gates), self-organization (#4 — modular skills + post-mortem harvesting), goals (#3 — <code>GOALS.yaml</code> + <code>/evolve</code>), paradigms (#2 — context quality as primary lever). The bet: changing the loop beats tuning the output.
</details>

### Goal-Driven Mode: `/evolve`

Define fitness goals in `GOALS.yaml`, then `/evolve` measures them, picks the worst gap, runs `/rpi` to fix it, re-measures ALL goals (regressed commits auto-revert), and loops. Kill switch: `echo "stop" > ~/.config/evolve/KILL`

```yaml
# GOALS.yaml — define goals, walk away
version: 1
goals:
  - id: test-pass-rate
    description: "All tests pass"
    check: "make test"
    weight: 10
```

Each `/rpi` cycle is smarter than the last because it learned from every cycle before it.

---

## From Vision to Execution

`/product` defines the vision. `/research` explores the landscape. `/plan` decomposes into issues with dependency waves. `/crank` spawns fresh-context workers per wave. `/vibe` validates. `/post-mortem` extracts learnings and suggests the next `/rpi` command. `/evolve` loops until all `GOALS.yaml` fitness goals pass.

You define the goal. The system builds it piece by piece — each cycle compounds on the last.

---

## Skills

38 skills: 28 user-facing, 10 internal (fire automatically). Start anywhere on this ladder — each level composes the ones below it.

| Scope | Skill | What it does |
|-------|-------|-------------|
| **Single review** | `/council` | Multiple judges (Claude + Codex) debate, surface disagreement, converge on a verdict |
| **Single issue** | `/implement` | Full lifecycle for one task — research, plan, build, validate, learn |
| **Multi-issue waves** | `/crank` | Parallel agents in dependency-ordered waves with fresh context per worker |
| **Full lifecycle** | `/rpi` | Research → Plan → Pre-mortem → Crank → Vibe → Post-mortem — one command, zero prompts |
| **Hands-free loop** | `/evolve` | Measures fitness goals, picks the worst gap, ships a fix, rolls back regressions, repeats |

**Supporting skills:** `/research`, `/plan`, `/vibe`, `/pre-mortem`, `/post-mortem`, `/status`, `/quickstart`, `/bug-hunt`, `/doc`, `/release`, `/knowledge`, `/handoff`

Full reference with all 38 skills: [docs/SKILLS.md](docs/SKILLS.md)

---

## Cross-Runtime Orchestration

AgentOps doesn't lock you into one agent runtime — it orchestrates across them. Claude can lead a team of Codex workers. Codex judges can review Claude's output. Mix and match:

| Spawning Backend | How it works | Best for |
|-----------------|-------------|----------|
| **Native teams** | `TeamCreate` + `SendMessage` — built into Claude Code | Tight coordination, debate, real-time messaging |
| **Background tasks** | `Task(run_in_background=true)` — fire-and-forget | Quick parallel work, no team overhead |
| **Codex sub-agents** | `/codex-team` — Claude orchestrates Codex workers | Cross-vendor validation, GPT judges on Claude code |
| **tmux + Agent Mail** | `/swarm --mode=distributed` — full process isolation | Long-running work, crash recovery, debugging stuck workers |

**The primitives are composable.** The RPI workflow is how I use them. You can:
- Have Claude plan and Codex implement (or vice versa)
- Run a `/council` with 3 Claude judges and 3 Codex judges simultaneously (`--mixed`)
- Spawn a distributed swarm that survives terminal disconnects
- Build your own orchestration patterns on top of these building blocks

Take the pieces. Experiment. Create your own workflow. If you find something that works better, I'd love to hear about it.

| Runtime | Support level | What works |
|---------|:---:|-------------|
| **Claude Code** | Best-in-class | Native teams + lifecycle hooks + full gate enforcement |
| **Codex CLI** | Strong | Skills + artifacts + `/codex-team` parallel execution |
| **OpenCode** | Good | Plugin with tool enrichment, audit logging, compaction resilience |

---

## How AgentOps Fits With Other Tools

These aren't competitors — they're fellow experiments in making coding agents actually work. The whole point is to customize your workflow. Use pieces from any of them. Mix and match. None of these should feel like vendor lock-in.

| Alternative | What it does well | Where AgentOps focuses differently |
|-------------|-------------------|-------------------------------------|
| **Direct agent use** (Claude Code, Cursor) | Full autonomy, simple | Adds cross-vendor councils, fresh-context waves, and cross-session memory |
| **Custom prompts** (.cursorrules, CLAUDE.md) | Flexible, version-controlled | Adds auto-extracted learnings that compound — static instructions can't do that |
| **Orchestrators** (CrewAI, AutoGen, LangGraph) | Multi-agent task routing | Focuses on what's *in* each agent's context window, not just routing between them |
| **CI/CD gates** (GitHub Actions, pre-commit) | Automated, enforced | Runs validation *before* coding (/pre-mortem) and *before* push (/vibe), not just after |
| **[GSD](https://github.com/glittercowboy/get-shit-done)** | Clean subagent spawning, fights context rot | Cross-session memory (GSD keeps context fresh *within* a session; AgentOps carries knowledge *between* sessions) |
| **[Compound Engineer](https://github.com/EveryInc/compound-engineering-plugin)** | Knowledge compounding, structured loop | Multi-model councils and validation gates — independent judges debating before and after code ships |

You don't pick one. You build a workflow that fits how you think.

[Detailed comparisons →](docs/comparisons/)

---

## How It Works

Parallel agents produce noisy output; councils filter it; ratchets lock progress so it can never regress. Every worker gets fresh context — no bleed-through between waves. 12 hooks enforce the workflow automatically (kill switch: `AGENTOPS_HOOKS_DISABLED=1`).

Deep dive: [docs/how-it-works.md](docs/how-it-works.md) — Brownian Ratchet, Ralph Loops, agent backends, hooks, context windowing.

---

## The `ao` CLI

Optional but recommended. The CLI is plumbing — skills and hooks call it automatically. You install it, your agent uses it.

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops
ao hooks install --full # All 12 hooks across 8 lifecycle event types
```

**The three commands you'll actually type:**
```bash
ao rpi phased "goal"   # Full RPI lifecycle, fresh context per phase
ao search "query"      # Search knowledge across files and chat history
ao demo                # Interactive demo of capabilities
```

Everything else runs automatically. Full reference: [CLI Commands](cli/docs/COMMANDS.md)

---

## FAQ

**No data leaves your machine.** All state lives in `.agents/` (local; git-ignored by default). No telemetry, no cloud. Works with Claude Code, Codex CLI, Cursor, OpenCode — anything supporting [Skills](https://skills.sh).

More questions: [docs/FAQ.md](docs/FAQ.md) — comparisons, limitations, subagent nesting, PRODUCT.md, uninstall.

---

<details>
<summary><b>Built on</b> — Ralph Wiggum, Multiclaude, beads, CASS, MemRL</summary>
<a href="https://ghuntley.com/ralph/">Ralph Wiggum</a> (fresh context per agent) · <a href="https://github.com/dlorenc/multiclaude">Multiclaude</a> (validation gates) · <a href="https://github.com/steveyegge/beads">beads</a> (git-native issues) · <a href="https://github.com/Dicklesworthstone/coding_agent_session_search">CASS</a> (session search) · <a href="https://arxiv.org/abs/2502.06173">MemRL</a> (cross-session memory)
</details>

## Contributing

<details>
<summary><b>Issue tracking</b> — Beads / <code>bd</code></summary>
Git-native issues in <code>.beads/</code>. <code>bd onboard</code> (setup) · <code>bd ready</code> (find work) · <code>bd show &lt;id&gt;</code> · <code>bd close &lt;id&gt;</code> · <code>bd sync</code>. More: <a href="AGENTS.md">AGENTS.md</a>
</details>

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [How It Works](docs/how-it-works.md) · [FAQ](docs/FAQ.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
