# AgentOps

> **"Every AI coding tool forgets. AgentOps remembers."**

AI-assisted development workflows that compound knowledge over time.

---

## Why AgentOps?

### The Problem: AI Amnesia

Your AI assistant is brilliant in the moment and amnesiac across sessions.

You debug an OAuth token refresh issue. 45 minutes, $2.40 in tokens. Done.

Three weeks later, same issue. Claude starts fresh. Another 45 minutes. Another $2.40.

**This isn't a bug. It's thermodynamics.**

### The Science: Knowledge Decays

Research (Darr, Argote, & Epple) measured organizational knowledge decay: **17% per week**.

```
Week 0: 100% ████████████████████
Week 1:  83% ████████████████░░░░
Week 2:  69% █████████████░░░░░░░
Week 4:  47% █████████░░░░░░░░░░░
```

Every AI session starts at Week 0. By the time you need that knowledge again, it's gone.

### The Math: Escape Velocity

Knowledge evolution follows a differential equation:

```
dK/dt = I(t) - δ·K + σ·ρ·K

Where:
  I(t) = new knowledge input
  δ    = decay rate (0.17/week)
  σ    = retrieval effectiveness (can you find it?)
  ρ    = citation rate (do you use it?)
```

**The critical insight:** `σ·ρ·K` is the compounding term. When retrieval × usage exceeds decay, knowledge grows exponentially.

```
Escape velocity: σ × ρ > δ

Without AgentOps:  0 × 0 = 0.00 < 0.17  → Always decaying
With AgentOps:    0.7 × 0.3 = 0.21 > 0.17  → Compounding
```

That 0.04 difference compounds. After a year, it's the difference between an assistant that knows nothing and one that knows your entire codebase.

### The Result: Compounding Intelligence

```
WITHOUT AGENTOPS:
  Session 1: Debug OAuth refresh (45 min, $2.40)
  Session 2: Same issue, fresh start (45 min, $2.40)
  Session 3: Same issue, fresh start (45 min, $2.40)
  Total: 135 min, $7.20 — and still forgetting

WITH AGENTOPS:
  Session 1: Debug OAuth, capture pattern (45 min, $2.40)
  Session 2: "I see we solved this before" (3 min, $0.15)
  Session 3: Instant recall (1 min, $0.05)
  Total: 49 min, $2.60 — and getting faster
```

**64% time savings. 64% cost savings. And it compounds.**

---

## How It Works

### The Knowledge Flywheel

```
    CAPTURE ──────► STORE ──────► RECALL
        │             │             │
        │             │             ▼
        │             │          APPLY
        │             │             │
        │             │             ▼
        │             │          LEARN
        │             │             │
        └─────────────┴─────────────┘
                      │
              (compounds forever)
```

1. **Capture** — `/forge` extracts decisions and learnings from sessions
2. **Store** — Dual format (human-readable + machine-queryable) in `.agents/`
3. **Recall** — `/inject` retrieves relevant knowledge at session start
4. **Apply** — Knowledge used in your work
5. **Learn** — Utility scoring tracks what's actually helpful

Each turn makes the next turn easier. That's the flywheel.

### The Brownian Ratchet

Progress is permanent. You can always add more chaos, but you can't un-ratchet.

```
CHAOS ────► FILTER ────► RATCHET
  │            │            │
  │            │            └── Merged code, closed issues, stored learnings
  │            └── Tests, CI, /vibe, /pre-mortem
  └── Multiple attempts, parallel exploration
```

From physics: random motion through a one-way gate produces net forward movement.

In practice:
- **Chaos**: Explore multiple approaches (polecats, branches, experiments)
- **Filter**: Validate ruthlessly (tests pass, security clean, quality high)
- **Ratchet**: Lock progress permanently (merge, close, store)

Once knowledge is ratcheted, it never goes backward.

---

## Install

```bash
claude mcp add boshu2/agentops
```

## Quick Start

```bash
# Research a topic (builds knowledge)
/research authentication flows

# Plan work (decomposes into issues)
/plan add OAuth refresh token handling

# Implement (executes with full context)
/implement

# Validate (quality + security + architecture)
/vibe

# Extract learnings (closes the loop)
/retro
```

---

## The RPI Workflow

```
Research → Plan → Implement → Validate
    ↑                            │
    └──── Knowledge Flywheel ────┘
```

Every completion feeds back into research. Your assistant gets smarter about YOUR codebase.

---

## Skills

| Skill | Purpose | Trigger Phrases |
|-------|---------|-----------------|
| `/research` | Deep codebase exploration | "understand", "explore", "investigate" |
| `/plan` | Decompose goals into issues | "plan", "break down", "what issues" |
| `/implement` | Execute a single issue | "implement", "work on", "fix" |
| `/crank` | Autonomous multi-issue execution | "execute", "crank", "ship it" |
| `/vibe` | Validate code quality | "validate", "check", "review" |
| `/pre-mortem` | Simulate failures before building | "what could go wrong", "risks" |
| `/retro` | Extract learnings | "retrospective", "what did we learn" |
| `/post-mortem` | Full validation + extraction | "post-mortem", "wrap up" |
| `/forge` | Mine transcripts for knowledge | "forge", "extract knowledge" |
| `/inject` | Load relevant knowledge | "what do we know about" |
| `/beads` | Issue tracking | "create issue", "what's ready" |
| `/bug-hunt` | Root cause analysis | "investigate bug", "why is this broken" |

---

## ao CLI

The `ao` CLI provides the knowledge engine.

**Install:**
```bash
brew install agentops
# or build from source
cd cli && go build -o ao ./cmd/ao
```

**Commands:**
```bash
ao badge                    # Flywheel health (are you compounding?)
ao forge transcript <path>  # Extract knowledge from sessions
ao search <query>           # Find what you've learned
ao inject <context>         # Load knowledge for new session
ao ratchet status           # Workflow progress
ao feedback <id> <reward>   # Train utility scoring
```

**Example:**
```bash
$ ao badge
╔═══════════════════════════════════════════╗
║         🏛️  AGENTOPS KNOWLEDGE             ║
╠═══════════════════════════════════════════╣
║  Sessions Mined    │  47                  ║
║  Learnings         │  156                 ║
║  Patterns          │  23                  ║
║  Citations         │  89                  ║
╠═══════════════════════════════════════════╣
║  Retrieval (σ)     │  0.72  ███████░░░ ║
║  Citation Rate (ρ) │  0.34  ███░░░░░░░ ║
║  Decay (δ)         │  0.17  █░░░░░░░░░ ║
╠═══════════════════════════════════════════╣
║  σ×ρ = 0.24 > δ    │  🚀 COMPOUNDING    ║
╚═══════════════════════════════════════════╝
```

---

## Knowledge Storage

AgentOps stores knowledge in `.agents/`:

```
.agents/
├── learnings/    # Extracted lessons (what we learned)
├── patterns/     # Reusable solutions (how we solved it)
├── research/     # Exploration findings (what we found)
├── retros/       # Retrospectives (what went wrong/right)
├── products/     # Product briefs (why we're building)
└── ao/
    ├── sessions/ # Mined transcripts
    ├── index/    # Search index
    └── chain.jsonl  # Ratchet state
```

**Dual format:** Every artifact has `.md` (human-readable) and `.jsonl` (machine-queryable).

---

## The Science (Deep Dive)

### Cognitive Load Theory (Sweller, 1988)

The 40% rule isn't arbitrary. Research shows performance peaks at moderate cognitive load:

```
Performance
    │
100%│          ╭───╮
    │        ╭─╯   ╰─╮
 50%│      ╭─╯       ╰─── collapse
    │    ╭─╯
    └────────────────────────────
    0%   20%   40%   60%   80%
              Context Load
```

AgentOps checkpoints at 35%, alerts at 40%. You stay in the performance zone.

### MemRL (Zhang et al., 2026)

Not all knowledge is equally useful. MemRL uses reinforcement learning:

```python
# Utility updates based on feedback
utility = (1 - α) × old_utility + α × reward

# Retrieval ranks by freshness AND utility
score = z_norm(freshness) + λ × z_norm(utility)
```

The system learns what actually helps, not just what's recent.

### The Brownian Ratchet (Thermodynamics)

Random molecular motion + one-way gate = net forward movement.

In software: chaos (exploration) + filter (validation) + ratchet (merge) = permanent progress.

You can always add more chaos. You can't un-ratchet. Progress is locked.

---

## What Makes AgentOps Different

| Tool | Remembers? | Compounds? | Has Flywheel? |
|------|-----------|-----------|---------------|
| Cursor | No | No | No |
| Copilot | No | No | No |
| Devin | No | No | No |
| Claude Code | No | No | No |
| **+ AgentOps** | **Yes** | **Yes** | **Yes** |

Everyone has good AI. Nobody has the loop that makes it learn.

**The loop is the product.**

---

## Requirements

- [Claude Code](https://github.com/anthropics/claude-code) v1.0+
- Optional: [beads](https://github.com/beads-ai/beads) for issue tracking
- Optional: Go 1.22+ (to build ao CLI from source)

## Documentation

- **[docs/the-science.md](docs/the-science.md)** — Full research citations & math explained
- [docs/brownian-ratchet.md](docs/brownian-ratchet.md) — Core philosophy
- [docs/knowledge-flywheel.md](docs/knowledge-flywheel.md) — How compounding works
- [docs/ao-cli.md](docs/ao-cli.md) — CLI reference

## License

MIT

---

> **"You sleep. Code ships. Intelligence compounds."**
