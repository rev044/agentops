# /prime-simple - Quick Orientation for Straightforward Tasks

**Purpose:** Fast context loading for single-task execution

**Philosophy:** Minimum viable context for clear, simple tasks

**Token budget:** 2-3k tokens

**Output:** Ready to execute immediately

---

## When to Use

Use `/prime-simple` when:
- ✅ Task is clear and straightforward
- ✅ No research or planning needed
- ✅ Single file or small change
- ✅ Pattern already known
- ✅ Quick fix or update

**Don't use if:**
- ❌ Need to research approach
- ❌ Multi-step or complex task
- ❌ Unfamiliar with codebase area
- ❌ Architectural decisions needed

---

## How It Works

### Step 1: Load Minimal Context
```
Loading:
- CONSTITUTION.md (2k tokens)
- Task-specific pattern (if available, 500 tokens)
- Validation commands (300 tokens)

Total: ~2.5k tokens (1.25% context)
```

### Step 2: Execute Directly
**No intermediate phases:**
- Skip research (pattern known)
- Skip formal planning (straightforward)
- Execute change
- Validate
- Commit

### Step 3: Quick Validation
**Run:**
- Syntax check
- Basic functionality test
- Commit if passing

---

## Examples

### Example 1: Add Environment Variable
```
User: /prime-simple

User: Add DATABASE_URL to .env file

Agent: [Loads constitution + env file patterns]

Agent: Adding DATABASE_URL to .env...

File: .env:15
Add: DATABASE_URL=postgres://localhost:5432/mydb

Validation: ✅ Syntax valid

Commit: feat(config): add DATABASE_URL environment variable

Done! (Used 2.3k tokens, 1.15% context)
```

### Example 2: Fix Typo
```
User: /prime-simple

User: Fix typo in README.md line 45: "teh" → "the"

Agent: [Loads constitution]

Agent: Fixing typo...

File: README.md:45
Change: teh → the

Validation: ✅ Markdown valid

Commit: fix(docs): correct typo in README

Done! (Used 2.1k tokens, 1.05% context)
```

### Example 3: Update Dependency
```
User: /prime-simple

User: Update redis version to 7.0.12 in requirements.txt

Agent: [Loads constitution + dependency patterns]

Agent: Updating dependency...

File: requirements.txt:23
Change: redis==7.0.10 → redis==7.0.12

Validation: ✅ Syntax valid, no breaking changes

Commit: chore(deps): update redis to 7.0.12

Done! (Used 2.4k tokens, 1.2% context)
```

---

## What Tasks Qualify as Simple?

**✅ Simple (use prime-simple):**
- Fix typos or formatting
- Add/update environment variables
- Update version numbers
- Add comments or documentation
- Small config tweaks
- Known pattern application

**❌ Not Simple (use prime-complex):**
- New features
- Architectural changes
- Unfamiliar codebase areas
- Multiple interconnected changes
- Requires research or design

---

## Success Criteria

Prime-simple is successful when:
- ✅ Task completed in one pass
- ✅ No research or planning needed
- ✅ Validation passes
- ✅ Context under 3k tokens
- ✅ Total time < 5 minutes

---

## When to Escalate

If during execution you realize:
- Task is more complex than expected
- Need to research approach
- Multiple changes interconnected
- Validation failing repeatedly

**Then escalate:**
```
User: This is more complex than I thought

Agent: Let's use /prime-complex instead.
       Starting research phase...
```

---

## Integration with Workflow

```
Start: /prime-simple
  ↓
Understand task (clear and simple)
  ↓
Execute change directly
  ↓
Validate
  ↓
Pass → Commit → Done
Fail → Escalate to /prime-complex
```

---

*Use for: Quick fixes, known patterns, single-file changes*
*Skip for: Architecture, research, multi-step work*
