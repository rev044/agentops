# handoff

Create Codex-native handoffs that survive compaction: exact state, exact files, exact next step.

## Codex Execution Profile

1. Treat `skills/handoff/SKILL.md` as the canonical handoff contract and `skills-codex/handoff/SKILL.md` as the Codex-facing artifact.
2. Capture the current objective, completed work, unresolved blockers, and the next command or file to inspect.
3. Prefer durable paths, issue ids, and validation evidence over conversational summaries.

## Guardrails

1. Do not leave the next session guessing what to do first.
2. Keep unresolved risks separate from completed facts.
3. When handoff structure changes for Codex, update this override rather than the generated prompt directly.
