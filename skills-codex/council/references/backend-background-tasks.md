# Backend: Background Tasks (Fallback)

Fallback guidance for council when native Codex session agents are unavailable.

**Limitations:**
- No messaging
- No debate rounds
- No retry steering

## Spawn

Use the runtime's background-task primitive to start one judge per perspective.

```text
BACKGROUND_TASK(task_id="abc-123", prompt="You are judge-1 ...")
BACKGROUND_TASK(task_id="def-456", prompt="You are judge-2 ...")
```

## Wait

Wait for each task id and then verify the output file.

```text
BACKGROUND_WAIT(task_id="abc-123", timeout=120000)
BACKGROUND_WAIT(task_id="def-456", timeout=120000)
```

## Cleanup

Background tasks self-terminate when done. If a task stalls, re-spawn it from scratch.

## Key Rules

1. One task, one output file.
2. Verify files, not just completion status.
3. Prefer native Codex session agents when available.
