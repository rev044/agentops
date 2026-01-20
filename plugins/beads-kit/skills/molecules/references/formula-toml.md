# Formula TOML Syntax

Formulas are TOML templates that define reusable workflow structures. They are
"cooked" into protos which can then be poured (persistent) or wisped (ephemeral).

## Basic Structure

```toml
# Required top-level fields (NOT in a [formula] table!)
formula = "topic-slug"
description = "What this formula produces"
version = 2
type = "workflow"  # workflow | expansion | aspect

# Optional: Variables for parameterization
[vars]
component = "default-service"
base_path = "services/"

# Steps define the work items (become child issues when poured)
[[steps]]
id = "setup"
title = "Setup {{component}} environment"
description = """
Set up the {{component}} environment:
- Create directory structure at {{base_path}}{{component}}/
- Initialize configuration files
"""
needs = []  # No dependencies = Wave 1

[[steps]]
id = "implement"
title = "Implement {{component}}"
description = """
Implement core {{component}} functionality:
- Add main module
- Include error handling
"""
needs = []  # Parallel with setup (Wave 1)

[[steps]]
id = "test"
title = "Test {{component}}"
description = """
Add tests for {{component}}:
- Unit tests
- Integration tests
"""
needs = ["implement"]  # Depends on implement = Wave 2
```

## Field Reference

### Required Top-Level Fields

| Field | Type | Description |
|-------|------|-------------|
| `formula` | string | Unique identifier (slug) |
| `description` | string | What the formula creates |
| `version` | integer | Schema version (use `2`) |
| `type` | string | Must be: `workflow`, `expansion`, or `aspect` |

### Variables Section

```toml
[vars]
service_name = "default-service"
base_path = "services/"
requests_per_minute = "100"
```

- Simple key-value pairs only (no complex types)
- Use `{{var_name}}` syntax in step descriptions
- Variables substituted during `bd cook --mode=runtime` or `bd mol pour --var`

### Steps Array

Each `[[steps]]` entry requires:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique step identifier |
| `title` | string | Short step title |
| `description` | string | Detailed implementation guidance |
| `needs` | array | Step IDs this depends on (empty = Wave 1) |

**Wave Computation:**
- Wave 1: Steps with `needs = []`
- Wave N: Steps where all `needs` are in Wave N-1 or earlier

## Common Mistakes

**WRONG format (do NOT use):**

```toml
# WRONG: formula in a table
[formula]
name = "my-workflow"

# WRONG: version as string
version = "1.0.0"

# WRONG: invalid type
type = "feature"

# WRONG: tasks instead of steps
[[tasks]]
name = "do something"

# WRONG: unsupported step fields
[[steps]]
priority = 1
wave = 1
files = ["path/to/file.py"]
depends_on = ["other"]  # Use 'needs' not 'depends_on'
```

**CORRECT format:**

```toml
# TOP-LEVEL fields
formula = "my-workflow"
description = "..."
version = 2
type = "workflow"

[[steps]]
id = "core"
title = "..."
description = "..."
needs = []
```

## Full Example

```toml
# Formula: Rate Limiting
# Created: 2025-01-08

formula = "rate-limiting"
description = "Add rate limiting middleware to API gateway with configurable limits"
version = 2
type = "workflow"

[vars]
requests_per_minute = "100"
burst_size = "20"

[[steps]]
id = "middleware"
title = "Add rate limit middleware"
description = """
Implement rate limiting middleware:
- Add RateLimitMiddleware class to services/gateway/middleware.py
- Use token bucket algorithm with {{requests_per_minute}} rpm
- Burst allowance of {{burst_size}} requests
- Return 429 with Retry-After header
"""
needs = []

[[steps]]
id = "config"
title = "Add rate limit Helm config"
description = """
Add configuration for rate limiting:
- Add rateLimit section to charts/ai-platform/values.yaml
- Include requestsPerMinute and burstSize settings
- Document in values.yaml comments
"""
needs = []

[[steps]]
id = "tests"
title = "Rate limit integration tests"
description = """
Add comprehensive tests:
- Unit tests for middleware logic
- Integration tests for rate limiting behavior
- Test 429 response and headers
"""
needs = ["middleware"]
```

## Formula Types

| Type | Purpose |
|------|---------|
| `workflow` | Standard workflow with sequential/parallel steps |
| `expansion` | Expands into existing workflow (attachment) |
| `aspect` | Cross-cutting concern (applied to multiple workflows) |

## Search Paths

When cooking, bd searches for formulas in order:
1. `.beads/formulas/` (project)
2. `~/.beads/formulas/` (user)
3. `$GT_ROOT/.beads/formulas/` (orchestrator, if GT_ROOT set)

## Related

- **Cooking**: See `cooking.md` for `bd cook` usage
- **Lifecycle**: See `mol-lifecycle.md` for full molecule workflow
- **Skill**: See `formulate/` for formula creation
