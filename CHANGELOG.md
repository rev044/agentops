# Changelog

All notable changes to the AgentOps marketplace will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`bin/ralph`** — Full RPI loop script (Goal → Plan → Pre-mortem → Crank → Vibe → Post-mortem → PR). Each phase gets a fresh Claude context window (Ralph Wiggum pattern). Features: `--dry-run`, `--skip-pre-mortem`, `--branch`, `--spec` for acceptance criteria, `--resume` for checkpoint/resume, `--max-budget` and `--phase-timeout` for gutter detection.
- **`/codex-team` skill** — Spawn parallel Codex execution agents from Claude. Claude orchestrates task decomposition, Codex agents execute independently via `codex exec --full-auto`. Includes pre-flight checks, canonical command form, prompt guidelines, and fallback to `/swarm`.
- **`/codex-team` file-conflict prevention** — Team lead analyzes file targets before spawning: same-file tasks merge into one agent, dependent tasks sequence into waves with context injection, different-file tasks run in parallel. The orchestrator IS the lock manager.

### Changed

- **Codex model updated to `gpt-5.3-codex`** — All references across council, shared, and SKILL-TIERS updated from `gpt-5.3` to `gpt-5.3-codex` (canonical Codex model name).

## [1.6.0] - 2026-02-06

### Adoption Improvements

Driven by council analysis (3 judges + 6 explorers) and pre-mortem validation (2 judges, unanimous WARN → fixes applied).

#### README Overhaul

- **Tagline reframed** — "DevOps for AI agents" → "A knowledge flywheel for AI coding agents — your agent remembers across sessions." Leads with the differentiator (knowledge compounding), not the analogy
- **Tier table added** — Tier 0 (skills only) through Tier 3 (cross-vendor consensus) with graduation triggers. Uses "Tier" naming to avoid collision with existing L1-L5 learning path
- **What This Is reframed** — Flywheel narrative leads ("each session feeds the next"), ASCII diagram preserved in `<details>` block
- **Quick Start rewritten** — Self-contained with commands and context. `/quickstart` offered as optional guided tour (not primary path, due to known slash-command discoverability bug)
- **CLI Reference expanded** — MemRL retrieval, confidence decay, provenance tracking, escape velocity metrics. Leads with capabilities, not LOC count
- **"Why Agents Need DevOps" → "Why Agents Need This"** — Consistent with tagline reframe

#### Tier/Level Disambiguation

- **`docs/levels/README.md`** — Added "Tiers vs Levels" section explaining the two axes: Tiers (0-3) = what tools you install, Levels (L1-L5) = what concepts you learn. Cross-references README tier table
- **`skills/quickstart/SKILL.md`** — Added graduation hints to Step 7 based on detected CLI state (ao, beads presence). Natural language, not formal tier labels

### Native Teams Migration

**The big idea:** Council judges and swarm workers are no longer fire-and-forget background agents. They now spawn as teammates on native teams (`TeamCreate` + `SendMessage` + shared `TaskList`), enabling real-time coordination without re-spawning.

#### Council

- **Judges spawn as teammates** on a `council-YYYYMMDD-<target>` team instead of independent `Task(run_in_background=true)` calls
- **Debate R2 via SendMessage** — judges stay alive after R1 and receive other judges' full verdicts via `SendMessage`. No more re-spawning fresh R2 instances with truncated R1 summaries. Result: zero truncation loss, no spawn overhead, richer debate
- **Team cleanup** — `shutdown_request` each judge + `TeamDelete()` after consolidation
- **Communication rules** — judges message team lead only (prevents anchoring); no peer-to-peer, no TaskList access
- Updated architecture diagram with Phase 1a (Create Team) and Phase 3 (Cleanup)
- R2 output files unchanged (`.agents/council/YYYY-MM-DD-<target>-claude-{perspective}-r2.md`)

#### Swarm

- **Team-per-wave** — each wave creates a new team (`swarm-<epoch>`), preserving Ralph Wiggum fresh-context isolation
- **Workers as teammates** — workers join the wave team, claim tasks via `TaskUpdate`, and report via `SendMessage`
- **Retry via SendMessage** — failed workers receive retry instructions on their existing context (no re-spawn needed within a wave)
- **Workers access TaskList** — workers can claim and update their own tasks (previously Mayor had to reconcile everything)
- Step 5a added: team cleanup (`shutdown_request` workers + `TeamDelete`) after each wave

#### Crank

- **Diagram updated** to show swarm's team-based execution flow (`TeamCreate` per wave, `SendMessage` for reporting, `TeamDelete` after wave)
- Separation of concerns clarified: Crank = beads-aware orchestration, Swarm = team-based parallel execution

#### Shared

- **Native teams fallback** added to CLI availability/fallback table: if `TeamCreate` unavailable, fall back to `Task(run_in_background=true)` fire-and-forget
- Fallback degrades gracefully: council loses debate-via-message (reverts to R2 re-spawn with truncation), swarm loses retry-via-message (reverts to re-spawn)

### Hardening (ag-3p1)

Fixes from council validation of the native teams migration:

- **Codex model pre-flight** — council now tests model availability (not just CLI presence) before spawning Codex agents. Catches account-type restrictions (e.g. gpt-5.3-codex on ChatGPT accounts) and degrades to Claude-only
- **Debate fidelity marker** — debate reports include `**Fidelity:** full | degraded` so users know if `--debate` ran with full-context native teams or truncated fallback
- **Explicit R2 timeout** — `COUNCIL_R2_TIMEOUT` env var (default 90s) replaces vague "idle too long" with concrete timeout and fallback-to-R1 instruction
- **TeamDelete() documentation** — clarified that `TeamDelete()` targets the current session's team context; concurrent team scenarios (e.g. council inside crank) documented

### Simplification

Pre-release council validation (2 judges, unanimous WARN) identified over-engineering. Refactored before shipping:

- **Council task types 5 → 3** — merged critique→validate, analyze→research. Keeps validate, research, brainstorm
- **Removed `--perspectives-file`** — presets and `--perspectives="a,b,c"` cover all current use cases. Bring back when someone asks
- **Agent hard cap: MAX_AGENTS=12** — prevents resource bombs from `--mixed --deep --explorers=N` combinations. Pre-flight check errors if exceeded
- **`--debate` restricted to validate mode** — brainstorm and research don't produce PASS/WARN/FAIL verdicts; combining with --debate now errors instead of producing "awkward outputs"
- **`--debate` documented as Ralph exception** — judges intentionally persist across R1/R2 within one atomic invocation. Bounded, documented, justified
- **Distributed mode gated as experimental** — swarm and crank distributed mode (tmux + Agent Mail) labeled experimental. Local mode (native teams) is the recommended path
- **Crank validation simplified** — collapsed triple validation (per-task + per-issue + batched) to double (trust swarm + final batched vibe). Per-issue layer was redundant

### Documentation

- Added official Skills installer instructions: `npx skills@latest add boshu2/agentops --all -g`
- Added agent-scoped install example: `npx skills@latest add boshu2/agentops -g -a codex -s '*' -y`
- Clarified that session hooks are Claude Code plugin functionality (skills remain portable)

## [1.3.1] - 2026-02-01

### Documentation Reality Check

Swarm documentation updated to match tested behavior:

- **TaskCreate API**: Removed invalid `blockedBy` parameter from examples. Dependencies require separate `TaskUpdate(addBlockedBy=[...])` call
- **Terminology**: "crank loops" → "atomic agents" (agents don't loop internally)
- **Monitoring**: Replaced `TaskOutput` polling with automatic `<task-notification>` pattern
- **Agent isolation**: Documented that agents cannot access TaskList/TaskUpdate - Mayor must reconcile
- **Mayor reconciliation**: Added explicit verify → update status → spawn next wave step
- **Prompts**: Simplified from complex loop instructions to atomic task format

Meta-learning: Task decomposition matters. 6 "independent" doc tasks weren't independent - they shared a file. Consolidated to 2 truly parallel tasks.

## [1.3.0] - 2026-02-01

### Pure Claude-Native Swarm

**The big idea:** Why depend on tmux, external scripts, or complex tooling when Claude Code has everything we need built-in?

The `/swarm` skill now uses pure Claude Code primitives:
- `TaskCreate` / `TaskUpdate` / `TaskList` for state management
- `Task(run_in_background=true)` for spawning background agents
- `<task-notification>` for completion callbacks

No tmux sessions. No external scripts. No beads dependency. Just Claude Code.

### Ralph Wiggum Pattern

This release documents WHY the architecture works, based on the [Ralph Wiggum Pattern](https://ghuntley.com/ralph/):

```
Ralph's bash loop:          Our swarm:
while :; do                 Mayor spawns Task → fresh context
  cat PROMPT.md | claude    Mayor spawns Task → fresh context
done                        Mayor spawns Task → fresh context
```

**Key insight:** Each `Task(run_in_background=true)` spawn creates a fresh process with clean context. Making demigods loop internally would cause context to accumulate and degrade - violating Ralph's core principle.

The loop belongs in Mayor (orchestration). Fresh context belongs in demigods (work).

### Changed

- **`/swarm` skill** - Complete rewrite:
  - Removed tmux dependency
  - Removed external script requirements
  - Pure Task tool orchestration
  - Added Ralph Wiggum pattern documentation
  - Wave execution via `blockedBy` dependencies

- **L4-parallelization docs** - Modernized:
  - Updated from `/implement-wave` to `/swarm`
  - Added Ralph Wiggum pattern explanation
  - Demo uses TaskList/TaskUpdate flow

### Technical Details

The swarm loop:

1. Mayor calls `TaskList()` to find ready tasks (pending, no blockers)
2. For each ready task, Mayor spawns: `Task(run_in_background=true, ...)`
3. Claude sends `<task-notification>` when each agent completes
4. Mayor calls `TaskUpdate(status="completed")` for finished tasks
5. This unblocks dependent tasks → next wave becomes ready
6. Repeat until all tasks complete

Each demigod has fresh context. Mayor maintains state via TaskList. Files/commits persist work across spawns.

## [1.2.0] - 2026-01-31

### Parallel Wave Execution

**The big idea:** When you have multiple issues that can run in parallel (no dependencies between them), why run them one at a time?

Before v1.2.0, `/crank` executed issues sequentially - finish one, start the next. Fine for small epics, but painfully slow when you have 10 independent tasks that could run simultaneously.

Now `/crank` detects **waves** - groups of issues with no blockers - and executes them in parallel using subagents. Each issue gets its own isolated agent. Results flow back to the main session.

```
Before (sequential):
  issue-1 → done → issue-2 → done → issue-3 → done
  Time: 3x

After (parallel waves):
  Wave 1: [issue-1, issue-2, issue-3] → 3 subagents in parallel → all done
  Time: 1x
```

**Why max 3 agents per wave?** Context management. Each subagent returns results that accumulate in your session. We tested higher parallelism - context explodes on complex issues. 3 is the sweet spot: meaningful speedup without blowing your context budget.

### How Waves Work

Waves emerge naturally from beads dependencies:

1. **`/plan`** creates issues with `blocks` dependencies
2. Issues with NO blockers = Wave 1 (run in parallel)
3. Issues blocked by Wave 1 = Wave 2 (run after Wave 1 completes)
4. **`bd ready`** returns the current wave - all unblocked issues
5. **`/crank`** takes the wave and dispatches up to 3 subagents

The dependency graph IS your execution plan. No separate "wave configuration" needed.

### Full Pipeline

```
/research → understand the problem
     ↓
/plan → decompose into issues with dependencies
     ↓         (waves form automatically)
/crank → execute waves in parallel
     ↓         Wave 1: [a, b, c] → 3 agents
     ↓         Wave 2: [d, e] → 2 agents
     ↓         Wave 3: [f] → 1 agent
     ↓
/post-mortem → extract learnings
```

### What's Next: Olympus

This parallel wave model is designed for **single-session work** - one Claude session spawning subagents. It's the foundation for something bigger.

**Olympus** (coming soon) will handle true multi-session orchestration: separate Claude sessions, persistent workers, direct context management instead of subagent nesting. The beads dependency graph persists across sessions - that's the ratchet that survives context resets.

### Changed

- **`/crank` skill** - Parallel wave execution:
  - Added `MAX_PARALLEL_AGENTS = 3` limit per wave
  - Step 4 now dispatches subagents in parallel via Task tool
  - FIRE loop updated to show wave model
  - `bd ready` explicitly documented as "returns current wave"

- **`/plan` skill** - Explicit wave formation:
  - Step 7 now shows how to create `blocks` dependencies
  - Added explanation of how waves form from dependencies
  - Clarified that `bd ready` returns parallelizable work

- **L4 implement-wave docs** - Updated max from 8 to 3 agents per wave

### Technical Details

The key instruction for `/crank`:

> **All Task calls for a wave MUST be in a single message to enable parallel execution.**

When Claude sends multiple Task tool calls in one message, they execute concurrently. Sequential messages = sequential execution. This is how we get parallelism without external orchestration.

## [1.1.0] - 2026-01-26

### Added
- **Agent Farm** (`/farm` skill) - Parallel multi-agent execution:
  - `ao farm validate` - Pre-flight checks with cycle detection
  - `ao farm start --agents N` - Spawn N agents + witness in tmux sessions
  - `ao farm status` - Check farm progress and agent states
  - `ao farm stop` - Graceful shutdown with process cleanup
  - `ao farm resume` - Resume incomplete farm from metadata
- **Witness monitoring** - Background observer for agent farm:
  - `ao witness start` - Start witness process
  - `ao witness stop` - Stop witness
  - `ao witness status` - Check witness state
- **Agent messaging** - Communication between agents:
  - `ao inbox` - View messages from agents
  - `ao mail send --to <agent> --body <message>` - Send message to agent
- **Serial agent spawn** with 30s stagger (rate limit protection)
- **Circuit breaker** - Stops farm if >50% agents fail
- `prompts/witness_prompt.txt` - Witness agent prompt template

### Changed
- Updated `using-agentops` skill documentation to include `/farm`
- Bumped skill count to 22

## [0.4.0] - 2026-01-25

### Changed
- **Repository restructure** - Professional polish for cleaner organization:
  - Reduced root directories from 22 to 13
  - Consolidated `levels/`, `profiles/`, `reference/`, `templates/`, `workflows/` into `docs/`
  - Renamed `shared/` to `lib/`
  - Deleted `mail/` (empty) and `agents-archived/` (56 obsolete agents)

- **README rewrite** - Minimal and approachable (47 lines vs 350):
  - One install command, 4 key skills, "want more?" section
  - Moved all details to `docs/PLUGINS.md`
  - Progressive disclosure: start simple, discover more as needed

- **Plugin description** - Simplified from verbose to concise:
  - Old: "Complete Knowledge OS for Claude Code - Research/Plan/Implement workflow..."
  - New: "Plugin kits for Claude Code: RPI workflow, validation, multi-agent orchestration"

### Added
- **Thin commands** - 4 command files that delegate to skills:
  - `commands/research.md` → `solo-kit:research`
  - `commands/plan.md` → `core-kit:formulate`
  - `commands/execute.md` → `core-kit:crank`
  - `commands/validate.md` → `vibe-kit:vibe`

- **Session hooks** - `hooks/` directory with:
  - `hooks.json` - SessionStart hook configuration
  - `session-start.sh` - Creates `.agents/` directories, outputs context

- **Multi-platform support**:
  - `.codex/setup.md` - Codex installation instructions
  - `.opencode/setup.md` - OpenCode installation instructions

- **RELEASE-NOTES.md** - User-friendly version highlights

- **docs/PLUGINS.md** - Complete plugin catalog moved from README

- **Marketplace cleanup** - Removed email from author fields, use GitHub username instead

## [0.3.1] - 2026-01-24

### Changed
- **Standardized .agents/ paths** (core-kit v0.2.1, pr-kit v0.1.1) - All skills now use relative `.agents/` paths:
  - Removed `~/gt/.agents/<rig>/` pattern in favor of portable `.agents/`
  - Removed "Phase 0: Rig Detection" sections from all skills
  - Skills affected: research, plan, formulate, product, pre-mortem, retro, post-mortem, implement
  - PR skills affected: pr-research, pr-plan, pr-implement, pr-retro
  - Gas Town-specific skills (gastown-kit, dispatch-kit) retain their specialized paths

- **README mermaid diagrams** - Replaced ASCII art with GitHub-native mermaid:
  - RPI Workflow diagram: `/research → /pre-mortem → /formulate → /crank → /post-mortem`
  - Plan → Crank handoff diagram with pre-mortem and post-mortem
  - Upgrade Path diagram

### Added
- **RAG Formatting Standard** (domain-kit) - New reference for knowledge artifacts:
  - `standards/references/rag-formatting.md` - 200-400 char sections, frontmatter conventions
  - Knowledge Artifact Detection section in standards SKILL.md
  - No `confidence` column rule (query-time, not storage-time)

- **RAG references added** (core-kit) - Knowledge-producing skills now reference RAG standard:
  - research, plan, formulate, pre-mortem, retro, post-mortem

### Fixed
- **retro skill** - Removed incorrect "Confidence" column from Discovery Provenance template:
  - Confidence/relevance are query-time metrics, not storage-time
  - Added reference to RAG formatting standard

## [0.2.3] - 2026-01-24

### Fixed
- **Plugin JSON uniformity** - Standardized all 14 plugin.json files:
  - Added `$schema` to all plugins (was missing from all)
  - Added `license: "MIT"` to 9 plugins that were missing it
  - Added `keywords` array to all plugins for discoverability
  - All plugins now have identical field structure

## [0.2.2] - 2026-01-24

### Added
- **marketplace-release skill** (core-kit v0.1.2) - New skill for releasing Claude Code plugins:
  - Complete release workflow documentation
  - Version bumping guidance
  - Update propagation explanation
  - Common pitfalls and anti-patterns
  - Context mode reference (inline vs fork)

## [0.2.1] - 2026-01-24

### Fixed
- **Marketplace plugin skills** - Applied `context: inline` fix to distributed plugins:
  - `core-kit/crank` (v0.1.1) - Epic execution now sees conversation context
  - `vibe-kit/vibe` (v0.1.2) - Validation now sees conversation context
  - `general-kit/vibe` (v0.1.2) - Validation now sees conversation context
  - Users who install from marketplace now get the fix

## [0.2.0] - 2026-01-24

### Fixed
- **Skill context mode** - Changed `context: fork` to `context: inline` for conversation-aware skills:
  - `vibe` - Now has access to chat context for inferring validation targets
  - `crank` - Now can identify epics mentioned in conversation
  - `pre-mortem` - Now can analyze specs discussed in chat
  - `post-mortem` - Now can identify completed epics from conversation
  - Root cause: `context: fork` creates isolated execution without conversation history
  - See `.agents/patches/2026-01-24-skill-context-inline.md` for details

## [0.1.3] - 2026-01-21

### Added
- **Two-Tier Standards Architecture** - JIT loading strategy for language standards:
  - **Tier 1** (slim refs, ~4-5KB): Always loaded via standards skill
  - **Tier 2** (deep standards, ~15-25KB): Loaded with `--deep` flag
  - Languages: Python, TypeScript, Shell, Go, YAML, JSON, Markdown

- **domain-kit v0.1.1** - Tier 1 slim references:
  - `standards/references/python.md` - Quick reference, common errors, prescan checks
  - `standards/references/typescript.md` - Strict mode, ESLint, type patterns
  - `standards/references/shell.md` - Required flags, shellcheck, error handling
  - `standards/references/go.md` - Error patterns, interfaces, concurrency
  - `standards/references/yaml.md` - yamllint, Helm/Kustomize patterns
  - `standards/references/json.md` - Formatting, JSONL, schema validation
  - `standards/references/markdown.md` - AI optimization, structure, tables

- **vibe-kit v0.1.1** - Tier 2 deep standards:
  - `vibe/references/python-standards.md` - Full complexity patterns, compliance grading
  - `vibe/references/typescript-standards.md` - Discriminated unions, branded types
  - `vibe/references/shell-standards.md` - ERR traps, security patterns
  - `vibe/references/go-standards.md` - Custom errors, thread-safe patterns
  - `vibe/references/yaml-standards.md` - Full Helm/Kustomize conventions
  - `vibe/references/json-standards.md` - Configuration patterns, tooling
  - `vibe/references/markdown-standards.md` - AI-agent optimization principles

- **general-kit v0.1.1** - Tier 2 deep standards (zero-dependency version):
  - Same 7 `*-standards.md` files as vibe-kit
  - Standalone operation without beads integration

### Changed
- **vibe SKILL.md** (vibe-kit, general-kit) - Added "Two-Tier Standards Loading" documentation:
  - Explains Tier 1 vs Tier 2 loading behavior
  - Documents `--deep` flag for comprehensive audits
  - Usage examples for different scenarios

### Design Decisions
- **Progressive disclosure**: Tier 1 gives quick answers, Tier 2 provides comprehensive audit capability
- **Context efficiency**: Default validation stays under 40% context budget
- **Portable**: general-kit has same deep standards for zero-dependency environments

## [0.1.2] - 2026-01-20

### Added
- **Tiered Architecture** - Scalable plugin system from solo developer to multi-agent orchestration:
  - **Tier 1**: solo-kit (any developer, any project)
  - **Tier 2**: Language kits (plug in based on project)
  - **Tier 3**: Team workflows (beads-kit, pr-kit, dispatch-kit)
  - **Tier 4**: Multi-agent orchestration (crank-kit, gastown-kit)

- **solo-kit v0.1.2** - Foundation for any developer:
  - 7 skills: `/research`, `/vibe`, `/bug-hunt`, `/complexity`, `/doc`, `/oss-docs`, `/golden-init`
  - 2 agents: `code-reviewer`, `security-reviewer` (read-only review specialists)
  - Hooks: auto-format on save, console.log/debug warnings, git push review, debug audit on session end
  - Zero external dependencies - works with any project

- **python-kit v0.1.2** - Python language support:
  - Standards skill with `references/python.md`
  - Hooks: ruff format, ruff check, mypy type checking

- **go-kit v0.1.2** - Go language support:
  - Standards skill with `references/go.md`
  - Hooks: gofmt, golangci-lint, P13/P14 error handling checks

- **typescript-kit v0.1.2** - TypeScript/JavaScript support:
  - Standards skill with `references/typescript.md`
  - Hooks: prettier, tsc type checking, `any` type warnings

- **shell-kit v0.1.2** - Shell scripting support:
  - Standards skill with `references/shell.md`
  - Hooks: shellcheck, `set -euo pipefail` enforcement

- **ARCHITECTURE-PROPOSAL.md** - Documents the tiered architecture design and migration path

### Changed
- **README.md** - Major update with tiered architecture:
  - Added tiered install instructions
  - Added upgrade path diagram (solo-kit → language-kit → beads-kit → crank-kit → gastown-kit)
  - Clarified legacy plugins and migration targets

- **Argument Inference** - Enhanced `/crank` and `/vibe` to semantically infer targets:
  - `/crank creating beads` now extracts "beads" keyword and searches for matching epic
  - `/vibe the auth changes` now validates auth-related files from git diff
  - Priority: conversational keywords > beads/git discovery > ask user
  - Updated in: core-kit/crank, vibe-kit/vibe, general-kit/vibe

### Skill Counts
| Kit | Skills | Agents | Tier |
|-----|--------|--------|------|
| solo-kit | 7 | 2 | 1 |
| python-kit | 1 (standards) | - | 2 |
| go-kit | 1 (standards) | - | 2 |
| typescript-kit | 1 (standards) | - | 2 |
| shell-kit | 1 (standards) | - | 2 |

---

## [0.1.1] - 2026-01-20

### Added
- **general-kit v1.0.0** - Standalone plugin with zero dependencies:
  - `/research`, `/vibe`, `/vibe-docs`, `/bug-hunt`, `/complexity`, `/validation-chain`
  - `/doc`, `/oss-docs`, `/golden-init`
  - 4 expert agents: security-expert, architecture-expert, code-quality-expert, ux-expert
- **standards library skill** (domain-kit) - Language-specific validation rules:
  - Python, Go, TypeScript, Shell, YAML, Markdown, JSON references
  - OpenAI platform standards (prompts, functions, responses, reasoning, GPT-OSS)
- **Context inference** for vibe and crank skills - Auto-detect targets from conversation
- **Natural language triggers** - Skills activate from intent, not just slash commands

### Changed
- **README overhaul**:
  - Added ASCII art logo and workflow diagrams
  - "Just Talk Naturally" section showing intent-based triggering
  - "The Killer Workflow: Plan → Crank" section with Shift+Tab + /formulate pattern
  - Clarified this provides plugins FOR beads/gastown, not built on them
  - Added OpenCode compatibility section
- **vibe skill** - Now references standards library for language-specific validation
- **validation-chain skill** - Added standards dependency
- **vibe-docs skill** - Added standards dependency

### Fixed
- **Standards dependencies** - Added missing `standards` skill dependency to:
  - vibe (vibe-kit, general-kit)
  - validation-chain (vibe-kit, general-kit)
  - vibe-docs (vibe-kit, general-kit)
- **Vibe findings** - Addressed quality findings across all plugins
- **Cross-skill references** - Test validator now handles relative paths correctly
- **Personal identifiers** - Removed from public plugin files

---

## [0.1.0] - 2026-01-19

### Added
- **Unix Philosophy Restructure** - Plugins reorganized into 8 focused kits:
  - **core-kit v1.0.0** - Workflow: research, plan, formulate, product, implement, implement-wave, crank, retro
  - **vibe-kit v2.0.0** - Validation only: vibe, vibe-docs, validation-chain, bug-hunt, complexity (+ 4 expert agents)
  - **docs-kit v1.0.0** - Documentation: doc, oss-docs, golden-init
  - **beads-kit v1.0.0** - Issue tracking: beads, status, molecules
  - **dispatch-kit v1.0.0** - Orchestration: dispatch, handoff, roles, mail
  - **pr-kit v1.0.0** - OSS contribution: pr-research, pr-plan, pr-implement, pr-validate, pr-prep, pr-retro
  - **gastown-kit v1.0.0** - Gas Town: crew, polecat-lifecycle, gastown, bd-routing
  - **domain-kit v1.0.0** - Reference knowledge: 18 domain skills (languages, development, security, etc.)

### Changed
- **vibe-kit** - Slimmed down from 23 skills to 5 focused validation skills
- **gastown plugin** - Replaced by gastown-kit (Gas Town specific) + pr-kit (contribution workflow)
- **Main README** - Updated with Unix philosophy structure, recommended combinations, clearer skill guidance
- **Core kit README** - Added decision trees for implement vs crank vs implement-wave

### Removed
- **gastown plugin** - Split into gastown-kit and pr-kit for better modularity

### Fixed
- **vibe-kit missing skills** - Restored vibe and vibe-docs skills that were lost during restructure

### Consolidated
- **domain-kit v1.1.0** - Consolidated from 18 to 17 skills:
  - Removed `doc-curator` (redundant with docs-kit/doc)
  - Consolidated 7 `base/` utilities (audit-diataxis, audit-onboarding, audit-workflow, cleanup-deprecated, cleanup-docs, cleanup-plans, cleanup-repo) into single `maintenance` skill

### Skill Counts (Final)
| Kit | Skills |
|-----|--------|
| core-kit | 8 |
| vibe-kit | 5 |
| docs-kit | 3 |
| beads-kit | 3 |
| dispatch-kit | 4 |
| pr-kit | 6 |
| gastown-kit | 4 |
| domain-kit | 17 |
| **Total** | **50** |

---

### Previous Unreleased

#### Added
- **vibe-kit v1.1.0** - New skills added:
  - `implement-wave` - Parallel execution of multiple issues
  - `complexity` - Code complexity analysis using radon/gocyclo
  - `doc` - Documentation generation and validation
  - `oss-docs` - OSS documentation scaffolding (README, CONTRIBUTING, SECURITY)
  - `golden-init` - Repository initialization with Golden Template
  - `molecules` - Workflow templates and formula TOML patterns
- **Skills sync** - All skills updated to match latest local versions:
  - beads, bug-hunt, dispatch, implement, research, vibe, vibe-docs (vibe-kit)
  - All 18 gastown plugin skills updated

### Fixed
- **Painted doors removed** - Cleaned up non-functional references:
  - Removed empty `references/` directories (bug-hunt, implement, pr-research, pr-retro)
  - Fixed pr-research template reference to point to inline section

### Changed
- **Commands deprecated** - Commands directory marked as deprecated in favor of skills
  - Added deprecation notice to commands/INDEX.md
  - Added migration guide pointing to skill equivalents
  - Commands maintained for legacy compatibility only
- **vibe-kit plugin.json** updated to version 1.1.0 with new skills

### Previous Unreleased

- **vibe-check Integration** in session-management plugin
  - `/session-start` now captures baseline metrics via `vibe-check session start`
  - `/session-end` now captures session metrics and failure patterns via `vibe-check session end`
  - Automatic failure pattern detection (Debug Spiral, Context Amnesia, Velocity Crash, Trust Erosion, Flow Disruption)
  - Session entries in `claude-progress.json` now include metrics and retro blocks
  - `@boshu2/vibe-check` npm package added as plugin dependency
- **vibe-coding Plugin** added with commands:
  - `/vibe-check` - Run vibe-check analysis
  - `/vibe-level` - Declare vibe level for session
  - `/vibe-retro` - Run vibe-coding retrospective
- **constitution Plugin** added with:
  - laws-of-an-agent skill
  - context-engineering skill
  - git-discipline skill
  - guardian agent
- SECURITY.md with vulnerability reporting process
- CONTRIBUTING.md with comprehensive plugin submission guidelines
- CHANGELOG.md for version tracking
- CODE_OF_CONDUCT.md for community standards
- GitHub Actions CI/CD pipeline for automated validation
- GitHub issue templates for plugin submissions and bug reports
- GitHub PR template for structured contributions
- Test suite infrastructure with validation scripts
- Makefile for common development tasks
- ARCHITECTURE_REVIEW.md with comprehensive compliance analysis

### Changed
- Updated repository structure to follow GitHub best practices
- Enhanced documentation for better discoverability

### Security
- Established security policy and vulnerability reporting process
- Added automated security scanning (Dependabot, CodeQL)

## [1.0.0] - 2025-11-10

### Added
- Initial marketplace structure with `.claude-plugin/marketplace.json`
- Three core plugins:
  - **core-workflow**: Universal Research → Plan → Implement → Learn workflow
  - **devops-operations**: DevOps and platform engineering tools
  - **software-development**: Software development for Python, JavaScript, Go
- External marketplace references:
  - aitmpl.com/agents (63+ plugins)
  - wshobson/agents (open source collection)
- Comprehensive README with quick start guide
- Apache 2.0 license
- Plugin structure following Anthropic 2025 standards
- 12-Factor AgentOps integration in all agents
- Token budget estimation for plugins

### Agents (11 total)
- **core-workflow** (4 agents):
  - research-agent: Research phase with JIT context loading
  - plan-agent: Planning phase with detailed specifications
  - implement-agent: Implementation phase with validation
  - learn-agent: Learning extraction for continuous improvement
- **devops-operations** (3 agents):
  - devops-engineer: DevOps automation specialist
  - deployment-engineer: Deployment and release management
  - cicd-specialist: CI/CD pipeline expert
- **software-development** (3 agents):
  - software-engineer: General software development
  - code-reviewer: Code quality and review
  - test-engineer: Testing and quality assurance

### Commands (14 total)
- **core-workflow** (5 commands):
  - /research: Start research phase
  - /plan: Create implementation plan
  - /implement: Execute approved plan
  - /learn: Extract learnings
  - /workflow: Full workflow orchestration
- **devops-operations** (3 commands):
  - /deploy-app: Deploy applications
  - /setup-pipeline: Configure CI/CD pipelines
  - /rollback: Rollback deployments
- **software-development** (3 commands):
  - /create-feature: Create new features
  - /refactor-code: Refactor existing code
  - /add-tests: Add test coverage

### Skills (9 total)
- **core-workflow**: Universal workflow patterns
- **devops-operations** (3 skills):
  - gitops-patterns: GitOps workflow patterns
  - kubernetes-manifests: Kubernetes resource templates
  - helm-charts: Helm chart best practices
- **software-development** (3 skills):
  - python-testing: Python testing patterns
  - javascript-patterns: JavaScript/TypeScript patterns
  - go-best-practices: Go language best practices

### Documentation
- Comprehensive README.md with installation instructions
- Plugin-level README files for each plugin
- Agent documentation with examples and anti-patterns
- AgentOps principles integration
- External marketplace references

## Version History

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: New functionality (backwards-compatible)
- **PATCH** version: Bug fixes (backwards-compatible)

### Release Process

1. Update CHANGELOG.md with changes
2. Update version in `.claude-plugin/marketplace.json`
3. Update version in all plugin `plugin.json` files
4. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
5. Push tag: `git push origin v1.0.0`
6. Create GitHub release with changelog excerpt

## Links

- [Repository](https://github.com/boshu2/agentops)
- [Issues](https://github.com/boshu2/agentops/issues)
- [Pull Requests](https://github.com/boshu2/agentops/pulls)
- [Security Policy](SECURITY.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [12-Factor AgentOps Framework](https://github.com/boshu2/12-factor-agentops)

## Community

### How to Stay Updated

- Watch this repository on GitHub
- Check this CHANGELOG regularly
- Follow [@boshu2](https://github.com/boshu2) on GitHub

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on:
- How to add plugins
- Testing requirements
- Submission process
- Code of conduct

### Support

- **Documentation**: Check README.md and plugin docs
- **Issues**: [GitHub Issues](https://github.com/boshu2/agentops/issues)
- **Discussions**: [GitHub Discussions](https://github.com/boshu2/agentops/discussions)

---

**Note:** This changelog is automatically updated with each release. See [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) for format guidelines.
