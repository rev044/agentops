# Backend: Background Tasks (Fallback)

Concrete guidance for runtimes that do not provide native Codex session agents.

**Limitations:**
- Fire-and-forget
- No messaging
- No retry steering
- No shared state beyond files

## Spawn

Use the runtime's background-task primitive to start one worker per task.

```text
BACKGROUND_TASK(task_id="abc-123", prompt="You are judge-1 ...")
BACKGROUND_TASK(task_id="def-456", prompt="You are worker-3 ...")
```

The exact transport varies by runtime. The important part is the contract: one prompt, one output file, one worker.

## Wait

Poll for completion using the runtime's task wait primitive.

```text
BACKGROUND_WAIT(task_id="abc-123", timeout=120000)
BACKGROUND_WAIT(task_id="def-456", timeout=120000)
```

After completion, verify the worker wrote the expected result file.

## Cleanup

Background tasks self-terminate when done. If a worker stalls, treat it as failed and re-spawn a fresh one.

## Key Rules

1. Filesystem is the only communication channel.
2. No messaging means no follow-up round.
3. Verify the result file, not just the wait status.
4. Prefer native Codex session agents when available.
