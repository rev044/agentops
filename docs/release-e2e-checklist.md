# Release E2E Checklist

Use this checklist before tagging a release to verify the local gate and release-smoke paths.

## 1) Fast local gate (quick confidence)

Run:

```bash
bash scripts/ci-local-release.sh --fast --skip-e2e-install --jobs 4
```

Expect:

- Exit code `0`
- Output contains these markers:
  - `Codex runtime sections`
  - `Skill runtime parity`
  - `Hook install smoke (minimal + full)`
  - `ao init --hooks + ao rpi smoke`

Reference test: `tests/integration/test-release-e2e-validation.sh`.

## 2) Full local gate (pre-tag requirement)

Run:

```bash
bash scripts/ci-local-release.sh
```

Expect:

- Exit code `0`
- Final summary contains `LOCAL CI PASSED`

## 3) Codex runtime lint (focused check)

Run:

```bash
bash scripts/validate-codex-runtime-sections.sh
```

Use this when editing Codex runtime guidance or AGENTS runtime sections.

## 4) Hook install + init/RPI smoke expectations

The local gate includes these release E2E smoke checks:

- `Hook install smoke (minimal + full)` validates:
  - `ao hooks install`
  - `ao hooks show`
  - `ao hooks install --full --source-dir <repo-root> --force`
  - Hook artifacts are written (`~/.claude/settings.json` and `~/.agentops/hooks/session-start.sh`)
- `ao init --hooks + ao rpi smoke` validates (in a fresh git repo):
  - `ao init --hooks`
  - `ao rpi status`
  - `ao rpi --help`
  - `ao rpi phased --help`

## 5) Parity checks when workflow/docs contracts change

When changing CI workflow policy, hook/runtime docs, or required gate wording, run:

```bash
bash scripts/validate-ci-policy-parity.sh
bash scripts/validate-hooks-doc-parity.sh
bash scripts/validate-skill-runtime-parity.sh
bash scripts/validate-codex-runtime-sections.sh
bash scripts/validate-codex-install-bundle.sh
bash tests/docs/validate-doc-release.sh
```
