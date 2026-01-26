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

### ⚙️ The Execution Layer

* **`/research`**: Deep-scans your codebase *and* your knowledge base to build a mental model before acting.
* **`/plan`**: Converts vague requirements into tracked, bite-sized issues with acceptance criteria.
* **`/crank`**: The autonomous loop. It picks up a planned issue, writes code, verifies it, and commits—repeat.

### 🛡 The Quality Layer (Vibe Coding)

* **`/vibe`**: A sanity check command. It validates code against project conventions, security standards, and architectural fit.
* **`/pre-mortem`**: Simulates failure scenarios before implementation begins.
* **`/bug-hunt`**: Systematic root-cause analysis that refuses to guess.

---

## 🔄 The Workflow: A Day in the Life

Unlike standard chat sessions, an AgentOps session follows a strict arc to ensure knowledge capture.

| Phase | Action | What happens under the hood |
| --- | --- | --- |
| **1. Init** | **Session Start** | AgentOps scans `.agents/` and injects your team's coding standards and past bug fixes into context. |
| **2. Prep** | **`/research`** | The agent explores the code and cross-references it with the Knowledge Graph to identify risks. |
| **3. Build** | **`/crank`** | The agent executes the plan autonomously. It writes tests, implements code, and runs `/vibe` checks. |
| **4. Save** | **Session End** | The agent runs a generic retro. It distills the session into markdown artifacts stored in `.agents/learnings/`. |

---

## 📊 The Knowledge Flywheel

We measure the intelligence of your repo using three metrics:

* **Sigma (σ):** Retrieval rate. (How often do we find relevant old knowledge?)
* **Rho (ρ):** Citation rate. (How often is that knowledge actually useful?)
* **Delta (δ):** Decay rate. (Knowledge fades if not refreshed).

**Goal:** Achieve **Escape Velocity** (`σ × ρ > δ`).
When you hit this, your agent learns faster than your codebase changes. You will see your repo status shift from **🌱 STARTING** to **🚀 COMPOUNDING**.

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
