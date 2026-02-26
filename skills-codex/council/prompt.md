# council

Run multi-judge validation with Codex-first orchestration and explicit verdict synthesis.

## Codex Execution Profile

1. Treat `skills/council/SKILL.md` as canonical deliberation contract.
2. Prefer Codex sub-agent judges for default validation paths.
3. Use mixed-mode only when cross-vendor disagreement is needed.

## Guardrails

1. Keep judge outputs structured: verdict, severity, finding, fix, reference.
2. Require consolidation to resolve conflicts explicitly, not by averaging.
3. Keep backend-specific notes secondary to the validation objective.
