# /prime - Interactive JIT Context Router

**Purpose:** Start every session with intelligent context loading

**Philosophy:** Load only what you need, when you need it. Don't front-load the entire system.

**Token budget:** 2-5k tokens

**Output:** Prepared session with minimal context

---

## When to Use

Use `/prime` to start EVERY session before any work:
- ✅ Beginning of new session
- ✅ Switching between tasks
- ✅ Need guidance on what command to use next
- ✅ Unsure which workflow applies

**This is your entry point to AgentOps.**

---

## How It Works

### Step 1: Load Constitution
```
Loading: core/CONSTITUTION.md (2k tokens)
- Five Laws of an Agent
- 40% rule
- Core rules
```

### Step 2: Understand Your Task
**I will ask:**
- "What are you working on?"
- "What's the goal of this session?"
- "Do you have context from previous sessions?"

### Step 3: JIT Load Relevant Context
Based on your answer, I'll load ONLY relevant:
- Commands you might need
- Workflows that apply
- Patterns from similar work

**Not loaded:** Everything else (saves tokens)

### Step 4: Suggest Next Steps
**I will recommend:**
- `/prime-simple` if straightforward task
- `/prime-complex` if multi-phase work
- Specific workflow if clear match
- `/research` if exploration needed

---

## Example Session

```
User: /prime

Agent: Constitution loaded (2k tokens, 1% context)

What are you working on today?

User: I need to containerize a Python application

Agent: I'll load container and DevOps context...

Loaded (3k total):
- Container building patterns
- Dockerfile best practices
- DevOps profile commands

Suggested workflows:
1. /containerize-app (DevOps workflow - full process)
2. /prime-complex → /research (if unfamiliar with app)
3. /quick-fix (if just need Dockerfile)

Which approach fits your task?

User: Let's do the full containerize-app workflow

Agent: Great! Launching /containerize-app...
[Continues with workflow]
```

---

## What Gets Loaded

**Always loaded (2k):**
- CONSTITUTION.md (Five Laws, 40% rule)

**Conditionally loaded (1-3k each):**
- Profile-specific commands (if DevOps → container commands)
- Relevant workflows (if product dev → spec workflows)
- Pattern library (if similar work done before)
- Domain context (if K8s → K8s patterns)

**Never loaded until needed:**
- Full command catalog
- All agents
- All workflows
- Entire documentation

**Result:** Start with 2-5k tokens (1-2.5% context), not 40k

---

## Benefits

1. **Fast start** - 2-5k vs 40k tokens
2. **Guided** - Suggests relevant next steps
3. **Flexible** - Loads more if needed
4. **Smart** - Learns from your work patterns

---

## Integration with Other Commands

```
Session Start:
  /prime → Understand task → Suggest workflow

Simple Task:
  /prime → /prime-simple → Execute directly

Complex Task:
  /prime → /prime-complex → /research → /plan → /implement

Unclear Task:
  /prime → Ask questions → Clarify → Route appropriately
```

---

## Success Criteria

Prime is successful when:
- ✅ Constitution loaded
- ✅ User's intent understood
- ✅ Relevant context loaded (not everything)
- ✅ Next steps suggested
- ✅ Context under 5k tokens

---

*This is Law #5 in action: Guide with workflow suggestions, let user choose.*
