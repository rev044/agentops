---
name: council
tier: orchestration
description: 'Multi-model consensus council for validation, research, analysis, brainstorming, and critique. Spawns parallel judges with configurable perspectives and optional explorer sub-agents. Modes: validate, brainstorm, critique, research, analyze. Triggers: council, validate, brainstorm, critique, research, analyze, multi-model, consensus.'
dependencies:
  - standards   # optional - loaded for code validation context
replaces: judge
---

# /council — Multi-Model Consensus Council

Spawn parallel judges with different perspectives, consolidate into consensus. Works for any task — validation, research, analysis, brainstorming, critique.

## Quick Start

```bash
/council validate this plan                                    # validation
/council brainstorm caching approaches                         # brainstorm
/council critique the implementation                           # critique
/council research kubernetes upgrade strategies                # research
/council analyze the CI/CD pipeline bottlenecks                # analysis
/council --preset=security-audit validate the auth system      # preset personas
/council --deep --explorers=3 research upgrade automation      # deep + explorers
/council                                                       # infers from context
```

## Use Cases

Council is a general-purpose multi-model consensus tool. Use it for:

| Use Case | Example | Recommended Mode |
|----------|---------|-----------------|
| Code review | `/council validate recent` | validate |
| Plan validation | `/council validate the migration plan` | validate |
| Architecture analysis | `/council --preset=architecture analyze microservices boundaries` | analyze |
| Deep codebase research | `/council --deep --explorers=3 research the auth system` | research |
| Decision making | `/council brainstorm caching strategies` | brainstorm |
| Risk assessment | `/council --preset=ops critique the deployment pipeline` | critique |
| Security audit | `/council --preset=security-audit validate the API` | validate |
| Spec feedback | `/council critique the design doc` | critique |
| Technology comparison | `/council analyze Redis vs Memcached for our use case` | analyze |
| Incident investigation | `/council --deep research why deployments are slow` | research |

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
| **research** | research, investigate, deep dive, explore deeply | What can we discover? Each judge explores a different facet. |
| **analyze** | analyze, assess, examine, evaluate, compare | What are the properties, trade-offs, and structure? |

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
│  Files: .agents/      │           │  Files: .agents/      │
│    council/claude-*   │           │    council/codex-*    │
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
| Codex CLI not installed | Skip Codex agents, Claude-only (warn user) |
| Output dir missing | Create `.agents/council/` automatically |

Timeout: 120s per agent (configurable via `--timeout=N` in seconds).

**Minimum quorum:** At least 1 agent must respond for a valid council. If 0 agents respond, return error.

### Pre-Flight Checks

Before spawning agents, verify tools are available:

```bash
# Always available (Task tool is built-in)
# Claude agents: no pre-flight needed

# Codex agents (--mixed only)
if ! which codex > /dev/null 2>&1; then
  echo "⚠️ Codex CLI not found. Falling back to Claude-only."
  # Downgrade --mixed to --deep (3 Claude agents)
fi

# Create output directory
mkdir -p .agents/council
```

---

## Packet Format (JSON)

The packet sent to each agent. **File contents are included inline** — agents receive the actual code/plan text in the packet, not just paths. This ensures both Claude and Codex agents can analyze without needing file access.

```json
{
  "council_packet": {
    "version": "1.0",
    "mode": "validate | brainstorm | critique | research | analyze",
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

Simple name-based:
```bash
/council --perspectives="security,performance,ux" validate the API
```

### Persona Definitions

For richer control, use `--perspectives-file=<path>` pointing to a JSON file with full persona definitions:

```json
[
  {
    "name": "security-auditor",
    "focus": "OWASP top 10, authentication flows, data exposure",
    "instructions": "Analyze from the perspective of a security auditor performing a pentest",
    "explore_questions": [
      "What attack vectors exist?",
      "Where is input validation missing?",
      "Are secrets properly managed?"
    ]
  },
  {
    "name": "performance-engineer",
    "focus": "Latency, throughput, resource usage, caching",
    "instructions": "Analyze from the perspective of a performance engineer under load",
    "explore_questions": [
      "Where are the hot paths?",
      "What queries could be slow at scale?",
      "Where is caching missing or stale?"
    ]
  }
]
```

**Fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Short identifier for the perspective |
| `focus` | yes | Comma-separated focus areas injected into judge prompt |
| `instructions` | no | Full custom instruction override for the judge |
| `explore_questions` | no | Seed questions for `--explorers` (overrides auto-generation) |

### Built-in Presets

Use `--preset=<name>` for common persona configurations:

| Preset | Perspectives | Best For |
|--------|-------------|----------|
| `default` | pragmatist, skeptic, visionary | General validation |
| `security-audit` | attacker, defender, compliance | Security review |
| `architecture` | scalability, maintainability, simplicity | System design |
| `research` | breadth, depth, contrarian | Deep investigation |
| `ops` | reliability, observability, incident-response | Operations review |

```bash
/council --preset=security-audit validate the auth system
/council --preset=research --explorers=3 research upgrade automation
/council --preset=architecture analyze microservices boundaries
```

**Preset definitions** are equivalent to built-in perspective files. Custom `--perspectives-file` overrides any preset.

**Preset perspective details:**

```
security-audit:
  attacker:   "How would I exploit this? What's the weakest link?"
  defender:   "How do we detect and prevent attacks? What's our blast radius?"
  compliance: "Does this meet regulatory requirements? What's our audit trail?"

architecture:
  scalability:     "Will this handle 10x load? Where are the bottlenecks?"
  maintainability: "Can a new engineer understand this in a week? Where's the complexity?"
  simplicity:      "What can we remove? Is this the simplest solution?"

research:
  breadth:     "What's the full landscape? What options exist? What's adjacent?"
  depth:       "What are the deep technical details? What's under the surface?"
  contrarian:  "What's the conventional wisdom wrong about? What's overlooked?"

ops:
  reliability:       "What fails first? What's our recovery time? Where are SPOFs?"
  observability:     "Can we see what's happening? What metrics/logs/traces do we need?"
  incident-response: "When this breaks at 3am, what do we need? What's our runbook?"
```

---

## Explorer Sub-Agents

Judges can spawn explorer sub-agents for parallel deep-dive research. This is the key differentiator for `research` and `analyze` modes — massive parallel exploration.

### Flag

| Flag | Default | Max | Description |
|------|---------|-----|-------------|
| `--explorers=N` | 0 | 5 | Number of explorer sub-agents per judge |

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Judge (Pragmatist)                                              │
│                                                                  │
│  1. Receive packet + perspective                                 │
│  2. Identify N sub-questions to explore                          │
│  3. Spawn N explorers in parallel (Task tool, background)        │
│  4. Collect explorer results                                     │
│  5. Synthesize into final judge response                         │
└─────────────────────────────────────────────────────────────────┘
        │              │              │
        ▼              ▼              ▼
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │Explorer 1│  │Explorer 2│  │Explorer 3│
  │Sub-Q: "A"│  │Sub-Q: "B"│  │Sub-Q: "C"│
  │          │  │          │  │          │
  │Codebase  │  │Codebase  │  │Codebase  │
  │search +  │  │search +  │  │search +  │
  │analysis  │  │analysis  │  │analysis  │
  └──────────┘  └──────────┘  └──────────┘
```

**Total agents:** `judges * (1 + explorers)`

| Example | Judges | Explorers | Total Agents |
|---------|--------|-----------|--------------|
| `/council research X` | 2 | 0 | 2 |
| `/council --explorers=3 research X` | 2 | 3 | 2 + 6 = 8 |
| `/council --deep --explorers=3 research X` | 3 | 3 | 3 + 9 = 12 |
| `/council --mixed --explorers=3 research X` | 6 | 3 | 6 + 18 = 24 |

### Explorer Prompt

```
You are Explorer {M} for Council Judge {N} — THE {PERSPECTIVE}.

## Your Sub-Question

{SUB_QUESTION}

## Context

Working directory: {CWD}
Target: {TARGET}

## Instructions

1. Use available tools (Glob, Grep, Read, Bash) to investigate the sub-question
2. Search the codebase, documentation, and any relevant sources
3. Be thorough — your findings feed directly into the judge's analysis
4. Return a structured summary:

### Findings
<what you discovered>

### Evidence
<specific files, lines, patterns found>

### Assessment
<your interpretation of the findings>
```

### Explorer Execution

Explorers are spawned as `Explore`-type subagents for speed:

```
Task(
  description="Explorer for Judge {N}: {SUB_QUESTION_SHORT}",
  subagent_type="Explore",
  model="sonnet",
  run_in_background=true,
  prompt="{EXPLORER_PROMPT}"
)
```

**Model selection:** Explorers use `sonnet` by default (fast, good at search). Judges use `opus` (thorough analysis). Override with `--explorer-model=<model>`.

### Sub-Question Generation

When `--explorers=N` is set, the judge prompt includes:

```
Before analyzing, identify {N} specific sub-questions that would help you
answer from your {PERSPECTIVE} angle. For each sub-question, spawn an
explorer agent to investigate it. Use the explorer findings to inform
your final analysis.

Sub-questions should be:
- Specific and searchable (not vague)
- Complementary (cover different aspects)
- Relevant to your perspective
```

If a persona definition includes `explore_questions`, those are used instead of auto-generated sub-questions.

### Timeout

Explorer timeout: 60s (half of judge timeout). Judge timeout starts after all explorers complete.

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

Synthesize into a final council report.

For validate/critique modes:
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

For analyze mode:
1. **Properties**: Key properties assessed with confidence levels
2. **Trade-offs**: Current vs ideal state gap analysis
3. **Cross-Perspective Synthesis**: Where judges agreed and disagreed
4. **Recommendation**: Concrete next steps

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

**DISAGREE resolution:** When vendors disagree, the spawner presents both positions with reasoning and defers to the user. No automatic tie-breaking — cross-vendor disagreement is a signal worth human attention.

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

### Brainstorm Report

```markdown
## Council Brainstorm: <Topic>

**Target:** <what we're brainstorming>
**Judges:** <count and vendors>

### Options Explored

| Option | Pragmatist | Skeptic | Visionary |
|--------|------------|---------|-----------|
| Option A | Pros/cons | Risks | Potential |
| Option B | Pros/cons | Risks | Potential |

### Recommendation

<synthesized recommendation with reasoning>

*Council completed in Ns. N/N judges responded.*
```

**Write to:** `.agents/council/YYYY-MM-DD-brainstorm-<topic>.md`

### Critique Report

```markdown
## Council Critique: <Target>

**Target:** <what we're critiquing>
**Judges:** <count and vendors>

### Strengths
- ...

### Weaknesses

| Issue | Severity | Source | Recommendation |
|-------|----------|--------|----------------|
| ... | critical/significant/minor | Pragmatist/Skeptic/etc | ... |

### Improvement Roadmap
1. ...

*Council completed in Ns. N/N judges responded.*
```

**Write to:** `.agents/council/YYYY-MM-DD-critique-<topic>.md`

### Research Report

```markdown
## Council Research: <Topic>

**Target:** <what we're researching>
**Judges:** <count and vendors>

### Facets Explored

Each judge investigated a different aspect of the topic:

| Facet | Judge | Key Findings |
|-------|-------|-------------|
| <aspect 1> | Pragmatist | <summary> |
| <aspect 2> | Skeptic | <summary> |
| <aspect 3> | Visionary | <summary> |

### Synthesized Findings

<merged findings across all judges, organized by theme>

### Open Questions

- <questions that emerged during research>

### Recommendation

<synthesized recommendation with reasoning>

*Council completed in Ns. N/N judges responded.*
```

**Write to:** `.agents/council/YYYY-MM-DD-research-<topic>.md`

### Analyze Report

```markdown
## Council Analysis: <Target>

**Target:** <what we're analyzing>
**Judges:** <count and vendors>

### Properties

| Property | Assessment | Confidence |
|----------|-----------|------------|
| <property 1> | <assessment> | HIGH/MEDIUM/LOW |
| <property 2> | <assessment> | HIGH/MEDIUM/LOW |

### Trade-offs

| Dimension | Current State | Ideal State | Gap |
|-----------|--------------|-------------|-----|
| <dim 1> | <current> | <ideal> | <gap> |

### Cross-Perspective Synthesis

<where judges agreed and disagreed in their analysis>

### Recommendation

<concrete next steps based on analysis>

*Council completed in Ns. N/N judges responded.*
```

**Write to:** `.agents/council/YYYY-MM-DD-analyze-<topic>.md`

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COUNCIL_TIMEOUT` | 120 | Agent timeout in seconds |
| `COUNCIL_CODEX_MODEL` | gpt-5.2 | Default Codex model for --mixed |
| `COUNCIL_CLAUDE_MODEL` | opus | Claude model for agents |
| `COUNCIL_EXPLORER_MODEL` | sonnet | Model for explorer sub-agents |
| `COUNCIL_EXPLORER_TIMEOUT` | 60 | Explorer timeout in seconds |

### Flags

| Flag | Description |
|------|-------------|
| `--deep` | 3 Claude agents instead of 2 |
| `--mixed` | Add 3 Codex agents |
| `--timeout=N` | Override timeout in seconds (default: 120) |
| `--perspectives="a,b,c"` | Custom perspective names |
| `--perspectives-file=<path>` | JSON file with full persona definitions |
| `--preset=<name>` | Built-in persona preset (default, security-audit, architecture, research, ops) |
| `--count=N` | Override agent count per vendor (e.g., `--count=4` = 4 Claude, or 4+4 with --mixed) |
| `--explorers=N` | Explorer sub-agents per judge (default: 0, max: 5) |
| `--explorer-model=M` | Override explorer model (default: sonnet) |

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
claude --print "{JUDGE_PACKET}" > .agents/council/claude-pragmatist.md

# Or interactive with file output
claude -p "{JUDGE_PACKET}" --output-file .agents/council/claude-pragmatist.md
```

### Codex Agents (via Codex CLI)

**Canonical Codex command form:**

```bash
codex exec --full-auto -m gpt-5.2 -C "$(pwd)" -o .agents/council/codex-{perspective}.md "{PACKET}"
```

Always use this exact flag order: `--full-auto` → `-m` → `-C` → `-o` → prompt.

**Codex CLI flags (ONLY these are valid):**
- `--full-auto` — No approval prompts (REQUIRED, always first)
- `-m <model>` — Model override (default: gpt-5.2)
- `-C <dir>` — Working directory
- `-o <file>` — Output file (use `-o` not `--output`)

**DO NOT USE:** `-q` (doesn't exist), `--quiet` (doesn't exist)

### Parallel Spawning

**Spawn all agents in parallel:**

```
# Claude agents (Task tool, parallel)
Task(description="Judge 1", ..., run_in_background=true)
Task(description="Judge 2", ..., run_in_background=true)
Task(description="Judge 3", ..., run_in_background=true)

# Codex agents (Bash tool, parallel — canonical flag order)
Bash(command="codex exec --full-auto -m gpt-5.2 -C "$(pwd)" -o .agents/council/codex-pragmatist.md ...", run_in_background=true)
Bash(command="codex exec --full-auto -m gpt-5.2 -C "$(pwd)" -o .agents/council/codex-skeptic.md ...", run_in_background=true)
Bash(command="codex exec --full-auto -m gpt-5.2 -C "$(pwd)" -o .agents/council/codex-visionary.md ...", run_in_background=true)
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

All council outputs go to `.agents/council/`:

```bash
# Ensure directory exists
mkdir -p .agents/council

# Claude output
.agents/council/YYYY-MM-DD-<target>-claude-pragmatist.md

# Codex output
.agents/council/YYYY-MM-DD-<target>-codex-pragmatist.md

# Final consolidated report
.agents/council/YYYY-MM-DD-<target>-report.md
```

---

## Examples

### Validate Recent Changes

```bash
/council validate recent
```

2 Claude agents validate recent commits from pragmatist + skeptic perspectives.

### Deep Architecture Review

```bash
/council --deep --preset=architecture analyze the authentication system
```

3 Claude agents (scalability, maintainability, simplicity) analyze auth design.

### Cross-Vendor Validation

```bash
/council --mixed validate this plan
```

3 Claude + 3 Codex agents, cross-vendor synthesis.

### Deep Research with Explorers

```bash
/council --deep --explorers=3 research upgrade automation patterns
```

3 judges each spawn 3 explorers = 12 parallel research threads. Each judge explores a different facet of the topic with sub-agent support.

### Security Audit

```bash
/council --preset=security-audit --deep validate the API endpoints
```

3 judges (attacker, defender, compliance) review security posture.

### Brainstorm Approaches

```bash
/council brainstorm caching strategies for the API
```

2 Claude agents explore options, pros/cons, recommend one.

### Analyze Trade-offs

```bash
/council analyze Redis vs Memcached for session storage
```

2 judges assess properties, trade-offs, and gaps between options.

### Critique a Spec

```bash
/council critique the implementation plan in PLAN.md
```

2 Claude agents provide structured feedback on the plan.

### Custom Personas from File

```bash
/council --perspectives-file=./my-personas.json validate the migration
```

Load custom judge personas with tailored focus areas and explorer questions.

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

- `skills/vibe/SKILL.md` — Complexity + council for code validation (uses `--preset=default` + validate)
- `skills/pre-mortem/SKILL.md` — Plan validation (uses council validate)
- `skills/post-mortem/SKILL.md` — Work wrap-up (uses council validate + retro)
- `skills/swarm/SKILL.md` — Multi-agent orchestration
- `skills/standards/SKILL.md` — Language-specific coding standards
- `skills/research/SKILL.md` — Codebase exploration (complementary to council research mode)
