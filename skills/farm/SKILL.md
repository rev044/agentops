---
name: farm
description: 'Spawn Agent Farm for parallel issue execution. Mayor orchestrates demigods (independent Claude sessions) via tmux. Triggers: "farm", "spawn agents", "parallel work", "multi-agent".'
---

# Farm Skill

**YOU MUST EXECUTE THIS WORKFLOW. Do not just describe it.**

Spawn an Agent Farm - independent Claude Code sessions (demigods) in tmux that work on issues in parallel.

## Core Concept: Demigods

**Demigod** = Independent Claude Code session in its own tmux window
- Fully isolated from Mayor
- Survives Mayor disconnect
- Runs `/implement` loop on assigned work
- Signals completion via file or Agent Mail

**This is NOT subagents.** Subagents share context and die with the session. Demigods are independent processes.

## Architecture

```
Mayor (this session)
    |
    +-> Identify wave (ready issues with no blockers)
    |
    +-> For each issue in wave:
    |       |
    |       +-> tmux new-session -d -s demigod-N
    |       +-> claude --prompt "Run /implement on <issue-id>"
    |       +-> (30s stagger between spawns)
    |
    +-> Monitor: check tmux sessions, .demigod-status files
    |
    +-> When all demigods complete:
    |       +-> Review changes
    |       +-> Run /post-mortem
```

## Work Source

Farm uses the first available:
1. **Beads** (.beads/issues.jsonl) - preferred, git-native
2. **Native Tasks** (TaskList) - fallback, converted to temp beads

If using native tasks, convert them to a temp beads file first so demigods can use `bd` commands.

## Execution Steps

Given `/farm [--agents N]`:

### Step 1: Pre-Flight Validation

```bash
# Check beads exist
if [[ -f .beads/issues.jsonl ]]; then
    echo "Work source: beads"
    WORK_SOURCE="beads"
else
    echo "Work source: native tasks (will convert)"
    WORK_SOURCE="tasks"
fi

# Check ready issues
bd ready 2>/dev/null | head -20

# Check tmux available
command -v tmux >/dev/null || { echo "ERROR: tmux required"; exit 1; }

# Check claude available
command -v claude >/dev/null || { echo "ERROR: claude CLI required"; exit 1; }

# Check disk space
df -h . | awk 'NR==2 {print "Disk available: " $4}'
```

### Step 2: Identify Wave

A **wave** = set of ready issues that can run in parallel (no dependencies between them).

```bash
# Get ready issues (no blockers)
WAVE=$(bd ready --ids-only 2>/dev/null | head -${N:-5})
WAVE_SIZE=$(echo "$WAVE" | wc -l | tr -d ' ')

echo "Wave: $WAVE_SIZE issues ready for parallel execution"
echo "$WAVE"
```

If WAVE_SIZE = 0, STOP: "No ready issues. Check dependencies or run /plan."

### Step 3: Spawn Demigods

**For each issue in the wave, spawn a demigod:**

```bash
PROJECT=$(basename $(pwd))
WAVE_ID=$(date +%Y%m%d-%H%M%S)

for ISSUE_ID in $WAVE; do
    SESSION_NAME="demigod-${PROJECT}-${ISSUE_ID}"

    # Create status file
    echo "spawning" > ".demigod-${ISSUE_ID}.status"

    # Spawn demigod in new tmux session
    tmux new-session -d -s "$SESSION_NAME" \
        "claude --prompt 'You are a demigod. Run /implement on issue ${ISSUE_ID}. When done, write COMPLETE to .demigod-${ISSUE_ID}.status and exit.' 2>&1 | tee .demigod-${ISSUE_ID}.log; echo \$? > .demigod-${ISSUE_ID}.exit"

    echo "Spawned: $SESSION_NAME"

    # Stagger spawns to avoid rate limits
    sleep 30
done

echo ""
echo "Farm running: $WAVE_SIZE demigods spawned"
echo "Wave ID: $WAVE_ID"
```

### Step 4: Monitor Demigods

**Check demigod status periodically:**

```bash
# List active demigod sessions
tmux list-sessions 2>/dev/null | grep "demigod-"

# Check status files
for f in .demigod-*.status; do
    ISSUE=$(echo $f | sed 's/.*demigod-\(.*\)\.status/\1/')
    STATUS=$(cat $f 2>/dev/null || echo "unknown")
    echo "$ISSUE: $STATUS"
done

# Check for completions
grep -l "COMPLETE" .demigod-*.status 2>/dev/null | wc -l
```

**Show live output from a demigod:**
```bash
# Attach to view (Ctrl-B D to detach)
tmux attach -t demigod-<project>-<issue-id>

# Or tail the log
tail -f .demigod-<issue-id>.log
```

### Step 5: Handle Completion

**When all demigods complete:**

```bash
# Check all status files show COMPLETE
TOTAL=$(ls .demigod-*.status 2>/dev/null | wc -l)
COMPLETE=$(grep -l "COMPLETE" .demigod-*.status 2>/dev/null | wc -l)

if [[ $COMPLETE -eq $TOTAL ]]; then
    echo "Farm complete: $COMPLETE/$TOTAL demigods finished"

    # Clean up tmux sessions
    for s in $(tmux list-sessions -F '#{session_name}' 2>/dev/null | grep "demigod-"); do
        tmux kill-session -t "$s" 2>/dev/null
    done

    # Clean up status files
    rm -f .demigod-*.status .demigod-*.log .demigod-*.exit

    echo "Run /post-mortem to extract learnings"
else
    echo "In progress: $COMPLETE/$TOTAL complete"
fi
```

### Step 6: Review Changes

**After farm completes, Mayor reviews:**

```bash
# See what changed
git status
git diff --stat

# Review each demigod's work
git log --oneline -10
```

**Then run /post-mortem to extract learnings.**

## Demigod Behavior

Each demigod (Claude session):
1. Receives prompt: "Run /implement on issue <id>"
2. Runs `/implement` skill which:
   - Claims issue via `bd update --status in_progress`
   - Does the work
   - Commits changes
   - Closes issue via `bd close`
3. Writes "COMPLETE" to status file
4. Exits

Demigods are **fire-and-forget** - they run independently until done.

## Error Handling

### Check for Failed Demigods

```bash
# Check exit codes
for f in .demigod-*.exit; do
    CODE=$(cat $f 2>/dev/null)
    if [[ "$CODE" != "0" ]]; then
        ISSUE=$(echo $f | sed 's/.*demigod-\(.*\)\.exit/\1/')
        echo "FAILED: $ISSUE (exit $CODE)"
        echo "Check log: .demigod-${ISSUE}.log"
    fi
done
```

### Kill Stuck Demigods

```bash
# Kill specific demigod
tmux kill-session -t demigod-<project>-<issue-id>

# Kill all demigods
for s in $(tmux list-sessions -F '#{session_name}' | grep "demigod-"); do
    tmux kill-session -t "$s"
done
```

### Resume After Disconnect

If Mayor disconnects, demigods keep running. On reconnect:

```bash
# Check what's still running
tmux list-sessions | grep "demigod-"

# Check status files
cat .demigod-*.status
```

## Key Rules

- **Demigods are independent** - Not subagents, real Claude sessions
- **30s stagger** - Prevents API rate limits
- **Status files** - Primary completion signal
- **Fire and forget** - Demigods run to completion
- **Mayor reviews** - Always review changes after farm
- **Post-mortem** - Extract learnings from the work

## Wave Sizing

- **Default wave:** min(5, ready_issues)
- **Max recommended:** 10 (API rate limits)
- **Dependencies block:** Only ready (unblocked) issues in wave

## Quick Reference

```bash
# Start farm with 3 demigods
/farm --agents 3

# Check farm status
tmux list-sessions | grep demigod
cat .demigod-*.status

# View demigod output
tail -f .demigod-<issue>.log

# Kill all demigods
tmux kill-session -t demigod-*

# After completion
git status
/post-mortem
```
