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

| Skill | Tier | Description |
|-------|------|-------------|
| **council** | orchestration | Multi-model validation (core primitive) |
| beads | library | Issue tracking reference |
| standards | library | Coding standards reference |
| shared | library | Shared reference documents |
| crank | orchestration | Autonomous epic execution |
| swarm | orchestration | Parallel agent spawning |
| implement | team | Execute single issue |
| research | solo | Deep codebase exploration |
| plan | solo | Decompose epics into issues |
| **vibe** | solo | Complexity + council (validate code) |
| **pre-mortem** | solo | Council on plans |
| **post-mortem** | solo | Council + retro (wrap up work) |
| retro | solo | Extract learnings |
| complexity | solo | Cyclomatic analysis |
| knowledge | solo | Query knowledge artifacts |
| bug-hunt | solo | Investigate bugs |
| doc | solo | Generate documentation |
| handoff | solo | Session handoff |
| inbox | solo | Agent mail monitoring |
| trace | solo | Trace design decisions |
| extract | background | Extract from transcripts |
| inject | background | Load knowledge at session start |
| forge | background | Mine transcripts for knowledge |
| provenance | background | Trace knowledge lineage |
| ratchet | background | Progress gates |
| flywheel | background | Knowledge health monitoring |
| using-agentops | meta | AgentOps workflow guide |
| ~~judge~~ | deprecated | Replaced by /council |

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
| **vibe** | council, complexity, standards | required, required, optional |
| **pre-mortem** | council | required |
| **post-mortem** | council, retro, beads | required, required, optional |
| beads | - | - |
| bug-hunt | beads | optional |
| complexity | - | - |
| **crank** | swarm, vibe, implement, beads, post-mortem | required, required, required, required, optional |
| doc | standards | required |
| extract | - | - |
| flywheel | - | - |
| forge | - | - |
| handoff | retro | optional |
| **implement** | beads, standards | optional, required |
| inbox | - | - |
| inject | - | - |
| knowledge | - | - |
| **plan** | research, beads, pre-mortem, crank, implement | optional, required, optional, optional, optional |
| provenance | - | - |
| ratchet | - | - |
| research | knowledge, inject | optional, optional |
| retro | - | - |
| standards | - | - |
| **swarm** | implement, vibe | required, optional |
| trace | provenance | alternative |
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

POST-SHIP
─────────

┌─────────────┐
│ post-mortem │
│ (council +  │
│   retro)    │
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
| Codex | `codex` | `codex exec --full-auto -o output.md "prompt"` |
| OpenCode | `opencode` | (similar pattern) |

### Default Models

| Vendor | Model |
|--------|-------|
| Claude | Opus |
| Codex/OpenAI | GPT-5.2 |

### /council spawns both

```bash
# Claude agents (via Task tool)
Task(model="opus", run_in_background=true, prompt="...")

# Codex agents (via Bash tool)
codex exec -m gpt-5.2 --full-auto -o /tmp/output.md "..."
```

---

## Deprecated Skills

| Skill | Replaced By | Notes |
|-------|-------------|-------|
| `/judge` | `/council` | Council is the new multi-model validation primitive |

---

## See Also

- `skills/council/SKILL.md` — Core validation primitive
- `skills/vibe/SKILL.md` — Complexity + council for code
- `skills/pre-mortem/SKILL.md` — Council for plans
- `skills/post-mortem/SKILL.md` — Council + retro for wrap-up
