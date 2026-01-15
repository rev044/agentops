---
description: Extract learnings from completed work
version: 1.0.0
argument-hint: [topic-or-plan-file]
model: opus
---

# /retro

Invoke the **sk-retro** skill to extract learnings, patterns, and insights from completed work.

## Arguments

| Argument | Purpose | Default |
|----------|---------|---------|
| `topic` | Topic, plan file, or epic ID to analyze | Recent closed beads |

## Execution

This command invokes the `sk-retro` skill with the provided arguments.

The skill handles:
- Context gathering from git, beads, blackboard, and **conversation analysis**
- Friction detection and improvement proposals
- Supersession checking for existing artifacts
- Auto-update tier classification
- User review and change application
- Retro summary, learnings, and pattern extraction
- Memory storage to ai-platform (knowledge flywheel)

## Conversation Analysis (Knowledge Flywheel)

When a session ID is available, retro automatically analyzes the conversation:

```bash
python3 ~/.claude/scripts/analyze-sessions.py --session=$SESSION_ID --limit=50
```

This extracts decisions, friction, and patterns from the Claude Code chat history
and feeds them into the retro output artifacts.

## Outputs

| Output | Location |
|--------|----------|
| Retro summary | `.agents/retros/YYYY-MM-DD-{topic}.md` |
| Learnings | `.agents/learnings/YYYY-MM-DD-{topic}.md` |
| Patterns | `.agents/patterns/{pattern-name}.md` |

## Related

- **Skill**: `~/.claude/skills/sk-retro/SKILL.md`
- **Patterns**: `~/.claude/patterns/commands/retro/` (legacy, now in skill)
- **Standard**: `~/.claude/standards/command-skill-architecture.md`
