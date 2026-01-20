---
name: pr-implement
description: >
  Fork-based implementation with isolation check. Runs isolation check from pr-prep
  Phase 0 before starting work. Triggers: "implement PR", "implement contribution",
  "pr implement", "fork implementation".
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - pr-plan
  - pr-prep
---

# PR Implement Skill

Fork-based implementation for open source contributions with mandatory isolation check.

## Overview

Execute a contribution plan with fork isolation. This skill ensures PRs are clean
and focused by running isolation checks before and during implementation.

**Input**: Plan artifact from `/pr-plan` or repo URL

**When to Use**:
- Implementing a planned OSS contribution
- Need isolation enforcement for clean PRs
- Working in a fork worktree
- After completing `/pr-plan`

**When NOT to Use**:
- Internal project work (use `/implement`)
- Haven't planned yet (run `/pr-plan` first)
- Trivial changes (implement directly)

---

## Workflow

```
-1. Prior Work Check      -> BLOCKING: Check for competing PRs before starting
0.  Input Discovery       -> Find plan artifact or repo
1.  Fork Setup            -> Ensure fork exists and is current
2.  Worktree Creation     -> Create isolated worktree for changes
3.  Isolation Pre-Check   -> BLOCK if existing branch has mixed concerns
4.  Implementation        -> Execute plan with progress tracking
5.  Isolation Post-Check  -> BLOCK if implementation mixed concerns
6.  Commit Preparation    -> Stage changes with proper commit type
7.  Handoff               -> Ready for /pr-prep
```

---

## Phase -1: Prior Work Check (BLOCKING)

**CRITICAL**: Before starting implementation, verify no one else is working on this.

### -1.1 Check for Competing PRs

```bash
# Get target repo from plan or context
REPO="<owner/repo>"
TOPIC="<topic from plan>"

# Search for open PRs on this topic
gh pr list -R $REPO --state open --search "$TOPIC" --limit 10

# Check for PRs touching same files
gh pr list -R $REPO --state open --json files,number,title | \
  jq -r '.[] | select(.files[].path | contains("<target-file>")) | "#\(.number): \(.title)"'
```

### -1.2 Check Target Issue Status

```bash
# If implementing a specific issue, verify it's still available
gh issue view <issue-number> -R $REPO --json state,assignees,comments

# Check if someone claimed it
gh issue view <issue-number> -R $REPO --json comments --jq '.comments[-3:][].body' | grep -i "working on\|taking this\|claimed"
```

### -1.3 Prior Work Checklist

| Check | Finding | Action |
|-------|---------|--------|
| Open PR exists | PR #N on same topic | Coordinate or wait |
| Issue assigned | Someone claimed it | Coordinate or find alternative |
| Recent comment "working on this" | Active work | Coordinate or wait |
| No competing work | Clean | Proceed to Phase 0 |

### -1.4 If Competing Work Found

**DO NOT PROCEED** without resolution:

1. **Open PR exists**: Comment asking status, offer to help review instead
2. **Issue claimed**: Find alternative issue or coordinate
3. **Recent "working on it"**: Wait or ask for status update

**Only proceed to Phase 0 if no competing work exists.**

---

## Phase 0: Input Discovery

**Identify the plan artifact or repo to implement.**

### 0.1 If Plan Path Provided

```bash
# Verify plan artifact exists
ls -la "$INPUT"
cat "$INPUT" | head -30

# Extract repo and scope information
grep -E "^upstream:|^## Scope|^## Target" "$INPUT"
```

### 0.2 If Repo URL Provided

```bash
# Check for existing plan
ls ~/gt/.agents/*/plans/ 2>/dev/null | xargs grep -l "<repo-name>" 2>/dev/null

# If no plan found
echo "No plan found. Consider running /pr-plan first for structured approach."
```

### 0.3 Validation Checklist

| Check | Action |
|-------|--------|
| Plan exists | Extract scope, target, strategy |
| No plan | Suggest `/pr-plan` or proceed with caution |
| Repo accessible | Verify fork exists or can be created |

---

## Phase 1: Fork Setup

**Ensure fork exists and is synced with upstream.**

### 1.1 Check Fork Status

```bash
# Check if fork exists
gh repo view <your-username>/<repo> --json name 2>/dev/null

# If no fork, create one
gh repo fork <upstream/repo> --clone=false

# Verify fork is up-to-date with upstream
cd <fork-directory>
git fetch upstream
git log upstream/main..main --oneline
```

### 1.2 Sync Fork

```bash
# If behind upstream
git checkout main
git merge upstream/main --ff-only
git push origin main
```

### 1.3 Fork Checklist

| Check | Pass Criteria |
|-------|---------------|
| Fork exists | `gh repo view` succeeds |
| Fork synced | No commits behind upstream/main |
| Remote configured | `upstream` remote points to original |

---

## Phase 2: Worktree Creation

**Create isolated worktree for implementation.**

### 2.1 Create Feature Branch

```bash
# Determine branch name from plan scope
BRANCH_NAME="type/scope-description"  # e.g., refactor/suggest-validation

# Create branch from latest main
git checkout main
git pull upstream main
git checkout -b $BRANCH_NAME
```

### 2.2 Create Worktree (Optional)

For large contributions, use separate worktree:

```bash
# Create worktree
git worktree add ~/gt/.forks/<repo>-<feature> $BRANCH_NAME

# Work in isolation
cd ~/gt/.forks/<repo>-<feature>
```

### 2.3 Worktree Checklist

| Check | Pass Criteria |
|-------|---------------|
| Branch created | Based on latest main |
| Clean state | No uncommitted changes |
| Correct directory | Working in fork, not upstream |

---

## Phase 3: Isolation Pre-Check (BLOCKING)

**CRITICAL: Run isolation check BEFORE starting implementation.**

This is Phase 0 from pr-prep, run proactively.

### 3.1 Commit Type Analysis

```bash
# If branch has existing commits (resuming work)
git log --oneline main..HEAD | sed 's/^[^ ]* //' | grep -oE '^[a-z]+(\([^)]+\))?:' | sort -u

# Expected: empty (new branch) or single prefix
# RED FLAG: multiple prefixes = mixed concerns
```

### 3.2 File Theme Analysis

```bash
# If changes exist
git diff --name-only main..HEAD | cut -d'/' -f1-2 | sort -u

# All files should relate to PR scope from plan
# RED FLAG: unrelated directories = scope creep
```

### 3.3 Main Divergence Check

```bash
# What's been added to main since branch created?
git log --oneline HEAD..upstream/main

# If significant changes, consider rebasing before continuing
git fetch upstream
git rebase upstream/main
```

### 3.4 Isolation Pre-Check Checklist

| Check | Pass Criteria | Action if Fail |
|-------|---------------|----------------|
| Single commit type | 0 or 1 prefix | Split into separate branches |
| Thematic files | All match plan scope | Reset and re-commit |
| No main overlap | Changes not on main | Rebase to drop redundant |
| Branch fresh | Based on recent main | Rebase on upstream/main |

**DO NOT PROCEED TO PHASE 4 IF PRE-CHECK FAILS.**

### 3.5 Resolution Actions

If pre-check fails:

```bash
# Mixed commit types: Start fresh branch
git checkout main
git checkout -b $BRANCH_NAME-clean
git cherry-pick <only-relevant-commits>

# Unrelated files: Interactive rebase
git rebase -i main
# Mark unrelated commits as "drop"

# Already on main: Drop redundant commits
git rebase -i main
# Remove commits whose changes are on main
```

---

## Phase 4: Implementation

**Execute the plan with progress tracking.**

### 4.1 Load Plan Context

```bash
# Read scope and target from plan
cat "$PLAN_FILE" | grep -A 20 "## Scope"
cat "$PLAN_FILE" | grep -A 10 "## Implementation Strategy"
```

### 4.2 CONTRIBUTING.md Compliance Check

**CRITICAL**: Before writing any code, verify CONTRIBUTING.md requirements:

```bash
# Re-read CONTRIBUTING.md for implementation requirements
cat CONTRIBUTING.md 2>/dev/null || \
cat .github/CONTRIBUTING.md 2>/dev/null || \
cat docs/CONTRIBUTING.md 2>/dev/null

# Extract key implementation requirements
grep -iE "test|lint|style|format|sign-off|DCO" CONTRIBUTING.md .github/CONTRIBUTING.md 2>/dev/null | head -20
```

| Requirement | Check Before Commit |
|-------------|---------------------|
| **Code style** | Does your code match project conventions? |
| **Tests required** | Are you adding tests as required? |
| **Documentation** | Does CONTRIBUTING.md require doc updates? |
| **Sign-off** | Does project require DCO sign-off on commits? |

### 4.3 Implementation Guidelines

| Guideline | Why |
|-----------|-----|
| **Single concern** | Each commit = one logical change |
| **Match conventions** | Follow project style exactly (from CONTRIBUTING.md) |
| **Test incrementally** | Run tests after each change |
| **Document progress** | Update plan with status |

### 4.4 Commit Convention

Use conventional commits matching project:

```bash
# Check project convention
git log --oneline -10

# Standard format
git commit -m "type(scope): brief description

Longer explanation if needed.

Related: #issue-number"
```

### 4.4 Progress Tracking

Update plan file during implementation:

```markdown
## Implementation Progress

- [x] Step 1: Setup complete
- [x] Step 2: Core implementation done
- [ ] Step 3: Tests pending
- [ ] Step 4: Documentation update
```

---

## Phase 5: Isolation Post-Check (BLOCKING)

**Re-run isolation check after implementation.**

### 5.1 Full Isolation Validation

```bash
# Commit type analysis
git log --oneline main..HEAD | sed 's/^[^ ]* //' | grep -oE '^[a-z]+(\([^)]+\))?:' | sort -u

# File theme analysis
git diff --name-only main..HEAD

# Summary stats
git diff --stat main..HEAD
```

### 5.2 Post-Check Checklist

| Check | Pass Criteria |
|-------|---------------|
| **Single commit type** | All commits share same prefix |
| **Thematic files** | All files relate to PR title/scope |
| **No main overlap** | No redundant changes |
| **Atomic scope** | Can explain PR in one sentence |

### 5.3 Post-Check Resolution

If post-check fails:

| Issue | Resolution |
|-------|------------|
| Mixed types | Interactive rebase to squash or split |
| Unrelated files | `git reset` and selectively re-commit |
| Scope creep | Extract unrelated work to new branch |

```bash
# Example: Clean up mixed commits
git rebase -i main
# Squash related commits, drop unrelated

# Example: Extract scope creep
git checkout -b $BRANCH_NAME-extra
git cherry-pick <unrelated-commits>
git checkout $BRANCH_NAME
git rebase -i main  # drop those commits
```

**DO NOT PROCEED TO PHASE 6 IF POST-CHECK FAILS.**

---

## Phase 6: Commit Preparation

**Prepare commits for PR submission.**

### 6.1 Review Commits

```bash
# View all commits
git log --oneline main..HEAD

# View combined diff
git diff main..HEAD --stat

# Verify commit messages
git log main..HEAD --format="%s" | head -5
```

### 6.2 Squash if Needed

For cleaner history:

```bash
# Squash to single commit (common for small PRs)
git rebase -i main
# Mark all but first as "squash"

# Write clean commit message
git commit --amend
```

### 6.3 Push to Fork

```bash
# Push feature branch to fork
git push -u origin $BRANCH_NAME

# Verify
git log origin/$BRANCH_NAME --oneline -5
```

---

## Phase 7: Handoff

**Ready for PR submission via /pr-prep.**

### 7.1 Handoff Checklist

| Check | Status |
|-------|--------|
| Isolation checks pass | |
| All commits use consistent type | |
| Branch pushed to fork | |
| Tests pass locally | |
| Plan updated with progress | |

### 7.2 Output to User

```
Implementation complete. Isolation checks passed.

Branch: origin/$BRANCH_NAME
Commits: N commits, +X/-Y lines
Files: [list affected files]

Next step: /pr-prep

Isolation summary:
- Commit type: type(scope):
- Files changed: [thematic summary]
- Scope: [one sentence description]
```

### 7.3 Automatic Next Step

```bash
# Suggest pr-prep
echo "Ready to submit? Run: /pr-prep"
```

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| **Skip isolation pre-check** | Run Phase 3 FIRST - always |
| **Skip isolation post-check** | Run Phase 5 before push |
| **Mix concerns in commits** | One type prefix per PR |
| **Implement without plan** | Run /pr-plan first |
| **Push without tests** | Run tests locally |
| **Large monolithic commits** | Small, atomic commits |
| **Work in upstream clone** | Use fork worktree |

---

## Quick Example

**User**: `/pr-implement ~/gt/.agents/_oss/plans/2026-01-12-pr-plan-gastown.md`

**Agent workflow**:

```bash
# Phase 0: Input Discovery
cat ~/gt/.agents/_oss/plans/2026-01-12-pr-plan-<project>.md | head -30
# Found: Plan for <project> suggest validation

# Phase 1: Fork Setup
gh repo view <username>/<fork-repo> --json name  # Fork exists
cd ~/projects/<project>-fork
git fetch upstream

# Phase 2: Worktree Creation
git checkout main && git pull upstream main
git checkout -b refactor/suggest-validation

# Phase 3: Isolation Pre-Check
git log --oneline main..HEAD  # Empty (new branch) - PASS
# Pre-check PASSES

# Phase 4: Implementation
# Edit internal/suggest/suggest.go
# Add validation logic following plan

# Phase 5: Isolation Post-Check
git log --oneline main..HEAD
# refactor(suggest): add formula validation
# Single type, single scope - PASS

git diff --name-only main..HEAD
# internal/suggest/suggest.go
# internal/suggest/suggest_test.go
# Thematic files - PASS

# Phase 6: Commit Preparation
git push -u origin refactor/suggest-validation

# Phase 7: Handoff
# "Implementation complete. Ready for /pr-prep"
```

---

## References

- **PR Plan**: `pr-plan/SKILL.md`
- **PR Prep**: `pr-prep/SKILL.md`
- **Isolation Check**: pr-prep Phase 0
- **Internal Implementation**: `implement/SKILL.md`
