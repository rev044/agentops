# AgentOps

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**Your AI agent has amnesia. Let's fix that.**

> Other tools make Claude faster. AgentOps makes Claude *smarter*.

---

## The Workflow: 5 Commands

This is all you do. Everything else is automatic.

```
/research → /pre-mortem → /crank → /vibe → /post-mortem
```

| Step | Command | What Happens |
|------|---------|--------------|
| 1 | `/research` | Explore codebase, mine prior knowledge |
| 2 | `/pre-mortem` | Simulate failures before you build |
| 3 | `/crank` | Implement → validate → commit (loops until done) |
| 4 | `/vibe` | Validate: security, quality, architecture |
| 5 | `/post-mortem` | Extract learnings, update knowledge base |

**That's the whole workflow.**

The knowledge flywheel, the math, the retrieval, the linking — all handled in the background. You just run 5 commands.

---

## The Problem

Your agent solves a bug today. Tomorrow? Same bug, starts from scratch.

**Every session is day one.**

| Approach | Reality |
|----------|---------|
| Spec-driven dev | Specs are discovered during implementation |
| Better prompts | Knowledge dies when the session ends |
| RAG/embeddings | No learning, just lookup |
| Other workflows | Linear, no memory, no compounding |

---

## The Solution

AgentOps captures what your agent learns and feeds it back.

```
Session 1        Session 10       Session 100
+-----------+    +-----------+    +-----------+
|  Debug    |    |  "I've    |    |  Domain   |
|  auth bug |    |  seen     |    |  expert   |
|  (2 hrs)  | -> |  this"    | -> |  instant  |
|           |    |  (10 min) |    |  recall   |
+-----------+    +-----------+    +-----------+
```

**The math:** Knowledge decays ~17%/week. AgentOps retrieval > decay = compounding.

---

## Quick Start

```bash
# 1. Install CLI
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops

# 2. Install Plugin
claude plugin marketplace add boshu2/agentops
claude plugin install agentops

# 3. Initialize (in your project)
ao init && ao hooks install
```

**That's it.** Hooks handle knowledge capture automatically.

---

## How It Works

```
     CHAOS              FILTER              RATCHET
   (explore)          (validate)        (lock progress)
       │                  │                   │
       ▼                  ▼                   ▼
  ┌─────────┐        ┌─────────┐        ┌─────────┐
  │/research│───────▶│  /vibe  │───────▶│  commit │
  │/pre-mort│        │         │        │ /post-m │
  │ /crank  │        └────┬────┘        └────┬────┘
  └─────────┘             │                  │
       ▲                  │ (fail)           │
       └──────────────────┘                  │
                                             ▼
                                      ┌───────────┐
                                      │ .agents/  │──▶ next session
                                      │ (memory)  │
                                      └───────────┘
```

**The Brownian Ratchet:** Chaos generates options. Validation filters. Ratchet locks progress. Knowledge compounds.

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

## All 20 Skills

**You only need the 5 core commands.** The rest run automatically or are power-user utilities.

| Category | Skills | Notes |
|----------|--------|-------|
| **Core (you run these)** | `/research`, `/pre-mortem`, `/crank`, `/vibe`, `/post-mortem` | The main workflow |
| **Supporting** | `/plan`, `/implement`, `/retro` | Called by /crank automatically |
| **Validation** | `/bug-hunt`, `/complexity`, `/doc` | Optional deep-dives |
| **Knowledge (automatic)** | `/forge`, `/extract`, `/inject`, `/knowledge`, `/provenance`, `/flywheel`, `/ratchet` | Runs in hooks |
| **Integration** | `/beads`, `/using-agentops` | Issue tracking, onboarding |

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
