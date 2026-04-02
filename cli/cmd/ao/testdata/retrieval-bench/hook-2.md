---
type: learning
maturity: established
confidence: high
utility: 0.7
---
# Hook Authoring JSON Output Format

Structuring JSON output in hook authoring for Claude Code requires a specific envelope: `{"hookSpecificOutput": {"hookEventName": "...", "additionalContext": "..."}}`. Hook authoring JSON output that deviates from this format is silently ignored by the Claude Code runtime, making debugging extremely difficult. Always validate hook authoring JSON output with a schema check in the hook's own test suite before shipping.
