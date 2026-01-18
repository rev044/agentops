---
name: operations
description: >
  Use when: "incident", "outage", "production issue", "debug", "logs", "error",
  "postmortem", "root cause", "triage", "escalation", "P0", "P1", "on-call",
  "mitigation", "rollback".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Operations Skill

Incident response, debugging, and postmortem patterns.

## Quick Reference

| Area | Key Patterns | When to Use |
|------|--------------|-------------|
| **Incident Response** | Triage, mitigation, communication | Production issues |
| **Structured Response** | Procedures, escalation | Incident management |
| **Postmortems** | Root cause, blameless | After incidents |
| **Error Detection** | Log analysis, patterns | Debugging |

---

## Incident Response

### Severity Levels

| Level | Impact | Response Time | Examples |
|-------|--------|---------------|----------|
| **P0** | Complete outage | Immediate | Site down, data loss |
| **P1** | Major functionality broken | < 1 hour | Core feature broken |
| **P2** | Significant issues | < 4 hours | Degraded performance |
| **P3** | Minor issues | Next business day | UI bugs, minor errors |

### Immediate Actions (First 5 Minutes)

1. **Assess Severity**
   - User impact (how many, how severe)
   - Business impact (revenue, reputation)
   - System scope (which services affected)

2. **Stabilize**
   - Identify quick mitigation options
   - Implement temporary fixes if available
   - Communicate status clearly

3. **Gather Data**
   - Recent deployments or changes
   - Error logs and metrics
   - Similar past incidents

### Investigation Protocol

```markdown
## Log Analysis
1. Start with error aggregation
2. Identify error patterns
3. Trace to root cause
4. Check cascading failures

## Quick Fixes (in order of preference)
1. Rollback if recent deployment
2. Increase resources if load-related
3. Disable problematic features
4. Implement circuit breakers

## Communication
- Brief status updates every 15 minutes
- Technical details for engineers
- Business impact for stakeholders
- ETA when reasonable to estimate
```

### Fix Implementation

1. Minimal viable fix first
2. Test in staging if possible
3. Roll out with monitoring
4. Prepare rollback plan
5. Document changes made

---

## Structured Incident Response

### Triage Checklist

```markdown
## Incident Triage: [Title]

### Initial Assessment
- **Reported**: [Time]
- **Severity**: [P0/P1/P2/P3]
- **Impact**: [Description]
- **Affected Systems**: [List]

### Investigation
- [ ] Check recent deployments
- [ ] Review error logs
- [ ] Check monitoring dashboards
- [ ] Identify root cause hypothesis

### Actions Taken
1. [Time] - [Action] - [Result]
2. [Time] - [Action] - [Result]

### Current Status
- **Status**: [Investigating/Mitigating/Resolved]
- **ETA**: [If known]
- **Next Update**: [Time]
```

### Escalation Matrix

| Condition | Escalate To |
|-----------|-------------|
| P0 incident | On-call lead + management |
| > 30 min no progress | Senior engineer |
| Customer-facing | Customer success |
| Security-related | Security team |
| Data loss risk | Data team + legal |

---

## Blameless Postmortems

### Postmortem Template

```markdown
# Postmortem: [Incident Title]

**Date**: [Date]
**Duration**: [Start] - [End] ([Total])
**Severity**: [P0/P1/P2/P3]
**Author**: [Name]

## Summary
[2-3 sentence description of what happened]

## Impact
- **Users Affected**: [Number/percentage]
- **Duration**: [How long users impacted]
- **Revenue Impact**: [If applicable]

## Timeline (all times in UTC)
| Time | Event |
|------|-------|
| 14:00 | First alert fired |
| 14:05 | On-call acknowledged |
| 14:15 | Root cause identified |
| 14:30 | Fix deployed |
| 14:45 | Service restored |

## Root Cause
[Technical explanation of what caused the incident]

## Contributing Factors
1. [Factor 1] - [How it contributed]
2. [Factor 2] - [How it contributed]

## What Went Well
- [Positive 1]
- [Positive 2]

## What Went Wrong
- [Issue 1]
- [Issue 2]

## Action Items
| Action | Owner | Priority | Due Date |
|--------|-------|----------|----------|
| [Action 1] | [Name] | P1 | [Date] |
| [Action 2] | [Name] | P2 | [Date] |

## Lessons Learned
[Key takeaways for the team]
```

### Root Cause Analysis

**5 Whys Technique:**
1. Why did the service go down? → Database connection exhausted
2. Why were connections exhausted? → Connection pool too small
3. Why was pool too small? → Default configuration used
4. Why was default used? → No load testing done
5. Why no load testing? → Not in deployment checklist

**Root Cause**: Missing load testing in deployment process

---

## Error Detection

### Log Analysis Commands

```bash
# Recent errors
grep -i "error\|exception\|fatal" /var/log/app.log | tail -100

# Error frequency
grep -i "error" /var/log/app.log | cut -d' ' -f1-2 | uniq -c | sort -rn

# Specific time range
awk '/14:00/,/14:30/' /var/log/app.log | grep -i error

# Unique error types
grep -oE "Error: [^]]*" /var/log/app.log | sort | uniq -c | sort -rn

# Correlation with deployment
git log --since="1 hour ago" --oneline
```

### Error Pattern Detection

| Pattern | Indicates | Action |
|---------|-----------|--------|
| Sudden spike | New deployment issue | Rollback |
| Gradual increase | Resource exhaustion | Scale up |
| Periodic errors | Cron job issues | Check schedules |
| Cascade failures | Dependency down | Check upstream |
| Random errors | Infrastructure issue | Check hardware |

### Debugging Checklist

```markdown
## Debugging: [Issue]

### Symptoms
- [ ] Error message: [text]
- [ ] Frequency: [how often]
- [ ] First seen: [when]
- [ ] Affected: [what/who]

### Hypotheses
1. [Hypothesis 1] - [How to test]
2. [Hypothesis 2] - [How to test]

### Investigation
| Hypothesis | Test | Result |
|------------|------|--------|
| [H1] | [Test] | [✅/❌] |

### Root Cause
[Confirmed cause]

### Fix
[How it was/will be fixed]
```

---

## Runbook Integration

### Alert → Runbook Pattern

```yaml
# Alert definition
alert: HighErrorRate
expr: error_rate > 0.05
labels:
  severity: warning
  runbook: docs/runbooks/high-error-rate.md
annotations:
  summary: Error rate above 5%
  runbook_url: https://wiki/runbooks/high-error-rate
```

### Runbook Template

```markdown
# Runbook: [Alert Name]

## Alert Details
- **Condition**: [What triggers this]
- **Severity**: [warning/critical]
- **Threshold**: [Value]

## Quick Actions
1. [First thing to try]
2. [Second thing to try]

## Diagnosis
```bash
# Check current state
[command]

# Check recent changes
[command]
```

## Resolution Steps
### If [Condition A]
[Steps to resolve]

### If [Condition B]
[Steps to resolve]

## Escalation
If unresolved after 30 minutes:
- Page: [team]
- Slack: [channel]

## Prevention
[How to prevent recurrence]
```

---

## Key Principles

1. **Stabilize first, investigate later** - Get service back up before deep diving
2. **Communicate frequently** - Silence breeds anxiety
3. **Document as you go** - Memory fades, notes don't
4. **Blameless culture** - Focus on systems, not people
5. **Learn and improve** - Every incident is a learning opportunity
