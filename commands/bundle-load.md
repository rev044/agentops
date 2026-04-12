# /bundle-load - Restore Session Context from Bundle

**Purpose:** Load compressed bundle to resume work or reference knowledge

**Philosophy:** Multi-session work via context compression and restoration

**Token budget:** 1-3k tokens (bundle size)

**Output:** Context restored, ready to continue

---

## When to Use

Use `/bundle-load` to:

### 1. Continue Research → Plan
```
Session 1: /research → /bundle-save research
Session 2: /bundle-load research → /plan
```

### 2. Continue Plan → Implement
```
Session 1: /plan → /bundle-save plan
Session 2: /bundle-load plan → /implement
```

### 3. Resume Mid-Implementation
```
Session 1: /implement → context fills → /bundle-save progress
Session 2: /bundle-load progress → /implement --resume
```

### 4. Reference Previous Work
```
Working on similar task → /bundle-load previous-solution
Use patterns without re-research
```

---

## How It Works

### Basic Usage
```bash
/bundle-load {name}

# Examples:
/bundle-load redis-caching-research
/bundle-load api-optimization-plan
/bundle-load containerization-progress
```

### Step 1: Find Bundle
```
Searching for: redis-caching-research

Locations checked:
1. .agentops/bundles/redis-caching-research.md ✅ Found
2. Alternative names (fuzzy match)
3. UUID if provided

Found: redis-caching-research.md
Size: 1.9k tokens
Type: research
Created: 2025-11-07
```

### Step 2: Load Into Context
```
Loading bundle (1.9k tokens)...

Bundle content loaded:
- Executive summary
- Approaches evaluated
- Recommendation
- Constraints
- Next steps

Current context: 1.9k tokens (0.95%)
Safe to continue: YES
```

### Step 3: Verify Compatibility
```
Bundle check:
✅ Type matches phase (research bundle for planning)
✅ No missing dependencies
✅ Token budget allows continuation
✅ Git state compatible (if implementation bundle)

Ready to proceed
```

---

## Loading Different Bundle Types

### Research Bundle → Planning
```
/bundle-load redis-caching-research

Loaded (1.9k tokens):
- Recommendation: Redis with pub/sub
- Constraints: 500MB memory, circuit breaker needed
- Pattern: Service discovery integration

Ready for: /plan redis-caching-research.md
```

### Plan Bundle → Implementation
```
/bundle-load redis-caching-plan

Loaded (1.5k tokens):
- 5 files to create
- 3 files to modify
- Validation: make quick, integration tests
- Rollback: git revert procedure

Ready for: /implement redis-caching-plan.md
```

### Implementation Progress → Resume
```
/bundle-load redis-caching-progress

Loaded (2.5k tokens):
- Completed: 7 files (validated)
- Remaining: 13 files
- Git state: 7 files staged
- Estimated: 60k tokens to complete

Verifying git state...
✅ Git matches bundle (7 files staged)

Ready for: /implement --resume
```

---

## Resume Implementation Example

**Scenario:** Context filled mid-implementation

### Session 1: Save Progress
```
/implement redis-caching-plan.md

[Created 7 of 20 files]
[Context at 90k tokens, 45%]

/bundle-save redis-caching-progress --type implementation

Progress saved:
- 7 changes complete
- 13 changes remaining
- Bundle: 2.5k tokens

[End Session 1]
```

### Session 2: Resume
```
/prime-complex

/bundle-load redis-caching-progress

Loading implementation progress bundle (2.5k tokens)...

Progress loaded:
✅ 7 files created (validated)
✅ 7 files staged in git
⏸️ 13 changes remaining

Verifying git state...
Expected: 7 staged files
Actual: 7 staged files ✅ MATCH

Safe to resume: YES

Ready for: /implement --resume

[Agent continues from where it left off]
[Creates remaining 13 files]
[Validates, commits, pushes]

Implementation complete ✅
```

---

## Git State Verification

### For Implementation Bundles

**Why verify:**
- Ensure no changes since bundle created
- Prevent conflicts
- Confirm safe resumption

**What's checked:**
```
Bundle says:
- 7 files staged
- 0 files unstaged
- Branch: main

Git status shows:
- 7 files staged ✅
- 0 files unstaged ✅
- Branch: main ✅

Verification: PASS
```

**If mismatch:**
```
Bundle says: 7 files staged
Git shows: 5 files staged, 2 modified

⚠️ MISMATCH DETECTED

Options:
1. Abort load (safest)
2. Show diff (what changed)
3. Force load (risky)

Recommended: Abort, investigate difference
```

---

## Token Budget Management

### Before Loading
```
Current context: 10k tokens (5%)
Bundle size: 2.5k tokens
After load: 12.5k tokens (6.25%)

Safe to load: YES ✅
```

### If Near Threshold
```
Current context: 75k tokens (37.5%)
Bundle size: 2.5k tokens
After load: 77.5k tokens (38.75%)

⚠️ Approaching threshold

Recommendation: Proceed with caution
Estimated remaining work: 45k tokens
Projected total: 122.5k tokens (61%) ⚠️ EXCEEDS

Suggest: Save progress earlier if possible
```

---

## Multi-Bundle Loading

### Load Multiple Related Bundles
```
/bundle-load redis-caching-research

Loaded: 1.9k tokens

Related bundles available:
- redis-caching-plan (1.5k)
- redis-cluster-failover-pattern (1.2k)

Want to load more?

User: Yes, load redis-cluster-failover-pattern

Total context: 3.1k tokens (1.55%)
Safe to continue ✅
```

### Load Bundle Chain
```
Working on complex feature spanning multiple sessions:

Session 1: Research
  /research → /bundle-save feature-research

Session 2: Plan Phase 1
  /bundle-load feature-research
  /plan → /bundle-save feature-plan-phase1

Session 3: Implement Phase 1
  /bundle-load feature-plan-phase1
  /implement → /bundle-save feature-impl-phase1

Session 4: Plan Phase 2
  /bundle-load feature-impl-phase1  # Context of what's done
  /plan → /bundle-save feature-plan-phase2

Session 5: Implement Phase 2
  /bundle-load feature-plan-phase2
  /implement → Complete
```

---

## Bundle Discovery

### By Name (Exact)
```
/bundle-load redis-caching-research

✅ Found exact match
```

### By Name (Fuzzy)
```
/bundle-load redis-cache

Found multiple matches:
1. redis-caching-research (1.9k, 2025-11-07)
2. redis-caching-plan (1.5k, 2025-11-07)
3. redis-cluster-pattern (1.2k, 2025-11-05)

Which one? [1/2/3]
```

### By UUID
```
/bundle-load bundle-abc123def456

✅ Found by UUID
Loading: redis-caching-research.md
```

### By Tag
```
/bundle-load --tag redis --tag performance

Found 3 bundles with tags [redis, performance]:
1. redis-caching-research (1.9k)
2. api-redis-integration (1.4k)
3. redis-cluster-ha (1.7k)

Which one?
```

### Recent Bundles
```
/bundle-load --recent

5 most recently accessed:
1. redis-caching-progress (accessed 1 hour ago)
2. k8s-migration-plan (accessed 3 hours ago)
3. api-refactor-research (accessed 1 day ago)

Which one?
```

---

## Success Criteria

Bundle-load is successful when:
- ✅ Bundle found and loaded
- ✅ Context under 10% after load
- ✅ Git state verified (for implementation bundles)
- ✅ No dependency issues
- ✅ Ready to continue work

---

## Error Handling

### Bundle Not Found
```
/bundle-load nonexistent-bundle

❌ Bundle not found: nonexistent-bundle

Searched:
- .agentops/bundles/nonexistent-bundle.md
- Fuzzy matches
- UUID index

Suggestions:
- List all bundles: /bundle-list
- Check spelling
- Verify bundle was saved
```

### Git State Mismatch
```
/bundle-load redis-caching-progress

Bundle loaded but git state mismatch:

Expected: 7 files staged
Actual: 5 files staged, 2 modified

⚠️ Cannot safely resume

Action: Investigate what changed
- git status
- git diff
- Review recent commits

Resolution: Either:
1. Restore git state to match bundle
2. Create new bundle from current state
```

### Token Budget Exceeded
```
/bundle-load large-bundle

Current: 75k tokens (37.5%)
Bundle: 5k tokens
After: 80k tokens (40%)

⚠️ Would reach threshold

Options:
1. Load anyway (risky)
2. Load smaller bundle
3. Archive current context first
4. Start fresh session

Recommended: Start fresh session
```

---

## Commands Reference

```bash
# Load by name
/bundle-load {name}

# Load by UUID
/bundle-load bundle-abc123def456

# Load with filters
/bundle-load --tag redis
/bundle-load --recent
/bundle-load --type implementation

# Load and verify
/bundle-load {name} --verify-git

# List all bundles
/bundle-list
```

---

## Integration with Workflow

### Complete Cycle
```
Session 1: Research
  /prime-complex
  /research
  /bundle-save research
  [End session]

Session 2: Plan
  /prime-complex
  /bundle-load research ← Loads 1.9k
  /plan
  /bundle-save plan
  [End session]

Session 3: Implement
  /prime-complex
  /bundle-load plan ← Loads 1.5k
  /implement
  [Context fills at 40%]
  /bundle-save progress
  [End session]

Session 4: Resume
  /prime-complex
  /bundle-load progress ← Loads 2.5k
  /implement --resume
  Complete ✅
```

---

*Enables: Multi-session work, knowledge reuse, team collaboration*
*Token cost: 1-3k tokens (vs 40-60k without compression)*
*Compression: 37:1 average ratio*
