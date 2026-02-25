---
name: bug-hunt
description: 'Investigate suspected bugs or run proactive code audits. Triggers: "bug", "broken", "doesn''''t work", "failing", "investigate bug", "debug", "find the bug", "troubleshoot", "audit code", "find bugs in", "code audit", "hunt bugs".'
---


# Bug Hunt Skill

> **Quick Ref:** 4-phase investigation (Root Cause → Pattern → Hypothesis → Fix). Output: `.agents/research/YYYY-MM-DD-bug-*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Systematic investigation to find root cause and design a complete fix — or proactive audit to find hidden bugs before they bite.

**Requires:**
- session-start.sh has executed (creates `.agents/` directories for output)
- bd CLI (beads) for issue tracking if creating follow-up issues

## Modes

| Mode | Invocation | When |
|------|------------|------|
| **Investigation** | `$bug-hunt <symptom>` | You have a known bug or failure |
| **Audit** | `$bug-hunt --audit <scope>` | Proactive sweep for hidden bugs |

Investigation mode uses the 4-phase structure below. Audit mode uses systematic read-and-classify — see [Audit Mode](#audit-mode).

---

## The 4-Phase Structure (Investigation Mode)

| Phase | Focus | Output |
|-------|-------|--------|
| **1. Root Cause** | Find the actual bug location | file:line, commit |
| **2. Pattern** | Compare against working examples | Differences identified |
| **3. Hypothesis** | Form and test single hypothesis | Pass/fail for each |
| **4. Implementation** | Fix at root, not symptoms | Verified fix |

**For failure category taxonomy and the 3-failure rule, read `skills/bug-hunt/references/failure-categories.md`.**

## Execution Steps

Given `$bug-hunt <symptom>`:

---

## Phase 1: Root Cause Investigation

### Step 1.1: Confirm the Bug

First, reproduce the issue:
- What's the expected behavior?
- What's the actual behavior?
- Can you reproduce it consistently?

**Read error messages carefully.** Do not skip or skim them.

If the bug can't be reproduced, gather more information before proceeding.

### Step 1.2: Locate the Symptom

Find where the bug manifests:
```bash
# Search for error messages
grep -r "<error-text>" . --include="*.py" --include="*.ts" --include="*.go" 2>/dev/null | head -10

# Search for function/variable names
grep -r "<relevant-name>" . --include="*.py" --include="*.ts" --include="*.go" 2>/dev/null | head -10
```

### Step 1.3: Git Archaeology

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

### Step 1.4: Trace the Execution Path

**USE THE TASK TOOL** (subagent_type: "Explore") to trace the execution path:
- Find the entry point where the bug manifests
- Trace backward to find where bad data/state originates
- Identify all functions in the path and recent changes to them
- Return: execution path, likely root cause location, responsible changes

### Step 1.5: Identify Root Cause

Based on tracing, identify:
- **What** is wrong (the actual bug)
- **Where** it is (file:line)
- **When** it was introduced (commit)
- **Why** it happens (the logic error)

---

## Phase 2: Pattern Analysis

### Step 2.1: Find Working Examples

Search the codebase for similar functionality that WORKS:
```bash
# Find similar patterns
grep -r "<working-pattern>" . --include="*.py" --include="*.ts" --include="*.go" 2>/dev/null | head -10
```

### Step 2.2: Compare Against Reference

Identify ALL differences between:
- The broken code
- The working reference

Document each difference.

---

## Phase 3: Hypothesis and Testing

### Step 3.1: Form Single Hypothesis

State your hypothesis clearly:
> "I think X is wrong because Y"

**One hypothesis at a time.** Do not combine multiple guesses.

### Step 3.2: Test with Smallest Change

Make the SMALLEST possible change to test the hypothesis:
- If it works → proceed to Phase 4
- If it fails → record failure, form NEW hypothesis

### Step 3.3: Check Failure Counter

Check failure count per `skills/bug-hunt/references/failure-categories.md`. After 3 countable failures, escalate to architecture review.

---

## Phase 4: Implementation

### Step 4.1: Design the Fix

Before writing code, design the fix:
- What needs to change?
- What are the edge cases?
- Will this fix break anything else?
- Are there tests to update?

### Step 4.2: Create Failing Test (if possible)

Write a test that demonstrates the bug BEFORE fixing it.

### Step 4.3: Implement Single Fix

Fix at the ROOT CAUSE, not at symptoms.

### Step 4.4: Verify Fix

Run the failing test - it should now pass.

---

## Audit Mode

When invoked with `--audit`, bug-hunt switches to a proactive sweep. No symptom needed — you're hunting for bugs that haven't been reported yet.

```bash
$bug-hunt --audit cli/internal/goals/     # audit a package
$bug-hunt --audit src/auth/               # audit a directory
$bug-hunt --audit .                        # audit recent changes in repo
```

### Audit Step 1: Scope

Identify target files from the scope argument:

```bash
# Find source files in scope
find <scope> -name "*.go" -o -name "*.py" -o -name "*.ts" -o -name "*.rs" | head -50
```

If scope is `.` or broad (>50 files), narrow to recently changed files:

```bash
git log --since="2 weeks ago" --name-only --pretty=format: -- <scope> | sort -u | head -30
```

### Audit Step 2: Systematic Read

Read **every file** in scope line by line. For each file, check:

| Category | What to Look For |
|----------|-----------------|
| **Resource Leaks** | Unclosed handles, orphaned processes, missing cleanup/defer |
| **String Safety** | Byte-level truncation of UTF-8, unsanitized input |
| **Dead Code** | Unreachable branches, unused constants, shadowed variables |
| **Hardcoded Values** | Paths, URLs, repo-specific assumptions that won't work elsewhere |
| **Edge Cases** | Empty input, nil/zero values, boundary conditions |
| **Concurrency** | Unprotected shared state, goroutine leaks, missing signal handlers |
| **Error Handling** | Swallowed errors, missing context, wrong error types |

**Key discipline:** Read line by line. Do not skim. The proven methodology (5 bugs found, 0 hypothesis failures) came from careful reading, not heuristic scanning.

**USE THE TASK TOOL** (subagent_type: "Explore") for large scopes — split files across parallel agents.

### Audit Step 3: Classify Findings

For each finding, assign severity:

| Severity | Criteria | Examples |
|----------|----------|---------|
| **HIGH** | Data loss, security, resource leak, process orphaning | Zombie processes, SQL injection, file handle leak |
| **MEDIUM** | Wrong output, incorrect defaults, silent data corruption | UTF-8 truncation, hardcoded paths, wrong error code |
| **LOW** | Dead code, cosmetic, minor inconsistency | Unreachable branch, unused import, style violation |

### Audit Step 4: Write Audit Report

**For audit report format, read `skills/bug-hunt/references/audit-report-template.md`.**

Write to `.agents/research/YYYY-MM-DD-bug-<scope-slug>.md`.

Report to user with a summary table:

```
| # | Bug | Severity | File | Fix |
|---|-----|----------|------|-----|
| 1 | <description> | HIGH | <file:line> | <proposed fix> |
```

Include failure count (hypothesis tests that didn't confirm). Zero failures = clean audit.

---

## Step 5: Write Bug Report

**For bug report template, read `skills/bug-hunt/references/bug-report-template.md`.**

### Step 6: Report to User

Tell the user:
1. Root cause identified (or not yet)
2. Location of the bug (file:line)
3. Proposed fix
4. Location of bug report
5. Failure count and types encountered
6. Next step: implement fix or gather more info

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

## Examples

### Investigating a Test Failure

**User says:** `$bug-hunt "tests failing on CI but pass locally"`

**What happens:**
1. Agent confirms bug by checking CI logs vs local test output
2. Agent uses git archaeology to find recent changes to test files
3. Agent traces execution path to identify environment-specific differences
4. Agent forms hypothesis about missing environment variable
5. Agent creates failing test locally by unsetting the variable
6. Agent implements fix by adding default value
7. Bug report written to `.agents/research/2026-02-13-bug-test-failure.md`

**Result:** Root cause identified as missing ENV variable in CI configuration. Fix applied and verified.

### Tracking Down a Regression

**User says:** `$bug-hunt "feature X broke after yesterday's deployment"`

**What happens:**
1. Agent reproduces issue in current state
2. Agent uses `git log --since="2 days ago"` to find recent commits
3. Agent uses `git bisect` to identify exact breaking commit
4. Agent compares broken code against working examples in codebase
5. Agent forms hypothesis about introduced type mismatch
6. Agent implements minimal fix and verifies with existing tests
7. Bug report documents commit sha, root cause, and fix

**Result:** Regression traced to commit abc1234, type conversion error fixed at root cause in validation logic.

### Proactive Code Audit

**User says:** `$bug-hunt --audit cli/internal/goals/`

**What happens:**
1. Agent scopes to all `.go` files in the goals package
2. Agent reads each file line by line, checking for resource leaks, string safety, dead code, etc.
3. Agent finds 5 bugs: zombie process groups (HIGH), UTF-8 truncation (MEDIUM), hardcoded paths (MEDIUM), lost paragraph breaks (LOW), dead branch (LOW)
4. All findings confirmed on first pass — 0 hypothesis failures
5. Audit report written to `.agents/research/2026-02-24-bug-goals-go.md`

**Result:** 5 concrete bugs with severity, file:line, and proposed fix — ready for implementation without debugging.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Can't reproduce bug | Insufficient environment context or intermittent issue | Ask user for specific steps, environment variables, input data. Check for race conditions or timing issues. |
| Git archaeology returns too many commits | Broad search or high-churn file | Narrow timeframe with `--since` flag, focus on specific function with `git blame`, search commit messages for related keywords. |
| Hit 3-failure limit during hypothesis testing | Multiple incorrect hypotheses or complex root cause | Escalate to architecture review. Read `failure-categories.md` to determine if failures are countable. Consider asking for domain expert input. |
| Bug report missing key information | Incomplete investigation or skipped steps | Verify all 4 phases completed. Ensure root cause identified with file:line. Check git blame ran for responsible commit. |

## Reference Documents

- [references/audit-report-template.md](references/audit-report-template.md)
- [references/bug-report-template.md](references/bug-report-template.md)
- [references/failure-categories.md](references/failure-categories.md)

---

## References

### audit-report-template.md

# Audit Report Template

**Write to:** `.agents/research/YYYY-MM-DD-bug-<scope-slug>.md`

```markdown
# Bug Hunt: <Scope Description>

**Date:** YYYY-MM-DD
**Scope:** <files/directories audited>
**Failures:** N (hypothesis tests that didn't confirm)

## Summary

| # | Bug | Severity | File | Fix |
|---|-----|----------|------|-----|
| 1 | <short description> | HIGH | `<file:line>` | <proposed fix> |
| 2 | <short description> | MEDIUM | `<file:line>` | <proposed fix> |

## Findings

### BUG-1: <Title> (SEVERITY)

**File:** `<path:line>`
**Root cause:** <what's wrong and why>

**Observed:** <concrete evidence — error output, test failure, code path trace>

**Fix:** <specific change needed>

### BUG-2: <Title> (SEVERITY)

...
```

## Severity Criteria

| Severity | Criteria | Examples |
|----------|----------|---------|
| **HIGH** | Data loss, security, resource leak, process orphaning | Zombie processes, SQL injection, file handle leak |
| **MEDIUM** | Wrong output, incorrect defaults, silent data corruption | UTF-8 truncation, hardcoded paths, wrong error code |
| **LOW** | Dead code, cosmetic, minor inconsistency | Unreachable branch, unused import, style violation |

### bug-report-template.md

# Bug Report Template

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

### failure-categories.md

# Failure Categories Taxonomy

## Failure Tracking

**Track failures by TYPE - not all failures are equal:**

| Failure Type | Counts Toward Limit? | Action |
|--------------|----------------------|--------|
| `root_cause_not_found` | YES | Re-investigate from Phase 1 |
| `fix_failed_tests` | YES | New hypothesis in Phase 3 |
| `design_rejected` | YES | Rethink approach |
| `execution_timeout` | NO (reset counter) | Retry same approach |
| `external_dependency` | NO (escalate) | Report blocker |

## The 3-Failure Rule

- Count only `root_cause_not_found`, `fix_failed_tests`, `design_rejected`
- After 3 such failures: **STOP and question architecture**
- Output: "3+ fix attempts failed. Escalating to architecture review."
- Do NOT count timeouts or external blockers toward limit

## Track in Issue Notes

```bash
bd update <issue-id> --append-notes "FAILURE: <type> at $(date -Iseconds) - <reason>" 2>/dev/null
```

## Checking Failure Count

```bash
# Count failures (excluding timeouts and external blockers)
failures=$(bd show <issue-id> --json 2>/dev/null | jq '[.notes[]? | select(startswith("FAILURE:")) | select(contains("root_cause") or contains("fix_failed") or contains("design_rejected"))] | length')

if [[ "$failures" -ge 3 ]]; then
    echo "3+ fix attempts failed. Escalating to architecture review."
    bd update <issue-id> --append-notes "ESCALATION: Architecture review needed after 3 failures" 2>/dev/null
    exit 1
fi
```


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: bug-hunt" "grep -q '^name: bug-hunt' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions git log or git blame" "grep -qi 'git log\|git blame' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions root cause" "grep -qi 'root cause' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


