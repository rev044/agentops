# AgentOps for OpenCode

Setup instructions for using AgentOps with OpenCode.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/boshu2/agentops.git
   ```

2. Configure OpenCode to use AgentOps:
   ```bash
   # Reference the plugins directory in your OpenCode config
   opencode config set plugins.path /path/to/agentops/plugins
   ```

## Available Plugins

| Kit | Purpose | Skills |
|-----|---------|--------|
| solo-kit | Research and validation | research, vibe, bug-hunt |
| core-kit | RPI workflow | implement, crank, retro |
| vibe-kit | Code quality | vibe, vibe-docs |
| docs-kit | Documentation | doc, oss-docs |

## Adapting for OpenCode

Skills are markdown-based and portable:

```
plugins/
├── solo-kit/
│   └── skills/
│       └── research/
│           ├── SKILL.md          # Main instructions
│           └── references/       # Supporting docs
```

## Key Workflows

### The RPI Loop

```
/research → /plan → /execute → /validate → learn → repeat
```

1. **Research**: Understand the codebase
2. **Plan**: Decompose into tasks
3. **Execute**: Implement autonomously
4. **Validate**: Check quality

### Validation

The `/vibe` skill validates across 8 aspects:
- Security, Code Quality, Architecture, Accessibility
- Testing, Documentation, Performance, Dependencies

## Notes

- Skills contain all instructions in SKILL.md
- References provide domain knowledge loaded just-in-time
- The workflow concepts apply to any AI coding assistant
