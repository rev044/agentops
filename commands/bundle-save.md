# /bundle-save - Compress and Save Session Knowledge

**Purpose:** Compress session work for future resumption or team sharing

**Philosophy:** 37:1 compression enables multi-session work and knowledge reuse

**Token budget:** 5-10k tokens (to create bundle)

**Output:** Compressed bundle (1-3k tokens) saved to .agentops/bundles/

---

## When to Use

Use `/bundle-save` in these scenarios:

### 1. End of Research Phase
```
Just completed /research → Save findings → End session
Next session: Load bundle → /plan
```

### 2. End of Planning Phase
```
Just completed /plan → Save specifications → End session
Next session: Load bundle → /implement
```

### 3. Mid-Implementation (Context Filling)
```
Implementation at 40% context → Still have 60% work remaining
Save progress → End session → Resume in fresh context
```

### 4. Team Knowledge Sharing
```
Discovered valuable pattern → Compress → Save → Team can load
```

---

## How It Works

### Basic Usage
```bash
/bundle-save {name}

# Examples:
/bundle-save redis-caching-research
/bundle-save api-optimization-plan
/bundle-save containerization-progress
```

### Step 1: Analyze Session Content
```
Agent scans current session:
- What phase? (research/plan/implement)
- What's been accomplished?
- What remains (if incomplete)?
- What's the key finding/decision?
- What patterns emerged?
```

### Step 2: Compress Content
```
Original session: 45k tokens (research findings)
  ↓
Extract key information:
- Executive summary (200 tokens)
- Approaches evaluated (800 tokens)
- Recommendation + rationale (400 tokens)
- Implementation notes (300 tokens)
- Constraints discovered (200 tokens)
  ↓
Compressed bundle: 1.9k tokens

Compression ratio: 23:1
```

### Step 3: Save Bundle
```
Location: .agentops/bundles/{name}.md
Metadata: Timestamp, phase, tags, UUID
Size: 1.9k tokens
Reusable: YES
```

---

## Bundle Types

### Research Bundle
**Created after:** `/research` phase
**Contains:**
- Problem statement
- Approaches evaluated (pros/cons)
- Recommended approach + rationale
- Key constraints
- Implementation considerations

**Example:**
```markdown
# Redis Caching Research Bundle

## Executive Summary
Redis with pub/sub provides 10x response time improvement
with acceptable memory overhead (500MB).

## Evaluated Approaches
1. Redis (recommended) - Distributed, pub/sub, 10k msg/sec
2. Memcached - Simpler, no persistence, no pub/sub
3. Local cache - Fast but not distributed

## Recommendation
Use Redis cluster with:
- 500MB memory limit
- TTL: 3600s
- Circuit breaker for failures
- Pub/sub for cache invalidation

## Constraints
- Single instance initially (add failover in phase 2)
- Requires circuit breaker in API
- Must validate wal_level=logical

## Next Phase
Load this bundle → /plan detailed implementation
```

---

### Plan Bundle
**Created after:** `/plan` phase
**Contains:**
- Original plan summary
- Files to create/modify/delete (compressed list)
- Validation strategy
- Rollback procedure
- Approval status

**Example:**
```markdown
# Redis Caching Plan Bundle

## Approved Plan Summary
Deploy Redis instance, integrate with API via service discovery,
implement circuit breaker, validate 10x speedup.

## Key Changes
Create: 5 files (redis manifests, test scripts)
Modify: 3 files (API config, namespaces)
Delete: 1 file (legacy cache)

## Validation
- Syntax: make quick
- Integration: kubectl apply --dry-run
- Performance: cache response < 100ms
- Rollback: git revert tested

## Approval
✅ Approved 2025-11-07 by user
Safe to implement

## Next Phase
Load this bundle → /implement approved plan
```

---

### Implementation Progress Bundle (NEW)
**Created during:** `/implement` phase when context fills
**Contains:**
- Original plan (compressed)
- Completed changes (what's done)
- Remaining changes (what's left)
- Git state (staged files)
- Validation results (partial)
- Token estimate for continuation

**Example:**
```markdown
# Redis Caching Implementation Progress Bundle

## Status
Progress: 7 of 20 changes complete (35%)
Context: 90k tokens used (45%)
Estimated remaining: 60k tokens

## Original Plan
[Compressed plan summary - 500 tokens]

## Completed ✅
Created: 5 files (validated)
Modified: 2 files (validated)
Git: 7 files staged

## Remaining ⏸️
Create: 5 more files
Modify: 6 more files
Validate: Integration + performance
Commit: Final commit + push

## Git State
Branch: main
Staged: 7 files
Uncommitted: 0 files

## Resume In Next Session
1. /prime-complex
2. /bundle-load redis-caching-progress
3. /implement --resume
4. Continues from "Remaining" section

## Validation (Partial)
✅ Syntax: All created files pass
✅ YAML: All manifests valid
⏸️ Integration: Not yet tested
⏸️ Performance: Not yet measured
```

---

## Implementation Bundle vs Complete Bundle

### During Implementation (Progress Bundle)
```
Context at 40% → Still have work remaining

/bundle-save redis-caching-progress --type implementation

Creates: Implementation progress bundle
Contains: What's done + what remains + how to resume
Purpose: Multi-session implementation
```

### After Implementation (Complete Bundle)
```
Implementation finished → All validated → Committed

/bundle-save redis-caching-complete --type implementation

Creates: Implementation summary bundle
Contains: What was built + results + learnings
Purpose: Team knowledge + audit trail
```

---

## Compression Strategies

### Research Compression (40-60k → 1-2k)
**Keep:**
- Recommended approach
- Key constraints
- Critical decisions

**Discard:**
- Detailed exploration notes
- Rejected approaches (except why rejected)
- Verbose explanations

**Ratio:** 30-40:1

### Plan Compression (40-60k → 1.5-2.5k)
**Keep:**
- File:line specifications (compressed list)
- Validation commands
- Rollback procedure

**Discard:**
- Detailed rationale (in research bundle)
- Verbose file contents
- Repetitive validation steps

**Ratio:** 25-35:1

### Implementation Progress (90k → 2-3k)
**Keep:**
- Completed file list
- Remaining file list
- Git state
- Validation results

**Discard:**
- File contents (already in git)
- Verbose logs
- Intermediate steps

**Ratio:** 30-40:1

---

## Auto-Save Triggers

Agent should suggest bundle-save automatically:

### Research Phase
```
/research complete → findings ready
Agent: "Research complete. Save bundle?"
User: "Yes" → /bundle-save {name}-research
```

### Plan Phase
```
/plan approved → specifications ready
Agent: "Plan approved. Save bundle?"
User: "Yes" → /bundle-save {name}-plan
```

### Implementation (Context Threshold)
```
/implement in progress → context at 40%
Agent: "Context at 40%, estimated 60% remaining work.
       Save progress bundle and resume in fresh session?"
User: "Yes" → /bundle-save {name}-progress --type implementation
```

---

## Bundle Metadata

Every bundle includes:

```yaml
---
bundle_id: bundle-redis-caching-research-2025-11-07
created: 2025-11-07T10:30:00Z
type: research  # or: plan, implementation-progress, implementation-complete
phase: research
original_tokens: 45000
compressed_tokens: 1900
compression_ratio: 23.7
tags: [redis, caching, performance]
accessed_count: 0
last_accessed: null
approved: true  # for plans
git_sha: abc123def456  # commit when bundle created
---
```

---

## Bundle Naming Convention

```
Pattern: {topic}-{type}[-{version}]

Examples:
redis-caching-research          # Research findings
redis-caching-plan              # First plan
redis-caching-plan-v2           # Revised plan
redis-caching-progress          # Mid-implementation
redis-caching-complete          # Implementation done
k8s-migration-research          # Different topic
```

---

## Integration with Git

### Bundles Are Git-Tracked
```
.agentops/bundles/
├── redis-caching-research.md      # Committed
├── redis-caching-plan.md          # Committed
├── redis-caching-progress.md      # Committed (WIP)
└── redis-caching-complete.md      # Committed

Benefits:
- Version controlled
- Team accessible
- Audit trail
- Searchable history
```

### Sessions Are NOT Git-Tracked
```
.agentops/sessions/
├── 2025-11-07-10-30-research/     # Local only
└── 2025-11-07-14-15-plan/         # Local only

Reason:
- Too verbose (40-90k tokens)
- Temporary working space
- Pruned after bundle extraction
```

---

## Multi-Session Implementation Example

### Session 1: Start Implementation
```
/prime-complex
/bundle-load redis-caching-plan

/implement redis-caching-plan.md
[Creates 7 of 20 files]
[Context at 90k tokens, 45%]

Agent: "Context approaching threshold.
       7 changes done, 13 remaining.
       Save progress bundle?"

User: "Yes"

/bundle-save redis-caching-progress --type implementation

Saved: redis-caching-progress.md (2.5k tokens)
Git: 7 files staged (not committed)

[End Session 1]
```

### Session 2: Resume Implementation
```
/prime-complex
/bundle-load redis-caching-progress

Bundle loaded:
- 7 changes complete (validated)
- 13 changes remaining
- Git: 7 files staged

/implement --resume

[Continues from checkpoint]
[Creates remaining 13 files]
[Runs full validation]
[Creates commit]

✅ Implementation complete

/bundle-save redis-caching-complete --type implementation

Saved: redis-caching-complete.md (1.8k tokens)
Git: All changes committed + pushed

[End Session 2]
```

---

## Success Criteria

Bundle-save is successful when:
- ✅ Content compressed (target 37:1 ratio)
- ✅ Bundle saved to .agentops/bundles/
- ✅ Metadata complete
- ✅ Resumable (if progress bundle)
- ✅ Team accessible (git-tracked)
- ✅ Tagged appropriately

---

## Commands Reference

```bash
# Save research findings
/bundle-save {topic}-research

# Save approved plan
/bundle-save {topic}-plan

# Save implementation progress (mid-stream)
/bundle-save {topic}-progress --type implementation

# Save completed implementation
/bundle-save {topic}-complete --type implementation

# Save with custom tags
/bundle-save {topic}-research --tags "redis,performance,caching"
```

---

*Compression target: 37:1 ratio (40-60k → 1-2k tokens)*
*Enables: Multi-session work, team sharing, knowledge reuse*
*Integrates: Research → Plan → Implement → Learn cycle*
