---
name: golden-init
description: Detects when repos need Golden Template initialization and guides setup
version: 2.0.0
context: fork
triggers:
  - "repo needs setup"
  - "missing CI"
  - "no justfile"
  - "setup this repo"
  - "initialize repo"
  - "golden template"
  - "needs pre-commit"
  - "add CI/CD"
  - "no standards"
allowed-tools: Read, Bash, Grep, Glob, AskUserQuestion
---

# Golden-Init Skill

> Auto-triggers when repos need Golden Template initialization.
> For full init, run `/golden-init --apply`.

## Purpose

This skill **detects and advises** when repos are missing Golden Template components:
- `.agents/` directory structure
- `.beads/` issue tracking
- `justfile` command runner
- `.pre-commit-config.yaml`
- `.gitlab-ci.yml`
- `docs/standards/`
- Tekton integration (for container repos)

## Auto-Detection

When entering a repo, quickly check for missing components:

```bash
# Quick compliance check
missing=()
[[ ! -d ".agents" ]] && missing+=(".agents/")
[[ ! -d ".beads" ]] && missing+=(".beads/")
[[ ! -f "justfile" ]] && missing+=("justfile")
[[ ! -f ".pre-commit-config.yaml" ]] && missing+=("pre-commit")
[[ ! -f ".gitlab-ci.yml" ]] && missing+=("gitlab-ci")
[[ ! -d "docs/standards" ]] && missing+=("standards")

# OSS files (for public/OSS repos)
[[ ! -f "LICENSE" ]] && missing+=("LICENSE")
[[ ! -f "CONTRIBUTING.md" ]] && missing+=("CONTRIBUTING.md")
[[ ! -f "CODE_OF_CONDUCT.md" ]] && missing+=("CODE_OF_CONDUCT.md")
[[ ! -f "SECURITY.md" ]] && missing+=("SECURITY.md")
[[ ! -f "AGENTS.md" ]] && missing+=("AGENTS.md")
[[ ! -f "CHANGELOG.md" ]] && missing+=("CHANGELOG.md")
[[ ! -d ".github/ISSUE_TEMPLATE" ]] && missing+=(".github/ISSUE_TEMPLATE/")
[[ ! -f ".github/PULL_REQUEST_TEMPLATE.md" ]] && missing+=("PR template")
```

## Response Pattern

When components are missing, respond with:

```markdown
## Golden Template Status

This repo is missing Golden Template components:

### Core Infrastructure
| Component | Status |
|-----------|--------|
| .agents/ | {PASS/FAIL} |
| .beads/ | {PASS/FAIL} |
| justfile | {PASS/FAIL} |
| pre-commit | {PASS/FAIL} |
| gitlab-ci | {PASS/FAIL} |
| standards | {PASS/FAIL} |

### OSS Documentation
| Component | Status |
|-----------|--------|
| LICENSE | {PASS/FAIL} |
| CONTRIBUTING.md | {PASS/FAIL} |
| CODE_OF_CONDUCT.md | {PASS/FAIL} |
| SECURITY.md | {PASS/FAIL} |
| AGENTS.md | {PASS/FAIL} |
| CHANGELOG.md | {PASS/FAIL} |
| .github/ISSUE_TEMPLATE/ | {PASS/FAIL} |
| .github/PULL_REQUEST_TEMPLATE.md | {PASS/FAIL} |

**Detected:** {REPO_TYPE} repo with {LANGUAGES}

To initialize or update, run:
```
/golden-init --apply
```

Or for audit only:
```
/golden-init --audit
```
```

## Repo Type Detection

Quick detection without full analysis:

```bash
# Determine repo type
if [[ -f pyproject.toml ]] || [[ -f go.mod ]] || [[ -f package.json ]]; then
    echo "code"
elif [[ -d charts ]] || [[ -f Chart.yaml ]]; then
    echo "config"
else
    echo "unknown"
fi
```

## Standards Corpus Check

When user asks about standards or linting:

```markdown
## Standards Status

**Canonical source:** `gitops/docs/standards/`

| Language | Standards File | In gitops? |
|----------|----------------|------------|
| Python | python-style-guide.md | {YES/NO} |
| Shell | shell-script-standards.md | {YES/NO} |
| Go | go-style-guide.md | {YES/NO} |
| TypeScript | typescript-standards.md | {YES/NO} |

{IF_MISSING}
**Missing standards corpus!**
Create the missing standards in gitops before running golden-init.
{/IF_MISSING}
```

## Integration with /golden-init Command

This skill is lightweight - it detects and advises.
For actual initialization, it directs users to the full command:

- `/golden-init --audit` - Check compliance without changes
- `/golden-init --apply` - Create/update missing components
- `/golden-init --report` - Generate detailed report

## Pattern References

For detailed templates, the skill can reference:

| Pattern | Purpose |
|---------|---------|
| `repo-detection.md` | Full language detection |
| `justfile-templates.md` | Justfile generation |
| `cicd-templates.md` | CI/CD scaffolding |
| `tekton-integration.md` | Container builds |
| `oss-templates.md` | OSS documentation files |
| `file-templates.md` | CLAUDE.md and standards |

Located at: `~/.claude/patterns/commands/golden-init/`

## Usage Examples

### User: "This repo has no CI"

**Response:**
```
I notice this repo is missing CI/CD configuration.

Quick check shows:
- No `.gitlab-ci.yml` found
- No `.pre-commit-config.yaml` found
- Detected: Python repo (pyproject.toml present)

Run `/golden-init --apply` to scaffold:
- GitLab CI with lint, test, security stages
- Pre-commit hooks (ruff, yamllint, conventional commits)
- Security scanning (gitleaks, semgrep)
```

### User: "Setup this repo for development"

**Response:**
```
Let me check what this repo needs for Golden Template compliance...

## Current Status

| Component | Status |
|-----------|--------|
| .agents/ | [FAIL] Missing |
| .beads/ | [FAIL] Not initialized |
| justfile | [FAIL] Missing |
| pre-commit | [FAIL] Missing |
| CLAUDE.md | [PASS] Present |

**Repo Type:** Mixed (Python + Helm charts)
**Languages:** Python, Shell, YAML

To initialize everything:
```
/golden-init --apply
```

This will create:
- .agents/ directory structure
- Beads issue tracking
- Justfile with lint, test, ci recipes
- Pre-commit hooks for Python, YAML, Shell
- GitLab CI pipeline
- Standards synced from gitops
```

### User: "Why isn't linting working?"

**Response:**
```
Let me check the linting configuration...

**Pre-commit:** Not configured (`.pre-commit-config.yaml` missing)
**Standards:** `docs/standards/` directory missing

For Python, required files:
1. `.pre-commit-config.yaml` with ruff hooks
2. `pyproject.toml` with `[tool.ruff]` section
3. Standards from gitops for reference

Run `/golden-init --apply` to scaffold all linting infrastructure.
```

## Command/Skill Context Engineering Audit

When auditing a repo with commands/skills, check context engineering compliance:

```bash
# Check if audit script exists
if [[ -f "scripts/check-command-skill-size.sh" ]]; then
    ./scripts/check-command-skill-size.sh
fi

# Or use just command if available
if command -v just &>/dev/null && [[ -f "justfile" ]]; then
    just audit-commands 2>/dev/null || true
fi
```

### Context Engineering Standards

| Artifact | Max Lines | Purpose |
|----------|-----------|---------|
| Commands | 50 | Thin entry points - invoke skills |
| Skills | 100+ | Comprehensive intelligence |

### Audit Response Pattern

```markdown
## Command/Skill Context Engineering

| Type | Name | Lines | Status |
|------|------|-------|--------|
| Command | plan | 37 | ✓ Compliant |
| Command | complexity | 241 | ✗ Oversized (max: 50) |
| Skill | plan | 333 | ✓ Comprehensive |

**Issues Found:** 1 oversized command

**Fix:** Move intelligence from `complexity` command to a skill:
1. Create `complexity/SKILL.md`
2. Reduce command to thin invoker (~30-50 lines)
3. Reference the skill from the command
```

### Integration with Flywheel

Commands and skills participate in the knowledge flywheel:
- Track invocations via `/knowledge/usage` endpoint
- Promotion based on usage evidence
- Decay for unused artifacts (30/60/90 days)
- **Conversation analysis** extracts learnings from Claude Code sessions
- **Memory storage** enables semantic recall via ai-platform

**Report endpoint:** `POST /knowledge/usage/stats`

**See also:** `~/.claude/docs/KNOWLEDGE-FLYWHEEL.md` for full architecture

## Skill Boundaries

**DO:**
- Quick compliance checks
- Advise on missing components
- Direct to /golden-init command
- Explain what components do
- Audit command/skill sizes

**DON'T:**
- Create files directly (use command for that)
- Modify existing configurations
- Run lengthy operations
- Make changes without user consent

## Relationship to Beads Skill

This skill complements the beads skill:
- **golden-init**: Repo structure and CI/CD
- **beads**: Issue tracking workflow

Both can coexist - golden-init sets up the infrastructure, beads helps use it.
