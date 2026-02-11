# SDD Patterns — Boundaries and Conformance Checks

> Reference doc for /plan. Loaded JIT when agents need examples.

## What Are Boundaries?

Boundaries define the scope of a plan using three tiers:

| Tier | Purpose | Example |
|------|---------|---------|
| **Always** | Non-negotiable constraints applied to every issue | "All endpoints require auth middleware" |
| **Ask First** | Decisions requiring human input before proceeding | "Which rate limit values to use?" |
| **Never** | Explicit out-of-scope items preventing scope creep | "No new database tables" |

**Always** boundaries become cross-cutting constraints — /crank injects them into every worker task's validation metadata. **Ask First** boundaries are logged in auto mode and prompted in interactive mode. **Never** boundaries are guardrails for workers and pre-mortem judges.

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

## Cross-Cutting Constraints: How They Work

"Always" boundaries become cross-cutting constraints that /crank injects into **every** worker task:

```
Plan "Always" boundaries
    ↓
/crank reads plan → extracts Always
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
