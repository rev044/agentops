# Codex Parity Repair

Use this workflow when `skills/<name>/SKILL.md` is canonically correct but
`skills-codex/<name>/SKILL.md` has drifted into bad Codex UX after generation.

## Principles

1. `skills/<name>/SKILL.md` remains the canonical workflow contract.
2. `skills-codex/<name>/` is generated output and must not be hand-maintained.
3. Durable Codex-only body edits belong in `skills-codex-overrides/<name>/SKILL.md`.
4. Codex operator-layer prompt edits belong in `skills-codex-overrides/<name>/prompt.md`.

## Audit First

Run:

```bash
bash scripts/audit-codex-parity.sh
```

Or target one skill:

```bash
bash scripts/audit-codex-parity.sh --skill swarm
```

The audit flags the failure classes that the mechanical converter keeps
missing today:

- Claude-only backend references like `backend-codex-subagents.md` or `codex agents`
- duplicated runtime phrases created by blind search/replace

## Repair Loop

For each flagged skill:

1. Read `skills/<name>/SKILL.md` to confirm whether the canonical contract is correct.
2. Read `skills-codex/<name>/SKILL.md` to see the broken generated body.
3. Read `skills-codex-overrides/<name>/prompt.md` and `skills-codex-overrides/catalog.json`.
4. If the source contract is wrong, fix `skills/<name>/SKILL.md` first.
5. If the source is correct but Codex needs different body wording or tool semantics, create or update `skills-codex-overrides/<name>/SKILL.md`.
6. Re-run `bash scripts/sync-codex-native-skills.sh`.
7. Re-run validation:
   - `bash scripts/validate-codex-skill-parity.sh`
   - `bash scripts/validate-codex-generated-artifacts.sh --scope worktree`
   - `bash scripts/validate-codex-override-coverage.sh`

## LLM Repair Guidance

When doing the actual rewrite, the LLM should:

- preserve the behavior contract from `skills/<name>/SKILL.md`
- remove Claude-only primitive/tool names from the Codex body
- replace mechanical rewrites with real Codex-native instructions
- keep the final durable delta in `skills-codex-overrides/<name>/SKILL.md`, not in `skills-codex/<name>/SKILL.md`

If a skill keeps needing Codex-only body surgery, update
`skills-codex-overrides/catalog.json` so the treatment matches reality instead
of pretending the skill is still parity-only.
