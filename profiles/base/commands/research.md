---
description: Research phase - gather information and understand the problem
---

# /research - Phase 1: Gather & Understand

**Purpose:** Systematic research to understand problems before planning solutions.

**Philosophy:** Understanding is 80% of the work. Research deeply to plan effectively.

**Token budget:** 40-60k tokens (20-30% of context window)

---

## When to Use

Use `/research` when you need to:
- Understand a new problem domain
- Evaluate multiple solution approaches
- Gather evidence before making decisions
- Map existing systems and patterns
- Identify constraints and requirements

**Don't use if:**
- You already have a clear plan (use `/implement`)
- Problem is trivial and well-understood
- Just need to execute existing approach

---

## Research Process

### Step 1: Define Research Questions

What specifically do you need to understand?

**Examples:**
- How does the current system work?
- What are the constraints?
- What solutions exist?
- What are the tradeoffs?

### Step 2: Gather Information

**Sources:**
- Codebase exploration
- Documentation
- Similar implementations
- Technical specifications
- User requirements

### Step 3: Analyze & Synthesize

**Output:**
- Key findings
- Constraints identified
- Solution approaches
- Recommendations
- Next steps

### Step 4: Save Research Bundle

Compress findings for next phase:

```bash
/bundle-save research-[topic-name]
```

---

## Research Deliverables

### Research Document Structure

```markdown
# Research: [Topic Name]

## Problem Statement
[What are we trying to solve?]

## Key Findings
1. [Finding 1]
2. [Finding 2]
3. [Finding 3]

## Constraints
- [Constraint 1]
- [Constraint 2]

## Solution Approaches

### Approach A: [Name]
**Pros:** ...
**Cons:** ...
**Effort:** ...

### Approach B: [Name]
**Pros:** ...
**Cons:** ...
**Effort:** ...

## Recommendation
[Which approach and why?]

## Next Steps
1. Create implementation plan
2. Get approval
3. Execute
```

---

## Token Budget Management

```
Research Phase: 40-60k tokens (20-30%)

Breakdown:
- Exploration: 20-30k
- Analysis: 10-20k
- Documentation: 5-10k
- Reserve: 5-10k

Monitor: Stay under 40% total
```

---

## Transition to Planning

After research complete:

```bash
# Save research
/bundle-save research-[topic]

# Start new session for planning
# Load research bundle
/bundle-load research-[topic]

# Create plan
/plan
```

---

## Success Criteria

Research is complete when:

- [ ] Problem fully understood
- [ ] Constraints identified
- [ ] Multiple approaches evaluated
- [ ] Recommendation made with rationale
- [ ] Research bundle saved
- [ ] Ready for planning phase

---

**Next command:** `/plan` to create implementation specification
