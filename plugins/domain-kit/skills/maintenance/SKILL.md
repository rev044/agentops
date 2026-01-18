---
name: maintenance
description: >
  Repository maintenance utilities. Use for "audit docs", "cleanup repo",
  "check documentation structure", "remove deprecated", "audit workflow",
  "clean old plans", "check onboarding docs".
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
---

# Maintenance Skill

Consolidated repository maintenance utilities for auditing and cleanup.

## Commands

| Command | Purpose |
|---------|---------|
| `audit diataxis` | Check documentation follows Diataxis structure |
| `audit onboarding` | Verify onboarding docs exist and are complete |
| `audit workflow` | Validate workflow documentation |
| `cleanup deprecated` | Remove deprecated files and references |
| `cleanup docs` | Clean stale documentation |
| `cleanup plans` | Archive old plan files |
| `cleanup repo` | Full repository cleanup |

## Quick Reference

```bash
# Audit commands
/maintenance audit diataxis
/maintenance audit onboarding
/maintenance audit workflow

# Cleanup commands
/maintenance cleanup deprecated
/maintenance cleanup docs
/maintenance cleanup plans
/maintenance cleanup repo
```

## Audit: Diataxis

Check that documentation follows the Diataxis framework (tutorials, how-to, reference, explanation).

```bash
#!/bin/bash
# Check for Diataxis directories
for dir in tutorials how-to reference explanation; do
  if [ -d "docs/$dir" ]; then
    echo "✅ docs/$dir exists"
  else
    echo "❌ docs/$dir missing"
  fi
done

# Check for mixed content
grep -r "Step 1:" docs/reference/ && echo "⚠️ Tutorial content in reference/"
grep -r "API Reference" docs/tutorials/ && echo "⚠️ Reference content in tutorial/"
```

## Audit: Onboarding

Verify onboarding documentation is complete.

```bash
#!/bin/bash
REQUIRED_FILES=(
  "README.md"
  "docs/getting-started.md"
  "CONTRIBUTING.md"
)

for file in "${REQUIRED_FILES[@]}"; do
  [ -f "$file" ] && echo "✅ $file" || echo "❌ $file missing"
done
```

## Audit: Workflow

Validate workflow documentation matches actual workflows.

```bash
#!/bin/bash
# Check for documented workflows
ls .github/workflows/*.yml 2>/dev/null | while read wf; do
  name=$(basename "$wf" .yml)
  grep -q "$name" docs/ && echo "✅ $name documented" || echo "❌ $name undocumented"
done
```

## Cleanup: Deprecated

Remove files marked as deprecated.

```bash
#!/bin/bash
# Find deprecated markers
grep -rl "DEPRECATED" . --include="*.md" | while read file; do
  echo "Found deprecated: $file"
done

# Find old backup files
find . -name "*.bak" -o -name "*.old" -o -name "*~" | head -20
```

## Cleanup: Docs

Clean stale documentation.

```bash
#!/bin/bash
# Find docs older than 6 months with no recent updates
find docs/ -name "*.md" -mtime +180 | while read doc; do
  commits=$(git log --oneline --since="6 months ago" -- "$doc" | wc -l)
  [ "$commits" -eq 0 ] && echo "Stale: $doc"
done
```

## Cleanup: Plans

Archive old plan files from .agents/.

```bash
#!/bin/bash
ARCHIVE_DIR=".agents/.archive/$(date +%Y-%m)"
mkdir -p "$ARCHIVE_DIR"

# Move plans older than 30 days
find .agents/*/plans/ -name "*.md" -mtime +30 -exec mv {} "$ARCHIVE_DIR/" \;
echo "Archived old plans to $ARCHIVE_DIR"
```

## Cleanup: Repo

Full repository cleanup.

```bash
#!/bin/bash
echo "=== Repository Cleanup ==="

# Git cleanup
git gc --aggressive --prune=now

# Remove build artifacts
rm -rf node_modules/.cache dist/ build/ __pycache__/

# Clean test artifacts
rm -rf .pytest_cache/ .coverage coverage/

# Report
echo "Cleanup complete"
du -sh .
```

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Delete without review | List files first, confirm |
| Skip git commit | Commit cleanup as atomic change |
| Clean during active work | Clean at session boundaries |
| Ignore .gitignore | Respect ignored files |
