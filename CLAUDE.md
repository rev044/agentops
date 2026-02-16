# AgentOps

Your coding agent gets smarter every time you use it. Automates context, validation, execution, and learning — each session compounds on every session before it.

## Project Structure

```
.
├── .claude-plugin/
│   ├── plugin.json        # Plugin manifest (v2.7.1)
│   └── marketplace.json   # Marketplace metadata
├── skills/                # All 35 skills (25 user-facing, 10 internal)
│   ├── council/           # Multi-model validation (core primitive)
│   ├── crank/             # Hands-free epic execution
│   ├── swarm/             # Parallel agent spawning
│   ├── codex-team/        # Parallel Codex execution agents
│   ├── rpi/               # Full RPI lifecycle orchestrator
│   ├── evolve/            # Autonomous fitness-scored improvement loop
│   ├── implement/         # Execute single issue
│   ├── quickstart/        # Interactive onboarding
│   ├── status/            # Single-screen dashboard
│   ├── research/          # Deep codebase exploration
│   ├── plan/              # Decompose epics into issues
│   ├── product/           # Interactive PRODUCT.md generation
│   ├── vibe/              # Code validation (complexity + council)
│   ├── pre-mortem/        # Council on plans (failure simulation)
│   ├── post-mortem/       # Council + retro (wrap up work)
│   ├── retro/             # Extract learnings
│   ├── complexity/        # Cyclomatic analysis
│   ├── knowledge/         # Query knowledge artifacts
│   ├── bug-hunt/          # Investigate bugs
│   ├── doc/               # Generate documentation
│   ├── handoff/           # Session handoff
│   ├── inbox/             # Agent mail monitoring
│   ├── release/           # Pre-flight, changelog, tag
│   ├── trace/             # Trace design decisions
│   ├── beads/             # Issue tracking reference (internal)
│   ├── standards/         # Coding standards (internal)
│   ├── shared/            # Shared reference docs (internal)
│   ├── inject/            # Load knowledge at session start (internal)
│   ├── extract/           # Extract from transcripts (internal)
│   ├── forge/             # Mine transcripts (internal)
│   ├── provenance/        # Trace knowledge lineage (internal)
│   ├── ratchet/           # Progress gates (internal)
│   ├── flywheel/          # Knowledge health monitoring (internal)
│   └── using-agentops/    # Workflow guide (auto-injected, internal)
├── hooks/                 # Session and git hooks
│   ├── hooks.json
│   ├── session-start.sh
│   └── ...                # 13 hook scripts total
├── cli/                   # Go CLI (ao command)
├── lib/                   # Shared code
│   ├── skills-core.js
│   ├── hook-helpers.sh
│   ├── schemas/           # JSON schemas (team-spec, worker-output)
│   └── scripts/           # prescan.sh, team-runner.sh, watch-codex-stream.sh
├── docs/                  # Documentation
├── tests/                 # Validation and smoke tests
├── .agents/               # Knowledge artifacts (generated)
└── .beads/                # Issue tracking state
```

## Skill Structure

Each skill follows this structure:

```
skills/<skill-name>/
  SKILL.md          # Entry point with triggers, instructions (YAML frontmatter required)
  references/       # Progressive disclosure docs (loaded JIT)
  scripts/          # Validation scripts (optional)
```

See `skills/SKILL-TIERS.md` for the full skill taxonomy and dependency graph.

## Testing

```bash
# Validate all skills (static checks)
./tests/skills/run-all.sh

# Validate a specific skill
./tests/skills/validate-skill.sh skills/vibe

# Run smoke tests
./tests/smoke-test.sh

# Run marketplace e2e test
./tests/marketplace-e2e-test.sh

# Run full test suite (all tiers)
./tests/run-all.sh
```

## Common Tasks

**Create a new skill:**
```bash
# Create skill directory with SKILL.md
mkdir -p skills/new-skill
# Add SKILL.md with YAML frontmatter (name, description, tier)
```

**Test the plugin locally:**
```bash
claude --plugin ./
```

**Update dependencies:**
```bash
# Go modules
cd cli && go get -u ./... && go mod tidy

# Scan all deps (Go, GitHub Actions, Dockerfile) with Renovate
GITHUB_COM_TOKEN=$(gh auth token) renovate --platform=local
```

## Key Patterns

1. **This repo is the source of truth for skills** - Edit skills HERE, push to git, consumers install via `npx skills@latest add boshu2/agentops --all -g`. Never edit installed copies.
2. **SKILL.md is the entry point** - Triggers, instructions, allowed tools
3. **References are loaded JIT** - Keep SKILL.md lean, details in references/
4. **Scripts validate behavior** - Prove skills work, catch regressions
5. **Subagents are defined inline** - Agent behaviors live in SKILL.md files, not as separate files

## Development Pitfalls

These mistakes have been made repeatedly. Read before planning or implementing.

1. **Verify mechanically before planning** — Never write counts like "11 missing" or "28 need X" without running the actual command first. Grep for ALL invocation sites; don't list known ones. Plans built on assumed data have caused 43% scope underestimation.

2. **Go tasks need test validation** — Always add `"tests": "go test ./..."` to task validation metadata. Check for `_test.go` existence when creating new `.go` files. Run `go build` after each wave, not just at the end.

3. **TaskCreate AFTER TeamCreate** — Tasks created before spawning a team are invisible to teammates. Always: TeamCreate first, then TaskCreate. This has been re-learned 3 times.

4. **Lead-only commits in parallel work** — Workers write files but NEVER run `git add` or `git commit`. The team lead validates artifacts exist, then commits once after the wave. Worker commits are the dominant source of merge conflicts.

5. **Grep all call sites before changing signatures** — When a function signature changes, grep for ALL callers and update them in the same commit. Missed callers have broken `go build` repeatedly.

6. **No hardcoded counts without CI assertions** — README badges, prose counts ("34 skills", "10 internal") drift immediately. Pair every count in prose with a validation script that fails when reality diverges. This CLAUDE.md itself has drifted multiple times.

7. **Validate full corpus, not date-scoped subsets** — Scoping checks to `2026-02-1*` or similar date ranges masks non-compliant files outside that range. Only use date filtering for reporting, never for validation gates.

8. **Tighten validation regex patterns** — Grep-based checks cause false positives when patterns are too broad (matching comments, unrelated headers, etc.). Use structured parsing or scope patterns to specific file sections.

9. **Parallel agents must not share files** — When two agents modify the same file, the last write wins silently. Wave planning must ensure file disjointness. Always `git diff` after each wave to catch unexpected reversions.

## See Also

- [README.md](README.md) - Project overview and workflow guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guide
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [docs/SKILLS.md](docs/SKILLS.md) - Skills reference
- [skills/SKILL-TIERS.md](skills/SKILL-TIERS.md) - Skill taxonomy and dependencies
- [tests/](tests/) - Test infrastructure
