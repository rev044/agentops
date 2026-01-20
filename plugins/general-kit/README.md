# general-kit

Portable Claude Code skills that work out of the box - **no external dependencies required**.

## Install

```bash
claude plugin install general-kit@agentops-marketplace
```

## Skills

### Research & Exploration

| Skill | Purpose | Example |
|-------|---------|---------|
| `/research` | Deep codebase exploration | `/research how does auth work` |
| `/bug-hunt` | Git archaeology, root cause analysis | `/bug-hunt login fails on Safari` |

### Validation & Quality

| Skill | Purpose | Example |
|-------|---------|---------|
| `/vibe` | Comprehensive code validation | `/vibe src/auth/` |
| `/vibe-docs` | Validate docs match code | `/vibe-docs` |
| `/complexity` | Find refactoring targets | `/complexity services/` |
| `/validation-chain` | Multi-gate validation | `/validation-chain staged` |

### Documentation

| Skill | Purpose | Example |
|-------|---------|---------|
| `/doc` | Generate documentation | `/doc` |
| `/oss-docs` | Scaffold OSS docs | `/oss-docs` |
| `/golden-init` | Initialize repo structure | `/golden-init` |

## Expert Agents

Four specialized agents for deep analysis:

| Agent | Specialty |
|-------|-----------|
| `security-expert` | OWASP Top 10, vulnerability assessment |
| `architecture-expert` | System design, cross-cutting concerns |
| `code-quality-expert` | Complexity, patterns, maintainability |
| `ux-expert` | Accessibility, user experience |

**Usage:** Claude automatically delegates to these agents during `/vibe` checks, or you can invoke explicitly: "Use the security-expert agent to review this auth code"

## Quick Examples

```bash
# Explore a codebase
/research how does the API handle errors

# Validate your changes before commit
/vibe staged

# Find complex functions to refactor
/complexity src/

# Investigate a bug
/bug-hunt users can't reset password

# Generate project documentation
/doc

# Scaffold OSS files (README, CONTRIBUTING, etc.)
/oss-docs
```

## What's NOT Included

general-kit is intentionally minimal. For advanced workflows:

| Want | Install | Requires |
|------|---------|----------|
| Issue tracking | `beads-kit` | [beads](https://github.com/steveyegge/beads) CLI |
| Structured workflows | `core-kit` | beads |
| Multi-agent orchestration | `gastown-kit` | [gastown](https://github.com/steveyegge/gastown) CLI |

## Why general-kit?

- **Zero setup** - Works immediately after install
- **Standard git repos** - No special tooling required
- **Self-contained** - All skills work independently
- **Foundation** - Start here, add specialized kits as needed
