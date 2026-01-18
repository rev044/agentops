# Vibe Kit

Validation and quality assurance. 5 skills for vibe checks, bug hunting, and complexity analysis.

## Install

```bash
/plugin install vibe-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/vibe` | `/vibe [target]` | Comprehensive code validation |
| `/vibe-docs` | `/vibe-docs` | Validate documentation claims |
| `/validation-chain` | `/validation-chain` | Multi-stage validation pipeline |
| `/bug-hunt` | `/bug-hunt` | Git archaeology for root cause |
| `/complexity` | `/complexity` | Identify refactoring targets |

## When to Use Which

| Scenario | Skill | Why |
|----------|-------|-----|
| Before merging | `/vibe` | Catches issues early |
| After docs update | `/vibe-docs` | Ensures docs match reality |
| Security-critical code | `/validation-chain` | Full pipeline validation |
| Mysterious bug | `/bug-hunt` | Git history reveals cause |
| Before refactoring | `/complexity` | Find highest-impact targets |

## Expert Agents

This kit includes 4 expert agents for specialized validation:

| Agent | Focus |
|-------|-------|
| `security-expert` | OWASP Top 10, vulnerability assessment |
| `architecture-expert` | System design, cross-cutting concerns |
| `code-quality-expert` | Complexity, patterns, maintainability |
| `ux-expert` | Accessibility, UX, user-facing |

These agents are invoked automatically by `/vibe` when relevant.

## Examples

### Pre-merge validation

```bash
/vibe src/auth/
# Runs security, code quality, and architecture checks
```

### Bug investigation

```bash
/bug-hunt "login fails intermittently"
# Uses git blame, bisect, and log analysis to find root cause
```

### Find refactoring opportunities

```bash
/complexity
# Reports cyclomatic complexity, suggests refactor targets
```

## The 40% Rule

The most important principle:

- **Below 40% context** → 98% success rate
- **Above 60% context** → 24% success rate

Stay under 40% for reliable validation results.

## Philosophy

- **Validate before shipping** - catch issues early
- **Bug hunting with git archaeology** - history tells the story
- **Complexity analysis guides refactoring** - data-driven decisions

## Related Kits

- **core-kit** - Research and implementation
- **docs-kit** - Generate the docs that vibe-docs validates

---

*Less is more.*
