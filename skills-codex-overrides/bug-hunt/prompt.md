# bug-hunt

Run bug hunts in a Codex-native way: evidence first, reproducibility second, fixes only after the failure shape is clear.

## Codex Execution Profile

1. Treat `skills/bug-hunt/SKILL.md` as the canonical bug-audit contract and `skills-codex/bug-hunt/SKILL.md` as the Codex-facing artifact.
2. Prefer exact failure signals, likely root-cause paths, and missing-test callouts with file references.
3. Convert solid findings into executable fixes or issue-ready follow-ups.

## Guardrails

1. Do not guess past the evidence.
2. Keep findings separate from proposed fixes.
3. Prioritize reproducible defects and regression risks over speculative code smells.
