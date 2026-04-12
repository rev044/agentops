---
description: Guide deep research phase with multi-agent support
---

# /research - Deep Research Phase

**Purpose:** Systematically explore and understand before planning or implementing

**When to use:**
- Before major implementations (understand the system first)
- Investigating complex problems (root cause analysis)
- Evaluating technologies or approaches (compare alternatives)
- Exploring unfamiliar codebases (map the landscape)

**Token budget:** 40-60k tokens (20-30% of context window)

**Output:** Research bundle (500-1k tokens compressed)

---

## The Research Phase Philosophy

**"Planning IS the work" - but research comes first.**

Most implementation failures trace back to incomplete understanding:
- Started coding without understanding the system
- Missed critical dependencies or constraints
- Made wrong assumptions about how things work
- Didn't find existing patterns to reuse

**Research prevents this by:**
1. Mapping the system landscape
2. Finding similar implementations
3. Identifying constraints and edge cases
4. Locating proven patterns to adapt

---

## Step 1: Constitutional Foundation (Always Loaded)

The Five Laws of an Agent are already loaded in your context via CONSTITUTION.md.

**Key reminder for research:**
- **Law #1:** Extract learnings (document patterns discovered)
- **Law #2:** Improve system (identify better approaches)
- **Law #3:** Document context (why research was needed, what you found)

---

## Step 2: Define Research Scope

**What are you researching?**

Choose research type:

### Problem Investigation
- What's broken and why?
- Where's the root cause?
- What are the constraints?

### Technology Evaluation
- Which approach is best?
- What are the tradeoffs?
- What do similar projects use?

### System Understanding
- How does this component work?
- What are the dependencies?
- Where are the integration points?

### Pattern Discovery
- What patterns exist already?
- How have others solved this?
- What can be reused or adapted?

---

## Step 3: Research Activities (Pick What's Needed)

### Activity 1: Map the Landscape

**Goal:** Understand what exists and how it's organized

**Commands to run:**
```bash
# Find relevant files
find . -name "*keyword*" -type f

# Search for patterns
grep -r "pattern-name" --include="*.yaml" --include="*.md"

# Understand structure
tree -L 3 -d [relevant-directory]

# Check git history
git log --oneline --all --grep="keyword" | head -20
```

**Document:**
- Key files and their purposes
- Directory organization
- Naming conventions
- Related components

### Activity 2: Find Similar Implementations

**Goal:** Don't reinvent - find what's already proven

**Commands to run:**
```bash
# Search for similar features
grep -r "similar-feature" --include="*.yaml" | head -20

# Check examples directory
ls examples/ | grep -i "keyword"

# Look at git history for similar work
git log --all --oneline --grep="similar-work" | head -10
```

**Document:**
- Existing implementations that are similar
- Patterns used in those implementations
- What needs to be adapted
- What can be reused as-is

### Activity 3: Identify Constraints and Edge Cases

**Goal:** Know the boundaries and gotchas

**Questions to answer:**
- What dependencies exist?
- What validation is required?
- What are the failure modes?
- What are the edge cases?
- What are the security considerations?

**Document:**
- Technical constraints (APIs, versions, compatibility)
- Business constraints (requirements, policies)
- Edge cases to handle
- Known gotchas or pitfalls

### Activity 4: Read Relevant Documentation

**Goal:** Leverage institutional knowledge

**Check these locations:**
```bash
# README files
find . -name "README.md" | grep -i "keyword"

# Documentation directory
ls docs/ | grep -i "topic"

# How-to guides
ls docs/how-to/ 2>/dev/null

# Case studies or examples
ls docs/case-studies/ 2>/dev/null
```

**Document:**
- Key documentation files read
- Important patterns or principles
- References for future use

---

## Step 4: Organize Research Findings

**Create structured research output:**

```markdown
# Research: [Topic Name]

**Date:** [YYYY-MM-DD]
**Researcher:** [Human/Agent]
**Time spent:** [X hours/sessions]

## Problem Statement
[Why was this research needed?]

## Scope
[What was investigated? What was out of scope?]

## Key Findings

### System Landscape
- **Location:** [path/to/relevant/files]
- **Structure:** [how it's organized]
- **Patterns:** [conventions observed]

### Similar Implementations
- **Example 1:** [path/to/example] - [what it does]
  - Pattern used: [pattern-name]
  - Adaptation needed: [what to change]
- **Example 2:** [another example]

### Constraints Identified
- **Technical:** [API limitations, version requirements]
- **Business:** [policies, requirements]
- **Edge cases:** [scenarios to handle]

### Dependencies
- **Required:** [what must exist first]
- **Optional:** [what would be nice to have]
- **Conflicts:** [what might break]

## Recommended Approach
[Based on research, what's the best path forward?]

## Open Questions
[What still needs to be clarified?]

## References
- [path/to/file:line] - [description]
- [documentation URL] - [description]

## Next Steps
1. [What to do with this research]
2. [Planning phase activities]
3. [Validation needed]
```

---

## Step 5: Compress to Research Bundle

**After research is complete, compress findings:**

```bash
/bundle-save [topic-name]-research --type research
```

**This creates a reusable bundle:**
- Compresses 40-60k research → 500-1k token bundle
- Enables resumption in next phase
- Shareable with team (prevents duplicate research)
- Reusable for similar problems

**Bundle structure:**
```markdown
# Research Bundle: [Topic]

## TL;DR (3-5 sentences)
[Quick summary of findings and recommendation]

## Problem Location
- File: [path/to/file:line]
- Root cause: [why it's happening]

## Similar Patterns Found
- Pattern: [name]
- Location: [where to find it]
- Adaptation: [what to change]

## Files to Modify
1. [path/to/file1] - [what needs to change]
2. [path/to/file2] - [what needs to change]

## Constraints
- [Key constraint 1]
- [Key constraint 2]

## Recommended Approach
[1-2 paragraphs: what to do and why]

## References
- [key-file:line]
- [documentation-page]
```

---

## Step 6: Transition to Planning

**Research is complete when you can answer:**

✅ Where is the problem/feature located?
✅ What similar implementations exist?
✅ What constraints must be respected?
✅ What approach is recommended?
✅ What files need to change?

**Next phase:**
```bash
/plan [topic-name]-research.md
```

**This loads your research bundle and creates detailed implementation plan.**

---

## Research Patterns

### Pattern 1: Wide-Then-Deep

1. **Wide scan:** Find all relevant files/components
2. **Deep dive:** Focus on most relevant 2-3 areas
3. **Document:** Map relationships and patterns

**When:** Exploring unfamiliar codebase

### Pattern 2: Follow-The-Trail

1. **Start:** Entry point (user request, error log)
2. **Trace:** Follow code execution or data flow
3. **Document:** The path and decision points

**When:** Debugging or root cause analysis

### Pattern 3: Compare-And-Contrast

1. **Find:** 3-5 similar implementations
2. **Compare:** What's common? What differs?
3. **Extract:** The reusable pattern

**When:** Designing new feature or evaluating approaches

### Pattern 4: Git-History-Mining

1. **Search:** Git log for related work
2. **Read:** Commit messages and diffs
3. **Extract:** Why decisions were made (institutional memory)

**When:** Understanding context or avoiding past mistakes

---

## Multi-Agent Research (Advanced)

**For complex research, use parallel agents:**

```bash
/research-multi "[research-topic]"
```

**This launches 3 agents simultaneously:**
- Agent 1: Map the landscape (structure, files, patterns)
- Agent 2: Find similar implementations (examples, references)
- Agent 3: Identify constraints (dependencies, edge cases, gotchas)

**Result:** 3x wall-clock speedup (same token budget, parallel execution)

**See:** `/research-multi` command for details

---

## Token Budget Management

**Research phase target:** 20-30% of context window (40-60k tokens)

**If approaching 40%:**

```bash
# Option 1: Checkpoint progress
/bundle-save [topic]-research-partial --type research

# Option 2: Continue in fresh session
# Load partial bundle in next session
/bundle-load [topic]-research-partial

# Option 3: Switch to planning phase
# Convert current findings to plan
/plan [topic]-research-notes.md
```

**Why 40%?** Agents degrade beyond this threshold (hallucinations, context collapse)

---

## Success Criteria

**Research is successful when:**

✅ You understand the problem/system deeply
✅ You've found similar implementations to reference
✅ You've identified constraints and edge cases
✅ You have a recommended approach
✅ You've documented findings for future reuse
✅ You're ready to create a detailed plan

**Research is NOT:**
- Making changes (that's implementation phase)
- Creating detailed specifications (that's planning phase)
- Writing new code (that's implementation phase)

**Research IS:**
- Understanding deeply
- Finding patterns
- Identifying constraints
- Recommending approaches

---

## Common Mistakes to Avoid

❌ **Starting to code during research** - Stay focused on understanding
❌ **Researching forever** - Set time limit, move to planning
❌ **Not documenting findings** - Future agents need your discoveries
❌ **Ignoring git history** - Institutional memory is gold
❌ **Not compressing to bundle** - 40-60k context won't fit in next phase

✅ **Do:** Understand, document, compress, move to planning
✅ **Do:** Extract reusable patterns (Law #1)
✅ **Do:** Save as bundle for future reuse
✅ **Do:** Respect 40% context threshold

---

## Integration with Other Commands

**Before research:**
```bash
/prime-complex  # Load constitutional foundation
# Agent asks: What's your complex task?
# You describe the problem
# Agent guides you to /research
```

**After research:**
```bash
/bundle-save [topic]-research  # Compress findings
/plan [topic]-research.md      # Create implementation plan
```

**Alternative: Skip planning for simple tasks:**
```bash
/research "simple-topic"       # Quick investigation
/implement [topic]-research.md # Direct to implementation
# Only for straightforward changes
```

---

## Examples

### Example 1: Investigating Performance Issue

```bash
# Start research
/research "Redis caching performance degradation"

# Agent activities:
# 1. Map: Find Redis configuration files
# 2. Similar: Search for other caching implementations
# 3. Constraints: Check Redis version, dependencies
# 4. History: Look for related performance work

# Result: Research bundle (1.2k tokens)
# - Problem: Connection pool exhaustion
# - Pattern: Similar issue solved in auth-service
# - Approach: Increase pool size, add health checks
# - Files: config/redis.yaml, monitoring/alerts.yaml
```

### Example 2: Evaluating New Technology

```bash
# Start research
/research "Should we use Kyverno or OPA for policy enforcement?"

# Agent activities:
# 1. Compare: Kyverno vs OPA features
# 2. Examples: How others use each tool
# 3. Constraints: Kubernetes version, team expertise
# 4. Patterns: What patterns exist for policy management

# Result: Research bundle (0.9k tokens)
# - Recommendation: Kyverno (native K8s, simpler)
# - Evidence: 3 similar projects use Kyverno successfully
# - Approach: Start with validation policies, expand to mutation
# - Next: Create implementation plan
```

### Example 3: Understanding Unfamiliar Codebase

```bash
# Start research
/research "How does the authentication flow work?"

# Agent activities:
# 1. Entry point: Find login endpoint
# 2. Trace: Follow request through middleware
# 3. Dependencies: JWT validation, session management
# 4. Patterns: Standard OAuth2 flow with custom claims

# Result: Research bundle (1.1k tokens)
# - Flow: Login → JWT → Validation → Session
# - Files: auth/handlers.go, middleware/jwt.go, config/auth.yaml
# - Patterns: Standard OAuth2, can reuse for new endpoints
# - Next: Plan new auth feature using established patterns
```

---

## When to Use vs When to Skip

**Use /research when:**
- Problem is complex or unfamiliar
- High risk of wrong approach
- Need to understand system deeply
- Want to find reusable patterns
- Multiple possible approaches exist

**Skip research (use /prime-simple) when:**
- Problem is straightforward
- You already understand the system
- Change is trivial (typo, version bump)
- Pattern is obvious
- Time is critical and risk is low

---

## Related Commands

- **/prime-complex** - Load constitutional foundation before research
- **/research-multi** - Parallel multi-agent research (3x speedup)
- **/bundle-save** - Compress research findings for reuse
- **/bundle-load** - Resume research in fresh session
- **/plan** - Next phase after research
- **/learn** - Extract patterns from research for future reuse

---

**Ready to research? What are you investigating?**
