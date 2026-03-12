---

----|------|-------------|
| `current_streak_days` | integer | Consecutive calendar days with at least 1 rpi-state.json update. Minimum 1 if state file exists. |
| `last_rpi_date` | string (YYYY-MM-DD) | Date of most recent rpi-state.json update. |
| `total_rpi_sessions` | integer | Count of distinct `session_id` values found in rpi state history. Falls back to 1 if only current state exists. |

## Counting Rules

- **Session boundary:** Defined by `session_id` in `rpi-state.json`. Do NOT infer sessions from timestamps.
- **Consecutive days:** A streak breaks if a calendar day (UTC) passes with zero `rpi-state.json` updates.
- **Minimum streak:** If `rpi-state.json` exists and is valid, the minimum streak is 1.
- **Historical data:** If `.agents/rpi/` contains archived state files or `outcomes.jsonl`, use those to compute multi-day streaks. Otherwise, streak = 1 (current session only).

## Fallback Behavior

| Condition | Behavior |
|-----------|----------|
| `rpi-state.json` absent | Silent no-op. No tweetable line in report. |
| `rpi-state.json` unparseable (invalid JSON) | Silent no-op. Log warning to stderr only. |
| `rpi-state.json` missing expected fields | Use available fields, default missing to safe values. |
| No historical state files | Streak = 1, total sessions = 1 (current only). |

## Tweetable Summary Format

When streak data is available, add this line to the TOP of the post-mortem report (before the verdict table):

```
> RPI streak: 5 consecutive days | Sessions: 14 | Last verdict: PASS
```

If no verdict history exists, omit the verdict portion:

```
> RPI streak: 1 consecutive days | Sessions: 1
```
