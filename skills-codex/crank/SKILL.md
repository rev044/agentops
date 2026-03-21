---
name: crank
description: 'Hands-free epic execution. Runs until ALL children are CLOSED. Uses Codex session agents for parallel waves. NO human prompts, NO stopping. Triggers: "crank", "run epic", "execute epic", "run all tasks", "hands-free execution", "crank it".'
---

# $crank - Autonomous Epic Execution (Codex Native)

> **Quick Ref:** Execute every open issue in an epic via wave-based workers using `spawn_agent`, `wait_agent`, `send_input`, and `close_agent`. Output: closed issues + final validation.

**You must execute this workflow. Do not just describe it.**

## Architecture

```text
Crank (lead agent)
    |
    +-> bd ready (current wave)
    |
    +-> Build a wave task packet
    |
    +-> spawn_agent per issue (worker or explorer role)
    |
    +-> wait_agent for all worker ids
    |
    +-> Validate results + bd update
    |
    +-> Loop until epic DONE
```

## Backend Rules

1. Prefer Codex session agents when `spawn_agent` is available.
2. Use `agent_type=worker` for implementation agents and `agent_type=explorer` for discovery agents when the runtime exposes roles.
3. Use `send_input` only for short steering or retry prompts.
4. Use `close_agent` for stalled or unnecessary agents.
5. Never depend on legacy CSV fan-out or host-task result polling. Use `spawn_agent`, `wait_agent`, `send_input`, and `close_agent` instead.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--test-first` | off | SPEC -> TEST -> IMPL wave sequence. Workers classify tests by pyramid level (L0-L3) per the test pyramid standard (`test-pyramid.md` in the standards skill). When `$plan` includes `test_levels` metadata, carry it into `metadata.validation.test_levels`. |

## Global Limits

**MAX_EPIC_WAVES = 50** (hard limit). Typical epics use 5-10 waves.

## Completion Enforcement (Sisyphus Rule)

After each wave, output one of:
- `<promise>DONE</promise>` - epic complete, all issues closed
- `<promise>BLOCKED</promise>` - cannot proceed, with reason
- `<promise>PARTIAL</promise>` - incomplete, with remaining items

Never claim completion without the marker.

## Node Repair Operator

When a task fails during wave execution, classify as **RETRY** (transient — re-add with adjustment, max 2), **DECOMPOSE** (too complex — split into sub-issues, terminal), or **PRUNE** (blocked — escalate immediately). Budget: 2 per task.

## Execution Steps

Given `$crank [epic-id | plan-file.md | "description"]`:

### Step 0: Load Knowledge Context

```bash
if command -v ao &>/dev/null; then
    ao lookup --query "<epic-title>" --limit 5 2>/dev/null || true
    ao ratchet status 2>/dev/null || true
fi
```

### Step 0.5: Detect Tracking Mode

```bash
if command -v bd &>/dev/null; then
    TRACKING_MODE="beads"
else
    TRACKING_MODE="file"
fi
```

### Step 1: Identify the Epic

**Beads mode:**
- If epic ID provided: use it directly
- If no epic ID: `bd list --type epic --status open 2>/dev/null | head -5`

**File mode:**
- Read the plan file and extract tasks

### Step 2: Get Epic Details

```bash
bd show <epic-id> 2>/dev/null
```

### Step 3: List Ready Issues for the Current Wave

```bash
bd ready 2>/dev/null
```

`bd ready` returns all unblocked issues - these can run in parallel.

### Step 3a: Pre-flight Checks

1. Verify there are ready issues. Empty list is an error unless the epic is already complete.
2. If 3+ issues are ready, check `.agents/council/` for pre-mortem evidence.
3. For every string being modified, grep the codebase for stale cross-references.

### Step 3b: Language Standards Injection

Detect project language (`go.mod` -> Go, `pyproject.toml` -> Python, etc.) and read applicable standards from `$standards`. Include a Testing section in worker prompts.

### Step 4: Execute the Wave with Codex Session Agents

Crank follows the FIRE loop for each wave:
- **FIND:** locate the next ready set
- **IGNITE:** spawn workers
- **REAP:** wait, validate, and merge results
- **ESCALATE:** retry or block when needed

#### 4a: Build a Wave Task Packet

Create one packet per ready issue. Do not use CSV fan-out.

```bash
mkdir -p .agents/crank
cat > ".agents/crank/wave-${wave}-tasks.json" << EOF
{
  "wave": $wave,
  "epic_id": "$EPIC_ID",
  "tasks": [
    {
      "issue_id": "bd-123",
      "subject": "Short issue summary",
      "description": "Issue details and acceptance criteria",
      "files": ["path/to/file.go"],
      "validation_cmd": "go test ./...",
      "metadata": {
        "issue_type": "feature"
      }
    }
  ]
}
EOF
```

Each task packet must include `metadata.issue_type`.

#### 4b: Pre-spawn File Conflict Check

```text
wave_tasks = [tasks from packet]
all_files = {}
for task in wave_tasks:
    for f in task.files:
        if f in all_files:
            CONFLICT -> serialize into sub-waves
        all_files[f] = task.id
```

Display an ownership table before spawning workers. If conflicts exist, split into sub-waves and keep file ownership disjoint.

#### 4c: Spawn Workers

Spawn one agent per issue. Prefer `worker` roles for implementation and `explorer` roles for file discovery when the runtime exposes `agent_type`.

```text
spawn_agent(
  agent_type="worker",
  message="You are worker-<issue-id>.

Assignment: <subject>

<description>

FILE MANIFEST (files you are permitted to modify):
<list of files>

Rules:
1. Stay within your assigned files
2. Run validation: <validation_cmd>
3. Keep your response short
4. Write any durable notes to .agents/crank/results/<issue-id>.md or .agents/crank/results/<issue-id>.json

Use the repo's current Codex primitives only."
)
```

If a task is missing its file manifest, spawn a short-lived `explorer` agent first:

```text
spawn_agent(
  agent_type="explorer",
  message="You are explorer-<issue-id>.

Task: identify the files that must be created or modified for this issue.
Return a JSON array of paths only."
)
```

#### 4d: Wait for Workers

```text
wait_agent(ids=["agent-id-1", "agent-id-2"])
```

If a worker needs a short correction, use `send_input(id=..., message=...)`.

If a worker stalls or is no longer needed, use `close_agent(id=...)`.

### Step 5: Verify and Sync

For each completed worker:

1. PASS -> close the issue.
2. FAIL -> log the failure, keep the issue open, and retry only if the issue is still within the retry budget.
3. BLOCKED -> mark blocked with the reason and continue the wave.

Update beads:

```bash
bd close "$issue_id" 2>/dev/null
bd update "$issue_id" --status blocked --append-notes "Wave $wave FAIL: $reason" 2>/dev/null
```

### Step 5.5: Wave Acceptance Check

After all workers complete:
1. Compute `git diff` for the wave.
2. Run project-level tests appropriate to the wave.
3. If tests fail, identify which worker's changes broke things and requeue only that work.

### Step 5.7: Wave Checkpoint

```bash
cat > ".agents/crank/wave-${wave}-checkpoint.json" << EOF
{
  "wave": $wave,
  "epic_id": "$EPIC_ID",
  "completed": $COMPLETED_COUNT,
  "failed": $FAILED_COUNT,
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
```

### Step 6: Commit Wave Results

**Lead-only commit** - workers write files, lead validates and commits once per wave:

```bash
for f in $WORKER_FILES_CHANGED; do
    git add -- "$f"
done
git commit -m "feat(<scope>): wave $wave - $COMPLETED_COUNT issues completed"
```

### Step 7: Loop or Complete

```bash
wave=$((wave + 1))

if [[ $wave -ge 50 ]]; then
    echo "<promise>BLOCKED</promise>"
    echo "Global wave limit (50) reached."
    exit 1
fi

REMAINING=$(bd ready 2>/dev/null | wc -l)
if [[ $REMAINING -eq 0 ]]; then
    ALL_CLOSED=$(bd children "$EPIC_ID" 2>/dev/null | grep -c "CLOSED" || echo 0)
    ALL_TOTAL=$(bd children "$EPIC_ID" 2>/dev/null | wc -l || echo 0)

    if [[ $ALL_CLOSED -eq $ALL_TOTAL ]]; then
        echo "<promise>DONE</promise>"
    else
        echo "<promise>BLOCKED</promise>"
        echo "No ready issues but $((ALL_TOTAL - ALL_CLOSED)) issues remain unclosed."
    fi
else
    # Continue to next wave - return to Step 3
fi
```

### Step 8: Final Validation

When the epic is DONE:

```bash
$vibe validate the completed epic
```

## Retry Policy

- Max 2 retries per issue across all waves
- On third failure: mark BLOCKED and continue with remaining issues
- Track retries with `bd comments add "$issue_id" "retry $N: $reason"`

## Failure Recovery

| Scenario | Action |
|----------|--------|
| Worker timeout | Mark BLOCKED, log reason, continue wave |
| Test failure | Identify breaking change, retry once |
| All workers fail | `<promise>BLOCKED</promise>` with diagnostics |
| File conflict detected | Split into sub-waves, re-run |

## Reference Documents

- [references/commit-strategies.md](references/commit-strategies.md) - per-task vs wave-batch commits
- [references/contract-template.md](references/contract-template.md) - contract template for worker specs
- [references/failure-recovery.md](references/failure-recovery.md) - escalation and retry logic
- [references/failure-taxonomy.md](references/failure-taxonomy.md) - failure classification
- [references/fire.md](references/fire.md) - FIRE loop specification
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md) - Ralph Wiggum loop contract
- [references/taskcreate-examples.md](references/taskcreate-examples.md) - task creation examples
- [references/team-coordination.md](references/team-coordination.md) - worker coordination details
- [references/test-first-mode.md](references/test-first-mode.md) - test-first wave sequence
- [references/troubleshooting.md](references/troubleshooting.md) - common issues and fixes
- [references/uat-integration-wave.md](references/uat-integration-wave.md) - UAT integration wave patterns
- [references/wave-patterns.md](references/wave-patterns.md) - acceptance checks and checkpoints
- [references/wave1-spec-consistency-checklist.md](references/wave1-spec-consistency-checklist.md) - Wave 1 spec consistency checklist
- [references/worktree-per-worker.md](references/worktree-per-worker.md) - worktree isolation pattern
