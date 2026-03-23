---
name: validation
description: 'Full validation phase orchestrator. Vibe + post-mortem + retro + forge. Reviews implementation quality, extracts learnings, feeds the knowledge flywheel. Triggers: "validation", "validate", "validate work", "review and learn", "validation phase", "post-implementation review".'
---

# $validation — Full Validation Phase Orchestrator

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

## DAG — Execute This Sequentially

```
mkdir -p .agents/rpi
detect complexity from execution-packet or --complexity flag (default: standard)
detect ao CLI availability
```

**Run every step in order. Do not stop between steps.**

```
STEP 1  ──  $vibe recent [--quick]
              Use --quick for fast/standard. Full council for full.
              PASS/WARN? → continue
              FAIL?      → write summary, output <promise>FAIL</promise>, stop
                           (validation cannot fix code — caller decides retry)

STEP 1.5 ── Test pyramid coverage audit (advisory, append to summary)
              Check L0-L3 + BF1/BF4 per modified file. WARN only, not FAIL.

STEP 2  ──  if epic_id:
              $post-mortem <epic-id> [--quick]
              Use --quick for fast/standard. Full council for full.

STEP 3  ──  if not --no-retro:
              $retro

STEP 4  ──  if not --no-forge AND ao available:
              if Codex hookless mode:
                if epic_id:
                  $post-mortem owns Codex closeout for this run; skip duplicate ao codex stop here
                else:
                  inspect .agents/ao/codex/state.json when present
                  if the state is missing, unreadable, or last_stop.session_id != CODEX_THREAD_ID:
                    ao codex stop --auto-extract 2>/dev/null || true
                  else:
                    skip duplicate Codex closeout for this thread
              else:
                ao forge transcript --last-session --queue --quiet 2>/dev/null || true

STEP 5  ──  write phase summary to .agents/rpi/phase-3-summary-YYYY-MM-DD-<slug>.md
              ao ratchet record vibe 2>/dev/null || true
              output <promise>DONE</promise>
```

**That's it.** Steps 1→2→3→4→5. No stopping between steps.

---

## Setup Detail

**State:**
```
validation_state = {
  epic_id: "<epic-id or null>",
  complexity: <fast|standard|full>,
  no_retro: <true if --no-retro>,
  no_forge: <true if --no-forge>,
  vibe_verdict: null,
  post_mortem_verdict: null
}
```

**Load execution packet** (if available): read `complexity`, `contract_surfaces`, `done_criteria` from `.agents/rpi/execution-packet.json`.

## Gate Detail

**STEP 1 (vibe) is the only gate.** Validation cannot fix code — it can only report.

- **PASS/WARN:** Log verdict, continue to STEP 2.
- **FAIL:** Extract findings from `ls -t .agents/council/*vibe*.md | head -1`, write phase summary with FAIL status, output `<promise>FAIL</promise>` with findings attached. Suggest: `"Vibe FAIL. Fix findings, then re-run $validation [epic-id]"`.

## Phase Summary Format

Write to `.agents/rpi/phase-3-summary-YYYY-MM-DD-<slug>.md`:

```markdown
# Phase 3 Summary: Validation

- **Epic:** <epic-id or "standalone">
- **Vibe verdict:** <PASS|WARN|FAIL>
- **Post-mortem verdict:** <verdict or "skipped">
- **Retro:** <captured|skipped>
- **Forge:** <mined|skipped>
- **Complexity:** <fast|standard|full>
- **Status:** <DONE|FAIL>
- **Timestamp:** <ISO-8601>
```

## Phase Budgets

| Sub-step | `fast` | `standard` | `full` |
|----------|--------|------------|--------|
| Vibe | 2 min | 3 min | 5 min |
| Post-mortem | 2 min | 3 min | 5 min |
| Retro | 1 min | 1 min | 2 min |
| Forge | skip | 2 min | 3 min |

On budget expiry: allow in-flight calls to complete, write `[TIME-BOXED]` marker, proceed.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--complexity=<level>` | auto | Force complexity level (fast/standard/full) |
| `--no-retro` | off | Skip retro + forge steps |
| `--no-forge` | off | Skip forge step only |
| `--no-budget` | off | Disable phase time budgets |

## Quick Start

```bash
$validation ag-5k2                        # validate epic with full close-out
$validation                               # validate recent work (no epic)
$validation --complexity=full ag-5k2      # force full council ceremony
$validation --no-retro ag-5k2             # skip retro + forge
$validation --no-forge ag-5k2             # skip forge only
```

## Completion Markers

```
<promise>DONE</promise>    # Validation passed, learnings captured
<promise>FAIL</promise>    # Vibe failed, re-implementation needed (findings attached)
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Vibe FAIL on first run | Implementation has quality issues | Fix findings via `$crank`, then re-run `$validation` |
| Post-mortem skipped unexpectedly | No epic-id provided | Pass epic-id: `$validation ag-5k2` |
| Codex closeout missing | Codex has no session-end hook surface | Standalone `$validation` ensures `ao codex stop` once per thread; epic validation defers Codex closeout to `$post-mortem`. Use `ao codex status` to confirm lifecycle state manually |
| Forge produces no output | No ao CLI or no transcript content | Install ao CLI or run `$retro` manually |
| Stale execution-packet | Packet from a previous RPI cycle | Delete `.agents/rpi/execution-packet.json` and pass `--complexity` explicitly |

## Reference Documents

- [references/forge-scope.md](references/forge-scope.md) — forge session scoping and deduplication
- [references/idempotency-and-resume.md](references/idempotency-and-resume.md) — re-run behavior and standalone mode

## See Also

- [../vibe/SKILL.md](../vibe/SKILL.md) — code quality review
- [../post-mortem/SKILL.md](../post-mortem/SKILL.md) — retrospective analysis
- [../retro/SKILL.md](../retro/SKILL.md) — quick learning capture
- [../forge/SKILL.md](../forge/SKILL.md) — transcript mining
- [../crank/SKILL.md](../crank/SKILL.md) — previous phase (implementation)
- [../discovery/SKILL.md](../discovery/SKILL.md) — first phase (discovery)
- [../rpi/SKILL.md](../rpi/SKILL.md) — full lifecycle orchestrator

## Local Resources

### references/

- [references/forge-scope.md](references/forge-scope.md)
- [references/idempotency-and-resume.md](references/idempotency-and-resume.md)
