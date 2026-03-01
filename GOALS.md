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

cmd/ao has 12 zero-coverage functions (0 handler zeros). Each uncovered function is a bug hiding in plain sight. Drive this to ≤6 by targeting untested command handlers and error paths — then lower the `cmd-ao-coverage-floor` zero-max threshold to lock in the gain.

**Steer:** decrease

### 2. Tighten the cli/ complexity ceiling from 25 to 20

Three production functions exceed CC 20: `buildLastSessionSection` (CC 24), `collectPatterns` (CC 23), `runSeed` (CC 21). Refactor each to stay under 20, then lower the `go-complexity-ceiling` gate threshold from 25 to 21 to prevent future creep.

**Steer:** decrease

### 3. Ship one cross-runtime skill validation test

Skills run on Claude Code, Codex CLI, and OpenCode — but only Claude Code is gate-tested. Add one automated integration test that exercises a skill (e.g. `goals`, `inject`) via Codex CLI and verifies structured output. This anchors the multi-runtime value proposition with evidence rather than assumption.

**Steer:** increase

### 4. Evolve `ao rpi serve` into a multi-run mission control

`ao rpi serve` now shows one run. Extend it to aggregate all active runs simultaneously — a true mission control view where multiple parallel `ao rpi parallel` runs are visible in a single dashboard. Add run-switcher UI, comparative cost/timing, and worker heatmaps. The foundation (SSE + events.jsonl per run) is already there.

**Steer:** increase

### 5. Wire nudge protocol into the dashboard

`ao rpi nudge` exists and writes to `commands.jsonl`, but the dashboard has no send interface — you can only watch, not steer. Add a nudge composer to `ao rpi serve`: a text input that writes to the active run's commands.jsonl. The round-trip is: browser → POST /nudge → `ao rpi nudge` → tmux → agent.

**Steer:** increase

### 6. Run Athena knowledge cycle daily

The flywheel captures learnings reactively. Athena mines git, `.agents/`, and code hotspots to extract signal that sessions missed, then defrags stale/duplicate learnings and flags oscillating evolve goals before they waste cycles. Gate: `ao defrag` report is ≤26 hours old and stale learning count ≤5.

**Steer:** decrease (stale count, age)

### 7. Eliminate oscillating evolve goals

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
| go-complexity-ceiling | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/ --threshold 25` | 6 | No Go function in cli/ exceeds CC 24 (tightened from 35 after ag-atu refactoring) |
| go-internal-complexity | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/internal/ --threshold 18` | 5 | No function in cli/internal/ exceeds CC 17 (tightened from 27 after ag-atu refactoring) |
| go-coverage-floor | `cd cli && timeout 120 go test -cover ./... 2>&1 \| grep '^ok' \| sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p' \| awk '{s+=$1;n++} END{if(n>0 && s/n>=80) exit 0; else exit 1}'` | 4 | Average test coverage stays above 80% |
| cmd-ao-coverage-floor | `bash scripts/check-cmdao-coverage-floor.sh` | 6 | cmd/ao coverage floor and zero-coverage regression threshold are enforced |
| security-gate | `test -x scripts/security-gate.sh && timeout 60 bash tests/scripts/test-security-gate.sh` | 6 | Security toolchain gate is executable and passes |
| rpi-serve-smoke | `bash -c 'cd cli && go build -o /tmp/ao-rpi-serve-s ./cmd/ao && /tmp/ao-rpi-serve-s rpi serve --help 2>&1 \| grep -qF "port"'` | 5 | ao rpi serve subcommand exists and exposes expected flags |
| flywheel-compounding | `bash -c 'cd cli && go build -o /tmp/ao-fw-check ./cmd/ao && cd .. && /tmp/ao-fw-check flywheel status --json 2>/dev/null \| jq -e ".compounding == true"'` | 5 | Knowledge flywheel is above escape velocity (σρ > δ) |
| goals-validate | `bash -c 'cd cli && go build -o /tmp/ao-goals-val ./cmd/ao && cd .. && /tmp/ao-goals-val goals validate --json 2>/dev/null \| jq -e ".valid == true"'` | 5 | GOALS.md parses and validates without structural errors |
| athena-freshness | `bash scripts/check-athena-health.sh` | 4 | Athena defrag report ≤26h old, stale learnings ≤5 |
| athena-no-oscillation | `bash -c 'test -f .agents/defrag/latest.json && jq -e "(.oscillation.oscillating_goals // []) \| length == 0" .agents/defrag/latest.json'` | 4 | No evolve goals oscillating ≥3 consecutive cycles |
