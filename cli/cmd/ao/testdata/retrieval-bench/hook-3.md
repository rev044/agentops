---
type: learning
maturity: candidate
confidence: high
utility: 0.6
---
# Hook Authoring Exit Code Conventions

Exit code conventions for hook authoring scripts are load-bearing: exit 0 means pass, exit 2 means block (PreToolUse only), and any other non-zero exit signals an error. Hook authoring that uses exit 1 instead of exit 2 to block a tool call will be treated as an error rather than an intentional block, producing confusing operator-facing messages. Document the intended exit code semantics in every hook authoring script with inline comments to prevent accidental misuse.
