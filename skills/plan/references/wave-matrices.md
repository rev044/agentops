# Wave Matrices: File Conflicts and Cross-Wave Registries

> Extracted from plan/SKILL.md on 2026-04-11.
> File-level dependency matrix, cross-wave shared file registry, dependency necessity validation.

## File-Level Dependency Matrix (Mandatory)

Before assigning issues to waves, build a file-conflict matrix. For EACH issue, list which files it modifies. If any file appears in 2+ same-wave issues, either:
- **Serialize** them (move one to a later wave), or
- **Merge** them into a single issue assigned to one worker.

```markdown
## File-Conflict Matrix

| File | Issues |
|------|--------|
| `src/auth.go` | Issue 1, Issue 3 | ← CONFLICT: serialize or merge
| `src/config.go` | Issue 2 |
| `src/auth_test.go` | Issue 1 |
```

**Why:** Issue-level dependency graphs miss shared-file conflicts. In context-orchestration-leverage, two tracks both modified `rpi_phased_handoff.go` and required an unplanned Wave 2a/2b split. A file-conflict matrix would have caught this during planning.

## Cross-Wave Shared File Registry (Mandatory)

After computing waves, build a **cross-wave file registry** listing every file that appears in issues across different waves. These files are collision risks because later-wave worktrees are created from a base SHA that may not include earlier-wave changes.

```markdown
## Cross-Wave Shared Files

| File | Wave 1 Issues | Wave 2+ Issues | Mitigation |
|------|---------------|----------------|------------|
| `src/auth_test.go` | Issue 1 | Issue 5 | Wave 2 worktree must branch from post-Wave-1 SHA |
| `src/config.go` | Issue 2 | Issue 6 | Serial: Issue 6 blocked by Issue 2 |
```

**If any file appears in multiple waves:**
1. Ensure the later-wave issue explicitly declares a dependency on the earlier-wave issue that touches the same file (so `bd dep add` / `addBlockedBy` is set).
2. Flag the file in the plan's `## Cross-Wave Shared Files` section so `/crank` can enforce worktree base refresh between waves.
3. For test files shared across waves, prefer splitting test additions into the same wave as the code they test — avoid a separate "test coverage" issue that touches files already modified in an earlier wave.

**Why:** In na-vs9, Wave 2 agents started from pre-Wave-1 SHA. A Wave 2 test coverage issue overwrote Wave 1's `.md→.json` fix in `rpi_phased_test.go` because the worktree didn't include Wave 1's commit. The cross-wave registry makes these collisions visible during planning.

## Validate Dependency Necessity

For EACH declared dependency, verify:
1. Does the blocked issue modify a file that the blocker also modifies? → **Keep**
2. Does the blocked issue read output produced by the blocker? → **Keep**
3. Is the dependency only logical ordering (e.g., "specs before roles")? → **Remove**

False dependencies reduce parallelism. Pre-mortem judges will also flag these. In ol-571, unnecessary serialization between independent spec rewrites was caught by pre-mortem.
