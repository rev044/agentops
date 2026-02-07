---
name: codex-team
description: 'Use when you have 2+ independent tasks that Codex agents should execute in parallel. Claude orchestrates, Codex executes. Triggers: "codex team", "spawn codex", "codex agents", "use codex for", "codex fix".'
---

# Codex Team

Claude orchestrates, Codex agents execute in parallel. Each agent gets one focused task with a clear prompt and output file.

## When to Use

- You have 2+ independent tasks (bug fixes, implementations, refactors)
- Tasks are well-scoped with clear instructions
- You want cross-vendor execution (GPT-5.3-Codex alongside Claude)
- Tasks don't require Claude-specific tools (Task, SendMessage, TeamCreate)

**Don't use when:** Tasks need coordination, shared state, or Claude Code tools.

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

**Valid flags ONLY:** `--full-auto`, `-m`, `-C`, `-o`, `--json`, `--output-schema`

**DO NOT USE:** `-q`, `--quiet` (don't exist)

## Execution

### Step 1: Define Tasks

Break work into independent, focused tasks. Each task = one Codex agent.

### Step 2: Spawn Agents (Parallel)

Spawn all agents in a single response using Bash with `run_in_background=true`:

```
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/task-1.md "Fix the spec_path field in cmd/zeus.go line 245..."', run_in_background=true)
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/task-2.md "Remove the spurious beads field from the payload..."', run_in_background=true)
Bash(command='codex exec --full-auto -m "gpt-5.3-codex" -C "$(pwd)" -o .agents/codex-team/task-3.md "Fix the dispatch counter bug around line 196..."', run_in_background=true)
```

### Step 3: Wait for Completion

```
TaskOutput(task_id="<id-1>", block=true, timeout=120000)
TaskOutput(task_id="<id-2>", block=true, timeout=120000)
TaskOutput(task_id="<id-3>", block=true, timeout=120000)
```

### Step 4: Verify Results

- Read output files from `.agents/codex-team/`
- Check `git diff` for changes made by each agent
- Resolve any conflicts between agents editing the same file
- Run tests if applicable

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

## Limits

- **Max agents:** Keep to 6 or fewer (resource-reasonable)
- **Timeout:** 2 minutes default per agent. Increase with `timeout` param for larger tasks
- **Same-file edits:** If multiple agents edit the same file, review diffs carefully for conflicts

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
| Max agents | 6 recommended |
| Timeout | 120s default |
| Fallback | `/swarm` (Claude Task tool) |
