---
name: discovery
description: 'Full discovery phase orchestrator. Brainstorm + ao search + research + plan + pre-mortem gate. Produces epic-id and execution-packet for /crank. Triggers: "discovery", "discover", "explore and plan", "research and plan", "discovery phase".'
skill_api_version: 1
user-invocable: true
context:
  window: fork
  intent:
    mode: task
  sections:
    exclude: [HISTORY]
  intel_scope: full
metadata:
  tier: meta
  dependencies:
    - brainstorm  # optional - clarify WHAT before HOW
    - research    # required - codebase exploration
    - plan        # required - epic decomposition
    - pre-mortem  # required - validation gate
    - shared      # optional - CLI fallback table
---

# /discovery — Full Discovery Phase Orchestrator

> **Quick Ref:** Brainstorm → search history → research → plan → pre-mortem. Produces an epic-id and execution-packet ready for `/crank`.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## Quick Start

```bash
/discovery "add user authentication"              # full discovery
/discovery --interactive "refactor payment module" # human gates in research + plan
/discovery --skip-brainstorm "fix login bug"       # skip brainstorm for specific goals
/discovery --complexity=full "migrate to v2 API"   # force full council ceremony
```

## Architecture

```
/discovery <goal> [--interactive] [--complexity=<fast|standard|full>] [--skip-brainstorm]
  │
  ├── Step 1: Brainstorm (optional)
  │   └── /brainstorm <goal> — clarify WHAT before HOW
  │
  ├── Step 2: Search History
  │   └── ao search "<goal keywords>" — surface prior learnings
  │
  ├── Step 3: Research
  │   └── /research <goal> — deep codebase exploration
  │
  ├── Step 4: Plan
  │   └── /plan <goal> — decompose into epic + issues
  │
  └── Step 5: Pre-mortem (gate)
      └── /pre-mortem — council validates the plan
          PASS/WARN → output epic-id + execution-packet
          FAIL → re-plan with findings, max 3 attempts
```

## Execution Steps

### Step 0: Setup

```bash
mkdir -p .agents/rpi
```

Initialize state:

```
discovery_state = {
  goal: "<goal string>",
  interactive: <true if --interactive>,
  complexity: <fast|standard|full or null for auto-detect>,
  skip_brainstorm: <true if --skip-brainstorm or goal is >50 chars and specific>,
  epic_id: null,
  attempt: 1,
  verdict: null
}
```

**CLI dependency detection:**

```bash
# Tracking mode
if command -v bd &>/dev/null; then TRACKING_MODE="beads"; else TRACKING_MODE="tasklist"; fi

# Knowledge flywheel
if command -v ao &>/dev/null; then AO_AVAILABLE=true; else AO_AVAILABLE=false; fi
```

### Step 1: Brainstorm (optional)

**Skip if:** `--skip-brainstorm` flag, OR goal is >50 chars and contains no vague keywords (improve, better, something, somehow, maybe), OR a brainstorm artifact already exists for this goal in `.agents/brainstorm/`.

**Otherwise:** Invoke `/brainstorm` to clarify WHAT before HOW:

```
Skill(skill="brainstorm", args="<goal>")
```

If brainstorm produces a refined goal, use the refined goal for subsequent steps.

### Step 2: Search History

**Skip if:** `ao` CLI is not available.

**Otherwise:** Search for prior session learnings relevant to the goal:

```bash
ao search "<goal keywords>" 2>/dev/null || true
ao lookup --query "<goal keywords>" --limit 5 2>/dev/null || true
```

Summarize any relevant findings (prior attempts, related decisions, known constraints) and carry them forward as context for research.

### Step 3: Research

Invoke `/research` for deep codebase exploration:

```
Skill(skill="research", args="<goal> [--auto]")
```

Pass `--auto` unless `--interactive` was specified. Research output lands in `.agents/research/`.

### Step 4: Plan

Invoke `/plan` to decompose into an epic with trackable issues:

```
Skill(skill="plan", args="<goal> [--auto]")
```

Pass `--auto` unless `--interactive` was specified. Plan output lands in `.agents/plans/` and creates issues via `bd` or TaskList.

After plan completes:
1. Extract epic ID: `bd list --type epic --status open` (beads) or identify from TaskList.
2. Store in `discovery_state.epic_id`.
3. **Auto-detect complexity** (if not overridden):
   - Count issues: `bd children <epic-id> | wc -l`
   - Low: 1-2 issues → `fast`
   - Medium: 3-6 issues → `standard`
   - High: 7+ issues or 3+ waves → `full`

### Step 5: Pre-mortem (gate)

Invoke `/pre-mortem` to validate the plan:

```
Skill(skill="pre-mortem", args="[--quick]")
```

Use `--quick` for fast/standard complexity. Use full council (no `--quick`) for full complexity or `--deep` override.

**Gate logic (max 3 attempts):**

- **PASS/WARN:** Proceed. Store verdict in `discovery_state.verdict`.
- **FAIL:** Retry loop:
  1. Read pre-mortem report: `ls -t .agents/council/*pre-mortem*.md | head -1`
  2. Extract structured findings (description, fix, ref)
  3. Log: `"Pre-mortem: FAIL (attempt N/3) -- retrying plan with feedback"`
  4. Re-invoke `/plan` with findings context:
     ```
     Skill(skill="plan", args="<goal> --auto --context 'Pre-mortem FAIL: <findings>'")
     ```
  5. Re-invoke `/pre-mortem`
  6. If still FAIL after 3 total attempts, stop:
     ```
     "Pre-mortem failed 3 times. Last report: <path>. Manual intervention needed."
     ```
     Output: `<promise>BLOCKED</promise>`

### Step 6: Output

After successful gate (PASS/WARN):

1. **Write execution packet** to `.agents/rpi/execution-packet.json`:

```json
{
  "objective": "<goal>",
  "epic_id": "<epic-id>",
  "contract_surfaces": ["docs/contracts/repo-execution-profile.md"],
  "validation_commands": ["<from repo profile or defaults>"],
  "tracker_mode": "<beads|tasklist>",
  "done_criteria": ["<from repo profile or defaults>"],
  "complexity": "<fast|standard|full>",
  "pre_mortem_verdict": "<PASS|WARN>",
  "discovery_timestamp": "<ISO-8601>"
}
```

2. **Write phase summary** to `.agents/rpi/phase-1-summary-YYYY-MM-DD-<goal-slug>.md`:

```markdown
# Phase 1 Summary: Discovery

- **Goal:** <goal>
- **Epic:** <epic-id>
- **Issues:** <count>
- **Complexity:** <fast|standard|full>
- **Pre-mortem:** <PASS|WARN> (attempt <N>/3)
- **Brainstorm:** <used|skipped>
- **History search:** <findings count or skipped>
- **Status:** DONE
- **Timestamp:** <ISO-8601>
```

3. **Record ratchet and telemetry:**

```bash
ao ratchet record research 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-1" "discovery complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=1 phase_name=discovery 2>/dev/null || true
```

4. **Output completion marker:**

```
<promise>DONE</promise>
```

Report epic-id and suggest next step: `/crank <epic-id>`

## Phase Budgets

| Sub-step | `fast` | `standard` | `full` |
|----------|--------|------------|--------|
| Brainstorm | skip | 2 min | 3 min |
| History search | 1 min | 1 min | 2 min |
| Research | 3 min | 5 min | 10 min |
| Plan | 2 min | 5 min | 10 min |
| Pre-mortem | 1 min | 3 min | 5 min |

On budget expiry: allow in-flight calls to complete, write `[TIME-BOXED]` marker, proceed with whatever artifacts exist.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--interactive` | off | Human gates in research and plan |
| `--skip-brainstorm` | auto | Skip brainstorm step |
| `--complexity=<level>` | auto | Force complexity level (fast/standard/full) |
| `--no-budget` | off | Disable phase time budgets |

## Completion Markers

```
<promise>DONE</promise>      # Discovery complete, epic-id + execution-packet ready
<promise>BLOCKED</promise>   # Pre-mortem failed 3x, manual intervention needed
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Pre-mortem retries hit max | Plan has unresolvable risks | Review findings in `.agents/council/*pre-mortem*.md`, refine goal, re-run `/discovery` |
| No epic ID after plan | bd unavailable and TaskList empty | Check tracking mode, verify `/plan` produced output |
| Brainstorm loops without advancing | Goal too vague for automated clarification | Use `--interactive` or provide a specific goal |
| ao search returns nothing | No prior sessions on this topic | Normal — proceed without history context |

## See Also

- [skills/brainstorm/SKILL.md](../brainstorm/SKILL.md) — clarify WHAT before HOW
- [skills/research/SKILL.md](../research/SKILL.md) — deep codebase exploration
- [skills/plan/SKILL.md](../plan/SKILL.md) — epic decomposition
- [skills/pre-mortem/SKILL.md](../pre-mortem/SKILL.md) — plan validation gate
- [skills/crank/SKILL.md](../crank/SKILL.md) — next phase (implementation)
- [skills/rpi/SKILL.md](../rpi/SKILL.md) — full lifecycle orchestrator
