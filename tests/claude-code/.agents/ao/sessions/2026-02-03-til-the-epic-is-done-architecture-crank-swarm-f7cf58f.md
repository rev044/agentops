---
session_id: f7cf58f3-8f4a-4b6e-8d0a-6f77d73a81a9
date: 2026-02-03
summary: "til the epic is DONE.

## Architecture: Crank + Swarm

```
Crank (orchestrator)           Swarm (..."
tags:
  - olympus
  - session
  - 2026-02
---

# til the epic is DONE.

## Architecture: Crank + Swarm

```
Crank (orchestrator)           Swarm (...

**Session:** f7cf58f3-8f4a-4b6e-8d0a-6f77d73a81a9
**Date:** 2026-02-03

## Knowledge
- til the epic is DONE.

## Architecture: Crank + Swarm

```
Crank (orchestrator)           Swarm (executor)
    |                              |
    +-> bd ready (wave issues)     |
    |             ...
- til all done

## Example Flow

```
Mayor: "Let's build a user auth system"

1. /plan → Creates tasks:
   #1 [pending] Create User model
   #2 [pending] Add password hashing (blockedBy: #1)
   #3...
- til then:** I get progress updates like the ones you're seeing (token counts, tool usage) but those are just status - not completion.

Right now all 8 agents are still working. When they finish,...
- tility, confidence
2. **`ao pool show <id>`** - Shows full candidate details including provenance trail
3. **`ao pool promote <id>`** - Promotes staged candidates to knowledge base...
- till going? 
- til completion.

So the options are:
1. **Progress notifications** - Automatic, shows tool count and token usage
2. **TaskOutput(block=false)** - Confirms running/not_ready status
3....
- til done.

**Level 1 limitations:**
- Fire and forget
- No mid-execution messaging
- Can't inject prompts or steer
- Find out results only at completion

**Level 2 (Agent Mail) enables:**
-...
- till running...
</system-reminder>
```

2. **Completion notification** (when done):
```
<task-notification>
<task-id>abbb9cd</task-id>
<status>completed</status>
<result>...</result>
</task-notificati...
- til done | Could adjust between each |
| Complexity | Need to track multiple | Simple linear flow |
| File conflicts | Possible race conditions | No conflicts |

**When to use which:**

-...
- insight:**
Each spawn = fresh context = stays effective. If I tried to hold 8 implementation tasks in my head while editing files across all of them, I'd degrade. The agents stay sharp because they...
- till active - 14 tools used, 14k tokens just in that burst. It's working, just taking longer.

**Why swarm tests are slow:**
- Most complex task (testing the orchestration itself)
- Needs to...

## Files Changed
- [[/Users/fullerbt/gt/agentops/crew/nami/.agents/learnings/2026-02-03-cfa98d9a.md]]
- [[/Users/fullerbt/gt/agentops/crew/nami/.agents/plans/2026-02-03-ao-skills-validation.md]]
- [[/private/tmp/claude-501/-Users-fullerbt-gt-agentops-crew-nami/tasks/adb3838.output]]
- [[/private/tmp/claude-501/-Users-fullerbt-gt-agentops-crew-nami/tasks/a8f0caf.output]]

## Issues
- [[issues/ag-p43|ag-p43]]
- [[issues/pre-mortem|pre-mortem]]
- [[issues/ao-skills-validation|ao-skills-validation]]
- [[issues/pre-mortems|pre-mortems]]
- [[issues/per-wave|per-wave]]
- [[issues/rev-parse|rev-parse]]
- [[issues/mcp-agent-mail|mcp-agent-mail]]
- [[issues/ol-527|ol-527]]
- [[issues/ol-527-1|ol-527-1]]
- [[issues/ol-527-2|ol-527-2]]
- [[issues/ol-527-3|ol-527-3]]
- [[issues/mcp-tools|mcp-tools]]
- [[issues/max-workers|max-workers]]
- [[issues/new-session|new-session]]
- [[issues/has-session|has-session]]
- [[issues/per-bead|per-bead]]
- [[issues/mid-swarm|mid-swarm]]
- [[issues/bug-hunt|bug-hunt]]
- [[issues/non-zero|non-zero]]
- [[issues/go-test|go-test]]
- [[issues/run-all|run-all]]
- [[issues/ag-p43-1|ag-p43-1]]
- [[issues/and-forget|and-forget]]

## Tool Usage

| Tool | Count |
|------|-------|
| Bash | 43 |
| Read | 3 |
| Skill | 1 |
| Task | 8 |
| TaskCreate | 8 |
| TaskList | 1 |
| TaskOutput | 3 |
| TaskUpdate | 7 |
| Write | 1 |

## Tokens

- **Input:** 0
- **Output:** 0
- **Total:** ~110801 (estimated)
