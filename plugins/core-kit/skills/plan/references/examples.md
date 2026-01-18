# Plan Skill Examples

Detailed examples of plan creation scenarios.

---

## Example 1: Simple Multi-Feature Plan

**User Request**: "/plan Add OAuth2 support for GitLab integration"

**Agent Actions**:

```bash
# Phase 0: Context Discovery
mcp__smart-connections-work__lookup --query="OAuth2 GitLab authentication" --limit=10
cat docs/code-map/README.md | grep -i auth

# Phase 0.5: Prior Plan Check
grep -l "OAuth\|GitLab" .agents/plans/*.md
bd list --type=epic | grep -i oauth

# Phase 1: Research
# Found existing auth patterns in services/gateway/
# GitLab MCP server already exists at services/mcp-servers/gitlab/

# Phase 2: Decomposition
# Feature 1: Add OAuth2 callback endpoint (P1, no deps)
# Feature 2: Implement token exchange flow (P1, depends on F1)
# Feature 3: Add session token storage (P2, depends on F2)
# Feature 4: Update GitLab MCP to use OAuth tokens (P2, depends on F3)
# Feature 5: Add OAuth2 config to Helm values (P1, no deps)

# Phase 3: Create Issues
bd create "Epic: OAuth2 GitLab Integration" --type epic --priority P1
# Created: ai-platform-101

bd create "Add OAuth2 callback endpoint" --type feature --priority P1
# Created: ai-platform-102
bd comment ai-platform-102 "Files affected: services/gateway/routes.py, services/gateway/oauth.py"

bd create "Implement token exchange flow" --type feature --priority P1
# Created: ai-platform-103
bd dep add ai-platform-103 ai-platform-102

bd create "Add session token storage" --type feature --priority P2
# Created: ai-platform-104
bd dep add ai-platform-104 ai-platform-103

bd create "Update GitLab MCP to use OAuth tokens" --type feature --priority P2
# Created: ai-platform-105
bd dep add ai-platform-105 ai-platform-104

bd create "Add OAuth2 config to Helm values" --type feature --priority P1
# Created: ai-platform-106
# No deps - runs in parallel with ai-platform-102 (Wave 1)
bd comment ai-platform-106 "Files affected: charts/ai-platform/values.yaml"

# Track ALL children in epic via comment (NOT dependency!)
bd comments add ai-platform-101 "Children: ai-platform-102, ai-platform-103, ai-platform-104, ai-platform-105, ai-platform-106"

# Start the epic so children become ready
bd update ai-platform-101 --status in_progress
```

**Result**: 5 features under epic, waves computed:
- Wave 1: ai-platform-102, ai-platform-106 (parallel, different files)
- Wave 2: ai-platform-103
- Wave 3: ai-platform-104
- Wave 4: ai-platform-105

---

## Example 2: Extending Existing Plan

**Scenario**: User asks to plan work that overlaps with existing epic.

**Agent Actions**:

```bash
# Phase 0.5: Find existing work
bd list --type=epic | grep -i "auth"
# Found: ai-platform-101 "Epic: OAuth2 GitLab Integration" [in_progress]

bd show ai-platform-101
# Shows: 3 of 5 children completed
```

**Agent Response**:

"I found an existing epic `ai-platform-101` for OAuth2 GitLab Integration that's in progress
with 3 of 5 children completed. Do you want me to:

1. **Extend this epic** - Add new features as additional children
2. **Create a new epic** - Reference this one and build on its work
3. **Wait for completion** - This epic is close to done, we could finish it first"

**If user chooses "Extend":**

```bash
# Add new children to existing epic
bd create "Add OAuth2 refresh token rotation" --type feature --priority P2
# Created: ai-platform-107
bd dep add ai-platform-107 ai-platform-105  # Depends on OAuth MCP update

# Update children list
bd comments add ai-platform-101 "Added child: ai-platform-107 (refresh rotation)"
```

---

## Example 3: Plan From Research

**Scenario**: `/research` was already run, now planning from findings.

**Agent Actions**:

```bash
# Find recent research
ls -lt .agents/research/*.md | head -3
cat .agents/research/2026-01-03-oauth-investigation.md

# Research document contains:
# ## Recommendations
# 1. Implement token refresh (P1) - security critical
# 2. Add rate limiting (P2) - prevents abuse
# 3. Add audit logging (P2) - compliance requirement

# Create plan referencing research
bd create "Epic: OAuth Hardening" --type epic --priority P1
# Created: ai-platform-110
```

**Plan document includes:**
```markdown
## Research Reference
`.agents/research/2026-01-03-oauth-investigation.md`

Recommendations from research:
- Token refresh: IMPLEMENTED as ai-platform-111
- Rate limiting: IMPLEMENTED as ai-platform-112
- Audit logging: IMPLEMENTED as ai-platform-113
```

---

## Example 4: Complex Dependency Graph

**Scenario**: Feature with multiple parallel tracks that merge.

```bash
# Create epic
bd create "Epic: Multi-tenant Support" --type epic --priority P0

# Track 1: Database changes
bd create "Add tenant column to all tables" --type feature --priority P1
# -> ai-platform-201
bd create "Implement row-level security" --type feature --priority P1
# -> ai-platform-202
bd dep add ai-platform-202 ai-platform-201

# Track 2: API changes (parallel with Track 1)
bd create "Add tenant middleware" --type feature --priority P1
# -> ai-platform-203
bd create "Update all endpoints for tenant context" --type feature --priority P1
# -> ai-platform-204
bd dep add ai-platform-204 ai-platform-203

# Merge point: depends on both tracks
bd create "Integration tests for multi-tenant" --type feature --priority P1
# -> ai-platform-205
bd dep add ai-platform-205 ai-platform-202  # Waits for DB track
bd dep add ai-platform-205 ai-platform-204  # Waits for API track
```

**Result Waves:**
- Wave 1: ai-platform-201, ai-platform-203 (parallel start)
- Wave 2: ai-platform-202, ai-platform-204 (parallel, different tracks)
- Wave 3: ai-platform-205 (merge point, needs both tracks done)

---

## Example 5: Quick Plan (3 Features or Less)

**Scenario**: Small goal, doesn't need full epic overhead.

```bash
# For 1-2 features, skip epic
bd create "Add rate limiting to auth endpoint" --type feature --priority P1
# -> ai-platform-301
bd comment ai-platform-301 "Files affected: services/gateway/auth.py, services/gateway/middleware.py"

bd create "Add rate limit config to Helm values" --type feature --priority P2
# -> ai-platform-302
# No dependency - can run in parallel

# No epic needed - just output summary
```

**Summary output:**
```markdown
# Plan Complete: Rate Limiting for Auth

**Issues:** 2 features (no epic needed for small scope)

| ID | Feature | Priority | Ready |
|----|---------|----------|-------|
| ai-platform-301 | Add rate limiting to auth endpoint | P1 | Ready |
| ai-platform-302 | Add rate limit config to Helm values | P2 | Ready |

Both can run in parallel (different files). Use:
```bash
bd ready
/implement ai-platform-301
```
```

---

## Anti-Pattern Examples

### WRONG: Children Depending on Epic

```bash
# DON'T DO THIS
bd create "Epic: OAuth" --type epic
# -> ai-platform-400

bd create "Add callback" --type feature
# -> ai-platform-401

bd dep add ai-platform-401 ai-platform-400  # WRONG!
# Result: ai-platform-401 will NEVER become ready because
# the epic can't be closed until children are done (deadlock)
```

### WRONG: Skipping File Annotations

```bash
# DON'T DO THIS
bd create "Feature 1" --type feature
bd create "Feature 2" --type feature
# Both show as "ready" but /implement-wave can't detect
# if they conflict on files
```

### WRONG: Creating Too Many Features

```bash
# DON'T DO THIS
bd create "Epic: Huge Refactor" --type epic
# Then create 15 features at once

# Better: Break into 2-3 epics of 5 features each
# Or: Create first 5, complete, then plan next 5
```
