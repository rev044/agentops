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
| **Commands** | 3 | Orchestrators only (skills handle the rest) |
| **Skills** | 15 | Contribution workflow, beads, orchestration |

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

Skills auto-trigger for PR workflow - no explicit commands needed.

## Skills

### PR Workflow

| Skill | Triggers |
|-------|----------|
| `sk-pr-research` | "pr research", "upstream research", "contribution research" |
| `sk-pr-plan` | "pr plan", "contribution plan" |
| `sk-pr-implement` | "implement PR", "pr implement" |
| `sk-pr-validate` | "validate PR", "scope creep", "isolation check" |
| `sk-pr-prep` | "prepare PR", "submit PR", "create PR" |
| `sk-pr-retro` | "pr retro", "learn from PR", "PR outcome" |

### Gas Town Orchestration

| Skill | Triggers |
|-------|----------|
| `beads` | "track issues", "create beads issue", "show blockers" |
| `sk-dispatch` | "gt sling", "gt hook", "gt convoy" |
| `sk-roles` | "Mayor", "Crew", "Polecat", "Witness", "Refinery" |
| `sk-mail` | "gt mail", "send mail", "check inbox" |
| `sk-handoff` | "hand off", "gt handoff", "context cycling" |
| `sk-crew` | "crew workspace", "persistent workspace" |
| `sk-polecat-lifecycle` | "spawn polecat", "nuke polecat", "reset polecat" |
| `sk-gastown` | "Gas Town", "gt status", "rig list" |
| `sk-bd-routing` | "beads routing", "prefix routing", "BEADS_DIR" |

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
