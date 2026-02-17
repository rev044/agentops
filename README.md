<div align="center">

<img src="docs/assets/hero.jpg" alt="Context engineering — crafting what enters the window" width="700">

# AgentOps

### The missing DevOps layer for coding agents. Give it a goal, it ships validated code — and remembers what worked.

*Context orchestration for every phase — research, planning, validation, execution. Each session learns from the last, so your agent compounds knowledge over time.*

[![Version](https://img.shields.io/github/v/tag/boshu2/agentops?display_name=tag&sort=semver&label=version&color=8b5cf6)](CHANGELOG.md)

[See It Work](#see-it-work) · [Orchestration](#cross-runtime-agent-orchestration) · [Install](#install) · [The Workflow](#the-workflow) · [The Flywheel](#the-flywheel) · [Skills](#skills) · [The Science](#the-science) · [CLI](#the-ao-cli) · [FAQ](#faq)

</div>

---

**Quickstart (TL;DR):**

Skills install (start here):

```bash
npx skills@latest add boshu2/agentops --all -g
```

Requires: Node.js 18+.

Optional: Claude Code plugin install (if you prefer that):

```bash
claude plugin add boshu2/agentops
```

Optional: OpenCode install:

```bash
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-opencode.sh | bash
```

Then in your coding agent chat:

```text
/quickstart
```

Next (full lifecycle):

```text
/rpi "your goal"
```

If slash commands don’t appear: restart your agent; then try `npx skills@latest update`. Optional CLI + hooks: see [Install](#install). Bigger goal? See [Phased RPI](#phased-rpi-own-your-context-window).

**Every coding agent session starts at zero.** Knowledge decays at 17% per week without reinforcement ([Darr 1995](docs/the-science.md)). By week 4, half of what your agent learned is gone. By week 8, it's running on 22% of what it once knew.

I come from DevOps, so I started treating my agent like a pipeline — isolated stages, validated gates, fresh context at each phase. Then I built a knowledge flywheel on top so each session could build on the last instead of starting over.

**Measured results:**

| Metric | Without AgentOps | With AgentOps |
|--------|:---:|:---:|
| Same-issue resolution | 45 min | **3 min (15x)** |
| Token cost per issue | $2.40 | **$0.15 (16x)** |
| Context collapse | ~65% at 60% load | **Eliminated** |
| Pre-mortem ROI | — | **7 consecutive epics, 10/10 findings pre-code** |

<sub>Source: [docs/the-science.md](docs/the-science.md) Part 8 (resolution, cost, context). Pre-mortem ROI from internal testing across 7 epics.</sub>

**What's worked for me:**

1. **Cross-session memory.** The system extracts what worked, what failed, and what patterns emerged — then injects quality-gated knowledge into the next session. Session 10 is smarter than session 1 because it learned from 1–9. There's a [formal threshold](docs/the-science.md) for when this tips from decay to compounding.
2. **Multi-model validation.** Pre-mortem simulates failures on the plan *before* coding. Council reviews the code *after* — Claude and Codex judges debating each other. Failures retry automatically with context.
3. **Composable pieces.** Use one skill or all of them. Wire them together when you're ready. `/rpi "goal"` runs the full lifecycle, but you don't have to start there.

[Detailed comparisons →](docs/comparisons/) · [Glossary →](docs/GLOSSARY.md) · [How it works →](docs/how-it-works.md) · [The Science →](docs/the-science.md)

---

## See It Work

**Use one piece.** No pipeline required:
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
> ao search "rate limiting"

1. .agents/learnings/2026-01-28-rate-limiting.md  (score: 0.92)
   [established] Token bucket with Redis — chose over sliding window for burst tolerance
2. .agents/patterns/api-middleware.md  (score: 0.84)
   Pattern: rate limit at middleware layer, not per-handler
```
Your agent reads these automatically at session start. No copy-paste, no "remember last time we..."

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

Session 2 was faster and better because session 1's learnings were already in context. That's the flywheel.

<details>
<summary>More examples</summary>

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

**From vision to town** (big goal → many small pieces):
```text
> /product                    # define mission, personas, value props
> /research "build auth system"
> /plan "build auth system"   # → 8 issues, 3 waves

> /evolve --max-cycles=3
[cycle-1] /rpi "add user model + migrations" → 2 issues, 1 wave ✓
[cycle-2] /rpi "add login/signup endpoints" → 3 issues, 1 wave ✓
[cycle-3] /rpi "add JWT refresh + middleware" → 3 issues, 2 waves ✓
[teardown] 9 learnings extracted. All goals met.
```

</details>

---

## Cross-Runtime Agent Orchestration

AgentOps doesn't lock you into one agent runtime — it orchestrates across them. Claude can lead a team of Codex workers. Codex judges can review Claude's output. A tmux swarm can run persistent workers that survive disconnects. Mix and match:

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
- Use background tasks for quick parallel exploration, native teams for coordinated waves
- Build your own orchestration patterns on top of these building blocks

Take the pieces. Experiment. Create your own workflow. If you find something that works better, I'd love to hear about it.

| Runtime | Support level | What works |
|---------|:---:|-------------|
| **Claude Code** | Best-in-class | Native teams + lifecycle hooks + full gate enforcement |
| **Codex CLI** | Strong | Skills + artifacts + `/codex-team` parallel execution |
| **OpenCode** | Good | Plugin with tool enrichment, audit logging, compaction resilience |

---

## Install

**Requires:** Node.js 18+ and a coding agent that supports [Skills](https://skills.sh) (Claude Code, Codex CLI, Cursor, Open Code).

```bash
npx skills@latest add boshu2/agentops --all -g
```

This installs all 37 skills. For lifecycle hooks (coding standards, git safety, task validation), see **Full setup** below.

Then open your coding agent and type `/quickstart`. That's it.

<details>
<summary>Full setup (CLI + hooks)</summary>

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops
ao init              # Directories + .gitignore (idempotent)
ao init --hooks      # + minimal hooks (SessionStart + Stop only — 2/8 events)
ao init --hooks --full  # + all 12 hook scripts across 8 lifecycle events (recommended)
```

The `ao` CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle. `ao init` is the canonical setup command — creates all `.agents/` directories, configures `.gitignore`, and optionally registers hooks. All 37 skills work without it.

</details>

<details>
<summary>OpenCode install</summary>

```bash
curl -fsSL https://raw.githubusercontent.com/boshu2/agentops/main/scripts/install-opencode.sh | bash
```

This installs the AgentOps plugin (7 hooks for tool enrichment, audit logging, and compaction resilience) and symlinks all 37 skills. Restart OpenCode after install.

**Key difference from Claude Code:** OpenCode's `skill` tool is **read-only** — it loads skill content into context instead of executing it. The plugin handles this automatically with prescriptive tool mapping so models like Devstral know exactly which tools to call for each skill.

Full details: [.opencode/INSTALL.md](.opencode/INSTALL.md)

</details>

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

</details>

<details>
<summary>Troubleshooting</summary>

If slash commands don't appear: `npx skills@latest update`

More: [docs/troubleshooting.md](docs/troubleshooting.md)

</details>

<details>
<summary><strong>.gitignore</strong></summary>

`ao init` automatically adds `.agents/` to your `.gitignore`. If you prefer stealth mode (no `.gitignore` modification), use `ao init --stealth` to write to `.git/info/exclude` instead. The session-start hook also auto-adds the entry as a safety net.

If you're using beads for issue tracking: do **not** add `.beads/` to `.gitignore` (it's where issues live and should be committed).

</details>

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

**RPI = Research → Plan → Implement.** `/rpi "goal"` runs the full lifecycle (with pre-mortem, vibe, and post-mortem gates around implementation).

1. **`/research`** — Explores your codebase. Produces a research artifact with findings and recommendations.

2. **`/plan`** — Decomposes the goal into issues with dependency waves. Derives three-tier boundaries (Always / Ask First / Never) to prevent scope creep, and conformance checks — verifiable assertions generated from the spec itself. Creates a [beads](https://github.com/steveyegge/beads) epic (git-native issue tracking) if `bd` is available; otherwise it still writes a plan artifact to `.agents/plans/`.

3. **`/pre-mortem`** — 4 judges simulate failures before you write code, including a spec-completeness judge that validates plan boundaries and conformance checks. FAIL? Re-plan with feedback and try again (max 3).

4. **`/crank`** — Spawns parallel agents in waves. Each worker gets fresh context. Cross-cutting constraints from the plan are injected into every wave's validation pass. `--test-first` uses a spec-first TDD model — specs and tests before implementation in every wave. Lead validates and commits. Runs until every issue is closed.

5. **`/vibe`** — 3 judges validate the code. FAIL? Re-crank with failure context and re-vibe (max 3).

6. **`/post-mortem`** — Council validates the implementation. Retro extracts learnings. Synthesizes process improvements. **Suggests the next `/rpi` command.**

`/rpi "goal"` runs all six, end to end (micro-epics of 2 or fewer issues auto-detect fast-path: inline validation instead of full council, ~15 min faster with no quality loss). Use `--interactive` if you want human gates at research and plan.

### Phased RPI: Own Your Context Window

For larger goals, `ao rpi phased "goal"` runs each phase in its own fresh Claude session — no context bleed between phases. Supports `--interactive` (human gates at research/plan), `--from=<phase>` (resume), and parallel worktrees.

```bash
ao rpi phased "add rate limiting"      # Hands-free, fresh context per phase
ao rpi phased "add auth" &             # Run multiple in parallel (auto-worktrees)
ao rpi phased --from=crank "fix perf"  # Resume from any phase
```

Use `/rpi` when context fits in one session. Use `ao rpi phased` when it doesn't.

---

## The Flywheel

Each session learns from every session before it. This is the part I'm most excited about.

```text
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

Post-mortem analyzes every learning from the retro, asks "what process would this improve?", and writes concrete improvement proposals. Then it hands you a ready-to-copy `/rpi` command targeting the highest-priority improvement. You come back, paste it, walk away. The knowledge base grows with each cycle.

Learnings pass quality gates (specificity, actionability, novelty) and land in gold/silver/bronze tiers. [MemRL](https://arxiv.org/abs/2502.06173)-inspired freshness decay ensures recent insights outweigh stale patterns.

**The escape-velocity condition:** Knowledge either compounds or decays to zero — there's no stable middle. The math: `σ × ρ > δ` (retrieval effectiveness × usage rate > decay rate of 0.17/week). When true, each session makes the next one better. When false, everything erodes. The whole system is designed around staying above this threshold. [The math →](docs/the-science.md)

### Systems Leverage (Meadows)

AgentOps is systems engineering. Donella Meadows' leverage points are a useful map: the highest leverage changes are not "tune a parameter" but "change how the system learns, adapts, and decides." AgentOps concentrates on the high-leverage end:

| Meadows leverage point (high leverage) | How AgentOps intervenes |
|---|---|
| **#6 Information flows** | Hooks + `ao search` + `.agents/` artifacts move validated prior decisions into the current context window automatically. |
| **#5 Rules and constraints** | Ratchet gates (push gate, worker guard, pre-mortem gate) make "validated work" the default, not a suggestion. |
| **#4 Self-organization** | Skills are modular, installable units; post-mortem generates process improvements and queues next work. |
| **#3 Goals** | `GOALS.yaml` makes the objective function explicit and executable; `/evolve` optimizes toward it. |
| **#2 Paradigms** | "Context quality is the primary lever" and "the cycle is the product" shift the work from prompt craft to workflow + feedback design. |
| **#1 Transcend paradigms** | Cross-runtime orchestration (Claude, Codex, OpenCode) and graceful degradation keep the loop usable as tools/models change. |

<details>
<summary>Full Meadows leverage-point ladder (12 to 1)</summary>

12. Constants, parameters, numbers
11. Buffers and stabilizing stocks
10. Material stocks and flows (structure)
9. Delays
8. Negative feedback loops
7. Positive feedback loops
6. Information flows
5. Rules
4. Self-organization
3. Goals
2. Paradigms
1. Transcend paradigms

<sub>Reference: Donella H. Meadows, "Places to Intervene in a System" (1999) / <em>Thinking in Systems</em> (2008).</sub>
</details>

AgentOps deliberately spends less energy on the low-leverage end (prompt knobs, bigger buffers, "just add context") because those gains do not compound. The bet is that changing the loop beats tuning the output.

### `/evolve`

`/evolve` ties the whole thing together — a hands-free loop that ships validated code toward measurable goals.

**How it works:**
1. You define fitness goals in `GOALS.yaml` (test pass rate, doc coverage, complexity targets — anything measurable)
2. `/evolve` measures all goals, selects the worst gap by weight, and runs a full `/rpi` lifecycle to fix it
3. After each cycle, it re-measures **ALL** goals — not just the one it worked on
4. Learnings from each cycle feed back into the flywheel before the next cycle starts
5. It loops until every goal passes or you pull the kill switch

Passing goals stay passing — if a cycle breaks something that was working, the commits get rolled back automatically.

```bash
# Define goals, walk away
echo "test-pass-rate: {weight: 10, command: 'make test'}" > GOALS.yaml
/evolve --max-cycles=5

# Kill switch (immediate stop after current cycle)
echo "stop" > ~/.config/evolve/KILL
```

Each `/rpi` cycle is smarter than the last because it learned from every cycle before it.

---

## Skills

37 skills: 27 user-facing, 10 internal (fire automatically). Start anywhere on this ladder — each level composes the ones below it.

| Scope | Skill | What it does |
|-------|-------|-------------|
| **Single review** | `/council` | Multiple judges (Claude + Codex) debate, surface disagreement, converge on a verdict |
| **Single issue** | `/implement` | Full lifecycle for one task — research, plan, build, validate, learn |
| **Multi-issue waves** | `/crank` | Parallel agents in dependency-ordered waves with fresh context per worker |
| **Full lifecycle** | `/rpi` | Research → Plan → Pre-mortem → Crank → Vibe → Post-mortem — one command, zero prompts |
| **Hands-free loop** | `/evolve` | Measures fitness goals, picks the worst gap, ships a fix, rolls back regressions, repeats |

**Supporting skills:** `/research`, `/plan`, `/vibe`, `/pre-mortem`, `/post-mortem`, `/status`, `/quickstart`, `/bug-hunt`, `/doc`, `/release`, `/knowledge`, `/handoff`

Full reference with all 37 skills: [docs/SKILLS.md](docs/SKILLS.md)

### Why Not Just Use...

| Alternative | What it does well | Where AgentOps focuses differently |
|-------------|-------------------|-------------------|
| **Direct agent use** (Claude Code, Cursor) | Full autonomy, simple | Adds cross-vendor councils, fresh-context waves, and cross-session memory |
| **Custom prompts** (.cursorrules, CLAUDE.md) | Flexible, version-controlled | Adds auto-extracted learnings that compound — static instructions can't do that |
| **Orchestrators** (CrewAI, AutoGen, LangGraph) | Multi-agent task routing | Focuses on what's *in* each agent's context window, not just routing between them |
| **CI/CD gates** (GitHub Actions, pre-commit) | Automated, enforced | Runs validation *before* coding (`/pre-mortem`) and *before* push (`/vibe`), not just after |

[Detailed comparisons →](docs/comparisons/)

---

## The Ratchet Guarantee

The idea comes from thermodynamics: a [Brownian Ratchet](docs/the-science.md) gets forward movement from random motion by only allowing progress in one direction. Same thing here — parallel agents produce noisy output, councils filter it, ratchets lock the gains.

**Enforcement mechanisms (12 hooks across 8 lifecycle events):**

| Gate | What it does | Enforcement |
|------|-------------|-------------|
| **Push gate** | `git push` is physically blocked until `/vibe` passes | PostToolUse hook on git push |
| **Pre-mortem gate** | `/crank` cannot start until `/pre-mortem` passes (3+ issue epics) | Task validation hook |
| **Worker guard** | Workers write files but cannot `git commit` — lead-only commits | PostToolUse hook on git commit |
| **Regression gate** | `/evolve` auto-reverts commits that cause passing goals to fail | Post-cycle measurement |
| **Dangerous git guard** | `force-push`, `reset --hard`, `clean -f` blocked by default | PostToolUse hook on git |

Every gate has a kill switch (`AGENTOPS_HOOKS_DISABLED=1`). Safety on by default, opt out when you want to.

Deep dive: [docs/how-it-works.md](docs/how-it-works.md) — Brownian Ratchet, Ralph Loops, agent backends, hooks, context windowing.

---

## The Science

I built AgentOps on top of research I found useful: knowledge decay rates (Darr 1995), cognitive load theory (Sweller 1988, Liu et al. 2023), reinforcement learning for memory (MemRL 2025), and the Brownian Ratchet from thermodynamics. The escape-velocity condition (`σ × ρ > δ`) is falsifiable — either your knowledge compounds or it doesn't.

Deep dive: [docs/the-science.md](docs/the-science.md) — the math, the evidence, the citations.

---

## The `ao` CLI

Optional but recommended. The CLI is plumbing — skills and hooks call it automatically. You install it, your agent uses it.

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops
ao hooks install --full # All 12 hooks across 8 lifecycle events
```

**The three commands you'll actually type:**
```bash
ao rpi phased "goal"   # Full RPI lifecycle, fresh context per phase
ao search "query"      # Search knowledge across files and chat history
ao demo                # Interactive demo of capabilities
```

Everything else runs automatically. 73 commands total — full reference: [CLI Commands](cli/docs/COMMANDS.md)

---

## FAQ

**No data leaves your machine.** All state lives in `.agents/` (local; git-ignored by default). No telemetry, no cloud. Works with Claude Code, Codex CLI, Cursor, Open Code — anything supporting [Skills](https://skills.sh).

More questions: [docs/FAQ.md](docs/FAQ.md) — comparisons, limitations, subagent nesting, PRODUCT.md, uninstall.

---

<details>
<summary><strong>Built on</strong></summary>

| Project | Role |
|---------|------|
| [Ralph Wiggum pattern](https://ghuntley.com/ralph/) | Fresh context per agent — no bleed-through |
| [Multiclaude](https://github.com/dlorenc/multiclaude) | Validation gates that lock — no regression |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search) | Unified search across coding agent chat histories |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

</details>

## Contributing

<details>
<summary><strong>Issue tracking (Beads / <code>bd</code>)</strong></summary>

This repo tracks work in `.beads/` (git-native issues).

```bash
bd onboard                           # one-time setup for this repo
bd ready                             # find available work
bd show <id>                         # view issue details
bd update <id> --status in_progress  # claim work
bd close <id>                        # complete work
bd sync                              # sync with git
```

More: [AGENTS.md](AGENTS.md)

</details>

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [How It Works](docs/how-it-works.md) · [FAQ](docs/FAQ.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
