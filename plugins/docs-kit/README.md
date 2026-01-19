# Docs Kit

Documentation generation and scaffolding. 5 skills for creating and maintaining docs.

## Install

```bash
/plugin install docs-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/doc` | `/doc` | Generate documentation from code |
| `/doc-creator` | `/doc-creator` | Create corpus or standards documentation |
| `/code-map-standard` | `/code-map-standard` | Generate code-map feature docs |
| `/oss-docs` | `/oss-docs` | Scaffold OSS documentation |
| `/golden-init` | `/golden-init` | Initialize with golden template |

## When to Use Which

| Scenario | Skill |
|----------|-------|
| Generate API docs | `/doc` |
| Create corpus/training docs | `/doc-creator` |
| Document features (code-map) | `/code-map-standard` |
| Prepare for open source | `/oss-docs` |
| Start new project | `/golden-init` |

## Examples

### Generate documentation

```bash
/doc
# Detects project type, generates appropriate docs
# Output: docs/code-map/, README updates
```

### Prepare for open source release

```bash
/oss-docs
# Creates: README.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md,
# LICENSE, SECURITY.md, AGENTS.md
```

### Initialize new repo

```bash
/golden-init
# Sets up: .github/, CLAUDE.md, .beads/, standard structure
```

## Philosophy

- **Documentation as code** - generated, not handwritten
- **OSS-ready from day one** - standard files in place
- **Golden templates for consistency** - same structure everywhere

## Related Kits

- **vibe-kit** - `/vibe-docs` validates the docs you generate
- **core-kit** - Research informs documentation
