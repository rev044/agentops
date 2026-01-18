# AgentOps - Claude Code Plugin Marketplace

Learn AI-assisted development through progressive levels. Start with 50-line commands that just work, advance to full automation when ready.

**[See it in action →](https://www.bodenfuller.com/workflow)**

## Quick Install

```bash
# Add marketplace (inside Claude Code)
/plugin marketplace add boshu2/agentops

# Install a plugin
/plugin install vibe-kit@boshu2-agentops     # Lean starter
/plugin install gastown@boshu2-agentops      # Multi-agent orchestration

# Or clone directly
git clone https://github.com/boshu2/agentops.git ~/.claude/plugins/agentops
```

## Getting Started

Start at **Level 1**. Each level adds ONE concept:

| Level | What You Learn | Commands |
|-------|----------------|----------|
| **[L1-basics](levels/L1-basics/)** | Run commands, see results | `/research`, `/implement` |
| **[L2-persistence](levels/L2-persistence/)** | Save findings to `.agents/` | + `/retro` |
| **[L3-state-management](levels/L3-state-management/)** | Track work with issues | + `/plan`, beads |
| **[L4-parallelization](levels/L4-parallelization/)** | Execute waves in parallel | + `/implement-wave` |
| **[L5-orchestration](levels/L5-orchestration/)** | Full automation | + `/autopilot` |

```bash
# Your first command (L1)
/research "how does authentication work"

# Progress when ready
/implement  # Make changes based on research
```

Each level has **demo transcripts** showing real sessions. See `levels/L1-basics/demo/`.

## Core Workflow

```
/research → /plan → /implement → /retro
    ↓         ↓          ↓          ↓
 explore   decompose   execute    learn
```

### By Level

| Level | Workflow |
|-------|----------|
| L1 | `/research` → `/implement` (single session) |
| L2 | `/research` → `/implement` → `/retro` (with persistence) |
| L3 | `/research` → `/plan` → `/implement <id>` → `/retro` (with issues) |
| L4 | `/research` → `/plan` → `/implement-wave` → `/retro` (parallel) |
| L5 | `/autopilot <epic>` (hands-off) |

## What's Included

| Category | Count | Description |
|----------|-------|-------------|
| **[Plugins](plugins/)** | 2 | Bundled skill packs (vibe-kit, gastown) |
| **[Levels](levels/)** | 5 | Progressive learning L1-L5 |
| **[Reference](reference/)** | 3 | PDC, FAAFO, Failure Patterns |
| **[Skills](skills/)** | 12 | Domain knowledge (55 areas consolidated) |
| **[Profiles](profiles/)** | 3 | Role-based configurations |

**Note:** Skills are directly invokable with `/skill-name` - no command wrappers needed.

## Plugins

Bundled skill packs for specific workflows. Install what you need.

| Plugin | Description | Install |
|--------|-------------|---------|
| **[vibe-kit](plugins/vibe-kit/)** | Lean starter - core commands, 40% rule | `/plugin install vibe-kit@boshu2-agentops` |
| **[gastown](plugins/gastown/)** | Gas Town contribution workflow, PR skills, beads | `/plugin install gastown@boshu2-agentops` |

### gastown plugin

**CONTRIBUTING.md for agents** - executable contribution workflows.

```bash
# Via Claude Code plugin system
/plugin marketplace add boshu2/agentops
/plugin install gastown@boshu2-agentops
```

Includes 18 skills:
- **PR workflow**: `/pr-research`, `/pr-plan`, `/pr-implement`, `/pr-validate`, `/pr-prep`, `/pr-retro`
- **Gas Town**: `/beads`, `/dispatch`, `/roles`, `/mail`, `/handoff`, `/crew`, `/polecat-lifecycle`, `/gastown`, `/bd-routing`, `/status`
- **Vibe validation**: `/vibe`, `/vibe-docs`
- **Phase -1**: Prior work check is BLOCKING - searches for existing issues/PRs before starting

For [steveyegge/gastown](https://github.com/steveyegge/gastown) contributors.

## Reference Documents

Deep framework content, consulted when needed:

| Document | Content |
|----------|---------|
| **[PDC Framework](reference/pdc-framework.md)** | Prevent, Detect, Correct methodology |
| **[FAAFO Alignment](reference/faafo-alignment.md)** | Fast, Ambitious, Autonomous, Fun, Optionality |
| **[Failure Patterns](reference/failure-patterns.md)** | 12 ways AI-assisted development goes wrong |

## Skills (12 Domains)

Skills load into main context with full tool access—no sub-agent limitations.

| Skill | Triggers | Areas |
|-------|----------|-------|
| **[languages](skills/languages/)** | Python, Go, Rust, Java, TypeScript | 6 |
| **[development](skills/development/)** | API, backend, frontend, mobile, LLM | 8 |
| **[documentation](skills/documentation/)** | docs, README, OpenAPI, Diátaxis | 4 |
| **[code-quality](skills/code-quality/)** | review, test, coverage | 3 |
| **[research](skills/research/)** | explore, find, analyze | 6 |
| **[validation](skills/validation/)** | validate, verify, tracer bullet | 4 |
| **[operations](skills/operations/)** | incident, debug, postmortem | 4 |
| **[monitoring](skills/monitoring/)** | metrics, alerts, SLO | 2 |
| **[security](skills/security/)** | pentest, SSL, secrets | 2 |
| **[data](skills/data/)** | ETL, Spark, ML, MLOps | 4 |
| **[meta](skills/meta/)** | context, session, workflow | 6 |
| **[specialized](skills/specialized/)** | accessibility, UX, risk | 6 |

## Vibe Coding Framework

Based on [Vibe Coding](https://itrevolution.com/product/vibe-coding-book/) by Gene Kim & Steve Yegge.

### Trust Levels

| Level | Trust | Use For |
|-------|-------|---------|
| **L5** | 95% | Formatting, linting |
| **L4** | 80% | Boilerplate, config |
| **L3** | 60% | Standard features |
| **L2** | 40% | New features |
| **L1** | 20% | Architecture, security |
| **L0** | 0% | Novel exploration |

### The 40% Rule

- **Below 40% context** → 98% success rate
- **Above 60% context** → 24% success rate

### FAAFO Promise

**F**ast (10-16x) · **A**mbitious (solo feasible) · **A**utonomous (team output) · **F**un (50% flow) · **O**ptionality (120x options)

## Repository Structure

```
agentops/
├── plugins/              # 📦 Bundled skill packs
│   ├── vibe-kit/         # Lean starter (core skills, 40% rule)
│   └── gastown/          # Gas Town contribution workflow
├── levels/               # 🎯 START HERE - Progressive learning
│   ├── L1-basics/        # Single-session, no state
│   ├── L2-persistence/   # Add .agents/ output
│   ├── L3-state-management/  # Add issue tracking
│   ├── L4-parallelization/   # Add wave execution
│   └── L5-orchestration/     # Full autopilot
├── reference/            # Framework documentation
│   ├── pdc-framework.md
│   ├── faafo-alignment.md
│   └── failure-patterns.md
├── .agents/              # AI memory system
│   ├── research/         # Deep exploration docs
│   ├── plans/            # Implementation roadmaps
│   ├── patterns/         # Reusable solutions
│   ├── learnings/        # Session insights
│   └── retros/           # Retrospectives
├── skills/               # Domain knowledge (12 domains)
├── profiles/             # Role configurations (3)
└── .beads/               # Git-based issue tracking
```

## Related

### Ecosystem

- **[12-Factor AgentOps](https://github.com/boshu2/12-factor-agentops)** - The methodology behind this marketplace
- **[12factoragentops.com](https://12factoragentops.com)** - Interactive documentation and examples
- **[vibe-kit](./plugins/vibe-kit/)** - Recommended starter plugin (lean, production-ready)
- **[gastown](./plugins/gastown/)** - Gas Town contribution workflow for [steveyegge/gastown](https://github.com/steveyegge/gastown)

### Inspiration

- [Vibe Coding Book](https://itrevolution.com/product/vibe-coding-book/) - Gene Kim & Steve Yegge
- [vibe-check](https://www.npmjs.com/package/@boshu2/vibe-check) - Metrics CLI tool
- [Kubernetes the Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way) - Inspiration for progressive levels

## License

MIT
