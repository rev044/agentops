---
type: learning
maturity: established
confidence: high
utility: 0.75
---
# Hook Authoring Kill Switch Pattern

Every hook authoring script must check the kill switch as its very first instruction before any logic executes. The hook authoring kill switch pattern — checking `AGENTOPS_HOOKS_DISABLED=1` at line one — ensures operators can disable all hooks instantly without redeploying. Omitting the kill switch from hook authoring work is a CI violation and will cause the hook to block operations during incident response when fast disablement is critical.
