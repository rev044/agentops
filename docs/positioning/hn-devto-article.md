# HN / dev.to Article Draft

## Working Title

Coding agents don't do their own bookkeeping

## Alternate Titles

- AgentOps: the operational layer for coding agents
- The problem with coding agents is not the model, it is the missing operational layer
- Coding agents need bookkeeping, validation, and flows

## Draft

Most discussion about coding agents still focuses on the wrong layer.

People ask:

- which model is best
- how many agents to run
- whether the workflow should be one-shot or multi-agent
- which prompt pattern gets the best output

Those questions matter, but they are not the core problem.

The core problem is that coding agents do not do their own bookkeeping.

They do not naturally preserve what mattered from prior work. They do not reliably challenge their own plans before implementation. They do not turn completed work into better next work unless someone builds a system around them that does exactly that.

That is the idea behind AgentOps.

The simplest way to describe it now is:

**AgentOps is the operational layer for coding agents.**

It gives coding agents four things they do not have by default:

1. **Bookkeeping**
   Sessions should not leave behind only chat history. They should leave behind reusable learnings, findings, and context that can be surfaced back later.
2. **Validation**
   Plans and code should be challenged before shipping, not only after the breakage shows up in CI or production.
3. **Primitives**
   Operators need reusable building blocks, not one giant all-or-nothing workflow.
4. **Flows**
   Those primitives need to connect into repeatable paths for discovery, implementation, validation, and knowledge extraction.

The reason this matters is compounding.

Without an operational layer, every session starts too close to zero. You re-brief the agent. You re-explain the repo. You rediscover old mistakes. You lose the result of yesterday's debugging the moment the context window moves on.

With an operational layer, the environment changes.

Research is written down. Findings are promoted. Learnings get retrieved. Validation gates get sharper. The next session starts inside a stronger operating context than the previous one had.

That is the product thesis.

Technically, I think the best internal description is that AgentOps acts like a **context compiler**.

A normal compiler takes raw source code and transforms it into an executable artifact with stronger guarantees than the original text had on its own.

AgentOps does something similar for agent work:

- raw transcripts, retros, and failures become learnings
- learnings become curated patterns and findings
- findings become rules, checks, and better planning context
- the next task starts with compiled context instead of vague memory

That framing is useful internally because it explains why the product is not just "memory" and not just "workflow tooling."

Publicly, though, I think the more useful sentence is still:

**coding agents don't do their own bookkeeping**

That line has become more convincing to me because I keep seeing the same convergence from different directions.

Karpathy has talked about the bookkeeping burden around an LLM wiki.

People at Block have described the moat as the feedback loop around the markdown file and the signal flowing through it.

And building AgentOps in production kept teaching the same lesson over and over:

the model is not the product.

The system around the model is the product.

That system needs:

- bookkeeping so knowledge does not disappear
- validation so errors get caught before shipping
- reusable primitives so operators can control the work
- connected flows so the whole thing compounds

If you already use coding agents every day, that is probably the pain you feel too.

Not "the model is dumb."

More like:

- "why did it forget the decision we made two days ago?"
- "why did it repeat the exact failed fix?"
- "why did it say this was done without actually validating it?"
- "why does every new session feel like I am onboarding a contractor with amnesia?"

That is the problem I think the category needs to solve.

Not more demos of autonomous generation.

Not bigger swarms for their own sake.

An operational layer.

That is what I am building toward with AgentOps.

## HN Blurb

AgentOps is an open-source operational layer for coding agents. The core thesis is that coding agents do not do their own bookkeeping, so work resets every session unless you build bookkeeping, validation, primitives, and flows around them. This draft is the longer version of that argument.

## Publish Checklist

- tighten opening paragraph for final voice
- add one concrete repo screenshot, issue, or nightly dream-cycle artifact
- publish to dev.to first, then adapt a shorter HN intro
