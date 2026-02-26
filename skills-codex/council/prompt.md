# council

Multi-model consensus council. Spawns parallel judges with configurable perspectives. Modes: validate, brainstorm, research. Triggers: "council", "get consensus", "multi-model review", "multi-perspective review", "council validate", "council brainstorm", "council research".

## Instructions

Load and follow the skill instructions from `~/.codex/skills/council/SKILL.md`.
For `validate` on plans/specs/contracts, enforce the first-pass contract completeness gate (canonical ack sequence, crash-safe consume protocol, precedence truth-table + anomaly codes, and boundary failpoint conformance). Do not return `PASS` if any gate item is missing, contradictory, or not mechanically verifiable.
