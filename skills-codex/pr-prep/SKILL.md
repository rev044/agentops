---
name: pr-prep
description: 'PR preparation: git archaeology, test validation, structured PR body generation. Mandatory user review gate before submission. Triggers: "prepare PR", "PR prep", "submit PR", "create PR body", "write PR description".'
---


# PR Preparation Skill

Systematic PR preparation that validates tests and generates high-quality PR bodies.

## Overview

Prepares contributions by analyzing the target repo's conventions, git history,
test coverage, and generating properly-formatted PR bodies.

**When to Use**:
- Preparing a PR for an external repository
- Contributing bug fixes or features

**When NOT to Use**:
- Internal commits (use normal git workflow)
- PRs to your own repositories

---

## Workflow

```
-1. Prior Work Check     -> BLOCKING: Final check for competing PRs
0.  Isolation Check      -> BLOCK if PR mixes unrelated changes
1.  Context Discovery    -> Understand target repo conventions
2.  Git Archaeology      -> Analyze commit patterns, PR history
3.  Pre-Flight Checks    -> Run tests, linting, build
4.  Change Analysis      -> Summarize what changed and why
5.  PR Body Generation   -> Create structured PR description
6.  USER REVIEW GATE     -> STOP. User must approve before submission.
7.  Submission           -> Only after explicit user approval
```

---

## Phase 0: Isolation Check (BLOCKING)

**CRITICAL**: Run this FIRST. Do not proceed if PR mixes unrelated changes.

### Commit Type Analysis

```bash
# Extract commit type prefixes from branch
git log --oneline main..HEAD | sed 's/^[^ ]* //' | grep -oE '^[a-z]+(\([^)]+\))?:' | sort -u
```

**Rule**: If more than one commit type prefix exists, the PR is mixing concerns.

### File Theme Analysis

```bash
# List all files changed vs main
git diff --name-only main..HEAD

# Group by directory
git diff --name-only main..HEAD | cut -d'/' -f1-2 | sort -u
```

### Isolation Checklist

| Check | Pass Criteria |
|-------|---------------|
| **Single commit type** | All commits share same prefix |
| **Thematic files** | All changed files relate to PR scope |
| **No main overlap** | Changes not already merged |
| **Atomic scope** | Can explain in one sentence |

**DO NOT PROCEED IF ISOLATION CHECK FAILS.**

---

## CRITICAL: User Review Gate

**NEVER submit a PR without explicit user approval.**

After generating the PR body (Phase 5), ALWAYS:

1. Write the PR body to a file for review
2. Show the user what will be submitted
3. **STOP and ask**: "Ready to submit? Review the PR body above."
4. Wait for explicit approval before running `gh pr create`

```bash
# Write PR body to file
cat > /tmp/pr-body.md << 'EOF'
<generated PR body>
EOF

# Show user
cat /tmp/pr-body.md

# ASK - do not proceed without answer
echo "Review complete. Submit this PR? [y/N]"
```

---

## Phase 3: Pre-Flight Checks

```bash
# Go projects
go build ./...
go vet ./...
go test ./... -v -count=1

# Node projects
npm run build
npm test

# Python projects
pytest -v
```

### Pre-Flight Checklist

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] No new linting warnings
- [ ] No secrets or credentials in code

---

## Phase 5: PR Body Generation

### Standard Format

```markdown
## Summary

Brief description of WHAT changed and WHY. 1-3 sentences.
Start with action verb (Add, Fix, Update, Refactor).

## Changes

Technical details of what was modified.

## Test plan

- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] Manual: <specific scenario tested>

Fixes #NNN
```

**Key conventions:**
- Test plan items are **checked** `[x]` (you ran them before PR)
- `Fixes #NNN` goes at the end

---

## Phase 7: Submission (After Approval Only)

```bash
# Create PR with reviewed body
gh pr create --title "type(scope): brief description" \
  --body "$(cat /tmp/pr-body.md)" \
  --base main
```

**Remember**: This command should ONLY run after user explicitly approves.

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| **Submit without approval** | **ALWAYS stop for user review** |
| **Skip isolation check** | **Run Phase 0 FIRST** |
| Bundle lint fixes into feature PRs | Lint fixes get their own PR |
| Giant PRs | Split into logical chunks |
| Vague PR body | Detailed summary with context |
| Skip pre-flight | Always run tests locally |

## Examples

### Prepare External PR Body

**User says:** "Prepare this branch for PR submission."

**What happens:**
1. Run isolation and pre-flight validation.
2. Build structured PR body with summary and test plan.
3. Pause for mandatory user review before submit.

### Evidence-First PR Packaging

**User says:** "Generate a high-quality PR description with clear verification steps."

**What happens:**
1. Gather git archaeology and test evidence.
2. Synthesize concise rationale and change list.
3. Produce submit-ready body pending approval.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| PR body is weak | Missing context from commits/tests | Re-run evidence collection and expand summary |
| Submission blocked | Mandatory review gate not passed | Get explicit user approval before `gh pr create` |
| Test plan incomplete | Commands/results not captured | Add executed checks and outcomes explicitly |
| Title/body mismatch | Scope drift during edits | Regenerate from latest branch diff and constraints |

## Reference Documents

- [references/case-study-historical-context.md](references/case-study-historical-context.md)
- [references/lessons-learned.md](references/lessons-learned.md)
- [references/package-extraction.md](references/package-extraction.md)

---

## References

### case-study-historical-context.md

# Case Study: PR #512 - Historical Context in Action

This PR contradicted an existing comment that said "do NOT set BEADS_DIR". Here's how historical context research proved the comment was wrong and got the PR accepted.

## The Problem

The code had a comment: `// do NOT set BEADS_DIR as that overrides routing and breaks resolution of rig-level beads`

But the PR needed to set BEADS_DIR to fix a bug. How to convince reviewers the change doesn't break what the comment warns about?

## Git Archaeology Commands Used

```bash
# 1. Find when the comment was introduced
git log -p --all -S "do NOT set BEADS_DIR" -- internal/cmd/sling_helpers.go

# 2. Find the commit that added it
git show 52ef89c5 --stat
# Result: "fix(sling): use bd native prefix routing instead of BEADS_DIR override"

# 3. Check if there were subsequent fixes
git log --oneline 52ef89c5..HEAD -- internal/beads/beads.go
# Found: 598a39e7 "fix: prevent inherited BEADS_DIR from causing prefix mismatch"

# 4. Compare the approaches
git show 598a39e7 --stat
# Result: This commit ALWAYS sets BEADS_DIR - opposite of what 52ef89c5 said!

# 5. Find who wrote the original helper functions
gh pr list --author boshu2 --state all --limit 20
# Found: PR #149 added ExtractPrefix, GetRigPathForPrefix
```

## Timeline Built

| Date | Commit/PR | Author | What Happened |
|------|-----------|--------|---------------|
| Jan 5 | PR #149 | boshu2 | Added correct helper functions |
| Jan 6 | 52ef89c5 | jack | Wrong fix: "do NOT set BEADS_DIR" (didn't use helpers) |
| Jan 11 | 598a39e7 | joe | Correct fix in beads.go: ALWAYS set BEADS_DIR |
| Jan 14 | PR #512 | boshu2 | Applies correct pattern using original helpers |

## Why This Worked

1. **Proved the comment was outdated** - It was from Jan 6, superseded on Jan 11
2. **Showed the correct pattern existed** - 598a39e7 established ALWAYS setting BEADS_DIR
3. **Connected to original work** - PR #149 had the helpers that should have been used
4. **Provided timeline** - Made it easy for reviewers to verify the history

## Key Takeaway

When contradicting existing code or comments:
1. Trace when it was introduced (`git log -S`)
2. Find if there were subsequent fixes
3. Build a timeline showing the evolution
4. Include this in the PR body

### lessons-learned.md

# Lessons Learned (Real PR Outcomes)

Based on actual PR submissions to steveyegge/gastown.

## What Got Accepted

| PR | Type | Why It Worked |
|----|------|---------------|
| #353 | refactor | **Single focus** - only touched one package (suggest.go), clear value prop |
| #149 | fix | **Solved real bug** - cross-rig beads weren't routing correctly |
| #512 | fix | **Historical context** - traced why old comment was wrong, built timeline |

**Patterns that work:**
- Small, focused changes (1 file or 1 package)
- Clear problem → solution narrative
- Tests pass, no lint issues
- Follows existing code conventions
- **When contradicting old code: provide timeline proving it was wrong**

## What Got Rejected

| PR | Type | Why It Failed |
|----|------|---------------|
| #236 | fix | **Wrong abstraction** - coupled refinery to convoy (ZFC violation) |
| #145 | fix | **Superseded by architecture change** - feature designed out |
| #118 | docs | **No feedback** - possibly stale or not needed |

**Patterns to avoid:**
- Fixing symptoms not root causes
- Adding coupling between components
- Docs for features that might change
- PRs during active architecture churn

## How to Improve Acceptance Rate

1. **Understand the architecture first**
   ```bash
   git log --oneline -20
   gh issue list --state open
   ```

2. **Ask before big changes**
   - Open an issue or discussion first
   - Propose approach, get feedback
   - Especially for architectural changes

3. **Target stable areas**
   - Refactors of established code (like #353)
   - Bug fixes with clear reproduction
   - Tests and docs for stable features
   - Avoid areas under active development

4. **Small PRs win**
   - 1 file > 5 files
   - 1 concern > 3 concerns
   - Easier to review = faster merge

5. **Track stats**
   ```bash
   gh pr list --author @me --state all --json state | \
     jq 'group_by(.state) | map({state: .[0].state, count: length})'
   ```

## Contribution Tracking

Not all contributions show on GitHub:
- **PRs merged via GitHub** → Shows on profile
- **Cherry-picked PRs** → Code ships, PR shows "closed"
- **Direct commits** → Only if email matches GitHub account

To ensure GitHub tracks contributions:
```bash
git config user.email "your-github-email@example.com"
```

### package-extraction.md

# Package Extraction Template

For extracting packages as standalone libraries.

## Pre-Extraction Checklist

- [ ] Package has minimal internal dependencies
- [ ] Test coverage > 50%
- [ ] Public API is clean and documented
- [ ] No hardcoded paths or configs
- [ ] External dependencies are minimal and stable

## Extraction Steps

1. **Copy Source**
   ```bash
   cp -r internal/package/ /new/repo/
   ```

2. **Remove Internal Imports**
   ```bash
   grep -r "github.com/original/repo/internal" /new/repo/
   # Remove or replace each import
   ```

3. **Update Module**
   ```bash
   go mod init github.com/you/package
   go mod tidy
   ```

4. **Verify Independence**
   ```bash
   go build ./...
   go test ./...
   ```

5. **Add Standalone Tests**
   - Integration tests that don't require original repo


---

## Scripts

### validate.sh

```bash
#!/bin/bash
# Validate pr-prep skill
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0

check() {
    local desc="$1"
    local cmd="$2"
    local expected="$3"

    CHECKS=$((CHECKS + 1))
    if eval "$cmd" 2>/dev/null | grep -qi "$expected"; then
        echo "✓ $desc"
    else
        echo "✗ $desc"
        echo "  Command: $cmd"
        echo "  Expected to find: $expected"
        ERRORS=$((ERRORS + 1))
    fi
}

check_pattern() {
    local desc="$1"
    local file="$2"
    local pattern="$3"

    CHECKS=$((CHECKS + 1))
    if grep -qiE "$pattern" "$file" 2>/dev/null; then
        echo "✓ $desc"
    else
        echo "✗ $desc (pattern '$pattern' not found in $file)"
        ERRORS=$((ERRORS + 1))
    fi
}

check_exists() {
    local desc="$1"
    local path="$2"

    CHECKS=$((CHECKS + 1))
    if [ -e "$path" ]; then
        echo "✓ $desc"
    else
        echo "✗ $desc ($path not found)"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== PR Prep Skill Validation ==="
echo ""


# Verify git is available
check "git binary exists" "which git" "git"

# Verify dependent skill exists
check_exists "Beads skill exists" "$HOME/.claude/skills/beads/SKILL.md"

# Verify pr-prep workflow patterns in SKILL.md
check_pattern "SKILL.md has git archaeology" "$SKILL_DIR/SKILL.md" "git|[Aa]rchaeology"
check_pattern "SKILL.md has test validation" "$SKILL_DIR/SKILL.md" "[Tt]est.*[Vv]alid"
check_pattern "SKILL.md has PR body generation" "$SKILL_DIR/SKILL.md" "PR.*[Bb]ody|[Bb]ody.*PR"
check_pattern "SKILL.md has user review gate" "$SKILL_DIR/SKILL.md" "[Rr]eview.*[Gg]ate|MANDATORY.*[Rr]eview"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: PR-prep skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: PR-prep skill validation passed"
    exit 0
fi
```


