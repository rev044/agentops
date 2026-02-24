# Plan: Extract all unconsumed improvements (2026-02-24 snapshot)

**Date:** 2026-02-24
**Source artifacts:** `.agents/rpi/next-work.jsonl` (unconsumed entries), `.agents/council/2026-02-24-post-mortem-recent.md`
**Status:** Extraction-only phase (no code changes applied yet)

## Baseline Audit

| Metric | Command | Result |
|---|---|---|
| Unconsumed backlog items in `next-work.jsonl` | `jq -s '[.[] | select(.consumed==false) | .items[]] | length' .agents/rpi/next-work.jsonl` | 21 |
| Unconsumed items by source epic | `jq -s '[.[] | select(.consumed==false) | .source_epic as $epic | .items[] | {epic: $epic}] | group_by(.epic) | map({epic: .[0].epic, count: length})' .agents/rpi/next-work.jsonl` | 230413f-tdd-hardening: 6, ag-poz: 10, recent: 5 |
| Unconsumed items by severity | `jq -s '[.[] | select(.consumed==false) | .items[] | {severity: .severity}] | group_by(.severity) | map({severity: .[0].severity, count: length})' .agents/rpi/next-work.jsonl` | high: 6, medium: 10, low: 5 |
| Unconsumed items by `target_repo` | `jq -s '[.[] | select(.consumed==false) | .items[] | {target_repo: .target_repo}] | group_by(.target_repo) | map({target_repo: .[0].target_repo, count: length})' .agents/rpi/next-work.jsonl` | agentops: 9, nami: 8, `*`: 4 |

## Baseline constraints and scope

- This extract is intentionally broad and includes everything still marked `consumed: false`.
- `.agents` is excluded from git; this plan is tracked in `docs/plans` for portability.
- No validation gates have been run in this extraction pass.

## Extraction Scope: 21 unconsumed improvements

### Source epic: `230413f-tdd-hardening` (6)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 1 | Clean dead code in `measure_test.go` | tech-debt | medium | council-finding | `agentops` | Remove longOutput/order variables and fix byte-vs-rune assertion in `TestMeasureOne_OutputTruncated`. |
| 2 | Add `goleak` for goroutine leak detection | improvement | medium | retro-learning | `agentops` | Replace `runtime.NumGoroutine+sleep` heuristic in leak tests. |
| 3 | Assert AntiStars values in `goals_init_test.go` | tech-debt | low | council-finding | `agentops` | Add assertions for `AntiStars[0]` and `AntiStars[1]`. |
| 4 | Fix `ci-local-release.sh` invocation inconsistency | tech-debt | low | council-finding | `agentops` | Use `./scripts/check-skill-flag-refs.sh` in `scripts/ci-local-release.sh`. |
| 5 | Add `MigrateV1ToV2` and `killAllChildren` package-level tests | improvement | medium | retro-learning | `agentops` | 0% coverage functions in `cli/internal/goals`. |
| 6 | Audit signal.Notify sites for goroutine leak | improvement | medium | retro-pattern | `agentops` | Ensure all `signal.Notify` sites follow deterministic teardown patterns. |

### Source epic: `ag-poz` (10)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 7 | Create `rust-cli.yaml` embedded template | tech-debt | high | council-finding | `nami` | `seed.go` recognizes template but file missing in `cli/embedded/templates/`. |
| 8 | Fix context_assemble `readIntelDir` to read `.json` files | tech-debt | high | council-finding | `nami` | Curated `.agents/learnings/` artifacts are `.json`, ignored by current collector. |
| 9 | Fix curate verify GOALS.md fallback | tech-debt | high | council-finding | `nami` | `runCurateVerify` hardcodes `GOALS.yaml`; fresh repos use `GOALS.md`. |
| 10 | Add wiring integration test `seed->curate->assemble->verify` | improvement | high | council-finding | `nami` | No pipeline test exists for this full flow. |
| 11 | Extract shared `detectTemplate` function | tech-debt | medium | council-finding | `nami` | `seed.go` and `goals_init.go` diverge. |
| 12 | Fix shell injection in `constraint-compiler.sh` JSON construction | tech-debt | medium | council-finding | `nami` | `TITLE`/`SUMMARY` interpolation without escaping in JSON literal path. |
| 13 | Fix jq-less fallback data loss in `constraint-compiler.sh` | tech-debt | medium | council-finding | `nami` | Fallback path overwrites index instead of merging entries. |
| 14 | Migrate `constraint.go` to root-level `GetOutput()` | tech-debt | low | council-finding | `nami` | `ao constraint list -o json` currently bypassed. |
| 15 | Require test file in same commit as command file | process-improvement | medium | retro-learning | `*` | Process policy to prevent Go command coverage regressions. |
| 16 | Formalize Wave 1 spec consistency checklist | process-improvement | low | retro-pattern | `*` | Make W1 consistency check explicit and tracked. |

### Source epic: `recent` (5)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 17 | Raise cmd/ao coverage floor and block regressions | tech-debt | high | council-finding | `agentops` | 961 zero-coverage functions, 103 files with 0% in `cmd/ao` (62.1%). |
| 18 | Add handler tests for batch_forge and batch_promote command paths | improvement | high | retro-pattern | `agentops` | 0% coverage for `runForgeBatch`, `loadAndFilterTranscripts`, `runBatchPromote`, `tryPromoteEntry`, `processPromotionCandidate`. |
| 19 | Require command-surface parity completion checklist before closing cycles | process-improvement | medium | retro-learning | `*` | Prevent false-close of cycles when broad command surface remains untested. |
| 20 | Add CI fallback mode for fork-sensitive Go suites | process-improvement | medium | retro-pattern | `*` | Keep validation deterministic when parallel jobs false-fail due env limits. |
| 21 | Add preflight existence checks for `checkpoint-policy` and `metadata-verification` docs | process-improvement | low | retro-learning | `agentops` | `/post-mortem` preflight currently blocked by missing references. |

## Files likely touched (inventory)

| File | Expected change |
|---|---|
| `cli/internal/goals/measure.go` | `signal.Notify` lifecycle handling pattern review (potential goroutine leak hardening). |
| `cli/internal/goals/measure_unix.go` | package-level tests for `killAllChildren`, nil-map guard behavior. |
| `cli/internal/goals/goals.go` | testability and coverage-oriented tests for `MigrateV1ToV2`. |
| `cli/internal/goals/goals_test.go` | new assertions for `AntiStars`, `MigrateV1ToV2` coverage. |
| `cli/internal/goals/measure_test.go` | dead code cleanup and `goleak` migration for goroutine leak tests. |
| `cli/cmd/ao/goals_init_test.go` | explicit `AntiStars` value assertions. |
| `scripts/ci-local-release.sh` | invocation normalization for nested script checks. |
| `scripts/check-skill-flag-refs.sh` | call-site compatibility if script path expectations are updated. |
| `scripts/validate-go-fast.sh` | `cmd/ao` coverage floor and zero-coverage handler gate. |
| `scripts/pre-push-gate.sh` | optional fork-fallback mode + cmd/ao floor invocation path. |
| `cli/embedded/templates/rust-cli.yaml` | add missing rust template asset. |
| `cli/cmd/ao/seed.go` | template discovery normalization and Rust detection behavior. |
| `cli/cmd/ao/goals_init.go` | shared template detection path with `seed.go`. |
| `cli/cmd/ao/context_assemble.go` | include `.json` artifacts in `readIntelDir`. |
| `cli/cmd/ao/curate.go` | GOALS.md fallback for verify flow. |
| `cli/cmd/ao/seed_test.go` | full pipeline integration assertions for seed→curate→assemble→verify. |
| `cli/cmd/ao/constraint.go` | consolidate output selection with root json flag path (`GetOutput()`). |
| `scripts/constraint-compiler.sh` | JSON-safe interpolation and jq-less merge-safe append behavior. |
| `skills/post-mortem/SKILL.md` | optional preflight for required docs with WARN-to-block semantics. |
| `skills/skill-name/SKILL.md` entries requiring W1/Wclose checks | add explicit file/test commitment rules where applicable. |

## Implementation waves (recommended)

### Wave 1: Stability gates and immediate risk reduction
- Item 17 — cmd/ao coverage floor gate
- Item 18 — batch command handler tests
- Item 4 — `ci-local-release.sh` path consistency
- Item 21 — post-mortem missing-doc preflight
- Item 15 — required test-file-in-command-commit rule (policy)

### Wave 2: Pipeline reliability and correctness
- Item 12, 13 — `constraint-compiler.sh` robustness
- Item 14 — `constraint.go` output path normalization
- Item 10 — seed→curate pipeline integration test
- Items 7, 8, 9, 11 — template/context/verify alignment in `nami` CLI flows

### Wave 3: Coverage depth and leakage cleanup
- Item 1, 2, 3, 5, 6 — goals/goals package coverage and signal/goroutine hygiene
- Item 20 — CI fork-fallback execution strategy
- Item 19, 16 — close-check and Wave 1 consistency controls

## Dependency rationale

- Process-gate hardening (Wave 1) blocks completion signals and should be in place before declaring further technical debt work complete.
- `nami` CLI correctness items (Wave 2) are grouped to avoid partial delivery of seed/curate behavior.
- Package-level reliability and test hardening (Wave 3) can proceed in parallel once major surface gates are enforceable.

## Suggested first commands to start execution

1. `jq -s '[.[] | select(.consumed==false) | .items[]] | length' .agents/rpi/next-work.jsonl`
2. `go test ./cli/internal/... -run TestMeasure -v`
3. `bash scripts/validate-go-fast.sh`
4. `bash scripts/pre-push-gate.sh`

## Tracking

- Source backlog is `.agents/rpi/next-work.jsonl`.
- `consumed` flags still `false` for all 21 items above.
- Next expected transition: materialize each row above into actionable implementation tasks with acceptance checks.
