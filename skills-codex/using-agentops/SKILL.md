---
name: using-agentops
description: 'Meta skill explaining the RPI workflow. Auto-injected on session start. Covers Research-Plan-Implement workflow, Knowledge Flywheel, and skill catalog.'
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
$research <topic>      # Deep codebase exploration
ao search "<query>"    # Search existing knowledge
ao lookup <id>         # Pull full content of specific learning
ao lookup --query "x"  # Search knowledge by relevance
```

**Output:** `.agents/research/<topic>.md`

### Plan Phase

```bash
$pre-mortem <spec>     # Simulate failures (error/rescue map, scope modes, prediction tracking)
$plan <goal>           # Decompose into trackable issues
```

**Output:** Beads issues with dependencies

### Implement Phase

```bash
$implement <issue>     # Single issue execution
$crank <epic>          # Autonomous epic loop (uses swarm for waves)
$swarm                 # Parallel execution (fresh context per agent)
```

**Output:** Code changes, tests, documentation

### Validate Phase

```bash
$vibe [target]         # Code validation (finding classification + suppression + domain checklists)
$post-mortem           # Validation + streak tracking + prediction accuracy + retro history
$retro                 # Quick-capture a single learning
```

**Output:** `.agents/learnings/`, `.agents/patterns/`

### Release Phase

```bash
$release [version]     # Full release: changelog + bump + commit + tag
$release --check       # Readiness validation only (GO/NO-GO)
$release --dry-run     # Preview without writing
```

**Output:** Updated CHANGELOG.md, version bumps, git tag, `.agents/releases/`

## Phase-to-Skill Mapping

| Phase | Primary Skill | Supporting Skills |
|-------|---------------|-------------------|
| **Discovery** | `$discovery` | `$brainstorm`, `$research`, `$plan`, `$pre-mortem` |
| **Implement** | `$crank` | `$implement` (single issue), `$swarm` (parallel execution) |
| **Validate** | `$validation` | `$vibe`, `$post-mortem`, `$retro`, `$forge` |
| **Release** | `$release` | — |

**Choosing the skill:**
- Use `$implement` for **single issue** execution. **Now defaults to TDD-first** — writes failing tests before implementing. Skip with `--no-tdd`.
- Use `$crank` for **autonomous epic execution** (loops waves via swarm until done). Auto-generates file-ownership maps to prevent worker conflicts.
- Use `$discovery` for the **discovery phase only** (brainstorm → search → research → plan → pre-mortem).
- Use `$validation` for the **validation phase only** (vibe → post-mortem → retro → forge).
- Use `$rpi` for **full lifecycle** — delegates to `$discovery` → `$crank` → `$validation`.
- Use `$ratchet` to **gate/record progress** through RPI.

## Available Skills

## Start Here (12 starters)

These are the skills every user needs first. Everything else is available when you need it.

| Skill | Purpose |
|-------|---------|
| `$quickstart` | Guided onboarding — run this first |
| `$research` | Deep codebase exploration |
| `$council` | Multi-model consensus review + finding auto-extraction |
| `$vibe` | Code validation (classification + suppression + domain checklists) |
| `$rpi` | Full RPI lifecycle orchestrator (research → plan → implement → validate) |
| `$implement` | Execute single issue |
| `$retro --quick` | Quick-capture a single learning into the flywheel |
| `$status` | Single-screen dashboard of current work and suggested next action |
| `$goals` | Maintain GOALS.yaml fitness specification |
| `$push` | Atomic test-commit-push workflow |
| `$flywheel` | Knowledge flywheel health monitoring (σ×ρ > δ) |

## Advanced Skills (when you need them)

| Skill | Purpose |
|-------|---------|
| `$athena` | Active knowledge intelligence — Mine → Grow → Defrag cycle |
| `$brainstorm` | Structured idea exploration before planning |
| `$discovery` | Full discovery phase orchestrator (brainstorm → search → research → plan → pre-mortem) |
| `$plan` | Epic decomposition into issues |
| `$pre-mortem` | Failure simulation (error/rescue, scope modes, temporal, predictions) |
| `$post-mortem` | Validation + streak tracking + prediction accuracy + retro history |
| `$bug-hunt` | Root cause analysis |
| `$release` | Pre-flight, changelog, version bumps, tag |
| `$crank` | Autonomous epic loop (uses swarm for each wave) |
| `$swarm` | Fresh-context parallel execution (Ralph pattern) |
| `$evolve` | Goal-driven fitness-scored improvement loop |
| `$doc` | Documentation generation |
| `$retro` | Quick-capture a learning (full retro → $post-mortem) |
| `$validation` | Full validation phase orchestrator (vibe → post-mortem → retro → forge) |
| `$ratchet` | Brownian Ratchet progress gates for RPI workflow |
| `$forge` | Mine transcripts for knowledge — decisions, learnings, patterns |
| `$readme` | Generate gold-standard README for any project |
| `$security` | Continuous repository security scanning and release gating |
| `$security-suite` | Binary and prompt-surface security suite — static analysis, dynamic tracing, offline redteam, policy gating |

## Expert Skills (specialized workflows)

| Skill | Purpose |
|-------|---------|
| `$grafana-platform-dashboard` | Build Grafana platform dashboards from templates/contracts |
| `$codex-team` | Parallel Codex agent execution |
| `$openai-docs` | Official OpenAI docs lookup with citations |
| `$oss-docs` | OSS documentation scaffold and audit |
| `$reverse-engineer-rpi` | Reverse-engineer a product into feature catalog and specs |
| `$pr-research` | Upstream repository research before contribution |
| `$pr-plan` | External contribution planning |
| `$pr-implement` | Fork-based PR implementation |
| `$pr-validate` | PR-specific validation and isolation checks |
| `$pr-prep` | PR preparation and structured body generation |
| `$pr-retro` | Learn from PR outcomes |
| `$complexity` | Code complexity analysis |
| `$product` | Interactive PRODUCT.md generation |
| `$handoff` | Session handoff for continuation |
| `$recover` | Post-compaction context recovery |
| `$trace` | Trace design decisions through history |
| `$provenance` | Trace artifact lineage to sources |
| `$beads` | Issue tracking operations |
| `$heal-skill` | Detect and fix skill hygiene issues |
| `$converter` | Convert skills to Codex/Cursor formats |
| `$update` | Reinstall all AgentOps skills from latest source |

## Knowledge Flywheel

Every `$post-mortem` feeds back to `$research`:

1. **Learnings** extracted → `.agents/learnings/`
2. **Patterns** discovered → `.agents/patterns/`
3. **Research** enriched → Future sessions benefit

## Issue Tracking

This workflow uses beads for git-native issue tracking:

```bash
bd ready              # Unblocked issues
bd show <id>          # Issue details
bd close <id>         # Close issue
bd vc status          # Inspect Dolt state if needed (JSONL auto-sync is automatic)
```

## Examples

### SessionStart Context Loading

**Hook triggers:** `session-start.sh` runs at session start

**What happens:**
1. In `manual` mode (default): MEMORY.md is auto-loaded by Codex; hook emits a pointer to on-demand retrieval (`ao search`, `ao lookup`)
2. In `lean` mode: hook extracts pending knowledge and injects prior learnings with a reduced token budget
3. Hook injects this skill automatically into session context
4. Agent loads RPI workflow overview, phase-to-skill mapping, trigger patterns
5. User says "check my code" → agent recognizes `$vibe` trigger naturally

**Result:** Agent knows the full skill catalog and workflow from session start. MEMORY.md is auto-loaded by default (`manual` mode). Set `AGENTOPS_STARTUP_CONTEXT_MODE=lean` for automatic knowledge injection alongside MEMORY.md.

### Workflow Reference During Planning

**User says:** "How should I approach this feature?"

**What happens:**
1. Agent references this skill's RPI workflow section
2. Agent recommends Research → Plan → Implement → Validate phases
3. Agent suggests `$research` for codebase exploration, `$plan` for decomposition
4. Agent explains `$pre-mortem` for failure simulation before implementation
5. User follows recommended workflow with agent guidance

**Result:** Agent provides structured workflow guidance based on this meta-skill, avoiding ad-hoc approaches.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Skill not auto-loaded | Hook not configured or SessionStart disabled | Verify hooks/session-start.sh exists; check hook enable flags |
| Outdated skill catalog | This file not synced with actual skills/ directory | Update skill list in this file after adding/removing skills |
| Wrong skill suggested | Natural language trigger ambiguous | User explicitly calls skill with `/skill-name` syntax |
| Workflow unclear | RPI phases not well-documented here | Read full workflow guide in README.md or docs/ARCHITECTURE.md |

## Local Resources

### scripts/

- `scripts/validate.sh`


