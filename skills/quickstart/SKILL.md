---
name: quickstart
description: 'Interactive onboarding for new AgentOps users. Guided RPI cycle on your actual codebase in under 10 minutes. Triggers: "quickstart", "get started", "onboarding", "how do I start".'
dependencies: []
---

# /quickstart — Get Started with AgentOps

> **Purpose:** Walk a new user through their first Research-Plan-Implement cycle on their actual codebase. Under 10 minutes to first value.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** None required. All external CLIs (bd, ao, gt) are optional enhancements.

## Execution Steps

### Step 1: Detect Project

```bash
# Detect language/framework
ls *.py setup.py pyproject.toml requirements.txt 2>/dev/null && echo "Python detected"
ls *.go go.mod go.sum 2>/dev/null && echo "Go detected"
ls *.ts *.tsx tsconfig.json package.json 2>/dev/null && echo "TypeScript detected"
ls *.sh Makefile Dockerfile 2>/dev/null && echo "Shell/Infra detected"

# Check git state
git log --oneline -5 2>/dev/null
git diff --stat HEAD~3 2>/dev/null | tail -5

# Check for existing AgentOps setup
ls .agents/ 2>/dev/null && echo "AgentOps artifacts found"
ls .beads/ 2>/dev/null && echo "Beads issue tracking found"
```

### Step 2: Welcome and Orient

Present this to the user:

```
Welcome to AgentOps! Here are the 3 skills that matter most:

  /research  — Deep dive into your codebase to understand it
  /plan      — Break a goal into trackable issues
  /vibe      — Validate code before shipping

Let's do a quick tour using YOUR code.
```

### Step 3: Mini Research

Run a focused research pass on the most recently changed area:

```bash
# Find what changed recently
git diff --name-only HEAD~5 2>/dev/null | head -10
```

Read 2-3 of the most recently changed files. Provide a brief summary:
- What area of the codebase is active
- What patterns are used
- One observation about code quality

Tell the user: "This is what `/research` does — deep exploration of your codebase. Use it before planning any significant work."

### Step 4: Mini Plan

Based on the research, suggest ONE concrete improvement:

```
Based on what I found, here's a task we could plan:

  "<specific improvement based on what was found>"

This is what /plan does — decomposes goals into trackable issues with
dependencies and waves.
```

### Step 5: Mini Vibe Check

Run a quick validation on recent changes:

```bash
# Get recent changes for vibe check
git diff --name-only HEAD~3 2>/dev/null | head -10
```

Perform a brief inline review (similar to `/council --quick`) of the most recent changes:
- Check for obvious issues
- Note any complexity concerns
- Provide a quick PASS/WARN/FAIL assessment

Tell the user: "This is what `/vibe` does — complexity analysis + multi-model council review. Use it before committing significant changes."

### Step 6: Show Available Tools

```
You've just completed a mini RPI cycle:
  Research → Plan → Validate

Here's what else is available:

  IMPLEMENT                    VALIDATE                 COLLABORATE
  /implement  - execute task   /vibe      - code check  /council   - multi-model review
  /crank      - run full epic  /pre-mortem - plan check  /swarm     - parallel agents
  /plan       - decompose work /post-mortem - wrap up    /handoff   - session handoff

  EXPLORE                      TRACK
  /research   - deep dive      /knowledge - query learnings
  /bug-hunt   - investigate    /inbox     - agent messages
  /complexity - code metrics   /trace     - decision history
  /doc        - generate docs  /retro     - extract learnings
```

### Step 7: Suggest Next Steps

Based on project state, suggest the most useful next action:

| State | Suggestion |
|-------|------------|
| Recent commits, no tests | "Try `/vibe recent` to check your latest changes" |
| Open issues/TODOs | "Try `/plan` to decompose a goal into trackable issues" |
| Complex codebase, new to it | "Try `/research <area>` to understand a specific area" |
| Bug reports or failures | "Try `/bug-hunt` to investigate systematically" |
| Clean state, looking for work | "Try `/research` to find improvement opportunities" |

---

## See Also

- `skills/vibe/SKILL.md` — Code validation
- `skills/research/SKILL.md` — Codebase exploration
- `skills/plan/SKILL.md` — Epic decomposition
