# 2026-04-09 Operating Review

## Current Product Story

**Public category:** AgentOps is the operational layer for coding agents.

**Public value:** It gives agents bookkeeping, validation, primitives, and flows so every session starts where the last one left off.

**Technical frame:** AgentOps is a context compiler. Raw session signal becomes reusable knowledge, compiled prevention, and better next work.

**Internal proof model:** The older three-gap contract still matters, but it should stay behind the public story. Use it to prove the product works, not to lead the category pitch.

## Audit Summary

| Surface | Status | Note |
|---------|--------|------|
| `README.md` | green | Product-first rewrite shipped. Install, proof, and operating model are now in the right order. |
| `skills/quickstart` + `skills-codex/quickstart` | green | Onboarding now matches the public category instead of the older memory/judgment/workflow pitch. |
| `skills/using-agentops` + `skills-codex/using-agentops` | green | Startup meta-skill now explains the operating model before the RPI flow. |
| `PRODUCT.md` | yellow | Top-level framing now aligns, but parts of the document still use the older three-gap proof language deeper down. |
| `GOALS.md` | yellow | Top-level framing and dream-cycle progress are corrected, but the goal system still uses the older technical contract by design. |
| `docs/newcomer-guide.md` | green | Intro now matches the product category. |
| `docs/software-factory.md`, `docs/context-lifecycle.md`, `docs/agentops-brief.md`, `docs/positioning/*` | red | These still leak the old public story and need one coordinated rewrite pass. |

## Executive Readout

### What is working

- The repo now has a credible public category: `operational layer for coding agents`.
- The technical story is strong and differentiated: `context compiler` + bookkeeping moat.
- The backlog already contains the right moat work: dream cycle, pattern-to-skill, Codex hooks default, runtime proof, and comparison freshness.

### What is still blocking GTM leverage

- The public story is not yet fully uniform across onboarding and positioning docs.
- The moat is described more clearly than it is demonstrated. Dream cycle automation and pattern-to-skill are still partially aspirational.
- The growth sprint has content tasks, but they need to be ordered around proof assets, not just prose.

## Owner Map

### CEO / GTM

- `na-gtm.3` — Pin convergence thesis Discussion
- `na-gtm.4` — Write and post HN/dev.to article
- `na-gtm.5` — Record 2-minute demo video
- `na-gtm.13` — Built with AgentOps gallery

### Product / Onboarding

- `na-gtm.6` — First-5-minutes auto-detection
- `na-b0j` — Unify legacy positioning docs with operational-layer framing

### Platform / Runtime

- `na-gtm.9` — Codex native hooks release
- `na-gtm.11` — Runtime smoke tests to CI
- `na-gtm.8` — Quarantine triage: promote or delete

### Knowledge Moat

- `na-gtm.7` — Dream cycle GitHub Action
- `na-gtm.10` — Pattern-to-skill pipeline

### Release

- `na-gtm.15` — v2.36 release

## Executive Scorecard

| Function | Question | Owning beads | Why it matters |
|----------|----------|--------------|----------------|
| Narrative discipline | Does every public surface sell the same category? | `na-b0j`, `na-gtm.6` | Category confusion kills conversion before product quality can matter. |
| Activation | Can a new operator reach first value in one session? | `na-gtm.6`, `na-gtm.9`, `na-gtm.11` | Sharp onboarding is the bridge from curiosity to retained use. |
| Proof | Can users see the flywheel compound without trusting a claim? | `na-gtm.5`, `na-gtm.7`, `na-gtm.10` | The moat has to be demonstrated, not narrated. |
| Distribution | Are we shipping artifacts that convert discovery into stars and installs? | `na-gtm.3`, `na-gtm.4`, `na-gtm.13` | Distribution only works when the story and proof are already coherent. |
| Release | Are we bundling the proof into a moment people can react to? | `na-gtm.15` | A release is the forcing function that turns scattered wins into momentum. |

## Recommended Sequence

1. Finish narrative unification so every onboarding and positioning surface says the same thing.
2. Ship the proof asset stack: demo video, Codex hooks default, and the new README together.
3. Make the dream cycle visible, not just real. Users need to see the flywheel run between sessions.
4. Ship pattern-to-skill to turn the moat from "good thesis" into "hard-to-copy behavior."
5. Bundle the above into `v2.36` and drive distribution through the article, Discussion, and comparison pages.

## CEO Call

If the goal is "unicorn status," the repo does not need more breadth right now. It needs tighter narrative discipline and more visible proof of the moat. The winning sequence is:

1. Category clarity
2. Activation
3. Proof
4. Distribution
5. Moat deepening

AgentOps is closest to winning when the public story is simple, the onboarding is sharp, and the repo visibly compounds better than competitors between sessions.
