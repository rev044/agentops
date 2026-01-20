---
name: doc
description: >
  This skill should be used when the user asks to "generate documentation",
  "validate docs", "check doc coverage", "find missing docs", "create code-map",
  "sync documentation", "update docs", or needs guidance on documentation
  generation and validation for any repository type.
  Triggers: doc, documentation, code-map, doc coverage, validate docs.
version: 2.0.0
context: fork
author: "AI Platform Team"
license: "MIT"
allowed-tools: "Read,Write,Edit,Glob,Grep,Bash,Task,mcp__smart-connections-work__lookup"
---

# Doc Skill

Universal documentation generation and validation for any project structure.

## Overview

Auto-detects project type (CODING, INFORMATIONAL, OPS), discovers documentable features, generates documentation, and validates coverage.

## Commands

| Command | Action |
|---------|--------|
| `discover` | Find undocumented features |
| `discover --create` | Generate stubs for undocumented |
| `gen [feature]` | Generate/update specific doc |
| `all` | Update all docs (CODING/OPS) or validate (INFORMATIONAL) |
| `sync` | Pull from canonical source |
| `coverage` | Validate docs match code |
| `coverage --create-issues` | Create tracking issues for gaps |

---

## Phase 0: Project Detection (Always First)

Run detection script to establish context:

```bash
./scripts/detect-project.sh
```

Returns JSON with type, confidence, and doc directories.

See `references/project-types.md` for signal weights and behaviors.

---

## Subcommand: discover

Find documentable features based on project type.

```bash
TYPE=$(./scripts/detect-project.sh | jq -r .type)
./scripts/discover-features.sh "$TYPE"
```

**CODING**: Finds services with endpoints, metrics, config vars (score >= 3)

**INFORMATIONAL**: Finds corpus sections, validates structure

**OPS**: Finds Helm charts, config files, runbooks

---

## Subcommand: gen

Generate or update documentation for a specific feature.

### Execution

1. Detect project type
2. Find source files (frontmatter or naming convention)
3. Extract content based on type:
   - CODING: endpoints, config, signposts, metrics
   - INFORMATIONAL: frontmatter, links, tokens
   - OPS: values.yaml, dependencies
4. Generate using templates from `references/generation-templates.md`
5. Validate output

**CODING repos**: Load `code-map-standard` skill first.

---

## Subcommand: all

Update all documentation or validate based on type.

| Type | Behavior |
|------|----------|
| CODING | Generate/update code-map docs from sources |
| INFORMATIONAL | Validate all docs, report issues |
| OPS | Generate/update Helm and config docs |

### Execution

1. Run Phase 0 (detection)
2. Run `discover` to find all features
3. For each feature with sources: run `gen`
4. Validate using type-specific rules
5. Report summary

---

## Subcommand: coverage

Validate documentation covers all actual features.

```bash
./scripts/detect-project.sh  # Get type
# Then validate per type
```

### Metrics (from `references/validation-rules.md`)

| Type | Key Metric | Target |
|------|-----------|--------|
| CODING | Entity Coverage | >= 90% |
| INFORMATIONAL | Links Valid | 100% |
| OPS | Values Coverage | >= 80% |

### Output

```
SUMMARY: 25 features, 22 documented (88%)
MISSING: 3 features need docs
ORPHANED: 1 doc has no source
```

---

## Subcommand: sync

Pull documentation from canonical source for multi-repo setups.

### Configuration

Check for sync config in:
1. `CLAUDE.md` - canonical source references
2. `.doc-sync.yaml` - explicit configuration
3. Document headers - `<!-- Canonical source: ... -->`

### Format

```yaml
# .doc-sync.yaml
canonical_repo: "../gitops"
sync_items:
  - src: "docs/standards/style-guide.md"
    dst: "docs/standards/style-guide.md"
```

---

## Integration Points

| Skill | Purpose |
|-------|---------|
| `code-map-standard` | **Required** for CODING doc generation |
| `beads` | For `--create-issues` tracking |
| `research` | Prior art discovery |

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "No features found" | Check source patterns, verify type detection |
| "Coverage 0%" | Check naming transformations, verify doc dir |
| Validation timeout | Use Python validator, not bash loops |

---

## References

- **Project Types**: `references/project-types.md`
- **Templates**: `references/generation-templates.md`
- **Validation**: `references/validation-rules.md`
- **Detection Script**: `scripts/detect-project.sh`
- **Discovery Script**: `scripts/discover-features.sh`
