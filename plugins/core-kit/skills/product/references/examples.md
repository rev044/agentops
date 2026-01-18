# Product Brief Examples

Good and bad examples to guide product brief creation.

---

## Good Example 1: API Rate Limiting

```markdown
# Product Brief: API Rate Limiting

**Date:** 2026-01-11
**Author:** Platform Team
**Status:** approved
**Rig:** ai-platform

---

## 1. The Headline (10 words or less)

Predictable API costs with automatic overrun protection.

---

## 2. Customer & Problem

**Who:** API consumers on usage-based billing plans (primarily startups and SMBs)

**Pain Points (ranked):**
1. **Unpredictable monthly bills** — Usage spikes during incidents or load tests cause surprise charges
2. **No visibility into consumption** — Can't see real-time usage until bill arrives
3. **Abrupt service cutoffs** — Hard limits cause production outages without warning

**Current Workarounds:**
Teams set calendar reminders to manually check usage dashboards, or build custom monitoring that queries billing APIs. This is error-prone and adds operational burden.

---

## 3. Solution

**What we're building:**
Self-service rate limiting that lets customers set usage caps with configurable alerts and graceful degradation. Users define their comfort zone; the system enforces it automatically.

**Problem→Solution Mapping:**

| Problem | Solution |
|---------|----------|
| Unpredictable bills | Configurable monthly/daily limits with hard caps |
| No visibility | Real-time usage dashboard + Slack/email alerts at thresholds |
| Abrupt cutoffs | Graceful degradation mode (429 with retry-after, not hard block) |

---

## 4. Customer Quote (Hypothetical)

> "I used to dread opening our cloud bill. Now I set a $500 limit, get alerts at 80%, and sleep soundly knowing we can't accidentally burn through our budget."
> — Sarah Chen, CTO at a 15-person startup

---

## 5. Success Criteria

| Metric | Target | How Measured |
|--------|--------|--------------|
| Adoption | 60% of paid users enable limits within 30 days | Analytics |
| Cost predictability | 40% reduction in billing support tickets | Support system |
| Customer satisfaction | NPS increase of 10 points | Quarterly survey |

---

## 6. Scope

**In Scope:**
- Per-API-key usage limits (requests and tokens)
- Real-time usage dashboard in UI
- Configurable alert thresholds (email, Slack webhook)
- Graceful degradation (429 + Retry-After header)
- Admin override for emergency lift

**Out of Scope (Non-Goals):**
- DDoS protection (separate security initiative)
- Per-endpoint rate limiting (future phase)
- Automatic budget scaling based on traffic patterns (v2)
- Changes to pricing model or tier structure

---

## 7. Open Questions

| Question | Impact | Owner |
|----------|--------|-------|
| Default limit for new accounts? | UX decision | Product |
| Alert fatigue threshold? | May need tuning post-launch | Engineering |
```

---

## Good Example 2: Code Search

```markdown
# Product Brief: Semantic Code Search

**Date:** 2026-01-10
**Author:** AI Platform Team
**Status:** review
**Rig:** ai-platform

---

## 1. The Headline (10 words or less)

Find any code in your repos in seconds, not minutes.

---

## 2. Customer & Problem

**Who:** Developers onboarding to large codebases (>100K LOC) or working across multiple repositories

**Pain Points (ranked):**
1. **Discovery friction** — "I know this exists somewhere" but grep/find take forever
2. **Context switching** — Must clone repos locally to search; breaks flow
3. **Semantic gaps** — Text search misses conceptually related code (e.g., "authentication" vs "login")

**Current Workarounds:**
Developers grep locally, ask teammates on Slack, or give up and reimplement. Senior engineers become human search indexes, interrupting their work to answer "where is X?" questions.

---

## 3. Solution

**What we're building:**
AI-powered code search that understands intent, not just keywords. Search across all indexed repos from the chat interface without cloning anything locally.

**Problem→Solution Mapping:**

| Problem | Solution |
|---------|----------|
| Discovery friction | Sub-second search across all repos via chat |
| Context switching | No clone needed; results include file links |
| Semantic gaps | Embedding-based search understands concepts |

---

## 4. Customer Quote (Hypothetical)

> "I asked 'how do we handle auth tokens?' and it showed me the exact middleware in a repo I didn't even know existed. What used to take an hour of Slack questions now takes 10 seconds."
> — Marcus, Senior Engineer joining a new team

---

## 5. Success Criteria

| Metric | Target | How Measured |
|--------|--------|--------------|
| Search latency | p95 < 2 seconds | APM metrics |
| Relevance | 80% of top-3 results rated useful | User feedback |
| Adoption | 50 searches/day/active user | Analytics |

---

## 6. Scope

**In Scope:**
- Semantic search over indexed GitLab repositories
- Chat interface integration ("search for X")
- Results with file paths, line numbers, snippets
- Re-indexing on push/merge

**Out of Scope (Non-Goals):**
- Code modification/refactoring suggestions (separate feature)
- Private repo access without explicit indexing permission
- Non-code files (docs, configs) — future phase
- Real-time streaming of changes (batch indexing only)

---

## 7. Open Questions

| Question | Impact | Owner |
|----------|--------|-------|
| Index all branches or just default? | Storage/performance | Engineering |
| How to handle large monorepos? | May need chunking strategy | Engineering |
```

---

## Bad Example: Generic Brief (Anti-Pattern)

**What's wrong:** Vague persona, unmeasurable success, features without problems, no non-goals.

```markdown
# Product Brief: Better API Experience

**Date:** 2026-01-11
**Author:** Team
**Status:** draft

---

## 1. The Headline

Improve the API experience.

❌ Too vague — what aspect? For whom?

---

## 2. Customer & Problem

**Who:** Users

❌ "Users" is not a persona

**Pain Points:**
- API could be better
- Some things are confusing
- Need more features

❌ Vague — what specifically is painful?

**Current Workarounds:** N/A

❌ If there are no workarounds, is this really a problem?

---

## 3. Solution

**What we're building:**
We will improve the API to make it better for users. We'll add features they want.

❌ Says nothing concrete

**Problem→Solution Mapping:**

| Problem | Solution |
|---------|----------|
| API could be better | Improvements |
| Confusing things | Make them clearer |

❌ Features don't map to specific problems

---

## 4. Customer Quote

> "This is great!"

❌ Generic — could apply to anything

---

## 5. Success Criteria

| Metric | Target | How Measured |
|--------|--------|--------------|
| User satisfaction | Higher | Survey |
| Usage | More | Analytics |

❌ "Higher" and "More" aren't measurable targets

---

## 6. Scope

**In Scope:**
- API improvements
- New features
- Better experience

❌ Vague — doesn't guide engineering

**Out of Scope:**
(none listed)

❌ Missing non-goals invites scope creep

---

## 7. Open Questions

(none)

❌ Pretending certainty when there isn't any
```

---

## When to Skip `/product`

Don't create a product brief for:

### 1. Bug Fixes

**Skip because:** The "why" is already clear (it's broken).

```
❌ /product "Fix null pointer in auth handler"
✓ Just fix it directly
```

### 2. Technical Debt

**Skip because:** No direct user impact to articulate.

```
❌ /product "Refactor database connection pooling"
✓ /research then /formulate directly
```

### 3. Single-Day Tasks

**Skip because:** Overhead exceeds value.

```
❌ /product "Add logging to payment endpoint"
✓ Just do it
```

### 4. PM Already Wrote PRD

**Skip because:** Don't duplicate work.

```
❌ /product (when PRD exists at docs/product/feature-x.md)
✓ Reference existing PRD in /formulate
```

### 5. Pure Infrastructure

**Skip because:** No end-user to describe.

```
❌ /product "Upgrade Kubernetes to 1.29"
✓ /research for compatibility, then /formulate
```

---

## Decision Flowchart

```
Is this user-facing work?
├─ No → Skip /product, go to /research or /formulate
└─ Yes
    └─ Will it take 3+ days?
        ├─ No → Skip /product
        └─ Yes
            └─ Does a PRD already exist?
                ├─ Yes → Skip /product, reference PRD
                └─ No → Use /product
```

---

## Common Mistakes and Fixes

| Mistake | Example | Fix |
|---------|---------|-----|
| Persona too vague | "Users" | "API consumers on metered plans" |
| Technical headline | "Implement caching layer" | "Instant responses for common queries" |
| Unmeasurable success | "Users are happier" | "NPS +10, response time <500ms" |
| No non-goals | (missing section) | List 2-3 things you're NOT building |
| Over-engineering | 2-hour brief | Timebox to 15 minutes |
| Skipping when needed | Multi-week feature with no brief | Take 15 minutes to articulate "why" |
