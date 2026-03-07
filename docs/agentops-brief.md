# AgentOps — One-Page Brief

**The problem:** AI coding tools behave like contractors with amnesia. Every session starts from zero — no memory of what broke last week, no record of decisions already made, no awareness of what was tried and abandoned. You brief them today. Tomorrow you brief them again.

---

## What It Is

A standard operating procedure system for AI coding tools.

AgentOps gives every AI session a structured briefing before work begins, enforces a consistent workflow from research through validation, requires independent review before anything is committed, and writes a debrief at the end that feeds the next session's briefing.

The institutional knowledge stops walking out the door.

---

## Four Components

### Skills — The SOPs
Structured, numbered checklists the AI follows for every task type: exploring a codebase, breaking work into tracked tasks, reviewing code before it ships, running a postmortem. ~40 workflows covering the full cycle. The AI does not improvise — it follows the procedure.

### Hooks — The Enforcement Layer
Shell scripts that fire automatically at lifecycle events: session start, session end, after every tool call, before every commit. They inject the briefing, harvest the debrief, and enforce quality gates that cannot be bypassed by an inattentive AI.

### `ao` CLI — The Memory System
A command-line tool that manages knowledge across sessions. At session end it extracts learnings (one to three sentences each). At session start it injects the most relevant, freshest ones into the briefing. Learnings that keep proving useful are promoted; stale ones decay automatically. Sessions compound.

### `.agents/` — The Knowledge Base
Plain files on disk. Learnings, patterns, research outputs, review verdicts, release notes. Each learning carries a utility score, a confidence level, a maturity rating, and a decay timestamp. Nothing proprietary — audit every file with a text editor.

---

## How a Session Works

```
Session starts
  → Prior knowledge injected into briefing automatically
  → AI follows structured SOP for the task

AI works
  → Automated gates enforce quality on every tool call
  → Independent reviewers validate before anything ships

Session ends
  → Learnings extracted and scored
  → Knowledge base updated

Next session starts smarter than this one did.
```

---

## Key Properties

| Property | Detail |
|----------|--------|
| **Local-only** | No telemetry, no cloud, no vendor accounts. Nothing phones home. |
| **Open source** | Every line auditable. Apache 2.0 licensed. |
| **Multi-tool** | Works with Claude Code, Codex, Cursor, OpenCode. Not locked to one vendor. |
| **Air-gap compatible** | Runs fully offline. Knowledge base is plain files. |
| **Auditable trail** | Every learning, decision, and review verdict written to `.agents/` with timestamps. |

---

## The Compound Effect

```
Without AgentOps:  [2 hrs] → [2 hrs] → [2 hrs] → [2 hrs]  =  8 hours total
With AgentOps:     [2 hrs] → [10 min] → [2 min] → instant  =  ~2.2 hours total
                    learn     recall     refine    mastered
```

By session 100, the AI already knows every bug fixed in this codebase, every architectural decision and the reasoning behind it, and every approach that failed.

---

## DevOps Parallel

DevOps applied one insight to infrastructure: reliability comes from feedback loops, not from better operators. Postmortems, runbooks, shift-left testing, continuous validation — each revolution of the loop made the next incident cheaper to handle.

AgentOps applies the same insight one layer up. Not better AI models. Better feedback loops around the models you already have.

---

*AgentOps — github.com/boshu2/agentops*

---

## Appendix: System Map

### Scale

```
┌──────────────────────────────────────────────────────────────────┐
│                    AgentOps at a Glance                          │
├───────────────────┬──────────────────────┬───────────────────────┤
│    49  Skills     │  121 CLI Commands    │    14 Hooks           │
│  (workflows)      │  (ao binary)         │  (auto-enforcement)   │
└───────────────────┴──────────────────────┴───────────────────────┘
```

### The Pipeline — Skills Calling Skills

`/rpi` chains the full workflow. Each node is a skill. Arrows show calls.

```
                         ┌─────────────┐
                         │   /evolve   │  ← loops /rpi overnight
                         └──────┬──────┘    fitness-gated
                                │
                                ▼
┌───────────────────────────────────────────────────────────────────┐
│                             /rpi                                  │
│                    (full pipeline orchestrator)                   │
└──┬──────────┬───────────┬─────────────┬──────────┬───────────────┘
   │          │           │             │          │
   ▼          ▼           ▼             ▼          ▼
/research   /plan    /pre-mortem     /crank    /post-mortem
                          │             │          │
                     calls /council     │          ├── calls /council
                                        │          └── calls /retro
                                        │
                               spawns N parallel
                                        │
                                   /implement  × N
                                        │
                                     /vibe  ── calls /council
                                            ── calls /complexity
```

### Judgment Layer — Everything Flows Through Council

```
                   ┌──────────────────────────────┐
                   │           /council            │
                   │  (independent reviewers       │
                   │   debate, verdict gates work) │
                   └───────────┬──────────────────┘
                               │ used by
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                    ▼
   /pre-mortem              /vibe              /post-mortem
   (validate plans          (validate code     (wrap-up +
    before building)         before shipping)   learnings)
```

### Knowledge Handoff — Skills Calling the CLI

```
   SKILL                   ao CLI COMMAND              RESULT
   ─────                   ──────────────              ──────
/research          →    ao lookup                  Prior knowledge loaded
/retro             →    ao forge transcript        Learnings extracted
/retro             →    ao pool promote            Learnings validated
/evolve            →    ao goals measure           Fitness checked
/rpi               →    ao ratchet record          Progress checkpointed
/implement         →    ao ratchet check           Gate verified
```

### Hooks — Automatic Enforcement

```
TRIGGER                   HOOK                        WHAT IT DOES
───────                   ────                        ────────────
Session starts         session-start.sh            Inject prior knowledge
Session ends           session-end-maintenance.sh  Harvest learnings
Agent stops            ao-flywheel-close.sh        Close learning loop
Every tool call        go-complexity-precommit.sh  Block over-complex code
Pre-commit             skill-lint-gate.sh          Reject malformed SOPs
Pre-commit             dangerous-git-guard.sh      Block force-pushes to main
Pre-commit             pre-mortem-gate.sh          Require review for large changes
Worker stop            subagent-stop.sh            Clean up parallel agent state
```

### CLI Command Groups

```
KNOWLEDGE FLYWHEEL          VALIDATION GATES         SESSION / LIFECYCLE
──────────────────          ────────────────         ───────────────────
ao forge                    ao gate pending          ao session close
ao pool ingest              ao gate approve          ao rpi status
ao pool promote             ao gate reject           ao hooks list
ao lookup                   ao ratchet status        ao config
ao lookup                   ao ratchet record
ao search                   ao ratchet check         METRICS / HEALTH
ao dedup                                             ────────────────
ao curate                   GOALS / FITNESS          ao metrics health
                            ───────────────          ao metrics flywheel
MEMORY TOOLS                ao goals measure         ao metrics report
────────────                ao goals steer           ao flywheel status
ao mind                     ao goals add             ao maturity
ao notebook                 ao goals prune           ao doctor
ao memory                   ao goals history
ao trace                    ao goals drift
ao extract
```
