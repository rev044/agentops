# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
