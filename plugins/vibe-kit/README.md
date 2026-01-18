# vibe-kit

A lean, production-ready Claude Code setup built on the 40% rule.

**Philosophy:** Simple beats clever. Skills replaced most agents. Complexity is where tokens go to die.

## What's Included

| Component | Count | Description |
|-----------|-------|-------------|
| Skills | 15 | General workflow + utilities |
| Agents | 4 | Parallel specialist review: security, architecture, quality, UX |
| Template | 1 | CLAUDE.md starter template |

**Note:** For Gas Town orchestration skills (beads, dispatch, vibe, etc.), install the [gastown plugin](../gastown/).

## Quick Start

```bash
# Clone to your plugins directory
git clone https://github.com/boshu2/agentops.git ~/.claude/plugins/agentops

# Copy vibe-kit to your .claude/
cp -r ~/.claude/plugins/agentops/plugins/vibe-kit/* ~/.claude/

# Or symlink for updates
ln -s ~/.claude/plugins/agentops/plugins/vibe-kit/commands/* ~/.claude/commands/
ln -s ~/.claude/plugins/agentops/plugins/vibe-kit/skills/* ~/.claude/skills/
ln -s ~/.claude/plugins/agentops/plugins/vibe-kit/agents/* ~/.claude/agents/
```

## The 40% Rule

The most important thing in this plugin:

- **Below 40% context** → 98% success rate
- **Above 60% context** → 24% success rate

Above 40%, the model doesn't degrade. It lies.

### How to stay under 40%

1. Don't load everything at startup
2. Use skills with JIT (just-in-time) loading
3. Compact frequently—write summaries to files, start fresh sessions
4. Kill agents that return too much context

## Skills

Skills auto-trigger on phrases or can be invoked directly with `/skill-name`.

### Core Workflow

| Skill | Invoke | Triggers |
|-------|--------|----------|
| `research` | `/research` | "research this", "explore the codebase" |
| `implement` | `/implement` | "implement this", "work on task" |
| `implement-wave` | `/implement-wave` | "run a wave", "parallel implementation" |
| `crank` | `/crank` | "execute this", "run crank", "autonomous execution" |
| `retro` | `/retro` | "run retrospective", "extract learnings" |
| `plan` | `/plan` | "create a plan", "plan implementation" |
| `formulate` | `/formulate` | "create a formula", "formulate this" |
| `product` | `/product` | "product brief", "who is this for", "PR/FAQ" |

### Utilities

| Skill | Invoke | Triggers |
|-------|--------|----------|
| `bug-hunt` | `/bug-hunt` | "investigate bug", "find root cause" |
| `validation-chain` | `/validation-chain` | "validate changes", "run validation gate" |
| `doc` | `/doc` | "generate docs", "create documentation" |
| `oss-docs` | `/oss-docs` | "add OSS docs", "create README" |
| `complexity` | `/complexity` | "analyze complexity", "find refactor targets" |
| `golden-init` | `/golden-init` | "initialize repo", "golden template" |
| `molecules` | `/molecules` | "workflow template", "formula" |

## Agents

Four domain experts for parallel review:

| Agent | Focus |
|-------|-------|
| `security-expert` | OWASP Top 10, vulnerabilities |
| `architecture-expert` | System design, cross-cutting concerns |
| `code-quality-expert` | Complexity, patterns |
| `ux-expert` | Accessibility, user-facing |

**When to use agents:** Only for parallel specialist review before merge. If you're not doing parallel review, you probably don't need custom agents.

## CLAUDE.md Template

Copy `CLAUDE.md.template` to `~/.claude/CLAUDE.md` and adapt it:

```bash
cp ~/.claude/plugins/agentops/plugins/vibe-kit/CLAUDE.md.template ~/.claude/CLAUDE.md
```

Keep it under 200 lines. Anything longer and you're wasting context.

## Related

- **[12-Factor AgentOps](https://12factoragentops.com)** - The methodology
- **[AgentOps Marketplace](https://github.com/boshu2/agentops)** - More plugins
- **[Devlog](https://bodenfuller.com)** - How this evolved

---

*Less is more.*
