---
name: codex-team
description: 'Use when you have 2+ tasks that Codex agents should execute. Claude orchestrates, Codex executes. Handles file conflicts via merge/wave strategies. Triggers: "codex team", "spawn codex", "codex agents", "use codex for", "codex fix".'
---

# Codex Team

Claude orchestrates, Codex agents execute. Each agent gets one focused task. The team lead prevents file conflicts before spawning — the orchestrator IS the lock manager.

## When to Use

- You have 2+ tasks (bug fixes, implementations, refactors)
- Tasks are well-scoped with clear instructions
- You want cross-vendor execution (GPT-5.3-Codex alongside Claude)
- Tasks don't require Claude-specific tools (Task, SendMessage, TeamCreate)

**Don't use when:** Tasks need real-time coordination, shared state, or Claude Code tools.

## Pre-Flight

```
# REQUIRED before spawning
if ! which codex > /dev/null 2>&1; then
  echo "Codex CLI not found. Install: npm i -g @openai/codex"
  # Fallback: use /swarm with Task tool instead
fi

# Model availability test
CODEX_MODEL="${CODEX_MODEL:-gpt-5.3-codex}"
if ! codex exec --full-auto -m "$CODEX_MODEL" -C "$(pwd)" "echo ok" > /dev/null 2>&1; then
  echo "Codex model $CODEX_MODEL unavailable. Falling back to Claude-only."
  # Fallback: use /swarm with Task tool instead
fi
```

## Canonical Command

```bash
codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o <output-file> "<prompt>"
```

Flag order: `--full-auto` -> `-m` -> `-C` -> `-o` -> prompt. Always this order.

**Valid flags:** `--full-auto`, `-m`, `-C`, `-o`, `--json`, `--output-schema`, `--add-dir`, `-s`

**DO NOT USE:** `-q`, `--quiet` (don't exist)

## Cross-Project Tasks

When tasks span multiple repos/directories, use `--add-dir` to grant access:

```bash
codex exec --full-auto -m gpt-5.3-codex -C "$(pwd)" --add-dir /path/to/other/repo -o output.md "prompt"
```

The `--add-dir` flag is repeatable for multiple additional directories.

## Progress Monitoring (optional)

Add `--json` to stream JSONL events to stdout for real-time monitoring:

```bash
codex exec --full-auto --json -m gpt-5.3-codex -C "$(pwd)" -o output.md "prompt" 2>/dev/null
```

Key events:
- `turn.started` / `turn.completed` — track progress
- `turn.completed` includes token `usage` field
- No events for 60s → agent likely stuck

## Sandbox Levels

Use `-s` to control the sandbox:

| Level | Flag | Use When |
|-------|------|----------|
| Read-only | `-s read-only` | Judges, reviewers (no file writes needed) |
| Workspace write | `-s workspace-write` | Default with `--full-auto` |
| Full access | `-s danger-full-access` | Only in externally sandboxed environments |

For code review and analysis tasks, prefer `-s read-only` over `--full-auto`.

## Execution

### Step 1: Define Tasks

Break work into focused tasks. Each task = one Codex agent (unless merged).

### Step 2: Analyze File Targets (REQUIRED)

**Before spawning, identify which files each task will edit.** Codex agents are headless — they can't negotiate locks or wait turns. All conflict prevention happens here.

For each task, list the target files. Then apply the right strategy:

| File Overlap | Strategy | Action |
|-------------|----------|--------|
| All tasks touch same file | **Merge** | Combine into 1 agent with all fixes |
| Some tasks share files | **Multi-wave** | Shared-file tasks go sequential across waves |
| No overlap | **Parallel** | Spawn all agents at once |

```
# Decision logic (team lead performs this mentally):

tasks = [
  {name: "fix spec_path",    files: ["cmd/zeus.go"]},
  {name: "remove beads field", files: ["cmd/zeus.go"]},
  {name: "fix dispatch counter", files: ["cmd/zeus.go"]},
]

# All touch zeus.go → MERGE into 1 agent
```

```
tasks = [
  {name: "fix auth bug",     files: ["pkg/auth.go"]},
  {name: "add rate limiting", files: ["pkg/auth.go", "pkg/middleware.go"]},
  {name: "update config",    files: ["internal/config.go"]},
]

# Task 1 and 2 share auth.go → MULTI-WAVE (1+3 parallel, then 2)
# Task 3 is independent → runs in Wave 1 alongside Task 1
```

```
tasks = [
  {name: "fix auth",    files: ["pkg/auth.go"]},
  {name: "fix config",  files: ["internal/config.go"]},
  {name: "fix logging", files: ["pkg/log.go"]},
]

# No overlap → PARALLEL (all 3 at once)
```

### Step 3: Spawn Agents

**Strategy: Parallel (no file overlap)**

Spawn all agents in a single response using Bash with `run_in_background=true`:

```
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/auth-fix.md "Fix the null check in pkg/auth.go:validateToken around line 89..."', run_in_background=true)
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/config-fix.md "Add timeout field to internal/config.go:Config struct..."', run_in_background=true)
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/logging-fix.md "Fix log rotation in pkg/log.go:rotateLogFile..."', run_in_background=true)
```

**Strategy: Merge (same file)**

Combine all fixes into a single agent prompt:

```
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/zeus-fixes.md \
  "Fix these 3 issues in cmd/zeus.go: \
   (1) Line 245: rename spec_path to spec_location in QUEST_REQUEST payload \
   (2) Line 250: remove the spurious beads field from the payload \
   (3) Line 196: fix dispatch counter — increment inside the loop, not outside"', run_in_background=true)
```

One agent, one file, no conflicts possible.

**Strategy: Multi-wave (partial overlap)**

```
# Wave 1: non-overlapping tasks
Bash(command='codex exec ... -o .agents/codex-team/auth-fix.md "Fix null check in pkg/auth.go:89..."', run_in_background=true)
Bash(command='codex exec ... -o .agents/codex-team/config-fix.md "Add timeout to internal/config.go..."', run_in_background=true)

# Wait for Wave 1
TaskOutput(task_id="<id-1>", block=true, timeout=120000)
TaskOutput(task_id="<id-2>", block=true, timeout=120000)

# Read Wave 1 results — understand what changed
Read(.agents/codex-team/auth-fix.md)
git diff pkg/auth.go

# Wave 2: task that shares files with Wave 1
Bash(command='codex exec ... -o .agents/codex-team/rate-limit.md \
  "Add rate limiting to pkg/auth.go and pkg/middleware.go. \
   Note: pkg/auth.go was recently modified — the validateToken function now has a null check at line 89. \
   Build on the current state of the file."', run_in_background=true)

TaskOutput(task_id="<id-3>", block=true, timeout=120000)
```

The team lead synthesizes Wave 1 results and injects relevant context into Wave 2 prompts. Don't dump raw diffs — describe what changed and why it matters for the next task.

### Step 4: Wait for Completion

```
TaskOutput(task_id="<id-1>", block=true, timeout=120000)
TaskOutput(task_id="<id-2>", block=true, timeout=120000)
TaskOutput(task_id="<id-3>", block=true, timeout=120000)
```

### Step 5: Verify Results

- Read output files from `.agents/codex-team/`
- Check `git diff` for changes made by each agent
- Run tests if applicable
- For multi-wave: verify Wave 2 agents built correctly on Wave 1 changes

## Output Directory

```
mkdir -p .agents/codex-team
```

Output files: `.agents/codex-team/<task-name>.md`

## Prompt Guidelines

Good Codex prompts are **specific and self-contained**:

```
# GOOD: Specific file, line, exact change
"Fix in cmd/zeus.go line 245: rename spec_path to spec_location in the QUEST_REQUEST payload struct"

# BAD: Vague, requires exploration
"Fix the spec path issue somewhere in the codebase"
```

Include in each prompt:
- Exact file path(s)
- Line numbers or function names
- What to change and why
- Any constraints (don't touch other files, preserve API compatibility)

For multi-wave Wave 2+ prompts, also include:
- What changed in prior waves (summarized, not raw diffs)
- Current state of shared files after prior edits

## Limits

- **Max agents:** 6 per wave (resource-reasonable)
- **Timeout:** 2 minutes default per agent. Increase with `timeout` param for larger tasks
- **Max waves:** 3 recommended. If you need more, reconsider task decomposition

## Fallback

If Codex is unavailable, use `/swarm` with Claude-native Task tool agents:

```
Task(description="Fix task 1", subagent_type="general-purpose", run_in_background=true, prompt="...")
```

## Quick Reference

| Item | Value |
|------|-------|
| Model | `gpt-5.3-codex` |
| Command | `codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o <file> "prompt"` |
| Output dir | `.agents/codex-team/` |
| Max agents/wave | 6 recommended |
| Timeout | 120s default |
| Strategies | Parallel (no overlap), Merge (same file), Multi-wave (partial overlap) |
| Fallback | `/swarm` (Claude Task tool) |
