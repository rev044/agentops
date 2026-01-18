# Troubleshooting Polecats

Common issues and fixes for stuck, broken, or misbehaving polecats.

## Quick Diagnostics

```bash
# Status check
gt polecat status <rig>/<name>

# Git state (what would be lost)
gt polecat git-state <rig>/<name>

# Recovery assessment
gt polecat check-recovery <rig>/<name>

# Peek at tmux session
tmux capture-pane -t gt-<rig>-<name> -p | tail -30
```

## Common Issues

### Polecat Hit Usage Limit

**Symptoms:**
- Session shows "You've hit your limit"
- No progress on work
- Polecat appears stuck

**Diagnosis:**
```bash
tmux capture-pane -t gt-daedalus-Toast -p | grep -i "limit"
```

**Fix:**
```bash
# Option A: Wait for limit reset (usually 5 hours)

# Option B: Nuke and re-dispatch after /login
gt polecat nuke daedalus/Toast --force
# ... login to reset limits ...
gt sling <bead> daedalus
```

**Prevention:**
- Use `--account work` to specify backup account
- Stagger batch dispatches
- Monitor convoy progress

---

### Polecat Seems Stuck

**Symptoms:**
- No activity in tmux
- Bead still open
- Convoy shows no progress

**Diagnosis:**
```bash
# Check what it's doing
tmux capture-pane -t gt-daedalus-Toast -p | tail -40

# Check status
gt polecat status daedalus/Toast

# Check if it has work
bd show <bead-id>
```

**Possible Causes:**

| Cause | Signs | Fix |
|-------|-------|-----|
| Waiting for input | "?" prompt | Send nudge |
| Error loop | Same error repeating | Nuke and restart |
| Resource wait | "waiting for..." | Check dependency |
| Session crashed | No tmux output | Restart session |
| Rate limited | Limit message | Wait or switch account |

**Fixes:**

```bash
# Nudge - send input to continue
tmux send-keys -t gt-daedalus-Toast "continue with your task" Enter

# Restart Claude in session
tmux send-keys -t gt-daedalus-Toast "claude" Enter

# Nuclear option
gt polecat nuke daedalus/Toast --force
gt sling <bead> daedalus
```

---

### Session Crashed / Won't Start

**Symptoms:**
- No tmux session found
- `gt polecat status` shows "no session"

**Diagnosis:**
```bash
# List sessions
tmux ls | grep gt-daedalus

# Check worktree exists
ls ~/gt/daedalus/polecats/Toast
```

**Fixes:**

```bash
# Restart session manually
cd ~/gt/daedalus/polecats/Toast
tmux new-session -d -s gt-daedalus-Toast 'claude'

# Or nuke and re-sling
gt polecat nuke daedalus/Toast --force
gt sling <bead> daedalus
```

---

### Work Not on Hook

**Symptoms:**
- Polecat is idle
- `gt hook` shows nothing
- Work was slung but not picked up

**Diagnosis:**
```bash
# Check hook status
cd ~/gt/daedalus/polecats/Toast
gt hook

# Check bead assignment
bd show <bead-id>
```

**Fixes:**

```bash
# Re-hook the work
cd ~/gt/daedalus/polecats/Toast
gt hook <bead>

# Or nudge the session
tmux send-keys -t gt-daedalus-Toast "gt hook && gt prime" Enter
```

---

### Branch Conflicts

**Symptoms:**
- Push rejected
- Merge conflicts on beads

**Diagnosis:**
```bash
cd ~/gt/daedalus/polecats/Toast
git status
git log --oneline -5
git diff origin/main...HEAD
```

**Fixes:**

```bash
# For beads conflicts (append-only, safe to accept theirs)
git checkout --theirs .beads/issues.jsonl
git add .beads/issues.jsonl
git commit -m "merge: resolve beads conflict"

# For code conflicts - manual resolution needed
git merge origin/main
# ... resolve conflicts ...
git commit

# Nuclear option - lose local, restart fresh
gt polecat nuke daedalus/Toast --force
gt sling <bead> daedalus
```

---

### Worktree Corrupted

**Symptoms:**
- Git errors in worktree
- Missing files
- Inconsistent state

**Diagnosis:**
```bash
cd ~/gt/daedalus/polecats/Toast
git status
git fsck
```

**Fixes:**

```bash
# If work is pushed, safe to nuke
gt polecat check-recovery daedalus/Toast
gt polecat nuke daedalus/Toast --force

# If work is NOT pushed, try to recover
cd ~/gt/daedalus/polecats/Toast
git stash
git checkout origin/main
git stash pop
# ... verify state ...
git push -u origin HEAD
# THEN nuke
gt polecat nuke daedalus/Toast
```

---

## Recovery vs Nuke Decision

Use `gt polecat check-recovery` to decide:

```bash
gt polecat check-recovery daedalus/Toast
```

| Result | Meaning | Action |
|--------|---------|--------|
| "Safe to nuke" | No valuable work | `gt polecat nuke --force` |
| "Recovery recommended" | Unpushed work exists | Push first, then nuke |
| "Manual review needed" | Complex state | Inspect git-state |

## Escalation Path

1. **Try nudge** - Simple input might unstick
2. **Check status** - Understand the state
3. **Check git-state** - Know what would be lost
4. **check-recovery** - Automated recommendation
5. **Nuke with --force** - Last resort

## Monitoring Best Practices

### Convoy Dashboard

```bash
gt convoy list
```

Shows all active work. Stalled entries need attention.

### Periodic Checks

```bash
# All polecats in rig
gt polecat list daedalus

# Stale detection
gt polecat stale daedalus

# Peek at each
for p in $(gt polecat list daedalus --names-only); do
  echo "=== $p ==="
  tmux capture-pane -t gt-daedalus-$p -p | tail -10
done
```

### Alert Signs

- Convoy showing same status for >30min
- Multiple polecats stuck simultaneously
- Usage limit messages appearing
- Beads not closing despite activity

## Prevention

1. **Stagger dispatches** - Don't sling 10 beads at once
2. **Use multiple accounts** - Spread usage limits
3. **Monitor convoys** - Catch issues early
4. **Run gc regularly** - Keep clean state
5. **Push frequently** - Reduce recovery complexity
