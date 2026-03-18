# Backend: Codex Session Agents

Concrete agent calls for council judges when running inside a Codex session.

---

## Spawn

Spawn one judge per perspective.

```text
spawn_agent(message="You are judge-1.

Perspective: correctness
Task: validate the target
Target files: ...

Write your analysis to .agents/council/judge-1.md.")

spawn_agent(message="You are judge-2.

Perspective: completeness
Task: validate the target
Target files: ...

Write your analysis to .agents/council/judge-2.md.")
```

## Wait

Wait for the agent ids returned by `spawn_agent`.

```text
wait_agent(ids=["agent-id-1", "agent-id-2"])
```

If one judge needs a correction, use `send_input` with a short follow-up prompt.

## Cleanup

Use `close_agent` for any judge you no longer need.

```text
close_agent(id="agent-id-1")
```

## Key Rules

1. One judge, one perspective.
2. Keep the durable analysis in the output file.
3. Use `send_input` only for short steering messages.
