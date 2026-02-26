---
name: crank
description: 'Hands-free epic execution. Runs until ALL children are CLOSED. Uses $swarm with runtime-native spawning (Codex sub-agents or Claude teams). NO human prompts, NO stopping. Triggers: "crank", "run epic", "execute epic", "run all tasks", "hands-free execution", "crank it".'
---


# Crank Skill

> **Quick Ref:** Autonomous epic execution. `$swarm` for each wave with runtime-native spawning. Output: closed issues + final vibe.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Autonomous execution: implement all issues until the epic is DONE.

**CLI dependencies:** bd (issue tracking), ao (knowledge flywheel). Both optional — see `skills/shared/SKILL.md` for fallback table. If bd is unavailable, use TaskList for issue tracking and skip beads sync. If ao is unavailable, skip knowledge injection/extraction.

For Claude runtime feature coverage (agents/hooks/worktree/settings), see `..$shared/references/claude-code-latest-features.md`.

## Architecture: Crank + Swarm

**Beads mode** (bd available):
```
Crank (orchestrator)           Swarm (executor)
    |                              |
    +-> bd ready (wave issues)     |
    |                              |
    +-> TaskCreate from beads  --->+-> Select spawn backend (codex sub-agents | claude teams | fallback)
    |                              |
    +-> $swarm                 --->+-> Spawn workers per backend
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
    +-> $swarm                         --->+-> Select spawn backend per wave
    |                                      |
    +-> Verify via TaskList()          <---+-> Workers report via backend channel
    |                                      |
    +-> Loop until all completed       <---+-> Cleanup backend resources after wave
```

**Separation of concerns:**
- **Crank** = Orchestration, epic/task lifecycle, knowledge flywheel
- **Swarm** = Runtime-native parallel execution (Ralph Wiggum pattern via fresh worker set per wave)

Ralph alignment source: `..$shared/references/ralph-loop-contract.md` (fresh context, scheduler/worker split, disk-backed state, backpressure).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--test-first` | off | Enable spec-first TDD: SPEC WAVE generates contracts, TEST WAVE generates failing tests, IMPL WAVES make tests pass |
| `--per-task-commits` | off | Opt-in per-task commit strategy. Falls back to wave-batch when file boundaries overlap. See `references/commit-strategies.md`. |

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

Given `$crank [epic-id | plan-file.md | "description"]`:

### Step 0: Load Knowledge Context (ao Integration)

**Search for relevant learnings before starting the epic:**

```bash
# If ao CLI available, pull relevant knowledge for this epic
if command -v ao &>/dev/null; then
    # Pull knowledge scoped to the epic
    ao know lookup --query "<epic-title>" --limit 5 2>/dev/null || \
      ao know search "epic execution implementation patterns" 2>/dev/null | head -20

    # Check flywheel status
    ao quality flywheel status 2>/dev/null

    # Get current ratchet state
    ao work ratchet status 2>/dev/null
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
    # Classify issues by type
    # spec-eligible: feature, bug, task → SPEC + TEST waves apply
    # skip: chore, epic, docs → standard implementation waves only
    SPEC_ELIGIBLE=()
    SPEC_SKIP=()

    if [[ "$TRACKING_MODE" == "beads" ]]; then
        for issue in $READY_ISSUES; do
            ISSUE_TYPE=$(bd show "$issue" 2>/dev/null | grep "Type:" | head -1 | awk '{print tolower($NF)}')
            case "$ISSUE_TYPE" in
                feature|bug|task) SPEC_ELIGIBLE+=("$issue") ;;
                chore|epic|docs) SPEC_SKIP+=("$issue") ;;
                *)
                    echo "WARNING: Issue $issue has unknown type '$ISSUE_TYPE'. Defaulting to spec-eligible."
                    SPEC_ELIGIBLE+=("$issue")
                    ;;
            esac
        done
    else
        # TaskList mode: no bd available, default all to spec-eligible
        SPEC_ELIGIBLE=($READY_ISSUES)
        echo "TaskList mode: all ${#SPEC_ELIGIBLE[@]} issues defaulted to spec-eligible (no bd type info)"
    fi
    echo "Test-first mode: ${#SPEC_ELIGIBLE[@]} spec-eligible, ${#SPEC_SKIP[@]} skipped (chore/epic/docs)"
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
   - Epic has no child issues (run $plan first)
   - All issues already completed"
```

Also verify: epic has at least 1 child issue total. An epic with 0 children means $plan was not run.

Do NOT proceed with empty issue list - this produces false "epic complete" status.

### Step 3a.1: Pre-flight Check - Pre-Mortem Required (3+ issues)

**If the epic has 3 or more child issues, require pre-mortem evidence before proceeding.**

```bash
# Count child issues (beads mode)
if [[ "$TRACKING_MODE" == "beads" ]]; then
    CHILD_COUNT=$(bd show "$EPIC_ID" 2>/dev/null | grep -c "↳")
else
    CHILD_COUNT=$(TaskList | grep -c "pending\|in_progress")
fi

if [[ "$CHILD_COUNT" -ge 3 ]]; then
    # Look for pre-mortem report in .agents/council/
    PRE_MORTEM=$(ls -t .agents/council/*pre-mortem* 2>/dev/null | head -1)
    if [[ -z "$PRE_MORTEM" ]]; then
        echo "STOP: Epic has $CHILD_COUNT issues but no pre-mortem evidence found."
        echo "Run '$pre-mortem' first to validate the plan before cranking."
        echo "<promise>BLOCKED</promise>"
        echo "Reason: pre-mortem required for epics with 3+ issues"
        # STOP - do not continue
        exit 1
    fi
    echo "Pre-mortem evidence found: $PRE_MORTEM"
fi
```

**Why:** 7 consecutive epics (ag-oke through ag-9ad) showed positive ROI from pre-mortem validation. For epics with 3+ issues, the cost of a pre-mortem (~2 min) is negligible compared to the cost of cranking a flawed plan.

### Step 3b: SPEC WAVE (--test-first only)

**Skip if `--test-first` is NOT set or if no spec-eligible issues exist.**

For each spec-eligible issue (feature/bug/task):
1. **TaskCreate** with subject `SPEC: <issue-title>`
2. Worker receives: issue description, plan boundaries, contract template (`skills/crank/references/contract-template.md`), codebase access (read-only)
3. Worker generates: `.agents/specs/contract-<issue-id>.md`
4. **Validation:** files_exist + content_check for `## Invariants` AND `## Test Cases`
5. **Wave 1 spec consistency checklist (MANDATORY):** run `skills/crank/references/wave1-spec-consistency-checklist.md` across all contracts in this wave. If any item fails, re-run SPEC workers for affected issues and do NOT proceed to TEST WAVE.
6. Lead commits all specs after validation

For BLOCKED recovery and full worker prompt, read `skills/crank/references/test-first-mode.md`.

### Step 3c: TEST WAVE (--test-first only)

**Skip if `--test-first` is NOT set or if no spec-eligible issues exist.**

For each spec-eligible issue:
1. **TaskCreate** with subject `TEST: <issue-title>`
2. Worker receives: contract-<issue-id>.md + codebase types (NOT implementation code)
3. Worker generates: failing test files in appropriate location
4. **RED Gate:** Lead runs test suite — ALL new tests must FAIL
5. Lead commits test harness after RED Gate passes

For RED Gate enforcement and retry logic, read `skills/crank/references/test-first-mode.md`.

**Summary:** SPEC WAVE generates contracts from issues → TEST WAVE generates failing tests from contracts → RED Gate verifies all new tests fail before proceeding. Docs/chore/ci issues bypass both waves.

### Step 3b.1: Build Context Briefing (Before Worker Dispatch)

```bash
if command -v ao &>/dev/null; then
    ao work context assemble --task='<epic title>: wave $wave'
fi
```

This produces a 5-section briefing (GOALS, HISTORY, INTEL, TASK, PROTOCOL) at `.agents/rpi/briefing-current.md` with secrets redacted. Include the briefing path in each worker's TaskCreate description so workers start with full project context.

### Step 4: Execute Wave via Swarm

**GREEN mode (--test-first only):** If `--test-first` is set and SPEC/TEST waves have completed, modify worker prompts for spec-eligible issues:
- Include in each worker's TaskCreate: `"Failing tests exist at <test-file-paths>. Make them pass. Do NOT modify test files. See GREEN Mode rules in $implement SKILL.md."`
- Workers receive: failing tests (immutable), contract, issue description
- Workers follow GREEN Mode rules from `$implement` SKILL.md
- Docs/chore/ci issues (skipped by SPEC/TEST waves) use standard worker prompts unchanged

**File manifests (REQUIRED):** Include a `metadata.files` array in every TaskCreate listing the files that worker will modify. This feeds into swarm's pre-spawn conflict detection -- two workers claiming the same file in the same wave get serialized or worktree-isolated automatically. Derive file lists from the issue description, plan, or codebase exploration during planning.

```
TaskCreate(
  subject="ag-1234: Add auth middleware",
  description="...",
  activeForm="Implementing ag-1234",
  metadata={
    "files": ["src/middleware/auth.py", "tests/test_auth.py"],
    "validation": {
      "tests": "pytest tests/test_auth.py -v",
      "files_exist": ["src/middleware/auth.py", "tests/test_auth.py"]
    }
  }
)
```

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

### Step 5.5: Wave Acceptance Check (MANDATORY)

> **Principle:** Verify each wave meets acceptance criteria using lightweight inline judges. No skill invocations — prevents context explosion in the orchestrator loop.

**For acceptance check details (diff computation, inline judges, verdict gating), read `skills/crank/references/wave-patterns.md`.**

### Step 5.7: Wave Checkpoint

After each wave completes (post-vibe-gate, pre-next-wave), write a checkpoint file:

```bash
mkdir -p .agents/crank

cat > ".agents/crank/wave-${wave}-checkpoint.json" <<EOF
{
  "schema_version": 1,
  "wave": ${wave},
  "timestamp": "$(date -Iseconds)",
  "tasks_completed": $(echo "$COMPLETED_IDS" | jq -R 'split(" ")'),
  "tasks_failed": $(echo "$FAILED_IDS" | jq -R 'split(" ")'),
  "files_changed": $(git diff --name-only "${WAVE_START_SHA}..HEAD" | jq -R . | jq -s .),
  "git_sha": "$(git rev-parse HEAD)",
  "acceptance_verdict": "<PASS|WARN|FAIL>",
  "commit_strategy": "<per-task|wave-batch|wave-batch-fallback>"
}
EOF
```

- `COMPLETED_IDS` / `FAILED_IDS`: space-separated issue IDs from the wave results.
- `acceptance_verdict`: verdict from the Wave Acceptance Check (Step 5.5). Used by final validation to skip redundant $vibe on clean epics.
- On retry of the same wave, the file is overwritten (same path).

### Step 6: Check for More Work

After completing a wave, check for newly unblocked issues (beads: `bd ready`, TaskList: `TaskList()`). Loop back to Step 4 if work remains, or proceed to Step 7 when done.

**For detailed check/retry logic, read `skills/crank/references/team-coordination.md`.**

### Step 7: Final Batched Validation

When all issues complete, run ONE comprehensive vibe on recent changes. Fix CRITICAL issues before completion.

If hooks or `lib/hook-helpers.sh` were modified, verify embedded copies are in sync: `cd cli && make sync-hooks`.

**For detailed validation steps, read `skills/crank/references/failure-recovery.md`.**

### Step 8: Extract Learnings (ao Integration)

If ao CLI available: run `ao know forge transcript`, `ao quality flywheel close-loop --quiet`, `ao quality flywheel status`, and `ao quality pool list --status=pending` to extract and review learnings. If ao unavailable, skip and recommend `$post-mortem` manually.

### Step 9: Report Completion

Tell the user:
1. Epic ID and title
2. Number of issues completed
3. Total iterations used (of 50 max)
4. Final vibe results
5. Flywheel status (if ao available)
6. Suggest running `$post-mortem` to review and promote learnings

**Output completion marker:**
```
<promise>DONE</promise>
Epic: <epic-id>
Issues completed: N
Iterations: M/50
Flywheel: <status from ao quality flywheel status>
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

**For FIRE loop details, parallel wave models, and wave acceptance check, read `skills/crank/references/wave-patterns.md`.**

## Key Rules

- **Auto-detect tracking** - check for `bd` at start; use TaskList if absent
- **Plan files as input** - `$crank plan.md` decomposes plan into tasks automatically
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

## Examples

### Execute Epic with Beads Tracking

**User says:** `$crank ag-m0r`

Loads learnings (`ao know inject`), gets epic details (`bd show`), finds unblocked issues (`bd ready`), creates TaskList, invokes `$swarm` per wave with runtime-native spawning. Workers execute in parallel; lead verifies, commits per wave. Loops until all issues closed, then batched vibe + `ao know forge transcript`.

### Execute from Plan File (TaskList Mode)

**User says:** `$crank .agents/plans/auth-refactor.md`

Reads plan file, decomposes into TaskList tasks with dependencies. Invokes `$swarm` per wave, lead verifies and commits. Loops until all tasks completed, then final vibe.

### Test-First Epic with Contract-Based TDD

**User says:** `$crank --test-first ag-xj9`

Runs: classify issues → SPEC WAVE (contracts) → TEST WAVE (failing tests, no impl access) → RED Gate (tests must fail) → GREEN IMPL WAVES (make tests pass) → final vibe. See `skills/crank/references/test-first-mode.md`.

### Recovery from Blocked State

If all remaining issues are blocked (e.g., circular dependencies), crank outputs `<promise>BLOCKED</promise>` with the blocking chains and exits cleanly. See `skills/crank/references/failure-recovery.md`.

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "No ready issues found" | Epic has no children or all blocked | Run `$plan` first or check deps with `bd show <id>` |
| "Global wave limit (50) reached" | Excessive retries or circular deps | Review `.agents/crank/wave-N-checkpoint.json`, fix blockers manually |
| Wave vibe gate fails repeatedly | Workers producing non-conforming code | Check `.agents/council/` vibe reports, refine constraints |
| Workers complete but files missing | Permission errors or wrong paths | Check swarm output files, verify write permissions |
| RED Gate passes (tests don't fail) | Test wave workers wrote implementation | Re-run TEST WAVE with no-implementation-access prompt |
| TaskList mode can't find epic | bd CLI required for beads tracking | Provide plan file (`.md`) instead, or install bd |

See `skills/crank/references/troubleshooting.md` for extended troubleshooting.

---

## References

- **Wave patterns:** `skills/crank/references/wave-patterns.md`
- **Team coordination:** `skills/crank/references/team-coordination.md`
- **Failure recovery:** `skills/crank/references/failure-recovery.md`
- **Failure Taxonomy:** `references/failure-taxonomy.md`
- **FIRE Protocol:** `references/fire.md`

## Reference Documents

- [references/claude-code-latest-features.md](references/claude-code-latest-features.md)
- [references/commit-strategies.md](references/commit-strategies.md)
- [references/contract-template.md](references/contract-template.md)
- [references/failure-recovery.md](references/failure-recovery.md)
- [references/failure-taxonomy.md](references/failure-taxonomy.md)
- [references/fire.md](references/fire.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)
- [references/taskcreate-examples.md](references/taskcreate-examples.md)
- [references/team-coordination.md](references/team-coordination.md)
- [references/test-first-mode.md](references/test-first-mode.md)
- [references/troubleshooting.md](references/troubleshooting.md)
- [references/wave1-spec-consistency-checklist.md](references/wave1-spec-consistency-checklist.md)
- [references/wave-patterns.md](references/wave-patterns.md)

---

## References

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

### commit-strategies.md

# Commit Strategies

Crank supports two commit strategies for how changes are committed after task completion.

## wave-batch (default)

The team lead commits once after all workers in a wave have completed and passed validation.

**Pros:**
- No merge conflicts (single committer)
- Clean git history (one commit per wave)
- Proven pattern across 7+ epics

**Cons:**
- Coarse bisectability (entire wave in one commit)
- Harder to attribute changes to specific issues

**Commit message format:** `feat(<epic-id>): wave N - <summary of changes>`

## per-task (opt-in via `--per-task-commits`)

Workers commit after their individual task passes validation.

**Pros:**
- Fine-grained git bisect (one commit per issue)
- Per-issue traceability in git history
- Better attribution

**Cons:**
- Merge conflict risk when multiple workers modify overlapping files
- Requires parallel-wave guard for safety

**Commit message format:** `feat(<issue-id>): <issue-title>`

## Parallel-Wave Guard (mandatory for per-task)

When a wave has 2+ workers modifying overlapping files, per-task commits are automatically disabled for that wave:

1. Before wave start, check file boundaries from plan/task metadata
2. **If file boundaries are absent** for any worker in a multi-worker wave → fall back to wave-batch (safe default). Only allow per-task when ALL workers have explicit boundary declarations.
3. If any file appears in 2+ workers' boundaries → fall back to wave-batch
3. Log: "Per-task commits disabled for wave N (overlapping file boundaries: <files>). Using wave-batch."
4. Record fallback in wave checkpoint JSON: `"commit_strategy": "wave-batch-fallback"`

Single-worker waves are always safe for per-task commits (no conflict possible).

## State Tracking

When `--per-task-commits` is active:
- `crank_state.per_task_commits = true`
- Each wave checkpoint includes: `"commit_strategy": "per-task" | "wave-batch" | "wave-batch-fallback"`

### contract-template.md

# Contract Template

> One contract per issue. Spec workers fill this out before implementation begins.

---

```yaml
# --- Contract Frontmatter ---
issue:      # e.g., ag-abc.3
framework:  # go | python | typescript | rust | shell
category:   # feature | bugfix | refactor | docs | chore | ci
```

---

## Problem

<!-- 1-2 sentences. What is broken, missing, or suboptimal? -->

## Inputs

<!-- Bullet list: name, type, description -->

- `inputName` (type) — description

## Outputs

<!-- Bullet list: name, type, description -->

- `outputName` (type) — description

## Invariants

<!-- Numbered list. Minimum 3. These are properties that must ALWAYS hold. -->

1. ...
2. ...
3. ...

## Failure Modes

<!-- Numbered list: what could go wrong → expected behavior -->

1. **Condition** → expected behavior
2. **Condition** → expected behavior

## Out of Scope

<!-- Explicitly excluded items — prevents scope creep -->

- ...

## Test Cases

<!-- Map each test case to an invariant. Cover: boundaries, errors, success path. -->

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | ... | ... | #1 |
| 2 | ... | ... | #2 |
| 3 | ... | ... | #3 |

## Contract Granularity

- **1 contract per issue.** Do not combine multiple issues into one contract.
- **Test boundaries, errors, and success paths.** Every contract must have at least one test case for each category.
- **Acceptance criteria are user-facing.** Invariants describe system properties; acceptance criteria describe what the user sees. Keep them separate — invariants go here, acceptance criteria go in the issue.

---

# EXAMPLE: Add Rate Limiting Middleware

```yaml
issue:      ag-xyz.5
framework:  go
category:   feature
```

## Problem

API endpoints accept unlimited requests per client, enabling abuse and risking resource exhaustion under load.

## Inputs

- `request` (*http.Request) — incoming HTTP request with client IP in RemoteAddr
- `config.RateLimit` (int) — max requests per window per client (default: 100)
- `config.RateWindow` (time.Duration) — sliding window duration (default: 1 minute)

## Outputs

- **Pass-through** — request forwarded to next handler with `X-RateLimit-Remaining` header
- **429 response** — JSON error body `{"error": "rate limit exceeded", "retry_after": <seconds>}` with `Retry-After` header

## Invariants

1. A client sending ≤ `RateLimit` requests within `RateWindow` is never rejected.
2. A client exceeding `RateLimit` within `RateWindow` receives HTTP 429 for every subsequent request until the window expires.
3. Rate limit state for one client never affects another client's quota.
4. The middleware adds < 1ms p99 latency to the request path.
5. If the rate limit store is unavailable, requests pass through (fail-open) and an error is logged.

## Failure Modes

1. **Rate store unreachable** → fail-open, log error, increment `ratelimit_store_errors_total` metric.
2. **Malformed RemoteAddr** → treat as unknown client, apply default limit, log warning.
3. **Clock skew between instances** → accept up to 2× burst during window overlap (documented trade-off).

## Out of Scope

- Distributed rate limiting across multiple instances (future work).
- Per-endpoint rate limits (all endpoints share the same limit).
- Authentication-aware rate limiting (keyed by IP only).

## Test Cases

| # | Input | Expected | Validates Invariant |
|---|-------|----------|---------------------|
| 1 | 100 requests from same IP in 60s | All 100 return 200 | #1 — at-limit success |
| 2 | 101st request from same IP in 60s | Returns 429 with `Retry-After` header | #2 — over-limit rejection |
| 3 | 100 requests from IP-A, 100 from IP-B | All 200 return 200 | #3 — client isolation |
| 4 | Request when rate store returns error | Returns 200, error logged | #5 — fail-open behavior |
| 5 | Wait for window expiry after 429 | Next request returns 200 | #2 — window reset |
| 6 | Benchmark 10k requests | p99 < 1ms overhead | #4 — latency bound |

### failure-recovery.md

# Failure Recovery

## Validation Failure Handling

**On swarm validation failure:**

1. Do NOT close the beads issue
2. Add failure context:
   ```bash
   bd comments add <issue-id> "Validation failed: <reason>. Retrying..." 2>/dev/null
   ```
3. Re-add to next wave
4. After 3 failures, escalate:
   ```bash
   bd update <issue-id> --labels BLOCKER 2>/dev/null
   bd comments add <issue-id> "ESCALATED: 3 validation failures. Human review required." 2>/dev/null
   ```

## Wave Limit Enforcement

```bash
# CHECK GLOBAL LIMIT before each wave
if [[ $wave -ge 50 ]]; then
    echo "<promise>BLOCKED</promise>"
    echo "Global wave limit (50) reached. Remaining issues:"
    # Beads mode: bd children <epic-id> --status open
    # TaskList mode: TaskList() → pending tasks
    # STOP - do not continue
fi
```

## Pre-flight Check: Issues Exist

**Verify there are issues to work on:**

**If 0 ready issues found (beads mode) or 0 pending unblocked tasks (TaskList mode):**
```
STOP and return error:
  "No ready issues found for this epic. Either:
   - All issues are blocked (check dependencies)
   - Epic has no child issues (run $plan first)
   - All issues already completed"
```

Also verify: epic has at least 1 child issue total. An epic with 0 children means $plan was not run.

Do NOT proceed with empty issue list - this produces false "epic complete" status.

## Final Batched Validation

When all issues complete, check whether a full $vibe is needed:

```bash
# Check wave checkpoint verdicts — skip final vibe if ALL waves passed clean
ALL_PASS=true
for checkpoint in .agents/crank/wave-*-checkpoint.json; do
    verdict=$(jq -r '.acceptance_verdict // "UNKNOWN"' "$checkpoint" 2>/dev/null)
    if [[ "$verdict" != "PASS" ]]; then
        ALL_PASS=false
        break
    fi
done
```

**If ALL waves passed acceptance check with PASS verdict (no WARNs, no retries):**
Skip the final $vibe — per-wave acceptance checks already validated acceptance criteria. Proceed directly to Step 8 (learnings extraction).

**If ANY wave had WARN, FAIL, or missing verdicts:**
Run ONE comprehensive vibe on recent changes:

```bash
# Get list of changed files from recent commits
git diff --name-only HEAD~10 2>/dev/null | sort -u
```

```
Tool: Skill
Parameters:
  skill: "agentops:vibe"
  args: "recent"
```

**If CRITICAL issues found:**
1. Fix them
2. Re-run vibe on affected files
3. Only proceed to completion when clean

## Retry Strategy

| Failure Type | Action |
|--------------|--------|
| Validation failure | Re-add to next wave (max 3 attempts) |
| Blocked dependencies | Escalate after 3 checks |
| Context exhaustion (distributed) | Checkpoint + spawn replacement |
| Build failure | Re-add to retry queue |
| Spec impossible | Mark blocked, escalate immediately |

## Escalation

When issues cannot be resolved automatically:
- Mark with BLOCKER label (beads mode)
- Output `<promise>BLOCKED</promise>` with reason
- List remaining issues for human review

### failure-taxonomy.md

# Failure Taxonomy

> **Classification and handling of failures in autonomous execution.**

## Overview

Failures in crank execution fall into distinct categories, each with specific detection methods and remediation strategies. The goal is always: **continue the epic, escalate what can't be fixed**.

## Failure Categories

### 1. Polecat Stuck

**Symptoms**:
- No status change for 5+ poll intervals (2.5 min)
- Convoy shows `running` but no progress
- tmux pane shows idle prompt or repeated error

**Detection**:

```bash
# Check convoy status history
gt convoy status <id> --history               # FUTURE: gt convoy not yet implemented

# Peek at polecat
tmux capture-pane -t gt-<rig>-<polecat> -p | tail -30
```

**Causes**:
- Waiting for user input
- Infinite loop in code
- External service timeout
- Claude usage limit hit

**Remediation**:

```bash
# Step 1: Nudge the polecat
tmux send-keys -t gt-<rig>-<polecat> "continue with your assigned task" Enter

# Step 2: Wait one poll interval (30s)

# Step 3: If still stuck, check for usage limit
tmux capture-pane -t gt-<rig>-<polecat> -p | grep -i "limit"

# Step 4: If usage limit, nuke and re-sling after cooldown
# WARNING: This destroys the polecat session. Ensure work is saved.
gt polecat nuke <rig>/<name> --force
# Wait for limit reset, then:
gt sling <issue> <rig>

# Step 5: If other cause, nuke and re-sling immediately
# WARNING: This destroys the polecat session. Ensure work is saved.
gt polecat nuke <rig>/<name> --force
gt sling <issue> <rig>
```

**Escalation trigger**: 3 consecutive nudge failures

---

### 2. Validation Failure

**Symptoms**:
- Polecat completes but issue not closed
- `.agents/validations/` contains failure artifacts
- Commit exists but tests/lint failing

**Detection**:

```bash
# Check polecat output
tmux capture-pane -t gt-<rig>-<polecat> -p | grep -i "fail\|error"

# Check validation artifacts
ls ./polecats/<polecat>/.agents/validations/

# Check CI if applicable
git -C ./polecats/<polecat> log -1 --format="%H" | xargs gh run list --commit
```

**Causes**:
- Tests failing
- Lint errors
- Type check failures
- Security scan findings
- Build failures

**Remediation**:

```bash
# Step 1: Add failure context to issue
bd comments add <issue> "Validation failed: $(cat validation-output.txt | head -50)"

# Step 2: Re-sling with hint
bd comments add <issue> "HINT: Focus on fixing <specific failure>"
gt sling <issue> <rig>

# Step 3: If second failure, be more specific
bd comments add <issue> "EXPLICIT: The test_auth_flow test fails because X. Fix by Y."
gt sling <issue> <rig>
```

**Escalation trigger**: 3 validation failures (may need human insight)

---

### 3. Dependency Deadlock

**Symptoms**:
- Multiple issues show as `blocked`
- No issues in `ready` state
- Circular dependency detected

**Detection**:

```bash
# Check for circular deps
bd blocked --parent=<epic> --show-deps

# Manual trace
bd show <issue-a> | grep "blocked by"
bd show <issue-b> | grep "blocked by"
# If A -> B -> A, deadlock exists
```

**Causes**:
- Incorrectly specified dependencies
- Missing issue that should break the cycle
- Overly aggressive blocking

**Remediation**:

```bash
# Step 1: Identify the cycle
bd dep graph <epic>  # Visual if available

# Step 2: Remove weakest dependency
bd dep remove <issue> <blocking-issue>

# Step 3: Add comment explaining
bd comments add <issue> "Removed dep on <blocking> to break deadlock.
May need manual integration after both complete."

# Step 4: Continue cranking
# Issues should now become ready
```

**Escalation trigger**: Immediate if auto-resolution fails

---

### 4. Context Limit

**Symptoms**:
- Polecat stops mid-work
- Message about "context limit" or "token limit"
- Partial work committed

**Detection**:

```bash
tmux capture-pane -t gt-<rig>-<polecat> -p | grep -i "context\|token\|limit"
```

**Causes**:
- Large files read into context
- Long conversation history
- Complex multi-file changes

**Remediation**:

```bash
# Step 1: Checkpoint current progress
git -C ./polecats/<polecat> stash  # If uncommitted work

# Step 2: Check what was accomplished
git -C ./polecats/<polecat> log --oneline -5

# Step 3: Update issue with progress
bd comments add <issue> "Partial progress: <what was done>. Remaining: <what's left>"

# Step 4: Fresh polecat
gt polecat nuke <rig>/<name> --force
gt sling <issue> <rig>

# The new polecat reads the comment and continues from there
```

**Escalation trigger**: 2 context limit failures (may need issue decomposition)

---

### 5. Git Conflict

**Symptoms**:
- Merge/rebase fails
- `.beads/` conflicts
- Branch divergence

**Detection**:

```bash
git -C ./polecats/<polecat> status | grep -i "conflict\|diverged"
```

**Causes**:
- Parallel work on same files
- Stale branch
- Beads sync race

**Remediation**:

```bash
# For beads conflicts (most common)
git -C ./polecats/<polecat> checkout --theirs .beads/issues.jsonl
git -C ./polecats/<polecat> add .beads/issues.jsonl
git -C ./polecats/<polecat> commit -m "merge: resolve beads conflict"

# For code conflicts
# Step 1: Check if conflict is trivial
git -C ./polecats/<polecat> diff --name-only --diff-filter=U

# Step 2: If simple, nudge polecat to resolve
tmux send-keys -t gt-<rig>-<polecat> "resolve the git conflicts and continue" Enter

# Step 3: If complex, abort and re-sling with fresh base
git -C ./polecats/<polecat> merge --abort
git -C ./polecats/<polecat> fetch origin
git -C ./polecats/<polecat> reset --hard origin/main
gt sling <issue> <rig>
```

**Escalation trigger**: 2 conflict failures on same files (architectural issue)

---

### 6. External Service Failure

**Symptoms**:
- Timeouts in polecat output
- API errors (429, 500, etc.)
- Network connectivity issues

**Detection**:

```bash
tmux capture-pane -t gt-<rig>-<polecat> -p | grep -i "timeout\|429\|500\|network\|connection"
```

**Causes**:
- Rate limiting
- Service outage
- Network partition
- API credential expiry

**Remediation**:

```bash
# Step 1: Identify the service
# (from polecat output)

# Step 2: Check service status
# (manual or via status page)

# Step 3: If rate limit, apply backoff
# Wait BACKOFF_BASE * 2^attempt before retry

# Step 4: If outage, pause affected issues
bd update <issue> --labels=WAITING_EXTERNAL
bd comments add <issue> "Paused: <service> outage. Resume when service recovers."

# Step 5: Continue other issues
# External failures shouldn't block entire epic
```

**Escalation trigger**: Immediate for credential issues, after recovery for outages

---

### 7. Polecat Crash

**Symptoms**:
- tmux session gone
- No polecat in `gt polecat list`
- Issue still shows `in_progress`

**Detection**:

```bash
gt polecat list <rig> | grep <polecat-name>
tmux has-session -t gt-<rig>-<polecat> 2>/dev/null && echo "exists" || echo "gone"
```

**Causes**:
- OOM kill
- Segfault in tooling
- System restart
- Manual termination

**Remediation**:

```bash
# Step 1: Clean up orphaned state
gt polecat nuke <rig>/<name> --force 2>/dev/null || true

# Step 2: Reset issue status
bd update <issue> --status=open

# Step 3: Add crash context
bd comments add <issue> "Previous polecat crashed. No partial work recovered."

# Step 4: Re-sling
gt sling <issue> <rig>
```

**Escalation trigger**: 2 crashes (may indicate systemic issue)

---

## Failure Handling Matrix

| Failure Type | Detection Cost | Auto-Recovery | Retry Limit | Escalation Action |
|--------------|----------------|---------------|-------------|-------------------|
| Polecat Stuck | ~200 tokens | Nudge, nuke | 3 | BLOCKER + mail |
| Validation Fail | ~150 tokens | Hints | 3 | BLOCKER + mail |
| Dependency Deadlock | ~100 tokens | Remove dep | 1 | Immediate mail |
| Context Limit | ~50 tokens | Checkpoint, re-sling | 2 | Decompose issue |
| Git Conflict | ~100 tokens | Auto-resolve | 2 | BLOCKER + mail |
| External Service | ~50 tokens | Backoff | 5 | WAITING label |
| Polecat Crash | ~50 tokens | Clean re-sling | 2 | Check system |

## Escalation Protocol

When MAX_RETRIES exhausted:

```bash
# 1. Mark issue as BLOCKER
bd update <issue> --labels=BLOCKER

# 2. Add detailed failure report
bd comments add <issue> "$(cat <<'EOF'
## AUTO-ESCALATION REPORT

**Issue**: <issue-id>
**Epic**: <epic-id>
**Failure Type**: <type>
**Attempts**: 3/3

### Attempt 1
- Polecat: <name>
- Duration: <time>
- Failure: <reason>

### Attempt 2
- Polecat: <name>
- Duration: <time>
- Failure: <reason>

### Attempt 3
- Polecat: <name>
- Duration: <time>
- Failure: <reason>

### Recommendation
<what human should investigate>
EOF
)"

# 3. Mail human (--human, not mayor/ since we ARE mayor)
gt mail send --human -s "BLOCKER: <issue> - <failure-type>" -m "See issue for details"

# 4. Continue epic (don't halt for one blocker)
# Other issues can still proceed
```

## Post-Failure Analysis

After epic completion (or major milestone), analyze failures:

```bash
# List all issues that required retries
bd list --parent=<epic> --has-label=BLOCKER
bd list --parent=<epic> --has-label=WAITING_EXTERNAL

# Check retry patterns
# (requires custom tooling or log analysis)

# Feed into retrospective
$retro --topic="crank failures on <epic>"
```

## Prevention Strategies

Based on failure patterns:

| Pattern | Prevention |
|---------|------------|
| Frequent context limits | Decompose large issues |
| Repeated validation fails | Add pre-validation to issues |
| Git conflicts | Smaller, focused changes |
| External service issues | Add circuit breaker patterns |
| Polecat crashes | Monitor system resources |

### fire.md

# FIRE Loop Specification

> **Find-Ignite-Reap-Escalate**: The Brownian Ratchet engine powering autonomous execution.

## Overview

FIRE is the reconciliation loop that extracts progress from chaos. Like a forge that transforms raw ore into refined steel, FIRE continuously drives an epic toward completion through parallel attempts filtered by validation.

**Design philosophy**: Chaos + Filter + Ratchet = Progress.

```
    ┌──────────────────────────────────────────────────────────┐
    │                       FIRE LOOP                           │
    │                                                           │
    │     FIND ────► IGNITE ────► REAP ────► ESCALATE          │
    │    (state)    (chaos)    (ratchet)   (recovery)          │
    │       │                                   │               │
    │       └───────────────────────────────────┘               │
    │                      (loop)                               │
    │                                                           │
    │     EXIT when: all children closed                        │
    └──────────────────────────────────────────────────────────┘
```

## The Brownian Ratchet

| Phase | Ratchet Role | Description |
|-------|--------------|-------------|
| **FIND** | Observe | Read current state, identify ready work |
| **IGNITE** | **Chaos** | Spark parallel polecats, embrace variance |
| **REAP** | **Filter + Ratchet** | Harvest results, validate, merge (permanent) |
| **ESCALATE** | Recovery | Handle failures, retry or escalate to human |

**Key insight**: Polecats can fail independently. Each successful merge ratchets forward. The system extracts progress from parallel attempts, filtering failures automatically.

---

## Loop Phases

### FIND Phase

**Purpose**: Build current state snapshot. What's ready? What's burning? What's done?

**Commands**:

```bash
bd ready --parent=<epic>                    # Ready to ignite
bd list --parent=<epic> --status=in_progress  # Currently burning
bd list --parent=<epic> --status=closed       # Reaped
bd blocked --parent=<epic>                    # Waiting on deps
gt convoy list                               # Active convoys  <!-- FUTURE: gt convoy not yet implemented -->
```

**State object**:

```yaml
fire_state:
  epic_id: gt-0100
  total_children: 8

  # Work pools
  ready: [gt-0101, gt-0102]      # Can ignite
  burning: [gt-0103, gt-0104]    # In-flight
  reaped: [gt-0105, gt-0106]     # Completed
  blocked: [gt-0107, gt-0108]    # Waiting

  # Derived
  remaining: 6                    # total - reaped
  capacity: 2                     # MAX_POLECATS - burning
  complete: false                 # remaining == 0
```

**Token cost**: ~200-300 tokens

---

### IGNITE Phase

**Purpose**: Spark parallel polecats. This is the CHAOS - multiple independent attempts.

**Decision logic**:

```python
def ignite_phase(state, retry_queue):
    to_ignite = []

    # Priority 1: Scheduled retries that are due
    for issue, scheduled_time in retry_queue:
        if now() >= scheduled_time:
            to_ignite.append(issue)
            retry_queue.remove(issue)

    # Priority 2: Fresh ready issues
    for issue in state.ready:
        if issue not in to_ignite:
            to_ignite.append(issue)

    # Respect capacity
    to_ignite = to_ignite[:state.capacity]

    # IGNITE - spark the chaos
    for issue in to_ignite:
        gt_sling(issue, rig)

    return to_ignite
```

**Commands**:

```bash
# Batch ignite - preferred (each issue gets own polecat)
gt sling <issue1> <issue2> <issue3> <rig>

# Single ignite
gt sling <issue> <rig>

# Find stranded convoys (ready work, no workers)
gt convoy stranded                           # FUTURE: gt convoy not yet implemented
```

**Token cost**: ~50 tokens per dispatch

---

### REAP Phase

**Purpose**: Harvest results. This is the FILTER + RATCHET - validate completions, merge permanently.

The REAP phase combines monitoring and collection into a single harvest operation:

1. **Monitor** - Poll for completion
2. **Validate** - Verify work quality (the FILTER)
3. **Merge** - Lock progress (the RATCHET)

**Monitoring**:

```bash
# Primary: Convoy dashboard (lowest token cost)
gt convoy status <convoy-id>                 # FUTURE: gt convoy not yet implemented

# Secondary: Individual polecat check
gt polecat status <rig>/<name>

# Tertiary: Peek at work (debugging only)
tmux capture-pane -t gt-<rig>-<polecat> -p | tail -20
```

**Poll interval**: 30 seconds

| Convoy Status | Meaning | Action |
|---------------|---------|--------|
| `running` | Polecats burning | Continue monitoring |
| `partial` | Some done | Reap completed, continue |
| `complete` | All done | Reap all |
| `failed` | Some failed | Reap successes, escalate failures |
| `stalled` | No progress 5+ polls | Investigate |

**Validation (the FILTER)**:

```python
def validate_completion(issue, polecat):
    """Filter: only valid completions ratchet forward."""

    # Check beads status
    status = bd_show(issue).status
    if status != 'closed':
        return False, "Status not closed"

    # Check git work exists
    commits = git_log(polecat_path, count=1)
    if not commits:
        return False, "No commits found"

    # Check commit references issue
    if issue not in commits[0].message:
        return False, "Commit doesn't reference issue"

    return True, "Validated"
```

**Merge (the RATCHET)**:

```bash
# Polecats self-merge via gt done:
# push → submit to merge queue → exit

# Post-merge cleanup
gt polecat gc <rig>  # Clean merged branches
```

**Key property**: Once merged, work is PERMANENT. The ratchet doesn't go backward.

**Token cost**: ~250 tokens per reap cycle

---

### ESCALATE Phase

**Purpose**: Handle failures with backoff and human escalation. Failed attempts re-enter the chaos pool or get escalated.

**Retry policy**:

| Attempt | Backoff | Action |
|---------|---------|--------|
| 1 | 30s | Re-ignite fresh polecat |
| 2 | 60s | Re-ignite with context |
| 3 | 120s | Re-ignite with explicit hints |
| 4+ | - | **ESCALATE**: BLOCKER + mail human |

**Backoff calculation**:

```python
def calculate_backoff(attempt):
    """Exponential backoff: 30s * 2^(attempt-1)"""
    return 30 * (2 ** (attempt - 1))
```

**Retry (back to chaos pool)**:

```bash
# Re-ignite with failure context
bd comments add <issue> "Previous attempt failed: <reason>. Try: <hint>"
gt sling <issue> <rig>
```

**Escalation (exit chaos pool)**:

```bash
# Mark as blocker
bd update <issue> --labels=BLOCKER

# Document failure history
bd comments add <issue> "AUTO-ESCALATED: Failed 3 attempts.
Reasons: 1) <reason1> 2) <reason2> 3) <reason3>
Human review required."

# Mail human
gt mail send --human -s "BLOCKER: <issue> failed 3 attempts" -m "..."

# Continue with other issues (don't halt epic)
```

**Token cost**: ~100 tokens per escalation

---

## State Machine

```
                    ┌─────────────────────────────────┐
                    │                                 │
                    ▼                                 │
┌────────┐    ┌─────────┐    ┌────────┐    ┌──────────┐
│  FIND  │───►│  IGNITE │───►│  REAP  │───►│ ESCALATE │
└────────┘    └─────────┘    └────────┘    └──────────┘
     │           chaos        ratchet          │
     │                                         │
     │ (all reaped)                            │
     ▼                                         │
┌────────┐                                     │
│  EXIT  │◄────────────────────────────────────┘
└────────┘         (retry scheduled)
```

---

## Loop Invariants

1. **Progress**: Each iteration must make progress OR escalate
2. **Bounded**: Retry counts are bounded, escalation is guaranteed
3. **Idempotent**: Re-running FIND produces same state for same beads
4. **Recoverable**: State can be reconstructed from beads alone
5. **Ratchet**: Merged work never goes backward

---

## Concurrency Model

**Single Mayor, Ephemeral Polecats**:

```
Mayor (FIRE Loop)
    │
    ├── Polecat 1 (burning gt-0101) → reaped → nuked
    ├── Polecat 2 (burning gt-0102) → reaped → nuked
    ├── Polecat 3 (burning gt-0103) → failed → escalated
    └── Polecat 4 (burning gt-0104) → reaped → nuked
```

**Polecat lifecycle**:
1. `gt sling` ignites polecat with hooked work
2. Polecat executes via `$implement`
3. On completion: `gt done` → push → merge queue → exit
4. Witness nukes sandbox after merge
5. No idle state - polecats don't wait

**Coordination via beads**:
- Mayor updates status via `bd update`
- Polecats work independently
- Status synced via `bd sync`

---

## Token Budget

Per FIRE iteration (30s):

| Phase | Tokens | Notes |
|-------|--------|-------|
| FIND | ~300 | bd queries |
| IGNITE | ~100 | gt sling commands |
| REAP | ~250 | monitoring + validation |
| ESCALATE | ~100 | if failures |
| **Total** | ~750 | per iteration |

**Per hour**: ~90,000 tokens (120 iterations)
**Per 8-hour run**: ~720,000 tokens

Sustainable for long-running autonomous execution.

---

## Error Recovery

**Mayor session crash**:
```bash
# State is in beads, not memory
$crank <epic> <rig>  # Resumes from beads state
```

**Polecat stalled**:
```bash
gt polecat stale <rig>                    # Find stale
gt polecat check-recovery <rig>/<name>    # Decide: recover | nuke
gt polecat nuke <rig>/<name> --force      # Destroy
gt sling <issue> <rig>                    # Re-ignite
```

**Beads sync conflict**:
```bash
git checkout --theirs .beads/issues.jsonl
git add .beads/issues.jsonl
bd sync
```

---

## Tuning Parameters

| Parameter | Default | Tuning Guidance |
|-----------|---------|-----------------|
| `MAX_POLECATS` | 4 | Increase for large epics, decrease for complex issues |
| `POLL_INTERVAL` | 30s | Decrease for fast issues, increase to save tokens |
| `MAX_RETRIES` | 3 | Increase for flaky tests, decrease for clean codebases |
| `BACKOFF_BASE` | 30s | Increase for rate-limited APIs |
| `STALL_THRESHOLD` | 5 polls | Decrease for tight deadlines |

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

### taskcreate-examples.md

# TaskCreate Examples

> Copy-paste-ready TaskCreate patterns for each crank mode.

---

## SPEC WAVE TaskCreate

Use when `--test-first` is set and issue is spec-eligible (feature/bugfix/refactor).

```
TaskCreate(
  subject="SPEC: <issue-title>",
  description="Generate contract for beads issue <issue-id>.

Details from beads:
<paste issue details from bd show>

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

```validation
files_exist:
  - .agents/specs/contract-<issue-id>.md
content_check:
  - file: .agents/specs/contract-<issue-id>.md
    patterns:
      - "## Invariants"
      - "## Test Cases"
```

Mark task complete when contract is written and validation passes.",
  activeForm="Writing spec for <issue-id>"
)
```

---

## TEST WAVE TaskCreate

Use when `--test-first` is set, SPEC WAVE is complete, and issue is spec-eligible.

```
TaskCreate(
  subject="TEST: <issue-title>",
  description="Generate FAILING tests for beads issue <issue-id>.

Details from beads:
<paste issue details from bd show>

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

```validation
files_exist:
  - <test-file-path-1>
  - <test-file-path-2>
red_gate:
  command: "<test-command>"
  expected: "FAIL"
  reason: "All new tests must fail (RED state) before implementation"
```

Mark task complete when tests are written and ALL tests FAIL.",
  activeForm="Writing tests for <issue-id>"
)
```

---

## GREEN Mode TaskCreate

Use when `--test-first` is set, SPEC and TEST waves are complete, and issue is spec-eligible.

```
TaskCreate(
  subject="<issue-id>: <issue-title>",
  description="Implement beads issue <issue-id> (GREEN mode).

Details from beads:
<paste issue details from bd show>

**GREEN Mode:** Failing tests exist. Make them pass. Do NOT modify test files.

Failing tests are at:
- <test-file-path-1>
- <test-file-path-2>

Contract is at: .agents/specs/contract-<issue-id>.md

Follow GREEN Mode rules from $implement SKILL.md:
1. Read failing tests and contract FIRST
2. Write minimal implementation to pass tests
3. Do NOT modify test files
4. Do NOT add tests (already written)
5. Validate by running test suite

Execute using $implement <issue-id>. Mark complete when all tests pass.",
  activeForm="Implementing <issue-id> (GREEN)"
)
```

---

## Standard IMPL TaskCreate

Use for non-spec-eligible issues (docs/chore/ci) or when `--test-first` is NOT set.

```
TaskCreate(
  subject="<issue-id>: <issue-title>",
  description="Implement beads issue <issue-id>.

Details from beads:
<paste issue details from bd show>

Execute using $implement <issue-id>. Mark complete when done.

```validation
<optional validation metadata specific to this issue>
build:
  command: "<build-command>"
  expected: "success"
tests:
  command: "<test-command>"
  expected: "pass"
files_exist:
  - <expected-output-file-1>
  - <expected-output-file-2>
```
",
  activeForm="Implementing <issue-id>",
  metadata={
    "files": ["<expected-modified-file-1>", "<expected-modified-file-2>"],
    "validation": {
      "tests": "<test-command>",
      "files_exist": ["<expected-output-file-1>"]
    }
  }
)
```

---

## Notes

- **Subject patterns:**
  - SPEC WAVE: `SPEC: <issue-title>` (no issue ID)
  - TEST WAVE: `TEST: <issue-title>` (no issue ID)
  - GREEN/IMPL: `<issue-id>: <issue-title>` (with issue ID)

- **Validation blocks:**
  - Fenced with triple backticks and `validation` language tag
  - Always include for SPEC and TEST waves
  - Optional but recommended for GREEN/IMPL waves
  - Consumed by lead during wave validation

- **activeForm:**
  - Shows in TaskList UI while worker is active
  - Keep concise (3-5 words)
  - Include issue ID for easy tracking

- **Worker context:**
  - SPEC: codebase read access, contract template
  - TEST: contract only, codebase structure (not implementations)
  - GREEN: failing tests (immutable), contract, issue description
  - IMPL: full codebase access, issue description

- **File manifests (`metadata.files`):**
  - **Required** for all TaskCreate entries — list every file the worker will modify
  - Swarm uses manifests for pre-spawn conflict detection (overlapping files = serialize or isolate)
  - Workers receive the manifest in their prompt and must stay within it
  - Derive from issue description, plan, or codebase exploration during planning

- **Category-based skipping:**
  - docs/chore/ci issues bypass SPEC and TEST waves
  - Use standard IMPL TaskCreate for these even if `--test-first` is set

### team-coordination.md

# Team Coordination

## Wave Execution via Swarm

### Beads Mode

1. **Get ready issues from current wave**
2. **Create TaskList tasks from beads issues:**

For each ready beads issue, create a corresponding TaskList task:
```
TaskCreate(
  subject="<issue-id>: <issue-title>",
  description="Implement beads issue <issue-id>.

Details from beads:
<paste issue details from bd show>

Execute using $implement <issue-id>. Mark complete when done.",
  activeForm="Implementing <issue-id>"
)
```

3. **Add dependencies if issues have beads blockedBy:**
```
TaskUpdate(taskId="2", addBlockedBy=["1"])
```

4. **Invoke swarm to execute the wave:**
```
Tool: Skill
Parameters:
  skill: "agentops:swarm"
```

5. **After swarm completes, verify beads status:**
```bash
# For each completed TaskList task, close the beads issue
bd update <issue-id> --status closed 2>/dev/null
```

### TaskList Mode

Tasks already exist in TaskList (created in Step 1 from plan file/description, or pre-existing). Just invoke swarm directly:

```
Tool: Skill
Parameters:
  skill: "agentops:swarm"
```

Swarm finds unblocked TaskList tasks and executes them.

### Both Modes — Swarm Will:

- Find all unblocked TaskList tasks
- Select runtime backend for the wave (runtime-native first: Claude sessions -> `TeamCreate`, Codex sessions -> `spawn_agent`, fallback tasks only if needed)
- Spawn workers with fresh context (Ralph pattern)
- Workers execute in parallel and report via backend channel (`wait`/`SendMessage`/`TaskOutput`)
- Team lead validates, then cleans up backend resources (`close_agent`/`TeamDelete`/none)

## Verify and Sync to Beads (MANDATORY)

> Swarm executes per-task validation (see `skills/shared/validation-contract.md`). Crank trusts swarm validation and focuses on beads sync.

**For each issue reported complete by swarm:**

1. **Verify swarm task completed:**
   ```
   TaskList() → check task status == "completed"
   ```
   If task is still pending/blocked, swarm validation failed — add to retry queue.

2. **Sync to beads:**
   ```bash
   bd update <issue-id> --status closed 2>/dev/null
   ```

3. **On sync failure** (bd unavailable or error):
   - Log warning but do NOT block the wave
   - Track for manual sync after epic completes

4. **Record ratchet progress (ao integration):**
   ```bash
   if command -v ao &>/dev/null; then
       ao work ratchet record implement 2>/dev/null
   fi
   ```

**Note:** Per-issue review is handled by swarm validation. Wave-level semantic review happens in the Wave Acceptance Check.

## Check for More Work

After completing a wave:

### Beads Mode
1. Clear completed tasks from TaskList
2. Check if new beads issues are now unblocked: `bd ready`
3. If yes, return to wave execution (create new TaskList tasks, invoke swarm)
4. If no more issues after 3 retry attempts, proceed to final validation

### TaskList Mode
1. `TaskList()` → any remaining pending tasks with no blockers?
2. If yes, loop back to wave execution
3. If all completed, proceed to final validation

### Both Modes
- **Max retries:** If issues remain blocked after 3 checks, escalate: "Epic blocked - cannot unblock remaining issues"

### test-first-mode.md

# Test-First Mode (--test-first)

> Reference for crank's `--test-first` flag. Covers SPEC WAVE, TEST WAVE, and RED Gate enforcement.

## SPEC WAVE

> **Purpose:** Generate contracts that ground implementation in verified requirements.

**Skip this step if `--test-first` is NOT set or if no spec-eligible issues exist.**

For each **spec-eligible** issue (feature/bugfix/refactor):

1. **TaskCreate** with subject `SPEC: <issue-title>`
2. **Worker prompt:**
   ```
   You are a spec writer. Generate a contract for this issue.

   FIRST: Explore the codebase to understand existing patterns, types, and interfaces
   relevant to this issue. Use Glob and Read to examine the code.

   THEN: Read the contract template at:
   skills/crank/references/contract-template.md

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
6. **Wave 1 consistency checklist (MANDATORY):** run `skills/crank/references/wave1-spec-consistency-checklist.md` across the full SPEC wave set before advancing to TEST WAVE.

**Category-based skip:** Issues categorized as docs/chore/ci bypass SPEC and TEST waves entirely and proceed directly to standard implementation waves.

### Wave 1 Consistency Gate

Run the checklist once per SPEC wave:

```bash
# Mechanical gate: all contracts in this wave satisfy checklist criteria
# (frontmatter completeness, invariant/test-case minimums, and consistency checks)
cat skills/crank/references/wave1-spec-consistency-checklist.md
```

If any checklist item fails:
1. Re-run SPEC worker(s) for affected issue(s)
2. Re-validate the full SPEC wave
3. Do not start TEST WAVE until checklist passes

### SPEC WAVE BLOCKED Recovery

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

## TEST WAVE

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

## RED Gate Enforcement

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

## Test Framework Detection

> Spec workers use this heuristic when the issue doesn't specify a test framework. First match wins.

**Detection priority (check in order, first match wins):**

| Priority | File Present | Check | Framework | Test Command | Contract `framework:` |
|----------|-------------|-------|-----------|-------------|----------------------|
| 1 | `Cargo.toml` | file exists | Rust | `cargo test` | `rust` |
| 2 | `go.mod` | file exists | Go | `go test ./...` | `go` |
| 3 | `pyproject.toml` or `pytest.ini` | file exists | pytest | `pytest` | `python` |
| 4 | `package.json` | `devDependencies.vitest` key exists | Vitest | `npx vitest run` | `typescript` |
| 5 | `package.json` | `devDependencies.jest` key exists | Jest | `npx jest` | `typescript` |
| 6 | `package.json` | file exists (no jest/vitest) | Node | `npm test` | `typescript` |
| 7 | `*.test.sh` or `tests/*.sh` | glob match | Shell | `bash <test-file>` | `shell` |

**For SPEC WAVE workers:** Detect the project framework using the heuristic above. Set `framework:` in the contract YAML frontmatter.

**For TEST WAVE workers:** Read the `framework:` field from the contract to determine which test runner to use. Generate tests following the project's existing test patterns.

**Fallback:** If no framework detected, spec worker writes `framework: unknown` and TEST WAVE skips that issue (falls back to standard IMPL without TDD).

**Polyglot repos:** If multiple frameworks match (e.g., Go backend + Node tooling), use the framework that matches the issue's target files. If ambiguous, use the highest-priority match.

### troubleshooting.md

# Crank Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "No ready issues found for this epic" | Epic has no child issues or all blocked | Run `$plan <epic-id>` first to decompose epic into issues. Check dependencies with `bd show <id>`. |
| "Global wave limit (50) reached" | Excessive retries or circular dependencies | Review failed waves in `.agents/crank/wave-N-checkpoint.json`. Fix blocking issues manually or break circular deps with `bd dep remove`. |
| Wave vibe gate fails repeatedly | Workers producing non-conforming code | Check `.agents/council/YYYY-MM-DD-vibe-wave-N.md` for specific findings. Add cross-cutting constraints to task metadata or refine worker prompts. |
| Workers report completion but files missing | Permission errors or workers writing to wrong paths | Check `.agents/swarm/<team>/worker-N-output.json` for file paths. Verify write permissions with `ls -ld`. |
| RED Gate passes (tests don't fail) | Test wave workers wrote implementation code | Re-run TEST WAVE with explicit "no implementation code access" in worker prompts. Tests must fail before GREEN waves start. |
| TaskList mode can't find epic ID | bd CLI required for beads epic tracking | Provide plan file path (`.md`) or task description string instead of epic ID. Or install bd CLI with `brew install bd`. |

### wave-patterns.md

# Wave Patterns

## The FIRE Loop

Crank follows FIRE for each wave:

| Phase | Beads Mode | TaskList Mode |
|-------|-----------|--------------|
| **FIND** | `bd ready` — get unblocked beads issues | `TaskList()` → pending, unblocked |
| **IGNITE** | TaskCreate from beads + `$swarm` | `$swarm` (tasks already in TaskList) |
| **REAP** | Swarm results + `bd update --status closed` | Swarm results (TaskUpdate by workers) |
| **CHECK** | Wave acceptance check (2 inline judges) → PASS/WARN/FAIL | Same |
| **ESCALATE** | `bd comments add` + retry | Update task description + retry |

**With `--test-first` flag, FIRE extends with two pre-implementation phases:**

| Phase | Description |
|-------|-------------|
| **SPEC** | Generate contracts per issue → `.agents/specs/contract-<id>.md` |
| **TEST** | Generate failing tests from contracts → RED gate (all must fail) |

## Parallel Wave Model

### Beads Mode

```
Wave 1: bd ready → [issue-1, issue-2, issue-3]
        ↓
        TaskCreate for each issue
        ↓
        $swarm → spawns 3 fresh-context agents
                  ↓         ↓         ↓
               DONE      DONE      BLOCKED
                                     ↓
                               (retry in next wave)
        ↓
        bd update --status closed for completed

Wave 2: bd ready → [issue-4, issue-3-retry]
        ↓
        TaskCreate for each
        ↓
        $swarm → spawns 2 fresh-context agents
        ↓
        bd update for completed

Final vibe on all changes → Epic DONE
```

### TaskList Mode

```
Wave 1: TaskList() → [task-1, task-2, task-3] (pending, unblocked)
        ↓
        $swarm → spawns 3 fresh-context agents
                  ↓         ↓         ↓
               DONE      DONE      BLOCKED
                                     ↓
                               (reset to pending, retry next wave)

Wave 2: TaskList() → [task-4, task-3-retry] (pending, unblocked)
        ↓
        $swarm → spawns 2 fresh-context agents
        ↓
        TaskUpdate → completed

Final vibe on all changes → All tasks DONE
```

Loop until all issues are CLOSED (beads) or all tasks are completed (TaskList).

## Spec-First Wave Model (--test-first)

When `--test-first` is enabled, crank runs 4 wave types instead of 1:

```
SPEC WAVE (conditional on --test-first)
  Workers: 1 per spec-eligible issue
  Input: issue description + plan boundaries + codebase (read-only)
  Output: .agents/specs/contract-{issue-id}.md
  Gate: Lead validates completeness (all issues have contracts)
                    ↓
TEST WAVE (conditional on --test-first)
  Workers: 1 per spec-eligible issue
  Input: contract-{issue-id}.md + codebase types (NOT implementation code)
  Output: test files committed to repo
  Gate: RED confirmation — ALL new tests must FAIL
                    ↓
IMPL WAVE (standard, enhanced with GREEN mode)
  Workers: 1 per issue (full access)
  Input: failing tests + contract + issue description
  Output: implementation code
  Gate: GREEN confirmation — ALL tests must PASS + wave acceptance check
                    ↓
[Optional] REFACTOR WAVE
  Workers: 1 per changed file group
  Input: passing tests + implementation
  Output: diff-only cleanup
  Gate: All tests still PASS
```

### Category-Based Skip

Issues categorized as docs, chore, or ci skip SPEC and TEST waves entirely:
- **feature / bugfix / refactor** → full pipeline (SPEC → TEST → IMPL)
- **docs / chore / ci** → standard implementation waves only

### RED Confirmation Gate

After TEST WAVE, the lead runs the test suite. ALL new tests must FAIL:
- If a new test passes → the test validates existing behavior, not new requirements
- Tests that pass are removed or flagged for rewrite
- Only proceed to IMPL when all new tests are confirmed RED

### RED Gate Failure Recovery

When the RED gate detects unexpected test passes:

1. **Identify cause:** Tests that pass against current code validate existing behavior, not new requirements from the contract
2. **Retry:** Re-spawn test writer with the unexpected-pass list and "must fail" constraint (max 2 retries)
3. **Escalate:** After 2 retries, mark the issue as BLOCKER and fall back to standard IMPL (no TDD for that issue)
4. **Log:** Record RED gate failure in wave checkpoint for post-mortem analysis

```bash
# RED gate failure tracking
if [[ ${#UNEXPECTED_PASSES[@]} -gt 0 ]]; then
    bd comments add <issue-id> "RED GATE: ${#UNEXPECTED_PASSES[@]} tests passed unexpectedly. Retry $RETRY_COUNT/2." 2>/dev/null
fi
```

### GREEN Confirmation Gate

After IMPL WAVE, the lead runs the test suite. ALL tests must PASS:
- New tests (from TEST WAVE) must now pass
- Existing tests must still pass (no regressions)
- Standard wave acceptance check also applies

### Contract Validation

SPEC WAVE workers explore the codebase before writing contracts (not fully isolated). This prevents generic, ungrounded specs. Workers read:
- Existing types, interfaces, and patterns
- Related test files for style reference
- Module structure and dependencies

But do NOT read implementation details of the specific feature being specified.

## Wave Acceptance Check (MANDATORY)

> **Principle:** Verify each wave meets acceptance criteria before advancing. Uses lightweight inline judges — no skill invocations, no context explosion.

**After closing all beads in a wave, before advancing to the next wave:**

**Note:** SPEC WAVE has its own validation (contract completeness check) and TEST WAVE has the RED gate. The Wave Acceptance Check applies only to IMPL and REFACTOR waves.

1. **Compute wave diff** (WAVE_START_SHA recorded in Step 4):
   ```bash
   git diff $WAVE_START_SHA HEAD --name-only
   WAVE_DIFF=$(git diff $WAVE_START_SHA HEAD)
   ```

2. **Load acceptance criteria** for all issues closed in this wave:
   ```bash
   # For each closed issue in the wave:
   bd show <issue-id>  # extract ACCEPTANCE CRITERIA section
   ```

3. **Spawn 2 inline judges** (Task agents, NOT skill invocations):

   ```
   # Judge 1: Spec compliance
   Tool: Task
   Parameters:
     subagent_type: "general-purpose"
     model: "haiku"
     description: "Wave N spec-compliance check"
     prompt: |
       Review this git diff against the acceptance criteria below.
       Does the implementation satisfy all acceptance criteria?
       Return: PASS, WARN (minor gaps), or FAIL (criteria not met) with brief justification.

       ## Acceptance Criteria
       <acceptance criteria from step 2>

       ## Git Diff
       <wave diff>

   # Judge 2: Error paths
   Tool: Task
   Parameters:
     subagent_type: "general-purpose"
     model: "haiku"
     description: "Wave N error-paths check"
     prompt: |
       Review this git diff for error handling and edge cases.
       Are error paths handled? Any unhandled exceptions or missing validations?
       Return: PASS, WARN (minor gaps), or FAIL (critical gaps) with brief justification.

       ## Git Diff
       <wave diff>
   ```

   **Dispatch both judges in parallel** (single message, 2 Task tool calls).

4. **Aggregate verdicts:**
   - Both PASS → **PASS**
   - Any FAIL → **FAIL**
   - Otherwise → **WARN**

5. **Gate on verdict:**

   | Verdict | Action |
   |---------|--------|
   | **PASS** | Record verdict in epic notes. Advance to next wave. |
   | **WARN** | Create fix beads as children of the epic (`bd create`). Execute fixes inline (small) or as wave N.5 via swarm. Re-run acceptance check. If PASS on re-check, advance. If still WARN after 2 attempts, treat as FAIL. |
   | **FAIL** | Record verdict in epic notes. Output `<promise>BLOCKED</promise>` and exit. Human review required. |

   ```bash
   # Record verdict in epic notes
   bd update <epic-id> --append-notes "CRANK_ACCEPT: wave=$wave verdict=<PASS|WARN|FAIL> at $(date -Iseconds)"
   ```

### wave1-spec-consistency-checklist.md

# Wave 1 Spec Consistency Checklist

Use this checklist after SPEC WAVE and before TEST WAVE.

## Required Per-Contract Checks

For every `contract-<issue-id>.md` produced in the current SPEC wave:

1. Frontmatter completeness:
- [ ] `issue` is present and matches the issue being implemented.
- [ ] `framework` is present (`go|python|typescript|rust|shell|unknown`).
- [ ] `category` is present (`feature|bugfix|refactor|docs|chore|ci`).

2. Structural completeness:
- [ ] `## Invariants` exists with at least 3 numbered invariants.
- [ ] `## Test Cases` exists with at least 3 rows.
- [ ] Every test case has a non-empty `Validates Invariant` value.

3. Implementability:
- [ ] Inputs/outputs reference concrete codebase concepts (not placeholders).
- [ ] Failure modes describe expected behavior, not only symptoms.

## Wave-Level Consistency Checks

Across all contracts in the wave:

1. Scope consistency:
- [ ] No contract combines multiple issue IDs.
- [ ] Each spec-eligible issue has exactly one contract.

2. Terminology consistency:
- [ ] Shared domain terms are used consistently between contracts.
- [ ] Conflicting invariants are resolved before TEST WAVE starts.

3. Test readiness:
- [ ] Every contract includes at least one success-path and one error-path test case.
- [ ] No contract is marked `BLOCKED` without a corresponding issue comment/escalation.

## Gate Rule

If any required item fails:
1. Re-run SPEC worker(s) for affected issue(s).
2. Re-run this checklist across the full wave.
3. Do NOT proceed to TEST WAVE until all required checks pass.


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
check "SKILL.md has name: crank" "grep -q '^name: crank' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "SKILL.md mentions wave concept" "grep -qi 'wave' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions worker concept" "grep -qi 'worker' '$SKILL_DIR/SKILL.md'"
check "Lead-only commit pattern documented" "grep -rqi 'lead.*commit\|lead-only' '$SKILL_DIR/'"
check "FIRE loop documented" "grep -q 'FIRE' '$SKILL_DIR/SKILL.md'"
check "No phantom bd cook refs" "! grep -q 'bd cook' '$SKILL_DIR/SKILL.md'"
check "No phantom gt convoy refs" "! grep -q 'gt convoy' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


