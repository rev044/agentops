# Plan: Extract all unconsumed improvements (2026-02-24 snapshot)

**Date:** 2026-02-24
**Source artifact:** `.agents/rpi/next-work.jsonl` (all `consumed:false` entries)
**Status:** Extraction-only phase (no code changes applied yet)

## Baseline Audit

| Metric | Command | Result |
|---|---|---|
| Total extraction items | n/a | 29 |
| Unconsumed backlog items in `next-work.jsonl` | `jq -R -s 'split("\n") | map(select(length>0) | fromjson | select(.consumed==false) as $e | .items[] | . + {source_epic: $e.source_epic}) | length' .agents/rpi/next-work.jsonl` | 29 |
| Unconsumed items by source epic | `jq -R -s 'split("\n") | map(select(length>0) | fromjson | select(.consumed==false) as $e | .items[] | . + {source_epic: $e.source_epic}) | group_by(.source_epic) | map({source_epic: .[0].source_epic, count: length})' .agents/rpi/next-work.jsonl` | 230413f-tdd-hardening: 6, ag-poz: 10, post-mortem-all-since-v2.17: 8, recent: 5 |
| Included historical backlog additions | n/a | `post-mortem-all-since-v2.17` (8) |
| Items by severity (combined) | n/a | high: 8, medium: 14, low: 7 |
| Items by `target_repo` (combined) | n/a | agentops: 9, nami: 14, `*`: 6 |

## Baseline constraints and scope

- This extract includes all unconsumed backlog items from `.agents/rpi/next-work.jsonl`: 29 total (including the 8 requested historical items represented by `post-mortem-all-since-v2.17`).
- `.agents` is excluded from git; this plan is tracked in `docs/plans` for portability.
- No validation gates have been run in this extraction pass.

## Extraction Scope: 29 improvements

### Source epic: `230413f-tdd-hardening` (6)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 1 | Clean dead code in `measure_test.go` | tech-debt | medium | council-finding | `agentops` | Remove longOutput/order variables and fix byte-vs-rune assertion in `TestMeasureOne_OutputTruncated`. |
| 2 | Add `goleak` for goroutine leak detection | improvement | medium | retro-learning | `agentops` | Replace `runtime.NumGoroutine+sleep` heuristic in leak tests. |
| 3 | Assert AntiStars values in `goals_init_test.go` | tech-debt | low | council-finding | `agentops` | Add assertions for `AntiStars[0]` and `AntiStars[1]`. |
| 4 | Fix `ci-local-release.sh` invocation inconsistency | tech-debt | low | council-finding | `agentops` | Use `./scripts/check-skill-flag-refs.sh` for consistency. |
| 5 | Add `MigrateV1ToV2` and `killAllChildren` package-level tests | improvement | medium | retro-learning | `agentops` | 0% coverage functions in `cli/internal/goals`. |
| 6 | Audit signal.Notify sites for goroutine leak | improvement | medium | retro-pattern | `agentops` | Generalize cleanup pattern used by one confirmed leak fix. |

### Source epic: `ag-poz` (10)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 7 | Create `rust-cli.yaml` embedded template | tech-debt | high | council-finding | `nami` | `seed.go` lists rust-cli as template but `cli/embedded/templates/rust-cli.yaml` is missing. |
| 8 | Fix context_assemble `readIntelDir` to read `.json` files | tech-debt | high | council-finding | `nami` | Curated artifacts written to `.json` are currently invisible to context assembly. |
| 9 | Fix curate verify GOALS.md fallback | tech-debt | high | council-finding | `nami` | `runCurateVerify` hardcodes `GOALS.yaml`, breaking fresh repos with `GOALS.md`. |
| 10 | Add wiring integration test seed->curate->assemble->verify | improvement | high | council-finding | `nami` | No end-to-end test validates full pipeline. |
| 11 | Extract shared detectTemplate function | tech-debt | medium | council-finding | `nami` | `seed.go` and `goals_init.go` diverge on detection semantics. |
| 12 | Fix shell injection in `constraint-compiler.sh` JSON construction | tech-debt | medium | council-finding | `nami` | Unescaped TITLE/SUMMARY interpolation in JSON literal. |
| 13 | Fix jq-less fallback data loss in `constraint-compiler.sh` | tech-debt | medium | council-finding | `nami` | Fallback overwrite path currently drops existing constraints index. |
| 14 | Migrate `constraint.go` to root-level `GetOutput()` | tech-debt | low | council-finding | `nami` | `ao constraint list -o json` does not use unified output mode. |
| 15 | Require test file in same commit as command file | process-improvement | medium | retro-learning | `*` | Add process rule to prevent command-file coverage gaps. |
| 16 | Formalize Wave 1 spec consistency checklist | process-improvement | low | retro-pattern | `*` | Wave 1 checklist is currently implicit and inconsistently enforced. |

### Source epic: `post-mortem-all-since-v2.17` (8)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 17 | Fix metrics_health computeLoopDominance stale threshold (30d->90d) | tech-debt | high | council-finding | `nami` | Stale threshold logic appears inverted in health metric retention handling. |
| 18 | Fix constraint.go file permissions to 0600 | tech-debt | high | council-finding | `nami` | File permissions for generated constraints are inconsistent with policy. |
| 19 | Add curateParseFrontmatter YAML library parsing | tech-debt | medium | council-finding | `nami` | Current parser risks malformed data handling for frontmatter. |
| 20 | Add BATS tests for `constraint-compiler.sh` | improvement | medium | council-finding | `nami` | No test coverage for this shell script path. |
| 21 | Establish prior-findings resolution tracking across releases | process-improvement | medium | retro-learning | `*` | Current prior-findings resolution rate (17%) is unsustainable. |
| 22 | Add template-existence lint test for validTemplates maps | process-improvement | medium | retro-pattern | `*` | Template map mismatch anti-pattern can go undetected. |
| 23 | Fix readIntelDir I/O efficiency (single pass batch read) | improvement | low | council-finding | `nami` | Current approach scales poorly with large file sets. |
| 24 | Standardize JSON output flag across all commands | tech-debt | low | council-finding | `nami` | Inconsistent `json` output flag behavior across 5+ command paths. |

### Source epic: `recent` (5)

| # | Title | Type | Severity | Source | Target | Evidence |
|---|---|---|---|---|---|---|
| 25 | Raise cmd/ao coverage floor and block regressions | tech-debt | high | council-finding | `agentops` | cmd/ao has ~961 zero-coverage functions and ~62.1% coverage. |
| 26 | Add handler tests for batch_forge and batch_promote command paths | improvement | high | retro-pattern | `agentops` | `runForgeBatch`, `loadAndFilterTranscripts`, `runBatchPromote`, `tryPromoteEntry`, `processPromotionCandidate` are untested. |
| 27 | Require command-surface parity completion checklist before closing cycles | process-improvement | medium | retro-learning | `*` | Coverage for command command-surface still broad despite narrowed closure check. |
| 28 | Add CI fallback mode for fork-sensitive Go suites | process-improvement | medium | retro-pattern | `*` | Parallel suite instability should auto-fallback to serial mode. |
| 29 | Add preflight existence checks for checkpoint-policy and metadata-verification docs | process-improvement | low | retro-learning | `agentops` | Post-mortem preflight references absent files in this repo. |

## Files likely touched (inventory)

| File | Expected change |
|---|---|
| `cli/internal/goals/measure.go` | Review and harden `signal.Notify` lifecycle handling where applicable. |
| `cli/internal/goals/measure_unix.go` | Add package-level tests for `killAllChildren`, including nil-map guard. |
| `cli/internal/goals/goals.go` | Add `MigrateV1ToV2` tests. |
| `cli/internal/goals/goals_test.go` | Add `AntiStars` value assertions. |
| `cli/internal/goals/measure_test.go` | Remove dead test code and migrate to `goleak`. |
| `cli/cmd/ao/goals_init_test.go` | Add explicit `AntiStars` value assertions. |
| `scripts/ci-local-release.sh` | Inconsistency fix for `check-skill-flag-refs.sh` invocation and related script checks. |
| `scripts/check-skill-flag-refs.sh` | Adjust path/call expectations if changed by the release gate. |
| `cli/embedded/templates/rust-cli.yaml` | Add missing rust template asset. |
| `cli/cmd/ao/seed.go` | Normalize template discovery and avoid cross-reference drift. |
| `cli/cmd/ao/constraint.go` | Template and output-mode harmonization for constraint command behaviors. |
| `cli/cmd/ao/context_assemble.go` | Include `.json` files and improve `readIntelDir` performance. |
| `cli/cmd/ao/curate.go` | Fix GOALS.md fallback and parse frontmatter via YAML library. |
| `cli/cmd/ao/metrics_health.go` | Fix computeLoopDominance stale threshold and JSON output consistency in metrics. |
| `scripts/constraint-compiler.sh` | Escape JSON-safe values and keep jq-less merge behavior. |
| `tests/constraint-compiler.bats` | Add BATS-backed coverage for `constraint-compiler.sh` behavior. |
| `skills/post-mortem/SKILL.md` | Add prior-findings resolution metadata checks. |
| `scripts/pre-push-gate.sh` | Add fork-safe fallback and command-surface completion checks. |

## Implementation waves (recommended)

### Wave 1: Stability gates and immediate risk reduction
- Item 25 — cmd/ao coverage floor gate
- Item 26 — batch handler tests
- Item 4 — `ci-local-release.sh` path consistency
- Item 29 — post-mortem missing-doc preflight
- Item 15 — required test-file-in-command commit rule

### Wave 2: Pipeline reliability and correctness
- Item 12, 13, 20 — `constraint-compiler.sh` hardening and test coverage
- Item 14 — `constraint.go` output mode normalization
- Item 10 — seed→curate→assemble→verify pipeline test
- Items 7, 8, 9, 11, 14, 18, 19, 23, 24 — template/context/verify/metrics/permissions/capability alignment in `nami` CLI flows

### Wave 3: Coverage depth and process hardening
- Item 1, 2, 3, 5, 6 — goals package reliability and goroutine hygiene
- Item 27, 16, 21, 22 — closure + consistency + findings tracking gates

## Dependency rationale

- Process gates (Wave 1) are prerequisites for claiming completion of surface-level items in Waves 2 and 3.
- `nami` corrections are grouped to avoid partial delivery on the template/curate/mutation pipeline.
- Coverage and leak hardening (Wave 3) can run in parallel once gating and pipeline consistency are established.

## Suggested first commands to start execution

1. `jq -R -s 'split("\n") | map(select(length>0) | fromjson | select(.consumed==false) as $e | .items[] | . + {source_epic: $e.source_epic}) | length' .agents/rpi/next-work.jsonl`
2. `jq -R -s 'split("\n") | map(select(length>0) | fromjson | select(.consumed==false) | .items[] | {title}' .agents/rpi/next-work.jsonl | sort`
3. `bash scripts/ci-local-release.sh`

## Tracking

- Source backlog is `.agents/rpi/next-work.jsonl`.
- `consumed` flags remain `false` for all 29 items until implementation closes them.
- Next transition: materialize each row above into actionable implementation tasks with acceptance checks.
