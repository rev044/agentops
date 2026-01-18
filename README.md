# AgentOps - Claude Code Plugin Marketplace

Learn AI-assisted development through progressive levels. Start with 50-line commands that just work, advance to full automation when ready.

**[See it in action →](https://www.bodenfuller.com/workflow)**

## Quick Install

```bash
# Add marketplace (inside Claude Code)
/plugin marketplace add boshu2/agentops

# Install kits you need (Unix philosophy: small, focused tools)
/plugin install core-kit@boshu2-agentops     # Workflow essentials
/plugin install vibe-kit@boshu2-agentops     # Validation

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
| **[L5-orchestration](levels/L5-orchestration/)** | Full automation | + `/crank` |

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
| L5 | `/crank` (fully autonomous) |

## Plugins (Unix Philosophy)

**Small, focused kits that do one thing well.** Install only what you need.

| Kit | Skills | Purpose | Install |
|-----|--------|---------|---------|
| **[core-kit](plugins/core-kit/)** | 8 | Workflow: research, plan, implement, crank | `/plugin install core-kit@boshu2-agentops` |
| **[vibe-kit](plugins/vibe-kit/)** | 5 | Validation: vibe, bug-hunt, complexity | `/plugin install vibe-kit@boshu2-agentops` |
| **[docs-kit](plugins/docs-kit/)** | 3 | Documentation: doc, oss-docs, golden-init | `/plugin install docs-kit@boshu2-agentops` |
| **[beads-kit](plugins/beads-kit/)** | 3 | Issue tracking: beads, status, molecules | `/plugin install beads-kit@boshu2-agentops` |
| **[dispatch-kit](plugins/dispatch-kit/)** | 4 | Orchestration: dispatch, handoff, roles, mail | `/plugin install dispatch-kit@boshu2-agentops` |
| **[pr-kit](plugins/pr-kit/)** | 6 | OSS contribution: pr-research through pr-retro | `/plugin install pr-kit@boshu2-agentops` |
| **[gastown-kit](plugins/gastown-kit/)** | 4 | Gas Town: crew, polecat, gastown, bd-routing | `/plugin install gastown-kit@boshu2-agentops` |
| **[domain-kit](plugins/domain-kit/)** | 18 | Reference knowledge: languages, ops, security | `/plugin install domain-kit@boshu2-agentops` |

### Recommended Combinations

| Use Case | Install |
|----------|---------|
| **Getting started** | `core-kit` + `vibe-kit` |
| **Full development** | `core-kit` + `vibe-kit` + `beads-kit` + `docs-kit` |
| **Multi-agent orchestration** | Add `dispatch-kit` + `gastown-kit` |
| **OSS contributions** | Add `pr-kit` |
| **Domain expertise** | Add `domain-kit` |

### Kit Details

#### core-kit (Recommended Start)

The complete workflow from exploration to execution.

```bash
/plugin install core-kit@boshu2-agentops
```

Skills: `/research`, `/plan`, `/formulate`, `/product`, `/implement`, `/implement-wave`, `/crank`, `/retro`

**When to use which:**

| Skill | Use When |
|-------|----------|
| `/implement` | Single issue, want to review each step |
| `/implement-wave` | Multiple independent issues, parallel execution |
| `/crank` | Full epic, trust the plan, run to completion |
| `/plan` | One-off decomposition into beads issues |
| `/formulate` | Repeatable pattern, creates reusable `.formula.toml` |

#### vibe-kit

Validation and quality assurance.

```bash
/plugin install vibe-kit@boshu2-agentops
```

Skills: `/vibe`, `/vibe-docs`, `/validation-chain`, `/bug-hunt`, `/complexity`

Includes 4 expert agents: security, architecture, code-quality, UX.

#### pr-kit

Open source contribution workflow with Phase -1 prior work check.

```bash
/plugin install pr-kit@boshu2-agentops
```

Skills: `/pr-research` → `/pr-plan` → `/pr-implement` → `/pr-validate` → `/pr-prep` → `/pr-retro`

## Reference Documents

Deep framework content, consulted when needed:

| Document | Content |
|----------|---------|
| **[PDC Framework](reference/pdc-framework.md)** | Prevent, Detect, Correct methodology |
| **[FAAFO Alignment](reference/faafo-alignment.md)** | Fast, Ambitious, Autonomous, Fun, Optionality |
| **[Failure Patterns](reference/failure-patterns.md)** | 12 ways AI-assisted development goes wrong |

## Vibe Coding Framework

Based on [Vibe Coding](https://itrevolution.com/product/vibe-coding-book/) by Gene Kim & Steve Yegge.

### The 40% Rule

- **Below 40% context** → 98% success rate
- **Above 60% context** → 24% success rate

### Trust Levels

| Level | Trust | Use For |
|-------|-------|---------|
| **L5** | 95% | Formatting, linting |
| **L4** | 80% | Boilerplate, config |
| **L3** | 60% | Standard features |
| **L2** | 40% | New features |
| **L1** | 20% | Architecture, security |
| **L0** | 0% | Novel exploration |

### FAAFO Promise

**F**ast (10-16x) · **A**mbitious (solo feasible) · **A**utonomous (team output) · **F**un (50% flow) · **O**ptionality (120x options)

## Repository Structure

```
agentops/
├── plugins/              # 📦 Unix-style kits (8 focused plugins)
│   ├── core-kit/         # Workflow: research, plan, implement
│   ├── vibe-kit/         # Validation: vibe, bug-hunt, complexity
│   ├── docs-kit/         # Documentation generation
│   ├── beads-kit/        # Issue tracking
│   ├── dispatch-kit/     # Work assignment
│   ├── pr-kit/           # OSS contribution
│   ├── gastown-kit/      # Multi-agent orchestration
│   └── domain-kit/       # Reference knowledge (18 domains)
├── levels/               # 🎯 START HERE - Progressive learning
│   ├── L1-basics/        # Single-session, no state
│   ├── L2-persistence/   # Add .agents/ output
│   ├── L3-state-management/  # Add issue tracking
│   ├── L4-parallelization/   # Add wave execution
│   └── L5-orchestration/     # Full automation
├── reference/            # Framework documentation
├── skills/               # Domain knowledge (reference)
├── profiles/             # Role configurations
└── .beads/               # Git-based issue tracking
```

## Related

### Ecosystem

- **[12-Factor AgentOps](https://github.com/boshu2/12-factor-agentops)** - The methodology behind this marketplace
- **[12factoragentops.com](https://12factoragentops.com)** - Interactive documentation and examples

### Inspiration

- [Vibe Coding Book](https://itrevolution.com/product/vibe-coding-book/) - Gene Kim & Steve Yegge
- [vibe-check](https://www.npmjs.com/package/@boshu2/vibe-check) - Metrics CLI tool
- [Kubernetes the Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way) - Inspiration for progressive levels

## License

MIT
