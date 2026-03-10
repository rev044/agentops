# Discovery Output Templates

## Execution Packet

Write to `.agents/rpi/execution-packet.json`:

```json
{
  "objective": "<goal>",
  "epic_id": "<epic-id>",
  "contract_surfaces": ["docs/contracts/repo-execution-profile.md"],
  "validation_commands": ["<from repo profile or defaults>"],
  "tracker_mode": "<beads|tasklist>",
  "done_criteria": ["<from repo profile or defaults>"],
  "complexity": "<fast|standard|full>",
  "pre_mortem_verdict": "<PASS|WARN>",
  "discovery_timestamp": "<ISO-8601>"
}
```

## Phase Summary

Write to `.agents/rpi/phase-1-summary-YYYY-MM-DD-<goal-slug>.md`:

```markdown
# Phase 1 Summary: Discovery

- **Goal:** <goal>
- **Epic:** <epic-id>
- **Issues:** <count>
- **Complexity:** <fast|standard|full>
- **Pre-mortem:** <PASS|WARN> (attempt <N>/3)
- **Brainstorm:** <used|skipped>
- **History search:** <findings count or skipped>
- **Status:** DONE
- **Timestamp:** <ISO-8601>
```

## Ratchet and Telemetry

```bash
ao ratchet record research 2>/dev/null || true
bash scripts/checkpoint-commit.sh rpi "phase-1" "discovery complete" 2>/dev/null || true
bash scripts/log-telemetry.sh rpi phase-complete phase=1 phase_name=discovery 2>/dev/null || true
```
