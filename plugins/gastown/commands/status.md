---
description: Quick status check of current work state
version: 1.0.0
model: haiku
allowed-tools: Bash, Read
---

# Status Command

Quick snapshot of current work state.

## Active Work

```bash
echo "=== Active Issues ==="
bd list --status in_progress 2>/dev/null || echo "No beads configured"

echo ""
echo "=== Ready Issues (Top 5) ==="
bd ready 2>/dev/null | head -5 || echo "No ready issues"
```

## Session Context

```bash
echo ""
echo "=== Recent Files (last 30min) ==="
find . -type f \( -name "*.py" -o -name "*.md" -o -name "*.ts" \) -mmin -30 2>/dev/null | grep -v node_modules | grep -v __pycache__ | head -10 || echo "No recent changes"

echo ""
echo "=== Git Status ==="
git status -s 2>/dev/null | head -10 || echo "Not a git repo"
```

## TodoWrite State

If TodoWrite has items, display current in_progress and pending items.

## Summary

Report format:
- X issues in progress
- Y issues ready
- Z files modified recently
- Git: clean/dirty
