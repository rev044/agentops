---
name: council
description: 'Multi-model consensus council. Spawns parallel judges with configurable perspectives. Modes: validate, brainstorm, research. Triggers: "council", "get consensus", "multi-model review", "multi-perspective review", "council validate", "council brainstorm", "council research".'
---


# $council — Multi-Model Consensus Council

Spawn parallel judges with different perspectives, consolidate into consensus. Works for any task — validation, research, brainstorming.

## Quick Start

```bash
$council --quick validate recent                               # fast inline check
$council validate this plan                                    # validation (2 agents)
$council brainstorm caching approaches                         # brainstorm
$council validate the implementation                          # validation (critique triggers map here)
$council research kubernetes upgrade strategies                # research
$council research the CI/CD pipeline bottlenecks               # research (analyze triggers map here)
$council --preset=security-audit validate the auth system      # preset personas
$council --deep --explorers=3 research upgrade automation      # deep + explorers
$council --debate validate the auth system                # adversarial 2-round review
$council --deep --debate validate the migration plan      # thorough + debate
$council                                                       # infers from context
```

Council works independently — no RPI workflow, no ratchet chain, no `ao` CLI required. Zero setup beyond initial install.

## Modes

| Mode | Agents | Execution Backend | Use Case |
|------|--------|-------------------|----------|
| `--quick` | 0 (inline) | Self | Fast single-agent check, no spawning |
| default | 2 | Runtime-native (Codex sub-agents preferred; Claude teams fallback) | Independent judges (no perspective labels) |
| `--deep` | 3 | Runtime-native | Thorough review |
| `--mixed` | 3+3 | Runtime-native + Codex CLI | Cross-vendor consensus |
| `--debate` | 2+ | Runtime-native | Adversarial refinement (2 rounds) |

```bash
$council --quick validate recent   # inline single-agent check, no spawning
$council recent                    # 2 runtime-native judges
$council --deep recent             # 3 runtime-native judges
$council --mixed recent            # runtime-native + Codex CLI
```

### Spawn Backend (MANDATORY)

Council requires a runtime that can **spawn parallel subagents** and (for `--debate`) **send messages between agents**. Use whatever multi-agent primitives your runtime provides. If no multi-agent capability is detected, fall back to `--quick` (inline single-agent).

**Required capabilities:**
- **Spawn subagent** — create a parallel agent with a prompt (required for all modes except `--quick`)
- **Agent messaging** — send a message to a specific agent (required for `--debate`)

Skills describe WHAT to do, not WHICH tool to call. See `skills/shared/SKILL.md` for the capability contract.

**After detecting your backend, read the matching reference for concrete spawn/wait/message/cleanup examples:**
- Claude feature contract → `..$shared/references/claude-code-latest-features.md`
- Claude Native Teams → `..$shared/references/backend-claude-teams.md`
- Codex Sub-Agents / CLI → `..$shared/references/backend-codex-subagents.md`
- Background Tasks → `..$shared/references/backend-background-tasks.md`
- Inline (`--quick`) → `..$shared/references/backend-inline.md`

See also `references/cli-spawning.md` for council-specific spawning flow (phases, timeouts, output collection).

## When to Use `--debate`

Use `--debate` for high-stakes or ambiguous reviews where judges are likely to disagree:
- Security audits, architecture decisions, migration plans
- Reviews where multiple valid perspectives exist
- Cases where a missed finding has real consequences

Skip `--debate` for routine validation where consensus is expected. Debate adds R2 latency (judges stay alive and process a second round via backend messaging).

**Incompatibilities:**
- `--quick` and `--debate` cannot be combined. `--quick` runs inline with no spawning; `--debate` requires multi-agent rounds. If both are passed, exit with error: "Error: --quick and --debate are incompatible."
- `--debate` is only supported with validate mode. Brainstorm and research do not produce PASS/WARN/FAIL verdicts. If combined, exit with error: "Error: --debate is only supported with validate mode."

## Task Types

| Type | Trigger Words | Perspective Focus |
|------|---------------|-------------------|
| **validate** | validate, check, review, assess, critique, feedback, improve | Is this correct? What's wrong? What could be better? |
| **brainstorm** | brainstorm, explore, options, approaches | What are the alternatives? Pros/cons? |
| **research** | research, investigate, deep dive, explore deeply, analyze, examine, evaluate, compare | What can we discover? What are the properties, trade-offs, and structure? |

Natural language works — the skill infers task type from your prompt.

### First-pass rigor gate for plan/spec validation (MANDATORY)

When mode is `validate` and the target is a plan/spec/contract (or contains boundary rules, state transitions, or conformance tables), judges must apply this gate before returning `PASS`:

1. Canonical mutation + ack sequence is explicit, single-path, and non-contradictory.
2. Consume-at-most-once path is crash-safe with explicit atomic boundary and restart recovery semantics.
3. Status/precedence behavior is defined with a field-level truth table and anomaly reason codes for conflicting evidence.
4. Conformance includes explicit boundary failpoint tests and deterministic assertions for replay/no-duplicate-effect outcomes.

Verdict policy for this gate:
- Missing or contradictory gate item: minimum `WARN`.
- Missing deterministic conformance coverage for any gate item: minimum `WARN`.
- Critical lifecycle invariant not mechanically verifiable: `FAIL`.

---

## Architecture

### Context Budget Rule (CRITICAL)

Judges write ALL analysis to output files. Messages to the lead contain ONLY a
minimal completion signal: `{"type":"verdict","verdict":"...","confidence":"...","file":"..."}`.
The lead reads output files during consolidation. This prevents N judges from
exploding the lead's context window with N full reports via backend messaging.

**Consolidation runs inline as the lead** — no separate chairman agent. The lead
reads each judge's output file sequentially with the Read tool and synthesizes.

### Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Phase 1: Build Packet (JSON)                                   │
│  - Task type (validate/brainstorm/research)                      │
│  - Target description                                           │
│  - Context (files, diffs, prior decisions)                      │
│  - Perspectives to assign                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 1a: Select spawn backend                                  │
│  codex_subagents | claude_teams | background_fallback            │
│  Team lead = spawner (this agent)                                │
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┴─────────────────┐
            ▼                                   ▼
┌───────────────────────┐           ┌───────────────────────┐
│  RUNTIME-NATIVE JUDGES│           │     CODEX AGENTS      │
│ (spawn_agent or teams)│           │  (Bash tool, parallel)│
│                       │           │  Agent 1 (independent │
│  Agent 1 (independent │           │    or with preset)    │
│    or with preset)    │           │  Agent 2              │
│  Agent 2              │           │  Agent 3              │
│  Agent 3 (--deep only)│           │  (--mixed only)       │
│  (--deep/--mixed only)│           │                       │
│                       │           │  Output: JSON + MD    │
│  Write files, then    │           │  Files: .agents/      │
│  wait()/signal to     │           │    council/codex-*    │
│  lead                 │           │                       │
│  Files: .agents/      │           └───────────────────────┘
│    council/judge-*    │                       │
└───────────────────────┘                       │
            │                                   │
            └─────────────────┬─────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 2: Consolidation (Team Lead — inline, no extra agent)    │
│  - Receive MINIMAL completion signals (verdict + file path)     │
│  - Read each judge's output file with Read tool                 │
│  - If schema_version is missing from a judge's output, treat    │
│    as version 0 (backward compatibility)                        │
│  - Compute consensus verdict                                    │
│  - Identify shared findings                                     │
│  - Surface disagreements with attribution                       │
│  - Generate Markdown report for human                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 3: Cleanup                                               │
│  - Cleanup backend resources (close_agent / TeamDelete / none)  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Output: Markdown Council Report                                │
│  - Consensus: PASS/WARN/FAIL                                    │
│  - Shared findings                                              │
│  - Disagreements (if any)                                       │
│  - Recommendations                                              │
└─────────────────────────────────────────────────────────────────┘
```

### Graceful Degradation

| Failure | Behavior |
|---------|----------|
| 1 of N agents times out | Proceed with N-1, note in report |
| All Codex CLI agents fail | Proceed with runtime-native judges only, note degradation |
| All agents fail | Return error, suggest retry |
| Codex CLI not installed | Skip Codex CLI judges, continue with runtime judges only (warn user) |
| No multi-agent capability | Fall back to `--quick` (inline single-agent review) |
| No agent messaging | `--debate` unavailable, single-round review only |
| Output dir missing | Create `.agents/council/` automatically |

Timeout: 120s per agent (configurable via `--timeout=N` in seconds).

**Minimum quorum:** At least 1 agent must respond for a valid council. If 0 agents respond, return error.

### Pre-Flight Checks

1. **Multi-agent capability:** Detect whether runtime supports spawning parallel subagents. If not, degrade to `--quick`.
2. **Agent messaging:** Detect whether runtime supports agent-to-agent messaging. If not, disable `--debate`.
3. **Codex CLI judges (--mixed only):** Check `which codex`, test model availability, test `--output-schema` support. Downgrade mixed mode when unavailable.
4. **Agent count:** Verify `judges * (1 + explorers) <= MAX_AGENTS (12)`
5. **Output dir:** `mkdir -p .agents/council`

---

## Quick Mode (`--quick`)

Single-agent inline validation. No subprocess spawning, no Task tool, no Codex. The current agent performs a structured self-review using the same output schema as a full council.

**When to use:** Routine checks, mid-implementation sanity checks, pre-commit quick scan.

**Execution:** Gather context (files, diffs) -> perform structured self-review inline using the council output_schema (verdict, confidence, findings, recommendation) -> write report to `.agents/council/YYYY-MM-DD-quick-<target>.md` labeled as `Mode: quick (single-agent)`.

**Limitations:** No cross-perspective disagreement, no cross-vendor insights, lower confidence ceiling. Not suitable for security audits or architecture decisions.

---

## Packet Format (JSON)

The packet sent to each agent. **File contents are included inline** — agents receive the actual code/plan text in the packet, not just paths. This ensures both Claude and Codex agents can analyze without needing file access.

If `.agents/ao/environment.json` exists, include it in the context packet so judges can reason about available tools and environment state.

```json
{
  "council_packet": {
    "version": "1.0",
    "mode": "validate | brainstorm | research",
    "target": "Implementation of user authentication system",
    "context": {
      "files": [
        {
          "path": "src/auth/jwt.py",
          "content": "<file contents inlined here>"
        },
        {
          "path": "src/auth/middleware.py",
          "content": "<file contents inlined here>"
        }
      ],
      "diff": "git diff output if applicable",
      "spec": {
        "source": "bead na-0042 | plan doc | none",
        "content": "The spec/bead description text (optional — included when wrapper provides it)"
      },
      "prior_decisions": [
        "Using JWT, not sessions",
        "Refresh tokens required"
      ]
    },
    "perspective": "skeptic (only when --preset or --perspectives used)",
    "perspective_description": "What could go wrong? (only when --preset or --perspectives used)",
    "output_schema": {
      "verdict": "PASS | WARN | FAIL",
      "confidence": "HIGH | MEDIUM | LOW",
      "key_insight": "Single sentence summary",
      "findings": [
        {
          "severity": "critical | significant | minor",
          "category": "security | architecture | performance | style",
          "description": "What was found",
          "location": "file:line if applicable",
          "recommendation": "How to address",
          "fix": "Specific action to resolve this finding",
          "why": "Root cause or rationale",
          "ref": "File path, spec anchor, or doc reference"
        }
      ],
      "recommendation": "Concrete next step",
      "schema_version": 2
    }
  }
}
```

---

## Perspectives

> **Perspectives & Presets:** Use `Read` tool on `skills/council/references/personas.md` for persona definitions, preset configurations, and custom perspective details.

**Auto-Escalation:** When `--preset` or `--perspectives` specifies more perspectives than the current judge count, automatically escalate judge count to match. The `--count` flag overrides auto-escalation.

---

## Named Perspectives

Named perspectives assign each judge a specific viewpoint. Pass `--perspectives="a,b,c"` for free-form names, or `--perspectives-file=<path>` for YAML with focus descriptions:

```bash
$council --perspectives="security-auditor,performance-critic,simplicity-advocate" validate src/auth/
$council --perspectives-file=.agents/perspectives/api-review.yaml validate src/api/
```

**YAML format** for `--perspectives-file`:

```yaml
perspectives:
  - name: security-auditor
    focus: Find security vulnerabilities and trust boundary violations
  - name: performance-critic
    focus: Identify performance bottlenecks and scaling risks
```

**Flag priority:** `--perspectives`/`--perspectives-file` override `--preset` perspectives. `--count` always overrides judge count. Without `--count`, judge count auto-escalates to match perspective count.

See [references/personas.md](references/personas.md) for all built-in presets and their perspective definitions.

---

## Explorer Sub-Agents

> **Explorer Details:** Use `Read` tool on `skills/council/references/explorers.md` for explorer architecture, prompts, sub-question generation, and timeout configuration.

**Summary:** Judges can spawn explorer sub-agents (`--explorers=N`, max 5) for parallel deep-dive research. Total agents = `judges * (1 + explorers)`, capped at MAX_AGENTS=12.

---

## Debate Phase (`--debate`)

> **Debate Protocol:** Use `Read` tool on `skills/council/references/debate-protocol.md` for full debate execution flow, R1-to-R2 verdict injection, timeout handling, and cost analysis.

**Summary:** Two-round adversarial review. R1 produces independent verdicts. R2 sends other judges' verdicts via backend messaging (`send_input` for Codex sub-agents, runtime-native messaging for other backends) for steel-manning and revision. Only supported with validate mode.

---

## Agent Prompts

> **Agent Prompts:** Use `Read` tool on `skills/council/references/agent-prompts.md` for judge prompts (default and perspective-based), consolidation prompt, and debate R2 message template.

---

## Consensus Rules

| Condition | Verdict |
|-----------|---------|
| All PASS | PASS |
| Any FAIL | FAIL |
| Mixed PASS/WARN | WARN |
| All WARN | WARN |

Disagreement handling:
- If Claude says PASS and Codex says FAIL → DISAGREE (surface both)
- Severity-weighted: Security FAIL outweighs style WARN

**DISAGREE resolution:** When vendors disagree, the spawner presents both positions with reasoning and defers to the user. No automatic tie-breaking — cross-vendor disagreement is a signal worth human attention.

---

## Output Format

> **Report Templates:** Use `Read` tool on `skills/council/references/output-format.md` for full report templates (validate, brainstorm, research) and debate report additions (verdict shifts, convergence detection).

All reports write to `.agents/council/YYYY-MM-DD-<type>-<target>.md`.


---

## Configuration

### Partial Completion

**Minimum quorum:** 1 agent. **Recommended:** 80% of judges. On timeout, proceed with remaining judges and note in report. On user cancellation, shutdown all judges and generate partial report with INCOMPLETE marker.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COUNCIL_TIMEOUT` | 120 | Agent timeout in seconds |
| `COUNCIL_CODEX_MODEL` | (user's default) | Override Codex model for --mixed. Omit `-m` flag to use the user's configured default. |
| `COUNCIL_CLAUDE_MODEL` | sonnet | Claude model for judges (sonnet default — use opus for high-stakes via `--profile=thorough`) |
| `COUNCIL_EXPLORER_MODEL` | sonnet | Model for explorer sub-agents |
| `COUNCIL_EXPLORER_TIMEOUT` | 60 | Explorer timeout in seconds |
| `COUNCIL_R2_TIMEOUT` | 90 | Maximum wait time for R2 debate completion after sending debate messages. Shorter than R1 since judges already have context. |

### Flags

| Flag | Description |
|------|-------------|
| `--deep` | 3 Claude agents instead of 2 |
| `--mixed` | Add 3 Codex agents |
| `--debate` | Enable adversarial debate round (2 rounds via backend messaging, same agents). Incompatible with `--quick`. |
| `--timeout=N` | Override timeout in seconds (default: 120) |
| `--perspectives="a,b,c"` | Custom perspective names (each name sets the judge's system prompt to adopt that viewpoint) |
| `--perspectives-file=<path>` | Load named perspectives from a YAML file (see Named Perspectives below) |
| `--preset=<name>` | Built-in persona preset (security-audit, architecture, research, ops, code-review, plan-review, doc-review, retrospective, product, developer-experience) |
| `--count=N` | Override agent count per vendor (e.g., `--count=4` = 4 Claude, or 4+4 with --mixed). Subject to MAX_AGENTS=12 cap. |
| `--explorers=N` | Explorer sub-agents per judge (default: 0, max: 5). Max effective value depends on judge count. Total agents capped at 12. |
| `--explorer-model=M` | Override explorer model (default: sonnet) |
| `--technique=<name>` | Brainstorm technique (scamper, six-hats, reverse). Case-insensitive. Only applicable to brainstorm mode — error if combined with validate/research. If omitted, unstructured brainstorm (current behavior). See `references/brainstorm-techniques.md`. |
| `--profile=<name>` | Model quality profile (thorough, balanced, fast). Error if unrecognized name. Overridden by `COUNCIL_CLAUDE_MODEL` env var (highest priority), then by explicit `--count`/`--deep`/`--mixed`. See `references/model-profiles.md`. |

---

## CLI Spawning Commands

> **CLI Spawning:** Use `Read` tool on `skills/council/references/cli-spawning.md` for team setup, Claude/Codex agent spawning, parallel execution, debate R2 commands, cleanup, and model selection.

---

## Examples

```bash
$council validate recent                                        # 2 judges, recent commits
$council --deep --preset=architecture research the auth system  # 3 judges with architecture personas
$council --mixed validate this plan                             # 3 Claude + 3 Codex
$council --deep --explorers=3 research upgrade patterns         # 12 agents (3 judges x 4)
$council --preset=security-audit --deep validate the API        # attacker, defender, compliance
$council --preset=doc-review validate README.md                  # 4 doc judges with named perspectives
$council brainstorm caching strategies for the API              # 2 judges explore options
$council --technique=scamper brainstorm API improvements               # structured SCAMPER brainstorm
$council --technique=six-hats brainstorm migration strategy            # parallel perspectives brainstorm
$council --profile=thorough validate the security architecture       # opus, 3 judges, 120s timeout
$council --profile=fast validate recent                               # haiku, 2 judges, 60s timeout
$council research Redis vs Memcached for session storage        # 2 judges assess trade-offs
$council validate the implementation plan in PLAN.md            # structured plan feedback
$council --preset=doc-review validate docs/ARCHITECTURE.md             # 4 doc review judges
$council --perspectives="security-auditor,perf-critic" validate src/   # named perspectives
$council --perspectives-file=.agents/perspectives/custom.yaml validate # perspectives from file
```

### Fast Single-Agent Validation

**User says:** `$council --quick validate recent`

**What happens:**
1. Agent gathers context (recent diffs, files) inline without spawning
2. Agent performs structured self-review using council output schema
3. Report written to `.agents/council/YYYY-MM-DD-quick-<target>.md` labeled `Mode: quick (single-agent)`

**Result:** Fast sanity check for routine validation (no cross-perspective insights or debate).

### Adversarial Debate Review

**User says:** `$council --debate validate the auth system`

**What happens:**
1. Agent spawns 2 judges (runtime-native backend) with independent perspectives
2. R1: Judges assess independently, write verdicts to `.agents/council/`
3. R2: Team lead sends other judges' verdicts via backend messaging
4. Judges revise positions based on cross-perspective evidence
5. Consolidation: Team lead computes consensus with convergence detection

**Result:** Two-round review with steel-manning and revision, useful for high-stakes decisions.

### Cross-Vendor Consensus with Explorers

**User says:** `$council --mixed --explorers=2 research Kubernetes upgrade strategies`

**What happens:**
1. Agent spawns 3 Claude judges + 3 Codex judges (6 total)
2. Each judge spawns 2 explorer sub-agents (6 x 3 = 18 total agents, exceeds MAX_AGENTS)
3. Agent auto-scales to 2 judges per vendor (4 x 3 = 12 agents at limit)
4. Explorers perform parallel deep-dives, return sub-findings to judges
5. Judges consolidate explorer findings with own research

**Result:** Cross-vendor research with deep exploration, capped at 12 total agents.

---

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "Error: --quick and --debate are incompatible" | Both flags passed together | Use `--quick` for fast inline check OR `--debate` for multi-round review, not both |
| "Error: --debate is only supported with validate mode" | Debate flag used with brainstorm/research | Remove `--debate` or switch to validate mode — brainstorming/research have no PASS/FAIL verdicts |
| Council spawns fewer agents than expected | `--explorers=N` exceeds MAX_AGENTS (12) | Agent auto-scales judge count. Check report header for actual judge count. Reduce `--explorers` or use `--count` to manually set judges |
| Codex judges skipped in --mixed mode | Codex CLI not on PATH | Install Codex CLI (`brew install codex`). Model uses user's configured default — no specific model required. |
| No output files in `.agents/council/` | Permission error or disk full | Check directory permissions with `ls -ld .agents/council/`. Council auto-creates missing dirs. |
| Agent timeout after 120s | Slow file reads or network issues | Increase timeout with `--timeout=300` or check `COUNCIL_TIMEOUT` env var. Default: 120s. |

---

## Migration from judge

`$council` replaces the old judge skill. Migration:

| Old | New |
|-----|-----|
| judge recent | `$council validate recent` |
| judge 2 opus | `$council recent` (default) |
| judge 3 opus | `$council --deep recent` |

The judge skill is deprecated. Use `$council`.

---

## Multi-Agent Architecture

Council uses whatever multi-agent primitives your runtime provides. Each judge is a parallel subagent that writes output to a file and sends a minimal completion signal to the lead.

### Deliberation Protocol

The `--debate` flag implements the **deliberation protocol** pattern:
> Independent assessment → evidence exchange → position revision → convergence analysis

- **R1:** Spawn judges as parallel subagents. Each assesses independently, writes verdict to file, signals completion.
- **R2:** Lead sends other judges' verdict summaries to each judge via agent messaging. Judges revise and write R2 files.
- **Consolidation:** Lead reads all output files, computes consensus.
- **Cleanup:** Shut down judges via runtime's cleanup mechanism.

### Communication Rules

- **Judges → lead only.** Judges never message each other directly. This prevents anchoring.
- **Lead → judges.** Only the lead sends follow-ups (for debate R2).
- **No shared task mutation by judges.** Lead manages coordination state.

### Ralph Wiggum Compliance

Council maintains fresh-context isolation (Ralph Wiggum pattern) with one documented exception:

**`--debate` reuses judge context across R1 and R2.** This is intentional. Judges persist within a single atomic council invocation — they do NOT persist across separate council calls. The rationale:

- Judges benefit from their own R1 analytical context (reasoning chain, not just the verdict JSON) when evaluating other judges' positions in R2
- Re-spawning with only the verdict summary (~200 tokens) would lose the judge's working memory of WHY they reached their verdict
- The exception is bounded: max 2 rounds, within one invocation, with explicit cleanup

Without `--debate`, council is fully Ralph-compliant: each judge is a fresh spawn, executes once, writes output, and terminates.

### Degradation

If no multi-agent capability is detected, council falls back to `--quick` (inline single-agent review). If agent messaging is unavailable, `--debate` degrades to single-round review with a note in the report.

### Judge Naming

Convention: `council-YYYYMMDD-<target>` (e.g., `council-20260206-auth-system`).

Judge names: `judge-{N}` for independent judges (e.g., `judge-1`, `judge-2`), or `judge-{perspective}` when using presets/perspectives (e.g., `judge-error-paths`, `judge-feasibility`). Use the same logical names across both Codex and Claude backends.

---

## See Also

- `skills/vibe/SKILL.md` — Complexity + council for code validation (uses `--preset=code-review` when spec found)
- `skills/pre-mortem/SKILL.md` — Plan validation (uses `--preset=plan-review`, always 3 judges)
- `skills/post-mortem/SKILL.md` — Work wrap-up (uses `--preset=retrospective`, always 3 judges + retro)
- `skills/swarm/SKILL.md` — Multi-agent orchestration
- `skills/standards/SKILL.md` — Language-specific coding standards
- `skills/research/SKILL.md` — Codebase exploration (complementary to council research mode)

## Reference Documents

- [references/backend-background-tasks.md](references/backend-background-tasks.md)
- [references/backend-claude-teams.md](references/backend-claude-teams.md)
- [references/backend-codex-subagents.md](references/backend-codex-subagents.md)
- [references/backend-inline.md](references/backend-inline.md)
- [references/brainstorm-techniques.md](references/brainstorm-techniques.md)
- [references/claude-code-latest-features.md](references/claude-code-latest-features.md)
- [references/model-profiles.md](references/model-profiles.md)
- [references/presets.md](references/presets.md)
- [references/quick-mode.md](references/quick-mode.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)

---

## References

### agent-prompts.md

# Agent Prompts

## Judge Agent Prompt -- Default (Independent, No Perspectives)

Used when no `--preset` or `--perspectives` flag is provided:

```
You are Council Judge {N}. You are one of {TOTAL} independent judges evaluating the same target.
You are a teammate on team "{TEAM_NAME}".

{JSON_PACKET}

Instructions:
1. Analyze the target thoroughly
2. Write your FULL analysis to: .agents/council/{OUTPUT_FILENAME}
   - Start with a JSON code block matching the output_schema
   - Follow with Markdown explanation
   - ALL detailed findings, reasoning, and recommendations go in this file
3. Send a SHORT completion signal to the team lead (see format below)
4. You may receive follow-up messages (e.g., debate round 2). Process and respond.

Your job is to find problems. A PASS with caveats is less valuable than a specific FAIL.

FIRST-PASS CONTRACT COMPLETENESS GATE (validate mode):
If the target is a plan/spec/contract, you MUST audit these before returning PASS:
1. Canonical mutation + ack sequence is explicit, single-path, and non-contradictory.
2. Consume-at-most-once flow is crash-safe with explicit atomic boundary and restart recovery semantics.
3. Status/precedence behavior is specified with a field-level truth table and anomaly reason codes.
4. Conformance includes boundary failpoint tests with deterministic replay/no-duplicate-effect assertions.

Gate verdict rules:
- Any missing/contradictory item -> WARN minimum.
- Any missing deterministic conformance coverage -> WARN minimum.
- Any critical lifecycle invariant not mechanically verifiable -> FAIL.

For every finding, include these structured remediation fields:
- fix: Specific action to resolve this finding
- why: Root cause or rationale
- ref: File path, spec anchor, or doc reference that supports this finding

CONTEXT BUDGET: Your message to the team lead must be MINIMAL to avoid exploding
the lead's context window. Send ONLY the completion signal below — the lead reads
your full analysis from the output file.

Your message MUST be exactly this JSON block and nothing else:

\`\`\`json
{
  "type": "verdict",
  "verdict": "PASS | WARN | FAIL",
  "confidence": "HIGH | MEDIUM | LOW",
  "file": ".agents/council/{OUTPUT_FILENAME}"
}
\`\`\`

Do NOT include key_insight, findings, or explanation in the message.
Do NOT add prose before or after the JSON block.
The lead reads your output file for all details.

Rules:
- Do NOT message other judges -- all communication through team lead
- Do NOT access shared task state -- lead manages coordination
- Keep messages to lead MINIMAL -- file has the details
```

## Judge Agent Prompt -- With Perspectives (Preset or Custom)

Used when `--preset` or `--perspectives` flag is provided. When using a preset with named personas, the persona name replaces the generic "Council Member N" label via `{PERSONA_NAME}`. For custom perspectives without names, `{PERSONA_NAME}` falls back to `Council Member {N}`.

```
You are **{PERSONA_NAME}**, the {PERSPECTIVE} judge.
You are a teammate on team "{TEAM_NAME}".

{JSON_PACKET}

Your angle: {PERSPECTIVE_DESCRIPTION}

FIRST-PASS CONTRACT COMPLETENESS GATE (validate mode):
If the target is a plan/spec/contract, you MUST audit these before returning PASS:
1. Canonical mutation + ack sequence is explicit, single-path, and non-contradictory.
2. Consume-at-most-once flow is crash-safe with explicit atomic boundary and restart recovery semantics.
3. Status/precedence behavior is specified with a field-level truth table and anomaly reason codes.
4. Conformance includes boundary failpoint tests with deterministic replay/no-duplicate-effect assertions.

Gate verdict rules:
- Any missing/contradictory item -> WARN minimum.
- Any missing deterministic conformance coverage -> WARN minimum.
- Any critical lifecycle invariant not mechanically verifiable -> FAIL.

For every finding, include these structured remediation fields:
- fix: Specific action to resolve this finding
- why: Root cause or rationale
- ref: File path, spec anchor, or doc reference that supports this finding

Instructions:
1. Analyze the target from your perspective
2. Write your FULL analysis to: .agents/council/{OUTPUT_FILENAME}
   - Start with a JSON code block matching the output_schema
   - Follow with Markdown explanation
   - ALL detailed findings, reasoning, and recommendations go in this file
3. Send a SHORT completion signal to the team lead (see format below)
4. You may receive follow-up messages (e.g., debate round 2). Process and respond.

CONTEXT BUDGET: Your message to the team lead must be MINIMAL to avoid exploding
the lead's context window. Send ONLY the completion signal below — the lead reads
your full analysis from the output file.

Your message MUST be exactly this JSON block and nothing else:

\`\`\`json
{
  "type": "verdict",
  "verdict": "PASS | WARN | FAIL",
  "confidence": "HIGH | MEDIUM | LOW",
  "file": ".agents/council/{OUTPUT_FILENAME}"
}
\`\`\`

Do NOT include key_insight, findings, or explanation in the message.
Do NOT add prose before or after the JSON block.
The lead reads your output file for all details.

Rules:
- Do NOT message other judges -- all communication through team lead
- Do NOT access shared task state -- lead manages coordination
- Keep messages to lead MINIMAL -- file has the details
```

## Debate Round 2 Message (via agent messaging)

When `--debate` is active, the team lead sends this message to each judge after R1 completes. The judge already has its own R1 analysis in context (no truncation needed).

```
## Debate Round 2

## Anti-Anchoring Protocol

Before reviewing other judges' verdicts:

1. **RESTATE your R1 position** -- Write 2-3 sentences summarizing your own R1 verdict
   and the key evidence that led to it. This anchors you to YOUR OWN reasoning before
   exposure to others.

2. **Then review other verdicts** -- Only after restating your position, read the
   other judges' JSON verdicts below.

3. **Evidence bar for changing verdict** -- You may only change your verdict if you can
   cite a SPECIFIC technical detail, code location, or factual error that you missed
   in R1. "Judge 2 made a good point" is NOT sufficient. "Judge 2 found an unchecked
   error path at auth.py:45 that I missed" IS sufficient.

Other judges' R1 verdicts (SUMMARY ONLY — read their output files for full analysis):

### {OTHER_JUDGE_PERSPECTIVE}
Verdict: {VERDICT_VALUE} | Confidence: {CONFIDENCE} | File: {R1_OUTPUT_FILE}
(repeat for each other judge)

To review a judge's full reasoning, read their output file listed above.

## Debate Instructions

You MUST follow this structure:

**IF judges disagreed in R1 (different verdicts):**

1. **STEEL-MAN**: State the strongest version of an argument from another
   judge that you initially disagree with. Show you understand it fully
   before responding to it.

2. **CHALLENGE**: Identify at least one specific claim from another judge
   that you believe is wrong or incomplete. Cite evidence.

3. **ACKNOWLEDGE**: Identify at least one point from another judge that
   strengthens, modifies, or adds to your analysis.

4. **REVISE OR CONFIRM**: State your final verdict with specific reasoning.
   If changing from R1, explain exactly what new evidence changed your mind.
   If confirming, explain why the opposing arguments did not persuade you.

**IF all judges agreed in R1 (same verdict):**

Do NOT invent disagreement. Instead, stress-test the consensus:

1. **DEVIL'S ADVOCATE**: What is the strongest argument AGAINST the consensus?
2. **BLIND SPOT**: What perspective or risk did all judges overlook?
3. **CONFIRM OR REVISE**: Does the consensus hold under scrutiny?

Do NOT change your verdict merely because others disagree.
Do NOT defensively maintain without engaging with opposing arguments.

Write revised verdict to: .agents/council/{R2_OUTPUT_FILENAME}
Include "debate_notes" field in your JSON.

CONTEXT BUDGET: After writing your R2 file, send the SAME minimal completion signal
as R1 — verdict, confidence, file path only. No prose, no findings in the message.

\`\`\`json
{
  "type": "verdict",
  "verdict": "PASS | WARN | FAIL",
  "confidence": "HIGH | MEDIUM | LOW",
  "file": ".agents/council/{R2_OUTPUT_FILENAME}"
}
\`\`\`

Required JSON format for the OUTPUT FILE (not the message):

{
  "verdict": "PASS | WARN | FAIL",
  "confidence": "HIGH | MEDIUM | LOW",
  "key_insight": "...",
  "findings": [...],
  "recommendation": "...",
  "debate_notes": {
    "revised_from": "original verdict if changed, or null if unchanged",
    "steel_man": "strongest opposing argument I considered",
    "challenges": [
      {
        "target_judge": "judge-2",
        "claim": "what they claimed",
        "response": "why I agree or disagree"
      }
    ],
    "acknowledgments": [
      {
        "source_judge": "judge-1",
        "point": "what they found",
        "impact": "how it affected my analysis"
      }
    ]
  }
}

Then provide a Markdown explanation of your debate reasoning.
```

## Brainstorm Technique Injection

When `--technique` is specified, inject the technique's prompt template into each judge's instructions BEFORE the standard brainstorm prompt. The technique prompt is loaded from `references/brainstorm-techniques.md`.

Template:
```
{TECHNIQUE_PROMPT_INJECTION}

In addition to the technique framework above, also provide your independent creative analysis.
```

If `--technique` is used with a non-brainstorm task type (validate, research), exit with error: "Error: --technique is only applicable to brainstorm mode." This matches the --quick/--debate incompatibility pattern.

## Consolidation Prompt

**Codex output format note:** When Codex judges used `--output-schema`, their output files are pure JSON (`.json` extension) conforming to `skills/council/schemas/verdict.json`. Parse directly with JSON. When fallback was used, output files are markdown (`.md` extension) with a JSON code block that must be extracted (current behavior). Check file extension to determine parse strategy.

The consolidation phase runs as the TEAM LEAD (this agent), NOT as a separate spawned agent.
This avoids creating yet another context window. The lead reads each judge's output file
directly using the Read tool, then synthesizes inline.

**Consolidation procedure:**

1. Collect the list of judge output files from their completion signals
2. Read each file with the Read tool (one file at a time to manage context)
3. Extract the JSON verdict block from each file
4. Synthesize the final report

When synthesizing, follow these guidelines. When judges have persona names (from presets), use those names in attribution (e.g., "**Red** (attacker) found...") instead of generic "Judge 1" labels.

Ensure all consolidated findings have fix/why/ref populated. If a judge omitted these fields, infer them from the judge's analysis. Use fallbacks:
- fix = finding.fix || finding.recommendation || "No fix specified"
- why = finding.why || "No root cause specified"
- ref = finding.ref || finding.location || "No reference"

For validate mode:
1. **Consensus Verdict**: PASS if all PASS, FAIL if any FAIL, else WARN
2. **Shared Findings**: Points all judges agree on
3. **Disagreements**: Where judges differ (with attribution)
4. **Cross-Vendor Insights**: (if --mixed) Unique findings per vendor
5. **Final Recommendation**: Concrete next step

For brainstorm mode:
1. **Options Explored**: Each option with multi-perspective assessment
2. **Trade-offs**: Pros/cons matrix
3. **Recommendation**: Synthesized best approach

For research mode:
1. **Facets Explored**: What each judge investigated
2. **Synthesized Findings**: Merged findings organized by theme
3. **Open Questions**: What remains unknown
4. **Recommendation**: Next steps for further investigation or action

Output format: Markdown report for human consumption.
```

## Consolidation Prompt -- Debate Additions

When `--debate` is used, append this to the consolidation prompt:

```
## Additional Instructions (Debate Mode)

You have received TWO rounds of judge reports.

Round 1 (independent assessment): Each judge evaluated independently.
Round 2 (post-debate revision): Each judge reviewed all other judges' findings and revised.

When synthesizing:
1. Use Round 2 verdicts for the CONSENSUS VERDICT computation (PASS/WARN/FAIL)
2. Use Round 1 verdicts for FINDING COMPLETENESS -- a finding in R1 but dropped in R2 without explanation deserves mention
3. Compare R1 and R2 to identify position shifts
4. Flag judges who changed verdict without citing a specific technical detail, a misinterpretation they corrected, or a finding they missed (possible anchoring)
   Flag judges who changed verdict without citing:
   - A specific file:line or code location
   - A factual error in their R1 analysis
   - A missing test case or edge case
   These are "weak flips" -- potential anchoring, not genuine persuasion.
5. If R1 had at least 2 judges with different verdicts AND R2 is unanimous, note "Convergence detected -- review reasoning for anchoring risk"
6. In the report, include the Verdict Shifts table showing R1->R2 changes per judge
7. Detect whether debate ran via native teams (judges stayed alive between rounds) or fallback (R2 judges were re-spawned with truncated R1 verdicts). Include the `**Fidelity:**` field in the report header: "full" for native teams, "degraded" for fallback.

When a Round 2 verdict is unavailable (timeout fallback):
- Read the full R1 output file (.agents/council/YYYY-MM-DD-<target>-claude-{perspective}.md)
- Extract the JSON verdict block (first JSON code block in the file)
- Use this as the judge's verdict for consolidation
- Mark in report: "Judge {perspective}: R1 verdict (R2 timeout)"
```

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

### brainstorm-techniques.md

# Brainstorming Techniques

Structured techniques for `$council brainstorm` mode. Use `--technique=<name>` to activate.

When no technique is specified, brainstorm mode uses unstructured exploration (current behavior).

## Technique Names

Canonical allowlist for `--technique=<name>` (case-insensitive):

| Name | Section |
|------|---------|
| `scamper` | SCAMPER |
| `six-hats` | Six Thinking Hats |
| `reverse` | Reverse Brainstorming |

## SCAMPER

Systematic derivative generation. Each judge applies the SCAMPER framework to the brainstorm topic:

- **S**ubstitute — What can be replaced?
- **C**ombine — What can be merged?
- **A**dapt — What can be borrowed from elsewhere?
- **M**odify — What can be enlarged, minimized, or altered?
- **P**ut to other uses — What else could this be used for?
- **E**liminate — What can be removed?
- **R**everse — What if we did the opposite?

**When to use:** Feature ideation, product improvement, exploring variations of existing solutions.

**Judge prompt injection:**
```
Apply the SCAMPER framework to this brainstorm topic. For each of the 7 SCAMPER lenses (Substitute, Combine, Adapt, Modify, Put to other uses, Eliminate, Reverse), generate at least one concrete idea. Prioritize the top 3 ideas by feasibility and impact.
```

## Six Thinking Hats

Parallel perspectives technique. Each judge is assigned a "hat" that determines their analysis angle:

| Hat | Color | Focus |
|-----|-------|-------|
| White | Facts | What data do we have? What's missing? |
| Red | Emotions | Gut reactions, intuitions, feelings about the ideas |
| Black | Caution | Risks, dangers, what could go wrong |
| Yellow | Benefits | Value, advantages, why it might work |
| Green | Creativity | New ideas, alternatives, provocations |
| Blue | Process | Meta-view: what should we explore next? |

**When to use:** Complex decisions requiring multiple analytical angles, team alignment on priorities.

**Judge prompt injection:**
```
Apply the Six Thinking Hats framework. Analyze this topic from ALL six perspectives (White=facts, Red=emotions, Black=caution, Yellow=benefits, Green=creativity, Blue=process). Structure your response with a section per hat. Highlight which hat reveals the most critical insight.
```

## Reverse Brainstorming

"How could we make this worse?" then invert. Judges deliberately brainstorm ways to fail, then flip each failure into a solution.

**When to use:** Problem-solving when direct approaches feel stuck, identifying hidden risks, stress-testing plans.

**Judge prompt injection:**
```
Use Reverse Brainstorming. First, brainstorm at least 5 ways to make this problem WORSE or guarantee failure. Then, for each failure mode, invert it into a specific, actionable solution. The inversions are your recommendations.
```

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

### cli-spawning.md

# Spawning Judges

## Capability Contract

Council requires these runtime capabilities. Map them to whatever your agent harness provides.

**For concrete tool call examples per backend, read the matching shared reference:**
- Claude Native Teams → `skills/shared/references/backend-claude-teams.md`
- Codex Sub-Agents / CLI → `skills/shared/references/backend-codex-subagents.md`
- Background Tasks → `skills/shared/references/backend-background-tasks.md`
- Inline → `skills/shared/references/backend-inline.md`

| Capability | Required for | What it does |
|------------|-------------|-------------|
| **Spawn parallel subagent** | All modes except `--quick` | Create N judges that run concurrently, each with a prompt |
| **Agent-to-agent messaging** | `--debate` only | Send a message to a running judge (for R2 verdict exchange) |
| **Graceful shutdown** | Cleanup | Terminate judges after consolidation |
| **Shared filesystem** | All modes | Judges write output files to `.agents/council/` |

If **spawn** is unavailable, degrade to `--quick` (inline single-agent).
If **messaging** is unavailable, `--debate` degrades to single-round review.

## Spawning Flow

### Phase 1: Spawn Judges in Parallel

For each judge (N = 2 default, 3 with `--deep`):

1. Spawn a subagent with the judge prompt (see `agent-prompts.md`)
2. Each judge receives the full context packet as its prompt
3. Track the mapping: `judge-{N}` → agent handle (for messaging and cleanup)

All judges spawn in parallel. Do not wait for one before spawning the next.

### Phase 2: Wait for Completion

Judges write output files, then send a MINIMAL completion signal:

```json
{
  "type": "verdict",
  "verdict": "PASS | WARN | FAIL",
  "confidence": "HIGH | MEDIUM | LOW",
  "file": ".agents/council/YYYY-MM-DD-<target>-judge-1.md"
}
```

Wait for all judges to signal (up to `COUNCIL_TIMEOUT`, default 120s). If a judge times out, proceed with N-1 and note in report.

### Phase 3: Debate R2 (if `--debate`)

After R1 completes, send each judge a message containing:
- Verdict summaries of OTHER judges (verdict + confidence + file path only)
- Instructions to read other judges' files for full reasoning
- The debate protocol (see `agent-prompts.md`)

CONTEXT BUDGET: Send only verdict summaries, NOT full JSON findings. Judges read files for detail.

Wait up to `COUNCIL_R2_TIMEOUT` (default 90s). If a judge doesn't respond, use their R1 verdict.

### Phase 4: Consolidation

Lead reads each judge's output file (one at a time), extracts JSON verdict, synthesizes final report. No separate agent — consolidation runs inline.

### Phase 5: Cleanup

Shut down all judges via runtime's shutdown mechanism. Cleanup MUST succeed even on partial failures:

1. Request graceful shutdown for each judge
2. Wait up to 30s for acknowledgment
3. If any judge doesn't respond, log warning, proceed anyway
4. Always run cleanup — lingering agents pollute future sessions

## Codex CLI Judges (--mixed mode)

For cross-vendor consensus, run Codex CLI processes alongside runtime-native judges:

```bash
# With structured output (preferred)
codex exec -s read-only -C "$(pwd)" --output-schema skills/council/schemas/verdict.json -o .agents/council/codex-{N}.json "{PACKET}"

# Fallback (if --output-schema unsupported)
codex exec --full-auto -C "$(pwd)" -o .agents/council/codex-{N}.md "{PACKET}"
```

Uses the user's default Codex model. Only pass `-m` if `COUNCIL_CODEX_MODEL` is explicitly set.

Flag order: `-s`/`--full-auto` → `-C` → `--output-schema` → `-o` → prompt (add `-m` before `-C` only if overriding model).

**Valid flags:** `--full-auto`, `-s`, `-m`, `-C`, `--output-schema`, `-o`, `--add-dir`
**Invalid flags:** `-q` (doesn't exist), `--quiet` (doesn't exist)

Codex CLI processes run as background shell commands — this is fine (they're separate OS processes, not agent background tasks).

## Timeout Configuration

| Timeout | Default | Description |
|---------|---------|-------------|
| Judge timeout | 120s | Max time for judge to complete (per round) |
| Shutdown grace period | 30s | Time to wait for shutdown acknowledgment |
| R2 debate timeout | 90s | Max time for R2 after sending debate messages |

## Model Selection

| Vendor | Default | Override |
|--------|---------|----------|
| Claude | sonnet | `--claude-model=opus` |
| Codex | (user's default) | `--codex-model=<model>` or `COUNCIL_CODEX_MODEL` env var |

## Output Collection

All council outputs go to `.agents/council/`:

```bash
mkdir -p .agents/council

# Judge output (R1)
.agents/council/YYYY-MM-DD-<target>-judge-1.md
.agents/council/YYYY-MM-DD-<target>-judge-error-paths.md

# Judge output (R2, when --debate)
.agents/council/YYYY-MM-DD-<target>-judge-1-r2.md

# Codex CLI output (--mixed)
.agents/council/YYYY-MM-DD-<target>-codex-1.json   # with --output-schema
.agents/council/YYYY-MM-DD-<target>-codex-1.md      # fallback

# Final consolidated report
.agents/council/YYYY-MM-DD-<target>-report.md
```

## Judge Naming

Convention: `council-YYYYMMDD-<target>` for the team/session name.

Judge names: `judge-{N}` for independent judges, `judge-{perspective}` when using presets (e.g., `judge-error-paths`, `judge-feasibility`).

### debate-protocol.md

# Debate Phase (`--debate`)

When `--debate` is passed, council runs two rounds instead of one. Round 1 produces independent verdicts. Round 2 lets judges review each other's work and revise.

**Native teams unlock the key advantage:** Judges stay alive after R1. Instead of re-spawning fresh R2 judges with truncated R1 verdicts, the team lead sends other judges' full R1 verdicts via `SendMessage`. Judges wake from idle, process R2 with full context (their own R1 analysis + others' verdicts), and write R2 files. Result: no truncation loss, no spawn overhead, richer debate.

## Execution Flow (with --debate)

```
Phase 1: Build Packet + Create Team + Spawn R1 judges as teammates
                              |
                    Collect all R1 verdicts (via SendMessage)
                    Judges go idle after R1 (stay alive)
                              |
Phase 1.5: Prepare R2 context (--debate only)
  - For each OTHER judge's R1 verdict, extract full JSON verdict
  - Team lead sends R2 instructions to each judge via SendMessage
  - Each judge already has its own R1 in context (no truncation needed)
  - Each judge receives other judges' verdicts (full JSON, not truncated)
                              |
Phase 2: Judges wake up for Round 2 (--debate only)
  - Same judge instances as R1 (not re-spawned)
  - Each judge processes via SendMessage:
    - Other judges' full R1 JSON verdicts
    - Steel-manning rebuttal prompt
    - Branch: disagreed OR agreed
  - Judges write R2 files + send completion message
                              |
                    Collect all R2 verdicts
                              |
Phase 3: Consolidation (uses R2 verdicts when --debate)
                              |
Phase 4: shutdown_request each judge, TeamDelete()
```

## Round 2 via Agent Messaging

**Branch selection (lead responsibility):**
```
r1_verdicts = [extract JSON verdict from each R1 output file]
r1_unanimous = all verdicts have same verdict value (PASS/WARN/FAIL)

For each judge:
  other_verdicts = [v for v in r1_verdicts if v.judge != this_judge]
  branch = "agreed" if r1_unanimous else "disagreed"
  send message to judge with build_r2_message(other_verdicts, branch)
```

**R2 message content:** See "Debate Round 2 Message" in `agent-prompts.md`.

**R2 output files:** Use `-r2` suffix to preserve R1 files:
```
.agents/council/YYYY-MM-DD-<target>-claude-{perspective}-r2.md
```

**With --explorers:** Explorers run in R1 only. R2 judges do not spawn explorers. Explorer findings from R1 are already in the judge's context (no truncation loss).

**With --mixed:** Only Claude judges participate in R2 (they stay alive on the team). Codex agents run once in R1 (Bash-spawned, cannot join teams). For consolidation, use Claude R2 verdicts + Codex R1 verdicts.

## R1 Verdict Injection for R2

Since judges stay alive, truncation is no longer needed for a judge's **own** R1 verdict -- it's already in their context.

For **other judges' verdicts** sent via SendMessage, include the full JSON verdict block:

```json
{
  "judge": "judge-1 (or perspective name when using presets)",
  "verdict": "WARN",
  "confidence": "HIGH",
  "key_insight": "Rate limiting missing on auth endpoints",
  "findings": [
    {"severity": "significant", "description": "No rate limiting on /login"},
    {"severity": "significant", "description": "JWT expiry too long (1h)"},
    {"severity": "minor", "description": "Missing request ID in error responses"}
  ],
  "recommendation": "Add rate limiting to auth endpoints"
}
```

Full Markdown analysis remains in `.agents/council/YYYY-MM-DD-<target>-claude-{perspective}.md` files and can be referenced by the team lead during consolidation.

## Anti-Anchoring Protocol

Before reviewing other judges' verdicts in R2:

1. **RESTATE your R1 position** -- Write 2-3 sentences summarizing your own R1 verdict and the key evidence that led to it. This anchors you to YOUR OWN reasoning before exposure to others.

2. **Then review other verdicts** -- Only after restating your position, read the other judges' JSON verdicts.

3. **Evidence bar for changing verdict** -- You may only change your verdict if you can cite a SPECIFIC technical detail, code location, or factual error that you missed in R1. "Judge 2 made a good point" is NOT sufficient. "Judge 2 found an unchecked error path at auth.py:45 that I missed" IS sufficient.

## Timeout and Failure Handling

| Scenario | Behavior |
|----------|----------|
| No R2 completion message within `COUNCIL_R2_TIMEOUT` (default 90s) | Read their R1 output file, use R1 verdict for consolidation |
| All judges fail R2 | Fall back to R1-only consolidation (note in report) |
| R1 judge timed out | No R2 message for that perspective (N-1 in R2) |
| Mixed R2 timeout | Consolidate with available R2 verdicts + R1 fallbacks |

## Cost and Latency

`--debate` adds R2 latency but **reduces spawn overhead** vs the old re-spawn approach:
- **Agents spawned:** N judges total (same instances for both rounds, not 2N)
- **Wall time:** R1 time + R2 time (sequential rounds, but R2 is faster -- no spawn delay)
- **With --mixed:** Only Claude judges get R2. Codex agents run once (Bash-spawned, cannot join teams). For consolidation, use Claude R2 verdicts + Codex R1 verdicts for consensus computation.
- **With --explorers:** Explorers run in R1 only. R2 cost = judge processing time (no explorer multiplication).
- **Non-verdict modes:** `--debate` is only supported with validate mode. If combined with brainstorm or research, exit with error: "Error: --debate is only supported with validate mode. Debate requires PASS/WARN/FAIL verdicts."

### explorers.md

# Explorer Sub-Agents

Judges can spawn explorer sub-agents for parallel deep-dive research. This is the key differentiator for `research` mode -- massive parallel exploration.

## Flag

| Flag | Default | Max | Description |
|------|---------|-----|-------------|
| `--explorers=N` | 0 | 5 | Number of explorer sub-agents per judge |

## Architecture

```
+------------------------------------------------------------------+
|  Judge (independent or with perspective)                          |
|                                                                   |
|  1. Receive packet + perspective                                  |
|  2. Identify N sub-questions to explore                           |
|  3. Spawn N explorers in parallel (Task tool, background)         |
|  4. Collect explorer results                                      |
|  5. Synthesize into final judge response                          |
+------------------------------------------------------------------+
        |              |              |
        v              v              v
  +----------+  +----------+  +----------+
  |Explorer 1|  |Explorer 2|  |Explorer 3|
  |Sub-Q: "A"|  |Sub-Q: "B"|  |Sub-Q: "C"|
  |          |  |          |  |          |
  |Codebase  |  |Codebase  |  |Codebase  |
  |search +  |  |search +  |  |search +  |
  |analysis  |  |analysis  |  |analysis  |
  +----------+  +----------+  +----------+
```

**Total agents:** `judges * (1 + explorers)`

**MAX_AGENTS = 12** (hard limit). If total agents (judges x (1 + explorers)) exceeds 12, exit with error: "Error: Total agent count {N} exceeds MAX_AGENTS (12). Reduce --explorers or remove --mixed."

| Example | Judges | Explorers | Total Agents | Status |
|---------|--------|-----------|--------------|--------|
| `$council research X` | 2 | 0 | 2 | Valid |
| `$council --explorers=3 research X` | 2 | 3 | 8 | Valid |
| `$council --deep --explorers=3 research X` | 3 | 3 | 12 | Valid (at cap) |
| `$council --mixed --explorers=3 research X` | 6 | 3 | 24 | BLOCKED (exceeds 12) |
| `$council --mixed research X` | 6 | 0 | 6 | Valid |
| `$council --mixed --explorers=1 research X` | 6 | 1 | 12 | Valid (at cap) |

## Explorer Prompt

```
You are Explorer {M} for Council Judge {N}{PERSPECTIVE_SUFFIX}.
(PERSPECTIVE_SUFFIX is " -- THE {PERSPECTIVE}" when using presets, or empty for independent judges)

## Your Sub-Question

{SUB_QUESTION}

## Context

Working directory: {CWD}
Target: {TARGET}

## Instructions

1. Use available tools (Glob, Grep, Read, Bash) to investigate the sub-question
2. Search the codebase, documentation, and any relevant sources
3. Be thorough -- your findings feed directly into the judge's analysis
4. Return a structured summary:

### Findings
<what you discovered>

### Evidence
<specific files, lines, patterns found>

### Assessment
<your interpretation of the findings>
```

## Explorer Execution

Explorers are spawned as lightweight subagents optimized for search/read (not editing). Each explorer receives a sub-question and returns findings to its parent judge.

**Model selection:** Explorers use `sonnet` by default (fast, good at search). Judges also use `sonnet` by default (use `--profile=thorough` or `COUNCIL_CLAUDE_MODEL=opus` for high-stakes reviews). Override explorer model with `--explorer-model=<model>`.

## Sub-Question Generation

When `--explorers=N` is set, the judge prompt includes:

```
Before analyzing, identify {N} specific sub-questions that would help you
answer thoroughly. For each sub-question, spawn an explorer agent to
investigate it. Use the explorer findings to inform your final analysis.

Sub-questions should be:
- Specific and searchable (not vague)
- Complementary (cover different aspects)
- Relevant to your analysis angle (perspective if assigned, or general if independent)
```

## Timeout

Explorer timeout: 60s (half of judge timeout). Judge timeout starts after all explorers complete.

### model-profiles.md

# Model Quality Profiles

Pre-configured model + judge count combinations for different use cases.

Use `--profile=<name>` to select a profile. Profiles set environment variables before agent spawning.

## Profiles

| Profile | COUNCIL_CLAUDE_MODEL | Judge Count | COUNCIL_TIMEOUT | Use Case |
|---------|---------------------|-------------|-----------------|----------|
| `thorough` | opus | 3 | 120 | Architecture decisions, security audits |
| `balanced` | sonnet | 2 | 120 | Default validation, routine reviews |
| `fast` | haiku | 2 | 60 | Quick checks, mid-implementation sanity |

## Precedence

Profiles are a convenience shortcut. Explicit flags and env vars always override:

1. Explicit env var (`COUNCIL_CLAUDE_MODEL=...`) --- highest priority
2. Explicit flags (`--count=N`, `--deep`, `--mixed`) --- override profile settings
3. `--profile=<name>` --- sets defaults
4. Built-in defaults --- lowest priority

When `--profile=thorough` is set but `--count=4` is also provided, the count comes from `--count` (4 judges), while the model comes from the profile (opus).

## Report Header

When a profile is used, include in the council report header:
```
**Profile:** <name>
```

## Env Var Mapping

Each profile sets these env vars before agent spawning:

```
thorough:
  COUNCIL_CLAUDE_MODEL=opus
  COUNCIL_JUDGE_COUNT=3
  COUNCIL_TIMEOUT=120

balanced:
  COUNCIL_CLAUDE_MODEL=sonnet
  COUNCIL_JUDGE_COUNT=2
  COUNCIL_TIMEOUT=120

fast:
  COUNCIL_CLAUDE_MODEL=haiku
  COUNCIL_JUDGE_COUNT=2
  COUNCIL_TIMEOUT=60
```

### output-format.md

# Output Format

## Council Report (Markdown) — Validate Mode

```markdown
## Council Consensus: WARN

**Target:** Implementation of user authentication
**Modes:** validate, --mixed
**Judges:** 3 Claude (Opus 4.6) + 3 Codex (GPT-5.3-Codex)

---

### Verdicts

| Vendor | Judge 1 | Judge 2 | Judge 3 |
|--------|---------|---------|---------|
| Claude | PASS | WARN | PASS |
| Codex | WARN | WARN | WARN |

*(With `--preset`, column headers reflect perspective names instead of Judge N)*

---

### Shared Findings

| Finding | Severity | Fix | Ref |
|---------|----------|-----|-----|
| JWT implementation follows best practices | minor | No action needed | src/auth/jwt.py:15 |
| Refresh token rotation is correctly implemented | minor | No action needed | src/auth/refresh.py:42 |
| Rate limiting missing on auth endpoints | significant | Add rate limiting middleware to /auth/* routes | OWASP Authentication Cheatsheet |

### Disagreements

| Issue | Claude | Codex |
|-------|--------|-------|
| Rate limiting | Optional for internal APIs | Required per OWASP |
| Token expiry | 1 hour acceptable | Should be 15 minutes |

### Cross-Vendor Insights

**Claude-only:** Noted UX friction in token refresh flow
**Codex-only:** Flagged potential timing attack in token comparison

---

### Recommendation

Add rate limiting to auth endpoints. Consider reducing token expiry to 30 minutes as compromise.

---

*Council completed in 45s. 6/6 judges responded.*
```

## Brainstorm Report

```markdown
## Council Brainstorm: <Topic>

**Target:** <what we're brainstorming>
**Judges:** <count and vendors>

### Options Explored

| Option | Judge 1 | Judge 2 | Judge 3 |
|--------|---------|---------|---------|
| Option A | Assessment | Assessment | Assessment |
| Option B | Assessment | Assessment | Assessment |

### Recommendation

<synthesized recommendation with reasoning>

*Council completed in Ns. N/N judges responded.*
```

**Write to:** `.agents/council/YYYY-MM-DD-brainstorm-<topic>.md`

## Research Report

```markdown
## Council Research: <Topic>

**Target:** <what we're researching>
**Judges:** <count and vendors>

### Facets Explored

Each judge investigated a different aspect of the topic:

| Facet | Judge | Key Findings |
|-------|-------|-------------|
| <aspect 1> | Judge 1 | <summary> |
| <aspect 2> | Judge 2 | <summary> |
| <aspect 3> | Judge 3 | <summary> |

### Synthesized Findings

<merged findings across all judges, organized by theme>

### Open Questions

- <questions that emerged during research>

### Recommendation

<synthesized recommendation with reasoning>

*Council completed in Ns. N/N judges responded.*
```

**Write to:** `.agents/council/YYYY-MM-DD-research-<topic>.md`

## Debate Report Additions

When `--debate` is used, add these sections to any report format:

**Header addition:**
```markdown
**Mode:** {task_type}, --debate
**Rounds:** 2 (independent assessment + adversarial debate)
**Fidelity:** full (native teams -- judges retained full R1 context for R2)
```

If debate ran in fallback mode (re-spawned with truncated R1 verdicts), use instead:
```markdown
**Mode:** {task_type}, --debate
**Rounds:** 2 (independent assessment + adversarial debate)
**Fidelity:** degraded (fallback -- R1 verdicts truncated for R2 re-spawn)
```

**After the Verdicts table, add:**

```markdown
### Verdict Shifts (R1 -> R2)

| Judge | R1 Verdict | R2 Verdict | Changed? | Reason |
|-------|-----------|-----------|----------|--------|
| Judge 1 (or Perspective) | PASS | WARN | Yes | Accepted Judge 2's finding on rate limiting |
| Judge 2 (or Perspective) | WARN | WARN | No | Confirmed after reviewing counterarguments |
| Judge 3 (or Perspective) | PASS | PASS | No | Maintained -- challenged Judge 2's scope concern |

### Debate Notes

**Key Exchanges:**
- **Judge 1 <- Judge 2:** [what was exchanged and its impact]
- **Judge 3 vs Judge 2:** [where they disagreed and why]

**Steel-Man Highlights:**
- Judge 1 steel-manned: "[strongest opposing argument they engaged with]"
- Judge 2 steel-manned: "[strongest opposing argument they engaged with]"
```

**Convergence Detection:**

If Round 1 had at least 2 judges with different verdicts AND Round 2 is unanimous, add this flag:

```markdown
> **Convergence Detected:** Judges who disagreed in Round 1 now agree in Round 2.
> Review debate reasoning to verify this reflects genuine persuasion, not anchoring.
> Round 1 verdicts preserved above for comparison.
```

**Footer update:**
```markdown
*Council completed in {R1_time + R2_time}. {N}/{N} judges responded in R1, {M}/{N} in R2.*
```

### personas.md

# Perspectives

## Default: Independent Judges (No Perspectives)

When no `--preset` or `--perspectives` flag is provided, all judges get the **same prompt** with no perspective label. Diversity comes from independent sampling, not personality labels.

| Judge | Prompt | Assigned To |
|-------|--------|-------------|
| **Judge 1** | Independent judge — same prompt as all others | Agent 1 |
| **Judge 2** | Independent judge — same prompt as all others | Agent 2 |
| **Judge 3** | Independent judge — same prompt as all others | Agent 3 (--deep/--mixed) |

The default judge prompt (no perspective labels):

```
You are Council Judge {N}. You are one of {TOTAL} independent judges evaluating the same target.

{JSON_PACKET}

Instructions:
1. Analyze the target thoroughly
2. Write your analysis to: .agents/council/{OUTPUT_FILENAME}
   - Start with a JSON code block matching the output_schema
   - Follow with Markdown explanation
3. Send verdict to team lead

Your job is to find problems. A PASS with caveats is less valuable than a specific FAIL.
```

When `--preset` or `--perspectives` is used, judges receive the perspective-labeled prompt instead (see Agent Prompts section).

## Custom Perspectives

Simple name-based:
```bash
$council --perspectives="security,performance,ux" validate the API
```

## Built-in Presets

Use `--preset=<name>` for common persona configurations:

| Preset | Perspectives | Best For |
|--------|-------------|----------|
| `default` | (none — independent judges) | General validation |
| `security-audit` | attacker, defender, compliance | Security review |
| `architecture` | scalability, maintainability, simplicity | System design |
| `research` | breadth, depth, contrarian | Deep investigation |
| `ops` | reliability, observability, incident-response | Operations review |
| `code-review` | error-paths, api-surface, spec-compliance | Code validation (used by $vibe) |
| `plan-review` | missing-requirements, feasibility, scope, spec-completeness | Plan validation (used by $pre-mortem) |
| `doc-review` | clarity-editor, accuracy-verifier, completeness-auditor, audience-advocate | Documentation quality review |
| `retrospective` | plan-compliance, tech-debt, learnings | Post-implementation review (used by $post-mortem) |
| `product` | user-value, adoption-barriers, competitive-position | Product-market fit review (used by $pre-mortem when PRODUCT.md exists) |
| `developer-experience` | api-clarity, error-experience, discoverability | Developer UX review (used by $vibe when PRODUCT.md exists) |

```bash
$council --preset=security-audit validate the auth system
$council --preset=research --explorers=3 research upgrade automation
$council --preset=architecture research microservices boundaries
```

**Preset definitions** are built-in perspective configurations.

**Persona name mappings:**

| Preset | Name | Perspective |
|--------|------|-------------|
| security-audit | **Red** | attacker |
| security-audit | **Blue** | defender |
| security-audit | **Auditor** | compliance |
| architecture | **Scale** | scalability |
| architecture | **Craft** | maintainability |
| architecture | **Razor** | simplicity |
| code-review | **Pathfinder** | error-paths |
| code-review | **Surface** | api-surface |
| code-review | **Spec** | spec-compliance |
| plan-review | **Gaps** | missing-requirements |
| plan-review | **Reality** | feasibility |
| plan-review | **Scope** | scope |
| plan-review | **Completeness** | spec-completeness |
| retrospective | **Compass** | plan-compliance |
| retrospective | **Debt** | tech-debt |
| retrospective | **Harvest** | learnings |
| product | **User** | user-value |
| product | **Friction** | adoption-barriers |
| product | **Edge** | competitive-position |
| developer-experience | **Signal** | api-clarity |
| developer-experience | **SOS** | error-experience |
| developer-experience | **Beacon** | discoverability |
| doc-review | **Clarity** | clarity-editor |
| doc-review | **Accuracy** | accuracy-verifier |
| doc-review | **Coverage** | completeness-auditor |
| doc-review | **Audience** | audience-advocate |
| research | **Wide** | breadth |
| research | **Deep** | depth |
| research | **Contrarian** | contrarian |
| ops | **Uptime** | reliability |
| ops | **Lens** | observability |
| ops | **Oncall** | incident-response |

**Preset perspective details:**

```
security-audit:
  attacker:   {name: Red}       "How would I exploit this? What's the weakest link?"
  defender:   {name: Blue}      "How do we detect and prevent attacks? What's our blast radius?"
  compliance: {name: Auditor}   "Does this meet regulatory requirements? What's our audit trail?"

architecture:
  scalability:     {name: Scale}  "Will this handle 10x load? Where are the bottlenecks?"
  maintainability: {name: Craft}  "Can a new engineer understand this in a week? Where's the complexity?"
  simplicity:      {name: Razor}  "What can we remove? Is this the simplest solution?"

research:
  breadth:     {name: Wide}       "What's the full landscape? What options exist? What's adjacent?"
  depth:       {name: Deep}       "What are the deep technical details? What's under the surface?"
  contrarian:  {name: Contrarian} "What's the conventional wisdom wrong about? What's overlooked?"

ops:
  reliability:       {name: Uptime}  "What fails first? What's our recovery time? Where are SPOFs?"
  observability:     {name: Lens}    "Can we see what's happening? What metrics/logs/traces do we need?"
  incident-response: {name: Oncall}  "When this breaks at 3am, what do we need? What's our runbook?"

code-review:
  error-paths:      {name: Pathfinder}  "Trace every error handling path. What's uncaught? What fails silently?"
  api-surface:      {name: Surface}     "Review every public interface. Is the contract clear? Breaking changes?"
  spec-compliance:  {name: Spec}        "Compare implementation against the spec/bead. What's missing? What diverges?"
  # Note: spec-compliance gracefully degrades to general correctness review when no spec
  # is present in context.spec. The judge reviews code on its own merits.

plan-review:
  missing-requirements: {name: Gaps}         "What's not in the spec that should be? What questions haven't been asked?"
  feasibility:          {name: Reality}       "What's technically hard or impossible here? What will take 3x longer than estimated?"
  scope:                {name: Scope}         "What's unnecessary? What's missing? Where will scope creep?"
  spec-completeness:    {name: Completeness}  "Are boundaries defined (Always/Ask First/Never)? Do conformance checks cover all acceptance criteria? Can every acceptance criterion be mechanically verified? Are schema enum values and field names domain-neutral (meaningful in ANY codebase, not just this repo)? Also enforce lifecycle contract completeness: canonical mutation+ack sequence, crash-safe consume protocol with atomic boundary + restart recovery, field-level precedence truth table with anomaly codes, and boundary failpoint conformance tests. Missing/contradictory checklist items are WARN minimum; critical non-mechanically-verifiable invariants are FAIL."

retrospective:
  plan-compliance: {name: Compass}  "What was planned vs what was delivered? What's missing? What was added?"
  tech-debt:       {name: Debt}     "What shortcuts were taken? What will bite us later? What needs cleanup?"
  learnings:       {name: Harvest}  "What patterns emerged? What should be extracted as reusable knowledge?"

product:
  user-value:            {name: User}      "What user problem does this solve? Who benefits and how?"
  adoption-barriers:     {name: Friction}  "What makes this hard to discover, learn, or use? What's the friction?"
  competitive-position:  {name: Edge}      "How does this compare to alternatives? What's our differentiation?"

doc-review:
  clarity-editor:       {name: Clarity}   "Is every sentence unambiguous? Can a reader understand without re-reading? Where's the jargon?"
  accuracy-verifier:    {name: Accuracy}  "Do code examples match the actual API? Are version numbers current? Do links resolve?"
  completeness-auditor: {name: Coverage}  "What's documented but not explained? What's missing entirely? Are edge cases covered?"
  audience-advocate:    {name: Audience}  "Who is the reader? Is the assumed knowledge level consistent? Would a newcomer get lost?"

developer-experience:
  api-clarity:     {name: Signal}  "Is every public interface self-documenting? Can a user predict behavior from names alone?"
  error-experience: {name: SOS}    "When something goes wrong, does the user know what happened, why, and what to do next?"
  discoverability: {name: Beacon}  "Can a new user find this feature without reading docs? Is the happy path obvious?"
```

### presets.md

# Built-in Presets

Presets are covered in `personas.md`. This file exists as a redirect.

**For all preset definitions and perspective details, read `skills/council/references/personas.md`.**

### quick-mode.md

# Quick Mode (`--quick`)

Single-agent inline validation. No subprocess spawning, no Task tool, no Codex. The current agent performs a structured self-review using the same output schema as a full council.

**When to use:** Routine checks, mid-implementation sanity checks, pre-commit quick scan. Use full council for important decisions, final reviews, or when cross-perspective disagreement is valuable.

## Quick Mode Execution

1. **Gather context** (same as full council -- read target files, get diffs)
2. **Skip agent spawning** -- no Task tool, no background agents
3. **Perform structured self-review inline** using this template:

```
Analyze the target as a single independent reviewer.

Target: {TARGET_DESCRIPTION}

Context:
{FILES_AND_DIFFS}

Respond with:
1. A JSON block matching the council output_schema:
   {
     "verdict": "PASS | WARN | FAIL",
     "confidence": "HIGH | MEDIUM | LOW",
     "key_insight": "Single sentence summary",
     "findings": [
       {
         "severity": "critical | significant | minor",
         "category": "security | architecture | performance | style",
         "description": "What was found",
         "location": "file:line if applicable",
         "recommendation": "How to address"
       }
     ],
     "recommendation": "Concrete next step"
   }
2. A brief Markdown explanation (2-5 paragraphs max)
```

4. **Write report** to `.agents/council/YYYY-MM-DD-quick-<target>.md`
5. **Label clearly** as `Mode: quick (single-agent)` in the report header

## Quick Mode Report Format

```markdown
# Council Quick Check: <target>

**Date:** YYYY-MM-DD
**Mode:** quick (single-agent, no multi-perspective spawning)
**Target:** <description>

## Verdict: PASS | WARN | FAIL

<JSON block>

## Analysis

<Brief explanation>

---
*Quick check -- for thorough multi-perspective review, run `$council validate` (default mode).*
```

## Quick Mode Limitations

- No cross-perspective disagreement (single viewpoint)
- No cross-vendor insights (no Codex)
- Lower confidence ceiling than full council
- Not suitable for security audits or architecture decisions -- use `--deep` or `--mixed` for those

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
| Fresh context per loop | New workers/teams per wave in `$swarm`; fresh phase context in `ao rpi phased` |
| Main context as scheduler | Mayor/lead orchestration in `$swarm` and `$crank` |
| Plan file as state | `bd` issue graph, TaskList state, plan artifacts in `.agents/plans/` |
| One task per pass | One issue per worker assignment in swarm/crank waves |
| Backpressure | `$vibe`, task validation hooks, tests/lint gates, push/pre-mortem gates |
| Outer loop restart | Wave loop in `$crank`; phase loop in `ao rpi phased` |

## Implementation Notes

- Keep worker prompts concise and operational.
- Keep state in files/issue trackers, not long conversational memory.
- Prefer deterministic checks over subjective completion.


---

## Scripts

### validate-council.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# Council Validation - Packet structure, config, and output checks
# Validates council output directory contents and SKILL.md references
#
# Usage: validate-council.sh [council-output-dir]
# Exit 0 on pass, non-zero on failure

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$SKILL_DIR/../.." && pwd)"

# Validate output directory argument to prevent argument injection
OUTPUT_DIR="${1:-.agents/council}"
if [[ "$OUTPUT_DIR" =~ ^- ]]; then
    echo "Error: OUTPUT_DIR cannot start with a dash (prevents argument injection)" >&2
    exit 1
fi

DATE=$(date +%Y-%m-%d)
PASS=0
FAIL=0
WARN=0

pass() {
    echo "PASS: $1"
    PASS=$((PASS + 1))
}

fail() {
    echo "FAIL: $1"
    FAIL=$((FAIL + 1))
}

warn() {
    echo "WARN: $1"
    WARN=$((WARN + 1))
}

# ── Section 1: SKILL.md structural checks ──────────────────────────

echo "=== SKILL.md Structure ==="

if [[ -f "$SKILL_DIR/SKILL.md" ]]; then
    pass "SKILL.md exists"
else
    fail "SKILL.md missing"
fi

if head -1 "$SKILL_DIR/SKILL.md" 2>/dev/null | grep -q '^---$'; then
    pass "SKILL.md has YAML frontmatter"
else
    fail "SKILL.md missing YAML frontmatter"
fi

# ── Section 2: Reference file reachability ──────────────────────────

echo ""
echo "=== Reference File Reachability ==="

# Extract all references/ paths mentioned in SKILL.md
REFERENCED_FILES=$(grep -oE 'references/[a-zA-Z0-9_-]+\.md' "$SKILL_DIR/SKILL.md" 2>/dev/null | sort -u || true)

if [[ -z "$REFERENCED_FILES" ]]; then
    warn "No reference files mentioned in SKILL.md"
else
    while IFS= read -r ref; do
        ref_path="$SKILL_DIR/$ref"
        if [[ -f "$ref_path" ]]; then
            pass "Referenced file exists: $ref"
        else
            fail "Referenced file missing: $ref (mentioned in SKILL.md)"
        fi
    done <<< "$REFERENCED_FILES"
fi

# Check for orphaned reference files (exist on disk but not mentioned in SKILL.md)
if [[ -d "$SKILL_DIR/references" ]]; then
    for ref_file in "$SKILL_DIR"/references/*.md; do
        [[ -f "$ref_file" ]] || continue
        ref_basename="references/$(basename "$ref_file")"
        if echo "$REFERENCED_FILES" | grep -qF "$ref_basename"; then
            pass "Reference file reachable from SKILL.md: $ref_basename"
        else
            warn "Orphaned reference file (not mentioned in SKILL.md): $ref_basename"
        fi
    done
fi

# ── Section 3: Files referenced in SKILL.md that should exist ──────

echo ""
echo "=== External File References ==="

# Check schemas referenced in SKILL.md
SCHEMA_REFS=$(grep -oE 'schemas/[a-zA-Z0-9_-]+\.json' "$SKILL_DIR/SKILL.md" 2>/dev/null | sort -u || true)
if [[ -n "$SCHEMA_REFS" ]]; then
    while IFS= read -r schema; do
        schema_path="$SKILL_DIR/$schema"
        if [[ -f "$schema_path" ]]; then
            pass "Schema file exists: $schema"
        else
            fail "Schema file missing: $schema (referenced in SKILL.md)"
        fi
    done <<< "$SCHEMA_REFS"
fi

# Check skill cross-references (skills/*/SKILL.md)
SKILL_REFS=$(grep -oE 'skills/[a-zA-Z0-9_-]+/SKILL\.md' "$SKILL_DIR/SKILL.md" 2>/dev/null | sort -u || true)
if [[ -n "$SKILL_REFS" ]]; then
    while IFS= read -r skill_ref; do
        skill_path="$REPO_ROOT/$skill_ref"
        if [[ -f "$skill_path" ]]; then
            pass "Cross-referenced skill exists: $skill_ref"
        else
            fail "Cross-referenced skill missing: $skill_ref"
        fi
    done <<< "$SKILL_REFS"
fi

# ── Section 4: Judge count by mode ─────────────────────────────────

echo ""
echo "=== Judge Count Validation ==="

# Validate documented judge counts against SKILL.md mode table
# Expected: default=2, --deep=3, --mixed=3+3=6
SKILL_CONTENT=$(cat "$SKILL_DIR/SKILL.md" 2>/dev/null || true)

# Check default mode documents 2 agents
if echo "$SKILL_CONTENT" | grep -qE 'default.*\|.*2.*\|'; then
    pass "Default mode documents 2 judges"
else
    fail "Default mode should document 2 judges"
fi

# Check --deep mode documents 3 agents
if echo "$SKILL_CONTENT" | grep -qE '\-\-deep.*\|.*3.*\|'; then
    pass "--deep mode documents 3 judges"
else
    fail "--deep mode should document 3 judges"
fi

# Check --mixed mode documents 3+3 agents
if echo "$SKILL_CONTENT" | grep -qE '\-\-mixed.*\|.*3\+3'; then
    pass "--mixed mode documents 3+3 judges"
else
    fail "--mixed mode should document 3+3 judges"
fi

# ── Section 5: Output file naming convention ────────────────────────

echo ""
echo "=== Output File Naming Convention ==="

# Expected pattern: YYYY-MM-DD-<type>-<target>.md
# Per SKILL.md: .agents/council/YYYY-MM-DD-<type>-<target>.md
NAMING_PATTERN='YYYY-MM-DD-<type>-<target>'
if echo "$SKILL_CONTENT" | grep -qF "$NAMING_PATTERN"; then
    pass "Naming convention documented: $NAMING_PATTERN"
else
    fail "Naming convention not documented in SKILL.md"
fi

# If the output directory exists, validate actual files match the convention
if [[ -d "$REPO_ROOT/$OUTPUT_DIR" ]]; then
    BAD_NAMES=0
    CHECKED_FILES=0
    for council_file in "$REPO_ROOT/$OUTPUT_DIR"/*.md; do
        [[ -f "$council_file" ]] || continue
        fname=$(basename "$council_file")
        CHECKED_FILES=$((CHECKED_FILES + 1))

        # Validate: YYYY-MM-DD-<type>-<target>.md
        # Core council types: validate, brainstorm, research, quick
        # Wrapper skill types: vibe, pre-mortem, postmortem, council (from wrapper skills)
        # Suffixes: optional -<vendor>-<perspective-or-id> for per-judge files
        #           optional -report for consolidated reports
        #           optional -judge-N for numbered judges
        VALID_TYPES="validate|brainstorm|research|quick|vibe|pre-mortem|postmortem|council|analyze|beads|consistency|final|justify|native|release|skills|codex|debate|adoption"
        if [[ "$fname" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}-[a-zA-Z0-9_-]+\.md$ ]]; then
            # Broad date-prefixed naming accepted; strict type check is a warning
            if [[ "$fname" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}-(${VALID_TYPES})-[a-zA-Z0-9_-]+\.md$ ]]; then
                pass "File matches naming convention: $fname"
            else
                warn "File is date-prefixed but uses non-standard type: $fname"
            fi
        else
            fail "File does not match naming convention (missing YYYY-MM-DD prefix): $fname"
            BAD_NAMES=$((BAD_NAMES + 1))
        fi
    done

    if [[ $CHECKED_FILES -eq 0 ]]; then
        warn "No .md files found in $OUTPUT_DIR to validate"
    fi
else
    warn "Output directory $OUTPUT_DIR does not exist (skipping file name checks)"
fi

# ── Section 6: Perspective uniqueness ───────────────────────────────

echo ""
echo "=== Perspective Uniqueness ==="

# Check that built-in presets define unique perspectives
if [[ -f "$SKILL_DIR/references/personas.md" ]]; then
    # Extract perspective names from personas.md (lines starting with ### or | name |)
    PERSPECTIVES=$(grep -E '^\| \*\*' "$SKILL_DIR/references/personas.md" 2>/dev/null | \
        sed 's/| \*\*\([^*]*\)\*\*.*/\1/' | sort || true)

    if [[ -n "$PERSPECTIVES" ]]; then
        UNIQUE_PERSPECTIVES=$(echo "$PERSPECTIVES" | sort -u)
        TOTAL=$(echo "$PERSPECTIVES" | wc -l | tr -d ' ')
        UNIQUE=$(echo "$UNIQUE_PERSPECTIVES" | wc -l | tr -d ' ')

        if [[ "$TOTAL" -eq "$UNIQUE" ]]; then
            pass "All $TOTAL perspective names in personas.md are unique"
        else
            DUPES=$(echo "$PERSPECTIVES" | sort | uniq -d)
            fail "Duplicate perspectives found in personas.md: $DUPES"
        fi
    else
        warn "Could not extract perspective names from personas.md"
    fi

    # Check presets.md for duplicate perspectives within each preset
    if [[ -f "$SKILL_DIR/references/presets.md" ]]; then
        PRESET_SECTIONS=$(grep -n '^### ' "$SKILL_DIR/references/presets.md" 2>/dev/null || true)
        if [[ -n "$PRESET_SECTIONS" ]]; then
            pass "Presets reference file exists with sections"
        else
            warn "No preset sections found in presets.md"
        fi
    else
        warn "presets.md not found (perspective preset validation skipped)"
    fi
else
    warn "personas.md not found (perspective uniqueness check skipped)"
fi

# If output directory has council files, check for duplicate perspectives in packets
if [[ -d "$REPO_ROOT/$OUTPUT_DIR" ]]; then
    for council_file in "$REPO_ROOT/$OUTPUT_DIR"/*.md; do
        [[ -f "$council_file" ]] || continue
        # Extract perspective assignments from council reports
        ASSIGNED=$(grep -oE 'Perspective:.*' "$council_file" 2>/dev/null | \
            sed 's/Perspective: *//' | sort || true)
        if [[ -n "$ASSIGNED" ]]; then
            fname=$(basename "$council_file")
            TOTAL_P=$(echo "$ASSIGNED" | wc -l | tr -d ' ')
            UNIQUE_P=$(echo "$ASSIGNED" | sort -u | wc -l | tr -d ' ')
            if [[ "$TOTAL_P" -eq "$UNIQUE_P" ]]; then
                pass "No duplicate perspectives in $fname"
            else
                DUPES_P=$(echo "$ASSIGNED" | sort | uniq -d)
                fail "Duplicate perspective assignment in $fname: $DUPES_P"
            fi
        fi
    done
fi

# ── Section 7: Consensus rules documented ──────────────────────────

echo ""
echo "=== Consensus Rules ==="

if echo "$SKILL_CONTENT" | grep -q 'All PASS.*PASS'; then
    pass "Consensus rule: All PASS -> PASS documented"
else
    fail "Missing consensus rule: All PASS -> PASS"
fi

if echo "$SKILL_CONTENT" | grep -q 'Any FAIL.*FAIL'; then
    pass "Consensus rule: Any FAIL -> FAIL documented"
else
    fail "Missing consensus rule: Any FAIL -> FAIL"
fi

if echo "$SKILL_CONTENT" | grep -qE 'Mixed.*WARN|DISAGREE'; then
    pass "Consensus rule: Mixed/disagreement handling documented"
else
    fail "Missing consensus rule: disagreement handling"
fi

# ── Summary ─────────────────────────────────────────────────────────

echo ""
echo "=== Summary ==="
echo ""
echo "| Result | Count |"
echo "|--------|-------|"
echo "| PASS   | $PASS |"
echo "| FAIL   | $FAIL |"
echo "| WARN   | $WARN |"
echo ""

if [[ $FAIL -gt 0 ]]; then
    echo "Status: FAILED ($FAIL failures)"
    exit 1
elif [[ $WARN -gt 0 ]]; then
    echo "Status: PASS with warnings ($WARN warnings)"
    exit 0
else
    echo "Status: PASS (all checks passed)"
    exit 0
fi
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
check "SKILL.md has name: council" "grep -q '^name: council' '$SKILL_DIR/SKILL.md'"
check "schemas/ directory exists" "[ -d '$SKILL_DIR/schemas' ]"
check "verdict.json exists" "[ -f '$SKILL_DIR/schemas/verdict.json' ]"
check "verdict.json is valid JSON" "python3 -m json.tool '$SKILL_DIR/schemas/verdict.json' >/dev/null 2>&1"
check "verdict.json has verdict field" "grep -q '\"verdict\"' '$SKILL_DIR/schemas/verdict.json'"
check "verdict.json has confidence field" "grep -q '\"confidence\"' '$SKILL_DIR/schemas/verdict.json'"
check "verdict.json has key_insight field" "grep -q '\"key_insight\"' '$SKILL_DIR/schemas/verdict.json'"
check "verdict.json has findings field" "grep -q '\"findings\"' '$SKILL_DIR/schemas/verdict.json'"
check "verdict.json has recommendation field" "grep -q '\"recommendation\"' '$SKILL_DIR/schemas/verdict.json'"
check "verdict.json has additionalProperties:false at root" "python3 -c \"import json,sys; d=json.load(open('$SKILL_DIR/schemas/verdict.json')); sys.exit(0 if d.get('additionalProperties') == False else 1)\""
check "verdict.json has additionalProperties:false in findings items" "python3 -c \"import json,sys; d=json.load(open('$SKILL_DIR/schemas/verdict.json')); sys.exit(0 if d['properties']['findings']['items'].get('additionalProperties') == False else 1)\""
check "references/ has at least 5 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 5 ]"
check "SKILL.md mentions default mode" "grep -q 'default' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --deep mode" "grep -q '\-\-deep' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --mixed mode" "grep -q '\-\-mixed' '$SKILL_DIR/SKILL.md'"
check "Output directory pattern documented" "grep -q '\.agents/council/' '$SKILL_DIR/SKILL.md'"
# Behavioral contracts: verify key features remain documented
check "SKILL.md mentions --debate mode" "grep -q '\-\-debate' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --quick mode" "grep -q '\-\-quick' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions verdict field" "grep -q 'verdict' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions PASS/WARN/FAIL" "grep -q 'PASS.*WARN.*FAIL\|PASS | WARN | FAIL' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```
