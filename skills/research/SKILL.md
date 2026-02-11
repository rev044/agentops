---
name: research
tier: solo
description: 'Deep codebase exploration. Triggers: research, explore, investigate, understand, deep dive, current state.'
dependencies:
  - knowledge # optional - queries existing knowledge
  - inject    # optional - injects prior context
---

# Research Skill

> **Quick Ref:** Deep codebase exploration with multi-angle analysis. Output: `.agents/research/*.md`

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

**CLI dependencies:** ao (knowledge injection — optional). If ao is unavailable, skip prior knowledge search and proceed with direct codebase exploration.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--auto` | off | Skip human approval gate. Used by `/rpi --auto` for fully autonomous lifecycle. |

## Execution Steps

Given `/research <topic> [--auto]`:

### Step 1: Create Output Directory
```bash
mkdir -p .agents/research
```

### Step 2: Check Prior Art

**First, search and inject existing knowledge (if ao available):**

```bash
# Search knowledge base for relevant learnings, patterns, and prior research
ao search "<topic>" 2>/dev/null || echo "ao not available, skipping knowledge search"

# Inject relevant context into this session
ao inject "<topic>" 2>/dev/null || echo "ao not available, skipping knowledge injection"
```

**Review ao search results:** If ao returns relevant learnings or patterns, incorporate them into your research strategy. Look for:
- Prior research on this topic or related topics
- Known patterns or anti-patterns
- Lessons learned from similar investigations

**Search local research artifacts:**
```bash
ls -la .agents/research/ 2>/dev/null | grep -i "<topic>" || echo "No prior research found"
```

Also use Grep to search `.agents/` for related content. Check TEMPERED learnings:
```bash
ls -la .agents/learnings/ .agents/patterns/ 2>/dev/null | head -10
```

### Step 2.5: Pre-Flight — Detect Spawn Backend

Before launching the explore agent, detect which backend is available:

1. Check if `spawn_agent` is available → log `"Backend: codex-sub-agents"`
2. Else check if `TeamCreate` is available → log `"Backend: claude-native-teams"`
3. Else check if `Task` is available → log `"Backend: background-task-fallback"`
4. Else → log `"Backend: inline (no spawn available)"`

Record the selected backend — it will be included in the research output document for traceability.

### Step 3: Launch Explore Agent

**YOU MUST DISPATCH AN EXPLORATION AGENT NOW.** Select the backend using capability detection:

#### Backend Selection (MANDATORY)

1. If `spawn_agent` is available → **Codex sub-agent**
2. Else if `TeamCreate` is available → **Claude native team** (Explore agent)
3. Else → **Background task fallback**

#### Exploration Prompt (all backends)

Use this prompt for whichever backend is selected:

```
Thoroughly investigate: <topic>

Search strategy:
1. Glob for relevant files (*.md, *.py, *.ts, *.go, etc.)
2. Grep for keywords related to <topic>
3. Read key files and understand the architecture
4. Check docs/ and .agents/ for existing documentation

Return a detailed report with:
- Key files found (with paths)
- How the system works
- Important patterns or conventions
- Any issues or concerns

Cite specific file:line references for all claims.
```

#### Backend 1: Codex Sub-Agents (`spawn_agent` available)

```
agent_id = spawn_agent(message="""
You are a codebase research agent. Your task:

<exploration prompt from above>

Use file reading, grep, and glob to explore the codebase at: <cwd>
Write your findings as structured markdown.
""")

result = wait(ids=[agent_id])
# Extract findings from result
```

#### Backend 2: Claude Native Teams (`TeamCreate` available)

```
Tool: Task
Parameters:
  subagent_type: "Explore"
  description: "Research: <topic>"
  prompt: |
    <exploration prompt from above>
```

#### Backend 3: Background Task Fallback

```
Tool: Task
Parameters:
  subagent_type: "general-purpose"
  run_in_background: true
  description: "Research: <topic>"
  prompt: |
    <exploration prompt from above>
```

Then retrieve results with `TaskOutput(task_id=..., block=true)`.

#### No Backend Available

If none of the above are available, perform the exploration **inline** in the current session using Glob, Grep, and Read tools directly. Log: `"Note: No spawn backend available. Performing inline exploration."`

### Step 4: Validate Research Quality (Optional)

**For thorough research, perform quality validation:**

#### 4a. Coverage Validation
Check: Did we look everywhere we should? Any unexplored areas?
- List directories/files explored
- Identify gaps in coverage
- Note areas that need deeper investigation

#### 4b. Depth Validation
Check: Do we UNDERSTAND the critical parts? HOW and WHY, not just WHAT?
- Rate depth (0-4) for each critical area
- Flag areas with shallow understanding
- Identify what needs more investigation

#### 4c. Gap Identification
Check: What DON'T we know that we SHOULD know?
- List critical gaps
- Prioritize what must be filled before proceeding
- Note what can be deferred

#### 4d. Assumption Challenge
Check: What assumptions are we building on? Are they verified?
- List assumptions made
- Flag high-risk unverified assumptions
- Note what needs verification

### Step 5: Synthesize Findings

After the Explore agent and validation swarm return, write findings to:
`.agents/research/YYYY-MM-DD-<topic-slug>.md`

Use this format:
```markdown
# Research: <Topic>

**Date:** YYYY-MM-DD
**Backend:** <codex-sub-agents | claude-native-teams | background-task-fallback | inline>
**Scope:** <what was investigated>

## Summary
<2-3 sentence overview>

## Key Files
| File | Purpose |
|------|---------|
| path/to/file.py | Description |

## Findings
<detailed findings with file:line citations>

## Recommendations
<next steps or actions>
```

### Step 6: Request Human Approval (Gate 1)

**Skip this step if `--auto` flag is set.** In auto mode, proceed directly to Step 7.

**USE AskUserQuestion tool:**

```
Tool: AskUserQuestion
Parameters:
  questions:
    - question: "Research complete. Approve to proceed to planning?"
      header: "Gate 1"
      options:
        - label: "Approve"
          description: "Research is sufficient, proceed to /plan"
        - label: "Revise"
          description: "Need deeper research on specific areas"
        - label: "Abandon"
          description: "Stop this line of investigation"
      multiSelect: false
```

**Wait for approval before reporting completion.**

### Step 7: Report to User

Tell the user:
1. What you found
2. Where the research doc is saved
3. Gate 1 approval status
4. Next step: `/plan` to create implementation plan

## Key Rules

- **Actually dispatch the Explore agent** - don't just describe doing it
- **Scope searches** - use the topic to narrow file patterns
- **Cite evidence** - every claim needs `file:line`
- **Write output** - research must produce a `.agents/research/` artifact

## Thoroughness Levels

Include in your Explore agent prompt:
- "quick" - for simple questions
- "medium" - for feature exploration
- "very thorough" - for architecture/cross-cutting concerns
