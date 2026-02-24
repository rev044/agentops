# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Work-scoped knowledge injection: `ao inject --bead <id>` boosts learnings tagged with the active bead
- Predecessor context injection: `ao inject --predecessor <handoff-path>` surfaces structured handoff context
- Learning schema extended with `source_bead` and `source_phase` fields for work-context tracking
- `ao extract --bead <id>` tags extracted learnings with the active bead ID
- Citation-to-utility feedback pipeline in flywheel close-loop (stage 5)
- Bead metadata resolver reads from env vars (`HOOK_BEAD_TITLE`, `HOOK_BEAD_LABELS`) or cache file
- SessionStart hook auto-wires `--bead` and `--predecessor` when Gas Town env vars are available
- Type 3 setup hook template: `hooks/examples/50-agentops-bootstrap.sh`

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
