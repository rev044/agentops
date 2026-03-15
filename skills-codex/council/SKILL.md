---
name: council
description: 'Multi-model consensus council. Spawns parallel judges via spawn_agents_on_csv. Modes: validate, brainstorm, research. Triggers: "council", "get consensus", "multi-model review", "multi-perspective review", "council validate", "council brainstorm", "council research".'
metadata:
  tier: judgment
---

# $council — Multi-Model Consensus Council (Codex Native)

Spawn parallel judges with different perspectives via `spawn_agents_on_csv`, consolidate into consensus.

## Quick Start

```bash
$council --quick validate recent                               # fast inline check
$council validate this plan                                    # validation (2 agents)
$council brainstorm caching approaches                         # brainstorm
$council --deep validate the implementation                    # 3 agents
$council --preset=security-audit validate the auth system      # preset personas
```

## Modes

| Mode | Agents | Method | Use Case |
|------|--------|--------|----------|
| `--quick` | 0 (inline) | Self | Fast single-agent check, no spawning |
| default | 2 | `spawn_agents_on_csv` | Independent judges |
| `--deep` | 3 | `spawn_agents_on_csv` | Thorough review |

**Note:** `--debate` (multi-round adversarial) requires agent messaging which is not available in Codex. Use `--deep` with 3 judges instead for thorough review.

**Note:** `--mixed` (cross-vendor) is not applicable in Codex — all judges use the same model.

## Task Types

| Type | Trigger Words | Focus |
|------|---------------|-------|
| **validate** | validate, check, review, assess, critique | Is this correct? What's wrong? |
| **brainstorm** | brainstorm, explore, options, approaches | Alternatives? Pros/cons? |
| **research** | research, investigate, deep dive, analyze | What can we discover? |

## Execution Flow

### Phase 1: Build Packet

1. Determine task type from user prompt
2. Identify target (files, diffs, plan, code)
3. Read relevant context files
4. Select perspectives (or use preset)

### Phase 1a: Build Judge CSV

```bash
mkdir -p .agents/council
CSV_FILE=".agents/council/judges-$(date +%Y%m%d-%H%M%S).csv"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Build CSV with judge assignments
echo "judge_id,perspective,target_files,task_type,context" > "$CSV_FILE"

# Default: 2 judges, --deep: 3 judges
echo "\"judge-1\",\"correctness\",\"$TARGET_FILES\",\"$TASK_TYPE\",\"$CONTEXT\"" >> "$CSV_FILE"
echo "\"judge-2\",\"completeness\",\"$TARGET_FILES\",\"$TASK_TYPE\",\"$CONTEXT\"" >> "$CSV_FILE"
# --deep adds:
# echo "\"judge-3\",\"edge-cases\",\"$TARGET_FILES\",\"$TASK_TYPE\",\"$CONTEXT\"" >> "$CSV_FILE"
```

### Phase 1b: Spawn Judges

```
spawn_agents_on_csv(
    csv_path=".agents/council/judges-{timestamp}.csv",
    instruction="You are judge {judge_id} reviewing from a {perspective} perspective.

Task: {task_type} the following target.
Target files: {target_files}

Context:
{context}

INSTRUCTIONS:
1. Read the target files
2. Analyze from your assigned perspective ({perspective})
3. Write your full analysis to .agents/council/{judge_id}-{timestamp}.md
4. Report your verdict via report_agent_job_result

For validate tasks, use verdicts: PASS, WARN, FAIL
For brainstorm tasks, list options with pros/cons
For research tasks, report findings with evidence

Write ALL analysis to your output file. Keep the result report minimal.",
    id_column="judge_id",
    output_schema={
        "type": "object",
        "properties": {
            "judge_id": {"type": "string"},
            "verdict": {"type": "string"},
            "confidence": {"type": "string", "enum": ["high", "medium", "low"]},
            "finding_count": {"type": "integer"},
            "output_file": {"type": "string"},
            "summary": {"type": "string"}
        },
        "required": ["judge_id", "verdict", "confidence", "finding_count", "output_file", "summary"],
        "additionalProperties": false
    },
    max_concurrency=3,
    max_runtime_seconds=120
)
```

### Phase 1c: Wait for Judges

```
wait(timeout_seconds=300)
```

### Phase 2: Consolidation (Lead — Inline)

The lead reads each judge's output file and synthesizes:

1. Read each `.agents/council/{judge_id}-*.md` file
2. Compute consensus verdict:
   - **PASS:** All judges PASS (or majority PASS, none FAIL)
   - **WARN:** Any judge WARN, none FAIL
   - **FAIL:** Any judge FAIL
3. Identify shared findings across judges
4. Surface disagreements with attribution
5. Generate final report

### Phase 3: Write Report

Save to `.agents/council/YYYY-MM-DD-<type>-<target>.md`:

```markdown
# Council Report: <type> <target>

**Consensus:** PASS/WARN/FAIL
**Judges:** N responded / N spawned
**Date:** YYYY-MM-DD

## Shared Findings
- Finding 1 (judges 1, 2)
- Finding 2 (judges 1, 3)

## Disagreements
- Judge 1 says X, Judge 2 says Y

## Recommendations
1. ...
2. ...

## Individual Verdicts
| Judge | Perspective | Verdict | Confidence | Findings |
|-------|-------------|---------|------------|----------|
| judge-1 | correctness | PASS | high | 3 |
| judge-2 | completeness | WARN | medium | 5 |
```

## Presets

| Preset | Perspectives |
|--------|-------------|
| default | correctness, completeness |
| security-audit | vulnerability, attack-surface, data-flow |
| architecture | coupling, scalability, maintainability |
| research | breadth, depth, contrarian |
| ops | reliability, observability, failure-modes |

Use: `$council --preset=security-audit validate the auth system`

## Graceful Degradation

| Failure | Behavior |
|---------|----------|
| 1 of N judges timeout | Proceed with N-1, note in report |
| All judges fail | Return error, suggest retry |
| No multi-agent capability | Fall back to `--quick` (inline) |

## Context Budget Rule

Judges write ALL analysis to output files. Results to the lead contain ONLY minimal signals (verdict + file path). This prevents N judges from flooding the lead's context.

## Standards Integration

If `$standards` is available and the target includes code files, load applicable language standards and include in judge context packet.

## First-Pass Rigor Gate (validate mode)

When validating plans/specs, judges must check:
1. Mutation + ack sequence is explicit and non-contradictory
2. Consume-at-most-once path is crash-safe
3. Status/precedence behavior has field-level truth table
4. Conformance includes boundary failpoint tests

Missing gate item → minimum WARN. Critical unverifiable invariant → FAIL.

## Reference Documents

- [references/agent-prompts.md](references/agent-prompts.md)
- [references/backend-background-tasks.md](references/backend-background-tasks.md)
- [references/backend-codex-subagents.md](references/backend-codex-subagents.md)
- [references/backend-inline.md](references/backend-inline.md)
- [references/brainstorm-techniques.md](references/brainstorm-techniques.md)
- [references/caching-guidance.md](references/caching-guidance.md)
- [references/cli-spawning.md](references/cli-spawning.md)
- [references/debate-protocol.md](references/debate-protocol.md)
- [references/explorers.md](references/explorers.md)
- [references/finding-extraction.md](references/finding-extraction.md)
- [references/model-profiles.md](references/model-profiles.md)
- [references/output-format.md](references/output-format.md)
- [references/personas.md](references/personas.md)
- [references/presets.md](references/presets.md)
- [references/quick-mode.md](references/quick-mode.md)
- [references/ralph-loop-contract.md](references/ralph-loop-contract.md)
