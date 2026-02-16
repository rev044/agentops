---
name: using-agentops
description: 'Meta skill explaining the RPI workflow. Auto-injected on session start. Covers Research-Plan-Implement workflow, Knowledge Flywheel, and skill catalog.'
metadata:
  tier: meta
  dependencies: []
  internal: true
---

# RPI Workflow

You have access to workflow skills for structured development.

## The RPI Workflow

```
Research → Plan → Implement → Validate
    ↑                            │
    └──── Knowledge Flywheel ────┘
```

### Research Phase

```bash
/research <topic>      # Deep codebase exploration
/knowledge <query>     # Query existing knowledge
```

**Output:** `.agents/research/<topic>.md`

### Plan Phase

```bash
/pre-mortem <spec>     # Simulate failures before implementing
/plan <goal>           # Decompose into trackable issues
```

**Output:** Beads issues with dependencies

### Implement Phase

```bash
/implement <issue>     # Single issue execution
/crank <epic>          # Autonomous epic loop (uses swarm for waves)
/swarm                 # Parallel execution (fresh context per agent)
```

**Output:** Code changes, tests, documentation

### Validate Phase

```bash
/vibe [target]         # Code validation (security, quality, architecture)
/post-mortem           # Extract learnings after completion
/retro                 # Quick retrospective
```

**Output:** `.agents/learnings/`, `.agents/patterns/`

### Release Phase

```bash
/release [version]     # Full release: changelog + bump + commit + tag
/release --check       # Readiness validation only (GO/NO-GO)
/release --dry-run     # Preview without writing
```

**Output:** Updated CHANGELOG.md, version bumps, git tag, `.agents/releases/`

## Phase-to-Skill Mapping

| Phase | Primary Skill | Supporting Skills |
|-------|---------------|-------------------|
| **Research** | `/research` | `/knowledge`, `/inject` |
| **Plan** | `/plan` | `/pre-mortem` |
| **Implement** | `/implement` | `/crank` (epic loop), `/swarm` (parallel execution) |
| **Validate** | `/vibe` | `/retro`, `/post-mortem` |
| **Release** | `/release` | — |

**Choosing the skill:**
- Use `/implement` for **single issue** execution.
- Use `/crank` for **autonomous epic execution** (loops waves via swarm until done).
- Use `/swarm` directly for **parallel execution** without beads (TaskList only).
- Use `/ratchet` to **gate/record progress** through RPI.

## Available Skills (26 user-facing)

| Skill | Purpose |
|-------|---------|
| `/quickstart` | Interactive onboarding — first RPI cycle in under 10 minutes |
| `/status` | Single-screen dashboard of current work and suggested next action |
| `/research` | Deep codebase exploration |
| `/plan` | Epic decomposition into issues |
| `/pre-mortem` | Failure simulation before implementing |
| `/implement` | Execute single issue |
| `/crank` | Autonomous epic loop (uses swarm for each wave) |
| `/swarm` | Fresh-context parallel execution (Ralph pattern) |
| `/rpi` | Full RPI lifecycle orchestrator (research → plan → implement → validate) |
| `/evolve` | Goal-driven fitness-scored improvement loop |
| `/vibe` | Code validation (complexity + multi-model council) |
| `/council` | Multi-model consensus review (validate, brainstorm, research) |
| `/codex-team` | Parallel Codex agent execution |
| `/retro` | Extract learnings from completed work |
| `/post-mortem` | Full validation + knowledge extraction |
| `/release` | Pre-flight, changelog, version bumps, tag |
| `/bug-hunt` | Root cause analysis |
| `/complexity` | Code complexity analysis |
| `/doc` | Documentation generation |
| `/product` | Interactive PRODUCT.md generation |
| `/knowledge` | Query knowledge artifacts |
| `/beads` | Issue tracking operations |
| `/handoff` | Session handoff for continuation |
| `/inbox` | Agent mail monitoring |
| `/recover` | Post-compaction context recovery |
| `/trace` | Trace design decisions through history |
| `/provenance` | Trace artifact lineage to sources |
| `/update` | Reinstall all AgentOps skills from latest source |

## Knowledge Flywheel

Every `/post-mortem` feeds back to `/research`:

1. **Learnings** extracted → `.agents/learnings/`
2. **Patterns** discovered → `.agents/patterns/`
3. **Research** enriched → Future sessions benefit

## Natural Language Triggers

Skills auto-trigger from conversation:

| Say This | Runs |
|----------|------|
| "I need to understand how auth works" | `/research` |
| "Check my code for issues" | `/vibe` |
| "What could go wrong with this?" | `/pre-mortem` |
| "Let's execute this epic" | `/crank` |
| "Spawn agents to work in parallel" | `/swarm` |
| "How did we decide on this?" | `/trace` |
| "Where did this learning come from?" | `/provenance` |
| "Cut a release" | `/release` |
| "Are we ready to release?" | `/release --check` |
| "What am I working on?" | `/status` |
| "Get started" / "How do I start?" | `/quickstart` |
| "Define the product" | `/product` |
| "Run the full lifecycle" | `/rpi` |
| "Improve toward goals" | `/evolve` |
| "Where was I?" / "Lost context" | `/recover` |
| "End session" / "Pick up later" | `/handoff` |
| "Update skills" | `/update` |

## Issue Tracking

This workflow uses beads for git-native issue tracking:

```bash
bd ready              # Unblocked issues
bd show <id>          # Issue details
bd close <id>         # Close issue
bd sync               # Sync with git
```

## Examples

### SessionStart Auto-Injection

**Hook triggers:** `session-start.sh` runs at session start

**What happens:**
1. Hook injects this skill automatically into session context
2. Agent loads RPI workflow overview, phase-to-skill mapping, trigger patterns
3. Agent understands available skills without user prompting
4. User says "check my code" → agent recognizes `/vibe` trigger naturally
5. Agent executes skill using workflow knowledge from this reference

**Result:** Agent knows the full skill catalog and workflow from session start, enabling natural language skill invocation.

### Workflow Reference During Planning

**User says:** "How should I approach this feature?"

**What happens:**
1. Agent references this skill's RPI workflow section
2. Agent recommends Research → Plan → Implement → Validate phases
3. Agent suggests `/research` for codebase exploration, `/plan` for decomposition
4. Agent explains `/pre-mortem` for failure simulation before implementation
5. User follows recommended workflow with agent guidance

**Result:** Agent provides structured workflow guidance based on this meta-skill, avoiding ad-hoc approaches.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Skill not auto-loaded | Hook not configured or SessionStart disabled | Verify hooks/session-start.sh exists; check hook enable flags |
| Outdated skill catalog | This file not synced with actual skills/ directory | Update skill list in this file after adding/removing skills |
| Wrong skill suggested | Natural language trigger ambiguous | User explicitly calls skill with `/skill-name` syntax |
| Workflow unclear | RPI phases not well-documented here | Read full workflow guide in README.md or docs/ARCHITECTURE.md |
