# Product Brief Template

Use this template for `/product` outputs. Target completion time: **15 minutes**.

---

## Template

```markdown
# Product Brief: [Feature Name]

**Date:** YYYY-MM-DD
**Author:** [your name or "AI-assisted"]
**Status:** draft | review | approved
**Rig:** [target rig]

---

## 1. The Headline (10 words or less)

[What would the announcement say? Focus on customer value, not technical details.]

---

## 2. Customer & Problem

**Who:** [Specific persona—NOT "users". Example: "API consumers on usage-based billing plans"]

**Pain Points (ranked):**
1. [Most painful problem this solves]
2. [Second problem]
3. [Third problem (optional)]

**Current Workarounds:**
[How do they solve this today? Why is that insufficient?]

---

## 3. Solution

**What we're building:**
[2-3 sentences describing the solution from the customer's perspective, not technical internals]

**Problem→Solution Mapping:**

| Problem | Solution |
|---------|----------|
| [Pain point 1] | [Feature that addresses it] |
| [Pain point 2] | [Feature that addresses it] |
| [Pain point 3] | [Feature that addresses it] |

---

## 4. Customer Quote (Hypothetical)

> "[What would a happy user say after using this? Make it realistic and specific.]"
> — [Persona Name], [Role/Title]

---

## 5. Success Criteria

| Metric | Target | How Measured |
|--------|--------|--------------|
| [Adoption metric] | [X% or X users] | [Analytics/survey] |
| [Value metric] | [Quantified improvement] | [Data source] |
| [Quality metric] | [Error/complaint reduction] | [Logs/support] |

---

## 6. Scope

**In Scope:**
- [What we ARE building—be specific]
- [These become inputs to /formulate]
- [Keep to 3-5 items]

**Out of Scope (Non-Goals):**
- [What we are NOT building—even if someone might expect it]
- [Adjacent features explicitly deferred]
- [Things that could cause scope creep if not stated]

---

## 7. Open Questions

| Question | Impact | Owner |
|----------|--------|-------|
| [Unresolved decision] | [What it blocks] | [Who decides] |
| [Technical uncertainty] | [Risk level] | [Who investigates] |

---

**Next:** `/formulate` to decompose into engineering tasks
```

---

## Guidance Notes

### Section 1: Headline

The headline test is brutal but effective. If you can't state the value in 10 words,
you don't understand what you're building.

**Bad headlines:**
- "Implement rate limiting middleware" (technical, not value)
- "Improve API experience" (vague)
- "Add new feature for users" (says nothing)

**Good headlines:**
- "Predictable API costs with automatic overrun protection"
- "Find any code in seconds, not minutes"
- "One-click deployment to any environment"

### Section 2: Customer & Problem

**Persona specificity matters:**

| Too Vague | Just Right |
|-----------|------------|
| "Users" | "DevOps engineers managing 10+ services" |
| "Customers" | "API consumers on usage-based billing" |
| "Engineers" | "Junior developers onboarding to new codebase" |

**Pain point ranking:**
- Put the most urgent/frequent problem first
- If you can't rank, you don't understand the problem well enough
- Three problems is usually enough; more suggests scope is too broad

### Section 3: Solution

**The mapping table is critical.** Every feature must trace to a stated problem.
If you have features that don't map, ask:
- Is this actually needed?
- Is there an unstated problem?
- Is this scope creep?

### Section 4: Customer Quote

This feels awkward but works. Writing what a happy user would say forces empathy.

**Bad quote:**
> "This feature is great!"

**Good quote:**
> "I used to spend 20 minutes every morning checking if any API calls went over budget. Now I get a Slack alert only when I need to act."

### Section 5: Success Criteria

**SMART criteria:**
- **S**pecific: "80% of API users" not "most users"
- **M**easurable: "30% reduction" not "fewer"
- **A**chievable: Don't promise 100% unless you mean it
- **R**elevant: Tied to stated problems
- **T**ime-bound: "Within 3 months of launch"

### Section 6: Scope

**Non-goals are as important as goals.**

Good non-goals:
- Prevent someone from asking "but what about X?"
- Clarify boundaries for engineering
- Enable faster decisions during implementation

If you can't think of non-goals, your scope is probably too narrow or you haven't
thought about adjacent features.

### Section 7: Open Questions

**Be honest about uncertainty.** It's better to flag unknowns than to pretend
you have answers.

Questions that belong here:
- Technical feasibility uncertainties
- Business decisions not yet made
- Dependencies on other teams
- Pricing/positioning questions

---

## Time Management

| Section | Target Time |
|---------|-------------|
| Headline | 2 min |
| Customer & Problem | 4 min |
| Solution | 3 min |
| Customer Quote | 2 min |
| Success Criteria | 2 min |
| Scope | 2 min |
| Open Questions | 1 min |
| **Total** | **~15 min** |

If you're spending significantly longer:
- Scope is too broad → narrow it
- You lack context → do `/research` first
- You're over-engineering → simplify

---

## Anti-Patterns to Avoid

| Anti-Pattern | Why It's Bad | Fix |
|--------------|--------------|-----|
| "Users" as persona | Too vague to guide decisions | Name specific persona |
| Features without problems | May build wrong thing | Map every feature |
| "Improve" as metric | Unmeasurable | Quantify the improvement |
| No non-goals | Invites scope creep | List 2-3 explicit non-goals |
| 2-hour brief | Diminishing returns | Timebox to 15 min |
| Technical jargon in headline | Misses customer focus | Rewrite from user POV |
