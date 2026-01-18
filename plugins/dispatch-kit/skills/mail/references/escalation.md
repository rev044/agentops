# Escalation Patterns

## Overview

Escalation is the process of surfacing issues that require human decision or intervention. Proper escalation ensures blockers are resolved while maintaining autonomous operation.

## When to Escalate

### Always Escalate

| Situation | Why |
|-----------|-----|
| Security concerns | Human must assess risk |
| Data loss risk | Human must authorize action |
| Production impact | Human accountability required |
| Missing credentials | Human must provide/configure |
| Out-of-scope requests | Human must expand or reject scope |
| Ambiguous requirements | Human must clarify intent |
| Repeated failures (3+) | Autonomous recovery not working |

### Consider Escalating

| Situation | Escalate If |
|-----------|-------------|
| Blocked by dependency | No workaround after 2 attempts |
| Test failures | Cannot fix without requirement change |
| Architecture conflict | Change affects multiple systems |
| Performance concern | Trade-off decision needed |

### Do NOT Escalate

| Situation | Instead |
|-----------|---------|
| Normal errors | Retry with backoff |
| Missing context | Read bead, search codebase |
| Build failures | Fix and retry |
| Merge conflicts | Resolve and continue |

## How to Escalate

### To Human

```bash
gt mail send --human -s "ESCALATION: <brief>" -m "$(cat <<'EOF'
## Issue
<bead-id>: <title>

## Problem
<What's blocking progress>

## Attempted
<What was tried>

## Need
<Specific decision or action required>

## Impact
<What's blocked if not resolved>
EOF
)" --urgent
```

**Example:**

```bash
gt mail send --human -s "ESCALATION: Missing API credentials for gt-abc" -m "$(cat <<'EOF'
## Issue
gt-abc: Integrate external payment service

## Problem
Cannot find API credentials for payment service in:
- Environment variables
- Config files
- Secrets management

## Attempted
1. Checked .env files across environments
2. Searched for PAYMENT_* in codebase
3. Reviewed secrets in Vault

## Need
Human to provide or configure payment service credentials

## Impact
- gt-abc blocked
- gt-def and gt-ghi depend on gt-abc
- Wave 2 cannot start
EOF
)" --urgent
```

### To Mayor (from Polecat/Witness)

```bash
gt mail send mayor/ -s "ESCALATION: <brief>" -m "<details>"
```

### Blocker Comment Pattern

When escalating, also add blocker comment to bead:

```bash
bd comments add gt-abc "BLOCKER: Missing API credentials. Escalated to human."
```

This allows convoy monitoring to detect blocked issues.

## Escalation Severity

| Severity | Subject Prefix | Priority | Notification |
|----------|---------------|----------|--------------|
| Critical | `ESCALATION:` | `--urgent` | Yes |
| High | `ATTENTION:` | `--priority 1` | Optional |
| Normal | `QUESTION:` | Default | No |

## Escalation Templates

### Missing Resources

```bash
gt mail send --human -s "ESCALATION: Missing <resource>" -m "$(cat <<'EOF'
Issue: <id>
Resource: <what's missing>
Searched: <where checked>
Need: <provision/configure/access grant>
EOF
)" --urgent
```

### Requirement Clarification

```bash
gt mail send --human -s "QUESTION: Clarification needed for <feature>" -m "$(cat <<'EOF'
Issue: <id>
Question: <specific question>
Options:
A) <option 1> - <implications>
B) <option 2> - <implications>
Recommendation: <if any>
EOF
)"
```

### Security Concern

```bash
gt mail send --human -s "ESCALATION: Security concern in <area>" -m "$(cat <<'EOF'
Issue: <id>
Concern: <security issue found>
Risk: <potential impact>
Evidence: <what was observed>
Recommendation: <suggested action>
EOF
)" --urgent
```

### Architecture Decision

```bash
gt mail send --human -s "QUESTION: Architecture decision for <feature>" -m "$(cat <<'EOF'
Issue: <id>
Decision: <what needs deciding>
Context: <background>
Options:
1) <approach 1>
   Pros: ...
   Cons: ...
2) <approach 2>
   Pros: ...
   Cons: ...
Recommendation: <if any>
Trade-offs: <what's being balanced>
EOF
)"
```

### Repeated Failure

```bash
gt mail send --human -s "ESCALATION: Repeated failure on <task>" -m "$(cat <<'EOF'
Issue: <id>
Failure: <what keeps failing>
Attempts: <count>
Error: <consistent error message>
Hypothesis: <suspected root cause>
Need: <human investigation/fix>
EOF
)" --urgent
```

## After Escalation

1. **Add blocker comment** - `bd comments add <id> "BLOCKER: <reason>"`
2. **Continue other work** - Don't wait, work on unblocked issues
3. **Check periodically** - Human may respond via mail

## Response Handling

When human responds:

```bash
# Check inbox
gt mail inbox

# Read response
gt mail read <msg-id>

# Remove blocker if resolved
bd comments add <id> "BLOCKER RESOLVED: <resolution>"

# Continue work
bd update <id> --status in_progress
```

## Best Practices

1. **Be specific** - "Need X" not "Something's wrong"
2. **Include context** - Bead ID, what was tried
3. **Suggest options** - Help human decide quickly
4. **Add blocker comment** - Enable convoy tracking
5. **Continue other work** - Don't block on escalation
6. **Follow up** - Check for response, resume when resolved
7. **Don't over-escalate** - Try autonomous resolution first

## Escalation Chain

```
Polecat → Witness → Mayor → Human
             ↓         ↓
         (local)   (strategic)
```

- **Polecat → Witness**: Session issues, resource problems
- **Witness → Mayor**: Rig-wide issues, pattern detection
- **Mayor → Human**: Strategic decisions, security, out-of-scope

## Crank Escalation Mode

During autonomous `/crank` execution:

```bash
# Blocker detected - escalate to HUMAN, not Mayor (we ARE Mayor in crank)
gt mail send --human -s "CRANK ESCALATION: $ISSUE_ID" -m "$(cat <<EOF
Issue: $ISSUE_ID
Epic: $EPIC_ID
Problem: $BLOCKER_REASON
Wave: $WAVE_NUM
Blocked children: $BLOCKED_COUNT
EOF
)" --urgent

# Mark and continue
bd comments add $ISSUE_ID "BLOCKER: $BLOCKER_REASON. Escalated."
# Continue with other issues in wave
```
