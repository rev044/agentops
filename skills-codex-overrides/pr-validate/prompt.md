# pr-validate

Validate PR work in Codex with findings-first output: isolation, upstream fit, regression risk, and scope creep.

## Codex Execution Profile

1. Treat `skills/pr-validate/SKILL.md` as the canonical PR validation contract and `skills-codex/pr-validate/SKILL.md` as the Codex-facing artifact.
2. Lead with concrete findings and missing-proof gaps that would matter to an upstream reviewer.
3. Keep validation aligned with the originally planned contribution boundary.

## Guardrails

1. Do not bury isolation or scope-creep problems in summary prose.
2. Prefer exact file references and missing-test callouts.
3. If the PR is not ready, say why in reviewer-facing terms.
