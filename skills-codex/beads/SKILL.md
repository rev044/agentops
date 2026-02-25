---
name: beads
description: 'This skill should be used when the user asks to "track issues", "create beads issue", "show blockers", "what''''s ready to work on", "beads routing", "prefix routing", "cross-rig beads", "BEADS_DIR", "two-level beads", "town vs rig beads", "slingable beads", or needs guidance on git-based issue tracking with the bd CLI.'
---


# Beads - Persistent Task Memory for AI Agents

Graph-based issue tracker that survives conversation compaction.

## Overview

**bd (beads)** replaces markdown task lists with a dependency-aware graph stored in git.

**Key Distinction**:
- **bd**: Multi-session work, dependencies, survives compaction, git-backed
- **Task tools (TaskCreate/TaskUpdate/TaskList)**: Single-session tasks, status tracking, conversation-scoped

**Decision Rule**: If resuming in 2 weeks would be hard without bd, use bd.

## Prerequisites

- **bd CLI**: Version 0.34.0+ installed and in PATH
- **Git Repository**: Current directory must be a git repo
- **Initialization**: `bd init` run once (humans do this, not agents)

## Examples

### Skill Loading from $vibe

**User says:** `$vibe`

**What happens:**
1. Agent loads beads skill automatically via dependency
2. Agent calls `bd show <id>` to read issue metadata
3. Agent links validation findings to the issue being checked
4. Output references issue ID in validation report

**Result:** Validation report includes issue context, no manual bd lookups needed.

### Skill Loading from $implement

**User says:** `$implement ag-xyz-123`

**What happens:**
1. Agent loads beads skill to understand issue structure
2. Agent calls `bd show ag-xyz-123` to read issue body
3. Agent checks dependencies with bd output
4. Agent closes issue with `bd close ag-xyz-123` after completion

**Result:** Issue lifecycle managed automatically during implementation.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| bd command not found | bd CLI not installed or not in PATH | Install bd: `brew install bd` or check PATH |
| "not a git repository" error | bd requires git repo, current dir not initialized | Run `git init` or navigate to git repo root |
| "beads not initialized" error | .beads/ directory missing | Human runs `bd init --prefix <prefix>` once |
| Issue ID format errors | Wrong prefix or malformed ID | Check rigs.json for correct prefix, follow `<prefix>-<tag>-<num>` format |

## Reference Documents

- [references/ANTI_PATTERNS.md](references/ANTI_PATTERNS.md)
- [references/BOUNDARIES.md](references/BOUNDARIES.md)
- [references/CLI_REFERENCE.md](references/CLI_REFERENCE.md)
- [references/DEPENDENCIES.md](references/DEPENDENCIES.md)
- [references/INTEGRATION_PATTERNS.md](references/INTEGRATION_PATTERNS.md)
- [references/ISSUE_CREATION.md](references/ISSUE_CREATION.md)
- [references/MOLECULES.md](references/MOLECULES.md)
- [references/PATTERNS.md](references/PATTERNS.md)
- [references/RESUMABILITY.md](references/RESUMABILITY.md)
- [references/ROUTING.md](references/ROUTING.md)
- [references/STATIC_DATA.md](references/STATIC_DATA.md)
- [references/TROUBLESHOOTING.md](references/TROUBLESHOOTING.md)
- [references/WORKFLOWS.md](references/WORKFLOWS.md)

---

## References

### ANTI_PATTERNS.md

# Beads Anti-Patterns

Hard-won lessons from production beads usage. Avoid these mistakes.

---

## Critical Anti-Patterns

### 1. Molecule-Style Issue IDs

**DON'T**: Create issues with dot-separated hierarchical IDs

```bash
# WRONG - These IDs corrupt the database
code-map-validation.calculate-coverage
etl-throughput-optimization.enable-parallel-sync
kagent-openwebui-bridge.admin-functions
```

**DO**: Use standard `prefix-xxxx` format

```bash
# CORRECT - Standard beads ID format
ap-7tc6
ap-euoy
ap-cr7k
```

**Why it breaks**:
- bd expects IDs in `prefix-hash` format
- Dot-separated IDs fail prefix validation during import
- `bd sync --import-only` errors with "invalid suffix"
- Database becomes corrupted, requiring full rebuild

**Root cause**: Early formula/molecule templates created non-standard IDs. This was a design mistake.

**Fix**: If you have molecule-style IDs, filter them out or rebuild:
```bash
# Filter to standard format only
grep -E '"id":"[a-z]+-[a-z0-9]+' .beads/issues.jsonl > clean.jsonl
mv clean.jsonl .beads/issues.jsonl
bd sync --import-only
```

---

### 2. Prefix Proliferation

**DON'T**: Mix multiple prefixes in one database

```bash
# WRONG - Multiple prefixes in same .beads/
code-map-validation
etl-throughput-optimization
kagent-openwebui-bridge
ap-1234
```

**DO**: One prefix per beads database

```bash
# CORRECT - Single prefix
ap-1234
ap-5678
ap-abcd
```

**Why it breaks**:
- `bd sync --import-only` fails with "prefix mismatch detected"
- Database configured for one prefix rejects others
- Cross-prefix dependencies don't resolve correctly

**Root cause**: Formulas/molecules created issues with their own prefixes instead of the database's prefix.

**Fix**: Enforce single prefix policy:
```bash
# Check for prefix violations
grep -o '"id":"[^-]*' .beads/issues.jsonl | sort -u
# Should show only ONE prefix
```

---

### 3. Skipping Session End Protocol

**DON'T**: Stop work without syncing

```bash
# WRONG - Work not persisted
bd close ap-1234 --reason "Done"
# ... session ends without sync
```

**DO**: Always sync and push before stopping

```bash
# CORRECT - Full session end protocol
bd close ap-1234 --reason "Done"
bd sync                    # Commit beads changes
git add .beads/            # Stage if needed
git commit -m "beads: close ap-1234"
git push                   # Push to remote
```

**Why it matters**:
- Beads changes live in `.beads/issues.jsonl`
- Without commit+push, changes lost on branch switch
- Other agents/sessions won't see your updates
- Merge conflicts accumulate if not synced regularly

---

### 4. Mayor Implementing Instead of Dispatching

**DON'T**: Mayor role edits code directly

```bash
# WRONG - Mayor implementing
cd ~/gt/ai_platform/mayor/rig
vim services/etl/app/main.py  # NO!
```

**DO**: Mayor dispatches to polecats

```bash
# CORRECT - Mayor dispatches
gt sling ap-1234 ai_platform
# Polecat does the work, Mayor monitors
gt convoy list  <!-- FUTURE: gt convoy not yet implemented -->
```

**Why it matters**:
- Mayor context is precious (coordinates across rigs)
- Polecat isolation provides 100x context reduction
- Task agent returns ~10KB, polecat status ~100 tokens
- Mayor implementing causes context bloat

**Rule**: If you're Mayor, NEVER edit code. Even "quick fixes" go through `gt sling`.

---

### 5. Stale MR Issue Accumulation

**DON'T**: Let merge request issues pile up

```bash
bd list --type=merge-request
# 35 stale MR issues from months ago
```

**DO**: Clean up MRs when branches merge

```bash
# After merge, close the MR issue
bd close ap-mr-123 --reason "Branch merged"

# Regular cleanup
bd list --status=open --type=merge-request | while read id; do
    # Check if branch still exists
    git branch -r | grep -q "origin/$branch" || bd close $id --reason "Branch merged/deleted"
done
```

**Why it matters**:
- Stale MRs create noise in `bd list`
- `bd ready` shows work that doesn't exist
- Database bloat from abandoned tracking issues

---

### 6. Using Short IDs

**DON'T**: Use abbreviated issue IDs

```bash
# WRONG - Ambiguous
bd show 1234
bd close xyz
```

**DO**: Use full prefix-hash IDs

```bash
# CORRECT - Unambiguous
bd show ap-1234
bd close ap-xyz5
```

**Why it matters**:
- Short IDs can match multiple issues
- Cross-rig work requires full IDs for routing
- Gas Town dispatch needs full IDs

---

### 7. Creating Issues Without Context

**DON'T**: Create issues with minimal information

```bash
# WRONG - No context for future agents
bd create "Fix the bug"
```

**DO**: Include enough context for resumption

```bash
# CORRECT - Self-contained context
bd create "Fix authentication timeout in OAuth flow" \
  --description "Users report 30s timeout during OAuth callback.
Error in services/gateway/oauth.py:142.
Reproduce: Login with Google SSO on slow network.
Fix: Increase timeout or add retry logic." \
  --type bug \
  --priority 1
```

**Why it matters**:
- Issues survive compaction, conversations don't
- Future agent needs full context from issue alone
- 2-week resumption test: Could you restart this work from the issue text?

---

## Database Health Commands

### Check for Problems

```bash
# Check prefix consistency
grep -o '"id":"[^-]*' .beads/issues.jsonl | sort -u

# Check for molecule-style IDs
grep -E '"id":"[^"]+\.[^"]+' .beads/issues.jsonl

# Check issue count
wc -l .beads/issues.jsonl

# Check database vs JSONL sync
bd doctor
```

### Maintenance Commands

```bash
# Weekly cleanup
bd list --status=tombstone  # Review tombstones
bd doctor                   # Health check

# Before major work
bd sync --status            # Check sync state
bd ready                    # Verify ready queue

# After git pull
bd sync --import-only       # Import remote changes
```

### Nuclear Options

> **WARNING: DESTRUCTIVE OPERATIONS BELOW**
> These commands permanently delete data. Before running:
> 1. Ensure you have a backup: `cp -r .beads/ .beads.backup/`
> 2. Verify you're in the correct directory
> 3. Understand that this cannot be undone

```bash
# Full database rebuild (DESTRUCTIVE)
rm -rf .beads/*.db
bd sync --import-only

# Complete reset (VERY DESTRUCTIVE)
rm -rf .beads/
bd init --prefix=ap
```

---

## Gas Town Integration Rules

When using beads with Gas Town:

| Role | Can Create Issues | Can Edit Code | Uses |
|------|-------------------|---------------|------|
| Mayor | Yes (HQ beads) | NO | gt sling, gt convoy <!-- FUTURE: gt convoy not yet implemented --> |
| Crew | Yes (rig beads) | Yes | bd commands directly |
| Polecat | Update only | Yes | bd update, bd close |

**Prefix routing**:
- HQ beads: `hq-*` prefix, stored at `~/gt/.beads/`
- Rig beads: Project prefix (e.g., `ap-*`), stored at `~/gt/<rig>/.beads/`

**Creating slingable beads from Mayor**:
```bash
# Mayor can't hook hq- beads to polecats
# Create in rig database instead:
BEADS_DIR=~/gt/ai_platform/mayor/rig/.beads bd create --title="Task" --type=task
```

---

## Related

- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Error resolution
- [ROUTING.md](ROUTING.md) - Multi-rig prefix routing
- [WORKFLOWS.md](WORKFLOWS.md) - Correct workflow patterns

### BOUNDARIES.md

# Boundaries: When to Use bd vs Task Tools

This reference provides detailed decision criteria for choosing between bd issue tracking and Task tools (`TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`) for task management.

## Contents

- [The Core Question](#the-core-question)
- [Decision Matrix](#decision-matrix)
  - [Use bd for](#use-bd-for): Multi-Session Work, Complex Dependencies, Knowledge Work, Side Quests, Project Memory
  - [Use Task tools for](#use-task-tools-for): Single-Session Tasks, Linear Execution, Immediate Context, Simple Tracking
- [Detailed Comparison](#detailed-comparison)
- [Integration Patterns](#integration-patterns)
  - Pattern 1: bd as Strategic, Task tools as Tactical
  - Pattern 2: Task tools as Working Copy of bd
  - Pattern 3: Transition Mid-Session
- [Real-World Examples](#real-world-examples)
  - Strategic Document Development, Simple Feature Implementation, Bug Investigation, Refactoring with Dependencies
- [Common Mistakes](#common-mistakes)
  - Using Task tools for multi-session work, using bd for simple tasks, not transitioning when complexity emerges, creating too many bd issues, never using bd
- [The Transition Point](#the-transition-point)
- [Summary Heuristics](#summary-heuristics)

## The Core Question

**"Could I resume this work after 2 weeks away?"**

- If bd would help you resume → **use bd**
- If markdown skim would suffice → **Task tools are fine**

This heuristic captures the essential difference: bd provides structured context that persists across long gaps, while Task tools excel at immediate session tracking.

## Decision Matrix

### Use bd for:

#### Multi-Session Work
Work spanning multiple compaction cycles or days where context needs to persist.

**Examples:**
- Strategic document development requiring research across multiple sessions
- Feature implementation split across several coding sessions
- Bug investigation requiring experimentation over time
- Architecture design evolving through multiple iterations

**Why bd wins**: Issues capture context that survives compaction. Return weeks later and see full history, design decisions, and current status.

#### Complex Dependencies
Work with blockers, prerequisites, or hierarchical structure.

**Examples:**
- OAuth integration requiring database setup, endpoint creation, and frontend changes
- Research project with multiple parallel investigation threads
- Refactoring with dependencies between different code areas
- Migration requiring sequential steps in specific order

**Why bd wins**: Dependency graph shows what's blocking what. `bd ready` automatically surfaces unblocked work. No manual tracking required.

#### Knowledge Work
Tasks with fuzzy boundaries, exploration, or strategic thinking.

**Examples:**
- Architecture decision requiring research into frameworks and trade-offs
- API design requiring research into multiple options
- Performance optimization requiring measurement and experimentation
- Documentation requiring understanding system architecture

**Why bd wins**: `design` and `acceptance_criteria` fields capture evolving understanding. Issues can be refined as exploration reveals more information.

#### Side Quests
Exploratory work that might pause the main task.

**Examples:**
- During feature work, discover a better pattern worth exploring
- While debugging, notice related architectural issue
- During code review, identify potential improvement
- While writing tests, find edge case requiring research

**Why bd wins**: Create issue with `discovered-from` dependency, pause main work safely. Context preserved for both tracks. Resume either one later.

#### Project Memory
Need to resume work after significant time with full context.

**Examples:**
- Open source contributions across months
- Part-time projects with irregular schedule
- Complex features split across sprints
- Research projects with long investigation periods

**Why bd wins**: Git-backed database persists indefinitely. All context, decisions, and history available on resume. No relying on conversation scrollback or markdown files.

---

### Use Task tools for:

#### Single-Session Tasks
Work that completes within current conversation.

**Examples:**
- Implementing a single function based on clear spec
- Fixing a bug with known root cause
- Adding unit tests for existing code
- Updating documentation for recent changes

**Why Task tools win**: Lightweight task tracking is perfect for linear execution. No need for persistence or dependencies. Clear completion within session.

#### Linear Execution
Straightforward step-by-step tasks with no branching.

**Examples:**
- Database migration with clear sequence
- Deployment checklist
- Code style cleanup across files
- Dependency updates following upgrade guide

**Why Task tools win**: Steps are predetermined and sequential. No discovery, no blockers, no side quests. Just execute top to bottom.

#### Immediate Context
All information already in conversation.

**Examples:**
- User provides complete spec and asks for implementation
- Bug report with reproduction steps and fix approach
- Refactoring request with clear before/after vision
- Config changes based on user preferences

**Why Task tools win**: No external context to track. Everything needed is in current conversation. Task tools provide user visibility, nothing more needed.

#### Simple Tracking
Just need a checklist to show progress to user.

**Examples:**
- Breaking down implementation into visible steps
- Showing validation workflow progress
- Demonstrating systematic approach
- Providing reassurance work is proceeding

**Why Task tools win**: User wants to see thinking and progress. Task tools are visible in conversation. bd is invisible background structure.

---

## Detailed Comparison

| Aspect | bd | Task tools |
|--------|-----|-----------|
| **Persistence** | Git-backed, survives compaction | Session-only, task list resets when session ends |
| **Dependencies** | Graph-based, automatic ready detection | Supports blockedBy/blocks via TaskUpdate |
| **Discoverability** | `bd ready` surfaces work | TaskList shows all tasks with status |
| **Complexity** | Handles nested epics, blockers | Flat task list with status and dependencies |
| **Visibility** | Background structure, not in conversation | Visible to user in chat |
| **Setup** | Requires `.beads/` directory in project | Always available |
| **Best for** | Complex, multi-session, explorative | Simple, single-session, linear |
| **Context capture** | Design notes, acceptance criteria, links | Subject and description |
| **Evolution** | Issues can be updated, refined over time | Tasks updated via TaskUpdate as work progresses |
| **Audit trail** | Full history of changes | Only visible in conversation |

## Integration Patterns

bd and Task tools can coexist effectively in a session. Use both strategically.

### Pattern 1: bd as Strategic, Task tools as Tactical

**Setup:**
- bd tracks high-level issues and dependencies
- Task tools track current session's execution steps

**Example:**
```
bd issue: "Implement user authentication" (epic)
  ├─ Child issue: "Create login endpoint"
  ├─ Child issue: "Add JWT token validation"  ← Currently working on this
  └─ Child issue: "Implement logout"

Create session tasks:
  TaskCreate: "Install JWT library" (pending)
  TaskCreate: "Create token validation middleware" (pending)
  TaskCreate: "Add tests for token expiry" (pending)
  TaskCreate: "Update API documentation" (pending)
Mark completed via TaskUpdate as you go.
```

**When to use:**
- Complex features with clear implementation steps
- User wants to see current progress but larger context exists
- Multi-session work currently in single-session execution phase

### Pattern 2: Task tools as Working Copy of bd

**Setup:**
- Start with bd issue containing full context
- Create session tasks from bd issue's acceptance criteria
- Update bd as tasks complete

**Example:**
```
Session start:
- Check bd: "issue-auth-42: Add JWT token validation" is ready
- Extract acceptance criteria into session tasks via TaskCreate
- Mark bd issue as in_progress
- Work through tasks, marking completed via TaskUpdate
- Update bd design notes as you learn
- When all tasks complete, close bd issue
```

**When to use:**
- bd issue is ready but execution is straightforward
- User wants visible progress tracking
- Need structured approach to larger issue

### Pattern 3: Transition Mid-Session

**From Task tools to bd:**

Recognize mid-execution that work is more complex than anticipated.

**Trigger signals:**
- Discovering blockers or dependencies
- Realizing work won't complete this session
- Finding side quests or related issues
- Needing to pause and resume later

**How to transition:**
```
1. Create bd issue with current task list content
2. Note: "Discovered this is multi-session work during implementation"
3. Add dependencies as discovered
4. Keep task list for current session
5. Update bd issue before session ends
6. Next session: resume from bd, create new tasks if needed
```

**From bd to Task tools:**

Rare, but happens when bd issue turns out simpler than expected.

**Trigger signals:**
- All context already clear
- No dependencies discovered
- Can complete within session
- User wants execution visibility

**How to transition:**
```
1. Keep bd issue for historical record
2. Create session tasks from issue description via TaskCreate
3. Execute via task list
4. Close bd issue when done
5. Note: "Completed in single session, simpler than expected"
```

## Real-World Examples

### Example 1: Database Migration Planning

**Scenario**: Planning migration from MySQL to PostgreSQL for production application.

**Why bd**:
- Multi-session work across days/weeks
- Fuzzy boundaries - scope emerges through investigation
- Side quests - discover schema incompatibilities requiring refactoring
- Dependencies - can't migrate data until schema validated
- Project memory - need to resume after interruptions

**bd structure**:
```
db-epic: "Migrate production database to PostgreSQL"
  ├─ db-1: "Audit current MySQL schema and queries"
  ├─ db-2: "Research PostgreSQL equivalents for MySQL features" (blocks schema design)
  ├─ db-3: "Design PostgreSQL schema with type mappings"
  └─ db-4: "Create migration scripts and test data integrity" (blocked by db-3)
```

**Task tools role**: None initially. Might use Task tools for single-session testing sprints once migration scripts ready.

### Example 2: Simple Feature Implementation

**Scenario**: Add logging to existing endpoint based on clear specification.

**Why Task tools**:
- Single session work
- Linear execution - add import, call logger, add test
- All context in user message
- Completes within conversation

**Task tools**:
```
TaskCreate: "Import logging library" (pending)
TaskCreate: "Add log statements to endpoint" (pending)
TaskCreate: "Add test for log output" (pending)
TaskCreate: "Run tests" (pending)
```

**bd role**: None. Overkill for straightforward task.

### Example 3: Bug Investigation

**Initial assessment**: Seems simple, try Task tools first.

**Task tools**:
```
TaskCreate: "Reproduce bug" (pending)
TaskCreate: "Identify root cause" (pending)
TaskCreate: "Implement fix" (pending)
TaskCreate: "Add regression test" (pending)
```

**What actually happens**: Reproducing bug reveals it's intermittent. Root cause investigation shows multiple potential issues. Needs time to investigate.

**Transition to bd**:
```
Create bd issue: "Fix intermittent auth failure in production"
  - Description: Initially seemed simple but reproduction shows complex race condition
  - Design: Three potential causes identified, need to test each
  - Created issues for each hypothesis with discovered-from dependency

Pause for day, resume next session from bd context
```

### Example 4: Refactoring with Dependencies

**Scenario**: Extract common validation logic from three controllers.

**Why bd**:
- Dependencies - must extract before modifying callers
- Multi-file changes need coordination
- Potential side quest - might discover better pattern during extraction
- Need to track which controllers updated

**bd structure**:
```
refactor-1: "Create shared validation module"
  → blocks refactor-2, refactor-3, refactor-4

refactor-2: "Update auth controller to use shared validation"
refactor-3: "Update user controller to use shared validation"
refactor-4: "Update payment controller to use shared validation"
```

**Task tools role**: Could use Task tools for individual controller updates as implementing.

**Why this works**: bd ensures you don't forget to update a controller. `bd ready` shows next available work. Dependencies prevent starting controller update before extraction complete.

## Common Mistakes

### Mistake 1: Using Task tools for Multi-Session Work

**What happens**:
- Next session, forget what was done
- Scroll conversation history to reconstruct
- Lose design decisions made during implementation
- Start over or duplicate work

**Solution**: Create bd issue instead. Persist context across sessions.

### Mistake 2: Using bd for Simple Linear Tasks

**What happens**:
- Overhead of creating issue not justified
- User can't see progress in conversation
- Extra tool use for no benefit

**Solution**: Use Task tools. They're designed for exactly this case.

### Mistake 3: Not Transitioning When Complexity Emerges

**What happens**:
- Start with Task tools for "simple" task
- Discover blockers and dependencies mid-way
- Keep using Task tools despite poor fit
- Lose context when conversation ends

**Solution**: Transition to bd when complexity signal appears. Not too late mid-session.

### Mistake 4: Creating Too Many bd Issues

**What happens**:
- Every tiny task gets an issue
- Database cluttered with trivial items
- Hard to find meaningful work in `bd ready`

**Solution**: Reserve bd for work that actually benefits from persistence. Use "2 week test" - would bd help resume after 2 weeks? If no, skip it.

### Mistake 5: Never Using bd Because Task tools are Familiar

**What happens**:
- Multi-session projects become markdown swamps
- Lose track of dependencies and blockers
- Can't resume work effectively
- Rotten half-implemented plans

**Solution**: Force yourself to use bd for next multi-session project. Experience the difference in organization and resumability.

### Mistake 6: Always Asking Before Creating Issues (or Never Asking)

**When to create directly** (no user question needed):
- **Bug reports**: Clear scope, specific problem ("Found: auth doesn't check profile permissions")
- **Research tasks**: Investigative work ("Research workaround for Slides export")
- **Technical TODOs**: Discovered during implementation ("Add validation to form handler")
- **Side quest capture**: Discoveries that need tracking ("Issue: MCP can't read Shared Drive files")

**Why create directly**: Asking slows discovery capture. User expects proactive issue creation for clear-cut problems.

**When to ask first** (get user input):
- **Strategic work**: Fuzzy boundaries, multiple valid approaches ("Should we implement X or Y pattern?")
- **Potential duplicates**: Might overlap with existing work
- **Large epics**: Multiple approaches, unclear scope ("Plan migration strategy")
- **Major scope changes**: Changing direction of existing issue

**Why ask**: Ensures alignment on fuzzy work, prevents duplicate effort, clarifies scope before investment.

**Rule of thumb**: If you can write a clear, specific issue title and description in one sentence, create directly. If you need user input to clarify the work, ask first.

**Examples**:
- Create directly: "workspace MCP: Google Doc -> .docx export fails with UTF-8 encoding error"
- Create directly: "Research: Workarounds for reading Google Slides from Shared Drives"
- Ask first: "Should we refactor the auth system now or later?" (strategic decision)
- Ask first: "I found several data validation issues, should I file them all?" (potential overwhelming)

## The Transition Point

Most work starts with an implicit mental model:

**"This looks straightforward"** → Task tools

**As work progresses:**

**Stays straightforward** → Continue with Task tools, complete in session

**Complexity emerges** → Transition to bd, preserve context

The skill is recognizing the transition point:

**Transition signals:**
- "This is taking longer than expected"
- "I've discovered a blocker"
- "This needs more research"
- "I should pause this and investigate X first"
- "The user might not be available to continue today"
- "I found three related issues while working on this"

**When you notice these signals**: Create bd issue, preserve context, work from structured foundation.

## Summary Heuristics

Quick decision guides:

**Time horizon:**
- Same session → Task tools
- Multiple sessions → bd

**Dependency structure:**
- Linear steps → Task tools
- Blockers/prerequisites → bd

**Scope clarity:**
- Well-defined → Task tools
- Exploratory → bd

**Context complexity:**
- Conversation has everything → Task tools
- External context needed → bd

**User interaction:**
- User watching progress → Task tools visible in chat
- Background work → bd invisible structure

**Resume difficulty:**
- Easy from markdown → Task tools
- Need structured history → bd

When in doubt: **Use the 2-week test**. If you'd struggle to resume this work after 2 weeks without bd, use bd.

### CLI_REFERENCE.md

# CLI Command Reference

**For:** AI agents and developers using bd command-line interface
**Version:** 0.21.0+

## Quick Navigation

- [Basic Operations](#basic-operations)
- [Issue Management](#issue-management)
- [Dependencies & Labels](#dependencies--labels)
- [Filtering & Search](#filtering--search)
- [Advanced Operations](#advanced-operations)
- [Database Management](#database-management)

## Basic Operations

### Check Status

```bash
# Check database path and daemon status
bd info --json

# Example output:
# {
#   "database_path": "/path/to/.beads/beads.db",
#   "issue_prefix": "bd",
#   "daemon_running": true
# }
```

### Find Work

```bash
# Find ready work (no blockers)
bd ready --json

# Find stale issues (not updated recently)
bd stale --days 30 --json                    # Default: 30 days
bd stale --days 90 --status in_progress --json  # Filter by status
bd stale --limit 20 --json                   # Limit results
```

## Issue Management

### Create Issues

```bash
# Basic creation
# IMPORTANT: Always quote titles and descriptions with double quotes
bd create "Issue title" -t bug|feature|task -p 0-4 -d "Description" --json

# Create with explicit ID (for parallel workers)
bd create "Issue title" --id worker1-100 -p 1 --json

# Create with labels (--labels or --label work)
bd create "Issue title" -t bug -p 1 -l bug,critical --json
bd create "Issue title" -t bug -p 1 --label bug,critical --json

# Examples with special characters (all require quoting):
bd create "Fix: auth doesn't validate tokens" -t bug -p 1 --json
bd create "Add support for OAuth 2.0" -d "Implement RFC 6749 (OAuth 2.0 spec)" --json

# Create multiple issues from markdown file
bd create -f feature-plan.md --json

# Create epic with hierarchical child tasks
bd create "Auth System" -t epic -p 1 --json         # Returns: bd-a3f8e9
bd create "Login UI" -p 1 --json                     # Auto-assigned: bd-a3f8e9.1
bd create "Backend validation" -p 1 --json           # Auto-assigned: bd-a3f8e9.2
bd create "Tests" -p 1 --json                        # Auto-assigned: bd-a3f8e9.3

# Create and link discovered work (one command)
bd create "Found bug" -t bug -p 1 --deps discovered-from:<parent-id> --json
```

### Update Issues

```bash
# Update one or more issues
bd update <id> [<id>...] --status in_progress --json
bd update <id> [<id>...] --priority 1 --json

# Edit issue fields in $EDITOR (HUMANS ONLY - not for agents)
# NOTE: This command is intentionally NOT exposed via the MCP server
# Agents should use 'bd update' with field-specific parameters instead
bd edit <id>                    # Edit description
bd edit <id> --title            # Edit title
bd edit <id> --design           # Edit design notes
bd edit <id> --notes            # Edit notes
bd edit <id> --acceptance       # Edit acceptance criteria
```

### Close/Reopen Issues

```bash
# Complete work (supports multiple IDs)
bd close <id> [<id>...] --reason "Done" --json

# Reopen closed issues (supports multiple IDs)
bd reopen <id> [<id>...] --reason "Reopening" --json
```

### View Issues

```bash
# Show dependency tree
bd dep tree <id>

# Get issue details (supports multiple IDs)
bd show <id> [<id>...] --json
```

## Dependencies & Labels

### Dependencies

```bash
# Link discovered work (old way - two commands)
bd dep add <discovered-id> <parent-id> --type discovered-from

# Create and link in one command (new way - preferred)
bd create "Issue title" -t bug -p 1 --deps discovered-from:<parent-id> --json
```

### Labels

```bash
# Label management (supports multiple IDs)
bd label add <id> [<id>...] <label> --json
bd label remove <id> [<id>...] <label> --json
bd label list <id> --json
bd label list-all --json
```

## Filtering & Search

### Basic Filters

```bash
# Filter by status, priority, type
bd list --status open --priority 1 --json               # Status and priority
bd list --assignee alice --json                         # By assignee
bd list --type bug --json                               # By issue type
bd list --id bd-123,bd-456 --json                       # Specific IDs
```

### Label Filters

```bash
# Labels (AND: must have ALL)
bd list --label bug,critical --json

# Labels (OR: has ANY)
bd list --label-any frontend,backend --json
```

### Text Search

```bash
# Title search (substring)
bd list --title "auth" --json

# Pattern matching (case-insensitive substring)
bd list --title-contains "auth" --json                  # Search in title
bd list --desc-contains "implement" --json              # Search in description
bd list --notes-contains "TODO" --json                  # Search in notes
```

### Date Range Filters

```bash
# Date range filters (YYYY-MM-DD or RFC3339)
bd list --created-after 2024-01-01 --json               # Created after date
bd list --created-before 2024-12-31 --json              # Created before date
bd list --updated-after 2024-06-01 --json               # Updated after date
bd list --updated-before 2024-12-31 --json              # Updated before date
bd list --closed-after 2024-01-01 --json                # Closed after date
bd list --closed-before 2024-12-31 --json               # Closed before date
```

### Empty/Null Checks

```bash
# Empty/null checks
bd list --empty-description --json                      # Issues with no description
bd list --no-assignee --json                            # Unassigned issues
bd list --no-labels --json                              # Issues with no labels
```

### Priority Ranges

```bash
# Priority ranges
bd list --priority-min 0 --priority-max 1 --json        # P0 and P1 only
bd list --priority-min 2 --json                         # P2 and below
```

### Combine Filters

```bash
# Combine multiple filters
bd list --status open --priority 1 --label-any urgent,critical --no-assignee --json
```

## Global Flags

Global flags work with any bd command and must appear **before** the subcommand.

### Sandbox Mode

**Auto-detection (v0.21.1+):** bd automatically detects sandboxed environments and enables sandbox mode.

When detected, you'll see: `ℹ️  Sandbox detected, using direct mode`

**Manual override:**

```bash
# Explicitly enable sandbox mode
bd --sandbox <command>

# Equivalent to combining these flags:
bd --no-daemon --no-auto-flush --no-auto-import <command>
```

**What it does:**
- Disables daemon (uses direct SQLite mode)
- Disables auto-export to JSONL
- Disables auto-import from JSONL

**When to use:** Sandboxed environments where daemon can't be controlled (permission restrictions), or when auto-detection doesn't trigger.

### Staleness Control

```bash
# Skip staleness check (emergency escape hatch)
bd --allow-stale <command>

# Example: access database even if out of sync with JSONL
bd --allow-stale ready --json
bd --allow-stale list --status open --json
```

**Shows:** `⚠️  Staleness check skipped (--allow-stale), data may be out of sync`

**⚠️ Caution:** May show stale or incomplete data. Use only when stuck and other options fail.

### Force Import

```bash
# Force metadata update even when DB appears synced
bd import --force -i .beads/issues.jsonl
```

**When to use:** `bd import` reports "0 created, 0 updated" but staleness errors persist.

**Shows:** `Metadata updated (database already in sync with JSONL)`

### Other Global Flags

```bash
# JSON output for programmatic use
bd --json <command>

# Force direct mode (bypass daemon)
bd --no-daemon <command>

# Disable auto-sync
bd --no-auto-flush <command>    # Disable auto-export to JSONL
bd --no-auto-import <command>   # Disable auto-import from JSONL

# Custom database path
bd --db /path/to/.beads/beads.db <command>

# Custom actor for audit trail
bd --actor alice <command>
```

**See also:**
- [TROUBLESHOOTING.md - Sandboxed environments](TROUBLESHOOTING.md#sandboxed-environments-codex-claude-code-etc) for detailed sandbox troubleshooting
- [DAEMON.md](DAEMON.md) for daemon mode details

## Advanced Operations

### Cleanup

```bash
# Clean up closed issues (bulk deletion)
bd admin cleanup --force --json                                   # Delete ALL closed issues
bd admin cleanup --older-than 30 --force --json                   # Delete closed >30 days ago
bd admin cleanup --dry-run --json                                 # Preview what would be deleted
bd admin cleanup --older-than 90 --cascade --force --json         # Delete old + dependents
```

### Duplicate Detection & Merging

```bash
# Find and merge duplicate issues
bd duplicates                                          # Show all duplicates
bd duplicates --auto-merge                             # Automatically merge all
bd duplicates --dry-run                                # Preview merge operations

# Merge specific duplicate issues
bd merge <source-id...> --into <target-id> --json      # Consolidate duplicates
bd merge bd-42 bd-43 --into bd-41 --dry-run            # Preview merge
```

### Compaction (Memory Decay)

```bash
# Agent-driven compaction
bd admin compact --analyze --json                           # Get candidates for review
bd admin compact --analyze --tier 1 --limit 10 --json       # Limited batch
bd admin compact --apply --id bd-42 --summary summary.txt   # Apply compaction
bd admin compact --apply --id bd-42 --summary - < summary.txt  # From stdin
bd admin compact --stats --json                             # Show statistics

# Legacy AI-powered compaction (requires ANTHROPIC_API_KEY)
bd admin compact --auto --dry-run --all                     # Preview
bd admin compact --auto --all --tier 1                      # Auto-compact tier 1

# Restore compacted issue from git history
bd restore <id>  # View full history at time of compaction
```

### Rename Prefix

```bash
# Rename issue prefix (e.g., from 'knowledge-work-' to 'kw-')
bd rename-prefix kw- --dry-run  # Preview changes
bd rename-prefix kw- --json     # Apply rename
```

## Database Management

### Import/Export

```bash
# Import issues from JSONL
bd import -i .beads/issues.jsonl --dry-run      # Preview changes
bd import -i .beads/issues.jsonl                # Import and update issues
bd import -i .beads/issues.jsonl --dedupe-after # Import + detect duplicates

# Handle missing parents during import
bd import -i issues.jsonl --orphan-handling allow      # Default: import orphans without validation
bd import -i issues.jsonl --orphan-handling resurrect  # Auto-resurrect deleted parents as tombstones
bd import -i issues.jsonl --orphan-handling skip       # Skip orphans with warning
bd import -i issues.jsonl --orphan-handling strict     # Fail if parent is missing

# Configure default orphan handling behavior
bd config set import.orphan_handling "resurrect"
bd sync  # Now uses resurrect mode by default
```

**Orphan handling modes:**

- **`allow` (default)** - Import orphaned children without parent validation. Most permissive, ensures no data loss even if hierarchy is temporarily broken.
- **`resurrect`** - Search JSONL history for deleted parents and recreate them as tombstones (Status=Closed, Priority=4). Preserves hierarchy with minimal data. Dependencies are also resurrected on best-effort basis.
- **`skip`** - Skip orphaned children with warning. Partial import succeeds but some issues are excluded.
- **`strict`** - Fail import immediately if a child's parent is missing. Use when database integrity is critical.

**When to use:**
- Use `allow` (default) for daily imports and auto-sync
- Use `resurrect` when importing from databases with deleted parents
- Use `strict` for controlled imports requiring guaranteed parent existence
- Use `skip` rarely - only for selective imports

See [CONFIG.md](CONFIG.md#example-import-orphan-handling) and [TROUBLESHOOTING.md](TROUBLESHOOTING.md#import-fails-with-missing-parent-errors) for more details.

### Migration

```bash
# Migrate databases after version upgrade
bd migrate                                             # Detect and migrate old databases
bd migrate --dry-run                                   # Preview migration
bd migrate --cleanup --yes                             # Migrate and remove old files

# AI-supervised migration (check before running bd migrate)
bd migrate --inspect --json                            # Show migration plan for AI agents
bd info --schema --json                                # Get schema, tables, config, sample IDs
```

**Migration workflow for AI agents:**

1. Run `--inspect` to see pending migrations and warnings
2. Check for `missing_config` (like issue_prefix)
3. Review `invariants_to_check` for safety guarantees
4. If warnings exist, fix config issues first
5. Then run `bd migrate` safely

**Migration safety invariants:**

- **required_config_present**: Ensures issue_prefix and schema_version are set
- **foreign_keys_valid**: No orphaned dependencies or labels
- **issue_count_stable**: Issue count doesn't decrease unexpectedly

These invariants prevent data loss and would have caught issues like GH #201 (missing issue_prefix after migration).

### Daemon Management

See [docs/DAEMON.md](DAEMON.md) for complete daemon management reference.

```bash
# List all running daemons
bd daemons list --json

# Check health (version mismatches, stale sockets)
bd daemons health --json

# Stop/restart specific daemon
bd daemons stop /path/to/workspace --json
bd daemons restart 12345 --json  # By PID

# View daemon logs
bd daemons logs /path/to/workspace -n 100
bd daemons logs 12345 -f  # Follow mode

# Stop all daemons
bd daemons killall --json
bd daemons killall --force --json  # Force kill if graceful fails
```

### Sync Operations

```bash
# Manual sync (force immediate export/import/commit/push)
bd sync

# What it does:
# 1. Export pending changes to JSONL
# 2. Commit to git
# 3. Pull from remote
# 4. Import any updates
# 5. Push to remote
```

## Issue Types

- `bug` - Something broken that needs fixing
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature composed of multiple issues (supports hierarchical children)
- `chore` - Maintenance work (dependencies, tooling)

**Hierarchical children:** Epics can have child issues with dotted IDs (e.g., `bd-a3f8e9.1`, `bd-a3f8e9.2`). Children are auto-numbered sequentially. Up to 3 levels of nesting supported.

## Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (nice-to-have features, minor bugs)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

## Dependency Types

- `blocks` - Hard dependency (issue X blocks issue Y)
- `related` - Soft relationship (issues are connected)
- `parent-child` - Epic/subtask relationship
- `discovered-from` - Track issues discovered during work

Only `blocks` dependencies affect the ready work queue.

**Note:** When creating an issue with a `discovered-from` dependency, the new issue automatically inherits the parent's `source_repo` field.

## Output Formats

### JSON Output (Recommended for Agents)

Always use `--json` flag for programmatic use:

```bash
# Single issue
bd show bd-42 --json

# List of issues
bd ready --json

# Operation result
bd create "Issue" -p 1 --json
```

### Human-Readable Output

Default output without `--json`:

```bash
bd ready
# bd-42  Fix authentication bug  [P1, bug, in_progress]
# bd-43  Add user settings page  [P2, feature, open]
```

## Common Patterns for AI Agents

### Claim and Complete Work

```bash
# 1. Find available work
bd ready --json

# 2. Claim issue
bd update bd-42 --status in_progress --json

# 3. Work on it...

# 4. Close when done
bd close bd-42 --reason "Implemented and tested" --json
```

### Discover and Link Work

```bash
# While working on bd-100, discover a bug

# Old way (two commands):
bd create "Found auth bug" -t bug -p 1 --json  # Returns bd-101
bd dep add bd-101 bd-100 --type discovered-from

# New way (one command):
bd create "Found auth bug" -t bug -p 1 --deps discovered-from:bd-100 --json
```

### Batch Operations

```bash
# Update multiple issues at once
bd update bd-41 bd-42 bd-43 --priority 0 --json

# Close multiple issues
bd close bd-41 bd-42 bd-43 --reason "Batch completion" --json

# Add label to multiple issues
bd label add bd-41 bd-42 bd-43 urgent --json
```

### Session Workflow

```bash
# Start of session
bd ready --json  # Find work

# During session
bd create "..." -p 1 --json
bd update bd-42 --status in_progress --json
# ... work ...

# End of session (IMPORTANT!)
bd sync  # Force immediate sync, bypass debounce
```

**ALWAYS run `bd sync` at end of agent sessions** to ensure changes are committed/pushed immediately.

## See Also

- [AGENTS.md](../AGENTS.md) - Main agent workflow guide
- [DAEMON.md](DAEMON.md) - Daemon management and event-driven mode
- [GIT_INTEGRATION.md](GIT_INTEGRATION.md) - Git workflows and merge strategies
- [LABELS.md](../LABELS.md) - Label system guide
- [README.md](../README.md) - User documentation

### DEPENDENCIES.md

# Dependency Types Guide

Deep dive into bd's four dependency types: blocks, related, parent-child, and discovered-from.

## Contents

- [Overview](#overview) - Four types at a glance, which affect bd ready?
- [blocks - Hard Blocker](#blocks---hard-blocker)
  - [When to Use](#when-to-use) - Prerequisites, sequential steps, build order
  - [When NOT to Use](#when-not-to-use) - Soft preferences, parallel work
  - [Examples](#examples) - API development, migrations, library dependencies
  - [Creating blocks Dependencies](#creating-blocks-dependencies)
  - [Common Patterns](#common-patterns) - Build foundation first, migration sequences, testing gates
  - [Automatic Unblocking](#automatic-unblocking)
- [related - Soft Link](#related---soft-link)
  - [When to Use](#when-to-use-1) - Context, related features, parallel work
  - [When NOT to Use](#when-not-to-use-1)
  - [Examples](#examples-1) - Feature context, research links, parallel development
  - [Creating related Dependencies](#creating-related-dependencies)
  - [Common Patterns](#common-patterns-1) - Context clusters, research threads, feature families
- [parent-child - Hierarchical](#parent-child---hierarchical)
  - [When to Use](#when-to-use-2) - Epics/subtasks, phases
  - [When NOT to Use](#when-not-to-use-2)
  - [Examples](#examples-2) - Epic with subtasks, phased projects
  - [Creating parent-child Dependencies](#creating-parent-child-dependencies)
  - [Combining with blocks](#combining-with-blocks)
  - [Common Patterns](#common-patterns-2) - Epic decomposition, nested hierarchies
- [discovered-from - Provenance](#discovered-from---provenance)
  - [When to Use](#when-to-use-3) - Side quests, research findings
  - [Why This Matters](#why-this-matters)
  - [Examples](#examples-3) - Bug discovered during feature work, research branches
  - [Creating discovered-from Dependencies](#creating-discovered-from-dependencies)
  - [Common Patterns](#common-patterns-3) - Discovery during implementation, research expansion
  - [Combining with blocks](#combining-with-blocks-1)
- [Decision Guide](#decision-guide)
  - [Decision Tree](#decision-tree)
  - [Quick Reference by Situation](#quick-reference-by-situation)
- [Common Mistakes](#common-mistakes)
  - Using blocks for preferences, using discovered-from for planning, not using dependencies, over-using blocks, wrong direction
- [Advanced Patterns](#advanced-patterns)
  - Diamond dependencies, optional dependencies, discovery cascade, epic with phases
- [Visualization](#visualization)
- [Summary](#summary)

## Overview

bd supports four dependency types that serve different purposes in organizing and tracking work:

| Type | Purpose | Affects `bd ready`? | Common Use |
|------|---------|---------------------|------------|
| **blocks** | Hard blocker | Yes - blocked issues excluded | Sequential work, prerequisites |
| **related** | Soft link | No - just informational | Context, related work |
| **parent-child** | Hierarchy | No - structural only | Epics and subtasks |
| **discovered-from** | Provenance | No - tracks origin | Side quests, research findings |

**Key insight**: Only `blocks` dependencies affect what work is ready. The other three provide structure and context.

---

## blocks - Hard Blocker

**Semantics**: Issue A blocks issue B. B cannot start until A is complete.

**Effect**: Issue B disappears from `bd ready` until issue A is closed.

### When to Use

Use `blocks` when work literally cannot proceed:

- **Prerequisites**: Database schema must exist before endpoints can use it
- **Sequential steps**: Migration step 1 must complete before step 2
- **Build order**: Foundation must be done before building on top
- **Technical blockers**: Library must be installed before code can use it

### When NOT to Use

Don't use `blocks` for:

- **Soft preferences**: "Should do X before Y but could do either"
- **Parallel work**: Both can proceed independently
- **Information links**: Just want to note relationship
- **Recommendations**: "Would be better if done in this order"

Use `related` instead for soft connections.

### Examples

**Example 1: API Development**

```
db-schema-1: "Create users table"
  blocks
api-endpoint-2: "Add GET /users endpoint"

Why: Endpoint literally needs table to exist
Effect: api-endpoint-2 won't show in bd ready until db-schema-1 closed
```

**Example 2: Migration Sequence**

```
migrate-1: "Backup production database"
  blocks
migrate-2: "Run schema migration"
  blocks
migrate-3: "Verify data integrity"

Why: Each step must complete before next can safely proceed
Effect: bd ready shows only migrate-1; closing it reveals migrate-2, etc.
```

**Example 3: Library Installation**

```
setup-1: "Install JWT library"
  blocks
auth-2: "Implement JWT validation"

Why: Code won't compile/run without library
Effect: Can't start auth-2 until setup-1 complete
```

### Creating blocks Dependencies

```bash
bd dep add prerequisite-issue blocked-issue
# or explicitly:
bd dep add prerequisite-issue blocked-issue --type blocks
```

**Direction matters**: `from_id` blocks `to_id`. Think: "prerequisite blocks dependent".

### Common Patterns

**Pattern: Build Foundation First**

```
foundation-1: "Set up authentication system"
  blocks all of:
    - feature-2: "Add user profiles"
    - feature-3: "Add admin panel"
    - feature-4: "Add API access"

One foundational issue blocks multiple dependent features.
```

**Pattern: Sequential Pipeline**

```
step-1 blocks step-2 blocks step-3 blocks step-4

Linear chain where each step depends on previous.
bd ready shows only current step.
```

**Pattern: Parallel Then Merge**

```
research-1: "Investigate option A"
research-2: "Investigate option B"
research-3: "Investigate option C"
All three block:
  decision-4: "Choose approach based on research"

Multiple parallel tasks must complete before next step.
```

### Automatic Unblocking

When you close an issue that's blocking others:

```
1. Close db-schema-1
2. bd automatically updates: api-endpoint-2 is now ready
3. bd ready shows api-endpoint-2
4. No manual unblocking needed
```

This is why `blocks` is powerful - bd maintains ready state automatically.

---

## related - Soft Link

**Semantics**: Issues are related but neither blocks the other.

**Effect**: No impact on `bd ready`. Pure informational link.

### When to Use

Use `related` for context and discoverability:

- **Similar work**: "These tackle the same problem from different angles"
- **Shared context**: "Working on one provides insight for the other"
- **Alternative approaches**: "These are different ways to solve X"
- **Complementary features**: "These work well together but aren't required"

### When NOT to Use

Don't use `related` if:

- One actually blocks the other → use `blocks`
- One discovered the other → use `discovered-from`
- One is parent of the other → use `parent-child`

### Examples

**Example 1: Related Refactoring**

```
refactor-1: "Extract validation logic"
  related to
refactor-2: "Extract error handling logic"

Why: Both are refactoring efforts, similar patterns, but independent
Effect: None on ready state; just notes the relationship
```

**Example 2: Documentation and Code**

```
feature-1: "Add OAuth login"
  related to
docs-2: "Document OAuth setup"

Why: Docs and feature go together, but can be done in any order
Effect: Can work on either whenever; just notes they're connected
```

**Example 3: Alternative Approaches**

```
perf-1: "Investigate Redis caching"
  related to
perf-2: "Investigate CDN caching"

Why: Both address performance, different approaches, explore both
Effect: Both show in bd ready; choosing one doesn't block the other
```

### Creating related Dependencies

```bash
bd dep add issue-1 issue-2 --type related
```

**Direction doesn't matter** for `related` - it's a symmetric link.

### Common Patterns

**Pattern: Cluster Related Work**

```
api-redesign related to:
  - api-docs-update
  - api-client-update
  - api-tests-update
  - api-versioning

Group of issues all related to API work.
Use related to show they're part of same initiative.
```

**Pattern: Cross-Cutting Concerns**

```
security-audit related to:
  - auth-module
  - api-endpoints
  - database-access
  - frontend-forms

Security audit touches multiple areas.
Related links show what areas it covers.
```

---

## parent-child - Hierarchical

**Semantics**: Issue A is parent of issue B. Typically A is an epic, B is a subtask.

**Effect**: No impact on `bd ready`. Creates hierarchical structure.

### When to Use

Use `parent-child` for breaking down large work:

- **Epics and subtasks**: Big feature split into smaller pieces
- **Hierarchical organization**: Logical grouping of related tasks
- **Progress tracking**: See completion of children relative to parent
- **Work breakdown structure**: Decompose complex work

### When NOT to Use

Don't use `parent-child` if:

- Siblings need ordering → add `blocks` between children
- Relationship is equality → use `related`
- Just discovered one from the other → use `discovered-from`

### Examples

**Example 1: Feature Epic**

```
oauth-epic: "Implement OAuth integration" (epic)
  parent of:
    - oauth-1: "Set up OAuth credentials" (task)
    - oauth-2: "Implement authorization flow" (task)
    - oauth-3: "Add token refresh" (task)
    - oauth-4: "Create login UI" (task)

Why: Epic decomposed into implementable tasks
Effect: Hierarchical structure; all show in bd ready (unless blocked)
```

**Example 2: Research with Findings**

```
research-epic: "Investigate caching strategies" (epic)
  parent of:
    - research-1: "Redis evaluation"
    - research-2: "Memcached evaluation"
    - research-3: "CDN evaluation"
    - decision-4: "Choose caching approach"

Why: Research project with multiple investigation threads
Effect: Can track progress across all investigations
```

### Creating parent-child Dependencies

```bash
bd dep add child-task-id parent-epic-id --type parent-child
```

**Direction matters**: The child depends on the parent. Think: "child depends on parent" or "task is part of epic".

### Combining with blocks

Parent-child gives structure; blocks gives ordering:

```
auth-epic (parent of all)
  ├─ auth-1: "Install library"
  ├─ auth-2: "Create middleware" (blocked by auth-1)
  ├─ auth-3: "Add endpoints" (blocked by auth-2)
  └─ auth-4: "Add tests" (blocked by auth-3)

parent-child: Shows these are all part of auth epic
blocks: Shows they must be done in order
```

### Common Patterns

**Pattern: Epic with Independent Subtasks**

```
Epic with no ordering between children:
All children show in bd ready immediately.
Work on any child in any order.
Close epic when all children complete.
```

**Pattern: Epic with Sequential Subtasks**

```
Epic with blocks dependencies between children:
bd ready shows only first child.
Closing each child unblocks next.
Epic provides structure, blocks provides order.
```

**Pattern: Nested Epics**

```
major-epic
  ├─ sub-epic-1
  │   ├─ task-1a
  │   └─ task-1b
  └─ sub-epic-2
      ├─ task-2a
      └─ task-2b

Multiple levels of hierarchy for complex projects.
```

---

## discovered-from - Provenance

**Semantics**: Issue B was discovered while working on issue A.

**Effect**: No impact on `bd ready`. Tracks origin and provides context.

### When to Use

Use `discovered-from` to preserve discovery context:

- **Side quests**: Found new work during implementation
- **Research findings**: Discovered issue while investigating
- **Bug found during feature work**: Context of discovery matters
- **Follow-up work**: Identified next steps during current work

### Why This Matters

Knowing where an issue came from helps:

- **Understand context**: Why was this created?
- **Reconstruct thinking**: What led to this discovery?
- **Assess relevance**: Is this still important given original context?
- **Track exploration**: See what emerged from research

### Examples

**Example 1: Bug During Feature**

```
feature-10: "Add user profiles"
  discovered-from leads to
bug-11: "Existing auth doesn't handle profile permissions"

Why: While adding profiles, discovered auth system inadequate
Context: Bug might not exist if profiles weren't being added
```

**Example 2: Research Findings**

```
research-5: "Investigate caching options"
  discovered-from leads to
finding-6: "Redis supports persistence unlike Memcached"
finding-7: "CDN caching incompatible with our auth model"
decision-8: "Choose Redis based on findings"

Why: Research generated specific findings
Context: Findings only relevant in context of research question
```

**Example 3: Refactoring Reveals Technical Debt**

```
refactor-20: "Extract validation logic"
  discovered-from leads to
debt-21: "Validation inconsistent across controllers"
debt-22: "No validation for edge cases"
improvement-23: "Could add validation library"

Why: Refactoring work revealed multiple related issues
Context: Issues discovered as side effect of refactoring
```

### Creating discovered-from Dependencies

```bash
bd dep add original-work-id discovered-issue-id --type discovered-from
```

**Direction matters**: `to_id` was discovered while working on `from_id`.

### Common Patterns

**Pattern: Exploration Tree**

```
spike-1: "Investigate API redesign"
  discovered-from →
    finding-2: "Current API mixes REST and GraphQL"
    finding-3: "Authentication not consistent"
    finding-4: "Rate limiting missing"

One exploration generates multiple findings.
Tree structure shows exploration process.
```

**Pattern: Bug Investigation Chain**

```
bug-1: "Login fails intermittently"
  discovered-from →
    bug-2: "Race condition in session creation"
      discovered-from →
        bug-3: "Database connection pool too small"

Investigation of one bug reveals root cause as another bug.
Chain shows how you got from symptom to cause.
```

**Pattern: Feature Implementation Side Quests**

```
feature-main: "Add shopping cart"
  discovered-from →
    improvement-a: "Product images should be cached"
    bug-b: "Price formatting wrong for some locales"
    debt-c: "Inventory system needs refactoring"

Main feature work generates tangential discoveries.
Captured for later without derailing main work.
```

### Combining with blocks

Can use both together:

```
feature-10: "Add user profiles"
  discovered-from →
    bug-11: "Auth system needs role-based access"
      blocks →
        feature-10: "Add user profiles"

Discovery: Found bug during feature work
Assessment: Bug actually blocks feature
Actions: Mark feature blocked, work on bug first
```

---

## Decision Guide

**"Which dependency type should I use?"**

### Decision Tree

```
Does Issue A prevent Issue B from starting?
  YES → blocks
  NO ↓

Is Issue B a subtask of Issue A?
  YES → parent-child (A parent, B child)
  NO ↓

Was Issue B discovered while working on Issue A?
  YES → discovered-from (A original, B discovered)
  NO ↓

Are Issues A and B just related?
  YES → related
```

### Quick Reference by Situation

| Situation | Use |
|-----------|-----|
| B needs A complete to start | blocks |
| B is part of A (epic/task) | parent-child |
| Found B while working on A | discovered-from |
| A and B are similar/connected | related |
| B should come after A but could start | related + note |
| A and B are alternatives | related |
| B is follow-up to A | discovered-from |

---

## Common Mistakes

### Mistake 1: Using blocks for Preferences

**Wrong**:
```
docs-1: "Update documentation"
  blocks
feature-2: "Add new feature"

Reason: "We prefer to update docs first"
```

**Problem**: Documentation doesn't actually block feature implementation.

**Right**: Use `related` or don't link at all. If you want ordering, note it in issue descriptions but don't enforce with blocks.

### Mistake 2: Using discovered-from for Planning

**Wrong**:
```
epic-1: "OAuth integration"
  discovered-from →
    task-2: "Set up OAuth credentials"

Reason: "I'm planning these tasks from the epic"
```

**Problem**: `discovered-from` is for emergent discoveries, not planned decomposition.

**Right**: Use `parent-child` for planned task breakdown.

### Mistake 3: Not Using Any Dependencies

**Symptom**: Long list of issues with no structure.

**Problem**: Can't tell what's blocked, what's related, how work is organized.

**Solution**: Add structure with dependencies:
- Group with parent-child
- Order with blocks
- Link with related
- Track discovery with discovered-from

### Mistake 4: Over-Using blocks

**Wrong**:
```
Everything blocks everything else in strict sequential order.
```

**Problem**: No parallel work possible; `bd ready` shows only one issue.

**Right**: Only use `blocks` for actual technical dependencies. Allow parallel work where possible.

### Mistake 5: Wrong Direction

**Wrong**:
```bash
bd dep add api-endpoint database-schema

Meaning: api-endpoint blocks database-schema
```

**Problem**: Backwards! Schema should block endpoint, not other way around.

**Right**:
```bash
bd dep add database-schema api-endpoint

Meaning: database-schema blocks api-endpoint
```

**Mnemonic**: "from_id blocks to_id" or "prerequisite blocks dependent"

---

## Advanced Patterns

### Pattern: Diamond Dependencies

```
        setup
       /    \
   impl-a  impl-b
       \    /
       testing

setup blocks both impl-a and impl-b
both impl-a and impl-b block testing
```

Both implementations must complete before testing can begin.

### Pattern: Optional Dependencies

```
core-feature (ready immediately)
  related to
nice-to-have (ready immediately)

Both can be done, neither blocks the other.
Use related to show they're connected.
```

### Pattern: Discovery Cascade

```
research-main
  discovered-from → finding-1
  discovered-from → finding-2
    discovered-from → deep-finding-3

Research generates findings.
Findings generate deeper findings.
Tree shows discovery process.
```

### Pattern: Epic with Phases

```
auth-epic
  parent of phase-1-epic
    parent of: setup-1, setup-2, setup-3
  parent of phase-2-epic
    parent of: implement-1, implement-2
  parent of phase-3-epic
    parent of: test-1, test-2

phase-1-epic blocks phase-2-epic blocks phase-3-epic

Nested hierarchy with phase ordering.
```

---

## Visualization

When you run `bd show issue-id` on an issue, you see:

```
Issue: feature-10
Dependencies (blocks this issue):
  - setup-5: "Install library"
  - config-6: "Add configuration"

Dependents (blocked by this issue):
  - test-12: "Add integration tests"
  - docs-13: "Document new feature"

Related:
  - refactor-8: "Similar refactoring effort"

Discovered from:
  - research-3: "API investigation"
```

This shows the full dependency context for an issue.

---

## Summary

**Four dependency types, four different purposes:**

1. **blocks**: Sequential work, prerequisites, hard blockers
   - Affects bd ready
   - Use for technical dependencies only

2. **related**: Context, similar work, soft connections
   - Informational only
   - Use liberally for discoverability

3. **parent-child**: Epics and subtasks, hierarchical structure
   - Organizational only
   - Use for work breakdown

4. **discovered-from**: Side quests, research findings, provenance
   - Context preservation
   - Use to track emergence

**Key insight**: Only `blocks` affects what work is ready. The other three provide rich context without constraining execution.

Use dependencies to create a graph that:
- Automatically maintains ready work
- Preserves discovery context
- Shows project structure
- Links related work

This graph becomes the persistent memory that survives compaction and enables long-horizon agent work.

### INTEGRATION_PATTERNS.md

# Integration Patterns with Other Skills

How bd-issue-tracking integrates with Task tools (TaskCreate, TaskUpdate, TaskList, TaskGet), writing-plans, and other skills for optimal workflow.

## Contents

- [Task Tools Integration](#task-tools-integration) - Temporal layering pattern
- [writing-plans Integration](#writing-plans-integration) - Detailed implementation plans
- [Cross-Skill Workflows](#cross-skill-workflows) - Using multiple skills together
- [Decision Framework](#decision-framework) - When to use which tool

---

## Task Tools Integration

**Both tools complement each other at different timescales:**

### Temporal Layering Pattern

**Task tools** (short-term working memory - this hour):
- Tactical execution: "Review Section 3", "Expand Q&A answers"
- Marked completed via TaskUpdate as you go
- Present/future tense ("Review", "Expand", "Create")
- Ephemeral: Task list resets when session ends

**Beads** (long-term episodic memory - this week/month):
- Strategic objectives: "Continue work on strategic planning document"
- Key decisions and outcomes in notes field
- Past tense in notes ("COMPLETED", "Discovered", "Blocked by")
- Persistent: Survives compaction and session boundaries

**Key insight**: Task tools = working copy for the current hour. Beads = project journal for the current month.

### The Handoff Pattern

1. **Session start**: Read bead → Create tasks via TaskCreate for immediate actions
2. **During work**: Mark tasks completed via TaskUpdate as you go
3. **Reach milestone**: Update bead notes with outcomes + context
4. **Session end**: Task list resets, bead survives with enriched notes

**After compaction**: Task list is gone, but bead notes reconstruct what happened.

### Example: Task tools track execution, Beads capture meaning

**Task tools (ephemeral execution view):**
```
Create session tasks:
  TaskCreate: "Implement login endpoint" (pending)
  TaskCreate: "Add password hashing with bcrypt" (pending)
  TaskCreate: "Create session middleware" (pending)
Mark completed via TaskUpdate as you go. Check TaskList for progress.
```

**Corresponding bead notes (persistent context):**
```bash
bd update issue-123 --notes "COMPLETED: Login endpoint with bcrypt password
hashing (12 rounds). KEY DECISION: Using JWT tokens (not sessions) for stateless
auth - simplifies horizontal scaling. IN PROGRESS: Session middleware implementation.
NEXT: Need user input on token expiry time (1hr vs 24hr trade-off)."
```

**What's different**:
- Task tools: Task names (what to do)
- Beads: Outcomes and decisions (what was learned, why it matters)

**Don't duplicate**: Task tools track execution, Beads captures meaning and context.

### When to Update Each Tool

**Update Task tools** (frequently):
- Mark task completed via TaskUpdate as you finish each one
- Add new tasks via TaskCreate as you break down work
- Set in_progress via TaskUpdate when switching tasks

**Update Beads** (at milestones):
- Completed a significant piece of work
- Made a key decision that needs documentation
- Hit a blocker that pauses progress
- About to ask user for input
- Session token usage > 70%
- End of session

**Pattern**: Task tools change every few minutes. Beads updates every hour or at natural breakpoints.

### Full Workflow Example

**Scenario**: Implement OAuth authentication (multi-session work)

**Session 1 - Planning**:
```bash
# Create bd issue
bd create "Implement OAuth authentication" -t feature -p 0 --design "
JWT tokens with refresh rotation.
See BOUNDARIES.md for bd vs Task tools decision.
"

# Mark in_progress
bd update oauth-1 --status in_progress

# Create session tasks for today's work
Create session tasks:
  TaskCreate: "Research OAuth 2.0 refresh token flow" (pending)
  TaskCreate: "Design token schema" (pending)
  TaskCreate: "Set up test environment" (pending)
Mark completed via TaskUpdate as you go. Check TaskList for progress.
```

**End of Session 1**:
```bash
# Update bd with outcomes
bd update oauth-1 --notes "COMPLETED: Researched OAuth2 refresh flow. Decided on 7-day refresh tokens.
KEY DECISION: RS256 over HS256 (enables key rotation per security review).
IN PROGRESS: Need to set up test OAuth provider.
NEXT: Configure test provider, then implement token endpoint."

# Task list resets when session ends
```

**Session 2 - Implementation** (after compaction):
```bash
# Read bd to reconstruct context
bd show oauth-1
# See: COMPLETED research, NEXT is configure test provider

# Create fresh session tasks from NEXT
Create session tasks:
  TaskCreate: "Configure test OAuth provider" (pending)
  TaskCreate: "Implement token endpoint" (pending)
  TaskCreate: "Add basic tests" (pending)
Mark completed via TaskUpdate as you go. Check TaskList for progress.

# Work proceeds...

# Update bd at milestone
bd update oauth-1 --notes "COMPLETED: Test provider configured, token endpoint implemented.
TESTS: 5 passing (token generation, validation, expiry).
IN PROGRESS: Adding refresh token rotation.
NEXT: Implement rotation, add rate limiting, security review."
```

**For complete decision criteria and boundaries, see:** [BOUNDARIES.md](BOUNDARIES.md)

---

## writing-plans Integration

**For complex multi-step features**, the design field in bd issues can link to detailed implementation plans that break work into bite-sized RED-GREEN-REFACTOR steps.

### When to Create Detailed Plans

**Use detailed plans for:**
- Complex features with multiple components
- Multi-session work requiring systematic breakdown
- Features where TDD discipline adds value (core logic, critical paths)
- Work that benefits from explicit task sequencing

**Skip detailed plans for:**
- Simple features (single function, straightforward logic)
- Exploratory work (API testing, pattern discovery)
- Infrastructure setup (configuration, wiring)

**The test:** If you can implement it in one session without a checklist, skip the detailed plan.

### Using the writing-plans Skill

When design field needs detailed breakdown, reference the **writing-plans** skill:

**Pattern:**
```bash
# Create issue with high-level design
bd create "Implement OAuth token refresh" --design "
Add JWT refresh token flow with rotation.
See docs/plans/2025-10-23-oauth-refresh-design.md for detailed plan.
"

# Then use writing-plans skill to create detailed plan
# The skill creates: docs/plans/YYYY-MM-DD-<feature-name>.md
```

**Detailed plan structure** (from writing-plans):
- Bite-sized tasks (2-5 minutes each)
- Explicit RED-GREEN-REFACTOR steps per task
- Exact file paths and complete code
- Verification commands with expected output
- Frequent commit points

**Example task from detailed plan:**
```markdown
### Task 1: Token Refresh Endpoint

**Files:**
- Create: `src/auth/refresh.py`
- Test: `tests/auth/test_refresh.py`

**Step 1: Write failing test**
```python
def test_refresh_token_returns_new_access_token():
    refresh_token = create_valid_refresh_token()
    response = refresh_endpoint(refresh_token)
    assert response.status == 200
    assert response.access_token is not None
```

**Step 2: Run test to verify it fails**
Run: `pytest tests/auth/test_refresh.py::test_refresh_token_returns_new_access_token -v`
Expected: FAIL with "refresh_endpoint not defined"

**Step 3: Implement minimal code**
[... exact implementation ...]

**Step 4: Verify test passes**
[... verification ...]

**Step 5: Commit**
```bash
git add tests/auth/test_refresh.py src/auth/refresh.py
git commit -m "feat: add token refresh endpoint"
```
```

### Integration with bd Workflow

**Three-layer structure**:
1. **bd issue**: Strategic objective + high-level design
2. **Detailed plan** (writing-plans): Step-by-step execution guide
3. **Task tools**: Current task within the plan

**During planning phase:**
1. Create bd issue with high-level design
2. If complex: Use writing-plans skill to create detailed plan
3. Link plan in design field: `See docs/plans/YYYY-MM-DD-<topic>.md`

**During execution phase:**
1. Open detailed plan (if exists)
2. Use Task tools to track current task within plan
3. Update bd notes at milestones, not per-task
4. Close bd issue when all plan tasks complete

**Don't duplicate:** Detailed plan = execution steps. BD notes = outcomes and decisions.

**Example bd notes after using detailed plan:**
```bash
bd update oauth-5 --notes "COMPLETED: Token refresh endpoint (5 tasks from plan: endpoint + rotation + tests)
KEY DECISION: 7-day refresh tokens (vs 30-day) - reduces risk of token theft
TESTS: All 12 tests passing (auth, rotation, expiry, error handling)"
```

### When NOT to Use Detailed Plans

**Red flags:**
- Feature is simple enough to implement in one pass
- Work is exploratory (discovering patterns, testing APIs)
- Infrastructure work (OAuth setup, MCP configuration)
- Would spend more time planning than implementing

**Rule of thumb:** Use detailed plans when systematic breakdown prevents mistakes, not for ceremony.

**Pattern summary**:
- **Simple feature**: bd issue only
- **Complex feature**: bd issue + Task tools
- **Very complex feature**: bd issue + writing-plans + Task tools

---

## Cross-Skill Workflows

### Pattern: Research Document with Strategic Planning

**Scenario**: User asks "Help me write a strategic planning document for Q4"

**Tools used**: bd-issue-tracking + developing-strategic-documents skill

**Workflow**:
1. Create bd issue for tracking:
   ```bash
   bd create "Q4 strategic planning document" -t task -p 0
   bd update strat-1 --status in_progress
   ```

2. Use developing-strategic-documents skill for research and writing

3. Update bd notes at milestones:
   ```bash
   bd update strat-1 --notes "COMPLETED: Research phase (reviewed 5 competitor docs, 3 internal reports)
   KEY DECISION: Focus on market expansion over cost optimization per exec input
   IN PROGRESS: Drafting recommendations section
   NEXT: Get exec review of draft recommendations before finalizing"
   ```

4. Task tools track immediate writing tasks:
   ```
   Create session tasks:
     TaskCreate: "Draft recommendation 1: Market expansion" (pending)
     TaskCreate: "Add supporting data from research" (pending)
     TaskCreate: "Create budget estimates" (pending)
   Mark completed via TaskUpdate as you go. Check TaskList for progress.
   ```

**Why this works**: bd preserves context across sessions (document might take days), skill provides writing framework, Task tools track current work.

### Pattern: Multi-File Refactoring

**Scenario**: Refactor authentication system across 8 files

**Tools used**: bd-issue-tracking + systematic-debugging (if issues found)

**Workflow**:
1. Create epic and subtasks:
   ```bash
   bd create "Refactor auth system to use JWT" -t epic -p 0
   bd create "Update login endpoint" -t task
   bd create "Update token validation" -t task
   bd create "Update middleware" -t task
   bd create "Update tests" -t task

   # Link hierarchy
   bd dep add auth-epic login-1 --type parent-child
   bd dep add auth-epic validation-2 --type parent-child
   bd dep add auth-epic middleware-3 --type parent-child
   bd dep add auth-epic tests-4 --type parent-child

   # Add ordering
   bd dep add validation-2 login-1  # validation depends on login
   bd dep add middleware-3 validation-2  # middleware depends on validation
   bd dep add tests-4 middleware-3  # tests depend on middleware
   ```

2. Work through subtasks in order, using Task tools for each:
   ```
   Current: login-1
   Create session tasks:
     TaskCreate: "Update login route signature" (pending)
     TaskCreate: "Add JWT generation" (pending)
     TaskCreate: "Update tests" (pending)
     TaskCreate: "Verify backward compatibility" (pending)
   Mark completed via TaskUpdate as you go. Check TaskList for progress.
   ```

3. Update bd notes as each completes:
   ```bash
   bd close login-1 --reason "Updated to JWT. Tests passing. Backward compatible with session auth."
   ```

4. If issues discovered, use systematic-debugging skill + create blocker issues

**Why this works**: bd tracks dependencies and progress across files, Task tools focus on current file, skills provide specialized frameworks when needed.

---

## Decision Framework

### Which Tool for Which Purpose?

| Need | Tool | Why |
|------|------|-----|
| Track today's execution | Task tools | Lightweight, shows current progress |
| Preserve context across sessions | bd | Survives compaction, persistent memory |
| Detailed implementation steps | writing-plans | RED-GREEN-REFACTOR breakdown |
| Research document structure | developing-strategic-documents | Domain-specific framework |
| Debug complex issue | systematic-debugging | Structured debugging protocol |

### Decision Tree

```
Is this work done in this session?
├─ Yes → Use Task tools only
└─ No → Use bd
    ├─ Simple feature → bd issue + Task tools
    └─ Complex feature → bd issue + writing-plans + Task tools

Will conversation history get compacted?
├─ Likely → Use bd (context survives)
└─ Unlikely → Task tools are sufficient

Does work have dependencies or blockers?
├─ Yes → Use bd (tracks relationships)
└─ No → Task tools are sufficient

Is this specialized domain work?
├─ Research/writing → developing-strategic-documents
├─ Complex debugging → systematic-debugging
├─ Detailed implementation → writing-plans
└─ General tracking → bd + Task tools
```

### Integration Anti-Patterns

**Don't**:
- Duplicate session tasks into bd notes (different purposes)
- Create bd issues for single-session linear work (use Task tools)
- Put detailed implementation steps in bd notes (use writing-plans)
- Update bd after every task completion (update at milestones)
- Use writing-plans for exploratory work (defeats the purpose)

**Do**:
- Update bd when changing tools or reaching milestones
- Use Task tools as "working copy" of bd's NEXT section
- Link between tools (bd design field → writing-plans file path)
- Choose the right level of formality for the work complexity

---

## Summary

**Key principle**: Each tool operates at a different timescale and level of detail.

- **Task tools**: Minutes to hours (current execution)
- **bd**: Hours to weeks (persistent context)
- **writing-plans**: Days to weeks (detailed breakdown)
- **Other skills**: As needed (domain frameworks)

**Integration pattern**: Use the lightest tool sufficient for the task, add heavier tools only when complexity demands it.

**For complete boundaries and decision criteria, see:** [BOUNDARIES.md](BOUNDARIES.md)

### ISSUE_CREATION.md

# Issue Creation Guidelines

Guidance on when and how to create bd issues for maximum effectiveness.

## Contents

- [When to Ask First vs Create Directly](#when-to-ask)
- [Issue Quality](#quality)
- [Making Issues Resumable](#resumable)
- [Design vs Acceptance Criteria](#design-vs-acceptance)

## When to Ask First vs Create Directly {#when-to-ask}

### Ask the user before creating when:
- Knowledge work with fuzzy boundaries
- Task scope is unclear
- Multiple valid approaches exist
- User's intent needs clarification

### Create directly when:
- Clear bug discovered during implementation
- Obvious follow-up work identified
- Technical debt with clear scope
- Dependency or blocker found

**Why ask first for knowledge work?** Task boundaries in strategic/research work are often unclear until discussed, whereas technical implementation tasks are usually well-defined. Discussion helps structure the work properly before creating issues, preventing poorly-scoped issues that need immediate revision.

## Issue Quality {#quality}

Use clear, specific titles and include sufficient context in descriptions to resume work later.

### Field Usage

**Use --design flag for:**
- Implementation approach decisions
- Architecture notes
- Trade-offs considered

**Use --acceptance flag for:**
- Definition of done
- Testing requirements
- Success metrics

## Making Issues Resumable (Complex Technical Work) {#resumable}

For complex technical features spanning multiple sessions, enhance notes field with implementation details.

### Optional but valuable for technical work:
- Working API query code (tested, with response structure)
- Sample API responses showing actual data
- Desired output format examples (show, don't describe)
- Research context (why this approach, what was discovered)

### Example pattern:

```markdown
bd update issue-9 --notes "IMPLEMENTATION GUIDE:
WORKING CODE: service.about().get(fields='importFormats')
Returns: dict with 49 entries like {'text/markdown': [...]}
OUTPUT FORMAT: # Drive Import Formats (markdown with categorized list)
CONTEXT: text/markdown support added July 2024, not in static docs"
```

**When to add:** Multi-session technical features with APIs or specific formats. Skip for simple tasks.

**For detailed patterns and examples, read:** [RESUMABILITY.md](RESUMABILITY.md)

## Design vs Acceptance Criteria (Critical Distinction) {#design-vs-acceptance}

Common mistake: Putting implementation details in acceptance criteria. Here's the difference:

### DESIGN field (HOW to build it):
- "Use two-phase batchUpdate approach: insert text first, then apply formatting"
- "Parse with regex to find * and _ markers"
- "Use JWT tokens with 1-hour expiry"
- Trade-offs: "Chose batchUpdate over streaming API for atomicity"

### ACCEPTANCE CRITERIA (WHAT SUCCESS LOOKS LIKE):
- "Bold and italic markdown formatting renders correctly in the Doc"
- "Solution accepts markdown input and creates Doc with specified title"
- "Returns doc_id and webViewLink to caller"
- "User tokens persist across sessions and refresh automatically"

### Why this matters:
- Design can change during implementation (e.g., use library instead of regex)
- Acceptance criteria should remain stable across sessions
- Criteria should be **outcome-focused** ("what must be true?") not **step-focused** ("do these steps")
- Each criterion should be **verifiable** - you can definitively say yes/no

### The pitfall

Writing criteria like "- [ ] Use batchUpdate approach" locks you into one implementation.

Better: "- [ ] Formatting is applied atomically (all at once or not at all)" - allows flexible implementation.

### Test yourself

If you rewrote the solution using a different approach, would the acceptance criteria still apply? If not, they're design notes, not criteria.

### Example of correct structure

✅ **Design field:**
```
Two-phase Docs API approach:
1. Parse markdown to positions
2. Create doc + insert text in one call
3. Apply formatting in second call
Rationale: Atomic operations, easier to debug formatting separately
```

✅ **Acceptance criteria:**
```
- [ ] Markdown formatting renders in Doc (bold, italic, headings)
- [ ] Lists preserve order and nesting
- [ ] Links are clickable
- [ ] Large documents (>50KB) process without timeout
```

❌ **Wrong (design masquerading as criteria):**
```
- [ ] Use two-phase batchUpdate approach
- [ ] Apply formatting in second batchUpdate call
```

## Quick Reference

**Creating good issues:**

1. **Title**: Clear, specific, action-oriented
2. **Description**: Problem statement, context, why it matters
3. **Design**: Approach, architecture, trade-offs (can change)
4. **Acceptance**: Outcomes, success criteria (should be stable)
5. **Notes**: Implementation details, session handoffs (evolves over time)

**Common mistakes:**

- Vague titles: "Fix bug" → "Fix: auth token expires before refresh"
- Implementation in acceptance: "Use JWT" → "Auth tokens persist across sessions"
- Missing context: "Update database" → "Update database: add user_last_login for session analytics"

### MOLECULES.md

# Molecules and Wisps Reference

> **Status: NOT YET IMPLEMENTED — Design Spec Only**
>
> The `bd mol` commands described below are a planned feature. They do not exist yet in the bd CLI.
> This document serves as the design specification for the molecule subsystem.
> Do not reference these commands in active skill code paths.

This reference covers bd's molecular chemistry system for reusable work templates and ephemeral workflows.

> **WARNING**: Molecule/formula templates must use standard `prefix-hash` ID format.
> Never create IDs like `my-template.step-name` - these corrupt the database.
> See [ANTI_PATTERNS.md](ANTI_PATTERNS.md) for details.

## The Chemistry Metaphor

bd v0.34.0 introduces a chemistry-inspired workflow system:

| Phase | Name | Storage | Synced? | Use Case |
|-------|------|---------|---------|----------|
| **Solid** | Proto | `.beads/` | Yes | Reusable template (epic with `template` label) |
| **Liquid** | Mol | `.beads/` | Yes | Persistent instance (real issues from template) |
| **Vapor** | Wisp | `.beads-wisp/` | No | Ephemeral instance (operational work, no audit trail) |

**Phase transitions:**
- `spawn` / `pour`: Solid (proto) → Liquid (mol)
- `wisp create`: Solid (proto) → Vapor (wisp)
- `squash`: Vapor (wisp) → Digest (permanent summary)
- `burn`: Vapor (wisp) → Nothing (deleted, no trace)
- `distill`: Liquid (ad-hoc epic) → Solid (proto)

## When to Use Molecules

### Use Protos/Mols When:
- **Repeatable patterns** - Same workflow structure used multiple times (releases, reviews, onboarding)
- **Team knowledge capture** - Encoding tribal knowledge as executable templates
- **Audit trail matters** - Work that needs to be tracked and reviewed later
- **Cross-session persistence** - Work spanning multiple days/sessions

### Use Wisps When:
- **Operational loops** - Patrol cycles, health checks, routine monitoring
- **One-shot orchestration** - Temporary coordination that shouldn't clutter history
- **Diagnostic runs** - Debugging workflows with no archival value
- **High-frequency ephemeral work** - Would create noise in permanent database

**Key insight:** Wisps prevent database bloat from routine operations while still providing structure during execution.

---

## Proto Management

### Creating a Proto

Protos are epics with the `template` label. Create manually or distill from existing work:

```bash
# Manual creation
bd create "Release Workflow" --type epic --label template
bd create "Run tests for {{component}}" --type task
bd dep add task-id epic-id --type parent-child

# Distill from ad-hoc work (extracts template from existing epic)
bd mol distill bd-abc123 --as "Release Workflow" --var version=1.0.0
```

**Proto naming convention:** Use `mol-` prefix for clarity (e.g., `mol-release`, `mol-patrol`).

> **CRITICAL**: When protos spawn children, children MUST use the database's prefix
> (e.g., `ap-xxxx`), NOT a custom prefix like `release-xxxx` or hierarchical IDs
> like `mol-release.step1`. Hierarchical IDs corrupt the database.

### Listing Formulas

```bash
bd formula list                 # List all formulas (protos)
bd formula list --json          # Machine-readable
```

### Viewing Proto Structure

```bash
bd mol show mol-release         # Show template structure and variables
bd mol show mol-release --json  # Machine-readable
```

---

## Spawning Molecules

### Basic Spawn (Creates Wisp by Default)

```bash
bd mol spawn mol-patrol                    # Creates wisp (ephemeral)
bd mol spawn mol-feature --pour            # Creates mol (persistent)
bd mol spawn mol-release --var version=2.0 # With variable substitution
```

**Chemistry shortcuts:**
```bash
bd mol pour mol-feature                    # Shortcut for spawn --pour
bd mol wisp mol-patrol                     # Explicit wisp creation
```

### Spawn with Immediate Execution

```bash
bd mol run mol-release --var version=2.0
```

`bd mol run` does three things:
1. Spawns the molecule (persistent)
2. Assigns root issue to caller
3. Pins root issue for session recovery

**Use `mol run` when:** Starting durable work that should survive crashes. The pin ensures `bd ready` shows the work after restart.

### Spawn with Attachments

Attach additional protos in a single command:

```bash
bd mol spawn mol-feature --attach mol-testing --var name=auth
# Spawns mol-feature, then spawns mol-testing and bonds them
```

**Attach types:**
- `sequential` (default) - Attached runs after primary completes
- `parallel` - Attached runs alongside primary
- `conditional` - Attached runs only if primary fails

```bash
bd mol spawn mol-deploy --attach mol-rollback --attach-type conditional
```

---

## Bonding Molecules

### Bond Types

```bash
bd mol bond A B                    # Sequential: B runs after A
bd mol bond A B --type parallel    # Parallel: B runs alongside A
bd mol bond A B --type conditional # Conditional: B runs if A fails
```

### Operand Combinations

| A | B | Result |
|---|---|--------|
| proto | proto | Compound proto (reusable template) |
| proto | mol | Spawn proto, attach to molecule |
| mol | proto | Spawn proto, attach to molecule |
| mol | mol | Join into compound molecule |

### Phase Control in Bonds

By default, spawned protos inherit target's phase. Override with flags:

```bash
# Found bug during wisp patrol? Persist it:
bd mol bond mol-critical-bug wisp-patrol --pour

# Need ephemeral diagnostic on persistent feature?
bd mol bond mol-temp-check bd-feature --wisp
```

### Custom Compound Names

```bash
bd mol bond mol-feature mol-deploy --as "Feature with Deploy"
```

---

## Wisp Lifecycle

### Creating Wisps

```bash
bd mol wisp mol-patrol                       # From proto
bd mol spawn mol-patrol                      # Same (spawn defaults to wisp)
bd mol spawn mol-check --var target=db       # With variables
```

### Listing Wisps

```bash
bd mol wisp list                     # List all wisps
bd mol wisp list --json              # Machine-readable
```

### Ending Wisps

**Option 1: Squash (compress to digest)**
```bash
bd mol squash wisp-abc123                              # Auto-generate summary
bd mol squash wisp-abc123 --summary "Completed patrol" # Agent-provided summary
bd mol squash wisp-abc123 --keep-children              # Keep children, just create digest
bd mol squash wisp-abc123 --dry-run                    # Preview
```

Squash creates a permanent digest issue summarizing the wisp's work, then deletes the wisp children.

**Option 2: Burn (delete without trace)**
```bash
bd mol burn wisp-abc123                    # Delete wisp, no digest
```

Use burn for routine work with no archival value.

### Garbage Collection

```bash
bd mol wisp gc                       # Clean up orphaned wisps
```

---

## Distilling Protos

Extract a reusable template from ad-hoc work:

```bash
bd mol distill bd-o5xe --as "Release Workflow"
bd mol distill bd-abc --var feature_name=auth-refactor --var version=1.0.0
```

**What distill does:**
1. Loads existing epic and all children
2. Clones structure as new proto (adds `template` label)
3. Replaces concrete values with `{{variable}}` placeholders

**Variable syntax (both work):**
```bash
--var branch=feature-auth      # variable=value (recommended)
--var feature-auth=branch      # value=variable (auto-detected)
```

**Use cases:**
- Team develops good workflow organically, wants to reuse it
- Capture tribal knowledge as executable templates
- Create starting point for similar future work

---

## Cross-Project Dependencies

### Concept

Projects can depend on capabilities shipped by other projects:

```bash
# Project A ships a capability
bd ship auth-api                # Marks capability as available

# Project B depends on it
bd dep add bd-123 external:project-a:auth-api
```

### Shipping Capabilities

```bash
bd ship <capability>            # Ship capability (requires closed issue)
bd ship <capability> --force    # Ship even if issue not closed
bd ship <capability> --dry-run  # Preview
```

**How it works:**
1. Find issue with `export:<capability>` label
2. Validate issue is closed
3. Add `provides:<capability>` label

### Depending on External Capabilities

```bash
bd dep add <issue> external:<project>:<capability>
```

The dependency is satisfied when the external project has a closed issue with `provides:<capability>` label.

**`bd ready` respects external deps:** Issues blocked by unsatisfied external dependencies won't appear in ready list.

---

## Common Patterns

### Pattern: Weekly Review Proto

```bash
# Create proto
bd create "Weekly Review" --type epic --label template
bd create "Review open issues" --type task
bd create "Update priorities" --type task
bd create "Archive stale work" --type task
# Link as children...

# Use each week
bd mol spawn mol-weekly-review --pour
```

### Pattern: Ephemeral Patrol Cycle

```bash
# Patrol proto exists
bd mol wisp mol-patrol

# Execute patrol work...

# End patrol
bd mol squash wisp-abc123 --summary "Patrol complete: 3 issues found, 2 resolved"
```

### Pattern: Feature with Rollback

```bash
bd mol spawn mol-deploy --attach mol-rollback --attach-type conditional
# If deploy fails, rollback automatically becomes unblocked
```

### Pattern: Capture Tribal Knowledge

```bash
# After completing a good workflow organically
bd mol distill bd-release-epic --as "Release Process" --var version=X.Y.Z
# Now team can: bd mol spawn mol-release-process --var version=2.0.0
```

---

## CLI Quick Reference

| Command | Purpose |
|---------|---------|
| `bd formula list` | List available formulas/protos |
| `bd mol show <id>` | Show proto/mol structure |
| `bd mol spawn <proto>` | Create wisp from proto (default) |
| `bd mol spawn <proto> --pour` | Create persistent mol from proto |
| `bd mol run <proto>` | Spawn + assign + pin (durable execution) |
| `bd mol bond <A> <B>` | Combine protos or molecules |
| `bd mol distill <epic>` | Extract proto from ad-hoc work |
| `bd mol squash <mol>` | Compress wisp children to digest |
| `bd mol burn <wisp>` | Delete wisp without trace |
| `bd mol pour <proto>` | Shortcut for `spawn --pour` |
| `bd mol wisp <proto>` | Create ephemeral wisp |
| `bd mol wisp list` | List all wisps |
| `bd mol wisp gc` | Garbage collect orphaned wisps |
| `bd ship <capability>` | Publish capability for cross-project deps |

---

## Troubleshooting

**"Proto not found"**
- Check `bd formula list` for available formulas/protos
- Protos need `template` label on the epic

**"Variable not substituted"**
- Use `--var key=value` syntax
- Check proto for `{{key}}` placeholders with `bd mol show`

**"Wisp commands fail"**
- Wisps stored in `.beads-wisp/` (separate from `.beads/`)
- Check `bd mol wisp list` for active wisps

**"External dependency not satisfied"**
- Target project must have closed issue with `provides:<capability>` label
- Use `bd ship <capability>` in target project first

### PATTERNS.md

# Common Usage Patterns

Practical patterns for using bd effectively across different scenarios.

## Contents

- [Knowledge Work Session](#knowledge-work-session) - Resume long-running research or writing tasks
- [Side Quest Handling](#side-quest-handling) - Capture discovered work without losing context
- [Multi-Session Project Resume](#multi-session-project-resume) - Pick up work after time away
- [Status Transitions](#status-transitions) - When to change issue status
- [Compaction Recovery](#compaction-recovery) - Resume after conversation history is lost
- [Issue Closure](#issue-closure) - Documenting completion properly

---

## Knowledge Work Session

**Scenario**: User asks "Help me write a proposal for expanding the analytics platform"

**What you see**:
```bash
$ bd ready
# Returns: bd-42 "Research analytics platform expansion proposal" (in_progress)

$ bd show bd-42
Notes: "COMPLETED: Reviewed current stack (Mixpanel, Amplitude)
IN PROGRESS: Drafting cost-benefit analysis section
NEXT: Need user input on budget constraints before finalizing recommendations"
```

**What you do**:
1. Read notes to understand current state
2. Create Task tools for immediate work:
   ```
   - [ ] Draft cost-benefit analysis
   - [ ] Ask user about budget constraints
   - [ ] Finalize recommendations
   ```
3. Work on tasks, mark Task tools items completed
4. At milestone, update bd notes:
   ```bash
   bd update bd-42 --notes "COMPLETED: Cost-benefit analysis drafted.
   KEY DECISION: User confirmed $50k budget cap - ruled out enterprise options.
   IN PROGRESS: Finalizing recommendations (Posthog + custom ETL).
   NEXT: Get user review of draft before closing issue."
   ```

**Outcome**: Task tools disappears at session end, but bd notes preserve context for next session.

**Key insight**: Notes field captures the "why" and context, Task tools tracks the "doing" right now.

---

## Side Quest Handling

**Scenario**: During main task, discover a problem that needs attention.

**Pattern**:
1. Create issue immediately: `bd create "Found: inventory system needs refactoring"`
2. Link provenance: `bd dep add main-task new-issue --type discovered-from`
3. Assess urgency: blocker or can defer?
4. **If blocker**:
   - `bd update main-task --status blocked`
   - `bd update new-issue --status in_progress`
   - Work on the blocker
5. **If deferrable**:
   - Note in new issue's design field
   - Continue main task
   - New issue persists for later

**Why this works**: Captures context immediately (before forgetting), preserves relationship to main work, allows flexible prioritization.

**Example (with MCP):**

Working on "Implement checkout flow" (checkout-1), discover payment validation security hole:

1. Create bug issue: `mcp__plugin_beads_beads__create` with `{title: "Fix: payment validation bypasses card expiry check", type: "bug", priority: 0}`
2. Link discovery: `mcp__plugin_beads_beads__dep` with `{from_issue: "checkout-1", to_issue: "payment-bug-2", type: "discovered-from"}`
3. Block current work: `mcp__plugin_beads_beads__update` with `{issue_id: "checkout-1", status: "blocked", notes: "Blocked by payment-bug-2: security hole in validation"}`
4. Start new work: `mcp__plugin_beads_beads__update` with `{issue_id: "payment-bug-2", status: "in_progress"}`

(CLI: `bd create "Fix: payment validation..." -t bug -p 0` then `bd dep add` and `bd update` commands)

---

## Multi-Session Project Resume

**Scenario**: Starting work after days or weeks away from a project.

**Pattern (with MCP)**:
1. **Check what's ready**: Use `mcp__plugin_beads_beads__ready` to see available work
2. **Check what's stuck**: Use `mcp__plugin_beads_beads__blocked` to understand blockers
3. **Check recent progress**: Use `mcp__plugin_beads_beads__list` with `status:"closed"` to see completions
4. **Read detailed context**: Use `mcp__plugin_beads_beads__show` for the issue you'll work on
5. **Update status**: Use `mcp__plugin_beads_beads__update` with `status:"in_progress"`
6. **Begin work**: Create Task tools from notes field's NEXT section

(CLI: `bd ready`, `bd blocked`, `bd list --status closed`, `bd show <id>`, `bd update <id> --status in_progress`)

**Example**:
```bash
$ bd ready
Ready to work on (3):
  auth-5: "Add OAuth refresh token rotation" (priority: 0)
  api-12: "Document REST API endpoints" (priority: 1)
  test-8: "Add integration tests for payment flow" (priority: 2)

$ bd show auth-5
Title: Add OAuth refresh token rotation
Status: open
Priority: 0 (critical)

Notes:
COMPLETED: Basic JWT auth working
IN PROGRESS: Need to add token refresh
NEXT: Implement rotation per OWASP guidelines (7-day refresh tokens)
BLOCKER: None - ready to proceed

$ bd update auth-5 --status in_progress
# Now create Task tools based on NEXT section
```

**For complete session start workflow with checklist, see:** [WORKFLOWS.md](WORKFLOWS.md#session-start)

---

## Status Transitions

Understanding when to change issue status.

### Status Lifecycle

```
open → in_progress → closed
  ↓         ↓
blocked   blocked
```

### When to Use Each Status

**open** (default):
- Issue created but not started
- Waiting for dependencies to clear
- Planned work not yet begun
- **Command**: Issues start as `open` by default

**in_progress**:
- Actively working on this issue right now
- Has been read and understood
- Making commits or changes related to this
- **Command**: `bd update issue-id --status in_progress`
- **When**: Start of work session on this issue

**blocked**:
- Cannot proceed due to external blocker
- Waiting for user input/decision
- Dependency not completed
- Technical blocker discovered
- **Command**: `bd update issue-id --status blocked`
- **When**: Hit a blocker, capture what blocks you in notes
- **Note**: Document blocker in notes field: "BLOCKER: Waiting for API key from ops team"

**closed**:
- Work completed and verified
- Tests passing
- Acceptance criteria met
- **Command**: `bd close issue-id --reason "Implemented with tests passing"`
- **When**: All work done, ready to move on
- **Note**: Issues remain in database, just marked complete

### Transition Examples

**Starting work**:
```bash
bd ready  # See what's available
bd update auth-5 --status in_progress
# Begin working
```

**Hit a blocker**:
```bash
bd update auth-5 --status blocked --notes "BLOCKER: Need OAuth client ID from product team. Emailed Jane on 2025-10-23."
# Switch to different issue or create new work
```

**Unblocking**:
```bash
# Once blocker resolved
bd update auth-5 --status in_progress --notes "UNBLOCKED: Received OAuth credentials. Resuming implementation."
```

**Completing**:
```bash
bd close auth-5 --reason "Implemented OAuth refresh with 7-day rotation. Tests passing. PR #42 merged."
```

---

## Compaction Recovery

**Scenario**: Conversation history has been compacted. You need to resume work with zero conversation context.

**What survives compaction**:
- All bd issues and notes
- Complete work history
- Dependencies and relationships

**What's lost**:
- Conversation history
- Task tools lists
- Recent discussion

### Recovery Pattern

1. **Check in-progress work**:
   ```bash
   bd list --status in_progress
   ```

2. **Read notes for context**:
   ```bash
   bd show issue-id
   # Read notes field - should explain current state
   ```

3. **Reconstruct Task tools from notes**:
   - COMPLETED section: Done, skip
   - IN PROGRESS section: Current state
   - NEXT section: **This becomes your Task tools list**

4. **Report to user**:
   ```
   "From bd notes: [summary of COMPLETED]. Currently [IN PROGRESS].
   Next steps: [from NEXT]. Should I continue with that?"
   ```

### Example Recovery

**bd show returns**:
```
Issue: bd-42 "OAuth refresh token implementation"
Status: in_progress
Notes:
COMPLETED: Basic JWT validation working (RS256, 1hr access tokens)
KEY DECISION: 7-day refresh tokens per security review
IN PROGRESS: Implementing token rotation endpoint
NEXT: Add rate limiting (5 refresh attempts per 15min), then write tests
BLOCKER: None
```

**Recovery actions**:
1. Read notes, understand context
2. Create Task tools:
   ```
   - [ ] Implement rate limiting on refresh endpoint
   - [ ] Write tests for token rotation
   - [ ] Verify security guidelines met
   ```
3. Report: "From notes: JWT validation is done with 7-day refresh tokens. Currently implementing rotation endpoint. Next: add rate limiting and tests. Should I continue?"
4. Resume work based on user response

**For complete compaction survival workflow, see:** [WORKFLOWS.md](WORKFLOWS.md#compaction-survival)

---

## Issue Closure

**Scenario**: Work is complete. How to close properly?

### Closure Checklist

Before closing, verify:
- [ ] **Acceptance criteria met**: All items checked off
- [ ] **Tests passing**: If applicable
- [ ] **Documentation updated**: If needed
- [ ] **Follow-up work filed**: New issues created for discovered work
- [ ] **Key decisions documented**: In notes field

### Closure Pattern

**Minimal closure** (simple tasks):
```bash
bd close task-123 --reason "Implemented feature X"
```

**Detailed closure** (complex work):
```bash
# Update notes with final state
bd update task-123 --notes "COMPLETED: OAuth refresh with 7-day rotation
KEY DECISION: RS256 over HS256 per security review
TESTS: 12 tests passing (auth, rotation, expiry, errors)
FOLLOW-UP: Filed perf-99 for token cleanup job"

# Close with summary
bd close task-123 --reason "Implemented OAuth refresh token rotation with rate limiting. All security guidelines met. Tests passing."
```

### Documenting Resolution (Outcome vs Design)

For issues where the outcome differed from initial design, use `--notes` to document what actually happened:

```bash
# Initial design was hypothesis - document actual outcome in notes
bd update bug-456 --notes "RESOLUTION: Not a bug - behavior is correct per OAuth spec. Documentation was unclear. Filed docs-789 to clarify auth flow in user guide."

bd close bug-456 --reason "Resolved: documentation issue, not bug"
```

**Pattern**: Design field = initial approach. Notes field = what actually happened (prefix with RESOLUTION: for clarity).

### Discovering Follow-up Work

When closing reveals new work:

```bash
# While closing auth feature, realize performance needs work
bd create "Optimize token lookup query" -t task -p 2

# Link the provenance
bd dep add auth-5 perf-99 --type discovered-from

# Now close original
bd close auth-5 --reason "OAuth refresh implemented. Discovered perf optimization needed (filed perf-99)."
```

**Why link with discovered-from**: Preserves the context of how you found the new work. Future you will appreciate knowing it came from the auth implementation.

---

## Pattern Summary

| Pattern | When to Use | Key Command | Preserves |
|---------|-------------|-------------|-----------|
| **Knowledge Work** | Long-running research, writing | `bd update --notes` | Context across sessions |
| **Side Quest** | Discovered during other work | `bd dep add --type discovered-from` | Relationship to original |
| **Multi-Session Resume** | Returning after time away | `bd ready`, `bd show` | Full project state |
| **Status Transitions** | Tracking work state | `bd update --status` | Current state |
| **Compaction Recovery** | History lost | Read notes field | All context in notes |
| **Issue Closure** | Completing work | `bd close --reason` | Decisions and outcomes |

**For detailed workflows with step-by-step checklists, see:** [WORKFLOWS.md](WORKFLOWS.md)

### RESUMABILITY.md

# Making Issues Resumable Across Sessions

## When Resumability Matters

**Use enhanced documentation for:**
- Multi-session technical features with API integration
- Complex algorithms requiring code examples
- Features with specific output format requirements
- Work with "occult" APIs (undocumented capabilities)

**Skip for:**
- Simple bug fixes with clear scope
- Well-understood patterns (CRUD operations, etc.)
- Single-session tasks
- Work with obvious acceptance criteria

**The test:** Would a fresh Claude instance (or you after 2 weeks) struggle to resume this work from the description alone? If yes, add implementation details.

## Anatomy of a Resumable Issue

### Minimal (Always Include)
```markdown
Description: What needs to be built and why
Acceptance Criteria: Concrete, testable outcomes (WHAT not HOW)
```

### Enhanced (Complex Technical Work)
```markdown
Notes Field - IMPLEMENTATION GUIDE:

WORKING CODE:
```python
# Tested code that queries the API
service = build('drive', 'v3', credentials=creds)
result = service.about().get(fields='importFormats').execute()
# Returns: {'text/markdown': ['application/vnd.google-apps.document'], ...}
```

API RESPONSE SAMPLE:
Shows actual data structure (not docs description)

DESIRED OUTPUT FORMAT:
```markdown
# Example of what the output should look like
Not just "return markdown" but actual structure
```

RESEARCH CONTEXT:
Why this approach? What alternatives were considered?
Key discoveries that informed the design.
```

## Real Example: Before vs After

### ❌ Not Resumable
```
Title: Add dynamic capabilities resources
Description: Query Google APIs for capabilities and return as resources
Acceptance: Resources return capability info
```

**Problem:** Future Claude doesn't know:
- Which API endpoints to call
- What the responses look like
- What format to return

### ✅ Resumable
```
Title: Add dynamic capabilities resources
Description: Query Google APIs for system capabilities (import formats,
themes, quotas) that aren't in static docs. Makes server self-documenting.

Notes: IMPLEMENTATION GUIDE

WORKING CODE (tested):
```python
from workspace_mcp.tools.drive import get_credentials
from googleapiclient.discovery import build

creds = get_credentials()
service = build('drive', 'v3', credentials=creds)
about = service.about().get(
    fields='importFormats,exportFormats,folderColorPalette'
).execute()

# Returns:
# - importFormats: dict, 49 entries like {'text/markdown': [...]}
# - exportFormats: dict, 10 entries
# - folderColorPalette: list, 24 hex strings
```

OUTPUT FORMAT EXAMPLE:
```markdown
# Drive Import Formats

Google Drive supports 49 import formats:

## Text Formats
- **text/markdown** → Google Docs ✨ (NEW July 2024)
- text/plain → Google Docs
...
```

RESEARCH CONTEXT:
text/markdown support announced July 2024 but NOT in static Google docs.
Google's workspace-developer MCP server doesn't expose this.
This is why dynamic resources matter.

Acceptance Criteria:
- User queries workspace://capabilities/drive/import-formats
- Response shows all 49 formats including text/markdown
- Output is readable markdown, not raw JSON
- Queries live API (not static data)
```

**Result:** Fresh Claude instance can:
1. See working API query code
2. Understand response structure
3. Know desired output format
4. Implement with context

## Optional Template

Copy this into notes field for complex technical features:

```markdown
IMPLEMENTATION GUIDE FOR FUTURE SESSIONS:

WORKING CODE (tested):
```language
# Paste actual code that works
# Include imports and setup
# Show what it returns
```

API RESPONSE SAMPLE:
```json
{
  "actualField": "actualValue",
  "structure": "as returned by API"
}
```

DESIRED OUTPUT FORMAT:
```
Show what the final output should look like
Not just "markdown" but actual structure/style
```

RESEARCH CONTEXT:
- Why this approach?
- What alternatives considered?
- Key discoveries?
- Links to relevant docs/examples?
```

## Anti-Patterns

### ❌ Over-Documenting Simple Work
```markdown
Title: Fix typo in README
Notes: IMPLEMENTATION GUIDE
WORKING CODE: Open README.md, change "teh" to "the"...
```
**Problem:** Wastes tokens on obvious work.

### ❌ Design Details in Acceptance Criteria
```markdown
Acceptance:
- [ ] Use batchUpdate approach
- [ ] Call API with fields parameter
- [ ] Format as markdown with ## headers
```
**Problem:** Locks implementation. Should be in Design/Notes, not Acceptance.

### ❌ Raw JSON Dumps
```markdown
API RESPONSE:
{giant unformatted JSON blob spanning 100 lines}
```
**Problem:** Hard to read. Extract relevant parts, show structure.

### ✅ Right Balance
```markdown
API RESPONSE SAMPLE:
Returns dict with 49 entries. Example entries:
- 'text/markdown': ['application/vnd.google-apps.document']
- 'text/plain': ['application/vnd.google-apps.document']
- 'application/pdf': ['application/vnd.google-apps.document']
```

## When to Add This Detail

**During issue creation:**
- Already have working code from research? Include it.
- Clear output format in mind? Show example.

**During work (update notes):**
- Just got API query working? Add to notes.
- Discovered important context? Document it.
- Made key decision? Explain rationale.

**Session end:**
- If resuming will be hard, add implementation guide.
- If obvious, skip it.

**The principle:** Help your future self (or next Claude) resume without rediscovering everything.

### ROUTING.md

# Beads Routing Architecture

**For:** AI agents working in multi-workspace environments (Gas Town)
**Applies to:** Two-level beads deployments with Town + Rig structure

## Overview

In multi-agent environments, beads operates at two levels with automatic prefix-based routing.

## Two-Level Architecture

| Level | Location | sync-branch | Prefix | Purpose |
|-------|----------|-------------|--------|---------|
| Town | `~/gt/.beads/` | NOT set | `hq-*` | Mail, HQ coordination |
| Rig | `<rig>/crew/*/.beads/` | `beads-sync` | Olympian prefix | Project issues |

**Key points:**
- **Town beads**: Mail and coordination. Commits to main (single clone, no sync needed)
- **Rig beads**: Project work in git worktrees (crew/*, polecats/*)
- The rig-level `<rig>/.beads/` is **gitignored** (local runtime state)
- Rig beads use `beads-sync` branch for multi-clone coordination

## Prefix-Based Routing

`bd` commands automatically route to the correct rig based on issue ID prefix:

```bash
bd show gt-xyz   # Routes to daedalus beads
bd show ap-abc   # Routes to athena beads
bd show hq-123   # Routes to town beads
```

**How it works:**
- Routes defined in `~/gt/.beads/routes.jsonl`
- `gt rig add` auto-registers new rig prefixes
- Each rig's prefix (e.g., `gt-`) maps to its beads location

**Debug routing:**
```bash
BD_DEBUG_ROUTING=1 bd show <id>
```

**Conflicts:** If two rigs share a prefix, use `bd rename-prefix <new>` to fix.

### Common Prefixes

| Prefix | Rig | Prefix | Rig |
|--------|-----|--------|-----|
| `hq` | town (coordination) | `gt` | daedalus |
| `ap` | athena | `ho` | argus |
| `be` | chronicle | `gitops` | gitops |
| `starport` | starport | `fr` | cyclopes |

## Creating Beads for Rig Work

**HQ beads (`hq-*`) CANNOT be hooked by polecats!**

`gt sling` uses `bd update` which lacks cross-database routing. Beads must exist
in the target rig's database to be hookable.

| Work Type | Create From | Gets Prefix | Can Sling? |
|-----------|-------------|-------------|------------|
| Mayor coordination | `~/gt` | `hq-*` | No |
| Rig bug/feature | Rig's beads | `gt-*`, `ap-*`, etc. | Yes |

**From Mayor, to create a slingable bead:**

```bash
# Use BEADS_DIR to target the rig's beads database
BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create --title="Fix X" --type=bug
# Creates: gt-xxxxx (daedalus prefix)

BEADS_DIR=~/gt/athena/mayor/rig/.beads bd create --title="Add Y" --type=feature
# Creates: ap-xxxxx (athena prefix)
```

**Then sling normally:**
```bash
gt sling gt-xxxxx daedalus   # Works - bead is in daedalus's database
gt sling ap-xxxxx athena     # Works - bead is in athena's database
```

## Common Gotchas

### Wrong Prefix for Rig Work

Creating from `~/gt` gives `hq-*` which polecats can't hook.

- **WRONG:** `bd create --title="daedalus bug"` from town root
  - Creates `hq-xxx` (unhookable by polecats)
- **RIGHT:** `BEADS_DIR=~/gt/daedalus/mayor/rig/.beads bd create ...`
  - Creates `gt-xxx` (slingable)

### GitHub URLs

Use `git remote -v` to verify repo URLs - never assume orgs like `anthropics/`.

### Temporal Language Inverts Dependencies

"Phase 1 blocks Phase 2" is backwards in dependency semantics:

```bash
# WRONG (temporal thinking: "1 before 2")
bd dep add phase1 phase2

# RIGHT (requirement thinking: "2 needs 1")
bd dep add phase2 phase1
```

**Rule:** Think "X needs Y", not "X comes before Y". Verify with `bd blocked`.

## Sync Behavior by Level

### Town Level
- Single clone, no sync branch needed
- Commits directly to main
- Used for: mail, HQ coordination beads

### Rig Level
- Multiple worktrees (crew/*, polecats/*)
- Uses `beads-sync` branch for coordination
- `bd sync` handles cross-worktree synchronization
- `bd sync --from-main` pulls updates from main (for ephemeral branches)

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Bead not found | Check prefix matches rig: `BD_DEBUG_ROUTING=1 bd show <id>` |
| Can't sling HQ bead | Create bead in rig's database with `BEADS_DIR` |
| Prefix conflict | Run `bd rename-prefix <new>` on one rig |
| Sync fails | Ensure `beads-sync` branch exists: `git branch beads-sync` |

### STATIC_DATA.md

# Using bd for Static Reference Data

bd is primarily designed for work tracking, but can also serve as a queryable database for static reference data with some adaptations.

## Work Tracking (Primary Use Case)

Standard bd workflow:
- Issues flow through states (open → in_progress → closed)
- Priorities and dependencies matter
- Status tracking is essential
- IDs are sufficient for referencing

## Reference Databases / Glossaries (Alternative Use)

When using bd for static data (terminology, glossaries, reference information):

**Characteristics:**
- Entities are mostly static (typically always open)
- No real workflow or state transitions
- Names/titles more important than IDs
- Minimal or no dependencies

**Recommended approach:**
- Use separate database (not mixed with work tracking) to avoid confusion
- Consider dual format: maintain markdown version alongside database for name-based lookup
- Example: A terminology database could use both `terms.db` (queryable via bd) and `GLOSSARY.md` (browsable by name)

**Key difference**: Work items have lifecycle; reference entities are stable knowledge.

## When to Use This Pattern

**Good fit:**
- Technical glossaries or terminology databases
- Reference documentation that needs dependency tracking
- Knowledge bases with relationships between entries
- Structured data that benefits from queryability

**Poor fit:**
- Data that changes frequently (use work tracking pattern)
- Simple lists (markdown is simpler)
- Data that needs complex queries (use proper database)

## Limitations

**bd show requires IDs, not names:**
- `bd show term-42` works
- `bd show "API endpoint"` doesn't work
- Workaround: `bd list | grep -i "api endpoint"` to find ID first
- This is why dual format (bd + markdown) is recommended for reference data

**No search by content:**
- bd searches by ID, title filters, status, labels
- For full-text search across descriptions/notes, use grep on the JSONL file
- Example: `grep -i "authentication" .beads/issues.jsonl`

### TROUBLESHOOTING.md

# Troubleshooting Guide

Common issues encountered when using bd and how to resolve them.

**See also**: [ANTI_PATTERNS.md](ANTI_PATTERNS.md) for preventable mistakes.

## Interface-Specific Troubleshooting

**MCP tools (local environment):**
- MCP tools require bd daemon running
- Check daemon status: `bd daemon --status` (CLI)
- If MCP tools fail, verify daemon is running and restart if needed
- MCP tools automatically use daemon mode (no --no-daemon option)

**CLI (web environment or local):**
- CLI can use daemon mode (default) or direct mode (--no-daemon)
- Direct mode has 3-5 second sync delay
- Web environment: Install via `npm install -g @beads/cli`
- Web environment: Initialize via `bd init <prefix>` before first use

**Most issues below apply to both interfaces** - the underlying database and daemon behavior is the same.

## Contents

- [Database Out of Sync / Prefix Mismatch](#database-out-of-sync--prefix-mismatch)
- [Molecule-Style ID Corruption](#molecule-style-id-corruption)
- [Dependencies Not Persisting](#dependencies-not-persisting)
- [Status Updates Not Visible](#status-updates-not-visible)
- [Daemon Won't Start](#daemon-wont-start)
- [Database Errors on Cloud Storage](#database-errors-on-cloud-storage)
- [JSONL File Not Created](#jsonl-file-not-created)
- [Version Requirements](#version-requirements)

---

## Database Out of Sync / Prefix Mismatch

### Symptom
```bash
bd list
# Error: Database out of sync with JSONL. Run 'bd sync --import-only' to fix.

bd sync --import-only
# Error: prefix mismatch detected: database uses 'ap-' but found issues with prefixes:
# [code- (14 issues) etl- (8 issues) hybrid- (11 issues)]
```

### Root Cause
Multiple prefixes accumulated in the JSONL file, but the database is configured for a single prefix. This typically happens when:
1. Formula/molecule templates created issues with their own prefixes
2. Issues were manually added with wrong prefixes
3. JSONL files from different projects were merged

### Resolution

**Option 1: Filter to single prefix (preserves valid data)**
```bash
cd .beads
cp issues.jsonl issues.jsonl.bak

# Keep only issues with correct prefix (e.g., ap-)
grep -E '"id":"ap-' issues.jsonl.bak > issues.jsonl

# Reimport
bd sync --import-only
```

**Option 2: Use --rename-on-import (if IDs are standard format)**
```bash
bd sync --import-only --rename-on-import
# Renames all issues to database prefix
```

**Option 3: Nuclear rebuild**
```bash
rm -rf .beads/*.db
bd sync --import-only
```

### Prevention
- **Single prefix per database**: Never create issues with different prefixes
- **Formula discipline**: Ensure formulas use parent epic's prefix
- **Regular audits**: `grep -o '"id":"[^-]*' .beads/issues.jsonl | sort -u`

---

## Molecule-Style ID Corruption

### Symptom
```bash
bd sync --import-only --rename-on-import
# Error: cannot rename issue code-map-validation: invalid suffix 'map-validation'
```

### Root Cause
Issues were created with dot-hierarchical IDs like:
- `code-map-validation.calculate-coverage`
- `etl-throughput-optimization.enable-parallel-sync`

These don't match the standard `prefix-hash` format and can't be renamed.

### Resolution

**Filter out malformed IDs**:
```bash
cd .beads
cp issues.jsonl issues.jsonl.bak

# Keep only standard format IDs (prefix-alphanumeric, no dots)
grep -E '"id":"[a-z]+-[a-z0-9]+"' issues.jsonl.bak > issues.jsonl

# Check what was removed
diff <(wc -l < issues.jsonl.bak) <(wc -l < issues.jsonl)

# Reimport
rm -f *.db
bd sync --import-only
```

**Full reset** (if too corrupted):
```bash
rm -rf .beads/
bd init --prefix=ap
```

### Prevention
- **Never use dots in IDs**: IDs must be `prefix-hash` format
- **Audit formula outputs**: Check that spawned issues have correct format
- **Reject molecule templates that create custom prefixes**

---

## Dependencies Not Persisting

### Symptom
```bash
bd dep add issue-2 issue-1 --type blocks
# Reports: ✓ Added dependency
bd show issue-2
# Shows: No dependencies listed
```

### Root Cause (Fixed in v0.15.0+)
This was a **bug in bd** (GitHub issue #101) where the daemon ignored dependencies during issue creation. **Fixed in bd v0.15.0** (Oct 21, 2025).

### Resolution

**1. Check your bd version:**
```bash
bd version
```

**2. If version < 0.15.0, update bd:**
```bash
# Via Homebrew (macOS/Linux)
brew upgrade bd

# Via go install
go install github.com/steveyegge/beads/cmd/bd@latest

# Via package manager
# See https://github.com/steveyegge/beads#installing
```

**3. Restart daemon after upgrade:**
```bash
pkill -f "bd daemon"  # Kill old daemon
bd daemon             # Start new daemon with fix
```

**4. Test dependency creation:**
```bash
bd create "Test A" -t task
bd create "Test B" -t task
bd dep add <B-id> <A-id> --type blocks
bd show <B-id>
# Should show: "Depends on (1): → <A-id>"
```

### Still Not Working?

If dependencies still don't persist after updating:

1. **Check daemon is running:**
   ```bash
   ps aux | grep "bd daemon"
   ```

2. **Try without --no-daemon flag:**
   ```bash
   # Instead of: bd --no-daemon dep add ...
   # Use: bd dep add ...  (let daemon handle it)
   ```

3. **Check JSONL file:**
   ```bash
   cat .beads/issues.jsonl | jq '.dependencies'
   # Should show dependency array
   ```

4. **Report to beads GitHub** with:
   - `bd version` output
   - Operating system
   - Reproducible test case

---

## Status Updates Not Visible

### Symptom
```bash
bd --no-daemon update issue-1 --status in_progress
# Reports: ✓ Updated issue: issue-1
bd show issue-1
# Shows: Status: open (not in_progress!)
```

### Root Cause
This is **expected behavior**, not a bug. Understanding requires knowing bd's architecture:

**BD Architecture:**
- **JSONL files** (`.beads/issues.jsonl`): Human-readable export format
- **SQLite database** (`.beads/*.db`): Source of truth for queries
- **Daemon**: Syncs JSONL ↔ SQLite every 5 minutes

**What `--no-daemon` actually does:**
- **Writes**: Go directly to JSONL file
- **Reads**: Still come from SQLite database
- **Sync delay**: Daemon imports JSONL → SQLite periodically

### Resolution

**Option 1: Use daemon mode (recommended)**
```bash
# Don't use --no-daemon for CRUD operations
bd update issue-1 --status in_progress
bd show issue-1
# ✓ Status reflects immediately
```

**Option 2: Wait for sync (if using --no-daemon)**
```bash
bd --no-daemon update issue-1 --status in_progress
# Wait 3-5 seconds for daemon to sync
sleep 5
bd show issue-1
# ✓ Status should reflect now
```

**Option 3: Manual sync trigger**
```bash
bd --no-daemon update issue-1 --status in_progress
# Trigger sync by exporting/importing
bd export > /dev/null 2>&1  # Forces sync
bd show issue-1
```

### When to Use `--no-daemon`

**Use --no-daemon for:**
- Batch import scripts (performance)
- CI/CD environments (no persistent daemon)
- Testing/debugging

**Don't use --no-daemon for:**
- Interactive development
- Real-time status checks
- When you need immediate query results

---

## Daemon Won't Start

### Symptom
```bash
bd daemon
# Error: not in a git repository
# Hint: run 'git init' to initialize a repository
```

### Root Cause
bd daemon requires a **git repository** because it uses git for:
- Syncing issues to git remote (optional)
- Version control of `.beads/*.jsonl` files
- Commit history of issue changes

### Resolution

**Initialize git repository:**
```bash
# In your project directory
git init
bd daemon
# ✓ Daemon should start now
```

**Prevent git remote operations:**
```bash
# If you don't want daemon to pull from remote
bd daemon --global=false
```

**Flags:**
- `--global=false`: Don't sync with git remote
- `--interval=10m`: Custom sync interval (default: 5m)
- `--auto-commit=true`: Auto-commit JSONL changes

---

## Database Errors on Cloud Storage

### Symptom
```bash
# In directory: /Users/name/Google Drive/...
bd init myproject
# Error: disk I/O error (522)
# OR: Error: database is locked
```

### Root Cause
**SQLite incompatibility with cloud sync filesystems.**

Cloud services (Google Drive, Dropbox, OneDrive, iCloud) don't support:
- POSIX file locking (required by SQLite)
- Consistent file handles across sync operations
- Atomic write operations

This is a **known SQLite limitation**, not a bd bug.

### Resolution

**Move bd database to local filesystem:**

```bash
# Wrong location (cloud sync)
~/Google Drive/My Work/project/.beads/  # ✗ Will fail

# Correct location (local disk)
~/Repos/project/.beads/                 # ✓ Works reliably
~/Projects/project/.beads/              # ✓ Works reliably
```

**Migration steps:**

1. **Move project to local disk:**
   ```bash
   mv ~/Google\ Drive/project ~/Repos/project
   cd ~/Repos/project
   ```

2. **Re-initialize bd (if needed):**
   ```bash
   bd init myproject
   ```

3. **Import existing issues (if you had JSONL export):**
   ```bash
   bd import < issues-backup.jsonl
   ```

**Alternative: Use global `~/.beads/` database**

If you must keep work on cloud storage:
```bash
# Don't initialize bd in cloud-synced directory
# Use global database instead
cd ~/Google\ Drive/project
bd create "My task"
# Uses ~/.beads/default.db (on local disk)
```

**Workaround limitations:**
- No per-project database isolation
- All projects share same issue prefix
- Manual tracking of which issues belong to which project

**Recommendation:** Keep code/projects on local disk, sync final deliverables to cloud.

---

## JSONL File Not Created

### Symptom
```bash
bd init myproject
bd --no-daemon create "Test" -t task
ls .beads/
# Only shows: .gitignore, myproject.db
# Missing: issues.jsonl
```

### Root Cause
**JSONL initialization coupling.** The `issues.jsonl` file is created by daemon on first startup, not by `bd init`.

### Resolution

**Start daemon once to initialize JSONL:**
```bash
bd daemon --global=false &
# Wait for initialization
sleep 2

# Now JSONL file exists
ls .beads/issues.jsonl
# ✓ File created

# Subsequent --no-daemon operations work
bd --no-daemon create "Task 1" -t task
cat .beads/issues.jsonl
# ✓ Shows task data
```

**Why this matters:**
- Daemon owns the JSONL export format
- First daemon run creates empty JSONL skeleton
- `--no-daemon` operations assume JSONL exists

**Pattern for batch scripts:**
```bash
#!/bin/bash
# Batch import script

bd init myproject
bd daemon --global=false &  # Start daemon
sleep 3                     # Wait for initialization

# Now safe to use --no-daemon for performance
for item in "${items[@]}"; do
    bd --no-daemon create "$item" -t feature
done

# Daemon syncs JSONL → SQLite in background
sleep 5  # Wait for final sync

# Query results
bd stats
```

---

## Version Requirements

### Minimum Version for Dependency Persistence

**Issue:** Dependencies created but don't appear in `bd show` or dependency tree.

**Fix:** Upgrade to **bd v0.15.0+** (released Oct 2025)

**Check version:**
```bash
bd version
# Should show: bd version 0.15.0 or higher
```

**If using MCP plugin:**
```bash
# Update Codex beads plugin
claude plugin update beads
```

### Breaking Changes

**v0.15.0:**
- MCP parameter names changed from `from_id/to_id` to `issue_id/depends_on_id`
- Dependency creation now persists correctly in daemon mode

**v0.14.0:**
- Daemon architecture changes
- Auto-sync JSONL behavior introduced

---

## MCP-Specific Issues

### Dependencies Created Backwards

**Symptom:**
Using MCP tools, dependencies end up reversed from intended.

**Example:**
```python
# Want: "task-2 depends on task-1" (task-1 blocks task-2)
beads_add_dependency(issue_id="task-1", depends_on_id="task-2")
# Wrong! This makes task-1 depend on task-2
```

**Root Cause:**
Parameter confusion between old (`from_id/to_id`) and new (`issue_id/depends_on_id`) names.

**Resolution:**

**Correct MCP usage (bd v0.15.0+):**
```python
# Correct: task-2 depends on task-1
beads_add_dependency(
    issue_id="task-2",        # Issue that has dependency
    depends_on_id="task-1",   # Issue that must complete first
    dep_type="blocks"
)
```

**Mnemonic:**
- `issue_id`: The issue that **waits**
- `depends_on_id`: The issue that **must finish first**

**Equivalent CLI:**
```bash
bd dep add task-2 task-1 --type blocks
# Meaning: task-2 depends on task-1
```

**Verify dependency direction:**
```bash
bd show task-2
# Should show: "Depends on: task-1"
# Not the other way around
```

---

## Getting Help

### Debug Checklist

Before reporting issues, collect this information:

```bash
# 1. Version
bd version

# 2. Daemon status
ps aux | grep "bd daemon"

# 3. Database location
echo $PWD/.beads/*.db
ls -la .beads/

# 4. Git status
git status
git log --oneline -1

# 5. JSONL contents (for dependency issues)
cat .beads/issues.jsonl | jq '.' | head -50
```

### Report to beads GitHub

If problems persist:

1. **Check existing issues:** https://github.com/steveyegge/beads/issues
2. **Create new issue** with:
   - bd version (`bd version`)
   - Operating system
   - Debug checklist output (above)
   - Minimal reproducible example
   - Expected vs actual behavior

### Codex Skill Issues

If the **bd-issue-tracking skill** provides incorrect guidance:

1. **Check skill version:**
   ```bash
   ls -la beads/
   head -20 beads/SKILL.md
   ```

2. **Report via Codex feedback** or user's GitHub

---

## Quick Reference: Common Fixes

| Problem | Quick Fix |
|---------|-----------|
| Dependencies not saving | Upgrade to bd v0.15.0+ |
| Status updates lag | Use daemon mode (not `--no-daemon`) |
| Daemon won't start | Run `git init` first |
| Database errors on Google Drive | Move to local filesystem |
| JSONL file missing | Start daemon once: `bd daemon &` |
| Dependencies backwards (MCP) | Update to v0.15.0+, use `issue_id/depends_on_id` correctly |

---

## Related Documentation

- [CLI Reference](CLI_REFERENCE.md) - Complete command documentation
- [Dependencies Guide](DEPENDENCIES.md) - Understanding dependency types
- [Workflows](WORKFLOWS.md) - Step-by-step workflow guides
- [beads GitHub](https://github.com/steveyegge/beads) - Official documentation

### WORKFLOWS.md

# Workflows and Checklists

Detailed step-by-step workflows for common bd usage patterns with checklists.

## Contents

- [Session Start Workflow](#session-start) - Check bd ready, establish context
- [Compaction Survival](#compaction-survival) - Recovering after compaction events
- [Discovery and Issue Creation](#discovery) - Proactive issue creation during work
- [Status Maintenance](#status-maintenance) - Keeping bd status current
- [Epic Planning](#epic-planning) - Structuring complex work with dependencies
- [Side Quest Handling](#side-quests) - Discovery during main task, assessing blocker vs deferrable, resuming
- [Multi-Session Resume](#resume) - Returning after days/weeks away
- [Session Handoff Workflow](#session-handoff) - Collaborative handoff between sessions
- [Unblocking Work](#unblocking) - Handling blocked issues
- [Integration with Task tools](#integration-with-todowrite) - Using both tools together
- [Common Workflow Patterns](#common-workflow-patterns)
  - Systematic Exploration, Bug Investigation, Refactoring with Dependencies, Spike Investigation
- [Checklist Templates](#checklist-templates)
  - Starting Any Work Session, Creating Issues During Work, Completing Work, Planning Complex Features
- [Decision Points](#decision-points)
- [Troubleshooting Workflows](#troubleshooting-workflows)

## Session Start Workflow {#session-start}

**bd is available when**:
- Project has `.beads/` directory (project-local), OR
- `~/.beads/` exists (global fallback for any directory)

**Automatic checklist at session start:**

```
Session Start (when bd is available):
- [ ] Run bd ready --json
- [ ] Report: "X items ready to work on: [summary]"
- [ ] If using global ~/.beads, note this in report
- [ ] If none ready, check bd blocked --json
- [ ] Suggest next action based on findings
```

**Pattern**: Always run `bd ready` when starting work where bd is available. Report status immediately to establish shared context.

**Database selection**: bd auto-discovers which database to use (project-local `.beads/` takes precedence over global `~/.beads/`).

---

## Compaction Survival {#compaction-survival}

**Critical**: After compaction events, conversation history is deleted but bd state persists. Beads are your only memory.

**Post-compaction recovery checklist:**

```
After Compaction:
- [ ] Run bd list --status in_progress to see active work
- [ ] Run bd show <issue-id> for each in_progress issue
- [ ] Read notes field to understand: COMPLETED, IN PROGRESS, BLOCKERS, KEY DECISIONS
- [ ] Check dependencies: bd dep tree <issue-id> for context
- [ ] If notes insufficient, check bd list --status open for related issues
- [ ] Reconstruct Task tools list from notes if needed
```

**Pattern**: Well-written notes enable full context recovery even with zero conversation history.

**Writing notes for compaction survival:**

**Good note (enables recovery):**
```
bd update issue-42 --notes "COMPLETED: User authentication - added JWT token
generation with 1hr expiry, implemented refresh token endpoint using rotating
tokens pattern. IN PROGRESS: Password reset flow. Email service integration
working. NEXT: Need to add rate limiting to reset endpoint (currently unlimited
requests). KEY DECISION: Using bcrypt with 12 rounds after reviewing OWASP
recommendations, tech lead concerned about response time but benchmarks show <100ms."
```

**Bad note (insufficient for recovery):**
```
bd update issue-42 --notes "Working on auth feature. Made some progress.
More to do later."
```

The good note contains:
- Specific accomplishments (what was implemented/configured)
- Current state (which part is working, what's in progress)
- Next concrete step (not just "continue")
- Key context (team concerns, technical decisions with rationale)

**After compaction**: `bd show issue-42` reconstructs the full context needed to continue work.

---

## Discovery and Issue Creation {#discovery}

**When encountering new work during implementation:**

```
Discovery Workflow:
- [ ] Notice bug, improvement, or follow-up work
- [ ] Assess: Can defer or is blocker?
- [ ] Create issue with bd create "Issue title"
- [ ] Add discovered-from dependency: bd dep add current-id new-id --type discovered-from
- [ ] If blocker: pause and switch; if not: continue current work
- [ ] Issue persists for future sessions
```

**Pattern**: Proactively file issues as you discover work. Context captured immediately instead of lost when session ends.

**When to ask first**:
- Knowledge work with fuzzy scope
- User intent unclear
- Multiple valid approaches

**When to create directly**:
- Clear bug found
- Obvious follow-up work
- Technical debt with clear scope

---

## Status Maintenance {#status-maintenance}

**Throughout work on an issue:**

```
Issue Lifecycle:
- [ ] Start: Update status to in_progress
- [ ] During: Add design notes as decisions made
- [ ] During: Update acceptance criteria if requirements clarify
- [ ] During: Add dependencies if blockers discovered
- [ ] Complete: Close with summary of what was done
- [ ] After: Check bd ready to see what unblocked
```

**Pattern**: Keep bd status current so project state is always accurate.

**Status transitions**:
- `open` → `in_progress` when starting work
- `in_progress` → `blocked` if blocker discovered
- `blocked` → `in_progress` when unblocked
- `in_progress` → `closed` when complete

---

## Epic Planning {#epic-planning}

**For complex multi-step features, think in Ready Fronts, not phases.**

### The Ready Front Model

A **Ready Front** is the set of issues with all dependencies satisfied - what can be worked on *right now*. As issues close, the front advances. The dependency DAG IS the execution plan.

```
Ready Front = Issues where all dependencies are closed
              (no blockers remaining)

Static view:  Natural topology in the DAG (sync points, bottlenecks)
Dynamic view: Current wavefront of in-progress work
```

**Why Ready Fronts, not Phases?**

"Phases" trigger temporal reasoning that inverts dependencies:

```
⚠️ COGNITIVE TRAP:
"Phase 1 before Phase 2" → brain thinks "Phase 1 blocks Phase 2"
                         → WRONG: bd dep add phase1 phase2

Correct: "Phase 2 needs Phase 1" → bd dep add phase2 phase1
```

**The fix**: Name issues by what they ARE, think about what they NEED.

### Epic Planning Workflow

```
Epic Planning with Ready Fronts:
- [ ] Create epic issue for high-level goal
- [ ] Walk backward from goal: "What does the end state need?"
- [ ] Create child issues named by WHAT, not WHEN
- [ ] Add deps using requirement language: "X needs Y" → bd dep add X Y
- [ ] Verify with bd blocked (tasks blocked BY prerequisites, not dependents)
- [ ] Use bd ready to work through in dependency order
```

### The Graph Walk Pattern

Walk **backward** from the goal to get correct dependencies:

```
Start: "What's the final deliverable?"
       ↓
       "Integration tests passing" → gt-integration
       ↓
"What does that need?"
       ↓
       "Streaming support" → gt-streaming
       "Header display" → gt-header
       ↓
"What do those need?"
       ↓
       "Message rendering" → gt-messages
       ↓
"What does that need?"
       ↓
       "Buffer layout" → gt-buffer (foundation, no deps)
```

This produces correct deps because you're asking "X needs Y", not "X before Y".

### Ready Fronts Visualized

```
Ready Front 1:  gt-buffer (foundation)
Ready Front 2:  gt-messages (needs buffer)
Ready Front 3:  gt-streaming, gt-header (parallel, need messages)
Ready Front 4:  gt-integration (needs streaming, header)
```

At any moment, `bd ready` shows the current front. As issues close, blocked work becomes ready.

### Example: OAuth Integration

```bash
# Create epic (the goal)
bd create "OAuth integration" -t epic

# Walk backward: What does OAuth need?
bd create "Login/logout endpoints" -t task        # needs token storage
bd create "Token storage and refresh" -t task     # needs auth flow
bd create "Authorization code flow" -t task       # needs credentials
bd create "OAuth client credentials" -t task      # foundation

# Add deps using requirement language: "X needs Y"
bd dep add endpoints storage      # endpoints need storage
bd dep add storage flow           # storage needs flow
bd dep add flow credentials       # flow needs credentials
# credentials has no deps - it's Ready Front 1

# Verify: bd blocked should show sensible blocking
bd blocked
# endpoints blocked by storage ✓
# storage blocked by flow ✓
# flow blocked by credentials ✓
# credentials ready ✓
```

### Validation

After adding deps, verify with `bd blocked`:
- Tasks should be blocked BY their prerequisites
- NOT blocked by their dependents

If `gt-integration` is blocked by `gt-setup` → correct
If `gt-setup` is blocked by `gt-integration` → deps are inverted, fix them

---

## Side Quest Handling {#side-quests}

**When discovering work that pauses main task:**

```
Side Quest Workflow:
- [ ] During main work, discover problem or opportunity
- [ ] Create issue for side quest
- [ ] Add discovered-from dependency linking to main work
- [ ] Assess: blocker or can defer?
- [ ] If blocker: mark main work blocked, switch to side quest
- [ ] If deferrable: note in issue, continue main work
- [ ] Update statuses to reflect current focus
```

**Example**: During feature implementation, discover architectural issue

```
Main task: Adding user profiles

Discovery: Notice auth system should use role-based access

Actions:
1. Create issue: "Implement role-based access control"
2. Link: discovered-from "user-profiles-feature"
3. Assess: Blocker for profiles feature
4. Mark profiles as blocked
5. Switch to RBAC implementation
6. Complete RBAC, unblocks profiles
7. Resume profiles work
```

---

## Multi-Session Resume {#resume}

**Starting work after days/weeks away:**

```
Resume Workflow:
- [ ] Run bd ready to see available work
- [ ] Run bd stats for project overview
- [ ] List recent closed issues for context
- [ ] Show details on issue to work on
- [ ] Review design notes and acceptance criteria
- [ ] Update status to in_progress
- [ ] Begin work with full context
```

**Why this works**: bd preserves design decisions, acceptance criteria, and dependency context. No scrolling conversation history or reconstructing from markdown.

---

## Session Handoff Workflow {#session-handoff}

**Collaborative handoff between sessions using notes field:**

This workflow enables smooth work resumption by updating beads notes when stopping, then reading them when resuming. Works in conjunction with compaction survival - creates continuity even after conversation history is deleted.

### At Session Start (Claude's responsibility)

```
Session Start with in_progress issues:
- [ ] Run bd list --status in_progress
- [ ] For each in_progress issue: bd show <issue-id>
- [ ] Read notes field to understand: COMPLETED, IN PROGRESS, NEXT
- [ ] Report to user with context from notes field
- [ ] Example: "workspace-mcp-server-2 is in_progress. Last session:
       completed tidying. No code written yet. Next step: create
       markdown_to_docs.py. Should I continue with that?"
- [ ] Wait for user confirmation or direction
```

**Pattern**: Notes field is the "read me first" guide for resuming work.

### At Session End (Claude prompts user)

When wrapping up work on an issue:

```
Session End Handoff:
- [ ] Notice work reaching a stopping point
- [ ] Prompt user: "We just completed X and started Y on <issue-id>.
       Should I update the beads notes for next session?"
- [ ] If yes, suggest command:
       bd update <issue-id> --notes "COMPLETED: X. IN PROGRESS: Y. NEXT: Z"
- [ ] User reviews and confirms
- [ ] Claude executes the update
- [ ] Notes saved for next session's resumption
```

**Pattern**: Update notes at logical stopping points, not after every keystroke.

### Notes Format (Current State, Not Cumulative)

```
Good handoff note (current state):
COMPLETED: Parsed markdown into structured format
IN PROGRESS: Implementing Docs API insertion
NEXT: Debug batchUpdate call - getting 400 error on formatting
BLOCKER: None
KEY DECISION: Using two-phase approach (insert text, then apply formatting) based on reference implementation

Bad handoff note (not useful):
Updated some stuff. Will continue later.
```

**Rules for handoff notes:**
- Current state only (overwrite previous notes, not append)
- Specific accomplishments (not vague progress)
- Concrete next step (not "continue working")
- Optional: Blockers, key decisions, references
- Written for someone with zero conversation context

### Session Handoff Checklist

For Claude at session end:

```
Session End Checklist:
- [ ] Work reaching logical stopping point
- [ ] Prompt user about updating notes
- [ ] If approved:
  - [ ] Craft note with COMPLETED/IN_PROGRESS/NEXT
  - [ ] Include blocker if stuck
  - [ ] Include key decisions if relevant
  - [ ] Suggest bd update command
- [ ] Execute approved update
- [ ] Confirm: "Saved handoff notes for next session"
```

For user (optional, but helpful):

```
User Tips:
- [ ] When stopping work: Let Claude suggest notes update
- [ ] When resuming: Let Claude read notes and report context
- [ ] Avoid: Trying to remember context manually (that's what notes are for!)
- [ ] Trust: Well-written notes will help next session pick up instantly
```

### Example: Real Session Handoff

**Scenario:** Implementing markdown→Docs feature (workspace-mcp-server-2)

**At End of Session 1:**
```bash
bd update workspace-mcp-server-2 --notes "COMPLETED: Set up skeleton with Docs
API connection verified. Markdown parsing logic 80% done (handles *, _ modifiers).
IN PROGRESS: Testing edge cases for nested formatting. NEXT: Implement
batchUpdate call structure for text insertion. REFERENCE: Reference pattern at
docs/markdown-to-docs-reference.md. No blockers, moving well."
```

**At Start of Session 2:**
```bash
bd show workspace-mcp-server-2
# Output includes notes field showing exactly where we left off
# Claude reports: "Markdown→Docs feature is 80% parsed. We were testing
# edge cases and need to implement batchUpdate next. Want to continue?"
```

Session resumes instantly with full context, no history scrolling needed.

---

## Unblocking Work {#unblocking}

**When ready list is empty:**

```
Unblocking Workflow:
- [ ] Run bd blocked --json to see what's stuck
- [ ] Show details on blocked issues: bd show issue-id
- [ ] Identify blocker issues
- [ ] Choose: work on blocker, or reassess dependency
- [ ] If reassess: remove incorrect dependency
- [ ] If work on blocker: close blocker, check ready again
- [ ] Blocked issues automatically become ready when blockers close
```

**Pattern**: bd automatically maintains ready state based on dependencies. Closing a blocker makes blocked work ready.

**Example**:

```
Situation: bd ready shows nothing

Actions:
1. bd blocked shows: "api-endpoint blocked by db-schema"
2. Show db-schema: "Create user table schema"
3. Work on db-schema issue
4. Close db-schema when done
5. bd ready now shows: "api-endpoint" (automatically unblocked)
```

---

## Integration with Task tools

**Using both tools in one session:**

```
Hybrid Workflow:
- [ ] Check bd for high-level context
- [ ] Choose bd issue to work on
- [ ] Mark bd issue in_progress
- [ ] Create Task tools from acceptance criteria for execution
- [ ] Work through Task tools items
- [ ] Update bd design notes as you learn
- [ ] When Task tools complete, close bd issue
```

**Why hybrid**: bd provides persistent structure, Task tools provides visible progress.

---

## Common Workflow Patterns

### Pattern: Systematic Exploration

Research or investigation work:

```
1. Create research issue with question to answer
2. Update design field with findings as you go
3. Create new issues for discoveries
4. Link discoveries with discovered-from
5. Close research issue with conclusion
```

### Pattern: Bug Investigation

```
1. Create bug issue
2. Reproduce: note steps in description
3. Investigate: track hypotheses in design field
4. Fix: implement solution
5. Test: verify in acceptance criteria
6. Close with explanation of root cause and fix
```

### Pattern: Refactoring with Dependencies

```
1. Create issues for each refactoring step
2. Add blocks dependencies for correct order
3. Work through in dependency order
4. bd ready automatically shows next step
5. Each completion unblocks next work
```

### Pattern: Spike Investigation

```
1. Create spike issue: "Investigate caching options"
2. Time-box exploration
3. Document findings in design field
4. Create follow-up issues for chosen approach
5. Link follow-ups with discovered-from
6. Close spike with recommendation
```

---

## Checklist Templates

### Starting Any Work Session

```
- [ ] Check for .beads/ directory
- [ ] If exists: bd ready
- [ ] Report status to user
- [ ] Get user input on what to work on
- [ ] Show issue details
- [ ] Update to in_progress
- [ ] Begin work
```

### Creating Issues During Work

```
- [ ] Notice new work needed
- [ ] Create issue with clear title
- [ ] Add context in description
- [ ] Link with discovered-from to current work
- [ ] Assess blocker vs deferrable
- [ ] Update statuses appropriately
```

### Completing Work

```
- [ ] Implementation done
- [ ] Tests passing
- [ ] Close issue with summary
- [ ] Check bd ready for unblocked work
- [ ] Report completion and next available work
```

### Planning Complex Features

```
- [ ] Create epic for overall goal
- [ ] Break into child tasks
- [ ] Create all child issues
- [ ] Link with parent-child dependencies
- [ ] Add blocks between children if order matters
- [ ] Work through in dependency order
```

---

## Decision Points

**Should I create a bd issue or use Task tools?**
→ See [BOUNDARIES.md](BOUNDARIES.md) for decision matrix

**Should I ask user before creating issue?**
→ Ask if scope unclear; create if obvious follow-up work

**Should I mark work as blocked or just note dependency?**
→ Blocked = can't proceed; dependency = need to track relationship

**Should I create epic or just tasks?**
→ Epic if 5+ related tasks; tasks if simpler structure

**Should I update status frequently or just at start/end?**
→ Start and end minimum; during work if significant changes

---

## Troubleshooting Workflows

**"I can't find any ready work"**
1. Run bd blocked
2. Identify what's blocking progress
3. Either work on blockers or create new work

**"I created an issue but it's not showing in ready"**
1. Run bd show on the issue
2. Check dependencies field
3. If blocked, resolve blocker first
4. If incorrectly blocked, remove dependency

**"Work is more complex than expected"**
1. Transition from Task tools to bd mid-session
2. Create bd issue with current context
3. Note: "Discovered complexity during implementation"
4. Add dependencies as discovered
5. Continue with bd tracking

**"I closed an issue but work isn't done"**
1. Reopen with bd update status=open
2. Or create new issue linking to closed one
3. Note what's still needed
4. Closed issues can't be reopened in some systems, so create new if needed

**"Too many issues, can't find what matters"**
1. Use bd list with filters (priority, issue_type)
2. Use bd ready to focus on unblocked work
3. Consider closing old issues that no longer matter
4. Use labels for organization


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
check "name is beads" "grep -q '^name: beads' '$SKILL_DIR/SKILL.md'"
check "mentions bd CLI" "grep -q 'bd' '$SKILL_DIR/SKILL.md'"
check "mentions issue tracking" "grep -qi 'issue track' '$SKILL_DIR/SKILL.md'"
check "mentions dependency-aware" "grep -qi 'dependency' '$SKILL_DIR/SKILL.md'"
check "mentions git-backed" "grep -qi 'git' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


