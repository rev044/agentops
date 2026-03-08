# release

Prepare releases for Codex with explicit boundaries: preflight gates, versioning steps, and tag-ready output.

## Codex Execution Profile

1. Treat `skills/release/SKILL.md` as the canonical release contract and `skills-codex/release/SKILL.md` as the Codex-facing artifact.
2. Keep the release boundary explicit: everything up to the tag, with validations and changelog evidence called out.
3. Prefer deterministic command sequences and clear rollback points over narrative release notes during execution.

## Guardrails

1. Do not blur preparation work with post-tag publishing tasks.
2. Keep version bumps, changelog generation, and release commits auditable.
3. Surface release blockers immediately when gates fail.
