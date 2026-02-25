---
name: oss-docs
description: 'Scaffold and audit OSS documentation packs for open source projects. Triggers: "add OSS docs", "setup contributing guide", "add changelog", "prepare for open source", "add AGENTS.md", "OSS documentation".'
---


# OSS Documentation Skill

Scaffold and audit documentation for open source projects.

## Overview

This skill helps prepare repositories for open source release by:
1. Auditing existing documentation completeness
2. Scaffolding missing standard files
3. Generating content tailored to project type

## Commands

| Command | Action |
|---------|--------|
| `audit` | Check which OSS docs exist/missing |
| `scaffold` | Create all missing standard files |
| `scaffold [file]` | Create specific file |
| `update` | Refresh existing docs with latest patterns |
| `validate` | Check docs follow best practices |

---

## Phase 0: Project Detection

```bash
# Determine project type and language
PROJECT_NAME=$(basename $(pwd))
LANGUAGES=()

[[ -f go.mod ]] && LANGUAGES+=("go")
[[ -f pyproject.toml ]] || [[ -f setup.py ]] && LANGUAGES+=("python")
[[ -f package.json ]] && LANGUAGES+=("javascript")
[[ -f Cargo.toml ]] && LANGUAGES+=("rust")

# Detect project category
if [[ -f Dockerfile ]] && [[ -d cmd ]]; then
    PROJECT_TYPE="cli"
elif [[ -d config/crd ]]; then
    PROJECT_TYPE="operator"
elif [[ -f Chart.yaml ]]; then
    PROJECT_TYPE="helm"
else
    PROJECT_TYPE="library"
fi
```

---

## Subcommand: audit

### Required Files (Tier 1 - Core)

| File | Purpose |
|------|---------|
| `LICENSE` | Legal terms |
| `README.md` | Project overview |
| `CONTRIBUTING.md` | How to contribute |
| `CODE_OF_CONDUCT.md` | Community standards |

### Recommended Files (Tier 2 - Standard)

| File | Purpose |
|------|---------|
| `SECURITY.md` | Vulnerability reporting |
| `CHANGELOG.md` | Version history |
| `AGENTS.md` | AI assistant context |
| `.github/ISSUE_TEMPLATE/` | Issue templates |
| `.github/PULL_REQUEST_TEMPLATE.md` | PR template |

### Optional Files (Tier 3 - Enhanced)

| File | When Needed |
|------|-------------|
| `docs/QUICKSTART.md` | Complex setup |
| `docs/ARCHITECTURE.md` | Non-trivial codebase |
| `docs/CLI_REFERENCE.md` | CLI tools |
| `docs/CONFIG.md` | Configurable software |
| `examples/` | Complex workflows |

---

## Subcommand: scaffold

### Template Selection

| Project Type | Focus |
|--------------|-------|
| `cli` | Installation, commands, examples |
| `operator` | K8s CRDs, RBAC, deployment |
| `service` | API, configuration, deployment |
| `library` | API reference, examples |
| `helm` | Values, dependencies, upgrading |

---

## Documentation Organization

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
│   └── CONFIG.md          # Configuration options
└── examples/
    └── README.md          # Examples index
```

---

## AGENTS.md Pattern

```markdown
# Agent Instructions

This project uses **<tool>** for <purpose>. Run `<onboard-cmd>` to get started.

## Quick Reference

```bash
<cmd1>              # Do thing 1
<cmd2>              # Do thing 2
```

## Landing the Plane (Session Completion)

**MANDATORY WORKFLOW:**

1. **Run quality gates** - Tests, linters, builds
2. **Commit changes** - Meaningful commit message
3. **PUSH TO REMOTE** - This is MANDATORY
4. **Verify** - All changes committed AND pushed
```

---

## Style Guidelines

1. **Be direct** - Get to the point quickly
2. **Be friendly** - Welcome contributions
3. **Be concise** - Avoid boilerplate
4. **Use tables** - For commands, options, features
5. **Show examples** - Code blocks over prose
6. **Link liberally** - Cross-reference related docs

---

## Skill Boundaries

**DO:**
- Audit existing documentation
- Generate standard OSS files
- Validate documentation quality

**DON'T:**
- Overwrite existing content without confirmation
- Generate code documentation (use `$doc`)
- Create CI/CD files (out of scope — configure CI/CD separately)

## Examples

### OSS Readiness Audit

**User says:** "Audit this repo for open-source documentation readiness."

**What happens:**
1. Evaluate presence/quality of core OSS docs.
2. Identify missing or weak sections.
3. Output prioritized documentation actions.

### Scaffold Missing Docs

**User says:** "Generate missing OSS docs for this project."

**What happens:**
1. Detect project type and documentation gaps.
2. Scaffold standard files with project-aware content.
3. Produce a checklist for final review and landing.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Generated docs feel generic | Project signals too sparse | Add concrete repo context (commands, architecture, workflows) |
| Existing docs conflict | Legacy text diverges from current behavior | Reconcile with current code/process and mark obsolete sections |
| Contributor path unclear | Missing setup/testing guidance | Add explicit quickstart and validation commands |
| Open-source handoff incomplete | Session-end workflow not reflected | Add landing-the-plane and release hygiene steps |

## Reference Documents

- [references/beads-patterns.md](references/beads-patterns.md)
- [references/documentation-tiers.md](references/documentation-tiers.md)
- [references/project-types.md](references/project-types.md)

---

## References

### beads-patterns.md

# Documentation Patterns from Beads

> Extracted from analysis of beads (chronicle) repository.
> Beads demonstrates exemplary OSS documentation practices.

## Overview

Beads is a Git-backed issue tracker for AI-supervised coding workflows.
As of v0.48.0, it has ~90 markdown files with comprehensive documentation.

**Why study beads?**
- Actively maintained OSS project
- Targets AI-assisted development (similar audience)
- Extensive documentation coverage
- Clear writing style

---

## README.md Pattern

### Structure

```
1. Project name + one-liner
2. Quick install (single command)
3. Quick start (3-5 commands)
4. Key features (bullet list)
5. Documentation links
6. Community/contributing
7. License
```

### Key Elements

**Title + Tagline:**
```markdown
# beads (bd)

> Git-backed issue tracker for AI-supervised coding workflows.
```

**Quick Install:**
```markdown
## Installation

```bash
brew tap steveyegge/beads && brew install bd
```
```

**Quick Start:**
```markdown
## Quick Start

```bash
bd init                  # Initialize in project
bd create "Fix bug" -p 1 # Create issue
bd ready                 # Find unblocked work
bd sync                  # Sync to git
```
```

**Features as Bullets (not walls of text):**
```markdown
## Features

- **Zero setup** - `bd init` creates project-local database
- **Dependency tracking** - Four dependency types
- **Ready work detection** - Find issues with no blockers
- **Agent-friendly** - `--json` flags for programmatic use
```

---

## AGENTS.md Pattern

### Structure

```
1. Quick reference commands
2. Session close protocol
3. Workflow overview
4. Common operations
```

### Key Elements

**Command Quick Reference:**
```markdown
## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```
```

**Session Close Protocol (Critical):**
```markdown
## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below.
Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work**
2. **Run quality gates** (if code changed)
3. **Update issue status**
4. **PUSH TO REMOTE** - This is MANDATORY
5. **Verify** - All changes committed AND pushed
```

**Emphasis on Critical Rules:**
```markdown
**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing
- NEVER say "ready to push when you are" - YOU must push
```

---

## CLI_REFERENCE.md Pattern

### Structure

```
1. Overview table of all commands
2. Global flags section
3. Each command with:
   - Synopsis
   - Description
   - Flags table
   - Examples
```

### Key Elements

**Command Synopsis:**
```markdown
## bd create

Create a new issue.

### Synopsis

```
bd create <title> [flags]
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | `-t` | `task` | Issue type |
| `--priority` | `-p` | `2` | Priority (0-4) |
| `--description` | `-d` | | Issue description |
| `--json` | | `false` | Output as JSON |

### Examples

```bash
# Create a bug with high priority
bd create "Login fails on Safari" -t bug -p 1

# Create with description
bd create "Add dark mode" -d "Support system preference"
```
```

---

## TROUBLESHOOTING.md Pattern

### Structure

```
1. Quick fixes section (most common issues)
2. Categorized issues
3. Each issue with:
   - Symptoms
   - Cause
   - Solution
4. Recovery procedures
```

### Key Elements

**Issue Format:**
```markdown
### Issue: Database is locked

**Symptoms:**
```
bd: database is locked (SQLITE_BUSY)
```

**Cause:** Another process has the database open.

**Solutions:**

1. **Stop daemon and retry:**
   ```bash
   bd daemon stop
   bd <your-command>
   ```

2. **Check for hung processes:**
   ```bash
   ps aux | grep bd
   kill <pid>
   ```
```

**Quick Fixes Section:**
```markdown
## Quick Fixes

| Problem | Solution |
|---------|----------|
| "database is locked" | `bd daemon stop && bd daemon start` |
| "JSONL conflict markers" | `git checkout --theirs .beads/issues.jsonl` |
| "circular dependency" | `bd doctor` (diagnose only) |
```

---

## CONFIG.md Pattern

### Structure

```
1. Configuration overview
2. Configuration levels (project, user, env)
3. Settings table with all options
4. Examples for common scenarios
```

### Key Elements

**Configuration Levels:**
```markdown
## Configuration Precedence

1. **Environment variables** (highest) - `BEADS_*`
2. **CLI flags** - `--flag`
3. **Project config** - `.beads/config.yaml`
4. **User config** - `~/.config/beads/config.yaml`
5. **Defaults** (lowest)
```

**Settings Table:**
```markdown
## Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `sync.auto_commit` | bool | `true` | Auto-commit on sync |
| `sync.auto_push` | bool | `false` | Auto-push on sync |
| `sync.branch` | string | | Separate sync branch |
| `daemon.port` | int | `0` | Daemon port (0=auto) |
```

---

## CHANGELOG.md Pattern

### Format

Based on [Keep a Changelog](https://keepachangelog.com/):

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.48.0] - 2026-01-17

### Added
- **VersionedStorage interface** - Abstract storage layer
- **`bd types` command** - List valid issue types (#1102)

### Fixed
- **Doctor sync branch health check** - Removed destructive --fix (GH#1062)
- **Duplicate merge target selection** - Use combined weight (GH#1022)

### Changed
- **Daemon CLI refactor** - Consolidated subcommands

### Documentation
- Add lazybeads TUI to community tools (#951)
```

### Key Practices

1. **Link issue numbers** - `(#123)` or `(GH#123)`
2. **Bold feature names** - `**Feature name**`
3. **Categorize changes** - Added, Fixed, Changed, etc.
4. **Include dates** - `[0.48.0] - 2026-01-17`
5. **Link comparison URLs** at bottom

---

## Documentation Organization

### Directory Structure

```
beads/
├── README.md                 # Overview + quick start
├── AGENTS.md                 # AI assistant guide
├── CONTRIBUTING.md           # Contributor guide
├── CHANGELOG.md              # Version history
├── SECURITY.md               # Vulnerability reporting
│
├── docs/
│   ├── QUICKSTART.md         # Detailed getting started
│   ├── CLI_REFERENCE.md      # Complete command reference
│   ├── ARCHITECTURE.md       # System design
│   ├── CONFIG.md             # Configuration options
│   ├── TROUBLESHOOTING.md    # Common issues
│   ├── FAQ.md                # Frequently asked questions
│   ├── GIT_INTEGRATION.md    # Git workflows
│   ├── WORKTREES.md          # Git worktree support
│   ├── MULTI_REPO_*.md       # Multi-repo patterns
│   └── <FEATURE>.md          # Feature-specific docs
│
├── examples/
│   ├── README.md             # Examples index
│   ├── python-agent/         # Python integration
│   ├── bash-agent/           # Shell scripts
│   └── <pattern>/            # Usage patterns
│
└── integrations/
    ├── beads-mcp/            # MCP server
    └── claude-code/          # Codex plugin
```

### Navigation Principles

1. **README links to docs/** - Don't duplicate, link
2. **Each doc is self-contained** - Can be read standalone
3. **Cross-references** - "See also" sections
4. **Index pages** - examples/README.md lists all examples

---

## Style Guidelines

### From Beads' Writing Style

1. **Direct language** - "Run this command" not "You may want to run"
2. **Active voice** - "The daemon exports" not "Issues are exported by"
3. **Tables for structured data** - Commands, flags, options
4. **Code blocks for examples** - Always with language hint
5. **Warnings are prominent** - Use blockquotes or boxes
6. **No jargon without definition** - Explain terms on first use

### Warning Format

```markdown
**⚠️ WARNING:** Daemon mode does NOT work correctly with git worktrees.
```

Or:

```markdown
> **Note:** For environments with shell access, CLI is recommended over MCP.
```

### Example Quality

Bad:
```markdown
Run the create command to create an issue.
```

Good:
```markdown
```bash
bd create "Fix authentication bug" -t bug -p 1 --json
```
```

---

## Metrics from Beads

| Metric | Value |
|--------|-------|
| Total .md files | ~90 |
| README.md length | ~200 lines |
| CLI_REFERENCE.md | ~800 lines |
| TROUBLESHOOTING.md | ~845 lines |
| CONFIG.md | ~615 lines |
| CHANGELOG entries | 48+ versions |
| Integration guides | 4+ (MCP, Codex, Aider, etc.) |

### Coverage Analysis

- **Tier 1:** 4/4 (LICENSE, README, CONTRIBUTING, CODE_OF_CONDUCT)
- **Tier 2:** 5/5 (SECURITY, CHANGELOG, AGENTS, templates)
- **Tier 3:** 6+/6 (QUICKSTART, ARCHITECTURE, CLI_REFERENCE, CONFIG, TROUBLESHOOTING, examples)

**Score: 100% coverage across all tiers**

---

## Applying to Your Project

1. **Start with README.md** - Use beads' structure as template
2. **Add AGENTS.md early** - AI assistants need context
3. **Document commands** - CLI_REFERENCE.md for any CLI
4. **Anticipate problems** - TROUBLESHOOTING.md saves support time
5. **Keep CHANGELOG** - Start from v0.1.0, update every release

### documentation-tiers.md

# Documentation Tiers

> Prioritized documentation requirements for OSS projects.
> Based on analysis of successful open source projects.

## Overview

Not all documentation is created equal. This tiered approach ensures
critical files are prioritized while allowing progressive enhancement.

---

## Tier 1: Required (Legal + Essential)

**Must have for any public repository.**

| File | Purpose | Template |
|------|---------|----------|
| `LICENSE` | Legal terms for usage | Apache 2.0, MIT, etc. |
| `README.md` | First impression, quick start | Project-type specific |
| `CONTRIBUTING.md` | How to contribute | Fork/PR workflow |
| `CODE_OF_CONDUCT.md` | Community standards | Contributor Covenant |

### Why These Are Required

- **LICENSE**: Without a license, code is "all rights reserved" by default
- **README.md**: First file GitHub displays, defines project identity
- **CONTRIBUTING.md**: Reduces friction for new contributors
- **CODE_OF_CONDUCT.md**: Sets expectations, required by many organizations

### Audit Check

```bash
TIER1_SCORE=0
[[ -f LICENSE ]] && ((TIER1_SCORE++))
[[ -f README.md ]] && ((TIER1_SCORE++))
[[ -f CONTRIBUTING.md ]] && ((TIER1_SCORE++))
[[ -f CODE_OF_CONDUCT.md ]] && ((TIER1_SCORE++))
echo "Tier 1: $TIER1_SCORE/4"
```

---

## Tier 2: Standard (Professional Quality)

**Expected for production-quality projects.**

| File | Purpose | When Critical |
|------|---------|---------------|
| `SECURITY.md` | Vulnerability reporting | Always |
| `CHANGELOG.md` | Version history | Versioned releases |
| `AGENTS.md` | AI assistant context | AI-assisted development |
| `.github/ISSUE_TEMPLATE/` | Structured issue reports | Public issue tracker |
| `.github/PULL_REQUEST_TEMPLATE.md` | PR checklist | Active contributions |

### Why These Matter

- **SECURITY.md**: Private vulnerability disclosure channel
- **CHANGELOG.md**: Users need to know what changed between versions
- **AGENTS.md**: AI assistants (Claude, Copilot) work better with context
- **Issue Templates**: Reduce noise, get structured reports
- **PR Template**: Ensure consistency, remind of checklist items

### Audit Check

```bash
TIER2_SCORE=0
[[ -f SECURITY.md ]] && ((TIER2_SCORE++))
[[ -f CHANGELOG.md ]] && ((TIER2_SCORE++))
[[ -f AGENTS.md ]] && ((TIER2_SCORE++))
[[ -d .github/ISSUE_TEMPLATE ]] && ((TIER2_SCORE++))
[[ -f .github/PULL_REQUEST_TEMPLATE.md ]] && ((TIER2_SCORE++))
echo "Tier 2: $TIER2_SCORE/5"
```

---

## Tier 3: Enhanced (Comprehensive)

**For mature projects with complex functionality.**

| File | Purpose | Recommended When |
|------|---------|------------------|
| `docs/QUICKSTART.md` | Detailed getting started | Complex setup |
| `docs/ARCHITECTURE.md` | System design | Non-trivial codebase |
| `docs/CLI_REFERENCE.md` | Command documentation | CLI tools |
| `docs/CONFIG.md` | Configuration options | Configurable software |
| `docs/TROUBLESHOOTING.md` | Common issues | Production software |
| `docs/FAQ.md` | Frequently asked questions | Recurring questions |
| `examples/README.md` | Example index | Multiple examples |

### Recommendation Matrix

| Project Characteristic | Recommended Docs |
|------------------------|------------------|
| CLI tool | CLI_REFERENCE.md, QUICKSTART.md |
| Kubernetes operator | ARCHITECTURE.md, CONFIG.md |
| Library | API.md, examples/ |
| Complex config | CONFIG.md, TROUBLESHOOTING.md |
| Large codebase | ARCHITECTURE.md, INTERNALS.md |

### Audit Check

```bash
TIER3_SCORE=0
[[ -f docs/QUICKSTART.md ]] && ((TIER3_SCORE++))
[[ -f docs/ARCHITECTURE.md ]] && ((TIER3_SCORE++))
[[ -f docs/CLI_REFERENCE.md ]] && ((TIER3_SCORE++))
[[ -f docs/CONFIG.md ]] && ((TIER3_SCORE++))
[[ -f docs/TROUBLESHOOTING.md ]] && ((TIER3_SCORE++))
[[ -d examples ]] && ((TIER3_SCORE++))
echo "Tier 3: $TIER3_SCORE/6"
```

---

## Tier 4: Specialized

**Domain-specific documentation.**

| Category | Files |
|----------|-------|
| **API** | `docs/API.md`, OpenAPI spec |
| **Helm** | `docs/VALUES.md`, upgrade guides |
| **Operator** | CRD references, RBAC docs |
| **Protocol** | Wire format, versioning |
| **MCP** | Server setup, tool documentation |

---

## Scoring Guide

| Score Range | Status | Action |
|-------------|--------|--------|
| Tier 1 < 4 | Incomplete | Add missing required files |
| Tier 1 = 4, Tier 2 < 3 | Basic | Add standard files |
| Tier 1 = 4, Tier 2 >= 3 | Standard | Consider Tier 3 |
| All tiers complete | Comprehensive | Maintain and update |

---

## Progressive Enhancement Strategy

### Phase 1: Go Public (Tier 1)

Before making a repo public:
1. Add LICENSE (choose appropriate license)
2. Write README.md with basic info
3. Add CONTRIBUTING.md (fork/PR workflow)
4. Add CODE_OF_CONDUCT.md (Contributor Covenant)

### Phase 2: Attract Contributors (Tier 2)

After initial public release:
1. Add SECURITY.md for vulnerability reports
2. Start CHANGELOG.md for version tracking
3. Add issue/PR templates
4. Create AGENTS.md for AI assistants

### Phase 3: Scale (Tier 3)

As project grows:
1. Split README content into docs/
2. Add troubleshooting for common issues
3. Document architecture for contributors
4. Create comprehensive examples

---

## Examples from Beads

Beads (chronicle) demonstrates excellent documentation coverage:

**Tier 1 (all present):**
- LICENSE (MIT)
- README.md (comprehensive overview)
- CONTRIBUTING.md (detailed guide)
- CODE_OF_CONDUCT.md (Contributor Covenant)

**Tier 2 (all present):**
- SECURITY.md (vulnerability reporting)
- CHANGELOG.md (Keep a Changelog format)
- AGENTS.md (AI workflow guide)
- Issue templates (bug report, feature request)
- PR template

**Tier 3 (extensive):**
- docs/QUICKSTART.md
- docs/ARCHITECTURE.md
- docs/CLI_REFERENCE.md (~800 lines)
- docs/CONFIG.md (~615 lines)
- docs/TROUBLESHOOTING.md (~845 lines)
- docs/FAQ.md
- docs/GIT_INTEGRATION.md
- docs/WORKTREES.md
- examples/ directory with multiple patterns

**Key Patterns:**
- Clear separation between user docs and developer docs
- Extensive troubleshooting documentation
- Multiple integration guides (MCP, Codex, etc.)
- Active CHANGELOG with detailed version notes

### project-types.md

# Project Types Reference

> Documentation patterns by project category.
> Templates adapt to project type for relevant content.

## Type Detection

```bash
#!/bin/bash
# Detect project type based on file patterns

detect_project_type() {
    local type="unknown"
    local confidence=0

    # CLI Tool (Go)
    if [[ -f go.mod ]] && [[ -d cmd ]]; then
        type="cli-go"
        confidence=90

    # CLI Tool (Python)
    elif [[ -f pyproject.toml ]] && grep -q "scripts" pyproject.toml 2>/dev/null; then
        type="cli-python"
        confidence=85

    # Kubernetes Operator
    elif [[ -f PROJECT ]] || [[ -d config/crd ]] || [[ -f Makefile ]] && grep -q "controller-gen" Makefile 2>/dev/null; then
        type="operator"
        confidence=95

    # Helm Chart
    elif [[ -f Chart.yaml ]]; then
        type="helm"
        confidence=100

    # Go Library
    elif [[ -f go.mod ]] && [[ ! -d cmd ]]; then
        type="library-go"
        confidence=80

    # Python Library
    elif [[ -f pyproject.toml ]] || [[ -f setup.py ]]; then
        type="library-python"
        confidence=75

    # Node.js
    elif [[ -f package.json ]]; then
        if grep -q '"bin"' package.json 2>/dev/null; then
            type="cli-node"
            confidence=85
        else
            type="library-node"
            confidence=75
        fi

    # Rust
    elif [[ -f Cargo.toml ]]; then
        if [[ -d src/bin ]] || grep -q '^\[\[bin\]\]' Cargo.toml 2>/dev/null; then
            type="cli-rust"
            confidence=85
        else
            type="library-rust"
            confidence=80
        fi

    # Documentation/Informational
    elif [[ -d docs ]] && [[ $(find . -maxdepth 1 -name "*.md" | wc -l) -gt 5 ]]; then
        type="docs"
        confidence=70
    fi

    echo "$type:$confidence"
}
```

---

## Type: cli-go

**Go CLI tools (like beads, gastown)**

### Detection Signals
- `go.mod` present
- `cmd/` directory with main packages
- Often has `internal/` for private packages

### Recommended Documentation

| File | Priority | Content Focus |
|------|----------|---------------|
| `README.md` | Required | Installation (brew, go install), quick start |
| `docs/CLI_REFERENCE.md` | High | All commands with flags |
| `docs/QUICKSTART.md` | High | First-run experience |
| `docs/CONFIG.md` | Medium | Config files, env vars |
| `docs/TROUBLESHOOTING.md` | Medium | Common errors, fixes |
| `examples/` | Medium | Usage examples |

### README Template Key Sections

```markdown
## Installation

```bash
# Homebrew (recommended)
brew install <name>

# Go install
go install <module>/cmd/<name>@latest

# From source
git clone <repo>
cd <repo>
go build -o <name> ./cmd/<name>
```

## Quick Start

```bash
<name> init
<name> <primary-command>
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize configuration |
| `<cmd>` | Primary operation |
| `help` | Show help |
```

---

## Type: operator

**Kubernetes Operators (kubebuilder, operator-sdk)**

### Detection Signals
- `PROJECT` file (kubebuilder marker)
- `config/crd/` directory
- `Makefile` with controller-gen references
- `api/` or `apis/` directory with types

### Recommended Documentation

| File | Priority | Content Focus |
|------|----------|---------------|
| `README.md` | Required | What it manages, quick install |
| `docs/ARCHITECTURE.md` | High | Controllers, reconciliation |
| `docs/CONFIG.md` | High | CRD spec fields |
| `SECURITY.md` | High | RBAC, pod security |
| `docs/TROUBLESHOOTING.md` | Medium | Common issues |

### README Template Key Sections

```markdown
## Installation

```bash
kubectl apply -f https://github.com/<owner>/<repo>/releases/latest/download/install.yaml
```

Or with Helm:
```bash
helm install <name> <repo>/<chart>
```

## CRDs

| Kind | API Version | Description |
|------|-------------|-------------|
| `<Kind>` | `<group>/<version>` | Manages... |

## Quick Start

```yaml
apiVersion: <group>/<version>
kind: <Kind>
metadata:
  name: example
spec:
  # minimal spec
```

## RBAC Requirements

The operator requires the following permissions:
- `<resource>`: create, get, list, watch, update, delete
```

### SECURITY.md Focus

```markdown
## Security Considerations

- **Pod Security:** Runs with restricted security context
- **RBAC:** Minimal permissions following least-privilege
- **Secrets:** Never logged, stored encrypted at rest
- **Network:** Egress to API server only
```

---

## Type: helm

**Helm Charts**

### Detection Signals
- `Chart.yaml` present
- `values.yaml` present
- `templates/` directory

### Recommended Documentation

| File | Priority | Content Focus |
|------|----------|---------------|
| `README.md` | Required | Installation, basic values |
| `docs/VALUES.md` | High | All values documented |
| `docs/UPGRADING.md` | Medium | Version migration |

### README Template Key Sections

```markdown
## Installation

```bash
helm repo add <repo> <url>
helm install <release> <repo>/<chart>
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image name | `<default>` |
| `image.tag` | Image tag | `latest` |
| `replicas` | Pod replicas | `1` |

See `values.yaml` for all options.

## Upgrading

```bash
helm upgrade <release> <repo>/<chart>
```
```

---

## Type: library-go

**Go Libraries**

### Detection Signals
- `go.mod` present
- No `cmd/` directory
- Public package exports

### Recommended Documentation

| File | Priority | Content Focus |
|------|----------|---------------|
| `README.md` | Required | Installation, basic usage |
| `docs/API.md` | High | Public API reference |
| `examples/` | High | Usage patterns |

### README Template Key Sections

```markdown
## Installation

```bash
go get <module>
```

## Usage

```go
import "<module>"

func main() {
    client := pkg.New()
    result, err := client.DoSomething()
}
```

## API

See [pkg.go.dev](https://pkg.go.dev/<module>) for complete API documentation.
```

---

## Type: library-python

**Python Libraries**

### Detection Signals
- `pyproject.toml` or `setup.py`
- `src/` or package directory
- No CLI entry points

### Recommended Documentation

| File | Priority | Content Focus |
|------|----------|---------------|
| `README.md` | Required | Installation, basic usage |
| `docs/API.md` | High | Public API reference |
| `examples/` | High | Usage notebooks/scripts |

### README Template Key Sections

```markdown
## Installation

```bash
pip install <package>
# or
uv pip install <package>
```

## Usage

```python
from <package> import Client

client = Client()
result = client.do_something()
```

## API Documentation

See your hosted API documentation URL for complete API reference.
```

---

## Type: cli-python

**Python CLI Tools**

### Detection Signals
- `pyproject.toml` with `[project.scripts]`
- Click, Typer, or argparse usage
- Entry point defined

### Recommended Documentation

Similar to cli-go but with Python installation methods:

```markdown
## Installation

```bash
# pip
pip install <package>

# pipx (recommended for CLI tools)
pipx install <package>

# uv
uv tool install <package>
```
```

---

## Type: docs

**Documentation-Only Repositories**

### Detection Signals
- Heavy markdown content
- `docs/` directory dominant
- Minimal code

### Recommended Documentation

| File | Priority | Content Focus |
|------|----------|---------------|
| `README.md` | Required | Navigation, purpose |
| `CONTRIBUTING.md` | High | How to contribute docs |
| `docs/index.md` | High | Main entry point |

---

## Language Detection

```bash
#!/bin/bash
# Detect languages in project

detect_languages() {
    local langs=()

    [[ -f go.mod ]] && langs+=("go")
    [[ -f pyproject.toml ]] || [[ -f setup.py ]] && langs+=("python")
    [[ -f package.json ]] && langs+=("javascript")
    [[ -f Cargo.toml ]] && langs+=("rust")
    [[ -f Makefile ]] && langs+=("make")
    [[ $(find . -name "*.sh" -maxdepth 2 | wc -l) -gt 0 ]] && langs+=("shell")
    [[ -f Dockerfile ]] && langs+=("docker")
    [[ -f Chart.yaml ]] && langs+=("helm")

    echo "${langs[*]}"
}
```

---

## Command Extraction

For CLI tools, extract commands for documentation:

### Go (cobra)

```bash
# Find cobra commands
grep -r "func.*Command\(\)" cmd/ --include="*.go" | \
    sed 's/.*func \(.*\)Command.*/\1/'
```

### Python (click/typer)

```bash
# Find click commands
grep -r "@click.command\|@app.command" --include="*.py" | \
    sed 's/.*def \([a-z_]*\).*/\1/'
```

---

## Test Command Detection

```bash
detect_test_command() {
    if [[ -f go.mod ]]; then
        echo "go test ./..."
    elif [[ -f pyproject.toml ]]; then
        if grep -q "pytest" pyproject.toml; then
            echo "pytest"
        else
            echo "python -m pytest"
        fi
    elif [[ -f package.json ]]; then
        echo "npm test"
    elif [[ -f Cargo.toml ]]; then
        echo "cargo test"
    elif [[ -f Makefile ]] && grep -q "^test:" Makefile; then
        echo "make test"
    else
        echo "<TEST_COMMAND>"
    fi
}
```


---

## Scripts

### audit-oss-docs.sh

```bash
#!/bin/bash
# OSS Documentation Audit Script
# Usage: audit-oss-docs.sh [--json]
#
# Checks for presence of standard OSS documentation files
# and reports coverage across tiers.

set -e

JSON_OUTPUT=false
[[ "$1" == "--json" ]] && JSON_OUTPUT=true

# Colors (disabled for JSON output)
if [[ "$JSON_OUTPUT" == "false" ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

# Project detection
PROJECT_NAME=$(basename "$(pwd)")
GIT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")

# Detect project type
# Order matters: more specific types checked first
detect_type() {
    # Kubernetes Operator (kubebuilder/operator-sdk) - check BEFORE cli-go
    # because operators also have go.mod + cmd/
    if [[ -f PROJECT ]] || [[ -d config/crd ]] || [[ -d config/rbac ]]; then
        echo "operator"
    # Helm Chart
    elif [[ -f Chart.yaml ]]; then
        echo "helm"
    # Go CLI Tool
    elif [[ -f go.mod ]] && [[ -d cmd ]]; then
        echo "cli-go"
    # Python CLI Tool (has entry points)
    elif [[ -f pyproject.toml ]] && grep -q "\[project.scripts\]" pyproject.toml 2>/dev/null; then
        echo "cli-python"
    # Go Library (go.mod but no cmd/)
    elif [[ -f go.mod ]]; then
        echo "library-go"
    # Python Library
    elif [[ -f pyproject.toml ]] || [[ -f setup.py ]]; then
        echo "library-python"
    # Node.js
    elif [[ -f package.json ]]; then
        if grep -q '"bin"' package.json 2>/dev/null; then
            echo "cli-node"
        else
            echo "library-node"
        fi
    # Rust
    elif [[ -f Cargo.toml ]]; then
        if [[ -d src/bin ]] || grep -q '^\[\[bin\]\]' Cargo.toml 2>/dev/null; then
            echo "cli-rust"
        else
            echo "library-rust"
        fi
    else
        echo "unknown"
    fi
}

# Detect languages
detect_languages() {
    local langs=()
    [[ -f go.mod ]] && langs+=("go")
    [[ -f pyproject.toml ]] || [[ -f setup.py ]] && langs+=("python")
    [[ -f package.json ]] && langs+=("javascript")
    [[ -f Cargo.toml ]] && langs+=("rust")
    [[ -f Makefile ]] && langs+=("make")
    [[ -f Dockerfile ]] && langs+=("docker")
    [[ -f Chart.yaml ]] && langs+=("helm")
    echo "${langs[*]}"
}

PROJECT_TYPE=$(detect_type)
LANGUAGES=$(detect_languages)

# Tier 1: Required
check_tier1() {
    local score=0
    local total=4
    local results=()

    if [[ -f LICENSE ]]; then
        results+=("LICENSE:pass")
        ((score++))
    else
        results+=("LICENSE:fail")
    fi

    if [[ -f README.md ]]; then
        results+=("README.md:pass")
        ((score++))
    else
        results+=("README.md:fail")
    fi

    if [[ -f CONTRIBUTING.md ]]; then
        results+=("CONTRIBUTING.md:pass")
        ((score++))
    else
        results+=("CONTRIBUTING.md:fail")
    fi

    if [[ -f CODE_OF_CONDUCT.md ]]; then
        results+=("CODE_OF_CONDUCT.md:pass")
        ((score++))
    else
        results+=("CODE_OF_CONDUCT.md:fail")
    fi

    echo "$score:$total:${results[*]}"
}

# Tier 2: Standard
check_tier2() {
    local score=0
    local total=5
    local results=()

    if [[ -f SECURITY.md ]]; then
        results+=("SECURITY.md:pass")
        ((score++))
    else
        results+=("SECURITY.md:fail")
    fi

    if [[ -f CHANGELOG.md ]]; then
        results+=("CHANGELOG.md:pass")
        ((score++))
    else
        results+=("CHANGELOG.md:fail")
    fi

    if [[ -f AGENTS.md ]]; then
        results+=("AGENTS.md:pass")
        ((score++))
    else
        results+=("AGENTS.md:fail")
    fi

    if [[ -d .github/ISSUE_TEMPLATE ]]; then
        results+=("issue_templates:pass")
        ((score++))
    else
        results+=("issue_templates:fail")
    fi

    if [[ -f .github/PULL_REQUEST_TEMPLATE.md ]]; then
        results+=("pr_template:pass")
        ((score++))
    else
        results+=("pr_template:fail")
    fi

    echo "$score:$total:${results[*]}"
}

# Tier 3: Enhanced (with recommendations)
check_tier3() {
    local score=0
    local total=6
    local results=()

    # QUICKSTART - recommended for all
    if [[ -f docs/QUICKSTART.md ]]; then
        results+=("docs/QUICKSTART.md:pass:recommended")
        ((score++))
    else
        results+=("docs/QUICKSTART.md:fail:recommended")
    fi

    # ARCHITECTURE - recommended for non-trivial projects
    if [[ -f docs/ARCHITECTURE.md ]]; then
        results+=("docs/ARCHITECTURE.md:pass:conditional")
        ((score++))
    else
        local rec="optional"
        # Recommend if large codebase
        [[ $(find . -name "*.go" -o -name "*.py" 2>/dev/null | wc -l) -gt 20 ]] && rec="recommended"
        results+=("docs/ARCHITECTURE.md:fail:$rec")
    fi

    # CLI_REFERENCE - recommended for CLI tools
    # CRD_REFERENCE - recommended for operators (check for either)
    if [[ -f docs/CLI_REFERENCE.md ]] || [[ -f docs/CRD_REFERENCE.md ]]; then
        local found_file="docs/CLI_REFERENCE.md"
        [[ -f docs/CRD_REFERENCE.md ]] && found_file="docs/CRD_REFERENCE.md"
        results+=("$found_file:pass:conditional")
        ((score++))
    else
        local rec="optional"
        local check_file="docs/CLI_REFERENCE.md"
        if [[ "$PROJECT_TYPE" == "operator" ]]; then
            check_file="docs/CRD_REFERENCE.md"
            rec="recommended"
        elif [[ "$PROJECT_TYPE" == "cli-go" ]] || [[ "$PROJECT_TYPE" == "cli-python" ]] || [[ "$PROJECT_TYPE" == "cli-node" ]] || [[ "$PROJECT_TYPE" == "cli-rust" ]]; then
            rec="recommended"
        fi
        results+=("$check_file:fail:$rec")
    fi

    # CONFIG - recommended if configurable or operator
    if [[ -f docs/CONFIG.md ]]; then
        results+=("docs/CONFIG.md:pass:conditional")
        ((score++))
    else
        local rec="optional"
        # Operators should document CRD spec fields
        [[ "$PROJECT_TYPE" == "operator" ]] && rec="recommended"
        [[ -f config.yaml ]] || [[ -d config ]] && rec="recommended"
        results+=("docs/CONFIG.md:fail:$rec")
    fi

    # TROUBLESHOOTING - recommended for production software
    if [[ -f docs/TROUBLESHOOTING.md ]]; then
        results+=("docs/TROUBLESHOOTING.md:pass:conditional")
        ((score++))
    else
        results+=("docs/TROUBLESHOOTING.md:fail:optional")
    fi

    # examples/ directory
    if [[ -d examples ]]; then
        results+=("examples/:pass:recommended")
        ((score++))
    else
        results+=("examples/:fail:optional")
    fi

    echo "$score:$total:${results[*]}"
}

# Parse tier results
parse_results() {
    local tier_data="$1"
    local score="${tier_data%%:*}"
    local rest="${tier_data#*:}"
    local total="${rest%%:*}"
    local items="${rest#*:}"
    echo "$score" "$total" "$items"
}

# Run checks
TIER1=$(check_tier1)
TIER2=$(check_tier2)
TIER3=$(check_tier3)

read -r T1_SCORE T1_TOTAL T1_ITEMS <<< "$(parse_results "$TIER1")"
read -r T2_SCORE T2_TOTAL T2_ITEMS <<< "$(parse_results "$TIER2")"
read -r T3_SCORE T3_TOTAL T3_ITEMS <<< "$(parse_results "$TIER3")"

TOTAL_SCORE=$((T1_SCORE + T2_SCORE + T3_SCORE))
TOTAL_POSSIBLE=$((T1_TOTAL + T2_TOTAL + T3_TOTAL))

# Output
if [[ "$JSON_OUTPUT" == "true" ]]; then
    # JSON output
    cat <<EOF
{
  "project": "$PROJECT_NAME",
  "type": "$PROJECT_TYPE",
  "languages": "$(echo $LANGUAGES | tr ' ' ',')",
  "tier1": {
    "score": $T1_SCORE,
    "total": $T1_TOTAL,
    "items": [$(echo "$T1_ITEMS" | tr ' ' '\n' | sed 's/\(.*\):\(.*\)/{"file":"\1","status":"\2"}/' | tr '\n' ',' | sed 's/,$//' )]
  },
  "tier2": {
    "score": $T2_SCORE,
    "total": $T2_TOTAL,
    "items": [$(echo "$T2_ITEMS" | tr ' ' '\n' | sed 's/\(.*\):\(.*\)/{"file":"\1","status":"\2"}/' | tr '\n' ',' | sed 's/,$//' )]
  },
  "tier3": {
    "score": $T3_SCORE,
    "total": $T3_TOTAL,
    "items": [$(echo "$T3_ITEMS" | tr ' ' '\n' | sed 's/\([^:]*\):\([^:]*\):\(.*\)/{"file":"\1","status":"\2","recommendation":"\3"}/' | tr '\n' ',' | sed 's/,$//' )]
  },
  "total_score": $TOTAL_SCORE,
  "total_possible": $TOTAL_POSSIBLE
}
EOF
else
    # Human-readable output
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  OSS Documentation Audit: ${PROJECT_NAME}${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "Project Type: ${YELLOW}$PROJECT_TYPE${NC}"
    echo -e "Languages: ${YELLOW}$LANGUAGES${NC}"
    echo ""

    # Tier 1
    echo -e "${BLUE}── Tier 1: Required ──${NC}"
    for item in $T1_ITEMS; do
        file="${item%%:*}"
        status="${item##*:}"
        if [[ "$status" == "pass" ]]; then
            echo -e "  ${GREEN}✓${NC} $file"
        else
            echo -e "  ${RED}✗${NC} $file"
        fi
    done
    echo -e "  Score: ${T1_SCORE}/${T1_TOTAL}"
    echo ""

    # Tier 2
    echo -e "${BLUE}── Tier 2: Standard ──${NC}"
    for item in $T2_ITEMS; do
        file="${item%%:*}"
        status="${item##*:}"
        if [[ "$status" == "pass" ]]; then
            echo -e "  ${GREEN}✓${NC} $file"
        else
            echo -e "  ${RED}✗${NC} $file"
        fi
    done
    echo -e "  Score: ${T2_SCORE}/${T2_TOTAL}"
    echo ""

    # Tier 3
    echo -e "${BLUE}── Tier 3: Enhanced ──${NC}"
    for item in $T3_ITEMS; do
        IFS=':' read -r file status rec <<< "$item"
        if [[ "$status" == "pass" ]]; then
            echo -e "  ${GREEN}✓${NC} $file"
        elif [[ "$rec" == "recommended" ]]; then
            echo -e "  ${YELLOW}✗${NC} $file (recommended)"
        else
            echo -e "  ${NC}○${NC} $file (optional)"
        fi
    done
    echo -e "  Score: ${T3_SCORE}/${T3_TOTAL}"
    echo ""

    # Summary
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    if [[ $T1_SCORE -lt $T1_TOTAL ]]; then
        echo -e "${RED}  Status: INCOMPLETE - Missing required files${NC}"
    elif [[ $T2_SCORE -lt 3 ]]; then
        echo -e "${YELLOW}  Status: BASIC - Consider adding standard files${NC}"
    elif [[ $T3_SCORE -lt 3 ]]; then
        echo -e "${GREEN}  Status: STANDARD - Ready for public${NC}"
    else
        echo -e "${GREEN}  Status: COMPREHENSIVE - Well documented${NC}"
    fi
    echo -e "  Total Score: ${TOTAL_SCORE}/${TOTAL_POSSIBLE}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"

    # Scaffold hint
    if [[ $TOTAL_SCORE -lt $TOTAL_POSSIBLE ]]; then
        echo ""
        echo "To scaffold missing files:"
        echo "  $oss-docs scaffold"
    fi
fi
```

### validate.sh

```bash
#!/bin/bash
# Validate oss-docs skill
set -euo pipefail

# Determine SKILL_DIR relative to this script (works in plugins or ~/.claude)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

ERRORS=0
CHECKS=0

check_pattern() {
    local desc="$1"
    local file="$2"
    local pattern="$3"

    CHECKS=$((CHECKS + 1))
    if grep -qiE "$pattern" "$file" 2>/dev/null; then
        echo "✓ $desc"
    else
        echo "✗ $desc (pattern '$pattern' not found in $file)"
        ERRORS=$((ERRORS + 1))
    fi
}

check_exists() {
    local desc="$1"
    local path="$2"

    CHECKS=$((CHECKS + 1))
    if [ -e "$path" ]; then
        echo "✓ $desc"
    else
        echo "✗ $desc ($path not found)"
        ERRORS=$((ERRORS + 1))
    fi
}

echo "=== OSS-Docs Skill Validation ==="
echo ""


# Verify dependent skill exists
check_exists "Standards skill exists" "$HOME/.claude/skills/standards/SKILL.md"

# Verify oss-docs workflow patterns in SKILL.md
check_pattern "SKILL.md has README documentation" "$SKILL_DIR/SKILL.md" "README"
check_pattern "SKILL.md has CONTRIBUTING documentation" "$SKILL_DIR/SKILL.md" "CONTRIBUTING"
check_pattern "SKILL.md has open source patterns" "$SKILL_DIR/SKILL.md" "[Oo]pen [Ss]ource|OSS"
check_pattern "SKILL.md mentions AGENTS.md" "$SKILL_DIR/SKILL.md" "AGENTS.md"

echo ""
echo "=== Results ==="
echo "Checks: $CHECKS"
echo "Errors: $ERRORS"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "FAIL: OSS-docs skill validation failed"
    exit 1
else
    echo ""
    echo "PASS: OSS-docs skill validation passed"
    exit 0
fi
```


