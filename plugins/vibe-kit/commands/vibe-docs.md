---
description: Validate documentation claims against deployment reality
version: 1.0.0
argument-hint: [status|agents|images|full]
model: haiku
---

# /vibe-docs

Invoke the **vibe-docs** skill for semantic documentation validation.

## Arguments

| Argument | Purpose |
|----------|---------|
| `status` | Check status claims against deployment |
| `agents` | Verify agent catalog matches cluster |
| `images` | Verify IMAGE-LIST matches registry |
| `full` | Run all checks |

## Execution

Validates that documentation claims match deployment reality:

1. **Gather Claims** - Extract status/deployment claims from docs
2. **Gather Reality** - Query cluster state (oc get pods, agents)
3. **Compare** - Find mismatches between claims and reality
4. **Report** - Output issues by severity (CRITICAL → LOW)

## Use Cases

- After deployment changes
- Before releases
- Periodic audits (monthly recommended)
- When docs feel "stale"

## Comparison with /doc

| `/doc` | `/vibe-docs` |
|--------|--------------|
| Structure validation | Semantic validation |
| Links work? | Claims true? |
| Sections present? | Status matches reality? |
| Coverage metrics | Accuracy metrics |

## Related

- **Skill**: `~/.claude/skills/vibe-docs/SKILL.md`
- **Ground Truth**: `~/.claude/skills/vibe-docs/references/ground-truth-patterns.md`
- **Structure Validation**: `/doc coverage`
