---
name: doc
description: 'This skill should be used when the user asks to "generate documentation", "validate docs", "check doc coverage", "find missing docs", "create code-map", "sync documentation", "update docs", or needs guidance on documentation generation and validation for any repository type. Triggers: doc, documentation, code-map, doc coverage, validate docs.'
---


# Doc Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Generate and validate documentation for any project.

## Execution Steps

Given `$doc [command] [target]`:

### Step 1: Detect Project Type

```bash
# Check for indicators
ls package.json pyproject.toml go.mod Cargo.toml 2>/dev/null

# Check for existing docs
ls -d docs/ doc/ documentation/ 2>/dev/null
```

Classify as:
- **CODING**: Has source code, needs API docs
- **INFORMATIONAL**: Primarily documentation (wiki, knowledge base)
- **OPS**: Infrastructure, deployment, runbooks

### Step 2: Execute Command

**discover** - Find undocumented features:
```bash
# Find public functions without docstrings (Python)
grep -r "^def " --include="*.py" | grep -v '"""' | head -20

# Find exported functions without comments (Go)
grep -r "^func [A-Z]" --include="*.go" | head -20
```

**coverage** - Check documentation coverage:
```bash
# Count documented vs undocumented
TOTAL=$(grep -r "^def \|^func \|^class " --include="*.py" --include="*.go" | wc -l)
DOCUMENTED=$(grep -r '"""' --include="*.py" | wc -l)
echo "Coverage: $DOCUMENTED / $TOTAL"
```

**gen [feature]** - Generate documentation:
1. Read the code for the feature
2. Understand what it does
3. Generate appropriate documentation
4. Write to docs/ directory

**all** - Update all documentation:
1. Run discover to find gaps
2. Generate docs for each undocumented feature
3. Validate existing docs are current

### Step 3: Generate Documentation

When generating docs, include:

**For Functions/Methods:**
```markdown
## function_name

**Purpose:** What it does

**Parameters:**
- `param1` (type): Description
- `param2` (type): Description

**Returns:** What it returns

**Example:**
```python
result = function_name(arg1, arg2)
```

**Notes:** Any important caveats
```

**For Classes:**
```markdown
## ClassName

**Purpose:** What this class represents

**Attributes:**
- `attr1`: Description
- `attr2`: Description

**Methods:**
- `method1()`: What it does
- `method2()`: What it does

**Usage:**
```python
obj = ClassName()
obj.method1()
```
```

### Step 4: Create Code-Map (if requested)

**Write to:** `docs/code-map/`

```markdown
# Code Map: <Project>

## Overview
<High-level architecture>

## Directory Structure
```
src/
├── module1/     # Purpose
├── module2/     # Purpose
└── utils/       # Shared utilities
```

## Key Components

### Module 1
- **Purpose:** What it does
- **Entry point:** `main.py`
- **Key files:** `handler.py`, `models.py`

### Module 2
...

## Data Flow
<How data moves through the system>

## Dependencies
<External dependencies and why>
```

### Step 5: Validate Documentation

Check for:
- Out-of-date docs (code changed, docs didn't)
- Missing sections (no examples, no parameters)
- Broken links
- Inconsistent formatting

### Step 6: Write Report

**Write to:** `.agents/doc/YYYY-MM-DD-<target>.md`

```markdown
# Documentation Report: <Target>

**Date:** YYYY-MM-DD
**Project Type:** <CODING/INFORMATIONAL/OPS>

## Coverage
- Total documentable items: <count>
- Documented: <count>
- Coverage: <percentage>%

## Generated
- <list of docs generated>

## Gaps Found
- <undocumented item 1>
- <undocumented item 2>

## Validation Issues
- <issue 1>
- <issue 2>

## Next Steps
- [ ] Document remaining gaps
- [ ] Fix validation issues
```

### Step 7: Report to User

Tell the user:
1. Documentation coverage percentage
2. Docs generated/updated
3. Gaps remaining
4. Location of report

## Key Rules

- **Detect project type first** - approach varies
- **Generate meaningful docs** - not just stubs
- **Include examples** - always show usage
- **Validate existing** - docs can go stale
- **Write the report** - track coverage over time

## Commands Summary

| Command | Action |
|---------|--------|
| `discover` | Find undocumented features |
| `coverage` | Check documentation coverage |
| `gen [feature]` | Generate docs for specific feature |
| `all` | Update all documentation |
| `validate` | Check docs match code |

## Examples

### Generating API Documentation

**User says:** `$doc gen authentication`

**What happens:**
1. Agent detects project type by checking for `package.json` and finding Node.js project
2. Agent searches codebase for authentication-related functions using grep
3. Agent reads authentication module files to understand implementation
4. Agent generates documentation with purpose, parameters, returns, and usage examples
5. Agent writes to `docs/api/authentication.md` with code samples
6. Agent validates generated docs match actual function signatures

**Result:** Complete API documentation created for authentication module with working code examples.

### Checking Documentation Coverage

**User says:** `$doc coverage`

**What happens:**
1. Agent detects Python project from `pyproject.toml`
2. Agent counts total functions/classes with `grep -r "^def \|^class "`
3. Agent counts documented items by searching for docstrings (`"""`)
4. Agent calculates coverage: 45/67 items = 67% coverage
5. Agent writes report to `.agents/doc/2026-02-13-coverage.md`
6. Agent lists 22 undocumented functions as gaps

**Result:** Documentation coverage report shows 67% coverage with specific list of 22 functions needing docs.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Coverage calculation inaccurate | Grep pattern doesn't match all code styles | Adjust pattern for project conventions. For Python, check for `async def` and class methods. For Go, check both `func` and `type` definitions. |
| Generated docs lack examples | Missing context about typical usage | Read existing tests to find usage patterns. Check README for code samples. Ask user for typical use case if unclear. |
| Discover command finds too many items | Low existing documentation coverage | Prioritize by running `discover` on specific subdirectories. Focus on public API first, internal utilities later. Use `--limit` to process in batches. |
| Validation shows docs out of sync | Code changed after docs written | Re-run `gen` command for affected features. Consider adding git hook to flag doc updates needed when code changes. |

## Reference Documents

- [references/generation-templates.md](references/generation-templates.md)
- [references/project-types.md](references/project-types.md)
- [references/validation-rules.md](references/validation-rules.md)

---

## References

### generation-templates.md

# Documentation Generation Templates

## CODING: Code-Map Template

**CRITICAL**: Load `code-map-standard` skill before generating.

```markdown
---
title: "[Feature Name]"
sources: [path/to/main.py]
last_updated: YYYY-MM-DD
---

# [Feature Name]

## Current Status

[One-liner with date]

## Overview

[2-3 sentences]

## State Machine

[ASCII diagram if applicable]

## Inputs/Outputs

| Type | Name | Description |
|------|------|-------------|

## Data Flow

[ASCII diagram]

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|

## Code Signposts

| Component | Location | Purpose |
|-----------|----------|---------|

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|

## Prometheus Metrics

| Metric | Type | Labels | PromQL Example |
|--------|------|--------|----------------|

## Error Handling

| Error | Cause | Resolution |
|-------|-------|------------|

## Unit Tests

| Test File | Coverage |
|-----------|----------|

## Integration Tests

| Test | What It Validates |
|------|-------------------|

## Example Usage

### curl
### SDK

## Related Features

## Known Limitations

## Learnings

### What Worked
### What We'd Change
```

---

## INFORMATIONAL: Corpus Section Template

```markdown
---
title: "Document Title"
summary: "One-line summary for search"
tags: [tag1, tag2]
tokens: 1500
last_updated: YYYY-MM-DD
---

# Title

## Overview

[Introduction paragraph]

## Key Concepts

### Concept 1
### Concept 2

## Practical Application

## Related Topics

- [Link 1](../path/to/doc.md)
- [Link 2](../path/to/doc.md)

## References

- External sources
```

---

## OPS: Helm Chart Template

```markdown
# [Chart Name]

## Overview

[Description from Chart.yaml]

## Quick Start

```bash
helm install [release] ./charts/[name]
```

## Values Reference

| Key | Type | Default | Description |
|-----|------|---------|-------------|

## Dependencies

| Chart | Version | Condition |
|-------|---------|-----------|

## Common Overrides

### Development
### Staging
### Production

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
```

---

## Stub Template (--create mode)

For undocumented features:

```markdown
---
title: "[Feature Name]"
status: STUB
created: YYYY-MM-DD
sources: [detected source files]
---

# [Feature Name]

> AUTO-GENERATED STUB - Replace with actual content

## Current Status

[Discovered but not documented]

## Overview

[Brief description of this feature]

## Sources

- `path/to/source.py`

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
```

---

## Section Markers

Use markers to control auto-generation behavior:

```markdown
<!-- HUMAN-MAINTAINED: Do not auto-generate -->
[This section is preserved during updates]

<!-- AUTO-GENERATED: Safe to replace -->
[This section is regenerated from source]
```

**Merge Strategy**:
1. HUMAN-MAINTAINED sections: Always preserve
2. AUTO-GENERATED sections: Replace with fresh data
3. Frontmatter: Merge (add missing, update tokens/dates)

### project-types.md

# Project Type Detection

Score-based classification into CODING, INFORMATIONAL, or OPS.

## CODING Signals

| Signal | Weight | Detection |
|--------|--------|-----------|
| `services/` directory | +3 | `[[ -d services ]]` |
| `src/` directory | +2 | `[[ -d src ]]` |
| `pyproject.toml` or `package.json` | +2 | Config file exists |
| `docs/code-map/` directory | +3 | Code-map docs exist |
| >50 Python/TypeScript files | +2 | File count |
| FastAPI/Express routes | +2 | `@app.get`, `router.` patterns |

**Threshold**: Score >= 5 = Likely CODING repo

---

## INFORMATIONAL Signals

| Signal | Weight | Detection |
|--------|--------|-----------|
| `docs/corpus/` directory | +3 | Knowledge corpus |
| `docs/standards/` directory | +2 | Standards docs |
| >100 markdown files | +3 | High doc count |
| No `services/` or `src/` | +2 | Not a code repo |
| Diataxis structure | +2 | `tutorials/`, `how-to/`, `reference/`, `explanation/` |

**Threshold**: Score >= 5 = Likely INFORMATIONAL repo

---

## OPS Signals

| Signal | Weight | Detection |
|--------|--------|-----------|
| `charts/` directory | +3 | Helm charts |
| `apps/` or `applications/` | +2 | ArgoCD apps |
| >5 `values.yaml` files | +3 | Multi-environment Helm |
| `config.env` files | +2 | Config rendering |
| ArgoCD manifests | +2 | `Application` kind |

**Threshold**: Score >= 5 = Likely OPS repo

---

## Tie-Breaking

When scores are equal: **CODING > OPS > INFORMATIONAL**

Rationale: Code repos need more precise docs, ops is next most critical.

---

## Type-Specific Behaviors

| Type | `$doc all` | `$doc discover` | `$doc coverage` |
|------|------------|-----------------|-----------------|
| CODING | Generate code-maps | Find services, endpoints | Entity coverage |
| INFORMATIONAL | Validate all docs | Find corpus sections | Link validation |
| OPS | Generate Helm docs | Find charts, configs | Values coverage |

### validation-rules.md

# Documentation Validation Rules

## Coverage Metrics by Type

| Type | Key Metric | Target | How Measured |
|------|-----------|--------|--------------|
| CODING | Entity Coverage | >= 90% | Documented services / total services |
| CODING | Signpost Accuracy | 100% | Referenced functions exist |
| INFORMATIONAL | Frontmatter Valid | >= 95% | Required fields present |
| INFORMATIONAL | Links Valid | 100% | All internal links resolve |
| OPS | Values.yaml Coverage | >= 80% | Documented keys / total keys |
| OPS | Golden Completeness | 100% | Required sections present |

---

## INFORMATIONAL Validation

Use Python validator for fast, exhaustive checking:

```bash
python3 ~/.claude/scripts/doc-validate.py docs/
```

### Checks Performed

1. **Broken Links** - ALL internal .md links resolved
2. **Orphaned Docs** - Files not referenced from any index
3. **Index Completeness** - READMEs reference all subdirectories
4. **Hardcoded Paths** - Absolute paths like /Users/, /home/

### Why Python, Not Bash?

- Bash loops are O(n*m) and timeout on large repos
- Python processes 350+ files in <5 seconds
- Regex extraction is cleaner and more reliable

### Output Format

```
CRITICAL: Broken Links (81)
   file.md:42 -> missing.md (not found)

MEDIUM: Orphaned Documents (13)
   path/to/orphan.md

LOW: Hardcoded Paths (2)
   file.md:156 -> /Users/...

SUMMARY: 96 issues (81 critical, 13 medium, 2 low)
```

---

## CODING Validation

### Required Sections (16)

From `code-map-standard` skill:

1. Current Status (one-liner with date)
2. Overview (2-3 sentences)
3. State Machine (ASCII diagram if applicable)
4. Inputs/Outputs (table)
5. Data Flow (ASCII diagram)
6. API Endpoints (table with curl examples)
7. Code Signposts (NO line numbers)
8. Configuration (table)
9. Prometheus Metrics (table + PromQL examples)
10. Error Handling (table)
11. Unit Tests (table)
12. Integration Tests (separate from unit)
13. Example Usage (curl + SDK)
14. Related Features (cross-links)
15. Known Limitations
16. Learnings (What Worked + What We'd Change)

### Signpost Rules

- **NO line numbers** - Functions/classes only
- References must exist in source files
- Use semantic names: `authenticate()`, `UserService`

---

## OPS Validation

### Required Sections

1. Overview with Chart.yaml description
2. Quick Start with install command
3. Values Reference table
4. Dependencies table
5. Environment overrides (dev/staging/prod)
6. Troubleshooting table

### Values.yaml Coverage

Every key in values.yaml should have:
- Description comment or doc reference
- Type specification
- Default value explanation

---

## Coverage Report Format

```
===================================================================
              DOCUMENTATION COVERAGE REPORT
===================================================================
Repository: [REPO_NAME]
Type: [CODING|INFORMATIONAL|OPS]
Generated: [date]

SUMMARY
-------------------------------------------------------------------
Total Features: 25
Documented: 22 (88%)
Missing: 3
Orphaned: 1

MISSING DOCUMENTATION
-------------------------------------------------------------------
| Feature | Priority | Source Files |
|---------|----------|--------------|
| auth-service | P1 | services/auth/*.py |

ORPHANED DOCUMENTATION
-------------------------------------------------------------------
| Document | Last Updated | Action |
|----------|--------------|--------|
| legacy-api.md | 2023-06-15 | Remove |

===================================================================
```

---

## --create-issues Flag

Auto-create tracking issues for gaps:

```bash
# Prefer beads
bd create --title "docs: create code-map for $FEATURE" \
          --type task --priority P1

# Fallback to GitHub
gh issue create --title "docs: create code-map for $FEATURE" \
                --label documentation
```

---

## Semantic Validation (CODING repos)

**Structure vs Semantic:** Structural validation checks formatting. Semantic validation checks if claims are TRUE.

### Semantic Metrics

| Check | How | Target |
|-------|-----|--------|
| Status Accuracy | Compare "Status: X" to deployment state | 100% |
| Claim Verification | Cross-ref with ground truth file | 100% |
| Validation Freshness | Status includes date | < 30 days |

### Ground Truth Pattern

Establish ONE authoritative file per domain. Other docs MUST reference, not duplicate.

| Domain | Ground Truth | Pattern |
|--------|--------------|---------|
| Agents | `docs/agents/catalog.md` | Reference via link |
| Images | `charts/*/IMAGE-LIST.md` | Reference via link |
| Config | `values.yaml` | Generate docs from source |

### Status Validation

Valid status formats:

```markdown
## Current Status: ✅ RUNNING
Validated: 2026-01-04 against ocppoc cluster

## Current Status: ❌ FAILED
Status: Accepted=False (CRD exists but not running)
Validated: 2026-01-04 against ocppoc cluster

## Current Status: 📝 PLANNED
Not yet deployed - template only
```

### Semantic Validation Commands

```bash
# Check status claims against cluster (manual)
oc get pods -n ai-platform | grep <service>
oc get agents.kagent.dev -n ai-platform

# Cross-reference with ground truth
diff <(grep "Status:" docs/code-map/services/*.md) <(cat docs/agents/catalog.md)
```

### --verify-claims Flag

When running `$doc coverage --verify-claims`:

1. Extract all "Status: X" claims from docs
2. Query deployment state (oc get pods, oc get agents)
3. Report mismatches as CRITICAL
4. Flag stale validation dates (>30 days) as WARNING

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Sample 20 files, declare "healthy" | Scan ALL files |
| Say "healthy" with broken links | Report exact issue counts |
| Skip validation for "organized" repos | Validate regardless |
| Use bash loops on large repos | Use Python validator |
| Claim "deployed" without verification | Validate against cluster first |
| Duplicate ground truth data | Reference authoritative file |
| Omit validation dates | Include "Validated: DATE against SOURCE" |


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
check "SKILL.md has name: doc" "grep -q '^name: doc' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions documentation generation" "grep -qi 'generate.*doc\|documentation' '$SKILL_DIR/SKILL.md'"
check "SKILL.md mentions code-map" "grep -qi 'code-map\|code map' '$SKILL_DIR/SKILL.md'"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```


