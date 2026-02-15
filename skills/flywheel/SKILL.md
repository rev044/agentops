---
name: flywheel
description: 'Knowledge flywheel health monitoring. Checks velocity, pool depths, staleness. Triggers: "flywheel status", "knowledge health", "is knowledge compounding".'
metadata:
  tier: background
  dependencies: []
  internal: true
---

# Flywheel Skill

Monitor the knowledge flywheel health.

## The Flywheel Model

```
Sessions → Transcripts → Forge → Pool → Promote → Knowledge
     ↑                                               │
     └───────────────────────────────────────────────┘
                    Future sessions find it
```

**Velocity** = Rate of knowledge flowing through
**Friction** = Bottlenecks slowing the flywheel

## Execution Steps

Given `/flywheel`:

### Step 1: Measure Knowledge Pools

```bash
# Count learnings
LEARNINGS=$(ls .agents/learnings/ 2>/dev/null | wc -l)

# Count patterns
PATTERNS=$(ls .agents/patterns/ 2>/dev/null | wc -l)

# Count research
RESEARCH=$(ls .agents/research/ 2>/dev/null | wc -l)

# Count retros
RETROS=$(ls .agents/retros/ 2>/dev/null | wc -l)

echo "Learnings: $LEARNINGS"
echo "Patterns: $PATTERNS"
echo "Research: $RESEARCH"
echo "Retros: $RETROS"
```

### Step 2: Check Recent Activity

```bash
# Recent learnings (last 7 days)
find .agents/learnings/ -mtime -7 2>/dev/null | wc -l

# Recent research
find .agents/research/ -mtime -7 2>/dev/null | wc -l
```

### Step 3: Detect Staleness

```bash
# Old artifacts (> 30 days without modification)
find .agents/ -name "*.md" -mtime +30 2>/dev/null | wc -l
```

### Step 4: Check ao CLI Status

```bash
if command -v ao &>/dev/null; then
  ao forge status 2>/dev/null || echo "ao forge status unavailable"
  ao maturity --scan 2>/dev/null || echo "ao maturity unavailable"
  ao promote-anti-patterns --dry-run 2>/dev/null || echo "ao promote-anti-patterns unavailable"
  ao badge 2>/dev/null || echo "ao badge unavailable"
else
  echo "ao CLI not available"
fi
```

### Step 5: Validate Artifact Consistency

Cross-reference validation: scan knowledge artifacts for broken internal references.
Read `references/artifact-consistency.md` for validation details.

Health indicator: >90% = Healthy, 70-90% = Warning, <70% = Critical.

### Step 6: Write Health Report

**Write to:** `.agents/flywheel-status.md`

```markdown
# Knowledge Flywheel Health

**Date:** YYYY-MM-DD

## Pool Depths
| Pool | Count | Recent (7d) |
|------|-------|-------------|
| Learnings | <count> | <count> |
| Patterns | <count> | <count> |
| Research | <count> | <count> |
| Retros | <count> | <count> |

## Velocity (Last 7 Days)
- Sessions with extractions: <count>
- New learnings: <count>
- New patterns: <count>

## Artifact Consistency
- References scanned: <count>
- Broken references: <count>
- Consistency score: <percentage>%
- Status: <Healthy/Warning/Critical>

## Health Status
<Healthy/Warning/Critical>

## Friction Points
- <issue 1>
- <issue 2>

## Recommendations
1. <recommendation>
2. <recommendation>
```

### Step 7: Report to User

Tell the user:
1. Overall flywheel health
2. Knowledge pool depths
3. Recent activity
4. Any friction points
5. Recommendations

## Health Indicators

| Metric | Healthy | Warning | Critical |
|--------|---------|---------|----------|
| Learnings/week | 3+ | 1-2 | 0 |
| Stale artifacts | <20% | 20-50% | >50% |
| Research/plan ratio | >0.5 | 0.2-0.5 | <0.2 |

## Key Rules

- **Monitor regularly** - flywheel needs attention
- **Address friction** - bottlenecks slow compounding
- **Feed the flywheel** - run /retro and /post-mortem
- **Prune stale knowledge** - archive old artifacts

## Examples

### Status Check Invocation

**User says:** `/flywheel` or "check knowledge health"

**What happens:**
1. Agent counts artifacts in `.agents/learnings/`, `.agents/patterns/`, `.agents/research/`, `.agents/retros/`
2. Agent checks recent activity with `find -mtime -7`
3. Agent detects stale artifacts with `find -mtime +30`
4. Agent calls `ao forge status` to check CLI state
5. Agent validates artifact consistency (cross-references)
6. Agent writes health report to `.agents/flywheel-status.md`
7. Agent reports overall health, friction points, recommendations

**Result:** Single-screen dashboard showing knowledge flywheel velocity, pool depths, and health status.

### Automated Health Monitoring

**Hook triggers:** Periodic check or after `/post-mortem`

**What happens:**
1. Hook calls flywheel skill to measure pools
2. Agent compares current vs historical metrics
3. Agent detects velocity drops (learnings/week < threshold)
4. Agent flags friction points (e.g., stale artifacts >50%)
5. Agent recommends actions to restore velocity

**Result:** Proactive alerts when knowledge flywheel slows or stalls, enabling intervention before bottlenecks harden.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| All pool counts zero | `.agents/` directory missing or empty | Run `/post-mortem` or `/retro` to seed knowledge pools |
| Velocity always zero | No recent extractions (last 7 days) | Run `/forge` + `/extract` to process pending sessions |
| "ao CLI not available" | ao command not installed or not in PATH | Install ao CLI or use manual pool counting fallback |
| Stale artifacts >50% | Long time since last session or inactive repo | Run `/provenance --stale` to audit and archive old artifacts |
