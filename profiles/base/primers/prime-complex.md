---
description: Multi-phase Research→Plan→Implement with interactive JIT loading
---

# /prime-complex - Research→Plan→Implement Context Router

**Purpose:** Guide complex multi-step tasks through research, planning, and implementation phases.

**Workflow:** Research → Plan → Implement (3 phases, human review between each)

**Token budget:** <40% per phase (80k of 200k per phase)

---

## Step 1: Constitutional Baseline (Always Enforced)

{{cat .claude/CONSTITUTION.md}}

**Status:** ✅ Constitutional foundation loaded (2k tokens)

You are now operating under AgentOps constitution with multi-phase workflow:
- **Research phase:** Gather widely, understand system (40-60k tokens)
- **Plan phase:** Specify changes, get approval (40-60k tokens)
- **Implement phase:** Execute plan, validate (40-80k tokens)

**Key insight:** Planning IS the work. If you're shouting at Claude during implementation, the plan was incomplete.

---

## Step 2: Understand Your Complex Task

**What complex task are you working on?**

Choose from common complex workflows:

### Creating Complex Systems
- **New multi-service application** - App with multiple dependencies
- **Infrastructure provisioning** - Crossplane, Kyverno policies
- **Platform feature** - Cross-cutting functionality

### Major Modifications
- **Architecture refactoring** - Restructure existing systems
- **Security hardening** - Credential management, policy enforcement
- **Performance optimization** - Profiling, tuning, scaling

### Investigation & Analysis
- **Architecture review** - Identify improvement opportunities
- **Incident investigation** - Complex failure analysis
- **Technical research** - Evaluate new technologies

### Documentation & Planning
- **Comprehensive documentation** - Multi-doc projects
- **Migration planning** - Move between systems
- **Capacity planning** - Resource forecasting

### Other
- Describe your complex task in your own words

---

## Step 3: The Three-Phase Pattern

### Phase 1: Research (30-40% of time)

**Goal:** Understand the system deeply before planning

**Activities:**
1. Map the system (files, dependencies, patterns)
2. Find similar implementations (search codex, git history)
3. Locate the problem or requirement
4. Identify constraints and edge cases

**Output:** Research bundle (500-1k tokens)
```markdown
# Research Findings: [Task Name]

## Problem Location
- File: [path/to/file.yaml:line]
- Root cause: [why it's happening]

## Similar Implementations
- Pattern: [existing pattern name]
- Location: [where to find it]
- Adaptation needed: [what to change]

## Files Involved
1. [path/to/file1] - [what needs to change]
2. [path/to/file2] - [what needs to change]
```

**Token budget:** 40-60k (20-30% of window)

**Human review checkpoint:** ✓

---

### Phase 2: Plan (30-40% of time)

**Goal:** Specify EVERY change with file:line precision

**Activities:**
1. Load research bundle + constitution (fresh context)
2. Specify exact changes (file:line, what to edit)
3. Define test strategy
4. Document rollback procedure

**Output:** Plan bundle (1-2k tokens)
```markdown
# Implementation Plan: [Task Name]

## Changes Specified
1. Edit [file:line] - [specific change]
2. Add [file] - [what to add]
3. Validate with [command]

## Test Strategy
- Unit: [test commands]
- Integration: [validation steps]
- Rollback: [how to undo if needed]

## Approval
✓ Human reviewed: [date]
✓ Approach validated
```

**Token budget:** 40-60k (20-30% of window)

**Human review checkpoint:** ✓

---

### Phase 3: Implement (20-30% of time)

**Goal:** Execute the plan (trust it, don't redesign)

**Activities:**
1. Load plan bundle + constitution (fresh context)
2. Execute each change in plan
3. Validate as you go
4. Capture learnings for future

**Output:** Commit with Context/Solution/Learning/Impact

**Token budget:** 40-80k (20-40% of window)

**Human review checkpoint:** ✓ (final validation)

---

## Step 4: JIT Load Domain Context

**[After you describe your task, I will:]**

1. Analyze task complexity and domain
2. Load relevant patterns from `docs/reference/workflows/`
3. Suggest complex workflows from `docs/reference/workflows/COMMON_WORKFLOWS.md`
4. Guide you through Research phase first

**Examples of what gets loaded in each phase:**

→ **Task:** "Create new multi-service application"

**Research Phase:**
- Constitution (2k)
- application-creation.md (1k)
- helm-patterns.md (0.8k)
- Search existing apps (20-40k exploration)
- Total: 23-43k (11-21%)

**Plan Phase:**
- Constitution (2k)
- Research bundle (1k)
- deployment-strategies.md (0.7k)
- Specify changes (20-40k planning)
- Total: 23-43k (11-21%)

**Implement Phase:**
- Constitution (2k)
- Plan bundle (1.5k)
- Execute changes (30-60k work)
- Total: 33-63k (16-31%)

---

→ **Task:** "Refactor authentication system"

**Research Phase:**
- Constitution (2k)
- security-hardening.md (0.7k)
- Explore current implementation (30-50k)
- Total: 32-52k (16-26%)

**Plan Phase:**
- Constitution (2k)
- Research bundle (1k)
- Specify refactoring (30-50k)
- Total: 33-53k (16-26%)

**Implement Phase:**
- Constitution (2k)
- Plan bundle (1.5k)
- Execute refactoring (40-70k)
- Total: 43-73k (21-36%)

---

## Step 5: Context Bundle Pattern

Between phases, I'll save bundles for continuity:

**After Research:**
```bash
cat > .agents/bundles/task-[name]-research.md
```

**After Plan:**
```bash
cat > .agents/bundles/task-[name]-plan.md
```

**Each new phase:**
1. Start with fresh context window
2. Load constitution (2k)
3. Load previous bundle (1-2k)
4. Continue with full headroom (195-197k available)

**This prevents context collapse across phases!**

---

## Token Budget Per Phase

```text
Phase 1: Research
  Constitution:    ████░░░░░░  2k/200k   (1%)
  Domain pattern:  ██░░░░░░░░  1k/200k   (0.5%)
  Exploration:     ████████████ 40k/200k  (20%)
  Reserved:        ░░░░░░░░░░  157k/200k (78.5%)
  Status: 🟢 GREEN

Phase 2: Plan
  Constitution:    ████░░░░░░  2k/200k   (1%)
  Research bundle: ██░░░░░░░░  1k/200k   (0.5%)
  Planning:        ████████████ 40k/200k  (20%)
  Reserved:        ░░░░░░░░░░  157k/200k (78.5%)
  Status: 🟢 GREEN

Phase 3: Implement
  Constitution:    ████░░░░░░  2k/200k   (1%)
  Plan bundle:     ███░░░░░░░  1.5k/200k (0.75%)
  Execution:       ████████████ 50k/200k  (25%)
  Reserved:        ░░░░░░░░░░  146k/200k (73%)
  Status: 🟢 GREEN
```

---

## What Happens Next?

**I will:**
1. ✅ Understand your complex task (from your description)
2. ✅ Guide you through Phase 1: Research first
3. ✅ JIT load relevant patterns for research
4. ✅ Create research bundle after exploration
5. ✅ Get your review before moving to Phase 2
6. ✅ Repeat for Plan and Implement phases

**You do:**
1. Describe your complex task (see Step 2 above)
2. Review research findings (end of Phase 1)
3. Approve plan (end of Phase 2)
4. Validate implementation (end of Phase 3)

---

## Why Three Phases?

**Based on learning science:**
- **Research = Schema formation** (80% of cognitive effort)
- **Plan = Knowledge organization** (making it actionable)
- **Implement = Execution** (20% of cognitive effort)

**"Planning IS the work"** - If implementation feels hard, the plan was incomplete.

---

## Related Patterns

- **Context bundles:** Prevent context collapse between phases
- **40% rule:** Stay under budget per phase
- **Human review gates:** Validate approach before execution
- **JIT loading:** Load patterns when needed

---

**Ready! What complex task are you working on?** (Describe or choose from categories above)
