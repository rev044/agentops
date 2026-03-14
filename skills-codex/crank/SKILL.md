---
name: crank
description: 'Hands-free epic execution. Runs until ALL children are CLOSED. Uses spawn_agents_on_csv for parallel waves. NO human prompts, NO stopping. Triggers: "crank", "run epic", "execute epic", "run all tasks", "hands-free execution", "crank it".'
---

# $crank — Autonomous Epic Execution (Codex Native)

> **Quick Ref:** Execute all issues in an epic via wave-based parallel workers using `spawn_agents_on_csv`. Output: closed issues + final validation.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## Architecture

```
Crank (lead agent)
    |
    +-> bd ready (wave issues)
    |
    +-> Build CSV from ready issues
    |
    +-> spawn_agents_on_csv (parallel workers)
    |
    +-> wait for workers
    |
    +-> Verify results + bd update
    |
    +-> Loop until epic DONE
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--test-first` | off | SPEC → TEST → IMPL wave sequence. Workers classify tests by pyramid level (L0-L3) per the test pyramid standard (`test-pyramid.md` in the standards skill). When `$plan` includes `test_levels` metadata, carry it into `metadata.validation.test_levels`. |

## Global Limits

**MAX_EPIC_WAVES = 50** (hard limit). Typical epics use 5–10 waves.

## Completion Enforcement (Sisyphus Rule)

After each wave, output:
- `<promise>DONE</promise>` — epic complete, all issues closed
- `<promise>BLOCKED</promise>` — cannot proceed (with reason)
- `<promise>PARTIAL</promise>` — incomplete (with remaining items)

**Never claim completion without the marker.**

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

### Step 3: List Ready Issues (Current Wave)

```bash
bd ready 2>/dev/null
```

`bd ready` returns all unblocked issues — these can run in parallel.

### Step 3a: Pre-flight Checks

1. **Issues exist:** Verify there are ready issues. Empty list = error.
2. **Pre-mortem (3+ issues):** Check `.agents/council/` for pre-mortem evidence.
3. **Changed-string grep:** For every string being modified, grep the codebase for stale cross-references.

### Step 3b: Language Standards Injection

Detect project language (`go.mod` → Go, `pyproject.toml` → Python, etc.) and read applicable standards from `$standards`. Include Testing section in worker prompts.

### Step 4: Execute Wave via spawn_agents_on_csv

#### 4a: Build Wave CSV

Create a CSV file with one row per ready issue:

```bash
# Build CSV: id, subject, description, files, validation_cmd
CSV_FILE=".agents/crank/wave-${wave}-tasks.csv"
mkdir -p .agents/crank

echo "id,subject,description,files,validation_cmd" > "$CSV_FILE"

for issue_id in $READY_ISSUES; do
    ISSUE_DATA=$(bd show "$issue_id" 2>/dev/null)
    SUBJECT=$(echo "$ISSUE_DATA" | head -1 | sed 's/^[^·]*· //' | sed 's/  *\[.*//')
    FILES=$(echo "$ISSUE_DATA" | grep -oE '[a-zA-Z0-9_/.-]+\.(go|py|ts|sh|md|yaml|json)' | sort -u | paste -sd';')

    echo "\"$issue_id\",\"$SUBJECT\",\"$(echo "$ISSUE_DATA" | tail -n +3)\",\"$FILES\",\"go test ./...\"" >> "$CSV_FILE"
done
```

#### 4b: Pre-Spawn File Conflict Check

```
wave_tasks = [tasks from CSV]
all_files = {}
for task in wave_tasks:
    for f in task.files:
        if f in all_files:
            CONFLICT → serialize into sub-waves
        all_files[f] = task.id
```

Display ownership table. If conflicts > 0, split into sub-waves.

#### 4c: Spawn Workers

```
spawn_agents_on_csv(
    csv_path=".agents/crank/wave-{wave}-tasks.csv",
    instruction="You are a worker implementing issue {id}: {subject}.

{description}

Files to modify: {files}

Rules:
1. Stay within your assigned files
2. Run validation: {validation_cmd}
3. Report result via report_agent_job_result

Knowledge artifacts are in .agents/. See .agents/AGENTS.md for navigation.",
    id_column="id",
    output_schema={
        "type": "object",
        "properties": {
            "issue_id": {"type": "string"},
            "status": {"type": "string", "enum": ["PASS", "FAIL", "BLOCKED"]},
            "files_changed": {"type": "array", "items": {"type": "string"}},
            "reason": {"type": "string"}
        },
        "required": ["issue_id", "status", "files_changed", "reason"],
        "additionalProperties": false
    },
    max_concurrency=4,
    max_runtime_seconds=600
)
```

#### 4d: Wait for Workers

```
wait(timeout_seconds=1800)
```

Collect all `report_agent_job_result` outputs.

### Step 5: Verify and Sync

For each completed worker:

1. **Check result:** PASS → close issue, FAIL → log retry
2. **Run validation:** Execute the validation command from metadata
3. **Update beads:**
   ```bash
   bd close "$issue_id" 2>/dev/null   # On PASS
   bd update "$issue_id" --status blocked --append-notes "Wave $wave FAIL: $reason" 2>/dev/null  # On FAIL
   ```

### Step 5.5: Wave Acceptance Check

After all workers complete:
1. Compute `git diff` for the wave
2. Run project-level tests (`go test ./...` or equivalent)
3. If tests fail, identify which worker's changes broke things

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

**Lead-only commit** — workers write files, lead validates and commits once per wave:

```bash
# Stage only files reported by workers (avoid untracked temp files)
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

# Check remaining work
REMAINING=$(bd ready 2>/dev/null | wc -l)
if [[ $REMAINING -eq 0 ]]; then
    # Check if ALL issues are closed
    ALL_CLOSED=$(bd children "$EPIC_ID" 2>/dev/null | grep -c "CLOSED" || echo 0)
    ALL_TOTAL=$(bd children "$EPIC_ID" 2>/dev/null | wc -l || echo 0)

    if [[ $ALL_CLOSED -eq $ALL_TOTAL ]]; then
        echo "<promise>DONE</promise>"
    else
        echo "<promise>BLOCKED</promise>"
        echo "No ready issues but $((ALL_TOTAL - ALL_CLOSED)) issues remain unclosed."
    fi
else
    # Continue to next wave — return to Step 3
fi
```

### Step 8: Final Validation

When epic is DONE:

```bash
$vibe validate the completed epic
```

### Retry Policy

- **Max 2 retries per issue** across all waves
- On third failure: mark BLOCKED, continue with remaining issues
- Track retries: `bd comments add "$issue_id" "retry $N: $reason"`

### Failure Recovery

| Scenario | Action |
|----------|--------|
| Worker timeout | Mark BLOCKED, log reason, continue wave |
| Test failure | Identify breaking change, retry once |
| All workers fail | `<promise>BLOCKED</promise>` with diagnostics |
| File conflict detected | Split into sub-waves, re-run |

## References

For detailed patterns, read:
- `references/team-coordination.md` — worker coordination details
- `references/failure-recovery.md` — escalation and retry logic
- `references/wave-patterns.md` — acceptance checks and checkpoints
- `references/commit-strategies.md` — per-task vs wave-batch commits
