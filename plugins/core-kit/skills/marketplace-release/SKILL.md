---
name: marketplace-release
description: >
  Claude Code marketplace plugin release workflow. Handles version bumping,
  changelog updates, git tagging, and publishing. Ensures updates propagate
  to users who install from the marketplace.
version: 1.0.0
author: "AgentOps Team"
license: "MIT"
context: inline
triggers:
  - "release plugin"
  - "publish marketplace"
  - "bump version"
  - "create release"
  - "marketplace release"
allowed-tools: "Read,Edit,Write,Bash,Glob,Grep"
---

# Marketplace Release Skill

Release Claude Code plugins to the marketplace with proper versioning.

## Quick Start

```bash
/marketplace-release              # Interactive release flow
/marketplace-release patch        # Bump patch version (0.1.0 -> 0.1.1)
/marketplace-release minor        # Bump minor version (0.1.0 -> 0.2.0)
/marketplace-release major        # Bump major version (0.1.0 -> 1.0.0)
```

---

## How Claude Code Plugin Updates Work

### Distribution Model

Claude Code uses a **decentralized marketplace model**:

1. **No central registry** - Marketplaces are just Git repos with `marketplace.json`
2. **Users add marketplaces** - `/plugin marketplace add owner/repo`
3. **Users install plugins** - `/plugin install plugin-name@marketplace-name`
4. **Auto-update at startup** - If enabled, pulls latest versions

### Update Propagation

| Trigger | When Updates Reach Users |
|---------|--------------------------|
| **Auto-update enabled** | Next Claude Code startup |
| **Auto-update disabled** | Manual: `/plugin update plugin-name@marketplace` |
| **Version pinned** | Never (until user unpins) |

---

## Release Checklist

### 1. Identify Changed Plugins

```bash
# See what plugin files changed
git diff --name-only HEAD~5 | grep "plugins/" | cut -d'/' -f2 | sort -u
```

### 2. Update Plugin Versions

Each changed plugin needs its version bumped in `.claude-plugin/plugin.json`:

```bash
# Check current versions
grep -r '"version"' plugins/*/.claude-plugin/plugin.json

# Edit each changed plugin
# plugins/<name>/.claude-plugin/plugin.json
{
  "name": "my-plugin",
  "version": "0.1.1",  # <- Bump this
  ...
}
```

**Semantic Versioning:**
- `MAJOR.MINOR.PATCH`
- **PATCH** (0.0.X): Bug fixes, no new features
- **MINOR** (0.X.0): New features, backward compatible
- **MAJOR** (X.0.0): Breaking changes

### 3. Update Root Marketplace Version

```bash
# .claude-plugin/plugin.json (root)
{
  "name": "agentops",
  "version": "0.2.1",  # <- Bump this
  ...
}
```

### 4. Update CHANGELOG.md

```markdown
## [Unreleased]

## [X.Y.Z] - YYYY-MM-DD

### Added
- New feature description

### Changed
- Changed behavior description

### Fixed
- Bug fix description
- **plugin-name** (vX.Y.Z) - What was fixed
```

### 5. Run Validation

```bash
./tests/marketplace-e2e-test.sh
```

Must pass with 0 errors (warnings OK).

### 6. Commit and Tag

```bash
git add -A
git commit -m "release: vX.Y.Z - Brief description"
git tag -a vX.Y.Z -m "Release vX.Y.Z - Description"
```

### 7. Push

```bash
git push && git push origin vX.Y.Z
```

---

## File Reference

### Required Files

| File | Purpose | When to Update |
|------|---------|----------------|
| `plugins/<kit>/.claude-plugin/plugin.json` | Plugin version | When plugin changes |
| `.claude-plugin/plugin.json` | Root marketplace version | Every release |
| `CHANGELOG.md` | Change documentation | Every release |

### Optional Files

| File | Purpose |
|------|---------|
| `.claude-plugin/marketplace.json` | Plugin catalog (auto-discovered) |
| `plugins/<kit>/CLAUDE.md` | Plugin-specific instructions |

---

## Common Patterns

### Fix Applied to Local Skills but Not Marketplace

**Problem:** You fixed `~/.claude/skills/foo/SKILL.md` but users installing from
the marketplace don't get the fix.

**Solution:** Also fix `plugins/<kit>/skills/foo/SKILL.md` and bump plugin version.

```bash
# 1. Find the marketplace skill
find plugins -name "SKILL.md" -path "*foo*"

# 2. Apply the same fix
# 3. Bump the plugin version
# 4. Release
```

### Skill Exists Locally but Not in Marketplace

Some skills in `~/.claude/skills/` may not be distributed:
- Personal/experimental skills
- Skills with external dependencies
- Work-in-progress

**To add to marketplace:**
```bash
cp -r ~/.claude/skills/my-skill plugins/<appropriate-kit>/skills/
# Edit to remove personal paths/config
# Bump plugin version
# Release
```

### Multiple Plugins Changed

Bump each changed plugin individually:

```bash
# core-kit changed
plugins/core-kit/.claude-plugin/plugin.json  # 0.1.0 -> 0.1.1

# vibe-kit changed
plugins/vibe-kit/.claude-plugin/plugin.json  # 0.1.1 -> 0.1.2

# Root version
.claude-plugin/plugin.json                    # 0.2.0 -> 0.2.1
```

---

## Troubleshooting

### Users Not Getting Updates

1. **Check auto-update setting:**
   ```bash
   /plugin marketplace list  # Shows auto-update status
   ```

2. **Force update:**
   ```bash
   /plugin update plugin-name@marketplace-name
   ```

3. **Check version pinning:**
   Users may have pinned to specific version in their settings.

### Validation Fails

```bash
# Run with verbose output
./tests/marketplace-e2e-test.sh 2>&1 | less

# Common issues:
# - Invalid JSON in plugin.json
# - Missing required fields
# - Broken skill references
```

### Tag Already Exists

```bash
# Delete local tag
git tag -d vX.Y.Z

# Delete remote tag (if pushed)
git push origin :refs/tags/vX.Y.Z

# Recreate
git tag -a vX.Y.Z -m "message"
git push origin vX.Y.Z
```

---

## Anti-Patterns

| DON'T | DO INSTEAD |
|-------|------------|
| Only update root version | Update each changed plugin version |
| Skip CHANGELOG | Document all changes |
| Push without validation | Run e2e tests first |
| Fix local skills only | Also fix marketplace plugins |
| Use same version twice | Always increment version |

---

## Context Mode Reference

When adding/modifying skills, understand `context` frontmatter:

| Setting | Behavior | Use When |
|---------|----------|----------|
| `context: inline` | Skill sees conversation history | Skill needs chat context (vibe, crank) |
| `context: fork` | Skill runs in isolation | Long autonomous tasks, no context needed |

**Default:** If omitted, defaults to `inline`.

**Common mistake:** Using `context: fork` for skills that should infer targets
from conversation (e.g., "/vibe the auth code we discussed").

---

## Example Release Session

```bash
# 1. Check what changed
git diff --name-only HEAD~3

# 2. Identify affected plugins
# Output: plugins/vibe-kit/skills/vibe/SKILL.md changed

# 3. Bump vibe-kit version
vim plugins/vibe-kit/.claude-plugin/plugin.json
# Change: "version": "0.1.1" -> "0.1.2"

# 4. Bump root version
vim .claude-plugin/plugin.json
# Change: "version": "0.2.0" -> "0.2.1"

# 5. Update CHANGELOG
vim CHANGELOG.md
# Add [0.2.1] section with changes

# 6. Validate
./tests/marketplace-e2e-test.sh
# Ensure: 0 errors

# 7. Commit
git add -A
git commit -m "release: v0.2.1 - Fix vibe skill context mode"

# 8. Tag
git tag -a v0.2.1 -m "Release v0.2.1"

# 9. Push
git push && git push origin v0.2.1

# Done! Users will get update on next Claude Code startup
```

---

## Related Skills

- `/golden-init` - Initialize new plugin structure
- `/vibe` - Validate plugin code quality
- `/doc` - Generate plugin documentation

---

## References

- [Claude Code Plugin Docs](https://docs.anthropic.com/claude-code/plugins)
- [Plugin Marketplace Guide](https://docs.anthropic.com/claude-code/plugin-marketplaces)
- [Semantic Versioning](https://semver.org/)
