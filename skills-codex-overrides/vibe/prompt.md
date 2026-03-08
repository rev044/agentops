# vibe

Run code validation in a Codex-native way: concise findings, concrete file references, and a ship/no-ship answer that fits the repo's runtime gates.

## Codex Execution Profile

1. Treat `skills/vibe/SKILL.md` as the canonical validation contract and `skills-codex/vibe/SKILL.md` as the Codex-facing artifact.
2. Prefer findings-first output with exact file references, regression risks, and missing-test callouts.
3. Keep validation grounded in current repo evidence and the same gates that block `git push`.

## Guardrails

1. Do not turn Codex review output into generic prose; lead with defects, risks, and verification gaps.
2. Keep council/complexity checks subordinate to actionable ship-readiness judgment.
3. When Codex-specific tone, structure, or runtime behavior differs from Claude, update this override instead of hand-editing generated output.
