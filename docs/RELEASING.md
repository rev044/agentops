# Releasing AgentOps

This document describes the release process for the `ao` CLI and AgentOps plugin.

## Overview

Releases are triggered by git tags and follow a 3-stage pipeline:

```
git tag v1.0.12
    ↓
┌─────────────────────────────────────────────────┐
│              release.yml workflow               │
├─────────────────────────────────────────────────┤
│  BUILD    → GoReleaser builds 4 binaries        │
│            (darwin/linux × amd64/arm64)         │
│            Uploads as workflow artifacts        │
├─────────────────────────────────────────────────┤
│  VALIDATE → Downloads darwin-arm64 artifact     │
│            Runs scripts/validate-release.sh     │
│            Checks version, size, executability  │
├─────────────────────────────────────────────────┤
│  PUBLISH  → Creates GitHub Release              │
│            Uploads binaries                     │
│            Pushes Homebrew formula              │
└─────────────────────────────────────────────────┘
```

## Making a Release

### 1. Pre-release Checklist

- [ ] All tests pass locally
- [ ] Version number follows semver (vX.Y.Z)
- [ ] CHANGELOG updated (if maintained)
- [ ] No uncommitted changes on main

### 2. Create and Push Tag

```bash
# Ensure you're on main and up to date
git checkout main
git pull

# Create annotated tag
git tag -a v1.0.12 -m "Release v1.0.12"

# Push tag (triggers release workflow)
git push origin v1.0.12
```

### 3. Monitor the Workflow

Watch the release at: https://github.com/boshu2/agentops/actions

The workflow runs three jobs sequentially:
1. **build** - Creates binaries (~2 min)
2. **validate** - Tests darwin-arm64 binary (~1 min)
3. **publish** - Creates release and updates Homebrew (~1 min)

### 4. Verify the Release

After the workflow completes:

```bash
# Update Homebrew
brew update

# Upgrade ao
brew upgrade ao

# Verify version
ao version
# Should show: ao version 1.0.12
```

## Validation Checks

The `validate` stage runs `scripts/validate-release.sh` which checks:

| Check | What it Catches |
|-------|-----------------|
| Binary exists | Build produced no output |
| Size > 1MB | Truncated or corrupted binary |
| File is executable | Wrong file type in archive |
| Version matches tag | ldflags injection failed |
| `--help` works | Binary crashes on startup |
| `-h` works | Flag parsing broken |
| `status` runs | Basic command execution |

## Failure Modes

### Validation Fails

If validation fails, the release is NOT published. The tag exists but no release was created.

**To fix:**
1. Identify the issue from the workflow logs
2. Fix the code
3. Delete the tag: `git tag -d v1.0.12 && git push origin :refs/tags/v1.0.12`
4. Create a new tag after fixing

### Publish Fails

If publish fails after validation passed, the release may be in a partial state.

**To fix:**
1. Check if GitHub release was created (may need manual cleanup)
2. Check if Homebrew formula was pushed
3. Use `workflow_dispatch` to re-run manually if needed

### Homebrew Token Expired

The `HOMEBREW_TAP_GITHUB_TOKEN` secret is validated before publish. If it fails:

1. Generate a new PAT at https://github.com/settings/tokens
2. Scope: `public_repo` (for homebrew-agentops)
3. Update the secret in repository settings

## Manual Release (workflow_dispatch)

If you need to re-run a release without pushing a new tag:

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Enter the tag (e.g., `v1.0.12`)
4. Click "Run workflow"

## Local Testing

Before tagging, you can test the build locally:

```bash
# Install goreleaser
brew install goreleaser

# Build snapshot (doesn't require tag)
goreleaser build --snapshot --clean --single-target

# Test the binary
./dist/ao_darwin_arm64/ao version
./dist/ao_darwin_arm64/ao --help
./dist/ao_darwin_arm64/ao status
```

## Configuration Files

| File | Purpose |
|------|---------|
| `.goreleaser.yml` | Build configuration, binary naming, Homebrew formula |
| `.github/workflows/release.yml` | 3-stage release workflow |
| `scripts/validate-release.sh` | Binary validation script |

## Homebrew Tap

The Homebrew formula is automatically pushed to:
https://github.com/boshu2/homebrew-agentops

Users install with:
```bash
brew tap boshu2/agentops
brew install ao
```

## Troubleshooting

### "ao version" shows "dev"

The ldflags version injection failed. Check `.goreleaser.yml`:
```yaml
ldflags:
  - -s -w -X main.version={{ .Version }}
```

The validation stage should catch this before publish.

### Binary not found in tarball

GoReleaser archive naming doesn't match extraction pattern. Check:
- `.goreleaser.yml` `archives` section
- Workflow's `find` command for tarball

### Homebrew formula not updated

1. Check `HOMEBREW_TAP_GITHUB_TOKEN` is valid
2. Check workflow logs for push errors
3. Verify formula at homebrew-agentops repo
