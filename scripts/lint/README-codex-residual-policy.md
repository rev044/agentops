# Codex residual marker allowlist policy

This policy governs `scripts/lint/codex-residual-allowlist.txt`, the canonical machine-readable allowlist for residual mixed-runtime markers in `skills-codex/**/SKILL.md`.

## Purpose

`skills-codex` is Codex-first, but a small set of Claude markers is intentionally retained for mixed-runtime flows (`--mixed`) and runtime-native fallback documentation. The allowlist defines those exceptions explicitly so lint can fail on accidental runtime drift.

## Authoring rules

1. One POSIX ERE pattern per line in the allowlist file.
2. Keep patterns narrow and stable. Prefer exact phrases, backend IDs, reference filenames, and primitive names.
3. Do not use broad wildcards or generic vendor tokens (`Claude`, `claude`, `.*`, `.*claude.*`).
4. Every new entry must map to an intentional mixed-runtime contract in `skills-codex/**/SKILL.md`.
5. Remove entries once no longer referenced.

## Review checklist for allowlist changes

1. Is this marker required for mixed-runtime behavior?
2. Is the pattern as specific as possible?
3. Could this be expressed by an existing allowlisted marker?
4. Will this hide unintended Codex-to-Claude regressions?
