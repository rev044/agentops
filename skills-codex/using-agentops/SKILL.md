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
$knowledge <query>     # Query existing knowledge
ao lookup <id>         # Pull full content of specific learning
ao lookup --query "x"  # Search knowledge by relevance
```

**Output:** `.agents/research/<topic>.md`

### Plan Phase

```bash
$pre-mortem <spec>     # Simulate failures before implementing
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
$vibe [target]         # Code validation (security, quality, architecture)
$post-mortem           # Extract learnings after completion
$retro                 # Quick retrospective
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
| **Research** | `$research` | `$knowledge`, `$inject` |
| **Plan** | `$plan` | `$pre-mortem` |
| **Implement** | `$implement` | `$crank` (epic loop), `$swarm` (parallel execution) |
| **Validate** | `$vibe` | `$retro`, `$post-mortem` |
| **Release** | `$release` | — |

**Choosing the skill:**
- Use `$implement` for **single issue** execution.
- Use `$crank` for **autonomous epic execution** (loops waves via swarm until done).
- Use `$swarm` directly for **parallel execution** without beads (TaskList only).
- Use `$ratchet` to **gate/record progress** through RPI.

## Available Skills

## Start Here (11 starters)

These are the skills every user needs first. Everything else is available when you need it.

| Skill | Purpose |
|-------|---------|
| `$quickstart` | Guided onboarding — run this first |
| `$research` | Deep codebase exploration |
| `$council` | Multi-model consensus review (validate, brainstorm, research) |
| `$vibe` | Code validation (complexity + multi-model council) |
| `$rpi` | Full RPI lifecycle orchestrator (research → plan → implement → validate) |
| `$implement` | Execute single issue |
| `$learn` | Capture knowledge manually into the flywheel |
| `$status` | Single-screen dashboard of current work and suggested next action |
| `$goals` | Maintain GOALS.yaml fitness specification |
| `$flywheel` | Knowledge flywheel health monitoring (σ×ρ > δ) |
| `$inbox` | Agent mail inbox — check pending messages and HELP_REQUESTs |

## Advanced Skills (when you need them)

| Skill | Purpose |
|-------|---------|
| `$brainstorm` | Structured idea exploration before planning |
| `$plan` | Epic decomposition into issues |
| `$pre-mortem` | Failure simulation before implementing |
| `$post-mortem` | Full validation + knowledge extraction |
| `$bug-hunt` | Root cause analysis |
| `$release` | Pre-flight, changelog, version bumps, tag |
| `$crank` | Autonomous epic loop (uses swarm for each wave) |
| `$swarm` | Fresh-context parallel execution (Ralph pattern) |
| `$evolve` | Goal-driven fitness-scored improvement loop |
| `$doc` | Documentation generation |
| `$retro` | Extract learnings from completed work |
| `$knowledge` | Query knowledge artifacts |
| `$ratchet` | Brownian Ratchet progress gates for RPI workflow |
| `$extract` | Extract decisions and learnings from session transcripts |
| `$forge` | Mine transcripts for knowledge — decisions, learnings, patterns |
| `$readme` | Generate gold-standard README for any project |
| `$security` | Continuous repository security scanning and release gating |
| `$security-suite` | Binary security suite — static analysis, dynamic tracing, policy gating |

## Expert Skills (specialized workflows)

| Skill | Purpose |
|-------|---------|
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
bd sync               # Sync with git
```

## Examples

### SessionStart Context Loading

**Hook triggers:** `session-start.sh` runs at session start

**What happens:**
1. In `manual` mode (default): MEMORY.md is auto-loaded by Codex; hook emits a pointer to on-demand retrieval (`ao search`, `ao lookup`)
2. In `lean`/`legacy` modes: hook extracts pending knowledge and injects prior learnings
3. Hook injects this skill automatically into session context
4. Agent loads RPI workflow overview, phase-to-skill mapping, trigger patterns
5. User says "check my code" → agent recognizes `$vibe` trigger naturally

**Result:** Agent knows the full skill catalog and workflow from session start. Knowledge retrieval is on-demand by default (MEMORY.md + `ao search`), with opt-in startup injection via `AGENTOPS_STARTUP_CONTEXT_MODE=lean`.

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

---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0
check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "name is using-agentops" "grep -q '^name: using-agentops' '$SKILL_DIR/SKILL.md'"
check "mentions RPI or workflow" "grep -qiE 'RPI|workflow' '$SKILL_DIR/SKILL.md'"
check "mentions Knowledge Flywheel" "grep -qi 'Knowledge Flywheel' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


