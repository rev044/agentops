# Codex Skill Standards

## Canonical Contract

Source of truth: `docs/contracts/codex-skill-api.md`

## Frontmatter

Codex SKILL.md frontmatter must contain **only** `name` and `description`:

```yaml
---
name: skill-name
description: 'When this skill triggers and when it does not.'
---
```

**Prohibited fields** (Claude-internal, ignored by Codex):
`skill_api_version`, `context`, `metadata`, `allowed-tools`, `model`, `user-invocable`, `output_contract`

## Tool References

Skills must reference only tools available in Codex sessions:

| Codex Tool | Purpose | Claude Equivalent |
|------------|---------|-------------------|
| `read_file` | Read file contents | `Read` |
| `apply_patch` | Apply file edits | `Edit` |
| `rg` | Search file contents | `Grep` |
| `glob_file_search` | Find files by pattern | `Glob` |
| `cmd` | Shell execution | `Bash` |
| `git` | Git operations | `Bash(git ...)` |
| `list_dir` | List directory | `Bash(ls)` |
| `wait` | Long-poll sub-agents | Built into Agent tool |

### Prohibited Tool References

These Codex primitives have **no Codex equivalent** and must not appear:

- `Skill(skill=...)` — Codex uses `$skill-name` invocation, not a Skill tool
- `Agent(subagent_type=...)` — Codex uses agent roles, not subagent_type

### Mapped Forms Also Prohibited

Lowercase-hyphenated forms are equally invalid (`task-create`, `team-create`, `send-message`).

The previously-mapped `todo_write` and `update_plan` are **not available** as general-purpose tools in Codex sessions (empirically verified via `codex exec`).

## Skill Discovery Paths

| Scope | Path |
|-------|------|
| Repo | `.agents/skills/` |
| User | `~/.agents/skills/` |
| Admin | `/etc/codex/skills/` |

**Prohibited paths:** `~/.agents/skills/`, `~/.codex/skills/`

## Sub-Agent Patterns

Codex orchestration uses:

| Pattern | Tool | Use Case |
|---------|------|----------|
| Batch parallel | `spawn_agents_on_csv` | Many similar tasks from CSV |
| Agent roles | `[agents.<name>]` in config | Specialized sub-agents (worker, explorer, monitor) |
| Shell orchestration | `cmd` + `bd` CLI | Issue tracking, wave management |


## Common Issues

| Pattern | Problem | Fix |
|---------|---------|-----|
| `$vibe ` | Claude Skill tool, doesn't exist | Use `$vibe` invocation syntax |
| `context.window: fork` | Claude frontmatter, ignored | Remove from Codex SKILL.md |
| `~/.agents/skills/` | Wrong path | Use `.agents/skills/` |
| `todo_write(...)` | Not available in Codex sessions | Use `bd` CLI or file-based tracking |

## Testing Codex Skills

Validate empirically with headless Codex:

```bash
# Check if skill loads
codex exec -s read-only -C "$(pwd)" \
  "List your available skills. Does \$skill-name appear?"

# Check for broken references
codex exec -s read-only -C "$(pwd)" \
  "Read \$skill-name. List any tools or primitives it references that you don't have."
```

## Checklist

When reviewing Codex skills (`skills-codex/*/SKILL.md`):

- [ ] Frontmatter has only `name` + `description`
- [ ] No Claude primitive names (PascalCase or lowercase-hyphenated)
- [ ] No `~/.codex/` paths
- [ ] No `Skill(skill=...)` tool invocations
- [ ] No `Agent(subagent_type=...)` tool invocations
- [ ] No `context.*` or `metadata.*` frontmatter
- [ ] Reference files (`references/*.md`) also free of Claude primitives
- [ ] Instructions are actionable for a Codex agent with only Codex tools
