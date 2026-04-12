---
description: Extract reusable patterns from completed work for institutional memory
---

# /learn - Pattern Extraction & Learning

**Purpose:** Extract reusable patterns from work for future agents and developers

**When to use:**
- After completing implementation (work is done)
- After solving complex problem (capture solution)
- After discovering efficient pattern (share with team)
- Periodic retrospectives (extract learnings from recent work)

**Token budget:** 10-20k tokens (5-10% of context window)

**Output:** Learning bundle (300-500 tokens) + pattern catalog update

---

## The Learning Philosophy

**"Institutional memory compounds over time."**

Every problem solved is an opportunity to improve the system:
- Extract the pattern that worked
- Document why it worked
- Make it reusable for next time
- Prevent solving the same problem twice

**Learning fulfills Law #1: ALWAYS Extract Learnings**

Without learning:
- Same problems solved repeatedly
- Knowledge lives only in commit messages
- New agents start from zero
- Efficiency stays flat

With learning:
- Patterns become reusable
- Knowledge accessible and searchable
- New agents benefit from past work
- Efficiency compounds

---

## What Gets Learned?

### Pattern Types

**1. Implementation Patterns**
- How to solve a specific problem
- What approach works best
- What to avoid (gotchas, pitfalls)

**Example:**
```markdown
# Pattern: Redis Connection Pooling

## Problem
Redis connections exhaust under burst traffic

## Solution
Increase pool size, add health checks, implement circuit breaker

## When to Use
- High-traffic services using Redis
- Burst traffic patterns (5pm spike)
- Connection limits approaching maximum

## Implementation
1. config/redis.yaml:15 - pool_size: 100
2. app/cache.go:34 - Initialize with pool
3. app/cache.go:89 - Add health check endpoint

## Gotchas
- Pool size too high exhausts memory
- Health check must verify Redis connectivity, not just pool existence

## Evidence
- Reduced connection errors by 99%
- P95 latency improved from 500ms → 50ms
- Handled 3x traffic during peak
```

**2. Debugging Patterns**
- How to diagnose a class of problems
- What tools or commands revealed the issue
- What symptoms to look for

**Example:**
```markdown
# Pattern: Diagnosing Intermittent Auth Failures

## Symptoms
- Auth succeeds 95% of time, fails 5%
- Failures clustered around 5pm
- No error logs during failures

## Diagnosis Steps
1. Check metrics: High traffic at 5pm (burst pattern)
2. Check pools: Connection pool exhausted
3. Check logs: Timeouts during exhaustion

## Root Cause
Connection pool too small for burst traffic

## Solution Pattern
See "Redis Connection Pooling" pattern

## Reusable Diagnostic
- Check traffic patterns (bursts vs steady)
- Check resource limits (pools, memory, CPU)
- Correlate failures with resource exhaustion
```

**3. Architecture Patterns**
- How to structure code/config for a domain
- What organization works well
- What scales as project grows

**Example:**
```markdown
# Pattern: Base + Overlay Kustomize Structure

## Problem
50+ applications with mostly shared config, some overrides

## Solution
Base directory (common config) + overlays (site-specific)

## Structure
```
base/
  deployment.yaml  # Common to all sites
  service.yaml     # Common to all sites
overlays/
  site-a/
    kustomization.yaml  # References base, adds overrides
  site-b/
    kustomization.yaml  # Different overrides
```

## When to Use
- Multiple environments (dev, staging, prod)
- Multiple sites with shared infrastructure
- Common base with customization points

## Benefits
- DRY: Config defined once in base
- Override only what differs per site
- Easy to add new sites (copy overlay template)

## Evidence
- Reduced config duplication from 80% → 5%
- New site setup: 2 hours → 15 minutes
- Config errors dropped 60% (one source of truth)
```

**4. Automation Patterns**
- What to automate and how
- What tools work well together
- What workflows save time

**Example:**
```markdown
# Pattern: Bundle Checkpointing for Long Implementations

## Problem
Large implementations exceed single-session context (>40%)

## Solution
Checkpoint progress mid-implementation, resume in fresh session

## Workflow
1. Implementation runs, context approaches 40%
2. Auto-save progress bundle (what's done, what remains)
3. Fresh session loads bundle (2-3k tokens)
4. Continue from checkpoint with full context headroom

## When to Use
- Implementations >40k tokens
- Multi-day projects
- Complex changes with many files

## Benefits
- No context collapse (fresh session = better decisions)
- Can pause/resume anytime
- Multiple sessions, sustained quality

## Evidence
- Completed implementations 50% larger than before
- Zero context-related errors
- Sustained quality across multi-day projects
```

---

## Step 1: Identify What to Extract

**After completing work, I analyze:**

### Questions I Ask

**What problem was solved?**
- What was broken or missing?
- Why did it need solving?
- What was the impact?

**What approach worked?**
- What solution was implemented?
- Why did this approach work?
- What alternatives were considered?

**What was learned?**
- What would you do differently next time?
- What gotchas were discovered?
- What patterns emerged?

**What is reusable?**
- Could this pattern apply to other problems?
- What generalizes? What's specific?
- How should it be documented?

---

## Step 2: Extract the Pattern

**I create structured pattern documentation:**

```markdown
# Pattern: [Name]

**Category:** [Implementation | Debugging | Architecture | Automation]
**Difficulty:** [Simple | Moderate | Complex]
**Domain:** [General | Go | Python | Kubernetes | etc.]

## Problem Statement
[What problem does this solve? When does it occur?]

## Solution Summary
[High-level approach, 2-3 sentences]

## When to Use
- [Scenario 1]
- [Scenario 2]
- [Scenario 3]

## When NOT to Use
- [Anti-pattern scenario 1]
- [Anti-pattern scenario 2]

## Implementation Steps
1. [Step 1 with file:line references]
2. [Step 2 with commands]
3. [Step 3 with validation]

## Code Example
```language
// Before
[problematic code]

// After
[improved code]
```

## Validation
[How to verify the pattern worked]
- Command: [test command]
- Expected: [outcome]

## Gotchas
- **[Gotcha 1]:** [What to watch for]
- **[Gotcha 2]:** [Common mistake]

## Evidence
- [Metric 1]: [improvement]
- [Metric 2]: [improvement]
- [Experience]: [number of times used successfully]

## Related Patterns
- [Pattern A] - [relationship]
- [Pattern B] - [relationship]

## References
- [file:line] - [where implemented]
- [doc URL] - [related documentation]
- [commit SHA] - [when introduced]

## Author
[Who discovered/documented this]

## Date
[When pattern was discovered]
```

---

## Step 3: Categorize and Store

**I save pattern to appropriate location:**

### Pattern Storage Structure

```
.agentops/patterns/
├── implementation/
│   ├── redis-connection-pooling.md
│   ├── jwt-validation.md
│   └── error-handling-wrapper.md
├── debugging/
│   ├── diagnose-intermittent-auth.md
│   ├── trace-slow-queries.md
│   └── memory-leak-detection.md
├── architecture/
│   ├── base-overlay-kustomize.md
│   ├── microservice-boundaries.md
│   └── config-management.md
├── automation/
│   ├── bundle-checkpointing.md
│   ├── multi-agent-research.md
│   └── continuous-validation.md
└── INDEX.md  # Searchable catalog
```

### Pattern Index Update

**I update the pattern catalog:**

```markdown
# Pattern Catalog Index

## By Category

### Implementation Patterns (15)
- [Redis Connection Pooling](implementation/redis-connection-pooling.md) - Handle burst traffic
- [JWT Validation](implementation/jwt-validation.md) - Secure API authentication
- [Error Handling Wrapper](implementation/error-handling-wrapper.md) - Consistent error responses

### Debugging Patterns (8)
- [Diagnose Intermittent Auth](debugging/diagnose-intermittent-auth.md) - Connection pool exhaustion
- [Trace Slow Queries](debugging/trace-slow-queries.md) - Database performance issues
- [Memory Leak Detection](debugging/memory-leak-detection.md) - Go memory profiling

### Architecture Patterns (12)
- [Base + Overlay Kustomize](architecture/base-overlay-kustomize.md) - Multi-site config
- [Microservice Boundaries](architecture/microservice-boundaries.md) - Domain separation
- [Config Management](architecture/config-management.md) - Environment-specific settings

### Automation Patterns (6)
- [Bundle Checkpointing](automation/bundle-checkpointing.md) - Long implementations
- [Multi-Agent Research](automation/multi-agent-research.md) - 3x speedup
- [Continuous Validation](automation/continuous-validation.md) - Pre-commit checks

## By Problem Domain

### Performance (8 patterns)
### Security (5 patterns)
### Scalability (7 patterns)
### Maintainability (10 patterns)

## Most Used (Last 30 Days)
1. Redis Connection Pooling (12 uses)
2. Base + Overlay Kustomize (8 uses)
3. Bundle Checkpointing (5 uses)

## Recently Added
- 2025-11-07: Bundle Checkpointing
- 2025-11-06: JWT Validation
- 2025-11-05: Diagnose Intermittent Auth
```

---

## Step 4: Make Pattern Discoverable

**I ensure future agents/developers can find patterns:**

### Tagging

**Add searchable tags:**

```markdown
# Pattern: Redis Connection Pooling

**Tags:** #redis #performance #connection-pool #burst-traffic #caching
```

### Cross-Referencing

**Link related patterns:**

```markdown
## Related Patterns
- [Circuit Breaker](architecture/circuit-breaker.md) - Combine with connection pooling
- [Health Checks](debugging/health-check-design.md) - Validate pool status
- [Rate Limiting](implementation/rate-limiting.md) - Alternative approach to burst traffic
```

### Search Integration

**Add to git commit messages:**

```bash
git commit -m "feat(cache): Add Redis connection pooling

## Pattern Extracted
Documented in: .agentops/patterns/implementation/redis-connection-pooling.md
Tags: #redis #performance #connection-pool

See pattern for reusable approach to handling burst traffic.
"
```

---

## Step 5: Share with Team

**I make learning accessible:**

### Option 1: Bundle for Sharing

```bash
# Create learning bundle
/bundle-save [pattern-name]-learning --type learning

# Share bundle URL or UUID with team
# Team loads: /bundle-load [pattern-name]-learning
```

### Option 2: Documentation Update

```bash
# Add to team documentation
cp .agentops/patterns/[category]/[pattern].md docs/patterns/

# Commit to shared repo
git add docs/patterns/
git commit -m "docs(patterns): Add [pattern-name] pattern"
git push
```

### Option 3: Pattern Demo

**Create runnable example:**

```bash
# Add to examples directory
examples/[pattern-name]/
├── README.md         # Pattern explanation
├── before/          # Problem state
│   └── code.go      # Problematic code
├── after/           # Solution state
│   └── code.go      # Improved code
└── tests/           # Verification
    └── pattern_test.go
```

---

## Learning Patterns

### Pattern 1: Retrospective Learning (Periodic)

**Scan recent work for patterns:**

```bash
# Weekly or monthly
/learn --retrospective --since "2025-11-01"

# I analyze:
# - Last 30 days of commits
# - Common problems solved
# - Repeated approaches
# - Patterns that emerged
#
# Extract:
# - 3-5 reusable patterns
# - Updated pattern catalog
```

### Pattern 2: Immediate Learning (After Each Task)

**Extract pattern right after solving:**

```bash
# After implementation
/implement [topic]-plan
# Implementation complete

# Immediately extract
/learn [topic]

# I analyze:
# - What was implemented
# - What approach worked
# - What was learned
#
# Extract:
# - Single pattern
# - Add to catalog
```

### Pattern 3: Failure Learning (After Mistakes)

**Learn from what didn't work:**

```bash
# After failed approach
/learn --failure "[what-went-wrong]"

# I analyze:
# - What was attempted
# - Why it failed
# - What succeeded instead
#
# Extract:
# - Anti-pattern (what not to do)
# - Correct pattern (what to do)
# - Warning in pattern catalog
```

### Pattern 4: Comparative Learning (Evaluate Alternatives)

**Compare multiple approaches:**

```bash
# After evaluating alternatives
/learn --compare "[approach-A]" "[approach-B]"

# I analyze:
# - Approach A: pros, cons, use cases
# - Approach B: pros, cons, use cases
# - When to use each
#
# Extract:
# - Decision framework
# - When to use which approach
```

---

## Integration with Laws of an Agent

**Learning directly fulfills AgentOps Laws:**

### Law #1: ALWAYS Extract Learnings
**`/learn` is the mechanism:**
- Documents patterns discovered
- Makes learnings reusable
- Compounds institutional memory

### Law #2: ALWAYS Improve Self or System
**Learning improves system:**
- Pattern catalog grows
- Future work becomes faster
- Quality improves over time

### Law #3: ALWAYS Document Context
**Learning captures context:**
- Why solution was needed
- What was implemented
- How to reuse

---

## Success Criteria

**Learning succeeds when:**

✅ Pattern is clearly documented
✅ Pattern is discoverable (tags, index, search)
✅ Pattern is reusable (generalizes beyond specific instance)
✅ Evidence supports pattern (metrics, uses)
✅ Related patterns cross-referenced
✅ Stored in appropriate location

**Learning does NOT:**
- Document implementation details (that's in code)
- Duplicate existing documentation (patterns are new insights)
- Create one-time solutions (patterns must generalize)

---

## Common Learning Mistakes

❌ **Too specific** - Pattern only applies to exact scenario
❌ **Too vague** - Pattern doesn't provide actionable steps
❌ **No evidence** - Claims benefits without proof
❌ **Not discoverable** - Stored but not indexed or tagged
❌ **Duplicates existing** - Pattern already documented

✅ **Do:** Generalize from specific to reusable
✅ **Do:** Provide clear steps and examples
✅ **Do:** Include evidence (metrics, uses)
✅ **Do:** Make searchable (tags, index)
✅ **Do:** Check existing patterns first

---

## Token Budget Management

**Learning phase target:** 5-10% of context window (10-20k tokens)

**Breakdown:**
- Analyze recent work: 5-10k tokens
- Extract patterns: 3-5k tokens
- Document patterns: 2-5k tokens
- Update catalog: 1-2k tokens

**If approaching 40%:**

```bash
# Learning is lightweight, shouldn't approach 40%
# If it does, you're analyzing too much context
# Focus on recent work only (last commit or session)
```

---

## Examples

### Example 1: Immediate Pattern Extraction

```bash
# After implementing Redis caching
/learn redis-caching-implementation

# I analyze:
# - Commit: feat(cache): Add Redis connection pooling
# - Files: config/redis.yaml, app/cache.go, app/health.go
# - Tests: All passed
# - Evidence: P95 latency 500ms → 50ms
#
# Extract:
# Pattern: Redis Connection Pooling
# Category: Implementation
# Evidence: 10x latency improvement
#
# Saved: .agentops/patterns/implementation/redis-connection-pooling.md
# Updated: .agentops/patterns/INDEX.md
```

### Example 2: Retrospective Learning

```bash
# Monthly retrospective
/learn --retrospective --since "2025-11-01"

# I analyze last 30 days:
# - 15 commits reviewed
# - 3 patterns emerged:
#   1. Base + overlay Kustomize (used 8 times)
#   2. Bundle checkpointing (used 5 times)
#   3. Multi-agent research (used 3 times)
#
# Extract all 3 patterns
# Update pattern catalog
# Mark most-used patterns
```

### Example 3: Failure Learning

```bash
# After failed approach
/learn --failure "JWT validation without library"

# I analyze:
# - Attempted: Hand-rolled JWT validation
# - Failed: Security vulnerabilities, edge cases missed
# - Succeeded: Using golang-jwt/jwt library
#
# Extract:
# Anti-pattern: Don't hand-roll crypto
# Correct pattern: Use vetted libraries
# Warning: Added to pattern catalog
```

---

## Related Commands

- **/implement** - Execute plan (learning comes after)
- **/validate** - Verify implementation (learning extracts patterns)
- **/bundle-save** - Save learning bundles for sharing
- **/research** - Research patterns (learning documents new ones)

---

**Ready to extract patterns? Use /learn [topic] after completing work.**
