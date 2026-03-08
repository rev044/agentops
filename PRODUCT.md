---
last_reviewed: 2026-02-25
---

# PRODUCT.md

## Mission

AgentOps treats context quality as the primary lever for agent output quality — orchestrating what information enters each agent's window at each phase so every decision is made with the right context and nothing else, then compounding those results through a knowledge flywheel that makes each successive session smarter.

## Vision

Make coding agents feel like a real engineering organization: validated work, institutional memory, and continuous improvement by default.

AgentOps is the local-first DevOps layer around your coding agent. Execution happens in sessions, but knowledge persists across sessions: plans, gates, outcomes, and learnings get captured and fed forward so the same codebase becomes faster and safer to change over time.

## Target Personas

### Persona 1: The Solo Developer
- **Goal:** Ship features faster while maintaining code quality — without manual code review or multi-person coordination overhead.
- **Pain point:** High-velocity solo development means either skipping validation (shipping broken code) or spending hours on manual testing and review prep. Each session starts from scratch with no memory of what worked before.

### Persona 2: The Scaling Tech Lead
- **Goal:** Keep a large backlog moving predictably while preventing agents from shipping conflicting changes or breaking shared systems. Need end-to-end visibility into what's being worked on, why, and what the system learned.
- **Pain point:** Managing parallel work across agents creates cascading blockers — specs change mid-cycle, cross-cutting constraints get violated, learnings from failed attempts aren't captured for next time. Manual ticket grooming and post-mortems burn cycles.

### Persona 3: The Quality-First Maintainer
- **Goal:** Ship fewer but higher-confidence releases. Prevent regressions in critical code paths. Maintain institutional knowledge even when team members change.
- **Pain point:** Every regression requires debugging, patching, coordinating hotfixes. Test coverage stalls because writing tests is slower than writing features. Design decisions get lost in commit messages. New agents repeat mistakes because knowledge isn't captured.

## Core Value Propositions

- **Compound Intelligence Across Sessions** — Each session captures learnings that pass quality gates (scored on specificity, actionability, novelty, context, and confidence) into gold/silver/bronze tiers. Freshness decay ensures recent insights outweigh stale patterns. The system doesn't just remember — it curates.
- **Ship With Confidence, Not Caution** — Least-privilege context loading gives each agent only the information relevant to its task, preventing context contamination. Parallel model validation and self-correcting workflows catch issues before deployment, letting teams move faster without sacrificing quality.
- **Hands-Free Goal Achievement** — Spawn agents that work independently toward your goals (via `/evolve` and `/crank`), validate their work through multi-model consensus, and commit only when passing quality gates. `/evolve` is built for overnight runs — cycle state is disk-based (survives context compaction), regression gates are hard-gated (no snapshot = stop), and every cycle writes a verifiable audit trail. Validated in production: 116 cycles ran ~7 hours unattended, delivering test coverage from ~85% to ~97%, zero high-complexity functions, and modern idiomatic Go across 203 files.
- **Compose What You Need** — Skills are standalone building blocks you compose however you want. Use one (`/council validate this PR`), chain several (`/plan` → `/pre-mortem` → `/crank`), or run the full pipeline (`/rpi`). The same recursive shape — lead decomposes work, workers execute atomically, validation gates lock progress — repeats at every scale from a single `/implement` to a full `/evolve` run.
- **Multi-Runtime, Multi-Model** — Works across Claude Code, Codex CLI, Cursor, and OpenCode. Skills are portable across runtimes (`/converter` exports to native formats). Codex-native skill format ships alongside Claude-native. Independent judges (Claude + Codex) debate before code ships — not advisory, validation gates block merges until they pass.
- **Progressive Approachability** — New users see 11 starter skills (`/quickstart`, `/research`, `/council`, `/vibe`, `/rpi`, `/implement`, `/retro --quick`, `/status`, `/goals`, `/flywheel`, `/inbox`), not 53. Plain-English verb aliases let you type what you mean — "review this code" triggers `/vibe`, "execute this epic" triggers `/crank`. The 18 advanced skills (planning, orchestration, validation) and 20 expert skills (cross-vendor, PR workflows, traceability) reveal themselves as you grow.
- **Zero Setup, Zero Telemetry** — All state lives in git-tracked `.agents/` directories with no cloud dependency, giving teams full control and auditability. 52 skills, 3 hooks, and the knowledge flywheel work independently with no external daemon.

## Design Principles

**Theoretical foundation — four pillars:**

1. **[Systems theory (Meadows)](https://en.wikipedia.org/wiki/Twelve_leverage_points)** — Target the high-leverage end of the hierarchy: information flows (#6), rules (#5), self-organization (#4), goals (#3). Changing the loop beats tuning the output.
2. **[DevOps (Three Ways)](docs/the-science.md#part-3-devops-foundation-the-three-ways)** — Flow, feedback, continual learning — applied to the agent loop instead of the deploy pipeline.
3. **[Brownian Ratchet](docs/brownian-ratchet.md)** — Embrace agent variance, filter aggressively, ratchet successes. Chaos + filter + one-way gate = net forward progress.
4. **[Knowledge Flywheel (escape velocity)](docs/the-science.md#the-escape-velocity-condition)** — If retrieval rate × usage rate exceeds decay rate (σ×ρ > δ), knowledge compounds. If not, it decays to zero. The flywheel exists to stay above that threshold.

**Operational principles:**

1. **Context quality determines output quality.** Every skill, hook, and flywheel component exists to ensure the right context is in the right window at the right time.
2. **Least-privilege loading.** Agents receive only the context necessary for their task — phase-specific, role-scoped, freshness-weighted.
3. **The cycle is the product.** No single skill is the value. The compounding loop — research, plan, validate, build, validate, learn, repeat — is what makes the system improve.
4. **Two-tier execution.** Orchestrators (`/evolve`, `/rpi`, `/crank`) stay in the main session so you see progress and can intervene. Workers they spawn fork into subagents where results merge back via the filesystem — never accumulated chat context.
5. **Dormancy is last resort.** When goals pass and the current backlog is empty, the system keeps generating productive work from tests, validation gaps, bug hunts, drift, and feature suggestions before it finally goes dormant.

## Usage

This file enables product-aware council reviews:

- **`/pre-mortem`** — Automatically includes `product` perspectives (user-value, adoption-barriers, competitive-position) alongside plan-review judges when this file exists.
- **`/vibe`** — Automatically includes `developer-experience` perspectives (api-clarity, error-experience, discoverability) alongside code-review judges when this file exists.
- **`/council --preset=product`** — Run product review on demand.
- **`/council --preset=developer-experience`** — Run DX review on demand.

Explicit `--preset` overrides from the user skip auto-include (user intent takes precedence).

## See Also

- [Scale Without Swarms](docs/scale-without-swarms.md) — why 3-5 focused agents with fresh context and regression gates outperform massive uncoordinated swarms; the AgentOps model of waves, isolation, and gates explained.
