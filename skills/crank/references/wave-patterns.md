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
