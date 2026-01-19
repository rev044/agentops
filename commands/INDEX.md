# Commands Index

**DEPRECATED**: Commands are replaced by **skills** in `/plugins/`. Skills are directly invokable with `/skill-name` and provide better context loading.

## Migration Guide

| Old Command | New Skill |
|-------------|-----------|
| `/research` | `/research` (core-kit) |
| `/plan` | Claude's native plan mode or `/formulate` |
| `/implement` | `/implement` (core-kit) |
| `/retro` | `/retro` (core-kit) |
| `/vibe-check` | `/vibe` (vibe-kit) |
| `/code-review` | `/vibe` (vibe-kit) |
| `/architecture-review` | `/vibe` (vibe-kit) |
| `/update-docs` | `/doc` (docs-kit) |
| `/research-multi` | `/implement-wave` (core-kit) |
| Session commands | Claude's native `/rename`, `/resume`, `--continue` |

## Install Plugins Instead

```bash
/plugin install core-kit@boshu2-agentops
/plugin install vibe-kit@boshu2-agentops
/plugin install docs-kit@boshu2-agentops
```

See [../plugins/](../plugins/) for the full list.
