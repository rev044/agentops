# RPI Integration (Research → Plan → Implement)

## Overview

The `--full` flag chains the complete workflow: research the codebase, create a plan with issues, execute via polecats.

## Full Workflow

```bash
/gastown "Add OAuth support to the API" --full
```

Executes:

```
1. Research Phase
   └─ Task(subagent_type="Explore", model="haiku")
   └─ Output: understanding of codebase
   └─ Context cost: LOW (haiku, exploration only)

2. Plan Phase
   └─ Create epic with bd create
   └─ Create child issues
   └─ Set dependencies with bd dep add
   └─ Context cost: LOW (local operations)

3. Execute Phase
   └─ Compute waves from dependencies
   └─ For each wave:
       └─ gt convoy create
       └─ gt sling × N
       └─ Poll gt convoy status
   └─ Context cost: LOW (status only)

4. Report Phase
   └─ Synthesize from beads
   └─ Close epic
   └─ Output summary
```

## Implementation

```python
def gastown_full(goal: str, rig: str = "gastown"):
    # Phase 1: Research
    research = Task(
        subagent_type="Explore",
        model="haiku",
        prompt=f"Research codebase for: {goal}. Find relevant files, patterns, constraints."
    )
    # Haiku output is small - OK in context

    # Phase 2: Plan
    # Create epic
    epic_id = bash(f"bd create 'Epic: {goal}' --type epic --priority 1")

    # Decompose into features based on research
    features = decompose_goal(goal, research)
    issue_ids = []

    for feature in features:
        issue_id = bash(f"bd create '{feature.title}' --type feature --priority 2")
        issue_ids.append(issue_id)

        # Set dependencies
        for dep in feature.depends_on:
            bash(f"bd dep add {issue_id} {dep}")

    # Track children
    bash(f"bd comments add {epic_id} 'Children: {', '.join(issue_ids)}'")
    bash(f"bd update {epic_id} --status in_progress")

    # Phase 3: Execute
    waves = compute_waves(issue_ids)

    for wave_num, wave_issues in enumerate(waves, 1):
        # Create convoy
        convoy_id = bash(f"gt convoy create 'Wave {wave_num}' {' '.join(wave_issues)}")

        # Dispatch to polecats
        for issue in wave_issues:
            bash(f"gt sling {issue} {rig}")

        # Monitor (small output)
        result = monitor_convoy(convoy_id, timeout_minutes=60)

        if result == "blocked":
            escalate_to_human()
            return

    # Phase 4: Report
    summary = synthesize_results(epic_id, issue_ids)
    bash(f"bd close {epic_id} --reason '{summary}'")

    print(f"Complete: {epic_id}")
    print(summary)
```

## Research Phase Details

Use haiku for exploration - fast and cheap:

```python
research = Task(
    subagent_type="Explore",
    model="haiku",
    prompt=f"""
    Research the codebase for implementing: {goal}

    Find:
    1. Relevant existing code (file paths, key functions)
    2. Existing patterns to follow
    3. Related tests
    4. Potential constraints or blockers

    Return concise findings.
    """
)
```

**Why haiku:** Exploration output is small. No need for opus-level reasoning.

## Plan Phase Details

Decompose into discrete features:

```python
def decompose_goal(goal: str, research: str) -> list:
    # Based on research, identify discrete units of work
    # Each feature should be:
    # - Completable in single session
    # - Testable independently
    # - Following existing patterns

    features = []

    # Example decomposition for "Add OAuth"
    features.append(Feature(
        title="Add OAuth provider configuration",
        depends_on=[]
    ))
    features.append(Feature(
        title="Implement token refresh middleware",
        depends_on=[]
    ))
    features.append(Feature(
        title="Add login button component",
        depends_on=["oauth-config"]
    ))
    features.append(Feature(
        title="Integration tests for OAuth flow",
        depends_on=["oauth-config", "token-refresh"]
    ))

    return features
```

## Execute Phase Details

Waves computed from dependencies:

```python
def compute_waves(issues: list) -> list:
    waves = []
    remaining = set(issues)
    completed = set()

    while remaining:
        # Find issues with all deps satisfied
        ready = [i for i in remaining if all_deps_in(i, completed)]

        if not ready:
            raise Exception("Circular dependency detected")

        waves.append(ready)
        completed.update(ready)
        remaining -= set(ready)

    return waves
```

## Context Budget

| Phase | Agent | Tokens |
|-------|-------|--------|
| Research | haiku Task | ~2-5K (small output) |
| Plan | local | ~500 (bd commands) |
| Execute | polecats | ~1K per wave (status only) |
| Report | local | ~500 |

**Total:** ~5-10K tokens for full workflow.

Compare to Task-based: 80K+ for 8 parallel agents.

## Error Handling

```python
def gastown_full_safe(goal: str, rig: str):
    try:
        # Check prerequisites
        if not gas_town_available():
            print("Gas Town not available. Use /crank instead.")
            return

        gastown_full(goal, rig)

    except Exception as e:
        # Save state for resume
        if epic_id:
            bash(f"bd comments add {epic_id} 'CHECKPOINT: Error - {e}'")
        raise
```

## Resume After Error

```bash
# If full workflow fails partway
/gastown <epic-id> --resume

# Reads state from beads, continues from last wave
```
