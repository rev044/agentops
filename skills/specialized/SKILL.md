---
name: specialized
description: >
  Use when: "accessibility", "WCAG", "a11y", "customer support", "FAQ", "tickets",
  "UI", "UX", "design system", "wireframe", "Obsidian", "knowledge graph", "links",
  "task decomposition", "workflow", "risk assessment", "edge cases", "failure modes".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Specialized Skill

Domain-specific patterns for accessibility, support, design, knowledge management, planning, and risk.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Accessibility** | WCAG, ARIA, screen readers | Inclusive design |
| **Customer Support** | Tickets, FAQs, troubleshooting | Support content |
| **UI/UX Design** | Wireframes, design systems | User experience |
| **Knowledge Graphs** | Obsidian links, connections | Knowledge management |
| **Task Decomposition** | Goal breakdown, workflows | Complex planning |
| **Risk Assessment** | Edge cases, failure modes | Risk analysis |

---

## Accessibility

### WCAG Compliance Levels

| Level | Requirement | Examples |
|-------|-------------|----------|
| **A** | Minimum | Alt text, keyboard access |
| **AA** | Standard | Color contrast 4.5:1, focus visible |
| **AAA** | Enhanced | Contrast 7:1, sign language |

### Quick Checks

```markdown
## Accessibility Checklist

### Perceivable
- [ ] Images have alt text
- [ ] Videos have captions
- [ ] Color contrast meets 4.5:1 (AA)
- [ ] Text resizable to 200%

### Operable
- [ ] All functionality via keyboard
- [ ] Focus visible and logical
- [ ] No keyboard traps
- [ ] Skip links available

### Understandable
- [ ] Language declared
- [ ] Error messages helpful
- [ ] Consistent navigation
- [ ] Labels on form fields

### Robust
- [ ] Valid HTML
- [ ] ARIA used correctly
- [ ] Works with assistive tech
```

### ARIA Patterns

```html
<!-- Button -->
<button aria-label="Close dialog" aria-pressed="false">×</button>

<!-- Navigation -->
<nav aria-label="Main navigation">
  <ul role="menubar">
    <li role="menuitem"><a href="/">Home</a></li>
  </ul>
</nav>

<!-- Live region -->
<div aria-live="polite" aria-atomic="true">
  Status message will be announced
</div>

<!-- Modal -->
<div role="dialog" aria-modal="true" aria-labelledby="modal-title">
  <h2 id="modal-title">Dialog Title</h2>
</div>
```

### Testing Tools

| Tool | Purpose |
|------|---------|
| **axe** | Automated testing |
| **WAVE** | Visual feedback |
| **Lighthouse** | Performance + a11y |
| **NVDA/VoiceOver** | Screen reader testing |

---

## Customer Support

### Ticket Response Template

```markdown
## Response: [Ticket #]

Hi [Name],

Thank you for reaching out about [issue summary].

### Understanding
[Confirm understanding of the issue]

### Solution
[Step-by-step resolution]

1. [Step 1]
2. [Step 2]
3. [Step 3]

### If This Doesn't Work
[Alternative approaches or escalation]

### Prevention
[How to avoid this in future, if applicable]

Let me know if you have any questions!

Best,
[Support Agent]
```

### FAQ Structure

```markdown
## FAQ: [Topic]

### Q: [Common question]
**A:** [Clear, concise answer]

**Steps:**
1. [Step 1]
2. [Step 2]

**Related:** [Link to detailed guide]

---

### Q: [Another question]
**A:** [Answer]
```

### Troubleshooting Guide Template

```markdown
## Troubleshooting: [Issue]

### Symptoms
- [Symptom 1]
- [Symptom 2]

### Quick Fixes
1. **[Fix 1]**: [Instructions]
2. **[Fix 2]**: [Instructions]

### Diagnosis

| Check | Expected | If Different |
|-------|----------|--------------|
| [Check 1] | [Expected] | [Action] |
| [Check 2] | [Expected] | [Action] |

### Advanced Troubleshooting
[For complex cases]

### Escalation
If unresolved, contact [team] with:
- [Required info 1]
- [Required info 2]
```

---

## UI/UX Design

### Design System Components

```markdown
## Component: [Name]

### Variants
| Variant | Use Case |
|---------|----------|
| Primary | Main actions |
| Secondary | Supporting actions |
| Danger | Destructive actions |

### States
- Default
- Hover
- Active
- Disabled
- Loading

### Accessibility
- Minimum touch target: 44x44px
- Focus indicator visible
- ARIA attributes required

### Usage
[When to use this component]

### Don't
[Common misuses to avoid]
```

### Wireframe Notation

```
┌─────────────────────────────────┐
│ [Logo]        [Nav] [Nav] [Nav] │ ← Header
├─────────────────────────────────┤
│                                 │
│  ┌─────────┐  ┌─────────┐      │
│  │ Card    │  │ Card    │      │ ← Content
│  │         │  │         │      │
│  └─────────┘  └─────────┘      │
│                                 │
├─────────────────────────────────┤
│ [Footer]                        │ ← Footer
└─────────────────────────────────┘
```

### User Research Template

```markdown
## User Research: [Feature]

### Research Questions
1. [Question 1]
2. [Question 2]

### Methodology
- [ ] User interviews (n=X)
- [ ] Surveys (n=X)
- [ ] Usability testing
- [ ] Analytics review

### Key Findings
1. **[Finding 1]**: [Evidence]
2. **[Finding 2]**: [Evidence]

### Recommendations
| Priority | Recommendation | Effort |
|----------|----------------|--------|
| P0 | [Must do] | Low |
| P1 | [Should do] | Medium |

### Next Steps
[How to act on findings]
```

---

## Knowledge Graphs (Obsidian)

### Link Analysis

```bash
# Find orphan notes (no incoming links)
grep -rL "\[\[" vault/*.md

# Find broken links
grep -ohE "\[\[[^\]]+\]\]" vault/*.md | sort | uniq -c | sort -rn

# Most linked notes
grep -oh "\[\[[^\]]*\]\]" vault/*.md | sort | uniq -c | sort -rn | head -20
```

### Connection Patterns

```markdown
## Link Types

### Hierarchical
- Parent → Child: `[[Parent]] > [[Child]]`
- Index → Item: `[[MOC]] contains [[Note]]`

### Associative
- Related: `[[Note A]] relates to [[Note B]]`
- See also: `See [[Related Topic]]`

### Temporal
- Follows: `[[Step 1]] → [[Step 2]]`
- Version: `[[v1]] superseded by [[v2]]`
```

### Vault Health Check

```markdown
## Vault Health Report

### Statistics
- Total notes: X
- Total links: X
- Orphan notes: X
- Broken links: X

### Most Connected (hubs)
1. [[Note 1]] - 50 links
2. [[Note 2]] - 35 links

### Orphans (need integration)
- [[Orphan 1]]
- [[Orphan 2]]

### Broken Links (need fixing)
- [[Missing Note]] referenced in 5 places

### Recommendations
1. [Connect orphans]
2. [Fix broken links]
3. [Create MOC for cluster]
```

---

## Task Decomposition

### Goal Breakdown

```markdown
## Goal: [High-level goal]

### Success Criteria
- [ ] [Measurable outcome 1]
- [ ] [Measurable outcome 2]

### Decomposition

#### Phase 1: [Name]
| Task | Dependencies | Effort |
|------|--------------|--------|
| [Task 1] | None | S |
| [Task 2] | Task 1 | M |

#### Phase 2: [Name]
| Task | Dependencies | Effort |
|------|--------------|--------|
| [Task 3] | Phase 1 | L |

### Critical Path
Task 1 → Task 2 → Task 3 (bottleneck)

### Risks
- [Risk 1]: [Mitigation]
```

### Workflow Architecture

```markdown
## Workflow: [Name]

### Triggers
- [What starts this workflow]

### Steps
```
[Input] → [Step 1] → [Step 2] → [Output]
              ↓
         [Branch if condition]
              ↓
         [Alternative step]
```

### Error Handling
| Error | Recovery |
|-------|----------|
| [Error 1] | [Action] |

### Monitoring
- [Metric to track]
```

---

## Risk Assessment

### Risk Matrix

| Likelihood / Impact | Low | Medium | High |
|---------------------|-----|--------|------|
| **High** | Medium | High | Critical |
| **Medium** | Low | Medium | High |
| **Low** | Low | Low | Medium |

### Risk Register Template

```markdown
## Risk Register: [Project]

| ID | Risk | Likelihood | Impact | Score | Mitigation | Owner |
|----|------|------------|--------|-------|------------|-------|
| R1 | [Risk] | High | Medium | High | [Action] | [Name] |
| R2 | [Risk] | Low | High | Medium | [Action] | [Name] |
```

### Edge Case Analysis

```markdown
## Edge Cases: [Feature]

### Input Edge Cases
| Case | Input | Expected | Risk |
|------|-------|----------|------|
| Empty | "" | Error message | Low |
| Max length | 10000 chars | Truncate | Medium |
| Special chars | "<script>" | Sanitize | High |

### State Edge Cases
| Case | State | Expected | Risk |
|------|-------|----------|------|
| Concurrent | Two users edit | Merge conflict | Medium |
| Timeout | Network fail | Retry | Low |

### Failure Modes
| Mode | Cause | Detection | Recovery |
|------|-------|-----------|----------|
| [Mode 1] | [Cause] | [How to detect] | [Recovery] |
```

### Failure Mode Analysis

```markdown
## FMEA: [Component]

| Failure Mode | Effect | Severity | Cause | Occurrence | Detection | RPN | Action |
|--------------|--------|----------|-------|------------|-----------|-----|--------|
| [Mode] | [Effect] | 8 | [Cause] | 3 | 5 | 120 | [Action] |

**RPN** = Severity × Occurrence × Detection (lower is better)

### Priority Actions
1. [Highest RPN item] - [Action]
2. [Next highest] - [Action]
```
