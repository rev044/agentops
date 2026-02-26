---
name: swarm
description: 'Spawn isolated agents for parallel task execution. Auto-selects runtime-native teams (Claude Native Teams in Claude sessions, Codex sub-agents in Codex sessions). Triggers: "swarm", "spawn agents", "parallel work", "run in parallel", "parallel execution".'
---


# Swarm Skill

Spawn isolated agents to execute tasks in parallel. Fresh context per agent (Ralph Wiggum pattern).

**Integration modes:**
- **Direct** - Create TaskList tasks, invoke `$swarm`
- **Via Crank** - `$crank` creates tasks from beads, invokes `$swarm` for each wave

> **Requires multi-agent runtime.** Swarm needs a runtime that can spawn parallel subagents. If unavailable, work must be done sequentially in the current session.

## Architecture (Mayor-First)

```
Mayor (this session)
    |
    +-> Plan: TaskCreate with dependencies
    |
    +-> Identify wave: tasks with no blockers
    |
    +-> Select spawn backend (runtime-native first: Claude teams in Claude runtime, Codex sub-agents in Codex runtime; fallback tasks if unavailable)
    |
    +-> Assign: TaskUpdate(taskId, owner="worker-<id>", status="in_progress")
    |
    +-> Spawn workers via selected backend
    |       Workers receive pre-assigned task, execute atomically
    |
    +-> Wait for completion (wait() | runtime-native signal | TaskOutput)
    |
    +-> Validate: Review changes when complete
    |
    +-> Cleanup backend resources (close_agent | runtime-native cleanup | none)
    |
    +-> Repeat: New team + new plan if more work needed
```

## Execution

Given `$swarm`:

### Step 0: Detect Multi-Agent Capabilities (MANDATORY)

Use runtime capability detection, not hardcoded tool names. Swarm requires:
- **Spawn parallel subagents** — create workers that run concurrently
- **Agent messaging** (optional) — for coordination and retry

See `skills/shared/SKILL.md` for the capability contract.

**After detecting your backend, read the matching reference for concrete spawn/wait/message/cleanup examples:**
- Claude feature contract → `..$shared/references/claude-code-latest-features.md`
- Claude Native Teams → `..$shared/references/backend-claude-teams.md`
- Codex Sub-Agents / CLI → `..$shared/references/backend-codex-subagents.md`
- Background Tasks → `..$shared/references/backend-background-tasks.md`
- Inline (no spawn) → `..$shared/references/backend-inline.md`

See also `references/local-mode.md` for swarm-specific execution details (worktrees, validation, git commit policy, wave repeat).

### Step 1: Ensure Tasks Exist

Use TaskList to see current tasks. If none, create them:

```
TaskCreate(subject="Implement feature X", description="Full details...",
  metadata={"files": ["src/feature_x.py", "tests/test_feature_x.py"], "validation": {...}})
TaskUpdate(taskId="2", addBlockedBy=["1"])  # Add dependencies after creation
```

#### File Manifest

Every TaskCreate **must** include a `metadata.files` array listing the files that worker is expected to modify. This enables mechanical conflict detection before spawning a wave.

- Pull file lists from the plan, issue description, or codebase exploration during planning.
- If you cannot enumerate files yet, add a planning step to identify them before spawning workers. An empty or missing manifest signals the need for more planning, not unconstrained workers.
- Workers receive the manifest in their prompt and are instructed to stay within it (see `references/local-mode.md` worker prompt template).

```json
{
  "files": ["cli/cmd/ao/goals.go", "cli/cmd/ao/goals_test.go"],
  "validation": {
    "tests": "go test ./cli/cmd/ao/...",
    "files_exist": ["cli/cmd/ao/goals.go"]
  }
}
```

### Step 1a: Build Context Briefing (Before Worker Dispatch)

```bash
if command -v ao &>/dev/null; then
    ao work context assemble --task='<swarm objective or wave description>'
fi
```

This produces a 5-section briefing (GOALS, HISTORY, INTEL, TASK, PROTOCOL) at `.agents/rpi/briefing-current.md` with secrets redacted. Include the briefing path in each worker's TaskCreate description so workers start with full project context.

### Step 2: Identify Wave

Find tasks that are:
- Status: `pending`
- No blockedBy (or all blockers completed)

These can run in parallel.

#### Pre-Spawn Conflict Check

Before spawning a wave, scan all worker file manifests for overlapping files:

```
wave_tasks = [tasks with status=pending and no blockers]
all_files = {}
for task in wave_tasks:
    for f in task.metadata.files:
        if f in all_files:
            CONFLICT: f is claimed by both all_files[f] and task.id
        all_files[f] = task.id
```

**On conflict detection:**
- **Serialize** the conflicting workers into separate sub-waves (preferred -- simplest fix), OR
- **Isolate** them with worktree isolation (`--worktrees`) so each operates on a separate branch.

Do not spawn workers with overlapping file manifests into the same shared-worktree wave. This is the primary cause of build breaks and merge conflicts in parallel execution.

### Steps 3-6: Spawn Workers, Validate, Finalize

**For detailed local mode execution (team creation, worker spawning, race condition prevention, git commit policy, validation contract, cleanup, and repeat logic), read `skills/swarm/references/local-mode.md`.**

> **Platform pitfalls:** Include relevant pitfalls from `references/worker-pitfalls.md` in worker prompts for the target language/platform. For example, inject the Bash section for shell script tasks, the Go section for Go tasks, etc. This prevents common worker failures from known platform gotchas.

## Example Flow

```
Mayor: "Let's build a user auth system"

1. $plan -> Creates tasks:
   #1 [pending] Create User model
   #2 [pending] Add password hashing (blockedBy: #1)
   #3 [pending] Create login endpoint (blockedBy: #1)
   #4 [pending] Add JWT tokens (blockedBy: #3)
   #5 [pending] Write tests (blockedBy: #2, #3, #4)

2. $swarm -> Spawns agent for #1 (only unblocked task)

3. Agent #1 completes -> #1 now completed
   -> #2 and #3 become unblocked

4. $swarm -> Spawns agents for #2 and #3 in parallel

5. Continue until #5 completes

6. $vibe -> Validate everything
```

## Key Points

- **Runtime-native local mode** - Auto-selects the native backend for the current runtime (Claude teams or Codex sub-agents)
- **Universal orchestration contract** - Same swarm behavior across Claude and Codex sessions
- **Pre-assigned tasks** - Mayor assigns tasks before spawning; workers never race-claim
- **Fresh worker contexts** - New sub-agents/teammates per wave preserve Ralph isolation
- **Wave execution** - Only unblocked tasks spawn
- **Mayor orchestrates** - You control the flow, workers write results to disk
- **Thin results** - Workers write `.agents/swarm/results/<id>.json`, orchestrator reads files (NOT Task returns or messaging content)
- **Retry via message/input** - Use `send_input` (Codex sub-agents) or runtime-native messaging for coordination only
- **Atomic execution** - Each worker works until task done
- **Graceful degradation** - If multi-agent unavailable, work executes sequentially in current session

## Workflow Integration

This ties into the full workflow:

```
$research -> Understand the problem
$plan -> Decompose into beads issues
$crank -> Autonomous epic loop
    +-- $swarm -> Execute each wave in parallel
$vibe -> Validate results
$post-mortem -> Extract learnings
```

**Direct use (no beads):**
```
TaskCreate -> Define tasks
$swarm -> Execute in parallel
```

The knowledge flywheel captures learnings from each agent.

## Task Management Commands

```
# List all tasks
TaskList()

# Mark task complete after notification
TaskUpdate(taskId="1", status="completed")

# Add dependency between tasks
TaskUpdate(taskId="2", addBlockedBy=["1"])
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--max-workers=N` | Max concurrent workers | 5 |
| `--from-wave <json-file>` | Load wave from OL hero hunt output (see OL Wave Integration) | - |
| `--per-task-commits` | Commit per task instead of per wave (for attribution/audit) | Off (per-wave) |

## When to Use Swarm

| Scenario | Use |
|----------|-----|
| Multiple independent tasks | `$swarm` (parallel) |
| Sequential dependencies | `$swarm` with blockedBy |
| Mix of both | `$swarm` spawns waves, each wave parallel |

## Why This Works: Ralph Wiggum Pattern

Follows the [Ralph Wiggum Pattern](https://ghuntley.com/ralph/): **fresh context per execution unit**.

- **Wave-scoped worker set** = spawn workers -> execute -> cleanup -> repeat (fresh context each wave)
- **Mayor IS the loop** - Orchestration layer, manages state across waves
- **Workers are atomic** - One task, one spawn, one result
- **TaskList as memory** - State persists in task status, not agent context
- **Filesystem for EVERYTHING** - Code artifacts AND result status written to disk, not passed through context
- **Backend messaging for signals only** - Short coordination signals (under 100 tokens), never work details

Ralph alignment source: `..$shared/references/ralph-loop-contract.md`.

## Integration with Crank

When `$crank` invokes `$swarm`: Crank bridges beads to TaskList, swarm executes with fresh-context agents, crank syncs results back.

| You Want | Use | Why |
|----------|-----|-----|
| Fresh-context parallel execution | `$swarm` | Each spawned agent is a clean slate |
| Autonomous epic loop | `$crank` | Loops waves via swarm until epic closes |
| Just swarm, no beads | `$swarm` directly | TaskList only, skip beads |
| RPI progress gates | `$ratchet` | Tracks progress; does not execute work |

---

## OL Wave Integration

When `$swarm --from-wave <json-file>` is invoked, the swarm reads wave data from an OL hero hunt output file and executes it with completion backflow to OL.

### Pre-flight

```bash
# --from-wave requires ol CLI on PATH
which ol >/dev/null 2>&1 || {
    echo "Error: ol CLI required for --from-wave. Install ol or use swarm without wave integration."
    exit 1
}
```

If `ol` is not on PATH, exit immediately with the error above. Do not fall back to normal swarm mode.

### Input Format

The `--from-wave` JSON file contains `ol hero hunt` output:

```json
{
  "wave": [
    {"id": "ol-527.1", "title": "Add auth middleware", "spec_path": "quests/ol-527/specs/ol-527.1.md", "priority": 1},
    {"id": "ol-527.2", "title": "Fix rate limiting", "spec_path": "quests/ol-527/specs/ol-527.2.md", "priority": 2}
  ],
  "blocked": [
    {"id": "ol-527.3", "title": "Integration tests", "blocked_by": ["ol-527.1", "ol-527.2"]}
  ],
  "completed": [
    {"id": "ol-527.0", "title": "Project setup"}
  ]
}
```

### Execution

1. **Parse the JSON file** and extract the `wave` array.

2. **Create TaskList tasks** from wave entries (one `TaskCreate` per entry):

```
for each entry in wave:
    TaskCreate(
        subject="[{entry.id}] {entry.title}",
        description="OL bead {entry.id}\nSpec: {entry.spec_path}\nPriority: {entry.priority}\n\nRead the spec file at {entry.spec_path} for full requirements.",
        metadata={
            "ol_bead_id": entry.id,
            "ol_spec_path": entry.spec_path,
            "ol_priority": entry.priority
        }
    )
```

3. **Execute swarm normally** on those tasks (Step 2 onward from main execution flow). Tasks are ordered by priority (lower number = higher priority).

4. **Completion backflow**: After each worker completes a bead task AND passes validation, the team lead runs the OL ratchet command to report completion back to OL:

```bash
# Extract quest ID from bead ID (e.g., ol-527.1 -> ol-527)
QUEST_ID=$(echo "$BEAD_ID" | sed 's/\.[^.]*$//')

ol hero ratchet "$BEAD_ID" --quest "$QUEST_ID"
```

**Ratchet result handling:**

| Exit Code | Meaning | Action |
|-----------|---------|--------|
| 0 | Bead complete in OL | Mark task completed, log success |
| 1 | Ratchet validation failed | Mark task as failed, log the validation error from stderr |

5. **After all wave tasks complete**, report a summary that includes both swarm results and OL ratchet status for each bead.

### Example

```
$swarm --from-wave /tmp/wave-ol-527.json

# Reads wave JSON -> creates 2 tasks from wave entries
# Spawns workers for ol-527.1 and ol-527.2
# On completion of ol-527.1:
#   ol hero ratchet ol-527.1 --quest ol-527 -> exit 0 -> bead complete
# On completion of ol-527.2:
#   ol hero ratchet ol-527.2 --quest ol-527 -> exit 0 -> bead complete
# Wave done: 2/2 beads ratcheted in OL
```

---

## References

- **Local Mode Details:** `skills/swarm/references/local-mode.md`
- **Validation Contract:** `skills/swarm/references/validation-contract.md`

---

## Examples

### Building a User Auth System

**User says:** `$swarm`

**What happens:**
1. Agent identifies unblocked tasks from TaskList (e.g., "Create User model")
2. Agent selects spawn backend using runtime-native priority (Claude session -> Claude teams; Codex session -> Codex sub-agents)
3. Agent spawns worker for task #1, assigns ownership via TaskUpdate
4. Worker completes, team lead validates changes
5. Agent identifies next wave (tasks #2 and #3 now unblocked)
6. Agent spawns two workers in parallel for Wave 2

**Result:** Multi-wave execution with fresh-context workers per wave, zero race conditions.

### Direct Swarm Without Beads

**User says:** Create three tasks for API refactor, then `$swarm`

**What happens:**
1. User creates TaskList tasks with TaskCreate
2. Agent calls `$swarm` without beads integration
3. Agent identifies parallel tasks (no dependencies)
4. Agent spawns all three workers simultaneously
5. Workers execute atomically, report to team lead via backend messaging or task completion
6. Team lead validates all changes, commits once per wave

**Result:** Parallel execution of independent tasks using TaskList only.

### Loading Wave from OL

**User says:** `$swarm --from-wave /tmp/wave-ol-527.json`

**What happens:**
1. Agent validates `ol` CLI is on PATH (pre-flight check)
2. Agent reads wave JSON from OL hero hunt output
3. Agent creates TaskList tasks from wave entries (priority-sorted)
4. Agent spawns workers for all unblocked beads
5. On completion, agent runs `ol hero ratchet <bead-id> --quest <quest-id>` for each bead
6. Agent reports backflow status to user

**Result:** OL beads executed with completion reporting back to Olympus.

---

## Worktree Isolation (Multi-Epic Dispatch)

**Default behavior:** Auto-detect and prefer runtime-native isolation first.

In Claude runtime, first verify teammate profiles with `claude agents` and use agent definitions with `isolation: worktree` for write-heavy parallel waves. If native isolation is unavailable, use manual `git worktree` fallback below.

### Isolation Semantics Per Spawn Backend

| Backend | Isolation Mechanism | How It Works |
|---------|-------------------|--------------|
| **Claude teams** (`Task` with `team_name`) | `isolation: worktree` in agent definition | Runtime creates an isolated git worktree per teammate; changes are invisible to other agents and the main tree until merged |
| **Background tasks** (`Task` with `run_in_background`) | `isolation: worktree` in agent definition | Same worktree isolation as teams; each background agent gets its own worktree |
| **Inline** (no spawn) | None | Operates directly on the main working tree; no isolation possible |

**Key diagnostic:** When `isolation: worktree` is specified but worker changes appear in the main working tree (no separate worktree path in the Task result), **isolation did NOT engage**. This is a silent failure — the runtime accepted the parameter but did not create a worktree.

### Post-Spawn Isolation Verification

After spawning workers with `isolation: worktree`, the lead MUST verify isolation engaged:

1. **Check Task result** for a `worktreePath` field. If present, isolation is active.
2. **If `worktreePath` is absent** but `isolation: worktree` was specified:
   - Log warning: "Isolation did not engage for worker-N. Changes may be in main working tree."
   - **For waves with 2+ workers touching overlapping files:** abort the wave, fall back to serial execution to prevent conflicts.
   - **For waves with fully independent file sets:** may proceed with caution, but monitor for conflicts.
3. **If isolation consistently fails:** fall back to manual `git worktree` creation (see below) or switch to serial inline execution.

**When to use worktrees:** Activate worktree isolation when:
- Dispatching workers across **multiple epics** (each epic touches different packages)
- Wave has **>3 workers touching overlapping files** (detected via `git diff --name-only`)
- Tasks span **independent branches** that shouldn't cross-contaminate

Evidence: 4 parallel agents in shared worktree produced 1 build break and 1 algorithm duplication (see `.agents/evolve/dispatch-comparison.md`). Worktree isolation prevents collisions by construction.

### Detection: Do I Need Worktrees?

```bash
# Heuristic: multi-epic = worktrees needed
# Single epic with independent files = shared worktree OK

# Check if tasks span multiple epics
# e.g., task subjects contain different epic IDs (ol-527, ol-531, ...)
# If yes: use worktrees
# If no: proceed with default shared worktree
```

### Creation: One Worktree Per Epic

Before spawning workers, create an isolated worktree per epic:

```bash
# For each epic ID in the wave:
git worktree add /tmp/swarm-<epic-id> -b swarm/<epic-id>
```

Example for 3 epics:
```bash
git worktree add /tmp/swarm-ol-527 -b swarm/ol-527
git worktree add /tmp/swarm-ol-531 -b swarm/ol-531
git worktree add /tmp/swarm-ol-535 -b swarm/ol-535
```

Each worktree starts at HEAD of current branch. The worker branch (`swarm/<epic-id>`) is ephemeral — deleted after merge.

### Worker Routing: Inject Worktree Path

Pass the worktree path as the working directory in each worker prompt:

```
WORKING DIRECTORY: /tmp/swarm-<epic-id>

All file reads, writes, and edits MUST use paths rooted at /tmp/swarm-<epic-id>.
Do NOT operate on /path/to/main/repo directly.
```

Workers run in isolation — changes in one worktree cannot conflict with another.

**Result file path:** Workers still write results to the main repo's `.agents/swarm/results/`:
```bash
# Worker writes to main repo result path (not the worktree)
RESULT_DIR=/path/to/main/repo/.agents/swarm/results
```

The orchestrator path for `.agents/swarm/results/` is always the main repo, not the worktree.

### Merge-Back: After Validation

After a worker's task passes validation, merge the worktree branch back to main:

```bash
# From the main repo (not worktree)
git merge --no-ff swarm/<epic-id> -m "chore: merge swarm/<epic-id> (epic <epic-id>)"
```

Merge order: respect task dependencies. If epic B blocked by epic A, merge A before B.

**On merge conflict:** The team lead resolves conflicts manually. Workers must not merge — lead-only commit policy still applies.

### Cleanup: Remove Worktrees After Merge

```bash
# After successful merge:
git worktree remove /tmp/swarm-<epic-id>
git branch -d swarm/<epic-id>
```

Run cleanup even on partial failures (same reaper pattern as team cleanup).

### Full Pre-Spawn Sequence (Worktree Mode)

```
1. Detect: does this wave need worktrees? (multi-epic or file overlap)
2. For each epic:
   a. git worktree add /tmp/swarm-<epic-id> -b swarm/<epic-id>
3. Spawn workers with worktree path injected into prompt
4. Wait for completion (same as shared mode)
5. Validate each worker's changes (run tests inside worktree)
6. For each passing epic:
   a. git merge --no-ff swarm/<epic-id>
   b. git worktree remove /tmp/swarm-<epic-id>
   c. git branch -d swarm/<epic-id>
7. Commit all merged changes (team lead, sole committer)
```

### Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `--worktrees` | Force worktree isolation for this wave | Off (auto-detect) |
| `--no-worktrees` | Force shared worktree even for multi-epic | Off |

---

## Troubleshooting

### Worktree isolation did not engage
Cause: `isolation: worktree` was specified but the Task result has no `worktreePath` — worker changes land in the main tree.
Solution: Verify agent definitions include `isolation: worktree`. If the runtime does not support declarative isolation, fall back to manual `git worktree add` (see Worktree Isolation section). For overlapping-file waves, abort and switch to serial execution.

### Workers produce file conflicts
Cause: Multiple workers editing the same file in parallel.
Solution: Use worktree isolation (`--worktrees`) for multi-epic dispatch. For single-epic waves, use wave decomposition to group workers by file scope. Homogeneous waves (all Go, all docs) prevent conflicts.

### Team creation fails
Cause: Stale team from prior session not cleaned up.
Solution: Run `rm -rf ~/.codex/teams/<team-name>` then retry.

### Codex agents unavailable
Cause: `codex` CLI not installed or API key not configured.
Solution: Run `which codex` to verify installation. Check `~/.codex/config.toml` for API credentials.

### Workers timeout or hang
Cause: Worker task too large or blocked on external dependency.
Solution: Break tasks into smaller units. Add timeout metadata to worker tasks.

### OL wave integration fails with "ol CLI required"
Cause: `--from-wave` used but `ol` CLI not on PATH.
Solution: Install Olympus CLI or run swarm without `--from-wave` flag.

### Tasks assigned but workers never spawn
Cause: Backend selection failed or spawning API unavailable.
Solution: Check which spawn backend was selected (look for "Using: <backend>" message). Verify Codex CLI (`which codex`) or native team API availability.

## Reference Documents

- [references/backend-background-tasks.md](references/backend-background-tasks.md)
- [references/backend-claude-teams.md](references/backend-claude-teams.md)
- [references/backend-codex-subagents.md](references/backend-codex-subagents.md)
- [references/backend-inline.md](references/backend-inline.md)
- [references/claude-code-latest-features.md](references/claude-code-latest-features.md)
- [references/local-mode.md](references/local-mode.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)
- [references/validation-contract.md](references/validation-contract.md)
- [references/worker-pitfalls.md](references/worker-pitfalls.md)

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

### local-mode.md

# Swarm Local Mode: Runtime-Aware Detailed Execution

## Context Budget Rule

> **Workers write results to disk. The orchestrator reads only thin status files.**
>
> When N workers finish, their full output (file reads, tool calls, reasoning) must NOT flood back into the orchestrator context. This is the #1 cause of context explosion in multi-wave epics.

**Result protocol:**
1. Workers write `.agents/swarm/results/<task-id>.json` on completion
2. Orchestrator checks for result files (Glob/Read), NOT full Task/SendMessage output
3. SendMessage used only for coordination signals (blocked, need help) — kept under 100 tokens
4. Task tool return values are acknowledged but NOT parsed for work details

```bash
# Orchestrator creates result directory before spawning
mkdir -p .agents/swarm/results
```

## Step 2b: Pre-Spawn Worktree Setup (Multi-Epic Waves)

> **Skip this step** for single-epic waves or when `--no-worktrees` is set.
> **Required** for multi-epic dispatch or when `--worktrees` is set.

Evidence: shared-worktree multi-epic dispatch produced build breaks and algorithm duplication (`.agents/evolve/dispatch-comparison.md`).

### Claude-Native Isolation (preferred when available)

If running in Claude runtime with modern agent definitions, prefer declarative isolation first:

1. Confirm teammate profiles with `claude agents`
2. Use teammate definitions that set `isolation: worktree`
3. For long-running workers, set `background: true`

Only fall back to manual `git worktree` management when declarative isolation is unavailable.

### Detection

```bash
# Multi-epic: check if tasks span more than one epic prefix
# If wave tasks have subjects like "[ol-527] ..." and "[ol-531] ...", use worktrees.
# Single-epic: tasks share one prefix (e.g., all ol-527.*) → shared worktree OK.
```

### Create Worktrees

```bash
# For each epic ID in the wave:
git worktree add /tmp/swarm-<epic-id> -b swarm/<epic-id>
```

Track the mapping:
```
epic_worktrees = {
  "<epic-id>": "/tmp/swarm-<epic-id>",
  ...
}
```

### Inject into Worker Prompts

Each worker prompt must include:
```
WORKING DIRECTORY: /tmp/swarm-<epic-id>

All file reads, writes, and edits MUST use absolute paths rooted at /tmp/swarm-<epic-id>.
Do NOT operate on the main repo directly.
Result file: write to <main-repo>/.agents/swarm/results/<task-id>.json (always main repo path).
```

### Merge-Back After Validation

After each worker's task passes validation:
```bash
# From main repo:
git merge --no-ff swarm/<epic-id> -m "chore: merge swarm/<epic-id>"
git worktree remove /tmp/swarm-<epic-id>
git branch -d swarm/<epic-id>
```

Merge order must respect task blockedBy dependencies.

---

## Step 3: Spawn Workers

Use whatever multi-agent primitives your runtime provides to spawn parallel workers. Each worker receives a pre-assigned task in its prompt.

### Model Selection

Workers should use **sonnet** (not opus) to minimize cost. The orchestrator (lead) stays on opus for coordination and validation.

When spawning via the Task tool, pass `model: "sonnet"`. When spawning via native teams, teammates inherit from the session model unless overridden — set `COUNCIL_CLAUDE_MODEL=sonnet` or use `model: "sonnet"` in the Task call. For longer tasks, prefer teammate profiles with `background: true`.

| Role | Model | Rationale |
|------|-------|-----------|
| Lead/orchestrator | opus (session default) | Coordination, validation, state management |
| Workers | sonnet | Focused single-task execution, 3-5x cheaper |
| Explorers | sonnet | Read-only search tasks |

### Spawn Protocol

For each ready task:

1. **Pre-assign** — Mark the task owned by `worker-<task-id>` before spawning (prevents race conditions)
2. **Spawn** — Create a parallel subagent with the worker prompt (see below)
3. **Track** — Map `worker-<task-id>` to agent handle for waits/retries/cleanup

All workers in a wave spawn in parallel. New team/agent-group per wave = fresh context (Ralph Wiggum preserved).

### Worker Prompt Template

Every worker receives this prompt (adapt to your runtime's spawn mechanism):

```
You are worker-<task-id>.

Your Assignment: Task #<id>: <subject>
<description>

FILE MANIFEST (files you are permitted to modify):
<list of files from plan — one per line>

You MUST NOT modify files outside this manifest. If you need to read other files for context, that is fine.
If your task requires modifying a file not in this manifest, write a blocked result instead.

Instructions:
1. Execute your pre-assigned task independently — create/edit files as needed, verify your work
2. Write your result to .agents/swarm/results/<task-id>.json (see format below)
3. Send a SHORT completion signal to the lead (under 100 tokens)
4. If blocked, write blocked result to same path and signal the lead

RESULT FILE FORMAT (MANDATORY — write this BEFORE sending any signal):

On success:
{"type":"completion","issue_id":"<task-id>","status":"done","detail":"<one-line summary max 100 chars>","artifacts":["path/to/file1","path/to/file2"]}

If blocked:
{"type":"blocked","issue_id":"<task-id>","status":"blocked","detail":"<reason max 200 chars>"}

CONTEXT BUDGET RULE:
Your message to the lead must be under 100 tokens.
Do NOT include file contents, diffs, or detailed explanations in messages.
The result JSON file IS your full report. The lead reads the file, not your message.

Rules:
- Work only on YOUR pre-assigned task
- Do NOT claim other tasks
- Do NOT message other workers
- Do NOT run git add, git commit, or git push — the lead commits
```

> **Orchestrator note — populating the FILE MANIFEST:** When building each worker prompt, replace
> `<list of files from plan — one per line>` with the explicit file paths assigned to that task in
> your plan. Pull these from the task's `metadata.files` field if present, or derive them from the
> task description during planning. Example populated manifest:
>
> ```
> FILE MANIFEST (files you are permitted to modify):
> src/middleware/auth.py
> tests/test_auth.py
> ```
>
> If the plan does not yet enumerate files, add a planning step to identify them before spawning
> workers. An empty or missing manifest is a signal to pause and plan further — not to let workers
> operate unconstrained.

## Race Condition Prevention

Workers do NOT race-claim tasks from TaskList. The team lead assigns each task
to a specific worker BEFORE spawning. This prevents:
- Two workers claiming the same task
- Workers seeing stale TaskList state
- Non-deterministic assignment order

Workers only transition their assigned task: in_progress -> completed.

## Git Commit Policy

**Workers MUST NOT commit.** The team lead is the sole committer.

| Actor | Git Permissions |
|-------|----------------|
| Team lead (mayor) | git add, commit, push |
| Workers | Read-only git. Write files only. |

**Rationale:**
- Workers share a worktree (per native teams semantics research)
- Concurrent git add/commit from multiple workers corrupts the index
- Lead-only commits ensure atomic, reviewable changesets per wave

**Worker instructions:** Include in every worker prompt:
"Do NOT run git add, git commit, or git push. Write your result to .agents/swarm/results/<task-id>.json, then send a short signal (under 100 tokens) via your runtime channel. The team lead reads result files, not messages."

### Wave Commit Cadence

**Best practice: one commit per completed wave** (not one massive commit for the entire swarm run).

| Cadence | When to use | Commit message format |
|---------|-------------|----------------------|
| **Per wave** (default) | Standard swarm execution | `chore(wave-N): close ag-xxxx, ag-yyyy` |
| **Per task** (`--per-task-commits`) | When per-task attribution is required (audits, blame tracking) | `chore(ag-xxxx): <task subject>` |
| **End of swarm** | Never recommended — loses wave attribution and makes rollback harder | - |

**Why per-wave:** Each wave is an atomic unit of parallel work. A single commit per wave provides:
- Clean rollback boundary (revert one wave without touching others)
- Clear attribution of which wave introduced a change
- Issue IDs in commit message for traceability

**Commit message convention:**
```
chore(wave-1): close ag-1234, ag-1235

- ag-1234: Add authentication middleware
- ag-1235: Create user model schema
```

The lead commits after all tasks in a wave pass validation (Step 4a), before spawning the next wave.

## Step 4: Wait for Completion

Wait for all workers to signal completion using your runtime's wait mechanism. Workers write result files to disk and send a minimal signal.

**Result data** — always read from disk, never from agent messages:

Check `.agents/swarm/results/<task-id>.json` for each worker. These are ~200 bytes each. Do NOT parse agent messages or return values for work details — they contain the worker's full conversation (5-20K tokens per worker) and will explode the lead's context.

**CRITICAL**: Do NOT mark complete yet — validation required first.

## Step 4a: Validate Before Accepting (MANDATORY)

> **TRUST ISSUE**: Agent completion claims are NOT trusted. Verify then trust.

**The Validation Contract**: Before marking any task complete, Mayor MUST run validation checks. See `skills/shared/validation-contract.md` for full specification.

**Validation flow:**

```
<task-notification> arrives
        |
        v
    RUN VALIDATION
        |
    +---+---+
    |       |
  PASS    FAIL
    |       |
    v       v
 complete  retry/escalate
```

**For each completed task notification:**

1. **Check task metadata for validation requirements:**
   ```
   TaskList() -> find task -> check metadata.validation
   ```

2. **Execute validation checks (in order):**

   | Check Type | Command | Pass Condition |
   |------------|---------|----------------|
   | `files_exist` | `ls -la <paths>` | All files exist |
   | `command` | Run specified command | Exit code 0 |
   | `content_check` | `grep <pattern> <file>` | Pattern found |
   | `tests` | `<test_command>` | Tests pass |
   | `lint` | `<lint_command>` | No errors |

3. **On validation PASS:**
   ```
   TaskUpdate(taskId="<id>", status="completed")
   ```

4. **On validation FAIL:**
   - Increment retry count for task
   - If retries < MAX_RETRIES (default: 3): send a follow-up message to the worker via your runtime's messaging mechanism: "Validation failed: <specific failure>. Fix and retry."
   - If retries >= MAX_RETRIES: mark task as blocked and escalate to user

**Minimal validation (when no metadata):**

If task has no explicit validation requirements, apply default checks:

```bash
# Check that worker wrote the expected files
git status --porcelain  # Should show unstaged changes from worker

# Check for obvious failures in recent output
# (agent should not have ended with errors)
```

**Example task with validation metadata:**

```
TaskCreate(
  subject="Add authentication middleware",
  description="...",
  metadata={
    "validation": {
      "files_exist": ["src/middleware/auth.py", "tests/test_auth.py"],
      "command": "pytest tests/test_auth.py -v",
      "content_check": {"file": "src/middleware/auth.py", "pattern": "def authenticate"}
    }
  }
)
```

## Step 5: Review & Finalize

When workers complete AND pass validation:
1. Check git status for changes (workers wrote files but did not commit)
2. Review diffs
3. Run any additional tests/validation
4. Team lead commits all changes for the wave (sole committer)

## Step 5a: Cleanup

After wave completes:

```bash
# Clean up result files from this wave (prevent stale reads in next wave)
rm -f .agents/swarm/results/*.json
```

Shut down all workers via your runtime's cleanup mechanism. Then clean up the agent group/team.

### Reaper Cleanup Pattern

Cleanup MUST succeed even on partial failures:

1. Request graceful shutdown for each worker
2. Wait up to 30s for acknowledgment
3. If any worker doesn't respond, log warning, proceed anyway
4. Always run cleanup — lingering agents pollute future sessions

### Timeout Configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| Worker timeout | 180s | Max time for worker to complete its task |
| Shutdown grace period | 30s | Time to wait for shutdown acknowledgment |
| Wave timeout | 600s | Max time for entire wave before forced cleanup |

## Step 6: Repeat if Needed

If more tasks remain:
1. Check TaskList for next wave
2. Spawn a NEW wave worker set (new sub-agents or new team) for fresh context
3. Execute the next wave
4. Continue until all done

## Partial Completion

**Worker timeout:** 180s (3 minutes) default per worker.

**Timeout behavior:**
1. Log warning: "Worker {name} timed out on task {id}"
2. Mark task as failed with reason "timeout"
3. Add to retry queue for next wave
4. Continue with remaining workers

**Quorum:** Swarm does not require quorum -- each worker is independent.
Each completed task is accepted individually.

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

### validation-contract.md

# Validation Contract (Moved)

Source of truth: `skills/shared/validation-contract.md`

This file is a compatibility shim for older links and references.

### worker-pitfalls.md

# Worker Pitfalls: Platform-Specific Gotchas

Inject relevant sections into worker prompts based on the task's target language/platform.

---

## Bash

**Subshell variable scoping** -- Variables set inside a pipe subshell do not propagate.
```bash
# BROKEN: count stays 0 (while runs in subshell)
count=0; cat file.txt | while read line; do count=$((count+1)); done
# FIX: redirect instead of pipe
while read line; do count=$((count+1)); done < file.txt
```

**macOS vs GNU tools** -- BSD sed/awk/head flags differ from GNU.
```bash
# BROKEN on macOS:
sed -i 's/old/new/' file.txt
# FIX (macOS): sed -i '' 's/old/new/' file.txt
```

**rm alias hangs workers** -- Some systems alias `rm` to `rm -i`, blocking on confirmation.
```bash
# FIX: bypass aliases
/bin/rm -f somefile
```

**Silent pipe failures** -- Pipeline exit code is the last command's. Earlier failures are hidden.
```bash
# FIX: enable pipefail at top of script
set -o pipefail
```

**Unquoted variables** -- Word splitting breaks paths with spaces.
```bash
# BROKEN: cat "my" and "report.txt" separately
file="my report.txt"; cat $file
# FIX: always double-quote: cat "$file"
```

---

## Go

**Build tag placement** -- `//go:build` must be first line, blank line before `package`.
```go
// BROKEN:
package main
//go:build linux
// FIX:
//go:build linux

package main
```

**Module path vs imports** -- Import paths must match the module path in go.mod exactly.
```
go.mod: module github.com/user/repo
BROKEN: import "github.com/user/repo/v2/pkg"  (module is not v2)
FIX:    import "github.com/user/repo/pkg"
```

**Test naming** -- Files must end `_test.go`. Functions must be `TestXxx` (capital after Test).
```
BROKEN: auth_tests.go, func testAuth(t *testing.T)
FIX:    auth_test.go,  func TestAuth(t *testing.T)
```

**Unused imports fail build** -- Go refuses to compile with unused imports.
```go
// FIX: remove unused imports, or blank-import for side effects:
import _ "github.com/lib/pq"
```

**Unused variables fail build** -- Declared-but-unused locals are a compile error.
```go
// BROKEN: result declared, only err used
result, err := doSomething()
// FIX: blank identifier
_, err := doSomething()
```

---

## Git

**Worktree isolation** -- Changes in a worktree are invisible to main tree until merged. Workers in `/tmp/swarm-epic-1/` do not affect `/repo/`.

**Detached HEAD** -- Worktrees created without `-b` start detached; commits may be lost.
```bash
# BROKEN: git worktree add /tmp/task1 HEAD
# FIX:    git worktree add /tmp/task1 -b swarm/task1
```

**Never commit from a worker** -- Concurrent `git add`/`git commit` corrupts the index. Workers write files only. The team lead is the sole committer.

---

## Skills / Docs

**Source of truth** -- Edit skills in `skills/` in this repo, NOT `~/.codex/skills/` (installed copies are overwritten on update).

**Reference linkage** -- Every file under `skills/<name>/references/` must be linked from that skill's SKILL.md. `heal.sh --strict` enforces this.

**No symlinks** -- The plugin-load-test rejects symlinks. Copy files instead of symlinking.

**Skill count sync** -- Adding or removing a skill directory requires `scripts/sync-skill-counts.sh`. CI fails otherwise.


---

## Scripts

### ol-ratchet.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# OL Ratchet Backflow - Feedback path from AO back to OL
# Called per-bead after worker validation passes.
# Notifies Olympus that a bead completed successfully.

BEAD_ID="${1:?Usage: ol-ratchet.sh <bead-id>}"

# Extract quest ID: strip the trailing sub-bead suffix (e.g., ol-527.1 -> ol-527)
QUEST_ID="${BEAD_ID%.*}"

echo "ol-ratchet: bead=${BEAD_ID} quest=${QUEST_ID}"

# Call Olympus hero ratchet, capturing stderr for error reporting
if stderr=$(ol hero ratchet "$BEAD_ID" --quest "$QUEST_ID" 2>&1 1>/dev/null); then
    echo "ol-ratchet: success — ratchet complete for ${BEAD_ID}"
    exit 0
else
    echo "ol-ratchet: validation failed for ${BEAD_ID}" >&2
    if [[ -n "${stderr}" ]]; then
        echo "ol-ratchet: error: ${stderr}" >&2
    fi
    exit 1
fi
```

### ol-wave-loader.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# OL Wave Loader - Bridges OL hero hunt output into AO swarm execution
# Extracts and validates wave data from ol hero hunt JSON output.
# Called once at swarm startup when --from-wave is passed.

WAVE_FILE="${1:?Usage: ol-wave-loader.sh <wave-json-file>}"

# Validate file exists and is readable
if [[ ! -f "$WAVE_FILE" ]]; then
    echo "Error: Wave file not found: $WAVE_FILE" >&2
    exit 1
fi

if [[ ! -r "$WAVE_FILE" ]]; then
    echo "Error: Wave file not readable: $WAVE_FILE" >&2
    exit 1
fi

# Extract and validate the wave array from ol hero hunt output
# Check that each wave entry has required fields: id, title, spec_path, priority
wave_entries=$(jq -c '.wave[]?' "$WAVE_FILE" 2>/dev/null) || {
    echo "Error: Failed to parse wave array from $WAVE_FILE (not valid JSON or missing 'wave' key)" >&2
    exit 1
}

# Process each wave entry
if [[ -z "$wave_entries" ]]; then
    echo "Error: No wave entries found in $WAVE_FILE" >&2
    exit 1
fi

# Validate and output sorted entries
results=()
while IFS= read -r entry; do
    # Validate required fields
    id=$(echo "$entry" | jq -r '.id // empty' 2>/dev/null)
    if [[ -z "$id" ]]; then
        echo "Error: Missing or invalid 'id' field in wave entry: $entry" >&2
        exit 1
    fi

    title=$(echo "$entry" | jq -r '.title // empty' 2>/dev/null)
    if [[ -z "$title" ]]; then
        echo "Error: Missing or invalid 'title' field in wave entry: $entry" >&2
        exit 1
    fi

    spec_path=$(echo "$entry" | jq -r '.spec_path // empty' 2>/dev/null)
    if [[ -z "$spec_path" ]]; then
        echo "Error: Missing or invalid 'spec_path' field in wave entry: $entry" >&2
        exit 1
    fi

    priority=$(echo "$entry" | jq -r '.priority // empty' 2>/dev/null)
    if [[ -z "$priority" ]]; then
        echo "Error: Missing or invalid 'priority' field in wave entry: $entry" >&2
        exit 1
    fi

    # Validate priority is a number
    if ! [[ "$priority" =~ ^[0-9]+$ ]]; then
        echo "Error: Priority must be a number, got '$priority' for bead $id" >&2
        exit 1
    fi

    # Store result for sorting (priority first for sorting, then output columns)
    results+=("$priority"$'\t'"$id"$'\t'"$title"$'\t'"$spec_path")
done <<< "$wave_entries"

# Sort by priority (column 1, numeric) and output in correct order (id, title, spec_path, priority)
printf '%s\n' "${results[@]}" | sort -t$'\t' -k1 -n | awk -F$'\t' '{print $2"\t"$3"\t"$4"\t"$1}'
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: swarm" "grep -q '^name: swarm' '$SKILL_DIR/SKILL.md'"
check "Local mode documented" "grep -q 'Local' '$SKILL_DIR/SKILL.md'"
check "Backend references documented" "grep -q 'backend-claude-teams' '$SKILL_DIR/SKILL.md'"
check "Shared backend docs exist" "[ -f '$SKILL_DIR/..$shared/references/backend-claude-teams.md' ]"
check "Cleanup lifecycle documented" "grep -qE 'TeamDelete|close_agent|cleanup' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


