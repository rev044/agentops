# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.20.1] - 2026-03-07

### Fixed
- Codex install workflow now uses `~/.agents/skills` as the single raw skill home and stops recreating an AgentOps mirror in `~/.codex/skills`
- Native Codex plugin refresh now archives overlapping legacy `~/.codex/skills` AgentOps folders instead of repopulating them
- Codex install docs now consistently describe the `~/.agents/skills` workflow and the need for a fresh Codex session after install
- Codex skill conversion now preserves multiline YAML `description` fields correctly, fixing malformed generated metadata for skills such as Athena
- `ao doctor` now treats plugin-cache plus `~/.agents/skills` as the supported Codex layout and reports manifest drift with accurate wording

## [2.20.0] - 2026-03-05

### Added
- Flywheel loop closure — `ao session close --auto-extract` produces lightweight learnings and auto-handoff at session boundary
- Handoff-to-learnings bridge — `ao handoff` now extracts decisions into `.agents/learnings/` automatically
- Session-type scoring in `ao inject --session-type` — 30% boost for matching session context (career, debug, research, brainstorm)
- Identity artifact support — `ao inject --profile` surfaces `.agents/profile.md` in session context
- MEMORY.md auto-promotion in `ao flywheel close-loop` (Step 7) after maturity transitions
- Session-type detection in `ao forge` output metadata
- Production RPI orchestration engine — `ao rpi serve <goal>` with SSE streaming and auto mode
- Knowledge mining — `ao mine` and `ao defrag` commands for automated codebase intelligence
- Context declarations — `ao inject --for <skill>` reads skill frontmatter `context:` block for scoped retrieval
- Sections include allowlist and context artifact directories for skill-scoped injection
- `ao handoff` command for structured session boundary isolation
- Behavioral guardrails — 3-layer hook defense-in-depth (intent-echo, research-loop-detector, task-validation-gate)
- Context enforcement hook and run-id namespaced artifact paths
- Headless invocation standards and RPI phase runner
- Nightly CI athena job for automated knowledge warmup
- Coverage ratchet gate with BATS integration tests for shell scripts
- Fuzz targets, property tests, and golden file contracts for CLI
- Git worker guard, embedded parity gate, and swarm evidence validation hooks
- Release cadence gate warns on releases within 7 days of previous

### Changed
- Coverage floor raised to 84% for `cmd/ao`, average floor to 95%
- Complexity ceiling tightened to 20 (from 25)
- Default session-start hook mode switched from manual to lean
- Hard quality gate on injection — maturity + utility filter
- Post-mortem redesigned as knowledge lifecycle processor
- RPI god-file split — 1,363 lines reduced to 203 with structured handoff schema
- Legacy RPI orchestrator retired — serve now uses phased engine (-1,121 lines)
- Council V2 findings synthesized into agent instructions and skill contracts
- 10k LOC of coverage-padding tests deleted; 72 stale tests quarantined
- Skill hardening — web security controls across 5 skills, CSRF protection, crank pre-flight
- Session-end hook wires `ao session close --auto-extract` before existing forge pipeline

### Fixed
- Flywheel signal chain — confidence decay, close-loop ordering, glob errors
- Path traversal in context enforcement hook and frontmatter parsing
- Race condition in handoff consumption at session boundary
- `ao mine` stabilized — dedup IDs, error propagation, `--since` window, empty output guard
- Hook test assertions aligned with warn-then-fail ratchet pattern (strict env required)
- Pre-mortem gate exit code corrected to 2 in strict mode (was 1)
- RPI serve event pipeline and coherence gate hardened
- jq injection via bare 8-hex run IDs in serve classifier
- Goals parser edge cases — paired backtick strip and rune-aware truncation
- UTF-8 truncation across six functions converted to rune-safe slicing
- CORS headers and stale doc references cleaned up
- Cross-wave worktree file collisions prevented
- hookEventName added to hookSpecificOutput JSON schema

## [2.19.3] - 2026-02-27

### Changed
- README highlights `ao search` (built on CASS) — indexes all chat sessions from every runtime unconditionally; adds Second Brain + Obsidian vault section with Smart Connections local/GPU embeddings and MCP semantic retrieval

## [2.19.2] - 2026-02-27

### Fixed
- CHANGELOG retrospectively updated to document all v2.19.1 post-tag commits (skills namespace fixes were shipped but not recorded)

## [2.19.1] - 2026-02-27

### Fixed
- Quickstart skill rewritten from 275 lines to 68 lines — removes 90-line ASCII diagram and 50-line intent router that caused 3+ minute runtime; now outputs ~8 lines and completes in under 30 seconds
- `truncateText` edge case: maxLen 1–3 now returns `"..."[:maxLen]` instead of the original string unchanged
- Dead anti-pattern promotion functions removed from `ao maturity` (`promoteAntiPatternsCmd`, `filterTransitionsByNewMaturity`, `displayAntiPatternCandidates`, ~99 LOC)
- Windows file-lock and signal support — replace no-op `filelock_windows.go` with real `LockFileEx`/`UnlockFileEx` via kernel32.dll; extract `syscall.Flock` and `syscall.Kill` into platform-specific helpers so the binary compiles on Windows without POSIX-only syscalls
- `heal.sh` Check 7 false positive — script reference integrity check now strips URLs before pattern matching, preventing remote `https://…/scripts/foo.sh` references from being validated as local files
- Security gate `BLOCKED_HIGH` — three persistent findings resolved: gosec G118 false positive (context cancel func returned to caller), golangci-lint nolint syntax (space in `// nolint:` directive), radon double-counting `reverse_engineer_rpi.py` from `skills-codex/` copy
- 71 stale `ao know *` and `ao quality *` namespace references replaced across 17 `skills-codex/` SKILL.md files — agents running rpi/evolve/crank were invoking non-existent commands from the pre-flatten CLI namespace
- Three HIGH-severity stale command references fixed across `skills/` and `skills-codex/`: `ao flywheel status` → `ao metrics flywheel status`, `ao settings notebook update` → `ao notebook update`, `ao start seed/init` → `ao seed`/`ao init`

### Added
- Spec-consistency gate (`scripts/spec-consistency-gate.sh`) validates contract files before crank spawns workers
- Command-surface parity gate (`scripts/check-cmdao-surface-parity.sh`) ensures all CLI leaf commands are tested
- `scripts/post-merge-check.sh` now validates `go mod tidy` sync and blocks on symlinks
- `scripts/merge-worktrees.sh` now propagates file deletions and preserves permissions
- Post-mortem preflight script checks reference file existence before council runs
- Hooks.json preflight validates script existence
- Windows binaries added to GoReleaser and SLSA attestation subject list

### Changed
- Coverage floor raised 78% → 80% with CI enforcement gate; Codecov threshold aligned to 75%
- Six truncation functions converted to rune-safe Unicode slicing
- `truncateID` in pool.go delegates to shared `truncateText`
- Crank skill invokes spec-consistency gate before spawning workers
- Vibe skill carries forward unconsumed high-severity next-work items as pre-flight context
- Release skill warns on unconsumed high-severity next-work items
- next-work JSONL schema formalized to v1.2
- Skills installation switched from `npx skills` to native curl installer (`bash <(curl -fsSL …/install.sh)`)
- README updated with 5-command summary, compound effect section, and `/vibe` breakdown

## [2.19.0] - 2026-02-27

### Added
- `ao mind` command for knowledge graph operations.
- New RPI operator surfaces: normalized C2/event plumbing plus `ao rpi stream`, `ao rpi workers`, and tmux worker nudge visibility.
- Codex install/bootstrap improvements, including native `~/.codex/skills` install and one-line installer flow.
- Windows binaries added to GoReleaser build outputs.

### Changed
- CLI namespace migration completed and aligned across hooks, docs, integration tests, and generated command references.
- Codex skill system moved to regenerated modular layout with codex-specific overrides and runtime prompt tailoring.
- CI/release gates hardened (codex runtime sections, release e2e validation, parity checks, stricter policy enforcement).
- High-complexity CLI paths refactored (`runRPIParallel`, `runDedup`, `parseGatesTable`) to lower cyclomatic complexity.

### Fixed
- Multiple post-mortem remediation waves landed for CLI/RPI/swarm reliability and edge-case handling.
- Hook delegation and integration behavior corrected for flat command namespace.
- `heal.sh` false-positive behavior reduced and doctor stale-path detection improved.
- Skill/doc parity and cross-reference drift issues corrected across codex and core skill catalogs.

### Removed
- Legacy inbox/mail command surface and stale/dead skill references from active catalogs.

## [2.18.2] - 2026-02-25

### Fixed
- `ao seed` now creates `.gitignore` and storage directories — reuses `setupGitProtection`, `ensureNestedAgentsGitignore`, and `initStorage` from `ao init`
- `ao seed` text updated from stale `ao inject`/`ao forge` to current MEMORY.md + session hooks paradigm
- MemRL feedback loop closed — `ao feedback-loop` command wired, `ao maturity --recalibrate` dry-run guard added
- Quickstart skill updated to reference `ao seed` and current flywheel docs
- CLI reference regenerated after `ao feedback-loop` and seed help text changes

### Changed
- `.agents/` session artifacts removed from git tracking
- PRODUCT.md updated — Olympus section removed, value props and skill tier counts corrected
- GOALS.md coverage directive updated to measured 78.8% (target 85%)

## [2.18.1] - 2026-02-25

### Changed
- SessionStart hook default mode changed from `manual` to `lean` — flywheel injection now fires every session
- Auto-prune enabled by default (`AGENTOPS_AUTO_PRUNE` defaults to `1`, opt-out via `=0`)
- Anti-pattern detection threshold lowered from `harmful_count >= 5` to `>= 3`
- Eviction confidence threshold relaxed from `< 0.2` to `< 0.3`
- Maturity promotion threshold in `--help` text synced with code (`0.7` → `0.55`)

### Fixed
- Empty learnings no longer inflate flywheel metrics — extract prompt skips empty files, pool ingest rejects "no significant learnings" stubs
- `ao pool ingest` now runs automatically in session-end hook after forge (was manual-only)
- 8 stale doc/comment references to old thresholds updated across hooks, ENV-VARS.md, HOOKS.md, using-agentops skill
- 13 empty stub learnings removed from `.agents/learnings/`

## [2.18.0] - 2026-02-25

### Added
- `ao notebook update` command — compound MEMORY.md loop that merges latest session insights into structured sections
- `ao memory sync` command — sync session history to repo-root MEMORY.md with managed block markers for cross-runtime access (Codex, OpenCode)
- `ao seed` command — plant AgentOps in any repository with auto-detected templates (go-cli, python-lib, web-app, rust-cli, generic)
- `ao lookup` command — retrieve specific knowledge artifacts by ID or relevance query (two-phase complement to `ao inject --index-only`)
- `ao constraint` command family — manage compiled constraints (list, activate, retire, review)
- `ao curate` command family — curation pipeline operations (catalog, verify, status)
- `ao dedup` command — detect near-duplicate learnings with optional `--merge` auto-resolution
- `ao contradict` command — detect potentially contradictory learnings
- `ao metrics health` subcommand — flywheel health metrics (sigma, rho, delta, escape velocity)
- `ao context assemble` command — build 5-section context packet briefings for tasks
- Work-scoped knowledge injection: `ao inject --bead <id>` boosts learnings tagged with the active bead
- Predecessor context injection: `ao inject --predecessor <handoff-path>` surfaces structured handoff context
- Compact knowledge index: `ao inject --index-only` outputs ~200 token index table for JIT retrieval
- Learning schema extended with `source_bead` and `source_phase` fields for work-context tracking
- `ao extract --bead <id>` tags extracted learnings with the active bead ID
- Citation-to-utility feedback pipeline in flywheel close-loop (stage 5)
- Global `~/.agents/` knowledge tier for cross-repo learning sharing (0.8 weight penalty, deduped)
- Bead metadata resolver reads from env vars (`HOOK_BEAD_TITLE`, `HOOK_BEAD_LABELS`) or cache file
- Goal templates embedded in binary (go-cli, python-lib, web-app, rust-cli, generic) for `ao goals init --template` and `ao seed`
- Platform-specific process-group isolation for goal check timeouts (Unix: SIGKILL pgid, Windows: taskkill /T)
- SessionStart hook rewritten with 3 startup modes: lean (default), manual, legacy — via `AGENTOPS_STARTUP_CONTEXT_MODE`
- SessionEnd hook now gates notebook update and memory sync on successful forge
- Type 3 setup hook template: `hooks/examples/50-agentops-bootstrap.sh`
- Constraint compiler hook: `hooks/constraint-compiler.sh`
- Codex-native skill format (`skills-codex/`) with install and sync scripts for cross-runtime skill delivery
- Comprehensive cmd/ao test coverage push — 500+ tests across 5 waves reaching 79.2% statement coverage (13 untestable functions excluded)

### Changed
- SessionStart hook default mode changed from full inject to `lean` (extract + lean inject, shrinks when MEMORY.md is fresh)
- `ao flywheel close-loop` now applies ALL maturity transitions (not just anti-pattern)
- `ao hooks` generated config uses script-based commands instead of inline ao invocations
- `ao rpi` prefers epic-type issues before falling back to any open issue

### Fixed
- `truncateText` now uses rune-safe `[]rune` slicing to avoid breaking multi-byte UTF-8 characters
- `syncMemory` extracted from Cobra handler for testability
- `parseManagedBlock` detects duplicate markers and refuses to parse (prevents data loss)
- `readNLatestSessionEntries` warns on skipped unreadable session files
- `readSessionByID` detects ambiguous matches and returns error instead of first substring match
- `findMemoryFile` broad contains-fallback removed (was matching wrong projects)
- `pruneNotebook` iteration capped at 100 to prevent runaway loops
- `MEMORY_AGE_DAYS` sentinel initialized to -1 (was 0, causing false lean-mode activation when file missing)
- Lean-mode guard now requires `MEMORY_AGE_DAYS >= 0` before comparing freshness
- Memory sync moved inside forge success gate in session-end hook
- `ao search --json` returns `[]` (empty JSON array) when no results, instead of human-readable text
- `ao doctor` returns `DEGRADED` status for warnings without failures (previously only HEALTHY/UNHEALTHY)
- `ao rpi status` goroutine leak fix — signal channel properly cleaned up
- Inline rune truncation in `formatMemoryEntry` replaced with shared `truncateText`
- 6 new tests for dedup, ambiguity detection, iteration cap, duplicate markers
- Cobra pflag state pollution between test invocations — explicit flag reset in `executeCommand()` helper
- Goals validate.sh outdated checks and missing validate.sh for 7 skills
- 10 tech debt findings from ag-8km+ag-chm post-mortem (stale nudge, scanner, docs)
- ao binary codesigned with stable Mach-O identifier
- Hook integration tests updated — removed 8 stale standalone ao-* hook tests consolidated into session-end-maintenance.sh

## [2.17.0] - 2026-02-24

### Added
- GOALS.md (v4) OODA-driven intent layer — markdown-based goals format with mission, north/anti stars, and steerable directives
- `ao goals init` interactive GOALS.md bootstrap with `--non-interactive` mode
- `ao goals steer` command to add, remove, and prioritize directives
- `ao goals prune` command to remove stale gates referencing missing paths
- `ao goals migrate --to-md` converter from GOALS.yaml to GOALS.md format
- `ao goals measure --directives` JSON output of active directives
- `ao goals validate` reports format and directive count
- Format-aware `ao goals add` writeback (auto-detects md or yaml)
- Go markdown parser library with case-insensitive heading matching and round-trip rendering (26 tests)
- `/goals` skill rewritten with 5 OODA verbs (init/measure/steer/validate/prune)
- `/evolve` Step 3 rewritten with directive-based cascade for idle reduction

### Fixed
- `ao rpi` falls back to any open issue when no epic exists (#50)
- RPI phased processing tests added (~230 lines) for writePhaseResult, validatePriorPhaseResult, heartbeat, and registry directory

## [2.16.0] - 2026-02-23

### Added
- Evolve idle hardening — disk-derived stagnation detection, 60-minute circuit breaker, rolling fitness files, no idle commits
- Evolve `--quality` mode — findings-first priority cascade that prioritizes post-mortem findings over goals
- Evolve cycle-history.jsonl canonical schema standardization and artifact-only commit gating
- `heal-skill` checks 7-10 with `--strict` CI gate for automated skill maintenance
- 6-phase E2E validation test suite for RPI lifecycle (gate retries, complexity scaling, phase summaries, promise tags)
- Fixture-based CLI regression and parity tests
- `ao goals migrate` command for v1→v2 GOALS.yaml migration with deprecation warning (#48)
- Goal failure taxonomy script and tests

### Changed
- CLI taxonomy, shared resolver, skill versioning, and doctor dogfooding improvements (6 architecture concerns)
- GoReleaser action bumped from v6 to v7
- Evolve build detection generalized from hardcoded Go gate to multi-language detection

### Fixed
- `ao pool list --wide` flag and `pool show` prefix matching (#47)
- Consistent artifact counts across `doctor`, `badge`, and `metrics` (#46)
- Double multiplication in `vibe-check` score display (#45)
- Skills installed as symlinks now detected and checked in both directories (#44)
- Learnings resolved by frontmatter ID; `.md` file count in maturity scan (#43)
- JSON output truncated at clean object boundaries (#42)
- Misleading hook event count removed from display (#41)
- Post-mortem schema `model` field and resolver `DiscoverAll` migration
- 15+ missing skills added to catalog tables in `using-agentops`
- Handoff example filename format corrected to `YYYYMMDDTHHMMSSZ` spec
- Quickstart step numbering corrected (7 before 8)
- OpenAI docs skill: added Claude Code MCP alternative to Codex-only fallback
- Dead link to `conflict-resolution-algorithm.md` removed from post-mortem
- `ao forge search` → `ao search` in provenance and knowledge skills
- OSS docs: root-level doc path checks, removed golden-init reference
- Reverse-engineer-rpi fixture paths and contract refs corrected
- Crank: removed missing script refs, moved orphans to references
- Codex-team: removed vaporware Team Runner Backend section
- Security skill: bundled `security-gate.sh`, fixed `security-suite` path
- Evolve oscillation detection and TodoWrite→Task tools migration
- Wired `check-contract-compatibility.sh` into GOALS.yaml
- Synced embedded skills and regenerated CLI docs
