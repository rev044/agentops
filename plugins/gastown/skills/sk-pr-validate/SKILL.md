---
name: sk-pr-validate
description: >
  PR-specific validation: isolation, upstream alignment, quality, scope creep detection.
  Reuses sk-pr-prep Phase 0. Triggers: "validate PR", "pr validation", "check PR",
  "scope creep", "isolation check".
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Bash,Grep,Glob"
skills:
  - sk-pr-prep
---

# PR Validate Skill

PR-specific validation that ensures changes are clean, focused, and ready for submission.
Runs before /pr-prep to catch issues early.

## Overview

Validates a PR branch for submission readiness by checking isolation, upstream alignment,
scope containment, and basic quality gates.

**Input**: Branch name (default: current branch)

**When to Use**:
- Before running /pr-prep
- After /pr-implement completes
- When suspicious of scope creep
- Before any PR submission

**When NOT to Use**:
- Internal project validation (use /validation-chain)
- Already submitted PR (validate before submission)
- Trivial single-file changes (skip validation)

---

## Workflow

```
1.  Branch Discovery     -> Identify branch and upstream
1b. PR Status Check      -> If PR exists, check draft status (WARNING)
2.  Upstream Alignment   -> FIRST: Check rebase status (BLOCKING)
3.  CONTRIBUTING.md      -> Verify compliance with contribution guidelines (BLOCKING)
4.  Isolation Check      -> Single type, thematic files (Phase 0 from sk-pr-prep)
5.  Scope Check          -> Verify changes match intended scope
6.  Quality Gate         -> Tests, linting (non-blocking)
7.  Report Generation    -> Summary with pass/fail and resolutions
```

**Critical**: Upstream alignment is checked FIRST. No point analyzing isolation/scope on a stale branch.

---

## Phase 1: Branch Discovery

**Identify the branch to validate and upstream reference.**

### 1.1 Determine Branch

```bash
# Current branch if not specified
BRANCH=$(git branch --show-current)

# Upstream reference (default: main)
UPSTREAM="${UPSTREAM:-main}"

# Verify branch exists
git rev-parse --verify $BRANCH
git rev-parse --verify $UPSTREAM
```

### 1.2 Branch Context

```bash
# When was branch created?
git merge-base $UPSTREAM $BRANCH

# Commits since branch point
git log --oneline $UPSTREAM..$BRANCH

# Summary stats
echo "Branch: $BRANCH"
echo "Upstream: $UPSTREAM"
echo "Commits: $(git rev-list --count $UPSTREAM..$BRANCH)"
```

---

## Phase 1b: PR Status Check (WARNING)

**If a PR already exists for this branch, check its status.**

Draft PRs won't be reviewed. Catch this before wasting time on other validation.

### 1b.1 Check for Existing PR

```bash
# Check if PR exists for current branch
PR_JSON=$(gh pr view --json number,state,isDraft,url 2>/dev/null || echo "")

if [ -z "$PR_JSON" ]; then
    echo "PR STATUS: No PR exists yet (will be created by /pr-prep)"
else
    PR_NUMBER=$(echo "$PR_JSON" | jq -r '.number')
    PR_STATE=$(echo "$PR_JSON" | jq -r '.state')
    IS_DRAFT=$(echo "$PR_JSON" | jq -r '.isDraft')
    PR_URL=$(echo "$PR_JSON" | jq -r '.url')

    echo "PR: #$PR_NUMBER ($PR_STATE)"
    echo "URL: $PR_URL"
fi
```

### 1b.2 Draft Detection

```bash
if [ "$IS_DRAFT" = "true" ]; then
    echo "PR STATUS: WARNING - PR is a DRAFT"
    echo ""
    echo "Draft PRs are not reviewed by maintainers."
    echo "Resolution: gh pr ready $PR_NUMBER"
    echo ""
    DRAFT_WARNING=true
else
    echo "PR STATUS: OK (ready for review)"
fi
```

### 1b.3 PR Status Summary

| Status | Meaning | Action |
|--------|---------|--------|
| No PR | Branch not yet submitted | Continue validation, then /pr-prep |
| Open (ready) | PR awaiting review | Continue validation |
| Open (draft) | **WARNING** | Convert to ready: `gh pr ready` |
| Closed | PR was closed/rejected | Check if should reopen or abandon |
| Merged | Already accepted | No validation needed |

**Note**: Draft status is a WARNING, not a blocker. Creating drafts for early feedback is valid.
The warning ensures you don't forget to mark ready before expecting review.

---

## Phase 3: CONTRIBUTING.md Compliance (BLOCKING)

**After upstream check passes. Verify PR complies with contribution guidelines.**

### 3.1 Find and Check CONTRIBUTING.md

```bash
# Find CONTRIBUTING.md
CONTRIB_FILE=""
for f in CONTRIBUTING.md .github/CONTRIBUTING.md docs/CONTRIBUTING.md; do
    if [ -f "$f" ]; then
        CONTRIB_FILE="$f"
        break
    fi
done

if [ -z "$CONTRIB_FILE" ]; then
    echo "WARNING: No CONTRIBUTING.md found - high risk of rejection"
else
    echo "Found: $CONTRIB_FILE"
fi
```

### 3.2 Extract and Verify Requirements

```bash
# Check commit message format requirement
grep -iE "commit message|conventional commit" "$CONTRIB_FILE" 2>/dev/null | head -5

# Check DCO/sign-off requirement
grep -iE "sign-off|DCO|Developer Certificate" "$CONTRIB_FILE" 2>/dev/null | head -3

# Check test requirements
grep -iE "test|coverage" "$CONTRIB_FILE" 2>/dev/null | head -5

# Verify your commits comply
git log --format="%s" $UPSTREAM..$BRANCH | head -5
```

### 3.3 CONTRIBUTING.md Compliance Checklist

| Check | Pass Criteria | How to Verify |
|-------|---------------|---------------|
| **Commit format** | Matches spec in CONTRIBUTING.md | Compare `git log` with documented format |
| **Sign-off** | Present if required | `git log --format="%s%n%b" \| grep "Signed-off-by"` |
| **Tests added** | If CONTRIBUTING.md requires | `git diff --name-only \| grep test` |
| **Lint passes** | If CONTRIBUTING.md requires | Run project's lint command |
| **PR template** | Used if provided | Check `.github/PULL_REQUEST_TEMPLATE.md` |

### 3.4 Compliance Result

```bash
# Example compliance check
CONTRIB_PASS=true

# Check for sign-off if required
if grep -qi "sign-off" "$CONTRIB_FILE" 2>/dev/null; then
    if ! git log --format="%b" $UPSTREAM..$BRANCH | grep -q "Signed-off-by"; then
        echo "CONTRIB: FAIL (sign-off required but missing)"
        CONTRIB_PASS=false
    fi
fi

if [ "$CONTRIB_PASS" = true ]; then
    echo "CONTRIB: PASS (complies with CONTRIBUTING.md)"
fi
```

**DO NOT PROCEED IF CONTRIBUTING.md REQUIREMENTS ARE NOT MET.**

---

## Phase 4: Isolation Check (BLOCKING)

**After CONTRIBUTING.md check passes. Reuses sk-pr-prep Phase 0. Ensures single concern per PR.**

### 4.1 Commit Type Analysis

```bash
# Extract commit type prefixes from branch
git log --oneline $UPSTREAM..$BRANCH | sed 's/^[^ ]* //' | grep -oE '^[a-z]+(\([^)]+\))?:' | sort -u

# Expected output for clean PR:
#   refactor(suggest):
#
# RED FLAG - mixed types like:
#   refactor(suggest):
#   fix(lint):
#   docs:
```

**Rule**: If more than one commit type prefix exists, the PR is mixing concerns.

### 4.2 File Theme Analysis

```bash
# List all files changed vs upstream
git diff --name-only $UPSTREAM..$BRANCH

# Group by directory (first 2 levels)
git diff --name-only $UPSTREAM..$BRANCH | cut -d'/' -f1-2 | sort -u

# RED FLAG: Changes spanning unrelated packages
```

### 4.3 Isolation Checklist

| Check | Pass Criteria |
|-------|---------------|
| **Single commit type** | All commits share same prefix (fix, feat, refactor, docs, test) |
| **Thematic files** | All changed files relate to PR title/scope |
| **Atomic scope** | Could explain PR in one sentence without "and also..." |

### 4.4 Isolation Result

```bash
TYPES=$(git log --oneline $UPSTREAM..$BRANCH | sed 's/^[^ ]* //' | grep -oE '^[a-z]+(\([^)]+\))?:' | sort -u | wc -l)

if [ "$TYPES" -gt 1 ]; then
    echo "ISOLATION: FAIL (mixed commit types)"
    echo "Resolution: Split into separate branches by type"
else
    echo "ISOLATION: PASS (single type)"
fi
```

---

## Phase 2: Upstream Alignment (BLOCKING - CHECK FIRST)

**FIRST CHECK: Ensure branch is up to date before any other analysis.**

No point analyzing isolation or scope on a stale branch. Rebase first, then validate.

### 2.1 Divergence Check

```bash
# Fetch latest upstream
git fetch origin $UPSTREAM

# How many commits behind?
BEHIND=$(git rev-list --count $BRANCH..origin/$UPSTREAM)
echo "Behind upstream: $BEHIND commits"

# What's been added to upstream since branch?
git log --oneline $BRANCH..origin/$UPSTREAM | head -10
```

### 2.2 Conflict Detection

```bash
# Dry-run merge to detect conflicts
git merge-tree $(git merge-base $BRANCH origin/$UPSTREAM) $BRANCH origin/$UPSTREAM | grep -c "conflict" || true

# Or check with rebase dry-run (safer)
git rebase --dry-run origin/$UPSTREAM 2>&1 | grep -i conflict || echo "No conflicts detected"
```

### 2.3 Alignment Checklist

| Check | Pass Criteria |
|-------|---------------|
| **Minimal divergence** | < 20 commits behind upstream |
| **No conflicts** | Merge/rebase would succeed |
| **No redundant changes** | Files not already modified on upstream |

### 2.4 Redundancy Check

```bash
# Files modified in both branch and upstream
BRANCH_FILES=$(git diff --name-only $UPSTREAM..$BRANCH)
UPSTREAM_FILES=$(git diff --name-only $BRANCH..origin/$UPSTREAM)

# Find overlap
comm -12 <(echo "$BRANCH_FILES" | sort) <(echo "$UPSTREAM_FILES" | sort)
```

---

## Phase 5: Scope Check (BLOCKING)

**After isolation check passes.**

**Verify changes match intended scope.**

### 5.1 Scope Inference

```bash
# Infer scope from commit messages
SCOPE=$(git log --format="%s" $UPSTREAM..$BRANCH | grep -oE '\([^)]+\)' | sort -u | head -1 | tr -d '()')

# Or from primary directory
SCOPE=$(git diff --name-only $UPSTREAM..$BRANCH | cut -d'/' -f1-2 | sort | uniq -c | sort -rn | head -1 | awk '{print $2}')
```

### 5.2 Scope Validation

```bash
# All files should be within expected scope
TOTAL_FILES=$(git diff --name-only $UPSTREAM..$BRANCH | wc -l)
SCOPE_FILES=$(git diff --name-only $UPSTREAM..$BRANCH | grep -c "$SCOPE" || echo 0)
SCOPE_PERCENT=$((SCOPE_FILES * 100 / TOTAL_FILES))

if [ "$SCOPE_PERCENT" -lt 80 ]; then
    echo "SCOPE: WARN (only ${SCOPE_PERCENT}% of files in primary scope)"
else
    echo "SCOPE: PASS (${SCOPE_PERCENT}% of files in scope: $SCOPE)"
fi
```

### 5.3 Scope Creep Detection

Look for patterns indicating scope creep:

| Pattern | Indicator |
|---------|-----------|
| `*.md` outside docs | Added documentation for unrelated areas |
| `*.test.*` for untouched code | Testing code not part of contribution |
| Config files | Changes to CI, linting, etc. |
| Unrelated packages | Files in different domains |

```bash
# List files outside primary scope
git diff --name-only $UPSTREAM..$BRANCH | grep -v "$SCOPE"
```

---

## Phase 6: Quality Gate (Non-Blocking)

**Run basic quality checks. Failures are warnings, not blockers.**

### 6.1 Test Execution

```bash
# Detect project type and run appropriate tests
if [ -f "go.mod" ]; then
    go test ./... -short 2>&1 | tail -5
elif [ -f "package.json" ]; then
    npm test 2>&1 | tail -10
elif [ -f "pyproject.toml" ] || [ -f "setup.py" ]; then
    pytest -x 2>&1 | tail -10
fi
```

### 6.2 Lint Check

```bash
# Language-specific linting
if [ -f "go.mod" ]; then
    golangci-lint run --new-from-rev=$UPSTREAM 2>&1 | head -20
elif [ -f "package.json" ]; then
    npm run lint 2>&1 | head -20
elif [ -f "pyproject.toml" ]; then
    ruff check . 2>&1 | head -20
fi
```

### 6.3 Quality Summary

```bash
# Non-blocking - report status but don't fail
TESTS_PASS=$?
if [ "$TESTS_PASS" -eq 0 ]; then
    echo "QUALITY: PASS (tests pass, lint clean)"
else
    echo "QUALITY: WARN (see test output above)"
fi
```

---

## Phase 7: Report Generation

**Generate summary report with actionable resolutions.**

### 7.1 Report Format

```markdown
## PR Validation Report

**Branch:** $BRANCH
**Upstream:** $UPSTREAM
**Commits:** N
**Files Changed:** N

### Validation Results

| Check | Status | Details |
|-------|--------|---------|
| PR Status | OK/WARN/N/A | Draft detection |
| CONTRIBUTING.md | PASS/FAIL | Complies with guidelines |
| Isolation | PASS/FAIL | Single type: refactor |
| Upstream | PASS/FAIL | 0 commits behind |
| Scope | PASS/WARN | 95% in internal/suggest/ |
| Quality | PASS/WARN | Tests pass |

### Overall Status: PASS / BLOCKED / WARN

### Resolutions (if any)

- [Issue]: [Resolution action]
```

### 7.2 Pass Output

```
PR Validation: PASS

Branch: feature/suggest-validation (5 commits)
Upstream: main (in sync)
PR: #542 (ready for review)

Checks:
  [OK] PR Status: Ready for review
  [OK] CONTRIBUTING.md: Complies with guidelines
  [OK] Isolation: Single type (refactor)
  [OK] Upstream: 0 commits behind, no conflicts
  [OK] Scope: 100% in internal/suggest/
  [OK] Quality: Tests pass

Ready for /pr-prep
```

### 7.3 Fail Output

```
PR Validation: BLOCKED

Branch: feature/mixed-changes (8 commits)
Upstream: main (15 behind)
PR: #570 (DRAFT)

Checks:
  [WARN] PR Status: DRAFT - won't be reviewed until marked ready
  [FAIL] CONTRIBUTING.md: Missing sign-off (DCO required)
  [FAIL] Isolation: Mixed types (refactor, fix, docs)
  [WARN] Upstream: 15 commits behind (rebase recommended)
  [FAIL] Scope: Only 60% in primary scope
  [OK] Quality: Tests pass

Resolutions:
0. DRAFT: Mark PR ready for review
   gh pr ready 570

1. ISOLATION: Split branch by commit type
   git checkout main && git checkout -b refactor/suggest
   git cherry-pick <refactor-commits>

2. SCOPE: Extract unrelated changes
   git reset HEAD~2 && git stash
   git checkout -b fix/unrelated
   git stash pop

Run /pr-validate again after resolution.
```

---

## Resolution Actions

### Mixed Commit Types

```bash
# Option 1: Cherry-pick to clean branch
git checkout $UPSTREAM
git checkout -b ${BRANCH}-clean
git cherry-pick <relevant-commits-only>

# Option 2: Interactive rebase to drop
git rebase -i $UPSTREAM
# Mark unrelated commits as "drop"
```

### Upstream Divergence

```bash
# Rebase on latest upstream
git fetch origin $UPSTREAM
git rebase origin/$UPSTREAM

# If conflicts, resolve then continue
git add .
git rebase --continue
```

### Scope Creep

```bash
# Extract out-of-scope changes
git checkout -b ${BRANCH}-extra
git cherry-pick <out-of-scope-commits>

# Clean original branch
git checkout $BRANCH
git rebase -i $UPSTREAM
# Drop out-of-scope commits
```

---

## Quick Examples

**Example 1: Clean Branch**

```bash
$ /pr-validate
PR Validation: PASS
  Isolation: OK (type: refactor)
  Upstream: OK (in sync)
  Scope: OK (internal/suggest/)
  Quality: OK
Ready for /pr-prep
```

**Example 2: Mixed Concerns**

```bash
$ /pr-validate
PR Validation: BLOCKED
  Isolation: FAIL
    Found types: refactor(suggest), fix(lint), docs
    Resolution: Split into 3 PRs by type
```

**Example 3: Behind Upstream**

```bash
$ /pr-validate
PR Validation: WARN
  Upstream: 23 commits behind
  Conflicts: 2 files would conflict
    Resolution: git fetch origin main && git rebase origin/main
```

---

## Integration Points

### Before /pr-prep

```bash
# Recommended workflow
/pr-implement <plan>
/pr-validate           # Catch issues before prep
/pr-prep               # Only if validation passes
```

### With /pr-implement

The sk-pr-implement skill runs isolation checks as Phase 3 (pre) and Phase 5 (post).
This skill provides the same checks as a standalone command for manual use.

### Automated Gate

Can be integrated into CI:

```yaml
# .github/workflows/pr-validate.yml
- name: PR Validation
  run: |
    # Run isolation check
    TYPES=$(git log --oneline origin/main..HEAD | sed 's/^[^ ]* //' | grep -oE '^[a-z]+:' | sort -u | wc -l)
    if [ "$TYPES" -gt 1 ]; then
      echo "::error::PR mixes multiple commit types"
      exit 1
    fi
```

---

## Maintainer Risk Considerations

When validating PRs, consider what maintainers will scrutinize:

| Risk Factor | Why It Matters | How to Address |
|-------------|----------------|----------------|
| **False positives** | Features that accidentally affect users | Add safeguards, document edge cases |
| **Large PRs** | Hard to review, hard to revert | Split into smaller focused PRs |
| **Config changes** | Break existing deployments | Add migration path, validation |
| **Testing gaps** | Unverified behavior | Complete manual test plan |
| **Signal handling** | Break containers/init systems | Document signal behavior |

### Maintainer-Focused Questions

Before marking validation as PASS, ask:

1. **Would I merge this if I were the maintainer?**
2. **What could go wrong for users?**
3. **Is the scope minimal enough to revert easily?**
4. **Are there any "trust me" assumptions?**

### Analysis vs Approval

When reporting validation results:

| DON'T | DO |
|-------|-----|
| "LGTM - ready to merge" | "Validation passes, ready for /pr-prep" |
| "This is correct" | "Tests pass, isolation clean" |
| "Approved" | "No blocking issues found" |

We validate. Maintainers approve.

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| **Analyze stale branch** | Check upstream alignment FIRST before any other validation |
| **Skip validation** | Run /pr-validate before /pr-prep |
| **Ignore scope creep** | Extract unrelated changes to new branch |
| **Submit behind upstream** | Rebase before validation |
| **Mix fix and refactor** | Separate PRs by commit type |
| **Validate after submission** | Validate before running /pr-prep |
| **Approve PRs** | Validate and report - maintainer approves |

---

## References

- **Command**: `~/.claude/commands/pr-validate.md`
- **Isolation Source**: sk-pr-prep Phase 0
- **Implementation**: sk-pr-implement (includes validation)
- **Full Validation**: sk-validation-chain (internal projects)
