# Demo: /implement-wave Session

A session showing parallel execution of multiple issues.

---

## Before

Plan created with 6 issues for "Add REST API endpoints":

```
$ bd list --status open
1. [P1] agentops-a1: Create users endpoint      (ready)
2. [P1] agentops-b2: Create orders endpoint     (ready)
3. [P1] agentops-c3: Create products endpoint   (ready)
4. [P2] agentops-d4: Add authentication         (blocked by a1)
5. [P2] agentops-e5: Add order validation       (blocked by b2)
6. [P3] agentops-f6: Add integration tests      (blocked by d4, e5)
```

Wave 1: 3 issues ready (no dependencies)

---

## Session

```
You: /implement-wave

Claude: Checking ready issues...

$ bd ready
1. [P1] agentops-a1: Create users endpoint
2. [P1] agentops-b2: Create orders endpoint
3. [P1] agentops-c3: Create products endpoint

**Wave 1: 3 issues**

Launching sub-agents in parallel...

[Agent 1: agentops-a1] Creating src/routes/users.ts...
[Agent 2: agentops-b2] Creating src/routes/orders.ts...
[Agent 3: agentops-c3] Creating src/routes/products.ts...

[Agents work simultaneously...]

**Agent 1 complete:** users endpoint with CRUD operations
**Agent 2 complete:** orders endpoint with CRUD operations
**Agent 3 complete:** products endpoint with CRUD operations

**Validating combined changes:**

$ npm test
12 passed

$ npm run lint
No issues

**Committing wave:**

$ git add src/routes/
$ git commit -m "feat: add REST API endpoints for users, orders, products

- GET/POST/PUT/DELETE for each resource
- Request validation with zod
- Error handling middleware

Closes: agentops-a1, agentops-b2, agentops-c3"

$ bd close agentops-a1 agentops-b2 agentops-c3
✓ Closed 3 issues

**Wave 1 complete!**

$ bd ready
1. [P2] agentops-d4: Add authentication         (unblocked!)
2. [P2] agentops-e5: Add order validation       (unblocked!)

Wave 2 has 2 issues ready.
```

---

## After Wave 1

- 3 files created in parallel
- 3 issues closed in single commit
- 2 more issues now unblocked

---

## Continuing with Wave 2

```
You: /implement-wave

Claude: Wave 2: 2 issues

[2 agents work in parallel...]

$ git commit -m "feat: add auth and validation middleware"
$ bd close agentops-d4 agentops-e5

Wave 2 complete!

$ bd ready
1. [P3] agentops-f6: Add integration tests

Wave 3: 1 issue (final wave)
```

---

## What You Learned

1. Waves execute all ready issues in parallel
2. Sub-agents are independent - no shared state
3. Single commit captures entire wave
4. Closing issues unblocks dependents for next wave
5. ~3x faster than sequential `/implement`
