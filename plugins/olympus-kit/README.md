# olympus-kit

**Knowledge Forge** - Mine transcripts, track provenance, compound intelligence.

> **"You sleep. Code ships. Intelligence compounds."**

## What It Does

AI coding assistants are brilliant but amnesiac. Every session starts from zero. Your team debugs the same issues, rediscovers the same patterns, makes the same mistakes—over and over.

**olympus-kit fixes this.** Every session makes the next one smarter.

```
WITHOUT OLYMPUS:
  Session 1: Debug OAuth refresh (45 min)
  Session 2: Same issue, fresh start (45 min)
  Session 3: Same issue, fresh start (45 min)

WITH OLYMPUS:
  Session 1: Debug issue, capture pattern (45 min)
  Session 2: "Solved this before" (3 min)
  Session 3: Instant recall (1 min)
```

## Installation

```bash
# From agentops plugin directory
cd plugins/olympus-kit
./install.sh

# Verify
ol --version
```

## Quick Start

```bash
# Mine your Claude transcripts
ol forge transcript ~/.claude/projects/**/*.jsonl

# Search what you've learned
ol search "authentication"

# Check workflow progress
ol ratchet status

# Inject knowledge at session start
ol inject
```

## Skills

| Skill | Command | Purpose |
|-------|---------|---------|
| `/forge` | `ol forge transcript` | Mine transcripts for knowledge |
| `/inject` | `ol inject` | Recall relevant knowledge |
| `/ratchet` | `ol ratchet status` | Track RPI workflow progress |
| `/flywheel` | `ol metrics` | Knowledge health metrics |
| `/provenance` | `ol ratchet trace` | Trace knowledge lineage |

## Session Hooks

olympus-kit includes automatic session hooks:

| Hook | Action |
|------|--------|
| **SessionStart** | `ol inject` - Load relevant knowledge |
| **SessionEnd** | `ol forge transcript --last-session` - Extract learnings |

To enable, merge `hooks/hooks.json` into your Claude Code settings.

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

## Knowledge Stores

| Location | Content |
|----------|---------|
| `.agents/learnings/` | Lessons learned |
| `.agents/patterns/` | Reusable patterns |
| `.agents/retros/` | Retrospectives |
| `.agents/olympus/` | Session index, provenance graph |

## CLI Commands

```bash
ol forge transcript <files>   # Mine transcripts
ol forge badge                # Session badge with outcome
ol inject [--context <ctx>]   # Recall knowledge
ol search <query>             # Semantic search
ol ratchet status             # Workflow status
ol ratchet check <step>       # Validate gate
ol ratchet record <step>      # Record completion
ol ratchet trace <step>       # Trace provenance
ol metrics                    # Flywheel health
```

## Version

0.1.0

## License

MIT
