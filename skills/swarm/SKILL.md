---
name: swarm
tier: orchestration
description: 'Spawn isolated agents for parallel task execution. Local mode auto-selects Codex sub-agents or Claude teams. Distributed mode uses tmux + Agent Mail (process isolation, persistence). Triggers: "swarm", "spawn agents", "parallel work".'
dependencies:
  - implement # required - executes `/implement <bead-id>` in distributed mode
  - vibe      # optional - integration with validation
---

# Swarm Skill

Spawn isolated agents to execute tasks in parallel. Fresh context per agent (Ralph Wiggum pattern).

**Execution Modes:**
- **Local** (default) - Runtime-native spawning (Codex sub-agents when available, otherwise Claude teams/task agents)
- **Distributed** (`--mode=distributed`) - tmux sessions + Agent Mail for robust coordination

**Integration modes:**
- **Direct** - Create TaskList tasks, invoke `/swarm`
- **Via Crank** - `/crank` creates tasks from beads, invokes `/swarm` for each wave

## Architecture (Mayor-First)

```
Mayor (this session)
    |
    +-> Plan: TaskCreate with dependencies
    |
    +-> Identify wave: tasks with no blockers
    |
    +-> Select spawn backend (Codex sub-agents | Claude teams | fallback tasks)
    |
    +-> Assign: TaskUpdate(taskId, owner="worker-<id>", status="in_progress")
    |
    +-> Spawn workers via selected backend
    |       Workers receive pre-assigned task, execute atomically
    |
    +-> Wait for completion (wait() | SendMessage | TaskOutput)
    |
    +-> Validate: Review changes when complete
    |
    +-> Cleanup backend resources (close_agent | TeamDelete | none)
    |
    +-> Repeat: New team + new plan if more work needed
```

## Execution

Given `/swarm`:

### Step 0: Select Local Spawn Backend (MANDATORY)

Use runtime capability detection, not hardcoded assumptions:

1. If `spawn_agent` is available, use **Codex experimental sub-agents**
2. Else if `TeamCreate` is available, use **Claude native teams**
3. Else use **background task fallback** (`Task(run_in_background=true)`)

See `skills/shared/SKILL.md` ("Runtime-Native Spawn Backend Selection") for the shared contract used by all orchestration skills.

### Step 1: Ensure Tasks Exist

Use TaskList to see current tasks. If none, create them:

```
TaskCreate(subject="Implement feature X", description="Full details...")
TaskUpdate(taskId="2", addBlockedBy=["1"])  # Add dependencies after creation
```

### Step 2: Identify Wave

Find tasks that are:
- Status: `pending`
- No blockedBy (or all blockers completed)

These can run in parallel.

### Steps 3-6: Spawn Workers, Validate, Finalize

**For detailed local mode execution (team creation, worker spawning, race condition prevention, git commit policy, validation contract, cleanup, and repeat logic), read `skills/swarm/references/local-mode.md`.**

## Example Flow

```
Mayor: "Let's build a user auth system"

1. /plan -> Creates tasks:
   #1 [pending] Create User model
   #2 [pending] Add password hashing (blockedBy: #1)
   #3 [pending] Create login endpoint (blockedBy: #1)
   #4 [pending] Add JWT tokens (blockedBy: #3)
   #5 [pending] Write tests (blockedBy: #2, #3, #4)

2. /swarm -> Spawns agent for #1 (only unblocked task)

3. Agent #1 completes -> #1 now completed
   -> #2 and #3 become unblocked

4. /swarm -> Spawns agents for #2 and #3 in parallel

5. Continue until #5 completes

6. /vibe -> Validate everything
```

## Key Points

- **Runtime-native local mode** - Auto-selects Codex sub-agents or Claude teams
- **Universal orchestration contract** - Same swarm behavior across Claude and Codex sessions
- **Pre-assigned tasks** - Mayor assigns tasks before spawning; workers never race-claim
- **Fresh worker contexts** - New sub-agents/teammates per wave preserve Ralph isolation
- **Wave execution** - Only unblocked tasks spawn
- **Mayor orchestrates** - You control the flow, workers report via backend channel
- **Retry via message/input** - Use `send_input` (Codex) or `SendMessage` (Claude)
- **Atomic execution** - Each worker works until task done
- **Graceful fallback** - If richer APIs unavailable, fall back to `Task(run_in_background=true)`

## Integration with AgentOps

This ties into the full workflow:

```
/research -> Understand the problem
/plan -> Decompose into beads issues
/crank -> Autonomous epic loop
    +-- /swarm -> Execute each wave in parallel
/vibe -> Validate results
/post-mortem -> Extract learnings
```

**Direct use (no beads):**
```
TaskCreate -> Define tasks
/swarm -> Execute in parallel
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
| `--mode=local\|distributed` | Execution mode | `local` |
| `--max-workers=N` | Max concurrent workers | 5 |
| `--from-wave <json-file>` | Load wave from OL hero hunt output (see OL Wave Integration) | - |
| `--bead-ids` | Specific beads to work (comma-separated, distributed mode) | Auto from `bd ready` |
| `--wait` | Wait for all workers to complete (distributed mode) | false |
| `--timeout` | Max time to wait if `--wait` (distributed mode) | 30m |

## When to Use Swarm

| Scenario | Use |
|----------|-----|
| Multiple independent tasks | `/swarm` (parallel) |
| Sequential dependencies | `/swarm` with blockedBy |
| Mix of both | `/swarm` spawns waves, each wave parallel |

## Why This Works: Ralph Wiggum Pattern

Follows the [Ralph Wiggum Pattern](https://ghuntley.com/ralph/): **fresh context per execution unit**.

- **Wave-scoped worker set** = spawn workers -> execute -> cleanup -> repeat (fresh context each wave)
- **Mayor IS the loop** - Orchestration layer, manages state across waves
- **Workers are atomic** - One task, one spawn, one result
- **TaskList as memory** - State persists in task status, not agent context
- **Filesystem for artifacts** - Files written by workers, committed by team lead
- **Backend messaging for coordination** - Workers report to team lead, never to each other

## Integration with Crank

When `/crank` invokes `/swarm`: Crank bridges beads to TaskList, swarm executes with fresh-context agents, crank syncs results back.

| You Want | Use | Why |
|----------|-----|-----|
| Fresh-context parallel execution | `/swarm` | Each spawned agent is a clean slate |
| Autonomous epic loop | `/crank` | Loops waves via swarm until epic closes |
| Just swarm, no beads | `/swarm` directly | TaskList only, skip beads |
| RPI progress gates | `/ratchet` | Tracks progress; does not execute work |

---

## Distributed Mode

**For the full distributed mode specification (tmux + Agent Mail, experimental), read `skills/swarm/references/distributed-mode.md`.**
Run `scripts/multi-agent-preflight.sh --workflow swarm` before starting distributed mode.

---

## OL Wave Integration

When `/swarm --from-wave <json-file>` is invoked, the swarm reads wave data from an OL hero hunt output file and executes it with completion backflow to OL.

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
/swarm --from-wave /tmp/wave-ol-527.json

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
- **Distributed Mode:** `skills/swarm/references/distributed-mode.md`
- **Validation Contract:** `skills/swarm/references/validation-contract.md`
- **Agent Mail Protocol:** See `skills/shared/agent-mail-protocol.md` for message format specifications
- **Parser (Go):** `cli/internal/agentmail/` - shared parser for all message types
