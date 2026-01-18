# Mail-to-Self Handoff Pattern

Using self-addressed mail for cross-session context.

## Overview

Mail-to-self is a pattern for passing arbitrary context to your future session.
Unlike beads (structured work items), mail can contain free-form instructions,
notes, or context that doesn't fit the bead model.

## Basic Pattern

```bash
# Send mail to yourself
gt mail send --self -s "HANDOFF: Topic" -m "Context here"

# Hook the mail for next session
gt hook attach <mail-id>
```

## When to Use

| Use Mail-to-Self | Use Bead Comments |
|------------------|-------------------|
| Ad-hoc instructions | Structured state |
| Priority shifts | Epic orchestration |
| Context notes | Progress tracking |
| One-time handoffs | Recurring patterns |

## Examples

### Priority Shift

```bash
gt mail send --self -s "HANDOFF: Priority shift" -m "$(cat <<'EOF'
Human requested: Focus on security fixes first.
Defer feature work until security audit complete.

Check: bd list --label=security
Start with: gt-sec-001 (highest priority)
EOF
)"
```

### Context Dump

```bash
gt mail send --self -s "HANDOFF: Auth feature" -m "$(cat <<'EOF'
## Working On
Implementing OAuth2 flow for ticket gt-abc.

## Key Files
- src/auth/oauth.py - Main logic (80% done)
- src/auth/tokens.py - Token management (done)
- tests/auth/test_oauth.py - Need to write

## Gotchas
- Token refresh has race condition (see oauth.py:142)
- Don't use the old session store (deprecated)

## Next Steps
1. Finish refresh token logic
2. Write tests
3. Update docs
EOF
)"
```

### Investigation Notes

```bash
gt mail send --self -s "HANDOFF: Bug investigation" -m "$(cat <<'EOF'
## Bug: gt-bug-123
Users seeing 500 errors on checkout.

## Investigation So Far
- Not database (queries normal)
- Not rate limiting (under threshold)
- Possibly payment gateway timeout

## Leads
- Check payment_gateway.py:process()
- Look at logs around 14:00 UTC
- Possibly related to gt-bug-100 (similar symptoms)

## Next
Add more logging to payment flow and reproduce.
EOF
)"
```

## Hooking Mail

Mail can be hooked just like beads:

```bash
# Find the mail ID
gt mail inbox
# → msg-abc123   self   HANDOFF: Auth feature   2026-01-08

# Hook it
gt hook attach msg-abc123

# Verify
gt hook
# → msg-abc123 (HANDOFF: Auth feature)
```

When the next session starts:

```bash
gt hook
# → msg-abc123

gt mail read msg-abc123
# → Full context displayed
```

## Hookable Mail Pattern

Combine mail with hook for reliable handoff:

```bash
# Create and hook in one flow
msg_id=$(gt mail send --self -s "HANDOFF: Work X" -m "Context" | grep -o 'msg-[a-z0-9]*')
gt hook attach $msg_id
gt handoff
```

Or use `gt handoff` with message directly:

```bash
gt handoff -s "Work X" -m "Context"
# Automatically sends mail and handles handoff
```

## Best Practices

### Structure Your Message

```markdown
## Context
What and why (1-2 sentences)

## Completed
- Item 1
- Item 2

## In Progress
- Current item (state: ...)

## Next
What to do next

## Files
Key files: path/file.py:42
```

### Be Specific

```bash
# Bad
gt mail send --self -s "HANDOFF" -m "Continue the work"

# Good
gt mail send --self -s "HANDOFF: Auth feature gt-abc" -m "$(cat <<'EOF'
Continue OAuth implementation.
Left off at: refresh token validation
File: src/auth/oauth.py:142
Next: Handle token expiry edge case
EOF
)"
```

### Include File References

```bash
-m "Key files changed:
- src/api/auth.py:89 (new endpoint)
- src/services/session.py:156 (cleanup logic)
- tests/api/test_auth.py (need to write)"
```

## Limitations

- Mail doesn't have structured fields (unlike beads)
- No dependency tracking
- No status transitions
- Can't be queried/filtered like beads

For structured work, use beads. For free-form context, use mail-to-self.

## See Also

- `gt mail send` - Mail command reference
- `gt hook attach` - Hooking mail
- `gt handoff` - Combined handoff command
