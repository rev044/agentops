# Skill Tier Taxonomy

This document defines the `tier` field used in skill frontmatter to categorize skills by their role in the AgentOps workflow.

## Tier Values

| Tier | Description | Examples |
|------|-------------|----------|
| **solo** | Standalone skills invoked directly by users | research, plan, vibe, implement |
| **library** | Reference skills loaded JIT by other skills | beads, standards |
| **orchestration** | Multi-skill coordinators that run other skills | crank |
| **team** | Skills requiring human collaboration | implement (guided mode) |
| **background** | Hook-triggered or automatic skills | inject, forge, extract |
| **meta** | Skills about skills (documentation, validation) | using-agentops |

## Tier vs Context Discovery Tiers

**Important:** The skill `tier` field is **different** from the 6-tier context discovery hierarchy.

| Concept | Purpose | Values |
|---------|---------|--------|
| **Skill tier** | Categorizes skill role in workflow | solo, library, orchestration, team, background, meta |
| **Context discovery tier** | Prioritizes where to find information | Code-map, Semantic, Grep, Source, .agents/, External |

The context discovery hierarchy (1-6) describes WHERE to look for information during research.
The skill tier describes WHAT KIND of skill it is.

## Usage in Frontmatter

```yaml
---
name: vibe
tier: solo
description: Comprehensive code validation
---
```

## Tier Selection Guide

| If the skill... | Use tier |
|-----------------|----------|
| Is invoked directly via `/skill-name` | `solo` |
| Provides reference docs for other skills | `library` |
| Runs multiple other skills in sequence | `orchestration` |
| Requires human in the loop | `team` |
| Runs automatically via hooks or internally | `background` |
| Documents or validates other skills | `meta` |

## Current Skill Tiers

| Skill | Tier |
|-------|------|
| beads | library |
| standards | library |
| crank | orchestration |
| implement | team |
| research | solo |
| plan | solo |
| vibe | solo |
| pre-mortem | solo |
| post-mortem | solo |
| retro | solo |
| knowledge | solo |
| bug-hunt | solo |
| complexity | solo |
| doc | solo |
| extract | background |
| inject | background |
| forge | background |
| provenance | background |
| ratchet | background |
| flywheel | background |
| handoff | solo |
| inbox | solo |
| swarm | orchestration |
| judge | orchestration |
| trace | solo |
| using-agentops | meta |

---

## Skill Dependency Graph

This section documents the dependencies between skills. Dependencies are declared in each skill's frontmatter via the `dependencies:` field.

### Dependency Types

| Type | Meaning |
|------|---------|
| **required** | Skill invokes or relies on this dependency |
| **optional** | Skill can use this dependency if available |
| **alternative** | Related skill for different use cases |

### Dependency Table

| Skill | Dependencies | Type |
|-------|--------------|------|
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
| post-mortem | beads, retro | optional, implicit |
| pre-mortem | - | - |
| provenance | - | - |
| ratchet | - | - |
| research | knowledge, inject | optional, optional |
| retro | vibe | optional |
| standards | - | - |
| **swarm** | implement, vibe | required, optional |
| judge | vibe, swarm | optional, optional |
| trace | provenance | alternative |
| using-agentops | - | - |
| vibe | standards | required |

### Dependency Graph Visualization

```
                    ┌─────────────────────────────────────┐
                    │           ORCHESTRATION             │
                    │                                     │
                    │  ┌───────┐         ┌───────┐       │
                    │  │ crank │─────────│ swarm │       │
                    │  └───┬───┘         └───┬───┘       │
                    │      │                 │           │
                    └──────┼─────────────────┼───────────┘
                           │                 │
              ┌────────────┼─────────────────┼────────────┐
              │            ▼                 ▼            │
              │       ┌─────────┐      ┌─────────┐        │
              │       │implement│◄─────│  vibe   │        │
              │       └────┬────┘      └────┬────┘        │
              │            │                │             │
              │            ▼                ▼             │
              │       ┌─────────┐      ┌─────────┐        │
              │       │standards│      │standards│        │
              │       └─────────┘      └─────────┘        │
              │                                           │
              │              TEAM / SOLO                  │
              └───────────────────────────────────────────┘

                    ┌─────────────────────────────────────┐
                    │            PLANNING                 │
                    │                                     │
                    │  ┌───────┐         ┌──────────┐    │
                    │  │ plan  │─────────│ research │    │
                    │  └───┬───┘         └────┬─────┘    │
                    │      │                  │          │
                    │      ▼                  ▼          │
                    │  ┌───────┐         ┌─────────┐     │
                    │  │ beads │         │knowledge│     │
                    │  └───────┘         │ inject  │     │
                    │                    └─────────┘     │
                    └─────────────────────────────────────┘

                    ┌─────────────────────────────────────┐
                    │           VALIDATION                │
                    │                                     │
                    │  ┌──────────┐      ┌───────┐       │
                    │  │post-mort.│──────│ retro │       │
                    │  └──────────┘      └───┬───┘       │
                    │                        │           │
                    │                        ▼           │
                    │                   ┌───────┐        │
                    │                   │ judge │        │
                    │                   └───┬───┘        │
                    │                       │            │
                    │                       ▼            │
                    │                   ┌───────┐        │
                    │                   │ vibe  │        │
                    │                   └───────┘        │
                    └─────────────────────────────────────┘

                    ┌─────────────────────────────────────┐
                    │            LIBRARY                  │
                    │                                     │
                    │  ┌─────────┐      ┌───────┐        │
                    │  │standards│      │ beads │        │
                    │  └─────────┘      └───────┘        │
                    │        ▲               ▲           │
                    │        │               │           │
                    │   Used by:        Used by:         │
                    │   vibe, doc,      implement,       │
                    │   implement       bug-hunt,        │
                    │                   plan, crank      │
                    └─────────────────────────────────────┘
```

### Circular Dependency Analysis

**No circular dependencies found.**

The dependency graph is acyclic (DAG). Key observations:

1. **Library skills** (`standards`, `beads`) have no dependencies and are leaf nodes
2. **Background skills** (`inject`, `forge`, `extract`, `flywheel`, `ratchet`, `provenance`) are independent
3. **Orchestration skills** (`crank`, `swarm`) depend on execution skills but not vice versa
4. **Plan** depends on many skills but none depend on plan (except optionally crank)

### Dependency Chains

Longest dependency chains (for context loading):

1. `crank → swarm → implement → standards` (depth: 4)
2. `crank → vibe → standards` (depth: 3)
3. `plan → research → knowledge` (depth: 3)
4. `post-mortem → retro → vibe → standards` (depth: 4)

### Usage in Frontmatter

```yaml
---
name: crank
description: 'Fully autonomous epic execution...'
dependencies:
  - swarm       # required - executes each wave
  - vibe        # required - final validation
  - implement   # required - individual issue execution
  - beads       # required - issue tracking via bd CLI
  - post-mortem # optional - suggested for learnings extraction
---
```
