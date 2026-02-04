---
name: pre-mortem
description: 'Tools-first spec validation. Runs toolchain-validate.sh BEFORE any manual review. Stops on CRITICAL (exit 2), warns on HIGH (exit 3). Then generates checklists from failure-taxonomy.md. Triggers: "pre-mortem", "validate spec", "what could go wrong".'
---

# Pre-Mortem Skill

> **Quick Ref:** Toolchain FIRST -> Checklist generation -> Manual verification. Output: `.agents/pre-mortems/*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**Architecture:** Run toolchain FIRST (gate on CRITICAL). Then generate explicit checklist from failure-taxonomy.md. Tools verify what's verifiable; manual review checks what requires judgment. Every finding must have location (line number), verification status (tool-verified or manual), and specific fix.

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

### Step 1a: Search for Prior Failure Learnings (if ao available)

**Before generating the checklist, search for relevant past failures:**

```bash
# Search for prior failure patterns related to this topic
ao search "failure <topic>" 2>/dev/null || echo "ao not available, skipping failure pattern search"

# Also search for incidents and anti-patterns
ao search "incident <topic>" 2>/dev/null
ao anti-patterns 2>/dev/null | grep -i "<topic>" || true
```

**Review ao search results:** If ao returns relevant learnings, incorporate them:
- **Prior incidents:** Add checks for conditions that caused past failures
- **Known anti-patterns:** Add specific checklist items to verify they're avoided
- **Lessons learned:** Use these to generate additional verification questions

**Note:** This search runs BEFORE generating the checklist so prior knowledge can inform what to check.

### Step 2: Run Toolchain Validation (MANDATORY GATE)

**Before ANY other validation, run the toolchain:**

```bash
./scripts/toolchain-validate.sh --gate --json 2>&1 | tee .agents/tooling/pre-mortem-run.log
TOOL_EXIT=$?

# Also capture structured output for agent context
cat .agents/tooling/summary.json
```

**Exit Code Handling:**

| Exit Code | Meaning | Action |
|-----------|---------|--------|
| 0 | All tools pass | Proceed to checklist generation |
| 2 | CRITICAL findings | **STOP. Report tool findings. Do NOT proceed.** |
| 3 | HIGH findings only | WARN, include in pre-mortem context |

**If TOOL_EXIT == 2 (CRITICAL):**

```
STOP and report to user:

  Pre-Mortem: BLOCKED BY TOOLCHAIN

  Toolchain found CRITICAL issues that must be fixed first:
  - See .agents/tooling/summary.json for summary
  - See .agents/tooling/<tool>.txt for details

  Fix these issues before running pre-mortem.

  Tool Findings (verified):
  | Tool | Status | Finding Count | Details File |
  |------|--------|---------------|--------------|
  | gitleaks | <status> | <count> | .agents/tooling/gitleaks.txt |
  | semgrep | <status> | <count> | .agents/tooling/semgrep.txt |
  | ... | ... | ... | ... |
```

**DO NOT proceed to checklist generation if tools found CRITICAL issues.** This prevents theater where manual review ignores definitive tool failures.

**If TOOL_EXIT == 3 (HIGH only):**

Continue to Step 2a, but include tool findings in the pre-mortem report:
```markdown
## Toolchain Findings (HIGH severity)

The following issues were found by automated tools. Include these in agent context:

<paste structured output from .agents/tooling/summary.json>
```

### Step 2a: Generate Explicit Checklist

**Load the failure taxonomy:**
```
Tool: Read
Parameters:
  file_path: skills/pre-mortem/references/failure-taxonomy.md
```

**For each category in the taxonomy, generate specific questions:**

| Category | Checklist Item | Verification Method |
|----------|----------------|---------------------|
| Interface Mismatch | Does spec define API schema? | Search for `schema:` or `interface` |
| Timing | Does spec define timeouts? | Search for `timeout:` |
| Error Handling | Does spec define error states? | Search for `error:` or `failure:` |
| Safety | Does spec require confirmation for destructive ops? | Search for `confirm` |
| Integration | Does spec list dependencies? | Search for `depends:` or `requires:` |
| Rollback | Does spec define rollback procedure? | Search for `rollback` or `revert` |
| State | Does spec define state transitions? | Search for `state:` or `transition` |

**The checklist above covers 7 essential items. For comprehensive validation, use all 10 categories from `references/failure-taxonomy.md`.**

**Build the checklist BEFORE reading the spec.** This prevents pattern-matching bias.

### Step 3: Run Mechanical Cross-Reference Check

**Run automated cross-reference:**

```bash
./scripts/spec-cross-reference.sh <spec-file> | tee .agents/pre-mortems/cross-ref.md
```

This catches mechanically:
- File paths that don't exist
- Function/type references that aren't defined
- Broken markdown links

Include output in pre-mortem report under "## Cross-Reference Verification"

### Step 4: Verify Checklist Items

**Verify checklist items systematically. Do not "simulate failures" - check for specific gaps.**

#### 4a. Verify Spec Completeness

For EACH checklist item from failure-taxonomy.md, read the spec and answer:

| Checklist Item | Present? | Location (line) | Complete? |
|----------------|----------|-----------------|-----------|
| API schema defined | yes/no | line N | yes/partial/no |
| Timeouts specified | yes/no | line N | yes/partial/no |
| Error states listed | yes/no | line N | yes/partial/no |
| Rollback procedure | yes/no | line N | yes/partial/no |
| Dependencies listed | yes/no | line N | yes/partial/no |
| State transitions | yes/no | line N | yes/partial/no |
| Confirmation for destructive ops | yes/no | line N | yes/partial/no |

For items marked "no" or "partial": flag as GAP with specific fix.

#### 4b. Find Implicit Assumptions

Find statements that assume something without stating it:
- "The user will..." -> What if they don't?
- "The API returns..." -> What if it errors?
- "This runs after..." -> What if order changes?

Record assumptions:
| Location (line) | Assumption | What If Wrong? | Specific Clarification Needed |
|-----------------|------------|----------------|-------------------------------|

Every finding MUST have a line number and specific fix.

#### 4c. Find Boundary Conditions

For each input/parameter in the spec:
| Location (line) | Input | Type | Min/Max Stated? | Empty Handling? | Invalid Handling? |
|-----------------|-------|------|-----------------|-----------------|-------------------|

Flag any inputs without explicit boundary handling.
Every finding MUST have a line number and specific fix.

### Step 5: Categorize Findings

Combine agent outputs and categorize:

| Severity | Definition | Action |
|----------|------------|--------|
| **CRITICAL** | Spec has fundamental gap (no rollback, no error handling) | Must fix before implementation |
| **HIGH** | Spec makes unstated assumptions | Should clarify |
| **MEDIUM** | Spec could be clearer | Worth noting |
| **LOW** | Minor improvements | Optional |

**Every finding must have:**
1. Location (line number in spec)
2. Description (what's missing or unclear)
3. Specific fix (exact text to add/change)

### Apply Enhancement Patterns

For each finding, identify the applicable pattern from `references/enhancement-patterns.md`:

| Gap Type | Enhancement Pattern |
|----------|---------------------|
| Missing schema | "Schema from Code" - Extract schema from actual code |
| Missing error handling | "Error Recovery Matrix" - Map all error types to actions |
| Missing timeouts | "Per-Tool Timeout Configuration" - Add per-operation timeouts |
| Missing safety info | "Mandatory Safety Display" - Add safety level classification |
| Missing progress feedback | "Progress Feedback Specification" - Add update frequency |
| Missing escalation | "Escalation Flow" - Define when/how to escalate |
| Missing audit trail | "Audit Trail Requirements" - Add logging requirements |

### Step 6: Write Pre-Mortem Report

**Write to:** `.agents/pre-mortems/YYYY-MM-DD-<topic>.md`

```markdown
# Pre-Mortem: <Topic>

**Date:** YYYY-MM-DD
**Spec:** <path to spec/plan>

## Toolchain Verification (Gate)

| Tool | Status | Findings | Details |
|------|--------|----------|---------|
| gitleaks | pass/findings/skipped | N CRITICAL | .agents/tooling/gitleaks.txt |
| semgrep | pass/findings/skipped | N HIGH | .agents/tooling/semgrep.txt |
| golangci-lint | pass/findings/skipped | N issues | .agents/tooling/golangci-lint.txt |
| shellcheck | pass/findings/skipped | N issues | .agents/tooling/shellcheck.txt |

**Gate Result:** PASS / BLOCKED (exit code 2) / WARN (exit code 3)

## Checklist Verification

| Category | Item | Verified | Method | Location |
|----------|------|----------|--------|----------|
| Interface | API schema | yes/no | tool: spec-cross-reference / manual | line N |
| Timing | Timeouts | yes/no | manual | line N |
| Error | Error states | yes/no | manual | line N |
| Safety | Confirmation | yes/no | manual | line N |
| Rollback | Rollback procedure | yes/no | manual | line N |
| Deps | Dependencies listed | yes/no | tool: spec-cross-reference / manual | line N |
| State | State transitions | yes/no | manual | line N |

## Findings

### CRITICAL (Must Fix Before Implementation)
1. **<Issue>**: <Description>
   - **Location:** line N
   - **Verified:** yes (tool: <tool-name>) / no (manual review)
   - **Why Critical:** <explanation>
   - **Specific Fix:** <exact text to add to spec>

### HIGH (Should Clarify)
1. **<Issue>**: <Description>
   - **Location:** line N
   - **Verified:** yes (tool: <tool-name>) / no (manual review)
   - **Assumption Made:** <what's assumed>
   - **Specific Clarification:** <exact question to answer>

### MEDIUM
- **<Issue>** (line N, verified: yes/no): <issue and specific fix>

## Implicit Assumptions Found
| Location | Assumption | Risk | Clarification Needed |
|----------|------------|------|---------------------|

## Edge Cases Without Handling
| Location | Input | Boundary Missing | Suggested Handling |
|----------|-------|-----------------|-------------------|

## Verdict

[ ] READY - Toolchain passed, all checklist items verified, no CRITICAL gaps
[ ] BLOCKED - Toolchain found CRITICAL issues (exit code 2)
[ ] NEEDS WORK - <count> CRITICAL gaps must be addressed
```

### Step 7: Request Human Approval (Gate 3)

**Gate Criteria:**
- **READY**: 0 CRITICAL gaps, <=2 HIGH gaps
- **NEEDS WORK**: 1+ CRITICAL or >2 HIGH gaps

```
Tool: AskUserQuestion
Parameters:
  questions:
    - question: "Pre-mortem found <N> gaps. Proceed to implementation?"
      header: "Gate 3"
      options:
        - label: "Proceed"
          description: "Gaps acceptable, start implementation"
        - label: "Fix Spec"
          description: "Address gaps before implementing"
        - label: "More Research"
          description: "Need more information"
      multiSelect: false
```

### Step 8: Report to User

Tell the user:
1. Checklist verification results (table)
2. Number of gaps by severity
3. Top 3 items that need attention (with line numbers)
4. Location of pre-mortem report
5. Gate 3 decision

## Key Differences from Previous Version

| Before | After |
|--------|-------|
| "Simulate failures" (vague) | Verify checklist items (specific) |
| Agents dispatched first | Toolchain runs FIRST (gate on CRITICAL) |
| Gestalt impression | Explicit checklist + location (line number) |
| Hallucinated confidence scores | Verification status: "verified: yes/no" with tool citation |
| Agents ignore tool failures | STOP if exit code 2 (CRITICAL), WARN if exit code 3 (HIGH) |
| Agents get vague "find issues" prompt | Agents receive structured tool output as context |
| No methodology | Uses failure-taxonomy.md |
| Findings without location | Every finding has line number and specific fix |
| Pattern matching | Mechanical verification first, manual review second |
