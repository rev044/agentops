# Pinned Queue: Format, Blocker Resolution, and State Persistence

> Pinned queue mode processes an ordered user-defined roadmap before falling back to fitness-driven selection.

## Queue File Format

An ordered markdown checklist. Each numbered item is a work unit:

```markdown
1. `rig-difc` — Extract PFN/E2E subsystems and remove them.
   blocker: `rig-8z29` (land as minimum unblocker first)
2. `rig-f81f.1` — Converge operational workflow domains.
3. `rig-imjw` — Converge cluster/env/inventory/release.
```

### Parsing Rules

- Lines matching `^\s*[0-9]+\.` are queue items, processed in order.
- The **first backtick-delimited string** on each line is the item ID.
- An optional `blocker: \`<id>\`` annotation declares a known blocker.
- Items without blockers proceed directly to `/rpi`.
- Indented sub-lines (not starting with a number) are description context passed to `/rpi`.

### macOS-Safe Parsing

Use `sed` (not `grep -oP`) for portability:

```bash
ITEM_ID=$(echo "$line" | sed -n 's/.*`\([^`]*\)`.*/\1/p' | head -1)
BLOCKER=$(echo "$line" | sed -n 's/.*blocker:[[:space:]]*`\([^`]*\)`.*/\1/p')
```

## --queue Flag Behavior

```bash
/evolve --queue=.agents/evolve/roadmap.md       # File-based queue
/evolve --queue=.agents/evolve/roadmap.md --test-first  # With strict quality
```

When `--queue` value is not a file path (file does not exist at that path), evolve auto-writes the value to `.agents/evolve/roadmap.md` and uses that file. This enables inline/prompt-based roadmaps.

## Item-to-Prompt Mapping

When processing a queue item:

1. If the item ID matches a bead (`bd show $ITEM_ID` succeeds), use:
   ```
   /rpi "Land $ITEM_ID: $(bd show $ITEM_ID --json | jq -r .title)" --auto --max-cycles=1
   ```
2. Otherwise, use the full queue line as a freeform prompt:
   ```
   /rpi "$FULL_LINE" --auto --max-cycles=1
   ```

## Blocker Resolution Protocol

When a queue item has a declared blocker or `/rpi` fails revealing an undeclared one:

```
detect_blocker(item) →
  if UNBLOCK_DEPTH > MAX_DEPTH (2):
    ESCALATE → write to escalated.md, skip item, continue to next
  spawn /rpi on blocker (--auto --max-cycles=1)
  if success: resume original item
  if failure:
    UNBLOCK_FAILURES++
    if UNBLOCK_FAILURES >= 3: ESCALATE
    check for deeper blocker via dynamic detection
    if deeper found AND depth < 2: recurse
    else: retry with --quality (narrowed scope + pre-mortem)
```

### Dynamic Blocker Detection

When `/rpi` fails, scan the failure output for:

- **Bead IDs** mentioned in error context (`bd show $ID` succeeds)
- **Dependency keywords**: "blocked by", "requires", "depends on"
- **Build/import failures** pointing to missing prerequisites

If no signal is detected, retry once with narrowed scope, then escalate.

### Escalation Cascade

When an item is escalated (skipped), check if subsequent items declare the escalated item as a `blocker:`. If so, those dependent items are also marked escalated. This prevents attempting work that depends on a skipped item.

### Guardrails

| Guardrail | Value | Effect |
|-----------|-------|--------|
| Max unblock depth | 2 | Blocker-of-blocker-of-blocker escalates |
| Max consecutive unblock failures | 3 | Per-item, then escalate and skip |
| Kill switch | `.agents/evolve/STOP` or `~/.config/evolve/KILL` | Checked at top of every cycle AND sub-cycle |
| Consecutive failure circuit breaker | 5 | 5 consecutive failed cycles in pinned mode → teardown |

## Escalation Output

Escalated items are written to `.agents/evolve/escalated.md`:

```markdown
## Escalated Items

| Item | Reason | Cycle | Blocker Chain |
|------|--------|-------|---------------|
| `rig-difc` | 3 consecutive unblock failures on `rig-8z29` | 4 | rig-8z29 |
| `rig-f81f.1` | Cascade: depends on escalated `rig-difc` | 4 | rig-difc → rig-8z29 |
```

Summary also printed to stdout. Evolve continues to next non-escalated item.

## Queue State Persistence

State file: `.agents/evolve/pinned-queue-state.json`

### Schema

```json
{
  "queue_file": "string — path to queue file",
  "current_index": 0,
  "completed": ["string — completed item IDs"],
  "in_progress": null,
  "escalated": [
    {"id": "string", "reason": "string", "cycle": 0, "chain": ["string"]}
  ],
  "unblock_chain": []
}
```

### Atomic Write

Always write via temp file + JSON validation + `mv`:

```bash
TMP=$(mktemp .agents/evolve/pinned-queue-state.XXXXXX.json)
jq -n --arg file "$QUEUE_FILE" --argjson idx "$QUEUE_INDEX" \
  --argjson completed "$(printf '%s\n' "${PINNED_COMPLETED[@]}" | jq -R . | jq -s .)" \
  --argjson escalated "$PINNED_ESCALATED" \
  '{queue_file: $file, current_index: $idx, completed: $completed, in_progress: null, escalated: $escalated, unblock_chain: []}' \
  > "$TMP" && jq . "$TMP" >/dev/null 2>&1 && mv "$TMP" .agents/evolve/pinned-queue-state.json
```

### Resume Across Sessions

On session restart, evolve reads `pinned-queue-state.json` and resumes from `current_index`. Completed items are not re-processed. Escalated items are skipped with a log message.

## Cycle History Fields (Pinned Mode)

Pinned queue cycles add these fields to `cycle-history.jsonl`:

```json
{
  "cycle": 7,
  "mode": "pinned",
  "queue_item": "rig-difc",
  "queue_index": 2,
  "queue_total": 9,
  "unblock_target": "rig-8z29",
  "unblock_depth": 1,
  "result": "improved"
}
```

## Examples

### Simple Roadmap

```markdown
1. `add-auth` — Add JWT authentication middleware
2. `add-rbac` — Add role-based access control
   blocker: `add-auth`
3. `add-audit-log` — Add audit logging for auth events
   blocker: `add-rbac`
```

### Roadmap with Shared Dependency (Fan-Out Blocker)

```markdown
1. `extract-core` — Extract core library from monolith
2. `migrate-db` — Migrate to new database schema
   blocker: `extract-core`
3. `update-api` — Update API to use new core
   blocker: `extract-core`
4. `integration-tests` — Full integration test suite
   blocker: `update-api`
```
