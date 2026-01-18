---
name: implement-wave
description: >
  Parallel execution of multiple issues. Triggers: "run a wave", "parallel
  implementation", "batch execute", "implement in parallel".
version: 2.1.0
context: fork
author: "AI Platform Team"
license: "MIT"
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, Task, TaskOutput, TodoWrite
skills:
  - beads
  - implement
---

# Implement Wave Skill

Execute a wave of independent issues in parallel using sub-agents.

## Overview

Orchestrate parallel implementation with conflict detection, batch closing, and single-commit consolidation.

**When to Use**: >=2 ready issues, parallelism benefit.

**When NOT to Use**: Single issue (use `/implement`), complex dependencies.

---

## Workflow

```
0. Wave Identification  -> bd ready, filter up to 8
1. Conflict Detection   -> File overlap, greedy partition
2. Launch Sub-Agents    -> Parallel Task() calls
3. Monitor Results      -> Poll TaskOutput
4. Batch Close + Commit -> Single wave commit
5. Handle Failures      -> Leave failed in_progress
```

---

## Phase 0: Wave Identification

```bash
bd ready
```

**Selection**: Up to 8 issues, priority ordered (P0 > P1 > P2 > P3).

---

## Phase 1: Conflict Detection

**CRITICAL**: Detect file overlaps before parallel execution.

**Algorithm**: For each issue pair, check if predicted file changes overlap:
1. Read issue descriptions to identify target files
2. Build overlap matrix: `overlap[i][j] = files_touched(i) ∩ files_touched(j)`
3. If overlap, place in sequential groups (greedy partition)

If overlaps found, partition into sequential groups.

---

## Phase 2: Launch Sub-Agents

```markdown
Task(
    subagent_type="general-purpose",
    model="sonnet",
    prompt="Implement issue <id>. Add 'READY_FOR_BATCH_CLOSE' when done.
            Do NOT commit - orchestrator handles commits.",
    run_in_background=True
)
```

**Model**: Use `sonnet` for implementation (cost-effective for code generation).

**Constraints**: Maximum 8 concurrent sub-agents.

---

## Phase 3-4: Collect & Close

Poll for completion, check signals:
- `READY_FOR_BATCH_CLOSE` → Success
- `BLOCKER:` → Failure

```bash
bd close <id1> <id2> ...  # Only successful issues
git add -A && git commit -m "feat(wave): complete N issues"
bd sync && git push
```

---

## Phase 5: Handle Failures

- Leave failed issues `in_progress`
- Do NOT auto-close
- Suggest: `/implement <id>` to retry

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Launch >8 agents | Limit to 8 max |
| Auto-close failed | Leave in_progress |
| Commit per issue | Single wave commit |
| Ignore overlaps | Detect and partition |

---

## References

- `/implement` - Single issue execution
- `~/.claude/standards/model-routing.md` - Model selection guidance
