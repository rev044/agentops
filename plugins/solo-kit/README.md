# Solo Kit

> Essential skills for any developer, any project.

The foundation kit for AgentOps. Works with any language, any codebase.

## Install

```bash
/plugin install solo-kit@agentops
```

## What's Included

### Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/research` | `/research <topic>` | Deep codebase exploration |
| `/vibe` | `/vibe [target]` | Comprehensive code validation |
| `/bug-hunt` | `/bug-hunt` | Git archaeology for root cause |
| `/complexity` | `/complexity` | Find refactoring targets |
| `/doc` | `/doc` | Generate documentation |
| `/oss-docs` | `/oss-docs` | OSS scaffolding (README, etc.) |
| `/golden-init` | `/golden-init` | Initialize repo with best practices |

### Agents

| Agent | Purpose | Invocation |
|-------|---------|------------|
| `code-reviewer` | Quality review (read-only) | Auto or explicit request |
| `security-reviewer` | Security scan (read-only) | Auto or explicit request |

### Hooks (Auto-Enabled)

| Hook | Trigger | What It Does |
|------|---------|--------------|
| Git push review | `git push` | Pause to review changes |
| Console.log warning | Edit JS/TS | Warn about debug statements |
| Print warning | Edit Python | Note print() usage |
| Auto-format | Edit any | Run formatter if available |
| Shellcheck | Edit shell | Validate shell scripts |
| PR logging | `gh pr create` | Log PR URL |
| Debug audit | Session end | Final check for debug code |

## Quick Start

```bash
# Research first
/research "authentication flow"

# Validate your work
/vibe src/

# Hunt down a bug
/bug-hunt "login fails intermittently"

# Find complexity hotspots
/complexity
```

## Philosophy

**Language-agnostic.** Works whether you're writing Python, Go, TypeScript, Rust, or shell scripts.

**Essential only.** 7 skills, 2 agents. No bloat.

**Hooks that help.** Auto-format, warn about debug code, catch issues before commit.

## Add Language Support

For language-specific standards and deeper validation:

```bash
/plugin install python-kit@agentops    # Python
/plugin install go-kit@agentops        # Go
/plugin install typescript-kit@agentops # TypeScript
/plugin install shell-kit@agentops     # Shell/Bash
```

Language kits add:
- Language-specific standards
- Specialized linting hooks
- Testing patterns
- Common error detection

## Upgrade Path

```
solo-kit              # Any developer
  ↓
+ beads-kit           # Track work across sessions
  ↓
+ pr-kit              # PR workflows
  ↓
+ crank-kit           # Autonomous execution
  ↓
+ gastown-kit         # Multi-agent orchestration
```

## Requirements

- Claude Code CLI
- Git (for git-based features)
- Optional: prettier, ruff, gofmt, shellcheck (for auto-format/lint hooks)

## License

MIT
