# Skill Tier Taxonomy

This document defines the `tier` field used in skill frontmatter to categorize skills by their role in the AgentOps workflow.

## Tier Values

| Tier | Description | Examples |
|------|-------------|----------|
| **solo** | Standalone skills invoked directly by users | research, plan, vibe, implement |
| **library** | Reference skills loaded JIT by other skills | beads, standards |
| **orchestration** | Multi-skill coordinators that run other skills | crank, council |
| **team** | Skills requiring human collaboration | implement (guided mode) |
| **background** | Hook-triggered or automatic skills | inject, forge, extract |
| **meta** | Skills about skills (documentation, validation) | using-agentops |

## Current Skill Tiers

### User-Facing Skills (42)

| Skill | Tier | Description |
|-------|------|-------------|
| **council** | orchestration | Multi-model validation (core primitive) |
| **crank** | orchestration | Autonomous epic execution |
| **swarm** | orchestration | Parallel agent spawning |
| **codex-team** | orchestration | Spawn parallel Codex execution agents |
| **openai-docs** | solo | Authoritative OpenAI docs lookup with citations |
| **rpi** | orchestration | Full RPI lifecycle orchestrator (research → post-mortem) |
| **evolve** | orchestration | Autonomous fitness-scored improvement loop |
| **pr-research** | solo | Upstream repository research before contribution |
| **pr-plan** | solo | Contribution planning for external PRs |
| **pr-implement** | solo | Fork-based implementation for external PRs |
| **pr-validate** | solo | PR-specific isolation and scope validation |
| **pr-prep** | solo | PR preparation and structured PR body generation |
| **pr-retro** | solo | Learn from accepted/rejected PR outcomes |
| **oss-docs** | solo | Scaffold and audit OSS documentation packs |
| **implement** | team | Execute single issue |
| **quickstart** | solo | Interactive onboarding (mini RPI cycle) |
| **status** | solo | Single-screen dashboard |
| **research** | solo | Deep codebase exploration |
| **plan** | solo | Decompose epics into issues |
| **vibe** | solo | Complexity + council (validate code) |
| **pre-mortem** | solo | Council on plans |
| **post-mortem** | solo | Council + retro (wrap up work) |
| **retro** | solo | Extract learnings |
| **complexity** | solo | Cyclomatic analysis |
| **knowledge** | solo | Query knowledge artifacts |
| **bug-hunt** | solo | Investigate bugs |
| **doc** | solo | Generate documentation |
| **handoff** | solo | Session handoff |
| **inbox** | solo | Agent mail monitoring |
| **release** | solo | Pre-flight, changelog, version bumps, tag |
| **security** | solo | Continuous security scanning and release gating |
| **product** | solo | Interactive PRODUCT.md generation |
| **recover** | solo | Post-compaction context recovery |
| **trace** | solo | Trace design decisions |
| **learn** | solo | Manual knowledge capture |
| **reverse-engineer-rpi** | research | Reverse-engineer a product into feature catalog + code map + specs |
| **readme** | solo | Gold-standard README generation with council validation |
| **brainstorm** | solo | Structured idea exploration before planning |
| **heal-skill** | solo | Detect and fix skill hygiene issues |
| **converter** | solo | Cross-platform skill converter (Codex, Cursor) |
| **goals** | solo | Maintain GOALS.yaml fitness specification |
| **update** | solo | Reinstall all AgentOps skills globally |

### Internal Skills (10) — `metadata.internal: true`

These are hidden from interactive `npx skills add` discovery. They are loaded JIT
by other skills via Read or auto-triggered by hooks. Not intended for direct user invocation.

| Skill | Tier | Purpose |
|-------|------|---------|
| beads | library | Issue tracking reference (loaded by /implement, /plan) |
| standards | library | Coding standards (loaded by /vibe, /implement, /doc) |
| shared | library | Shared reference documents (distributed mode) |
| inject | background | Load knowledge at session start (hook-triggered) |
| extract | background | Extract from transcripts (hook-triggered) |
| forge | background | Mine transcripts for knowledge |
| provenance | background | Trace knowledge lineage |
| ratchet | background | Progress gates |
| flywheel | background | Knowledge health monitoring |
| using-agentops | meta | AgentOps workflow guide (auto-injected) |

---

## Skill Dependency Graph

### Core Primitive: /council

All validation skills depend on `/council`:

```
                         ┌──────────┐
                         │ council  │  ← Multi-model judgment
                         └────┬─────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
  ┌────────────┐        ┌─────────┐         ┌─────────────┐
  │ pre-mortem │        │  vibe   │         │ post-mortem │
  │ (plans)    │        │ (code)  │         │ (wrap up)   │
  └────────────┘        └────┬────┘         └──────┬──────┘
                             │                     │
                             ▼                     ▼
                       ┌────────────┐         ┌─────────┐
                       │ complexity │         │  retro  │
                       └────────────┘         └─────────┘
```

### Dependency Table

| Skill | Dependencies | Type |
|-------|--------------|------|
| **council** | - | - (core primitive) |
| **vibe** | council, complexity, standards | required, optional (graceful skip), optional |
| **pre-mortem** | council | required |
| **post-mortem** | council, retro, beads | required, optional (graceful skip), optional |
| beads | - | - |
| bug-hunt | beads | optional |
| complexity | - | - |
| **codex-team** | - | - (standalone, fallback to swarm) |
| **crank** | swarm, vibe, implement, beads, post-mortem | required, required, required, optional, optional |
| doc | standards | required |
| extract | - | - |
| flywheel | - | - |
| forge | - | - |
| handoff | retro | optional |
| **implement** | beads, standards | optional, required |
| inbox | - | - |
| inject | - | - |
| knowledge | - | - |
| **openai-docs** | - | - (standalone) |
| **plan** | research, beads, pre-mortem, crank, implement | optional, optional, optional, optional, optional |
| **product** | - | - (standalone) |
| **pr-research** | - | - (standalone) |
| **pr-plan** | pr-research | optional |
| **pr-implement** | pr-plan, pr-validate | optional, optional |
| **pr-validate** | - | - (standalone) |
| **pr-prep** | pr-validate | optional |
| **pr-retro** | pr-prep | optional |
| **oss-docs** | doc | optional |
| provenance | - | - |
| **quickstart** | - | - (zero dependencies) |
| **rpi** | research, plan, pre-mortem, crank, vibe, post-mortem, ratchet | all required |
| **evolve** | rpi | required (rpi pulls in all sub-skills) |
| **release** | - | - (standalone) |
| **security** | - | - (standalone) |
| ratchet | - | - |
| **recover** | - | - (standalone) |
| **reverse-engineer-rpi** | - | - (standalone, research tier) |
| research | knowledge, inject | optional, optional |
| retro | - | - |
| standards | - | - |
| **goals** | - | - (reads GOALS.yaml directly) |
| **status** | - | - (all CLIs optional) |
| **swarm** | implement, vibe | required, optional |
| trace | provenance | alternative |
| **update** | - | - (standalone) |
| using-agentops | - | - |

### RPI Workflow

```
RESEARCH          PLAN              IMPLEMENT           VALIDATE
────────          ────              ─────────           ────────

┌──────────┐    ┌──────────┐      ┌───────────┐      ┌──────────┐
│ research │───►│   plan   │─────►│ implement │─────►│   vibe   │
└──────────┘    └────┬─────┘      └─────┬─────┘      └────┬─────┘
                     │                  │                 │
                     ▼                  │                 │
               ┌────────────┐           │                 │
               │ pre-mortem │           │                 │
               │ (council)  │           │                 │
               └────────────┘           │                 │
                                        │                 │
                                        ▼                 ▼
                                   ┌─────────┐      ┌───────────┐
                                   │  swarm  │      │complexity │
                                   └────┬────┘      │ + council │
                                        │          └───────────┘
                                        ▼
                                   ┌─────────┐
                                   │  crank  │
                                   └─────────┘

POST-SHIP                             ONBOARDING / STATUS
─────────                             ───────────────────

┌─────────────┐                       ┌────────────┐
│ post-mortem │                       │ quickstart │ (first-time tour)
│ (council +  │                       └────────────┘
│   retro)    │                       ┌────────────┐
└──────┬──────┘                       │   status   │ (dashboard)
       │                              └────────────┘
       ▼
┌─────────────┐
│   release   │ (changelog, version bump, tag)
└─────────────┘
```

### Knowledge Flywheel

```
┌─────────┐     ┌─────────┐     ┌──────────┐     ┌──────────┐
│ extract │────►│  forge  │────►│ knowledge│────►│  inject  │
└─────────┘     └─────────┘     └──────────┘     └──────────┘
     ▲                                                 │
     │              ┌──────────┐                       │
     └──────────────│ flywheel │◄──────────────────────┘
                    └──────────┘

Supporting: provenance, trace, ratchet
```

---

## CLI Integration

### Spawning Agents

| Vendor | CLI | Command |
|--------|-----|---------|
| Claude | `claude` | `claude --print "prompt" > output.md` |
| Codex | `codex` | `codex exec --full-auto -m gpt-5.3-codex -C "$(pwd)" -o output.md "prompt"` |
| OpenCode | `opencode` | (similar pattern) |

### Default Models

| Vendor | Model |
|--------|-------|
| Claude | Opus 4.6 |
| Codex/OpenAI | GPT-5.3-Codex |

### /council spawns both

```bash
# Runtime-native judges (spawn via whatever multi-agent primitive your runtime provides)
# Each judge receives a prompt, writes output to .agents/council/, signals completion

# Codex CLI judges (--mixed mode, via shell)
codex exec --full-auto -m gpt-5.3-codex -C "$(pwd)" -o .agents/council/codex-output.md "..."
```

### Consolidated Output

All council-based skills write to `.agents/council/`:

| Skill / Mode | Output Pattern |
|--------------|----------------|
| `/council validate` | `.agents/council/YYYY-MM-DD-<target>-report.md` |
| `/council brainstorm` | `.agents/council/YYYY-MM-DD-brainstorm-<topic>.md` |
| `/council research` | `.agents/council/YYYY-MM-DD-research-<topic>.md` |
| `/vibe` | `.agents/council/YYYY-MM-DD-vibe-<target>.md` |
| `/pre-mortem` | `.agents/council/YYYY-MM-DD-pre-mortem-<topic>.md` |
| `/post-mortem` | `.agents/council/YYYY-MM-DD-post-mortem-<topic>.md` |

Individual judge outputs also go to `.agents/council/`:
- `YYYY-MM-DD-<target>-claude-pragmatist.md`, `...-claude-skeptic.md`, `...-claude-visionary.md`
- `YYYY-MM-DD-<target>-codex-pragmatist.md`, `...-codex-skeptic.md`, `...-codex-visionary.md`

---

---

## See Also

- `skills/council/SKILL.md` — Core validation primitive
- `skills/vibe/SKILL.md` — Complexity + council for code
- `skills/pre-mortem/SKILL.md` — Council for plans
- `skills/post-mortem/SKILL.md` — Council + retro for wrap-up
