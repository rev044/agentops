# Runtime Spawning Commands

## Backend Selection (MANDATORY)

Select backend in this order:

1. `spawn_agent` available -> **Codex experimental sub-agents**
2. `TeamCreate` available -> **Claude native teams**
3. Otherwise -> **Foreground `Task()`** (one-shot subagents, write to files)

> **NEVER use `Task(run_in_background=true)` for Claude agents.** It causes instability. Use foreground `Task()` calls (parallel via multiple calls in one message) or native teams. `Bash(run_in_background=true)` for Codex CLI processes is fine.

This keeps `/council` universal across Codex and Claude runtimes.

## Codex Judges (experimental sub-agents)

Spawn judges directly as sub-agents:

```
spawn_agent(message="{JUDGE_PACKET for judge-1}")
spawn_agent(message="{JUDGE_PACKET for judge-2}")
spawn_agent(message="{JUDGE_PACKET for judge-3}")
```

Track judge mapping (`judge-1 -> <agent-id>`) for debate and cleanup.

Wait for completion:

```
wait(ids=["<judge-1-id>", "<judge-2-id>", "<judge-3-id>"], timeout_ms=120000)
```

Debate R2 follow-up:

```
send_input(
  id="<judge-1-id>",
  message="## Debate Round 2\n\nOther judges' R1 verdicts:\n\n{OTHER_VERDICTS_JSON}\n\n{DEBATE_INSTRUCTIONS}"
)
```

Cleanup:

```
close_agent(id="<judge-1-id>")
close_agent(id="<judge-2-id>")
close_agent(id="<judge-3-id>")
```

## Claude Judges (via Native Teams)

Create the council team before spawning judges:

```
TeamCreate(team_name="council-YYYYMMDD-<target>")
```

Team naming convention: `council-YYYYMMDD-<target>` (e.g., `council-20260206-auth-system`).

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

Spawn judges as foreground Task calls. Multiple Task calls in the same message run in parallel.

```
Task(
  description="Council judge 1",
  subagent_type="general-purpose",
  model="opus",
  prompt="{JUDGE_PACKET}"
)
Task(
  description="Council judge 2",
  subagent_type="general-purpose",
  model="opus",
  prompt="{JUDGE_PACKET}"
)
```

## Codex Judges (via Codex CLI for --mixed)

Use Codex CLI to add cross-vendor judges in `--mixed` mode:

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
# Codex runtime backend:
spawn_agent(message="{judge-1 packet}")
spawn_agent(message="{judge-2 packet}")
spawn_agent(message="{judge-3 packet}")

# Claude runtime backend:
TeamCreate(team_name="council-YYYYMMDD-<target>")
Task(description="Judge 1", team_name="council-...", name="judge-1", ...)
Task(description="Judge 2", team_name="council-...", name="judge-2", ...)
Task(description="Judge 3", team_name="council-...", name="judge-3", ...)

# Optional Codex CLI judges for --mixed mode:
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

Judges send a MINIMAL completion signal (verdict + confidence + file path only).
The lead reads their output files for full analysis during consolidation.

- Codex sub-agent backend: `wait(ids=[...])`
- Claude teams backend: completion via `SendMessage` (minimal JSON signal)
- Codex CLI judges: `TaskOutput(task_id="...", block=true)` — output file on disk

## Debate Round 2 (via send_input or SendMessage)

**After R1 completes, send R2 instructions to existing judges (no re-spawn):**

CONTEXT BUDGET: Send only verdict summaries (verdict + confidence + file path) of other
judges, NOT their full JSON findings. Judges read other judges' files if they need detail.

```
# Determine branch
r1_unanimous = all R1 verdicts have same value

# Codex backend:
send_input(id="<judge-1-id>", message="## Debate Round 2\n\nOther judges' R1 verdicts:\n\njudge-2: WARN (HIGH) — file: .agents/council/...-claude-2.md\njudge-3: PASS (MEDIUM) — file: .agents/council/...-claude-3.md\n\nRead their files for full reasoning.\n\n{DEBATE_INSTRUCTIONS}")

# Claude teams backend:
SendMessage(type="message", recipient="judge-1", content="## Debate Round 2\n\nOther judges' R1 verdicts:\n\njudge-2: WARN (HIGH) — file: .agents/council/...-claude-2.md\njudge-3: PASS (MEDIUM) — file: .agents/council/...-claude-3.md\n\nRead their files for full reasoning.\n\n{DEBATE_INSTRUCTIONS}", summary="Debate R2: review other verdicts")
```

Judges wake from idle, read other judges' files, process R2, write R2 files, send minimal completion signal.

**R2 completion wait:** wait up to `COUNCIL_R2_TIMEOUT` (default 90s) using backend channel (`wait` or `SendMessage`). If a judge does not respond, read their R1 output file and use the R1 verdict for consolidation. Log: `Judge <name> R2 timeout -- using R1 verdict.`

## Cleanup

**After consolidation:**

```
# Codex backend:
close_agent(id="<judge-1-id>")
close_agent(id="<judge-2-id>")
close_agent(id="<judge-3-id>")

# Claude teams backend:
SendMessage(type="shutdown_request", recipient="judge-1", content="Council complete")
SendMessage(type="shutdown_request", recipient="judge-2", content="Council complete")
SendMessage(type="shutdown_request", recipient="judge-3", content="Council complete")
TeamDelete()
```

> **Note:** `TeamDelete()` applies only to Claude team backend. Codex backend cleanup is explicit via `close_agent()`.

## Reaper Cleanup Pattern

Cleanup MUST succeed even on partial failures. Follow this sequence:

1. **Attempt graceful close/shutdown:** `close_agent` (Codex) or `shutdown_request` (Claude)
2. **Wait up to 30s** for shutdown_approved responses
3. **If any judge doesn't respond:** Log warning, proceed anyway
4. **Always run backend cleanup** (`close_agent` or `TeamDelete`)
5. **Cleanup must complete best-effort**

**Failure modes and recovery:**

| Failure | Behavior |
|---------|----------|
| Judge hangs (no response) | 30s timeout -> proceed with cleanup |
| close/shutdown fails | Log warning -> continue cleanup |
| TeamDelete fails | Log error -> team orphaned (manual cleanup: delete ~/.claude/teams/<name>/) |
| close_agent fails | Log error -> leaked sub-agent handle (best effort close later) |
| Lead crashes mid-council | Cleanup may be deferred to session end |

**Never skip cleanup.** Lingering workers pollute future sessions.

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
