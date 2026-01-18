# PR Kit

Open source contribution workflow. 6 skills from research to retrospective.

## Install

```bash
/plugin install pr-kit@boshu2-agentops
```

## Skills

| Skill | Invoke | Purpose |
|-------|--------|---------|
| `/pr-research` | `/pr-research <repo>` | Upstream codebase exploration |
| `/pr-plan` | `/pr-plan` | Strategic contribution planning |
| `/pr-implement` | `/pr-implement` | Fork-based implementation |
| `/pr-validate` | `/pr-validate` | Isolation/scope validation |
| `/pr-prep` | `/pr-prep` | PR body generation |
| `/pr-retro` | `/pr-retro <pr>` | Learn from PR outcomes |

## Complete Workflow

```
/pr-research → /pr-plan → /pr-implement → /pr-validate → /pr-prep → /pr-retro
     ↓            ↓            ↓              ↓            ↓          ↓
  explore     scope it      code it       verify it    submit it   learn
```

## Phase -1: Prior Work Check

**BLOCKING** - Before ANY PR work:

1. Search existing issues
2. Check open/closed PRs
3. Read CONTRIBUTING.md
4. Understand maintainer expectations

This prevents wasted effort on duplicate or unwanted contributions.

## Examples

### Full contribution workflow

```bash
/pr-research kubernetes/kubernetes
# Creates .agents/research/ artifact with guidelines, patterns

/pr-plan
# Creates .agents/plans/ with scope, acceptance criteria

/pr-implement
# Runs isolation check, creates fork, implements

/pr-validate
# Checks isolation, upstream alignment, scope creep

/pr-prep
# Generates PR body with structured format

# After PR merged/rejected:
/pr-retro #1234
# Captures learnings, updates lessons-learned.md
```

### Quick contribution

```bash
# For simple, well-understood changes:
/pr-implement
# Includes Phase 0 isolation check automatically
```

## Philosophy

- **CONTRIBUTING.md for agents, not just humans**
- **Phase -1: Prior work check is BLOCKING**
- **Research maintainer expectations first**
- **Validate isolation before submission**

## Related Kits

- **core-kit** - `/research` for codebase exploration
- **vibe-kit** - Validate before submitting
