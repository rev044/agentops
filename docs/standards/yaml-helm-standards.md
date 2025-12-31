# YAML/Helm Standards

> **Purpose**: YAML and Helm chart standards for this repository.

## Quick Reference

| Aspect | Standard |
|--------|----------|
| **Indentation** | 2 spaces |
| **Linter** | yamllint |
| **Helm** | helm lint, helm template |

## yamllint Configuration

Create `.yamllint.yml`:
```yaml
extends: default
rules:
  line-length:
    max: 120
  indentation:
    spaces: 2
  truthy:
    check-keys: false
```

## Helm Validation

```bash
# Lint chart
helm lint ./charts/myapp

# Template and validate
helm template myapp ./charts/myapp | kubectl apply --dry-run=client -f -
```

## Key Conventions

1. Use 2-space indentation
2. Quote strings that look like numbers or booleans
3. Use `|` for multi-line strings
4. Document values.yaml with comments
5. Use `{{- include ... }}` for templates

## values.yaml Best Practices

```yaml
# ✅ GOOD - Documented, typed, with defaults
# replicaCount is the number of pod replicas
# @type: integer
replicaCount: 1

# image configuration
image:
  # repository is the container image repository
  repository: nginx
  # tag is the image tag (immutable tags recommended)
  tag: "1.25.0"  # quoted to prevent YAML interpretation

# ❌ BAD - Undocumented, unquoted
replicaCount: 1
image:
  repository: nginx
  tag: 1.25.0  # could be interpreted as float
```

## Multi-line Strings

```yaml
# ✅ GOOD - Literal block scalar preserves newlines
description: |
  This is a multi-line description.
  Each line is preserved exactly as written.

# ✅ GOOD - Folded block scalar for flowing text
summary: >
  This is a long description that will be
  folded into a single line with spaces.

# ❌ BAD - Escaped newlines are hard to read
description: "Line 1\nLine 2\nLine 3"
```

## Helm Template Patterns

```yaml
# Use include for reusable templates
labels:
  {{- include "myapp.labels" . | nindent 4 }}

# Use with for scoped context
{{- with .Values.nodeSelector }}
nodeSelector:
  {{- toYaml . | nindent 2 }}
{{- end }}

# Use range for lists
{{- range .Values.extraVolumes }}
- name: {{ .name }}
  {{- if .configMap }}
  configMap:
    name: {{ .configMap.name }}
  {{- end }}
{{- end }}
```

## TODO

- [ ] Add project-specific patterns
- [ ] Document Kustomize conventions if used
- [ ] Add CI validation steps
