# AgentOps for Claude Code

[![Release](https://img.shields.io/github/v/release/boshu2/agentops?style=flat-square)](https://github.com/boshu2/agentops/releases)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude%20Code-Plugin-purple?style=flat-square)](https://docs.anthropic.com/en/docs/agents-and-tools/claude-code)

### **The Missing Memory Layer for AI Development**
**Your AI assistant shouldn't start from zero every session.**

---

## **Why AgentOps?**

You've been there: You spend 45 minutes teaching Claude how to debug a specific OAuth issue in your repo. Two weeks later, the issue returns. You open a new session, and **Claude has forgotten everything.** You pay the time (and token) cost all over again.

**AgentOps fixes this.** It runs in the background to:
1.  **Capture** successful patterns from your coding sessions.
2.  **Store** them in your repo (`.agents/`) as permanent knowledge.
3.  **Inject** that context automatically when you start a new task.

**Result:** Your assistant gets smarter—effectively *compounding* knowledge instead of resetting it.

---

## Quick Start

### 1. Install the Plugin
Add the AgentOps toolset to your Claude Code configuration.

```bash
claude mcp add boshu2/agentops
```

### 2. Install the CLI

The `ao` CLI manages your knowledge base and search index.

```bash
brew install agentops
# Or build from source:
# cd cli && go build -o ao ./cmd/ao
```

### 3. Start Your Flywheel

Initialize the knowledge base in your repository.

```bash
ao init
```

---

## The Workflow: R.P.I.

Stop "vibing" random code. Use the **RPI** loop to build software systematically.

| Phase | Command | What it does |
| --- | --- | --- |
| **1. Research** | `/research` | Deep scans your codebase & loads past learnings. |
| **2. Plan** | `/plan` | Breaks your goal into tracked issues/steps. |
| **3. Implement** | `/implement` | Executes the code. |
| **4. Validate** | `/vibe` | Runs tests, linters, and checks quality. |
| **5. Learn** | `/retro` | **Crucial:** Extracts what worked and saves it for next time. |

---

## Features

### Permanent Memory (.agents/)

Instead of ephemeral chat logs, we store knowledge in your repo.

* **Learnings:** High-level takeaways (e.g., "The auth service requires a 2-second delay on retry").
* **Patterns:** Reusable code snippets and architecture decisions.
* **Searchable:** The `ao` CLI indexes this so Claude finds it instantly.

### Powerful Skills

Your Claude instance gains specialized commands:

* **`/crank`**: Autonomous execution. Give it a goal, and it loops through RPI until done.
* **`/pre-mortem`**: Scans for risks *before* writing code.
* **`/bug-hunt`**: Specialized root-cause analysis workflow.
* **`/forge`**: Mines old transcripts to extract gold nuggets of knowledge you missed.

### The 'ao' CLI

Manage your AI's brain from the terminal.

```bash
# Check your knowledge health
$ ao badge
> COMPOUNDING (Retrieval Rate: 72%)

# Search your team's collective learnings
$ ao search "oauth retry"

# Manually ingest a past session
$ ao forge transcript ./session-logs/debug_session.txt
```

---

## Comparison

| Feature | Standard Claude Code | **With AgentOps** |
| --- | --- | --- |
| **Session Memory** | Gone when tab closes | **Persisted forever** |
| **Context** | Generic training data | **Your specific codebase history** |
| **Improvement** | Static (wait for model updates) | **Compounds daily** |
| **Cost** | Re-explain everything ($) | **Recall instantly (¢)** |

---

## All Skills

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
| `/doc` | Generate documentation | "generate docs", "doc coverage" |
| `/complexity` | Analyze code complexity | "complexity", "refactor targets" |
| `/knowledge` | Query knowledge artifacts | "find learnings", "search patterns" |

---

## Knowledge Storage

AgentOps stores knowledge in `.agents/`:

```
.agents/
├── learnings/    # Extracted lessons (what we learned)
├── patterns/     # Reusable solutions (how we solved it)
├── research/     # Exploration findings (what we found)
├── retros/       # Retrospectives (what went wrong/right)
└── ao/
    ├── sessions/ # Mined transcripts
    └── index/    # Search index
```

**Dual format:** Every artifact has `.md` (human-readable) and `.jsonl` (machine-queryable).

---

## Requirements

- [Claude Code](https://github.com/anthropics/claude-code) v1.0+
- Optional: [beads](https://github.com/beads-ai/beads) for issue tracking
- Optional: Go 1.22+ (to build ao CLI from source)

## Documentation

- [docs/brownian-ratchet.md](docs/brownian-ratchet.md) — Core philosophy
- [docs/knowledge-flywheel.md](docs/knowledge-flywheel.md) — How compounding works

## Contributing

We are building the standard for AI-assisted development workflows.

* **Issues:** [GitHub Issues](https://github.com/boshu2/agentops/issues)

## License

MIT

---

*Stop renting intelligence. Own it.*
