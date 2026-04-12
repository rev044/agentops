---
description: Execute approved plan with validation and auto-checkpoint
---

# /implement - Implementation Phase

**Purpose:** Execute approved plan mechanically with continuous validation

**When to use:**
- After planning phase (detailed spec exists)
- Approved plan ready for execution
- Changes are well-specified (file:line precision)
- Test strategy is defined

**Token budget:** 40-80k tokens (20-40% of context window)

**Output:** Working implementation + commit with learnings

---

## The Implementation Philosophy

**"Trust the plan. Don't redesign during implementation."**

If the plan is good, implementation feels mechanical:
- Read change specification
- Make the exact change
- Validate immediately
- Move to next change

**If implementation feels hard:**
- The plan was incomplete (missing details)
- The plan was wrong (wrong approach)
→ STOP, return to planning phase

**Good implementation:**
```
Load plan → Execute change 1 → Validate → Execute change 2 → Validate → Done
```

**Bad implementation:**
```
Load plan → Redesign approach → Try change → Fails → Redesign again → ...
```

---

## Step 1: Load Plan Context

**Implementation executes a plan. Load your plan bundle:**

```bash
/implement [plan-bundle-name]

# Or resume from checkpoint:
/implement --resume [implementation-progress-bundle]
```

**I will load:**
- Constitutional foundation (CONSTITUTION.md)
- Your plan bundle (1-2k tokens)
- Fresh context (196k tokens available)

**Total context after load:** ~3-4k tokens (1.5-2%)

---

## Step 2: Verify Plan Quality

**Before starting, I check if the plan is implementable:**

**Quality checks:**
✅ All files are specified (paths exist or will be created)
✅ Changes are detailed (file:line with before/after)
✅ Test strategy is defined (commands to run)
✅ Implementation order is clear (what comes first)
✅ Rollback plan exists (how to undo)

**If plan fails quality check:**
```
⚠️ Plan quality issue detected:
- Missing: Test strategy undefined
- Suggestion: Return to /plan and add test specification

Proceed anyway? (not recommended)
```

**Your options:**
1. Return to `/plan [topic]-research` to improve plan
2. Proceed anyway (risky - implementation may fail)

---

## Step 3: Execute Changes Sequentially

**I follow the plan's implementation order:**

### For Each Change:

**1. Read specification from plan**
```markdown
Change 3: Add JWT validation
File: auth/middleware.go:45
Before: if token != nil {
After:  if token != nil && validateJWT(token) {
```

**2. Make the exact change**
```bash
# I use Edit tool
Edit file: auth/middleware.go
Old: if token != nil {
New: if token != nil && validateJWT(token) {
```

**3. Validate immediately**
```bash
# Run validation from plan's test strategy
go build ./auth/...
# ✅ Build succeeds

go test ./auth/...
# ✅ Tests pass
```

**4. Document progress**
```markdown
✅ Change 3 complete: JWT validation added
   - File: auth/middleware.go:45
   - Validation: Build + tests passed
   - Time: 2 minutes
```

**5. Checkpoint if needed (auto)**
```
Context check: 35k tokens (17.5%) - OK, continue
```

**6. Move to next change**

---

## Step 4: Auto-Checkpointing

**If context approaches 40% during implementation:**

```
⚠️ Context at 78k tokens (39%) - Approaching threshold

Auto-checkpointing:
1. Saving implementation progress bundle
2. Current state: 5 of 12 changes complete
3. Resume with: /implement --resume [bundle-name]

Bundle saved: [topic]-implementation-progress.md
```

**What gets saved:**
- Changes completed so far
- Changes remaining
- Current git state (commit SHA, branch)
- Validation results
- Token budget used

**Resume in next session:**
```bash
/implement --resume [topic]-implementation-progress
```

**I will:**
- Load progress bundle (2-3k tokens)
- Verify git state matches checkpoint
- Continue from where you left off

---

## Step 5: Continuous Validation

**After each change, I run validation from plan:**

### Validation Types

**1. Syntax Validation (Immediate)**
```bash
# For YAML
yamllint file.yaml

# For Go
go build ./...

# For Python
python -m py_compile file.py

# For Shell
bash -n script.sh
```

**2. Unit Tests (Per Change)**
```bash
# Run tests related to changed file
go test ./path/to/changed/...
pytest tests/test_feature.py
```

**3. Integration Tests (After Multiple Changes)**
```bash
# Run broader test suite
make test
npm test
go test ./...
```

**4. Full Validation (Before Commit)**
```bash
# Run complete validation suite
make ci-all
./run-all-tests.sh
```

**Validation failure:**
```
❌ Validation failed: tests/auth_test.go:34 - JWT validation test failed

Options:
1. Debug and fix (recommended)
2. Rollback last change
3. Skip validation (NOT recommended)
```

---

## Step 6: Commit with Context

**When all changes complete and validation passes:**

```bash
git add .
git commit -m "$(cat <<'EOF'
<type>(<scope>): <subject>

## Context
[Why was this change needed? What problem does it solve?]

## Solution
[What was implemented? How does it work?]

## Learning
[What patterns were discovered? What would you do differently?]

## Impact
[What's the measured improvement? Time saved? Quality gained?]

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

**Semantic commit format:**
- `feat(auth): Add JWT signature validation`
- `fix(cache): Increase Redis connection pool size`
- `refactor(handlers): Extract JWT logic to separate package`
- `docs(readme): Document JWT configuration`

---

## Implementation Patterns

### Pattern 1: Sequential Execution (Default)

**Follow plan order exactly:**

1. Execute change 1 → Validate
2. Execute change 2 → Validate
3. Execute change 3 → Validate
4. Complete → Full validation → Commit

**When:** Changes have dependencies (A requires B)

### Pattern 2: Batch & Validate

**Group related changes, validate together:**

1. Execute changes 1-3 (same file)
2. Validate all 3 together
3. Execute changes 4-6 (different file)
4. Validate all 3 together
5. Complete → Full validation → Commit

**When:** Changes are independent within groups

### Pattern 3: Test-Driven Implementation

**Write tests first, then implement:**

1. Create test file (from plan)
2. Run tests → Should fail (no implementation)
3. Implement change
4. Run tests → Should pass
5. Repeat for each change

**When:** Plan uses test-first approach

### Pattern 4: Incremental with Rollback Points

**Commit after each major change:**

1. Execute change 1 → Validate → Commit
2. Execute change 2 → Validate → Commit
3. Execute change 3 → Validate → Commit
4. Complete → Full validation

**When:** High-risk changes (easy rollback needed)

---

## Handling Implementation Issues

### Issue: Change Fails Validation

**Problem:** Test fails after making planned change

**Response:**
1. **Check:** Did I make the change exactly as planned?
   - If no: Fix the mistake, try again
   - If yes: Plan was wrong

2. **If plan was wrong:**
   - STOP implementation
   - Save progress bundle: `/bundle-save [topic]-implementation-partial`
   - Return to planning: `/plan [topic]-research --revise`
   - Update plan with correct approach
   - Resume implementation: `/implement [topic]-plan-revised`

### Issue: Context Fills During Implementation

**Problem:** Context approaching 40% threshold

**Response (automatic):**
```
⚠️ Context at 76k tokens (38%)

Auto-checkpointing in progress...
✅ Saved: [topic]-implementation-progress.md

Resume in next session with:
/implement --resume [topic]-implementation-progress

Current state: 7 of 15 changes complete
Remaining: 8 changes (~30k tokens estimated)
```

**You continue in fresh session:**
```bash
# New session, fresh context
/implement --resume [topic]-implementation-progress

# I load progress bundle (2.5k tokens)
# Continue from change 8
```

### Issue: Plan Was Incomplete

**Problem:** Plan says "Fix authentication" (too vague)

**Response:**
```
⚠️ Plan quality issue: Change specification too vague

Change 5: "Fix authentication"
Missing: File path, line numbers, exact changes

Cannot implement. Return to /plan to add detail.
```

**You must:**
1. `/plan [topic]-research --revise`
2. Add specific details (file:line, before/after)
3. Approve revised plan
4. `/implement [topic]-plan-revised`

### Issue: Unexpected Dependency Found

**Problem:** Change requires something not in plan

**Response:**
```
⚠️ Dependency discovered: Change 3 requires JWT library

Plan missing: go get github.com/golang-jwt/jwt/v4

Options:
1. Add dependency now (recommended)
2. Update plan for next time (document learning)
```

**I will:**
1. Add missing dependency
2. Document in commit message (learning section)
3. Continue implementation

---

## Success Criteria

**Implementation succeeds when:**

✅ All changes from plan executed
✅ All validation passes (syntax, unit, integration)
✅ Full test suite passes
✅ Build succeeds
✅ Manual smoke test passes (if in plan)
✅ Commit created with context/solution/learning/impact
✅ Ready for deployment (if applicable)

**Implementation does NOT:**
- Redesign during execution (trust the plan)
- Skip validation (always verify)
- Make unplanned changes (stick to spec)
- Commit without testing (validate first)

---

## Token Budget Management

**Implementation phase target:** 20-40% of context window (40-80k tokens)

**Breakdown:**
- Load plan bundle: 1.5k tokens
- Execute changes: 30-60k tokens
- Validation output: 5-10k tokens
- Progress tracking: 2-5k tokens
- Documentation: 2-5k tokens

**If approaching 40%:**

```bash
# Automatic checkpoint triggers
Context: 78k tokens (39%)
→ Auto-save progress bundle
→ Resume in next session

# Or manually checkpoint:
/bundle-save [topic]-implementation-checkpoint --type implementation
```

---

## Multi-Session Implementation

**For large implementations (>40% context):**

**Session 1:**
```bash
/implement [topic]-plan
# Complete changes 1-7
# Context at 38% (76k tokens)
# Auto-checkpoint: [topic]-implementation-progress.md
```

**Session 2:**
```bash
/implement --resume [topic]-implementation-progress
# Load bundle (2.5k tokens) - Know state of changes 1-7
# Complete changes 8-15
# Full validation
# Commit with complete context
```

**Key advantage:** Fresh context = better decisions, no degradation

---

## Resume from Checkpoint

**When resuming implementation:**

**I verify git state matches checkpoint:**
```bash
# Checkpoint saved at: commit abc123, branch main, 7 changes done

# Resume checks:
✅ Current branch: main (matches)
✅ Current commit: abc123 (matches)
✅ Working directory: clean (no uncommitted changes)

# Safe to resume
Loading progress: 7 of 15 changes complete
Continuing from: Change 8 (Add test coverage)
```

**If git state doesn't match:**
```
⚠️ Git state mismatch

Checkpoint: commit abc123, branch main
Current:    commit def456, branch feature-x

Cannot safely resume. Options:
1. Checkout original state: git checkout abc123
2. Start fresh: /implement [topic]-plan
3. Manual reconciliation (advanced)
```

---

## Integration with Other Commands

**Before implementation:**
```bash
/research [topic]           # Understand
/plan [topic]-research      # Specify
# Get approval
/implement [topic]-plan     # Execute
```

**During implementation:**
```bash
# Auto-checkpoint if context fills
# Resume with:
/implement --resume [topic]-implementation-progress
```

**After implementation:**
```bash
/validate                   # Full validation pass
/learn [topic]              # Extract patterns for future
```

---

## Examples

### Example 1: Simple Implementation

```bash
/implement redis-caching-plan

# I will:
# 1. Load plan bundle (1.5k tokens)
# 2. Execute change 1: Edit config/redis.yaml:15
#    → Add pool_size: 100
#    → Validate: yamllint config/redis.yaml ✅
# 3. Execute change 2: Edit app/cache.go:34
#    → Add pool initialization
#    → Validate: go build ./app/... ✅
# 4. Execute change 3: Create tests/cache_test.go
#    → Add test cases
#    → Validate: go test ./tests/... ✅
# 5. Full validation: make test ✅
# 6. Commit with context/solution/learning
#
# Total: 45k tokens (22.5%), single session
# Result: feat(cache): Increase Redis connection pool size
```

### Example 2: Multi-Session Implementation

```bash
# Session 1
/implement kubernetes-upgrade-plan

# Changes 1-10 complete (35 files modified)
# Context at 76k tokens (38%)
# Auto-checkpoint: k8s-upgrade-progress.md

# Session 2 (next day)
/implement --resume k8s-upgrade-progress

# Load bundle (2.3k tokens)
# Know state: Changes 1-10 done, 35 files modified
# Continue: Changes 11-20
# Complete: Full validation, commit
#
# Total: Session 1 (76k) + Session 2 (55k) = 131k across 2 sessions
# Average: 65k per session (32.5%) - sustainable!
```

### Example 3: Implementation with Issues

```bash
/implement auth-refactor-plan

# Change 1-3: Complete ✅
# Change 4: Execute
#   → Edit auth/handler.go:50
#   → Validate: go test ./auth/...
#   ❌ Tests fail: undefined: validateJWT

# Plan was incomplete (missing function definition)
# STOP implementation
# Save progress: auth-refactor-partial.md

# Return to planning
/plan auth-refactor-research --revise
# Add: Create auth/jwt.go with validateJWT function

# Resume with revised plan
/implement auth-refactor-plan-revised
# Changes 1-3: Already done (from checkpoint)
# Change 4 (revised): Create auth/jwt.go first
# Change 5: Edit auth/handler.go:50 (now works!)
# Complete ✅
```

---

## When to Skip to Implementation (No Plan)

**Implementation without explicit plan:**

```bash
# For trivial changes
/prime-simple
# Make change directly
# Validate
# Commit
```

**OK for:**
- Single line changes (typo, version bump)
- Well-known patterns (done 10+ times)
- Very low risk (comments, docs)

**NOT OK for:**
- Multiple files
- Complex logic
- Critical systems
- Unfamiliar territory

---

## Related Commands

- **/prime-complex** - Load constitutional foundation
- **/research** - Understand before planning
- **/plan** - Create detailed specification
- **/bundle-load** - Load plan in fresh session
- **/bundle-save** - Checkpoint implementation progress
- **/validate** - Full validation pass after implementation
- **/learn** - Extract patterns after implementation

---

**Ready to implement? Load your plan bundle or resume from checkpoint.**
