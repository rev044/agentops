# CLI ↔ Skills/Hooks Wiring Map

> Which `ao` commands are called by which skills and hooks — and vice versa.

Auto-audited 2026-02-25. 52 CLI commands, 53 skills, 3 active runtime hooks.

## Summary

| Category | Count |
|----------|-------|
| CLI commands with skill/hook callers | 29 |
| Orphan commands (user utilities, hidden, CI-only) | 23 |
| Phantom subcommands (bugs) | 2 |

---

## CLI Commands → Callers

Every `ao` command that is actively called by at least one skill or hook.

| Command | Skill Callers | Hook Callers |
|---------|--------------|--------------|
| `ao know inject` | crank, evolve, implement, inject, learn, recover, research, retro | session-start.sh, worktree-setup.sh |
| `ao know forge` | extract, flywheel, forge, post-mortem, retro, vibe, evolve, crank | session-end-maintenance.sh |
| `ao work ratchet` | crank, handoff, implement, plan, pre-mortem, ratchet, rpi, status, vibe | ratchet-advance.sh, stop-auto-handoff.sh, prompt-nudge.sh, precompact-snapshot.sh |
| `ao work goals` | goals, evolve | — |
| `ao know search` | crank, inject, knowledge, plan, pre-mortem, provenance, research, using-agentops, vibe | session-start.sh |
| `ao work rpi` | council, crank, plan, quickstart, research, rpi, shared, swarm | — |
| `ao quality flywheel` | crank, evolve, flywheel, post-mortem, quickstart, retro, status | ao-flywheel-close.sh |
| `ao quality pool` | crank, learn, status | session-end-maintenance.sh |
| `ao know lookup` | crank, implement, inject, plan, research, using-agentops | session-start.sh |
| `ao context` | crank, implement, swarm | context-guard.sh |
| `ao quality maturity` | flywheel | session-end-maintenance.sh |
| `ao quality constraint` | flywheel, post-mortem, retro | — |
| `ao quality badge` | flywheel, status | — |
| `ao start seed` | quickstart | — |
| `ao settings notebook` | retro | session-start.sh |
| `ao settings memory` | — | session-end-maintenance.sh |
| `ao quality dedup` | flywheel | session-end-maintenance.sh |
| `ao quality contradict` | flywheel | session-end-maintenance.sh |
| `ao quality metrics` | flywheel | — |
| `ao extract` | extract | session-start.sh |
| `ao settings hooks` | quickstart | — |
| `ao start init` | quickstart | — |
| `ao work session` | post-mortem, retro | — |
| `ao temper` | post-mortem | — |
| `ao quality curate` | flywheel | — |
| `ao status` | flywheel, quickstart | — |
| `ao task-feedback` | retro | — |
| `ao task-status` | status | — |
| `ao promote-anti-patterns` | flywheel | — |

---

## Skills → Commands

Which `ao` commands each skill invokes.

| Skill | ao Commands Used |
|-------|-----------------|
| **crank** | `context assemble`, `flywheel close-loop`, `flywheel status`, `forge transcript`, `inject`, `lookup`, `pool list`, `ratchet record`, `ratchet status`, `rpi phased`, `search` |
| **evolve** | `forge`, `goals measure`, `inject` |
| **extract** | `extract`, `forge` |
| **flywheel** | `badge`, `constraint review`, `contradict`, `curate status`, `dedup`, `maturity`, `metrics cite-report`, `metrics health`, `promote-anti-patterns`, `status` |
| **forge** | `forge markdown`, `forge transcript` |
| **goals** | `goals add`, `goals drift`, `goals export`, `goals history`, `goals init`, `goals measure`, `goals meta`, `goals migrate`, `goals prune`, `goals steer`, `goals validate` |
| **handoff** | `ratchet status` |
| **implement** | `context assemble`, `lookup`, `ratchet record`, `ratchet skip`, `ratchet spec`, `ratchet status` |
| **inject** | `inject`, `lookup`, `search` |
| **knowledge** | `search` |
| **learn** | `inject`, `pool ingest`, `pool list`, `pool promote`, `pool stage` |
| **plan** | `ratchet record`, `rpi cleanup`, `rpi status`, `search` |
| **post-mortem** | `constraint activate`, `flywheel close-loop`, `forge`, `forge markdown`, `session close`, `temper validate` |
| **pre-mortem** | `ratchet record`, `search` |
| **provenance** | `search` |
| **quickstart** | `flywheel status`, `hooks install`, `hooks test`, `init`, `rpi phased`, `seed`, `status` |
| **ratchet** | `ratchet check`, `ratchet record`, `ratchet skip`, `ratchet status` |
| **recover** | `inject` |
| **research** | `inject`, `lookup`, `rpi phased`, `search` |
| **retro** | `constraint activate`, `constraint review`, `flywheel close-loop`, `forge`, `forge markdown`, `notebook update`, `session close`, `task-feedback` |
| **rpi** | `ratchet record`, `rpi cancel`, `rpi cleanup` |
| **status** | `badge`, `flywheel status`, `pool list`, `pool promote`, `pool stage`, `ratchet status`, `task-status` |
| **swarm** | `context assemble`, `rpi phased` |
| **using-agentops** | `lookup`, `search` |
| **vibe** | `forge markdown`, `ratchet record`, `search` |
| council | `rpi phased` |
| shared | `rpi phased` |

Skills with **no ao commands**: beads, brainstorm, bug-hunt, codex-team, complexity, converter, doc, heal-skill, inbox, openai-docs, oss-docs, pr-implement, pr-plan, pr-prep, pr-research, pr-retro, pr-validate, product, readme, release, reverse-engineer-rpi, security, security-suite, standards, trace, update.

---

## Hooks → Commands

Which `ao` commands each hook invokes.

| Hook File | Event | ao Commands |
|-----------|-------|-------------|
| **session-start.sh** | SessionStart | `extract`, `inject`, `lookup`, `notebook update`, `search` |
| **session-end-maintenance.sh** | SessionEnd | `contradict`, `dedup`, `forge transcript`, `maturity`, `memory sync`, `notebook update`, `pool ingest` |
| **ao-flywheel-close.sh** | Stop | `flywheel close-loop` |
| **ratchet-advance.sh** | PostToolUse | `ratchet record` |
| **context-guard.sh** | UserPromptSubmit | `context guard` |
| **prompt-nudge.sh** | UserPromptSubmit | `ratchet status` |
| **precompact-snapshot.sh** | PreCompact | `ratchet status` |
| **stop-auto-handoff.sh** | Stop | `ratchet status` |
| **worktree-setup.sh** | (worktree init) | `inject` |

Hooks with **no ao commands**: citation-tracker.sh, config-change-monitor.sh, constraint-compiler.sh, dangerous-git-guard.sh, git-worker-guard.sh, pending-cleaner.sh, pre-mortem-gate.sh, skill-lint-gate.sh, standards-injector.sh, stop-team-guard.sh, subagent-stop.sh, task-validation-gate.sh, worktree-cleanup.sh.

---

## Orphan Commands

Commands that exist in the Go CLI but are not called by any skill or hook. All are intentionally uncalled — user utilities, hidden internals, or CI-only.

| Command | Category | Notes |
|---------|----------|-------|
| `ao completion` | User utility | Shell completion generation |
| `ao settings config` | User utility | Config management |
| `ao start demo` | User utility | Interactive demonstration |
| `ao doctor` | CI/install | Called by install.sh and release-smoke-test.sh |
| `ao version` | User utility | Version query |
| `ao start quick-start` | User utility | `/quickstart` skill is the orchestrator |
| `ao quality vibe-check` | User utility | `/vibe` skill orchestrates directly |
| `ao inbox` | User utility | `/inbox` skill works independently |
| `ao mail` | User utility | Alias for inbox operations |
| `ao settings plans` | User utility | Plan management |
| `ao know trace` | User utility | Artifact tracing |
| `ao quality gate` | CI/test | Promotion gate — called in test scripts |
| `ao feedback` | Hidden | UI for providing feedback on learnings |
| `ao work feedback-loop` | Internal | Async feedback processing |
| `ao batch-feedback` | Hidden | Batch feedback processing |
| `ao work session-outcome` | Hidden | Session outcome recording |
| `ao store` | Hidden | Vector store management |
| `ao index` | Hidden | Indexing utility |
| `ao task-sync` | Hidden | Task synchronization |
| `ao migrate` | Hidden | Migration utility (`migrate memrl`) |
| `ao worktree` | Hidden | Worktree GC utility |
| `ao quality anti-patterns` | Hidden | Anti-pattern list |
| `ao assemble` | Alias | Also registered as `ao context assemble` |

---

## Phantom Subcommands (Bugs)

References to subcommands that don't exist under their parent command.

| Phantom Call | Location | Problem | Fix |
|-------------|----------|---------|-----|
| `ao quality gate check` | `tests/rpi-e2e/run-full-rpi.sh:172,176,217` | `check` is a subcommand of `ao work ratchet`, not `ao quality gate` | Change to `ao work ratchet check` |
| `ao know forge index` | `scripts/test-flywheel.sh:94` | `index` doesn't exist under `forge` (has: `transcript`, `markdown`, `batch`) | Change to `ao know forge markdown` |

---

## Session Lifecycle Flow

How hooks chain `ao` commands across a session:

```
Session Start
  → session-start.sh
      → ao extract (lean mode: extract + inject with auto-shrink)
      → ao know inject
      → ao know lookup (JIT knowledge retrieval)

During Session
  → ratchet-advance.sh (PostToolUse)
      → ao work ratchet record
  → context-guard.sh (UserPromptSubmit)
      → ao context guard
  → prompt-nudge.sh (UserPromptSubmit)
      → ao work ratchet status
  → citation-tracker.sh (PostToolUse)
      → writes .agents/ao/citations.jsonl (no ao command)

Session End
  → session-end-maintenance.sh
      → ao know forge transcript
      → ao settings notebook update
      → ao settings memory sync
      → ao quality pool ingest
      → ao quality maturity --expire --evict
      → ao quality dedup
      → ao quality contradict

Stop Event
  → ao-flywheel-close.sh
      → ao quality flywheel close-loop (citation → utility → maturity)
  → stop-auto-handoff.sh
      → ao work ratchet status (check for incomplete gates)

Pre-Compaction
  → precompact-snapshot.sh
      → ao work ratchet status (snapshot before context loss)
```

---

## Regenerating

When skills, hooks, or command usage changes, refresh this map as follows:

1. Re-scan source invocations in: `skills/*/SKILL.md`, `skills-codex/*/SKILL.md`, `hooks/*.sh`, `hooks/hooks.json`.
2. Update the relevant rows in this document.
3. Run `bash scripts/validate-hooks-doc-parity.sh` and ensure no stale hook-count wording remains.
4. Update the audit header date above.

## Maintaining This Document

Re-audit when:
- Adding a new `ao` CLI command (check it has skill/hook callers or is intentionally orphaned)
- Adding a new skill that calls `ao` commands (verify the commands exist)
- Adding a new hook that calls `ao` commands
- Running `scripts/generate-cli-reference.sh` (companion to this doc)
