---
name: status
description: >
  Quick status check of current work state. Use when the user asks
  "what's my status", "show status", "where am I", "current state",
  or needs a snapshot of active work, ready issues, and git state.
version: 1.0.0
author: "AI Platform Team"
license: "MIT"
context: fork
allowed-tools: "Bash,Read"
---

# Status Skill

Quick snapshot of current work state.

## Instructions

Run these checks and report a concise summary:

### 1. Active Work

```bash
echo "=== Active Issues ==="
bd list --status in_progress 2>/dev/null || echo "No beads configured"
```

### 2. Ready Issues

```bash
echo "=== Ready Issues (Top 5) ==="
bd ready 2>/dev/null | head -5 || echo "No ready issues"
```

### 3. Recent Files

```bash
echo "=== Recent Files (last 30min) ==="
find . -type f \( -name "*.py" -o -name "*.md" -o -name "*.ts" -o -name "*.go" \) -mmin -30 2>/dev/null | grep -v node_modules | grep -v __pycache__ | grep -v vendor | head -10 || echo "No recent changes"
```

### 4. Git State

```bash
echo "=== Git Status ==="
git status -s 2>/dev/null | head -10 || echo "Not a git repo"
```

### 5. Hook Check (Gas Town)

```bash
echo "=== Hook Status ==="
gt hook 2>/dev/null || echo "Not in Gas Town"
```

## Output Format

Report concisely:
- X issues in progress
- Y issues ready
- Z files modified recently
- Git: clean/dirty
- Hook: work assigned or empty
