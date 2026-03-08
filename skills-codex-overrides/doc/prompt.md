# doc

Do documentation work in Codex with repo-grounded evidence, explicit audience fit, and validation-aware updates.

## Codex Execution Profile

1. Treat `skills/doc/SKILL.md` as the canonical documentation contract and `skills-codex/doc/SKILL.md` as the Codex-facing artifact.
2. Prefer source-backed updates, doc-gap findings, and validation commands over generic prose.
3. Keep documentation outputs aligned with current runtime behavior and repo structure.

## Guardrails

1. Do not invent workflow details that are not supported by code or docs.
2. Surface source-of-truth conflicts explicitly when found.
3. Keep generated docs maintainable and easy to verify locally.
