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

### 1. Harden parser edge cases

The GOALS.md parser and renderer need adversarial robustness — backtick corruption, Unicode handling, and table column misalignment are known gaps.

**Steer:** increase

### 2. Close coverage gaps in cmd/ao

The cmd/ao package sits at 58% statement coverage. Add targeted tests for untested command handlers and fallback paths to reach 70%.

**Steer:** increase

### 3. Reduce complexity hotspots

Two functions exceed CC 20 (runRPIParallel at 35, parseGatesTable at 27). Split Cobra-coupled handlers into parsing + execution layers.

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
| go-complexity-ceiling | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/ --threshold 36` | 6 | No Go function in cli/ exceeds CC 35 (ratchet — Directive 3 drives reduction) |
| go-internal-complexity | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/internal/ --threshold 28` | 5 | No function in cli/internal/ exceeds CC 27 (ratchet — Directive 3 drives reduction) |
| go-coverage-floor | `cd cli && timeout 120 go test -cover ./... 2>&1 \| grep '^ok' \| sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p' \| awk '{s+=$1;n++} END{if(n>0 && s/n>=80) exit 0; else exit 1}'` | 4 | Average test coverage stays above 80% |
| security-gate | `test -x scripts/security-gate.sh && timeout 60 bash tests/scripts/test-security-gate.sh` | 6 | Security toolchain gate is executable and passes |
| goals-measure-e2e | `bash -c 'cd cli && go build -o /tmp/ao-goals-test ./cmd/ao && cd .. && /tmp/ao-goals-test goals measure --json --timeout 30 2>/dev/null \| jq -e ".summary.total > 0 and .summary.passing > 0"'` | 5 | Fitness measurement pipeline runs end-to-end |
