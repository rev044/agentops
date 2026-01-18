# gastown

Gas Town multi-agent orchestration plugin for Claude Code.

**CONTRIBUTING.md for agents** - executable contribution workflows instead of docs to read.

## Quick Install

```bash
# Via Claude Code plugin system
/plugin marketplace add boshu2/agentops
/plugin install gastown@boshu2-agentops
```

## What's Included

| Component | Count | Description |
|-----------|-------|-------------|
| **Skills** | 18 | Contribution workflow, orchestration, validation |

Skills are directly invokable with `/skill-name` - no command wrappers needed.

## Core Workflow: PR Contribution

```
/pr-research → /pr-plan → /pr-implement → /pr-validate → /pr-prep → /pr-retro
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

## Skills

### PR Workflow

| Skill | Invoke | Triggers |
|-------|--------|----------|
| `pr-research` | `/pr-research` | "upstream research", "contribution research" |
| `pr-plan` | `/pr-plan` | "contribution plan" |
| `pr-implement` | `/pr-implement` | "implement PR" |
| `pr-validate` | `/pr-validate` | "scope creep", "isolation check" |
| `pr-prep` | `/pr-prep` | "prepare PR", "submit PR" |
| `pr-retro` | `/pr-retro` | "learn from PR", "PR outcome" |

### Gas Town Orchestration

| Skill | Invoke | Triggers |
|-------|--------|----------|
| `beads` | `/beads` | "track issues", "create beads issue", "show blockers" |
| `dispatch` | `/dispatch` | "gt sling", "gt hook", "gt convoy" |
| `roles` | `/roles` | "Mayor", "Crew", "Polecat", "Witness", "Refinery" |
| `mail` | `/mail` | "gt mail", "send mail", "check inbox" |
| `handoff` | `/handoff` | "gt handoff", "context cycling" |
| `crew` | `/crew` | "crew workspace", "persistent workspace" |
| `polecat-lifecycle` | `/polecat-lifecycle` | "spawn polecat", "nuke polecat", "reset polecat" |
| `gastown` | `/gastown` | "Gas Town", "gt status", "rig list" |
| `bd-routing` | `/bd-routing` | "beads routing", "prefix routing", "BEADS_DIR" |
| `status` | `/status` | "what's my status", "current state" |

### Vibe Validation

| Skill | Invoke | Triggers |
|-------|--------|----------|
| `vibe` | `/vibe` | "validate code", "check semantic faithfulness" |
| `vibe-docs` | `/vibe-docs` | "verify docs", "doc audit", "check doc claims" |

## Philosophy

```
"Check for existing work before starting"
```

### Principles

1. **Phase -1 is BLOCKING** - Prior work check before any implementation
2. **Skills are directly invokable** - `/skill-name` works, no command wrappers
3. **CONTRIBUTING.md for agents** - Humans read docs, agents run workflows
4. **Version controlled** - Community maintained, PRs welcome

## Related

- [steveyegge/gastown](https://github.com/steveyegge/gastown) - Gas Town multi-agent orchestrator
- [steveyegge/beads](https://github.com/steveyegge/beads) - Git-based issue tracking

## License

MIT
