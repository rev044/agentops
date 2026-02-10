# CLI Spawning Commands

## Team Setup

**Create the council team before spawning judges:**

```
TeamCreate(team_name="council-YYYYMMDD-<target>")
```

Team naming convention: `council-YYYYMMDD-<target>` (e.g., `council-20260206-auth-system`).

## Claude Agents (via Native Teams)

**Spawn judges as teammates on the council team:**

Default (independent judges, no perspectives):
```
Task(
  description="Council judge 1",
  subagent_type="general-purpose",
  model="opus",
  team_name="council-YYYYMMDD-<target>",
  name="judge-1",
  prompt="{JUDGE_DEFAULT_PROMPT}"
)
```

With perspectives (--preset or --perspectives):
```
Task(
  description="Council judge: Error-Paths",
  subagent_type="general-purpose",
  model="opus",
  team_name="council-YYYYMMDD-<target>",
  name="judge-error-paths",
  prompt="{JUDGE_PERSPECTIVE_PROMPT}"
)
```

Judges join the team, write output files, and send completion messages to the team lead via `SendMessage`.

**Fallback (if native teams unavailable):**

```
Task(
  description="Council judge 1",
  subagent_type="general-purpose",
  model="opus",
  run_in_background=true,
  prompt="{JUDGE_PACKET}"
)
```

## Codex Agents (via Codex CLI)

**Canonical Codex command form (unchanged -- Codex cannot join teams):**

```bash
# With structured output (preferred -- requires --output-schema support)
codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-{N}.json "{PACKET}"

# Fallback (if --output-schema unsupported by model)
codex exec --full-auto -m gpt-5.3-codex -C "$(pwd)" -o .agents/council/codex-{N}.md "{PACKET}"
```

Always use this exact flag order: `-s` / `--full-auto` -> `-m` -> `-C` -> `--output-schema` (if applicable) -> `-o` -> prompt.

**Codex CLI flags (ONLY these are valid):**
- `--full-auto` -- No approval prompts (REQUIRED for fallback, always first)
- `-s read-only` / `-s workspace-write` -- Sandbox level (read-only for judges, workspace-write for workers)
- `-m <model>` -- Model override (default: gpt-5.3-codex)
- `-C <dir>` -- Working directory
- `--output-schema <file>` -- Enforce structured JSON output (requires `additionalProperties: false` in schema)
- `-o <file>` -- Output file (use `-o` not `--output`). Extension `.json` when using `--output-schema`, `.md` for fallback.
- `--add-dir <dir>` -- Additional writable directories (repeatable)

**DO NOT USE:** `-q` (doesn't exist), `--quiet` (doesn't exist)

## Parallel Spawning

**Spawn all agents in parallel:**

```
# Step 1: Create team
TeamCreate(team_name="council-YYYYMMDD-<target>")

# Step 2: Spawn Claude judges as teammates (parallel)
# Default (independent -- no perspectives):
Task(description="Judge 1", team_name="council-...", name="judge-1", ...)
Task(description="Judge 2", team_name="council-...", name="judge-2", ...)
Task(description="Judge 3", team_name="council-...", name="judge-3", ...)
# With --preset or --perspectives:
# Task(description="Judge: Error-Paths", team_name="council-...", name="judge-error-paths", ...)

# Step 3: Spawn Codex agents (Bash tool, parallel -- cannot join teams)
# With --output-schema (preferred, when SCHEMA_SUPPORTED=true):
Bash(command="codex exec -s read-only -m gpt-5.3-codex -C \"$(pwd)\" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-1.json ...", run_in_background=true)
Bash(command="codex exec -s read-only -m gpt-5.3-codex -C \"$(pwd)\" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-2.json ...", run_in_background=true)
Bash(command="codex exec -s read-only -m gpt-5.3-codex -C \"$(pwd)\" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-3.json ...", run_in_background=true)
# Fallback (when SCHEMA_SUPPORTED=false):
# Bash(command="codex exec --full-auto -m gpt-5.3-codex -C \"$(pwd)\" -o .agents/council/codex-1.md ...", run_in_background=true)
# Bash(command="codex exec --full-auto -m gpt-5.3-codex -C \"$(pwd)\" -o .agents/council/codex-2.md ...", run_in_background=true)
# Bash(command="codex exec --full-auto -m gpt-5.3-codex -C \"$(pwd)\" -o .agents/council/codex-3.md ...", run_in_background=true)
```

**Wait for completion:**

Judges send completion messages to the team lead via `SendMessage`. These arrive automatically as conversation turns. For Codex agents, use `TaskOutput(task_id="...", block=true)`.

## Debate Round 2 (via SendMessage)

**After R1 completes, send R2 instructions to existing judges (no re-spawn):**

```
# Determine branch
r1_unanimous = all R1 verdicts have same value

# Send to each judge
SendMessage(
  type="message",
  recipient="judge-1",  # or "judge-{perspective}" with presets
  content="## Debate Round 2\n\nOther judges' R1 verdicts:\n\n{OTHER_VERDICTS_JSON}\n\n{DEBATE_INSTRUCTIONS_FOR_BRANCH}",
  summary="Debate R2: review other verdicts"
)
```

Judges wake from idle, process R2, write R2 files, send completion message.

**R2 completion wait:** After sending R2 debate messages to all judges, wait up to `COUNCIL_R2_TIMEOUT` (default 90s) for each judge's completion message via `SendMessage`. If a judge does not respond within the timeout, read their R1 output file (`.agents/council/YYYY-MM-DD-<target>-claude-{perspective}.md`) and use the R1 verdict for consolidation. Log: `Judge <name> R2 timeout -- using R1 verdict.`

## Team Cleanup

**After consolidation:**

```
# Shutdown each judge
SendMessage(type="shutdown_request", recipient="judge-1", content="Council complete")
SendMessage(type="shutdown_request", recipient="judge-2", content="Council complete")
SendMessage(type="shutdown_request", recipient="judge-3", content="Council complete")
# With presets: use judge-{perspective} names instead (e.g., judge-error-paths)

# Delete team
TeamDelete()
```

> **Note:** `TeamDelete()` deletes the team associated with this session's `TeamCreate()` call. If running concurrent teams (e.g., council inside crank), each team is cleaned up in the session that created it. No team name parameter is needed -- the API tracks the current session's team context automatically.

## Reaper Cleanup Pattern

Team cleanup MUST succeed even on partial failures. Follow this sequence:

1. **Attempt graceful shutdown:** Send shutdown_request to each judge
2. **Wait up to 30s** for shutdown_approved responses
3. **If any judge doesn't respond:** Log warning, proceed anyway
4. **Always call TeamDelete()** -- even if some judges are unresponsive
5. **TeamDelete cleans up** the team regardless of member state

**Failure modes and recovery:**

| Failure | Behavior |
|---------|----------|
| Judge hangs (no response) | 30s timeout -> proceed to TeamDelete |
| shutdown_request fails | Log warning -> proceed to TeamDelete |
| TeamDelete fails | Log error -> team orphaned (manual cleanup: delete ~/.claude/teams/<name>/) |
| Lead crashes mid-council | Team orphaned until session ends or manual cleanup |

**Never skip TeamDelete.** A lingering team config pollutes future sessions.

## Team Timeout Configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| Judge timeout | 120s | Max time for judge to complete (per round) |
| Shutdown grace period | 30s | Time to wait for shutdown_approved |
| R2 debate timeout | 90s | Max time for R2 completion after sending debate messages |

## Model Selection

| Vendor | Default | Override |
|--------|---------|----------|
| Claude | opus | `--claude-model=sonnet` |
| Codex | gpt-5.3-codex | `--codex-model=<model>` |

## Output Collection

All council outputs go to `.agents/council/`:

```bash
# Ensure directory exists
mkdir -p .agents/council

# Claude output (R1) -- independent judges
.agents/council/YYYY-MM-DD-<target>-claude-1.md

# Claude output (R1) -- with presets
.agents/council/YYYY-MM-DD-<target>-claude-error-paths.md

# Claude output (R2, when --debate)
.agents/council/YYYY-MM-DD-<target>-claude-1-r2.md

# Codex output (R1 only, even with --debate)
# When --output-schema is supported:
.agents/council/YYYY-MM-DD-<target>-codex-1.json
# Fallback (no --output-schema):
.agents/council/YYYY-MM-DD-<target>-codex-1.md

# Final consolidated report
.agents/council/YYYY-MM-DD-<target>-report.md
```
