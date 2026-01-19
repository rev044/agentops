# Contributing to AgentOps Marketplace

Thank you for your interest in contributing to the AgentOps marketplace! This guide will help you create and submit high-quality plugins.

## Table of Contents

- [Getting Started](#getting-started)
- [How to Add a Plugin](#how-to-add-a-plugin)
- [Plugin Structure](#plugin-structure)
- [Quality Standards](#quality-standards)
- [Testing Your Plugin](#testing-your-plugin)
- [Submission Process](#submission-process)
- [Code of Conduct](#code-of-conduct)

## Getting Started

### Prerequisites

- GitHub account
- Claude Code installed and configured
- Basic understanding of Claude Code plugins
- Familiarity with markdown and JSON

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/agentops.git
cd agentops
```

## How to Add a Plugin

### 1. Create Plugin Directory Structure

```bash
# Create your plugin directory
mkdir -p plugins/your-plugin-name/{.claude-plugin,agents,commands,skills}
```

### 2. Create Plugin Manifest

Create `plugins/your-plugin-name/.claude-plugin/plugin.json`:

```json
{
  "name": "your-plugin-name",
  "version": "1.0.0",
  "description": "Brief description of what your plugin does",
  "author": "Your Name",
  "license": "Apache-2.0",
  "keywords": ["relevant", "keywords", "here"],
  "components": {
    "agents": [
      "agents/your-agent.md"
    ],
    "commands": [
      "commands/your-command.md"
    ],
    "skills": [],
    "hooks": [],
    "mcp": []
  },
  "tokenBudget": {
    "estimated": 5000,
    "percentage": 2.5
  }
}
```

**Required Fields:**
- `name`: Kebab-case identifier (lowercase, hyphens)
- `version`: Semantic versioning (MAJOR.MINOR.PATCH)
- `description`: Clear, concise purpose (max 100 chars)
- `author`: Your name or organization
- `license`: Should be Apache-2.0 for consistency
- `keywords`: 3-7 searchable terms
- `components`: Paths to your plugin components

**Optional but Recommended:**
- `tokenBudget`: Estimate context usage (helps users)
- `homepage`: Link to documentation
- `repository`: Link to source code
- `dependencies`: Other plugins required

### 3. Create Plugin Components

#### Agents (`agents/your-agent.md`)

```markdown
---
name: your-agent
description: What this agent does
model: sonnet
tools: [Read, Write, Bash]
---

# Your Agent Name

**Purpose:** One-sentence purpose statement

**This agent demonstrates:**
- ✅ Factor II: JIT Context Loading
- ✅ Factor VI: Session Continuity
- ✅ Factor VII: Intelligent Routing

---

## 🔴 Laws of an Agent (MANDATORY)

[Include standard Laws section - see core-workflow examples]

---

## Role & Constraints

### What This Agent Does
[Detailed description]

### What This Agent Does NOT Do
[Clear boundaries]

---

## How to Execute

[Step-by-step usage guide]

---

## Examples

[Concrete usage examples]

---

## Success Criteria

[How to know the agent succeeded]
```

#### Commands (`commands/your-command.md`)

```markdown
# Command Description

Brief description of what this command does.

## Usage

\`\`\`bash
/your-command [arguments]
\`\`\`

## Arguments

- `arg1`: Description of first argument
- `arg2`: Description of second argument

## Examples

\`\`\`bash
/your-command --option value
\`\`\`

## Output

Description of expected output.
```

#### Skills (`skills/your-skill/README.md`)

```markdown
# Skill Name

## Purpose

What this skill provides to agents.

## Usage

How agents use this skill.

## Examples

Concrete examples of skill usage.
```

### 4. Create Plugin README

Create `plugins/your-plugin-name/README.md`:

```markdown
# Plugin Name

Brief description of your plugin.

## Features

- Feature 1
- Feature 2
- Feature 3

## Installation

\`\`\`bash
/plugin install your-plugin-name@agentops
\`\`\`

## Components

### Agents

- **agent-name** - Description

### Commands

- **/command-name** - Description

### Skills

- **skill-name** - Description

## Usage Examples

[Provide 2-3 concrete examples]

## Token Budget

Estimated: ~5,000 tokens (2.5% of context)

## Dependencies

- core-workflow (required)
- other-plugin (optional)

## License

Apache-2.0
```

### 5. Register in Marketplace

Add your plugin to `.claude-plugin/marketplace.json`:

```json
{
  "plugins": [
    {
      "name": "your-plugin-name",
      "source": "./your-plugin-name",
      "description": "Brief description",
      "version": "1.0.0",
      "author": "Your Name",
      "keywords": ["relevant", "keywords"],
      "dependencies": ["core-workflow"]
    }
  ]
}
```

**Important:** Add your plugin to the `plugins` array, maintaining alphabetical order.

## Plugin Structure

### Directory Layout

```
plugins/your-plugin-name/
├── .claude-plugin/
│   └── plugin.json           # Plugin manifest (required)
├── agents/                   # Agent definitions
│   └── your-agent.md
├── commands/                 # Slash commands
│   └── your-command.md
├── skills/                   # Agent skills
│   └── your-skill/
│       └── README.md
├── hooks/                    # Event hooks (optional)
│   └── hooks.json
├── .mcp.json                 # MCP servers (optional)
└── README.md                 # Plugin documentation (required)
```

### File Naming Conventions

- **Plugins:** `kebab-case` (e.g., `devops-operations`)
- **Agents:** `kebab-case.md` (e.g., `research-agent.md`)
- **Commands:** `kebab-case.md` (e.g., `deploy-app.md`)
- **Skills:** `kebab-case/` (e.g., `python-testing/`)

## Quality Standards

### Documentation Requirements

✅ **Every plugin must have:**
- Clear README.md with installation instructions
- plugin.json with all required fields
- Comprehensive agent documentation
- Usage examples
- Token budget estimate

✅ **Every agent must have:**
- Frontmatter with name, description, model, tools
- Laws of an Agent section
- Role & Constraints section
- How to Execute section
- Success Criteria section
- At least one concrete example

✅ **Every command must have:**
- Clear description
- Usage syntax
- Argument descriptions
- At least one example

### Code Quality

✅ **Required:**
- No hardcoded credentials or secrets
- No sensitive data in repository
- Valid JSON in all manifest files
- Valid YAML in agent frontmatter
- Working markdown links
- Accurate tool permissions

✅ **Recommended:**
- Follow 12-Factor AgentOps principles
- Include anti-patterns to avoid
- Provide troubleshooting guidance
- Document known limitations

### Token Budget Estimation

Estimate your plugin's context usage:

1. **Count words** in all markdown files
2. **Multiply by 1.3** (tokens ≈ words × 1.3)
3. **Calculate percentage** (tokens / 200,000 × 100)
4. **Add to plugin.json**

Example:
- Total words: ~3,000
- Estimated tokens: 3,000 × 1.3 = 3,900
- Percentage: 3,900 / 200,000 × 100 = 1.95%

```json
"tokenBudget": {
  "estimated": 3900,
  "percentage": 1.95
}
```

## Testing Your Plugin

### 1. Local Installation Test

```bash
# From repository root
/plugin install file://$(pwd)/plugins/your-plugin-name
```

### 2. Agent Functionality Test

```bash
# Test each agent
# Verify tools work as expected
# Check output is correct
# Confirm documentation is accurate
```

### 3. Command Execution Test

```bash
# Test each command
/your-command [arguments]

# Verify:
# - Command executes without errors
# - Output matches documentation
# - Arguments work as described
```

### 4. Validation Tests

```bash
# Validate JSON manifests
make validate

# Or manually:
python3 -m json.tool .claude-plugin/marketplace.json
python3 -m json.tool plugins/your-plugin-name/.claude-plugin/plugin.json
```

### 5. Integration Test

```bash
# Test with dependencies
/plugin install core-workflow
/plugin install your-plugin-name

# Verify:
# - Dependencies load correctly
# - Agents can use dependency features
# - No conflicts between plugins
```

### 6. Token Budget Verification

```bash
# Install plugin and check context usage
# Verify estimated tokens ≈ actual usage
# Update plugin.json if significantly different
```

## Submission Process

### 1. Create Feature Branch

```bash
git checkout -b add-your-plugin-name
```

### 2. Commit Changes

```bash
git add plugins/your-plugin-name/
git add .claude-plugin/marketplace.json

git commit -m "feat(marketplace): add your-plugin-name plugin

## Context
Adding plugin for [purpose] to help users [benefit].

## Solution
Created plugin with:
- X agents for [use case]
- Y commands for [use case]
- Z skills for [use case]

## Testing
- ✅ Local installation test passed
- ✅ All agents tested and working
- ✅ All commands tested and working
- ✅ JSON validation passed
- ✅ Token budget verified (~X tokens)

## Impact
Enables users to [specific capability].

```

### 3. Push and Create PR

```bash
git push origin add-your-plugin-name
```

Then create a Pull Request on GitHub with:

**Title:** `feat(marketplace): add [plugin-name] plugin`

**Description:**
```markdown
## Plugin Information

- **Name:** your-plugin-name
- **Version:** 1.0.0
- **Description:** [Brief description]
- **Token Budget:** ~X tokens (X%)

## Features

- [Feature 1]
- [Feature 2]
- [Feature 3]

## Testing Checklist

- [ ] Local installation successful
- [ ] All agents tested
- [ ] All commands tested
- [ ] JSON validation passed
- [ ] Token budget verified
- [ ] Documentation complete
- [ ] Examples provided
- [ ] README included

## Related Issues

Closes #X (if applicable)

## Screenshots/Examples

[Include usage examples or screenshots]
```

### 4. PR Review Process

**We will review:**
- Plugin structure and completeness
- Code quality and security
- Documentation quality
- Token budget accuracy
- Test coverage
- Adherence to standards

**Review timeline:**
- Initial review: 2-3 business days
- Feedback provided if changes needed
- Approval once all requirements met

### 5. After Merge

Once merged:
- Your plugin is available in the marketplace
- Users can install with: `/plugin install your-plugin-name@agentops`
- You'll be credited as author
- Consider writing a blog post or tutorial

## Code of Conduct

### Our Standards

**Positive behavior:**
- Be respectful and inclusive
- Provide constructive feedback
- Collaborate openly
- Welcome newcomers
- Share knowledge generously

**Unacceptable behavior:**
- Harassment or discrimination
- Trolling or insulting comments
- Personal or political attacks
- Publishing others' private information
- Other unprofessional conduct

### Enforcement

Violations may result in:
1. Warning from maintainers
2. Temporary ban from contributing
3. Permanent ban from project

Report issues to: fullerbt@users.noreply.github.com

## Getting Help

### Questions?

- **Documentation:** Check README.md and existing plugins
- **Examples:** Browse `plugins/core-workflow/` for reference
- **GitHub Issues:** Search existing issues or create new one
- **GitHub Discussions:** Ask questions, share ideas
- **Email:** fullerbt@users.noreply.github.com

### Useful Resources

**Official Documentation:**
- [Claude Code Plugins](https://docs.claude.com/en/docs/claude-code/plugins)
- [Plugin Marketplaces](https://docs.claude.com/en/docs/claude-code/plugin-marketplaces)
- [Subagents Guide](https://docs.claude.com/en/docs/claude-code/sub-agents)

**Example Plugins:**
- [core-workflow](plugins/core-workflow/) - Universal workflow
- [devops-operations](plugins/devops-operations/) - DevOps automation
- [software-development](plugins/software-development/) - Software dev tools

**12-Factor AgentOps:**
- [Framework](https://github.com/boshu2/12-factor-agentops)
- [Showcase](https://agentops-showcase.com)

## Maintainer Guidelines

### For Repository Maintainers

**Reviewing PRs:**
1. Check plugin structure completeness
2. Verify JSON manifests valid
3. Test local installation
4. Review documentation quality
5. Verify token budget reasonable
6. Check for security issues
7. Provide constructive feedback

**Merging:**
- Require all checks passed
- Require approval from at least 1 maintainer
- Use squash merge for cleaner history
- Update CHANGELOG.md

**Communication:**
- Respond to PRs within 2-3 business days
- Be welcoming and supportive
- Provide clear, actionable feedback
- Thank contributors for their work

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache License 2.0.

---

**Thank you for contributing to AgentOps marketplace! 🚀**

**Questions?** Open an issue or email fullerbt@users.noreply.github.com
