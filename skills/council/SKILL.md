---
name: council
tier: orchestration
description: 'Multi-model validation council. Spawns parallel judges with different perspectives, consolidates into consensus. Modes: validate, brainstorm, critique. Triggers: council, validate, brainstorm, critique, multi-model, consensus.'
dependencies:
  - vibe        # optional - can run vibe first for toolchain validation
  - standards   # optional - loaded for code validation context
replaces: judge
---

# /council — Multi-Model Validation Council

Spawn parallel judges with different perspectives, consolidate into consensus verdict.

## Quick Start

```bash
/council validate this plan
/council brainstorm caching approaches
/council critique the implementation
/council                              # infers from context
```

## Modes

| Mode | Agents | Vendors | Use Case |
|------|--------|---------|----------|
| default | 2 | Claude | Quick validation |
| `--deep` | 3 | Claude | Thorough review |
| `--mixed` | 3+3 | Claude + Codex | Cross-vendor consensus |

```bash
/council recent                    # 2 Claude agents
/council --deep recent             # 3 Claude agents
/council --mixed recent            # 3 Claude + 3 Codex
```

## Task Types

| Type | Trigger Words | Perspective Focus |
|------|---------------|-------------------|
| **validate** | validate, check, review, assess | Is this correct? What's wrong? |
| **brainstorm** | brainstorm, explore, options, approaches | What are the alternatives? Pros/cons? |
| **critique** | critique, feedback, improve | What could be better? |

Natural language works — the skill infers task type from your prompt.

---

## Architecture

### Execution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Phase 1: Build Packet (JSON)                                   │
│  - Task type (validate/brainstorm/critique)                     │
│  - Target description                                           │
│  - Context (files, diffs, prior decisions)                      │
│  - Perspectives to assign                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┴─────────────────┐
            ▼                                   ▼
┌───────────────────────┐           ┌───────────────────────┐
│     CLAUDE AGENTS     │           │     CODEX AGENTS      │
│  (Task tool, parallel)│           │  (Bash tool, parallel)│
│                       │           │                       │
│  Agent 1: Pragmatist  │           │  Agent 1: Pragmatist  │
│  Agent 2: Skeptic     │           │  Agent 2: Skeptic     │
│  Agent 3: Visionary   │           │  Agent 3: Visionary   │
│  (--deep/--mixed only)│           │  (--mixed only)       │
│                       │           │                       │
│  Output: JSON + MD    │           │  Output: JSON + MD    │
│  Files: /tmp/council- │           │  Files: /tmp/council- │
│         claude-*.json │           │         codex-*.json  │
└───────────────────────┘           └───────────────────────┘
            │                                   │
            └─────────────────┬─────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Phase 2: Consolidation (Spawner)                               │
│  - Read all agent outputs                                       │
│  - Compute consensus verdict                                    │
│  - Identify shared findings                                     │
│  - Surface disagreements with attribution                       │
│  - Generate Markdown report for human                           │
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
| All Codex agents fail | Proceed Claude-only, note degradation |
| All agents fail | Return error, suggest retry |

Timeout: 120s per agent (configurable via `--timeout`).

---

## Packet Format (JSON)

The packet sent to each agent:

```json
{
  "council_packet": {
    "version": "1.0",
    "mode": "validate",
    "target": "Implementation of user authentication system",
    "context": {
      "files": [
        "src/auth/jwt.py",
        "src/auth/middleware.py",
        "tests/test_auth.py"
      ],
      "diff": "git diff output if applicable",
      "prior_decisions": [
        "Using JWT, not sessions",
        "Refresh tokens required"
      ]
    },
    "perspective": "skeptic",
    "perspective_description": "What could go wrong? What's over-engineered? Where will this break?",
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
          "recommendation": "How to address"
        }
      ],
      "recommendation": "Concrete next step"
    }
  }
}
```

---

## Perspectives

### Default Perspectives

| Perspective | Focus | Assigned To |
|-------------|-------|-------------|
| **pragmatist** | What's simplest? What can we cut? Implementation risk. | Agent 1 |
| **skeptic** | What could go wrong? Failure modes. Over-engineering. | Agent 2 |
| **visionary** | Where does this lead? 10x version. Missing pieces. | Agent 3 (--deep/--mixed) |

### Custom Perspectives

```bash
/council --perspectives="security,performance,ux" validate the API
```

---

## Agent Prompts

### Judge Agent Prompt

```
You are Council Member {N} — THE {PERSPECTIVE}.

{JSON_PACKET}

Your angle: {PERSPECTIVE_DESCRIPTION}

Instructions:
1. Analyze the target from your perspective
2. Respond with JSON matching the output_schema
3. Then provide a Markdown explanation

Your response must start with a JSON code block, followed by Markdown analysis.
```

### Consolidation Prompt

```
You are the Council Chairman.

You have received {N} judge reports from {VENDORS}.

## Judge Reports

{JUDGE_OUTPUTS_JSON}

## Your Task

Synthesize into a final council report:

1. **Consensus Verdict**: PASS if all PASS, FAIL if any FAIL, else WARN
2. **Shared Findings**: Points all judges agree on
3. **Disagreements**: Where judges differ (with attribution)
4. **Cross-Vendor Insights**: (if --mixed) Unique findings per vendor
5. **Final Recommendation**: Concrete next step

Output format: Markdown report for human consumption.
```

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

---

## Output Format

### Council Report (Markdown)

```markdown
## Council Consensus: WARN

**Target:** Implementation of user authentication
**Modes:** validate, --mixed
**Judges:** 3 Claude (Opus) + 3 Codex (GPT-5.2)

---

### Verdicts

| Vendor | Pragmatist | Skeptic | Visionary |
|--------|------------|---------|-----------|
| Claude | PASS | WARN | PASS |
| Codex | WARN | WARN | WARN |

---

### Shared Findings

- JWT implementation follows best practices
- Refresh token rotation is correctly implemented
- Test coverage is adequate

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

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COUNCIL_TIMEOUT` | 120000 | Agent timeout in ms |
| `COUNCIL_CODEX_MODEL` | gpt-5.2 | Default Codex model for --mixed |
| `COUNCIL_CLAUDE_MODEL` | opus | Claude model for agents |

### Flags

| Flag | Description |
|------|-------------|
| `--deep` | 3 Claude agents instead of 2 |
| `--mixed` | Add 3 Codex agents |
| `--timeout=N` | Override timeout (seconds) |
| `--perspectives="a,b,c"` | Custom perspectives |
| `--count=N` | Override agent count |

---

## CLI Spawning Commands

### Claude Agents (via Claude Code CLI)

**Using Task tool (inside Claude Code):**

```
Task(
  description="Council judge: Pragmatist",
  subagent_type="general-purpose",
  model="opus",
  run_in_background=true,
  prompt="{JUDGE_PACKET}"
)
```

**Using claude CLI directly:**

```bash
# Spawn a Claude session with a prompt
claude --print "{JUDGE_PACKET}" > /tmp/council-claude-pragmatist.md

# Or interactive with file output
claude -p "{JUDGE_PACKET}" --output-file /tmp/council-claude-pragmatist.md
```

### Codex Agents (via Codex CLI)

**Using Bash tool (inside Claude Code):**

```bash
codex exec --full-auto -C "$(pwd)" -o /tmp/council-codex-{perspective}.md "{PACKET}"
```

**Codex CLI flags:**
- `--full-auto` — No approval prompts
- `-C <dir>` — Working directory
- `-o <file>` — Output file
- `-m <model>` — Model override (default: gpt-5.2)

**Using codex CLI directly:**

```bash
# Spawn Codex with a prompt
codex exec -m gpt-5.2 --full-auto -o /tmp/council-codex-pragmatist.md "You are THE PRAGMATIST..."
```

### Parallel Spawning

**Spawn all agents in parallel:**

```
# Claude agents (Task tool, parallel)
Task(description="Judge 1", ..., run_in_background=true)
Task(description="Judge 2", ..., run_in_background=true)
Task(description="Judge 3", ..., run_in_background=true)

# Codex agents (Bash tool, parallel)
Bash(command="codex exec ... -o /tmp/codex-1.md ...", run_in_background=true)
Bash(command="codex exec ... -o /tmp/codex-2.md ...", run_in_background=true)
Bash(command="codex exec ... -o /tmp/codex-3.md ...", run_in_background=true)
```

**Wait for completion:**

```
TaskOutput(task_id="...", block=true)
```

### Model Selection

| Vendor | Default | Override |
|--------|---------|----------|
| Claude | opus | `--claude-model=sonnet` |
| Codex | gpt-5.2 | `--codex-model=o3` |

### Output Collection

Both CLIs write to files. Read files after completion:

```bash
# Claude output
cat /tmp/council-claude-pragmatist.md

# Codex output
cat /tmp/council-codex-pragmatist.md
```

---

## Examples

### Validate Recent Changes

```bash
/council validate recent
```

Infers: validate mode, target = recent commits, 2 Claude agents.

### Deep Architecture Review

```bash
/council --deep validate the authentication system
```

3 Claude agents (pragmatist, skeptic, visionary) analyze auth implementation.

### Cross-Vendor Validation

```bash
/council --mixed validate this plan
```

3 Claude + 3 Codex agents, cross-vendor synthesis.

### Brainstorm Approaches

```bash
/council brainstorm caching strategies for the API
```

2 Claude agents explore options, pros/cons, recommend one.

### Critique a Spec

```bash
/council critique the implementation plan in PLAN.md
```

2 Claude agents provide feedback on the plan.

---

## Migration from /judge

`/council` replaces `/judge`. Migration:

| Old | New |
|-----|-----|
| `/judge recent` | `/council validate recent` |
| `/judge 2 opus` | `/council recent` (default) |
| `/judge 3 opus` | `/council --deep recent` |

The `/judge` skill is deprecated. Use `/council`.

---

## See Also

- `skills/vibe/SKILL.md` — Toolchain validation (run before council for code)
- `skills/swarm/SKILL.md` — Multi-agent orchestration
- `skills/standards/SKILL.md` — Language-specific coding standards
