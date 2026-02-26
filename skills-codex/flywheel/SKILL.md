---
name: flywheel
description: 'Knowledge flywheel health monitoring. Checks velocity, pool depths, staleness. Triggers: "flywheel status", "knowledge health", "is knowledge compounding".'
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

Given `$flywheel`:

### Step 1: Measure Knowledge Pools

```bash
# Count top-level artifact files (avoid counting directories)
LEARNINGS=$(find .agents/learnings -maxdepth 1 -type f 2>/dev/null | wc -l)

PATTERNS=$(find .agents/patterns -maxdepth 1 -type f 2>/dev/null | wc -l)

RESEARCH=$(find .agents/research -maxdepth 1 -type f 2>/dev/null | wc -l)

RETROS=$(find .agents/retros -maxdepth 1 -type f 2>/dev/null | wc -l)

echo "Learnings: $LEARNINGS"
echo "Patterns: $PATTERNS"
echo "Research: $RESEARCH"
echo "Retros: $RETROS"
```

### Step 2: Check Recent Activity

```bash
# Recent learnings (last 7 days)
find .agents/learnings -maxdepth 1 -type f -mtime -7 2>/dev/null | wc -l

# Recent research
find .agents/research -maxdepth 1 -type f -mtime -7 2>/dev/null | wc -l
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
  CITE_REPORT=$(ao quality metrics cite-report --json --days 30 2>/dev/null)
  if [ -n "$CITE_REPORT" ]; then
    HIT_RATE=$(echo "$CITE_REPORT" | jq -r '.hit_rate // "unknown"')
    UNCITED=$(echo "$CITE_REPORT" | jq -r '(.uncited_learnings // []) | length')
    STALE_90D=$(echo "$CITE_REPORT" | jq -r '.staleness["90d"] // 0')
    echo "Cache hit rate: $HIT_RATE"
    echo "Uncited learnings: $UNCITED"
    echo "Stale (90d uncited): $STALE_90D"
  fi
else
  # ao-free fallback: compute approximate metrics from files
  echo "Cache health (ao-free fallback):"

  # Learnings modified in last 30 days (active pool)
  ACTIVE_30D=$(find .agents/learnings/ -name "*.md" -mtime -30 2>/dev/null | wc -l | tr -d ' ')
  echo "Active learnings (30d): $ACTIVE_30D"

  # Forge candidates awaiting promotion
  FORGE_PENDING=$(ls .agents/forge/*.md 2>/dev/null | wc -l | tr -d ' ')
  echo "Forge candidates pending: $FORGE_PENDING"

  # Citation tracking (if citations.jsonl exists)
  if [ -f .agents/ao/citations.jsonl ]; then
    CITATION_COUNT=$(wc -l < .agents/ao/citations.jsonl | tr -d ' ')
    UNIQUE_CITED=$(grep -o '"learning_file":"[^"]*"' .agents/ao/citations.jsonl 2>/dev/null | sort -u | wc -l | tr -d ' ')
    echo "Total citations: $CITATION_COUNT"
    echo "Unique learnings cited: $UNIQUE_CITED"
  else
    echo "No citation data (citations.jsonl not found)"
  fi

  # Session outcomes (if outcomes.jsonl exists)
  if [ -f .agents/ao/outcomes.jsonl ]; then
    OUTCOME_COUNT=$(wc -l < .agents/ao/outcomes.jsonl | tr -d ' ')
    echo "Session outcomes recorded: $OUTCOME_COUNT"
  fi
fi
```

### Step 4: Check ao CLI Status

```bash
if command -v ao &>/dev/null; then
  ao quality flywheel status 2>/dev/null || echo "ao quality flywheel status unavailable"
  ao status 2>/dev/null || echo "ao status unavailable"
  ao quality maturity --scan 2>/dev/null || echo "ao quality maturity unavailable"
  ao promote-anti-patterns --dry-run 2>/dev/null || echo "ao promote-anti-patterns unavailable"
  ao badge 2>/dev/null || echo "ao badge unavailable"

  # Knowledge maintenance
  ao dedup --merge 2>/dev/null || true
  ao contradict 2>/dev/null || true
  ao quality constraint review 2>/dev/null || true
  ao curate status 2>/dev/null || true
  ao quality metrics health 2>/dev/null || true
  ao quality metrics cite-report --days 30 2>/dev/null || true

  # Active pruning: archive stale, evict low-utility
  ao quality maturity --expire --archive 2>/dev/null || true
  ao quality maturity --evict --archive 2>/dev/null || true
else
  echo "ao CLI not available — using file-based metrics"

  # Pool inventory
  echo "Pool depths:"
  for pool in learnings patterns forge knowledge research retros; do
    COUNT=$(ls .agents/${pool}/*.md 2>/dev/null | wc -l | tr -d ' ')
    echo "  $pool: $COUNT"
  done

  # Global patterns
  GLOBAL_COUNT=$(ls ~/.claude/patterns/*.md 2>/dev/null | wc -l | tr -d ' ')
  echo "  global patterns: $GLOBAL_COUNT"

  # Check for promotion-ready learnings (see references/promotion-tiers.md)
  echo "See: skills/flywheel/references/promotion-tiers.md for tier definitions"
fi
```

### Step 4.5: Process Metrics (from skill telemetry)

If `.agents/ao/skill-telemetry.jsonl` exists, use `jq` to extract: invocations by skill, average cycle time per skill, gate failure rates. Include in health report (Step 6) under `## Process Metrics`.

### Step 5: Validate Artifact Consistency

Cross-reference validation: scan knowledge artifacts for broken internal references.
Use `scripts/artifact-consistency.sh` (method documented in `references/artifact-consistency.md`).
Default allowlist lives at `references/artifact-consistency-allowlist.txt`; use `--no-allowlist` for a full raw audit.

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

Read `references/cache-eviction.md` for the full eviction pipeline (passive tracking → confidence decay → maturity scan → archive).

## Key Rules

- **Monitor regularly** - flywheel needs attention
- **Address friction** - bottlenecks slow compounding
- **Feed the flywheel** - run $retro and $post-mortem
- **Prune stale knowledge** - archive old artifacts

## Examples

**User says:** `$flywheel` — Counts pool depths, checks recent activity, validates artifact consistency, writes health report to `.agents/flywheel-status.md`.

**Hook trigger:** After `$post-mortem` — Compares current vs historical metrics, flags velocity drops and friction points.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| All pool counts zero | `.agents/` directory missing or empty | Run `$post-mortem` or `$retro` to seed knowledge pools |
| Velocity always zero | No recent extractions (last 7 days) | Run `$retro` or `$post-mortem` to extract and index learnings |
| "ao CLI not available" | ao command not installed or not in PATH | Install ao CLI or use manual pool counting fallback |
| Stale artifacts >50% | Long time since last session or inactive repo | Run `$provenance --stale` to audit and archive old artifacts |

## Reference Documents

- [references/artifact-consistency.md](references/artifact-consistency.md)
- [references/promotion-tiers.md](references/promotion-tiers.md)

---

## References

### artifact-consistency-allowlist.txt

```text
# Artifact consistency allowlist
#
# Format:
#   <source-glob> -> <target-glob>
#
# Use "*" as a wildcard in either side. Keep entries scoped to
# known non-literal or transient references so new actionable
# breakages still surface.

# Runtime telemetry artifacts (optional, not guaranteed in-repo)
* -> .agents/ao/*
* -> .agents/citations.jsonl
* -> .agents/worktrees.json

# RPI runtime outputs (session-specific artifacts)
* -> .agents/rpi/live-status.md
* -> .agents/rpi/session-log.jsonl
* -> .agents/rpi/retry-state.json
* -> .agents/rpi/phase-*-summary.md
* -> .agents/rpi/phase-*-handoff.md
* -> .agents/rpi/phase-*-stream.jsonl
* -> .agents/rpi/contracts/*.json

# Historical placeholders and synthetic examples
* -> .agents/crank/wave-*-checkpoint.json
* -> .agents/*/...*
* -> .agents/*/foo.*
* -> .agents/handoff/auto-99999999T999999Z.md
```

### artifact-consistency.md

# Artifact Consistency Validation

Cross-reference validation: scan knowledge artifacts for broken internal references.

```bash
# Preferred: run the helper script (handles fenced code blocks, placeholders, allowlist).
skills/flywheel/scripts/artifact-consistency.sh

# Optional: include each broken reference for cleanup work.
skills/flywheel/scripts/artifact-consistency.sh --verbose

# Optional: disable allowlist to inspect all historical breakage.
skills/flywheel/scripts/artifact-consistency.sh --no-allowlist

# Optional: use a custom allowlist.
skills/flywheel/scripts/artifact-consistency.sh --allowlist path/to/allowlist.txt
```

The helper script:
- Scans `.agents/**/*.md` excluding `.agents/ao/*`
- Ignores fenced code blocks
- Extracts references to `.agents/...(.md|.json|.jsonl)`
- Skips template placeholders (`YYYY`, `<...>`, `{...}`, wildcards, `...`)
- Applies allowlist patterns from `references/artifact-consistency-allowlist.txt`
- Reports `TOTAL_REFS`, `BROKEN_REFS`, `CONSISTENCY`, `STATUS`
- With `--verbose`, emits `BROKEN_REF=<source> -> <target>` lines

## Allowlist Format

`<source-glob> -> <target-glob>`

Examples:
- `* -> .agents/ao/*` (ignore transient runtime telemetry references)
- `.agents/research/* -> .agents/rpi/phase-*-summary.md` (scope to one source family)

Guidelines:
- Prefer narrow patterns first.
- Keep entries for historical or non-literal references only.
- Remove entries when underlying references are fixed or retired.

## Health Indicator

| Consistency | Status |
|-------------|--------|
| >90% | Healthy |
| 70-90% | Warning |
| <70% | Critical |

### cache-eviction.md

# Cache Eviction

The knowledge flywheel includes automated cache eviction to prevent unbounded growth:

```
Passive Read tracking → Confidence decay → Maturity scan → Archive
```

**How it works:**
1. **Passive tracking** — `ao quality maturity --scan` records when learnings are accessed
2. **Confidence decay** — Unused learnings lose confidence at 10%/week
3. **Composite criteria** — Learnings are eviction candidates when ALL conditions met:
   - Utility < 0.3 (low MemRL score)
   - No citation in 90+ days
   - Confidence < 0.2 (decayed from disuse)
   - Not established maturity (proven knowledge is protected)
4. **Archive** — Candidates move to `.agents/archive/learnings/` (never deleted)

**Commands:**
- `ao quality maturity --evict` — dry-run: show eviction candidates
- `ao quality maturity --evict --archive` — execute: archive candidates
- `ao quality metrics cite-report --days 30` — cache health report

**Kill switches:**
- `AGENTOPS_EVICTION_DISABLED=1` — disable SessionEnd auto-eviction
- `AGENTOPS_PRUNE_AUTO=0` — disable SessionStart auto-pruning (default: off)

### promotion-tiers.md

# Knowledge Promotion Tiers

Defines the maturity pipeline for knowledge artifacts.

## Tier 0: Forge Candidates (`.agents/forge/`)

- **Source:** `$forge` (transcript mining), SessionEnd hook
- **Confidence:** 0.0-0.6
- **Citations:** 0
- **Promotion criteria:** Auto-promote to Tier 1 when confidence >= 0.7 OR cited >= 2 times (ao-free fallback promotes automatically)
- **Eviction:** Candidates older than 90 days with 0 citations are archived

## Tier 1: Learnings (`.agents/learnings/`)

- **Source:** `$retro`, `$learn`, `$extract`, promoted from forge
- **Confidence:** 0.3-1.0
- **Citations:** 1+
- **Promotion criteria:** Promote to Tier 2 when confidence >= 0.8 AND cited >= 3 times AND age > 30 days
- **Eviction:** Learnings older than 90 days with 0 citations decay to archive

## Tier 2: Patterns (`.agents/patterns/`)

- **Source:** Promoted from learnings (manual or automated)
- **Confidence:** 0.8-1.0
- **Citations:** 3+
- **Age:** 30+ days
- **Eviction:** Protected — patterns are long-lived and rarely archived

## Cross-Repo Promotion

- Any tier can be promoted to `~/.claude/patterns/` via `$learn --global`
- Global patterns are user-level, shared across all repositories
- Promotion is a manual decision (human judgment on cross-repo applicability)
- Global patterns are found by `$research`, `$knowledge`, and `$inject` via grep

## Confidence Normalization

When comparing confidence across formats:
- Categorical: high = 0.9, medium = 0.6, low = 0.3
- Numeric: 0.0-1.0 pass through unchanged

## Citation Tracking

Citations are recorded in `.agents/ao/citations.jsonl`:
```json
{"learning_file": ".agents/learnings/example.md", "timestamp": "2026-02-19T12:00:00Z", "session": "session-id"}
```

The `$inject` skill records citations when knowledge is loaded into a session.
The `$post-mortem` skill processes citations to update confidence scores.
The `$flywheel` skill reports citation metrics in health checks.


---

## Scripts

### artifact-consistency.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

VERBOSE=0
ROOT=".agents"
ROOT_SET=0
ALLOWLIST="${ARTIFACT_CONSISTENCY_ALLOWLIST:-$SCRIPT_DIR/../references/artifact-consistency-allowlist.txt}"

usage() {
  cat <<EOF
Usage: $(basename "$0") [--verbose] [--allowlist <path> | --no-allowlist] [ROOT]

Scans markdown files for .agents artifact references and reports consistency.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --verbose)
      VERBOSE=1
      shift
      ;;
    --allowlist)
      if [[ $# -lt 2 ]]; then
        echo "ERROR: --allowlist requires a path" >&2
        exit 2
      fi
      ALLOWLIST="$2"
      shift 2
      ;;
    --no-allowlist)
      ALLOWLIST=""
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --*)
      echo "ERROR: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      if (( ROOT_SET )); then
        echo "ERROR: unexpected argument: $1" >&2
        usage >&2
        exit 2
      fi
      ROOT="$1"
      ROOT_SET=1
      shift
      ;;
  esac
done

trim() {
  sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//'
}

declare -a ALLOW_SOURCE_PATTERNS=()
declare -a ALLOW_TARGET_PATTERNS=()

if [[ -n "$ALLOWLIST" ]]; then
  if [[ ! -f "$ALLOWLIST" ]]; then
    echo "ERROR: allowlist not found: $ALLOWLIST" >&2
    exit 2
  fi

  while IFS= read -r raw_line || [[ -n "$raw_line" ]]; do
    line="$(printf '%s' "$raw_line" | trim)"
    [[ -z "$line" || "$line" == \#* ]] && continue

    source_pattern="*"
    target_pattern="$line"
    if [[ "$line" == *"->"* ]]; then
      source_pattern="$(printf '%s' "${line%%->*}" | trim)"
      target_pattern="$(printf '%s' "${line#*->}" | trim)"
    fi

    [[ -z "$source_pattern" || -z "$target_pattern" ]] && continue
    ALLOW_SOURCE_PATTERNS+=("$source_pattern")
    ALLOW_TARGET_PATTERNS+=("$target_pattern")
  done < "$ALLOWLIST"
fi

is_allowlisted() {
  local source_file="$1"
  local target_ref="$2"
  local i

  for i in "${!ALLOW_SOURCE_PATTERNS[@]}"; do
    if [[ "$source_file" == ${ALLOW_SOURCE_PATTERNS[$i]} ]] \
      && [[ "$target_ref" == ${ALLOW_TARGET_PATTERNS[$i]} ]]; then
      return 0
    fi
  done

  return 1
}

if [[ ! -d "$ROOT" ]]; then
  echo "TOTAL_REFS=0"
  echo "BROKEN_REFS=0"
  echo "CONSISTENCY=100"
  echo "STATUS=Healthy"
  exit 0
fi

total_refs=0
broken_refs=0
broken_lines=()

# Scan markdown while excluding ao telemetry/session data.
while IFS= read -r -d '' file; do
  # Strip fenced code blocks to avoid counting template snippets as broken links.
  refs=$(awk '
    BEGIN { in_code=0 }
    /^```/ { in_code=!in_code; next }
    in_code { next }
    {
      line=$0
      while (match(line, /\.agents\/[A-Za-z0-9._\/-]+\.(md|json|jsonl)/)) {
        print substr(line, RSTART, RLENGTH)
        line = substr(line, RSTART + RLENGTH)
      }
    }
  ' "$file" | sort -u)

  while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue

    # Skip template placeholders and non-literal paths.
    if [[ "$ref" =~ YYYY|\<|\>|\{|\}|\*|\.{3} ]]; then
      continue
    fi

    total_refs=$((total_refs + 1))

    # Normalize leading ./, then check relative to repo root.
    normalized="${ref#./}"
    if [[ ! -f "$normalized" ]]; then
      if is_allowlisted "$file" "$normalized"; then
        continue
      fi
      broken_refs=$((broken_refs + 1))
      if (( VERBOSE )); then
        broken_lines+=("$file -> $normalized")
      fi
    fi
  done <<< "$refs"
done < <(find "$ROOT" -type f -name "*.md" -not -path "$ROOT/ao/*" -print0)

if (( total_refs > 0 )); then
  consistency=$(( (total_refs - broken_refs) * 100 / total_refs ))
else
  consistency=100
fi

status="Critical"
if (( consistency > 90 )); then
  status="Healthy"
elif (( consistency >= 70 )); then
  status="Warning"
fi

echo "TOTAL_REFS=$total_refs"
echo "BROKEN_REFS=$broken_refs"
echo "CONSISTENCY=$consistency"
echo "STATUS=$status"

if (( VERBOSE )) && (( broken_refs > 0 )); then
  for line in "${broken_lines[@]}"; do
    echo "BROKEN_REF=$line"
  done
fi
```

### validate.sh

```bash
#!/usr/bin/env bash
set -euo pipefail
SKILL_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0; FAIL=0
check() { if bash -c "$2"; then echo "PASS: $1"; PASS=$((PASS + 1)); else echo "FAIL: $1"; FAIL=$((FAIL + 1)); fi; }

check "SKILL.md exists" "[ -f '$SKILL_DIR/SKILL.md' ]"
check "SKILL.md has YAML frontmatter" "head -1 '$SKILL_DIR/SKILL.md' | grep -q '^---$'"
check "name is flywheel" "grep -q '^name: flywheel' '$SKILL_DIR/SKILL.md'"
check "mentions health or velocity" "grep -qiE 'health|velocity' '$SKILL_DIR/SKILL.md'"
check "mentions pool" "grep -qi 'pool' '$SKILL_DIR/SKILL.md'"
check "artifact-consistency script exists" "[ -x '$SKILL_DIR/scripts/artifact-consistency.sh' ]"
check "artifact-consistency allowlist exists" "[ -f '$SKILL_DIR/references/artifact-consistency-allowlist.txt' ]"

echo ""; echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
```

