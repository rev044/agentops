# quickstart

Guide new users through AgentOps in a Codex-first flow.

## Codex Execution Profile

1. Treat `skills/quickstart/SKILL.md` as canonical workflow.
2. Prefer Codex tooling and command examples first.
3. Keep optional cross-runtime references brief and non-blocking.

## Guardrails

1. Do not require Claude CLI checks to proceed.
2. Avoid instructions that assume `.claude/` directories.
3. Be explicit that Codex CLI v0.115.0+ supports native hooks under `~/.codex/hooks.json`; older versions fall back to entry skills ensuring `ao codex ensure-start` once per thread and dedicated closeout skills ensuring `ao codex ensure-stop` once per thread.
4. Keep `ao codex status` as the manual lifecycle escape hatch, not the primary workflow.
5. Keep onboarding output action-oriented: next command, expected result, fallback.
