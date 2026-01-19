---
name: formulate
description: >
  This skill should be used when the user asks to "formulate", "create a formula",
  "create a plan", "plan implementation", "break down into tasks", "decompose into features",
  "create beads issues from research", "what issues should we create",
  "plan out the work", or needs to create reusable .formula.toml templates
  that convert goals into executable beads issues.
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Read,Write,Edit,Bash,Grep,Glob,Task"
skills:
  - beads
  - research
---

# Formulate Skill

Create reusable formula templates (.formula.toml) that define structured implementation
patterns. Formulas produce beads issues with proper dependencies and wave computation
for parallel execution.

## Overview

**Core Purpose**: Transform a goal into trackable beads issues with dependency ordering and
wave-based parallelization for `/crank` and `/implement-wave`.

**Typical Flow**:
1. `/research` - Deep exploration
2. **Native plan mode** (`Shift+Tab` x2) - Approach decisions, creates `plan.md`
3. Context clears on plan acceptance
4. **`/formulate`** - Creates beads issues from the plan
5. `/implement`, `/implement-wave`, or `/crank` - Execute

**Key Capabilities**:
- 6-tier context discovery hierarchy
- Prior formula discovery to prevent duplicates
- Feature decomposition with dependency modeling
- Formula template creation with proper TOML structure (optional)
- Auto-instantiation via `bd cook`
- Beads issue creation with epic-child relationships
- Wave computation for parallel execution

**Formulas vs Direct Issues**:
- **Formula**: Reusable template (.formula.toml) - can be instantiated multiple times
- **Direct (--immediate)**: One-time beads creation - specific to current goal

**When to Use**: After native plan mode clears context, or any time work needs 2+ discrete issues.
**When NOT to Use**: Single task (use `/implement`), exploratory (use `/research`).

### Flags

| Flag | Description |
|------|-------------|
| `--immediate` | Skip formula creation, directly create beads issues |
| `--cook` | Auto-run `bd cook` after formula creation |
| `--dry-run` | Preview formula output without writing files |

---

## Instructions

### Phase 0: Rig Detection

**CRITICAL**: All `.agents/` artifacts go to `~/gt/.agents/<rig>/` based on the primary codebase being formulated.

**Detection Logic**:
1. Identify which rig's code is being formulated (e.g., files in `~/gt/ai-platform/` -> `ai-platform`)
2. If formulating across multiple rigs, use `_cross-rig`
3. If unknown/unclear, ask user

| Files Being Read | Target Rig | Output Base |
|------------------|------------|-------------|
| `~/gt/athena/**` | `athena` | `~/gt/.agents/athena/` |
| `~/gt/hephaestus/**` | `hephaestus` | `~/gt/.agents/hephaestus/` |
| `~/gt/daedalus/**` | `daedalus` | `~/gt/.agents/daedalus/` |
| Multiple rigs | `_cross-rig` | `~/gt/.agents/_cross-rig/` |

```bash
# Set RIG variable for use in output paths
RIG="athena"  # or hephaestus, daedalus, _cross-rig
mkdir -p ~/gt/.agents/$RIG/formulas/
```

---

### Phase 1: Context Discovery

See `~/.claude/skills/research/references/context-discovery.md` for full 6-tier hierarchy.

**Quick version**: Code-Map -> Semantic Search -> Scoped Grep -> Source -> .agents/ -> External

**Checklist**:
- [ ] Checked code-map index
- [ ] Ran semantic search (if MCP available)
- [ ] Followed signposts to source
- [ ] Reviewed .agents/ patterns WITH verification

---

### Phase 1.5: Prior Formula Discovery

Before creating new formulas, check for existing work:

```bash
# Town-level formulas (Mayor/orchestration work)
grep -l "<goal keywords>" ~/gt/.agents/$RIG/formulas/*.toml 2>/dev/null | head -5
grep -l "<goal keywords>" ~/gt/.agents/_cross-rig/formulas/*.toml 2>/dev/null | head -5

# Crew workspace formulas (implementation work - may have older artifacts)
grep -l "<goal keywords>" ~/gt/$RIG/crew/boden/.agents/formulas/*.toml 2>/dev/null | head -5
grep -l "<goal keywords>" ~/gt/$RIG/crew/boden/.beads/formulas/*.toml 2>/dev/null | head -5

# Existing beads epics
bd list --type=epic | grep -i "<goal keywords>"
```

**Note**: Prior formulas may exist in either location:
- **Town-level** (`~/gt/.agents/<rig>/formulas/`) - Mayor/orchestration formulas
- **Crew workspace** (`~/gt/<rig>/crew/boden/.agents/formulas/`) - Implementation formulas

| Prior Formula Status | Action |
|----------------------|--------|
| EXISTS | Reference it, instantiate instead of creating new |
| SIMILAR | Extend existing formula or create variant |
| NONE | Create new formula |

---

### Phase 2: Research/Analysis

Deepen understanding with targeted exploration:

```
Task(
    subagent_type="Explore",
    model="haiku",
    prompt="Find all code related to: $GOAL"
)
```

**Model**: Use `haiku` for exploration (fast, cheap).

**Identify**: Affected files, existing patterns, related tests, blockers.

---

### Phase 3: Decompose into Features

Each feature should be:
- Completable in a single focused session
- Testable independently
- Following existing patterns

#### Dependency Direction (CRITICAL)

**Rule:** `bd dep add A B` means "A waits for B"

| Command | Meaning |
|---------|---------|
| `bd dep add A B` | A depends on B (B must complete first) |
| `bd dep add child epic` | **WRONG** - Creates deadlock! |

#### Epic-Child Relationship

**Epics and children have NO dependency relationship.** Track children via comment:

```bash
# 1. Create epic
bd create "Epic: OAuth2 Support" --type epic --priority P1
# -> ai-platform-101

# 2. Create children (NO dependency on epic!)
bd create "Add callback endpoint" --type feature --priority P1
# -> ai-platform-102

# 3. Dependencies ONLY between children
bd dep add ai-platform-103 ai-platform-102

# 4. Track children via comment (NOT dependency!)
bd comments add ai-platform-101 "Children: ai-platform-102, ai-platform-103"
```

#### Wave Computation

| Wave | Rule |
|------|------|
| Wave 1 | Issues with NO dependencies |
| Wave N | Issues where ALL deps are in Wave N-1 or earlier |

---

### Phase 3.5: Verify Agent Dependencies

For agent-related work (KAgents, MCP servers):

```bash
grep -A5 "agents:" charts/ai-platform/values.yaml
ls services/mcp-servers/
```

---

### Phase 4: Create Formula Template

Analyze the decomposed features and prepare the formula structure:

1. **Identify Variables**: Extract parameterizable values (service names, paths, etc.)
2. **Define Steps**: Map each feature to a [[steps]] entry with proper dependencies
3. **Compute Waves**: Group steps by dependency order for parallel execution

---

### Phase 5: Output Formula TOML File

**CRITICAL**: Write the formula to `~/gt/.agents/$RIG/formulas/{topic-slug}.formula.toml`

```toml
# Formula: {Goal Name}
# Created: YYYY-MM-DD

# Required top-level fields (NOT in a [formula] table!)
formula = "{topic-slug}"
description = "{Detailed description of what this formula produces}"
version = 2
type = "workflow"  # MUST be: workflow | expansion | aspect

# Optional: Variables for parameterization
# Use {{var_name}} in step descriptions
# Each variable needs its own table with description and optional default
[vars]
[vars.service_name]
description = "Name of the service being created"
default = "default-service"
[vars.base_path]
description = "Base path for service files"
default = "services/"

# Steps become child issues when the formula is poured
[[steps]]
id = "core"
title = "Create core implementation"
description = """
Implement the core {{service_name}} functionality:
- Add main module at {{base_path}}{{service_name}}/core.py
- Include error handling and logging
- Follow existing patterns in the codebase
"""
needs = []  # No dependencies - Wave 1

[[steps]]
id = "config"
title = "Add configuration"
description = """
Add configuration for {{service_name}}:
- Update charts/values.yaml with new config section
- Add environment variable mappings
- Document configuration options
"""
needs = []  # No dependencies - Wave 1 (parallel with core)

[[steps]]
id = "tests"
title = "Implement tests"
description = """
Add comprehensive tests for {{service_name}}:
- Unit tests for core functionality
- Integration tests for API endpoints
- Ensure >80% coverage
"""
needs = ["core"]  # Depends on core - Wave 2

[[steps]]
id = "docs"
title = "Add documentation"
description = """
Document {{service_name}}:
- API reference in docs/
- Update README with usage examples
- Add architecture decision record if needed
"""
needs = ["core", "config"]  # Depends on both - Wave 2
```

**Formula TOML Structure Reference**:

| Field | Required | Description |
|-------|----------|-------------|
| `formula` | Yes | Unique identifier (slug) - TOP LEVEL string |
| `description` | Yes | What the formula creates |
| `version` | Yes | Integer (use `2`) |
| `type` | Yes | Must be: `workflow`, `expansion`, or `aspect` |
| `[vars.name]` | No | Variable table with `description` and optional `default` |
| `[[steps]]` | Yes | Array of step definitions |
| `steps.id` | Yes | Unique step identifier |
| `steps.title` | Yes | Short step title |
| `steps.description` | Yes | Detailed implementation guidance |
| `steps.needs` | Yes | Array of step IDs this depends on (empty = Wave 1) |

**WRONG format (do NOT use):**
- `[formula]` table with `name = "..."` inside
- `version = "1.0.0"` (string)
- `type = "feature"` (invalid type)
- `[[tasks]]` instead of `[[steps]]`
- `depends_on`, `priority`, `wave`, `files` in steps
- `[vars]\nvar_name = "value"` (must use `[vars.var_name]\ndescription = "..."`)

**Wave Computation from `needs`**:
- Wave 1: Steps with `needs = []`
- Wave N: Steps where all `needs` are in Wave N-1 or earlier

---

### Phase 5.5: Cook and Pour Formula (Optional)

After writing the formula file, use the two-step process to create beads:

**Step 1: Cook** - Transform formula to proto (saves to database)
```bash
# Preview what would be created (dry run)
bd cook ~/gt/.agents/$RIG/formulas/{topic-slug}.formula.toml --dry-run

# Cook and persist proto to database
bd cook ~/gt/.agents/$RIG/formulas/{topic-slug}.formula.toml --persist

# With variable substitution
bd cook ~/gt/.agents/$RIG/formulas/{topic-slug}.formula.toml --persist \
  --var service_name=rate-limiter
```

**Step 2: Pour** - Instantiate proto into active mol (persistent beads)
```bash
# Pour the cooked proto (creates actual issues)
bd mol pour {topic-slug}

# Or with variable overrides
bd mol pour {topic-slug} --var service_name=rate-limiter

# Use wisp for ephemeral work (auto-cleaned)
bd mol wisp {topic-slug} --var service_name=rate-limiter
```

**Alternative: Direct beads creation (--immediate flag)**

For one-off work, skip formula files and create beads directly (see Phase 6).

**Note**: `bd cook --persist` saves the proto to the database. Then `bd mol pour`
creates actual beads issues from that proto.

---

### Phase 6: Instantiate Formula (--immediate or Manual)

**Option A: Using `bd mol pour` (Recommended)**

If `--cook` flag was passed or auto-instantiation is desired:

```bash
# Pour creates persistent beads from the formula proto
bd mol pour {topic-slug}
# Returns: Created mol ai-platform-xxx from proto {topic-slug}
```

**Option B: Using `--immediate` Flag**

When `--immediate` is passed, skip formula file creation entirely and create beads directly:

```bash
# Create epic
bd create "Epic: $GOAL" --type epic --priority P1

# Create features from decomposition
bd create "Feature description" --type feature --priority P2

# Set dependencies
bd dep add ai-platform-103 ai-platform-102

# REQUIRED: File annotations for wave parallelization
bd comment <id> "Files affected: src/auth/middleware.py, tests/test_auth.py"

# Track children on epic
bd comments add ai-platform-100 "Children: ai-platform-101, ai-platform-102, ai-platform-103"

# Start epic
bd update ai-platform-100 --status in_progress
```

**Option C: Manual Instantiation**

For formulas that need customization before instantiation:

```bash
# Edit the formula file to adjust variables
vim ~/gt/.agents/$RIG/formulas/{topic-slug}.formula.toml

# Preview with cook (outputs JSON proto)
bd cook ~/gt/.agents/$RIG/formulas/{topic-slug}.formula.toml --mode=runtime

# Then pour to create beads
bd mol pour {topic-slug} --var service_name=custom-name
```

---

### Phase 7: Write Companion Documentation

Create a companion markdown file at `~/gt/.agents/$RIG/formulas/{topic-slug}.md` with:
- Frontmatter with date, goal, epic ID, tags
- Features table with dependencies
- Wave execution order table
- Dependency graph (ASCII)
- Crank handoff section

See `references/templates.md` for full template.

---

### Phase 8: Output Summary

Output structured summary with crank handoff:

```markdown
# Formula Complete: [Goal]

**Formula:** `~/gt/.agents/$RIG/formulas/{goal-slug}.formula.toml`
**Epic:** `ai-platform-xxx`
**Plan:** `~/gt/.agents/$RIG/formulas/{goal-slug}.md`
**Issues:** N features across M waves

## Wave Execution Order
| Wave | Issues | Can Parallel |
|------|--------|--------------|
| 1 | ai-platform-102, ai-platform-106 | Yes |
| 2 | ai-platform-103 | No |

## Ready for Autopilot
```bash
/crank ai-platform-xxx --dry-run
/crank ai-platform-xxx
```
```

---

## Epic Lifecycle Checklist

- [ ] Create epic with `bd create --type epic`
- [ ] Create children (NO dependency on epic!)
- [ ] Add children list to epic via `bd comment`
- [ ] **Start epic**: `bd update <epic> --status in_progress`
- [ ] Implement children via `/implement` or `/implement-wave`
- [ ] **Close epic**: `bd close <epic>` after all children done
- [ ] Run `bd sync && git push` before session end

**Common Gotcha**: Children can't become `ready` until epic is `in_progress`.

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Create children depending on epic | Track via comment on epic |
| Skip file annotations | Add "Files affected: ..." |
| Create 10+ features at once | Group into 3-5 per wave |
| Skip prior formula check | Search ~/gt/.agents/$RIG/formulas/ first |
| Grep source blindly | Use code-map signposts |
| Forget to start epic | `bd update <epic> --status in_progress` |
| Create one-off plans for repeatable patterns | Create a formula template |

---

## Execution Checklist

### Standard Flow (Formula Creation)
- [ ] Detected target rig (Phase 0)
- [ ] Researched codebase context - 6-tier hierarchy (Phase 1)
- [ ] Checked for prior formulas (Phase 1.5)
- [ ] Decomposed into discrete features (Phase 3)
- [ ] Verified agent dependencies if applicable (Phase 3.5)
- [ ] Identified variables and step structure (Phase 4)
- [ ] Wrote .formula.toml with proper structure (Phase 5)
  - [ ] `formula`, `description`, `type`, `version` fields
  - [ ] `[vars]` section for parameterization
  - [ ] `[[steps]]` with id, title, description, needs
- [ ] Ran `bd cook` or instantiated manually (Phase 5.5/6)
- [ ] Wrote companion .md documentation (Phase 7)
- [ ] Output summary with crank handoff (Phase 8)
- [ ] Synced with `bd sync`

### Immediate Flow (--immediate flag)
- [ ] Detected target rig (Phase 0)
- [ ] Researched codebase context (Phase 1)
- [ ] Decomposed into discrete features (Phase 3)
- [ ] Created beads issues directly with `bd create`
- [ ] Set dependencies with `bd dep add`
- [ ] Added file annotations with `bd comment`
- [ ] Added Children comment to epic
- [ ] Started epic with `bd update <epic> --status in_progress`
- [ ] Output summary with crank handoff
- [ ] Synced with `bd sync`

---

## Quick Examples

See `references/examples.md` for detailed walkthroughs including:
- Simple multi-step formulas
- Reusing existing formulas with variables
- Creating formulas from research findings
- Complex dependency graphs with merge points
- Quick formulas (3 steps or less)
- Anti-pattern examples (wrong formats to avoid)

---

## References

### JIT-Loadable Documentation

| Topic | Reference |
|-------|-----------|
| Full templates | `references/templates.md` |
| Detailed examples | `references/examples.md` |
| Phase naming | `.claude/includes/phase-naming.md` |
| Beads workflows | `~/.claude/skills/beads/SKILL.md` |
| Decomposition patterns | `~/.claude/patterns/commands/plan/decomposition.md` |

### Essential Commands

| Command | Purpose |
|---------|---------|
| `bd create "Title" --type epic` | Create epic container |
| `bd create "Feature" --type feature` | Create feature issue |
| `bd cook <formula.toml> --dry-run` | Preview what would be created |
| `bd cook <formula.toml> --persist` | Cook and save proto to database |
| `bd mol pour <proto-id>` | Create beads from proto |
| `bd mol wisp <proto-id>` | Create ephemeral beads from proto |
| `bd dep add A B` | A waits for B |
| `bd comments add <id> "Children: ..."` | Track children on epic |
| `bd update <id> --status in_progress` | Start work |
| `bd show <id>` | View issue details |
| `bd ready` | Show unblocked issues |
| `bd sync` | Sync to git |

### Related Skills

- **beads**: Detailed issue management
- **research**: When goal needs exploration before formulating
- **implement**: When executing a single issue from the formula

---

**Progressive Disclosure**: This skill provides core formulation workflows. For detailed templates see `references/templates.md`, for examples see `references/examples.md`.
