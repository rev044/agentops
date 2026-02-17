# Complexity Scaling

Automatic complexity detection determines the level of validation ceremony applied to each RPI cycle.

## Classification Table

| Level | Issue Count | Wave Count | Ceremony |
|-------|------------|------------|----------|
| **low** | ≤2 | 1 | fast-path: `--quick` on pre-mortem, vibe, post-mortem |
| **medium** | 3-6 | 1-2 | default: standard council on all gates |
| **high** | 7+ OR 3+ waves | any | thorough: `--deep` on pre-mortem and vibe |

## Detection

Complexity is auto-detected after plan completes (Phase 2) by examining:
- Issue count: `bd children <epic-id> | wc -l`
- Wave count: derived from dependency depth

## Flag Precedence (explicit always wins)

| Flag | Effect |
|------|--------|
| `--fast-path` | Forces `low` regardless of auto-detection |
| `--deep` (passed to /rpi) | Forces `high` regardless of auto-detection |
| No flag | Auto-detect from epic structure |

Existing mandatory pre-mortem gate (3+ issues) still applies regardless of complexity level.
