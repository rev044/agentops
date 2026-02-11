---
name: implement
tier: team
description: 'Execute a single beads issue with full lifecycle. Triggers: "implement", "work on task", "fix bug", "start feature", "pick up next issue".'
dependencies:
  - beads     # optional - for issue tracking via bd CLI
  - standards # loads language-specific standards
---

# Implement Skill

> **Quick Ref:** Execute single issue end-to-end. Output: code changes + commit + closed issue.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Execute a single issue from start to finish.

**CLI dependencies:** bd (issue tracking), ao (ratchet gates). Both optional — see `skills/shared/SKILL.md` for fallback table. If bd is unavailable, use the issue description directly and track progress via TaskList instead of beads.

## Execution Steps

Given `/implement <issue-id-or-description>`:

### Step 0: Pre-Flight Checks (Resume + Gates)

**For resume protocol details, read `skills/implement/references/resume-protocol.md`.**

**For ratchet gate checks and pre-mortem gate details, read `skills/implement/references/gate-checks.md`.**

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

Based on the context gathered:

1. **Edit existing files** using the Edit tool (preferred)
2. **Write new files** only if necessary using the Write tool
3. **Follow existing patterns** in the codebase
4. **Keep changes minimal** - don't over-engineer

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

Tell user: "Implementation complete. Run /vibe to validate before pushing."

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

## Distributed Mode: Agent Mail Coordination

**For full distributed mode details, read `skills/implement/references/distributed-mode.md`.**
For standalone distributed runs, start with `scripts/multi-agent-preflight.sh --workflow implement`.

Distributed mode enhances /implement with real-time coordination via MCP Agent Mail when `--mode=distributed`, `--agent-mail`, or `$OLYMPUS_DEMIGOD_ID` is set.

---

## References

- **Agent Mail Protocol:** See `skills/shared/agent-mail-protocol.md` for message format specifications
- **Parser (Go):** `cli/internal/agentmail/` - shared parser for all message types
