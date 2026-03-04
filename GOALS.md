# Goals

Make every coding agent session smarter than the last.

## North Stars

- Every session compounds knowledge that future sessions retrieve
- Validation catches regressions before they reach users
- Zero manual intervention required for standard workflows

## Anti Stars

- Untested changes reaching main
- Goals that are trivially true or test implementation details
- Feature creep that adds complexity without user-visible value

## Directives

### 1. Cut the zero-coverage frontier in half

cmd/ao had 17 zero-coverage functions. Drove to 6 (all excluded: main, Execute, interactiveDemo, runRPIServe, runParallelEpic, runParallelEpicTmux). Coverage floor raised 80→84%, zero-max lowered 25→8, handler-zero-max lowered 20→5. Gate locked.

**Steer:** decrease

### 2. Tighten the cli/ complexity ceiling from 25 to 20

Three production functions exceeded CC 20. All refactored below 20 (collectPatterns 23→10, runMine 21→11, runSeed 21→11). Gate threshold lowered from 25 to 21. Max production CC is now 20.

**Steer:** decrease

### 3. Ship one cross-runtime skill validation test

Skills run on Claude Code, Codex CLI, and OpenCode — but only Claude Code is gate-tested. Shipped `tests/codex/test-skill-cross-runtime.sh` (2026-03-02) — exercises `ao inject` via both Claude Code and Codex CLI runtimes, verifies structured JSON consistency. Gate anchored.

**Steer:** decrease (maintain)

### 4. Prove flywheel compounds across sessions

The north star claims every session is smarter than the last. Instrumented and validated: `flywheel-proof` gate (automated proof-run.sh) and `flywheel-compounding` gate (σρ > δ) both passing green. Evidence chain: learnings captured → cited → applied across sessions.

**Steer:** decrease (maintain)

### 7. Enforce release cadence in pre-release gate

Policy: max 1 published release per week (security hotfixes exempt). Was violated (4 releases on 2026-02-27). Added `scripts/release-cadence-check.sh` with warn <7 days, block <1 day. Wired into `ci-local-release.sh` Phase 2. Gate: cadence check passes.

**Steer:** decrease (violation count)

### 5. Run Athena knowledge cycle daily

The flywheel captures learnings reactively. Athena mines git, `.agents/`, and code hotspots to extract signal that sessions missed, then defrags stale/duplicate learnings and flags oscillating evolve goals before they waste cycles. Gate: `ao defrag` report is ≤26 hours old and stale learning count ≤5.

**Steer:** decrease (stale count, age)

### 6. Eliminate oscillating evolve goals

Goals that alternate improved→fail for ≥3 consecutive cycles indicate the improvement approach isn't working — each cycle wastes tokens. Athena's oscillation sweep detects these and quarantines them. Gate: zero oscillating goals in cycle history.

**Steer:** decrease

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| go-cli-builds | `cd cli && go build -o /dev/null ./cmd/ao` | 8 | Go CLI compiles without errors |
| go-cli-tests | `cd cli && timeout 120 go test -race ./...` | 8 | All Go tests pass with race detector |
| go-vet-clean | `cd cli && go vet ./...` | 5 | No common bugs detected by vet |
| hook-preflight | `timeout 60 ./scripts/validate-hook-preflight.sh` | 6 | All hooks pass safety checks |
| manifest-versions-match | `test "$(jq -r '.metadata.version' .claude-plugin/marketplace.json)" = "$(jq -r '.version' .claude-plugin/plugin.json)"` | 5 | Plugin and marketplace versions in sync |
| skill-frontmatter | `bash -c 'for f in skills/*/SKILL.md; do head -5 "$f" \| grep -q "^---" && head -10 "$f" \| grep -q "^name:" && head -10 "$f" \| grep -q "^description:" \|\| { echo FAIL:$f; exit 1; }; done'` | 5 | Every skill has valid YAML frontmatter |
| contract-compatibility | `timeout 60 bash scripts/check-contract-compatibility.sh` | 5 | Contract schemas and references exist on disk |
| wiring-closure | `timeout 60 bash scripts/check-wiring-closure.sh` | 7 | All scripts, skills, and hooks referenced by registries |
| go-complexity-ceiling | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/ --threshold 21 && timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/internal/ --threshold 18` | 6 | No Go function exceeds CC thresholds (cli/: 21, cli/internal/: 18) |
| go-coverage-floor | `cd cli && timeout 120 go test -cover ./... 2>&1 \| grep '^ok' \| sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p' \| awk '{s+=$1;n++} END{if(n>0 && s/n>=95) exit 0; else exit 1}'` | 4 | Average test coverage stays above 95% |
| cmd-ao-coverage-floor | `bash scripts/check-cmdao-coverage-floor.sh` | 6 | cmd/ao coverage floor and zero-coverage regression threshold are enforced |
| security-gate | `test -x scripts/security-gate.sh && timeout 60 bash tests/scripts/test-security-gate.sh` | 6 | Security toolchain gate is executable and passes |
| flywheel-compounding | `bash -c 'cd cli && go build -o /tmp/ao-fw-check ./cmd/ao && cd .. && /tmp/ao-fw-check flywheel status --json 2>/dev/null \| jq -e ".compounding == true"'` | 5 | Knowledge flywheel is above escape velocity (σρ > δ) |
| goals-validate | `bash -c 'cd cli && go build -o /tmp/ao-goals-val ./cmd/ao && cd .. && /tmp/ao-goals-val goals validate --json 2>/dev/null \| jq -e ".valid == true"'` | 5 | GOALS.md parses and validates without structural errors |
| athena-freshness | `bash scripts/check-athena-health.sh` | 4 | Athena defrag report ≤26h old, stale learnings ≤5 |
| athena-no-oscillation | `bash -c 'test -f .agents/defrag/latest.json && jq -e "(.oscillation.oscillating_goals // []) \| length == 0" .agents/defrag/latest.json'` | 4 | No evolve goals oscillating ≥3 consecutive cycles |
| flywheel-proof | `bash scripts/proof-run.sh` | 7 | Flywheel compounds across sessions (automated proof) |
| release-cadence | `bash scripts/release-cadence-check.sh` | 3 | Release cadence policy enforced (warn <7d, block <1d) |
