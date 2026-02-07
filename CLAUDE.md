# AgentOps

A knowledge flywheel for AI coding agents — persistent memory, multi-model validation, and structured workflows for Claude Code.

## Project Structure

```
.
├── .claude-plugin/
│   ├── plugin.json        # Plugin manifest (v1.7.1)
│   └── marketplace.json   # Marketplace metadata
├── skills/                # All 32 skills (21 user-facing, 11 internal)
│   ├── council/           # Multi-model validation (core primitive)
│   ├── crank/             # Autonomous epic execution
│   ├── swarm/             # Parallel agent spawning
│   ├── codex-team/        # Parallel Codex execution agents
│   ├── implement/         # Execute single issue
│   ├── quickstart/        # Interactive onboarding
│   ├── status/            # Single-screen dashboard
│   ├── research/          # Deep codebase exploration
│   ├── plan/              # Decompose epics into issues
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
│   ├── using-agentops/    # Workflow guide (auto-injected, internal)
│   └── judge/             # DEPRECATED — use /council
├── hooks/                 # Session and git hooks
│   ├── hooks.json
│   ├── session-start.sh
│   └── ...                # 10 hook scripts total
├── cli/                   # Go CLI (ao command)
├── lib/                   # Shared code
│   ├── skills-core.js
│   └── scripts/prescan.sh
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

## Key Patterns

1. **SKILL.md is the entry point** - Triggers, instructions, allowed tools
2. **References are loaded JIT** - Keep SKILL.md lean, details in references/
3. **Scripts validate behavior** - Prove skills work, catch regressions
4. **Subagents are defined inline** - Agent behaviors live in SKILL.md files, not as separate files

## See Also

- [README.md](README.md) - Project overview and workflow guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guide
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [docs/SKILLS.md](docs/SKILLS.md) - Skills reference
- [skills/SKILL-TIERS.md](skills/SKILL-TIERS.md) - Skill taxonomy and dependencies
- [tests/](tests/) - Test infrastructure
