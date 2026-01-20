# Changelog

All notable changes to the AgentOps marketplace will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-01-20

### Added
- **general-kit v1.0.0** - Standalone plugin with zero dependencies:
  - `/research`, `/vibe`, `/vibe-docs`, `/bug-hunt`, `/complexity`, `/validation-chain`
  - `/doc`, `/oss-docs`, `/golden-init`
  - 4 expert agents: security-expert, architecture-expert, code-quality-expert, ux-expert
- **standards library skill** (domain-kit) - Language-specific validation rules:
  - Python, Go, TypeScript, Shell, YAML, Markdown, JSON references
  - OpenAI platform standards (prompts, functions, responses, reasoning, GPT-OSS)
- **Context inference** for vibe and crank skills - Auto-detect targets from conversation
- **Natural language triggers** - Skills activate from intent, not just slash commands

### Changed
- **README overhaul**:
  - Added ASCII art logo and workflow diagrams
  - "Just Talk Naturally" section showing intent-based triggering
  - "The Killer Workflow: Plan → Crank" section with Shift+Tab + /formulate pattern
  - Clarified this provides plugins FOR beads/gastown, not built on them
  - Added OpenCode compatibility section
- **vibe skill** - Now references standards library for language-specific validation
- **validation-chain skill** - Added standards dependency
- **vibe-docs skill** - Added standards dependency

### Fixed
- **Standards dependencies** - Added missing `standards` skill dependency to:
  - vibe (vibe-kit, general-kit)
  - validation-chain (vibe-kit, general-kit)
  - vibe-docs (vibe-kit, general-kit)
- **Vibe findings** - Addressed quality findings across all plugins
- **Cross-skill references** - Test validator now handles relative paths correctly
- **Personal identifiers** - Removed from public plugin files

---

## [0.1.0] - 2026-01-19

### Added
- **Unix Philosophy Restructure** - Plugins reorganized into 8 focused kits:
  - **core-kit v1.0.0** - Workflow: research, plan, formulate, product, implement, implement-wave, crank, retro
  - **vibe-kit v2.0.0** - Validation only: vibe, vibe-docs, validation-chain, bug-hunt, complexity (+ 4 expert agents)
  - **docs-kit v1.0.0** - Documentation: doc, oss-docs, golden-init
  - **beads-kit v1.0.0** - Issue tracking: beads, status, molecules
  - **dispatch-kit v1.0.0** - Orchestration: dispatch, handoff, roles, mail
  - **pr-kit v1.0.0** - OSS contribution: pr-research, pr-plan, pr-implement, pr-validate, pr-prep, pr-retro
  - **gastown-kit v1.0.0** - Gas Town: crew, polecat-lifecycle, gastown, bd-routing
  - **domain-kit v1.0.0** - Reference knowledge: 18 domain skills (languages, development, security, etc.)

### Changed
- **vibe-kit** - Slimmed down from 23 skills to 5 focused validation skills
- **gastown plugin** - Replaced by gastown-kit (Gas Town specific) + pr-kit (contribution workflow)
- **Main README** - Updated with Unix philosophy structure, recommended combinations, clearer skill guidance
- **Core kit README** - Added decision trees for implement vs crank vs implement-wave

### Removed
- **gastown plugin** - Split into gastown-kit and pr-kit for better modularity

### Fixed
- **vibe-kit missing skills** - Restored vibe and vibe-docs skills that were lost during restructure

### Consolidated
- **domain-kit v1.1.0** - Consolidated from 18 to 17 skills:
  - Removed `doc-curator` (redundant with docs-kit/doc)
  - Consolidated 7 `base/` utilities (audit-diataxis, audit-onboarding, audit-workflow, cleanup-deprecated, cleanup-docs, cleanup-plans, cleanup-repo) into single `maintenance` skill

### Skill Counts (Final)
| Kit | Skills |
|-----|--------|
| core-kit | 8 |
| vibe-kit | 5 |
| docs-kit | 3 |
| beads-kit | 3 |
| dispatch-kit | 4 |
| pr-kit | 6 |
| gastown-kit | 4 |
| domain-kit | 17 |
| **Total** | **50** |

---

### Previous Unreleased

#### Added
- **vibe-kit v1.1.0** - New skills added:
  - `implement-wave` - Parallel execution of multiple issues
  - `complexity` - Code complexity analysis using radon/gocyclo
  - `doc` - Documentation generation and validation
  - `oss-docs` - OSS documentation scaffolding (README, CONTRIBUTING, SECURITY)
  - `golden-init` - Repository initialization with Golden Template
  - `molecules` - Workflow templates and formula TOML patterns
- **Skills sync** - All skills updated to match latest local versions:
  - beads, bug-hunt, dispatch, implement, research, vibe, vibe-docs (vibe-kit)
  - All 18 gastown plugin skills updated

### Fixed
- **Painted doors removed** - Cleaned up non-functional references:
  - Removed empty `references/` directories (bug-hunt, implement, pr-research, pr-retro)
  - Fixed pr-research template reference to point to inline section

### Changed
- **Commands deprecated** - Commands directory marked as deprecated in favor of skills
  - Added deprecation notice to commands/INDEX.md
  - Added migration guide pointing to skill equivalents
  - Commands maintained for legacy compatibility only
- **vibe-kit plugin.json** updated to version 1.1.0 with new skills

### Previous Unreleased

- **vibe-check Integration** in session-management plugin
  - `/session-start` now captures baseline metrics via `vibe-check session start`
  - `/session-end` now captures session metrics and failure patterns via `vibe-check session end`
  - Automatic failure pattern detection (Debug Spiral, Context Amnesia, Velocity Crash, Trust Erosion, Flow Disruption)
  - Session entries in `claude-progress.json` now include metrics and retro blocks
  - `@boshu2/vibe-check` npm package added as plugin dependency
- **vibe-coding Plugin** added with commands:
  - `/vibe-check` - Run vibe-check analysis
  - `/vibe-level` - Declare vibe level for session
  - `/vibe-retro` - Run vibe-coding retrospective
- **constitution Plugin** added with:
  - laws-of-an-agent skill
  - context-engineering skill
  - git-discipline skill
  - guardian agent
- SECURITY.md with vulnerability reporting process
- CONTRIBUTING.md with comprehensive plugin submission guidelines
- CHANGELOG.md for version tracking
- CODE_OF_CONDUCT.md for community standards
- GitHub Actions CI/CD pipeline for automated validation
- GitHub issue templates for plugin submissions and bug reports
- GitHub PR template for structured contributions
- Test suite infrastructure with validation scripts
- Makefile for common development tasks
- ARCHITECTURE_REVIEW.md with comprehensive compliance analysis

### Changed
- Updated repository structure to follow GitHub best practices
- Enhanced documentation for better discoverability

### Security
- Established security policy and vulnerability reporting process
- Added automated security scanning (Dependabot, CodeQL)

## [1.0.0] - 2025-11-10

### Added
- Initial marketplace structure with `.claude-plugin/marketplace.json`
- Three core plugins:
  - **core-workflow**: Universal Research → Plan → Implement → Learn workflow
  - **devops-operations**: DevOps and platform engineering tools
  - **software-development**: Software development for Python, JavaScript, Go
- External marketplace references:
  - aitmpl.com/agents (63+ plugins)
  - wshobson/agents (open source collection)
- Comprehensive README with quick start guide
- Apache 2.0 license
- Plugin structure following Anthropic 2025 standards
- 12-Factor AgentOps integration in all agents
- Token budget estimation for plugins

### Agents (11 total)
- **core-workflow** (4 agents):
  - research-agent: Research phase with JIT context loading
  - plan-agent: Planning phase with detailed specifications
  - implement-agent: Implementation phase with validation
  - learn-agent: Learning extraction for continuous improvement
- **devops-operations** (3 agents):
  - devops-engineer: DevOps automation specialist
  - deployment-engineer: Deployment and release management
  - cicd-specialist: CI/CD pipeline expert
- **software-development** (3 agents):
  - software-engineer: General software development
  - code-reviewer: Code quality and review
  - test-engineer: Testing and quality assurance

### Commands (14 total)
- **core-workflow** (5 commands):
  - /research: Start research phase
  - /plan: Create implementation plan
  - /implement: Execute approved plan
  - /learn: Extract learnings
  - /workflow: Full workflow orchestration
- **devops-operations** (3 commands):
  - /deploy-app: Deploy applications
  - /setup-pipeline: Configure CI/CD pipelines
  - /rollback: Rollback deployments
- **software-development** (3 commands):
  - /create-feature: Create new features
  - /refactor-code: Refactor existing code
  - /add-tests: Add test coverage

### Skills (9 total)
- **core-workflow**: Universal workflow patterns
- **devops-operations** (3 skills):
  - gitops-patterns: GitOps workflow patterns
  - kubernetes-manifests: Kubernetes resource templates
  - helm-charts: Helm chart best practices
- **software-development** (3 skills):
  - python-testing: Python testing patterns
  - javascript-patterns: JavaScript/TypeScript patterns
  - go-best-practices: Go language best practices

### Documentation
- Comprehensive README.md with installation instructions
- Plugin-level README files for each plugin
- Agent documentation with examples and anti-patterns
- AgentOps principles integration
- External marketplace references

## Version History

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version: Incompatible API changes
- **MINOR** version: New functionality (backwards-compatible)
- **PATCH** version: Bug fixes (backwards-compatible)

### Release Process

1. Update CHANGELOG.md with changes
2. Update version in `.claude-plugin/marketplace.json`
3. Update version in all plugin `plugin.json` files
4. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
5. Push tag: `git push origin v1.0.0`
6. Create GitHub release with changelog excerpt

## Links

- [Repository](https://github.com/boshu2/agentops)
- [Issues](https://github.com/boshu2/agentops/issues)
- [Pull Requests](https://github.com/boshu2/agentops/pulls)
- [Security Policy](SECURITY.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [12-Factor AgentOps Framework](https://github.com/boshu2/12-factor-agentops)

## Community

### How to Stay Updated

- Watch this repository on GitHub
- Check this CHANGELOG regularly
- Follow [@boshu2](https://github.com/boshu2) on GitHub

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on:
- How to add plugins
- Testing requirements
- Submission process
- Code of conduct

### Support

- **Documentation**: Check README.md and plugin docs
- **Issues**: [GitHub Issues](https://github.com/boshu2/agentops/issues)
- **Discussions**: [GitHub Discussions](https://github.com/boshu2/agentops/discussions)

---

**Note:** This changelog is automatically updated with each release. See [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) for format guidelines.
