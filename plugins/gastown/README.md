# gastown

Gas Town multi-agent orchestration plugin for Claude Code.

**CONTRIBUTING.md for agents** - executable contribution workflows instead of docs to read.

## Quick Install

```bash
# Clone agentops
git clone https://github.com/boshu2/agentops.git ~/.claude/plugins/agentops

# Copy gastown plugin to your .claude/
cp -r ~/.claude/plugins/agentops/plugins/gastown/* ~/.claude/

# Or symlink for updates
ln -s ~/.claude/plugins/agentops/plugins/gastown/commands/* ~/.claude/commands/
ln -s ~/.claude/plugins/agentops/plugins/gastown/skills/* ~/.claude/skills/
```

## What's Included

| Component | Count | Description |
|-----------|-------|-------------|
| **Commands** | 8 | Orchestrators + vibe validation |
| **Skills** | 17 | Contribution workflow, beads, orchestration, vibe |

## Core Workflow: PR Contribution

```
/pr-research → /pr-plan → /pr-implement → /pr-validate → /pr-retro
```

### Phase -1: Prior Work Check (BLOCKING)

Every workflow starts by checking for existing work:

```bash
# Search for open issues on this topic
gh issue list -R <owner/repo> --state open --search "<topic>"

# Search for open PRs
gh pr list -R <owner/repo> --state open --search "<topic>"

# Check recently merged PRs
gh pr list -R <owner/repo> --state merged --search "<topic>"
```

**If prior work exists → DO NOT PROCEED.** Comment on existing issue or coordinate with PR author.

## Commands

| Command | Purpose |
|---------|---------|
| `/gastown` | Gas Town status and orchestration |
| `/beads-validate` | Validate beads state |
| `/status` | Quick status check |
| `/vibe` | Semantic validation of code behavior |
| `/vibe-docs` | Validate docs match deployment reality |
| `/vibe-plugin` | Validate plugins with L13 semantic verification |
| `/vibe-prescan` | Fast static scan for 6 failure patterns |
| `/vibe-semantic` | Orchestrate semantic faithfulness analyses |

Skills auto-trigger for PR workflow - no explicit commands needed.

## Skills

### PR Workflow

| Skill | Triggers |
|-------|----------|
| `pr-research` | "pr research", "upstream research", "contribution research" |
| `pr-plan` | "pr plan", "contribution plan" |
| `pr-implement` | "implement PR", "pr implement" |
| `pr-validate` | "validate PR", "scope creep", "isolation check" |
| `pr-prep` | "prepare PR", "submit PR", "create PR" |
| `pr-retro` | "pr retro", "learn from PR", "PR outcome" |

### Gas Town Orchestration

| Skill | Triggers |
|-------|----------|
| `beads` | "track issues", "create beads issue", "show blockers" |
| `dispatch` | "gt sling", "gt hook", "gt convoy" |
| `roles` | "Mayor", "Crew", "Polecat", "Witness", "Refinery" |
| `mail` | "gt mail", "send mail", "check inbox" |
| `handoff` | "hand off", "gt handoff", "context cycling" |
| `crew` | "crew workspace", "persistent workspace" |
| `polecat-lifecycle` | "spawn polecat", "nuke polecat", "reset polecat" |
| `gastown` | "Gas Town", "gt status", "rig list" |
| `bd-routing` | "beads routing", "prefix routing", "BEADS_DIR" |

### Vibe Validation

| Skill | Triggers |
|-------|----------|
| `vibe` | "validate code", "check semantic faithfulness", "run vibe" |
| `vibe-docs` | "verify docs", "doc audit", "check doc claims" |

## Philosophy

```
"Check for existing work before starting"
```

### Principles

1. **Phase -1 is BLOCKING** - Prior work check before any implementation
2. **Skills encode workflow** - Executable knowledge, not just documentation
3. **CONTRIBUTING.md for agents** - Humans read docs, agents run workflows
4. **Version controlled** - Community maintained, PRs welcome

## Related

- [steveyegge/gastown](https://github.com/steveyegge/gastown) - Gas Town multi-agent orchestrator
- [steveyegge/beads](https://github.com/steveyegge/beads) - Git-based issue tracking
- [JeremyKalmus/gastown-plugins](https://github.com/JeremyKalmus/gastown-plugins) - Community plugin collection

## License

MIT
