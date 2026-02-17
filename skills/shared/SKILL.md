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
3. **OpenCode subagents** (when `skill` tool is read-only — see detection below)
4. **Inline fallback** (no spawn capability — execute work in current turn)

Use capability detection, not hardcoded agent assumptions. The same skill must run across Claude Code, Codex, and OpenCode sessions.

**Detection heuristic:** If your `skill` tool loads content into context (returns `<skill_content>` blocks) rather than executing skills, you are in OpenCode. OpenCode has a `task` tool with built-in agent types: `general`, `explore`, `build`, `plan`.

| Operation | Codex Sub-Agents | Claude Native Teams | OpenCode Subagents | Inline Fallback |
|-----------|------------------|---------------------|--------------------|-----------------|
| Spawn | `spawn_agent(message=...)` | `TeamCreate` + `Task(team_name=...)` | `task(subagent_type="general", prompt=...)` | Execute inline |
| Spawn (read-only) | `spawn_agent(message=...)` | `Task(subagent_type="Explore")` | `task(subagent_type="explore", prompt=...)` | Execute inline |
| Wait | `wait(ids=[...])` | Completion via `SendMessage` | Task returns result directly | N/A |
| Retry/follow-up | `send_input(id=..., message=...)` | `SendMessage(type="message", ...)` | `task(task_id="<prior>", prompt=...)` | N/A |
| Cleanup | `close_agent(id=...)` | `shutdown_request` + `TeamDelete()` | None (sub-sessions auto-terminate) | N/A |
| Inter-agent messaging | `send_input` | `SendMessage` | Not available | N/A |
| Debate (R2) | Supported | Supported | **Not supported** (no messaging) | N/A |

**OpenCode limitations:**
- No inter-agent messaging — workers run as independent sub-sessions
- No debate mode (`--debate`) — requires messaging between judges
- `--quick` (inline) mode works identically across all backends

### Backend Capabilities Matrix

> **Prefer native teams over background tasks.** Native teams provide messaging, redirect, and graceful shutdown. Background tasks are fire-and-forget with no steering — only a speedometer and emergency brake.

| Capability | Codex Sub-Agents | Claude Native Teams | Background Tasks | Distributed (tmux) |
|------------|------------------|---------------------|------------------|---------------------|
| Observe output | `wait()` result | `SendMessage` delivery | `TaskOutput` (tail) | Agent Mail inbox |
| Send message mid-flight | `send_input` | `SendMessage` | **NO** | Agent Mail |
| Pause / resume | NO | Idle → wake via `SendMessage` | **NO** | `tmux` detach/attach |
| Graceful stop | `close_agent` | `shutdown_request` | **TaskStop (lossy)** | `tmux kill-session` |
| Redirect to different task | `send_input` | `SendMessage` | **NO** | Agent Mail |
| Adjust scope mid-flight | `send_input` | `SendMessage` | **NO** | Agent Mail |
| File conflict prevention | Worktree (planned) | Lead-only commits | None | File reservations |
| Crash recovery | NO | NO | NO | **YES** (tmux persists) |
| Process isolation | YES (sub-process) | Shared worktree | Shared worktree | **YES** (separate process) |

**When to use each:**

| Scenario | Backend |
|----------|---------|
| Quick parallel tasks, coordination needed | Claude Native Teams |
| Codex-specific execution | Codex Sub-Agents |
| Long-running work, need debug/recovery | Distributed (tmux + Agent Mail) |
| No team APIs available (last resort) | Background Tasks |

### Skill Invocation Across Runtimes

Skills that chain to other skills (e.g., `/rpi` calls `/research`, `/vibe` calls `/council`) MUST handle runtime differences:

| Runtime | Tool | Behavior | Pattern |
|---------|------|----------|---------|
| Claude Code | `Skill(skill="X", args="...")` | **Executable** — skill runs as a sub-invocation | `Skill(skill="council", args="--quick validate recent")` |
| Codex | N/A | Skills not available — inline the logic or skip | Check if `Skill` tool exists before calling |
| OpenCode | `skill` tool (read-only) | **Load-only** — returns `<skill_content>` blocks into context | Call `skill(skill="council")`, then follow the loaded instructions inline |

**OpenCode skill chaining rules:**
1. Call the `skill` tool to load the target skill's content into context
2. Read and follow the loaded instructions directly — do NOT expect automatic execution
3. **NEVER use slashcommand syntax** (e.g., `/council`) in OpenCode — it triggers a command lookup, not skill loading
4. If the loaded skill references tools by Claude Code names, use OpenCode equivalents (see tool mapping below)

**Cross-runtime tool mapping:**

| Claude Code | OpenCode | Notes |
|-------------|----------|-------|
| `Task(subagent_type="...")` | `task(subagent_type="...")` | Same semantics, different casing |
| `Skill(skill="X")` | `skill` tool (read-only) | Load content, then follow inline |
| `AskUserQuestion` | `question` | Same purpose, different name |
| `TodoWrite` | `todo` | Same purpose, different name |
| `Read`, `Write`, `Edit`, `Bash`, `Glob`, `Grep` | Same names | Identical across runtimes |

### Rules

1. **Never crash** — missing CLI = skip or fallback, not error
2. **Always inform** — tell the user what was skipped and how to enable it
3. **Preserve core function** — the skill's primary purpose must still work without optional CLIs
4. **Progressive enhancement** — CLIs add capabilities, their absence removes them cleanly
