# Backend: Background Tasks (Fallback)

This fallback is host-runtime specific and is **not** part of the Codex session surface. In Codex sessions, prefer the native sub-agent path in `backend-codex-subagents.md` and keep orchestration on `spawn_agent`, `wait_agent`, `send_input`, and `close_agent`.

**Limitations:**
- Fire-and-forget
- No inter-agent communication
- No debate mode
- No retry; failures require a fresh spawn

---

## Host Runtime Guidance

Use the host runtime's background-task API to spawn isolated workers and poll their completion handles. Keep each worker prompt self-contained, write results to files, and verify file output before proceeding.

### Recommended Pattern

- Put one worker per task or batch.
- Give each worker a clear file write target.
- Poll the worker handle until completion.
- Verify the result file exists before reading it.

---

## Completion Check

Poll the worker handle until it reports completion. If the host runtime times out, check the result file first, then decide whether to proceed with partial results or respawn.

**Fallback:** If background tasks fail despite detection, fall back to inline mode. See `backend-inline.md`.

---

## No Messaging

Background tasks cannot receive messages. This means:

- **No debate R2** — judges get one round only
- **No retry** — if validation fails, re-spawn a new agent from scratch
- **No scope adjustment** — the prompt is final at spawn time

---

## Cleanup

Background tasks self-terminate when done. For stuck tasks, use the host runtime's cleanup primitive if available. Partial work may be lost.

---

## Key Rules

1. **Filesystem is the only communication channel**
2. **No messaging = no debate**
3. **No retry = must respawn**
4. **Always check result files**
5. **Prefer native Codex sub-agents when available**
