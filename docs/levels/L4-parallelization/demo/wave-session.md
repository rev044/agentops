# Demo: /swarm Session

A session showing parallel execution of multiple tasks using the swarm pattern.

---

## Before

Plan created with tasks for "Add REST API endpoints":

```
$ TaskList()

#1 [pending] Create users endpoint
#2 [pending] Create orders endpoint
#3 [pending] Create products endpoint
#4 [pending] Add authentication (blockedBy: #1)
#5 [pending] Add order validation (blockedBy: #2)
#6 [pending] Add integration tests (blockedBy: #4, #5)
```

Wave 1: 3 tasks ready (no dependencies)

---

## Session

```
You: /swarm

Claude: Checking TaskList for ready tasks...

Ready for Wave 1:
#1 [pending] Create users endpoint
#2 [pending] Create orders endpoint
#3 [pending] Create products endpoint

Spawning 3 background agents...

Task(subagent_type="general-purpose", run_in_background=true, ...)
Task(subagent_type="general-purpose", run_in_background=true, ...)
Task(subagent_type="general-purpose", run_in_background=true, ...)

[Agents work in parallel, each with fresh context...]

<task-notification> Agent for #1 completed
<task-notification> Agent for #2 completed
<task-notification> Agent for #3 completed

All Wave 1 agents complete!

TaskUpdate(taskId="1", status="completed")
TaskUpdate(taskId="2", status="completed")
TaskUpdate(taskId="3", status="completed")

Validating combined changes...

$ npm test
12 passed

$ npm run lint
No issues

Wave 1 complete!

TaskList() shows:
#1 [completed] Create users endpoint
#2 [completed] Create orders endpoint
#3 [completed] Create products endpoint
#4 [pending] Add authentication         ← now unblocked!
#5 [pending] Add order validation       ← now unblocked!
#6 [pending] Add integration tests (blockedBy: #4, #5)
```

---

## After Wave 1

- 3 background agents ran in parallel
- Each had fresh context (Ralph Wiggum pattern)
- 3 tasks marked completed
- 2 more tasks now unblocked

---

## Continuing with Wave 2

```
You: /swarm

Claude: Wave 2: 2 tasks ready

Spawning background agents for #4 and #5...

[2 agents work in parallel with fresh context...]

<task-notification> Agent for #4 completed
<task-notification> Agent for #5 completed

Wave 2 complete!

TaskList() shows:
#6 [pending] Add integration tests      ← now unblocked!

Wave 3: 1 task (final wave)
```

---

## Wave 3 (Final)

```
You: /swarm

Claude: Wave 3: 1 task ready

Spawning background agent for #6...

<task-notification> Agent for #6 completed

All tasks complete!

$ git status
3 new files, 2 modified

$ git add -A && git commit -m "feat: complete REST API implementation"
```

---

## What You Learned

1. `/swarm` spawns background agents for all ready tasks
2. Each agent has fresh context (Ralph Wiggum pattern)
3. Mayor receives `<task-notification>` as agents complete
4. Mayor updates TaskList and triggers next wave
5. No context accumulation = stays effective at scale
