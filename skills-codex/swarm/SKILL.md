---
name: swarm
description: 'Spawn isolated agents for parallel task execution via spawn_agents_on_csv. Fresh context per agent (Ralph Wiggum pattern). Triggers: "swarm", "spawn agents", "parallel work", "run in parallel", "parallel execution".'
metadata:
  tier: orchestration
---

# $swarm — Parallel Agent Execution (Codex Native)

Spawn isolated agents to execute tasks in parallel using `spawn_agents_on_csv`. Fresh context per agent (Ralph Wiggum pattern).

**Integration modes:**
- **Via $crank** — crank creates tasks from beads, invokes swarm for each wave
- **Standalone** — direct invocation for ad-hoc parallel work

> **Requires Codex multi-agent.** Enable via `multi_agent = true` in `~/.codex/config.toml`.

## Architecture

```
Lead (this agent)
    |
    +-> Identify wave: tasks with no blockers
    |
    +-> Build CSV from tasks
    |
    +-> Pre-spawn conflict check (file ownership)
    |
    +-> spawn_agents_on_csv (parallel workers)
    |
    +-> wait for completion
    |
    +-> Validate: review changes, run tests
    |
    +-> Repeat if more work needed
```

## Execution

Given `$swarm`:

### Step 0: Verify Multi-Agent Available

Check that `spawn_agents_on_csv` is available. If not:
```
WARN: Multi-agent not available. Executing tasks sequentially in this session.
```
Fall back to serial execution within current session.

### Step 1: Ensure Tasks Exist

Tasks come from one of:
- `bd ready` output (beads mode)
- Explicit task list from `$crank`
- User-provided description (decompose first)

Each task needs:
- **id** — unique identifier
- **subject** — what to do
- **description** — detailed instructions
- **files** — file manifest (which files this worker owns)
- **validation** — how to verify completion

### Step 1.5: Auto-Populate File Manifests

If any task is missing its file manifest, spawn explorer agents to identify files:

```
spawn_agents_on_csv(
    csv_path=".agents/swarm/manifest-tasks.csv",
    instruction="Given this task: '{subject}', identify all files that will need to be created or modified. Return a JSON array of file paths.",
    id_column="id",
    output_schema={
        "type": "object",
        "properties": {
            "files": {"type": "array", "items": {"type": "string"}}
        },
        "required": ["files"],
        "additionalProperties": false
    },
    max_concurrency=6,
    max_runtime_seconds=120
)
```

### Step 2: Pre-Spawn Conflict Check

```
wave_tasks = [tasks with status=pending and no blockers]
all_files = {}
for task in wave_tasks:
    for f in task.files:
        if f in all_files:
            CONFLICT: f claimed by both all_files[f] and task.id
        all_files[f] = task.id
```

**On conflict:**
- **Serialize** conflicting workers into separate sub-waves (preferred)
- Do NOT spawn workers with overlapping file manifests in the same wave

**Display ownership table before spawning:**

```
File Ownership Map (Wave N):
┌─────────────────────────────┬──────────┬──────────┐
│ File                        │ Owner    │ Conflict │
├─────────────────────────────┼──────────┼──────────┤
│ src/auth/middleware.go       │ task-1   │          │
│ src/api/routes.go            │ task-2   │          │
└─────────────────────────────┴──────────┴──────────┘
Conflicts: 0
```

### Step 3: Build Worker CSV

```bash
CSV_FILE=".agents/swarm/wave-${wave}-workers.csv"
mkdir -p .agents/swarm

echo "id,subject,description,files,validation_cmd" > "$CSV_FILE"

for task in $WAVE_TASKS; do
    echo "\"$task_id\",\"$subject\",\"$description\",\"$files\",\"$validation\"" >> "$CSV_FILE"
done
```

### Step 4: Spawn Workers

```
spawn_agents_on_csv(
    csv_path=".agents/swarm/wave-{wave}-workers.csv",
    instruction="You are implementing: {subject}

{description}

YOUR FILE BOUNDARIES: {files}
Do NOT modify files outside this list.

After implementation:
1. Run validation: {validation_cmd}
2. Report result via report_agent_job_result with status PASS/FAIL

Knowledge artifacts are in .agents/. See .agents/AGENTS.md for navigation.
Use `ao lookup --query \"topic\"` for learnings if ao CLI is available.",
    id_column="id",
    output_schema={
        "type": "object",
        "properties": {
            "task_id": {"type": "string"},
            "status": {"type": "string", "enum": ["PASS", "FAIL", "BLOCKED"]},
            "files_changed": {"type": "array", "items": {"type": "string"}},
            "tests_passed": {"type": "boolean"},
            "reason": {"type": "string"}
        },
        "required": ["task_id", "status", "files_changed", "tests_passed", "reason"],
        "additionalProperties": false
    },
    max_concurrency=4,
    max_runtime_seconds=600
)
```

### Step 5: Wait and Collect Results

```
wait(timeout_seconds=1800)
```

Collect all `report_agent_job_result` outputs into a results table.

### Step 6: Validate Wave

For each worker result:

1. **PASS:** Accept changes
2. **FAIL:** Log failure, mark for retry (max 2 retries per task)
3. **BLOCKED:** Escalate to lead

After collecting results:

```bash
# Run project-level tests
go test ./... 2>&1  # or equivalent
```

If tests fail, identify which worker's changes caused the break.

### Step 7: Report Results

Output wave summary:

```
Wave N Results:
┌──────────┬──────────┬────────────────┐
│ Task     │ Status   │ Files Changed  │
├──────────┼──────────┼────────────────┤
│ ag-001   │ PASS     │ 3 files        │
│ ag-002   │ FAIL     │ 0 files        │
└──────────┴──────────┴────────────────┘
PASS: 1  FAIL: 1  BLOCKED: 0
```

### Test File Naming Validation

When workers create test files, validate naming:
- Go: `<source>_test.go` (reject `cov*_test.go`)
- Python: `test_<module>.py` or `<module>_test.py`

### Output Schema Size Guard

When 5+ workers share the same output schema, cache to `.agents/swarm/output-schema.json` and reference by path instead of inlining ~500 tokens per worker.

## Serial Fallback

If `spawn_agents_on_csv` is unavailable, execute tasks sequentially:

```
for task in wave_tasks:
    1. Read task details
    2. Implement changes
    3. Run validation
    4. Record result
```

This is slower but functionally identical.

## Reference Documents

- [references/backend-background-tasks.md](references/backend-background-tasks.md)
- [references/backend-codex-subagents.md](references/backend-codex-subagents.md)
- [references/backend-inline.md](references/backend-inline.md)
- [references/local-mode.md](references/local-mode.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)
- [references/validation-contract.md](references/validation-contract.md)
- [references/worker-pitfalls.md](references/worker-pitfalls.md)
