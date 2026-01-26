# AgentOps

**The Knowledge Engine for Claude Code**

> Stop starting from zero. Your agent learns, remembers, and compounds knowledge across sessions.

---

## The Problem

AI coding agents are brilliant but amnesiac. They solve a bug today, forget it tomorrow. You explain your architecture once, explain it again next week. Every session starts cold.

**AgentOps fixes this.** It gives your agent a persistent, git-tracked memory that compounds over time.

---

## How It Works

```
    ╭─────────────────╮                                    ╭─────────────────╮
    │  SESSION START  │                                    │    KNOWLEDGE    │
    │    (hooks)      │                                    │    .agents/     │
    ╰────────┬────────╯                                    ╰────────┬────────╯
             │                                                      │
             │  ao inject                                           │
             │  (prior knowledge)                                   │
             ▼                                                      │
    ╭────────────────────────────────────────────────────╮         │
    │                  YOUR WORKFLOW                      │         │
    │                                                     │         │
    │    /research ──▶ /plan ──▶ /pre-mortem             │         │
    │                                 │                   │         │
    │                                 ▼                   │         │
    │                             /crank ◀───────────╮    │         │
    │                                 │              │    │         │
    │                                 ▼          (retry)  │         │
    │                           /post-mortem ────────╯    │         │
    │                                                     │         │
    ╰─────────────────────────┬───────────────────────────╯         │
                              │                                     │
                              │  learnings                          │
                              ╰─────────────────────────────────────╯
                                          ▲
                                          │
                              ╭───────────┴───────────╮
                              │   FEEDS NEXT SESSION  │
                              ╰───────────────────────╯
```

| Step | What Happens |
|------|--------------|
| **Session Start** | Hooks inject relevant knowledge from past sessions |
| **/research** | Mine your knowledge base before diving in |
| **/plan** | Break work into tracked issues with dependencies |
| **/pre-mortem** | Simulate failures *before* they happen |
| **/crank** | Autonomous loop: implement → validate → commit → repeat |
| **/post-mortem** | Extract learnings, index for future sessions |

**The flywheel:** Each cycle feeds the next. Knowledge compounds.

---

## Installation

**Note:** Installation differs by platform. Claude Code has a built-in plugin system. Codex and OpenCode require manual setup.

### Claude Code (Recommended)

```bash
# 1. Install the CLI
brew install boshu2/agentops/agentops

# 2. Initialize your repo
ao init

# 3. Install hooks (this is where the magic happens)
ao hooks install

# 4. Verify
ao badge
```

### Codex

Tell Codex:

```
Fetch and follow instructions from https://raw.githubusercontent.com/boshu2/agentops/refs/heads/main/.codex/setup.md
```

### OpenCode

Tell OpenCode:

```
Fetch and follow instructions from https://raw.githubusercontent.com/boshu2/agentops/refs/heads/main/.opencode/setup.md
```

### Verify Installation

The hooks should fire automatically. Check that knowledge injection works:

```bash
# Start a Claude Code session - you should see injected context
claude

# Or manually test
ao inject --dry-run
```

---

## The Workflow

**Chaos + Filter + Ratchet = Progress**

Each phase produces chaos, filters it for quality, then ratchets progress permanently. You can always add more chaos, but you can't un-ratchet.

```
╔═══════════════════════════════════════════════════════════════════════════════╗
║                           THE BROWNIAN RATCHET                                 ║
╠═══════════════════════════════════════════════════════════════════════════════╣
║                                                                                ║
║   ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐           ║
║   │  1. RESEARCH    │    │    2. PLAN      │    │  3. PRE-MORTEM  │           ║
║   │                 │    │                 │    │                 │           ║
║   │  Mine prior     │───▶│  Define spec    │───▶│  Simulate N     │           ║
║   │  knowledge      │    │       │         │    │  iterations     │           ║
║   │       │         │    │       ▼         │    │       │         │           ║
║   │       ▼         │    │  Create beads   │    │       ▼         │           ║
║   │  Explore        │    │  issues         │    │  Find failure   │           ║
║   │  codebase       │    │       │         │    │  modes          │           ║
║   │       │         │    │       ▼         │    │       │         │           ║
║   │       ▼         │    │  Set            │    │       ▼         │           ║
║   │  Create         │    │  dependencies   │    │  Update spec    │           ║
║   │  synthesis doc  │    │                 │    │                 │           ║
║   └─────────────────┘    └─────────────────┘    └────────┬────────┘           ║
║                                                          │                    ║
║           ╭──────────────────────────────────────────────╯                    ║
║           │                                                                   ║
║           ▼                                                                   ║
║   ┌───────────────────────────────────────────┐    ┌─────────────────┐       ║
║   │              4. CRANK                      │    │  5. POST-MORTEM │       ║
║   │                                            │    │                 │       ║
║   │   ╭──────────────────────────────────╮    │    │  Extract        │       ║
║   │   │                                  │    │    │  learnings      │       ║
║   │   ▼                                  │    │    │       │         │       ║
║   │  Pick issue ──▶ Implement ──▶ Vibe? ─┤    │───▶│       ▼         │       ║
║   │   ▲                          │  │    │    │    │  Run full vibe  │       ║
║   │   │                         Yes No   │    │    │       │         │       ║
║   │   │                          │  │    │    │    │       ▼         │       ║
║   │   │                          ▼  ╰────╯    │    │  Matches spec?  │       ║
║   │   │                       Commit          │    │    │       │    │       ║
║   │   │                          │            │    │   Yes      No   │       ║
║   │   │                          ▼            │    │    │       │    │       ║
║   │   │                    More issues? ──────╯    │    │       ╰────╋──╮    ║
║   │   │                          │ No              │    │            │  │    ║
║   │   │                          ▼                 │    ▼            │  │    ║
║   │   │                        DONE ───────────────╯  Goal?          │  │    ║
║   │   │                                                │    │        │  │    ║
║   │   ╰────────────────────────────────────────────────╯   Yes       │  │    ║
║   │                     (retry if no match goal)            │        │  │    ║
║   └─────────────────────────────────────────────────────────┼────────╯  │    ║
║                                                             │           │    ║
╚═════════════════════════════════════════════════════════════╪═══════════╪════╝
                                                              │           │
                                                              ▼           │
                                              ╔═══════════════════════════╧════╗
                                              ║     🔒 RATCHET LOCKED          ║
                                              ╠════════════════════════════════╣
                                              ║                                ║
                                              ║  ao forge index ──▶ .agents/   ║
                                              ║         │                      ║
                                              ║         ▼                      ║
                                              ║     ao inject                  ║
                                              ║         │                      ║
                                              ║         ╰──▶ (next session)    ║
                                              ║                                ║
                                              ╚════════════════════════════════╝
```

---

## What Each Phase Does

| Phase | What Happens | Output |
|-------|--------------|--------|
| **Research** | Mine prior knowledge, explore codebase, synthesize findings | `.agents/research/` |
| **Plan** | Define spec, create tracked issues with dependencies | `.beads/` issues |
| **Pre-Mortem** | Simulate failures before they happen, update spec | Hardened spec |
| **Crank** | Autonomous loop: implement → validate → commit → repeat | Merged code |
| **Post-Mortem** | Extract learnings, validate against spec and goal | `.agents/learnings/` |

---

## How It's Automated

You don't manually run `ao` commands. Hooks do it for you.

```
                         ┌─────────────────────────────────────┐
                         │        CLAUDE CODE SESSION          │
                         │                                     │
   ╭───────────────╮     │  ┌─────────┐                       │
   │   ao inject   │────▶│  │  START  │                       │
   │               │     │  └────┬────┘                       │
   │ (loads prior  │     │       │                            │
   │  knowledge)   │     │       ▼                            │
   ╰───────┬───────╯     │  ┌─────────┐                       │
           │             │  │  YOUR   │                       │
           │             │  │  WORK   │                       │
           │             │  └────┬────┘                       │
           │             │       │                            │
           │             │       ▼                            │
           │             │  ┌─────────┐     ╭─────────────────╯
           │             │  │   END   │────▶│
           │             └──┴─────────┴─────┤
           │                                │
           │                                ▼
           │                     ╭─────────────────────╮
           │                     │  ao forge transcript │
           │                     │  (extracts learnings)│
           │                     ╰──────────┬──────────╯
           │                                │
           │                                ▼
           │                     ╭─────────────────────╮
           │                     │   ao forge index    │
           │                     ╰──────────┬──────────╯
           │                                │
           │    ┌───────────────────────────┼───────────────────────────┐
           │    │          YOUR REPO        │                           │
           │    │                           ▼                           │
           │    │    ┌──────────────────────────────────────────┐      │
           │    │    │              .agents/                     │      │
           │    │    │  ┌────────────┐ ┌──────────┐ ┌─────────┐ │      │
           │    │    │  │ learnings/ │ │patterns/ │ │research/│ │      │
           │    │    │  └─────┬──────┘ └────┬─────┘ └────┬────┘ │      │
           │    │    └────────┼─────────────┼────────────┼──────┘      │
           │    └─────────────┼─────────────┼────────────┼─────────────┘
           │                  │             │            │
           ╰──────────────────┴─────────────┴────────────╯
                              (feeds next session)
```

**SessionStart**: Injects relevant prior knowledge (weighted by freshness + utility)

**SessionEnd**: Extracts learnings and indexes them for future sessions

---

## The Escape Velocity Equation

Knowledge decays without reinforcement. But when retrieval × usage exceeds decay, knowledge compounds.

```
  ┌─────────────────────────────────────────────────────────────────────────┐
  │                         WITHOUT AGENTOPS                                 │
  │                                                                          │
  │   Session 1          Session 2          Session 3                        │
  │  ┌──────────┐       ┌──────────┐       ┌──────────┐                     │
  │  │ Debug    │ ───▶  │ Same bug │ ───▶  │ Same bug │     😩 Groundhog    │
  │  │ bug      │       │ start    │       │ start    │        Day          │
  │  │          │       │ fresh    │       │ fresh    │                     │
  │  └──────────┘       └──────────┘       └──────────┘                     │
  │       ⏱️ 2hr             ⏱️ 2hr             ⏱️ 2hr                        │
  └─────────────────────────────────────────────────────────────────────────┘

  ┌─────────────────────────────────────────────────────────────────────────┐
  │                          WITH AGENTOPS                                   │
  │                                                                          │
  │   Session 1          Session 2          Session 3                        │
  │  ┌──────────┐       ┌──────────┐       ┌──────────┐                     │
  │  │ Debug    │ ───▶  │ Recall   │ ───▶  │ Instant  │     🚀 Compounding  │
  │  │ bug +    │       │ pattern  │       │ recall   │        Knowledge    │
  │  │ capture  │       │ 3 min    │       │          │                     │
  │  └──────────┘       └──────────┘       └──────────┘                     │
  │       ⏱️ 2hr             ⏱️ 3min            ⏱️ 30sec                      │
  └─────────────────────────────────────────────────────────────────────────┘
```

**The Math:**

```
dK/dt = I(t) - δK + σρK

Where:
  δ = 0.17/week    (knowledge decay rate)
  σ = retrieval effectiveness
  ρ = citation rate

Goal: σ × ρ > δ → Knowledge compounds faster than it fades
```

---

## What's Inside

### Skills

| Skill | Triggers | What It Does |
|-------|----------|--------------|
| `/research` | "research", "explore", "investigate" | Deep codebase exploration with knowledge mining |
| `/plan` | "create a plan", "break down" | Convert goals into tracked beads issues |
| `/pre-mortem` | "what could go wrong", "simulate" | Find failure modes before implementation |
| `/crank` | "execute", "go", "ship it" | Autonomous implementation loop |
| `/vibe` | "validate", "check quality" | 9-aspect code validation |
| `/post-mortem` | "what did we learn", "wrap up" | Extract and index learnings |

### CLI Commands

| Command | Purpose |
|---------|---------|
| `ao inject` | Inject knowledge into current session |
| `ao forge search` | Search CASS-indexed sessions |
| `ao forge index` | Index artifacts for retrieval |
| `ao forge transcript` | Extract learnings from transcripts |
| `ao feedback` | Mark learnings as helpful/harmful |
| `ao ratchet` | Track provenance chain |
| `ao hooks install` | Install SessionStart/End hooks |

---

## Storage Architecture

Everything lives in your repo. Portable, version-controlled, yours.

```
.agents/
  learnings/     # Extracted insights (with confidence + maturity)
  patterns/      # Reusable solutions
  research/      # Deep dive outputs
  retros/        # Session retrospectives
  deltas/        # Spec vs reality mismatches
  specs/         # Validated specifications
  ao/            # Search indices
```

---

## The Science

Built on peer-reviewed research, not vibes.

| Concept | Source | Finding |
|---------|--------|---------|
| **Knowledge Decay** | Darr, Argote & Epple (1995) | Organizational knowledge depreciates ~17%/week without reinforcement |
| **Memory Reinforcement** | Ebbinghaus (1885) | Each retrieval strengthens memory and slows future decay |
| **MemRL** | Zhang et al. (2025) | Two-phase retrieval (semantic + utility) enables self-evolving agents |

---

## Credits

Built on excellent open-source work:

| Tool | Author | What We Use | Link |
|------|--------|-------------|------|
| **beads** | Steve Yegge | Git-native issue tracking | [steveyegge/beads](https://github.com/steveyegge/beads) |
| **CASS** | Dicklesworthstone | Session indexing and search | [coding_agent_session_search](https://github.com/Dicklesworthstone/coding_agent_session_search) |
| **cass-memory** | Dicklesworthstone | Confidence decay, maturity tracking | [cass_memory_system](https://github.com/Dicklesworthstone/cass_memory_system) |
| **multiclaude** | dlorenc | The "Brownian Ratchet" pattern | [dlorenc/multiclaude](https://github.com/dlorenc/multiclaude) |

---

## Optional: Parallel Execution

For larger projects, **gastown** enables parallel agent execution:

```
/crank (single agent) --> gastown (multiple polecats in parallel)
```

Each polecat works in isolation. CI validates. Passing work merges. Failures don't cascade.

---

## License

MIT
