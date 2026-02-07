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

## Orchestration Skills

### /council

Multi-model validation — the core primitive used by vibe, pre-mortem, and post-mortem.

```bash
/council validate recent
/council --deep recent
```

### /swarm

Parallel agent spawning for concurrent task execution.

```bash
/swarm <epic-id>
```

### /codex-team

Spawn parallel Codex execution agents.

```bash
/codex-team <epic-id>
```

---

## Additional Utility Skills

### /handoff

Session handoff — preserve context for continuation.

```bash
/handoff
```

### /inbox

Agent mail monitoring for distributed mode.

```bash
/inbox
```

### /release

Pre-flight checks, changelog generation, version bumps, and tagging.

```bash
/release
```

### /status

Single-screen dashboard of project state.

```bash
/status
```

### /quickstart

Interactive onboarding — mini RPI cycle for new users.

```bash
/quickstart
```

### /trace

Trace design decisions through knowledge artifacts.

```bash
/trace "why did we choose Redis?"
```

---

## Subagents

Subagent behaviors are defined inline within SKILL.md files (not as separate agent files). Skills that use subagents spawn them as Task agents during execution. 20 specialized roles are used across `/vibe`, `/pre-mortem`, `/post-mortem`, and `/research`.

| Agent Role | Used By | Focus |
|------------|---------|-------|
| Code reviewer | /vibe, /council | Quality, patterns, maintainability |
| Security reviewer | /vibe, /council | Vulnerabilities, OWASP |
| Security expert | /vibe, /council | Deep security analysis |
| Architecture expert | /vibe, /council | System design review |
| Code quality expert | /vibe, /council | Complexity and maintainability |
| UX expert | /vibe, /council | Accessibility and UX validation |
| Plan compliance expert | /post-mortem | Compare implementation to plan |
| Goal achievement expert | /post-mortem | Did we solve the problem? |
| Ratchet validator | /post-mortem | Verify gates are locked |
| Flywheel feeder | /post-mortem | Extract learnings with provenance |
| Technical learnings expert | /post-mortem | Technical patterns |
| Process learnings expert | /post-mortem | Process improvements |
| Integration failure expert | /pre-mortem | Integration risks |
| Ops failure expert | /pre-mortem | Operational risks |
| Data failure expert | /pre-mortem | Data integrity risks |
| Edge case hunter | /pre-mortem | Edge cases and exceptions |
| Coverage expert | /research | Research completeness |
| Depth expert | /research | Depth of analysis |
| Gap identifier | /research | Missing areas |
| Assumption challenger | /research | Challenge assumptions |

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
