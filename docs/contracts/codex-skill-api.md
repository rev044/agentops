# Codex Skill API Contract

> Source of truth for what the Codex runtime actually supports. All converter output and validation must conform to this contract.

**Official docs:**
- [Codex Skills](https://developers.openai.com/codex/skills/)
- [Codex Multi-Agent](https://developers.openai.com/codex/multi-agent/)
- [Codex CLI Features](https://developers.openai.com/codex/cli/features)

---

## SKILL.md Frontmatter

Codex recognizes **only** these frontmatter fields:

```yaml
---
name: skill-name
description: 'Explain when this skill triggers and when it does not.'
---
```

**Required:** `name`, `description`
**Everything else is ignored.** Fields like `skill_api_version`, `context`, `metadata`, `allowed-tools`, `model`, `user-invocable`, and `output_contract` are AgentOps-internal and must be stripped from Codex output.

---

## Optional: agents/openai.yaml

Codex skills may include `agents/openai.yaml` for display metadata and policy:

```yaml
interface:
  display_name: "User-facing name"
  short_description: "Brief description"
  icon_small: "./assets/small-logo.svg"
  icon_large: "./assets/large-logo.png"
  brand_color: "#3B82F6"
  default_prompt: "Optional surrounding prompt"

policy:
  allow_implicit_invocation: false

dependencies:
  tools:
    - type: "mcp"
      value: "toolName"
      description: "Tool description"
      transport: "streamable_http"
      url: "https://example.com"
```

| Field | Purpose |
|-------|---------|
| `interface.display_name` | User-visible name in Codex UI |
| `interface.short_description` | Brief description for skill browser |
| `policy.allow_implicit_invocation` | `false` prevents auto-activation (explicit `$skill` only) |
| `dependencies.tools` | MCP server dependencies |

---

## Skill Discovery Paths

Codex scans these directories (in order):

| Scope | Path | Use Case |
|-------|------|----------|
| Repo (nearest) | `.agents/skills/` from CWD | Folder-specific workflows |
| Repo (parent) | `../.agents/skills/` | Nested repo organization |
| Repo (root) | `$REPO_ROOT/.agents/skills/` | Organization-wide skills |
| User | `$HOME/.agents/skills/` | Personal skill collection |
| Admin | `/etc/codex/skills/` | System-wide defaults |
| System | Bundled with Codex | Built-in skills |

**NOT:** `~/.claude/skills/` or `~/.codex/skills/` — these are Claude Code paths.

---

## Skill Invocation

| Method | Syntax | Description |
|--------|--------|-------------|
| Explicit | `$skill-name` or `/skills` menu | User directly requests the skill |
| Implicit | Automatic | Codex matches task to skill description |

Skills are loaded via **progressive disclosure**: metadata first (name, description), full SKILL.md only when activated.

---

## Multi-Agent (Sub-Agents)

Codex multi-agent is experimental. Enable via `/experimental` or `multi_agent = true` in `~/.codex/config.toml`.

### Agent Roles

Configured in `[agents]` section of config files:

| Role | Purpose |
|------|---------|
| `default` | General-purpose fallback |
| `worker` | Execution-focused implementation |
| `explorer` | Read-heavy codebase exploration |
| `monitor` | Long-running command/task monitoring |

```toml
[agents]
max_threads = 6
max_depth = 1
job_max_runtime_seconds = 1800

[agents.reviewer]
description = "Code review specialist"
config_file = "codex-reviewer.toml"
```

### Batch Processing

`spawn_agents_on_csv` processes batches of similar tasks:

| Parameter | Description |
|-----------|-------------|
| `csv_path` | Source CSV file |
| `instruction` | Worker prompt template with `{column}` placeholders |
| `id_column` | Stable identifiers |
| `output_schema` | Fixed JSON structure for worker results |
| `max_concurrency` | Parallel worker limit |
| `max_runtime_seconds` | Worker timeout |

Workers call `report_agent_job_result` exactly once.

### What Does NOT Exist in Codex

These are **Claude Code primitives** with no Codex equivalent:

| Primitive | Claude Code | Codex |
|-----------|-------------|-------|
| `TaskCreate` | Create tasks in session | Does not exist |
| `TaskList` | List session tasks | Does not exist |
| `TaskUpdate` | Update task status | Does not exist |
| `TaskGet` | Get task details | Does not exist |
| `TaskStop` | Stop a task | Does not exist |
| `TeamCreate` | Create agent team | Does not exist |
| `TeamDelete` | Delete agent team | Does not exist |
| `SendMessage` | Message between agents | Does not exist |
| `EnterPlanMode` | Enter plan mode | Does not exist |
| `ExitPlanMode` | Exit plan mode | Does not exist |
| `EnterWorktree` | Enter git worktree | Does not exist |
| `context.window` | Knowledge context control | Does not exist |
| `context.sections.exclude` | Section filtering | Does not exist |
| `context.intel_scope` | Intelligence scoping | Does not exist |

Skills referencing these primitives produce **broken instructions** in Codex.

---

## Converter Requirements

When generating Codex skills from source skills:

1. **Strip all non-Codex frontmatter** — emit only `name` + `description`
2. **Remove Claude primitive references** — do not rename to lowercase (e.g., `team-create` doesn't exist either)
3. **Remove or replace orchestration sections** — sections describing TaskList workflows need Codex-native alternatives or removal
4. **Fix paths** — `~/.claude/` → `.agents/` (not `~/.codex/`)
5. **Generate `agents/openai.yaml`** when display metadata or implicit invocation policy applies
6. **Preserve skill body** — the SKILL.md body (instructions) is the skill's value; keep it functional

---

## Validation Criteria

A Codex-conformant skill must:

1. Have frontmatter with only `name` and `description`
2. Contain no Claude-only primitive names (TaskCreate, TeamCreate, SendMessage, etc.)
3. Contain no Claude-specific paths (`~/.claude/`, `~/.codex/`)
4. Have valid `agents/openai.yaml` if present
5. Not reference non-existent Codex features (context controls, plan mode, etc.)
