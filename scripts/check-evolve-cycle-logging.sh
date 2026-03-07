#!/usr/bin/env bash
set -euo pipefail

# Validate evolve cycle-history.jsonl integrity.
# Checks: file exists (if evolve has run), entries have required fields,
# cycle numbers are monotonically increasing, and productive entries carry
# the fields needed for trajectory plotting. Historical rows remain warning-
# tolerant so legacy schema drift does not block the repo. Numbering gaps are
# warnings by default; pass --strict-gaps to fail on them.

STRICT_GAPS=false
if [[ "${1:-}" == "--strict-gaps" ]]; then
  STRICT_GAPS=true
  shift
fi

if [[ $# -gt 0 ]]; then
  echo "Usage: $0 [--strict-gaps]"
  exit 2
fi

HISTORY=".agents/evolve/cycle-history.jsonl"

# If no evolve directory exists, skip gracefully (evolve hasn't run yet)
if [[ ! -d ".agents/evolve" ]]; then
  echo "No .agents/evolve/ directory — evolve has not run yet. Skipping."
  exit 0
fi

# If evolve directory exists but no history file, check for fitness snapshots
# which would indicate cycles ran without logging (the exact bug we're catching)
if [[ ! -f "$HISTORY" ]]; then
  SNAPSHOT_COUNT=$(find .agents/evolve -name 'fitness-*-post.json' 2>/dev/null | wc -l | tr -d ' ')
  if [[ "$SNAPSHOT_COUNT" -gt 0 ]]; then
    echo "ERROR: Found $SNAPSHOT_COUNT post-cycle fitness snapshots but no cycle-history.jsonl."
    echo "This indicates evolve cycles ran without logging — the tracking bug this goal prevents."
    exit 1
  fi
  echo "No cycle-history.jsonl and no post-cycle snapshots. Evolve has not completed any cycles."
  exit 0
fi

LINE_NUM=0
ERRORS=0
WARNINGS=0
PREV_CYCLE=-1
LAST_NON_EMPTY=0

while IFS= read -r line; do
  LINE_NUM=$((LINE_NUM + 1))

  # Skip empty lines
  [[ -z "$line" ]] && continue
  LAST_NON_EMPTY=$LINE_NUM

  # Validate JSON
  if ! echo "$line" | jq empty 2>/dev/null; then
    echo "ERROR: Line $LINE_NUM is not valid JSON"
    ERRORS=$((ERRORS + 1))
    continue
  fi

  # Check required scalar fields
  for field in cycle result timestamp; do
    VALUE=$(echo "$line" | jq -r ".$field // empty")
    if [[ -z "$VALUE" ]]; then
      echo "ERROR: Line $LINE_NUM missing required field: $field"
      ERRORS=$((ERRORS + 1))
    fi
  done

  TARGET=$(echo "$line" | jq -r '.target // .goal_id // empty')
  GOAL_IDS_LEN=$(echo "$line" | jq -r '(.goal_ids // []) | length')
  if [[ -z "$TARGET" && "$GOAL_IDS_LEN" -eq 0 ]]; then
    echo "ERROR: Line $LINE_NUM missing target/goal_id and goal_ids"
    ERRORS=$((ERRORS + 1))
  elif [[ -z "$TARGET" ]]; then
    echo "WARN: Line $LINE_NUM uses goal_ids parallel schema without target"
    WARNINGS=$((WARNINGS + 1))
  elif [[ "$(echo "$line" | jq -r 'has("target")')" != "true" ]]; then
    echo "WARN: Line $LINE_NUM uses legacy goal_id field"
    WARNINGS=$((WARNINGS + 1))
  fi

  RESULT=$(echo "$line" | jq -r '.result // empty')
  case "$RESULT" in
    improved|regressed|harvested)
      for field in sha goals_passing goals_total; do
        VALUE=$(echo "$line" | jq -r ".$field // empty")
        if [[ -z "$VALUE" ]]; then
          echo "WARN: Line $LINE_NUM missing productive field: $field"
          WARNINGS=$((WARNINGS + 1))
        fi
      done
      for field in goals_passing goals_total; do
        TYPE=$(echo "$line" | jq -r "if has(\"$field\") then (.$field | type) else \"missing\" end")
        if [[ "$TYPE" != "number" ]]; then
          echo "WARN: Line $LINE_NUM uses non-numeric $field ($TYPE)"
          WARNINGS=$((WARNINGS + 1))
        fi
      done
      CANONICAL_TYPE=$(echo "$line" | jq -r 'if has("canonical_sha") then (.canonical_sha | type) else "missing" end')
      if [[ "$CANONICAL_TYPE" != "missing" && "$CANONICAL_TYPE" != "string" ]]; then
        echo "WARN: Line $LINE_NUM uses non-string canonical_sha ($CANONICAL_TYPE)"
        WARNINGS=$((WARNINGS + 1))
      fi
      LOG_TYPE=$(echo "$line" | jq -r 'if has("log_sha") then (.log_sha | type) else "missing" end')
      if [[ "$LOG_TYPE" != "missing" && "$LOG_TYPE" != "string" ]]; then
        echo "WARN: Line $LINE_NUM uses non-string log_sha ($LOG_TYPE)"
        WARNINGS=$((WARNINGS + 1))
      fi
      SHA_VALUE=$(echo "$line" | jq -r '.sha // empty')
      CANONICAL_VALUE=$(echo "$line" | jq -r '.canonical_sha // empty')
      if [[ -n "$CANONICAL_VALUE" && -z "$SHA_VALUE" ]]; then
        echo "WARN: Line $LINE_NUM has canonical_sha but no compatibility sha"
        WARNINGS=$((WARNINGS + 1))
      elif [[ -n "$SHA_VALUE" && -n "$CANONICAL_VALUE" && "$SHA_VALUE" != "$CANONICAL_VALUE" ]]; then
        echo "WARN: Line $LINE_NUM has sha/canonical_sha mismatch"
        WARNINGS=$((WARNINGS + 1))
      fi
      ;;
    unchanged|quarantined)
      :
      ;;
    *)
      echo "WARN: Line $LINE_NUM uses unrecognized result: $RESULT"
      WARNINGS=$((WARNINGS + 1))
      ;;
  esac

  CYCLE_TYPE=$(echo "$line" | jq -r '.cycle | type')
  if [[ "$CYCLE_TYPE" != "number" ]]; then
    echo "ERROR: Line $LINE_NUM cycle is not numeric ($CYCLE_TYPE)"
    ERRORS=$((ERRORS + 1))
    continue
  fi

  # Check cycle number monotonicity
  CYCLE=$(echo "$line" | jq -r '.cycle // -1')
  if [[ "$PREV_CYCLE" -ge 0 ]]; then
    EXPECTED=$((PREV_CYCLE + 1))
    if [[ "$CYCLE" -le "$PREV_CYCLE" ]]; then
      echo "ERROR: Non-increasing cycle at line $LINE_NUM: previous $PREV_CYCLE, got $CYCLE"
      ERRORS=$((ERRORS + 1))
    elif [[ "$CYCLE" -ne "$EXPECTED" ]]; then
      if [[ "$STRICT_GAPS" == "true" ]]; then
        echo "ERROR: Cycle gap at line $LINE_NUM: expected cycle $EXPECTED, got $CYCLE"
        ERRORS=$((ERRORS + 1))
      else
        echo "WARN: Cycle gap at line $LINE_NUM: expected cycle $EXPECTED, got $CYCLE"
        WARNINGS=$((WARNINGS + 1))
      fi
    fi
  elif [[ "$CYCLE" -ne 1 ]]; then
    echo "WARN: First logged cycle is $CYCLE (expected 1)"
    WARNINGS=$((WARNINGS + 1))
  fi
  PREV_CYCLE="$CYCLE"

done < "$HISTORY"

if [[ "$LAST_NON_EMPTY" -eq 0 ]]; then
  echo "WARN: cycle-history.jsonl exists but is empty."
  exit 0
fi

if [[ "$ERRORS" -gt 0 ]]; then
  echo
  echo "ERROR: $ERRORS integrity issues found in cycle-history.jsonl ($LINE_NUM entries checked)."
  exit 1
fi

if [[ "$WARNINGS" -gt 0 ]]; then
  echo "cycle-history.jsonl OK with warnings: $LAST_NON_EMPTY entries checked, cycles 1-$PREV_CYCLE, warnings=$WARNINGS."
  exit 0
fi

echo "cycle-history.jsonl OK: $LAST_NON_EMPTY entries, cycles 1-$PREV_CYCLE, required fields present."
exit 0
