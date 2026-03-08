# evolve

Run `$evolve` as an always-on Codex loop over `$rpi`: keep selecting productive work, compound via harvested follow-ups, and only go dormant as a last resort.

## Codex Execution Profile

1. Treat `skills/evolve/SKILL.md` as the canonical loop contract and `skills-codex/evolve/SKILL.md` as the Codex-facing artifact.
2. Use Codex commentary updates to show cycle boundaries, selection source, queue refreshes, and stop reasons.
3. Prefer Codex sub-agents for generator layers and sidecar audits when they can run in parallel without blocking the main loop.
4. Persist loop state under `.agents/evolve/` and recover from disk instead of relying on live context.

## Guardrails

1. Do not treat empty initial queues as success; run the full fallback ladder before dormancy.
2. Re-enter selection after every `$rpi` cycle and re-read harvested work immediately.
3. Keep kill-switch, regression gates, and stagnation protection active without short-circuiting useful work discovery.
4. If a Codex-native override and the source skill diverge, keep behavior aligned with the source contract and then update the override.
