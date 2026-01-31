# REBRAND COORDINATION: DevOps for Vibe-Coding Pivot

**Date:** 2026-01-31
**Status:** Wave 0 - Pre-Requisite Before ag-xrd Execution
**Purpose:** Address 4 CRITICAL issues from pre-mortem before /crank

---

## 1. Terminology Map

### Primary Messaging Change

| OLD | NEW |
|-----|-----|
| "Operational principles for reliable AI agents" | "DevOps for Vibe-Coding" |
| "12-Factor AgentOps" | "12-Factor AgentOps" (KEEP - don't rename) |
| "AI agents" | "Coding agents" |
| "Production-grade agents" | "Shift-left validation" |
| "How do we take AI agents from 80% reliable to production-grade?" | "How do we validate AI-generated code before it ships?" |

### Secondary Phrases to Update

| OLD PHRASE | NEW PHRASE |
|------------|------------|
| "AI-assisted development" | "Validation-first coding" |
| "Production-ready plugins" | "Shift-left validation plugins" |
| "Operate them reliably" | "Validate before you ship" |

### Phrases to KEEP (Not Changing)

- "12-Factor AgentOps" (repo name unchanged)
- "Vibe Coding" (reference to book)
- "Knowledge Flywheel"
- "Brownian Ratchet"
- "RPI workflow"

---

## 2. Scope Definition

### What IS a "Coding Agent"

- **Claude Code** running in terminal/IDE
- AI assistants that write, modify, or review code
- Agents using tools like Read, Edit, Write, Bash for development
- AI pair programming sessions
- Code generation with validation workflows

### What IS NOT (Out of Scope)

- Chatbots for customer service
- RAG-based Q&A systems
- Multi-modal agents (image, voice)
- Autonomous production agents (Olympus handles this)
- General-purpose AI assistants

### Where Out-of-Scope Users Should Go

| Use Case | Recommended Resource |
|----------|---------------------|
| Building autonomous agents | [12-Factor Agents](https://github.com/humanlayer/12-factor-agents) by Dex Horthy |
| Production agent orchestration | Olympus (multi-session, temporal-based) |
| Enterprise AI platform | Athena (RAG, memory, observability) |

---

## 3. File Audit

### 12factor Repository (~/gt/12factor)

Files containing old tagline "operational principles":

| File | Action | Priority |
|------|--------|----------|
| `README.md` | UPDATE | P0 - Hero statement |
| `docs/README.md` | UPDATE | P1 |
| `docs/00-SUMMARY.md` | UPDATE | P1 |
| `docs/ecosystem.md` | UPDATE | P1 |
| `docs/principles/README.md` | UPDATE | P2 |
| `docs/principles/constraint-based-engineering.md` | UPDATE | P2 |
| `docs/explanation/from-theory-to-production.md` | UPDATE | P2 |
| `docs/explanation/ecosystem-position.md` | UPDATE | P2 |
| `factors/README.md` | UPDATE | P2 |
| `.agents/research/2025-12-27-*.md` | SKIP | Historical research |

### agentops Repository (~/gt/agentops)

Files requiring update:

| File | Action | Priority |
|------|--------|----------|
| `.claude-plugin/plugin.json` | UPDATE | P0 - Marketplace |
| `README.md` (if exists) | UPDATE | P1 |
| Skill SKILL.md files | REFRAME | P2 (Issue ag-xrd.9) |

Note: `.agents/` files referencing old terminology are expected (historical context).

### personal_site Repository (~/gt/personal_site)

Primary files (crew/neo is the active worktree):

| File | Action | Priority |
|------|--------|----------|
| `crew/neo/src/content/writing/12-factor-agentops.mdx` | UPDATE | P1 |
| `crew/neo/src/content/writing/gutenberg-moment-for-code.mdx` | UPDATE | P2 |

Note: `mayor/rig/` and `refinery/rig/` are duplicates - only update active crew.

---

## 4. Archive Plan

### Archive Strategy: DIRECTORY (Not Delete)

All archived content moves to `docs/_archived/` with README explaining scope change.

### Files to Archive

| File | Reason | Archive Path |
|------|--------|--------------|
| `docs/domain-guides/platform-engineering-agent.md` | Infrastructure automation, not coding | `docs/_archived/domain-guides/` |
| `docs/case-studies/production/` (entire dir) | Production ops focus, not coding validation | `docs/_archived/case-studies/` |

### Files to KEEP (Despite General Focus)

| File | Reason to Keep |
|------|----------------|
| `docs/case-studies/enterprise-validation.md` | References Vibe Coding book, coding-relevant |
| All factor definitions in `factors/` | Factors apply to coding agents |

### Pre-Archive Checklist

- [ ] Create backup tag: `git tag archive/v2.0-pre-rebrand`
- [ ] Run: `grep -r "_archived" .agents/learnings/` to check for backlinks
- [ ] Verify no beads issues reference archived files

### Archive README Template

```markdown
# Archived Content

**Archived:** 2026-01-31
**Reason:** Scope narrowed from "reliable AI agents" to "coding agents"

These files document operational patterns for general production agents.
For general agent patterns, see [12-Factor Agents](https://github.com/humanlayer/12-factor-agents).

This content remains for historical reference but is no longer actively maintained.
```

---

## 5. Rollout Sequence

### Pre-Launch (48h Before)

**T-48h: Preparation**
- [ ] Create this coordination document (DONE)
- [ ] Create backup tag: `git tag archive/v2.0-pre-rebrand` on all 3 repos
- [ ] Notify Gene Kim/Steve Yegge about scope change (if possible)
- [ ] Verify plugin.json update propagation (test in staging if available)

**T-24h: Final Checks**
- [ ] Run final grep for old taglines across all repos
- [ ] Verify all issues in ag-xrd have clear acceptance criteria
- [ ] Create rollback commit (ready but not pushed)

### Launch Day: Atomic 2-Hour Window

| Time | Repo | Action | Verification |
|------|------|--------|--------------|
| 10:00 | 12factor | Push README + docs updates | `grep "operational principles" README.md` returns empty |
| 10:15 | 12factor | Push archive moves | `ls docs/_archived/` shows archived files |
| 10:30 | agentops | Push plugin.json + README updates | Check marketplace (note: cache delay) |
| 10:45 | agentops | Push skill SKILL.md reframes | Run `/help` to verify descriptions |
| 11:00 | personal_site | Push article updates | Check local build renders correctly |
| 11:15 | personal_site | Deploy to production | Verify live site |
| 11:30 | - | Write blog post explaining pivot | Draft ready in docs/blog/ |
| 12:00 | - | WAIT for cache propagation | Do NOT announce until +48h |

### Post-Launch (+48h)

**T+48h: Marketing Window Opens**
- [ ] Verify plugin marketplace shows new description
- [ ] Verify CDN caches refreshed
- [ ] Announce on social media
- [ ] Submit to relevant newsletters/aggregators

**T+2 weeks: Monitoring**
- [ ] Check GitHub issues for confused users
- [ ] Track SEO rankings for old vs new keywords
- [ ] Run `/retro` to capture learnings
- [ ] Close ag-xrd epic if all children complete

---

## 6. Rollback Plan

### Trigger Conditions for Rollback

- Plugin marketplace shows incorrect metadata after 72h
- More than 5 confused issues filed in first week
- Critical negative feedback from Vibe Coding authors
- Major SEO ranking loss detected

### Rollback Procedure

```bash
# Step 1: Restore from backup tag
cd ~/gt/12factor
git checkout archive/v2.0-pre-rebrand
git checkout -b rollback/2026-01-31
git push -u origin rollback/2026-01-31

# Step 2: Create revert PR
gh pr create --title "Rollback: Revert DevOps rebrand" \
  --body "Rolling back scope pivot due to [REASON]. See REBRAND-COORDINATION.md for details."

# Step 3: Update other repos similarly
cd ~/gt/agentops && git checkout archive/v2.0-pre-rebrand
cd ~/gt/personal_site && git checkout archive/v2.0-pre-rebrand

# Step 4: Re-push plugin.json with original content
# Wait 48h for cache propagation

# Step 5: Issue public statement
# Create .agents/post-mortems/rollback-[date].md
```

### Partial Rollback Options

| Issue | Partial Fix |
|-------|-------------|
| Only plugin marketplace wrong | Revert only plugin.json, wait for cache |
| Only personal_site confusing | Revert only that repo |
| Archive broke backlinks | Move archived files back, update references |

---

## 7. CI Checks to Add

After rebrand, add these checks to prevent regression:

### GitHub Actions Workflow

```yaml
# .github/workflows/terminology-check.yml
name: Terminology Check
on: [push, pull_request]

jobs:
  check-old-taglines:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Check for old taglines
        run: |
          if grep -rI "operational principles for reliable AI agents" --include="*.md" .; then
            echo "ERROR: Found old tagline in markdown files"
            exit 1
          fi
          if grep -rI "production-grade agents" --include="*.md" docs/; then
            echo "WARNING: Found 'production-grade agents' - verify context"
          fi
```

### Pre-Commit Hook

```yaml
# .pre-commit-config.yaml (add to existing)
- repo: local
  hooks:
    - id: check-old-taglines
      name: Check for old taglines
      entry: bash -c 'if grep -rI "operational principles for reliable" .; then exit 1; fi'
      language: system
      types: [markdown]
```

---

## 8. Success Criteria for Wave 0

Wave 0 is complete when:

- [x] REBRAND-COORDINATION.md created (this file)
- [x] Backup tags created on all 3 repos (2026-01-31)
    - 12factor: `archive/v2.0-pre-rebrand`
    - agentops: `archive/v2.0-pre-rebrand`
    - personal_site: `archive/v2.0-pre-rebrand`
- [ ] Explicit archive file list reviewed and approved
- [ ] Rollout sequence reviewed and approved
- [ ] CI checks drafted (will add after rebrand)

### Gate Decision

Once all checkboxes above are complete:
1. Run `/crank ag-xrd` for autonomous execution
2. Follow rollout sequence timing
3. Monitor for 2 weeks
4. Run `/post-mortem ag-xrd` to extract learnings

---

## 9. Plugin Cache Strategy (CRITICAL)

### The Problem

Claude Code marketplace caches plugin.json for 24-48+ hours. Users installing during this window see stale description.

### The Solution

1. **Update plugin.json FIRST** (at T+10:30)
2. **Wait 48h before any marketing** (until T+60h)
3. **Monitor propagation** with test installs

### Verification Commands

```bash
# Check current marketplace state (when tools available)
# For now, manually check Claude Code plugin marketplace

# After update, verify local build
cd ~/gt/agentops
cat .claude-plugin/plugin.json | jq '.description'
```

### Fallback

If cache doesn't clear after 72h:
- File issue with Anthropic support
- Use README/SKILL.md to clarify (these aren't cached)
- Note discrepancy in blog post announcement

---

## 10. External Coordination

### People to Notify

| Who | Why | When | How |
|-----|-----|------|-----|
| Steve Yegge | Vibe Coding book author, credits us | T-1 week | Twitter DM or email |
| Gene Kim | Vibe Coding co-author | T-1 week | Via Steve or direct |
| Dex Horthy | 12-Factor Agents author | T-1 week | Twitter DM |

### Suggested Notification Template

```
Hi [Name],

I'm pivoting 12-Factor AgentOps to focus specifically on coding agents
with a "DevOps for Vibe-Coding" positioning. The repo name stays the same,
but the messaging is shifting from "reliable AI agents" to "shift-left
validation for coding agents."

Wanted to give you a heads up since [you're credited/you inspired the work].

Rolling out [DATE]. Let me know if you have concerns.

- Boden
```

---

*Coordination document complete. Ready for human review before /crank.*
