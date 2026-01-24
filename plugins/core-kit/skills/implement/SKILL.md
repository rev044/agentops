---
name: implement
description: >
  Execute a single beads issue with full lifecycle. Triggers: "implement",
  "work on task", "fix bug", "start feature", "pick up next issue".
version: 2.1.0
tier: team
context: inline
author: "AI Platform Team"
license: "MIT"
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - beads
  - standards
---

# Implement Skill

Execute a SINGLE beads issue from `open` to `closed`.

## Role in the Brownian Ratchet

Implement is the **micro-ratchet** - the atomic unit of progress:

| Component | Implement's Role |
|-----------|------------------|
| **Chaos** | Coding attempts, debugging, iteration |
| **Filter** | Tests must pass, lint must pass |
| **Ratchet** | Issue status: `open` → `in_progress` → `closed` |

> **The issue lifecycle IS the ratchet. Once closed, work is permanent.**

Each `/implement` cycle is a complete micro-ratchet:
```
open → in_progress (chaos) → tests pass (filter) → closed (ratchet)
```

**Key property:** Issues don't go backward. `closed` is permanent.
Failed attempts stay `in_progress` until fixed or marked blocked.

## Overview

Take a beads issue through: context → implement → test → close → commit.

**When to Use**: Any beads issue needs execution.

**When NOT to Use**: Creating issues (`/plan`), research (`/research`), bulk (`/implement-wave`).

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

**STOP if no issues**: Inform user to run `/plan`.

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

## Context Management (L3)

**From 2026-01 Post-Mortem:** Context rot is THE failure mode.

| Threshold | Status | Action |
|-----------|--------|--------|
| < 35% | Green | Continue |
| 35-40% | Yellow | Checkpoint soon |
| 40-60% | Red | Stop, checkpoint, fresh session |
| > 60% | Collapse | Context lost (99% info loss) |

**Checkpoints:** After each issue in a multi-issue wave, check context utilization.
If approaching 40%, commit progress and start fresh session.

**See:** `~/.claude/CLAUDE.md` Post-Mortem Learnings (2026-01)

---

## First Epic Buffer (L8)

**From 2026-01 Post-Mortem:** First epics grow 30-50% beyond estimate.

When implementing the first epic in a new domain:
- Research phase doesn't fully map cross-component dependencies
- Accept that scope growth is natural for exploratory work
- Budget 30% buffer for unexpected tasks

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
- `/plan`: Create issues
- `/implement-wave`: Bulk execution

---

## Standards Loading

During Phase 0 (Context Discovery), load relevant standards based on files being modified:

| File Pattern | Load Reference |
|--------------|----------------|
| `*.py` | `~/.claude/skills/standards/references/python.md` |
| `*.go` | `~/.claude/skills/standards/references/go.md` |
| `*.ts`, `*.tsx` | `~/.claude/skills/standards/references/typescript.md` |
| `*.sh` | `~/.claude/skills/standards/references/shell.md` |
| `*.yaml`, `*.yml` | `~/.claude/skills/standards/references/yaml.md` |
| `*.md` | `~/.claude/skills/standards/references/markdown.md` |
| `*.json`, `*.jsonl` | `~/.claude/skills/standards/references/json.md` |

**Usage**: Reference standards during implementation for consistent code style and patterns.

---

## Phase Completion (RPI Workflow)

When implementation of an epic or major milestone is complete:

```bash
~/.claude/scripts/checkpoint.sh implement "Brief description of work completed"
```

This will:
1. Save a checkpoint to `~/gt/.agents/$RIG/checkpoints/`
2. Remind you to start a fresh session for the next epic
3. Provide recovery commands

**When to checkpoint:**
- After completing all issues in a wave
- After completing an entire epic
- Before starting a different type of work

**Why fresh session?** Implementation context (code patterns, debugging state) should
not carry over to the next epic or research phase. Clean sessions prevent context pollution.
