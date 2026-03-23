# release

Prepare releases for Codex with explicit boundaries: preflight gates, versioning steps, and tag-ready output.


<!-- BEGIN AGENTOPS OPERATOR CONTRACT -->
<!-- Generated from skills-codex-overrides/catalog.json for release. -->

## Codex Execution Profile

1. Keep the release boundary explicit: everything up to the tag, with validations and changelog evidence called out.
2. Prefer deterministic command sequences and clear rollback points over narrative release notes during execution.
3. If this release ends a Codex hookless thread, inspect `.agents/ao/codex/state.json` and run `ao codex stop --auto-extract` only when `last_stop.session_id` does not match the current thread.

## Guardrails

1. Do not blur preparation work with post-tag publishing tasks.

<!-- END AGENTOPS OPERATOR CONTRACT -->
