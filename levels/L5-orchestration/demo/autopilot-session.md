# Demo: /autopilot Session

A session showing autonomous epic execution with validation gates.

---

## Before

Epic created with `/plan`:

```
$ bd show agentops-epic-xyz
Epic: "Add real-time notifications"
Status: open
Blocks: 6 child issues

Issues:
  [P1] agentops-a1: Set up WebSocket server
  [P1] agentops-b2: Create notification service
  [P2] agentops-c3: Add browser client (blocked by a1)
  [P2] agentops-d4: Create notification types (blocked by b2)
  [P2] agentops-e5: Add user preferences (blocked by b2)
  [P3] agentops-f6: Integration tests (blocked by c3, d4, e5)
```

---

## Session

```
You: /autopilot agentops-epic-xyz

Claude: Loading epic...

**Epic:** Add real-time notifications
**Issues:** 6 total
**Waves:** 3

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Wave 1/3:** 2 issues

Launching:
  → agentops-a1: Set up WebSocket server
  → agentops-b2: Create notification service

[Sub-agents working...]

✓ agentops-a1 complete
✓ agentops-b2 complete

**Validation Gate:**
$ npm test
24 passed
$ npm run lint
No issues

✓ Wave 1 PASSED

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Wave 2/3:** 3 issues

Launching:
  → agentops-c3: Add browser client
  → agentops-d4: Create notification types
  → agentops-e5: Add user preferences

[Sub-agents working...]

✓ agentops-c3 complete
✓ agentops-d4 complete
✓ agentops-e5 complete

**Validation Gate:**
$ npm test
31 passed
$ npm run lint
1 warning: unused import in preferences.ts

Severity: LOW - Continuing...

✓ Wave 2 PASSED

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Wave 3/3:** 1 issue

Launching:
  → agentops-f6: Integration tests

[Sub-agent working...]

✓ agentops-f6 complete

**Validation Gate:**
$ npm test
45 passed (14 new integration tests)

✓ Wave 3 PASSED

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Epic Complete!**

Summary:
- 6/6 issues closed
- 3 waves executed
- 45 tests passing
- 1 low-severity warning (non-blocking)

Running /retro...

Learnings saved to: .agents/retros/2025-01-15-notifications-epic.md

$ bd sync && git push
```

---

## Checkpoint Example

If Wave 2 had a HIGH severity issue:

```
**Validation Gate:**
$ npm run security-scan
HIGH: Potential XSS in notification renderer

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⚠️  PAUSED - Human checkpoint required

Issue: Potential XSS vulnerability
File: src/notifications/renderer.ts:42
Severity: HIGH

Options:
  [continue] - Proceed anyway (not recommended)
  [fix]      - Fix the issue, then continue
  [abort]    - Stop autopilot, manual intervention

Your choice: fix

Claude: Fixing XSS issue...

[Sanitizes user content in renderer.ts]

$ npm run security-scan
No issues

Resuming autopilot...
```

---

## What You Learned

1. Autopilot chains waves automatically
2. Validation gates run between waves
3. Human checkpoints catch critical issues
4. Epic complete = all issues closed + retro run
5. Full autonomy with safety guardrails
