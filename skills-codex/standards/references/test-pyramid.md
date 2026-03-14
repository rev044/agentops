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

## Bug-Finding Levels (Agent-Autonomous)

> **Proven 2026-03-14 on jren-cm:** 3,321 L1 unit tests found 0 new bugs. These levels found 8.
> Evidence: `/Users/fullerbt/gt/jren_cm/crew/ichigo/scripts/.agents/council/2026-03-14-post-mortem-full-session-methodology.md`

L0–L3 are the **coverage pyramid** — they verify code works as designed.
These are the **bug-finding pyramid** — they find bugs the coverage pyramid misses.

| Level | Name | What It Finds | Agent Writes? | Bugs Found (jren-cm) |
|-------|------|---------------|---------------|---------------------|
| BF1 | Property-Based | Edge cases from randomized inputs (non-IPv4, empty strings, unicode) | Yes | 1 crash |
| BF2 | Golden/Snapshot | Output drift, template regression | Yes | 0 (regression guard) |
| BF3 | Mutation | Untested code paths (operator flips that tests don't catch) | Yes (config) | 13 gaps |
| BF4 | Chaos/Negative | Unhandled exceptions at system boundaries (corrupt DB, permissions, timeouts) | Yes | 4 bugs |
| BF5 | Script Functional | Bash runtime crashes, jq logic errors, undefined functions | Yes | 2 critical bugs |

### Bug-Finding Level Details

**BF1 — Property-Based Testing**

Randomize inputs to every data transformation. Catches crashes on inputs no human would write.

| Language | Tool | Example |
|----------|------|---------|
| Python | `hypothesis` | `@given(st.text())` on parsers |
| Go | `gopter` or `rapid` | `rapid.Check(t, func(t *rapid.T) { ... })` |
| TypeScript | `fast-check` | `fc.assert(fc.property(fc.string(), (s) => ...))` |
| Rust | `proptest` | `proptest! { fn test(s in ".*") { ... } }` |

**BF2 — Golden/Snapshot Testing**

Generate canonical output, save as golden file. Test asserts exact match. Use `GOLDEN_UPDATE=1` env var to regenerate.

| Language | Tool | Pattern |
|----------|------|---------|
| Python | `pytest` + `tmp_path` | `assert output == golden_path.read_text()` |
| Go | `testdata/` convention | `golden.Update(t, got)` / `golden.Get(t)` |
| TypeScript | `vitest` snapshot | `expect(output).toMatchSnapshot()` |
| Rust | `insta` | `insta::assert_snapshot!(output)` |

**BF3 — Mutation Testing**

Flip operators and conditions in source. If tests still pass, there's an untested path.

| Language | Tool | Command |
|----------|------|---------|
| Python | `mutmut` | `mutmut run --paths-to-mutate src/critical.py` |
| Go | `go-mutesting` | `go-mutesting ./pkg/critical/...` |
| TypeScript | `stryker` | `npx stryker run --mutate 'src/critical/**'` |
| Rust | `cargo-mutants` | `cargo mutants --package critical` |

**Target critical modules only.** Full-codebase mutation is too slow (hours). Pick the 3-5 modules with highest blast radius.

**BF4 — Chaos/Negative Testing**

Inject failures at every system boundary. The bugs L1 mocks away.

```python
# Python pattern
@patch("module.external_call", side_effect=TimeoutError)
def test_timeout_returns_specific_error(mock):
    result = function_under_test()
    assert result.error_class == "timeout"  # NOT "unknown"
```

```go
// Go pattern
func TestHandlesCorruptDB(t *testing.T) {
    db := newCorruptDB()
    store, err := NewStore(db)
    require.NoError(t, err)
    assert.False(t, store.Enabled())  // graceful degradation, not crash
}
```

**Checklist for every subsystem:**
- [ ] External API timeout → specific error, not crash
- [ ] External API connection refused → specific error
- [ ] File permission denied → graceful failure
- [ ] Database corruption → degraded mode, not crash
- [ ] Invalid/missing config → clear error message

**BF5 — Script Functional Testing**

Stub external tools via PATH override, source the script, call its functions, verify output.

```bash
# Create stub directory with fake oc/kubectl
STUB_DIR=$(mktemp -d)
cat > "${STUB_DIR}/oc" << 'EOF'
#!/bin/bash
echo '{"items": [{"status": {"phase": "Failed"}}]}'
EOF
chmod +x "${STUB_DIR}/oc"

# Run specialist with stubbed tools
PATH="${STUB_DIR}:${ORIGINAL_PATH}" source specialist.sh
result=$(scan 2>&1)
echo "$result" | python3 -c "import json,sys; json.load(sys.stdin)"  # valid JSON?
```

### When to Use Bug-Finding Levels

```
After L0–L3 coverage is complete, run bug-finding levels:

  Has data transformations (parse/render/serialize)?
    YES → BF1 (property-based) — randomize all inputs

  Has output generators (config files, reports, manifests)?
    YES → BF2 (golden) — snapshot every output

  Has critical modules (auth, state, error handling)?
    YES → BF3 (mutation) — targeted mutation on those modules

  Has external boundaries (APIs, databases, filesystems)?
    YES → BF4 (chaos) — inject failures at every boundary

  Has bash/shell scripts?
    YES → BF5 (functional) — stub tools, verify behavior
```

### Bug-Finding in RPI Phases

| RPI Phase | Bug-Finding Action |
|-----------|--------------------|
| `$plan` | Classify which BF levels apply per issue |
| `$pre-mortem` | Verify BF levels are planned for boundary-touching code |
| `$implement` | Write BF tests alongside L0–L3 (or as separate wave) |
| `$vibe` | **Check BF coverage before council** — flag missing chaos/property tests on boundary code |
| `$post-mortem` | Assess BF bug discovery count. If BF4 found 0 bugs → either code is solid or chaos tests are too weak |

## Coverage Assessment Template

Used by `$post-mortem` and `$vibe` to assess test pyramid health:

### Coverage Pyramid (L0–L3)

| Level | Tests Exist? | Tests Pass? | Coverage Gap? | Action |
|-------|-------------|-------------|---------------|--------|
| L0 Contract | yes/no | yes/no/na | description | add/fix/ok |
| L1 Unit | yes/no | yes/no/na | description | add/fix/ok |
| L2 Integration | yes/no | yes/no/na | description | add/fix/ok |
| L3 Component | yes/no | yes/no/na | description | add/fix/ok |
| L4+ | human-gated | — | — | defer to human |

### Bug-Finding Pyramid (BF1–BF5)

| Level | Tests Exist? | Bugs Found? | Gap? | Action |
|-------|-------------|-------------|------|--------|
| BF1 Property | yes/no | count | data transforms without property tests | add for parsers/renderers |
| BF2 Golden | yes/no | N/A | output generators without golden files | add for config/report generators |
| BF3 Mutation | yes/no | surviving mutants | critical modules with <80% mutation score | target top 3 modules |
| BF4 Chaos | yes/no | count | external boundaries without failure injection | add for every API/DB/FS boundary |
| BF5 Script Functional | yes/no | count | bash scripts without functional tests | add for top 10 by complexity |
