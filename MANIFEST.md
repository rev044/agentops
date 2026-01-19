# AgentOps Plugin Manifest

A marketplace for Claude Code plugins. Start with `vibe-kit` for a production-ready workflow, or build your own.

## Quick Start

```bash
# Install the recommended starter plugin
/plugin install boshu2/agentops/vibe-kit
```

## Available Plugins

### vibe-kit (Recommended)

A lean, production-ready Claude Code setup built on the 40% rule.

| Component | Count | Description |
|-----------|-------|-------------|
| Commands | 5 | Core workflows: `/research`, `/crank`, `/implement`, `/retro` |
| Skills | 5 | Validation: bug-hunt, complexity, validation-chain, vibe-docs, vibe |
| Agents | 4 | Parallel specialist review: security, architecture, quality, UX |
| CLAUDE.md | 1 | Universal config template |

**Philosophy:** Simple beats clever. Skills replaced most agents. Stay under 40% context utilization.

### Creating Your Own Plugin

Plugins are directories containing:

```
my-plugin/
├── plugin.json       # Metadata
├── commands/         # Explicit /workflows
├── skills/           # Auto-triggered context
├── agents/           # Parallel specialists (use sparingly)
└── README.md         # Documentation
```

See [12-Factor AgentOps](https://12factoragentops.com) for methodology.

## Plugin Format

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "What this plugin does",
  "commands": ["command1.md", "command2.md"],
  "skills": ["skill1/", "skill2/"],
  "agents": ["agent1.md"]
}
```

## Related

- **[12-Factor AgentOps](https://12factoragentops.com)** - The methodology
- **[vibe-kit source](./plugins/vibe-kit/)** - The starter plugin

---

*Less is more. Complexity is where tokens go to die.*
