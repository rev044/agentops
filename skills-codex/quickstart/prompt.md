# quickstart

Guide new users through AgentOps in a Codex-first flow.

## Codex Execution Profile

1. Treat `skills/quickstart/SKILL.md` as canonical workflow.
2. Prefer Codex tooling and command examples first.
3. Keep optional cross-runtime references brief and non-blocking.

## Guardrails

1. Do not require Claude CLI checks to proceed.
2. Avoid instructions that assume `.claude/` directories.
3. Be explicit that Codex has no startup or session-end hook surface under `~/.codex`.
4. Point Codex users to `ao codex start` after init and `ao codex stop` for closeout instead of implying hidden automation.
5. Keep onboarding output action-oriented: next command, expected result, fallback.
