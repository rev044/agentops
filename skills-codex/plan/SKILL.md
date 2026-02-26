---
name: plan
description: 'Epic decomposition into trackable issues. Triggers: "create a plan", "plan implementation", "break down into tasks", "decompose into features", "create beads issues from research", "what issues should we create", "plan out the work".'
---


# Plan Skill

> **Quick Ref:** Decompose goal into trackable issues with waves. Output: `.agents/plans/*.md` + bd issues.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** bd (issue creation). If bd is unavailable, write the plan to `.agents/plans/` as markdown with issue descriptions, and use TaskList for tracking instead. The plan document is always created regardless of bd availability.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--auto` | off | Skip human approval gate. Used by `$rpi --auto` for fully autonomous lifecycle. |

## Execution Steps

Given `$plan <goal> [--auto]`:

### Step 1: Setup
```bash
mkdir -p .agents/plans
```

### Step 2: Check for Prior Research

Look for existing research on this topic:
```bash
ls -la .agents/research/ 2>/dev/null | head -10
```

Use Grep to search `.agents/` for related content. If research exists, read it with the Read tool to understand the context before planning.

**Search knowledge flywheel for prior planning patterns:**
```bash
if command -v ao &>/dev/null; then
    ao know search "<topic> plan decomposition patterns" 2>/dev/null | head -10
fi
```
If ao returns relevant learnings or patterns, incorporate them into the plan. Skip silently if ao is unavailable or returns no results.

### Step 3: Explore the Codebase (if needed)

**USE THE TASK TOOL** to dispatch an Explore agent. The explore prompt MUST request symbol-level detail:

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  description: "Understand codebase for: <goal>"
  prompt: |
    Explore the codebase to understand what's needed for: <goal>

    1. Find relevant files and modules
    2. Understand current architecture
    3. Identify what needs to change

    For EACH file that needs modification, return:
    - Exact function/method signatures that need changes
    - Struct/type definitions that need new fields
    - Key functions to reuse (with file:line references)
    - Existing test file locations and naming conventions (e.g., TestFoo_Bar)
    - Import paths and package relationships

    Return: file inventory, per-file symbol details, reuse points with line numbers, test patterns
```

#### Pre-Planning Baseline Audit (Mandatory)

**Before decomposing into issues**, run a quantitative baseline audit to ground the plan in verified numbers. This is mandatory for ALL plans — not just cleanup/refactor. Any plan that makes quantitative claims (counts, sizes, coverage) must verify them mechanically.

Run grep/wc/ls commands to count the current state of what you're changing:

- **Files to change:** count with `ls`/`find`/`wc -l`
- **Sections to add/remove:** count with `grep -l`/`grep -L`
- **Code to modify:** count LOC, packages, import references
- **Coverage gaps:** count missing items with `grep -L` or `find`

**Record the verification commands alongside their results.** These become pre-mortem evidence and acceptance criteria.

| Bad | Good |
|-----|------|
| "14 missing refs/" | "14 missing refs/ (verified: `ls -d skills/*/references/ \| wc -l` = 20 of 34)" |
| "clean up dead code" | "Delete 3,003 LOC across 3 packages (verified: `find src/old -name '*.go' \| xargs wc -l`)" |
| "update stale docs" | "Rewrite 4 specs (verified: `ls docs/specs/*.md \| wc -l` = 4)" |
| "add missing sections" | "Add Examples to 27 skills (verified: `grep -L '## Examples' skills/*/SKILL.md \| wc -l` = 27)" |

Ground truth with numbers prevents scope creep and makes completion verifiable. In ol-571, the audit found 5,752 LOC to remove — without it, the plan would have been vague. In ag-dnu, wrong counts (11 vs 14, 0 vs 7) caused a pre-mortem FAIL that a simple grep audit would have prevented.

### Step 3.5: Generate Implementation Detail (Mandatory)

**After exploring the codebase**, generate symbol-level implementation detail for EVERY file in the plan. This is what separates actionable specs from vague descriptions. A worker reading the plan should know exactly what to write without rediscovering function names, parameters, or code locations.

#### File Inventory Table

Start with a `## Files to Modify` table listing EVERY file the plan touches:

```markdown
## Files to Modify

| File | Change |
|------|--------|
| `src/auth/middleware.go` | Add rate limit check to `AuthMiddleware` |
| `src/config/config.go` | Add `RateLimit` section to `Config` struct |
| `src/auth/middleware_test.go` | **NEW** — rate limit middleware tests |
```

Mark new files with `**NEW**`. This table gives the implementer the full blast radius in 30 seconds.

#### Per-Section Implementation Specs

For each logical change group, provide symbol-level detail:

1. **Exact function signatures** — name the function, its parameters, and what changes:
   - "Add `worktreePath string` parameter to `classifyRunStatus`"
   - "Create new `RPIConfig` struct with `WorktreeMode string` field"

2. **Key functions to reuse** — with `file:line` references from the explore step:
   - "Reuse `readRunHeartbeat()` at `rpi_phased.go:1963`"
   - "Call existing `parsePhasedState()` at `rpi_phased.go:1924`"

3. **Inline code blocks** — for non-obvious constructs (struct definitions, CLI flags, config snippets):
   ```go
   type RPIConfig struct {
       WorktreeMode string `yaml:"worktree_mode" json:"worktree_mode"`
   }
   ```

4. **New struct fields with tags** — exact field names and JSON/YAML tags

5. **CLI flag definitions** — exact flag names, types, defaults, and help text

#### Named Test Functions

For each test file, list specific test functions with one-line descriptions:

```markdown
**`src/auth/middleware_test.go`** — add:
- `TestRateLimitMiddleware_UnderLimit`: Request within limit returns 200
- `TestRateLimitMiddleware_OverLimit`: Request exceeding limit returns 429
- `TestRateLimitMiddleware_ResetAfterWindow`: Counter resets after time window
```

#### Verification Procedures

Add a `## Verification` section with runnable bash sequences that reproduce the scenario and confirm the fix:

```markdown
## Verification

1. **Unit tests**: `go test ./src/auth/ -run "TestRateLimit" -v`
2. **Build check**: `go build ./...`
3. **Manual simulation**:
   ```bash
   # Start server
   go run ./cmd/server/ &
   # Hit endpoint 11 times (limit is 10)
   for i in $(seq 1 11); do curl -s -o /dev/null -w "%{http_code}\n" localhost:8080/api; done
   # Last request should return 429
   ```
```

**Why this matters:** The golden plan pattern (file tables + symbol-level specs + verification procedures) enabled single-pass implementation of an 8-file, 5-area change with zero ambiguity. Category-level specs ("modify classifyRunStatus") force implementers to rediscover symbols, causing divergence and rework.

### Step 4: Decompose into Issues

Analyze the goal and break it into discrete, implementable issues. For each issue define:
- **Title**: Clear action verb (e.g., "Add authentication middleware")
- **Description**: What needs to be done
- **Dependencies**: Which issues must complete first (if any)
- **Acceptance criteria**: How to verify it's done

#### Design Briefs for Rewrites

For any issue that says "rewrite", "redesign", or "create from scratch":
Include a **design brief** (3+ sentences) covering:
1. **Purpose** — what does this component do in the new architecture?
2. **Key artifacts** — what files/interfaces define success?
3. **Workflows** — what sequences must work?

Without a design brief, workers invent design decisions. In ol-571, a spec rewrite issue without a design brief produced output that diverged from the intended architecture.

#### Issue Granularity

- **1-2 independent files** → 1 issue
- **3+ independent files with no code deps** → split into sub-issues (one per file)
  - Example: "Rewrite 4 specs" → 4 sub-issues (4.1, 4.2, 4.3, 4.4)
  - Enables N parallel workers instead of 1 serial worker
- **Shared files between issues** → serialize or assign to same worker

#### Conformance Checks

For each issue's acceptance criteria, derive at least one **mechanically verifiable** conformance check using validation-contract.md types. These checks bridge the gap between spec intent and implementation verification.

| Acceptance Criteria | Conformance Check |
|-----|------|
| "File X exists" | `files_exist: ["X"]` |
| "Function Y is implemented" | `content_check: {file: "src/foo.go", pattern: "func Y"}` |
| "Tests pass" | `tests: "go test ./..."` |
| "Endpoint returns 200" | `command: "curl -s -o /dev/null -w '%{http_code}' localhost:8080/api \| grep 200"` |
| "Config has setting Z" | `content_check: {file: "config.yaml", pattern: "setting_z:"}` |

**Rules:**
- Every issue MUST have at least one conformance check
- Checks MUST use validation-contract.md types: `files_exist`, `content_check`, `command`, `tests`, `lint`
- Prefer `content_check` and `files_exist` (fast, deterministic) over `command` (slower, environment-dependent)
- If acceptance criteria cannot be mechanically verified, flag it as underspecified

### Step 5: Compute Waves

Group issues by dependencies for parallel execution:
- **Wave 1**: Issues with no dependencies (can run in parallel)
- **Wave 2**: Issues depending only on Wave 1
- **Wave 3**: Issues depending on Wave 2
- Continue until all issues assigned

#### Validate Dependency Necessity

For EACH declared dependency, verify:
1. Does the blocked issue modify a file that the blocker also modifies? → **Keep**
2. Does the blocked issue read output produced by the blocker? → **Keep**
3. Is the dependency only logical ordering (e.g., "specs before roles")? → **Remove**

False dependencies reduce parallelism. Pre-mortem judges will also flag these. In ol-571, unnecessary serialization between independent spec rewrites was caught by pre-mortem.

### Step 6: Write Plan Document

**Write to:** `.agents/plans/YYYY-MM-DD-<goal-slug>.md`

```markdown
# Plan: <Goal>

**Date:** YYYY-MM-DD
**Source:** <research doc if any>

## Context
<1-2 paragraphs explaining the problem, current state, and why this change is needed>

## Files to Modify

| File | Change |
|------|--------|
| `path/to/file.go` | Description of change |
| `path/to/new_file.go` | **NEW** — description |

## Boundaries

**Always:** <non-negotiable requirements — security, backward compat, testing, etc.>
**Ask First:** <decisions needing human input before proceeding — in auto mode, logged only>
**Never:** <explicit out-of-scope items preventing scope creep>

## Baseline Audit

| Metric | Command | Result |
|--------|---------|--------|
| <what was measured> | `<grep/wc/ls command used>` | <result> |

## Implementation

### 1. <Change Group Name>

In `path/to/file.go`:

- **Modify `functionName`**: Add `paramName Type` parameter. If `paramName != ""` and condition, return `"value"`.

- **Add `NewStruct`**:
  ```go
  type NewStruct struct {
      FieldName string `json:"field_name,omitempty"`
  }
  ```

- **Key functions to reuse:**
  - `existingHelper()` at `path/to/file.go:123`
  - `anotherFunc()` at `path/to/other.go:456`

### 2. <Next Change Group>

<Same pattern — exact symbols, inline code, reuse references>

## Tests

**`path/to/file_test.go`** — add:
- `TestFunctionName_ScenarioA`: Input X produces output Y
- `TestFunctionName_ScenarioB`: Edge case Z handled correctly

**`path/to/new_test.go`** — **NEW**:
- `TestNewFeature_HappyPath`: Normal flow succeeds
- `TestNewFeature_ErrorCase`: Bad input returns error

## Conformance Checks

| Issue | Check Type | Check |
|-------|-----------|-------|
| Issue 1 | content_check | `{file: "src/auth.go", pattern: "func Authenticate"}` |
| Issue 1 | tests | `go test ./src/auth/...` |
| Issue 2 | files_exist | `["docs/api-v2.md"]` |

## Verification

1. **Unit tests**: `go test ./path/to/ -run "TestFoo" -v`
2. **Full suite**: `go test ./... -short -timeout 120s`
3. **Manual simulation**:
   ```bash
   # Create test scenario
   mkdir -p .test/data
   echo '{"key": "value"}' > .test/data/input.json
   # Run the tool
   ./bin/tool --flag value
   # Verify expected output
   cat .test/data/output.json  # Should show "result"
   ```

## Issues

### Issue 1: <Title>
**Dependencies:** None
**Acceptance:** <how to verify>
**Description:** <what to do — reference Implementation section for symbol-level detail>

### Issue 2: <Title>
**Dependencies:** Issue 1
**Acceptance:** <how to verify>
**Description:** <what to do>

## Execution Order

**Wave 1** (parallel): Issue 1, Issue 3
**Wave 2** (after Wave 1): Issue 2, Issue 4
**Wave 3** (after Wave 2): Issue 5

## Next Steps
- Run `$pre-mortem` to validate plan
- Run `$crank` for autonomous execution
- Or `$implement <issue>` for single issue
```

### Step 7: Create Tasks for In-Session Tracking

**Use TaskCreate tool** for each issue:

```
Tool: TaskCreate
Parameters:
  subject: "<issue title>"
  description: |
    <Full description including:>
    - What to do
    - Acceptance criteria
    - Dependencies: [list task IDs that must complete first]
  activeForm: "<-ing verb form of the task>"
```

**After creating all tasks, set up dependencies:**

```
Tool: TaskUpdate
Parameters:
  taskId: "<task-id>"
  addBlockedBy: ["<dependency-task-id>"]
```

**IMPORTANT: Create persistent issues for ratchet tracking:**

If bd CLI available, create beads issues to enable progress tracking across sessions:
```bash
# Create epic first
bd create --title "<goal>" --type epic --label "planned"

# Create child issues (note the IDs returned)
bd create --title "<wave-1-task>" --body "<description>" --parent <epic-id> --label "planned"
# Returns: na-0001

bd create --title "<wave-2-task-depends-on-wave-1>" --body "<description>" --parent <epic-id> --label "planned"
# Returns: na-0002

# Add blocking dependencies to form waves
bd dep add na-0001 na-0002
# Now na-0002 is blocked by na-0001 → Wave 2
```

**Include conformance checks in issue bodies:**

When creating beads issues, embed the conformance checks from the plan as a fenced validation block in the issue description. This flows to worker validation metadata via $crank:

````
bd create --title "<task>" --body "Description...

\`\`\`validation
{\"files_exist\": [\"src/auth.go\"], \"content_check\": {\"file\": \"src/auth.go\", \"pattern\": \"func Authenticate\"}}
\`\`\`
" --parent <epic-id>
````

**Include cross-cutting constraints in epic description:**

"Always" boundaries from the plan should be added to the epic's description as a `## Cross-Cutting Constraints` section. $crank reads these from the epic (not per-issue) and injects them into every worker task's validation metadata.

**Waves are formed by `blocks` dependencies:**
- Issues with NO blockers → Wave 1 (appear in `bd ready` immediately)
- Issues blocked by Wave 1 → Wave 2 (appear when Wave 1 closes)
- Issues blocked by Wave 2 → Wave 3 (appear when Wave 2 closes)

**`bd ready` returns the current wave** - all unblocked issues that can run in parallel.

Without bd issues, the ratchet validator cannot track gate progress. This is required for `$crank` autonomous execution and `$post-mortem` validation.

### Step 8: Request Human Approval (Gate 2)

**Skip this step if `--auto` flag is set.** In auto mode, proceed directly to Step 9.

**USE AskUserQuestion tool:**

```
Tool: AskUserQuestion
Parameters:
  questions:
    - question: "Plan complete with N tasks in M waves. Approve to proceed?"
      header: "Gate 2"
      options:
        - label: "Approve"
          description: "Proceed to $pre-mortem or $crank"
        - label: "Revise"
          description: "Modify the plan before proceeding"
        - label: "Back to Research"
          description: "Need more research before planning"
      multiSelect: false
```

**Wait for approval before reporting completion.**

### Step 9: Record Ratchet Progress

```bash
ao work ratchet record plan 2>/dev/null || true
```

### Step 10: Report to User

Tell the user:
1. Plan document location
2. Number of issues identified
3. Wave structure for parallel execution
4. Tasks created (in-session task IDs)
5. Next step: `$pre-mortem` for failure simulation, then `$crank` for execution

## Key Rules

- **Read research first** if it exists
- **Explore codebase** to understand current state
- **Identify dependencies** between issues
- **Compute waves** for parallel execution
- **Always write the plan** to `.agents/plans/`

## Examples

### Plan from Research

**User says:** `$plan "add user authentication"`

**What happens:**
1. Agent reads recent research from `.agents/research/2026-02-13-authentication-system.md`
2. Explores codebase to identify integration points
3. Decomposes into 5 issues: middleware, session store, token validation, tests, docs
4. Creates epic `ag-5k2` with 5 child issues in 2 waves
5. Output written to `.agents/plans/2026-02-13-add-user-authentication.md`

**Result:** Epic with dependency graph, conformance checks, and wave structure for parallel execution.

### Plan with Auto Mode

**User says:** `$plan --auto "refactor payment module"`

**What happens:**
1. Agent skips human approval gates
2. Searches knowledge base for refactoring patterns
3. Creates epic and child issues automatically
4. Records ratchet progress

**Result:** Fully autonomous plan creation with 3 waves, 8 issues, ready for `$crank`.

### Plan Cleanup Epic with Audit

**User says:** `$plan "remove dead code"`

**What happens:**
1. Agent runs quantitative audit: 3,003 LOC across 3 packages
2. Creates issues grounded in audit numbers (not vague "cleanup")
3. Each issue specifies exact files and line count reduction
4. Output includes deletion verification checks

**Result:** Scoped cleanup plan with measurable completion criteria (e.g., "Delete 1,500 LOC from pkg/legacy").

### Plan with Implementation Detail (Symbol-Level)

**User says:** `$plan "add stale run detection to ao work rpi status"`

**What happens:**
1. Agent explores codebase, finds `classifyRunStatus` at `rpi_status.go:850`, `phasedState` at `rpi_phased.go:100`
2. Produces file inventory: 4 files to modify, 2 new files
3. Each implementation section names exact functions, parameters, struct fields with JSON tags
4. Tests section lists `TestClassifyRunStatus_StaleWorktree`, `TestDetermineRunLiveness_MissingWorktree` with descriptions
5. Verification section provides manual simulation: create fake stale run, check `ao work rpi status` output

**Result:** Implementer can execute the plan in a single pass without rediscovering any symbol names, reducing implementation time by ~50% and eliminating spec-divergence rework.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| bd create fails | Beads not initialized in repo | Run `bd init --prefix <prefix>` first |
| Dependencies not created | Issues created without explicit `bd dep add` calls | Verify plan output includes dependency commands. Re-run to regenerate |
| Plan too large | Research scope was too broad, resulting in >20 issues | Narrow the goal or split into multiple epics |
| Wave structure incorrect | False dependencies declared (logical ordering, not file conflicts) | Review dependency necessity: does blocked issue modify blocker's files? |
| Conformance checks missing | Acceptance criteria not mechanically verifiable | Add `files_exist`, `content_check`, `tests`, or `command` checks per validation-contract.md |
| Epic has no children | Plan created but bd commands failed silently | Check `bd list --type epic` output; re-run plan with bd CLI available |

## Reference Documents

- [references/examples.md](references/examples.md)
- [references/sdd-patterns.md](references/sdd-patterns.md)
- [references/templates.md](references/templates.md)

---

## References

### examples.md

# Formulate Skill Examples

Detailed examples of formula creation and instantiation scenarios.

---

## Example 1: Simple Multi-Step Formula

**User Request**: "/formulate Add OAuth2 support for GitLab integration"

**Agent Actions**:

```bash
# Phase 0: Context Discovery
mcp__smart-connections-work__lookup --query="OAuth2 GitLab authentication" --limit=10
cat docs/code-map/README.md | grep -i auth

# Phase 0.5: Prior Formula Check
ls .agents/formulas/*.formula.toml | xargs grep -l "OAuth\|GitLab"
bd list --type=epic | grep -i oauth

# Phase 1: Research
# Found existing auth patterns in services/gateway/
# GitLab MCP server already exists at services/mcp-servers/gitlab/

# Phase 2: Decomposition
# Step 1: Add OAuth2 callback endpoint (no deps)
# Step 2: Implement token exchange flow (depends on callback)
# Step 3: Add session token storage (depends on token-exchange)
# Step 4: Update GitLab MCP to use OAuth tokens (depends on session-storage)
# Step 5: Add OAuth2 config to Helm values (no deps, parallel)

# Phase 3: Create Formula Template
```

**Formula output** (`.agents/formulas/oauth-gitlab.formula.toml`):

```toml
# Formula: OAuth2 GitLab Integration
# Reusable pattern for adding OAuth2 to service integrations

formula = "oauth-gitlab"
description = "Add OAuth2 authentication for GitLab integration"
version = 2
type = "workflow"

[vars]
provider = "gitlab"

[[steps]]
id = "callback-endpoint"
title = "Add OAuth2 callback endpoint"
description = """
Add OAuth2 callback endpoint for {{provider}}:
- Create callback route in services/gateway/routes.py
- Implement callback handler in services/gateway/oauth.py
- Handle authorization code exchange
"""
needs = []

[[steps]]
id = "token-exchange"
title = "Implement token exchange flow"
description = """
Implement OAuth2 token exchange:
- Exchange authorization code for access token
- Handle refresh tokens
- Store tokens securely
"""
needs = ["callback-endpoint"]

[[steps]]
id = "session-storage"
title = "Add session token storage"
description = """
Add session-based token storage:
- Create session module in services/gateway/session.py
- Integrate with Redis for token storage
- Handle session expiration
"""
needs = ["token-exchange"]

[[steps]]
id = "mcp-update"
title = "Update {{provider}} MCP to use OAuth tokens"
description = """
Update {{provider}} MCP server to use OAuth:
- Modify services/mcp-servers/{{provider}}/ to use stored tokens
- Remove hardcoded credentials
- Add token refresh logic
"""
needs = ["session-storage"]

[[steps]]
id = "helm-config"
title = "Add OAuth2 config to Helm values"
description = """
Add OAuth2 configuration:
- Add oauth section to charts/ai-platform/values.yaml
- Include clientId, clientSecret placeholders
- Document configuration options
"""
needs = []
```

```bash
# Phase 4: Cook and Pour
# NOTE: bd cook is a planned feature, not yet implemented
bd cook .agents/formulas/oauth-gitlab.formula.toml --persist
# -> Cooked proto: oauth-gitlab

# FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec.
bd mol pour oauth-gitlab
# -> Created mol ai-platform-101 with children:
#    ai-platform-101.callback-endpoint, ai-platform-101.token-exchange, etc.
```

**Result**: 5 steps, computed waves:
- Wave 1: callback-endpoint, helm-config (parallel, different files)
- Wave 2: token-exchange
- Wave 3: session-storage
- Wave 4: mcp-update

---

## Example 2: Reusing an Existing Formula

**Scenario**: User wants to add OAuth2 for a different provider.

**Agent Actions**:

```bash
# Phase 0.5: Find existing formula
ls .agents/formulas/*.formula.toml | xargs grep -l "oauth"
# Found: oauth-gitlab.formula.toml

cat .agents/formulas/oauth-gitlab.formula.toml
# Formula is parameterized with {{provider}}
```

**Agent Response**:

"I found an existing formula `oauth-gitlab.formula.toml` that's parameterized for different
OAuth providers. I can instantiate it with `provider=github`:

```bash
# Pour with different provider variable
# FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec.
bd mol pour oauth-gitlab --var provider=github
```

This will create the same 5-step structure but for GitHub instead of GitLab."

---

## Example 3: Formula From Research

**Scenario**: `$research` was already run, now creating a formula from findings.

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
```

**Formula output** (`.agents/formulas/oauth-hardening.formula.toml`):

```toml
# Formula: OAuth Hardening
# Based on research: 2026-01-03-oauth-investigation.md

formula = "oauth-hardening"
description = "Security hardening for OAuth implementations"
version = 2
type = "workflow"

[[steps]]
id = "token-refresh"
title = "Implement token refresh mechanism"
description = """
Implement automatic token refresh:
- Add refresh token handling
- Implement token rotation
- Handle refresh failures gracefully

Based on research: 2026-01-03-oauth-investigation.md
"""
needs = []

[[steps]]
id = "rate-limiting"
title = "Add OAuth rate limiting"
description = """
Add rate limiting to OAuth endpoints:
- Limit auth attempts per IP
- Add exponential backoff on failures
- Log rate limit events
"""
needs = []

[[steps]]
id = "audit-logging"
title = "Add OAuth audit logging"
description = """
Add comprehensive audit logging:
- Log all auth events
- Include user, timestamp, outcome
- Integrate with SIEM if available
"""
needs = []
```

**Result**: All 3 steps can run in parallel (Wave 1) - no dependencies between them.

---

## Example 4: Complex Dependency Graph Formula

**Scenario**: Feature with multiple parallel tracks that merge.

**Formula output** (`.agents/formulas/multi-tenant.formula.toml`):

```toml
# Formula: Multi-tenant Support
# Complex formula with parallel tracks and merge points

formula = "multi-tenant"
description = "Add multi-tenant support to the platform"
version = 2
type = "workflow"

# Track 1: Database changes
[[steps]]
id = "db-tenant-column"
title = "Add tenant column to all tables"
description = """
Database schema changes for multi-tenancy:
- Add tenant_id column to all entity tables
- Create migration scripts
- Update indexes for tenant queries
"""
needs = []

[[steps]]
id = "db-rls"
title = "Implement row-level security"
description = """
Add PostgreSQL row-level security:
- Create RLS policies per table
- Test isolation between tenants
- Document RLS configuration
"""
needs = ["db-tenant-column"]

# Track 2: API changes (parallel with Track 1)
[[steps]]
id = "api-middleware"
title = "Add tenant middleware"
description = """
Add tenant context middleware:
- Extract tenant from request headers/JWT
- Inject tenant context into request
- Handle missing tenant gracefully
"""
needs = []

[[steps]]
id = "api-endpoints"
title = "Update all endpoints for tenant context"
description = """
Update API endpoints:
- Add tenant filter to all queries
- Validate tenant access on mutations
- Update OpenAPI specs
"""
needs = ["api-middleware"]

# Merge point: depends on both tracks
[[steps]]
id = "integration-tests"
title = "Integration tests for multi-tenant"
description = """
Comprehensive integration testing:
- Test tenant isolation
- Test cross-tenant access prevention
- Performance testing with multiple tenants
"""
needs = ["db-rls", "api-endpoints"]
```

**Computed Waves:**
- Wave 1: db-tenant-column, api-middleware (parallel start)
- Wave 2: db-rls, api-endpoints (parallel, different tracks)
- Wave 3: integration-tests (merge point, needs both tracks done)

---

## Example 5: Quick Formula (3 Steps or Less)

**Scenario**: Small goal, simple formula.

**Formula output** (`.agents/formulas/rate-limiting-quick.formula.toml`):

```toml
# Formula: Rate Limiting (Quick)
# Simple formula for adding rate limiting

formula = "rate-limiting-quick"
description = "Quick rate limiting implementation"
version = 2
type = "workflow"

[vars]
endpoint = "/api/auth"

[[steps]]
id = "impl"
title = "Add rate limiting to {{endpoint}}"
description = """
Implement rate limiting:
- Add RateLimitMiddleware to services/gateway/middleware.py
- Configure limits for {{endpoint}}
- Return 429 with Retry-After header
"""
needs = []

[[steps]]
id = "config"
title = "Add rate limit config to Helm values"
description = """
Add configuration:
- Add rateLimit section to charts/ai-platform/values.yaml
- Include requestsPerMinute and burstSize
- Document in values.yaml comments
"""
needs = []
```

**Summary output:**
```markdown
# Formula Instantiated: Rate Limiting for Auth

**Formula:** `.agents/formulas/rate-limiting-quick.formula.toml`
**Steps:** 2 steps (Wave 1 only - all parallel)

| ID | Step | Ready |
|----|------|-------|
| xxx.impl | Add rate limiting to /api/auth | Ready |
| xxx.config | Add rate limit config to Helm values | Ready |

Both can run in parallel (different files). Use:
```bash
bd ready
$implement xxx.impl
```
```

---

## Anti-Pattern Examples

### WRONG: Using Old Formula Format

```toml
# WRONG - This format will fail with bd cook! (NOTE: bd cook is planned, not yet implemented)
[formula]                              # WRONG: use top-level `formula = "..."`
name = "oauth-gitlab"
version = "1.0.0"                      # WRONG: use integer version = 2

[variables]                            # WRONG: use [vars] with simple values
provider = { type = "string" }         # WRONG: complex type definitions

[[tasks]]                              # WRONG: use [[steps]]
id = "callback-endpoint"
title = "Add callback"
type = "feature"                       # WRONG: no type in steps
priority = "P1"                        # WRONG: no priority in steps
depends_on = []                        # WRONG: use needs = []
files_affected = ["..."]               # WRONG: no files_affected

[waves]                                # WRONG: waves are computed, not declared
1 = ["callback-endpoint", "helm-config"]
```

**Correct version:**

```toml
formula = "oauth-gitlab"
description = "Add OAuth2 authentication for GitLab integration"
version = 2
type = "workflow"

[vars]
provider = "gitlab"

[[steps]]
id = "callback-endpoint"
title = "Add callback"
description = "Add OAuth2 callback endpoint..."
needs = []
```

### WRONG: Children Depending on Epic

```bash
# DON'T DO THIS with --immediate mode
bd create "Epic: OAuth" --type epic
# -> ai-platform-400

bd create "Add callback" --type feature
# -> ai-platform-401

bd dep add ai-platform-401 ai-platform-400  # WRONG!
# Result: ai-platform-401 will NEVER become ready because
# the epic can't be closed until children are done (deadlock)
```

### WRONG: Skipping Description in Steps

```toml
# WRONG - description is required!
[[steps]]
id = "impl"
title = "Implement feature"
needs = []
# Missing description field - will fail validation!
```

### WRONG: Creating Too Many Steps

```toml
# DON'T DO THIS - 15+ steps is too many
# Better: Break into 2-3 formulas of 5 steps each
# Or: Create first 5, complete, then plan next 5
```

### WRONG: One-Off Plans for Repeatable Patterns

```bash
# DON'T DO THIS
# Creating a new plan document every time you add OAuth to a service

# Better: Create a formula template once, pour with variables
# FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec.
bd mol pour oauth-gitlab --var provider=github
```

### sdd-patterns.md

# SDD Patterns — Boundaries and Conformance Checks

> Reference doc for $plan. Loaded JIT when agents need examples.

## What Are Boundaries?

Boundaries define the scope of a plan using three tiers:

| Tier | Purpose | Example |
|------|---------|---------|
| **Always** | Non-negotiable constraints applied to every issue | "All endpoints require auth middleware" |
| **Ask First** | Decisions requiring human input before proceeding | "Which rate limit values to use?" |
| **Never** | Explicit out-of-scope items preventing scope creep | "No new database tables" |

**Always** boundaries become cross-cutting constraints — $crank injects them into every worker task's validation metadata. **Ask First** boundaries are logged in auto mode and prompted in interactive mode. **Never** boundaries are guardrails for workers and pre-mortem judges.

## What Are Conformance Checks?

Conformance checks are mechanically verifiable assertions derived from acceptance criteria. They bridge the gap between "what success looks like" (prose) and "how to verify it" (automation).

**The derivation chain:**
```
Acceptance Criteria (prose) → Conformance Check (validation-contract.md type) → Worker Validation Metadata
```

Each check uses one of the validation-contract.md types:

| Type | Use When | Example |
|------|----------|---------|
| `files_exist` | Task creates new files | `["src/auth/middleware.go", "tests/auth_test.go"]` |
| `content_check` | Task implements specific functions/patterns | `{file: "src/auth.go", pattern: "func Authenticate"}` |
| `command` | Task produces verifiable runtime behavior | `"go build ./..."` |
| `tests` | Task has associated tests | `"go test ./src/auth/..."` |
| `lint` | Task must maintain code quality | `"ruff check src/"` |

**Rules:**
- Every acceptance criterion MUST have at least one conformance check
- Prefer `content_check` and `files_exist` (fast, deterministic) over `command` (environment-dependent)
- If an acceptance criterion can't be mechanically verified, it's underspecified — rewrite it

## Example 1: API Feature — "Add Rate Limiting"

### Boundaries

**Always:**
- Backward compatible — existing endpoints continue to work without rate limit headers
- All rate-limited endpoints require auth middleware
- Tests cover both under-limit and over-limit cases

**Ask First:**
- Rate limit values (requests per minute) — depends on infrastructure capacity
- Whether to rate-limit internal service-to-service calls

**Never:**
- Rate limiting on health check endpoints (`/healthz`, `/readyz`)
- Custom rate limit configuration per user (that's a separate feature)

### Conformance Checks

| Issue | Check Type | Check |
|-------|-----------|-------|
| Add rate limit middleware | content_check | `{file: "src/middleware/ratelimit.go", pattern: "func RateLimitMiddleware"}` |
| Add rate limit middleware | tests | `go test ./src/middleware/...` |
| Add rate limit middleware | content_check | `{file: "src/middleware/ratelimit.go", pattern: "X-RateLimit-Remaining"}` |
| Wire middleware to routes | content_check | `{file: "src/routes/api.go", pattern: "RateLimitMiddleware"}` |
| Wire middleware to routes | command | `go build ./...` |
| Add rate limit tests | files_exist | `["tests/ratelimit_test.go"]` |
| Add rate limit tests | tests | `go test ./tests/ratelimit_test.go -v` |

### Cross-Cutting Constraints (from "Always")

```json
[
  {"name": "auth-required", "type": "content_check", "file": "src/routes/api.go", "pattern": "AuthMiddleware"},
  {"name": "builds-clean", "type": "command", "command": "go build ./..."},
  {"name": "tests-pass", "type": "tests", "command": "go test ./..."}
]
```

## Example 2: Refactoring — "Extract Shared Library"

### Boundaries

**Always:**
- No behavior change — all existing tests must pass before and after
- Extracted functions maintain the same signatures
- No new dependencies added

**Ask First:**
- Package naming conventions (e.g., `pkg/shared` vs `internal/common`)
- Whether to add godoc comments during extraction

**Never:**
- New features or behavior changes during extraction
- Refactoring unrelated code "while we're at it"

### Conformance Checks

| Issue | Check Type | Check |
|-------|-----------|-------|
| Create shared package | files_exist | `["pkg/shared/helpers.go"]` |
| Create shared package | content_check | `{file: "pkg/shared/helpers.go", pattern: "package shared"}` |
| Move functions to shared | content_check | `{file: "pkg/shared/helpers.go", pattern: "func ParseConfig"}` |
| Move functions to shared | content_check | `{file: "pkg/shared/helpers.go", pattern: "func ValidateInput"}` |
| Update imports in callers | command | `go build ./...` |
| Update imports in callers | tests | `go test ./...` |
| Remove duplicates from source | command | `! grep -r 'func ParseConfig' src/old/ 2>/dev/null` |

### Cross-Cutting Constraints

```json
[
  {"name": "tests-unchanged", "type": "tests", "command": "go test ./..."},
  {"name": "no-new-deps", "type": "command", "command": "go mod tidy && git diff --exit-code go.mod"}
]
```

## Example 3: Documentation — "Rewrite API Docs"

### Boundaries

**Always:**
- All public endpoints documented
- Each endpoint has request/response examples
- Links to source code reference valid files

**Ask First:**
- Whether to include curl examples or SDK examples
- Documentation framework (plain markdown vs generated)

**Never:**
- Implementation details or internal architecture
- Auto-generated API reference (that's a separate tool)

### Conformance Checks

| Issue | Check Type | Check |
|-------|-----------|-------|
| Write endpoint docs | files_exist | `["docs/api/endpoints.md"]` |
| Write endpoint docs | content_check | `{file: "docs/api/endpoints.md", pattern: "## GET /api/users"}` |
| Write endpoint docs | content_check | `{file: "docs/api/endpoints.md", pattern: "## POST /api/users"}` |
| Write auth docs | files_exist | `["docs/api/authentication.md"]` |
| Write auth docs | content_check | `{file: "docs/api/authentication.md", pattern: "Authorization: Bearer"}` |
| Add examples | content_check | `{file: "docs/api/endpoints.md", pattern: "### Example"}` |
| Validate links | command | `./scripts/check-doc-links.sh docs/api/` |

### Cross-Cutting Constraints

```json
[
  {"name": "all-endpoints-covered", "type": "content_check", "file": "docs/api/endpoints.md", "pattern": "## (GET|POST|PUT|DELETE)"},
  {"name": "examples-present", "type": "content_check", "file": "docs/api/endpoints.md", "pattern": "### Example"}
]
```

## Example 4: Implementation Detail — "Add Stale Run Detection"

This example demonstrates symbol-level implementation detail — the key differentiator between vague plans and actionable specs.

### Files to Modify

| File | Change |
|------|--------|
| `cli/cmd/ao/rpi_status.go` | Add worktree check to `classifyRunStatus`, add `Reason` field to `rpiRunInfo` |
| `cli/cmd/ao/rpi_cleanup.go` | **NEW** — `ao work rpi cleanup` command |
| `cli/cmd/ao/rpi_phased.go` | Add terminal metadata fields to `phasedState` |
| `cli/internal/config/config.go` | Add `RPIConfig` with `WorktreeMode` |

### Implementation (Symbol-Level)

#### 1. Stale Run Detection in `rpi_status.go`

- **Modify `classifyRunStatus`**: Add check for `state.TerminalStatus != ""` — return it directly. Add check for `state.WorktreePath != ""` with `os.Stat()` — if directory gone, return `"stale"`.

- **Add `Reason` field to `rpiRunInfo`**:
  ```go
  Reason string `json:"reason,omitempty"` // why a run is stale/failed
  ```

- **Modify `determineRunLiveness`**: If `state.WorktreePath != ""` and `os.Stat(state.WorktreePath)` fails, short-circuit to `return false, hb` without probing tmux.

- **Key functions to reuse:**
  - `readRunHeartbeat()` at `rpi_phased.go:1963`
  - `checkTmuxSessionAlive()` at `rpi_status.go:896`
  - `parsePhasedState()` at `rpi_phased.go:1924`

#### 2. Terminal Metadata in `rpi_phased.go`

- **Add fields to `phasedState`**:
  ```go
  TerminalStatus string `json:"terminal_status,omitempty"` // interrupted, failed, stale, completed
  TerminalReason string `json:"terminal_reason,omitempty"`
  TerminatedAt   string `json:"terminated_at,omitempty"`
  ```

### Tests (Named Functions)

**`cli/cmd/ao/rpi_status_test.go`** — add:
- `TestClassifyRunStatus_StaleWorktree`: Run with `worktree_path` pointing to nonexistent dir → status "stale"
- `TestClassifyRunStatus_TerminalMetadata`: Run with `terminal_status` set → uses that status directly
- `TestDetermineRunLiveness_MissingWorktree`: Worktree path gone → not active

**`cli/cmd/ao/rpi_cleanup_test.go`** — **NEW**:
- `TestCleanupStaleRun`: Create stale registry entry, run cleanup, verify terminal metadata written
- `TestCleanupActiveRunUntouched`: Create active (fresh heartbeat) entry, verify unchanged
- `TestCleanupDryRun`: Dry-run produces output but doesn't modify state

### Verification

1. **Unit tests**: `cd cli && go test ./cmd/ao/ -run "TestClassifyRunStatus|TestCleanup" -v`
2. **Manual stale simulation**:
   ```bash
   mkdir -p .agents/rpi/runs/fakestale
   echo '{"schema_version":1,"run_id":"fakestale","phase":2,"worktree_path":"/nonexistent"}' \
     > .agents/rpi/runs/fakestale/phased-state.json
   ao work rpi status           # Should show "stale" not "running"
   ao work rpi cleanup --all --dry-run   # Preview
   ao work rpi cleanup --all             # Fix
   ao work rpi status                    # Should show "stale" with reason
   ```

### Why This Format Works

Compared to a category-level spec like "Add stale worktree detection to `classifyRunStatus`", the implementation detail above tells the worker:
- The exact parameter name (`state.TerminalStatus`)
- The exact condition (`os.Stat(state.WorktreePath)` fails)
- The exact return value (`"stale"`)
- Where to find existing code (`readRunHeartbeat()` at `rpi_phased.go:1963`)
- What to name tests (`TestClassifyRunStatus_StaleWorktree`)
- How to verify manually (create fake stale run, check output)

This enabled single-pass implementation of an 8-file change with zero spec-divergence.

## Cross-Cutting Constraints: How They Work

"Always" boundaries become cross-cutting constraints that $crank injects into **every** worker task:

```
Plan "Always" boundaries
    ↓
$crank reads plan → extracts Always
    ↓
Converts to validation-contract.md checks (flat array):
  [{"name": "...", "type": "content_check|command|tests|...", ...fields...}]
    ↓
Injected into every TaskCreate's metadata.validation.cross_cutting
    ↓
Workers validated against per-task checks + cross-cutting checks
```

**Schema:** Each cross-cutting check is a flat object with:
- `name` (string): Human-readable label
- `type` (string): One of `files_exist`, `content_check`, `command`, `tests`, `lint`
- Remaining fields: Same as the corresponding validation-contract.md type

This keeps the schema flat and consistent with existing validation types — no nested meta-types.

### templates.md

# Formula Templates Reference

Detailed templates for formula files and plan summaries.

---

## Formula File Template (.formula.toml)

**Location:** `.agents/formulas/{topic-slug}.formula.toml`

### Structure Overview

| Section | Purpose |
|---------|---------|
| Top-level fields | Formula metadata (`formula`, `description`, `version`, `type`) |
| `[vars]` | Simple key-value pairs for parameterization |
| `[[steps]]` | Array of implementation steps with dependencies |

### Full Template

```toml
# Formula: {Goal Name}
# Reusable pattern for creating {description}
# Created: YYYY-MM-DD

# REQUIRED: Top-level fields (NOT in a [formula] table!)
formula = "{topic-slug}"
description = "{Detailed description of what this formula produces}"
version = 2
type = "workflow"  # MUST be: workflow | expansion | aspect

# OPTIONAL: Variables for parameterization
# Use {{var_name}} syntax in step descriptions
[vars]
service_name = "default-service"
base_path = "services/"

# Steps define the work items - each becomes a child issue when poured
# Order doesn't matter - dependencies define execution order via `needs`

[[steps]]
id = "core"
title = "Add {{service_name}} core implementation"
description = """
Implement the core {{service_name}} functionality:
- Add main module at {{base_path}}{{service_name}}/core.py
- Include error handling and logging
- Follow existing patterns in the codebase

Files affected:
- {{base_path}}{{service_name}}/core.py
- {{base_path}}{{service_name}}/__init__.py

Acceptance criteria:
- Module is importable
- Passes unit tests
- Handles edge cases gracefully
"""
needs = []  # Wave 1 - no dependencies

[[steps]]
id = "config"
title = "Add {{service_name}} configuration"
description = """
Add configuration for {{service_name}}:
- Update charts/ai-platform/values.yaml with new config section
- Add environment variable mappings
- Document configuration options in values.yaml comments

Files affected:
- charts/ai-platform/values.yaml
- charts/ai-platform/templates/configmap.yaml

Acceptance criteria:
- Config values documented
- Defaults are sensible
- Works in dev and prod environments
"""
needs = []  # Wave 1 - can run parallel with core

[[steps]]
id = "tests"
title = "{{service_name}} integration tests"
description = """
Add comprehensive tests for {{service_name}}:
- Unit tests for core functionality
- Integration tests for API endpoints
- Ensure >80% coverage

Files affected:
- tests/unit/test_{{service_name}}.py
- tests/integration/test_{{service_name}}_e2e.py

Acceptance criteria:
- Happy path covered
- Error cases handled
- CI passes
"""
needs = ["core"]  # Wave 2 - depends on core implementation

[[steps]]
id = "docs"
title = "{{service_name}} documentation"
description = """
Document {{service_name}}:
- API reference in docs/api/
- Update README with usage examples
- Add architecture decision record if needed

Files affected:
- docs/api/{{service_name}}.md
- README.md

Acceptance criteria:
- Usage examples work
- API fully documented
"""
needs = ["core", "config"]  # Wave 2 - depends on both core and config
```

### Field Reference

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `formula` | Yes | string | Unique identifier (slug) at TOP LEVEL |
| `description` | Yes | string | What the formula creates |
| `version` | Yes | integer | Use `2` |
| `type` | Yes | string | `workflow`, `expansion`, or `aspect` |
| `[vars]` | No | table | Simple `key = "value"` pairs |
| `[[steps]]` | Yes | array | Step definitions |
| `steps.id` | Yes | string | Unique step identifier |
| `steps.title` | Yes | string | Short step title (can use {{vars}}) |
| `steps.description` | Yes | string | Detailed implementation guidance |
| `steps.needs` | Yes | array | Step IDs this depends on (empty = Wave 1) |

### WRONG Format (Do NOT Use)

```toml
# WRONG - Do not use this format!
[formula]                              # WRONG: formula is a top-level string, not a table
name = "topic-slug"                    # WRONG: use `formula = "..."` at top level
version = "1.0.0"                      # WRONG: use integer `version = 2`

[variables]                            # WRONG: use [vars] with simple values
component = { type = "string" }        # WRONG: complex type definitions not supported

[[tasks]]                              # WRONG: use [[steps]]
title = "..."
type = "feature"                       # WRONG: no type field in steps
priority = "P1"                        # WRONG: no priority field in steps
wave = 1                               # WRONG: no wave field (computed from needs)
depends_on = ["..."]                   # WRONG: use needs = [...]
files = ["..."]                        # WRONG: no files field

[waves]                                # WRONG: waves are computed, not declared
1 = ["step1", "step2"]
```

### Wave Computation

Waves are computed from the `needs` field:

| Wave | Rule | Example |
|------|------|---------|
| Wave 1 | `needs = []` | core, config |
| Wave 2 | All `needs` are Wave 1 | tests (needs core), docs (needs core, config) |
| Wave N | All `needs` are Wave N-1 or earlier | - |

### Variable Substitution

Variables defined in `[vars]` can be used in step descriptions with `{{var}}` syntax:

```toml
[vars]
service_name = "rate-limiter"
requests_per_minute = "100"

[[steps]]
id = "impl"
title = "Implement {{service_name}}"
description = "Configure {{requests_per_minute}} requests per minute"
needs = []
```

### Using bd cook

> **Note:** `bd cook` is a planned feature, not yet implemented.

```bash
# Preview what would be created
bd cook .agents/formulas/{topic-slug}.formula.toml --dry-run

# Cook and save proto to database
bd cook .agents/formulas/{topic-slug}.formula.toml --persist

# Cook with variable overrides
bd cook .agents/formulas/{topic-slug}.formula.toml --persist \
  --var service_name=auth-middleware

# Then pour to create actual issues
# FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec.
bd mol pour {topic-slug}
```

---

## Companion Plan Document Template

**Location:** `.agents/formulas/{topic-slug}.md`

### Tag Vocabulary (REQUIRED)

Document type tag: `formula` (required first)

**Examples:**
- `[formula, agents, kagent]` - KAgent implementation formula
- `[formula, data, neo4j]` - GraphRAG implementation formula
- `[formula, auth, security]` - OAuth2 implementation formula
- `[formula, ci-cd, tekton]` - Tekton pipeline formula

### Full Template

```markdown
---
date: YYYY-MM-DD
type: Formula
goal: "[Goal description]"
tags: [formula, domain-tag, optional-tech-tag]
formula: "{topic-slug}.formula.toml"
epic: "[beads epic ID, if instantiated]"
status: TEMPLATE | INSTANTIATED
---

# Formula: [Goal]

## Overview
[2-3 sentence summary of what this formula creates and when to use it]

## Variables

| Variable | Default | Description |
|----------|---------|-------------|
| service_name | default-service | Name of the service |
| base_path | services/ | Base path for source files |

## Steps (Dependency Order)

| ID | Title | Needs | Wave |
|----|-------|-------|------|
| core | Add core implementation | - | 1 |
| config | Add configuration | - | 1 |
| tests | Integration tests | core | 2 |
| docs | Documentation | core, config | 2 |

## Dependency Graph

```
Wave 1 (No Dependencies):
  core: Add core implementation
  config: Add configuration
       |
       v unblocks
Wave 2 (Depends on Wave 1):
  tests: Integration tests (needs: core)
  docs: Documentation (needs: core, config)
```

## Wave Execution Order

| Wave | Steps | Can Parallel | Notes |
|------|-------|--------------|-------|
| 1 | core, config | Yes | No dependencies, different files |
| 2 | tests, docs | Yes | Both depend on Wave 1, different files |

**Wave Computation Rules:**
- **Wave 1:** All steps with `needs = []`
- **Wave N:** Steps where all `needs` are in Wave N-1 or earlier
- **Can Parallel:** "Yes" if steps in same wave affect different files

## Files to Modify

| File | Change |
|------|--------|
| `{{base_path}}{{service_name}}/core.py` | Add core module with main logic |
| `{{base_path}}{{service_name}}/config.py` | **NEW** — configuration handling |
| `tests/unit/test_{{service_name}}.py` | **NEW** — unit tests |

## Implementation

### 1. Core Module

In `{{base_path}}{{service_name}}/core.py`:

- **Create `ServiceHandler` class** with `__init__(self, config: ServiceConfig)` and `process(self, request: Request) -> Response`
- **Key functions to reuse:**
  - `validate_request()` at `{{base_path}}common/validation.py:45`
  - `format_response()` at `{{base_path}}common/response.py:23`

### 2. Configuration

In `{{base_path}}{{service_name}}/config.py`:

- **Add `ServiceConfig` dataclass:**
  ```python
  @dataclass
  class ServiceConfig:
      service_name: str = "{{service_name}}"
      base_path: str = "{{base_path}}"
  ```

## Tests

**`tests/unit/test_{{service_name}}.py`** — **NEW**:
- `test_service_handler_happy_path`: Valid request returns expected response
- `test_service_handler_invalid_input`: Bad request raises ValueError
- `test_config_defaults`: ServiceConfig has correct defaults

## Verification

1. **Unit tests**: `pytest tests/unit/test_{{service_name}}.py -v`
2. **Build check**: `python -c "from {{base_path.replace('/', '.')}}{{service_name}} import core"`
3. **Manual test**:
   ```bash
   python -c "
   from {{base_path.replace('/', '.')}}{{service_name}}.core import ServiceHandler
   handler = ServiceHandler()
   print(handler.process({'data': 'test'}))
   "
   ```

## Implementation Notes
[Key decisions, patterns to follow, risks identified]

## Usage

### Cook and Pour

> **Note:** `bd cook` is a planned feature, not yet implemented.

```bash
# Preview what would be created
bd cook .agents/formulas/{topic-slug}.formula.toml --dry-run

# Cook proto to database
bd cook .agents/formulas/{topic-slug}.formula.toml --persist

# Pour to create actual issues
# FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec.
bd mol pour {topic-slug}

# With variable overrides
# FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec.
bd mol pour {topic-slug} --var service_name=rate-limiter
```

## Next Steps
Run `$crank <epic-id>` for hands-free execution, or `/implement-wave <epic-id>` for supervised.
```

---

## Formula Summary Template (Crank Handoff)

Output this after cooking/pouring a formula. This is the **handoff to crank**.

```markdown
---

# Formula Instantiated: [Goal Description]

**Formula:** `.agents/formulas/{topic-slug}.formula.toml`
**Epic:** `<rig-prefix>-xxx`
**Plan:** `.agents/formulas/{topic-slug}.md`
**Steps:** N steps across M waves

---

## Wave Execution Order

| Wave | Steps | Can Parallel | Ready Now |
|------|-------|--------------|-----------|
| 1 | xxx.core, xxx.config | Yes | Ready |
| 2 | xxx.tests, xxx.docs | Yes | Blocked by Wave 1 |

## Steps Created

| ID | Step | Needs |
|----|------|-------|
| xxx.core | Add core implementation | - |
| xxx.config | Add configuration | - |
| xxx.tests | Integration tests | core |
| xxx.docs | Documentation | core, config |

## Dependency Graph

```
Wave 1 (needs = []):
  xxx.core: Add core implementation
  xxx.config: Add configuration
       |
       v unblocks
Wave 2 (depends on Wave 1):
  xxx.tests: Integration tests (needs: core)
  xxx.docs: Documentation (needs: core, config)
```

---

## Ready for Execution

### Pre-Flight Checklist

- [x] Formula cooked with `bd cook --persist` <!-- FUTURE: bd cook not yet implemented -->
- [x] Mol poured with `bd mol pour` <!-- FUTURE: bd mol not yet implemented. See skills/beads/references/MOLECULES.md for design spec. -->
- [x] Steps have proper dependencies via `needs`
- [ ] External requirements: [list any, e.g., "API key configured"]

### Execute

**Autonomous (overnight, parallel via polecats):**
```bash
$crank xxx              # Full auto until epic closed
```

**Supervised (sequential, same session):**
```bash
/implement-wave xxx     # One wave at a time
```

### Alternative: Manual Execution

```bash
# Implement one at a time
bd ready
$implement xxx.core
```
```


---

## Scripts

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0

check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "SKILL.md has name: plan" "grep -q '^name: plan' '$SKILL_DIR/SKILL.md'"
check "references/ directory exists" "[ -d '$SKILL_DIR/references' ]"
check "references/ has at least 2 files" "[ \$(ls '$SKILL_DIR/references/' | wc -l) -ge 2 ]"
check "SKILL.md mentions .agents/plans/ output path" "grep -q '\.agents/plans/' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions waves" "grep -qi 'wave' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions dependencies" "grep -qi 'dependencies\|depend' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions bd for issue tracking" "grep -q 'bd ' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions TaskList for tracking" "grep -q 'TaskList\|TaskCreate' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions conformance checks" "grep -qi 'conformance' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions --auto flag" "grep -q '\-\-auto' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions Explore agent" "grep -qi 'explore' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


