# YAML/Helm Standards Catalog - Vibe Canonical Reference

**Version:** 1.0.0
**Last Updated:** 2026-01-21
**Purpose:** Canonical YAML/Helm standards for vibe skill validation

---

## Table of Contents

1. [yamllint Configuration](#yamllint-configuration)
2. [Formatting Rules](#formatting-rules)
3. [Helm Chart Conventions](#helm-chart-conventions)
4. [Kustomize Patterns](#kustomize-patterns)
5. [Compliance Assessment](#compliance-assessment)

---

## yamllint Configuration

### Full Configuration

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
  comments:
    min-spaces-from-content: 1
  document-start: disable
  empty-lines:
    max: 2
```

### Usage

```bash
yamllint .
yamllint apps/
yamllint -f parsable .
```

---

## Formatting Rules

### Quoting Strings

```yaml
# Quote strings that look like other types
enabled: "true"      # String, not boolean
port: "8080"         # String, not integer

# No quotes for actual typed values
enabled: true        # Boolean
port: 8080           # Integer
```

### Multi-line Strings

```yaml
# Literal block scalar (preserves newlines)
script: |
  #!/bin/bash
  set -euo pipefail
  echo "Hello"

# Folded block scalar (folds newlines to spaces)
description: >
  This is a long description that will be
  folded into a single line with spaces.
```

---

## Helm Chart Conventions

### Chart Structure

```text
charts/<chart-name>/
├── Chart.yaml
├── values.yaml
├── values.schema.json
├── templates/
│   ├── _helpers.tpl
│   ├── deployment.yaml
│   └── service.yaml
└── charts/
```

### Validation Commands

```bash
# Lint chart
helm lint charts/<chart-name>/

# Template with values (dry-run)
helm template <release> charts/<chart-name>/ -f values.yaml

# Validate rendered output
helm template <release> charts/<chart-name>/ | kubectl apply --dry-run=client -f -
```

---

## Kustomize Patterns

### Overlay Structure

```text
apps/<app>/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   └── service.yaml
└── overlays/
    ├── dev/
    ├── staging/
    └── prod/
```

### kustomization.yaml Template

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - deployment.yaml
  - service.yaml

patches:
  - path: ./patches/replicas.yaml
    target:
      kind: Deployment
      name: my-app
```

---

## Compliance Assessment

**Use letter grades + evidence, NOT numeric scores.**

### Grading Scale

| Grade | Criteria |
|-------|----------|
| A+ | 0 yamllint errors, 0 helm lint errors, documented, 0 secrets |
| A | <3 yamllint warnings, <3 helm lint warnings, documented |
| A- | <10 warnings, partial docs |
| B+ | <20 warnings |
| B | <40 warnings, templates render |
| C | Significant issues |

### Validation Commands

```bash
# Lint YAML
yamllint .

# Check for tabs
grep -rP '\t' --include='*.yaml' . | wc -l

# Helm lint
for chart in charts/*/Chart.yaml; do
  helm lint "$(dirname "$chart")"
done

# Check for hardcoded secrets
grep -r "password:\|secret:\|token:" --include='*.yaml' apps/
```

---

## Additional Resources

- [YAML Spec](https://yaml.org/spec/)
- [Helm Documentation](https://helm.sh/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [yamllint Documentation](https://yamllint.readthedocs.io/)
