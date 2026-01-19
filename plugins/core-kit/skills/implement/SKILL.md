---
name: implement
description: >
  Execute a single beads issue with full lifecycle. Triggers: "implement",
  "work on task", "fix bug", "start feature", "pick up next issue".
version: 2.1.0
context: fork
author: "AI Platform Team"
license: "MIT"
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - beads
---

# Implement Skill

Execute a SINGLE beads issue from `open` to `closed`.

## Overview

Take a beads issue through: context → implement → test → close → commit.

**When to Use**: Any beads issue needs execution. Best for learning, complex bugs, unfamiliar code.

**When NOT to Use**: Creating issues (`/formulate`), research (`/research`), bulk (`/implement-wave`).

---

## Workflow

```
0. Context Discovery   -> 6-tier hierarchy
1. Select Issue        -> bd ready or specified ID
2. Start Work          -> bd update --status in_progress
3. Implement           -> Follow patterns, document progress
4. Validate            -> Lint, patterns
5. Test (MANDATORY)    -> just test
6. Close + Commit      -> bd close, git commit, bd sync
7. Next Steps          -> bd ready (STOP after ONE)
```

---

## Phase 0: Context Discovery

See `~/.claude/skills/research/references/context-discovery.md` for full 6-tier hierarchy.

**Quick version**: Code-Map → Semantic Search → Scoped Grep → Source → .agents/ → External

---

## Phase 1: Select Issue

```bash
# If ID provided
bd show $ARGUMENTS

# If no ID (auto-select)
bd ready  # Pick highest priority (P0 > P1 > P2 > P3)
```

**STOP if no issues**: Inform user to run `/formulate` to create issues.

---

## Phase 2-3: Start & Implement

```bash
bd update <id> --status in_progress
```

**Progress Updates**:
```bash
bd comment <id> "Implemented X in path/to/file.py"
```

Write notes assuming future Claude has ZERO history.

---

## Phase 4-5: Validate & Test

```bash
just lint      # Linting
just test      # MANDATORY - must pass
```

**If tests FAIL**: Fix or document blocker. Do NOT commit.

---

## Phase 6: Close & Commit

```bash
bd close <id> --reason "Completed: [summary]"

git add -A
git commit -m "feat(<scope>): <description>

Closes: <id>"

bd sync && git push
```

---

## Phase 7: Next Steps

```bash
bd ready
```

**STOP after ONE issue.** Let user decide whether to continue.

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Batch multiple issues | One per /implement |
| Skip tests | Always run tests |
| Commit with failing tests | Fix first |
| Abandon without documenting | Update status/comments |

---

## Essential Commands

| Command | Purpose |
|---------|---------|
| `bd ready` | Find unblocked issues |
| `bd show <id>` | View details |
| `bd update <id> --status in_progress` | Start |
| `bd comment <id> "msg"` | Progress note |
| `bd close <id> --reason "msg"` | Complete |
| `just test` | Run tests |
| `bd sync && git push` | Sync and push |

---

## References

- **beads skill**: Full CLI reference
- `/formulate`: Create issues from a goal
- `/implement-wave`: Bulk execution (parallel via Task() subagents)
- `/crank`: Full autonomous epic execution
