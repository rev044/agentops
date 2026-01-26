# AgentOps for Codex

Setup instructions for using AgentOps with OpenAI Codex.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/boshu2/agentops.git
   ```

2. Add to your Codex configuration:
   ```bash
   # In your project's .codex/config.yaml
   plugins:
     - path: /path/to/agentops
   ```

## Available Plugins

AgentOps provides plugin kits that can be adapted for Codex:

| Kit | Purpose |
|-----|---------|
| solo-kit | Research and validation |
| core-kit | RPI workflow execution |
| vibe-kit | Code validation |
| docs-kit | Documentation |

## Adapting for Codex

The skills in `plugins/*/skills/` contain markdown instructions that can be:
1. Copied to Codex's instruction format
2. Used as reference for custom Codex prompts
3. Integrated via Codex's plugin system

## Key Workflows

### Research
See `plugins/solo-kit/skills/research/SKILL.md`

### Validation
See `plugins/vibe-kit/skills/vibe/SKILL.md`

## Notes

- AgentOps was designed for Claude Code but concepts are portable
- The RPI workflow (Research → Plan → Implement → Validate) applies to any AI coding assistant
- Skill references in `plugins/*/skills/*/references/` contain domain knowledge
