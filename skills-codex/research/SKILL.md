---
name: research
description: 'Deep codebase exploration. Triggers: research, explore, investigate, understand, deep dive, current state.'
---


# Research Skill

> **Quick Ref:** Deep codebase exploration with multi-angle analysis. Output: `.agents/research/*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** ao (knowledge injection — optional). If ao is unavailable, skip prior knowledge search and proceed with direct codebase exploration.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--auto` | off | Skip human approval gate. Used by `$rpi --auto` for fully autonomous lifecycle. |

## Execution Steps

Given `$research <topic> [--auto]`:

### Step 1: Create Output Directory
```bash
mkdir -p .agents/research
```

### Step 2: Check Prior Art

**First, search and inject existing knowledge (if ao available):**

```bash
# Pull relevant prior knowledge for this topic
ao know lookup --query "<topic>" --limit 5 2>/dev/null || \
  ao know search "<topic>" 2>/dev/null || \
  echo "ao not available, skipping knowledge search"
```

**Review ao know search results:** If ao returns relevant learnings or patterns, incorporate them into your research strategy. Look for:
- Prior research on this topic or related topics
- Known patterns or anti-patterns
- Lessons learned from similar investigations

**Search ALL local knowledge locations by content (not just filename):**

Use Grep to search every knowledge directory for the topic. This catches learnings from `$learn`, retros, brainstorms, and plans — not just research artifacts.

```bash
# Search all knowledge locations by content
for dir in research learnings knowledge patterns retros plans brainstorm; do
  grep -r -l -i "<topic>" .agents/${dir}/ 2>/dev/null
done

# Search global patterns (cross-repo knowledge)
grep -r -l -i "<topic>" ~/.codex/patterns/ 2>/dev/null
```

If matches are found, read the relevant files with the Read tool before proceeding to exploration. Prior knowledge prevents redundant investigation.

### Step 2.5: Pre-Flight — Detect Spawn Backend

Before launching the explore agent, detect which backend is available:

1. Check if `spawn_agent` is available → log `"Backend: codex-sub-agents"`
2. Else check if runtime-native team spawning is available → log `"Backend: runtime-native-teams"`
3. Else check if `skill` tool is read-only (OpenCode) → log `"Backend: opencode-subagents"`
4. Else check if `Task` is available → log `"Backend: background-task-fallback"`
5. Else → log `"Backend: inline (no spawn available)"`

Record the selected backend — it will be included in the research output document for traceability.

**Read the matching backend reference for concrete tool call examples:**
- Claude feature contract → `..$shared/references/claude-code-latest-features.md`
- Codex → `..$shared/references/backend-codex-subagents.md`
- Claude Native Teams → `..$shared/references/backend-claude-teams.md`
- Background Tasks → `..$shared/references/backend-background-tasks.md`
- Inline → `..$shared/references/backend-inline.md`

### Step 3: Launch Explore Agent

**YOU MUST DISPATCH AN EXPLORATION AGENT NOW.** Select the backend using capability detection:

#### Backend Selection (MANDATORY)

1. If `spawn_agent` is available → **Codex sub-agent** (preferred)
2. Else if runtime-native team spawning is available → **Runtime-native team** (Explore agent)
3. Else if `skill` tool is read-only (OpenCode) → **OpenCode subagent** — `task(subagent_type="explore", description="Research: <topic>", prompt="<explore prompt>")`
4. Else → **Background task fallback**

#### Exploration Prompt (all backends)

Use this prompt for whichever backend is selected:

```
Thoroughly investigate: <topic>

Discovery tiers (execute in order, skip if source unavailable):

Tier 1 — Code-Map (fastest, authoritative):
  Read docs/code-map/README.md → find <topic> category
  Read docs/code-map/{feature}.md → get exact paths and function names
  Skip if: no docs/code-map/ directory

Tier 2 — Semantic Search (conceptual matches):
  mcp__smart-connections-work__lookup query="<topic>" limit=10
  Skip if: MCP not connected

Tier 2.5 — Git History (recent changes and decision context):
  git log --oneline -30 -- <topic-related-paths>   # scoped to relevant paths, cap 30 lines
  git log --all --oneline --grep="<topic>" -10      # cap 10 matches
  git blame <key-file> | grep -i "<topic>" | head -20  # cap 20 lines
  Skip if: not a git repo, no relevant history, or <topic> too broad (>100 matches)
  NEVER: git log on full repo without -- path filter (same principle as Tier 3 scoping)
  NOTE: This is git commit history, not session history. For session/handoff history, use $trace.

Tier 3 — Scoped Search (keyword precision):
  Grep("<topic>", path="<specific-dir>/")   # ALWAYS scope to a directory
  Glob("<specific-dir>/**/*.py")            # ALWAYS scope to a directory
  NEVER: Grep("<topic>") or Glob("**/*.py") on full repo — causes context overload

Tier 4 — Source Code (verify from signposts):
  Read files identified by Tiers 1-3 (including git history leads from Tier 2.5)
  Use function/class names, not line numbers

Tier 5 — Prior Knowledge (may be stale):
  Search ALL .agents/ knowledge dirs by content:
    for dir in research learnings knowledge patterns retros plans brainstorm; do
      grep -r -l -i "<topic>" .agents/${dir}/ 2>/dev/null
    done
  Read matched files. Cross-check findings against current source.

Tier 6 — External Docs (last resort):
  WebSearch for external APIs or standards
  Only when Tiers 1-5 are insufficient

Return a detailed report with:
- Key files found (with paths)
- How the system works
- Important patterns or conventions
- Any issues or concerns

Cite specific file:line references for all claims.
```

#### Spawn Research Agents

If your runtime supports spawning parallel subagents, spawn one or more research agents with the exploration prompt. Each agent explores independently and writes findings to `.agents/research/`.

If no multi-agent capability is available, perform the exploration **inline** in the current session using file reading, grep, and glob tools directly.

### Step 4: Validate Research Quality (Optional)

**For thorough research, perform quality validation:**

#### 4a. Coverage Validation
Check: Did we look everywhere we should? Any unexplored areas?
- List directories/files explored
- Identify gaps in coverage
- Note areas that need deeper investigation

#### 4b. Depth Validation
Check: Do we UNDERSTAND the critical parts? HOW and WHY, not just WHAT?
- Rate depth (0-4) for each critical area
- Flag areas with shallow understanding
- Identify what needs more investigation

#### 4c. Gap Identification
Check: What DON'T we know that we SHOULD know?
- List critical gaps
- Prioritize what must be filled before proceeding
- Note what can be deferred

#### 4d. Assumption Challenge
Check: What assumptions are we building on? Are they verified?
- List assumptions made
- Flag high-risk unverified assumptions
- Note what needs verification

### Step 5: Synthesize Findings

After the Explore agent and validation swarm return, write findings to:
`.agents/research/YYYY-MM-DD-<topic-slug>.md`

Use this format:
```markdown
# Research: <Topic>

**Date:** YYYY-MM-DD
**Backend:** <codex-sub-agents | claude-native-teams | background-task-fallback | inline>
**Scope:** <what was investigated>

## Summary
<2-3 sentence overview>

## Key Files
| File | Purpose |
|------|---------|
| path/to/file.py | Description |

## Findings
<detailed findings with file:line citations>

## Recommendations
<next steps or actions>
```

### Step 6: Request Human Approval (Gate 1)

**Skip this step if `--auto` flag is set.** In auto mode, proceed directly to Step 7.

**USE AskUserQuestion tool:**

```
Tool: AskUserQuestion
Parameters:
  questions:
    - question: "Research complete. Approve to proceed to planning?"
      header: "Gate 1"
      options:
        - label: "Approve"
          description: "Research is sufficient, proceed to $plan"
        - label: "Revise"
          description: "Need deeper research on specific areas"
        - label: "Abandon"
          description: "Stop this line of investigation"
      multiSelect: false
```

**Wait for approval before reporting completion.**

### Step 7: Report to User

Tell the user:
1. What you found
2. Where the research doc is saved
3. Gate 1 approval status
4. Next step: `$plan` to create implementation plan

## Key Rules

- **Actually dispatch the Explore agent** - don't just describe doing it
- **Scope searches** - use the topic to narrow file patterns
- **Cite evidence** - every claim needs `file:line`
- **Write output** - research must produce a `.agents/research/` artifact

## Thoroughness Levels

Include in your Explore agent prompt:
- "quick" - for simple questions
- "medium" - for feature exploration
- "very thorough" - for architecture/cross-cutting concerns

## Examples

### Investigate Authentication System

**User says:** `$research "authentication system"`

**What happens:**
1. Agent searches knowledge base for prior auth research
2. Explore agent investigates via Code-Map, Grep, and file reading
3. Findings synthesized with file:line citations
4. Output written to `.agents/research/2026-02-13-authentication-system.md`

**Result:** Detailed report identifying auth middleware location, session handling, and token validation patterns.

### Quick Exploration of Cache Layer

**User says:** `$research "cache implementation"`

**What happens:**
1. Agent uses Glob to find cache-related files
2. Explore agent reads key files and summarizes current state
3. No prior research found, proceeds with fresh exploration
4. Output written to `.agents/research/2026-02-13-cache-implementation.md`

**Result:** Summary of cache strategy, TTL settings, and eviction policies with file references.

### Deep Dive into Payment Flow

**User says:** `$research "payment processing flow"`

**What happens:**
1. Agent loads prior payment research from knowledge base
2. Explore agent traces flow through multiple services
3. Identifies integration points and error handling
4. Output written with cross-service file citations

**Result:** End-to-end payment flow diagram with file paths and critical decision points.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Research too shallow | Default exploration depth insufficient for the topic | Re-run with broader scope or specify additional search areas |
| Research output too large | Exploration covered too many tangential areas | Narrow the goal to a specific question rather than a broad topic |
| Missing file references | Codebase has changed since last exploration or files are in unexpected locations | Use Glob to verify file locations before citing them. Always use absolute paths |
| Auto mode skips important areas | Automated exploration prioritizes breadth over depth | Remove `--auto` flag to enable human approval gate for guided exploration |
| Explore agent times out | Topic too broad for single exploration pass | Split into smaller focused topics (e.g., "auth flow" vs "entire auth system") |
| No backend available for spawning | Running in environment without spawn_agent or runtime-native team support | Research runs inline — still functional but slower |

## Reference Documents

- [references/backend-background-tasks.md](references/backend-background-tasks.md)
- [references/backend-claude-teams.md](references/backend-claude-teams.md)
- [references/backend-codex-subagents.md](references/backend-codex-subagents.md)
- [references/backend-inline.md](references/backend-inline.md)
- [references/claude-code-latest-features.md](references/claude-code-latest-features.md)
- [references/context-discovery.md](references/context-discovery.md)
- [references/document-template.md](references/document-template.md)
- [references/failure-patterns.md](references/failure-patterns.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)
- [references/vibe-methodology.md](references/vibe-methodology.md)

---

## References

### backend-background-tasks.md

# Backend: Background Tasks (Fallback)

Concrete tool calls for spawning agents using `Task(run_in_background=true)`. This is the **last-resort fallback** when neither Codex sub-agents nor Claude native teams are available.

**When detected:** `Task` tool is available but `TeamCreate` and `spawn_agent` are not.

**Limitations:**
- Fire-and-forget — no messaging, no redirect, no scope adjustment
- No inter-agent communication
- No debate mode (R2 requires messaging)
- No retry (must re-spawn from scratch)
- No graceful shutdown (only `TaskStop`, which is lossy)

---

## Spawn: Background Agents

Spawn agents with `Task(run_in_background=true)`. Each call returns a `task_id` for later polling.

### Council Judges

```
Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are judge-1.\n\nYour perspective: Correctness & Completeness\n\n<PACKET>\n...\n</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-1.md\nThis is your ONLY output channel — there is no messaging.",
  description="Council judge-1"
)
# Returns: task_id="abc-123"

Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are judge-error-paths.\n\nYour perspective: Error Paths & Edge Cases\n\n<PACKET>...</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-error-paths.md",
  description="Council judge-error-paths"
)
# Returns: task_id="def-456"
```

Both `Task` calls go in the **same message** — they run in parallel.

### Swarm Workers

```
Task(
  subagent_type="general-purpose",
  run_in_background=true,
  prompt="You are worker-3.\n\nYour Assignment: Task #3: Add password hashing\n...\n\nWrite result to .agents/swarm/results/3.json\nDo NOT run git add/commit/push.",
  description="Swarm worker-3"
)
```

### Research Explorers

```
Task(
  subagent_type="Explore",
  run_in_background=true,
  prompt="Thoroughly investigate: authentication patterns...\n\nWrite findings to .agents/research/2026-02-17-auth.md",
  description="Research explorer"
)
```

---

## Wait: Poll for Completion

Background tasks have no messaging. Poll with `TaskOutput`.

```
TaskOutput(task_id="abc-123", block=true, timeout=120000)
TaskOutput(task_id="def-456", block=true, timeout=120000)
```

**Or non-blocking check:**

```
TaskOutput(task_id="abc-123", block=false, timeout=5000)
```

**After `TaskOutput` returns**, verify the agent wrote its result file:

```
Read(".agents/council/2026-02-17-auth-judge-1.md")
```

**Timeout behavior:** If `timeout` expires, `TaskOutput` returns with a timeout status — the agent may still be running. **Recovery:**
1. Check result file — agent may have written it but not finished cleanly
2. If result file exists → use it, `TaskStop` the agent
3. If no result file → agent failed silently. For council: proceed with N-1 verdicts, note in report. For swarm: add task back to retry queue, re-spawn a fresh agent.
4. Never assume `TaskOutput` completion means the result file was written — always verify

**Fallback:** If background tasks fail despite detection, fall back to inline mode. See `backend-inline.md`.

---

## No Messaging

Background tasks cannot receive messages. This means:

- **No debate R2** — judges get one round only
- **No retry** — if validation fails, re-spawn a new agent from scratch
- **No scope adjustment** — the prompt is final at spawn time

---

## Cleanup

Background tasks self-terminate when done. For stuck tasks:

```
TaskStop(task_id="abc-123")
```

This is lossy — partial work may be lost.

---

## Key Rules

1. **Filesystem is the only communication channel** — agents write files, lead reads files
2. **No messaging = no debate** — `--debate` is unavailable with this backend
3. **No retry = must re-spawn** — failed agents get a fresh `Task` call, not a message
4. **Always check result files** — `TaskOutput` completion doesn't guarantee the agent wrote its file
5. **Prefer native teams** — this backend is strictly inferior; use it only as last resort

### backend-claude-teams.md

# Backend: Claude Native Teams

Concrete tool calls for spawning agents using Codex native teams (`TeamCreate` + `SendMessage` + shared `TaskList`).

**When detected:** `TeamCreate` tool is available in your tool list.

---

## Pre-Flight: Confirm Modern Claude Features

Before spawning teammates, verify feature readiness:

1. `claude agents` succeeds (custom agents discoverable)
2. Teammate profiles for write tasks declare `isolation: worktree`
3. Long-running teammates prefer `background: true`
4. Hooks include worktree lifecycle coverage (`WorktreeCreate`, `WorktreeRemove`) and config auditing (`ConfigChange`) where policy requires it

For canonical feature details, read:
`skills/shared/references/claude-code-latest-features.md`.

---

## Setup: Create Team

Every spawn session starts by creating a team. One team per wave (fresh context = Ralph Wiggum preserved; see `skills/shared/references/ralph-loop-contract.md`).

```
TeamCreate(team_name="council-20260217-auth", description="Council validation of auth module")
```

```
TeamCreate(team_name="swarm-1739812345-w1", description="Wave 1: parallel implementation")
```

**Naming conventions:**
- Council: `council-YYYYMMDD-<target>`
- Swarm: `swarm-<epoch>-w<wave>`
- Crank: delegates to swarm naming

## Leader Contract (Native Teams)

Claude teams are leader-first orchestration:

1. One lead creates the team and assigns all work.
2. Teammates never self-assign from shared tasks.
3. Teammates report to lead via short `SendMessage` signals.
4. Lead reads result artifacts from disk, validates, and decides retries/escalation.

Recommended signal envelope (single-line JSON, under 100 tokens):

```json
{"type":"completion|blocked|help_request","agent":"worker-3","task":"3","detail":"short status","artifact":".agents/swarm/results/3.json"}
```

`completion`: task finished, artifact written.
`blocked`: cannot proceed safely.
`help_request`: teammate needs coordination or scope clarification.

### Peer Messaging (Allowed, Lead-Controlled)

Native teams support direct teammate-to-teammate messaging. Use this only for coordination handoffs; keep messages thin and always copy the lead in follow-up summaries.

```text
worker-2 -> worker-5: "Need auth schema constant name; please confirm from src/auth/schema.ts"
worker-5 -> lead: "Resolved peer question for worker-2; no scope change."
```

---

## Spawn: Create Workers/Judges

After `TeamCreate`, spawn each agent with `Task(team_name=..., name=...)`. All agents in a wave spawn in parallel (single message, multiple tool calls).

### Council Judges (parallel spawn)

```
Task(
  subagent_type="general-purpose",
  team_name="council-20260217-auth",
  name="judge-1",
  prompt="You are judge-1 on team council-20260217-auth.\n\nYour perspective: Correctness & Completeness\n\n<PACKET>\n...\n</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-1.md\nThen send a SHORT completion signal to the team lead (under 100 tokens).\nDo NOT include your full analysis in the message — the lead reads your file.",
  description="Council judge-1"
)

Task(
  subagent_type="general-purpose",
  team_name="council-20260217-auth",
  name="judge-error-paths",
  prompt="You are judge-error-paths on team council-20260217-auth.\n\nYour perspective: Error Paths & Edge Cases\n\n<PACKET>\n...\n</PACKET>\n\nWrite your verdict to .agents/council/2026-02-17-auth-judge-error-paths.md\nThen send a SHORT completion signal to the team lead (under 100 tokens).",
  description="Council judge-error-paths"
)
```

Both `Task` calls go in the **same message** — they spawn in parallel.

### Swarm Workers (parallel spawn)

```
Task(
  subagent_type="general-purpose",
  team_name="swarm-1739812345-w1",
  name="worker-3",
  prompt="You are worker-3 on team swarm-1739812345-w1.\n\nYour Assignment: Task #3: Add password hashing\n<description>...</description>\n\nInstructions:\n1. Execute your task — create/edit files as needed\n2. Write result to .agents/swarm/results/3.json\n3. Send a SHORT signal to team lead (under 100 tokens)\n4. Do NOT run git add/commit/push — the lead commits\n\nRESULT FORMAT:\n{\"type\":\"completion\",\"issue_id\":\"3\",\"status\":\"done\",\"detail\":\"one-line summary\",\"artifacts\":[\"path/to/file\"]}",
  description="Swarm worker-3"
)

Task(
  subagent_type="general-purpose",
  team_name="swarm-1739812345-w1",
  name="worker-5",
  prompt="You are worker-5 on team swarm-1739812345-w1.\n\nYour Assignment: Task #5: Create login endpoint\n...",
  description="Swarm worker-5"
)
```

### Research Explorers (read-only)

```
Task(
  subagent_type="Explore",
  team_name="research-20260217-auth",
  name="explorer-1",
  prompt="Thoroughly investigate: authentication patterns in this codebase\n\n...",
  description="Research explorer"
)
```

Use `subagent_type="Explore"` for read-only research agents. Use `"general-purpose"` for agents that need to write files.

---

## Wait: Receive Completion Signals

Workers/judges send completion signals via `SendMessage`. These are **automatically delivered** to the team lead — no polling needed.

When a teammate finishes, their message appears as a new conversation turn. The lead reads result files from disk, NOT from message content.

```
# Teammate message arrives automatically:
# "judge-1: Done. Verdict: WARN, confidence: HIGH. File: .agents/council/2026-02-17-auth-judge-1.md"

# Lead reads the file for full details:
Read(".agents/council/2026-02-17-auth-judge-1.md")
```

**Timeout handling (default: 120s per round, 90s for debate R2):**

If a teammate goes idle without sending a completion signal:
1. Check their result file — they may have written it but failed to message
2. If result file exists → read it and proceed (the message was the only thing missing)
3. If no result file → the agent failed silently. **Recovery:** proceed with N-1 judges/workers and note the failure in the report. For swarm workers, add the task back to the retry queue.
4. Never wait indefinitely — after the timeout, move on

See `skills/council/references/cli-spawning.md` for timeout configuration (`COUNCIL_TIMEOUT`, `COUNCIL_R2_TIMEOUT`).

**Fallback:** If native teams fail at runtime despite passing detection (e.g., `TeamCreate` succeeds but `Task` spawning fails), fall back to background tasks. See `backend-background-tasks.md`.

---

## Message: Debate R2 / Retry

Send messages to specific teammates using `SendMessage`. Teammates wake from idle when messaged.

### Council Debate R2

```
SendMessage(
  type="message",
  recipient="judge-1",
  content="DEBATE ROUND 2\n\nOther judges' verdicts:\n- judge-error-paths: FAIL (HIGH confidence) — file: .agents/council/2026-02-17-auth-judge-error-paths.md\n\nRead the other judge's file. Revise your assessment considering their perspective.\nWrite your R2 verdict to .agents/council/2026-02-17-auth-judge-1-r2.md\nThen send a completion signal.",
  summary="R2 debate instructions for judge-1"
)
```

**R2 timeout (default: 90s):** If a judge doesn't respond to R2 within `COUNCIL_R2_TIMEOUT`, use their R1 verdict for consolidation. See `skills/council/references/debate-protocol.md` for full timeout handling.

### Swarm Worker Retry

```
SendMessage(
  type="message",
  recipient="worker-3",
  content="Validation failed: pytest tests/test_auth.py returned exit code 1.\nFix the failing tests and rewrite your result to .agents/swarm/results/3.json",
  summary="Retry worker-3: test failure"
)
```

---

## Cleanup: Shutdown and Delete

After consolidation/validation, shut down all teammates then delete the team.

```
# Shutdown each teammate
SendMessage(type="shutdown_request", recipient="judge-1", content="Council complete")
SendMessage(type="shutdown_request", recipient="judge-error-paths", content="Council complete")

# After all teammates acknowledge shutdown:
TeamDelete()
```

**Reaper pattern:** If a teammate doesn't respond to shutdown within 30s, proceed with `TeamDelete()` anyway.

**If `TeamDelete` fails** (e.g., stale members): clean up manually with `rm -rf ~/.codex/teams/<team-name>/` then retry `TeamDelete()` to clear in-memory state.

---

## Multi-Wave Pattern

For crank/swarm with multiple waves, create a **new team per wave**:

```
# Wave 1
TeamCreate(team_name="swarm-1739812345-w1", description="Wave 1")
# ... spawn workers, wait, validate, commit ...
# ... shutdown teammates ...
TeamDelete()
# If TeamDelete fails: rm -rf ~/.codex/teams/swarm-1739812345-w1/ then retry

# Wave 2 (fresh context)
TeamCreate(team_name="swarm-1739812345-w2", description="Wave 2")
# ... spawn workers for newly-unblocked tasks ...
TeamDelete()
```

This ensures each wave's workers start with clean context (no leftover state from prior waves).

**If `TeamDelete` fails between waves**, the next `TeamCreate` may conflict. Always verify cleanup succeeded before creating the next wave team.

---

## Key Rules

1. **`TeamCreate` before `Task`** — tasks created before the team are invisible to teammates
2. **Pre-assign tasks before spawning** — workers do NOT race-claim from TaskList
3. **Lead-only commits** — workers write files, lead runs `git add` + `git commit`
4. **Thin messages** — workers send <100 token signals, full results go to disk
5. **New team per wave** — fresh context, Ralph Wiggum preserved
6. **Always cleanup** — `TeamDelete()` after every wave, even on partial failure

### backend-codex-subagents.md

# Backend: Codex Sub-Agents

Concrete tool calls for spawning agents using Codex CLI (`codex exec`). Used for `--mixed` mode cross-vendor consensus and as the primary backend when running inside a Codex session with `spawn_agent`.

---

## Variant A: Codex CLI (from any runtime)

Used when `codex` CLI is available on PATH. Agents run as background shell processes.

**When detected:** `which codex` succeeds.

### Spawn: Background Shell Processes

```bash
# With structured output (preferred for council judges)
Bash(
  command='codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-1.json "JUDGE PROMPT HERE"',
  run_in_background=true
)

# Without structured output (fallback)
Bash(
  command='codex exec --full-auto -m gpt-5.3-codex -C "$(pwd)" -o .agents/council/codex-1.md "JUDGE PROMPT HERE"',
  run_in_background=true
)
```

**Flag order:** `-s`/`--full-auto` → `-m` → `-C` → `--output-schema` → `-o` → prompt

**Valid flags:** `--full-auto`, `-s`, `-m`, `-C`, `--output-schema`, `-o`, `--add-dir`
**Invalid flags:** `-q` (doesn't exist), `--quiet` (doesn't exist)

### Wait: Poll Background Shell

```
TaskOutput(task_id="<shell-id>", block=true, timeout=120000)
```

Then read the output file:

```
Read(".agents/council/codex-1.json")
```

### Limitations

- No messaging — Codex CLI processes are fire-and-forget
- No debate R2 with Codex judges — they produce one verdict only
- `--output-schema` requires `additionalProperties: false` at all levels
- `--output-schema` requires ALL properties in `required` array
- `-s read-only` + `-o` works — `-o` is CLI-level post-processing, not sandbox I/O

---

## Variant B: Codex Sub-Agents (inside Codex runtime)

Used when running inside a Codex session where `spawn_agent` is available.

**When detected:** `spawn_agent` tool is in your tool list.

### Spawn

```
spawn_agent(message="You are judge-1.\n\nPerspective: Correctness & Completeness\n\n<PACKET>...</PACKET>\n\nWrite verdict to .agents/council/2026-02-17-auth-judge-1.md")
# Returns: agent_id

spawn_agent(message="You are worker-3.\n\nTask: Add password hashing\n...\n\nWrite result to .agents/swarm/results/3.json")
# Returns: agent_id
```

### Wait

```
wait(ids=["agent-id-1", "agent-id-2"])
```

**Timeout:** `wait()` blocks until completion. Set a timeout at the orchestration level (default: `COUNCIL_TIMEOUT=120s`). If an agent doesn't complete within the timeout, `close_agent` it and proceed with N-1 verdicts/workers.

### Message (retry/follow-up)

```
send_input(id="agent-id-1", message="Validation failed: fix tests and retry")
```

### Cleanup

```
close_agent(id="agent-id-1")
```

---

## Mixed Mode (Council)

For `--mixed` council, spawn runtime-native judges AND Codex CLI judges in parallel:

```
# Claude native team judges (via TeamCreate — see backend-claude-teams.md)
Task(subagent_type="general-purpose", team_name="council-20260217-auth", name="judge-1", prompt="...", description="Judge 1")
Task(subagent_type="general-purpose", team_name="council-20260217-auth", name="judge-2", prompt="...", description="Judge 2")

# Codex CLI judges (parallel background shells)
Bash(command='codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-1.json "PACKET"', run_in_background=true)
Bash(command='codex exec -s read-only -m gpt-5.3-codex -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-2.json "PACKET"', run_in_background=true)
```

All four spawn in the **same message** — maximum parallelism.

**Mixed mode quorum:** At least 1 judge from each vendor should respond for cross-vendor consensus. If all judges from one vendor fail, proceed as single-vendor council and note the degradation in the report.

---

## Key Rules

1. **Pre-flight check:** `which codex` before attempting Codex CLI spawning
2. **Model availability:** `gpt-5.3-codex` requires API account — fall back to `gpt-4o` if unavailable
3. **Flag order matters** — agents copy examples exactly
4. **`codex review` is a different command** with different flags — do not conflate with `codex exec`
5. **No debate with Codex judges** — they produce one verdict, Codex CLI has no messaging

### backend-inline.md

# Backend: Inline (No Spawn Available)

Degraded single-agent mode when no multi-agent primitives are detected. The current agent performs all work sequentially in its own context.

**When detected:** No `spawn_agent`, no `TeamCreate`, no `Task` tool available — or `--quick` flag was explicitly set.

---

## Council: Single Inline Judge

Instead of spawning parallel judges, the lead evaluates from each perspective sequentially:

```
1. Build the context packet (same as multi-agent mode)
2. For each perspective:
   a. Adopt the perspective mentally
   b. Write findings to .agents/council/YYYY-MM-DD-<target>-<perspective>.md
3. Synthesize into final report
```

Output format is identical — same file paths, same verdict schema. Downstream consumers (consolidation, report) don't know it was inline.

**No debate available** — debate requires messaging between agents.

---

## Swarm: Sequential Execution

Instead of parallel workers, execute each task sequentially:

```
1. TaskList() — find unblocked tasks
2. For each unblocked task (in order):
   a. Execute the task directly
   b. Write result to .agents/swarm/results/<task-id>.json
   c. TaskUpdate(taskId="<id>", status="completed")
3. Check for newly-unblocked tasks
4. Repeat until all tasks complete
```

Same result files, same validation — just sequential.

**Error handling:** If a task fails mid-execution:
1. Write failure result to `.agents/swarm/results/<task-id>.json` with `"status": "blocked"`
2. Check if downstream tasks depend on it (`blockedBy`)
3. Skip blocked downstream tasks, mark as skipped
4. Continue with independent tasks that don't depend on the failed one

---

## Research: Inline Exploration

Instead of spawning an Explore agent, perform the tiered search directly:

```
1. Read docs/code-map/ if present
2. Grep/Glob for relevant files
3. Read key files
4. Write findings to .agents/research/YYYY-MM-DD-<topic>.md
```

---

## Key Rules

1. **Same output format** — inline mode writes the same files as multi-agent mode
2. **Same validation** — all checks still apply
3. **Slower but functional** — no parallelism, but all skill capabilities preserved (except debate)
4. **Inform the user** — log "Running in inline mode (no multi-agent backend detected)"

### claude-code-latest-features.md

# Codex Latest Features Contract

This document is the shared source of truth for Codex feature usage across AgentOps skills.

## Baseline

- Target Codex release family: `2.1.x`
- Last verified against upstream changelog: `2.1.50`
- Changelog source: `https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md`

## Current Feature Set We Rely On

### 1. Core Slash Commands

Skills and docs should assume these commands exist and prefer them over legacy naming:

- `/agents`
- `/hooks`
- `/permissions`
- `/memory`
- `/mcp`
- `/output-style`

Reference: `https://code.claude.com/docs/en/slash-commands`

### 2. Agent Definitions

For custom teammates in `.claude/agents/*.md`, use modern frontmatter fields where applicable:

- `model`
- `description`
- `tools`
- `memory` (scope control)
- `background: true` for long-running teammates
- `isolation: worktree` for safe parallel write isolation

Reference: `https://code.claude.com/docs/en/sub-agents`

### 3. Worktree Isolation

When parallel workers may touch overlapping files, prefer Claude-native isolation features first:

- Session-level isolation: `claude --worktree` (`-w`)
- Agent-level isolation: `isolation: worktree`

If unavailable in a given runtime, fall back to manual `git worktree` orchestration.

Reference: changelog `2.1.49` and `2.1.50`.

### 4. Hooks and Governance Events

Hooks-based workflows should include modern event coverage:

- `WorktreeCreate`
- `WorktreeRemove`
- `ConfigChange`
- `SubagentStop`
- `TaskCompleted`
- `TeammateIdle`

Use these for auditability, policy enforcement, and cleanup.

Reference: hooks docs and changelog.

### 5. Settings Hierarchy

Skill guidance must respect settings precedence:

1. Enterprise managed policy
2. Command-line args
3. Local project settings
4. Shared project settings
5. User settings

Reference: `https://code.claude.com/docs/en/settings`

### 6. Agent Inventory Command

Use `claude agents` as the first CLI-level check to confirm configured teammate profiles before multi-agent runs.

Reference: changelog `2.1.50`.

## Skill Authoring Rules

1. Do not reference deprecated permission command names (`/allowed-tools`, `/approved-tools`).
2. Multi-agent skills (`council`, `swarm`, `research`, `crank`, `codex-team`) must explicitly point to this contract.
3. Prefer declarative agent isolation (`isolation: worktree`) over ad hoc branch/worktree shell choreography where runtime supports it.
4. Keep manual `git worktree` fallback documented for non-Claude runtimes.
5. For long-running explorers/judges/workers, document `background: true` as the default custom-agent policy.

## Review Cadence

- Re-verify this contract when:
  - Codex changelog introduces new `2.1.x` or `2.2.x` entries
  - any skill adds or changes multi-agent orchestration
  - hook event support changes

### context-discovery.md

# Context Discovery Tiers

**Purpose**: Systematic approach to finding code/context before implementing.

**Rule**: Work top-to-bottom. Skip tiers if source unavailable.

---

## Tier Order

| Tier | Source | Tool/Command | When to Skip |
|------|--------|--------------|--------------|
| **1** | Code-Map | `Read docs/code-map/README.md` | No code-map in repo |
| **2** | Semantic Search | `mcp__smart-connections-work__lookup` | MCP not connected |
| **3** | Scoped Search | `Grep/Glob` with path limits | - |
| **4** | Source Code | `Read` files from Tier 1-3 signposts | - |
| **5** | Prior Knowledge | `ls .agents/research/` | Verify against source |
| **6** | External Docs | Context7, WebSearch | Last resort |

---

## Tier Details

### Tier 1: Code-Map (Fastest)

```bash
Read docs/code-map/README.md   # Find category
Read docs/code-map/{feature}.md  # Get signposts
```

**Why first**: Local, instant, gives exact paths and function names.

### Tier 2: Semantic Search

```bash
mcp__smart-connections-work__lookup --query="$TOPIC" --limit=10
```

**Why second**: Finds conceptual matches code-map might miss. Requires MCP.

### Tier 3: Scoped Search

```bash
Grep("pattern", path="services/auth/")   # SCOPED
Glob("services/etl/**/*.py")             # SCOPED
```

**Never**: `Grep("pattern")` or `Glob("**/*.py")` on large repos.

### Tier 4: Source Code

Read files identified by Tiers 1-3. Use function/class names, not line numbers.

### Tier 5: Prior Knowledge

```bash
ls .agents/research/ | grep -i "$TOPIC"
```

**Caution**: May be stale. Always verify findings against current source.

### Tier 6: External

- **Context7**: Library documentation
- **WebSearch**: External APIs, standards

---

## Quick Reference

```
Code-Map → Semantic → Grep/Glob → Source → .agents/ → External
   ↓           ↓          ↓          ↓         ↓          ↓
 paths      meaning    keywords    code     history    docs
```

---

## Tier Weights (Flywheel-Optimized)

Default weights based on typical value. Adjust based on `GET /memories/analytics/sources`:

| Tier | Source Type | Default Weight | Notes |
|------|-------------|----------------|-------|
| 1 | `code-map` | 1.0 | Local, authoritative |
| 2 | `smart-connections` | 0.95 | High semantic match |
| 3 | `grep`, `glob` | 0.85 | Keyword precision |
| 4 | `read` | 0.80 | Direct source |
| 5 | `prior-research`, `memory-recall` | 0.70 | May be stale |
| 6 | `web-search`, `web-fetch` | 0.60 | External, verify |

**Optimization loop**:
```bash
# Query source analytics
curl -H "X-API-Key: $KEY" "$ETL_URL/memories/analytics/sources?collection=default"

# Response includes per-source value_score metrics:
# {
#   "sources": [
#     {"source_type": "smart-connections", "value_score": 0.72},
#     {"source_type": "grep", "value_score": 0.61},
#     ...
#   ],
#   "recommendations": [...]
# }

# Adjust weights based on value_score:
# value_score = (total_citations / memory_count) × avg_confidence × recency_factor
#
# - value_score > 0.5: Move source up in priority (increase weight)
# - value_score 0.3-0.5: Maintain current position
# - value_score < 0.3: Consider deprioritizing
# - value_score < 0.1 with high count: Review quality - many memories but rarely cited
```

**Tool to source_type mapping** (for session analyzer):
```python
WebSearch → "web-search"
WebFetch → "web-fetch"
mcp__smart-connections-work__lookup → "smart-connections"
mcp__smart-connections-personal__lookup → "smart-connections"
mcp__ai-platform__search_knowledge → "athena-knowledge"
mcp__ai-platform__memory_recall → "memory-recall"
Grep → "grep"
Glob → "glob"
Read → "read"
LSP → "lsp"
```

---

## Failure Pattern Prevention

Each tier helps prevent specific failure patterns from the Vibe-Coding methodology:

| Tier | Prevents Pattern | How |
|------|------------------|-----|
| 1 (Code-Map) | #9 Cargo Cult | Authoritative docs explain WHY patterns exist |
| 2 (Semantic) | #7 Zombie Resurrection | Finds prior art you might miss |
| 3 (Scoped Search) | #3 Context Amnesia | Scoping prevents context overload |
| 4 (Source Code) | #2 Confident Hallucination | Verify claims against actual code |
| 5 (Prior Knowledge) | #7 Zombie Resurrection | Don't re-solve solved problems |
| 6 (External) | #11 Security Theater | External standards for security |

### The 40% Context Rule

**Critical:** Never exceed 40% context utilization during discovery.

| Zone | Percentage | Action |
|------|-----------|--------|
| GREEN | <35% | Continue exploration |
| YELLOW | 35-40% | Summarize, prepare to output |
| RED | >40% | STOP. Write findings. Reset. |

**Why:** Above 40%, Pattern #3 (Context Amnesia) kicks in. Quality degrades exponentially.

### Defensive Epistemology

For each tier exploration, apply explicit reasoning:

```text
DOING: [search/read action]
EXPECT: [what I expect to find]
IF WRONG: [what I'll conclude]
```

After:

```text
RESULT: [what happened]
MATCHES: [yes/no]
THEREFORE: [conclusion]
```

This prevents Pattern #2 (Confident Hallucination) by forcing verification.

---

## Anti-Patterns

| DON'T | DO INSTEAD | Prevents Pattern |
|-------|------------|------------------|
| Start with Grep on full repo | Start with code-map | #3 Amnesia |
| Read source before knowing where | Find signposts first | #3 Amnesia |
| Trust .agents/ without verifying | Cross-check against source | #12 Doc Mirage |
| Web search for internal code | Use Tiers 1-4 | #9 Cargo Cult |
| Unscoped Glob/Grep | Always specify path | #3 Amnesia |
| "This API should work..." | Verify against actual docs | #2 Hallucination |
| "This code looks unused..." | Trace refs, check history | #6 Silent Deletion |
| Read entire large file | Targeted offset/limit | #3 Amnesia |

### document-template.md

# Research Document Template

## Filename Format

`.agents/research/YYYY-MM-DD-{topic-slug}.md`

Convert topic to kebab-case slug:
- "authentication flow" -> `2026-01-03-authentication-flow.md`
- "MCP server architecture" -> `2026-01-03-mcp-server-architecture.md`

---

## Required Sections

### 1. Frontmatter

```yaml
---
date: YYYY-MM-DD
type: Research
topic: "Topic Name"
tags: [research, domain, tech]
status: COMPLETE
supersedes: []
---
```

### 2. Executive Summary

2-3 sentences: what found, what recommend.

### 3. Current State

- What exists today
- Key files table: | File | Purpose |
- Existing patterns

### 4. Findings

Each finding with:
- Evidence: `file:line`
- Implications

### 5. Constraints

| Constraint | Impact | Mitigation |
|------------|--------|------------|

### 6. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|

### 7. Recommendation

- Recommended approach
- Rationale
- Alternatives considered and rejected

### 8. Discovery Provenance

Track which sources provided key insights (enables flywheel optimization).

**Purpose**: Create an audit trail showing which discovery method found each insight. This enables post-hoc analysis: "Which sources led to successful implementation?"

**When to complete**: As you research, add one row per significant finding showing its source.

**Example**:
```markdown
| Finding | Source Type | Source Detail | Confidence |
|---------|-------------|---------------|------------|
| Gateway request flow | code-map | docs/code-map/gateway.md | 1.0 |
| Middleware pattern | smart-connections | "request middleware chain" | 0.95 |
| Error handling at L45 | grep | services/gateway/middleware.py | 1.0 |
| Rate limiting precedent | prior-research | 2026-01-10-ratelimit.md | 0.85 |
| OAuth2 RFC | web-search | "RFC 6749 OAuth 2.0" | 0.80 |
```

**Source Types by Tier** (higher tier = better quality):

**Tier 1 (Authoritative)**
- `code-map` - Structured architecture documentation (highest confidence)

**Tier 2 (Semantic)**
- `smart-connections` - Obsidian semantic search
- `athena-knowledge` - MCP ai-platform search

**Tier 3 (Scoped Search)**
- `grep` - Pattern matching in code
- `glob` - File pattern matching

**Tier 4 (Source Code)**
- `read` - Direct file reading
- `lsp` - Language Server Protocol queries

**Tier 5 (Prior Art)**
- `prior-research` - Previous research documents
- `prior-retro` - Retrospective learnings
- `prior-pattern` - Reusable patterns
- `memory-recall` - Semantic memory search

**Tier 6 (External)**
- `web-search` - Web search results
- `web-fetch` - Direct URL fetch

**Other**
- `conversation` - User-provided context

**Confidence scoring**:
- `1.0` - Source is authoritative/written down
- `0.95` - Semantic match, high relevance
- `0.85` - Good match, may need verification
- `0.70` - Reasonable match, verify
- < 0.70 - Use sparingly, needs verification

### 9. Failure Pattern Risks

Identify which of the 12 failure patterns are risks for this work. This proactive assessment helps downstream implementation avoid known pitfalls.

**Required table:**
```markdown
## Failure Pattern Risks

| Pattern | Risk Level | Mitigation |
|---------|------------|------------|
| #N Pattern Name | HIGH/MEDIUM/LOW | Specific mitigation strategy |
```

**Pattern quick reference:**

| # | Pattern | Common Research Triggers |
|---|---------|-------------------------|
| 1 | Fix Spiral | Complex debugging, unclear root cause |
| 2 | Confident Hallucination | External APIs, unfamiliar libraries |
| 3 | Context Amnesia | Large codebase, many files to read |
| 4 | Tests Passing Lie | Weak test coverage, mocked dependencies |
| 5 | Eldritch Horror | Complex existing code, deep nesting |
| 6 | Silent Deletion | "Unused" code, cleanup opportunities |
| 7 | Zombie Resurrection | Prior failed attempts, known bugs |
| 8 | Gold Plating | Feature creep opportunities |
| 9 | Cargo Cult | New patterns, external examples |
| 10 | Premature Abstraction | Generic solutions proposed |
| 11 | Security Theater | Auth, crypto, access control |
| 12 | Documentation Mirage | Outdated docs, missing comments |

**Example:**
```markdown
## Failure Pattern Risks

| Pattern | Risk Level | Mitigation |
|---------|------------|------------|
| #2 Confident Hallucination | HIGH | External OAuth API - verify all claims against official docs |
| #5 Eldritch Horror | MEDIUM | Auth middleware is 400+ lines - document boundaries before changes |
| #9 Cargo Cult | MEDIUM | Using external OAuth example - understand why each step exists |
| #11 Security Theater | HIGH | Auth changes - use established patterns, get security review |
```

### 10. Next Steps

Point to `$plan` for implementation.

---

## Tag Vocabulary

**Rules:** 3-5 tags total. First tag MUST be `research`.

| Category | Valid Tags |
|----------|------------|
| **Core Domains** | `agents`, `data`, `api`, `infra`, `security`, `auth` |
| **Quality** | `testing`, `reliability`, `performance`, `monitoring` |
| **Process** | `ci-cd`, `workflow`, `ops`, `docs` |
| **Governance** | `architecture`, `compliance`, `standards`, `ui` |
| **Languages** | `python`, `shell`, `typescript`, `go`, `yaml` |
| **Platforms** | `helm`, `kubernetes`, `openshift`, `docker`, `argocd` |
| **AI Stack** | `mcp`, `litellm`, `neo4j`, `postgres`, `redis`, `fastapi` |

**Examples:**
- `[research, agents, mcp]` - MCP server research
- `[research, data, neo4j]` - Data storage research
- `[research, security, auth]` - Authentication research

---

## Status Values

| Status | Meaning |
|--------|---------|
| `COMPLETE` | Ready for planning |
| `IN_PROGRESS` | Ongoing research |
| `SUPERSEDED` | Newer research exists |

### failure-patterns.md

# The 12 Failure Patterns (Research Reference)

> Based on the Vibe-Coding methodology. Load this when you need full pattern details for risk assessment.

---

## Quick Reference

| # | Pattern | Key Symptom | First Action |
|---|---------|-------------|--------------|
| 1 | Fix Spiral | >3 attempts, circles | STOP, revert |
| 2 | Confident Hallucination | Non-existent APIs | Verify docs |
| 3 | Context Amnesia | Forgotten constraints | Save state |
| 4 | Tests Passing Lie | Green but broken | Manual test |
| 5 | Eldritch Horror | >200 line functions | Extract/refactor |
| 6 | Silent Deletion | Missing code | Check git history |
| 7 | Zombie Resurrection | Bugs return | Add regression test |
| 8 | Gold Plating | Unrequested features | Revert extras |
| 9 | Cargo Cult | Copied patterns | Understand why |
| 10 | Premature Abstraction | Generic w/ one use | Inline |
| 11 | Security Theater | Bypassable security | Audit |
| 12 | Documentation Mirage | Docs don't work | Test docs |

---

## Inner Loop Patterns (Seconds-Minutes)

### 1. The Fix Spiral

**Description:** Making a fix that breaks something else, then fixing that break which causes another issue, creating a cascading chain without resolution.

**Symptoms:**
- More than 3 fix attempts without convergence
- Changes oscillating between two states
- "This should work" appearing in explanations
- Error messages changing but not disappearing

**Research Defense:**
- Research root cause BEFORE attempting fix
- Document expected behavior vs actual behavior
- Identify all code paths affected

**Prevention:**
- Set hard limit: 3 attempts then STOP
- State explicit prediction before each fix
- Checkpoint working state before each attempt

---

### 2. The Confident Hallucination

**Description:** Generating plausible-sounding but factually incorrect information about APIs, libraries, or behavior.

**Symptoms:**
- Code references non-existent methods or parameters
- API usage that "looks right" but fails at runtime
- Overly specific technical claims without evidence
- Version-specific features applied to wrong versions

**Research Defense:**
- VERIFY all API claims against actual documentation
- Note confidence levels in provenance table
- Use Tier 6 (external docs) for unfamiliar APIs

**Prevention:**
- Test code in isolation before integration
- Use "I don't know" as valid response
- Run type checkers and linters early

---

### 3. The Context Amnesia

**Description:** As context window fills, losing track of earlier constraints, requirements, or decisions.

**Symptoms:**
- Reintroducing previously fixed bugs
- Contradicting earlier decisions
- Forgetting project-specific conventions
- Repeating completed work

**Research Defense:**
- Stay <40% context utilization
- Write findings to files immediately
- Use targeted reads (offset/limit) not full files

**Prevention:**
- Save progress frequently
- Start fresh sessions for distinct work
- Front-load critical constraints

---

### 4. The Tests Passing Lie

**Description:** Tests pass but code doesn't actually work - too narrow, wrong thing, mocks away behavior.

**Symptoms:**
- Green test suite but broken functionality
- Tests that test mocks instead of real behavior
- Coverage looks good but edge cases fail
- Tests modified in same PR as code they test

**Research Defense:**
- Find actual test coverage in research
- Identify what tests actually verify
- Note mocked vs real dependencies

**Prevention:**
- Run tests yourself; don't trust reported results
- Separate test changes from code changes
- Manual smoke test after suite passes

---

## Middle Loop Patterns (Hours-Days)

### 5. The Eldritch Horror

**Description:** Code becomes incomprehensible - functions spanning hundreds of lines, deeply nested logic, unclear naming.

**Symptoms:**
- Functions exceeding 200 lines
- Nesting depth beyond 4 levels
- Variable names like `temp2`, `data3`
- Comments that don't match behavior

**Research Defense:**
- Document complexity limits in findings
- Note current complexity metrics
- Identify refactoring boundaries

**Prevention:**
- Enforce hard limits: <200 lines per function
- Require meaningful names
- Use explicit interfaces

---

### 6. The Silent Deletion

**Description:** Removing code that appears unused but is actually necessary for edge cases, legacy support, or fallbacks.

**Symptoms:**
- "Cleanup" commits that remove "dead code"
- Features that worked yesterday now fail
- Error handling mysteriously missing
- Comments about "why" deleted along with code

**Research Defense:**
- Research WHY code exists before removal
- Check git history for context
- Trace all references including dynamic calls

**Prevention:**
- Never delete without understanding purpose
- Get human approval for deletion
- Keep deleted code in comments initially

---

### 7. The Zombie Resurrection

**Description:** Previously fixed bugs return because similar code regenerated without fix, or reverts during refactoring.

**Symptoms:**
- Bug reports for issues marked "fixed"
- Same error in different code paths
- Fixes lost during refactoring
- "I thought we fixed this" conversations

**Research Defense:**
- Prior art search prevents re-solving
- Check for existing regression tests
- Document root cause, not just fix

**Prevention:**
- Add regression tests for every fix
- Use automated checks for anti-patterns
- Keep lessons learned file

---

### 8. The Gold Plating

**Description:** Adding unrequested features, extra error handling, additional configurability beyond what was asked.

**Symptoms:**
- PR larger than expected
- New config options no one asked for
- "While I was here, I also..." explanations
- Abstraction layers for single use cases

**Research Defense:**
- Define explicit scope in research
- Note ONLY what's needed for the task
- Separate "nice to have" from "required"

**Prevention:**
- Define explicit scope before starting
- Reject changes outside stated scope
- Prefer boring, obvious solutions

---

## Outer Loop Patterns (Days-Weeks)

### 9. The Cargo Cult

**Description:** Copying patterns from examples without understanding why they work. May be inappropriate for context.

**Symptoms:**
- Copy-pasted code with irrelevant portions
- Patterns from different frameworks mixed
- "Best practices" where they don't fit
- Configuration copied without understanding

**Research Defense:**
- Understand WHY patterns exist
- Ask "why does this pattern exist?" for each
- Verify example matches your context

**Prevention:**
- Test copied code in isolation first
- Adapt patterns to local conventions
- Trace examples to their source

---

### 10. The Premature Abstraction

**Description:** Creating generic abstractions before concrete use cases exist. Abstractions don't match actual needs.

**Symptoms:**
- Generic interfaces with one implementation
- Factory patterns for single classes
- Configuration for cases that don't exist
- "Future-proofing" never used

**Research Defense:**
- Document concrete use cases first
- Require 3+ concrete cases before abstracting
- Note where duplication exists vs speculation

**Prevention:**
- Write concrete implementations first
- Prefer duplication over wrong abstraction
- Extract only when duplication appears

---

### 11. The Security Theater

**Description:** Code appears secure but isn't - validation that misses edge cases, encryption with hardcoded keys.

**Symptoms:**
- Security measures easily circumvented
- Validation on client but not server
- Hardcoded credentials or keys
- "Security by obscurity" approaches

**Research Defense:**
- Include security constraints in research
- Reference external security standards
- Note auth/crypto/access control patterns

**Prevention:**
- Use established security libraries
- Security review by qualified humans
- Static analysis for vulnerabilities

---

### 12. The Documentation Mirage

**Description:** Documentation exists but doesn't match reality - outdated READMEs, incorrect API docs.

**Symptoms:**
- Following docs leads to errors
- Comments contradict adjacent code
- Examples that don't compile
- Setup instructions that don't work

**Research Defense:**
- Verify docs match reality
- Test documentation by following it literally
- Note discrepancies in research findings

**Prevention:**
- Treat docs as code: test them
- Update docs in same PR as code
- Use executable documentation

---

## Pattern Frequency Tracking

Use this in research outputs to track which patterns are relevant:

```markdown
## Failure Pattern Risks

| Pattern | Risk Level | Mitigation |
|---------|------------|------------|
| #2 Confident Hallucination | HIGH | Verify external API claims |
| #5 Eldritch Horror | MEDIUM | Keep functions <200 lines |
| #9 Cargo Cult | MEDIUM | Understand why patterns exist |
```

Risk Levels:
- **HIGH**: Strong indicators in research, requires explicit mitigation
- **MEDIUM**: Some indicators, requires awareness
- **LOW**: Minor indicators, standard practices sufficient

---

## See Also

- `~/.codex/CLAUDE-base.md` - Core Vibe-Coding methodology
- `~/.codex/plugins/marketplaces/agentops-marketplace/reference/failure-patterns.md` - Full pattern reference
- `~/.codex/skills/crank/failure-taxonomy.md` - Execution failure taxonomy

### ralph-loop-contract.md

# Ralph Loop Contract (Reverse-Engineered)

This contract captures the operational Ralph mechanics reverse-engineered from:
- `https://github.com/ghuntley/how-to-ralph-wiggum`
- `.tmp/how-to-ralph-wiggum/README.md`
- `.tmp/how-to-ralph-wiggum/files/loop.sh`
- `.tmp/how-to-ralph-wiggum/files/PROMPT_plan.md`
- `.tmp/how-to-ralph-wiggum/files/PROMPT_build.md`

Use this as the source-of-truth for Ralph alignment in AgentOps orchestration skills.

## Core Contract

1. Fresh context every iteration/wave.
- Each execution unit starts clean; no carryover worker memory.

2. Scheduler-heavy, worker-light.
- The lead/orchestrator schedules and reconciles.
- Workers perform one scoped unit of work.

3. Disk-backed shared state.
- Loop continuity comes from filesystem state, not accumulated chat context.
- In classic Ralph: `IMPLEMENTATION_PLAN.md` and `AGENTS.md`.

4. One-task atomicity.
- Select one important task, execute, validate, persist state, then restart fresh.

5. Backpressure before completion.
- Build/tests/lint/gates must reject bad output before task completion/commit.

6. Observe and tune outside the loop.
- Humans (or lead agents) monitor outcomes and adjust prompts/constraints/contracts.

## AgentOps Mapping

| Ralph concept | AgentOps implementation |
|---|---|
| Fresh context per loop | New workers/teams per wave in `$swarm`; fresh phase context in `ao work rpi phased` |
| Main context as scheduler | Mayor/lead orchestration in `$swarm` and `$crank` |
| Plan file as state | `bd` issue graph, TaskList state, plan artifacts in `.agents/plans/` |
| One task per pass | One issue per worker assignment in swarm/crank waves |
| Backpressure | `$vibe`, task validation hooks, tests/lint gates, push/pre-mortem gates |
| Outer loop restart | Wave loop in `$crank`; phase loop in `ao work rpi phased` |

## Implementation Notes

- Keep worker prompts concise and operational.
- Keep state in files/issue trackers, not long conversational memory.
- Prefer deterministic checks over subjective completion.

### vibe-methodology.md

# Vibe Methodology

Core principles for AI-assisted development. "Vibe" = trust-but-verify.

---

## The 40% Rule

**Never exceed 40% context utilization.**

- Checkpoint at 35%
- Reset via session restart or `$research` artifact
- More context ≠ better results (hallucination risk increases)

---

## Three Levels of Verification

| Level | Vibe | Method | When |
|-------|------|--------|------|
| L1 | Accept | Structural check only | Boilerplate, formatting |
| L2 | Probe | Spot-check key logic | Normal implementation |
| L3 | Audit | Line-by-line review | Security, data handling |

**Default to L2.** Upgrade to L3 for:
- Authentication/authorization
- Financial calculations
- Data persistence
- External API calls

---

## Evidence Hierarchy

Trust in order:

1. **Running code** - Actually execute it
2. **Tests** - Passing tests prove behavior
3. **File contents** - Read the actual source
4. **Documentation** - May be stale
5. **Model claims** - Verify everything

---

## Working Patterns

### Incremental Verification
```
Write small piece → Test → Verify → Repeat
```

Don't write 500 lines then debug. Write 50, verify, continue.

### Checkpoint Often
- After each feature complete
- Before any risky change
- At natural boundaries

### Search Before Implement
```bash
# Always check for prior art
mcp__smart-connections-work__lookup --query="<topic>"
ls .agents/research/ | grep -i "<topic>"
```

---

## Anti-Patterns to Avoid

| Anti-Pattern | Why Bad | Instead |
|--------------|---------|---------|
| Trust-and-paste | Hallucinations slip through | Always read generated code |
| Context stuffing | Degrades quality | Stay under 40% |
| Fix spiraling | Compounds errors | Reset and rethink |
| Skipping verification | Builds on bad foundation | Verify incrementally |

---

## The Research Discipline

1. **Scope first** - Define what you're looking for
2. **Search smart** - Use semantic search before grep
3. **Read selectively** - Don't load whole files
4. **Cite everything** - `file:line` for all claims
5. **Synthesize** - Connect findings to goal

---

## Session Hygiene

```bash
# Start
gt hook              # Check assigned work
bd ready             # What's available

# Work
$research <topic>    # Creates artifact, saves context
$implement <issue>   # Focused execution

# End
bd sync              # Sync beads
git commit           # Commit changes
git push             # WORK IS NOT DONE UNTIL PUSHED
```

---

## References

- `failure-patterns.md` - 12 specific failure modes
- `context-discovery.md` - 6-tier exploration hierarchy
- `~/.codex/CLAUDE-base.md` - Full vibe methodology


---

## Scripts

### validate.md

```md
# Validation Script for Research Skill

## Overview

The `validate.sh` script ensures the `$research` skill meets basic quality and completeness standards. It runs a series of checks against the skill's structure, documentation, and references.

## Purpose

This validation script serves as a quality gate for the research skill, ensuring:

- Required files exist with correct structure
- Documentation includes essential patterns and concepts
- References directory contains sufficient resource materials

## Script Location

```
skills/research/scripts/validate.sh
```

## Script Execution

The script performs the following checks:

### Basic Structure Validation
- **SKILL.md exists**: Verifies the primary skill documentation file
- **SKILL.md has YAML frontmatter**: Ensures proper metadata formatting
- **name: research**: Confirms correct skill identification
- **references/ directory exists**: Validates reference materials directory
- **references/ has at least 3 files**: Ensures minimum reference coverage

### Documentation Content Validation
- **SKILL.md mentions .agents/research/ output path**: Confirms documented output location
- **SKILL.md mentions Explore agent**: Ensures agent reference is included
- **SKILL.md mentions --auto flag**: Validates feature documentation
- **SKILL.md mentions ao know inject**: Checks CLI integration documentation
- **SKILL.md mentions knowledge flywheel**: Confirms system architecture coverage
- **SKILL.md mentions backend detection**: Validates technical implementation details
- **SKILL.md mentions quality validation**: Ensures quality assurance documentation

## Usage

### Manual Execution

```bash
# From the project root directory
./skills/research/scripts/validate.sh
```

### Expected Output

```
PASS: SKILL.md exists
PASS: SKILL.md has YAML frontmatter
PASS: SKILL.md has name: research
PASS: references/ directory exists
PASS: references/ has at least 3 files
PASS: SKILL.md mentions .agents/research/ output path
PASS: SKILL.md mentions Explore agent
PASS: SKILL.md mentions --auto flag
PASS: SKILL.md mentions ao know inject
PASS: SKILL.md mentions knowledge flywheel
PASS: SKILL.md mentions backend detection
PASS: SKILL.md mentions quality validation

Results: 12 passed, 0 failed
```

## Integration with CI/CD

This script can be integrated into continuous integration workflows to ensure the research skill meets quality standards before deployment:

```yaml
# Example GitHub Actions workflow
- name: Validate Research Skill
  run: ./skills/research/scripts/validate.sh
```

## Exit Codes

- **0**: All checks passed (success)
- **1**: One or more checks failed
- **2**: Script execution error

## Development Workflow

### Adding New Features to Research Skill

1. **Implement the feature** in the skill's codebase
2. **Update SKILL.md** to document the new functionality
3. **Run validation script** to ensure documentation is complete:
   ```bash
   ./skills/research/scripts/validate.sh
   ```
4. **Address any failures** by updating documentation or code
5. **Commit changes** with confidence the skill meets quality standards

### Updating Validation Criteria

To modify validation criteria:

1. **Edit validate.sh** to add/remove checks as needed
2. **Update this documentation** to reflect new validation requirements
3. **Test the updated script** against the current skill implementation
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: research" "grep -q '^name: research' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 3 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 3 ]"
check "SKILL.md mentions .agents/research/ output path" "grep -q '\.agents/research/' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions Explore agent" "grep -qi 'explore' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --auto flag" "grep -q '\-\-auto' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions ao know inject" "grep -q 'ao know inject\|ao know search' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions knowledge flywheel" "grep -qi 'knowledge' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions backend detection" "grep -qi 'backend\|spawn' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions quality validation" "grep -qi 'coverage\|depth\|gap' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


