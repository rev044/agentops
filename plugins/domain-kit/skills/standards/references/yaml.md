# YAML/Helm Standards - Tier 1 Quick Reference

<!-- Tier 1: Generic standards (~4KB), always loaded -->
<!-- Tier 2: Deep standards in vibe/references/yaml-standards.md (~15KB), loaded on --deep -->
<!-- Last synced: 2026-01-21 -->

> **Purpose:** Quick reference for YAML/Helm standards. For comprehensive patterns, load Tier 2.

---

## Quick Reference

| Aspect | Standard | Validation |
|--------|----------|------------|
| **Indentation** | 2 spaces | yamllint |
| **Line length** | 120 chars max | yamllint |
| **Linter** | yamllint | `.yamllint.yml` at root |
| **Helm** | helm lint | Chart directory |
| **Kustomize** | kustomize build | Overlay directory |

---

## yamllint Configuration

```yaml
# .yamllint.yml
extends: default
rules:
  line-length:
    max: 120
    allow-non-breakable-inline-mappings: true
  indentation:
    spaces: 2
    indent-sequences: consistent
  truthy:
    check-keys: false
  document-start: disable
```

---

## Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| `mapping values not allowed` | Missing space after colon | `key: value` |
| `found duplicate key` | Repeated key | Remove duplicate |
| `could not find expected ':'` | Unquoted special chars | Quote the value |
| `helm: values don't align` | Wrong indentation | Use 2 spaces |
| Tab characters | Using tabs | Convert to spaces |
| `nil pointer evaluating` | Missing Helm value | Use `default` or `required` |

---

## Anti-Patterns

| Name | Pattern | Instead |
|------|---------|---------|
| Tab Indentation | Tabs | 2 spaces |
| Unquoted Versions | `version: 1.0` | `version: "1.0"` |
| Inline JSON | `data: {"key": "val"}` | Multi-line YAML |
| No Comments | Undocumented values.yaml | Section comments |
| Hardcoded Secrets | `password: hunter2` | External secret ref |
| Anchor Abuse | Complex `<<: *anchor` | Duplicate with comments |

---

## Quoting Rules

```yaml
# Quote strings that look like other types
enabled: "true"      # String, not boolean
port: "8080"         # String, not integer
version: "1.0"       # String, not float

# No quotes for actual typed values
enabled: true        # Boolean
port: 8080           # Integer
replicas: 3          # Integer
```

---

## Multi-line Strings

```yaml
# Literal block (preserves newlines)
script: |
  #!/bin/bash
  set -euo pipefail
  echo "Hello"

# Folded block (folds to single line)
description: >
  This is a long description
  folded into a single line.
```

---

## Helm vs Kustomize

| Use Helm | Use Kustomize |
|----------|---------------|
| Complex templating | Simple overlays |
| Chart distribution | Internal deployments |
| Computed values | Static patches |
| Repository ecosystem | Plain manifests |

---

## Summary Checklist

| Category | Requirement |
|----------|-------------|
| **Indentation** | 2 spaces, no tabs |
| **Linting** | yamllint passes |
| **Helm** | helm lint passes |
| **Versions** | Quoted: `"1.0"` |
| **Secrets** | Never hardcoded |
| **values.yaml** | Documented with comments |
| **Multi-line** | Use `\|` or `>` blocks |

---

## Talos Prescan Checks

| Check | Pattern | Rationale |
|-------|---------|-----------|
| PRE-002 | TODO/FIXME markers | Track technical debt |
| PRE-016 | Unquoted versions | Float parsing risk |
| PRE-017 | Hardcoded secrets | Security violation |

---

## JIT Loading

**Tier 2 (Deep Standards):** For comprehensive patterns including:
- Full yamllint configuration
- Helm chart structure and templates
- Kustomize overlay patterns
- Validation workflows
- Compliance assessment details

Load: `vibe/references/yaml-standards.md`
