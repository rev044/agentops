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
| using-agentops | meta |
