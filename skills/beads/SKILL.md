---
name: beads
description: >
  This skill should be used when the user asks to "track issues",
  "create beads issue", "show blockers", "what's ready to work on",
  "beads routing", "prefix routing", "cross-rig beads", "BEADS_DIR",
  "two-level beads", "town vs rig beads", "slingable beads",
  or needs guidance on git-based issue tracking with the bd CLI.
allowed-tools: "Read,Bash(bd:*)"
version: 0.34.1
context-budget:
  skill-md: 6KB
  references-total: 146KB
  typical-session: 20KB
author: "Steve Yegge <https://github.com/steveyegge>"
license: "MIT"
hooks:
  PostToolUse:
    - match: "Bash"
      action: "run"
      command: "[[ \"$INPUT\" == *'bd close'* && \"$INPUT\" == *'epic'* ]] && echo '[Flywheel] Epic closed - consider running /retro to extract learnings' || true"
---

# Beads - Persistent Task Memory for AI Agents

Graph-based issue tracker that survives conversation compaction.

## Overview

**bd (beads)** replaces markdown task lists with a dependency-aware graph stored in git.

**Key Distinction**:
- **bd**: Multi-session work, dependencies, survives compaction, git-backed
- **TodoWrite**: Single-session tasks, linear execution, conversation-scoped

**Decision Rule**: If resuming in 2 weeks would be hard without bd, use bd.

## Prerequisites

- **bd CLI**: Version 0.34.0+ installed and in PATH
- **Git Repository**: Current directory must be a git repo
- **Initialization**: `bd init` run once (humans do this, not agents)

---

## Session Protocol

### Start of Session

```bash
bd ready                              # Find unblocked work
bd show <id>                          # Get full context
bd update <id> --status in_progress   # Claim it
```

### During Work

```bash
bd update <id> --notes "COMPLETED: X. IN PROGRESS: Y. BLOCKED BY: Z"
```

**Critical**: Write notes as if explaining to a future agent with zero context.

### End of Session

```bash
bd close <id> --reason "What was accomplished"
bd sync                               # Sync to git
```

---

## Essential Commands (Top 10)

| Command | Purpose |
|---------|---------|
| `bd ready` | Show unblocked tasks |
| `bd create "Title" -p 1` | Create task (priority 0-4) |
| `bd show <id>` | View task details |
| `bd update <id> --status in_progress` | Start working |
| `bd update <id> --notes "Progress"` | Add notes |
| `bd close <id> --reason "Done"` | Complete task |
| `bd dep add <child> <parent>` | Add dependency |
| `bd list` | See all tasks |
| `bd search <query>` | Find by keyword |
| `bd sync` | Sync with git |

**Full command reference**: `references/CLI_REFERENCE.md`

---

## Task Creation

```bash
# Basic task
bd create "Fix authentication bug" -p 0 --type bug

# With description
bd create "Implement OAuth" -p 1 --description "Add OAuth2 for Google, GitHub"

# Epic with children
bd create "Epic: OAuth" -p 0 --type epic
bd create "Research providers" -p 1 --parent <epic-id>
bd create "Implement endpoints" -p 1 --parent <epic-id>
```

**Priority**: 0=critical, 1=high, 2=medium, 3=low, 4=backlog

**Types**: bug, feature, task, epic, chore

---

## Dependencies

```bash
bd dep add <child-id> <parent-id>    # Parent blocks child
bd dep list <id>                      # View dependencies
```

**Meaning**: `<parent-id>` must close before `<child-id>` becomes ready.

bd prevents circular dependencies automatically.

---

## Git Sync

```bash
bd sync                    # All-in-one: export, commit, pull, push
bd export -o backup.jsonl  # Export only
bd import -i backup.jsonl  # Import only
```

**Data stored in**: `.beads/issues.jsonl` (git-tracked)

---

## When to Use bd vs TodoWrite

| Question | YES → | NO → |
|----------|-------|------|
| Will I need this in 2 weeks? | bd | TodoWrite |
| Could conversation get compacted? | bd | TodoWrite |
| Does this have blockers/dependencies? | bd | TodoWrite |
| Is this fuzzy/exploratory? | bd | TodoWrite |
| Will this be done this session? | TodoWrite | bd |

---

## Critical Rules

**These rules prevent database corruption:**

1. **Single prefix per database** - Never mix prefixes (e.g., `ap-` with `etl-`)
2. **Standard ID format only** - IDs must be `prefix-hash`, never `prefix.step-name`
3. **Always sync before stopping** - `bd sync && git push` is mandatory
4. **Full IDs always** - Use `ap-1234`, never just `1234`

**If using Gas Town:**
- Mayor NEVER implements, always dispatches via `gt sling`
- Create rig beads with `BEADS_DIR=~/gt/<rig>/mayor/rig/.beads bd create`

---

## References

Load these JIT when needed:

| Reference | When to Load |
|-----------|--------------|
| `references/CLI_REFERENCE.md` | Full command syntax |
| `references/WORKFLOWS.md` | Complex workflow patterns |
| `references/DEPENDENCIES.md` | Dependency deep dive |
| `references/TROUBLESHOOTING.md` | Error resolution |
| `references/ANTI_PATTERNS.md` | **Mistakes that corrupt databases** |
| `references/RESUMABILITY.md` | Compaction survival |
| `references/BOUNDARIES.md` | bd vs TodoWrite details |
| `references/ROUTING.md` | Multi-rig prefix routing, two-level architecture |

---

## Quick Troubleshooting

| Error | Fix |
|-------|-----|
| `bd: command not found` | Install from github.com/steveyegge/beads |
| `No .beads database` | Run `bd init` |
| `Task not found` | Use `bd list` to verify ID |
| `Circular dependency` | Restructure - bd prevents cycles |
| `Database out of sync` | Run `bd sync --import-only` |
| `prefix mismatch detected` | Filter JSONL to single prefix, see TROUBLESHOOTING.md |
| `invalid suffix` | Molecule-style IDs - filter out or rebuild, see ANTI_PATTERNS.md |

**Full troubleshooting**: `references/TROUBLESHOOTING.md`
**Preventable mistakes**: `references/ANTI_PATTERNS.md`
