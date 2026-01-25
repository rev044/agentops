# Skills Reference

Complete reference for all AgentOps skills.

## Core Workflow Skills

### /research

Deep codebase exploration using Explore agents.

```bash
/research authentication flows in services/auth
```

**Output:** `.agents/research/<topic>.md`

### /plan

Decompose goals into trackable beads issues with dependencies.

```bash
/plan "Add user authentication with OAuth2"
```

**Output:** Beads issues with parent/child relationships

### /implement

Execute a single beads issue with full lifecycle.

```bash
/implement ap-1234
```

**Phases:** Context → Tests → Code → Validation → Commit

### /crank

Autonomous multi-issue execution. Runs until epic is CLOSED.

```bash
/crank <epic-id>
```

**Modes:** Crew (sequential) or Mayor (parallel via polecats)

### /vibe

Comprehensive code validation across 8 aspects.

```bash
/vibe services/auth/
```

**Checks:** Security, Quality, Architecture, Complexity, Testing, Accessibility, Performance, Documentation

### /retro

Extract learnings from completed work.

```bash
/retro "debugging memory leak"
```

**Output:** `.agents/retros/` and `.agents/learnings/`

### /post-mortem

Full validation + knowledge extraction. Combines retro + vibe.

```bash
/post-mortem <epic-id>
```

**Output:** Retro, learnings, patterns, security scan

---

## Utility Skills

### /beads

Git-native issue tracking operations.

```bash
bd ready              # Unblocked issues
bd show <id>          # Issue details
bd close <id>         # Close issue
```

### /bug-hunt

Root cause analysis with git archaeology.

```bash
/bug-hunt "login fails after password reset"
```

### /knowledge

Query knowledge artifacts across locations.

```bash
/knowledge "patterns for rate limiting"
```

**Searches:** `.agents/learnings/`, `.agents/patterns/`, `.agents/research/`

### /complexity

Code complexity analysis using radon (Python) or gocyclo (Go).

```bash
/complexity services/
```

**Threshold:** CC > 10 triggers refactoring issue

### /doc

Generate documentation for code.

```bash
/doc services/auth/
```

### /pre-mortem

Simulate failures before implementing.

```bash
/pre-mortem "add caching layer"
```

**Output:** Failure modes, mitigation strategies, spec improvements

---

## Subagents

Located in `agents/`:

| Agent | Purpose |
|-------|---------|
| `code-reviewer` | Code review for quality and patterns |
| `security-reviewer` | Security vulnerability analysis |
| `architecture-expert` | System design review |
| `code-quality-expert` | Complexity and maintainability |
| `security-expert` | OWASP Top 10 focus |
| `ux-expert` | Accessibility and UX validation |

---

## ao CLI Integration

Skills integrate with the ao CLI for orchestration:

| Skill | ao CLI Command |
|-------|----------------|
| `/research` | `ao forge search` |
| `/retro` | `ao forge index` |
| `/post-mortem` | `ao ratchet record` |
| `/implement` | `ao ratchet claim/record` |
| `/crank` | `ao ratchet verify` |
