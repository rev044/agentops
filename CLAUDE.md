# AgentOps Plugin Marketplace

Claude Code plugins for AI-assisted development workflows.

## Project Structure

```
plugins/           # Plugin packages (each installable independently)
  core-kit/        # Essential workflow skills (research, formulate, crank, etc.)
  vibe-kit/        # Code validation with domain expert agents
  general-kit/     # Zero-dependency portable version of vibe
  beads-kit/       # Git-based issue tracking integration
  docs-kit/        # Documentation generation and validation
  pr-kit/          # Pull request workflow skills
  dispatch-kit/    # Multi-agent coordination (mail, handoff, dispatch)
  gastown-kit/     # Gas Town orchestration skills
  domain-kit/      # Domain-specific skills (brand guidelines, etc.)
shared/            # Shared resources across plugins
  scripts/         # Common scripts (prescan.sh, etc.)
tests/             # Validation and smoke tests
templates/         # Plugin templates
```

## Plugin Structure Conventions

Each plugin follows this structure:

```
plugins/<name>-kit/
  .claude-plugin/
    plugin.json     # Manifest with name, version, dependencies
  skills/
    <skill-name>/
      SKILL.md      # Entry point with triggers, instructions
      references/   # Progressive disclosure docs (loaded JIT)
      scripts/      # Validation scripts (optional)
  agents/           # Custom agents (optional)
  commands/         # Custom commands (optional)
  CLAUDE.md         # Plugin-specific Claude instructions (optional)
```

## Testing

```bash
# Validate all skills (static checks)
./tests/skills/run-all.sh

# Validate a specific skill
./tests/skills/validate-skill.sh plugins/vibe-kit/skills/vibe

# Run smoke tests
./tests/smoke-test.sh

# Run marketplace e2e test
./tests/marketplace-e2e-test.sh
```

## Common Tasks

**Create a new skill:**
```bash
# Use the skill-creator if available, or copy from template
cp -r templates/skill-template plugins/your-kit/skills/new-skill
```

**Add shared scripts:**
- Place in `shared/scripts/`
- Symlink from plugin: `ln -s ../../../../../shared/scripts/script.sh script.sh`

**Test a plugin locally:**
```bash
claude --plugin ./plugins/your-kit
```

## Key Patterns

1. **SKILL.md is the entry point** - Triggers, instructions, allowed tools
2. **References are loaded JIT** - Keep SKILL.md lean, details in references/
3. **Scripts validate behavior** - Prove skills work, catch regressions
4. **Symlinks for shared code** - Avoid duplication across plugins

## See Also

- [CONTRIBUTING.md](CONTRIBUTING.md) - Full contribution guide
- [README.md](README.md) - Project overview and workflow guide
- [tests/](tests/) - Test infrastructure
