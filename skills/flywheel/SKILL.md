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

### Step 3.5: Check Cache Health

```bash
if command -v ao &>/dev/null; then
  # Get citation report (cache metrics)
  CITE_REPORT=$(ao metrics cite-report --json --days 30 2>/dev/null)
  if [ -n "$CITE_REPORT" ]; then
    HIT_RATE=$(echo "$CITE_REPORT" | jq -r '.hit_rate // "unknown"')
    UNCITED=$(echo "$CITE_REPORT" | jq -r '.uncited_learnings // 0')
    STALE_90D=$(echo "$CITE_REPORT" | jq -r '.staleness.days_90 // 0')
    echo "Cache hit rate: $HIT_RATE"
    echo "Uncited learnings: $UNCITED"
    echo "Stale (90d uncited): $STALE_90D"
  fi
fi
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

### Step 4.5: Process Metrics (from skill telemetry)

If `.agents/ao/skill-telemetry.jsonl` exists, include skill execution metrics in the health report:

```bash
if [ -f .agents/ao/skill-telemetry.jsonl ]; then
  echo "=== Skill Telemetry Summary ==="
  # Total skill invocations by skill name
  echo "--- Invocations by Skill ---"
  jq -s 'group_by(.skill) | map({skill: .[0].skill, count: length})' .agents/ao/skill-telemetry.jsonl 2>/dev/null || echo "No telemetry data"

  # Average cycle time per skill (requires duration_ms field)
  echo "--- Average Cycle Time ---"
  jq -s 'group_by(.skill) | map({skill: .[0].skill, avg_duration_ms: (map(.duration_ms // 0) | add / length | round)})' .agents/ao/skill-telemetry.jsonl 2>/dev/null || echo "No duration data"

  # Gate failure rates (count of events where verdict != PASS)
  echo "--- Gate Failure Rates ---"
  jq -s '[.[] | select(.verdict != null)] | group_by(.skill) | map({skill: .[0].skill, total: length, failures: [.[] | select(.verdict != "PASS")] | length})' .agents/ao/skill-telemetry.jsonl 2>/dev/null || echo "No verdict data"
fi
```

Include these metrics in the health report (Step 6) under a `## Process Metrics` section when data is available.

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

## Cache Health
- Hit rate: <percentage>%
- Uncited learnings: <count>
- Stale (90d uncited): <count>
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
| Cache hit rate | >80% | 50-80% | <50% |

## Cache Eviction

The knowledge flywheel includes automated cache eviction to prevent unbounded growth:

```
Passive Read tracking → Confidence decay → Maturity scan → Archive
```

**How it works:**
1. **Passive tracking** — PostToolUse(Read) hook records when learnings are accessed
2. **Confidence decay** — Unused learnings lose confidence at 10%/week
3. **Composite criteria** — Learnings are eviction candidates when ALL conditions met:
   - Utility < 0.3 (low MemRL score)
   - No citation in 90+ days
   - Confidence < 0.2 (decayed from disuse)
   - Not established maturity (proven knowledge is protected)
4. **Archive** — Candidates move to `.agents/archive/learnings/` (never deleted)

**Commands:**
- `ao maturity --evict` — dry-run: show eviction candidates
- `ao maturity --evict --archive` — execute: archive candidates
- `ao metrics cite-report --days 30` — cache health report

**Kill switches:**
- `AGENTOPS_EVICTION_DISABLED=1` — disable SessionEnd auto-eviction
- `AGENTOPS_PRUNE_AUTO=0` — disable SessionStart auto-pruning (default: off)

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
