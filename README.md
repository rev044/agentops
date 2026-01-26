# AgentOps

```
    _                    _    ___
   / \   __ _  ___ _ __ | |_ / _ \ _ __  ___
  / _ \ / _` |/ _ \ '_ \| __| | | | '_ \/ __|
 / ___ \ (_| |  __/ | | | |_| |_| | |_) \__ \
/_/   \_\__, |\___|_| |_|\__|\___/| .__/|___/
        |___/                     |_|
```

**The Knowledge Engine for Claude Code**

> *"The best time to plant a tree was 20 years ago. The second best time is now."*
>
> Your agent's memory works the same way. Start capturing knowledge today.

---

## The Problem

AI coding agents are **brilliant but amnesiac**. They solve a bug today, forget it tomorrow. You explain your architecture once, explain it again next week. Every session starts cold.

**AgentOps fixes this.** It gives your agent a persistent, git-tracked memory that compounds over time.

---

## How It Works

### The Physics: What's a Brownian Ratchet?

In physics, a **Brownian ratchet** extracts useful work from random thermal motion. Molecules bounce around chaotically, but a clever mechanism (the ratchet) only allows movement in one direction. Random noise becomes forward progress.

**AgentOps applies this to AI agents:**

| Physics | AgentOps |
|---------|----------|
| Random molecular motion | Chaotic exploration, trial and error |
| Ratchet mechanism | Validation gates (vibe checks) |
| Forward-only movement | Knowledge that persists and compounds |
| Thermal equilibrium | Session amnesia (what we're escaping) |

**The formula:** `Chaos + Filter + Ratchet = Progress`

```
+===========================================================================+
|                          THE BROWNIAN RATCHET                             |
+===========================================================================+
|                                                                           |
|  +---------------+     +---------------+     +---------------+            |
|  |  1. RESEARCH  |     |    2. PLAN    |     | 3. PRE-MORTEM |            |
|  |               |     |               |     |               |            |
|  | Mine prior    |---->| Define spec   |---->| Simulate N    |            |
|  | knowledge     |     |      |        |     | iterations    |            |
|  |      |        |     |      v        |     |      |        |            |
|  |      v        |     | Create beads  |     |      v        |            |
|  | Explore       |     | issues        |     | Find failure  |            |
|  | codebase      |     |      |        |     | modes         |            |
|  |      |        |     |      v        |     |      |        |            |
|  |      v        |     | Set deps      |     |      v        |            |
|  | Synthesis doc |     |               |     | Update spec   |            |
|  +---------------+     +---------------+     +-------+-------+            |
|                                                      |                    |
|         +--------------------------------------------+                    |
|         |                                                                 |
|         v                                                                 |
|  +-------------------------------+     +---------------+                  |
|  |          4. CRANK             |     | 5. POST-MORTEM|                  |
|  |                               |     |               |                  |
|  |  +-------------------------+  |     | Extract       |                  |
|  |  |                         |  |     | learnings     |                  |
|  |  v                         |  |     |      |        |                  |
|  | Pick --> Implement --> OK? |  |---->|      v        |                  |
|  |  ^           |        |  | |  |     | Run vibe      |                  |
|  |  |          Yes      No  | |  |     |      |        |                  |
|  |  |           |        +--+ |  |     |      v        |                  |
|  |  |           v             |  |     | Match spec?   |                  |
|  |  |        Commit           |  |     |   |     |     |                  |
|  |  |           |             |  |     |  Yes    No----+----+             |
|  |  |           v             |  |     |   |           |    |             |
|  |  +-----More? (yes)         |  |     |   v           |    |             |
|  |              |             |  |     | Match goal?   |    |             |
|  |              v (no)        |  |     |   |     |     |    |             |
|  |            DONE -----------+--+     |  Yes    No    |    |             |
|  +-------------------------------+     +---+-----+-----+    |             |
|                                            |     |          |             |
+============================================|=====|==========|=============+
                                             |     |          |
          +----------------------------------+     |          |
          |                                        |          |
          |     +----------------------------------+          |
          |     |  (retry from research if goal fails)       |
          |     +---------------------------------------------+
          |
          v
  +-------+-------+     +-----------+     +----------------+
  | RATCHET LOCKED|---->| ao index  |---->|    .agents/    |
  +---------------+     +-----------+     +-------+--------+
                                                  |
                                                  | (feeds next session)
                                                  v
                                          +---------------+
                                          |   ao inject   |
                                          +---------------+
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

Two parts: the **CLI** (for knowledge operations) and the **plugin** (for Claude Code integration).

### 1. Install the CLI

```bash
# Add the tap
brew tap boshu2/agentops https://github.com/boshu2/agentops

# Install
brew install agentops

# Verify
ao --version
```

### 2. Install the Plugin (Claude Code)

```bash
# Option A: Install from GitHub (recommended)
claude mcp add-plugin https://github.com/boshu2/agentops

# Option B: Install from local clone
git clone https://github.com/boshu2/agentops.git
claude mcp add-plugin ./agentops
```

### 3. Initialize Your Repo

```bash
cd your-project

# Initialize AgentOps in this repo
ao init

# Install hooks (this is where the magic happens)
ao hooks install

# Verify everything works
ao badge
```

### Other Platforms

**Codex:** Tell Codex:
```
Fetch and follow instructions from https://raw.githubusercontent.com/boshu2/agentops/refs/heads/main/.codex/setup.md
```

**OpenCode:** Tell OpenCode:
```
Fetch and follow instructions from https://raw.githubusercontent.com/boshu2/agentops/refs/heads/main/.opencode/setup.md
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
+---------------+                              +------------------+
|   ao inject   |                              |  ao forge index  |
| (prior know.) |                              | (extract + save) |
+-------+-------+                              +--------+---------+
        |                                               ^
        |  SessionStart                     SessionEnd  |
        v                                               |
+-------+-----------------------------------------------+-------+
|                    CLAUDE CODE SESSION                        |
|                                                               |
|    START ---> YOUR WORK ---> END                              |
|                                                               |
+---------------------------------------------------------------+
        |                                               ^
        |                                               |
        +---------------> learnings >------------------+
                              |
                              v
                    +-------------------+
                    |     .agents/      |
                    |  learnings/       |
                    |  patterns/        |
                    |  research/        |
                    +---------+---------+
                              |
                              | (feeds next session)
                              v
                       +------+------+
                       |  ao inject  |
                       +-------------+
```

**SessionStart**: Injects relevant prior knowledge (weighted by freshness + utility)

**SessionEnd**: Extracts learnings and indexes them for future sessions

---

## The Escape Velocity Equation

Knowledge decays without reinforcement. But when retrieval × usage exceeds decay, knowledge compounds.

```
WITHOUT AGENTOPS (Groundhog Day)
================================
Session 1        Session 2        Session 3
+-----------+    +-----------+    +-----------+
| Debug bug | -> | Same bug  | -> | Same bug  |
|           |    | start over|    | start over|
+-----------+    +-----------+    +-----------+
    2 hr             2 hr             2 hr         = 6 hours

WITH AGENTOPS (Compounding Knowledge)
=====================================
Session 1        Session 2        Session 3
+-----------+    +-----------+    +-----------+
| Debug bug | -> | Recall    | -> | Instant   |
| + capture |    | pattern   |    | recall    |
+-----------+    +-----------+    +-----------+
    2 hr            3 min           30 sec        = 2 hours total
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
