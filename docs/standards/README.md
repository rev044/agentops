# Coding Standards

<!-- Canonical source: gitops/docs/standards/README.md -->
<!-- Last synced: 2026-01-19 -->

> **This repository follows the canonical standards defined in gitops/docs/standards/.**
>
> When updating standards, first update gitops then sync here.

---

## Standards Index

| Language/Format | Document | Gate | Pre-commit | AI-Friendly |
|-----------------|----------|------|------------|-------------|
| **Python** | [python-style-guide.md](./python-style-guide.md) | CC ≤ 10 | ruff | ★★★★★ |
| **Shell/Bash** | [shell-script-standards.md](./shell-script-standards.md) | shellcheck | shellcheck | ★★★★★ |
| **Go** | [golang-style-guide.md](./golang-style-guide.md) | golangci-lint, CC ≤ 10 | - | ★★★★★ |
| **TypeScript** | [typescript-standards.md](./typescript-standards.md) | tsc --strict | eslint | ★★★★★ |
| **YAML/Helm** | [yaml-helm-standards.md](./yaml-helm-standards.md) | yamllint | yamllint | ★★★★★ |
| **Markdown** | [markdown-style-guide.md](./markdown-style-guide.md) | markdownlint | - | ★★★★★ |
| **JSON/JSONL** | [json-jsonl-standards.md](./json-jsonl-standards.md) | jq/prettier | - | ★★★★★ |
| **Documentation Tags** | [tag-vocabulary.md](./tag-vocabulary.md) | - | - | ★★★★★ |

---

## What Makes These Standards AI-Agent-Friendly

Every standard in this repository is optimized for AI agent execution:

| Feature | Implementation |
|---------|----------------|
| **Tables over prose** | Scannable, parallel parsing |
| **Decision trees** | Executable if/then logic |
| **Common Errors tables** | Symptom → Cause → Fix |
| **Anti-Patterns (named)** | Recognizable error states |
| **AI Agent Guidelines** | ALWAYS/NEVER rules |
| **Explicit thresholds** | Numbers, not "be careful" |
| **Copy-paste examples** | Ready to use, not fragments |

---

## Complexity Requirements

All functions MUST meet complexity thresholds:

| Language | Tool | Threshold | Style Guide |
|----------|------|-----------|-------------|
| Python | radon/xenon | CC ≤ 10 | [Yes](./python-style-guide.md#code-complexity) |
| Go | gocyclo | CC ≤ 10 | [Yes](./golang-style-guide.md#code-complexity) |
| Shell | shellcheck | Pass all | [Yes](./shell-script-standards.md) |
| TypeScript | tsc --strict | No errors | [Yes](./typescript-standards.md) |
| YAML/Helm | yamllint | Pass all | [Yes](./yaml-helm-standards.md) |

---

## Pre-commit Enforcement

Standards are enforced via pre-commit hooks where available.

### Quick Start

```bash
# Install hooks (one-time)
pre-commit install

# Run manually
pre-commit run --all-files

# Or skip hooks (not recommended)
git commit --no-verify
```

### Manual Validation Commands

| Language | Command |
|----------|---------|
| **Python** | `ruff check scripts/ && xenon scripts/ --max-absolute B` |
| **Go** | `golangci-lint run ./... && gocyclo -over 10 ./...` |
| **Shell** | `shellcheck scripts/*.sh` |
| **TypeScript** | `tsc --noEmit && eslint . --ext .ts,.tsx` |
| **YAML** | `yamllint .` |
| **Markdown** | `npx markdownlint '**/*.md'` |
| **JSON** | `jq empty config.json` (validates syntax) |

---

## Standard Document Structure

Each standard follows a consistent format:

```markdown
# Standard Name

## Quick Reference
[Table of key rules and values]

## [Topic Sections]
[Detailed guidance with examples]

## Common Errors
[Symptom | Cause | Fix table]

## Anti-Patterns
[Named patterns to avoid]

## AI Agent Guidelines
[ALWAYS/NEVER rules table]

## Summary
[Key takeaways list]
```

---

## Configuration Files

| File | Purpose |
|------|---------|
| `.pre-commit-config.yaml` | Hook definitions |
| `.yamllint.yml` | YAML linting rules |
| `.markdownlint.yml` | Markdown linting rules |
| `.shellcheckrc` | Shell linting rules |
| `.prettierrc` | JSON/Markdown formatting |
| `pyproject.toml` | Python tool settings (ruff, pytest) |
| `.golangci.yml` | Go linting configuration |
| `tsconfig.json` | TypeScript compiler settings |
| `eslint.config.js` | TypeScript/JS linting |

---

## Contributing

To update a standard:

1. First update the canonical source in `gitops/docs/standards/`
2. Copy here and update the sync date header
3. Customize any project-specific sections as needed
4. Run validation to ensure format compliance

### Adding a New Standard

1. Follow the document structure template above
2. Include: Quick Reference, Common Errors, Anti-Patterns, AI Guidelines
3. Add to this README's index table
4. Update related links in other standards

---

**Created:** 2024-12-30
**Last Updated:** 2026-01-19
**Related:** [Canonical Standards](https://github.com/your-org/gitops/tree/main/docs/standards)
