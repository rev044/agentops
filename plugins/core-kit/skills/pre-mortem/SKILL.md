---
name: pre-mortem
description: >
  Pre-mortem simulation for specs and designs. Simulates N iterations of
  implementation to identify failure modes before they happen. Triggers:
  "pre-mortem", "simulate spec", "stress test spec", "find spec gaps",
  "simulate implementation", "what could go wrong", "anticipate failures".
version: 1.0.0
tier: solo
author: "AI Platform Team"
license: "MIT"
context: inline
allowed-tools: "Read,Write,Bash,Grep,Glob,Task"
skills:
  - research
  - plan
---

# Pre-Mortem Skill

Pre-mortem simulation: find what will go wrong BEFORE implementation.

## Role in the Brownian Ratchet

Pre-mortem is the **pre-implementation filter** - it catches failures before they happen:

| Component | Pre-Mortem's Role |
|-----------|-------------------|
| **Chaos** | Simulate N iterations with different failure modes |
| **Filter** | Each iteration identifies problems before implementation |
| **Ratchet** | Enhanced spec locks lessons learned |

> **Pre-mortem filters BEFORE the chaos of implementation starts.**

Unlike /vibe which filters code after writing, pre-mortem filters specs
before implementation begins. This prevents entire classes of failures
from ever being attempted.

**The Economics:**
- Without pre-mortem: 10 implementation attempts × fix time = expensive
- With pre-mortem: 10 mental simulations + 1 correct implementation = cheap

## Quick Start

```bash
/pre-mortem .agents/specs/2026-01-22-feature-spec.md
```

## Philosophy

> "Simulate doing it 10 times and learn all the lessons so we don't have to."

Instead of:
1. Write spec
2. Implement
3. Hit problems
4. Fix spec
5. Repeat 10 times

Do:
1. Write spec v1
2. **Simulate 10 iterations mentally**
3. Extract ALL lessons upfront
4. Write spec v2 (battle-hardened)
5. Implement once, correctly

---

## Workflow

```
1. Intake         -> Read the spec
2. Simulate       -> Run N iterations (default: 10)
3. Analyze        -> Extract failure modes
4. Enhance        -> Apply lessons to spec
5. Output         -> Enhanced spec + analysis artifact
```

---

## Phase 1: Intake

Read and understand the spec structure:

```bash
# Read the target spec
Read($SPEC_PATH)

# Identify key components
- What does it define? (API, workflow, architecture)
- What are the critical paths?
- What are the integration points?
- What assumptions does it make?
```

**Checklist**:
- [ ] Read full spec
- [ ] Identified 3-5 critical components
- [ ] Listed external dependencies
- [ ] Noted implicit assumptions

---

## Phase 2: Simulate Iterations

For each iteration (1 to N), imagine implementing the spec and identify what goes wrong.

### Iteration Template

```markdown
## Iteration N: [Failure Category]

**What goes wrong:**
- Specific failure scenario
- Root cause
- How it manifests

**Lesson learned:**
- What assumption was wrong?
- What was missing from spec?

**Enhancement needed:**
- [ ] Concrete fix for the spec
```

### Common Failure Categories

| Category | Examples |
|----------|----------|
| **Interface Mismatch** | API doesn't match reality, JSON schema wrong |
| **Timeout/Performance** | Operations take longer than expected |
| **Error Handling** | Edge cases not covered, unclear recovery |
| **Safety/Security** | Missing guards, unclear permissions |
| **User Experience** | Confusing flows, missing feedback |
| **Integration** | External deps behave differently |
| **State Management** | Race conditions, partial failures |
| **Documentation** | Spec says X, reality is Y |
| **Tooling** | CLI behaves differently than documented |
| **Operational** | Monitoring, debugging, rollback gaps |

### Simulation Prompts

For each iteration, ask:

1. **What if the input isn't what we expect?**
2. **What if the external dependency fails?**
3. **What if this takes 10x longer?**
4. **What if the user skips reading instructions?**
5. **What if we need to rollback?**
6. **What if we're debugging this at 2 AM?**
7. **What happens on partial failure?**
8. **What if the user does this 100 times?**
9. **What if the environment is different?**
10. **What does the audit trail look like?**

---

## Phase 3: Analyze Patterns

After simulating all iterations, categorize findings:

### Severity Classification

| Severity | Criteria | Action |
|----------|----------|--------|
| **Critical** | Blocks v1, no workaround | Must fix before implementation |
| **Important** | Significant UX/reliability impact | Should fix for v1 |
| **Nice-to-Have** | Improvements, polish | Can defer to v1.1 |

### Finding Extraction

```markdown
## Summary: Enhancement Checklist

### Critical (Must have before v1)
1. [Finding 1] - Brief description
   - Fix: What to add/change
2. [Finding 2] - Brief description
   - Fix: What to add/change

### Important (Should have for v1)
3. [Finding 3] - Brief description
   - Fix: What to add/change

### Nice to Have (v1.1)
4. [Finding 4] - Brief description
   - Fix: What to add/change
```

---

## Phase 4: Enhance Spec

Apply lessons learned to create enhanced spec:

### Enhancement Types

| Type | Example |
|------|---------|
| **Schema Addition** | Add actual JSON schemas from code |
| **Error Matrix** | Map failure modes to recovery procedures |
| **Safety Gates** | Add explicit human gates for dangerous ops |
| **Timeout Config** | Per-operation timeout specifications |
| **Progress Feedback** | Expected durations, status patterns |
| **Audit Trail** | Logging requirements, session IDs |
| **Rollback Guide** | Recovery procedures for partial failures |

### Spec Enhancement Checklist

- [ ] Every interface has actual schema (not example)
- [ ] Every error has recovery guidance
- [ ] Every dangerous operation has explicit warning
- [ ] Every long operation has timeout + progress
- [ ] Every integration point has fallback
- [ ] Every assumption is documented

---

## Phase 5: Output

### Files to Create

1. **Analysis Artifact**: `~/gt/.agents/<rig>/specs/YYYY-MM-DD-{spec-name}-analysis.md`
   - All iterations with lessons
   - Categorized findings
   - Enhancement checklist

2. **Enhanced Spec**: `~/gt/.agents/<rig>/specs/YYYY-MM-DD-{spec-name}-v2.md`
   - Original spec with all enhancements applied
   - Change log from v1

### Output Template

```markdown
# Spec Enhancement Complete: [Spec Name]

**Original:** `path/to/spec-v1.md`
**Analysis:** `path/to/spec-analysis.md`
**Enhanced:** `path/to/spec-v2.md`

## Simulation Summary

| Iterations | Critical | Important | Nice-to-Have |
|------------|----------|-----------|--------------|
| 10         | 4        | 3         | 3            |

## Critical Enhancements Applied

1. **[Enhancement 1]**: Brief description
2. **[Enhancement 2]**: Brief description
...

## Next Steps

- [ ] Review enhanced spec for completeness
- [ ] Validate schemas against actual code
- [ ] Create implementation issues from spec
```

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Stop at 3 iterations | Run full 10 iterations minimum |
| Only consider happy path | Focus on failure modes |
| Make up JSON schemas | Extract from actual code |
| Skip severity classification | Categorize all findings |
| Leave findings as notes | Apply as concrete spec changes |
| Assume user reads carefully | Design for skimmers |

---

## Execution Checklist

- [ ] Read and understood original spec
- [ ] Ran 10+ simulation iterations
- [ ] Identified failure in each iteration
- [ ] Extracted concrete lessons
- [ ] Categorized by severity
- [ ] Created analysis artifact
- [ ] Applied all critical fixes to spec v2
- [ ] Applied important fixes to spec v2
- [ ] Documented nice-to-have for v1.1
- [ ] Created enhanced spec artifact
- [ ] Output summary with paths

---

## Example

**User**: "/pre-mortem .agents/specs/upgrade-assistant-spec.md"

**Agent workflow**:

```markdown
# Read spec
Found: 10 sections covering architecture, tools, safety, conversation flows

# Simulate Iteration 1: Interface Mismatch
What goes wrong: upgrade.py --json returns different fields than spec
Lesson: Need actual JSON schemas, not examples
Enhancement: [ ] Document actual upgrade.py JSON output schema

# Simulate Iteration 2: Timeout
What goes wrong: doctor scan takes 90s, tool times out at 60s
Lesson: Different operations need different timeouts
Enhancement: [ ] Per-tool timeout configuration

# Simulate Iteration 3: RAG Quality
What goes wrong: RAG returns irrelevant docs for specific errors
Lesson: Need fallback behavior when RAG doesn't match
Enhancement: [ ] Add "I don't know" escalation flow

... (7 more iterations)

# Analysis
Critical: JSON schemas, safety levels, wave-specific handling
Important: Timeouts, progress feedback, audit trail
Nice-to-have: Real-time updates, RAG testing

# Output
- Created: .agents/specs/2026-01-22-spec-analysis.md
- Created: .agents/specs/2026-01-22-spec-v2.md (714 lines, +137 from v1)
```

---

## References

### JIT-Loadable Documentation

| Topic | Reference |
|-------|-----------|
| Failure patterns | `references/failure-taxonomy.md` |
| Enhancement patterns | `references/enhancement-patterns.md` |
| Simulation prompts | `references/simulation-prompts.md` |
| RAG formatting | `domain-kit/skills/standards/references/rag-formatting.md` |

### Related Skills

- **research**: Deep exploration before writing spec
- **plan**: Decomposing spec into implementation issues
- **vibe**: Validating implementation against spec

---

## When to Use

| Scenario | Use Pre-Mortem? |
|----------|-----------------|
| Complex multi-component spec | Yes |
| External integrations | Yes |
| User-facing workflows | Yes |
| Simple API addition | Probably not |
| Bug fix | No |

**Rule of thumb**: If the spec took more than 30 minutes to write, it's worth 15 minutes of simulation.

---

**Progressive Disclosure**: This skill provides core simulation workflow. For detailed failure taxonomy see `references/failure-taxonomy.md`, for enhancement patterns see `references/enhancement-patterns.md`.
