---

----|--------|---------|
| Evidence Precedence | PASS/WARN/FAIL | N children resolved by commit/staged/worktree, M without evidence |
| Phantom Beads | PASS/WARN | N phantom beads detected |
| Orphaned Children | PASS/WARN | N orphans found |
| Multi-Wave Regression | PASS/FAIL | N regressions detected |
| Stretch Goals | PASS/WARN | N stretch goals closed without rationale |

### Findings
- <specific findings from each check>
```

## Integration with Council

Include closure integrity results in the council packet:

```json
{
  "context": {
    "closure_integrity": {
      "git_evidence_failures": [...],
      "evidence_modes": {
        "commit": [...],
        "staged": [...],
        "worktree": [...]
      },
      "phantom_beads": [...],
      "orphaned_children": [...],
      "wave_regressions": [...],
      "stretch_audit": [...]
    }
  }
}
```

The `plan-compliance` judge uses these to assess whether the epic should actually be marked complete.
