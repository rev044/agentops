---
name: oss-docs
description: >
  Scaffold and audit OSS documentation for open source projects.
  Triggers: "add OSS docs", "create README", "setup contributing",
  "add changelog", "missing SECURITY.md", "oss documentation",
  "prepare for open source", "github templates", "add AGENTS.md".
version: 1.0.0
context: fork
author: "Gas Town"
license: "MIT"
allowed-tools: "Read,Write,Edit,Glob,Grep,Bash,Task"
---

# OSS Documentation Skill

> Scaffold and audit documentation for open source projects.
> Based on patterns from [beads](https://github.com/steveyegge/beads) - a well-documented OSS project.

## Overview

This skill helps prepare repositories for open source release by:
1. Auditing existing documentation completeness
2. Scaffolding missing standard files
3. Generating content tailored to project type
4. Following proven patterns from successful OSS projects

## Commands

| Command | Action |
|---------|--------|
| `audit` | Check which OSS docs exist/missing |
| `scaffold` | Create all missing standard files |
| `scaffold [file]` | Create specific file (e.g., `scaffold CONTRIBUTING.md`) |
| `update` | Refresh existing docs with latest patterns |
| `validate` | Check docs follow best practices |

---

## Phase 0: Project Detection

Before scaffolding, detect project characteristics:

```bash
# Determine project type and language
PROJECT_NAME=$(basename $(pwd))
LANGUAGES=()

[[ -f go.mod ]] && LANGUAGES+=("go")
[[ -f pyproject.toml ]] || [[ -f setup.py ]] && LANGUAGES+=("python")
[[ -f package.json ]] && LANGUAGES+=("javascript")
[[ -f Cargo.toml ]] && LANGUAGES+=("rust")
[[ -f Chart.yaml ]] && LANGUAGES+=("helm")

# Detect project category
if [[ -f Dockerfile ]] && [[ -d cmd ]]; then
    PROJECT_TYPE="cli"
elif [[ -f PROJECT ]] || [[ -d config/crd ]]; then
    PROJECT_TYPE="operator"
elif [[ -f Chart.yaml ]]; then
    PROJECT_TYPE="helm"
elif [[ -d api ]] || [[ -d internal/server ]]; then
    PROJECT_TYPE="service"
else
    PROJECT_TYPE="library"
fi
```

---

## Subcommand: audit

Check which OSS documentation files exist:

### Required Files (Tier 1 - Core)

| File | Purpose | Status Check |
|------|---------|--------------|
| `LICENSE` | Legal terms | `[[ -f LICENSE ]]` |
| `README.md` | Project overview | `[[ -f README.md ]]` |
| `CONTRIBUTING.md` | How to contribute | `[[ -f CONTRIBUTING.md ]]` |
| `CODE_OF_CONDUCT.md` | Community standards | `[[ -f CODE_OF_CONDUCT.md ]]` |

### Recommended Files (Tier 2 - Standard)

| File | Purpose | Status Check |
|------|---------|--------------|
| `SECURITY.md` | Vulnerability reporting | `[[ -f SECURITY.md ]]` |
| `CHANGELOG.md` | Version history | `[[ -f CHANGELOG.md ]]` |
| `AGENTS.md` | AI assistant context | `[[ -f AGENTS.md ]]` |
| `.github/ISSUE_TEMPLATE/` | Issue templates | `[[ -d .github/ISSUE_TEMPLATE ]]` |
| `.github/PULL_REQUEST_TEMPLATE.md` | PR template | `[[ -f .github/PULL_REQUEST_TEMPLATE.md ]]` |

### Optional Files (Tier 3 - Enhanced)

| File | Purpose | When Needed |
|------|---------|-------------|
| `docs/QUICKSTART.md` | Getting started | Complex setup |
| `docs/ARCHITECTURE.md` | System design | Non-trivial codebase |
| `docs/CLI_REFERENCE.md` | Command docs | CLI tools |
| `docs/CRD_REFERENCE.md` | CRD spec docs | Kubernetes operators |
| `docs/CONFIG.md` | Configuration options | Configurable software |
| `docs/API.md` | API reference | Libraries/services |
| `docs/TROUBLESHOOTING.md` | Common issues | Production software |
| `examples/` | Usage examples | Complex workflows |

### Audit Output Pattern

```markdown
## OSS Documentation Audit: <PROJECT_NAME>

**Project Type:** <TYPE> | **Languages:** <LANGUAGES>

### Tier 1 (Required)
| File | Status |
|------|--------|
| LICENSE | [PASS/FAIL] |
| README.md | [PASS/FAIL] |
| CONTRIBUTING.md | [PASS/FAIL] |
| CODE_OF_CONDUCT.md | [PASS/FAIL] |

### Tier 2 (Standard)
| File | Status |
|------|--------|
| SECURITY.md | [PASS/FAIL] |
| CHANGELOG.md | [PASS/FAIL] |
| AGENTS.md | [PASS/FAIL] |
| Issue Templates | [PASS/FAIL] |
| PR Template | [PASS/FAIL] |

### Tier 3 (Enhanced)
| File | Status | Recommended? |
|------|--------|--------------|
| docs/QUICKSTART.md | [PASS/FAIL] | <YES/NO> |
| docs/ARCHITECTURE.md | [PASS/FAIL] | <YES/NO> |
| docs/CLI_REFERENCE.md | [PASS/FAIL] | <YES/NO> |

**Score:** X/Y files present

To scaffold missing files:
```
/oss-docs scaffold
```
```

---

## Subcommand: scaffold

Generate missing documentation files.

### Execution Flow

1. Run audit to identify gaps
2. Detect project type and extract metadata
3. Generate files using type-specific templates
4. Report what was created

### Template Selection

Templates are selected based on project type:

| Project Type | Template Variations |
|--------------|---------------------|
| `cli` | Focus on installation, commands, examples |
| `operator` | K8s CRDs, RBAC, deployment |
| `service` | API, configuration, deployment |
| `library` | API reference, examples, installation |
| `helm` | Values, dependencies, upgrading |

### Variable Extraction

```bash
# Extract from git/files
GIT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
REPO_OWNER=$(echo "$GIT_ORIGIN" | sed -E 's#.*[:/]([^/]+)/[^/]+\.git#\1#')
REPO_NAME=$(echo "$GIT_ORIGIN" | sed -E 's#.*/([^/]+)\.git#\1#')

# Extract from go.mod if present
if [[ -f go.mod ]]; then
    MODULE=$(head -1 go.mod | awk '{print $2}')
fi

# Extract from package.json if present
if [[ -f package.json ]]; then
    PKG_NAME=$(jq -r .name package.json)
    PKG_VERSION=$(jq -r .version package.json)
fi
```

---

## Subcommand: validate

Check documentation quality and consistency.

### Validation Rules

1. **Link Validation** - All markdown links resolve
2. **Code Block Accuracy** - Commands in code blocks are correct
3. **Version Consistency** - Versions match across files
4. **Placeholder Check** - No `<PLACEHOLDER>` text remaining
5. **Section Completeness** - Required sections present

### Validation Output

```markdown
## Documentation Validation: <PROJECT_NAME>

### Link Check
- README.md: 15 links, 14 valid, 1 broken (line 42: ./missing.md)

### Placeholder Check
- CONTRIBUTING.md: Found `<TEST_COMMAND>` on line 12

### Version Check
- README.md: v0.5.0
- CHANGELOG.md: v0.5.0
- package.json: 0.5.0
Status: CONSISTENT

### Quality Score: 85/100
```

---

## Project-Type Templates

### CLI Tools (Go)

**README.md structure:**
```markdown
# <name>

One-line description.

## Installation

```bash
brew install <name>
# or
go install <module>@latest
```

## Quick Start

```bash
<name> init
<name> <primary-command>
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize... |
| `<cmd>` | Do something... |

## Configuration

Configuration file: `~/.config/<name>/config.yaml`

## Documentation

- [Quick Start](docs/QUICKSTART.md)
- [CLI Reference](docs/CLI_REFERENCE.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
```

### Kubernetes Operators

**README.md structure:**
```markdown
# <name>

Kubernetes operator for <purpose>.

## Installation

```bash
kubectl apply -f https://github.com/<owner>/<repo>/releases/latest/download/install.yaml
```

## Quick Start

```yaml
apiVersion: <group>/<version>
kind: <Kind>
metadata:
  name: example
spec:
  # ...
```

## CRDs

| CRD | Description |
|-----|-------------|
| `<Kind>` | Manages... |

## Configuration

See [docs/CONFIG.md](docs/CONFIG.md) for all options.
```

### CRD_REFERENCE.md Pattern (Operators)

**For Kubernetes operators, document CRD spec fields:**

```markdown
# CRD Reference

## Polecat

A Polecat represents an ephemeral worker pod.

### Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `issueID` | string | Yes | Beads issue ID to work on |
| `rig` | string | Yes | Target rig name |
| `mode` | string | No | Execution mode (local, kubernetes) |
| `ttl` | duration | No | Time to live (default: 1h) |

### Status

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current phase (Pending, Running, Succeeded, Failed) |
| `startTime` | timestamp | When execution started |
| `completionTime` | timestamp | When execution completed |

### Example

```yaml
apiVersion: gastown.olympus.io/v1alpha1
kind: Polecat
metadata:
  name: polecat-abc123
spec:
  issueID: "gt-1234"
  rig: "athena"
  mode: "local"
```
```

---

## Integration Points

| Skill | Purpose |
|-------|---------|
| `golden-init` | Full repo setup (CI, linting, etc.) |
| `doc` | Code documentation generation |
| `beads` | Issue tracking for doc gaps |

---

## Beads Documentation Patterns

Based on analysis of beads (a well-documented OSS project):

### README Structure Pattern

```
1. Title + One-liner
2. Badges (optional)
3. Quick Install (1-2 commands)
4. Quick Start (3-5 commands)
5. Key Features (bullet list)
6. Documentation Links
7. Contributing/License
```

### Documentation Organization

```
project/
├── README.md              # Overview + quick start
├── AGENTS.md              # AI assistant context
├── CONTRIBUTING.md        # Contributor guide
├── CHANGELOG.md           # Keep a Changelog format
├── docs/
│   ├── QUICKSTART.md      # Detailed getting started
│   ├── CLI_REFERENCE.md   # Complete command reference
│   ├── ARCHITECTURE.md    # System design
│   ├── CONFIG.md          # Configuration options
│   ├── TROUBLESHOOTING.md # Common issues
│   └── <FEATURE>.md       # Feature-specific docs
└── examples/
    └── README.md          # Examples index
```

### AGENTS.md Pattern (from beads)

```markdown
# Agent Instructions

This project uses **<tool>** for <purpose>. Run `<onboard-cmd>` to get started.

## Quick Reference

```bash
<cmd1>              # Do thing 1
<cmd2>              # Do thing 2
<cmd3>              # Do thing 3
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below.

**MANDATORY WORKFLOW:**

1. **Run quality gates** - Tests, linters, builds
2. **Commit changes** - Meaningful commit message
3. **PUSH TO REMOTE** - This is MANDATORY
4. **Verify** - All changes committed AND pushed
```

---

## References

- **Documentation Tiers**: `references/documentation-tiers.md`
- **Project Types**: `references/project-types.md`
- **Beads Patterns**: `references/beads-patterns.md`
- **Audit Script**: `scripts/audit-oss-docs.sh`

---

## Style Guidelines

From successful OSS projects (gastown, beads):

1. **Be direct** - Get to the point quickly
2. **Be friendly** - Welcome contributions
3. **Be concise** - Avoid boilerplate
4. **Be specific** - Include project-specific details
5. **Use tables** - For commands, options, features
6. **Show examples** - Code blocks over prose
7. **Link liberally** - Cross-reference related docs

---

## Skill Boundaries

**DO:**
- Audit existing documentation
- Generate standard OSS files
- Validate documentation quality
- Suggest improvements

**DON'T:**
- Overwrite existing content without confirmation
- Generate code documentation (use doc for that)
- Create CI/CD files (use golden-init for that)
- Make changes without user consent
