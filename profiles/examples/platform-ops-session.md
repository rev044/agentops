# Example Session: Platform Operations

**Profile**: platform-ops
**Scenario**: P1 Incident - API latency spike
**Duration**: ~45 minutes

---

## Session Flow

### 1. Triage Phase (5 min)

```
Alert: API p99 latency > 2s (threshold: 500ms)
User: Help me investigate this latency spike
```

**Claude loads**: incident-responder

**Actions**:
- Acknowledge incident
- Check affected services
- Identify scope (all users vs subset)
- Start incident timeline

---

### 2. Investigation Phase (15 min)

```
User: Analyze logs for the affected service
```

**Claude loads**: error-detective, performance-engineer

**Actions**:
- Query application logs for errors
- Check database connection pool metrics
- Analyze request traces
- Identify bottleneck

**Finding**: Database connection pool exhausted

---

### 3. Mitigation Phase (10 min)

```
User: Apply temporary fix
```

**Claude loads**: change-executor

**Actions**:
- Increase connection pool size
- Restart affected pods
- Verify latency returning to normal
- Monitor for regression

**Result**: Latency back to normal

---

### 4. Root Cause Analysis (10 min)

```
User: What caused the connection pool exhaustion?
```

**Claude loads**: code-explorer, error-detective

**Actions**:
- Trace back to recent deployment
- Find unclosed database connections in new code
- Identify missing connection cleanup

**Root cause**: New feature missing `defer conn.Close()`

---

### 5. Postmortem Phase (5 min)

```
User: Create postmortem
```

**Claude loads**: incidents-postmortems

**Actions**:
- Document timeline
- Capture root cause
- Define action items
- Schedule follow-up review

**Output**: Blameless postmortem document

---

## Agents Used Summary

| Agent | When | Purpose |
|-------|------|---------|
| incident-responder | Triage | Initial response |
| error-detective | Investigate | Log analysis |
| performance-engineer | Investigate | Metrics analysis |
| change-executor | Mitigate | Apply fix |
| code-explorer | RCA | Find code issue |
| incidents-postmortems | Postmortem | Document learnings |

---

## Session Outcome

- ✅ Incident resolved in 30 min
- ✅ Root cause identified
- ✅ Postmortem completed
- ✅ Action items created

**MTTR**: 30 minutes (vs ~2 hours without structured approach)

---

## Follow-up Actions

1. [ ] Fix connection leak in code
2. [ ] Add connection pool monitoring alert
3. [ ] Review other services for similar patterns
4. [ ] Add runbook for connection pool issues
