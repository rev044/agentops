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
/council --quick validate recent                               # fast inline check
/council validate this plan                                    # validation (2 agents)
/council brainstorm caching approaches                         # brainstorm
/council critique the implementation                           # critique
/council research kubernetes upgrade strategies                # research
/council analyze the CI/CD pipeline bottlenecks                # analysis
/council --preset=security-audit validate the auth system      # preset personas
/council --deep --explorers=3 research upgrade automation      # deep + explorers
/council --debate validate the auth system                # adversarial 2-round review
/council --deep --debate validate the migration plan      # thorough + debate
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
| `--quick` | 0 (inline) | Self | Fast single-agent check, no spawning |
| default | 2 | Claude | Standard multi-perspective validation |
| `--deep` | 3 | Claude | Thorough review |
| `--mixed` | 3+3 | Claude + Codex | Cross-vendor consensus |
| `--debate` | 2+ | Claude | Adversarial refinement (2 rounds) |

```bash
/council --quick validate recent   # inline single-agent check, no spawning
/council recent                    # 2 Claude agents
/council --deep recent             # 3 Claude agents
/council --mixed recent            # 3 Claude + 3 Codex
```

## When to Use `--debate`

Use `--debate` for high-stakes or ambiguous reviews where judges are likely to disagree:
- Security audits, architecture decisions, migration plans
- Reviews where multiple valid perspectives exist
- Cases where a missed finding has real consequences

Skip `--debate` for routine validation where consensus is expected. Debate doubles cost and latency (two rounds of judge spawning).

**Incompatibility:** `--quick` and `--debate` cannot be combined. `--quick` runs inline with no spawning; `--debate` requires multi-agent rounds. If both are passed, exit with error: "Error: --quick and --debate are incompatible."

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
│  - Task type (validate/brainstorm/critique/research/analyze)     │
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

## Quick Mode (`--quick`)

Single-agent inline validation. No subprocess spawning, no Task tool, no Codex. The current agent performs a structured self-review using the same output schema as a full council.

**When to use:** Routine checks, mid-implementation sanity checks, pre-commit quick scan. Use full council for important decisions, final reviews, or when cross-perspective disagreement is valuable.

### Quick Mode Execution

1. **Gather context** (same as full council — read target files, get diffs)
2. **Skip agent spawning** — no Task tool, no background agents
3. **Perform structured self-review inline** using this template:

```
Analyze the target as a single reviewer covering all perspectives (pragmatist + skeptic).

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

### Quick Mode Report Format

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
*Quick check — for thorough multi-perspective review, run `/council validate` (default mode).*
```

### Quick Mode Limitations

- No cross-perspective disagreement (single viewpoint)
- No cross-vendor insights (no Codex)
- Lower confidence ceiling than full council
- Not suitable for security audits or architecture decisions — use `--deep` or `--mixed` for those

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

## Debate Phase (`--debate`)

When `--debate` is passed, council runs two rounds instead of one. Round 1 produces independent verdicts. Round 2 lets judges review each other's work and revise.

### Execution Flow (with --debate)

```
Phase 1: Build Packet + Spawn Round 1 judges (unchanged)
                              │
                    Collect all R1 verdicts
                              │
Phase 1.5: Prepare R2 context (--debate only)
  - For each R1 verdict, extract TRUNCATED summary:
    - JSON verdict block (verdict, confidence, key_insight)
    - Top 3 findings only (by severity)
    - ~1.5K tokens per verdict max
  - Full Markdown verdicts remain in .agents/council/ files
                              │
Phase 2: Spawn Round 2 judges (--debate only)
  - Same perspectives as R1
  - Each R2 judge receives:
    - Original packet (same as R1)
    - Their own R1 verdict (truncated)
    - All other R1 verdicts (truncated)
    - Steel-manning rebuttal prompt
  - Spawn via Task(run_in_background=true) — fresh instances
                              │
                    Collect all R2 verdicts
                              │
Phase 3: Consolidation (uses R2 verdicts when --debate)
```

### Round 2 Spawning

**Branch selection (spawner responsibility):**
```
r1_verdicts = [extract JSON verdict from each R1 output file]
r1_unanimous = all verdicts have same verdict value (PASS/WARN/FAIL)

For each perspective in [pragmatist, skeptic, visionary...]:
  prompt = build_r2_prompt(
    perspective,
    r1_verdicts,
    branch="agreed" if r1_unanimous else "disagreed"
  )
  # Inject ONLY the applicable branch (disagreed OR agreed), not both
  Task(
    description="Council judge R2: {perspective}",
    subagent_type="general-purpose",
    model="opus",
    run_in_background=true,
    prompt=prompt
  )
```

**R2 output files:** Use `-r2` suffix to preserve R1 files:
```
.agents/council/YYYY-MM-DD-<target>-claude-{perspective}-r2.md
```

**With --explorers:** Explorers run in R1 only. R2 judges do not spawn explorers. Explorer findings from R1 are captured in the R1 verdict's findings and key_insight, which are injected into R2 via truncation.

**With --mixed:** Only Claude judges participate in R2. Codex agents run once in R1. For consolidation, use Claude R2 verdicts + Codex R1 verdicts.

### R1 Verdict Truncation

R1 verdicts injected into R2 prompts are truncated to prevent context pressure.

**Extraction:** Read each R1 output file, extract the first JSON code block, then truncate:

```json
{
  "judge": "pragmatist",
  "verdict": "WARN",
  "confidence": "HIGH",
  "key_insight": "Rate limiting missing on auth endpoints",
  "top_findings": [
    {"severity": "significant", "description": "No rate limiting on /login"},
    {"severity": "significant", "description": "JWT expiry too long (1h)"},
    {"severity": "minor", "description": "Missing request ID in error responses"}
  ]
}
```

Full analysis remains in `.agents/council/YYYY-MM-DD-<target>-claude-{perspective}.md` files.

### Timeout and Failure Handling

| Scenario | Behavior |
|----------|----------|
| R2 judge times out | Use their R1 verdict for consolidation |
| All R2 judges time out | Fall back to R1-only consolidation (note in report) |
| R1 judge timed out | No R2 spawn for that perspective (N-1 in R2) |
| Mixed R2 timeout | Consolidate with available R2 verdicts + R1 fallbacks |

### Cost and Latency

`--debate` approximately doubles both:
- **Agents spawned:** N judges x 2 rounds (e.g., --deep = 6 total)
- **Wall time:** R1 time + R2 time (sequential rounds)
- **With --mixed:** Only Claude judges get R2. Codex agents run once (cannot participate in Task-tool debate). For consolidation, use Claude R2 verdicts + Codex R1 verdicts for consensus computation.
- **With --explorers:** Explorers run in R1 only. R2 cost = N judges (no explorer multiplication).
- **Non-verdict modes:** `--debate` is designed for verdict-producing modes (validate, critique). For brainstorm/research/analyze, `--debate` is not recommended and may produce awkward R2 outputs since these modes do not produce PASS/WARN/FAIL verdicts.

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

### Debate Round Judge Prompt (R2)

Used when `--debate` is active. Each R2 judge is a fresh Task instance (not the same agent as R1).

```
You are Council Member {N} — THE {PERSPECTIVE} (Debate Round).

## Prior Assessment from Your Perspective
A prior independent assessment from the {PERSPECTIVE} angle concluded:
{ROUND_1_VERDICT_TRUNCATED_JSON}

## All Round 1 Verdicts
{ALL_ROUND_1_VERDICTS_TRUNCATED}

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

{ORIGINAL_PACKET}

Respond with JSON matching the output_schema. You MUST include the "debate_notes" field:

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
        "target_judge": "skeptic",
        "claim": "what they claimed",
        "response": "why I agree or disagree"
      }
    ],
    "acknowledgments": [
      {
        "source_judge": "pragmatist",
        "point": "what they found",
        "impact": "how it affected my analysis"
      }
    ]
  }
}

Then provide a Markdown explanation of your debate reasoning.
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

### Consolidation Prompt — Debate Additions

When `--debate` is used, append this to the consolidation prompt:

```
## Additional Instructions (Debate Mode)

You have received TWO rounds of judge reports.

Round 1 (independent assessment): Each judge evaluated independently.
Round 2 (post-debate revision): Each judge reviewed all other judges' findings and revised.

When synthesizing:
1. Use Round 2 verdicts for the CONSENSUS VERDICT computation (PASS/WARN/FAIL)
2. Use Round 1 verdicts for FINDING COMPLETENESS — a finding in R1 but dropped in R2 without explanation deserves mention
3. Compare R1 and R2 to identify position shifts
4. Flag judges who changed verdict without citing a specific technical detail, a misinterpretation they corrected, or a finding they missed (possible anchoring)
5. If R1 had at least 2 judges with different verdicts AND R2 is unanimous, note "Convergence detected — review reasoning for anchoring risk"
6. In the report, include the Verdict Shifts table showing R1→R2 changes per judge

When a Round 2 verdict is unavailable (timeout fallback):
- Read the full R1 output file (.agents/council/YYYY-MM-DD-<target>-claude-{perspective}.md)
- Extract the JSON verdict block (first JSON code block in the file)
- Use this as the judge's verdict for consolidation
- Mark in report: "Judge {perspective}: R1 verdict (R2 timeout)"
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
**Judges:** 3 Claude (Opus 4.6) + 3 Codex (GPT-5.3)

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

### Debate Report Additions

When `--debate` is used, add these sections to any report format:

**Header addition:**
```markdown
**Mode:** {task_type}, --debate
**Rounds:** 2 (independent assessment + adversarial debate)
```

**After the Verdicts table, add:**

```markdown
### Verdict Shifts (R1 → R2)

| Judge | R1 Verdict | R2 Verdict | Changed? | Reason |
|-------|-----------|-----------|----------|--------|
| Pragmatist | PASS | WARN | Yes | Accepted Skeptic's finding on rate limiting |
| Skeptic | WARN | WARN | No | Confirmed after reviewing counterarguments |
| Visionary | PASS | PASS | No | Maintained — challenged Skeptic's scope concern |

### Debate Notes

**Key Exchanges:**
- **Pragmatist ← Skeptic:** [what was exchanged and its impact]
- **Visionary vs Skeptic:** [where they disagreed and why]

**Steel-Man Highlights:**
- Pragmatist steel-manned: "[strongest opposing argument they engaged with]"
- Skeptic steel-manned: "[strongest opposing argument they engaged with]"
```

**Convergence Detection:**

If Round 1 had at least 2 judges with different verdicts AND Round 2 is unanimous, add this flag:

```markdown
> **⚠ Convergence Detected:** Judges who disagreed in Round 1 now agree in Round 2.
> Review debate reasoning to verify this reflects genuine persuasion, not anchoring.
> Round 1 verdicts preserved above for comparison.
```

**Footer update:**
```markdown
*Council completed in {R1_time + R2_time}. {N}/{N} judges responded in R1, {M}/{N} in R2.*
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `COUNCIL_TIMEOUT` | 120 | Agent timeout in seconds |
| `COUNCIL_CODEX_MODEL` | gpt-5.3 | Default Codex model for --mixed |
| `COUNCIL_CLAUDE_MODEL` | opus | Claude model for agents |
| `COUNCIL_EXPLORER_MODEL` | sonnet | Model for explorer sub-agents |
| `COUNCIL_EXPLORER_TIMEOUT` | 60 | Explorer timeout in seconds |

### Flags

| Flag | Description |
|------|-------------|
| `--deep` | 3 Claude agents instead of 2 |
| `--mixed` | Add 3 Codex agents |
| `--debate` | Enable adversarial debate round (2 rounds, 2x cost). Incompatible with `--quick`. |
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
codex exec --full-auto -m gpt-5.3 -C "$(pwd)" -o .agents/council/codex-{perspective}.md "{PACKET}"
```

Always use this exact flag order: `--full-auto` → `-m` → `-C` → `-o` → prompt.

**Codex CLI flags (ONLY these are valid):**
- `--full-auto` — No approval prompts (REQUIRED, always first)
- `-m <model>` — Model override (default: gpt-5.3)
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
Bash(command="codex exec --full-auto -m gpt-5.3 -C "$(pwd)" -o .agents/council/codex-pragmatist.md ...", run_in_background=true)
Bash(command="codex exec --full-auto -m gpt-5.3 -C "$(pwd)" -o .agents/council/codex-skeptic.md ...", run_in_background=true)
Bash(command="codex exec --full-auto -m gpt-5.3 -C "$(pwd)" -o .agents/council/codex-visionary.md ...", run_in_background=true)
```

**Wait for completion:**

```
TaskOutput(task_id="...", block=true)
```

### Model Selection

| Vendor | Default | Override |
|--------|---------|----------|
| Claude | opus | `--claude-model=sonnet` |
| Codex | gpt-5.3 | `--codex-model=<model>` |

### Output Collection

All council outputs go to `.agents/council/`:

```bash
# Ensure directory exists
mkdir -p .agents/council

# Claude output (R1)
.agents/council/YYYY-MM-DD-<target>-claude-pragmatist.md

# Claude output (R2, when --debate)
.agents/council/YYYY-MM-DD-<target>-claude-pragmatist-r2.md

# Codex output (R1 only, even with --debate)
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

## Future: Native Teams

The `--debate` flag implements the **deliberation protocol** pattern:
> Independent assessment → evidence exchange → position revision → convergence analysis

This pattern is backend-agnostic. Current and future implementations:

- **Phase 1 (current):** `--debate` via Task tool re-spawning. Judges run as independent background agents.
- **Phase 2 (when `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` graduates):** Upgrade to native teams for selective engagement — judge A responds to judge B's specific claim, not batch all-to-all. Add `--team` flag.
- **Phase 3 (if Phase 2 proves value):** Deliberation protocol as a reusable primitive. Council-as-platform for any multi-agent coordination needing structured disagreement.

**Decision rationale:** See `.agents/council/2026-02-06-native-teams-plan-report.md` and `.agents/council/2026-02-06-plan-validation-report.md`.

---

## See Also

- `skills/vibe/SKILL.md` — Complexity + council for code validation (uses `--preset=default` + validate)
- `skills/pre-mortem/SKILL.md` — Plan validation (uses council validate)
- `skills/post-mortem/SKILL.md` — Work wrap-up (uses council validate + retro)
- `skills/swarm/SKILL.md` — Multi-agent orchestration
- `skills/standards/SKILL.md` — Language-specific coding standards
- `skills/research/SKILL.md` — Codebase exploration (complementary to council research mode)
