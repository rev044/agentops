---
name: shared
description: Shared reference documents for distributed mode skills (not directly invocable)
metadata:
  tier: library
  internal: true
---

# Shared References

This directory contains shared reference documents used by multiple skills:

- `agent-mail-protocol.md` - Message protocol for distributed mode coordination
- `validation-contract.md` - Verification requirements for accepting spawned work

These are **not directly invocable skills**. They are loaded by other skills (crank, swarm, inbox, implement) when needed for distributed mode operation.

---

## CLI Availability Pattern

All skills that reference external CLIs MUST degrade gracefully when those CLIs are absent.

### Check Pattern

```bash
# Before using any external CLI, check availability
if command -v bd &>/dev/null; then
  # Full behavior with bd
else
  echo "Note: bd CLI not installed. Using plain text tracking."
  # Fallback: use TaskList, plain markdown, or skip
fi
```

### Fallback Table

| Capability | When Missing | Fallback Behavior |
|------------|-------------|-------------------|
| `bd` | Issue tracking unavailable | Use TaskList for tracking. Note "install bd for persistent issue tracking" |
| `ao` | Knowledge flywheel unavailable | Write learnings to `.agents/knowledge/` directly. Skip flywheel metrics |
| `gt` | Workspace management unavailable | Work in current directory. Skip convoy/sling operations |
| `codex` | CLI missing or model unavailable | Fall back to runtime-native agents (Codex sub-agents if available, else Claude teams/task fallback). Council pre-flight checks CLI presence (`which codex`) and model availability for `--mixed` mode. |
| `cass` | Session search unavailable | Skip transcript search. Note "install cass for session history" |
| Native teams | Any capability missing | See "Native Teams Capability Bundle" below for per-capability degradation |

### Native Teams Capability Bundle

Native teams consist of multiple independent capabilities. Each can degrade independently:

| Capability | API | Degraded Behavior |
|------------|-----|-------------------|
| Team lifecycle | `TeamCreate`, `TeamDelete` | Fall back to `Task(run_in_background=true)`. No team cleanup needed. |
| Directed messaging | `SendMessage(type="message")` | Cannot send follow-up instructions. Debate R2 unavailable. Workers run fire-and-forget. |
| Broadcast | `SendMessage(type="broadcast")` | Cannot notify all workers. Use per-worker Task spawning instead. |
| Shutdown coordination | `SendMessage(type="shutdown_request/response")` | Workers terminate on their own when done. No graceful shutdown. |
| Shared task list | `TaskList`, `TaskCreate`, `TaskUpdate` | Use in-memory tracking. Workers cannot see shared state. |

**Degradation matrix:**

| Scenario | Team | Messaging | TaskList | Impact |
|----------|------|-----------|----------|--------|
| Full native teams | Yes | Yes | Yes | All features available |
| TeamCreate fails | No | No | Yes | Fire-and-forget workers, no debate |
| SendMessage fails | Yes | No | Yes | Workers isolated, no R2 debate |
| TaskList fails | Yes | Yes | No | Lead tracks manually, workers report via message |

### Runtime-Native Spawn Backend Selection

All orchestration skills that spawn parallel workers or judges MUST select backend in this order:

1. **Codex experimental sub-agents** (when `spawn_agent` is available)
2. **Claude native teams** (`TeamCreate` + `Task(team_name=...)` + `SendMessage`)
3. **Background task fallback** (`Task(run_in_background=true)`)

Use capability detection, not hardcoded agent assumptions. The same skill must run in both Claude and Codex sessions.

| Operation | Codex Sub-Agents | Claude Native Teams | Background Fallback |
|-----------|------------------|---------------------|---------------------|
| Spawn | `spawn_agent(message=...)` | `TeamCreate` + `Task(team_name=...)` | `Task(run_in_background=true)` |
| Wait | `wait(ids=[...])` | Completion via `SendMessage` | `TaskOutput(..., block=true)` |
| Retry/follow-up | `send_input(id=..., message=...)` | `SendMessage(type="message", ...)` | Re-spawn with revised prompt |
| Cleanup | `close_agent(id=...)` | `shutdown_request` + `TeamDelete()` | None |

### Rules

1. **Never crash** — missing CLI = skip or fallback, not error
2. **Always inform** — tell the user what was skipped and how to enable it
3. **Preserve core function** — the skill's primary purpose must still work without optional CLIs
4. **Progressive enhancement** — CLIs add capabilities, their absence removes them cleanly
