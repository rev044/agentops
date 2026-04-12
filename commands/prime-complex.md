# /prime-complex - Deep Orientation for Multi-Phase Work

**Purpose:** Comprehensive context loading for research → plan → implement cycles

**Philosophy:** Prepare for multi-day, multi-phase work with proper boundaries

**Token budget:** 5-10k tokens

**Output:** Ready for phase-based execution

---

## When to Use

Use `/prime-complex` when:
- ✅ Task requires research or exploration
- ✅ Multiple phases needed (research/plan/implement)
- ✅ Architectural decisions involved
- ✅ Unfamiliar codebase or domain
- ✅ Multi-day or multi-session work

**Don't use if:**
- ❌ Simple, straightforward change (use `/prime-simple`)
- ❌ Pattern already known
- ❌ Single file, quick fix

---

## How It Works

### Step 1: Load Comprehensive Context
```
Loading:
- CONSTITUTION.md (2k tokens)
- Phase-based workflow patterns (2k tokens)
- Multi-agent orchestration patterns (2k tokens)
- Context bundle system (1k tokens)
- Domain-specific patterns (if available, 1-2k tokens)

Total: 7-10k tokens (3.5-5% context)
```

### Step 2: Establish Phase Boundaries
**Explain the cycle:**
```
Phase 1: RESEARCH (40-60k tokens)
├─ Multi-agent exploration
├─ Gather comprehensive understanding
├─ Compress to bundle (1.2k)
└─ Save bundle, end session

Phase 2: PLAN (40-60k tokens, fresh session)
├─ Load research bundle (1.2k)
├─ Design detailed specifications
├─ File:line plans with validation
├─ Approval gate
├─ Compress to bundle (1.5k)
└─ Save bundle, end session

Phase 3: IMPLEMENT (40-80k tokens, fresh session)
├─ Load plan bundle (1.5k)
├─ Execute approved changes
├─ Continuous validation
├─ Create commit
└─ Push (final approval)
```

### Step 3: Clarify Current Phase
**I will ask:**
- "Are you starting fresh (research)?"
- "Do you have research to plan?"
- "Do you have plan to implement?"
- "Continuing previous work?"

### Step 4: Launch Appropriate Phase
**Based on answer:**
- Fresh start → `/research`
- Have research → `/plan`
- Have plan → `/implement`
- Continuing → Load relevant bundle

---

## Example Session (Full Cycle)

### Session 1: Research
```
User: /prime-complex

Agent: Loading multi-phase context (8k tokens)...

What are you working on?

User: Need to add Redis caching to our API

Agent: This requires research. Starting Phase 1...

Launching /research "Redis caching for API"...

[3 parallel agents explore for 20 min]

Research complete:
- Recommendation: Redis with pub/sub
- Bundle compressed: 45k → 1.2k tokens
- Saved: .agentops/bundles/redis-caching-research.md

Next step: Load this bundle in fresh session for planning

[End Session 1]
```

### Session 2: Planning
```
User: /prime-complex

Agent: Loading multi-phase context...

Continuing previous work?

User: Yes, I have research bundle: redis-caching-research

Agent: Loading bundle... (1.2k tokens loaded)

Launching /plan based on research...

Plan created:
- 5 files to create
- 3 files to modify
- Validation strategy defined
- Rollback procedure documented

Approval required before implementation.
Review plan?

User: Approved

Agent: Plan compressed: 50k → 1.5k tokens
Saved: .agentops/bundles/redis-caching-plan.md

Next step: Load this bundle in fresh session for implementation

[End Session 2]
```

### Session 3: Implementation
```
User: /prime-complex

Agent: Loading multi-phase context...

Continuing previous work?

User: Yes, I have plan bundle: redis-caching-plan

Agent: Loading approved plan... (1.5k tokens loaded)

Launching /implement...

Implementation complete:
✅ 5 files created
✅ 3 files modified
✅ All validations passed
✅ Commit created

Ready to git push?

User: Yes

Agent: Pushed successfully.

[End Session 3 - Task Complete]
```

---

## What Gets Loaded

**Always (8-10k):**
- CONSTITUTION.md
- Phase workflow patterns
- Multi-agent orchestration
- Context bundle system
- Validation framework

**Conditionally (2-3k each):**
- Previous phase bundle (if continuing)
- Domain patterns (if known)
- Similar work examples (if available)

**Never until needed:**
- Full codebase context
- All documentation
- Entire command catalog

**Result:** Prepared for multi-phase work, 10-13k total tokens (5-6.5%)

---

## Phase Management

### Phase Boundaries
**Each phase gets:**
- Fresh context window (new session)
- 40-60k token budget
- Specific deliverable (bundle)
- Approval gate (between phases)

### Context Compression
**Between phases:**
- Compress previous work (37:1 ratio)
- Load compressed bundle (1-2k)
- Continue with clean context

### Validation Gates
**Before next phase:**
- Research → Approve findings
- Plan → Approve specifications
- Implement → Review changes

---

## Benefits

1. **No context collapse** - Fresh window per phase
2. **Multi-day viable** - Bundles enable continuation
3. **Quality maintained** - 40% rule enforced
4. **Clear milestones** - Each phase has deliverable
5. **Reversible** - Can go back to earlier phase

---

## When to Use Each Phase

### Research Phase
**Use when:**
- Need to understand problem
- Evaluate multiple approaches
- Unfamiliar with domain
- Constraints unknown

### Plan Phase
**Use when:**
- Research complete
- Ready to specify changes
- Need file:line detail
- Want approval before coding

### Implement Phase
**Use when:**
- Plan approved
- Ready to execute
- All specs clear
- Validation defined

---

## Success Criteria

Prime-complex is successful when:
- ✅ Multi-phase workflow understood
- ✅ Current phase identified
- ✅ Appropriate phase launched
- ✅ Context under 10% after prime
- ✅ Phase boundaries respected

---

## Integration with Other Commands

```
Complex Work:
  /prime-complex
    ↓
  Identify phase
    ↓
  /research → bundle → [end session]
    ↓
  [new session] → /prime-complex → load bundle
    ↓
  /plan → bundle → [end session]
    ↓
  [new session] → /prime-complex → load bundle
    ↓
  /implement → commit → push
```

---

*Use for: Architecture, research, multi-step work, multi-day projects*
*Skip for: Quick fixes, known patterns, single changes*
