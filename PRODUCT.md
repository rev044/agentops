---
last_reviewed: 2026-02-16
---

# PRODUCT.md

## Mission

AgentOps is a skills plugin that treats context quality as the primary lever for agent output quality — orchestrating what information enters each agent's window at each phase so every decision is made with the right context and nothing else, then compounding those results through a knowledge flywheel that makes each successive context window smarter.

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
- **Hands-Free Goal Achievement** — Spawn agents that work independently toward your goals (via `/evolve` and `/crank`), validate their work through multi-model consensus, and commit only when passing quality gates.
- **Orchestration at Scale** — Big visions decompose into dependency-mapped waves of parallel workers, each running in fresh-context isolation. No subagent nesting required — workers are flat peers, not nested children. `/plan` creates the wave structure from dependencies; `/crank` executes it. The system handles coordination so you manage the roadmap, not the agents.
- **Progressive Approachability** — New users see 5 starter skills (`/quickstart`, `/council`, `/research`, `/vibe`, `/rpi`), not 27. Plain-English verb aliases let you type what you mean — "review this code" triggers `/vibe`, "execute this epic" triggers `/crank`. The `/learn` skill captures knowledge into the flywheel in one command. Power and expert tiers reveal themselves as you grow.
- **Zero Setup, Zero Telemetry** — All state lives in git-tracked `.agents/` directories with no cloud dependency, giving teams full control and auditability while working across any coding agent runtime.

## Competitive Landscape

| Alternative | Strength | Our Differentiation |
|-------------|----------|---------------------|
| Direct Agent Use (Claude Code, Cursor, Copilot) | Full autonomy; no overhead; simple to start | Adds multi-model councils, fresh-context waves, and knowledge compounding. A bare agent writes code once; ours extracts learnings and applies them next session. |
| Custom Prompt Engineering (.cursorrules, CLAUDE.md) | Flexible, version-controlled, lightweight | Static instructions don't compound. Our flywheel auto-extracts learnings and injects them back. `/post-mortem` proposes changes to the tools themselves. Progressive skill discovery shows 5 commands on day 1, not 27. |
| Agent Orchestrators (CrewAI, AutoGen, LangGraph) | Mature multi-language task scheduling | Those route work between agents but don't manage what's in each agent's context window. We treat context quality as the primary lever — fresh-context isolation per worker, phase-specific loading, knowledge quality gates. No external state backend — all learnings are git-tracked. |
| CI/CD Quality Gates (GitHub Actions, pre-commit) | Automated, enforced, industry standard | Gates run after code is written. Ours run before coding (`/pre-mortem`) and before push (`/vibe`). Failures retry with context, not human escalation. |
| Olympus (Mount Olympus) | Persistent daemon, context provenance, constraint injection (learnings become `*_test.go`), run ledger tracking every attempt | Complementary, not competing. AgentOps is the autonomous engine (skills, hooks, flywheel). Olympus is the power-user daemon layer that composes AgentOps for fully cross-session automation — nobody types `/rpi`, daemon polls and spawns. We stand alone; Olympus builds on top. |

## Relationship to Olympus

AgentOps is a complete, standalone system — autonomous within a session. You do not need Olympus to use AgentOps. The 39 skills, 12 hooks, knowledge flywheel, and RPI lifecycle work independently — `/rpi` ships features end-to-end, `/evolve` runs goal-driven improvement loops, and the flywheel compounds knowledge across sessions, all without any external daemon.

**Olympus is the power-user layer for people who want to go further.**

For users who've mastered AgentOps and want fully autonomous cross-session execution — no human types `/rpi`, no human opens Claude Code — Olympus provides:

- **A persistent daemon** that polls for ready work, spawns agent sessions, and monitors context saturation without human intervention
- **Context bundles with provenance** — hashable, diffable context assemblies that track exactly which learnings, specs, and prior failures were injected into each attempt
- **A run ledger** — append-only evidence of every execution attempt, what context produced what result, feeding failure context into the next spawn automatically
- **Constraint injection** — learnings compile to `*_test.go` files that fail the build, not markdown that might be ignored

AgentOps is where you learn to be a context engineer. Olympus is what you build when you've mastered it and want the machine to run without you.

**Repo:** [github.com/boshu2/olympus](https://github.com/boshu2/olympus)

## Design Principles

1. **Context quality determines output quality.** Every skill, hook, and flywheel component exists to ensure the right context is in the right window at the right time.
2. **Least-privilege loading.** Agents receive only the context necessary for their task — phase-specific, role-scoped, freshness-weighted.
3. **The cycle is the product.** No single skill is the value. The compounding loop — research, plan, validate, build, validate, learn, repeat — is what makes the system improve.
4. **Dormancy is success.** When all goals pass and no work remains, the system stops. It does not manufacture work to justify its existence.

## Usage

This file enables product-aware council reviews:

- **`/pre-mortem`** — Automatically includes `product` perspectives (user-value, adoption-barriers, competitive-position) alongside plan-review judges when this file exists.
- **`/vibe`** — Automatically includes `developer-experience` perspectives (api-clarity, error-experience, discoverability) alongside code-review judges when this file exists.
- **`/council --preset=product`** — Run product review on demand.
- **`/council --preset=developer-experience`** — Run DX review on demand.

Explicit `--preset` overrides from the user skip auto-include (user intent takes precedence).
