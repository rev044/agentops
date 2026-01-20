---
name: pr-retro
description: >
  Learn from PR outcomes. Analyzes accept/reject patterns and updates pr-prep lessons.
  Triggers: "pr retro", "learn from PR", "PR outcome", "why was PR rejected",
  "analyze PR feedback", "update pr lessons".
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,WebFetch"
skills:
  - beads
---

# PR Retro Skill

Learn from PR outcomes to improve future contribution success rates.

## Overview

Analyzes PR outcomes (merged, closed, feedback received) and extracts actionable
patterns. Updates the pr-prep skill's Lessons Learned section to improve
future PRs.

**When to Use**:
- After a PR is merged (capture what worked)
- After a PR is rejected/closed (learn from feedback)
- After receiving substantive review feedback

**When NOT to Use**:
- PR still open with no feedback
- Trivial PRs (typo fixes, doc tweaks)
- Auto-merged PRs (CI-only changes)

---

## Workflow

```
1. PR Identification   -> Parse input, validate PR exists
2. Outcome Analysis    -> Fetch PR state, reviews, comments
3. Feedback Extraction -> Extract maintainer expectations
4. Pattern Classification -> Categorize success/failure patterns
5. Lessons Update      -> Update pr-prep/SKILL.md
6. Retro Artifact      -> Write to .agents/retros/
```

---

## Phase 1: PR Identification

### 1.1 Parse Input

Support multiple input formats:

```bash
# PR number (requires context of current repo)
PR_NUM=236

# Full URL
PR_URL="https://github.com/steveyegge/gastown/pull/353"

# owner/repo#number format
PR_REF="steveyegge/gastown#353"
```

### 1.2 Normalize to Owner/Repo/Number

```bash
# Extract components
OWNER="steveyegge"
REPO="gastown"
PR_NUM="353"

# Verify PR exists
gh pr view $PR_NUM --repo $OWNER/$REPO --json state,title,body 2>/dev/null
```

---

## Phase 2: Outcome Analysis

### 2.1 Fetch PR State

```bash
# Get comprehensive PR data
gh pr view $PR_NUM --repo $OWNER/$REPO --json \
  state,title,body,author,createdAt,closedAt,mergedAt,\
  additions,deletions,changedFiles,\
  reviews,comments,labels,mergedBy,\
  baseRefName,headRefName

# State values: OPEN, CLOSED, MERGED
```

### 2.2 Classify Outcome

| State | Sub-Type | Classification |
|-------|----------|----------------|
| MERGED | - | SUCCESS |
| CLOSED | With review comments | REJECTED (feedback) |
| CLOSED | No comments | ABANDONED/STALE |
| OPEN | Changes requested | NEEDS_WORK |
| OPEN | Approved | PENDING_MERGE |
| OPEN | No reviews | AWAITING_REVIEW |

### 2.3 Extract Timeline

```bash
# When was PR created, reviewed, merged/closed?
gh pr view $PR_NUM --repo $OWNER/$REPO --json \
  createdAt,reviews,comments,closedAt,mergedAt \
  --jq '{
    created: .createdAt,
    first_review: .reviews[0].submittedAt,
    closed: .closedAt,
    merged: .mergedAt
  }'
```

---

## Phase 3: Feedback Extraction

### 3.1 Collect All Feedback

```bash
# Get review comments
gh api repos/$OWNER/$REPO/pulls/$PR_NUM/reviews --jq '.[].body'

# Get inline comments
gh api repos/$OWNER/$REPO/pulls/$PR_NUM/comments --jq '.[].body'

# Get issue comments (general discussion)
gh pr view $PR_NUM --repo $OWNER/$REPO --json comments --jq '.comments[].body'
```

### 3.2 Identify Maintainer Feedback

Focus on comments from:
- Repository maintainers (MEMBER, OWNER roles)
- Review state changes (APPROVED, CHANGES_REQUESTED)
- Merge commit messages

```bash
# Get reviews with author association
gh api repos/$OWNER/$REPO/pulls/$PR_NUM/reviews \
  --jq '.[] | select(.author_association == "MEMBER" or .author_association == "OWNER") | {author: .user.login, state: .state, body: .body}'
```

### 3.3 Extract Key Themes

Common feedback categories:

| Category | Keywords/Phrases |
|----------|------------------|
| **Scope** | "too large", "split this", "unrelated changes" |
| **Style** | "naming", "formatting", "conventions" |
| **Architecture** | "coupling", "abstraction", "design" |
| **Testing** | "add tests", "coverage", "edge cases" |
| **Documentation** | "add docs", "explain", "comments" |
| **Performance** | "slow", "optimize", "benchmark" |
| **Security** | "vulnerability", "sanitize", "validate" |

---

## Phase 4: Pattern Classification

### 4.1 Success Patterns (for MERGED PRs)

Extract what made the PR successful:

| Pattern | Evidence |
|---------|----------|
| **Small scope** | < 5 files changed, single concern |
| **Good tests** | Test files included, coverage up |
| **Clean commits** | Conventional commit messages |
| **Clear description** | PR body follows template |
| **Quick iteration** | Responded to feedback promptly |

### 4.2 Failure Patterns (for CLOSED PRs)

Extract what caused rejection:

| Pattern | Evidence |
|---------|----------|
| **Mixed concerns** | Multiple unrelated changes |
| **Missing tests** | No test changes, coverage down |
| **Style violations** | Lint failures, convention breaks |
| **Scope creep** | Added features beyond original intent |
| **Stale** | No response to feedback for 30+ days |
| **Superseded** | Another PR/approach preferred |
| **Architecture conflict** | Rejected abstraction/coupling |

### 4.3 Record Classification

```markdown
## Pattern Analysis

**PR:** owner/repo#123
**Outcome:** MERGED | REJECTED | ABANDONED
**Classification:** SUCCESS | FAILURE

### Success Factors (if merged)
- Small, focused changes (3 files)
- Clear problem statement
- Tests included

### Failure Factors (if rejected)
- Mixed refactor with fix
- Maintainer suggested different approach
- Superseded by #456
```

---

## Phase 5: Lessons Update

### 5.1 Target File

Update the "Lessons Learned" section in:
`pr-prep/SKILL.md`

### 5.2 Update Rules

| PR Outcome | Action |
|------------|--------|
| MERGED | Add to "What Got Accepted" table |
| REJECTED with feedback | Add to "What Got Rejected" table |
| ABANDONED (no feedback) | Skip (not informative) |

### 5.3 Entry Format

For accepted PRs:
```markdown
| #123 | type | **Single focus** - clear value prop, tests pass |
```

For rejected PRs:
```markdown
| #123 | type | **Scope issue** - mixed concerns, maintainer requested split |
```

### 5.4 Update Script

```bash
# Read current lessons
SKILL_FILE="$HOME/.claude/skills/pr-prep/SKILL.md"

# Find the correct section based on outcome
if [ "$OUTCOME" = "MERGED" ]; then
  SECTION="What Got Accepted"
else
  SECTION="What Got Rejected"
fi

# Insert new entry (use Edit tool for precision)
```

### 5.5 Pattern Updates

If new patterns emerge (not already documented), add to:
- "Patterns that work" (for success)
- "Patterns to avoid" (for failure)

---

## Phase 6: Retro Artifact

### 6.1 Rig Detection

Determine which rig this PR relates to:

```bash
# From repo name mapping
case "$REPO" in
  "ai-platform") RIG="ai-platform" ;;
  "gastown") RIG="gastown" ;;
  "fractal") RIG="fractal" ;;
  *) RIG="_external" ;;  # External repos
esac
```

### 6.2 Output Location

```bash
mkdir -p ~/gt/.agents/$RIG/retros/
RETRO_FILE="~/gt/.agents/$RIG/retros/$(date +%Y-%m-%d)-pr-${REPO}#${PR_NUM}.md"
```

### 6.3 Retro Template

```markdown
---
tags: [retro, pr-outcome, {repo}]
pr: {owner}/{repo}#{number}
outcome: {MERGED|REJECTED|ABANDONED}
date: {YYYY-MM-DD}
---

# PR Retro: {repo}#{number}

## Summary

**Title:** {PR title}
**Outcome:** {MERGED|REJECTED|ABANDONED}
**Time to resolution:** {N days}

## What Happened

{Brief narrative of the PR lifecycle}

## Feedback Analysis

### Maintainer Comments
{Key quotes from maintainers}

### Review Requests
{What changes were requested}

## Patterns Identified

### Success Factors
- {factor 1}
- {factor 2}

### Failure Factors
- {factor 1}
- {factor 2}

## Lessons for Future PRs

1. {Actionable lesson 1}
2. {Actionable lesson 2}

## Updates Made

- [ ] pr-prep/SKILL.md updated
- [ ] Pattern documented: {pattern-name}

## Related

- PR: {URL}
- Research: {link to pr-research artifact if exists}
- Prep: {link to pr-prep artifact if exists}
```

---

## Maintainer Perspective Analysis

When running retros, analyze from the maintainer's viewpoint:

### What Maintainers Care About

| Factor | Questions to Ask |
|--------|------------------|
| **User impact** | Did the change risk affecting users? False positives? |
| **Revert safety** | Could this be reverted easily if broken? |
| **Testing confidence** | Was the test plan complete? Manual tests executed? |
| **Scope containment** | Was the PR minimal, or did it touch too much? |
| **Migration path** | For config/API changes, was migration considered? |

### Retro Questions (Maintainer Lens)

When analyzing why a PR was accepted or rejected:

1. **If accepted**: What made the maintainer confident to merge?
2. **If rejected**: What specific risk did the maintainer identify?
3. **If slow review**: What information was missing upfront?

### Common Maintainer Concerns

From analysis of real PR feedback:

| Concern | Example Feedback | Lesson |
|---------|------------------|--------|
| **False positives** | "What if user runs Claude in cron?" | Add safeguards for edge cases |
| **PR too large** | "Can you split this?" | Keep PRs < 400 lines |
| **Testing gaps** | "Manual test items unchecked" | Complete test plan before PR |
| **Config migration** | "What about existing users?" | Document migration path |

### Updating Skills with Lessons

When a retro reveals new patterns, update:

1. `pr-prep/SKILL.md` - Lessons Learned tables
2. `pr-validate/SKILL.md` - Validation checks
3. `pr-plan/lessons/` - Create dated lesson file

**Lesson file format**: `YYYY-MM-DD-{topic}.md`

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Retro open PRs with no feedback | Wait for outcome |
| Blame maintainers for rejection | Extract actionable lessons |
| Add duplicate lessons | Check existing entries first |
| Skip rejection analysis | Most valuable learning comes from failures |
| Only track your own PRs | Learn from all team PRs |
| Use approval language in analysis | Provide analysis, maintainer decides |

---

## Integration with pr-prep

The lessons captured here feed directly into `/pr-prep`:

1. **Phase 0 Isolation Check** - Informed by rejection patterns
2. **Phase 5 PR Body** - Follows successful PR patterns
3. **How to Improve Acceptance Rate** - Updated with new insights

### Feedback Loop

```
/pr-research → implement → /pr-prep → submit → outcome → /pr-retro
                                                              ↓
                                                    pr-prep updated
                                                              ↓
                                                    Next /pr-prep improved
```

---

## References

- **PR Prep Skill**: `pr-prep/SKILL.md`
- **PR Research Skill**: `pr-research/SKILL.md`
- **Retro Skill**: `retro/SKILL.md`
- **GitHub CLI**: https://cli.github.com/manual/gh_pr_view
