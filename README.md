# AgentOps

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**Your AI agent has amnesia. Let's fix that.**

> Other tools make Claude faster. AgentOps makes Claude *smarter*.

---

## The Problem Everyone Ignores

Your agent solves a bug today. Tomorrow? Same bug, starts from scratch.

You explain your architecture once. Next week? Explain it again.

**Every session is day one.**

| Approach | Promise | Reality |
|----------|---------|---------|
| Spec-driven dev | "Perfect spec → perfect code" | Specs are discovered during implementation |
| Better prompts | "Just prompt engineer harder" | Knowledge dies when the session ends |
| RAG/embeddings | "Search your codebase" | No learning, just lookup |
| Other workflows | "Follow these 12 steps" | Linear, no memory, no compounding |

**The dirty secret:** None of these remember what they learned.

---

## The Solution: Compounding Knowledge

AgentOps captures what your agent learns and feeds it back. Every session makes the next one smarter.

```
Session 1        Session 10       Session 100
+-----------+    +-----------+    +-----------+
|  Debug    |    |  "I've    |    |  Domain   |
|  auth bug |    |  seen     |    |  expert   |
|  (2 hrs)  | -> |  this"    | -> |  instant  |
|           |    |  (10 min) |    |  recall   |
+-----------+    +-----------+    +-----------+

    Same effort        10x faster       100x faster
```

**The math:** Knowledge decays at ~17%/week without reinforcement.
AgentOps retrieval × usage > decay rate = **compounding, not forgetting**.

---

## Quick Start

```bash
# 1. Install CLI
brew tap boshu2/agentops https://github.com/boshu2/agentops
brew install agentops

# 2. Install Plugin
claude plugin marketplace add boshu2/agentops
claude plugin install agentops

# 3. Initialize (in your project)
ao init && ao hooks install
```

**That's it.** Hooks capture knowledge automatically. Every session feeds the next.

---

## How It Works

### The Brownian Ratchet

In physics, a Brownian ratchet extracts useful work from random motion. Molecules bounce chaotically, but the ratchet only allows forward movement.

**AgentOps applies this to AI agents:**

```
+------------------------------------------------------------------+
|                    THE BROWNIAN RATCHET                          |
|                                                                  |
|   CHAOS              FILTER              RATCHET                 |
|   (explore)          (validate)          (lock progress)         |
|                                                                  |
|   +---------+        +---------+        +---------+              |
|   | Research| -----> |  Vibe   | -----> | Commit  |              |
|   | & Try   |        |  Check  |        | & Learn |              |
|   +---------+        +----+----+        +---------+              |
|        ^                  |                  |                   |
|        |                  | (fail)           |                   |
|        +------------------+                  |                   |
|                                              v                   |
|                                       +-----------+              |
|                                       | .agents/  |              |
|                                       | (memory)  |              |
|                                       +-----+-----+              |
|                                             |                    |
|                                             | (next session)     |
|                                             v                    |
|                                       [ao inject]                |
+------------------------------------------------------------------+
```

| Phase | What Happens | You Type |
|-------|--------------|----------|
| **Research** | Mine prior knowledge, explore codebase | `/research` |
| **Plan** | Break into tracked issues | `/plan` |
| **Pre-mortem** | Simulate failures before they happen | `/pre-mortem` |
| **Crank** | Implement → validate → commit loop | `/crank` |
| **Post-mortem** | Extract learnings, lock the ratchet | `/post-mortem` |

---

## What Gets Captured

Everything lives in `.agents/` — git-tracked, portable, yours.

```
.agents/
├── learnings/     # "Auth bugs usually come from token refresh"
├── patterns/      # "Here's how we handle retries in this codebase"
├── research/      # Deep dive outputs
├── specs/         # Validated specifications
└── retros/        # Session retrospectives
```

**Hooks run automatically:**
- **SessionStart** → Injects relevant prior knowledge
- **SessionEnd** → Extracts and indexes learnings

You don't run `ao` commands manually. The flywheel turns itself.

---

## The Compound Effect

```
WITHOUT AGENTOPS
================
Session 1     Session 2     Session 3     Session 4
[2 hours] --> [2 hours] --> [2 hours] --> [2 hours]  = 8 hours
  (same problem, every time)

WITH AGENTOPS
=============
Session 1     Session 2     Session 3     Session 4
[2 hours] --> [10 min]  --> [2 min]   --> [instant]  = ~2.2 hours
  (learn)      (recall)     (refine)     (mastered)
```

**By session 100, your agent knows:**
- Every bug you've ever fixed
- Your architecture decisions and why
- Your team's coding patterns
- What approaches failed and why

---

## Skills Reference

| Skill | Trigger | What It Does |
|-------|---------|--------------|
| `/research` | "explore", "investigate" | Deep codebase exploration with knowledge mining |
| `/plan` | "create a plan" | Convert goals into tracked issues |
| `/pre-mortem` | "what could go wrong" | Find failure modes before implementation |
| `/crank` | "ship it", "execute" | Autonomous implement → validate → commit loop |
| `/vibe` | "validate", "check" | Multi-aspect code validation |
| `/post-mortem` | "what did we learn" | Extract and index learnings |

---

## CLI Reference

| Command | Purpose |
|---------|---------|
| `ao init` | Initialize AgentOps in a repo |
| `ao hooks install` | Install SessionStart/End hooks |
| `ao inject` | Manually inject knowledge |
| `ao forge search` | Search past sessions |
| `ao forge index` | Index artifacts |
| `ao feedback` | Mark learnings as helpful/harmful |

---

## The Science

Built on peer-reviewed research, not vibes.

| Concept | Source | Finding |
|---------|--------|---------|
| Knowledge Decay | Darr, Argote & Epple (1995) | Org knowledge depreciates ~17%/week |
| Memory Reinforcement | Ebbinghaus (1885) | Retrieval strengthens memory |
| MemRL | Zhang et al. (2025) | Two-phase retrieval enables self-evolving agents |

---

## Built On

| Tool | Author | What We Use |
|------|--------|-------------|
| [beads](https://github.com/steveyegge/beads) | Steve Yegge | Git-native issue tracking |
| [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search) | Dicklesworthstone | Session indexing |
| [multiclaude](https://github.com/dlorenc/multiclaude) | dlorenc | Brownian Ratchet pattern |

---

## License

MIT

---

<p align="center">
  <strong>Stop starting from zero.</strong><br>
  <em>Your agent's knowledge should compound, not reset.</em>
</p>
