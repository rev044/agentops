# How Compounding Works

Two mechanisms feed the loop:

**1. Knowledge flywheel (each cycle is smarter):**
```
Session 1:
  ao inject (nothing yet)         → cycle runs blind
  /rpi fixes test-pass-rate       → post-mortem runs ao forge
  Learnings extracted: "tests/skills/run-all.sh validates frontmatter"

Session 2:
  ao inject (loads Session 1 learnings)  → cycle knows about frontmatter validation
  /rpi fixes doc-coverage                → approach informed by prior learning
  Learnings extracted: "references/ dirs need at least one .md file"
```

**2. Work harvesting (each cycle discovers the next):**
```
Cycle 1: /rpi fixes test-pass-rate
  → post-mortem harvests: "add missing smoke test for /evolve" → next-work.jsonl

Cycle 2: all GOALS.yaml goals pass
  → /evolve reads next-work.jsonl → picks "add missing smoke test"
  → /rpi fixes it → post-mortem harvests: "update SKILL-TIERS count"

Cycle 3: reads next-work.jsonl → picks "update SKILL-TIERS count" → ...
```

The loop keeps running as long as post-mortem keeps finding follow-up work. Each /rpi cycle generates next-work items from its own post-mortem. The system feeds itself.

**Priority cascade:**
```
GOALS.yaml goals (explicit, human-authored)  → fix these first
next-work.jsonl (harvested from post-mortem) → work on these when goals pass
nothing left                                 → stop (human re-evaluates)
```
