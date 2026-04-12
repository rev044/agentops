---
description: Create detailed implementation specification from research
---

# /plan - Detailed Planning Phase

**Purpose:** Specify EVERY change with file:line precision before implementing

**When to use:**
- After research phase (you understand the system)
- Before implementation (need detailed specification)
- Complex changes (multiple files, dependencies)
- Team coordination (others need to review approach)

**Token budget:** 40-60k tokens (20-30% of context window)

**Output:** Plan bundle (1-2k tokens compressed)

---

## The Planning Phase Philosophy

**"Planning IS the work."**

If implementation feels hard, the plan was incomplete. A good plan makes implementation feel mechanical:

**Bad plan:**
```
1. Fix authentication
2. Update tests
3. Deploy
```

**Good plan:**
```
1. Edit auth/middleware.go:45 - Add JWT validation check
   Before: if token != nil {
   After:  if token != nil && validateJWT(token) {

2. Edit auth/middleware.go:12 - Import crypto/jwt package
   Add: import "github.com/golang-jwt/jwt/v4"

3. Create tests/auth_test.go:1-50 - Test JWT validation
   Test cases: valid token, expired token, invalid signature

4. Edit config/auth.yaml:8 - Add JWT secret environment variable
   Add: jwt_secret: ${JWT_SECRET}

5. Validate: go test ./auth/... && go build
```

**Key difference:** The good plan tells you EXACTLY what to change. Implementation becomes copy-paste.

---

## Step 1: Load Research Context

**Planning builds on research. Load your research bundle:**

```bash
/plan [research-bundle-name]

# Or if you just finished research in same session:
/plan ./research-findings.md
```

**I will load:**
- Constitutional foundation (CONSTITUTION.md)
- Your research bundle (500-1k tokens)
- Fresh context (195k tokens available)

**Total context after load:** ~3-5k tokens (1.5-2.5%)

---

## Step 2: Planning Activities

### Activity 1: List All Files to Change

**Based on research findings, identify every file that needs modification:**

```markdown
## Files to Modify

### Files to Edit
1. **auth/middleware.go** - Add JWT validation
   - Lines: 12 (import), 45 (validation check)
   - Type: Code change
   - Risk: Medium (auth is critical)

2. **config/auth.yaml** - Add JWT secret config
   - Lines: 8 (new env var)
   - Type: Configuration
   - Risk: Low (backwards compatible)

### Files to Create
1. **tests/auth_test.go** - JWT validation tests
   - Lines: 1-50 (new file)
   - Type: Test file
   - Risk: Low (tests only)

### Files to Delete
(None for this change)
```

### Activity 2: Specify Exact Changes

**For each file, specify EXACT changes with before/after:**

```markdown
## Change Specifications

### Change 1: Add JWT validation to middleware

**File:** auth/middleware.go:45
**Type:** Function modification

**Before:**
```go
if token != nil {
    ctx.Set("user", token.User)
    return next(ctx)
}
```

**After:**
```go
if token != nil && validateJWT(token) {
    ctx.Set("user", token.User)
    return next(ctx)
}
```

**Rationale:** Current code doesn't validate JWT signature, allowing forged tokens

---

### Change 2: Import JWT validation library

**File:** auth/middleware.go:12
**Type:** Import addition

**Before:**
```go
import (
    "net/http"
    "context"
)
```

**After:**
```go
import (
    "net/http"
    "context"
    "github.com/golang-jwt/jwt/v4"
)
```

**Rationale:** Need JWT library for signature validation

---

### Change 3: Add JWT secret configuration

**File:** config/auth.yaml:8
**Type:** Configuration addition

**Before:**
```yaml
auth:
  enabled: true
  session_timeout: 3600
```

**After:**
```yaml
auth:
  enabled: true
  session_timeout: 3600
  jwt_secret: ${JWT_SECRET}
```

**Rationale:** JWT validation requires secret key from environment
```

### Activity 3: Define Test Strategy

**How will you verify the changes work?**

```markdown
## Test Strategy

### Unit Tests
- **File:** tests/auth_test.go
- **Test cases:**
  1. Valid JWT token → Accept
  2. Expired JWT token → Reject
  3. Invalid signature → Reject
  4. Missing JWT token → Reject
  5. Malformed JWT token → Reject

- **Command:** `go test ./auth/...`
- **Expected:** All tests pass

### Integration Tests
- **Scenario:** Full authentication flow
- **Steps:**
  1. Request JWT token via /login
  2. Use token to access /protected endpoint
  3. Verify access granted

- **Command:** `go test ./integration/...`
- **Expected:** Authentication flow works end-to-end

### Manual Validation
- **Step 1:** Build application (`go build`)
- **Step 2:** Start server (`./app`)
- **Step 3:** Test with curl:
  ```bash
  # Get token
  TOKEN=$(curl -X POST /login -d '{"user":"test","pass":"test"}' | jq -r .token)

  # Use token
  curl -H "Authorization: Bearer $TOKEN" /protected
  ```
- **Expected:** Protected endpoint returns 200 OK

### Rollback Procedure
- **If tests fail:** Revert changes via `git checkout -- .`
- **If production issue:** Deploy previous version
- **Rollback time:** <5 minutes
```

### Activity 4: Identify Dependencies and Order

**What must happen first?**

```markdown
## Implementation Order

### Phase 1: Setup (Dependencies)
1. Install JWT library: `go get github.com/golang-jwt/jwt/v4`
2. Add JWT_SECRET to environment: `export JWT_SECRET=...`
3. Validate dependencies: `go mod tidy`

### Phase 2: Core Changes (Sequential)
1. Add import to middleware.go:12 (prerequisite for step 2)
2. Add validation to middleware.go:45 (uses import from step 1)
3. Add config to auth.yaml:8 (needed for step 2 to work)

### Phase 3: Testing (Parallel possible)
1. Create tests/auth_test.go (can develop alongside core changes)
2. Run unit tests: `go test ./auth/...`
3. Run integration tests: `go test ./integration/...`

### Phase 4: Validation (Sequential)
1. Build application: `go build`
2. Manual smoke test: `./app` and test with curl
3. Deploy to staging
4. Monitor for 10 minutes
5. Deploy to production

**Critical path:** Steps 1-2 in Phase 2 must be sequential. Everything else can be parallel or flexible.
```

### Activity 5: Risk Assessment

**What could go wrong?**

```markdown
## Risk Assessment

### High Risk
**None identified** - This is a well-understood change to a non-critical path

### Medium Risk
**Risk:** JWT validation breaks existing sessions
- **Mitigation:** Add feature flag to enable/disable validation
- **Rollback:** Revert via git if needed
- **Monitor:** Track auth failure rate in metrics

**Risk:** Performance impact from validation overhead
- **Mitigation:** Benchmark validation (should be <1ms)
- **Rollback:** Disable via feature flag
- **Monitor:** Track request latency p95/p99

### Low Risk
**Risk:** Tests are incomplete
- **Mitigation:** Code review before merge
- **Verification:** Coverage report shows >80%

**Risk:** Configuration missing in production
- **Mitigation:** Deployment checklist includes JWT_SECRET
- **Verification:** Startup check fails if JWT_SECRET not set
```

---

## Step 3: Create Implementation Checklist

**Turn your plan into a checklist for implementation:**

```markdown
## Implementation Checklist

### Setup
- [ ] Install JWT library: `go get github.com/golang-jwt/jwt/v4`
- [ ] Run `go mod tidy` to update dependencies
- [ ] Export JWT_SECRET environment variable

### Code Changes
- [ ] Edit auth/middleware.go:12 - Add import
- [ ] Edit auth/middleware.go:45 - Add validation check
- [ ] Edit config/auth.yaml:8 - Add jwt_secret configuration
- [ ] Create tests/auth_test.go - Add test cases

### Testing
- [ ] Run unit tests: `go test ./auth/...`
- [ ] Run integration tests: `go test ./integration/...`
- [ ] Build application: `go build`
- [ ] Manual smoke test with curl

### Validation
- [ ] Code review (if team workflow requires)
- [ ] Commit with semantic message
- [ ] Deploy to staging
- [ ] Monitor for 10 minutes
- [ ] Deploy to production

### Documentation
- [ ] Update auth/README.md with JWT validation details
- [ ] Document JWT_SECRET in deployment guide
- [ ] Add to CHANGELOG.md
```

---

## Step 4: Get Approval

**Before implementing, confirm the plan:**

**For simple changes:**
- Self-review the plan
- Check: Is every change specified?
- Check: Is test strategy clear?
- Proceed to implementation

**For complex changes:**
- Share plan with team
- Review in design doc or PR
- Address feedback
- Get approval before implementing

**For critical changes:**
- Architecture review required
- Security review if auth/authorization
- Performance review if high-traffic path
- Stakeholder sign-off

---

## Step 5: Compress to Plan Bundle

**Save your plan for implementation phase:**

```bash
/bundle-save [topic]-plan --type plan
```

**Bundle includes:**
- All file changes specified
- Test strategy
- Implementation order
- Risk assessment
- Checklist

**Compression:** 40-60k planning → 1-2k token bundle

---

## Planning Patterns

### Pattern 1: Bottom-Up Planning

**Start with smallest change, build up:**

1. Identify atomic changes (single line edits)
2. Group related changes (same file)
3. Order by dependencies (what needs what)
4. Add tests for each change
5. Define validation for each level

**When:** Well-understood problem, clear solution

### Pattern 2: Top-Down Planning

**Start with end goal, decompose:**

1. Define desired outcome (what does success look like?)
2. Identify major components needed
3. Break each component into files
4. Break each file into changes
5. Specify tests for each layer

**When:** Complex problem, multiple approaches possible

### Pattern 3: Risk-First Planning

**Identify risks, mitigate in plan:**

1. List all potential risks
2. Design changes to minimize risk
3. Add feature flags for high-risk changes
4. Plan rollback for each risk
5. Define monitoring for each risk

**When:** Critical systems, production changes

### Pattern 4: Test-First Planning

**Define tests, then changes to make them pass:**

1. Write test cases first (desired behavior)
2. Identify what changes are needed to pass tests
3. Specify exact changes
4. Implement changes
5. Verify tests pass

**When:** Behavior is clear, implementation uncertain

---

## Success Criteria

**Plan is complete when you can answer:**

✅ What files need to change? (every file listed)
✅ What exact changes are needed? (file:line specified)
✅ How will I test? (test strategy defined)
✅ What's the implementation order? (dependencies mapped)
✅ What could go wrong? (risks identified and mitigated)
✅ How do I rollback? (revert plan exists)

**If you can't answer these, your plan is incomplete.**

---

## Common Planning Mistakes

❌ **Vague changes** - "Fix the authentication" → Be specific!
❌ **Missing test strategy** - Hope it works → Define how to verify
❌ **No implementation order** - Start randomly → Follow dependencies
❌ **Ignoring risks** - Assume success → Plan for failure
❌ **No rollback plan** - Cross fingers → Always have an undo

✅ **Do:** Specify every change with file:line precision
✅ **Do:** Define clear test strategy with commands
✅ **Do:** Order changes by dependencies
✅ **Do:** Identify and mitigate risks
✅ **Do:** Plan rollback procedure

---

## Integration with Other Commands

**Before planning:**
```bash
/research [topic]           # Understand system first
/bundle-save [topic]-research # Compress findings

# Then in fresh session:
/plan [topic]-research      # Load research, create plan
```

**After planning:**
```bash
/bundle-save [topic]-plan   # Compress plan

# Get approval (human review)

# Then in fresh session:
/implement [topic]-plan     # Execute the plan
```

**For simple changes (skip planning):**
```bash
/research "simple-change"
/implement [topic]-research  # Direct from research to implementation
```

---

## Examples

### Example 1: Authentication Feature

```bash
# After research phase
/plan redis-caching-research

# I will:
# 1. Load research bundle (1.2k tokens)
# 2. Specify exact changes:
#    - config/redis.yaml:15 - Add connection pool size
#    - app/cache.go:34 - Initialize Redis with pool
#    - app/cache.go:89 - Add health check
# 3. Define test strategy:
#    - Unit: Test pool initialization
#    - Integration: Test cache operations
#    - Load: Test under high concurrency
# 4. Order implementation:
#    - Config first, code second, tests third
# 5. Assess risks:
#    - Pool exhaustion (mitigate with monitoring)
#    - Redis downtime (mitigate with circuit breaker)
# 6. Create plan bundle (1.5k tokens)
#
# Output: redis-caching-plan.md
# Ready for: /implement redis-caching-plan
```

### Example 2: Refactoring

```bash
# After research phase
/plan auth-refactoring-research

# I will:
# 1. Load research bundle (0.9k tokens)
# 2. Specify refactoring:
#    - Move auth/handler.go:50-100 → auth/jwt.go:1-50
#    - Move auth/handler.go:110-150 → auth/session.go:1-40
#    - Update auth/handler.go imports
# 3. Define test strategy:
#    - Existing tests must still pass
#    - No new functionality (pure refactor)
# 4. Order implementation:
#    - Create new files, copy code, update imports, delete old code
# 5. Assess risks:
#    - Tests break during refactor (mitigate with git commits per step)
# 6. Create plan bundle (1.3k tokens)
#
# Output: auth-refactoring-plan.md
```

### Example 3: Infrastructure Change

```bash
# After research phase
/plan kubernetes-upgrade-research

# I will:
# 1. Load research bundle (1.1k tokens)
# 2. Specify upgrade steps:
#    - Update manifests/deployment.yaml:5 - apiVersion: apps/v1
#    - Update manifests/service.yaml:5 - apiVersion: v1
#    - Add manifests/networkpolicy.yaml - New resource
# 3. Define test strategy:
#    - Validation: kubectl apply --dry-run
#    - Staging: Deploy to test cluster
#    - Production: Blue-green deployment
# 4. Order implementation:
#    - Update manifests, validate, test in staging, deploy to prod
# 5. Assess risks:
#    - API version incompatibility (mitigate with dry-run)
#    - Production downtime (mitigate with blue-green)
# 6. Create plan bundle (1.6k tokens)
#
# Output: kubernetes-upgrade-plan.md
```

---

## When to Skip Planning

**Plan is optional for:**

✅ **Trivial changes** - Single line, low risk (typo fix, version bump)
✅ **Well-known patterns** - You've done this 10+ times before
✅ **Emergency fixes** - Production is down, need fix NOW

**In these cases:**
```bash
/prime-simple             # Quick orientation
# Make change directly
# Validate immediately
```

**Plan is REQUIRED for:**

❌ **Complex changes** - Multiple files, dependencies
❌ **Critical systems** - Auth, payment, data loss risk
❌ **Team coordination** - Others need to review
❌ **Unfamiliar territory** - First time doing this
❌ **High risk** - Production impact possible

---

## Multi-Agent Planning (Advanced)

**For very complex planning, use parallel agents:**

```bash
/plan-multi [research-bundle]
```

**This launches 3 agents simultaneously:**
- Agent 1: Specify file changes
- Agent 2: Design test strategy
- Agent 3: Assess risks and rollback

**Result:** 3x faster planning, comprehensive specification

**See:** (Future command, not yet implemented)

---

## Token Budget Management

**Planning phase target:** 20-30% of context window (40-60k tokens)

**Breakdown:**
- Load research bundle: 1k tokens
- Specify changes: 20-30k tokens
- Design tests: 10-15k tokens
- Risk assessment: 5-10k tokens
- Create checklist: 2-5k tokens

**If approaching 40%:**

```bash
# Option 1: Checkpoint progress
/bundle-save [topic]-plan-partial --type plan

# Option 2: Simplify plan scope
# Focus on critical changes, defer nice-to-haves

# Option 3: Move to implementation
# Good enough plan → Start implementing
```

---

## Related Commands

- **/prime-complex** - Load constitutional foundation
- **/research** - Understand system before planning
- **/bundle-load** - Resume planning in fresh session
- **/bundle-save** - Compress plan for implementation
- **/implement** - Execute the plan
- **/validate** - Verify implementation matches plan

---

**Ready to plan? Load your research bundle or describe what you're planning.**
