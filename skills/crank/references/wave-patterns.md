# Wave Patterns

## The FIRE Loop

Crank follows FIRE for each wave:

| Phase | Beads Mode | TaskList Mode |
|-------|-----------|--------------|
| **FIND** | `bd ready` — get unblocked beads issues | `TaskList()` → pending, unblocked |
| **IGNITE** | TaskCreate from beads + `/swarm` | `/swarm` (tasks already in TaskList) |
| **REAP** | Swarm results + `bd update --status closed` | Swarm results (TaskUpdate by workers) |
| **VIBE** | Wave diff vs acceptance criteria → PASS/WARN/FAIL | Same |
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
        /swarm → spawns 3 fresh-context agents
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
        /swarm → spawns 2 fresh-context agents
        ↓
        bd update for completed

Final vibe on all changes → Epic DONE
```

### TaskList Mode

```
Wave 1: TaskList() → [task-1, task-2, task-3] (pending, unblocked)
        ↓
        /swarm → spawns 3 fresh-context agents
                  ↓         ↓         ↓
               DONE      DONE      BLOCKED
                                     ↓
                               (reset to pending, retry next wave)

Wave 2: TaskList() → [task-4, task-3-retry] (pending, unblocked)
        ↓
        /swarm → spawns 2 fresh-context agents
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
  Gate: GREEN confirmation — ALL tests must PASS + wave vibe gate
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

### GREEN Confirmation Gate

After IMPL WAVE, the lead runs the test suite. ALL tests must PASS:
- New tests (from TEST WAVE) must now pass
- Existing tests must still pass (no regressions)
- Standard wave vibe gate also applies

### Contract Validation

SPEC WAVE workers explore the codebase before writing contracts (not fully isolated). This prevents generic, ungrounded specs. Workers read:
- Existing types, interfaces, and patterns
- Related test files for style reference
- Module structure and dependencies

But do NOT read implementation details of the specific feature being specified.

## Wave Vibe Gate (MANDATORY)

> **Principle:** Fresh context catches what saturated context misses. No self-grading.

**After closing all beads in a wave, before advancing to the next wave:**

1. **Tag the wave start** (for diff):
   ```bash
   # Before each wave (Step 4), record the starting commit:
   WAVE_START_SHA=$(git rev-parse HEAD)
   ```

2. **Compute wave diff:**
   ```bash
   git diff $WAVE_START_SHA HEAD --name-only
   ```

3. **Load acceptance criteria** for all issues closed in this wave:
   ```bash
   # For each closed issue in the wave:
   bd show <issue-id>  # extract ACCEPTANCE CRITERIA section
   ```

4. **Run inline vibe** (spec-compliance + error-paths, 2 judges minimum):
   ```
   Tool: Skill
   Parameters:
     skill: "agentops:vibe"
     args: "--quick --diff $WAVE_START_SHA --criteria '<acceptance criteria>'"
   ```
   If /vibe doesn't support --quick, run a lightweight council:
   - Spawn 2 Task agents (spec-compliance judge, error-paths judge)
   - Each reviews the wave diff against acceptance criteria
   - Aggregate verdicts: all PASS = PASS, any FAIL = FAIL, else WARN

   **Note:** SPEC WAVE has its own validation (contract completeness check) and TEST WAVE has the RED gate. The Wave Vibe Gate applies only to IMPL and REFACTOR waves.

5. **Gate on verdict:**

   | Verdict | Action |
   |---------|--------|
   | **PASS** | Record `CRANK_VIBE: wave=N verdict=PASS` in epic notes. Advance to next wave. |
   | **WARN** | Create fix beads as children of the epic (`bd create`). Execute fixes inline (small) or as wave N.5 via swarm. Re-run vibe. If PASS on re-vibe, advance. If still WARN after 2 attempts, treat as FAIL. |
   | **FAIL** | Record `CRANK_VIBE: wave=N verdict=FAIL` in epic notes. Output `<promise>BLOCKED</promise>` and exit. Human review required. |

   ```bash
   # Record verdict in epic notes
   bd update <epic-id> --append-notes "CRANK_VIBE: wave=$wave verdict=<PASS|WARN|FAIL> at $(date -Iseconds)"
   ```
