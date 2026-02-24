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

Per-package coverage in cmd/ao sits below 60%. Add targeted tests for untested command handlers and fallback paths.

**Steer:** increase

### 3. Reduce complexity hotspots

Several functions exceed CC 10 in cmd/ao. Split Cobra-coupled handlers into parsing + execution layers.

**Steer:** decrease

## Gates

| ID | Check | Weight | Description |
|----|-------|--------|-------------|
| go-cli-builds | `cd cli && go build -o /dev/null ./cmd/ao` | 8 | Go CLI compiles without errors |
| go-cli-tests | `cd cli && go test -race ./...` | 8 | All Go tests pass with race detector |
| go-vet-clean | `cd cli && go vet ./...` | 5 | No common bugs detected by vet |
| hook-preflight | `timeout 60 ./scripts/validate-hook-preflight.sh` | 6 | All hooks pass safety checks |
| manifest-versions-match | `test "$(jq -r '.metadata.version' .claude-plugin/marketplace.json)" = "$(jq -r '.version' .claude-plugin/plugin.json)"` | 5 | Plugin and marketplace versions in sync |
| skill-frontmatter | `bash -c 'for f in skills/*/SKILL.md; do head -5 "$f" \| grep -q "^---" && head -10 "$f" \| grep -q "^name:" && head -10 "$f" \| grep -q "^description:" \|\| { echo FAIL:$f; exit 1; }; done'` | 5 | Every skill has valid YAML frontmatter |
| contract-compatibility | `timeout 60 bash scripts/check-contract-compatibility.sh` | 5 | Contract schemas and references exist on disk |
| wiring-closure | `timeout 60 bash scripts/check-wiring-closure.sh` | 7 | All scripts, skills, and hooks referenced by registries |
| go-complexity-ceiling | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/ --threshold 10` | 6 | No Go function in cli/ has cyclomatic complexity >= 10 |
| go-internal-complexity | `timeout 60 bash scripts/check-go-absolute-complexity.sh --dir cli/internal/ --threshold 8` | 5 | No function in cli/internal/ has CC >= 8 |
| go-coverage-floor | `cd cli && go test -cover ./... 2>&1 \| grep '^ok' \| sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p' \| awk '{s+=$1;n++} END{if(n>0 && s/n>=80) exit 0; else exit 1}'` | 4 | Average test coverage stays above 80% |
| security-gate | `test -x scripts/security-gate.sh && timeout 60 bash tests/scripts/test-security-gate.sh` | 6 | Security toolchain gate is executable and passes |
| evolve-kill-switch | `grep -q 'KILL' skills/evolve/SKILL.md && grep -q 'STOP' skills/evolve/SKILL.md` | 5 | Evolve has documented kill switch |
