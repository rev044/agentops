# AgentOps - Claude Code Plugin Marketplace

Production-ready patterns for AI-assisted development. The Research → Plan → Implement workflow, session persistence, and 12 domain skills consolidating 55 knowledge areas.

**[See it in action →](https://www.bodenfuller.com/workflow)**

## Quick Install

```bash
# Clone to your plugins directory
git clone https://github.com/boshu2/agentops.git ~/.claude/plugins/agentops

# Or add as marketplace
/plugin marketplace add boshu2/agentops
```

## What's Included

| Category | Count | Description |
|----------|-------|-------------|
| **Skills** | 12 | Domain knowledge packs (55 areas consolidated) |
| **Commands** | 28 | RPI workflow, sessions, bundles, quality, docs |
| **Profiles** | 3 | Role-based configurations |

## Core Workflow

```
/research → /plan → /implement → /retro
    ↓         ↓          ↓          ↓
 explore   specify    execute    learn
```

### Essential Commands

```bash
# RPI Workflow
/research          # Deep exploration before planning
/plan              # Create implementation plan with file:line specs
/implement         # Execute approved plan with validation

# Session Management
/session-start     # Initialize session with context
/session-end       # Save progress and summarize

# Context Persistence
/bundle-save       # Save context for multi-session work
/bundle-load       # Resume from saved context
```

## Skills (12 Domains)

Skills are knowledge modules that load into main context with full tool access—no sub-agent limitations.

| Skill | Triggers | Knowledge Areas |
|-------|----------|-----------------|
| **[languages](skills/languages/)** | Python, Go, Rust, Java, TypeScript, shell | 6 |
| **[development](skills/development/)** | API, backend, frontend, mobile, deploy, LLM | 8 |
| **[documentation](skills/documentation/)** | docs, README, OpenAPI, Diátaxis | 4 |
| **[code-quality](skills/code-quality/)** | review, test, coverage, quality | 3 |
| **[research](skills/research/)** | explore, find, analyze, git history | 6 |
| **[validation](skills/validation/)** | validate, verify, assumption, tracer bullet | 4 |
| **[operations](skills/operations/)** | incident, outage, debug, logs, postmortem | 4 |
| **[monitoring](skills/monitoring/)** | metrics, alerts, SLO, OpenTelemetry | 2 |
| **[security](skills/security/)** | pentest, SSL, TLS, firewall, secrets | 2 |
| **[data](skills/data/)** | pipeline, ETL, Spark, ML, MLOps | 4 |
| **[meta](skills/meta/)** | context, session, memory, retro, workflow | 6 |
| **[specialized](skills/specialized/)** | accessibility, WCAG, UX, Obsidian, risk | 6 |

### Skills vs Agents

| Aspect | Skills | Agents (Legacy) |
|--------|--------|-----------------|
| Execution | Main context | Sub-process |
| Tools | Full access | Limited |
| Context | Preserved | Isolated |
| MCP | Available | Unavailable |
| Chaining | Yes | No |

## Commands (28)

| Category | Commands |
|----------|----------|
| **RPI Workflow** | research, plan, implement |
| **Bundles** | bundle-save, bundle-load, bundle-search, bundle-list, bundle-prune, bundle-load-multi |
| **Sessions** | session-start, session-end, session-resume |
| **Metrics** | vibe-check, vibe-level |
| **Learning** | learn, retro |
| **Project** | project-init, progress-update |
| **Quality** | code-review, architecture-review, generate-tests |
| **Documentation** | update-docs, create-architecture-documentation, create-onboarding-guide |
| **Utilities** | ultra-think, maintain, containerize-application |
| **Multi-Agent** | research-multi |

## Vibe Coding Framework

Based on [Vibe Coding](https://itrevolution.com/product/vibe-coding-book/) by Gene Kim & Steve Yegge.

### Trust Levels (L0-L5)

| Level | Trust | Verification | Use For |
|-------|-------|--------------|---------|
| **L5** | 95% | Final only | Formatting, linting |
| **L4** | 80% | Spot check | Boilerplate, config |
| **L3** | 60% | Key outputs | Standard features |
| **L2** | 40% | Every change | New features |
| **L1** | 20% | Every line | Architecture, security |
| **L0** | 0% | Research only | Novel exploration |

### The 40% Rule

Context utilization matters:
- **Below 40%** → 98% success rate
- **Above 60%** → 24% success rate

### Three Feedback Loops

| Loop | Timeframe | Focus |
|------|-----------|-------|
| **Inner** | Seconds | Individual prompts |
| **Middle** | Hours | Work sessions |
| **Outer** | Days-weeks | Architecture |

### FAAFO Promise

**F**ast (10-16x) · **A**mbitious (solo feasible) · **A**utonomous (team output) · **F**un (50% flow) · **O**ptionality (120x options)

## Repository Structure

```
agentops/
├── .agents/              # AI memory system
│   ├── research/         # Deep exploration docs
│   ├── plans/            # Implementation roadmaps
│   ├── patterns/         # Reusable solutions
│   └── learnings/        # Session insights
├── .beads/               # Git-based issue tracking
├── .claude-plugin/       # Plugin manifest
├── commands/             # 28 slash commands
│   └── INDEX.md          # Command catalog
├── docs/standards/       # Coding standards
├── profiles/             # Role configurations
├── skills/               # 12 domain skills
│   ├── languages/        # Python, Go, Rust, Java, TS, Shell
│   ├── development/      # Backend, frontend, mobile, AI
│   ├── documentation/    # Diátaxis, OpenAPI
│   ├── code-quality/     # Review, testing
│   ├── research/         # Exploration, analysis
│   ├── validation/       # Verification patterns
│   ├── operations/       # Incidents, debugging
│   ├── monitoring/       # Observability
│   ├── security/         # Pentest, network
│   ├── data/             # ETL, ML, MLOps
│   ├── meta/             # Workflow coordination
│   └── specialized/      # Accessibility, UX, risk
└── agents-archived/      # Legacy agent definitions
```

## Related

- [Vibe Coding Book](https://itrevolution.com/product/vibe-coding-book/) - Gene Kim & Steve Yegge
- [Vibe Ecosystem](https://www.bodenfuller.com/builds/vibe-ecosystem) - Full documentation
- [vibe-check](https://www.npmjs.com/package/@boshu2/vibe-check) - Metrics CLI tool
- [bodenfuller.com/workflow](https://bodenfuller.com/workflow) - Video demos

## License

MIT
