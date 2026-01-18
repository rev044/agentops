# Lesson: Maintainer Perspective

**Date**: 2026-01-16
**Source**: Gas Town session lifecycle PR stack analysis
**Key Insight**: We provide analysis, not approvals

## What Happened

We analyzed 6 open PRs on upstream gastown and posted review comments. Initially, our comments included "LGTM" (Looks Good To Me) approval language. This was wrong.

## The Problem

- "LGTM" is approval language - it's the maintainer's job to approve
- We are not the maintainer
- Our role is to provide helpful analysis to support the review process

## Corrected Approach

### DON'T:
```markdown
✅ LGTM - this can merge independently.
✅ LGTM pending conflict check with #590.
The fix is correct. ✅ LGTM with the understanding that...
```

### DO:
```markdown
**Analysis Notes**

This PR addresses [problem] with [approach].

**Open questions**:
1. Is X sufficient for Y?
2. What happens if Z?

**Merge order**: Should merge after #NNN because [dependency].
```

## Maintainer Risk Calculus

When helping with PR review, consider what the maintainer cares about:

| Concern | Why It Matters |
|---------|----------------|
| **False positives** | Features that accidentally affect users are hard to undo |
| **Large PRs** | Harder to review, harder to revert |
| **Config changes** | Need migration paths for existing users |
| **Testing gaps** | Unchecked "manual test" items are red flags |
| **Signal handling** | Could break containers or init systems |
| **Merge conflicts** | PRs touching same files need coordination |

## PR Comment Structure

When reviewing PRs for a maintainer:

```markdown
## Analysis: [PR Stack Name]

[1-2 sentences on what the PR does and why it matters]

### Technical Notes

- [Key implementation detail]
- [Notable pattern or concern]

### Questions for Review

1. [Specific technical question]
2. [Edge case to consider]

### Merge Order Context

[Dependency chain and why]
```

## When Reviewing PR Stacks

For related PRs that should merge together:

1. **Map dependencies** - Which PRs depend on others?
2. **Detect conflicts** - Do any modify the same files?
3. **Identify blockers** - Which PR, if broken, blocks others?
4. **Suggest order** - Recommend merge sequence with rationale
5. **Note integration needs** - What testing should happen after merge?

## Key Takeaway

> **We provide analysis to help the maintainer's review. We don't make approval decisions.**

This is true even when our analysis suggests a PR is ready - the maintainer decides.
