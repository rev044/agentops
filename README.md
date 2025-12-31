# AgentOps - Claude Code Plugin Marketplace

Production-ready patterns for AI-assisted development. The Research → Plan → Implement workflow, session persistence, and 55 specialized agents organized into 12 domains.

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
| **Agents** | 55 | Specialized agents across 12 domains |
| **Commands** | 28 | RPI workflow, sessions, bundles, quality, docs |
| **Skills** | 7 | Domain knowledge packs |
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

## Agent Domains (12)

| Domain | Count | Examples |
|--------|-------|----------|
| **languages** | 6 | python-pro, golang-pro, rust-pro, java-pro, typescript-pro |
| **development** | 8 | backend-architect, frontend-developer, ai-engineer, fullstack-developer |
| **documentation** | 4 | documentation-create-docs, api-documenter, documentation-diataxis-auditor |
| **code_quality** | 3 | code-reviewer, code-review-improve, test-generator |
| **research** | 6 | code-explorer, history-explorer, spec-architect, doc-explorer |
| **validation** | 4 | assumption-validator, tracer-bullet-deployer, validation-planner |
| **operations** | 4 | incident-responder, error-detective, incidents-postmortems |
| **monitoring** | 2 | performance-engineer, monitoring-alerts-runbooks |
| **security** | 2 | penetration-tester, network-engineer |
| **data** | 4 | data-engineer, ml-engineer, mlops-engineer, data-scientist |
| **meta** | 6 | context-manager, meta-retro-analyzer, autonomous-worker |
| **specialized** | 6 | accessibility-specialist, ui-ux-designer, customer-support |

### All Agents

<details>
<summary>55 agents (click to expand)</summary>

**Languages (6)**
- golang-pro, java-pro, python-pro, rust-pro, shell-scripting-pro, typescript-pro

**Development (8)**
- ai-engineer, backend-architect, deployment-engineer, frontend-developer, fullstack-developer, ios-developer, mobile-developer, prompt-engineer

**Documentation (4)**
- api-documenter, documentation-create-docs, documentation-diataxis-auditor, documentation-optimize-docs

**Code Quality (3)**
- code-review-improve, code-reviewer, test-generator

**Research (6)**
- archive-researcher, code-explorer, doc-explorer, document-structure-analyzer, history-explorer, spec-architect

**Validation (4)**
- assumption-validator, continuous-validator, tracer-bullet-deployer, validation-planner

**Operations (4)**
- error-detective, incident-responder, incidents-postmortems, incidents-response

**Monitoring (2)**
- monitoring-alerts-runbooks, performance-engineer

**Security (2)**
- network-engineer, penetration-tester

**Data (4)**
- data-engineer, data-scientist, ml-engineer, mlops-engineer

**Meta (6)**
- autonomous-worker, change-executor, context-manager, meta-memory-manager, meta-observer, meta-retro-analyzer

**Specialized (6)**
- accessibility-specialist, connection-agent, customer-support, risk-assessor, task-decomposition-expert, ui-ux-designer

</details>

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

## Skills

| Skill | Purpose |
|-------|---------|
| **base** | Foundation patterns (7 audit/cleanup utilities) |
| **brand-guidelines** | Consistent documentation styling |
| **doc-curator** | Documentation quality management |
| **git-workflow** | Git best practices and automation |
| **skill-creator** | Create new skills from patterns |
| **test-gap-scanner** | Identify missing test coverage |
| **testing** | Test patterns and frameworks |

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
│   ├── learnings/        # Session insights
│   └── bundles/          # Context bundles
├── .beads/               # Git-based issue tracking
├── .claude-plugin/       # Plugin manifest
├── agents/               # 55 agent definitions
│   └── catalog.yaml      # Agent registry
├── commands/             # 28 slash commands
│   └── INDEX.md          # Command catalog
├── docs/standards/       # Coding standards
├── profiles/             # Role configurations
└── skills/               # Knowledge packs
```

## Related

- [Vibe Coding Book](https://itrevolution.com/product/vibe-coding-book/) - Gene Kim & Steve Yegge
- [Vibe Ecosystem](https://www.bodenfuller.com/builds/vibe-ecosystem) - Full documentation
- [vibe-check](https://www.npmjs.com/package/@boshu2/vibe-check) - Metrics CLI tool
- [bodenfuller.com/workflow](https://bodenfuller.com/workflow) - Video demos

## License

MIT
