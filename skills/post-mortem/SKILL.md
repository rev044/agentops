---
name: post-mortem
description: 'Tools-first post-implementation validation. Runs CI suite (linters, tests, scanners) BEFORE any agent dispatch. Blocks on failures. Agents synthesize tool output, not find issues. Triggers: "post-mortem", "validate completion", "final check", "wrap up epic".'
dependencies:
  - beads  # optional - for issue status
  - retro  # implicit - extracts learnings
---

# Post-Mortem Skill

> **Quick Ref:** CI Suite FIRST (gate) -> Triage Tool Findings -> Knowledge Extraction. Output: `.agents/retros/*.md` + `.agents/learnings/*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**Architecture:** CI suite runs FIRST and gates further work. Tools find issues. Agents synthesize and triage what tools found. Learnings have verification status ("verified: yes/no"), not confidence scores.

## Execution Steps

Given `/post-mortem [epic-id]`:

### Step 1: Run CI Suite (MANDATORY GATE)

**Before ANY other work, run the full CI suite including tests:**

```bash
./scripts/toolchain-validate.sh --gate --json 2>&1 | tee .agents/tooling/run.log
TOOL_EXIT=$?

# Also capture structured output for context
cat .agents/tooling/summary.json
```

**Exit Code Handling:**

| Exit Code | Meaning | Action |
|-----------|---------|--------|
| 0 | All tools pass (including tests) | Proceed to triage |
| 2 | CRITICAL findings (test failures, secrets, security) | **STOP. Report findings. Do NOT proceed.** |
| 3 | HIGH findings only | WARN, proceed with warnings in context |

### Step 1a: Block on CI Failures (Hard Gate)

**If TOOL_EXIT == 2 (CRITICAL):**

```
STOP and report to user:

  Post-Mortem: BLOCKED BY CI SUITE

  CI suite found CRITICAL issues that must be fixed first:
  - See .agents/tooling/summary.json for summary
  - See .agents/tooling/<tool>.txt for details

  Common CRITICAL issues:
  - pytest.txt: Test failures (MUST pass before post-mortem)
  - gotest.txt: Go test failures (MUST pass before post-mortem)
  - gitleaks.txt: Secret leaks detected
  - semgrep.txt: Security vulnerabilities

  Tool Findings (verified by CI):
  | Tool | Status | Finding Count | Details File |
  |------|--------|---------------|--------------|
  | pytest | <status> | <count> | .agents/tooling/pytest.txt |
  | gitleaks | <status> | <count> | .agents/tooling/gitleaks.txt |
  | ... | ... | ... | ... |

  Fix these issues, then re-run /post-mortem
```

**DO NOT proceed to agent triage.** Test failures and critical security issues block agent dispatch. This prevents theater where agents "analyze" code that doesn't even compile or pass tests.

**If TOOL_EXIT == 3 (HIGH only):**

Continue to Step 2, but include tool findings in agent context:
```markdown
## CI Suite Findings (HIGH severity)

The following issues were found by CI. Include in triage context:

<paste structured output from .agents/tooling/summary.json>
```

### Step 2: Identify What Was Completed

**If epic ID provided:** Use it directly.

**If no epic ID:** Find recently completed work:
```bash
bd list --status closed --since "7 days ago" 2>/dev/null | head -5
```

Or check recent git activity:
```bash
git log --oneline --since="24 hours ago" | head -10
```

### Step 3: Build Mechanical Comparison Table

**Generate checklist from plan before reading code:**

If plan exists (`.agents/plans/*.md`):
```
1. Extract all TODO/deliverable items from plan
2. For each item:
   - Expected file: <path>
   - File exists: yes/no
   - Implementation matches spec: yes/no (cite file:line)
```

This creates ground truth for plan-compliance, not gestalt impression.

Write comparison table to prompt text for plan-compliance-expert agent.

### Step 4: Synthesize Tool Output (Agents Receive Tool Data)

**Agents SYNTHESIZE tool output. They do not "find issues" - tools already found issues.**

**First, read tool outputs to provide as agent context:**
```bash
cat .agents/tooling/summary.json
cat .agents/tooling/gitleaks.txt
cat .agents/tooling/semgrep.txt
cat .agents/tooling/ruff.txt
cat .agents/tooling/golangci-lint.txt
cat .agents/tooling/pytest.txt
cat .agents/tooling/gotest.txt
```

**Agent prompts receive structured tool data, not vague instructions:**

```
WRONG: "Review this code and find security issues"
RIGHT: "Here are the semgrep findings. For each, read file:line and classify as true_pos/false_pos"
```

#### 4a. Triage Security Findings (Tool Output as Input)

**Input:** gitleaks.txt, semgrep.txt findings
**Task:** Classify each tool finding

For each finding from gitleaks/semgrep:
1. Read the cited file:line
2. Assess: true positive or false positive?
3. If true positive: severity (CRITICAL/HIGH/MEDIUM/LOW) and fix

Record in table:
| File:Line | Tool Finding | Verdict | Verified | Fix |
|-----------|--------------|---------|----------|-----|
| src/auth.go:42 | hardcoded secret | TRUE_POS | yes (tool: gitleaks) | Remove and use env var |
| tests/mock.py:15 | test credential | FALSE_POS | yes (tool: gitleaks, context: test fixture) | Ignore |

#### 4b. Triage Code Quality Findings (Tool Output as Input)

**Input:** ruff.txt, golangci-lint.txt, radon.txt findings
**Task:** Prioritize each tool finding

For each finding from linters:
1. Read the cited file:line
2. Assess: worth fixing now, tech debt, or noise?
3. If worth fixing: suggest specific change

Record in table:
| File:Line | Tool Finding | Priority | Verified | Suggested Fix |
|-----------|--------------|----------|----------|---------------|

#### 4c. Verify Plan Completion (Comparison Table as Input)

**Input:** Mechanical comparison table from Step 3
**Task:** Classify gaps

For each plan item marked "no" or "partial":
1. Is this a real gap or scope change?
2. Should it be tracked as follow-up issue?

Record in table:
| Plan Item | Status | Gap Type | Verified | Action |
|-----------|--------|----------|----------|--------|
| API endpoint | partial | scope change | yes (commit: abc123) | Document in ADR |

#### 4d. Extract Learnings with Verification Status

**From the completed work and changed files, extract learnings with explicit verification status.**

For each learning:
- ID: L-<date>-<N>
- Category: technical/process/architecture
- What: <1 sentence>
- Source: <file:line or commit hash>
- **Verified:** yes/no (with method)

**Use verification status, NOT confidence scores:**

```
WRONG: "Confidence: 0.92"
RIGHT: "Verified: yes (appeared in 3 files: src/a.go, src/b.go, src/c.go)"
RIGHT: "Verified: yes (commit abc123 shows fix worked)"
RIGHT: "Verified: no (single observation, needs confirmation)"
```

| ID | Learning | Source | Verified |
|----|----------|--------|----------|
| L-2024-01-15-1 | Pre-commit hooks catch 80% of lint issues | .agents/metrics/lint.csv | yes (6 months data) |
| L-2024-01-15-2 | Go context timeout should be 30s | cmd/server/main.go:42 | yes (production incident) |
| L-2024-01-15-3 | Python async needs explicit cleanup | | no (needs more testing) |

### Step 5a: Log Triage Decisions

**For each TRUE_POS or FALSE_POS verdict from agents, log it for accuracy tracking:**

```bash
# Log each triage decision
./scripts/log-triage-decision.sh "src/auth.go:42" "semgrep" "TRUE_POS" "security-reviewer"
./scripts/log-triage-decision.sh "tests/mock.py:15" "gitleaks" "FALSE_POS" "security-reviewer"
```

This enables accuracy tracking over time. Ground truth is added later when:
- CI confirms (test pass/fail)
- Production incident occurs
- Human reviews the decision

**View accuracy report:**
```bash
./scripts/compute-triage-accuracy.sh
```

### Step 5: Synthesize Results

Combine triage outputs:
1. Deduplicate findings by file:line
2. Sort by severity (CRITICAL -> HIGH -> MEDIUM -> LOW)
3. Count verified findings by status

**Grade based on CI SUITE results (tool findings), not agent opinions:**

```
Grade A: CI passed, 0 critical tool findings
Grade B: CI passed, 0 critical, <5 high tool findings
Grade C: CI passed, 0 critical, 5-15 high tool findings
Grade D: CI blocked (1+ critical tool findings)
Grade F: CI blocked (tests failing or multiple critical)
```

**Note:** If CI suite was blocked in Step 1, the grade is automatically D or F. There should be no post-mortem report generated for blocked CI - the user must fix issues first.

### Step 6: Request Human Approval (Gate 4)

```
Tool: AskUserQuestion
Parameters:
  questions:
    - question: "Post-mortem complete. Grade: <grade>. Tool findings triaged. Store learnings?"
      header: "Gate 4"
      options:
        - label: "TEMPER & STORE"
          description: "Learnings are good - lock and index"
        - label: "ITERATE"
          description: "Need another round of fixes"
      multiSelect: false
```

### Step 7: Write Post-Mortem Report

**Write to:** `.agents/retros/YYYY-MM-DD-post-mortem-<topic>.md`

**Only write report if CI suite passed (Step 1).** If CI was blocked, there is no post-mortem - user must fix issues first.

**Merge triage outputs:**
1. Collect tables from triage steps
2. Deduplicate by file:line (within 5 lines tolerance)
3. Sort by severity (CRITICAL -> HIGH -> MEDIUM -> LOW)

```markdown
# Post-Mortem: <Topic/Epic>

**Date:** YYYY-MM-DD
**Epic:** <epic-id or description>
**Duration:** <how long>

## CI Suite Results (Gate)

| Tool | Status | Findings | Details |
|------|--------|----------|---------|
| pytest | PASS/FAIL | <count> | .agents/tooling/pytest.txt |
| go test | PASS/FAIL | <count> | .agents/tooling/gotest.txt |
| gitleaks | PASS/FAIL | <count> | .agents/tooling/gitleaks.txt |
| semgrep | PASS/FAIL | <count> | .agents/tooling/semgrep.txt |
| ruff | PASS/FAIL | <count> | .agents/tooling/ruff.txt |

**Gate Status:** PASS (CI suite passed, post-mortem proceeds)

## Triaged Findings

### True Positives (actionable)
| File:Line | Tool Finding | Severity | Verified | Action |
|-----------|--------------|----------|----------|--------|

### False Positives (dismissed)
| File:Line | Tool Claimed | Verified | Why Dismissed |
|-----------|--------------|----------|---------------|

## Plan Compliance

<Mechanical comparison table from Step 3>

| Plan Item | Delivered | Verified | Notes |
|-----------|-----------|----------|-------|

## Learnings Extracted

| ID | Category | Learning | Source | Verified |
|----|----------|----------|--------|----------|
| L-<date>-1 | technical | <learning> | <file:line> | yes/no (method) |

**Note:** Learnings use "verified: yes/no" status, not confidence scores.

## Follow-up Issues

<Issues created from findings>

## Knowledge Flywheel Status

- **Learnings indexed:** <count from ao forge>
- **Session provenance:** <session-id>
- **ao forge status:** PASS/SKIP (not available)
```

### Step 7a: Index Learnings via ao forge (Knowledge Flywheel)

**If user approved TEMPER & STORE in Gate 4, index learnings into knowledge base:**

```bash
# Check if ao CLI is available
if command -v ao &>/dev/null; then
  # Index learnings from the retro/learnings directory
  ao forge index .agents/learnings/ 2>&1 | tee -a .agents/tooling/ao-forge.log
  AO_EXIT=$?

  # Add provenance tracking - link learnings to this session
  SESSION_ID=$(ao session id 2>/dev/null || echo "unknown")
  echo "Provenance: session=$SESSION_ID, timestamp=$(date -Iseconds)" >> .agents/learnings/provenance.txt

  # Check flywheel status
  FLYWHEEL_STATUS=$(ao flywheel status --json 2>/dev/null || echo '{"indexed": 0}')
  INDEXED_COUNT=$(echo "$FLYWHEEL_STATUS" | jq -r '.indexed // 0')

  if [ $AO_EXIT -eq 0 ]; then
    echo "Flywheel: Learnings indexed successfully"
  else
    echo "Flywheel: ao forge failed (exit $AO_EXIT) - learnings NOT indexed"
  fi
else
  echo "Flywheel: ao CLI not available - learnings written but NOT indexed"
  echo "  Install ao or run manually: ao forge index .agents/learnings/"
fi
```

**Fallback:** If ao is not available, learnings are still written to `.agents/learnings/*.md` but won't be searchable via `ao search`. The skill continues normally.

### Step 8: Report to User

Tell the user:
1. Toolchain results (which tools ran, pass/fail)
2. Grade (based on tool findings)
3. Key triaged findings
4. Learnings extracted
5. Gate 4 decision
6. **Flywheel status** (learnings indexed count)

## Key Differences from Previous Version

| Before | After |
|--------|-------|
| Agents find issues | CI suite finds issues, agents synthesize tool output |
| Agents dispatched first | CI suite runs FIRST (hard gate on failures) |
| Tests run during triage | Tests run BEFORE agent dispatch (blocks on failure) |
| Vague "find issues" prompts | Agents receive structured tool output as input |
| Confidence scores (0.92, 0.91) | Verification status ("verified: yes/no" with method) |
| Always produces report | Blocked if CI fails - no theater |
| Pattern matching validation | Mechanical verification via CI tools |
