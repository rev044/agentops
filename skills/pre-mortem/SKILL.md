---
name: pre-mortem
description: 'Pre-mortem simulation for specs and designs. Simulates N iterations of implementation to identify failure modes before they happen. Triggers: "pre-mortem", "simulate spec", "stress test spec", "find spec gaps", "simulate implementation", "what could go wrong", "anticipate failures".'
---

# Pre-Mortem Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Simulate implementation failures BEFORE building to catch problems early.

## Execution Steps

Given `/pre-mortem <spec-or-plan>`:

### Step 1: Load the Spec/Plan

If a path is provided, read it:
```
Tool: Read
Parameters:
  file_path: <provided path>
```

If no path, check for recent plans:
```bash
ls -lt .agents/plans/ .agents/specs/ 2>/dev/null | head -5
```

### Step 2: Understand What We're Building

Read the spec/plan carefully. Identify:
- **Goal**: What are we trying to build?
- **Components**: What pieces need to be created?
- **Integrations**: What does this connect to?
- **Constraints**: What limitations exist?

### Step 3: Dispatch Failure Expert Swarm

**Launch parallel failure simulations from different perspectives.**

```
Tool: Task
Parameters:
  subagent_type: "agentops:integration-failure-expert"
  description: "Integration failure simulation"
  prompt: |
    Simulate integration failures for this spec/plan:
    <spec-content>

    Identify: API mismatches, protocol issues, system boundary problems.
    Return findings with severity ratings.
```

```
Tool: Task
Parameters:
  subagent_type: "agentops:ops-failure-expert"
  description: "Operations failure simulation"
  prompt: |
    Simulate production operations failures for this spec/plan:
    <spec-content>

    Identify: Deployment risks, scaling issues, monitoring gaps, incident response holes.
    Return findings with severity ratings.
```

```
Tool: Task
Parameters:
  subagent_type: "agentops:data-failure-expert"
  description: "Data failure simulation"
  prompt: |
    Simulate data failures for this spec/plan:
    <spec-content>

    Identify: Corruption risks, consistency issues, migration problems, state management gaps.
    Return findings with severity ratings.
```

```
Tool: Task
Parameters:
  subagent_type: "agentops:edge-case-hunter"
  description: "Edge case hunting"
  prompt: |
    Hunt for unhandled edge cases in this spec/plan:
    <spec-content>

    Identify: Boundary conditions, unusual inputs, unexpected states, timing issues.
    Return findings with severity ratings.
```

**Wait for all agents to return, then synthesize findings.**

### Step 4: Categorize Findings

Group findings by severity:

| Severity | Definition | Action |
|----------|------------|--------|
| **CRITICAL** | Will definitely fail | Must fix in spec |
| **HIGH** | Likely to cause problems | Should address |
| **MEDIUM** | Could cause issues | Worth noting |
| **LOW** | Minor concerns | Optional |

### Step 5: Run Vibe on Spec (Optional)

If the spec is substantial, validate it:
```
Tool: Skill
Parameters:
  skill: "agentops:vibe"
  args: "<spec-path>"
```

### Step 6: Write Pre-Mortem Report

**Write to:** `.agents/pre-mortems/YYYY-MM-DD-<topic>.md`

```markdown
# Pre-Mortem: <Topic>

**Date:** YYYY-MM-DD
**Spec:** <path to spec/plan>

## Summary
<What we're building and key risks>

## Simulation Findings

### CRITICAL (Must Fix)
1. **<Issue>**: <Description>
   - **Why it will fail:** <explanation>
   - **Recommended fix:** <how to address>

### HIGH (Should Fix)
1. **<Issue>**: <Description>
   - **Risk:** <what could happen>
   - **Mitigation:** <how to reduce risk>

### MEDIUM (Consider)
- <issue and brief note>

## Ambiguities Found
- <unclear requirement 1>
- <unclear requirement 2>

## Spec Enhancement Recommendations
1. Add: <what to add>
2. Clarify: <what to clarify>
3. Remove: <what to remove>

## Verdict
[ ] READY - Proceed to implementation
[ ] NEEDS WORK - Address critical/high issues first
```

### Step 7: Request Human Approval (Gate 3)

**USE AskUserQuestion tool:**

```
Tool: AskUserQuestion
Parameters:
  questions:
    - question: "Pre-mortem found N critical, M high issues. Proceed to implementation?"
      header: "Gate 3"
      options:
        - label: "Proceed"
          description: "Accept risks, proceed to /crank"
        - label: "Fix Plan"
          description: "Address critical issues before implementing"
        - label: "Back to Research"
          description: "Need more research before proceeding"
      multiSelect: false
```

**Wait for approval before proceeding to implementation.**

### Step 8: Report to User

Tell the user:
1. Number of issues found by severity
2. Whether spec is ready or needs work
3. Top 3 most important fixes
4. Location of pre-mortem report
5. Gate 3 decision

## Key Rules

- **Simulate, don't just review** - mentally build it 10 times
- **Be adversarial** - look for ways it will fail
- **Categorize by severity** - not all issues are equal
- **Write findings** - always produce `.agents/pre-mortems/` artifact
- **Block on CRITICAL** - don't proceed with critical issues unresolved
