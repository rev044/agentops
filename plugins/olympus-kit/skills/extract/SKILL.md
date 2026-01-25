# Extract Skill

Process pending learning extractions from previous sessions.

## How It Works

This skill closes the knowledge loop by processing sessions queued for learning extraction.

```
Session N ends:
  → ao forge --last-session --queue
  → Session queued in .agents/ao/pending.jsonl

Session N+1 starts:
  → ao extract (this skill)
  → Outputs extraction prompt
  → Claude extracts learnings
  → Writes to .agents/learnings/
  → Loop closed
```

## Triggers

- SessionStart hook (automatic)
- "extract learnings"
- "process pending"

## Usage

```bash
# Check for pending extractions and output prompt
ao extract

# Clear pending queue without processing
ao extract --clear

# Limit content size in prompt
ao extract --max-content 4000
```

## What It Outputs

When pending extractions exist, outputs a structured prompt:

```markdown
---
# Knowledge Extraction Request

A previous session has been queued for learning extraction.

## Session Context
- **Session ID**: abc123
- **Summary**: Debugged OAuth token refresh issue

## Key Decisions
- Chose to use Redis for token storage
- ...

## Your Task
Extract **1-3 actionable learnings** and write to:
.agents/learnings/2026-01-25-abc123.md

[Format instructions...]
---
```

## The Closed Loop

| Step | Command | Output |
|------|---------|--------|
| Session ends | `ao forge --queue` | Queues session |
| Next session starts | `ao extract` | Outputs prompt |
| Claude processes | (automatic) | Writes learnings |
| Knowledge injected | `ao inject` | Loads learnings |

## Hook Configuration

```json
{
  "SessionStart": [
    {"command": "ao extract"},
    {"command": "ao inject"}
  ],
  "SessionEnd": [
    {"command": "ao forge transcript --last-session --queue --quiet"}
  ]
}
```

## See Also

- `/forge` - Mine transcripts
- `/inject` - Load knowledge
