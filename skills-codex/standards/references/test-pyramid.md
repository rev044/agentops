# Test Pyramid — L0 through L7

> Shared reference for RPI lifecycle skills. Loaded by `$plan`, `$pre-mortem`, `$implement`, `$crank`, `$validation`, and `$post-mortem`.

## The Full Testing Lifecycle

| Level | Name | What It Tests | When It Runs | Who Writes It | Context Needed |
|-------|------|---------------|--------------|---------------|----------------|
| L0 | Contract Tests | Spec boundaries — registration, imports, file existence | Every commit (CI) | Agent from SPEC.md | Just the spec |
| L1 | Unit Tests | Single function/class behavior in isolation | Every commit (CI) | Agent via TDD, before code | Spec + function signature |
| L2 | Integration Tests | Multiple modules working together within a subsystem | Every commit (CI) | Agent after units pass | Subsystem spec |
| L3 | Component Tests | Full subsystem end-to-end with mocked external deps | Pre-merge gate | Agent or human | Subsystem + adapter specs |
| L4 | Smoke Tests | Critical path works after deployment — "does it boot?" | Post-deploy (staging) | Human defines, agent implements | Deployment runbook |
| L5 | E2E Tests | Full system flow across subsystems, real infrastructure | Staging environment | Human designs, agent executes | Architecture doc |
| L6 | Acceptance Tests | Does it do what the user actually needed? | Staging with real data | Human validates | PRODUCT.md |
| L7 | Canary / Prod Validation | Does it work under real load with real users? | Production (gradual rollout) | Automated monitors + human judgment | Prod observability |

## Agent Autonomy Boundaries

```
┌─────────────────────────────────────────────────────┐
│  AGENT-AUTONOMOUS (L0–L3)                           │
│  Agent writes tests AND implementation.             │
│  No human input needed for test design.             │
│                                                     │
│  L0: Contract — from SPEC.md alone                  │
│  L1: Unit     — TDD RED→GREEN from spec             │
│  L2: Integration — from subsystem spec + adapters   │
│  L3: Component — agent writes, human defines scenarios│
├─────────────────────────────────────────────────────┤
│  HUMAN-GUIDED (L4–L7)                               │
│  Human defines WHAT to test.                        │
│  Agent builds the test infrastructure.              │
│                                                     │
│  L4: Smoke     — human defines "critical path"      │
│  L5: E2E       — human designs flow, agent harness  │
│  L6: Acceptance— human only validates               │
│  L7: Prod      — monitors + human judgment          │
└─────────────────────────────────────────────────────┘
```

## RPI Phase Mapping

| RPI Phase | Test Levels | What Happens |
|-----------|-------------|--------------|
| **Discovery** (`$discovery`, `$plan`) | L0–L3 scoping | Plan identifies which test levels apply. Issues include `test_level` metadata. |
| **Pre-mortem** (`$pre-mortem`) | L0–L3 coverage check | Validates plan covers appropriate test levels. Flags gaps. |
| **Implementation** (`$implement`, `$crank`) | L0–L2 writing + execution | TDD writes L1 tests first (RED). L0 contracts from specs. L2 after units pass. |
| **Validation** (`$vibe`, `$post-mortem`) | L0–L3 coverage audit | Assesses test pyramid coverage. Flags missing levels as findings. |

## Test Level Selection Guide

Use this decision tree when planning which test levels to include:

```
Does the change touch external APIs or I/O?
  YES → L0 (contract) + L1 (unit) + L2 (integration) minimum
  NO  → L1 (unit) minimum

Does it cross module boundaries?
  YES → Add L2 (integration)
  NO  → L1 sufficient

Does it affect a full subsystem workflow?
  YES → Add L3 (component)
  NO  → Skip L3

Is it deploying to staging/prod?
  YES → L4 (smoke) required, L5 (E2E) recommended
  NO  → Skip L4–L7
```

## Test Level Metadata for Issues

When creating issues in `$plan`, include test level metadata:

```json
{
  "test_levels": {
    "required": ["L0", "L1"],
    "recommended": ["L2"],
    "deferred": ["L3"],
    "rationale": "Pure internal refactor — L0 contracts verify spec, L1 units verify behavior, L2 recommended for cross-module calls"
  }
}
```

## Coverage Assessment Template

Used by `$post-mortem` and `$vibe` to assess test pyramid health:

| Level | Tests Exist? | Tests Pass? | Coverage Gap? | Action |
|-------|-------------|-------------|---------------|--------|
| L0 Contract | yes/no | yes/no/na | description | add/fix/ok |
| L1 Unit | yes/no | yes/no/na | description | add/fix/ok |
| L2 Integration | yes/no | yes/no/na | description | add/fix/ok |
| L3 Component | yes/no | yes/no/na | description | add/fix/ok |
| L4+ | human-gated | — | — | defer to human |
