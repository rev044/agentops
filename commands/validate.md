---
description: Run comprehensive validation suite after implementation
---

# /validate - Comprehensive Validation

**Purpose:** Verify implementation quality before deployment or commit

**When to use:**
- After implementation (changes complete)
- Before committing (final check)
- Before deployment (production readiness)
- Periodic validation (CI/CD pipeline)

**Token budget:** 10-20k tokens (5-10% of context window)

**Output:** Validation report with pass/fail status

---

## The Validation Philosophy

**"Trust, but verify."**

Implementation may look correct, but validation proves it:
- Syntax is correct (no typos, proper formatting)
- Tests pass (behavior is correct)
- Build succeeds (no compilation errors)
- Style is consistent (linting passes)
- Security is sound (no vulnerabilities)
- Performance is acceptable (no regressions)

**Validation catches:**
- Mistakes during implementation
- Edge cases not covered by tests
- Integration issues
- Performance degradation
- Security vulnerabilities

---

## Validation Levels

### Level 1: Syntax Validation (Fast)

**Goal:** Catch basic errors (5-10 seconds)

**For YAML:**
```bash
yamllint .
find . -name "*.yaml" -o -name "*.yml" | xargs yamllint
```

**For Go:**
```bash
go fmt ./...
go vet ./...
gofmt -l .
```

**For Python:**
```bash
python -m py_compile **/*.py
flake8 .
black --check .
```

**For Shell:**
```bash
find . -name "*.sh" | xargs bash -n
shellcheck **/*.sh
```

**For Markdown:**
```bash
markdownlint .
```

**For JSON:**
```bash
find . -name "*.json" | xargs jq empty
```

### Level 2: Unit Tests (Medium)

**Goal:** Verify component behavior (30-60 seconds)

**For Go:**
```bash
go test ./... -v
go test ./... -cover
```

**For Python:**
```bash
pytest tests/ -v
pytest --cov=src tests/
```

**For JavaScript:**
```bash
npm test
npm run test:coverage
```

**For Kubernetes:**
```bash
# Validate manifests
kubectl apply --dry-run=client -f manifests/
kustomize build . | kubectl apply --dry-run=server -f -
```

### Level 3: Integration Tests (Slow)

**Goal:** Verify system behavior (2-5 minutes)

**For API:**
```bash
# Run integration test suite
make test-integration
pytest tests/integration/ -v
```

**For Services:**
```bash
# Start services, run tests, stop services
docker-compose up -d
pytest tests/integration/
docker-compose down
```

**For End-to-End:**
```bash
# Full system test
make test-e2e
npm run test:e2e
```

### Level 4: Security Validation

**Goal:** Identify vulnerabilities

**For Dependencies:**
```bash
# Go
go list -json -m all | nancy sleuth

# Python
pip-audit
safety check

# JavaScript
npm audit
yarn audit
```

**For Secrets:**
```bash
# Detect secrets in code
trufflehog filesystem .
gitleaks detect
```

**For Containers:**
```bash
# Scan container images
trivy image myapp:latest
```

### Level 5: Performance Validation

**Goal:** Ensure no regressions

**For Load Testing:**
```bash
# HTTP load test
ab -n 1000 -c 10 http://localhost:8080/

# Custom load test
go test -bench=. -benchmem
pytest tests/performance/ --benchmark-only
```

**For Profiling:**
```bash
# CPU profiling
go test -cpuprofile=cpu.prof
python -m cProfile script.py
```

---

## Full Validation Suite

**I will run all applicable validations for your project:**

### Step 1: Detect Project Type

**I analyze your project structure:**

```bash
# Check for files
ls go.mod         # Go project
ls package.json   # JavaScript/Node project
ls requirements.txt # Python project
ls kustomization.yaml # Kubernetes project
```

### Step 2: Run Appropriate Validations

**For each detected type, I run validation:**

**Go Project:**
```bash
✓ Syntax: go fmt, go vet
✓ Build: go build ./...
✓ Tests: go test ./...
✓ Coverage: go test -cover ./...
✓ Linting: golangci-lint run
```

**Python Project:**
```bash
✓ Syntax: python -m py_compile
✓ Tests: pytest
✓ Coverage: pytest --cov
✓ Linting: flake8, black --check
✓ Types: mypy
```

**Kubernetes Project:**
```bash
✓ Syntax: yamllint
✓ Validation: kubectl apply --dry-run
✓ Build: kustomize build
✓ Security: kubesec scan
```

### Step 3: Generate Report

**I create validation report:**

```markdown
# Validation Report

**Date:** 2025-11-07 10:30
**Project:** agentops
**Commit:** abc123
**Status:** ✅ PASSED

## Summary
- Syntax: ✅ PASSED (0 errors)
- Build: ✅ PASSED
- Tests: ✅ PASSED (45/45)
- Coverage: ✅ PASSED (87% - meets 80% threshold)
- Linting: ⚠️ WARNING (3 style issues)
- Security: ✅ PASSED (0 vulnerabilities)

## Details

### Syntax Validation
✅ YAML: 127 files validated, 0 errors
✅ Go: All files formatted correctly
✅ Shell: 15 scripts validated

### Build
✅ go build ./... - SUCCESS (2.3s)

### Tests
✅ Unit tests: 45/45 passed (12.5s)
   - auth: 12/12 passed
   - cache: 8/8 passed
   - handlers: 15/15 passed
   - middleware: 10/10 passed

### Coverage
✅ Total: 87% (threshold: 80%)
   - auth: 92%
   - cache: 78% ⚠️ (below 80%)
   - handlers: 91%
   - middleware: 88%

### Linting
⚠️ 3 issues found:
   - auth/handler.go:45 - Line too long (golangci-lint)
   - cache/redis.go:12 - Unused variable 'ctx'
   - middleware/jwt.go:67 - Error not handled

### Security
✅ No vulnerabilities detected
✅ No secrets exposed
✅ Dependencies up to date

## Recommendations
1. Fix 3 linting issues
2. Increase cache package coverage to 80%+
3. Consider adding integration tests

## Action Required
⚠️ Fix linting issues before deployment
✅ OK to commit (linting is low priority)
```

---

## Validation Commands

### Basic Syntax (Fast Check)

```bash
/validate --quick

# Runs: Syntax only (5-10 seconds)
# Use: Before commit, fast feedback
```

### Standard Validation (Default)

```bash
/validate

# Runs: Syntax + Build + Tests (30-60 seconds)
# Use: After implementation, before commit
```

### Full Validation (Comprehensive)

```bash
/validate --full

# Runs: Everything including security, performance (5-10 minutes)
# Use: Before deployment, periodic checks
```

### Specific Validation

```bash
# Only syntax
/validate --syntax

# Only tests
/validate --tests

# Only security
/validate --security

# Only coverage
/validate --coverage
```

---

## Integration with CI/CD

**Validation in pipelines:**

### Pre-Commit Hook

```bash
# .git/hooks/pre-commit
/validate --quick
# Blocks commit if syntax fails
```

### Pull Request Check

```bash
# .github/workflows/pr.yml
- name: Validate
  run: /validate --full

# Blocks merge if validation fails
```

### Deployment Gate

```bash
# Before production deploy
/validate --full
# Requires all checks to pass
```

---

## Handling Validation Failures

### Syntax Errors

**Problem:** YAML syntax error detected

**Response:**
```
❌ Syntax validation failed

File: manifests/deployment.yaml:15
Error: Unexpected indent

Fix:
- Review file at line 15
- Correct indentation
- Run /validate --syntax to verify
```

### Test Failures

**Problem:** Unit tests failing

**Response:**
```
❌ Tests failed: 3/45

Failed tests:
- auth/jwt_test.go:34 - TestValidateToken
- cache/redis_test.go:56 - TestConnectionPool
- handlers/api_test.go:89 - TestHealthCheck

Debug:
1. Run failing tests: go test -v ./auth ./cache ./handlers
2. Review test output
3. Fix implementation
4. Re-run /validate
```

### Coverage Too Low

**Problem:** Coverage below threshold

**Response:**
```
⚠️ Coverage: 72% (threshold: 80%)

Low coverage packages:
- cache: 65% (needs 15% more)
- middleware: 74% (needs 6% more)

Recommendation:
1. Add tests for cache package
2. Add tests for middleware package
3. Re-run /validate --coverage
```

### Security Vulnerabilities

**Problem:** Vulnerabilities detected

**Response:**
```
❌ Security validation failed: 2 vulnerabilities

HIGH: CVE-2023-12345 in dependency X v1.2.3
- Impact: Remote code execution
- Fix: Upgrade to v1.2.4

MEDIUM: Exposed secret in config/auth.yaml:8
- Impact: Credentials leak
- Fix: Move to environment variable

Action required:
1. Upgrade dependency X
2. Remove secret from config
3. Re-run /validate --security
```

---

## Success Criteria

**Validation passes when:**

✅ All syntax checks pass
✅ Build succeeds
✅ All tests pass
✅ Coverage meets threshold (typically 80%)
✅ Linting has no errors (warnings acceptable)
✅ No security vulnerabilities
✅ Performance is acceptable

**Validation BLOCKS when:**

❌ Syntax errors (cannot build)
❌ Tests fail (behavior broken)
❌ Critical security vulnerabilities
❌ Coverage drops significantly (>10%)

**Validation WARNS when:**

⚠️ Linting issues (style, not correctness)
⚠️ Coverage slightly below threshold (<5%)
⚠️ Minor security issues (low severity)
⚠️ Performance regression (small)

---

## Token Budget Management

**Validation phase target:** 5-10% of context window (10-20k tokens)

**Breakdown:**
- Project detection: 1k tokens
- Run validations: 5-10k tokens
- Parse results: 2-5k tokens
- Generate report: 2-3k tokens

**If validation output is large:**

```bash
# Option 1: Filter output
/validate --quiet  # Show only failures

# Option 2: Save to file
/validate > validation-report.txt
# Review file instead of context
```

---

## Multi-Agent Validation (Advanced)

**For large projects, use parallel validation:**

```bash
/validate-multi
```

**This launches 3 agents simultaneously:**
- Agent 1: Syntax + Build validation
- Agent 2: Test execution + Coverage
- Agent 3: Security + Performance

**Result:** 3x faster validation (3 minutes instead of 9)

**See:** `/validate-multi` command for details

---

## Common Validation Patterns

### Pattern 1: Pre-Commit Validation

```bash
# Before every commit
/validate --quick

# If pass: Commit
# If fail: Fix, retry
```

### Pattern 2: Continuous Validation

```bash
# Watch mode (re-validate on file change)
/validate --watch

# Provides immediate feedback during development
```

### Pattern 3: Comprehensive Pre-Deploy

```bash
# Before deployment
/validate --full

# Must pass 100% before deploying
```

### Pattern 4: Incremental Validation

```bash
# Validate only changed files
/validate --changed

# Faster feedback during development
```

---

## Integration with Other Commands

**After implementation:**
```bash
/implement [topic]-plan     # Complete implementation
/validate                   # Verify quality
# If pass: commit and done
# If fail: fix and re-validate
```

**Before deployment:**
```bash
/validate --full           # Comprehensive check
# If pass: deploy
# If fail: fix before deploying
```

**Periodic validation:**
```bash
# Daily or weekly
/validate --full
# Catch regressions early
```

---

## Examples

### Example 1: Post-Implementation Validation

```bash
/validate

# I detect: Go project
# I run:
# ✅ go fmt ./... - No changes needed
# ✅ go vet ./... - No issues
# ✅ go build ./... - Build succeeds (2.1s)
# ✅ go test ./... - 45/45 tests passed (11.2s)
# ✅ go test -cover ./... - 87% coverage (threshold: 80%)
#
# Report: ✅ PASSED - OK to commit
```

### Example 2: Validation with Issues

```bash
/validate

# I detect: Python + Kubernetes project
# I run:
# ✅ YAML validation - 127 files, 0 errors
# ❌ pytest - 42/45 passed, 3 failed
#    - tests/auth_test.py:34 - FAILED
#    - tests/cache_test.py:56 - FAILED
#    - tests/api_test.py:89 - FAILED
# ⚠️ Coverage - 76% (threshold: 80%)
#
# Report: ❌ FAILED - Fix tests before commit
```

### Example 3: Security Validation

```bash
/validate --security

# I run:
# ✅ Secret detection - No secrets found
# ❌ Dependency scan - 2 vulnerabilities
#    - HIGH: CVE-2023-12345 in package X
#    - MEDIUM: CVE-2023-67890 in package Y
# ✅ Container scan - No issues
#
# Report: ❌ FAILED - Upgrade dependencies
```

---

## Related Commands

- **/implement** - Execute plan (validation comes after)
- **/validate-multi** - Parallel validation (3x faster)
- **/learn** - Extract patterns after validation passes
- **/prime-simple** - For trivial changes (inline validation)

---

**Ready to validate? Run /validate to check your implementation.**
