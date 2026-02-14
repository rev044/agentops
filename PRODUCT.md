---
last_reviewed: 2026-02-13
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
- **Zero Setup, Zero Telemetry** — All state lives in git-tracked `.agents/` directories with no cloud dependency, giving teams full control and auditability while working across any coding agent runtime.

## Competitive Landscape

| Alternative | Strength | Our Differentiation |
|-------------|----------|---------------------|
| Direct Agent Use (Claude Code, Cursor, Copilot) | Full autonomy; no overhead; simple to start | Adds multi-model councils, fresh-context waves, and knowledge compounding. A bare agent writes code once; ours extracts learnings and applies them next session. |
| Custom Prompt Engineering (.cursorrules, CLAUDE.md) | Flexible, version-controlled, lightweight | Static instructions don't compound. Our flywheel auto-extracts learnings and injects them back. `/post-mortem` proposes changes to the tools themselves. |
| Agent Orchestrators (CrewAI, AutoGen, LangGraph) | Mature multi-language task scheduling | Those route work between agents but don't manage what's in each agent's context window. We treat context quality as the primary lever — fresh-context isolation per worker, phase-specific loading, knowledge quality gates. No external state backend — all learnings are git-tracked. |
| CI/CD Quality Gates (GitHub Actions, pre-commit) | Automated, enforced, industry standard | Gates run after code is written. Ours run before coding (`/pre-mortem`) and before push (`/vibe`). Failures retry with context, not human escalation. |

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
