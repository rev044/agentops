# 🧠 AgentOps

### The Knowledge Engine for Claude Code

> **Stop starting from zero.** Standard AI agents have amnesia; they reset every session. AgentOps gives your agent a persistent, git-tracked memory that compounds over time.

---

## ⚡️ The "So What"

Most coding agents are brilliant but forgetful. They solve a complex bug today, but if you ask them to fix a similar issue next week, they have to figure it out from scratch.

**AgentOps changes the physics of AI coding.** It introduces a **Knowledge Flywheel**:

1. **Capture:** Every decision, pattern, and edge case is recorded.
2. **Index:** Knowledge is stored permanently in your repo (`.agents/`).
3. **Inject:** Relevant context is automatically loaded into future sessions.

After 5 sessions, your agent knows your tech stack. After 50, it knows your business logic better than you do.

---

## 🚀 Quick Start

### 1. Install the Core

Install the CLI tool that manages your knowledge base.

```bash
brew install boshu2/agentops/agentops
```

### 2. Connect to Claude

Bridge the CLI with Claude Code.

```bash
claude mcp add boshu2/agentops
```

### 3. Initialize Repository

Run this in the root of your project to create the memory structure.

```bash
ao init && ao hooks install
```

### 4. Verify

Check the health of your knowledge flywheel.

```bash
ao badge
# Output: 🌱 STARTING (This will grow as you code)
```

---

## 🛠 Capabilities

AgentOps isn't just a prompt library; it's a full-stack memory architecture.

### 🧠 The Memory Layer

* **Context Injection:** Automatically loads relevant docs and past learnings before the agent writes a line of code.
* **`ao forge`:** Mines chat transcripts to extract high-value patterns.
* **`ao retro`:** A mandatory cleanup step that saves "what worked" and "what didn't" to Git.

### ⚙️ The Execution Layer (Subagent-Driven Development)

* **`/research`**: Deep-scans your codebase *and* your knowledge base to build a mental model before acting.
* **`/plan`**: Converts vague requirements into tracked [beads](https://github.com/beads-ai/beads) issues with acceptance criteria.
* **`/crank`**: The autonomous loop. It picks up a planned issue, writes code, validates, commits—repeat.

**Subagent-Driven Development:** `/crank` dispatches fresh subagents per task. Each works in isolation, gets reviewed, and merges independently. Failures don't cascade.

### 🛡 The Quality Layer (Ratchet Validation)

Every stage has a gate. Progress locks in; failures don't propagate.

* **`/vibe`**: Validates code against project conventions, security standards, and architectural fit.
* **`/pre-mortem`**: Simulates failure scenarios *before* implementation begins.
* **`/bug-hunt`**: Systematic root-cause analysis that refuses to guess.

**The Brownian Ratchet:** Like a molecular ratchet that only moves forward, each validated commit is permanent progress. Subagents can fail independently—the system extracts success from parallel chaos.

---

## 🔄 The Workflow: A Day in the Life

Unlike standard chat sessions, AgentOps follows a strict arc with **ratchets at every stage**.

| Phase | Action | Ratchet (What Locks In) |
| --- | --- | --- |
| **1. Init** | **Session Start** | Knowledge injected from `.agents/` ✓ |
| **2. Research** | **`/research`** | Understanding documented in `.agents/research/` ✓ |
| **3. Plan** | **`/plan`** | Issues created in beads (git-tracked) ✓ |
| **4. Build** | **`/crank`** | Each issue: implement → `/vibe` → commit → close ✓ |
| **5. Save** | **Session End** | Learnings extracted to `.agents/learnings/` ✓ |

**Every stage produces an artifact.** Nothing is lost. Progress only moves forward.

---

## 📊 The Knowledge Flywheel

We measure the intelligence of your repo using three metrics:

* **Sigma (σ):** Retrieval rate. (How often do we find relevant old knowledge?)
* **Rho (ρ):** Citation rate. (How often is that knowledge actually useful?)
* **Delta (δ):** Decay rate. (Knowledge fades if not refreshed).

**Goal:** Achieve **Escape Velocity** (`σ × ρ > δ`).
When you hit this, your agent learns faster than your codebase changes. You will see your repo status shift from **🌱 STARTING** to **🚀 COMPOUNDING**.

---

## 📚 The Science

AgentOps isn't built on vibes. It's built on research.

### The Equation

```
dK/dt = I(t) - δ·K + σ·ρ·K
```

This models knowledge as a dynamic system: input (`I`), decay (`δ`), and compounding (`σ·ρ`). When retrieval × usage exceeds decay, the system grows.

### Key Research

| Concept | Source | Finding |
|---------|--------|---------|
| **Knowledge Decay** | Darr, Argote & Epple (1995) | Organizational knowledge depreciates ~17%/week without reinforcement |
| **Memory Reinforcement** | Ebbinghaus (1885) | Each retrieval strengthens memory and slows future decay |
| **Cognitive Load** | Sweller (1988), Paas & van Merriënboer (2020) | Performance peaks at moderate load (~40%), collapses at overload |
| **Lost in the Middle** | Liu et al. (2023) | LLMs lose information in crowded contexts; sparse is better |
| **MemRL** | Zhang et al. (2025) | Two-phase retrieval (semantic + utility) enables self-evolving agents |

### MemRL: The Memory Foundation

Our retrieval system is based on [MemRL](https://arxiv.org/abs/2601.03192) (Zhang, Wang, Zhou, et al., 2025):

> *"MemRL separates the stable reasoning of a frozen LLM from the plastic, evolving memory, allowing continuous runtime improvement through trial-and-error learning."*

Key insight: **Two-Phase Retrieval** filters candidates by semantic relevance first, then ranks by learned utility (Q-values). This is why `ao feedback` matters—it trains the system to surface what actually helps.

📖 **Deep Dive:** [docs/the-science.md](docs/the-science.md) — Full citations, equations, and the complete research stack.

---

## 📂 Architecture

Your knowledge base lives in your repo, not in the cloud. It is fully portable and version-controlled.

```
.agents/
├── learnings/     # Extracted wisdom (The "Long Term Memory")
├── patterns/      # Reusable code snippets and architectural decisions
├── research/      # Deep dive outputs (preventing re-work)
├── retros/        # Session logs and improvement vectors
└── ao/            # Search indices and graph data
```

> **Privacy Note:** Since `.agents/` is just a folder in your repo, your knowledge stays with your code. It works with local LLMs, Enterprise Claude, or any Git host.

---

## 🤝 Contributing

We are building the standard for stateful AI agents.

1. Fork the repo.
2. Create a branch for your new Skill or Memory Driver.
3. Submit a PR.

## 📄 License

MIT © [boshu2](https://github.com/boshu2)

---

<p align="center">
<em>Stop renting intelligence. Own it.</em>
</p>
