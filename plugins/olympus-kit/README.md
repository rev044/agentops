# olympus-kit

**Knowledge Forge** - Mine transcripts, track provenance, compound intelligence.

> **"You sleep. Code ships. Intelligence compounds."**

## The Closed Loop

Every session makes the next one smarter. **Automatically.**

```
Session N:
  work happens → session ends
                     ↓
                 ol forge --queue
                     ↓
              pending.jsonl (queued)

Session N+1:
  ol extract → outputs prompt → Claude extracts learnings
                                         ↓
                                .agents/learnings/
                                         ↓
  ol inject → loads learnings → work informed by past
```

**No manual steps.** The hooks handle everything.

## Quick Start

```bash
# Install
cd plugins/olympus-kit && ./install.sh

# Initialize in your project
cd ~/your-project && ol init

# The hooks do the rest:
# - SessionEnd queues your session
# - SessionStart extracts learnings + injects knowledge
```

## The Problem

```
WITHOUT OLYMPUS:
  Session 1: Debug OAuth refresh (45 min)
  Session 2: Same issue, fresh start (45 min)
  Session 3: Same issue, fresh start (45 min)

WITH OLYMPUS:
  Session 1: Debug issue → learning extracted (45 min)
  Session 2: "Solved this before" (3 min)
  Session 3: Instant recall (1 min)
```

## How the Loop Closes

| Hook | Command | What Happens |
|------|---------|--------------|
| **SessionEnd** | `ol forge --queue` | Parses transcript, queues for extraction |
| **SessionStart** | `ol extract` | Outputs prompt asking Claude to extract learnings |
| **SessionStart** | `ol inject` | Loads learnings into session context |

The key insight: **Claude does the extraction** using its own context at session start. No API keys needed.

## Skills

| Skill | Command | Purpose |
|-------|---------|---------|
| `/forge` | `ol forge transcript` | Mine transcripts for knowledge |
| `/extract` | `ol extract` | Process pending extractions |
| `/inject` | `ol inject` | Load relevant knowledge |
| `/ratchet` | `ol ratchet status` | Track RPI workflow |
| `/flywheel` | `ol metrics` | Knowledge health metrics |
| `/provenance` | `ol ratchet trace` | Trace knowledge lineage |

## CLI Commands

```bash
# Core loop
ol forge transcript --last-session --queue  # Queue session for extraction
ol extract                                   # Output extraction prompt
ol inject                                    # Load knowledge

# Search & status
ol search "authentication"                   # Semantic search
ol ratchet status                           # Workflow progress
ol metrics                                  # Flywheel health

# Management
ol extract --clear                          # Clear pending queue
ol inject --context "auth" --max-tokens 2000
```

## Knowledge Stores

| Location | Content | Written By |
|----------|---------|------------|
| `.agents/learnings/` | Actionable learnings | Claude via `/extract` |
| `.agents/patterns/` | Reusable patterns | `/retro`, manual |
| `.agents/olympus/sessions/` | Session summaries | `ol forge` |
| `.agents/olympus/pending.jsonl` | Extraction queue | `ol forge --queue` |

## The Brownian Ratchet

```
Progress = Chaos × Filter → Ratchet
```

| Phase | What Happens |
|-------|--------------|
| **Chaos** | Multiple parallel attempts |
| **Filter** | Validation gates (tests, review) |
| **Ratchet** | Lock progress permanently |

**Key insight:** You can't un-ratchet. Progress is permanent.

## Hook Configuration

Add to your Claude Code settings:

```json
{
  "hooks": {
    "SessionStart": [
      {"type": "command", "command": "ol extract 2>/dev/null || true"},
      {"type": "command", "command": "ol inject --format markdown --max-tokens 1000 2>/dev/null || true"}
    ],
    "SessionEnd": [
      {"type": "command", "command": "ol forge transcript --last-session --queue --quiet 2>/dev/null || true"}
    ]
  }
}
```

## The Compounding Effect

| Timeline | Claude Knows |
|----------|--------------|
| Day 1 | Nothing - fresh start |
| Week 1 | Your recent debugging sessions |
| Month 1 | Your codebase patterns |
| Month 3 | Your organization's history |

## Version

0.1.0

## License

MIT
