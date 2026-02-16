<div align="center">

# AgentOps

### Your coding agent gets smarter every time you use it.

[![GitHub stars](https://img.shields.io/github/stars/boshu2/agentops?style=social)](https://github.com/boshu2/agentops)
[![Version](https://img.shields.io/badge/version-2.9.0-brightgreen)](CHANGELOG.md)
[![Skills](https://img.shields.io/badge/skills-36-7c3aed)](skills/)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)

[See It Work](#see-it-work) · [Install](#install) · [The Workflow](#the-workflow) · [The Flywheel](#the-flywheel) · [Skills](#skills) · [CLI](#the-ao-cli) · [FAQ](#faq)

</div>

---

One command ships a feature end-to-end — researched, planned, validated by multiple AI models, implemented in parallel, and the system remembers what it learned for next time. The difference isn't smarter agents — it's controlling what context enters each agent's window at each phase, so every decision is made with the right information and nothing else. Every session compounds on the last. You stop managing your agent and start managing your roadmap.

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

**Search across knowledge and chat history:**
```
> ao search "mutex pattern"

Found 4 result(s) for: mutex pattern

1. .agents/learnings/2026-02-10-mutex-guard.md
   [established] Use sync.Mutex guard pattern for concurrent map access
2. .agents/sessions/2026-02-09-refactor.md
   Discussed mutex vs channel approaches for worker pool
3. .agents/patterns/concurrent-safety.md
   Pattern: always defer mutex.Unlock() immediately after Lock()

> ao search "auth retry" --cass    # session-aware, maturity-weighted

Found 3 result(s) for: auth retry

1. .agents/learnings/2026-02-12-auth-retry.md  (score: 0.92)
   [established] Exponential backoff on 401 with token refresh
2. .agents/sessions/2026-02-11-api-hardening.md  (score: 0.71)
   [candidate] Discussed retry budget per endpoint
```

**Council standalone** (no setup, no workflow):
```
> /council validate this PR
> /council brainstorm caching strategies for the API
> /council research Redis vs Memcached for our use case
```

</details>

---

## What It Does

- **Ships features end-to-end with one command.** `/rpi "goal"` runs six phases hands-free — research, plan, pre-mortem, implement, validate, post-mortem. Or use any skill standalone: `/council validate this PR` works with zero setup.
- **Catches bugs before they reach your branch.** Multi-model councils validate plans before coding (`/pre-mortem`) and code before shipping (`/vibe`). Failures retry with context — after 3 retries, the system surfaces the failure with full context for your decision.
- **Gets better the more you use it.** Post-mortem extracts what worked, what didn't, and how to improve the tools themselves. Then it suggests the next `/rpi` command. The system improves its own process.
- **Remembers everything across sessions.** Research loads prior knowledge. Each worker gets fresh context. Learnings persist in `.agents/` and git — no context resets between sessions. `ao search` finds knowledge across files and past chat history, with maturity-weighted ranking powered by [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search).
- **Enforces its own workflow.** 12 hooks across 8 lifecycle events block bad pushes, enforce lead-only commits, gate `/crank` on `/pre-mortem`, and auto-inject language-specific standards. The system doesn't just suggest good practice — it requires it.

Works with **Claude Code**, **Codex CLI**, **Cursor**, **Open Code** — any agent that supports [Skills](https://skills.sh). All state is local.

---

## Who This Is For

**Solo developers** who ship fast but skip validation — or burn hours doing it manually. AgentOps runs `/rpi` end-to-end so you get multi-model code review, retry-on-failure, and knowledge that persists across sessions. No team required.

**Tech leads scaling agent work** across a backlog. `/crank` runs parallel waves with fresh context per worker. `/status` shows what's in flight. `/post-mortem` captures what the system learned so the next cycle doesn't repeat mistakes. You manage the roadmap, not the agents.

**Quality-focused maintainers** who need high-confidence releases without manual regression hunting. `/pre-mortem` catches plan gaps before coding starts. `/vibe` validates code before push. The knowledge flywheel preserves institutional knowledge even when team members change.

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

**Where does the knowledge live?** Everything the system learns is stored in `.agents/` — a git-tracked knowledge vault in your repo. Learnings, research artifacts, plans, fitness snapshots, and session handoffs all live there. Because it's git-tracked, knowledge survives across sessions, machines, and team members. You never configure it; skills read from and write to it automatically.

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

`/rpi` runs all six phases in one Claude session. That works — but the context window fills up. Compaction kicks in, and while it's mostly fine (the real state lives in the plan and beads issues, not the conversation), you're still hoping compaction doesn't lose something important.

`ao rpi phased` solves this. Each phase runs in its own fresh Claude session. The Go CLI carries state between phases through filesystem artifacts — goal, verdicts, phase summaries — so each session starts clean with exactly the context it needs.

```
> ao rpi phased "add rate limiting to the API"

[phase 1/6] research  — spawning Claude session...done
[phase 2/6] plan      — spawning Claude session...done  (3 issues, 2 waves)
[phase 3/6] pre-mortem — spawning Claude session...done  (PASS)
[phase 4/6] crank     — spawning Claude session...done  (2 waves complete)
[phase 5/6] vibe      — spawning Claude session...done  (PASS)
[phase 6/6] post-mortem — spawning Claude session...done (3 learnings)
```

Three ways to use it:

- **Hands-free** — `ao rpi phased "goal"` runs start to finish, no prompts. Walk away.
- **Interactive** — `ao rpi phased --interactive "goal"` pauses at research and plan for your review. Step through it, approve each phase, keep full control.
- **Resume** — `ao rpi phased --from=crank "goal"` picks up from any phase. Session crashed during crank? Resume there. Want to re-run just validation? `--from=vibe`.

**Run multiple in parallel** — each run gets its own git worktree, so parallel invocations don't collide on state files or code changes:

```bash
ao rpi phased "add auth" &
ao rpi phased "fix perf" &
# Each runs in ../<repo>-rpi-<runID>/, merges back on success
```

ON by default. `--no-worktree` to opt out. On failure or Ctrl+C, the worktree is preserved for debugging.

The `/rpi` skill and `ao rpi phased` command do the same work. The difference is context control: one session vs. six fresh sessions. Use `/rpi` for small goals where context fits comfortably. Use `ao rpi phased` when the goal is big enough that you want each phase thinking clearly.

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

Learnings don't enter the knowledge base unchecked — they pass quality gates scored on specificity, actionability, novelty, context, and confidence, then land in gold/silver/bronze tiers. Good learnings get promoted over time; weak ones get demoted. Freshness decay (inspired by [MemRL](https://arxiv.org/abs/2502.06173) two-phase retrieval) ensures recent insights outweigh stale patterns automatically.

### Goal-Driven Mode: `/evolve`

Define what "done" looks like in `GOALS.yaml` — a quality contract with measurable goals and priority weights:

```yaml
goals:
  - id: test-pass-rate
    description: "All tests pass"
    check: "go test ./..."       # exit 0 = pass
    weight: 10                   # fix this first
  - id: doc-coverage
    description: "All public skills have reference docs"
    check: "test $(ls -d skills/*/references/ | wc -l) -ge 16"
    weight: 7
```

`/evolve` reads `GOALS.yaml`, measures every goal, picks the worst failing one (highest weight), runs a full `/rpi` cycle to fix it, re-measures, and loops. After each cycle, `/evolve` re-measures ALL goals — not just the target. If any goal regresses, every commit from that cycle is auto-reverted. This full regression gate is what makes "walk away" safe. Each cycle loads learnings from all prior cycles via the flywheel. When all goals pass, the system goes dormant — a valid success state, not a bug. You define the goals; the system handles the rest.

Kill switch at any time: `echo "stop" > ~/.config/evolve/KILL`

**Session 1:** Your agent ships a feature but the tests are weak.
**Session 2:** The flywheel already knows — `/vibe` now checks test assertion coverage because last cycle's retro proposed it.
**Session 10:** Your agent catches bugs it would have missed on day one. Not because you configured anything — because the system learned.

---

## The Design: Four Pillars

Every skill, goal, and hook in AgentOps maps to one of four pillars:

**Knowledge Compounding** — The system remembers. `/inject` loads prior learnings at session start. `/forge` mines transcripts at session end. `/retro` extracts what worked and what didn't. Each session is smarter than the last because the flywheel never stops turning.

**Validated Acceleration** — Speed without recklessness. `/council` spawns parallel judges for multi-model consensus. `/pre-mortem` catches plan gaps before coding. `/vibe` validates code before shipping. Failures retry with context — no human escalation needed.

**Goal-Driven Automation** — Define goals, not tasks. `/evolve` measures fitness against `GOALS.yaml` and runs `/rpi` cycles until all goals pass. `/crank` executes entire epics hands-free with wave-based parallelism. The system works toward outcomes, not checklists.

**Zero-Friction Workflow** — Start in 60 seconds. `/quickstart` runs a guided cycle on your actual codebase. `/implement` picks up a single issue end-to-end. `/handoff` preserves context across sessions. `/status` shows where you are and what to do next. No configuration required.

These pillars are codified in [`GOALS.yaml`](GOALS.yaml) — 47 measurable goals that define what "healthy" means for the system. `/evolve` measures them all.

---

## Skills

36 skills total: 26 user-facing across three tiers, plus 10 internal skills that fire automatically.

### Orchestration

| Skill | What it does |
|-------|-------------|
| `/rpi` | Goal to production — 6-phase lifecycle with self-correcting retry loops |
| `/council` | Multi-model consensus — parallel judges, consolidated verdict |
| `/crank` | Hands-free epic execution — runs waves until all issues closed (supports `--test-first` TDD) |
| `/swarm` | Parallel agents with fresh context — Codex sub-agents or Claude teams |
| `/codex-team` | Parallel Codex execution agents |
| `/evolve` | Goal-driven fitness loop — measures goals, fixes worst gap, compounds via flywheel |

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
| `/product` | Generate `PRODUCT.md` — unlocks product-aware judges in `/pre-mortem` and `/vibe` automatically |
| `/trace` | Trace design decisions through history |
| `/handoff` | Structured session handoff |
| `/inbox` | Agent Mail monitoring |
| `/recover` | Post-compaction context recovery |
| `/update` | Reinstall all AgentOps skills globally |

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

Agent output quality is determined by context input quality. Every pattern below — fresh context per worker, ratcheted progress, least-privilege loading — exists to ensure the right information is in the right window at the right time.

Parallel agents produce noisy output; councils filter it; ratchets lock progress so it can never regress.

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

Every wave gets a fresh worker set with clean context — no bleed-through between waves.

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
<summary><strong>Context Windowing</strong> — bounded execution for large codebases</summary>

For repos over ~1500 files, `/rpi` uses deterministic shards to keep each worker's context window bounded. Run `scripts/rpi/context-window-contract.sh` before `/rpi` to enable sharding. This prevents context overflow and keeps worker quality consistent regardless of codebase size.

</details>

---

## The `ao` CLI

Optional but recommended. The CLI is plumbing — skills and hooks call it automatically. You install it, your agent uses it. You don't type `ao` commands yourself (with three exceptions below).

**How it works:** 12 hooks fire `ao` commands at session lifecycle boundaries (start, stop, tool use, compaction). Skills call `ao` commands internally during `/rpi`, `/post-mortem`, `/status`, `/flywheel`, and other workflows. Every `ao` command is wired to at least one automated caller.

**What it enables:**

- **Search** — `ao search` finds knowledge across files and past chat history, with [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search)-powered maturity-weighted ranking. For full session search across Claude Code, Cursor, and other agents, use [`cass`](https://github.com/Dicklesworthstone/coding_agent_session_search) directly.
- **Knowledge curation** — Learnings flow through quality pools (`ao pool`), human review gates (`ao gate`), and maturity transitions (`ao maturity`). [MemRL](https://arxiv.org/abs/2502.06173)-inspired reward signals (`ao feedback`) update ranking automatically via hooks.
- **Forge→temper→store pipeline** — `ao forge` extracts knowledge from transcripts at session end, `ao temper` validates and locks it during `/post-mortem`, `ao store` indexes it for retrieval. Fully automated.
- **Feedback loop** — `ao feedback-loop` closes the MemRL reward cycle, `ao session-outcome` records composite reward signals, `ao task-feedback` applies outcomes to associated learnings. All fire automatically at session end.
- **Provenance** — `ao trace` follows any artifact back to the session transcript that created it.
- **Ratchet gates** — `ao ratchet` tracks RPI workflow progress and prevents regression. Once a phase passes, it stays passed.
- **Plans** — `ao plans` maintains a registry connecting research artifacts to [beads](https://github.com/steveyegge/beads) issues, with drift detection.
- **Metrics** — `ao metrics`, `ao badge`, and `ao task-status` give quantitative flywheel health. `/status` renders them into a dashboard.

<details>
<summary><strong>Automation map</strong> — which skills/hooks call which commands</summary>

| ao command | Called by |
|------------|----------|
| `inject`, `extract`, `ratchet status`, `maturity --scan` | SessionStart hooks |
| `forge transcript`, `session-outcome`, `feedback-loop`, `task-sync`, `batch-feedback` | SessionEnd hooks |
| `flywheel close-loop` | Stop hook |
| `ratchet record` | `/rpi` (each phase), ratchet-advance hook |
| `forge index`, `feedback-loop`, `session-outcome`, `temper validate`, `task-feedback` | `/post-mortem`, `/retro` |
| `badge`, `task-status`, `flywheel status`, `ratchet status` | `/status` |
| `maturity --scan`, `promote-anti-patterns`, `badge`, `forge status` | `/flywheel` |
| `search`, `pool`, `plans` | `/research`, `/knowledge`, `/plan` |

</details>

**Install:**
```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops
ao hooks install       # Flywheel hooks (SessionStart + Stop)
ao hooks install --full # All 12 hooks across 8 lifecycle events
```

**The three commands you'll actually type:**
```bash
ao rpi phased "goal"   # Full RPI lifecycle, fresh context per phase (see Phased RPI above)
ao search "query"      # Search knowledge (also: --cass, --use-sc, --type)
ao demo                # Interactive demo of capabilities
```

Everything else runs automatically. `ao quick-start` and `ao export-constraints` are the only other human-initiated commands (setup and debugging).

73 commands total. Full reference: [CLI Commands](cli/docs/COMMANDS.md)

---

## FAQ

<details>
<summary><strong>Why not just use my coding agent directly?</strong></summary>

Without AgentOps, every session starts from scratch. Your agent doesn't remember what failed last time, doesn't validate its plan before coding, doesn't check its code with a second opinion, and doesn't capture what it learned. You fill those gaps manually — re-explaining context, reviewing code, tracking what changed. With AgentOps, the system handles context, validation, and memory. You manage the roadmap.

</details>

<details>
<summary><strong>How does this compare to other approaches?</strong></summary>

| Approach | What it does well | What AgentOps adds |
|----------|------------------|--------------------|
| **Direct agent use** (Claude Code, Cursor, Copilot) | Full autonomy, simple to start | Multi-model councils, fresh-context waves, and knowledge that compounds across sessions. A bare agent starts fresh each session; ours extracts learnings and applies them next time. |
| **Custom prompts** (.cursorrules, CLAUDE.md) | Flexible, version-controlled | Static instructions don't compound. The flywheel auto-extracts learnings and injects them back. `/post-mortem` proposes changes to the tools themselves. |
| **Agent orchestrators** (CrewAI, AutoGen, LangGraph) | Multi-language task scheduling | Those choreograph sequential tasks; we compose parallel waves with validation at every stage. No external state backend — all learnings are git-tracked. |
| **CI/CD gates** (GitHub Actions, pre-commit) | Automated, industry standard | Gates run after code is written. Ours run before coding (`/pre-mortem`) and before push (`/vibe`). Failures retry with context, not human escalation. |

</details>

<details>
<summary><strong>What data leaves my machine?</strong></summary>

AgentOps itself stores nothing externally — all state lives in `.agents/` (git-tracked, local). No telemetry, no cloud, no external services. Your coding agent's normal API traffic to its LLM provider still applies.

</details>

<details>
<summary><strong>Can I use this with other AI coding tools?</strong></summary>

Yes — Claude Code, Codex CLI, Cursor, Open Code, anything supporting [Skills](https://skills.sh). The `--mixed` council mode adds Codex judges alongside Claude. Knowledge artifacts are plain markdown.

</details>

<details>
<summary><strong>What does PRODUCT.md do?</strong></summary>

Run `/product` to generate a `PRODUCT.md` describing your mission, personas, and competitive landscape. Once it exists, `/pre-mortem` automatically adds product perspectives (user-value, adoption-barriers) and `/vibe` adds developer-experience perspectives (api-clarity, error-experience) to their council reviews. Your agent understands what matters to your product — not just whether the code compiles.

</details>

<details>
<summary><strong>What are the current limitations?</strong></summary>

- **Single primary author so far.** The system works but hasn't been stress-tested across diverse codebases and team sizes. Looking for early adopters willing to break things.
- **Quality pool can over-promote.** Context-specific patterns sometimes get promoted as general knowledge. Freshness decay helps but doesn't fully solve stale injection.
- **Retry loops cap at 3.** If a council or crank wave fails three times, the system surfaces the failure to you rather than looping forever. This is intentional but means some edge cases need human judgment.
- **Knowledge curation is imperfect.** Freshness decay prevents the worst staleness, but the scoring heuristics (specificity, actionability, novelty) are tuned for one author's workflow. Your mileage may vary.

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
| [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search) | Unified search across coding agent chat histories |
| [MemRL](https://arxiv.org/abs/2502.06173) | Two-phase retrieval for cross-session memory |

</details>

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). If AgentOps helped you ship something, post in [Discussions](https://github.com/boshu2/agentops?tab=discussions).

## License

Apache-2.0 · [Docs](docs/INDEX.md) · [Glossary](docs/GLOSSARY.md) · [Architecture](docs/ARCHITECTURE.md) · [CLI Reference](cli/docs/COMMANDS.md) · [Changelog](CHANGELOG.md)
