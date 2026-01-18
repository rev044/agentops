# Cooking Formulas

The `bd cook` command transforms `.formula.toml` files into protos.
Cooking is the bridge between template definition and workflow instantiation.

## Basic Usage

```bash
# Preview what would be created
bd cook workflow.formula.toml --dry-run

# Cook to stdout (ephemeral proto)
bd cook workflow.formula.toml

# Persist proto to database
bd cook workflow.formula.toml --persist
```

## Cooking Modes

### Compile-Time Mode (Default)

Keeps `{{variable}}` placeholders intact:

```bash
bd cook workflow.formula.toml
# or explicitly
bd cook workflow.formula.toml --mode=compile
```

**Use for:**
- Modeling and estimation
- Contractor handoff
- Planning and review
- Template inspection

Output shows the template structure with variables unsubstituted.

### Runtime Mode

Substitutes all variables:

```bash
bd cook workflow.formula.toml --mode=runtime --var key=value
# or (--var implies runtime mode)
bd cook workflow.formula.toml --var component=auth --var version=2.0
```

**Use for:**
- Final validation before pour
- Seeing exact output
- Verifying variable substitution

All variables must have values (via `--var` or formula defaults).

## Command Options

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview what would be created |
| `--mode` | `compile` (keep placeholders) or `runtime` (substitute vars) |
| `--var key=value` | Variable substitution (enables runtime mode) |
| `--persist` | Write proto to database |
| `--force` | Replace existing proto (requires `--persist`) |
| `--prefix` | Prefix to prepend to proto ID |
| `--search-path` | Additional paths for formula inheritance |

## Workflow: Cook -> Pour

### Two-Step Process

```bash
# Step 1: Cook (preview first!)
bd cook workflow.formula.toml --dry-run
bd cook workflow.formula.toml --persist

# Step 2: Pour to create issues
bd mol pour workflow-name
bd mol pour workflow-name --var component=auth
```

### Direct Pour (Inline Cooking)

For most workflows, `bd mol pour` cooks inline:

```bash
# Cooks and pours in one step
bd mol pour workflow-name --var component=auth
```

**When to persist explicitly:**
- Reusing same proto multiple times
- Sharing proto across team
- Storing proto for later use

## Variable Handling

### In Formula

```toml
[vars]
service_name = "default-service"
base_path = "services/"
```

### During Cook

```bash
# Use defaults from formula
bd cook workflow.formula.toml

# Override defaults
bd cook workflow.formula.toml --var service_name=auth
```

### During Pour

```bash
# Override at pour time
bd mol pour workflow-name --var service_name=auth
```

## Search Paths

When cooking, bd searches for formulas in order:

1. `.beads/formulas/` (project)
2. `~/.beads/formulas/` (user)
3. `$GT_ROOT/.beads/formulas/` (orchestrator, if GT_ROOT set)

Override with `--search-path`:

```bash
bd cook my-formula.toml --search-path /path/to/custom/formulas
```

## Formula Inheritance

Formulas can extend other formulas:

```toml
# child.formula.toml
formula = "child-workflow"
extends = "base-workflow"
description = "Extended workflow"
version = 2
type = "workflow"

# Additional or overriding steps...
[[steps]]
id = "extra"
title = "Extra step"
description = "Added by child"
needs = []
```

Cook resolves inheritance automatically.

## Output Formats

### Default (JSON to stdout)

```bash
bd cook workflow.formula.toml
# Outputs JSON proto representation
```

### Dry Run (Human-readable)

```bash
bd cook workflow.formula.toml --dry-run
# Shows what would be created
```

### Persisted (Database)

```bash
bd cook workflow.formula.toml --persist
# Creates proto bead in database with:
# - ID matching formula name
# - "template" label
# - Child issues for each step
# - Dependencies from needs relationships
```

## Examples

### Basic Workflow

```bash
# Preview
bd cook rate-limiting.formula.toml --dry-run

# Output:
# Proto: rate-limiting
# Steps:
#   - middleware (Wave 1)
#   - config (Wave 1)
#   - tests (Wave 2, needs: middleware)

# Cook and persist
bd cook rate-limiting.formula.toml --persist

# Use later
bd mol pour rate-limiting --var requests_per_minute=200
```

### With Variables

```bash
# Compile-time: see placeholders
bd cook feature.formula.toml
# Shows: "Create {{service_name}} module"

# Runtime: see substituted values
bd cook feature.formula.toml --var service_name=auth
# Shows: "Create auth module"
```

### Replace Existing Proto

```bash
# First cook
bd cook workflow.formula.toml --persist

# Updated formula, need to replace
bd cook workflow.formula.toml --persist --force
```

## Troubleshooting

**"Formula not found"**
- Check formula exists in search paths
- Use full path: `bd cook path/to/formula.toml`

**"Variable not found"**
- Check formula has default or provide `--var`
- Runtime mode requires all variables

**"Proto already exists"**
- Use `--force` to replace
- Or use different `--prefix`

**"Invalid formula"**
- Check TOML syntax
- Ensure required fields: `formula`, `description`, `version`, `type`
- Use `[[steps]]` not `[[tasks]]`

## Related

- **Formula syntax**: See `formula-toml.md`
- **Molecule lifecycle**: See `mol-lifecycle.md`
- **Wisp patterns**: See `wisp-patterns.md`
