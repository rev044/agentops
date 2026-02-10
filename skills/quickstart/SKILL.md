---
name: quickstart
tier: solo
description: 'Interactive onboarding for new AgentOps users. Guided RPI cycle on your actual codebase in under 10 minutes. Triggers: "quickstart", "get started", "onboarding", "how do I start".'
dependencies: []
---

# /quickstart — Get Started with AgentOps

> **Purpose:** Walk a new user through their first Research-Plan-Implement cycle on their actual codebase. Under 10 minutes to first value.

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** None required. All external CLIs (bd, ao, gt) are optional enhancements.

**References (load as needed):**
- `references/getting-started.md` — Detailed first-time walkthrough
- `references/troubleshooting.md` — Common issues and fixes

## Execution Steps

### Step 0: Pre-flight

Check environment before starting. Failures here are informational, not blocking.

```bash
# 1. Git repo check
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "GIT_REPO=true"
else
  echo "GIT_REPO=false"
  echo "Not a git repo. Some features (recent changes, vibe) need git."
  echo "Options: run 'git init' to enable full features, or continue in manual mode."
fi

# 2. ao CLI availability
if command -v ao &>/dev/null; then
  echo "AO_CLI=true"
  ao status 2>/dev/null | head -3
else
  echo "AO_CLI=false — optional, enables persistent knowledge flywheel"
fi

# 3. .agents/ directory
if [ -d ".agents" ]; then
  echo "AGENTS_DIR=exists"
elif mkdir -p .agents 2>/dev/null; then
  echo "AGENTS_DIR=created"
  rmdir .agents 2>/dev/null  # clean up test dir
else
  echo "AGENTS_DIR=no_write — cannot create .agents/ directory (check permissions)"
fi

# 4. Claude Code version (informational)
claude --version 2>/dev/null || echo "Claude Code version: unknown"
```

**If GIT_REPO=false:** Continue the walkthrough but skip git-dependent steps (Steps 3, 5). Replace them with file-browsing equivalents. Tell the user which steps were adapted and why.

### Step 1: Detect Project

```bash
# Detect language/framework
ls *.py setup.py pyproject.toml requirements.txt 2>/dev/null && echo "Python detected"
ls *.go go.mod go.sum 2>/dev/null && echo "Go detected"
ls *.ts *.tsx tsconfig.json package.json 2>/dev/null && echo "TypeScript detected"
ls *.rs Cargo.toml 2>/dev/null && echo "Rust detected"
ls *.java pom.xml build.gradle 2>/dev/null && echo "Java detected"
ls *.sh Makefile Dockerfile 2>/dev/null && echo "Shell/Infra detected"
```

**If no language detected:** Tell the user: "I couldn't auto-detect a language. What is the primary language of this project? (Python, Go, TypeScript, Rust, Java, Shell, or other)" Then continue with whatever they choose.

```bash
# Check git state (skip if GIT_REPO=false)
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

**Graduation hints** (state-aware, based on pre-flight + detection):

```bash
# Gather state from pre-flight (Step 0)
command -v ao &>/dev/null && AO_AVAILABLE=true || AO_AVAILABLE=false
command -v bd &>/dev/null && BD_AVAILABLE=true || BD_AVAILABLE=false
command -v codex &>/dev/null && CODEX_AVAILABLE=true || CODEX_AVAILABLE=false
ls .agents/ &>/dev/null && AGENTS_DIR=true || AGENTS_DIR=false
ls .beads/ &>/dev/null && BEADS_DIR=true || BEADS_DIR=false
git rev-parse --is-inside-work-tree &>/dev/null && GIT_REPO=true || GIT_REPO=false
```

**Present ONLY the row matching current state (do not show all tiers):**

| Current State | Tier | Next Step |
|---------------|------|-----------|
| No git repo | — | "Initialize git with `git init` to unlock change tracking, `/vibe`, and full RPI workflow." |
| Git repo, no `ao`, no `.agents/` | Tier 0 | "You're at Tier 0 — skills work standalone. When you want learnings to persist across sessions, install the `ao` CLI: `brew install agentops && ao hooks install`" |
| `ao` installed, no `.agents/` yet | Tier 0+ | "Run `ao init` to create the `.agents/` directory. Then your knowledge flywheel starts capturing learnings automatically." |
| `ao` + `.agents/`, no beads | Tier 1 | "Knowledge flywheel is active. When you have multi-issue epics, add beads for issue tracking: `brew install beads && bd init --prefix <your-prefix>`" |
| `ao` + beads, no Codex | Tier 2 | "Full RPI stack. Try `/crank` for autonomous epic execution, or `/council --deep` for thorough multi-judge review." |
| `ao` + beads + Codex | Tier 2+ | "Full stack with cross-vendor. Try `/council --mixed` for Claude + Codex consensus, or `/vibe --mixed` for cross-vendor code review." |

---

## See Also

- `skills/vibe/SKILL.md` — Code validation
- `skills/research/SKILL.md` — Codebase exploration
- `skills/plan/SKILL.md` — Epic decomposition
