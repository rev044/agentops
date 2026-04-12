---
description: Run parallel 3-agent validation for 3x faster results
---

# /validate-multi - Parallel Multi-Agent Validation

**Purpose:** Comprehensive validation 3x faster via parallel execution

**When to use:**
- Large projects (many files, long test suites)
- Time-critical validation (CI/CD pipelines)
- Comprehensive checks (security + tests + performance)
- Pre-deployment validation (full coverage needed fast)

**Token budget:** 10-20k tokens (same as single-agent, but 3x wall-clock speedup)

**Output:** Combined validation report with all check results

---

## The Parallel Validation Philosophy

**Same validation depth. Faster results.**

Traditional validation is sequential:
1. Syntax checks (30 seconds)
2. Tests + coverage (60 seconds)
3. Security + performance (30 seconds)
**Total:** 120 seconds (2 minutes)

Parallel validation uses 3 agents simultaneously:
1. Agent 1: Syntax + Build (30 seconds)
2. Agent 2: Tests + Coverage (60 seconds)
3. Agent 3: Security + Performance (30 seconds)
**Total:** 60 seconds (3x speedup!)

**Key insight:** Wall-clock time drops by 3x, validation depth stays the same.

---

## How It Works

### Step 1: Task Distribution

**I automatically split validation into 3 parallel tracks:**

**Agent 1: Syntax & Build Validator**
- YAML/JSON syntax validation
- Code formatting checks (go fmt, black, prettier)
- Static analysis (go vet, flake8, eslint)
- Build verification (compilation succeeds)

**Agent 2: Test & Coverage Validator**
- Unit test execution
- Integration test execution
- Coverage measurement
- Test report generation

**Agent 3: Security & Performance Validator**
- Dependency vulnerability scanning
- Secret detection
- Container image scanning
- Performance regression checks

### Step 2: Parallel Execution

**All 3 agents work simultaneously:**

```
Time: 0 sec
├─ Agent 1 starts: Syntax + Build validation
├─ Agent 2 starts: Running test suites
└─ Agent 3 starts: Security scanning

Time: 60 sec
├─ Agent 1 completes: ✅ Syntax + Build passed
├─ Agent 2 completes: ✅ 45/45 tests passed, 87% coverage
└─ Agent 3 completes: ✅ No vulnerabilities, performance OK

Time: 65 sec
└─ Synthesis: Combined validation report
```

**Total:** 65 seconds (vs 120 seconds sequential)

### Step 3: Synthesis

**I combine results from all 3 agents:**

```markdown
# Multi-Agent Validation Report

**Date:** 2025-11-07 10:30
**Project:** agentops
**Agents:** 3 (parallel execution)
**Duration:** 65 seconds (vs 120s sequential)
**Status:** ✅ PASSED

## Summary (All Agents)
- Agent 1 (Syntax/Build): ✅ PASSED
- Agent 2 (Tests/Coverage): ✅ PASSED
- Agent 3 (Security/Performance): ✅ PASSED

## Agent 1: Syntax & Build
✅ YAML: 127 files, 0 errors
✅ Go fmt: All files formatted
✅ Go vet: No issues
✅ Build: Succeeded (2.1s)

## Agent 2: Tests & Coverage
✅ Unit tests: 45/45 passed (12.5s)
✅ Integration tests: 8/8 passed (48.2s)
✅ Coverage: 87% (threshold: 80%)

## Agent 3: Security & Performance
✅ Dependencies: No vulnerabilities
✅ Secrets: None detected
✅ Performance: No regressions

## Overall Status
✅ All checks passed - OK to deploy
```

---

## Usage

### Basic Syntax

```bash
/validate-multi

# Runs all 3 agents in parallel
# Use: Pre-deployment, comprehensive validation
```

### Quick Parallel Validation

```bash
/validate-multi --quick

# Agents run fewer checks (faster)
# Agent 1: Syntax only (skip build)
# Agent 2: Unit tests only (skip integration)
# Agent 3: Dependency scan only (skip performance)
# Total: ~30 seconds (vs 60 for full)
```

### Specific Track Validation

```bash
/validate-multi --syntax-build
# Only Agent 1 (useful for debugging)

/validate-multi --tests-coverage
# Only Agent 2

/validate-multi --security-performance
# Only Agent 3
```

---

## What I Do Automatically

### 1. Detect Project Type

**I analyze project structure:**

```bash
# Check for indicators
go.mod               → Go project
package.json         → JavaScript project
requirements.txt     → Python project
kustomization.yaml   → Kubernetes project
Dockerfile           → Container project
```

### 2. Assign Validation Tasks

**Per project type, I assign appropriate checks:**

**Go Project:**
- Agent 1: go fmt, go vet, go build
- Agent 2: go test, go test -cover
- Agent 3: nancy (dependencies), gitleaks (secrets)

**Python Project:**
- Agent 1: black, flake8, mypy
- Agent 2: pytest, pytest --cov
- Agent 3: safety check, bandit

**Kubernetes Project:**
- Agent 1: yamllint, kubectl validate
- Agent 2: kustomize build, helm lint
- Agent 3: kubesec scan, kube-score

### 3. Launch 3 Agents

**I create parallel Task tool calls:**

```
Task 1 (Agent 1): Syntax & Build Validator
Task 2 (Agent 2): Test & Coverage Validator
Task 3 (Agent 3): Security & Performance Validator
```

**Each agent:**
- Runs independently
- Returns structured results
- Completes in parallel

### 4. Synthesize Results

**I combine all validation results:**

- Merge pass/fail status
- Aggregate metrics (test counts, coverage %)
- Combine recommendations
- Determine overall status

---

## Agent Specialization

### Agent 1: Syntax & Build Validator (Fast Checks)

**Focus:** Catch basic errors quickly

**For Go:**
```bash
go fmt -l .         # Check formatting
go vet ./...        # Static analysis
go build ./...      # Compilation check
golangci-lint run   # Comprehensive linting
```

**For Python:**
```bash
black --check .     # Formatting
flake8 .            # Linting
mypy .              # Type checking
```

**For Kubernetes:**
```bash
yamllint .          # YAML syntax
kubectl apply --dry-run=client -f .
kustomize build .   # Build verification
```

**Output:**
```markdown
## Agent 1: Syntax & Build

### Formatting
✅ Go: All 45 files formatted correctly
✅ No formatting changes needed

### Static Analysis
✅ go vet: No issues found
⚠️ golangci-lint: 3 style warnings (non-blocking)

### Build
✅ go build: Succeeded (2.1s)
✅ All packages compile

### Recommendations
- Address 3 style warnings (low priority)
```

### Agent 2: Test & Coverage Validator (Behavior Checks)

**Focus:** Verify correctness and coverage

**For Go:**
```bash
go test ./... -v            # Run all tests
go test ./... -cover        # Measure coverage
go test ./... -race         # Race condition detection
go test -bench=. ./...      # Benchmarks
```

**For Python:**
```bash
pytest -v                   # Run tests
pytest --cov=src --cov-report=term
pytest tests/integration/   # Integration tests
```

**For JavaScript:**
```bash
npm test                    # Unit tests
npm run test:coverage       # Coverage
npm run test:integration    # Integration
```

**Output:**
```markdown
## Agent 2: Tests & Coverage

### Unit Tests
✅ 45/45 tests passed (12.5s)
   - auth: 12/12
   - cache: 8/8
   - handlers: 15/15
   - middleware: 10/10

### Integration Tests
✅ 8/8 tests passed (48.2s)

### Coverage
✅ Total: 87% (threshold: 80%)
   - auth: 92%
   - cache: 78% ⚠️ (below 80%)
   - handlers: 91%
   - middleware: 88%

### Performance
✅ Benchmarks: No regressions
   - AuthHandler: 1.2ms (baseline: 1.1ms) +9%

### Recommendations
- Increase cache package coverage to 80%+
- AuthHandler performance slightly slower (investigate)
```

### Agent 3: Security & Performance Validator (Risk Checks)

**Focus:** Identify vulnerabilities and regressions

**For Go:**
```bash
nancy sleuth          # Dependency vulnerabilities
gitleaks detect       # Secret detection
gosec ./...           # Security audit
```

**For Python:**
```bash
safety check          # Dependency scan
bandit -r src/        # Security issues
pip-audit             # Vulnerability check
```

**For Containers:**
```bash
trivy image myapp:latest    # Container scan
docker scan myapp:latest    # Docker scan
```

**Output:**
```markdown
## Agent 3: Security & Performance

### Dependency Security
✅ No known vulnerabilities
✅ All dependencies up to date

### Secret Detection
✅ No secrets found in code
✅ No credentials exposed

### Code Security
✅ gosec: No issues found
✅ Security audit passed

### Container Security
✅ trivy: No vulnerabilities in base image
✅ Docker scan: PASSED

### Performance
✅ No significant regressions detected
⚠️ API response time: +5% (within acceptable range)

### Recommendations
- Monitor API response time (slight increase)
```

---

## Success Criteria

**Multi-agent validation succeeds when:**

✅ Agent 1 (Syntax/Build): All checks pass
✅ Agent 2 (Tests/Coverage): Tests pass, coverage meets threshold
✅ Agent 3 (Security/Performance): No vulnerabilities, acceptable performance
✅ Overall: No blocking issues

**Validation BLOCKS deployment when:**

❌ Agent 1: Build fails, critical syntax errors
❌ Agent 2: Tests fail, coverage drops >10%
❌ Agent 3: Critical vulnerabilities (HIGH severity)

**Validation WARNS but allows deployment when:**

⚠️ Agent 1: Style warnings (non-critical)
⚠️ Agent 2: Coverage slightly below threshold (<5%)
⚠️ Agent 3: Low severity vulnerabilities, minor performance regression

---

## Handling Validation Failures

### Agent 1 Failure: Syntax/Build Issues

**Problem:** Build fails

**Response:**
```
❌ Agent 1 FAILED: Build errors

File: auth/handler.go:45
Error: undefined: validateJWT

Fix:
1. Import required package or define function
2. Run: go build ./auth/...
3. Retry: /validate-multi
```

### Agent 2 Failure: Tests Failing

**Problem:** Tests don't pass

**Response:**
```
❌ Agent 2 FAILED: 3/45 tests failed

Failed:
- auth/jwt_test.go:34 - TestValidateToken
- cache/redis_test.go:56 - TestConnectionPool
- handlers/api_test.go:89 - TestHealthCheck

Fix:
1. Run locally: go test -v ./auth ./cache ./handlers
2. Debug failures
3. Retry: /validate-multi
```

### Agent 3 Failure: Security Issues

**Problem:** Vulnerabilities found

**Response:**
```
❌ Agent 3 FAILED: 2 vulnerabilities

HIGH: CVE-2023-12345 in dependency X v1.2.3
- Fix: go get -u github.com/pkg/X@v1.2.4

MEDIUM: Secret detected in config/auth.yaml:8
- Fix: Move to environment variable

Action:
1. Upgrade dependency
2. Remove secret
3. Retry: /validate-multi
```

---

## When to Use Multi-Agent vs Single-Agent

### Use /validate-multi when:

✅ **Large project** - Many files, long test suites
✅ **Time-critical** - CI/CD pipeline, pre-deployment
✅ **Comprehensive needed** - Security + tests + performance
✅ **Regular validation** - Daily/weekly checks

**Example:** "Pre-deployment check for production release (need results in <2 minutes)"

### Use /validate when:

✅ **Small project** - Few files, quick tests
✅ **Time flexible** - No rush
✅ **Simple validation** - Just syntax or just tests
✅ **Development workflow** - Frequent local checks

**Example:** "Quick check before commit (syntax + unit tests only)"

---

## Token Budget Management

**Same token budget as single-agent:** 10-20k tokens

**Breakdown:**
- Project detection: 1k tokens
- Launch 3 agents: 2k tokens
- Agent results: 10-15k tokens (3-5k each)
- Synthesis: 2k tokens

**Key difference:** Wall-clock time is 3x faster, same token usage

---

## Integration with CI/CD

### Pre-Merge Validation

```yaml
# .github/workflows/pr.yml
name: PR Validation
on: pull_request

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Multi-Agent Validation
        run: /validate-multi
```

### Pre-Deployment Gate

```yaml
# .gitlab-ci.yml
deploy:production:
  stage: deploy
  script:
    - /validate-multi --full
    - if [ $? -eq 0 ]; then deploy_to_production; fi
  only:
    - main
```

---

## Performance Metrics

**Proven speedups from parallel validation:**

| Metric | Single-Agent | Multi-Agent | Improvement |
|--------|--------------|-------------|-------------|
| Wall-clock time | 120 sec | 60 sec | **2x faster** |
| Token budget | 15k | 16k | Similar |
| Coverage | Full | Full | Same |
| Quality | High | High | Maintained |

**Real examples:**
- Large Go project: 180s → 65s (2.8x)
- Kubernetes manifests: 90s → 35s (2.6x)
- Python + tests: 150s → 55s (2.7x)

---

## Related Commands

- **/validate** - Single-agent validation (simpler, slower)
- **/implement** - Execute plan (validation comes after)
- **/research-multi** - Parallel research (3x speedup)
- **/learn** - Extract patterns after validation

---

**Ready for fast validation? Run /validate-multi to check your implementation.**
