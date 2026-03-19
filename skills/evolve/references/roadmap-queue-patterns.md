# Roadmap Queue Patterns

## Parsing the Queue File

If `--queue` value is not a file path (file does not exist), auto-write the value to `.agents/evolve/roadmap.md` and use that file. This enables inline/prompt-based roadmaps.

```bash
if [ -n "$QUEUE_FILE" ] && [ -f "$QUEUE_FILE" ]; then
  QUEUE_ITEMS=()
  declare -A QUEUE_BLOCKERS
  declare -A QUEUE_LINES  # preserve full line per item ID for freeform prompts
  while IFS= read -r line; do
    ITEM_ID=$(echo "$line" | sed -n 's/.*`\([^`]*\)`.*/\1/p' | head -1)
    BLOCKER=$(echo "$line" | sed -n 's/.*blocker:[[:space:]]*`\([^`]*\)`.*/\1/p')
    if [ -n "$ITEM_ID" ]; then
      QUEUE_ITEMS+=("$ITEM_ID")
      QUEUE_LINES["$ITEM_ID"]="$line"
    fi
    [ -n "$BLOCKER" ] && QUEUE_BLOCKERS["$ITEM_ID"]="$BLOCKER"
  done < <(grep -E '^\s*[0-9]+\.' "$QUEUE_FILE")
  QUEUE_TOTAL=${#QUEUE_ITEMS[@]}
fi
```

## State Persistence and Resume

```bash
# Initialize tracking arrays
PINNED_COMPLETED=()
PINNED_ESCALATED='[]'

# Resume from persisted state
if [ -f .agents/evolve/pinned-queue-state.json ]; then
  QUEUE_INDEX=$(jq -r '.current_index // 0' .agents/evolve/pinned-queue-state.json)
  mapfile -t PINNED_COMPLETED < <(jq -r '.completed[]? // empty' .agents/evolve/pinned-queue-state.json 2>/dev/null)
  ESCALATED_IDS=$(jq -r '.escalated[]?.id // empty' .agents/evolve/pinned-queue-state.json 2>/dev/null)
  PINNED_ESCALATED=$(jq -c '.escalated // []' .agents/evolve/pinned-queue-state.json 2>/dev/null)
else
  QUEUE_INDEX=0
fi
```

See `references/pinned-queue.md` for format specification, blocker syntax, and state schema.

## Work Selection (Step 3.0)

If a pinned queue exists and `QUEUE_INDEX < QUEUE_TOTAL`:

1. Read the current item: `CURRENT_ITEM=${QUEUE_ITEMS[$QUEUE_INDEX]}`
2. Skip if this item ID is in the escalated list (log skip reason, advance index, re-enter)
3. Check for declared blocker: `QUEUE_BLOCKERS[$CURRENT_ITEM]`
4. If blocker exists AND not yet resolved (not in `pinned_queue_completed`): set `UNBLOCK_TARGET` and proceed to blocker resolution
5. If no blocker (or already resolved): proceed to execution with the queue item

**Item-to-prompt mapping:**
- If item ID matches a bead (`bd show $CURRENT_ITEM` succeeds): `/rpi "Land $CURRENT_ITEM: $(bd show $CURRENT_ITEM --json | jq -r .title)" --auto --max-cycles=1`
- Otherwise, use the preserved full queue line: `/rpi "${QUEUE_LINES[$CURRENT_ITEM]}" --auto --max-cycles=1`

**Escalation cascade guard:** When an item is escalated (skipped), check if subsequent items declare the escalated item as a `blocker:`. If so, those dependent items are also marked escalated.

When pinned queue is active, skip Steps 3.1-3.7 entirely. The queue IS the work source. When pinned queue is exhausted (`QUEUE_INDEX >= QUEUE_TOTAL`), fall through to normal selection (Steps 3.1-3.7).

## Blocker Resolution (Step 4.1)

If `UNBLOCK_TARGET` is set, enter the blocker resolution sub-loop:

```text
unblock_loop:
  if UNBLOCK_DEPTH > 2:
    ESCALATE: "Blocker chain too deep (>2 levels). Item: $ITEM_ID, chain: $UNBLOCK_CHAIN"
    Write escalation to .agents/evolve/escalated.md
    Mark item as escalated in pinned-queue-state.json
    Run escalation cascade for dependent items
    Advance QUEUE_INDEX to next non-escalated item
    Return to Step 3

  Run: /rpi "Unblock: land $UNBLOCK_TARGET as minimum unblocker" --auto --max-cycles=1

  if unblock succeeded:
    Close/update blocker bead if applicable (bd close $UNBLOCK_TARGET)
    Add UNBLOCK_TARGET to pinned_queue_completed
    Clear UNBLOCK_TARGET, reset UNBLOCK_DEPTH to 0
    UNBLOCK_FAILURES=0
    Persist queue state (atomic write)
    Return to Step 3.0 to re-check the original item

  if unblock failed:
    UNBLOCK_FAILURES++
    if UNBLOCK_FAILURES >= 3:
      ESCALATE: "3 consecutive unblock failures on $UNBLOCK_TARGET"
      Write escalation, mark escalated, run cascade
      Advance QUEUE_INDEX, return to Step 3

    Dynamic blocker detection — scan /rpi failure output for:
      - bead IDs mentioned in error context
      - dependency keywords ("blocked by", "requires", "depends on")
      - import/build failures pointing to missing prerequisites
    if deeper_blocker found AND UNBLOCK_DEPTH < 2:
      UNBLOCK_DEPTH++
      Push current UNBLOCK_TARGET to UNBLOCK_CHAIN
      Set UNBLOCK_TARGET = deeper_blocker
      goto unblock_loop
    else:
      Retry with narrowed scope and --quality
      goto unblock_loop
```

Kill switch is checked at the top of EVERY sub-`/rpi` invocation.

## Queue Advancement After Cycle

```bash
# Advance pinned queue after successful cycle (not unblock sub-cycles)
if [ -n "$QUEUE_FILE" ] && [ -z "$UNBLOCK_TARGET" ] && [ "$OUTCOME" != "regressed" ] && [ "$OUTCOME" != "failed" ]; then
  PINNED_COMPLETED+=("$CURRENT_ITEM")
  QUEUE_INDEX=$((QUEUE_INDEX + 1))
  # Persist queue state (atomic write via temp file)
  TMP=$(mktemp .agents/evolve/pinned-queue-state.XXXXXX.json)
  jq -n --arg file "$QUEUE_FILE" --argjson idx "$QUEUE_INDEX" \
    --argjson completed "$(printf '%s\n' "${PINNED_COMPLETED[@]}" | jq -R . | jq -s .)" \
    --argjson escalated "$PINNED_ESCALATED" \
    '{queue_file: $file, current_index: $idx, completed: $completed, in_progress: null, escalated: $escalated, unblock_chain: []}' \
    > "$TMP" && jq . "$TMP" >/dev/null 2>&1 && mv "$TMP" .agents/evolve/pinned-queue-state.json
fi
```

## Circuit Breakers (Queue Mode)

```bash
# Consecutive failure breaker (pinned queue mode)
if [ -n "$QUEUE_FILE" ]; then
  CONSEC_FAILURES=$(awk '/"result"\s*:\s*"(regressed|unchanged)"/{streak++; next} {streak=0} END{print streak+0}' \
    .agents/evolve/cycle-history.jsonl 2>/dev/null)
  if [ "$CONSEC_FAILURES" -ge 5 ]; then
    echo "CIRCUIT BREAKER: 5 consecutive failures in pinned queue mode. Stopping."
    # go to Teardown
  fi
fi
```
