<div align="center">

# AgentOps

### DevOps for AI agents. The system that gets smarter every time you use it.

[![GitHub stars](https://img.shields.io/github/stars/boshu2/agentops?style=social)](https://github.com/boshu2/agentops)
[![Version](https://img.shields.io/badge/version-2.5.0-brightgreen)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-34-7c3aed)](skills/)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

[Install](#install) · [See It Work](#see-it-work) · [The Workflow](#the-workflow) · [The Flywheel](#the-flywheel) · [Skills](#skills) · [FAQ](#faq)

</div>

---

Your coding agent can write code. But it doesn't know what it learned last session. It doesn't validate its own work. It doesn't plan before it builds, or extract learnings after it ships. You're the glue — managing context, remembering decisions, catching mistakes, figuring out what to work on next.

AgentOps automates all of that. Give it a goal. It researches, plans, validates the plan, implements in parallel, validates the code, extracts what it learned, and **tells you what to run next**. Each cycle makes the next one better. You stop managing your agent and start managing your roadmap.

---

## What It Does

- **Manages context perfectly.** Research loads prior knowledge. Each worker gets fresh context. Learnings persist across sessions in `.agents/` and git.
- **Validates at every stage.** Multi-model councils judge plans before coding and code before shipping. Failures retry with failure context — no human escalation.
- **Compounds intelligence.** Post-mortem extracts what worked, what didn't, and how to improve the tools themselves. Then it suggests the next `/rpi` command. The system improves its own skills.
- **One command, six phases.** `/rpi "goal"` runs the full lifecycle hands-free. Or use any skill standalone — `/council validate this PR` works with zero setup.

Works with **Claude Code**, **Codex CLI**, **Cursor**, **Open Code** — any agent that supports [Skills](https://skills.sh). All state is local.

---

## Install

```bash
npx skills@latest add boshu2/agentops --all -g
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops   # CLI (optional)
ao hooks install        # Flywheel hooks (SessionStart + Stop)
ao hooks install --full # All 8 events with safety gates, standards, validation
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
              Boundaries: 3 always · 2 ask-first · 2 never
              Conformance: 4 verifiable assertions derived from spec
[pre-mortem]  4 judges → Verdict: PASS (incl. spec-completeness)
[crank]       Wave 1: ███ 2/2 · Wave 2: █ 1/1
[vibe]        3 judges → Verdict: PASS
[post-mortem] 3 learnings extracted → .agents/
[flywheel]    Next: /rpi "add retry backoff to rate limiter"
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

**Autonomous improvement loop:**
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

2. **`/plan`** — Decomposes the goal into issues with dependency waves. Derives three-tier boundaries (Always / Ask First / Never) to prevent scope creep, and conformance checks — verifiable assertions generated from the spec itself. Creates a beads epic.

3. **`/pre-mortem`** — 4 judges simulate failures before you write code, including a spec-completeness judge that validates plan boundaries and conformance checks. FAIL? Re-plan with feedback and try again (max 3).

4. **`/crank`** — Spawns parallel agents in waves. Each worker gets fresh context. Cross-cutting constraints from the plan are injected into every wave's validation pass. `--test-first` uses a spec-first TDD model — specs and tests before implementation in every wave. Lead validates and commits. Runs until every issue is closed.

5. **`/vibe`** — 3 judges validate the code. FAIL? Re-crank with failure context and re-vibe (max 3).

6. **`/post-mortem`** — Council validates the implementation. Retro extracts learnings. Synthesizes process improvements. **Suggests the next `/rpi` command.**

`/rpi "goal"` runs all six, end to end. Use `--interactive` if you want human gates at research and plan.

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

**Or automate the whole thing:** `/evolve` reads `GOALS.yaml`, measures fitness, picks the worst failing goal, runs a full `/rpi` cycle, re-measures, and loops — compounding improvements until all goals pass or the cycle cap is hit. Each cycle loads learnings from all prior cycles via the flywheel. You define the goals; the system does the rest.

**Session 1:** Your agent ships a feature but the tests are weak.
**Session 2:** The flywheel already knows — `/vibe` now checks test assertion coverage because last cycle's retro proposed it.
**Session 10:** Your agent catches bugs it would have missed on day one. Not because you configured anything — because the system learned.

---

## Skills

34 skills total: 24 user-facing across three tiers, plus 10 internal skills that fire automatically.

### Orchestration

| Skill | What it does |
|-------|-------------|
| `/rpi` | Goal to production — 6-phase lifecycle with self-correcting retry loops |
| `/council` | Multi-model consensus — parallel judges, consolidated verdict |
| `/crank` | Autonomous epic execution — runs waves until all issues closed (supports `--test-first` TDD) |
| `/swarm` | Parallel agents with fresh context — Codex sub-agents or Claude teams |
| `/codex-team` | Parallel Codex execution agents |
| `/evolve` | Autonomous fitness loop — measures goals, fixes worst gap, compounds via flywheel |

### Workflow

| Skill | What it does |
|-------|-------------|
| `/research` | Deep codebase exploration |
| `/plan` | Decompose goal into issues with dependency waves, boundaries, and conformance checks |
| `/implement` | Single issue, full lifecycle |
| `/vibe` | Complexity analysis + multi-model validation gate |
| `/pre-mortem` | Simulate failures before coding (4 judges incl. spec-completeness) |
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
| `/product` | Interactive PRODUCT.md generation for product-aware reviews |
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

12 hooks. All have a kill switch: `AGENTOPS_HOOKS_DISABLED=1`.

| Hook | Trigger | What it does |
|------|---------|-------------|
| Push gate | `git push` | Blocks push if `/vibe` hasn't passed |
| Pre-mortem gate | `/crank` invocation | Blocks `/crank` if `/pre-mortem` hasn't passed |
| Worker guard | `git commit` | Blocks workers from committing (lead-only) |
| Dangerous git guard | `force-push`, `reset --hard` | Blocks destructive git commands |
| Standards injector | Write/Edit | Auto-injects language-specific coding rules |
| Ratchet nudge | Any prompt | "Run /vibe before pushing" |
| Task validation | Task completed | Validates metadata before accepting |
| Session start | Session start | Knowledge injection, stale state cleanup |
| Ratchet advance | After Bash | Locks progress gates |
| Stop team guard | Session stop | Prevents premature stop with active teams |
| Precompact snapshot | Before compaction | Saves state before context compaction |
| Pending cleaner | Session start | Cleans stale pending state |

All hooks use `lib/hook-helpers.sh` for structured error recovery — failures include suggested next actions and auto-handoff context.

</details>

<details>
<summary><strong>The <code>ao</code> CLI</strong> — optional engine for the flywheel</summary>

All 34 skills work without it. The CLI adds automatic knowledge injection/extraction, ratchet gates, and session lifecycle.

```bash
ao inject              # Load prior knowledge
ao forge transcript    # Extract learnings from session
ao ratchet status      # Check progress gates
ao flywheel status     # Knowledge health metrics
ao session close       # Full lifecycle close
ao hooks install       # Flywheel hooks (SessionStart + Stop)
ao hooks install --full # All 8 events (safety gates, standards, validation)
ao hooks show          # View installed hook coverage
ao hooks test          # Verify hook configuration
```

Install: `brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops`

</details>

---

## FAQ

<details>
<summary><strong>Why not just use my coding agent directly?</strong></summary>

Your coding agent writes code. AgentOps manages everything around the code — the context, the validation, the knowledge, the intent. It's the difference between a developer and a development team with process:

- **Context management** — injects prior knowledge, gives each worker fresh context, persists learnings across sessions
- **Quality gates** — multi-model councils validate plans before coding and code before shipping
- **Self-correction** — failures retry with failure context, not human escalation
- **Self-improvement** — every cycle proposes how to make the tools better, then suggests what to run next
- **Self-enforcement** — hooks block bad pushes, enforce lead-only commits, gate `/crank` on `/pre-mortem`, nudge agents through the workflow

Without AgentOps, you are the context manager, the quality gate, and the memory. With it, you manage the roadmap.

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

<details>
<summary><strong>Built on</strong></summary>

| Project | Role |
|---------|------|
| [Ralph Wiggum pattern](https://ghuntley.com/ralph/) | Fresh context per agent — no bleed-through |
| [Multiclaude](https://github.com/dlorenc/multiclaude) | Validation gates that lock — no regression |
| [beads](https://github.com/steveyegge/beads) | Git-native issue tracking |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

</details>

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [Glossary](docs/GLOSSARY.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
