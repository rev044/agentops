---
name: crank
tier: orchestration
description: 'Fully autonomous epic execution. Runs until ALL children are CLOSED. Local mode uses /swarm with runtime-native spawning (Codex sub-agents or Claude teams). Distributed mode uses /swarm --mode=distributed (tmux + Agent Mail) for persistence and coordination. NO human prompts, NO stopping.'
dependencies:
  - swarm       # required - executes each wave
  - vibe        # required - final validation
  - implement   # required - individual issue execution
  - beads       # optional - issue tracking via bd CLI (fallback: TaskList)
  - post-mortem # optional - suggested for learnings extraction
---

# Crank Skill

> **Quick Ref:** Autonomous epic execution. Local mode: `/swarm` for each wave with runtime-native spawning. Distributed mode: `/swarm --mode=distributed` (tmux + Agent Mail). Output: closed issues + final vibe.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Autonomous execution: implement all issues until the epic is DONE.

**CLI dependencies:** bd (issue tracking), ao (knowledge flywheel). Both optional — see `skills/shared/SKILL.md` for fallback table. If bd is unavailable, use TaskList for issue tracking and skip beads sync. If ao is unavailable, skip knowledge injection/extraction.

## Architecture: Crank + Swarm

**Beads mode** (bd available):
```
Crank (orchestrator)           Swarm (executor)
    |                              |
    +-> bd ready (wave issues)     |
    |                              |
    +-> TaskCreate from beads  --->+-> Select spawn backend (codex sub-agents | claude teams | fallback)
    |                              |
    +-> /swarm                 --->+-> Spawn workers per backend
    |                              |   (fresh context per wave)
    +-> Verify + bd update     <---+-> Workers report via backend channel
    |                              |
    +-> Loop until epic DONE   <---+-> Cleanup backend resources after wave
```

**TaskList mode** (bd unavailable):
```
Crank (orchestrator, TaskList mode)    Swarm (executor)
    |                                      |
    +-> TaskList() (wave tasks)            |
    |                                      |
    +-> /swarm                         --->+-> Select spawn backend per wave
    |                                      |
    +-> Verify via TaskList()          <---+-> Workers report via backend channel
    |                                      |
    +-> Loop until all completed       <---+-> Cleanup backend resources after wave
```

**Separation of concerns:**
- **Crank** = Orchestration, epic/task lifecycle, knowledge flywheel
- **Swarm** = Runtime-native parallel execution (Ralph Wiggum pattern via fresh worker set per wave)

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--test-first` | off | Enable spec-first TDD: SPEC WAVE generates contracts, TEST WAVE generates failing tests, IMPL WAVES make tests pass |

## Global Limits

**MAX_EPIC_WAVES = 50** (hard limit across entire epic)

This prevents infinite loops on circular dependencies or cascading failures.

**Why 50?**
- Typical epic: 5-10 issues
- With retries: ~5 waves max
- 50 = safe upper bound

## Completion Enforcement (The Sisyphus Rule)

**THE SISYPHUS RULE:** Not done until explicitly DONE.

After each wave, output completion marker:
- `<promise>DONE</promise>` - Epic truly complete, all issues closed
- `<promise>BLOCKED</promise>` - Cannot proceed (with reason)
- `<promise>PARTIAL</promise>` - Incomplete (with remaining items)

**Never claim completion without the marker.**

## Execution Steps

Given `/crank [epic-id | plan-file.md | "description"]`:

### Step 0: Load Knowledge Context (ao Integration)

**Search for relevant learnings before starting the epic:**

```bash
# If ao CLI available, inject prior knowledge about epic execution
if command -v ao &>/dev/null; then
    # Search for relevant learnings
    ao search "epic execution implementation patterns" 2>/dev/null | head -20

    # Check flywheel status
    ao flywheel status 2>/dev/null

    # Get current ratchet state
    ao ratchet status 2>/dev/null
fi
```

If ao not available, skip this step and proceed. The knowledge flywheel enhances but is not required.

### Step 0.5: Detect Tracking Mode

```bash
if command -v bd &>/dev/null; then
  TRACKING_MODE="beads"
else
  TRACKING_MODE="tasklist"
  echo "Note: bd CLI not found. Using TaskList for issue tracking."
fi
```

**Tracking mode determines the source of truth for the rest of the workflow:**

| | Beads Mode | TaskList Mode |
|---|---|---|
| **Source of truth** | `bd` (beads issues) | TaskList (Claude-native) |
| **Find work** | `bd ready` | `TaskList()` → pending, unblocked |
| **Get details** | `bd show <id>` | `TaskGet(taskId)` |
| **Mark complete** | `bd update <id> --status closed` | `TaskUpdate(taskId, status="completed")` |
| **Track retries** | `bd comments add` | Task description update |
| **Epic tracking** | `bd update <epic-id> --append-notes` | In-memory wave counter |

### Step 1: Identify the Epic / Work Source

**Beads mode:**

**If epic ID provided:** Use it directly. Do NOT ask for confirmation.

**If no epic ID:** Discover it:
```bash
bd list --type epic --status open 2>/dev/null | head -5
```

If multiple epics found, ask user which one.

**TaskList mode:**

If input is an epic ID → Error: "bd CLI required for beads epic tracking. Install bd or provide a plan file / task list."

If input is a plan file path (`.md`):
1. Read the plan file
2. Decompose into TaskList tasks (one `TaskCreate` per distinct work item)
3. Set up dependencies via `TaskUpdate(addBlockedBy)`
4. Proceed to Step 3

If no input:
1. Check `TaskList()` for existing pending tasks
2. If tasks exist, use them as the work items
3. If no tasks, ask user what to work on

If input is a description string:
1. Decompose into tasks (`TaskCreate` for each)
2. Set up dependencies
3. Proceed to Step 3

### Step 1a: Initialize Wave Counter

**Beads mode:**
```bash
# Initialize crank tracking in epic notes
bd update <epic-id> --append-notes "CRANK_START: wave=0 at $(date -Iseconds)" 2>/dev/null
```

**TaskList mode:** Track wave counter in memory only. No external state needed.

Track in memory: `wave=0`

### Step 1b: Detect Test-First Mode (--test-first only)

```bash
# Check for --test-first flag
if [[ "$TEST_FIRST" == "true" ]]; then
    # Classify issues by category
    # spec-eligible: feature, bugfix, refactor → SPEC + TEST waves apply
    # skip: docs, chore, ci → standard implementation waves only
    INVALID_ISSUES=()
    for issue in $READY_ISSUES; do
        CATEGORY=$(bd show "$issue" 2>/dev/null | grep -i "category:" | head -1 | awk '{print tolower($NF)}')
        # Validate category is non-empty and matches known values
        if [[ -z "$CATEGORY" ]]; then
            echo "ERROR: Issue $issue has no category field. Valid: feature|bugfix|refactor|docs|chore|ci"
            INVALID_ISSUES+=("$issue")
            continue
        fi
        case "$CATEGORY" in
            *feature*|*bugfix*|*refactor*) SPEC_ELIGIBLE+=("$issue") ;;
            *docs*|*chore*|*ci*) SPEC_SKIP+=("$issue") ;;
            *)
                echo "WARNING: Issue $issue has unknown category '$CATEGORY'. Defaulting to spec-eligible."
                SPEC_ELIGIBLE+=("$issue")
                ;;
        esac
    done
    # Fail-fast if any issues have missing categories
    if [[ ${#INVALID_ISSUES[@]} -gt 0 ]]; then
        echo "<promise>BLOCKED</promise>"
        echo "Issues with missing/invalid categories: ${INVALID_ISSUES[*]}"
        echo "Add category to each issue: bd update <id> --category <feature|bugfix|refactor|docs|chore|ci>"
        # STOP - do not continue
    fi
    echo "Test-first mode: ${#SPEC_ELIGIBLE[@]} spec-eligible, ${#SPEC_SKIP[@]} skipped (docs/chore/ci)"
fi
```

If `--test-first` is NOT set, skip Steps 3b and 3c entirely — behavior is unchanged.

### Step 2: Get Epic Details

**Beads mode:**
```bash
bd show <epic-id> 2>/dev/null
```

**TaskList mode:** `TaskList()` to see all tasks and their status/dependencies.

### Step 3: List Ready Issues (Current Wave)

**Beads mode:**

Find issues that can be worked on (no blockers):
```bash
bd ready 2>/dev/null
```

**`bd ready` returns the current wave** - all unblocked issues. These can be executed in parallel because they have no dependencies on each other.

**TaskList mode:**

`TaskList()` → filter for status=pending, no blockedBy (or all blockers completed). These are the current wave.

### Step 3a: Pre-flight Check - Issues Exist

**Verify there are issues to work on:**

**If 0 ready issues found (beads mode) or 0 pending unblocked tasks (TaskList mode):**
```
STOP and return error:
  "No ready issues found for this epic. Either:
   - All issues are blocked (check dependencies)
   - Epic has no child issues (run /plan first)
   - All issues already completed"
```

Also verify: epic has at least 1 child issue total. An epic with 0 children means /plan was not run.

Do NOT proceed with empty issue list - this produces false "epic complete" status.

### Step 3b: SPEC WAVE (--test-first only)

> **Purpose:** Generate contracts that ground implementation in verified requirements.

**Skip this step if `--test-first` is NOT set or if no spec-eligible issues exist.**

For each **spec-eligible** issue (feature/bugfix/refactor):

1. **TaskCreate** with subject `SPEC: <issue-title>`
2. **Worker prompt:**
   ```
   You are a spec writer. Generate a contract for this issue.

   FIRST: Explore the codebase to understand existing patterns, types, and interfaces
   relevant to this issue. Use Glob and Read to examine the code.

   THEN: Read the contract template at skills/crank/references/contract-template.md.

   Generate a contract following the template. Include:
   - At least 3 invariants
   - At least 3 test cases mapped to invariants
   - Concrete types and interfaces from the actual codebase

   If inputs are missing or the issue is underspecified, write BLOCKED with reason.

   Output: .agents/specs/contract-<issue-id>.md
   ```
3. **Worker receives:** Issue description, plan boundaries, contract template, codebase access (read-only)
4. **Validation:** files_exist + content_check for `## Invariants` AND `## Test Cases`
5. **Lead commits** all specs after validation: `git add .agents/specs/ && git commit -m "spec: contracts for <issue-ids>"`

**Category-based skip:** Issues categorized as docs/chore/ci bypass SPEC and TEST waves entirely and proceed directly to standard implementation waves.

#### SPEC WAVE BLOCKED Recovery

If a spec worker writes `BLOCKED` instead of a contract:

1. **Read the BLOCKED reason** from the worker output
2. **Add context to the issue:**
   ```bash
   bd comments add <issue-id> "SPEC BLOCKED: <reason>. Retrying with additional context..." 2>/dev/null
   ```
3. **Retry once** with enriched prompt (include the BLOCKED reason + additional codebase context)
4. **If still BLOCKED after 2 attempts**, escalate:
   ```bash
   bd update <issue-id> --labels BLOCKER 2>/dev/null
   bd comments add <issue-id> "ESCALATED: Spec generation failed 2x. Reason: <reason>. Human review required." 2>/dev/null
   ```
   Remove the issue from spec-eligible list and continue with remaining issues. Do NOT block the entire wave.

### Step 3c: TEST WAVE (--test-first only)

> **Purpose:** Generate failing tests from contracts. Tests must FAIL (RED confirmation).

**Skip this step if `--test-first` is NOT set or if no spec-eligible issues exist.**

For each **spec-eligible** issue:

1. **TaskCreate** with subject `TEST: <issue-title>`
2. **Worker prompt:**
   ```
   You are a test writer. Generate FAILING tests from the contract.

   Read ONLY the contract at .agents/specs/contract-<issue-id>.md.
   You may read codebase structure (imports, types, interfaces) but NOT existing
   implementation details.

   Generate tests that:
   - Cover ALL test cases from the contract's Test Cases table
   - Cover ALL invariants (at least one test per invariant)
   - All tests MUST FAIL when run (RED state)
   - Follow existing test patterns in the codebase

   Do NOT read or reference existing implementation code.
   Do NOT write implementation code.

   Output: test files in the appropriate location for the project's test framework.
   ```
3. **Worker receives:** contract-<issue-id>.md + codebase structure (imports, types) but NOT existing implementations
4. **Validation:** test files exist + RED confirmation (lead runs test suite, all new tests must fail)
5. **RED Gate:** Lead runs the test suite. ALL new tests must FAIL:
   ```bash
   # Run tests — expect failures for new tests
   # If any new test PASSES, the test is not meaningful (validates existing behavior, not new)
   ```
6. **Lead commits** test harness: `git add <test-files> && git commit -m "test: failing tests for <issue-ids> (RED)"`

#### RED Gate Enforcement

After TEST WAVE, the lead **must** verify RED state before proceeding:

```bash
# Run the test suite and capture results
TEST_OUTPUT=$(<test-command> 2>&1) || true
TEST_EXIT=$?

# Parse for unexpected passes among new test files
UNEXPECTED_PASSES=()
for test_file in $NEW_TEST_FILES; do
    # Check if tests in this file passed (framework-specific detection)
    if echo "$TEST_OUTPUT" | grep -q "PASS.*$(basename $test_file)"; then
        UNEXPECTED_PASSES+=("$test_file")
    fi
done

if [[ ${#UNEXPECTED_PASSES[@]} -gt 0 ]]; then
    echo "RED GATE FAILED: ${#UNEXPECTED_PASSES[@]} test file(s) passed unexpectedly:"
    printf '  - %s\n' "${UNEXPECTED_PASSES[@]}"
fi
```

**Decision tree for unexpected passes:**

| Condition | Action |
|-----------|--------|
| All new tests FAIL | PASS — proceed to IMPL wave |
| Some tests pass, some fail | Retry: re-generate passing tests with explicit "must fail" constraint |
| All new tests PASS | BLOCKED — tests validate existing behavior, not new requirements. Escalate to human. |

**On retry (max 2 attempts):**
1. Add the unexpected-pass context to the worker prompt
2. Re-spawn test writer with: "These tests passed unexpectedly: <list>. They must fail against current code. Rewrite them to test NEW behavior described in the contract."
3. If still passing after 2 retries, mark issue as BLOCKER and skip to standard IMPL

### Step 4: Execute Wave via Swarm

**GREEN mode (--test-first only):** If `--test-first` is set and SPEC/TEST waves have completed, modify worker prompts for spec-eligible issues:
- Include in each worker's TaskCreate: `"Failing tests exist at <test-file-paths>. Make them pass. Do NOT modify test files. See GREEN Mode rules in /implement SKILL.md."`
- Workers receive: failing tests (immutable), contract, issue description
- Workers follow GREEN Mode rules from `/implement` SKILL.md
- Docs/chore/ci issues (skipped by SPEC/TEST waves) use standard worker prompts unchanged

**BEFORE each wave:**
```bash
wave=$((wave + 1))
WAVE_START_SHA=$(git rev-parse HEAD)

if [[ "$TRACKING_MODE" == "beads" ]]; then
    bd update <epic-id> --append-notes "CRANK_WAVE: $wave at $(date -Iseconds)" 2>/dev/null
fi

# CHECK GLOBAL LIMIT
if [[ $wave -ge 50 ]]; then
    echo "<promise>BLOCKED</promise>"
    echo "Global wave limit (50) reached."
    # STOP - do not continue
fi
```

**Cross-cutting constraint injection (SDD):**

Before spawning workers, check for cross-cutting constraints:

```bash
# Guard clause: skip if plan has no boundaries (backward compat)
PLAN_FILE=$(ls -t .agents/plans/*.md 2>/dev/null | head -1)
if [[ -n "$PLAN_FILE" ]] && grep -q "## Boundaries" "$PLAN_FILE"; then
    # Extract "Always" boundaries and convert to cross_cutting checks
    # Read the plan's ## Cross-Cutting Constraints section or derive from ## Boundaries
    # Inject into every TaskCreate's metadata.validation.cross_cutting
fi
# "Ask First" boundaries: in auto mode, log as annotation only (no blocking)
# In --interactive mode, prompt before proceeding
```

When creating TaskCreate for each wave issue, include cross-cutting constraints in metadata:
```json
{
  "validation": {
    "files_exist": [...],
    "content_check": {...},
    "cross_cutting": [
      {"name": "...", "type": "content_check", "file": "...", "pattern": "..."}
    ]
  }
}
```

**For wave execution details (beads sync, TaskList bridging, swarm invocation), read `skills/crank/references/team-coordination.md`.**

**Cross-cutting validation (SDD):**

After per-task validation passes, run cross-cutting checks across all files modified in the wave:

```bash
# Only if cross_cutting constraints were injected
if [[ -n "$CROSS_CUTTING_CHECKS" ]]; then
    WAVE_FILES=$(git diff --name-only "${WAVE_START_SHA}..HEAD")
    for check in $CROSS_CUTTING_CHECKS; do
        run_validation_check "$check" "$WAVE_FILES"
    done
fi
```

### Step 5: Verify and Sync to Beads (MANDATORY)

> Swarm executes per-task validation (see `skills/shared/validation-contract.md`). Crank trusts swarm validation and focuses on beads sync.

**For verification details, retry logic, and failure escalation, read `skills/crank/references/team-coordination.md` and `skills/crank/references/failure-recovery.md`.**

### Step 5.5: Wave Vibe Gate (MANDATORY)

> **Principle:** Fresh context catches what saturated context misses. No self-grading.

**For wave vibe gate details (diff computation, acceptance criteria, verdict gating), read `skills/crank/references/wave-patterns.md`.**

### Step 5.7: Wave Checkpoint

After each wave completes (post-vibe-gate, pre-next-wave), write a checkpoint file:

```bash
mkdir -p .agents/crank

cat > ".agents/crank/wave-${wave}-checkpoint.json" <<EOF
{
  "wave": ${wave},
  "timestamp": "$(date -Iseconds)",
  "tasks_completed": $(echo "$COMPLETED_IDS" | jq -R 'split(" ")'),
  "tasks_failed": $(echo "$FAILED_IDS" | jq -R 'split(" ")'),
  "files_changed": $(git diff --name-only "${WAVE_START_SHA}..HEAD" | jq -R . | jq -s .),
  "git_sha": "$(git rev-parse HEAD)"
}
EOF
```

- `COMPLETED_IDS` / `FAILED_IDS`: space-separated issue IDs from the wave results.
- On retry of the same wave, the file is overwritten (same path).
- Checkpoint files are informational — no resume logic reads them yet (future work).

### Step 6: Check for More Work

After completing a wave, check for newly unblocked issues (beads: `bd ready`, TaskList: `TaskList()`). Loop back to Step 4 if work remains, or proceed to Step 7 when done.

**For detailed check/retry logic, read `skills/crank/references/team-coordination.md`.**

### Step 7: Final Batched Validation

When all issues complete, run ONE comprehensive vibe on recent changes. Fix CRITICAL issues before completion.

**For detailed validation steps, read `skills/crank/references/failure-recovery.md`.**

### Step 8: Extract Learnings (ao Integration)

If ao CLI available: run `ao forge transcript`, `ao flywheel status`, and `ao pool list --tier=pending` to extract and review learnings. If ao unavailable, skip and recommend `/post-mortem` manually.

### Step 9: Report Completion

Tell the user:
1. Epic ID and title
2. Number of issues completed
3. Total iterations used (of 50 max)
4. Final vibe results
5. Flywheel status (if ao available)
6. Suggest running `/post-mortem` to review and promote learnings

**Output completion marker:**
```
<promise>DONE</promise>
Epic: <epic-id>
Issues completed: N
Iterations: M/50
Flywheel: <status from ao flywheel status>
```

If stopped early:
```
<promise>BLOCKED</promise>
Reason: <global limit reached | unresolvable blockers>
Issues remaining: N
Iterations: M/50
```

## The FIRE Loop

Crank follows FIRE (Find → Ignite → Reap → Vibe → Escalate) for each wave. Loop until all issues are CLOSED (beads) or all tasks are completed (TaskList).

**For FIRE loop details, parallel wave models, and wave vibe gate, read `skills/crank/references/wave-patterns.md`.**

## Key Rules

- **Auto-detect tracking** - check for `bd` at start; use TaskList if absent
- **Plan files as input** - `/crank plan.md` decomposes plan into tasks automatically
- **If epic ID given, USE IT** - don't ask for confirmation (beads mode only)
- **Swarm for each wave** - delegates parallel execution to swarm
- **Fresh context per issue** - swarm provides Ralph pattern isolation
- **Batch validation at end** - ONE vibe at the end saves context
- **Fix CRITICAL before completion** - address findings before reporting done
- **Loop until done** - don't stop until all issues closed / tasks completed
- **Autonomous execution** - minimize human prompts
- **Respect wave limit** - STOP at 50 waves (hard limit)
- **Output completion markers** - DONE, BLOCKED, or PARTIAL (required)
- **Knowledge flywheel** - load learnings at start, forge at end (ao optional)
- **Beads ↔ TaskList sync** - in beads mode, crank bridges beads issues to TaskList for swarm

---

## Distributed Mode: Agent Mail Orchestration (Experimental)

> **Status: Experimental.** Local mode (TaskList + swarm) is the recommended execution method.

**For distributed mode details (architecture, execution steps, Chiron pattern, file reservations, checkpoint handling), read `skills/crank/references/distributed-mode.md`.**

---

## References

- **Wave patterns:** `skills/crank/references/wave-patterns.md`
- **Team coordination:** `skills/crank/references/team-coordination.md`
- **Failure recovery:** `skills/crank/references/failure-recovery.md`
- **Distributed mode:** `skills/crank/references/distributed-mode.md`
- **Agent Mail Protocol:** `skills/shared/agent-mail-protocol.md`
- **Parser (Go):** `cli/internal/agentmail/`
