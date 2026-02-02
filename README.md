<div align="center">

# AgentOps

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Skills](https://img.shields.io/badge/Skills-npx%20skills-7c3aed)](https://skills.sh/)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-blueviolet)](https://github.com/anthropics/claude-code)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

### DevOps for Vibe-Coding

**Shift-left validation for coding agents. Catch it before you ship it.**

</div>

---

<!-- Accessibility: Comparison showing traditional workflow vs shift-left workflow -->
```
+-----------------------------------------------------------------------+
|                                                                       |
|   TRADITIONAL WORKFLOW                SHIFT-LEFT WORKFLOW             |
|   -------------------                 -------------------             |
|                                                                       |
|   Write code                          /pre-mortem                     |
|      ↓                                   ↓                            |
|   Ship to CI                          Implement                       |
|      ↓                                   ↓                            |
|   CI catches problems                 /vibe (validate)                |
|      ↓                                   ↓                            |
|   Fix & repeat                        Commit (clean)                  |
|                                          ↓                            |
|   ========================            Knowledge compounds             |
|   Hope and pray                       ========================        |
|                                       Catch it before you ship it     |
|                                                                       |
+-----------------------------------------------------------------------+
```

---

## Why AgentOps?

AI-generated code is 80% valuable, 20% catastrophic. The difference is validation.

| Traditional | Shift-Left (AgentOps) |
|-------------|----------------------|
| Ship then validate | Validate then ship |
| CI catches problems | /pre-mortem catches problems |
| Hope the tests pass | /vibe confirms intent matches code |
| Same bugs rediscovered | Knowledge compounds, bugs remembered |

**The insight:** DevOps taught us to shift validation left for infrastructure. AgentOps applies that to coding agents.

---

## The Core Workflow: 3 Skills

The shift-left validation workflow:

```
                    THE SHIFT-LEFT WORKFLOW
                    -----------------------

    +-----------+         +-----------+         +-----------+
    |/PRE-MORTEM| ------> |  /CRANK   | ------> |   /VIBE   |
    | (before)  |         |(implement)|         | (before   |
    |           |         |           |         |  commit)  |
    +-----------+         +-----------+         +-----------+
         |                                            |
         |  "What could go wrong?"                    |  "Does code match intent?"
         |                                            |
         +--------------------------------------------+
                           |
                           v
                    +-----------+
                    | COMMIT    |
                    | (clean)   |
                    +-----------+

    Validation happens BEFORE you ship, not after.
```

| Skill | When | What It Does |
|-------|------|--------------|
| `/pre-mortem` | Before implementing | Simulates failure modes, catches risks before code exists |
| `/crank` | Implementation | Executes issues with validation gates at each step |
| `/vibe` | Before every commit | 8-aspect semantic check—does code match intent? |

**The insight:** CI catches problems after you ship. AgentOps catches them before.

---

## The Complete System

<!-- Accessibility: Comprehensive diagram showing validation-first workflow with 3 core skills, supporting skills, and knowledge flywheel -->
```
==============================================================================
                    AGENTOPS: DEVOPS FOR VIBE-CODING
==============================================================================

  THE SHIFT-LEFT VALIDATION WORKFLOW
  ----------------------------------

  +--------------------------------------------------------------------------+
  | STAGE 1: UNDERSTAND (optional but recommended)                           |
  |                                                                          |
  |  /research --> Deep-dive codebase before acting                          |
  |  /plan     --> Break goal into tracked issues                            |
  +--------------------------------------------------------------------------+
                                       |
                                       v
  +--------------------------------------------------------------------------+
  | STAGE 2: PRE-MORTEM (core skill)                      VALIDATION GATE 1  |
  |                                                                          |
  |  /pre-mortem --> BEFORE implementing, simulate failures                  |
  |       |                                                                  |
  |       |     4 failure experts:                                           |
  |       |     - integration-failure-expert  - ops-failure-expert           |
  |       |     - data-failure-expert         - edge-case-hunter             |
  |       |                                                                  |
  |       +---> OUTPUT: Risks identified, mitigations planned                |
  |                                                                          |
  |  "What could go wrong?" -- answered before code exists                   |
  +--------------------------------------------------------------------------+
                                       |
                                       v
  +--------------------------------------------------------------------------+
  | STAGE 3: CRANK (core skill)                           IMPLEMENTATION     |
  |                                                                          |
  |  /crank --> Execute issues with validation at each step                  |
  |       |                                                                  |
  |       |     FIRE LOOP (per issue):                                       |
  |       |     FIND ----> Get next unblocked issue                          |
  |       |     IGNITE --> Implement with validation                         |
  |       |     REAP ----> Commit with issue reference                       |
  |       |     ESCALATE > Handle failures, retry or escalate                |
  |       |                                                                  |
  |       +---> OUTPUT: Clean commits, issues closed                         |
  +--------------------------------------------------------------------------+
                                       |
                                       v
  +--------------------------------------------------------------------------+
  | STAGE 4: VIBE (core skill)                            VALIDATION GATE 2  |
  |                                                                          |
  |  /vibe --> BEFORE committing, semantic validation                        |
  |       |                                                                  |
  |       |     8-aspect check:                                              |
  |       |     - Semantic (does code match intent?)                         |
  |       |     - Security (SQL injection, auth bypass, secrets)             |
  |       |     - Quality (dead code, copy-paste, magic numbers)             |
  |       |     - Architecture (layer violations, circular deps)             |
  |       |     - Complexity (CC > 10, deep nesting)                         |
  |       |     - Performance (N+1 queries, resource leaks)                  |
  |       |     - Slop (AI hallucinations, cargo cult)                       |
  |       |     - Accessibility (ARIA, keyboard nav, contrast)               |
  |       |                                                                  |
  |       |     CRITICAL = 0 --> PASS (commit allowed)                       |
  |       |     CRITICAL > 0 --> BLOCK (fix before commit)                   |
  |       |                                                                  |
  |  "Does the code do what you intended?" -- answered before commit         |
  +--------------------------------------------------------------------------+
                                       |
                                       v
  +--------------------------------------------------------------------------+
  | STAGE 5: LEARN (closes the loop)                                         |
  |                                                                          |
  |  /retro, /post-mortem --> Extract learnings, feed the flywheel           |
  |                                                                          |
  |  "What makes the next session better?" -- every session compounds        |
  +--------------------------------------------------------------------------+
                                       |
                                       v
  +==========================================================================+
  |                        THE KNOWLEDGE FLYWHEEL                            |
  |                                                                          |
  |  SessionEnd --> Extract learnings --> Index for retrieval                |
  |  SessionStart --> Inject relevant knowledge --> Start smarter            |
  |                                                                          |
  |  Every session makes the next one better. This is the moat.              |
  +==========================================================================+

==============================================================================
            3 CORE SKILLS | 2 VALIDATION GATES | KNOWLEDGE COMPOUNDS
==============================================================================
```

---

## Quick Start

### 1. Install Skills (Any Agent)

AgentOps ships as portable skills. If your client supports the open Skills ecosystem, install with:

```bash
npx skills@latest add boshu2/agentops --all -g
```

This works across multiple clients (e.g. **Codex**, **OpenCode**, **Claude Code**, **Cursor**, etc.) — the installer writes to the right place for each agent.

Install to a specific agent only (example: Codex):

```bash
npx skills@latest add boshu2/agentops -g -a codex -s '*' -y
```

Update later:

```bash
npx skills@latest update
```

### 2. Install CLI (Optional)

```bash
brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops
brew install agentops
```

### 3. Install Plugin (Claude Code Only)

```bash
claude plugin add boshu2/agentops
```

### 4. Initialize in Your Project (Claude Code + CLI)

```bash
ao init && ao hooks install
```

Or just ask Claude: *"initialize agentops"*

### 5. Start With Validation

Before implementing your next feature:

```bash
/pre-mortem "add OAuth integration"
```

This simulates failures BEFORE you write code. Then implement with `/crank`, validate with `/vibe` before each commit.

> **Note:** There's a [known bug](https://github.com/anthropics/claude-code/issues/15178) where plugin skills don't appear when pressing `/`. Skills still work — just type them directly (e.g., `/pre-mortem`) or ask Claude to use them.

---

## Tool Dependencies

The `/vibe` and `/post-mortem` skills run `toolchain-validate.sh`, which uses available linters and scanners. **All tools are optional** — missing ones are skipped gracefully.

| Tool | Purpose | Install |
|------|---------|---------|
| **gitleaks** | Secret scanning | `brew install gitleaks` |
| **semgrep** | SAST security patterns | `brew install semgrep` |
| **trivy** | Dependency vulnerabilities | `brew install trivy` |
| **gosec** | Go security | `go install github.com/securego/gosec/v2/cmd/gosec@latest` |
| **hadolint** | Dockerfile linting | `brew install hadolint` |
| **ruff** | Python linting | `pip install ruff` |
| **radon** | Python complexity | `pip install radon` |
| **golangci-lint** | Go linting | `brew install golangci-lint` |
| **shellcheck** | Shell linting | `brew install shellcheck` |

**Quick install (recommended):**
```bash
brew install gitleaks semgrep trivy hadolint shellcheck golangci-lint
pip install ruff radon
```

More tools = more coverage. But even with zero tools installed, the workflow still runs.

---

## How AgentOps Fits In

**AgentOps is the validation layer.** Use it alongside your execution tools.

| Tool | What It Does | + AgentOps |
|------|--------------|------------|
| [Superpowers](https://github.com/obra/superpowers) | TDD, autonomous work | AgentOps adds shift-left validation |
| [Claude-Flow](https://github.com/ruvnet/claude-flow) | Multi-agent orchestration | AgentOps validates before commit |
| [cc-sdd](https://github.com/gotalab/cc-sdd) | Spec-driven development | AgentOps adds pre-mortem + vibe check |
| [GSD](https://github.com/glittercowboy/get-shit-done) | Fast shipping | AgentOps adds "catch before ship" |

*Feature comparisons as of January 2026. See [detailed comparisons](docs/comparisons/) for specifics.*

**What AgentOps uniquely adds:**

| Feature | Execution Tools | AgentOps |
|---------|:---------------:|:--------:|
| Pre-mortem failure simulation | ❌ | ✅ |
| 8-aspect semantic validation | ❌ | ✅ |
| Validation gates before commit | ❌ | ✅ |
| Knowledge that compounds | ❌ | ✅ |

> [Detailed comparisons →](docs/comparisons/)

---

## The `/vibe` Validator

Not just "does it compile?" — **does it match the spec?**

<!-- Accessibility: Table showing 8 validation aspects: Semantic, Security, Quality, Architecture, Complexity, Performance, Slop, Accessibility. Gate rule: 0 critical = pass, 1+ critical = blocked. -->
```
+------------------------------------------------------------------+
|                   8-ASPECT SEMANTIC VALIDATION                   |
+------------------------------------------------------------------+
|  [x] Semantic      Does code do what spec says?                  |
|  [x] Security      SQL injection, auth bypass, hardcoded secrets |
|  [x] Quality       Dead code, copy-paste, magic numbers          |
|  [x] Architecture  Layer violations, circular deps, god classes  |
|  [x] Complexity    CC > 10, deep nesting, parameter overload     |
|  [x] Performance   N+1 queries, unbounded loops, resource leaks  |
|  [x] Slop          AI hallucinations, cargo cult, over-engineering|
|  [x] Accessibility Missing ARIA, broken keyboard nav, contrast   |
+------------------------------------------------------------------+
|  GATE: 0 CRITICAL = pass  |  1+ CRITICAL = blocked until fixed   |
+------------------------------------------------------------------+
```

---

## The Validation Pattern

At every step: explore, validate, lock progress.

```
     +-----------------------------------------------------------+
     |                                                           |
     |   EXPLORE           VALIDATE            COMMIT            |
     |  (generate)        (check)             (lock)             |
     |                                                           |
     |  +---------+        +------+         +----------+         |
     |  | code    |  --->  |/vibe |  --->   | COMMIT   |         |
     |  +---------+        +--+---+         +----+-----+         |
     |       ^                | fail             |               |
     |       +----------------+                  v               |
     |    (fix and retry)                 +-----------+          |
     |                                    | .agents/  |          |
     |                                    | (memory)  |          |
     |                                    +-----+-----+          |
     |                                          |                |
     |                           inject learnings into           |
     |                               next session                |
     |                                                           |
     +-----------------------------------------------------------+

     Validation built in, not bolted on.
```

---

## What Gets Captured

Everything lives in `.agents/` — **git-tracked, portable, yours**.

```
.agents/
+-- research/      # Deep exploration outputs
+-- plans/         # Implementation plans
+-- pre-mortems/   # Failure simulations
+-- specs/         # Validated specifications
+-- learnings/     # Extracted insights ("Auth bugs stem from token refresh")
+-- patterns/      # Reusable patterns ("How we handle retries")
+-- retros/        # Session retrospectives
+-- vibe/          # Validation reports
+-- complexity/    # Complexity analysis
+-- ...            # + other skill outputs (doc, assessments, etc.)
```

**Automatic hooks (Claude Code plugin):**
- **SessionStart** → Injects relevant prior knowledge (with decay applied)
- **SessionEnd** → Extracts learnings, indexes for retrieval

You don't run `ao` commands manually. The flywheel turns itself.

---

## The Compound Effect

<!-- Accessibility: Comparison showing progression over 4 sessions. Without AgentOps: repeating. With AgentOps: compounding knowledge. -->
```
+--------------------------------------------------------------------+
|                                                                    |
|  WITHOUT AGENTOPS                                                  |
|  ================                                                  |
|                                                                    |
|  Session 1   Session 2   Session 3   Session 4                     |
|  +--------+  +--------+  +--------+  +--------+                    |
|  | repeat |->| repeat |->| repeat |->| repeat |   Repeating        |
|  +--------+  +--------+  +--------+  +--------+   (0 learning)     |
|                                                                    |
|  WITH AGENTOPS                                                     |
|  =============                                                     |
|                                                                    |
|  Session 1   Session 2   Session 3   Session 4                     |
|  +--------+  +--------+  +--------+  +--------+                    |
|  | learn  |->| recall |->| refine |->| expert |   Compounding      |
|  +--------+  +--------+  +--------+  +--------+   (mastered)       |
|                                                                    |
+--------------------------------------------------------------------+
```

**By session 100, your agent knows:**
- Every bug you've ever fixed
- Your architecture decisions and *why*
- Your team's coding patterns
- What approaches failed and why

---

## Agent Farm

Spawn multiple agents to work on issues in parallel:

```bash
ao farm validate          # Pre-flight checks (cycles, bd availability)
ao farm start --agents 3  # Spawn 3 agents + witness in tmux
ao farm status            # Check progress
ao inbox                  # View messages from agents
ao farm stop              # Graceful shutdown
```

**How it works:**
1. Validates issues have no cycles and bd CLI is available
2. Spawns N Claude agents in tmux sessions (30s stagger for rate limits)
3. Witness monitors progress, sends summaries to mayor
4. Circuit breaker stops farm if >50% agents fail
5. Use `ao farm resume` to continue after interruption

---

## All Skills

**Core workflow: 3 skills. The rest are supporting.**

| Category | Skills | Purpose |
|----------|--------|---------|
| **Core validation** | `/pre-mortem`, `/crank`, `/vibe` | The shift-left workflow |
| **Supporting** | `/research`, `/plan`, `/retro`, `/post-mortem` | Context and learning |
| **Multi-agent** | `/farm` | Spawn parallel agents |
| **Called by /crank** | `/implement` | Single issue execution |
| **Issue tracking** | `/beads` | Create and manage issues |
| **Language rules** | `/standards` | Apply language-specific patterns |
| **Deep dives** | `/bug-hunt`, `/complexity`, `/doc` | On-demand analysis |
| **Background** | `/forge`, `/extract`, `/inject`, `/knowledge`, `/flywheel` | Hooks run these |

---

## CLI Reference

```bash
ao init                # Initialize AgentOps in repo
ao hooks install       # Install session hooks
ao inject [topic]      # Manually inject knowledge
ao forge search        # Search past sessions
ao forge index         # Index artifacts
ao feedback            # Mark learnings as helpful/harmful
ao metrics             # View flywheel health

# Agent Farm
ao farm validate       # Pre-flight checks
ao farm start --agents 3  # Spawn agents
ao farm status         # Check progress
ao farm stop           # Graceful shutdown
ao inbox               # View agent messages
ao mail send --to mayor --body "message"  # Send message
```

---

## Known Limitations

Agent Farm is production-ready for supervised use (1-3 agents). Large/unattended farms have known edge cases:

| Issue | Impact | Workaround |
|-------|--------|------------|
| **Race condition on issue claims** | Multiple agents may claim same issue | Use `--agents 3` or fewer for reliability |
| **Witness crash not auto-detected** | Farm hangs if witness dies | Run `ao farm status` periodically to check |
| **Graceful shutdown required** | Ctrl+C may leave orphaned sessions | Always use `ao farm stop` |
| **Project name spaces** | tmux session names break with spaces | Avoid spaces in directory names |

These will be addressed in the Mt-Olympus orchestrator project.

---

## The Science

Built on peer-reviewed research, not vibes.

| Concept | Source | Finding |
|---------|--------|---------|
| Knowledge Decay | [Darr, Argote & Epple (1995)](https://pubsonline.informs.org/doi/abs/10.1287/mnsc.41.11.1750) | Org knowledge depreciates ~17%/week |
| Memory Reinforcement | [Ebbinghaus (1885)](https://en.wikipedia.org/wiki/Forgetting_curve) | Retrieval strengthens memory, slows decay |
| MemRL | [Zhang et al. (2025)](https://arxiv.org/abs/2502.06173) | Two-phase retrieval enables self-evolving agents |

**The math:**
```
dK/dt = I(t) - δ·K + σ·ρ·K

Where:
  δ = 0.17/week (decay rate)
  σ = retrieval effectiveness
  ρ = usage rate

Goal: σ·ρ > δ  →  Knowledge compounds instead of decays
```

---

## Built On

| Tool | Author | What We Use |
|------|--------|-------------|
| [beads](https://github.com/steveyegge/beads) | Steve Yegge | Git-native issue tracking |
| [CASS](https://github.com/Dicklesworthstone/coding_agent_session_search) | Dicklesworthstone | Session indexing |
| [multiclaude](https://github.com/dlorenc/multiclaude) | dlorenc | Brownian Ratchet pattern |

---

## License

Apache-2.0

---

<div align="center">

### Stop shipping and praying.

**Validation built in, not bolted on. Knowledge that compounds.**

[Get Started](#quick-start) · [Documentation](docs/) · [Comparisons](docs/comparisons/) · [Changelog](CHANGELOG.md)

</div>
