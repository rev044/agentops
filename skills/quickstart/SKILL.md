---
name: quickstart
description: 'Interactive onboarding for new users. Orient in under 2 minutes: skill map, intent router, flywheel status, next action. Triggers: "quickstart", "get started", "onboarding", "how do I start".'
skill_api_version: 1
metadata:
  tier: session
  dependencies: []
---

# /quickstart — Get Started

> **Purpose:** Orient a new user in under 2 minutes. Show what skills are available, how they compose, and what to do next. No demos — that's what the actual skills are for.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** None required. All external CLIs (bd, ao) are optional enhancements.

## Execution Steps

### Step 1: Pre-flight

```bash
git rev-parse --is-inside-work-tree >/dev/null 2>&1 && echo "GIT=true" || echo "GIT=false"
command -v ao >/dev/null && echo "AO=true" || echo "AO=false"
command -v bd >/dev/null && echo "BD=true" || echo "BD=false"
```

Record GIT, AO, BD — used in Step 4 for tier selection. That's it.

### Step 2: Orient

Present the welcome, then the skill map and intent router (text only — no bash):

```
AgentOps gives your coding agent three things it doesn't have:

  Memory    — Every session extracts learnings into .agents/.
              Next session, the best ones are injected automatically.
              Session 50 knows what session 1 learned.

  Judgment  — /council spawns independent judges (Claude + Codex)
              to validate plans and code before shipping.

  Skills    — Standalone primitives you use as-needed:
              /research, /plan, /vibe, /brainstorm, and more.
              Parallelize any of them with /swarm.
```

Show how skills compose:

```
                          YOU
                           │
            ┌──────────────┼──────────────┐
            │              │              │
       use one skill    compose a few   /rpi "goal"
       by itself        your way        full pipeline
            │              │              │
            ▼              ▼              ▼
   ┌─────────────────────────────────────────────────────┐
   │              HOW SKILLS COMPOSE                     │
   │                                                     │
   │  JUDGMENT (the foundation)                          │
   │  /council ──────► spawns independent judges         │
   │  /vibe ─────────► /complexity + /council            │
   │  /pre-mortem ───► /council (simulate failures)      │
   │  /post-mortem ──► /council + /retro                 │
   │                                                     │
   │  EXECUTION                                          │
   │  /research ─────► may trigger /brainstorm           │
   │  /plan ─────────► may call /pre-mortem to validate  │
   │  /implement ────► /research + /plan + build + /vibe │
   │  /crank ────────► /swarm ──► /implement (×N per     │
   │                   wave, fresh context each)         │
   │  /swarm ────────► parallelize any skill             │
   │                                                     │
   │  PIPELINE                                           │
   │  /rpi chains:  research → plan → pre-mortem →       │
   │                crank → vibe → post-mortem            │
   │  /evolve loops /rpi against fitness goals            │
   └─────────────────────────────────────────────────────┘
            │
            ▼
   ┌─────────────────┐
   │    .agents/     │  Append-only ledger.
   │    learnings    │  Every session writes.
   │    patterns     │  Freshness decay prunes.
   │    decisions    │  Next session injects the best.
   └─────────────────┘

QUICK REFERENCE
/research     explore and understand code
/council      independent judges validate plans or code
/vibe         code quality review (complexity + council)
/plan         break down a goal into tasks
/implement    execute a single task end-to-end
/crank        run a multi-issue epic in parallel waves
/swarm        parallelize any skill
/rpi          full pipeline — one command
/status       see what you're working on

INTENT ROUTER
What are you trying to do?
│
├─ "Not sure what to do yet"
│   └─ Generate options first ─────► /brainstorm
│
├─ "I have an idea"
│   └─ Understand code + context ──► /research
│
├─ "I know what I want to build"
│   └─ Break it into issues ───────► /plan
│
├─ "Now build it"
│   ├─ Small/single issue ─────────► /implement
│   ├─ Multi-issue epic ───────────► /crank <epic-id>
│   └─ Full flow in one command ───► /rpi "goal"
│
├─ "Fix a bug"
│   ├─ Know which file? ──────────► /implement <issue-id>
│   └─ Need to investigate? ──────► /bug-hunt
│
├─ "Build a feature"
│   ├─ Small (1-2 files) ─────────► /implement
│   ├─ Medium (3-6 issues) ───────► /plan → /crank
│   └─ Large (7+ issues) ─────────► /rpi (full pipeline)
│
├─ "Validate something"
│   ├─ Code ready to ship? ───────► /vibe
│   ├─ Plan ready to build? ──────► /pre-mortem
│   ├─ Work ready to close? ──────► /post-mortem
│   └─ Quick sanity check? ───────► /council --quick validate
│
├─ "Explore or research"
│   ├─ Understand this codebase ──► /research
│   ├─ Compare approaches ────────► /council research <topic>
│   └─ Generate ideas ────────────► /brainstorm
│
├─ "Learn from past work"
│   ├─ What do we know about X? ──► ao search "<query>"
│   ├─ Save this insight ─────────► /retro --quick "insight"
│   └─ Run a retrospective ───────► /retro
│
├─ "Parallelize work"
│   ├─ Multiple independent tasks ► /swarm
│   └─ Full epic with waves ──────► /crank <epic-id>
│
├─ "Ship a release"
│   └─ Changelog + tag ──────────► /release <version>
│
├─ "Session management"
│   ├─ Where was I? ──────────────► /status
│   ├─ Save for next session ─────► /handoff
│   └─ Recover after compaction ──► /recover
│
└─ "First time here" ────────────► /quickstart
```

### Step 3: Flywheel

```bash
count=$(ls .agents/learnings/ 2>/dev/null | wc -l | tr -d ' ')
if [ "$count" -gt 0 ]; then
  echo "Flywheel: $count learnings accumulated"
else
  echo "Flywheel: no learnings yet — first session"
fi
```

Explain what this means:

```
The flywheel turns automatically through hooks:

  SessionStart:  MEMORY.md auto-loaded (your knowledge is always there)
  During work:   Skills invoke ao commands at the right moments
  SessionEnd:    forge → pool ingest → notebook update → maturity scan
                 → dedup → contradict → expire/evict → prune
  Stop hook:     flywheel close-loop → citation feedback → promote

What it means for you:
  - Session 2 already knows what Session 1 learned (via MEMORY.md)
  - Useful learnings get promoted, stale ones decay and get archived
  - You never run ao commands directly — skills and hooks handle it

Verify any time:
  ao flywheel status             ← escape velocity check
  ao status                      ← current knowledge inventory
  ls .agents/learnings/          ← raw learning files
```

If no learnings yet, tell the user:

```
No learnings yet — that's expected on your first session.
Run /rpi "a small goal" to complete one full cycle.
The session-end hook will extract learnings automatically.
Next session, they'll be in MEMORY.md.
```

### Step 4: What's Next

Based on GIT/AO/BD from Step 1, show the ONE matching tier row:

| Current State | Tier | Next Step |
|---------------|------|-----------|
| GIT=false | — | "Initialize git with `git init` to unlock change tracking, `/vibe`, and full RPI workflow." |
| GIT=true, AO=false | Tier 0 | "Skills work standalone. For persistent learnings across sessions: `brew tap boshu2/agentops https://github.com/boshu2/homebrew-agentops && brew install agentops && ao seed && ao init --hooks`" |
| AO=true, no `.agents/` yet | Tier 0+ | "Run `ao seed` to bootstrap the flywheel, then `ao init --hooks` for SessionStart + SessionEnd automation." |
| AO=true, `.agents/` present, BD=false | Tier 1 | "Flywheel active. Add issue tracking: `brew install boshu2/agentops/beads && bd init --prefix <prefix>`" |
| AO=true, BD=true | Tier 2 | "Full RPI stack. Try `bd ready` (find work) or `/rpi \"goal\"` (full pipeline)." |

Then suggest one concrete next action based on project state:

| Project State | Try |
|---------------|-----|
| Recent commits | `/vibe recent` |
| Open issues | `/status` or `bd ready` |
| New to codebase | `/research <area>` |
| Bug reports | `/bug-hunt` |
| Ready to build | `/rpi "goal"` |

---

## Examples

### New User

**User says:** `/quickstart`

**What happens:**
1. Pre-flight: `GIT=true AO=false`
2. Welcome + skill map + intent router presented (text only, instant)
3. Flywheel: "no learnings yet — first session"
4. Tier 0 hint: install `ao` CLI

**Result:** Oriented in under 60 seconds. One clear next action.

### Returning User

**User says:** `/quickstart`

**What happens:**
1. Pre-flight: `GIT=true AO=true BD=true`
2. Skill map + intent router
3. Flywheel: "42 learnings accumulated"
4. Tier 2 hint: "Try `bd ready` or `/rpi \"goal\"`"

**Result:** Quick reorientation with one concrete next action.

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Wrong tier shown | Check AO/BD values from Step 1 output |
| Flywheel count is 0 | First session — run `/rpi "a small goal"` to start it |
| Skills not available | `npx skills@latest add boshu2/agentops --all -g` |

---

## See Also

- `skills/vibe/SKILL.md` — Code validation
- `skills/research/SKILL.md` — Codebase exploration
- `skills/plan/SKILL.md` — Epic decomposition

## Reference Documents

- [references/getting-started.md](references/getting-started.md)
- [references/troubleshooting.md](references/troubleshooting.md)
- [references/full-catalog.md](references/full-catalog.md)
