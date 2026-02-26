---
name: shared
description: 'Shared reference documents for multi-agent skills (not directly invocable)'
---


# Shared References

This directory contains shared reference documents used by multiple skills:

- `validation-contract.md` - Verification requirements for accepting spawned work
- `references/claude-code-latest-features.md` - Codex feature contract (slash commands, agent isolation, hooks, settings)
- `references/backend-claude-teams.md` - Portability: Claude native teams (non-Codex runtimes)
- `references/backend-codex-subagents.md` - Concrete examples for Codex CLI and Codex sub-agents
- `references/backend-background-tasks.md` - Fallback: `Task(run_in_background=true)`
- `references/backend-inline.md` - Degraded single-agent mode (no spawn)

These are **not directly invocable skills**. They are loaded by other skills (council, crank, swarm, research, implement) when needed.

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
| `codex` | CLI missing or model unavailable | Fall back to runtime-native agents. Council pre-flight checks CLI presence (`which codex`) and model availability for `--mixed` mode. |
| `cass` | Session search unavailable | Skip transcript search. Note "install cass for session history" |

### Required Multi-Agent Capabilities

Council, swarm, and crank require a runtime that provides these capabilities. If a capability is missing, the corresponding feature degrades.

| Capability | What it does | If missing |
|------------|-------------|------------|
| **Spawn subagent** | Create a parallel agent with a prompt | Cannot run multi-agent. Fall back to `--quick` (inline single-agent). |
| **Agent-to-agent messaging** | Send a message to a specific agent | No debate R2. Workers run fire-and-forget. |
| **Broadcast** | Message all agents at once | Per-agent messaging fallback. |
| **Graceful shutdown** | Request an agent to terminate | Agents terminate on their own when done. |
| **Shared task list** | Agents see shared work state | Lead tracks manually. |

Every runtime maps these capabilities to its own API. Skills describe WHAT to do, not WHICH tool to call.

**After detecting your backend (see Backend Detection below), load the matching reference for concrete tool call examples:**

| Backend | Reference |
|---------|-----------|
| Claude feature contract | `skills/shared/references/claude-code-latest-features.md` |
| Claude Native Teams | `skills/shared/references/backend-claude-teams.md` |
| Codex Sub-Agents / CLI | `skills/shared/references/backend-codex-subagents.md` |
| Background Tasks (fallback) | `skills/shared/references/backend-background-tasks.md` |
| Inline (no spawn) | `skills/shared/references/backend-inline.md` |

### Backend Detection

Use capability detection at runtime, not hardcoded tool names. The same skill must work across any agent harness that provides multi-agent primitives. If no multi-agent capability is detected, degrade to single-agent inline mode (`--quick`).

**Selection policy (Codex-first):**
1. If `spawn_agent` is available → use **Codex sub-agents** as the primary backend.
2. Else if runtime-native team spawning is available → use **runtime-native teams**.
3. If both are technically available, pick the backend native to the current runtime unless the user explicitly requests mixed/cross-vendor execution.
4. Only use background tasks when neither native backend is available.

| Operation | Codex Sub-Agents | Other Runtimes (see Portability Appendix) | Inline Fallback |
|-----------|------------------|-------------------------------------------|-----------------|
| Spawn | `spawn_agent(message=...)` | Runtime-native team spawn | Execute inline |
| Wait | `wait(ids=[...])` | Runtime-native completion signal | N/A |
| Retry/follow-up | `send_input(id=..., message=...)` | Runtime-native messaging | N/A |
| Cleanup | `close_agent(id=...)` | Runtime-native shutdown | N/A |
| Inter-agent messaging | `send_input` | Runtime-native messaging | N/A |
| Debate (R2) | Supported | Varies by runtime | N/A |

**OpenCode limitations:**
- No inter-agent messaging — workers run as independent sub-sessions
- No debate mode (`--debate`) — requires messaging between judges
- `--quick` (inline) mode works identically across all backends

### Backend Capabilities Matrix

> **Prefer native teams over background tasks.** Native teams provide messaging, redirect, and graceful shutdown. Background tasks are fire-and-forget with no steering — only a speedometer and emergency brake.

| Capability | Codex Sub-Agents | Background Tasks |
|------------|------------------|------------------|
| Observe output | `wait()` result | `TaskOutput` (tail) |
| Send message mid-flight | `send_input` | **NO** |
| Pause / resume | NO | **NO** |
| Graceful stop | `close_agent` | **TaskStop (lossy)** |
| Redirect to different task | `send_input` | **NO** |
| Adjust scope mid-flight | `send_input` | **NO** |
| File conflict prevention | Manual `git worktree` routing | None |
| Process isolation | YES (sub-process) | Shared worktree |

**When to use each:**

| Scenario | Backend |
|----------|---------|
| Parallel tasks with coordination | Codex Sub-Agents (preferred) |
| No sub-agent API available (last resort) | Background Tasks |

> See **Portability Appendix** in the inlined references for Claude Native Teams and OpenCode backend details.

### Skill Invocation Across Runtimes

Skills that chain to other skills (e.g., `$rpi` calls `$research`, `$vibe` calls `$council`) MUST handle runtime differences:

| Runtime | Tool | Behavior | Pattern |
|---------|------|----------|---------|
| Codex | `Skill(skill="X", args="...")` | **Executable** — skill runs as a sub-invocation | `Skill(skill="council", args="--quick validate recent")` |
| Codex | N/A | Skills not available — inline the logic or skip | Check if `Skill` tool exists before calling |
| OpenCode | `skill` tool (read-only) | **Load-only** — returns `<skill_content>` blocks into context | Call `skill(skill="council")`, then follow the loaded instructions inline |

**OpenCode skill chaining rules:**
1. Call the `skill` tool to load the target skill's content into context
2. Read and follow the loaded instructions directly — do NOT expect automatic execution
3. **NEVER use slashcommand syntax** (e.g., `$council`) in OpenCode — it triggers a command lookup, not skill loading
4. If the loaded skill references tools by Codex names, use OpenCode equivalents (see tool mapping below)

**Cross-runtime tool mapping:**

| Codex | OpenCode | Notes |
|-------------|----------|-------|
| `Task(subagent_type="...")` | `task(subagent_type="...")` | Same semantics, different casing |
| `Skill(skill="X")` | `skill` tool (read-only) | Load content, then follow inline |
| `AskUserQuestion` | `question` | Same purpose, different name |
| `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet` | `todo` | Task tracking (Claude uses 4 tools, OpenCode uses 1) |
| `Read`, `Write`, `Edit`, `Bash`, `Glob`, `Grep` | Same names | Identical across runtimes |

### Rules

1. **Never crash** — missing CLI = skip or fallback, not error
2. **Always inform** — tell the user what was skipped and how to enable it
3. **Preserve core function** — the skill's primary purpose must still work without optional CLIs
4. **Progressive enhancement** — CLIs add capabilities, their absence removes them cleanly

## Reference Documents

- [references/backend-background-tasks.md](references/backend-background-tasks.md)
- [references/backend-claude-teams.md](references/backend-claude-teams.md)
- [references/backend-codex-subagents.md](references/backend-codex-subagents.md)
- [references/backend-inline.md](references/backend-inline.md)
- [references/claude-code-latest-features.md](references/claude-code-latest-features.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)

---

## References

### backend-background-tasks.md

# Backend: Background Tasks (Fallback)

Concrete tool calls for spawning agents using `Task(run_in_background=true)`. This is the **last-resort fallback** when neither Codex sub-agents nor Claude native teams are available.

**When detected:** `Task` tool is available but `TeamCreate` and `spawn_agent` are not.

**Limitations:**
- Fire-and-forget — no messaging, no redirect, no scope adjustment
- No inter-agent communication
- No debate mode (R2 requires messaging)
- No retry (must re-spawn from scratch)
- No graceful shutdown (only `TaskStop`, which is lossy)

---

## Spawn: Background Agents

Spawn agents with `Task(run_in_background=true)`. Each call returns a `task_id` for later polling.

### Council Judges

```
Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are judge-1.\n\nYour perspective: Correctness & Completeness\n\n<PACKET>\n...\n</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-1.md\nThis is your ONLY output channel — there is no messaging.",
  description="Council judge-1"
)
# Returns: task_id="abc-123"

Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are judge-error-paths.\n\nYour perspective: Error Paths & Edge Cases\n\n<PACKET>...</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-error-paths.md",
  description="Council judge-error-paths"
)
# Returns: task_id="def-456"
```

Both `Task` calls go in the **same message** — they run in parallel.

### Swarm Workers

```
Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are worker-3.\n\nYour Assignment: Task #3: Add password hashing\n...\n\nWrite result to .agents/swarm/results/3.json\nDo NOT run git add/commit/push.",
  description="Swarm worker-3"
)
```

### Research Explorers

```
Task(
  subagent_type="Explore",
  run_in_background=true,
  prompt="Thoroughly investigate: authentication patterns...\n\nWrite findings to .agents/research/2026-02-17-auth.md",
  description="Research explorer"
)
```

---

## Wait: Poll for Completion

Background tasks have no messaging. Poll with `TaskOutput`.

```
TaskOutput(task_id="abc-123", block=true, timeout=120000)
TaskOutput(task_id="def-456", block=true, timeout=120000)
```

**Or non-blocking check:**

```
TaskOutput(task_id="abc-123", block=false, timeout=5000)
```

**After `TaskOutput` returns**, verify the agent wrote its result file:

```
Read(".agents/council/2026-02-17-auth-judge-1.md")
```

**Timeout behavior:** If `timeout` expires, `TaskOutput` returns with a timeout status — the agent may still be running. **Recovery:**
1. Check result file — agent may have written it but not finished cleanly
2. If result file exists → use it, `TaskStop` the agent
3. If no result file → agent failed silently. For council: proceed with N-1 verdicts, note in report. For swarm: add task back to retry queue, re-spawn a fresh agent.
4. Never assume `TaskOutput` completion means the result file was written — always verify

**Fallback:** If background tasks fail despite detection, fall back to inline mode. See `backend-inline.md`.

---

## No Messaging

Background tasks cannot receive messages. This means:

- **No debate R2** — judges get one round only
- **No retry** — if validation fails, re-spawn a new agent from scratch
- **No scope adjustment** — the prompt is final at spawn time

---

## Cleanup

Background tasks self-terminate when done. For stuck tasks:

```
TaskStop(task_id="abc-123")
```

This is lossy — partial work may be lost.

---

## Key Rules

1. **Filesystem is the only communication channel** — agents write files, lead reads files
2. **No messaging = no debate** — `--debate` is unavailable with this backend
3. **No retry = must re-spawn** — failed agents get a fresh `Task` call, not a message
4. **Always check result files** — `TaskOutput` completion doesn't guarantee the agent wrote its file
5. **Prefer native teams** — this backend is strictly inferior; use it only as last resort

### backend-claude-teams.md

# Backend: Claude Native Teams

Concrete tool calls for spawning agents using Codex native teams (`TeamCreate` + `SendMessage` + shared `TaskList`).

**When detected:** `TeamCreate` tool is available in your tool list.

---

## Pre-Flight: Confirm Modern Claude Features

Before spawning teammates, verify feature readiness:

1. `claude agents` succeeds (custom agents discoverable)
2. Teammate profiles for write tasks declare `isolation: worktree`
3. Long-running teammates prefer `background: true`
4. Hooks include worktree lifecycle coverage (`WorktreeCreate`, `WorktreeRemove`) and config auditing (`ConfigChange`) where policy requires it

For canonical feature details, read:
`skills/shared/references/claude-code-latest-features.md`.

---

## Setup: Create Team

Every spawn session starts by creating a team. One team per wave (fresh context = Ralph Wiggum preserved; see `skills/shared/references/ralph-loop-contract.md`).

```
TeamCreate(team_name="council-20260217-auth", description="Council validation of auth module")
```

```
TeamCreate(team_name="swarm-1739812345-w1", description="Wave 1: parallel implementation")
```

**Naming conventions:**
- Council: `council-YYYYMMDD-<target>`
- Swarm: `swarm-<epoch>-w<wave>`
- Crank: delegates to swarm naming

## Leader Contract (Native Teams)

Claude teams are leader-first orchestration:

1. One lead creates the team and assigns all work.
2. Teammates never self-assign from shared tasks.
3. Teammates report to lead via short `SendMessage` signals.
4. Lead reads result artifacts from disk, validates, and decides retries/escalation.

Recommended signal envelope (single-line JSON, under 100 tokens):

```json
{"type":"completion|blocked|help_request","agent":"worker-3","task":"3","detail":"short status","artifact":".agents/swarm/results/3.json"}
```

`completion`: task finished, artifact written.
`blocked`: cannot proceed safely.
`help_request`: teammate needs coordination or scope clarification.

### Peer Messaging (Allowed, Lead-Controlled)

Native teams support direct teammate-to-teammate messaging. Use this only for coordination handoffs; keep messages thin and always copy the lead in follow-up summaries.

```text
worker-2 -> worker-5: "Need auth schema constant name; please confirm from src/auth/schema.ts"
worker-5 -> lead: "Resolved peer question for worker-2; no scope change."
```

---

## Spawn: Create Workers/Judges

After `TeamCreate`, spawn each agent with `Task(team_name=..., name=...)`. All agents in a wave spawn in parallel (single message, multiple tool calls).

### Council Judges (parallel spawn)

```
Task(
  subagent_type="general-purpose",
  team_name="council-20260217-auth",
  name="judge-1",
  prompt="You are judge-1 on team council-20260217-auth.\n\nYour perspective: Correctness & Completeness\n\n<PACKET>\n...\n</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-1.md\nThen send a SHORT completion signal to the team lead (under 100 tokens).\nDo NOT include your full analysis in the message — the lead reads your file.",
  description="Council judge-1"
)

Task(
  subagent_type="general-purpose",
  team_name="council-20260217-auth",
  name="judge-error-paths",
  prompt="You are judge-error-paths on team council-20260217-auth.\n\nYour perspective: Error Paths & Edge Cases\n\n<PACKET>\n...\n</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-error-paths.md\nThen send a SHORT completion signal to the team lead (under 100 tokens).",
  description="Council judge-error-paths"
)
```

Both `Task` calls go in the **same message** — they spawn in parallel.

### Swarm Workers (parallel spawn)

```
Task(
  subagent_type="general-purpose",
  team_name="swarm-1739812345-w1",
  name="worker-3",
  prompt="You are worker-3 on team swarm-1739812345-w1.\n\nYour Assignment: Task #3: Add password hashing\n<description>...</description>\n\nInstructions:\n1. Execute your task — create/edit files as needed\n2. Write result to .agents/swarm/results/3.json\n3. Send a SHORT signal to team lead (under 100 tokens)\n4. Do NOT run git add/commit/push — the lead commits\n\nRESULT FORMAT:\n{\"type\":\"completion\",\"issue_id\":\"3\",\"status\":\"done\",\"detail\":\"one-line summary\",\"artifacts\":[\"path/to/file\"]}",
  description="Swarm worker-3"
)

Task(
  subagent_type="general-purpose",
  team_name="swarm-1739812345-w1",
  name="worker-5",
  prompt="You are worker-5 on team swarm-1739812345-w1.\n\nYour Assignment: Task #5: Create login endpoint\n...",
  description="Swarm worker-5"
)
```

### Research Explorers (read-only)

```
Task(
  subagent_type="Explore",
  team_name="research-20260217-auth",
  name="explorer-1",
  prompt="Thoroughly investigate: authentication patterns in this codebase\n\n...",
  description="Research explorer"
)
```

Use `subagent_type="Explore"` for read-only research agents. Use `"general-purpose"` for agents that need to write files.

---

## Wait: Receive Completion Signals

Workers/judges send completion signals via `SendMessage`. These are **automatically delivered** to the team lead — no polling needed.

When a teammate finishes, their message appears as a new conversation turn. The lead reads result files from disk, NOT from message content.

```
# Teammate message arrives automatically:
# "judge-1: Done. Verdict: WARN, confidence: HIGH. File: .agents/council/2026-02-17-auth-judge-1.md"

# Lead reads the file for full details:
Read(".agents/council/2026-02-17-auth-judge-1.md")
```

**Timeout handling (default: 120s per round, 90s for debate R2):**

If a teammate goes idle without sending a completion signal:
1. Check their result file — they may have written it but failed to message
2. If result file exists → read it and proceed (the message was the only thing missing)
3. If no result file → the agent failed silently. **Recovery:** proceed with N-1 judges/workers and note the failure in the report. For swarm workers, add the task back to the retry queue.
4. Never wait indefinitely — after the timeout, move on

See `skills/council/references/cli-spawning.md` for timeout configuration (`COUNCIL_TIMEOUT`, `COUNCIL_R2_TIMEOUT`).

**Fallback:** If native teams fail at runtime despite passing detection (e.g., `TeamCreate` succeeds but `Task` spawning fails), fall back to background tasks. See `backend-background-tasks.md`.

---

## Message: Debate R2 / Retry

Send messages to specific teammates using `SendMessage`. Teammates wake from idle when messaged.

### Council Debate R2

```
SendMessage(
  type="message",
  recipient="judge-1",
  content="DEBATE ROUND 2\n\nOther judges' verdicts:\n- judge-error-paths: FAIL (HIGH confidence) — file: .agents/council/2026-02-17-auth-judge-error-paths.md\n\nRead the other judge's file. Revise your assessment considering their perspective.\nWrite your R2 verdict to .agents/council/2026-02-17-auth-judge-1-r2.md\nThen send a completion signal.",
  summary="R2 debate instructions for judge-1"
)
```

**R2 timeout (default: 90s):** If a judge doesn't respond to R2 within `COUNCIL_R2_TIMEOUT`, use their R1 verdict for consolidation. See `skills/council/references/debate-protocol.md` for full timeout handling.

### Swarm Worker Retry

```
SendMessage(
  type="message",
  recipient="worker-3",
  content="Validation failed: pytest tests/test_auth.py returned exit code 1.\nFix the failing tests and rewrite your result to .agents/swarm/results/3.json",
  summary="Retry worker-3: test failure"
)
```

---

## Cleanup: Shutdown and Delete

After consolidation/validation, shut down all teammates then delete the team.

```
# Shutdown each teammate
SendMessage(type="shutdown_request", recipient="judge-1", content="Council complete")
SendMessage(type="shutdown_request", recipient="judge-error-paths", content="Council complete")

# After all teammates acknowledge shutdown:
TeamDelete()
```

**Reaper pattern:** If a teammate doesn't respond to shutdown within 30s, proceed with `TeamDelete()` anyway.

**If `TeamDelete` fails** (e.g., stale members): clean up manually with `rm -rf ~/.codex/teams/<team-name>/` then retry `TeamDelete()` to clear in-memory state.

---

## Multi-Wave Pattern

For crank/swarm with multiple waves, create a **new team per wave**:

```
# Wave 1
TeamCreate(team_name="swarm-1739812345-w1", description="Wave 1")
# ... spawn workers, wait, validate, commit ...
# ... shutdown teammates ...
TeamDelete()
# If TeamDelete fails: rm -rf ~/.codex/teams/swarm-1739812345-w1/ then retry

# Wave 2 (fresh context)
TeamCreate(team_name="swarm-1739812345-w2", description="Wave 2")
# ... spawn workers for newly-unblocked tasks ...
TeamDelete()
```

This ensures each wave's workers start with clean context (no leftover state from prior waves).

**If `TeamDelete` fails between waves**, the next `TeamCreate` may conflict. Always verify cleanup succeeded before creating the next wave team.

---

## Key Rules

1. **`TeamCreate` before `Task`** — tasks created before the team are invisible to teammates
2. **Pre-assign tasks before spawning** — workers do NOT race-claim from TaskList
3. **Lead-only commits** — workers write files, lead runs `git add` + `git commit`
4. **Thin messages** — workers send <100 token signals, full results go to disk
5. **New team per wave** — fresh context, Ralph Wiggum preserved
6. **Always cleanup** — `TeamDelete()` after every wave, even on partial failure

### backend-codex-subagents.md

# Backend: Codex Sub-Agents

Concrete tool calls for spawning agents using Codex CLI (`codex exec`). Used for `--mixed` mode cross-vendor consensus and as the primary backend when running inside a Codex session with `spawn_agent`.

---

## Variant A: Codex CLI (from any runtime)

Used when `codex` CLI is available on PATH. Agents run as background shell processes.

**When detected:** `which codex` succeeds.

### Spawn: Background Shell Processes

```bash
# With structured output (preferred for council judges)
Bash(
  command='codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-1.json "JUDGE PROMPT HERE"',
  run_in_background=true
)

# Without structured output (fallback)
Bash(
  command='codex exec --full-auto -m gpt-5.3-codex -C "$(pwd)" -o .agents/council/codex-1.md "JUDGE PROMPT HERE"',
  run_in_background=true
)
```

**Flag order:** `-s`/`--full-auto` → `-m` → `-C` → `--output-schema` → `-o` → prompt

**Valid flags:** `--full-auto`, `-s`, `-m`, `-C`, `--output-schema`, `-o`, `--add-dir`
**Invalid flags:** `-q` (doesn't exist), `--quiet` (doesn't exist)

### Wait: Poll Background Shell

```
TaskOutput(task_id="<shell-id>", block=true, timeout=120000)
```

Then read the output file:

```
Read(".agents/council/codex-1.json")
```

### Limitations

- No messaging — Codex CLI processes are fire-and-forget
- No debate R2 with Codex judges — they produce one verdict only
- `--output-schema` requires `additionalProperties: false` at all levels
- `--output-schema` requires ALL properties in `required` array
- `-s read-only` + `-o` works — `-o` is CLI-level post-processing, not sandbox I/O

---

## Variant B: Codex Sub-Agents (inside Codex runtime)

Used when running inside a Codex session where `spawn_agent` is available.

**When detected:** `spawn_agent` tool is in your tool list.

### Spawn

```
spawn_agent(message="You are judge-1.\n\nPerspective: Correctness & Completeness\n\n<PACKET>...</PACKET>\n\nWrite verdict to .agents/council/2026-02-17-auth-judge-1.md")
# Returns: agent_id

spawn_agent(message="You are worker-3.\n\nTask: Add password hashing\n...\n\nWrite result to .agents/swarm/results/3.json")
# Returns: agent_id
```

### Wait

```
wait(ids=["agent-id-1", "agent-id-2"])
```

**Timeout:** `wait()` blocks until completion. Set a timeout at the orchestration level (default: `COUNCIL_TIMEOUT=120s`). If an agent doesn't complete within the timeout, `close_agent` it and proceed with N-1 verdicts/workers.

### Message (retry/follow-up)

```
send_input(id="agent-id-1", message="Validation failed: fix tests and retry")
```

### Cleanup

```
close_agent(id="agent-id-1")
```

---

## Mixed Mode (Council)

For `--mixed` council, spawn runtime-native judges AND Codex CLI judges in parallel:

```
# Claude native team judges (via TeamCreate — see backend-claude-teams.md)
Task(subagent_type="general-purpose", team_name="council-20260217-auth", name="judge-1", prompt="...", description="Judge 1")
Task(subagent_type="general-purpose", team_name="council-20260217-auth", name="judge-2", prompt="...", description="Judge 2")

# Codex CLI judges (parallel background shells)
Bash(command='codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-1.json "PACKET"', run_in_background=true)
Bash(command='codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-2.json "PACKET"', run_in_background=true)
```

All four spawn in the **same message** — maximum parallelism.

**Mixed mode quorum:** At least 1 judge from each vendor should respond for cross-vendor consensus. If all judges from one vendor fail, proceed as single-vendor council and note the degradation in the report.

---

## Key Rules

1. **Pre-flight check:** `which codex` before attempting Codex CLI spawning
2. **Model availability:** `gpt-5.3-codex` requires API account — fall back to `gpt-4o` if unavailable
3. **Flag order matters** — agents copy examples exactly
4. **`codex review` is a different command** with different flags — do not conflate with `codex exec`
5. **No debate with Codex judges** — they produce one verdict, Codex CLI has no messaging

### backend-inline.md

# Backend: Inline (No Spawn Available)

Degraded single-agent mode when no multi-agent primitives are detected. The current agent performs all work sequentially in its own context.

**When detected:** No `spawn_agent`, no `TeamCreate`, no `Task` tool available — or `--quick` flag was explicitly set.

---

## Council: Single Inline Judge

Instead of spawning parallel judges, the lead evaluates from each perspective sequentially:

```
1. Build the context packet (same as multi-agent mode)
2. For each perspective:
   a. Adopt the perspective mentally
   b. Write findings to .agents/council/YYYY-MM-DD-<target>-<perspective>.md
3. Synthesize into final report
```

Output format is identical — same file paths, same verdict schema. Downstream consumers (consolidation, report) don't know it was inline.

**No debate available** — debate requires messaging between agents.

---

## Swarm: Sequential Execution

Instead of parallel workers, execute each task sequentially:

```
1. TaskList() — find unblocked tasks
2. For each unblocked task (in order):
   a. Execute the task directly
   b. Write result to .agents/swarm/results/<task-id>.json
   c. TaskUpdate(taskId="<id>", status="completed")
3. Check for newly-unblocked tasks
4. Repeat until all tasks complete
```

Same result files, same validation — just sequential.

**Error handling:** If a task fails mid-execution:
1. Write failure result to `.agents/swarm/results/<task-id>.json` with `"status": "blocked"`
2. Check if downstream tasks depend on it (`blockedBy`)
3. Skip blocked downstream tasks, mark as skipped
4. Continue with independent tasks that don't depend on the failed one

---

## Research: Inline Exploration

Instead of spawning an Explore agent, perform the tiered search directly:

```
1. Read docs/code-map/ if present
2. Grep/Glob for relevant files
3. Read key files
4. Write findings to .agents/research/YYYY-MM-DD-<topic>.md
```

---

## Key Rules

1. **Same output format** — inline mode writes the same files as multi-agent mode
2. **Same validation** — all checks still apply
3. **Slower but functional** — no parallelism, but all skill capabilities preserved (except debate)
4. **Inform the user** — log "Running in inline mode (no multi-agent backend detected)"

### claude-code-latest-features.md

# Codex Latest Features Contract

This document is the shared source of truth for Codex feature usage across AgentOps skills.

## Baseline

- Target Codex release family: `2.1.x`
- Last verified against upstream changelog: `2.1.50`
- Changelog source: `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`

## Current Feature Set We Rely On

### 1. Core Slash Commands

Skills and docs should assume these commands exist and prefer them over legacy naming:

- `/agents`
- `/hooks`
- `/permissions`
- `/memory`
- `/mcp`
- `/output-style`

Reference: `https://code.claude.com/docs/en/slash-commands`

### 2. Agent Definitions

For custom teammates in `.claude/agents/*.md`, use modern frontmatter fields where applicable:

- `model`
- `description`
- `tools`
- `memory` (scope control)
- `background: true` for long-running teammates
- `isolation: worktree` for safe parallel write isolation

Reference: `https://code.claude.com/docs/en/sub-agents`

### 3. Worktree Isolation

When parallel workers may touch overlapping files, prefer Claude-native isolation features first:

- Session-level isolation: `claude --worktree` (`-w`)
- Agent-level isolation: `isolation: worktree`

If unavailable in a given runtime, fall back to manual `git worktree` orchestration.

Reference: changelog `2.1.49` and `2.1.50`.

### 4. Hooks and Governance Events

Hooks-based workflows should include modern event coverage:

- `WorktreeCreate`
- `WorktreeRemove`
- `ConfigChange`
- `SubagentStop`
- `TaskCompleted`
- `TeammateIdle`

Use these for auditability, policy enforcement, and cleanup.

Reference: hooks docs and changelog.

### 5. Settings Hierarchy

Skill guidance must respect settings precedence:

1. Enterprise managed policy
2. Command-line args
3. Local project settings
4. Shared project settings
5. User settings

Reference: `https://code.claude.com/docs/en/settings`

### 6. Agent Inventory Command

Use `claude agents` as the first CLI-level check to confirm configured teammate profiles before multi-agent runs.

Reference: changelog `2.1.50`.

## Skill Authoring Rules

1. Do not reference deprecated permission command names (`/allowed-tools`, `/approved-tools`).
2. Multi-agent skills (`council`, `swarm`, `research`, `crank`, `codex-team`) must explicitly point to this contract.
3. Prefer declarative agent isolation (`isolation: worktree`) over ad hoc branch/worktree shell choreography where runtime supports it.
4. Keep manual `git worktree` fallback documented for non-Claude runtimes.
5. For long-running explorers/judges/workers, document `background: true` as the default custom-agent policy.

## Review Cadence

- Re-verify this contract when:
  - Codex changelog introduces new `2.1.x` or `2.2.x` entries
  - any skill adds or changes multi-agent orchestration
  - hook event support changes

### ralph-loop-contract.md

# Ralph Loop Contract (Reverse-Engineered)

This contract captures the operational Ralph mechanics reverse-engineered from:
- `https://github.com/ghuntley/how-to-ralph-wiggum`
- `.tmp/how-to-ralph-wiggum/README.md`
- `.tmp/how-to-ralph-wiggum/files/loop.sh`
- `.tmp/how-to-ralph-wiggum/files/PROMPT_plan.md`
- `.tmp/how-to-ralph-wiggum/files/PROMPT_build.md`

Use this as the source-of-truth for Ralph alignment in AgentOps orchestration skills.

## Core Contract

1. Fresh context every iteration/wave.
- Each execution unit starts clean; no carryover worker memory.

2. Scheduler-heavy, worker-light.
- The lead/orchestrator schedules and reconciles.
- Workers perform one scoped unit of work.

3. Disk-backed shared state.
- Loop continuity comes from filesystem state, not accumulated chat context.
- In classic Ralph: `IMPLEMENTATION_PLAN.md` and `AGENTS.md`.

4. One-task atomicity.
- Select one important task, execute, validate, persist state, then restart fresh.

5. Backpressure before completion.
- Build/tests/lint/gates must reject bad output before task completion/commit.

6. Observe and tune outside the loop.
- Humans (or lead agents) monitor outcomes and adjust prompts/constraints/contracts.

## AgentOps Mapping

| Ralph concept | AgentOps implementation |
|---|---|
| Fresh context per loop | New workers/teams per wave in `$swarm`; fresh phase context in `ao work rpi phased` |
| Main context as scheduler | Mayor/lead orchestration in `$swarm` and `$crank` |
| Plan file as state | `bd` issue graph, TaskList state, plan artifacts in `.agents/plans/` |
| One task per pass | One issue per worker assignment in swarm/crank waves |
| Backpressure | `$vibe`, task validation hooks, tests/lint gates, push/pre-mortem gates |
| Outer loop restart | Wave loop in `$crank`; phase loop in `ao work rpi phased` |

## Implementation Notes

- Keep worker prompts concise and operational.
- Keep state in files/issue trackers, not long conversational memory.
- Prefer deterministic checks over subjective completion.


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0
check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "name is shared" "grep -q '^name: shared' '$SKILL_DIR/SKILL.md'"
check "marked as internal" "grep -q 'internal: true' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


