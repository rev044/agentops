# Beads Kit

Git-based issue tracking. 3 skills for managing work with beads.

## Install

```bash
/plugin install beads-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/beads` | auto-triggered | Issue tracking with bd CLI |
| `/status` | `/status` | Quick status check |
| `/molecules` | auto-triggered | Workflow template guidance |

## Core Concepts

### Beads
Issues stored in `.beads/issues.jsonl` - version controlled, agent-friendly.

### Molecules
Workflow templates that encode multi-step processes. A molecule "cooks" to generate beads issues.

## Examples

### Check status

```bash
/status
# Shows: active issues, git state, ready work
```

### Work with issues

```bash
bd ready              # Find unblocked issues
bd show gt-1234       # View issue details
bd close gt-1234      # Mark complete
bd sync               # Push/pull beads
```

### Use molecules

```bash
gt mol cook           # Generate issues from template
gt mol status         # Check molecule progress
```

## Philosophy

- **Issues in git, not in the cloud** - travels with code
- **Agent-friendly workflows** - designed for automation
- **Molecule-based task composition** - reusable patterns

## Related Kits

- **core-kit** - `/plan` creates beads issues
- **dispatch-kit** - Work assignment uses beads
- **gastown-kit** - Multi-rig beads routing
