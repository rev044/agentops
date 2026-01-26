---
name: bug-hunt
description: 'Investigate suspected bugs with git archaeology and root cause analysis. Triggers: "bug", "broken", "doesn''t work", "failing", "investigate bug".'
---

# Bug Hunt Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Systematic investigation to find root cause and design a complete fix.

## Execution Steps

Given `/bug-hunt <symptom>`:

### Step 1: Confirm the Bug

First, reproduce the issue:
- What's the expected behavior?
- What's the actual behavior?
- Can you reproduce it consistently?

If the bug can't be reproduced, gather more information before proceeding.

### Step 2: Locate the Symptom

Find where the bug manifests:
```bash
# Search for error messages
grep -r "<error-text>" . --include="*.py" --include="*.ts" --include="*.go" 2>/dev/null | head -10

# Search for function/variable names
grep -r "<relevant-name>" . --include="*.py" --include="*.ts" --include="*.go" 2>/dev/null | head -10
```

### Step 3: Git Archaeology

Find when/how the bug was introduced:

```bash
# When was the file last changed?
git log --oneline -10 -- <file>

# What changed recently?
git diff HEAD~10 -- <file>

# Who changed it and why?
git blame <file> | grep -A2 -B2 "<suspicious-line>"

# Search for related commits
git log --oneline --grep="<keyword>" | head -10
```

### Step 4: Trace the Execution Path

**USE THE TASK TOOL** to explore the code:

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  description: "Trace bug execution path"
  prompt: |
    Trace the execution path for: <symptom>

    1. Find the entry point where the bug manifests
    2. Trace backward to find where bad data/state originates
    3. Identify all functions in the path
    4. Look for recent changes to these functions

    Return:
    - Execution path (function call chain)
    - Likely location of root cause
    - Recent changes that might be responsible
```

### Step 5: Identify Root Cause

Based on tracing, identify:
- **What** is wrong (the actual bug)
- **Where** it is (file:line)
- **When** it was introduced (commit)
- **Why** it happens (the logic error)

### Step 6: Design the Fix

Before writing code, design the fix:
- What needs to change?
- What are the edge cases?
- Will this fix break anything else?
- Are there tests to update?

### Step 7: Write Bug Report

**Write to:** `.agents/research/YYYY-MM-DD-bug-<slug>.md`

```markdown
# Bug Report: <Short Description>

**Date:** YYYY-MM-DD
**Severity:** <critical|high|medium|low>
**Status:** <investigating|root-cause-found|fix-designed>

## Symptom
<What the user sees>

## Expected Behavior
<What should happen>

## Reproduction Steps
1. <step 1>
2. <step 2>
3. <observe bug>

## Root Cause Analysis

### Location
- **File:** <path>
- **Line:** <line number>
- **Function:** <function name>

### Cause
<Explanation of what's wrong>

### When Introduced
- **Commit:** <hash>
- **Date:** <date>
- **Author:** <author>

## Proposed Fix

### Changes Required
1. <change 1>
2. <change 2>

### Risks
- <potential risk>

### Tests Needed
- <test to add/update>

## Related
- <related issues or PRs>
```

### Step 8: Report to User

Tell the user:
1. Root cause identified (or not yet)
2. Location of the bug (file:line)
3. Proposed fix
4. Location of bug report
5. Next step: implement fix or gather more info

## Key Rules

- **Reproduce first** - confirm the bug exists
- **Use git archaeology** - understand history
- **Trace systematically** - follow the execution path
- **Identify root cause** - not just symptoms
- **Design before fixing** - think through the solution
- **Document findings** - write the bug report

## Quick Checks

Common bug patterns to check:
- Off-by-one errors
- Null/undefined handling
- Race conditions
- Type mismatches
- Missing error handling
- State not reset
- Cache issues
