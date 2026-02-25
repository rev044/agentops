---
name: implement
description: 'Execute a single issue with full lifecycle. Triggers: "implement", "work on task", "build this", "start feature", "pick up next issue", "work on issue".'
---


# Implement Skill

> **Quick Ref:** Execute single issue end-to-end. Output: code changes + commit + closed issue.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Execute a single issue from start to finish.

**CLI dependencies:** bd (issue tracking), ao (ratchet gates). Both optional — see `skills/shared/SKILL.md` for fallback table. If bd is unavailable, use the issue description directly and track progress via TaskList instead of beads.

## Execution Steps

Given `$implement <issue-id-or-description>`:

### Step 0: Pre-Flight Checks (Resume + Gates)

**For resume protocol details, read `skills/implement/references/resume-protocol.md`.**

**For ratchet gate checks and pre-mortem gate details, read `skills/implement/references/gate-checks.md`.**

### Step 0.5: Pull Relevant Knowledge

```bash
# Pull knowledge scoped to this issue (if ao available)
ao lookup --bead <issue-id> --limit 3 2>/dev/null || true
```

### Step 1: Get Issue Details

**If beads issue ID provided** (e.g., `gt-123`):
```bash
bd show <issue-id> 2>/dev/null
```

**If plain description provided:** Use that as the task description.

**If no argument:** Check for ready work:
```bash
bd ready 2>/dev/null | head -3
```

### Step 2: Claim the Issue

```bash
bd update <issue-id> --status in_progress 2>/dev/null
```

### Step 2a: Build Context Briefing

```bash
if command -v ao &>/dev/null; then
    ao context assemble --task='<issue title and description>'
fi
```

This produces a 5-section briefing (GOALS, HISTORY, INTEL, TASK, PROTOCOL) at `.agents/rpi/briefing-current.md` with secrets redacted. Read it before gathering additional context.

### Step 3: Gather Context

**USE THE TASK TOOL** to explore relevant code:

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  description: "Gather context for: <issue title>"
  prompt: |
    Find code relevant to: <issue description>

    1. Search for related files (Glob)
    2. Search for relevant keywords (Grep)
    3. Read key files to understand current implementation
    4. Identify where changes need to be made

    Return:
    - Files to modify (paths)
    - Current implementation summary
    - Suggested approach
    - Any risks or concerns
```

### Step 4: Implement the Change

**GREEN Mode check:** If test files were provided (invoked by $crank --test-first):
1. Read all provided test files FIRST
2. Read the contract for invariants
3. Implement to make tests pass (do NOT modify test files)
4. Skip to Step 5 verification

Based on the context gathered:

1. **Edit existing files** using the Edit tool (preferred)
2. **Write new files** only if necessary using the Write tool
3. **Follow existing patterns** in the codebase
4. **Keep changes minimal** - don't over-engineer

### Step 4a: Build Verification (CLI repos only)

If the project has a Go `cmd/` directory or a Makefile with a `build` target, run build verification before proceeding to tests:

```bash
# Detect CLI repo
if [ -f go.mod ] && ls cmd/*/main.go &>/dev/null; then
    echo "CLI repo detected — running build verification..."

    # Build
    go build ./cmd/... 2>&1
    if [ $? -ne 0 ]; then
        echo "BUILD FAILED — fix compilation errors before proceeding"
        # Do NOT proceed to Step 5
    fi

    # Vet
    go vet ./cmd/... 2>&1

    # Smoke test: run the binary with --help
    BINARY=$(ls -t cmd/*/main.go | head -1 | xargs dirname | xargs basename)
    if [ -f "bin/$BINARY" ]; then
        ./bin/$BINARY --help > /dev/null 2>&1
        echo "Smoke test: $BINARY --help passed"
    fi
fi
```

**If build fails:** Fix compilation errors and re-run before proceeding. Do NOT skip to verification with a broken build.

**If not a CLI repo:** This step is a no-op — proceed directly to Step 5.

### Step 5: Verify the Change

**Success Criteria (all must pass):**
- [ ] All existing tests pass (no new failures introduced)
- [ ] New code compiles/parses without errors
- [ ] No new linter warnings (if linter available)
- [ ] Change achieves the stated goal

Check for test files and run them:
```bash
# Find tests
ls *test* tests/ test/ __tests__/ 2>/dev/null | head -5

# Run tests (adapt to project type)
# Python: pytest
# Go: go test ./...
# Node: npm test
# Rust: cargo test
```

**If tests exist:** All tests must pass. Any failure = verification failed.

**If no tests exist:** Manual verification required:
- [ ] Syntax check passes (file compiles/parses)
- [ ] Imports resolve correctly
- [ ] Can reproduce expected behavior manually
- [ ] Edge cases identified during implementation are handled

**If verification fails:** Do NOT proceed to Step 5a. Fix the issue first.

### Step 5a: Verification Gate (MANDATORY)

**THE IRON LAW:** NO COMPLETION CLAIMS WITHOUT FRESH VERIFICATION EVIDENCE

Before reporting success, you MUST:

1. **IDENTIFY** - What command proves this claim works?
2. **RUN** - Execute the FULL command (fresh, not cached output)
3. **READ** - Check full output AND exit code
4. **VERIFY** - Does output actually confirm the claim?
5. **ONLY THEN** - Make the completion claim

**Forbidden phrases without fresh verification evidence:**
- "should work", "probably fixed", "seems to be working"
- "Great!", "Perfect!", "Done!" (without output proof)
- "I just ran it" (must run it AGAIN, fresh)

#### Rationalization Table

| Excuse | Reality |
|--------|---------|
| "Too simple to verify" | Simple code breaks. Verification takes 10 seconds. |
| "I just ran it" | Run it AGAIN. Fresh output only. |
| "Tests passed earlier" | Run them NOW. State changes. |
| "It's obvious it works" | Nothing is obvious. Evidence or silence. |
| "The edit looks correct" | Looking != working. Run the code. |

**Store checkpoint:**
```bash
bd update <issue-id> --append-notes "CHECKPOINT: Step 5a verification passed at $(date -Iseconds)" 2>/dev/null
```

### GREEN Mode (Test-First Implementation)

When invoked by $crank with `--test-first`, the worker receives:
- **Failing tests** (immutable — DO NOT modify)
- **Contract** (contract-{issue-id}.md)
- **Issue description**

**GREEN Mode Rules:**

1. **Read failing tests FIRST** — understand what must pass
2. **Read contract** — understand invariants and failure modes
3. **Implement ONLY enough** to make all tests pass
4. **Do NOT modify test files** — tests are immutable in GREEN mode
5. **Do NOT add features** beyond what tests require
6. **BLOCKED if spec error** — if contract contradicts tests or is incomplete, write BLOCKED with reason

**Verification (GREEN Mode):**
1. Run test suite → ALL tests must PASS
2. Standard Iron Law (Step 5a) still applies — fresh verification evidence required
3. No untested code — every line must be reachable by a test

**Test Immutability Enforcement:**
- Workers may ADD new test files but MUST NOT modify existing test files provided by the TEST WAVE
- If a test appears wrong, write BLOCKED with the specific test and reason — do NOT fix it

### Step 6: Commit the Change

If the change is complete and verified:
```bash
git add <modified-files>
git commit -m "<descriptive message>

Implements: <issue-id>"
```

### Step 7: Close the Issue

```bash
bd update <issue-id> --status closed 2>/dev/null
```

### Step 7a: Record Implementation in Ratchet Chain

**After successful issue closure, record in ratchet:**

```bash
# Check if ao CLI is available
if command -v ao &>/dev/null; then
  # Get the commit hash as output artifact
  COMMIT_HASH=$(git rev-parse HEAD 2>/dev/null || echo "")
  CHANGED_FILES=$(git diff --name-only HEAD~1 2>/dev/null | tr '\n' ',' | sed 's/,$//')

  if [ -n "$COMMIT_HASH" ]; then
    # Record successful implementation
    ao ratchet record implement \
      --output "$COMMIT_HASH" \
      --files "$CHANGED_FILES" \
      --issue "<issue-id>" \
      2>&1 | tee -a .agents/ratchet.log

    if [ $? -eq 0 ]; then
      echo "Ratchet: Implementation recorded (commit: ${COMMIT_HASH:0:8})"
    else
      echo "Ratchet: Failed to record - chain.jsonl may need repair"
    fi
  else
    echo "Ratchet: No commit found - skipping record"
  fi
else
  echo "Ratchet: ao CLI not available - implementation NOT recorded"
  echo "  Run manually: ao ratchet record implement --output <commit>"
fi
```

**On failure/blocker:** Record the blocker in ratchet:

```bash
if command -v ao &>/dev/null; then
  ao ratchet record implement \
    --status blocked \
    --reason "<blocker description>" \
    2>/dev/null
fi
```

**Fallback:** If ao is not available, the issue is still closed via bd but won't be tracked in the ratchet chain. The skill continues normally.

### Step 7b: Post-Implementation Ratchet Record

After implementation is complete:

```bash
if command -v ao &>/dev/null; then
  ao ratchet record implement --output "<issue-id>" 2>/dev/null || true
fi
```

Tell user: "Implementation complete. Run $vibe to validate before pushing."

### Step 8: Report to User

Tell the user:
1. What was changed (files modified)
2. How it was verified (with actual command output)
3. Issue status (closed)
4. Any follow-up needed
5. **Ratchet status** (implementation recorded or skipped)

**Output completion marker:**
```
<promise>DONE</promise>
```

If blocked or incomplete:
```
<promise>BLOCKED</promise>
Reason: <why blocked>
```

```
<promise>PARTIAL</promise>
Remaining: <what's left>
```

## Key Rules

- **Explore first** - understand before changing
- **Edit, don't rewrite** - prefer Edit tool over Write tool
- **Follow patterns** - match existing code style
- **Verify changes** - run tests or sanity checks
- **Commit with context** - reference the issue ID
- **Close the issue** - update status when done

## Without Beads

If bd CLI not available:
1. Skip the claim/close status updates
2. Use the description as the task
3. Still commit with descriptive message
4. Report completion to user

---

## Examples

### Implement Specific Issue

**User says:** `$implement ag-5k2`

**What happens:**
1. Agent reads issue from beads: "Add JWT token validation middleware"
2. Explore agent finds relevant auth code and middleware patterns
3. Agent edits `middleware/auth.go` to add token validation
4. Runs `go test ./middleware/...` — all tests pass
5. Commits with message "Add JWT token validation middleware\n\nImplements: ag-5k2"
6. Closes issue via `bd update ag-5k2 --status closed`

**Result:** Issue implemented, verified, committed, and closed. Ratchet recorded.

### Pick Up Next Available Work

**User says:** `$implement`

**What happens:**
1. Agent runs `bd ready` — finds `ag-3b7` (first unblocked issue)
2. Claims issue via `bd update ag-3b7 --status in_progress`
3. Implements and verifies
4. Closes issue

**Result:** Autonomous work pickup and completion from ready queue.

### GREEN Mode (Test-First)

**User says:** `$implement ag-8h3` (invoked by `$crank --test-first`)

**What happens:**
1. Agent receives failing tests (immutable) and contract
2. Reads tests to understand expected behavior
3. Implements ONLY enough to make tests pass
4. Does NOT modify test files
5. Verification: all tests pass with fresh output

**Result:** Minimal implementation driven by tests, no over-engineering.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Issue not found | Issue ID doesn't exist or beads not synced | Run `bd sync` then `bd show <id>` to verify |
| GREEN mode violation | Edited a file not related to the issue scope | Revert unrelated changes. GREEN mode restricts edits to files relevant to the issue |
| Verification gate fails | Tests fail or build breaks after implementation | Read the verification output, fix the specific failures, re-run verification |
| "BLOCKED" status | Contract contradicts tests or is incomplete in GREEN mode | Write BLOCKED with specific reason, do NOT modify tests |
| Fresh verification missing | Agent claims success without running verification command | MUST run verification command fresh with full output before claiming completion |
| Ratchet record failed | ao CLI unavailable or chain.jsonl corrupted | Implementation still closes via bd, but ratchet chain needs manual repair |

## Reference Documents

- [references/gate-checks.md](references/gate-checks.md)
- [references/resume-protocol.md](references/resume-protocol.md)

---

## References

### gate-checks.md

# Gate Checks

> Extracted from implement SKILL.md Steps 0a-0b. Ratchet gate checks and pre-mortem validation prerequisites.

## Ratchet Status Check (RPI Workflow)

**Before implementation, verify prior workflow gates passed:**

```bash
# Check if ao CLI is available
if command -v ao &>/dev/null; then
  # Check if research and plan phases completed
  RATCHET_STATUS=$(ao ratchet status --json 2>/dev/null || echo '{}')
  RESEARCH_DONE=$(echo "$RATCHET_STATUS" | jq -r '.research.completed // false')
  PLAN_DONE=$(echo "$RATCHET_STATUS" | jq -r '.plan.completed // false')

  if [ "$RESEARCH_DONE" = "true" ] && [ "$PLAN_DONE" = "true" ]; then
    echo "Ratchet: Prior gates passed (research + plan complete)"
  elif [ "$RESEARCH_DONE" = "false" ] || [ "$PLAN_DONE" = "false" ]; then
    echo "WARNING: Prior gates not complete. Run $research and $plan first."
    echo "  Research: $RESEARCH_DONE"
    echo "  Plan: $PLAN_DONE"
    echo ""
    echo "Override with: ao ratchet skip <gate> --reason 'manual override'"
  fi

  # Get current spec path for reference
  SPEC_PATH=$(ao ratchet spec 2>/dev/null || echo "")
  if [ -n "$SPEC_PATH" ]; then
    echo "Ratchet: Current spec at $SPEC_PATH"
  fi
else
  echo "Ratchet: ao CLI not available - skipping gate check"
fi
```

**Fallback:** If ao is not available, proceed without ratchet checks. The skill continues normally.

## Pre-Flight Pre-Mortem Gate

**Before starting implementation, check if pre-mortem validation was run on the plan:**

```bash
if command -v ao &>/dev/null; then
  RATCHET_JSON=$(ao ratchet status --json 2>/dev/null || echo '{}')
  PRE_MORTEM_STATUS=$(echo "$RATCHET_JSON" | jq -r '.steps[]? | select(.name == "pre-mortem") | .status // "none"')
  PLAN_EXISTS=$(ls .agents/plans/*.md 2>/dev/null | head -1)

  if [ "$PRE_MORTEM_STATUS" = "pending" ] && [ -n "$PLAN_EXISTS" ]; then
    echo "Pre-mortem hasn't been run on your plan."
    echo "Options:"
    echo "  1. Run $pre-mortem first"
    echo "  2. Skip: ao ratchet skip pre-mortem --reason 'user chose to skip'"
    echo "  3. Proceed anyway"
    # Ask user: "Pre-mortem hasn't been run on your plan. Run $pre-mortem first, skip, or proceed?"
    # If skip: ao ratchet skip pre-mortem --reason "user chose to skip"
  fi
  # If ao unavailable or no chain: proceed silently
fi
```

**Fallback:** If ao is not available or no ratchet chain exists, proceed silently.

### resume-protocol.md

# Resume Protocol

> Extracted from implement SKILL.md Step 0. Handles session continuation and checkpoint detection.

## Check Issue State (Resume Logic)

Before starting implementation, check if resuming:

1. **Check if issue is in_progress:**
```bash
bd show <issue-id> --json 2>/dev/null | jq -r '.status'
```

2. **If status = in_progress AND assigned to you:**
   - Look for checkpoint in issue notes: `bd show <id> --json | jq -r '.notes'`
   - Resume from last checkpoint step
   - Announce: "Resuming issue from Step N"

3. **If status = in_progress AND assigned to another agent:**
   - Report: "Issue claimed by <agent> - use `bd update <id> --assignee self --force` to override"
   - Do NOT proceed without explicit override

4. **Store checkpoints after each major step:**
```bash
bd update <issue-id> --append-notes "CHECKPOINT: Step N completed at $(date -Iseconds)" 2>/dev/null
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
check "SKILL.md has name: implement" "grep -q '^name: implement' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 2 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 2 ]"
check "SKILL.md mentions bd for issue tracking" "grep -q 'bd ' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions beads" "grep -qi 'beads' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions $vibe for validation" "grep -q '$vibe' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions Explore agent" "grep -qi 'explore' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions verification gate" "grep -qi 'verification\|verify' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions ratchet record" "grep -q 'ratchet record' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions GREEN mode" "grep -q 'GREEN' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions DONE/BLOCKED/PARTIAL markers" "grep -q 'DONE\|BLOCKED\|PARTIAL' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


