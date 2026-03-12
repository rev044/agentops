---

-----------------|-----------|---------------------------|
| Chain loaded       | PASS/SKIP | path or "not found"       |
| Prior phases locked| PASS/WARN | list any unlocked         |
| No FAIL verdicts   | PASS/BLOCK| list any FAILed           |
| Artifacts exist    | PASS/WARN | list any missing          |
| Idempotency        | PASS/WARN/INFO | dedup status         |
```

## 6. Blocking Behavior

- **BLOCK** only on FAIL verdicts in prior gates (pre-mortem or vibe). If any check is BLOCK: stop post-mortem and report:
  > "Checkpoint-policy BLOCKED: `<reason>`. Fix the failing gate and re-run."
- **WARN** on everything else (missing phases, missing artifacts, idempotency). Warnings are logged, included in the council packet as `context.checkpoint_warnings`, and execution proceeds.
- **INFO** is purely informational — no action needed.
