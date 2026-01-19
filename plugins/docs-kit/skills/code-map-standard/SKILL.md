---
name: code-map-standard
description: >
  Gold standard for code-map feature documentation. Use when generating
  docs in docs/code-map/ or running /doc on CODING repositories.
version: 2.0.0
context: fork
allowed-tools: "Read,Grep,Glob"
skills:
  - standards
---

# Code-Map Documentation Standard

The authoritative standard for feature documentation in code-map format.

## When to Use

- Running `/doc gen` or `/doc all` on CODING repos
- Creating docs in `docs/code-map/`
- Updating existing code-map docs

**Do NOT use for:** Corpus docs, API reference, READMEs, ADRs.

---

## Required Sections Checklist

Every code-map doc MUST include ALL 16 sections:

| # | Section | Notes |
|---|---------|-------|
| 1 | **Current Status** | With validation date (see format below) |
| 2 | **Platform Parallel** | K8s/vCenter equivalent |
| 3 | **Inputs → Outputs** | Two tables |
| 4 | **Data Flow** | ASCII diagram |
| 5 | **State Machine** | If applicable |
| 6 | **API Endpoints** | Method, Path, Request, Response |
| 7 | **Code Signposts** | Function/class names (NEVER line numbers) |
| 8 | **Configuration** | Env vars + deployment config |
| 9 | **Prometheus Metrics** | Table + PromQL examples |
| 10 | **Unit Tests** | Separate section |
| 11 | **Integration Tests** | Separate from unit |
| 12 | **Dependencies** | Links to related docs |
| 13 | **Example - curl** | Complete with response |
| 14 | **Example - SDK** | TypeScript/Python |
| 15 | **What Worked Well** | 2-3 bullets |
| 16 | **What We'd Do Differently** | 1-2 bullets |

---

## Critical Rules

### Code Signposts

**NEVER use line numbers.** They drift after any edit.

| Valid | Invalid |
|-------|---------|
| `createMission()` | `lines 45-78` |
| `FleetStore` | `server/mission.ts:820` |

### Current Status Format (REQUIRED)

Every code-map doc MUST have a validated status with date and source:

```markdown
## Current Status: ✅ RUNNING
Validated: 2026-01-04 against ocppoc cluster

## Current Status: ❌ FAILED
Status: Accepted=False (CRD exists but not running)
Validated: 2026-01-04 against ocppoc cluster

## Current Status: 📝 PLANNED
Not yet deployed - template only
```

**Status Icons:**
| Icon | Meaning |
|------|---------|
| ✅ | Running and healthy |
| ❌ | Failed or errored |
| 📝 | Planned/templated only |
| 🔧 | In progress/partial |

**Validation Date:** Required for ✅ and ❌ statuses. Must be within 30 days.

### Section Markers

```markdown
<!-- HUMAN-MAINTAINED START -->
Architecture, examples, learnings
<!-- HUMAN-MAINTAINED END -->

<!-- AUTO-GENERATED START -->
Endpoints, signposts, config
<!-- AUTO-GENERATED END -->
```

**Rule:** Never overwrite HUMAN-MAINTAINED during regeneration.

---

## Anti-Patterns

| Don't | Do Instead |
|-------|------------|
| Line numbers in signposts | Function/class names |
| Combine Unit/Integration tests | Separate sections |
| Skip Prometheus metrics | Include + PromQL |
| Only curl examples | Both curl AND SDK |

---

## References

- **Full Template**: `references/full-template.md`
- **Gold Standard**: `ai-platform/docs/code-map/fleet-management.md`

---

## Standards Loading

When creating code-map documentation, reference markdown standard:

| File Pattern | Load Reference |
|--------------|----------------|
| `*.md` | `domain-kit/skills/standards/references/markdown.md` |
