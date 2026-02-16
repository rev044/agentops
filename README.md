<div align="center">

<img src="docs/assets/hero.jpg" alt="Context engineering — crafting what enters the window" width="700">

# AgentOps

### The missing DevOps layer for coding agents. Give it a goal, it ships validated code and gets smarter.

*Context orchestration for every phase — research, planning, validation, execution.*

[![GitHub stars](https://img.shields.io/github/stars/boshu2/agentops?style=social)](https://github.com/boshu2/agentops)
[![Version](https://img.shields.io/badge/version-2.9.0-brightgreen)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-36-7c3aed)](skills/)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

[See It Work](#see-it-work) · [Install](#install) · [The Workflow](#the-workflow) · [The Flywheel](#the-flywheel) · [Skills](#skills) · [CLI](#the-ao-cli) · [FAQ](#faq)

</div>

---

**Not another workflow tool.** SDD gives you a spec-first pipeline. GSD gives you a fast loop. Both forget everything when the session ends. AgentOps is different: it's composable primitives that remember what they learned.

This started because I was hand-crafting every turn my coding agents took — writing the prompt, reviewing the output, feeding context back in, repeat. So I built the pieces: a skill for research, a skill for validation, a skill for planning, hooks to enforce the workflow, a CLI to wire it together. Each piece works standalone — `/council` validates a PR, `/research` explores a codebase, `/vibe` checks code quality. You use what you need and skip the rest.

What makes this different from "write spec → implement → check against spec → done":

1. **It remembers across sessions.** The system extracts what worked, what failed, and what patterns emerged — then injects that knowledge into the next session. Session 10 is smarter than session 1 because it learned from 1–9.
2. **It self-corrects.** Validation happens *before* coding (pre-mortem simulates failures on the plan) and *after* (multi-model council reviews the code). Failures retry automatically with context. No human escalation unless it fails 3 times.
3. **It's composable, not prescribed.** Use one skill or all of them. Wire them together when you're ready. `/rpi "goal"` runs the full lifecycle, but you don't have to start there.

[Detailed comparisons →](docs/comparisons/) · Works with **Claude Code**, **Codex CLI**, **Cursor**, **Open Code** — any agent that supports [Skills](https://skills.sh).

---

## See It Work

**Use one piece.** No pipeline required:
```
> /council validate this PR

[council] 3 judges spawned
[judge-1] PASS — JWT implementation correct
[judge-2] WARN — rate limiting missing on /login
[judge-3] PASS — refresh rotation implemented
Consensus: WARN — add rate limiting before shipping
```

**It remembers.** Three weeks later, different session:
```
> ao search "rate limiting"

1. .agents/learnings/2026-01-28-rate-limiting.md  (score: 0.92)
   [established] Token bucket with Redis — chose over sliding window for burst tolerance
2. .agents/patterns/api-middleware.md  (score: 0.84)
   Pattern: rate limit at middleware layer, not per-handler
```
Your agent reads these automatically at session start. No copy-paste, no "remember last time we..."

**Wire it all together** when you're ready:
```
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
```
> /crank ag-0042

[crank] Epic: ag-0042 — 6 issues, 3 waves
[wave-1] ██████ 3/3 complete
[wave-2] ████── 2/2 complete
[wave-3] ██──── 1/1 complete
[vibe] PASS — all gates locked
[post-mortem] 4 learnings extracted
```

**Goal-driven improvement loop:**
```
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
```
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

## Install

**Requires:** Node.js 18+ and a coding agent that supports [Skills](https://skills.sh) (Claude Code, Codex CLI, Cursor, Open Code).

```bash
npx skills@latest add boshu2/agentops --all -g
```

Then open your coding agent and type `/quickstart`. That's it.

<details>
<summary>Full setup (CLI + hooks)</summary>

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops
ao hooks install        # Flywheel hooks (SessionStart + Stop)
ao hooks install --full # All 12 hooks across 8 lifecycle events
```

The `ao` CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle. All 36 skills work without it.

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
<summary><strong>Recommended .gitignore</strong></summary>

AgentOps writes session artifacts, validation reports, and knowledge to `.agents/` in your repo. These files may contain absolute paths and sensitive tool output (e.g., gitleaks results). Add this to your `.gitignore`:

```gitignore
# AgentOps session artifacts
.agents/
.beads/
```

</details>

---

## The Path

```
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

For larger goals, `ao rpi phased "goal"` runs each phase in its own fresh Claude session — no context bleed between phases. Supports `--interactive` (human gates at research/plan), `--from=<phase>` (resume), and parallel worktrees.

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

### Goal-Driven Mode: `/evolve`

Define fitness goals in `GOALS.yaml`, then `/evolve` measures them, picks the worst gap, runs `/rpi` to fix it, re-measures ALL goals (regressed commits auto-revert), and loops. Kill switch: `echo "stop" > ~/.config/evolve/KILL`

---

## From Vision to Execution

`/product` defines the vision. `/research` explores the landscape. `/plan` decomposes into issues with dependency waves. `/crank` spawns fresh-context workers per wave. `/vibe` validates. `/post-mortem` extracts learnings and suggests the next `/rpi` command. `/evolve` loops until all `GOALS.yaml` fitness goals pass.

You define the town. The system builds it house by house — each cycle compounds on the last.

---

## Skills

36 skills: 26 user-facing, 10 internal (fire automatically).

| | Key skills |
|---|---|
| **Orchestration** | `/rpi` (full lifecycle), `/council` (multi-model consensus), `/crank` (parallel waves), `/evolve` (goal-driven loop) |
| **Workflow** | `/research`, `/plan`, `/implement`, `/vibe` (validate code), `/pre-mortem` (validate plans), `/post-mortem` |
| **Utilities** | `/status`, `/quickstart`, `/bug-hunt`, `/doc`, `/release`, `/knowledge`, `/handoff` |

Full reference with all 36 skills: [docs/SKILLS.md](docs/SKILLS.md)

---

## How It Works

Parallel agents produce noisy output; councils filter it; ratchets lock progress so it can never regress. Every worker gets fresh context — no bleed-through between waves. 12 hooks enforce the workflow automatically (kill switch: `AGENTOPS_HOOKS_DISABLED=1`).

Deep dive: [docs/how-it-works.md](docs/how-it-works.md) — Brownian Ratchet, Ralph Loops, agent backends, hooks, context windowing.

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

**No data leaves your machine.** All state lives in `.agents/` (local, git-tracked). No telemetry, no cloud. Works with Claude Code, Codex CLI, Cursor, Open Code — anything supporting [Skills](https://skills.sh).

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

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [How It Works](docs/how-it-works.md) · [FAQ](docs/FAQ.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
