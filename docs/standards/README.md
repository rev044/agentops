# Standards Index

> Coding standards for this repository.

## Quick Reference

| Language | Document | Gate |
|----------|----------|------|
| Python | [python-style-guide.md](./python-style-guide.md) | CC ≤ 10 |
| Shell | [shell-script-standards.md](./shell-script-standards.md) | shellcheck |
| YAML | [yaml-helm-standards.md](./yaml-helm-standards.md) | yamllint |

## Complexity Requirements

| Language | Tool | Threshold |
|----------|------|-----------|
| Python | radon/xenon | CC ≤ 10 |
| Shell | shellcheck | Pass all |
| YAML | yamllint | Pass all |

## Pre-commit Checks

```bash
# Python
black --check scripts/
ruff check scripts/
xenon scripts/ --max-absolute B

# Shell
shellcheck scripts/*.sh

# YAML
yamllint .
```
