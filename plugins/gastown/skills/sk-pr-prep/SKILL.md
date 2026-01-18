---
name: sk-pr-prep
description: >
  PR preparation workflow with git archaeology, test validation, and structured
  PR body generation. INCLUDES MANDATORY USER REVIEW GATE before submission.
  Triggers: "prepare PR", "PR prep", "submit PR", "create PR", "open source contribution".
version: 1.4.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Bash,Grep,Glob,Task"
skills:
  - beads
---

# PR Preparation Skill

Systematic PR preparation that analyzes git history, validates tests, and generates
high-quality PR bodies following maintainer expectations.

## Overview

Prepares contributions for open-source repositories by analyzing the target repo's
conventions, git history, test coverage, and generating properly-formatted PR bodies.

**When to Use**:
- Preparing a PR for an external repository
- Extracting a module for open-source contribution
- Contributing bug fixes or features
- Package extraction PRs

**When NOT to Use**:
- Internal commits (use normal git workflow)
- PRs to your own repositories
- Trivial documentation-only changes

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

## Phase -1: Prior Work Check (BLOCKING)

**CRITICAL**: Final verification that no competing PR was submitted while you worked.

### -1.1 Check for Competing PRs

```bash
REPO="<owner/repo>"

# Search for open PRs on this topic (use your PR title keywords)
gh pr list -R $REPO --state open --search "<your PR topic>" --limit 10

# Check for PRs touching the same files you modified
FILES=$(git diff --name-only main..HEAD | head -5 | tr '\n' ' ')
gh pr list -R $REPO --state open --json files,number,title,url | \
  jq -r '.[] | "\(.number): \(.title) - \(.url)"'
```

### -1.2 Check for Recently Merged Work

```bash
# Check if someone merged a fix while you were working
gh pr list -R $REPO --state merged --search "<topic>" --limit 5

# Pull latest and check if your changes conflict or duplicate
git fetch upstream
git log HEAD..upstream/main --oneline | head -10
```

### -1.3 Prior Work Decision

| Finding | Action |
|---------|--------|
| Competing PR just opened | Coordinate - comment on theirs or differentiate scope |
| Similar PR merged | Pull latest, verify if still needed, may need rebase or abandon |
| No competing work | Proceed to Phase 0 |

**If competing work found**: Either coordinate, differentiate, or abandon. Don't create duplicate PRs.

---

## Phase 0: Isolation Check (BLOCKING)

**CRITICAL**: Run this FIRST. Do not proceed if PR mixes unrelated changes.

### 0.1 Commit Type Analysis

```bash
# Extract commit type prefixes from branch
git log --oneline main..HEAD | sed 's/^[^ ]* //' | grep -oE '^[a-z]+(\([^)]+\))?:' | sort -u

# Expected output for clean PR:
#   refactor(suggest):
#
# RED FLAG - mixed types like:
#   refactor(suggest):
#   fix(lint):
#   docs:
```

**Rule**: If more than one commit type prefix exists, the PR is mixing concerns.

### 0.2 File Theme Analysis

```bash
# List all files changed vs main
git diff --name-only main..HEAD

# Group by directory
git diff --name-only main..HEAD | cut -d'/' -f1-2 | sort -u

# RED FLAG: Changes spanning unrelated packages
# e.g., PR titled "refactor(suggest)" touching:
#   internal/suggest/suggest.go  ✓ expected
#   internal/cmd/down.go         ✗ unrelated
#   internal/cmd/root.go         ✗ unrelated
```

### 0.3 Main Divergence Check

```bash
# What's been added to main since branch created?
git log --oneline HEAD..origin/main

# Are any of your changes already on main?
git log --oneline main..HEAD -- <file>
# vs
git log --oneline HEAD..origin/main -- <file>

# If same file modified in both, likely redundant or conflict
```

### 0.4 Isolation Checklist

| Check | Pass Criteria |
|-------|---------------|
| **Single commit type** | All commits share same prefix (fix, feat, refactor, docs, test) |
| **Thematic files** | All changed files relate to PR title/scope |
| **No main overlap** | Changes not already merged to main via other PRs |
| **Atomic scope** | Could explain PR in one sentence without "and also..." |

### 0.5 Resolution Actions

If isolation check fails:

| Issue | Resolution |
|-------|------------|
| Mixed commit types | Split into separate PRs by type |
| Unrelated files | `git reset` and selectively re-commit |
| Already on main | Rebase to drop redundant commits |
| Scope creep | Extract unrelated work to new branch |

```bash
# Example: Remove lint fixes from refactor PR
git rebase -i main
# Mark lint commits as "drop"
# Or cherry-pick only the relevant commits to a clean branch

# Example: Check if changes redundant with main
git diff origin/main..HEAD -- internal/cmd/down.go
# Empty diff = already on main, drop the commit
```

**DO NOT PROCEED TO PHASE 1 IF ISOLATION CHECK FAILS.**

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

**Anti-pattern**: Generating PR body and immediately running `gh pr create`

---

## Phase 1: CONTRIBUTING.md Compliance (BLOCKING)

**CRITICAL**: Verify your PR complies with CONTRIBUTING.md before proceeding.

### 1.1 Re-Read CONTRIBUTING.md

Even if you ran /pr-research, re-check CONTRIBUTING.md before submission:

```bash
# Find and read CONTRIBUTING.md
cat CONTRIBUTING.md 2>/dev/null || \
cat .github/CONTRIBUTING.md 2>/dev/null || \
cat docs/CONTRIBUTING.md 2>/dev/null || \
echo "WARNING: No CONTRIBUTING.md found"

# Check for PR template
cat .github/PULL_REQUEST_TEMPLATE.md 2>/dev/null || echo "No PR template"
```

### 1.2 CONTRIBUTING.md Compliance Checklist

**YOU MUST VERIFY EACH ITEM BEFORE PROCEEDING:**

| Requirement | Your Compliance | How to Verify |
|-------------|-----------------|---------------|
| **Commit format** | [ ] Matches spec | `git log --oneline main..HEAD` |
| **PR title format** | [ ] Matches spec | Check your planned title |
| **Tests required** | [ ] Tests added | `git diff --name-only main..HEAD \| grep test` |
| **Linting** | [ ] Passes | Run project's lint command |
| **CLA/DCO** | [ ] Signed (if required) | Check CONTRIBUTING.md legal section |
| **Issue link** | [ ] Issue exists | Verify issue number |
| **Documentation** | [ ] Updated (if required) | Check if API/behavior changed |

### 1.3 Common CONTRIBUTING.md Requirements

Extract these from the project's CONTRIBUTING.md:

```bash
# DCO sign-off required?
grep -i "sign-off\|DCO\|Developer Certificate" CONTRIBUTING.md .github/CONTRIBUTING.md 2>/dev/null

# Commit message format?
grep -i "commit message\|conventional commit" CONTRIBUTING.md .github/CONTRIBUTING.md 2>/dev/null

# Required reviewers?
cat CODEOWNERS 2>/dev/null | head -10

# Check for commit conventions
git log --oneline -20 | head -10  # Recent commit style

# Check for required checks
cat .github/workflows/*.yml 2>/dev/null | grep -E "name:|runs-on:" | head -20
```

### 1.4 Maintainer Expectations Checklist

| Expectation | How to Check |
|-------------|--------------|
| Commit style | `git log --oneline -10` |
| PR template | `.github/PULL_REQUEST_TEMPLATE.md` |
| Required tests | CI workflow files |
| Linting | `.golangci.yml`, `eslint.config.js`, etc. |
| Code coverage | CI checks for coverage thresholds |

**DO NOT PROCEED TO PHASE 2 IF ANY CONTRIBUTING.md REQUIREMENTS ARE UNMET.**

---

## Phase 2: Git Archaeology

**CRITICAL**: Understand the history of files being modified.

### 2.1 File History

```bash
# When was this file last modified?
git log --oneline -5 -- <file>

# Who are the primary contributors?
git shortlog -sn -- <directory>

# What's the commit pattern for this area?
git log --oneline --since="3 months ago" -- <directory> | head -20
```

### 2.2 Related PRs (if GitHub)

```bash
# Find related merged PRs
gh pr list --state merged --limit 10 --search "<keywords>"

# Analyze successful PR patterns
gh pr view <number> --json title,body,files,additions,deletions
```

### 2.3 Historical Context Research

**CRITICAL**: If changing or removing code/comments, trace why they exist.

```bash
# Find when a specific comment or pattern was introduced
git log -p --all -S "do NOT set BEADS_DIR" -- <file> | head -80

# Find the commit that introduced this code
git log --oneline --all -S "<specific code pattern>" -- <file>

# View the full commit that introduced it
git show <commit-hash> --stat

# Check if there were subsequent fixes to this area
git log --oneline <introducing-commit>..HEAD -- <file>

# Find PRs by author (to see related work)
gh pr list --author <author> --state all --limit 20

# Check if a "correct" pattern exists elsewhere
grep -r "the correct pattern" --include="*.go" .
```

### 2.4 Building the Timeline

When your PR supersedes or contradicts existing code/comments, build a timeline:

| Date | Commit/PR | Author | What Happened |
|------|-----------|--------|---------------|
| ... | Original | ... | Introduced the pattern |
| ... | Fix attempt | ... | Tried to fix but wrong approach |
| ... | Correct fix | ... | Established correct pattern elsewhere |
| Now | This PR | You | Applies correct pattern here |

**Key questions:**
- When was this code/comment introduced?
- What problem was it trying to solve?
- Were there subsequent fixes that superseded it?
- Is there a "correct" pattern established elsewhere that should be applied here?

**Real example (gastown PR #512):**
```
| Jan 5  | PR #149   | boshu2 | Added ExtractPrefix, GetRigPathForPrefix helpers |
| Jan 6  | 52ef89c5  | jack   | Wrong fix: "do NOT set BEADS_DIR" (didn't use helpers) |
| Jan 11 | 598a39e7  | joe    | Correct fix in beads.go: ALWAYS set BEADS_DIR |
| Jan 14 | PR #512   | boshu2 | Applies correct pattern using original helpers |
```

### 2.5 Checklist

- [ ] Reviewed recent commits to understand style
- [ ] Identified primary maintainers for these files
- [ ] Found related merged PRs for reference
- [ ] Understood the evolution of this code area
- [ ] **If changing existing code: traced when/why it was introduced**
- [ ] **If superseding a fix: documented the timeline**
- [ ] **If contradicting a comment: proved the comment was wrong**

---

## Phase 3: Pre-Flight Checks

### 3.1 Build Verification

```bash
# Go projects
go build ./...
go vet ./...
golangci-lint run 2>/dev/null || echo "linter not installed"

# Node projects
npm run build 2>/dev/null || yarn build 2>/dev/null

# Python projects
python -m py_compile <files>
```

### 3.2 Test Execution

```bash
# Go - run tests for affected packages
go test ./... -v -count=1

# Go - check coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1

# Node
npm test

# Python
pytest -v
```

### 3.3 Pre-Flight Checklist

- [ ] Code compiles without errors
- [ ] All tests pass
- [ ] No new linting warnings
- [ ] Coverage maintained or improved
- [ ] No secrets or credentials in code

---

## Phase 4: Change Analysis

### 4.1 Diff Summary

```bash
# What files changed?
git diff --stat HEAD~1

# Lines added/removed
git diff --shortstat HEAD~1

# Specific changes
git diff HEAD~1 -- <key-file>
```

### 4.2 Impact Assessment

| Category | Questions |
|----------|-----------|
| **Scope** | How many files? Which packages? |
| **Risk** | Breaking changes? Migration needed? |
| **Dependencies** | New deps added? Versions changed? |
| **Tests** | New tests? Test coverage delta? |

### 4.3 Change Categories

Classify changes for PR summary:

| Category | Examples |
|----------|----------|
| **New Files** | New packages, new modules |
| **Modified** | Updated functions, fixed bugs |
| **Deleted** | Removed dead code, deprecated APIs |
| **Config** | go.mod, package.json, CI |
| **Docs** | README, inline comments |

---

## Phase 5: PR Body Generation

### 5.1 Standard Format (Focused PRs)

For bug fixes, small features, and focused changes. Based on successful upstream PR patterns:

```markdown
## Summary

Brief description of WHAT changed and WHY. 1-3 sentences.
Start with action verb (Add, Fix, Update, Refactor).

## Changes

Technical details of what was modified. Include code snippets when helpful.

For complex implementations, use `## Implementation` instead:
- Explain the approach
- Show before/after code patterns
- Note any tradeoffs

## Test plan

- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] Manual: <specific scenario tested>

Fixes #NNN
```

**Key conventions:**
- Test plan items are **checked** `[x]` (you ran them before PR)
- `Fixes #NNN` goes at the end
- Keep it concise - maintainers review many PRs

### 5.2 Detailed Format (Large PRs)

For PRs touching many files, adding new packages, or making structural changes:

```markdown
## Summary

2-4 sentences describing WHAT and WHY. Note any breaking changes upfront.

## Related Issue

Fixes #123 or Closes #123

## Changes

### New Files
- `path/to/file.go` - Brief description

### Modified Files
- `path/to/existing.go` - What changed and why

### Removed Files
- `path/to/old.go` - Why removed

## Test plan

- [x] Build passes (`go build ./...`)
- [x] All tests pass (`go test ./...`)
- [x] Manual testing: <specific scenario>
- [x] Coverage maintained

## Checklist

- [x] Code follows project style
- [x] Documentation updated (if applicable)
- [x] No breaking changes (or documented in summary)
- [x] Commit messages follow conventions
```

### 5.3 When to Add Historical Context

Only add `## Historical Context` when:
- Changing or removing existing code/comments
- Contradicting an existing comment that appears authoritative
- Superseding a previous fix attempt

When needed, build a timeline using git archaeology:

```markdown
## Historical Context

| Date | Commit/PR | Author | What Happened |
|------|-----------|--------|---------------|
| Jan 5 | PR #149 | author1 | Original pattern introduced |
| Jan 6 | 52ef89c5 | author2 | Wrong fix attempt |
| Jan 11 | 598a39e7 | author3 | Correct pattern established elsewhere |
| Now | This PR | you | Applies correct pattern here |

The old comment claimed "..." but git archaeology shows this was superseded by [commit].
```

### 5.4 Effective Summaries

**Good Summary Characteristics**:
- Starts with action verb (Add, Fix, Update, Refactor)
- Explains the "why" not just "what"
- Mentions affected components
- Notes any breaking changes upfront

**Example - Good**:
> Add standalone formula parsing library extracted from gastown.
> The formula package provides TOML-based workflow definitions with
> validation, cycle detection, and topological sorting. Extracted
> to enable reuse in other Go projects without gastown dependency.

**Example - Bad**:
> Updated formula code and added tests

---

## Phase 6: User Review Gate

**MANDATORY STOP POINT**

Before ANY submission, present the complete PR for review:

### 6.1 Generate Review Artifact

```bash
# Write PR body to reviewable file
cat > .pr-review.md << 'EOF'
# PR Review

**Title**: type(scope): description
**Target**: owner/repo (base: main)
**Branch**: your-branch-name

---

<full PR body here>
EOF

# Display for review
cat .pr-review.md
```

### 6.2 Request Approval

Use AskUserQuestion or direct prompt:

```
PR body generated. Please review above.

Options:
1. Submit as-is
2. Edit and resubmit
3. Cancel
```

**DO NOT PROCEED TO PHASE 7 WITHOUT EXPLICIT "submit" or "yes" FROM USER.**

---

## Phase 7: Submission (After Approval Only)

### 7.1 Final Checklist

```markdown
## PR Readiness Checklist

### Code Quality
- [ ] Tests pass locally
- [ ] Linting passes
- [ ] No TODO/FIXME introduced
- [ ] No debug code left

### Git Hygiene
- [ ] Commits are atomic
- [ ] Commit messages follow conventions
- [ ] Branch is rebased on latest main
- [ ] No merge commits (or squash merge planned)

### Documentation
- [ ] PR body complete
- [ ] Code comments where non-obvious
- [ ] README updated if public API changed

### Risk Assessment
- [ ] No secrets in code
- [ ] No breaking changes (or documented)
- [ ] Performance impact considered
```

### 7.2 PR Creation Command (Only After Phase 6 Approval)

```bash
# Create PR with reviewed body
gh pr create --title "type(scope): brief description" \
  --body "$(cat .pr-review.md)" \
  --base main
```

**Remember**: This command should ONLY run after user explicitly approves in Phase 6.

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| **Submit without approval** | **ALWAYS stop at Phase 6 for user review** |
| **Skip isolation check** | **Run Phase 0 FIRST - block if mixed concerns** |
| **Change code without history** | **Trace when/why code was introduced (git -S)** |
| **Contradict comments without proof** | **Build timeline showing comment was wrong** |
| **Bundle lint fixes into feature PRs** | Lint fixes get their own `fix(lint):` PR |
| **Add "while I'm here" changes** | Separate branch for unrelated improvements |
| Giant PRs | Split into logical chunks |
| Vague PR body | Detailed summary with context |
| Skip pre-flight | Always run tests locally |
| Ignore conventions | Match existing style |
| Mix concerns | One PR = one logical change |
| Force-push after review | Add fixup commits |
| Auto-submit after generating body | Write to file, show user, wait for approval |
| Assume branch is clean | Check for main divergence before PR |
| Assume old comments are correct | Verify with git archaeology and tests |

---

## Additional Resources

### Reference Files

For detailed patterns and examples, consult:

- **`references/case-study-historical-context.md`** - PR #512 walkthrough showing git archaeology for contradicting existing comments
- **`references/lessons-learned.md`** - Real PR outcomes, acceptance patterns, what got rejected
- **`references/package-extraction.md`** - Template for extracting standalone libraries

### External References

- **PR Template**: `.github/PULL_REQUEST_TEMPLATE.md`
- **Conventional Commits**: https://conventionalcommits.org
