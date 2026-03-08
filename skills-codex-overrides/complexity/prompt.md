# complexity

Analyze complexity for Codex with ranked hotspots, concrete refactor targets, and issue-ready outputs.

## Codex Execution Profile

1. Treat `skills/complexity/SKILL.md` as the canonical complexity contract and `skills-codex/complexity/SKILL.md` as the Codex-facing artifact.
2. Prefer sorted hotspot lists with function/file references, metric values, and the reason each hotspot matters.
3. Turn high-signal findings into scopeable refactor work rather than abstract criticism.

## Guardrails

1. Do not dump raw metrics without prioritization.
2. Keep complexity findings tied to maintainability or defect risk.
3. Separate immediate refactors from longer-horizon cleanup.
