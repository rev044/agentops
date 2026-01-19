# Code-Map Full Template

## Document Structure

```markdown
# Feature: [FEATURE_NAME]

> One-line description of what this feature does.

<!-- HUMAN-MAINTAINED START -->

## Overview

2-3 sentences explaining the feature's purpose.

**Current Status:** COMPLETE - [brief description of current state]

**K8s Parallel:** [How this maps to Kubernetes/vCenter/similar concepts]

## Inputs → Outputs

| Input | Type | Source | Description |
|-------|------|--------|-------------|
| `inputName` | `Type` | API/CLI/Event | What this input represents |

| Output | Type | Destination | Description |
|--------|------|-------------|-------------|
| `outputName` | `Type` | DB/File/API | What this output represents |

## Data Flow

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Trigger       │────▶│   Handler        │────▶│   Output        │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## State Machine: [Name]

> Include if feature has lifecycle/state transitions

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ STATE_A  │────▶│ STATE_B  │────▶│ STATE_C  │
└──────────┘     └──────────┘     └──────────┘
```

### State Transitions

| From | To | Trigger | Conditions |
|------|-----|---------|-----------|
| `STATE_A` | `STATE_B` | `trigger()` | Condition description |

<!-- HUMAN-MAINTAINED END -->

<!-- AUTO-GENERATED START -->

## API Endpoints

| Method | Path | Description | Request | Response |
|--------|------|-------------|---------|----------|
| `POST` | `/api/resource` | Create resource | `CreateRequest` | `Resource` |

## Code Signposts

| Component | File | Function/Class | Description |
|-----------|------|----------------|-------------|
| Core | `path/to/file.ts` | `ClassName` | What it does |

## Configuration

| Env Variable | Default | Description |
|--------------|---------|-------------|
| `FEATURE_VAR` | `default` | What it controls |

## Prometheus Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `feature_total` | Counter | Total operations | `tenant`, `status` |

### PromQL Query Examples

```promql
# Rate of operations
rate(feature_total[1h])

# Success rate
sum(rate(feature_total{status="success"}[5m])) / sum(rate(feature_total[5m]))
```

## Unit Tests

| Test | File | Description |
|------|------|-------------|
| `test_basic` | `tests/test_feature.py` | Basic functionality |

## Integration Tests

| Test | File | Prerequisites |
|------|------|---------------|
| `test_e2e` | `tests/e2e/test_feature.py` | Service running |

<!-- AUTO-GENERATED END -->

<!-- HUMAN-MAINTAINED START -->

## Dependencies

| Feature | Relationship | Description |
|---------|--------------|-------------|
| [Other Feature](./other.md) | requires | Must be configured first |

## Example Usage

### API (curl)

```bash
curl -X POST http://localhost:8000/api/resource \
  -H "Content-Type: application/json" \
  -d '{"key": "value"}'
```

### SDK/Programmatic Usage

```typescript
import { Client } from './client'
const result = await client.createResource({ key: 'value' })
```

## Related Features

- [Related Feature](./related.md) - How they interact

<!-- HUMAN-MAINTAINED END -->

## Learnings & Retrospectives

### What Worked Well

1. **[Strength 1]:** Description
2. **[Strength 2]:** Description

### What We'd Do Differently

1. **[Improvement 1]:** What we learned
```

---

## Platform Parallels

| Houston | K8s Equivalent |
|---------|----------------|
| Mission | Job |
| AgentPool | ReplicaSet |
| Policy | NetworkPolicy/OPA |
| Scheduler | kube-scheduler |

---

## Validation Script

```bash
DOC_FILE="path/to/doc.md"

grep -q "Current Status:" "$DOC_FILE" || echo "MISSING: Current Status"
grep -q "Parallel:" "$DOC_FILE" || echo "MISSING: Platform Parallel"
grep -q "Inputs.*Outputs" "$DOC_FILE" || echo "MISSING: Inputs/Outputs"
grep -q "Data Flow" "$DOC_FILE" || echo "MISSING: Data Flow"
grep -q "API Endpoints" "$DOC_FILE" || echo "MISSING: API Endpoints"
grep -q "Code Signposts" "$DOC_FILE" || echo "MISSING: Code Signposts"
grep -q "Prometheus Metrics\|PromQL" "$DOC_FILE" || echo "MISSING: Prometheus"
grep -q "Unit Tests" "$DOC_FILE" || echo "MISSING: Unit Tests"
grep -q "Integration Tests" "$DOC_FILE" || echo "MISSING: Integration Tests"
grep -q "What Worked Well" "$DOC_FILE" || echo "MISSING: What Worked Well"
```
